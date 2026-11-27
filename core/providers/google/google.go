// Package google is the Google provider: Gmail and Google Calendar of a
// Google account over their REST APIs. Sign-in is OAuth 2.0 for installed
// apps: the browser approves, Google redirects to a loopback address the
// core listens on (PKCE protects the code), and the refresh token goes to
// the vault. The OAuth client (a "Desktop app" client of a Google Cloud
// project with the Gmail and Calendar APIs enabled) is built in with
// GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET or asked for at sign-in.
//
// Mail ids are Gmail message ids; calendar ids and event ids are Google's.
package google

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/SelfHold/parley/core/crypto"
	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// Type is the provider type.
const Type = "google"

// Built-in OAuth client (make GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=...).
// A desktop client's secret is not a secret in Google's sense: it ships
// inside every installed app.
var (
	DefaultClientID     string
	DefaultClientSecret string
)

var scopes = []string{
	"openid", "email", "profile",
	"https://www.googleapis.com/auth/gmail.modify",
	"https://www.googleapis.com/auth/calendar",
}

// Provider is the Google provider. The URLs point elsewhere in tests.
type Provider struct {
	AuthURL  string // OAuth consent page
	TokenURL string
	API      string // REST APIs (Gmail, Calendar, userinfo), ending in "/"
	// Poll is how often mail is checked for changes; calendars every
	// third time.
	Poll time.Duration
}

// New returns the provider.
func New() *Provider {
	return &Provider{
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
		API:      "https://www.googleapis.com/",
		Poll:     time.Minute,
	}
}

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "Google",
		Methods: []provider.AuthMethod{{ID: "oauth", Name: "Google account", Description: "Gmail and Google Calendar; approve in the browser"}},
		Caps:    []provider.Capability{provider.CapMail, provider.CapCalendar},
	}
}

// --- sign-in -----------------------------------------------------------------

type login struct {
	p    *Provider
	http *http.Client
	prev *provider.Account

	clientID, secret string
	redirect, state  string
	verifier         string

	mu     sync.Mutex
	srv    *http.Server
	codeCh chan string
}

func (p *Provider) Login(_ context.Context, _ string, acct *provider.Account, env provider.Env) (provider.Login, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	return &login{p: p, http: hc, prev: acct}, nil
}

func (l *login) prevParam(k string) string {
	if l.prev == nil {
		return ""
	}
	return l.prev.Params[k]
}

func (l *login) clientStep(errMsg string) provider.Step {
	return provider.Step{
		Title: "Google — OAuth client",
		Message: "This build has no Google OAuth client. In console.cloud.google.com create a project, enable the Gmail API and the " +
			"Google Calendar API, set up the OAuth consent screen (add yourself as a test user) and create an OAuth client ID " +
			"of type \"Desktop app\".",
		Error: errMsg,
		Fields: []provider.Field{
			{Key: "client_id", Label: "Client ID", Default: l.prevParam("client_id")},
			{Key: "client_secret", Label: "Client secret", Secret: true, Default: l.prevParam("client_secret")},
		},
	}
}

func (l *login) authURL() string {
	sum := sha256.Sum256([]byte(l.verifier))
	q := url.Values{
		"client_id": {l.clientID}, "redirect_uri": {l.redirect}, "response_type": {"code"},
		"scope": {strings.Join(scopes, " ")}, "state": {l.state},
		"code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"},
		"access_type": {"offline"}, "prompt": {"consent"},
	}
	if l.prev != nil && l.prev.Login != "" {
		q.Set("login_hint", l.prev.Login)
	}
	return l.p.AuthURL + "?" + q.Encode()
}

func (l *login) waitStep(errMsg string) provider.Step {
	return provider.Step{
		Title:   "Google — approve",
		Message: "Choose your account and allow access to mail and calendar in the browser. If the page you land on does not open, copy its full address here.",
		Error:   errMsg,
		OpenURL: l.authURL(),
		Fields:  []provider.Field{{Key: "callback", Label: "Address after approving", Optional: true}},
	}
}

func (l *login) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if l.redirect == "" {
		id, secret := l.prevParam("client_id"), l.prevParam("client_secret")
		if id == "" {
			id, secret = DefaultClientID, DefaultClientSecret
		}
		if values != nil && values["client_id"] != "" {
			id, secret = strings.TrimSpace(values["client_id"]), strings.TrimSpace(values["client_secret"])
		} else if id == "" {
			if values == nil {
				return l.clientStep(""), nil
			}
			return l.clientStep("Client ID is required"), nil
		}
		l.clientID, l.secret = id, secret
		if err := l.listen(); err != nil {
			return l.clientStep(err.Error()), nil
		}
		return l.waitStep(""), nil
	}
	code := ""
	if cb := strings.TrimSpace(values["callback"]); cb != "" {
		u, err := url.Parse(cb)
		if err != nil {
			return l.waitStep("Not an address"), nil
		}
		q := u.Query()
		if e := q.Get("error"); e != "" {
			return l.waitStep("Google: " + e), nil
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
		return l.waitStep("Google did not return a code (access denied?)"), nil
	}
	return l.exchange(ctx, code)
}

// listen serves the loopback redirect on a free port.
func (l *login) listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("cannot listen for the redirect: %v", err)
	}
	l.state, l.verifier = crypto.RandomID(16), crypto.RandomID(48)
	l.codeCh = make(chan string, 1)
	l.redirect = fmt.Sprintf("http://127.0.0.1:%d/", ln.Addr().(*net.TCPAddr).Port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		if q.Get("error") != "" {
			fmt.Fprint(w, "<p>Access was not granted. You can close this page.</p>")
			return
		}
		fmt.Fprint(w, "<p>Signed in. You can close this page and go back to Parley.</p>")
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	l.mu.Lock()
	l.srv = srv
	l.mu.Unlock()
	go srv.Serve(ln)
	return nil
}

type tokenReply struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	Description  string `json:"error_description"`
}

func (p *Provider) token(ctx context.Context, hc *http.Client, form url.Values) (tokenReply, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenReply{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := hc.Do(req)
	if err != nil {
		return tokenReply{}, err
	}
	defer resp.Body.Close()
	var tr tokenReply
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tr); err != nil {
		return tokenReply{}, fmt.Errorf("google: token: %s", resp.Status)
	}
	if tr.Error != "" {
		err := fmt.Errorf("google: %s %s", tr.Error, tr.Description)
		if tr.Error == "invalid_grant" || tr.Error == "unauthorized_client" || tr.Error == "invalid_client" {
			err = fmt.Errorf("%w: %v", provider.ErrAuth, err)
		}
		return tokenReply{}, err
	}
	if tr.AccessToken == "" {
		return tokenReply{}, fmt.Errorf("google: token: %s", resp.Status)
	}
	return tr, nil
}

type userInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (l *login) exchange(ctx context.Context, code string) (provider.Step, error) {
	tr, err := l.p.token(ctx, l.http, url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {l.redirect},
		"client_id": {l.clientID}, "client_secret": {l.secret}, "code_verifier": {l.verifier},
	})
	if err != nil {
		return l.waitStep(err.Error()), nil
	}
	if tr.RefreshToken == "" {
		return provider.Step{}, errors.New("google: no refresh token in the answer")
	}
	c := &client{p: l.p, http: l.http, access: tr.AccessToken, expiry: time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)}
	var ui userInfo
	if err := c.do(ctx, http.MethodGet, "oauth2/v3/userinfo", nil, &ui); err != nil {
		return provider.Step{}, err
	}
	if ui.Email == "" {
		return provider.Step{}, errors.New("google: the account has no e-mail address")
	}
	l.Close()
	params := map[string]string{"refresh_token": tr.RefreshToken}
	if l.clientID != DefaultClientID {
		params["client_id"], params["client_secret"] = l.clientID, l.secret
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: strings.ToLower(ui.Email), Name: ui.Email, Params: params,
	}}, nil
}

func (l *login) Close() {
	l.mu.Lock()
	srv := l.srv
	l.srv = nil
	l.mu.Unlock()
	if srv != nil {
		srv.Close()
	}
}

// --- API client --------------------------------------------------------------

// client calls the REST APIs with an access token it renews by itself.
type client struct {
	p    *Provider
	http *http.Client

	clientID, secret, refresh string

	mu     sync.Mutex
	access string
	expiry time.Time
}

func (c *client) token(ctx context.Context, force bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !force && c.access != "" && time.Until(c.expiry) > time.Minute {
		return c.access, nil
	}
	if c.refresh == "" {
		return "", fmt.Errorf("google: %w", provider.ErrAuth)
	}
	tr, err := c.p.token(ctx, c.http, url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {c.refresh},
		"client_id": {c.clientID}, "client_secret": {c.secret},
	})
	if err != nil {
		return "", err
	}
	c.access, c.expiry = tr.AccessToken, time.Now().Add(time.Duration(tr.ExpiresIn)*time.Second)
	return c.access, nil
}

// apiError is a non-2xx answer.
type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string { return fmt.Sprintf("google: %d %s", e.Status, e.Message) }

func isStatus(err error, code int) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == code
}

// do calls path (relative to API, or an absolute URL) with a JSON body and
// decodes the JSON answer into out (nil: dropped).
func (c *client) do(ctx context.Context, method, path string, body, out any) error {
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			return err
		}
	}
	u := path
	if !strings.HasPrefix(u, "http") {
		u = c.p.API + path
	}
	for attempt := 0; ; attempt++ {
		tok, err := c.token(ctx, attempt > 0)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			continue // the token may have just expired
		}
		if resp.StatusCode/100 != 2 {
			var e struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			_ = json.Unmarshal(data, &e)
			ae := &apiError{Status: resp.StatusCode, Message: e.Error.Message}
			if resp.StatusCode == http.StatusUnauthorized {
				return fmt.Errorf("%w: %v", provider.ErrAuth, ae)
			}
			return ae
		}
		if out == nil || len(data) == 0 {
			return nil
		}
		return json.Unmarshal(data, out)
	}
}

// --- session -----------------------------------------------------------------

type session struct {
	p   *Provider
	c   *client
	env provider.Env

	mu        sync.Mutex
	self      model.Address
	historyID string
	calSeen   time.Time            // calendar changes are looked for after this
	rules     map[string]ruleEntry // series rules by calendar + series id
}

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	id, secret := acct.Params["client_id"], acct.Params["client_secret"]
	if id == "" {
		id, secret = DefaultClientID, DefaultClientSecret
	}
	c := &client{p: p, http: hc, clientID: id, secret: secret, refresh: acct.Params["refresh_token"]}
	return &session{p: p, c: c, env: env, self: model.Address{Email: acct.Login}}, nil
}

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	var ui userInfo
	if err := s.c.do(ctx, http.MethodGet, "oauth2/v3/userinfo", nil, &ui); err != nil {
		return err
	}
	s.mu.Lock()
	s.self = model.Address{Name: ui.Name, Email: ui.Email}
	s.calSeen = time.Now()
	s.mu.Unlock()
	if err := s.resetHistory(ctx); err != nil {
		return err
	}
	self := model.User{ID: ui.Sub, Name: ui.Name, Username: ui.Email}
	if self.Name == "" {
		self.Name = ui.Email
	}
	if ui.Picture != "" {
		self.Avatar = "photo:" + ui.Picture
	}
	sink.Connected(self)
	tick := time.NewTicker(s.p.Poll)
	defer tick.Stop()
	for n := 1; ; n++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
		if err := s.pollMail(ctx, sink); err != nil {
			return err
		}
		if n%3 == 0 {
			if err := s.pollCalendars(ctx, sink); err != nil {
				return err
			}
		}
	}
}

// Media fetches attachments ("att:<message>:<attachment>") and the
// account's photo ("photo:<url>", without credentials).
func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	kind, rest, _ := strings.Cut(ref, ":")
	switch kind {
	case "photo":
		u, err := url.Parse(rest)
		if err != nil || u.Scheme != "https" || !strings.HasSuffix(u.Hostname(), ".googleusercontent.com") {
			return nil, "", fmt.Errorf("google: bad photo ref")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rest, nil)
		if err != nil {
			return nil, "", err
		}
		resp, err := s.c.http.Do(req)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("google: photo: %s", resp.Status)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
		if int64(len(data)) > limit {
			return nil, "", provider.ErrTooLarge
		}
		return data, resp.Header.Get("Content-Type"), err
	case "att":
		return s.attachment(ctx, rest, limit)
	}
	return nil, "", fmt.Errorf("google: unknown media ref")
}
