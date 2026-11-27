// Package slack is the Slack provider. A workspace signs in either with
// the browser session (xoxc token + "d" cookie, what the web client uses;
// needs nothing from the workspace admin; live events over RTM), or with
// an official OAuth user token of your own Slack app (live events over
// Socket Mode when an app-level token is given, else by polling).
//
// Chat ids are Slack's conversation ids; message ids are message "ts".
package slack

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SelfHold/parley/core/crypto"
	"github.com/SelfHold/parley/core/provider"
)

// Type is the provider type.
const Type = "slack"

const (
	methodSession = "session"
	methodOAuth   = "oauth"
)

// userScopes is what the OAuth app asks for.
var userScopes = []string{
	"channels:history", "groups:history", "im:history", "mpim:history",
	"channels:read", "groups:read", "im:read", "mpim:read",
	"channels:write", "groups:write", "im:write", "mpim:write",
	"users:read", "chat:write",
}

// Provider is the Slack provider.
type Provider struct {
	// SlackURL is the OAuth host (tests point it elsewhere).
	SlackURL string
}

// New returns the provider.
func New() *Provider { return &Provider{SlackURL: "https://slack.com/"} }

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "Slack",
		Methods: []provider.AuthMethod{
			{ID: methodSession, Name: "Browser session", Description: "Token and cookie of the Slack web client; works in any workspace"},
			{ID: methodOAuth, Name: "Slack app (OAuth)", Description: "Your own Slack app installed in the workspace"},
		},
		Caps: []provider.Capability{provider.CapMessages},
	}
}

// client returns the API client of an account's params.
func (p *Provider) client(params map[string]string, hc *http.Client) *api {
	if params["method"] == methodOAuth {
		return &api{base: p.SlackURL + "api/", token: params["token"], http: hc}
	}
	return &api{base: params["workspace_url"] + "api/", token: params["token"], cookie: params["cookie"], http: hc}
}

type authTest struct {
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

func credentials(params map[string]string, at authTest) *provider.Credentials {
	params["team_id"], params["user_id"] = at.TeamID, at.UserID
	if at.URL != "" && params["method"] != methodOAuth {
		params["workspace_url"] = at.URL
	}
	return &provider.Credentials{Login: at.TeamID + ":" + at.UserID, Name: at.Team, Params: params}
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
	switch method {
	case methodSession, "":
		return &sessionLogin{p: p, http: hc, prev: prev}, nil
	case methodOAuth:
		return &oauthLogin{p: p, http: hc, prev: prev}, nil
	}
	return nil, fmt.Errorf("slack: unknown sign-in method %q", method)
}

// --- browser session -----------------------------------------------------

type sessionLogin struct {
	p    *Provider
	http *http.Client
	prev map[string]string
}

func (l *sessionLogin) step(errMsg string) provider.Step {
	ws := strings.TrimSuffix(strings.TrimPrefix(l.prev["workspace_url"], "https://"), "/")
	return provider.Step{
		Title: "Slack — browser session",
		Message: "Open the workspace in a browser and sign in. In the developer tools: the token is in " +
			"Application → Local Storage → localConfig_v2 (teams → token, starts with xoxc-); " +
			"the cookie is Application → Cookies → \"d\" (starts with xoxd-).",
		Error: errMsg,
		Fields: []provider.Field{
			{Key: "workspace", Label: "Workspace", Help: "acme, acme.slack.com or its URL", Default: ws},
			{Key: "token", Label: "Token (xoxc-…)", Secret: true},
			{Key: "cookie", Label: "Cookie d (xoxd-…)", Secret: true},
		},
	}
}

func (l *sessionLogin) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if values == nil {
		return l.step(""), nil
	}
	ws, err := workspaceURL(values["workspace"])
	if err != nil {
		return l.step(err.Error()), nil
	}
	token, cookie := strings.TrimSpace(values["token"]), strings.TrimSpace(values["cookie"])
	if cookie != "" {
		// Pasted from the browser it is usually URL-encoded already.
		if dec, err := url.QueryUnescape(cookie); err == nil {
			cookie = dec
		}
	}
	if !strings.HasPrefix(token, "xox") || cookie == "" {
		return l.step("Enter the token (xoxc-…) and the d cookie"), nil
	}
	params := map[string]string{"method": methodSession, "workspace_url": ws, "token": token, "cookie": cookie}
	var at authTest
	err = l.p.client(params, l.http).call(ctx, "auth.test", nil, &at)
	if isAuth(err) {
		return l.step("Slack did not accept the token and cookie (" + err.Error() + ")"), nil
	}
	if err != nil {
		return l.step(err.Error()), nil
	}
	return provider.Step{Done: true, Result: credentials(params, at)}, nil
}

func (l *sessionLogin) Close() {}

// --- OAuth ---------------------------------------------------------------

// oauthLogin asks for the app's credentials, sends the browser to Slack and
// takes the code from a redirect to this device (http://localhost) or from
// the address the user pastes (for an https redirect).
type oauthLogin struct {
	p    *Provider
	http *http.Client
	prev map[string]string

	clientID, secret, appToken, redirect, state string

	mu     sync.Mutex
	srv    *http.Server
	codeCh chan string
}

const defaultRedirect = "http://localhost:53682/slack/callback"

func (l *oauthLogin) appStep(errMsg string) provider.Step {
	redirect := l.prev["redirect_url"]
	if redirect == "" {
		redirect = defaultRedirect
	}
	return provider.Step{
		Title: "Slack — your app",
		Message: "Create an app at api.slack.com/apps, add the redirect URL below under OAuth & Permissions and the user scopes " +
			strings.Join(userScopes, ", ") + ". For live updates enable Socket Mode, subscribe to message.* events on behalf of users and create an app-level token (xapp-…).",
		Error: errMsg,
		Fields: []provider.Field{
			{Key: "client_id", Label: "Client ID", Default: l.prev["client_id"]},
			{Key: "client_secret", Label: "Client secret", Secret: true, Default: l.prev["client_secret"]},
			{Key: "app_token", Label: "App-level token (xapp-…)", Secret: true, Optional: true, Help: "For Socket Mode; without it new messages are polled", Default: l.prev["app_token"]},
			{Key: "redirect_url", Label: "Redirect URL", Default: redirect},
		},
	}
}

func (l *oauthLogin) authURL() string {
	q := url.Values{
		"client_id": {l.clientID}, "user_scope": {strings.Join(userScopes, ",")},
		"redirect_uri": {l.redirect}, "state": {l.state},
	}
	return l.p.SlackURL + "oauth/v2/authorize?" + q.Encode()
}

func (l *oauthLogin) waitStep(errMsg string) provider.Step {
	return provider.Step{
		Title:   "Slack — approve",
		Message: "Approve the app in the browser. If the page you land on does not load, copy its full address here.",
		Error:   errMsg,
		OpenURL: l.authURL(),
		Fields:  []provider.Field{{Key: "callback", Label: "Address after approving", Optional: true}},
	}
}

func (l *oauthLogin) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if l.clientID == "" {
		if values == nil {
			return l.appStep(""), nil
		}
		return l.configure(values)
	}
	code := ""
	if cb := strings.TrimSpace(values["callback"]); cb != "" {
		u, err := url.Parse(cb)
		if err != nil {
			return l.waitStep("Not an address"), nil
		}
		q := u.Query()
		if e := q.Get("error"); e != "" {
			return l.waitStep("Slack: " + e), nil
		}
		if q.Get("state") != l.state {
			return l.waitStep("This address belongs to another sign-in"), nil
		}
		code = q.Get("code")
	}
	if code == "" {
		select {
		case code = <-l.codeCh:
		case <-ctx.Done():
			return l.waitStep(""), nil // still waiting for the browser
		}
	}
	if code == "" {
		return l.waitStep("Slack did not return a code"), nil
	}
	return l.exchange(ctx, code)
}

func (l *oauthLogin) configure(values map[string]string) (provider.Step, error) {
	l.clientID = strings.TrimSpace(values["client_id"])
	l.secret = strings.TrimSpace(values["client_secret"])
	l.appToken = strings.TrimSpace(values["app_token"])
	l.redirect = strings.TrimSpace(values["redirect_url"])
	if l.redirect == "" {
		l.redirect = defaultRedirect
	}
	switch {
	case l.clientID == "" || l.secret == "":
		l.clientID = ""
		return l.appStep("Client ID and secret are required"), nil
	case l.appToken != "" && !strings.HasPrefix(l.appToken, "xapp-"):
		l.clientID = ""
		return l.appStep("The app-level token starts with xapp-"), nil
	}
	l.state = crypto.RandomID(16)
	l.codeCh = make(chan string, 1)
	if err := l.listen(); err != nil {
		l.clientID = ""
		return l.appStep(err.Error()), nil
	}
	return l.waitStep(""), nil
}

// listen serves the redirect when it points to this device over http.
func (l *oauthLogin) listen() error {
	u, err := url.Parse(l.redirect)
	if err != nil {
		return fmt.Errorf("bad redirect URL: %v", err)
	}
	host := u.Hostname()
	if u.Scheme != "http" || (host != "localhost" && host != "127.0.0.1") {
		return nil // the user pastes the address
	}
	port := u.Port()
	if port == "" {
		port = "80"
	}
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return fmt.Errorf("cannot listen for the redirect on port %s: %v", port, err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != l.state {
			http.Error(w, "unknown sign-in", http.StatusBadRequest)
			return
		}
		select {
		case l.codeCh <- q.Get("code"):
		default:
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<p>Signed in. You can close this page and go back to Parley.</p>")
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	l.mu.Lock()
	l.srv = srv
	l.mu.Unlock()
	go srv.Serve(ln)
	return nil
}

type oauthAccess struct {
	AuthedUser struct {
		ID           string `json:"id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	} `json:"authed_user"`
	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

func (l *oauthLogin) exchange(ctx context.Context, code string) (provider.Step, error) {
	var acc oauthAccess
	a := &api{base: l.p.SlackURL + "api/", http: l.http}
	err := a.call(ctx, "oauth.v2.access", url.Values{
		"client_id": {l.clientID}, "client_secret": {l.secret}, "code": {code}, "redirect_uri": {l.redirect},
	}, &acc)
	if err != nil {
		return l.waitStep(err.Error()), nil
	}
	if acc.AuthedUser.AccessToken == "" {
		return provider.Step{}, errors.New("slack: no user token in the answer (are user scopes set?)")
	}
	params := map[string]string{
		"method": methodOAuth, "token": acc.AuthedUser.AccessToken,
		"client_id": l.clientID, "client_secret": l.secret, "redirect_url": l.redirect,
	}
	if l.appToken != "" {
		params["app_token"] = l.appToken
	}
	setRefresh(params, acc.AuthedUser.RefreshToken, acc.AuthedUser.ExpiresIn)
	var at authTest
	if err := l.p.client(params, l.http).call(ctx, "auth.test", nil, &at); err != nil {
		return provider.Step{}, err
	}
	l.Close()
	return provider.Step{Done: true, Result: credentials(params, at)}, nil
}

// setRefresh records a rotating token's refresh token and expiry.
func setRefresh(params map[string]string, refresh string, expiresIn int) {
	if refresh == "" {
		return
	}
	params["refresh_token"] = refresh
	params["expires"] = strconv.FormatInt(time.Now().Add(time.Duration(expiresIn)*time.Second).Unix(), 10)
}

func (l *oauthLogin) Close() {
	l.mu.Lock()
	srv := l.srv
	l.srv = nil
	l.mu.Unlock()
	if srv != nil {
		srv.Close()
	}
}
