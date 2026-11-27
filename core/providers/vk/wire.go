package vk

// Everything about the VK (vk.com) messenger API lives in this file.
//
// VK has no public user-facing messages API: messages.* is open only to a
// handful of official applications. This provider therefore signs in and
// talks to the API as the official "VK Messenger" desktop client does; its
// application id and secret, its API version and its endpoints below are
// taken from that client (unpacked app.asar, package main). This is how
// Kate Mobile and similar clients work; it is against VK's terms, and VK
// may rotate the key at any time, after which sign-in breaks until these
// constants are updated.
//
// Token logic (as the official client does it):
//   - oauth.vk.com/get_anonym_token (app id + secret) → an anonymous token
//     with an expiry; refreshed when it runs out.
//   - QR sign-in: auth.getAuthCode (anonymous token) → an auth URL to show
//     as a QR code and an auth_hash; auth.checkAuthCode is polled until the
//     user approves it in VK on their phone and it yields an access_token.
//   - auth.getExchangeToken (access_token) → a per-user "common_token"
//     (exchange token) kept for silent re-login.
//   - re-login without the QR: oauth.vk.com/auth_by_exchange_token
//     (common_token) → a silent token; auth.exchangeSilentAuthToken
//     (anonymous token + silent token) → a fresh access_token. Error 1114
//     means the anonymous token expired: refetch it and retry.
//
// UNVERIFIED beyond the auth flow: the messages.* mapping below follows the
// public VK API (getConversations, getHistory, send, getLongPollServer and
// the long-poll stream). Field names are pinned by wire_test.go.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// From the VK Messenger desktop client (package main).
	clientID     = "51453752"
	clientSecret = "4UyuCUsdK8pVCNoeQuGi"
	apiVersion   = "5.231"
	userAgent    = "VKDesktopMessenger/6.0 (parley)"

	defaultAPIHost   = "api.vk.me"  // messages.*, auth.exchange*, users.get
	defaultAuthHost  = "api.vk.com" // auth.getAuthCode / auth.checkAuthCode
	defaultOAuthHost = "oauth.vk.com"

	// Long-poll stream version and mode (VK docs). mode 2|8|64|128 asks for
	// attachments, extra fields, random_id and message edits.
	pollVersion = "19"
	pollMode    = "234"
)

// QR approval status (ca enum in the client).
const (
	qrCreated  = 0
	qrOpened   = 1
	qrApproved = 2
	qrDeclined = 3
	qrExpired  = 4
)

// server holds where the API lives and how to reach it. Tests point every
// host at one httptest server.
type server struct {
	scheme    string // "https"; tests use "http"
	apiHost   string
	authHost  string
	oauthHost string
	version   string
	http      *http.Client
}

func (p *Provider) server(hc *http.Client) *server {
	scheme := p.scheme
	if scheme == "" {
		scheme = "https"
	}
	s := &server{
		scheme:    scheme,
		apiHost:   firstNonEmpty(p.apiHost, defaultAPIHost),
		authHost:  firstNonEmpty(p.authHost, defaultAuthHost),
		oauthHost: firstNonEmpty(p.oauthHost, defaultOAuthHost),
		version:   apiVersion,
		http:      hc,
	}
	return s
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// apiError is VK's {"error":{...}}.
type apiError struct {
	Code int    `json:"error_code"`
	Msg  string `json:"error_msg"`
}

func (e *apiError) Error() string { return fmt.Sprintf("vk: error %d: %s", e.Code, e.Msg) }

// authFailed reports whether an API error means the access token no longer
// works (and a refresh should be tried).
func authFailed(err error) bool {
	var ae *apiError
	return errors.As(err, &ae) && (ae.Code == 5 || ae.Code == 1116 || ae.Code == 1117)
}

// post sends a form POST and returns the decoded body. VK answers are JSON
// either way; GET would work too but the client POSTs.
func (s *server) post(ctx context.Context, host, path string, params map[string]string) (json.RawMessage, error) {
	form := url.Values{}
	for k, v := range params {
		if v != "" {
			form.Set(k, v)
		}
	}
	u := (&url.URL{Scheme: s.scheme, Host: host, Path: path}).String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vk: %s: HTTP %d", path, resp.StatusCode)
	}
	return body, nil
}

// method calls api.<name> with v and access_token and unwraps {"response"}.
func (s *server) method(ctx context.Context, name, accessToken string, params map[string]string) (json.RawMessage, error) {
	return s.methodOn(ctx, s.apiHost, name, accessToken, params)
}

func (s *server) methodOn(ctx context.Context, host, name, accessToken string, params map[string]string) (json.RawMessage, error) {
	p := map[string]string{"v": s.version, "lang": "ru"}
	for k, v := range params {
		p[k] = v
	}
	if accessToken != "" {
		p["access_token"] = accessToken
	}
	body, err := s.post(ctx, host, "/method/"+name, p)
	if err != nil {
		return nil, err
	}
	var env struct {
		Response json.RawMessage `json:"response"`
		Error    *apiError       `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("vk: %s: bad response: %w", name, err)
	}
	if env.Error != nil {
		return nil, env.Error
	}
	return env.Response, nil
}

// --- Token logic ---------------------------------------------------------

// anonymToken is oauth.vk.com/get_anonym_token's answer.
type anonymToken struct {
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expired_at"`
}

func (t anonymToken) valid() bool {
	return t.Token != "" && (t.ExpiredAt == 0 || time.Now().Unix() < t.ExpiredAt)
}

func (s *server) getAnonymToken(ctx context.Context) (anonymToken, error) {
	body, err := s.post(ctx, s.oauthHost, "/get_anonym_token", map[string]string{
		"v": s.version, "client_id": clientID, "client_secret": clientSecret,
	})
	if err != nil {
		return anonymToken{}, err
	}
	var t anonymToken
	if err := json.Unmarshal(body, &t); err != nil {
		return anonymToken{}, fmt.Errorf("vk: get_anonym_token: %w", err)
	}
	if t.Token == "" {
		return anonymToken{}, errors.New("vk: get_anonym_token: empty token")
	}
	return t, nil
}

// authCode is auth.getAuthCode's answer: the URL to show as a QR code.
type authCode struct {
	AuthURL  string `json:"auth_url"`
	AuthHash string `json:"auth_hash"`
}

func (s *server) getAuthCode(ctx context.Context, anonym, deviceName string) (authCode, error) {
	raw, err := s.methodOn(ctx, s.authHost, "auth.getAuthCode", "", map[string]string{
		"client_id": clientID, "device_name": deviceName,
		"auth_code_flow": "0", "anonymous_token": anonym,
	})
	if err != nil {
		return authCode{}, err
	}
	var c authCode
	if err := json.Unmarshal(raw, &c); err != nil {
		return authCode{}, fmt.Errorf("vk: auth.getAuthCode: %w", err)
	}
	return c, nil
}

// qrStatus is auth.checkAuthCode's answer.
type qrStatus struct {
	Status      int    `json:"status"`
	AccessToken string `json:"access_token"`
}

func (s *server) checkAuthCode(ctx context.Context, anonym, authHash string) (qrStatus, error) {
	raw, err := s.methodOn(ctx, s.authHost, "auth.checkAuthCode", "", map[string]string{
		"client_id": clientID, "auth_hash": authHash, "anonymous_token": anonym,
	})
	if err != nil {
		return qrStatus{}, err
	}
	var st qrStatus
	if err := json.Unmarshal(raw, &st); err != nil {
		return qrStatus{}, fmt.Errorf("vk: auth.checkAuthCode: %w", err)
	}
	return st, nil
}

// userToken is one entry of auth.getExchangeToken's users_exchange_tokens.
type userToken struct {
	UserID         int64  `json:"user_id"`
	CommonToken    string `json:"common_token"`
	ProfileType    int    `json:"profile_type"`
	SilentToken    string `json:"silent_token"`
	SilentTokenUID string `json:"silent_token_uuid"`
}

func (s *server) getExchangeToken(ctx context.Context, accessToken string) ([]userToken, error) {
	raw, err := s.method(ctx, "auth.getExchangeToken", accessToken, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Users []userToken `json:"users_exchange_tokens"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("vk: auth.getExchangeToken: %w", err)
	}
	return r.Users, nil
}

// silentToken is oauth.vk.com/auth_by_exchange_token's answer.
type silentToken struct {
	SilentToken    string `json:"silent_token"`
	SilentTokenUID string `json:"silent_token_uuid"`
}

func (s *server) authByExchangeToken(ctx context.Context, commonToken string) (silentToken, error) {
	body, err := s.post(ctx, s.oauthHost, "/auth_by_exchange_token", map[string]string{
		"v": s.version, "client_id": clientID, "exchange_token": commonToken,
	})
	if err != nil {
		return silentToken{}, err
	}
	// The endpoint wraps its payload in {"response":{...}}.
	var env struct {
		Response silentToken `json:"response"`
		Error    *apiError   `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return silentToken{}, fmt.Errorf("vk: auth_by_exchange_token: %w", err)
	}
	if env.Error != nil {
		return silentToken{}, env.Error
	}
	if env.Response.SilentToken == "" {
		// Some responses are flat, not wrapped.
		var flat silentToken
		_ = json.Unmarshal(body, &flat)
		return flat, nil
	}
	return env.Response, nil
}

// exchangeSilentAuthToken turns a silent token into an access token. It
// returns errAnonymExpired when the anonymous token must be refetched.
var errAnonymExpired = errors.New("vk: anonymous token expired")

func (s *server) exchangeSilentAuthToken(ctx context.Context, anonym, silent, uuid string) (string, error) {
	raw, err := s.method(ctx, "auth.exchangeSilentAuthToken", anonym, map[string]string{
		"token": silent, "uuid": uuid,
	})
	var ae *apiError
	if errors.As(err, &ae) && ae.Code == 1114 {
		return "", errAnonymExpired
	}
	if err != nil {
		return "", err
	}
	var r struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", fmt.Errorf("vk: auth.exchangeSilentAuthToken: %w", err)
	}
	if r.AccessToken == "" {
		return "", errors.New("vk: auth.exchangeSilentAuthToken: empty token")
	}
	return r.AccessToken, nil
}

// vkUser is users.get's shape (and the profiles of messages.*).
type vkUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ScreenneM string `json:"screen_name"`
	Photo     string `json:"photo_100"`
	Deactiv   string `json:"deactivated"`
	Online    int    `json:"online"`
}

func (u vkUser) name() string {
	n := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if n == "" {
		n = u.ScreenneM
	}
	return n
}

func (s *server) usersGet(ctx context.Context, accessToken, ids string) ([]vkUser, error) {
	raw, err := s.method(ctx, "users.get", accessToken, map[string]string{
		"user_ids": ids, "fields": "photo_100,screen_name,online",
	})
	if err != nil {
		return nil, err
	}
	var users []vkUser
	if err := json.Unmarshal(raw, &users); err != nil {
		return nil, fmt.Errorf("vk: users.get: %w", err)
	}
	return users, nil
}

// --- Long poll -----------------------------------------------------------

type longPollServer struct {
	Server string `json:"server"`
	Key    string `json:"key"`
	TS     int64  `json:"ts"`
}

func (s *server) getLongPollServer(ctx context.Context, accessToken string) (longPollServer, error) {
	raw, err := s.method(ctx, "messages.getLongPollServer", accessToken, map[string]string{
		"need_pts": "0", "lp_version": pollVersion,
	})
	if err != nil {
		return longPollServer{}, err
	}
	var lp longPollServer
	if err := json.Unmarshal(raw, &lp); err != nil {
		return longPollServer{}, fmt.Errorf("vk: getLongPollServer: %w", err)
	}
	return lp, nil
}

// longPollReply is one a_check answer. On failed != 0 the caller reacts:
// 1 → keep the new ts; 2/3 → fetch a fresh server; 4 → unsupported version.
type longPollReply struct {
	TS      json.Number       `json:"ts"`
	Updates []json.RawMessage `json:"updates"`
	Failed  int               `json:"failed"`
}

// poll makes one long-poll request. server is the bare host[/path] VK gave.
func (s *server) poll(ctx context.Context, lp longPollServer, wait int) (longPollReply, error) {
	host := lp.Server
	if !strings.Contains(host, "://") {
		host = s.scheme + "://" + host
	}
	u, err := url.Parse(host)
	if err != nil {
		return longPollReply{}, err
	}
	q := u.Query()
	q.Set("act", "a_check")
	q.Set("key", lp.Key)
	q.Set("ts", strconv.FormatInt(lp.TS, 10))
	q.Set("wait", strconv.Itoa(wait))
	q.Set("mode", pollMode)
	q.Set("version", pollVersion)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return longPollReply{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.http.Do(req)
	if err != nil {
		return longPollReply{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return longPollReply{}, err
	}
	var r longPollReply
	if err := json.Unmarshal(body, &r); err != nil {
		return longPollReply{}, fmt.Errorf("vk: long poll: %w", err)
	}
	return r, nil
}

// randomID is messages.send's idempotency id (int32, non-zero).
func randomID() string {
	n := rand.Int31()
	if n == 0 {
		n = 1
	}
	return strconv.FormatInt(int64(n), 10)
}
