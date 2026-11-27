// Package vkteams is the VK WorkSpace messenger (VK Teams) provider, as a
// user account over the client protocol (see wire.go: it is not publicly
// documented). Accounts sign in with login and password (the client key
// of the server is needed), or with the aimsid of a signed-in web client.
// The server is configurable: VK's cloud or a company's on-premises host.
//
// Chat ids are VK Teams "sn": a user's aimId (e-mail) or "<n>@chat.agent"
// for groups. Message ids are the server's history ids.
package vkteams

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SelfHold/parley/core/provider"
)

// Type is the provider type.
const Type = "vkteams"

const (
	methodPassword = "password"
	methodAimsid   = "aimsid"

	keyAimsid  = "aimsid"
	keyFetch   = "fetch_url"
	keyDialogs = "dialogs"
)

// Provider is the VK Teams provider.
type Provider struct {
	// scheme of the server URLs; tests set "http".
	scheme string
}

// New returns the provider.
func New() *Provider { return &Provider{} }

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "VK Teams",
		Methods: []provider.AuthMethod{
			{ID: methodPassword, Name: "Login and password", Description: "Account of VK WorkSpace (cloud or your company's server)"},
			{ID: methodAimsid, Name: "Web client session", Description: "aimsid of a signed-in VK Teams web client"},
		},
		Caps: []provider.Capability{provider.CapMessages},
	}
}

func (p *Provider) server(params map[string]string, hc *http.Client) *server {
	host, ver := params["host"], params["api_version"]
	if host == "" {
		host = defaultHost
	}
	if ver = normVersion(ver); ver == "" {
		ver = defaultVersion
	}
	scheme := p.scheme
	if scheme == "" {
		scheme = "https"
	}
	return &server{scheme: scheme, host: host, version: ver, http: hc}
}

func normHost(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "https://"), "http://")
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return s
}

func (p *Provider) Login(_ context.Context, method string, acct *provider.Account, env provider.Env) (provider.Login, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	prev := map[string]string{}
	if acct != nil {
		prev = acct.Params
	}
	if method == "" {
		method = methodPassword
	}
	if method != methodPassword && method != methodAimsid {
		return nil, errors.New("vk teams: unknown sign-in method " + method)
	}
	return &login{p: p, http: hc, prev: prev, method: method}, nil
}

type login struct {
	p      *Provider
	http   *http.Client
	prev   map[string]string
	method string
}

func (l *login) step(errMsg string) provider.Step {
	host := l.prev["host"]
	if host == "" {
		host = defaultHost
	}
	common := []provider.Field{
		{Key: "host", Label: "Server", Help: "API host of VK Teams; for on-premises installations your company's", Default: host},
		{Key: "api_version", Label: "API version", Optional: true, Help: "As in the client's requests: /api/v132/… (default v132)", Default: l.prev["api_version"]},
	}
	if l.method == methodAimsid {
		return provider.Step{
			Title: "VK Teams — web session",
			Message: "Sign in to the VK Teams web client, open the developer tools → Network, filter by " +
				"\"fetchEvents\" and copy the address of any such request (Copy → Copy URL); paste it here. " +
				"A bare aimsid works only if the server takes it without the address. " +
				"The session lasts until you sign out there.",
			Error: errMsg,
			Fields: append(common, provider.Field{Key: "aimsid", Label: "fetchEvents address or aimsid", Secret: true,
				Help: "https://…/api/v132/bos/…/aim/fetchEvents?aimsid=…"}),
		}
	}
	return provider.Step{
		Title:   "VK Teams",
		Message: "Sign in with your VK WorkSpace account. The password is used once and not stored.",
		Error:   errMsg,
		Fields: append(common,
			provider.Field{Key: "login", Label: "Login (e-mail)", Default: l.prev["login"]},
			provider.Field{Key: "password", Label: "Password", Secret: true},
			provider.Field{Key: "dev_id", Label: "Client key", Help: "The \"k\"/\"devId\" parameter of the official client's requests", Default: l.prev["dev_id"]},
		),
	}
}

func (l *login) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if values == nil {
		return l.step(""), nil
	}
	params := map[string]string{"method": l.method, "host": normHost(values["host"]), "api_version": normVersion(values["api_version"])}
	if l.method == methodAimsid {
		// A pasted fetchEvents address knows the server better than the fields.
		if host, ver, _, _, ok := parseFetchURL(values["aimsid"]); ok {
			params["host"] = host
			if ver != "" {
				params["api_version"] = ver
			}
		}
	}
	if params["host"] == "" {
		return l.step("Enter the server"), nil
	}
	srv := l.p.server(params, l.http)
	if l.method == methodAimsid {
		return l.withAimsid(ctx, srv, params, values["aimsid"])
	}
	return l.withPassword(ctx, srv, params, values)
}

func (l *login) withPassword(ctx context.Context, srv *server, params, values map[string]string) (provider.Step, error) {
	name, pw, dev := strings.TrimSpace(values["login"]), values["password"], strings.TrimSpace(values["dev_id"])
	if name == "" || pw == "" || dev == "" {
		return l.step("Login, password and client key are required"), nil
	}
	lr, err := srv.clientLogin(ctx, dev, name, pw)
	var le *errLogin
	if errors.As(err, &le) {
		return l.step("Wrong login or password (" + le.detail + ")"), nil
	}
	if err != nil {
		return l.step(err.Error()), nil
	}
	key := sessionKey(pw, lr.SessionSecret)
	offset := int64(0)
	if lr.HostTime > 0 {
		offset = lr.HostTime - time.Now().Unix()
	}
	si, err := srv.startSession(ctx, dev, lr.Token.A, key, offset)
	if err != nil {
		return l.step(err.Error()), nil
	}
	params["login"], params["dev_id"], params["token"], params["session_key"] = name, dev, lr.Token.A, key
	params["time_offset"] = strconv.FormatInt(offset, 10)
	id := si.MyInfo.id()
	if id == "" {
		id = name
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: params["host"] + "/" + id, Name: nameOr(si.MyInfo.name(), name), Params: params,
		State: map[string][]byte{keyAimsid: []byte(si.AimSID), keyFetch: []byte(si.FetchBaseURL)},
	}}, nil
}

func (l *login) withAimsid(ctx context.Context, srv *server, params map[string]string, input string) (provider.Step, error) {
	input = strings.TrimSpace(input)
	aimsid, first := input, ""
	if _, _, a, next, ok := parseFetchURL(input); ok {
		aimsid = a
		// Poll the pasted node, over this provider's scheme.
		if u, err := url.Parse(next); err == nil {
			u.Scheme = srv.scheme
			first = u.String()
		}
	} else if a, err := url.QueryUnescape(input); err == nil {
		aimsid = a
	}
	if aimsid == "" {
		return l.step("Enter the fetchEvents address or the aimsid"), nil
	}
	if strings.Contains(aimsid, "*") {
		return l.step("This aimsid is masked (the client's logs hide part of it); copy it from the web client"), nil
	}
	if first == "" {
		first = srv.firstFetchURL(aimsid)
	}
	r, err := srv.fetchEvents(ctx, first, time.Second)
	var se *errSession
	if errors.As(err, &se) {
		return l.step("The server does not know this aimsid (" + se.detail + ")"), nil
	}
	if err != nil {
		if !strings.Contains(input, "fetchEvents") {
			return l.step(err.Error() + "; paste the whole fetchEvents address instead of the aimsid"), nil
		}
		return l.step(err.Error()), nil
	}
	me := person{}
	for _, ev := range r.Events {
		if ev.Type == "myInfo" {
			_ = json.Unmarshal(ev.EventData, &me)
		}
	}
	id, name := me.id(), me.name()
	if id == "" { // a resumed poll does not repeat myInfo
		id = aimsidLogin(aimsid)
		name = id
	}
	if id == "" {
		return l.step("The server answered, but did not say who this session is"), nil
	}
	if name == "" || name == id {
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if p, err := srv.getUserInfo(cctx, aimsid, id); err == nil && p.name() != "" {
			name = p.name()
		}
		cancel()
	}
	next := r.FetchBaseURL
	if next == "" {
		next = first
	}
	params["aimsid"] = aimsid
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: params["host"] + "/" + id, Name: name, Params: params,
		State: map[string][]byte{keyAimsid: []byte(aimsid), keyFetch: []byte(next)},
	}}, nil
}

func (l *login) Close() {}

func nameOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
