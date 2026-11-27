// Package vk is the VK (vk.com) messenger provider: a user's personal VK
// chats (messages.*), signed in by QR code as the official VK Messenger
// desktop client. See wire.go for the API, the token logic and the caveat
// that this talks to VK as that official client.
//
// Chat ids are VK peer ids as decimal strings (a user id, 2000000000+<n>
// for a group chat, a negative community id). Message ids are VK's global
// message ids.
package vk

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"megachat/core/provider"
)

// Type is the provider type.
const Type = "vk"

const methodQR = "qr"

const (
	keyAccessToken = "access_token"
	keyCommonToken = "common_token"
	keyUserID      = "user_id"
)

// Provider is the VK provider. The host fields are empty in production and
// set by tests to point every endpoint at one httptest server.
type Provider struct {
	scheme    string
	apiHost   string
	authHost  string
	oauthHost string
}

// New returns the provider.
func New() *Provider { return &Provider{} }

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "VK",
		Methods: []provider.AuthMethod{
			{ID: methodQR, Name: "QR code", Description: "Scan with VK on your phone: open your profile → settings → log in by QR code"},
		},
		Caps: []provider.Capability{provider.CapMessages},
	}
}

func (p *Provider) Login(_ context.Context, method string, _ *provider.Account, env provider.Env) (provider.Login, error) {
	if method != "" && method != methodQR {
		return nil, errors.New("vk: unknown sign-in method " + method)
	}
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	return &login{srv: p.server(hc), env: env}, nil
}

type login struct {
	srv    *server
	env    provider.Env
	anonym anonymToken
	hash   string
	url    string
	shown  string // the QR last returned to the app
}

func (l *login) step(errMsg string) provider.Step {
	return provider.Step{
		Title: "VK", Error: errMsg, QR: l.url,
		Message: "On your phone open VK, go to your profile → settings → \"Log in by QR code\" and scan this code. " +
			"No code is sent anywhere.",
	}
}

// ensureQR fetches an anonymous token (if needed) and a fresh auth code.
func (l *login) ensureQR(ctx context.Context) error {
	if !l.anonym.valid() {
		t, err := l.srv.getAnonymToken(ctx)
		if err != nil {
			return err
		}
		l.anonym = t
	}
	c, err := l.srv.getAuthCode(ctx, l.anonym.Token, "MegaChat")
	if err != nil {
		return err
	}
	l.hash, l.url = c.AuthHash, c.AuthURL
	l.shown = ""
	return nil
}

func (l *login) Next(ctx context.Context, _ map[string]string) (provider.Step, error) {
	if l.url == "" {
		if err := l.ensureQR(ctx); err != nil {
			return provider.Step{}, err
		}
		l.shown = l.url
		return l.step(""), nil
	}
	// Poll until the user approves, the code dies, or ctx ends (then the
	// hub re-calls us and we poll again).
	for {
		st, err := l.srv.checkAuthCode(ctx, l.anonym.Token, l.hash)
		if err != nil {
			var ae *apiError
			if errors.As(err, &ae) {
				// A dead anonymous token or auth hash: start a new code.
				if err := l.ensureQR(ctx); err != nil {
					return provider.Step{}, err
				}
				l.shown = l.url
				return l.step("The code expired; scan the new one"), nil
			}
			return provider.Step{}, err
		}
		switch st.Status {
		case qrApproved:
			return l.finish(ctx, st.AccessToken)
		case qrDeclined, qrExpired:
			if err := l.ensureQR(ctx); err != nil {
				return provider.Step{}, err
			}
			l.shown = l.url
			msg := "The code expired; scan the new one"
			if st.Status == qrDeclined {
				msg = "Sign-in was declined; scan the new code to try again"
			}
			return l.step(msg), nil
		default: // qrCreated, qrOpened
		}
		select {
		case <-ctx.Done():
			return l.step(""), nil
		case <-time.After(2 * time.Second):
		}
	}
}

func (l *login) finish(ctx context.Context, accessToken string) (provider.Step, error) {
	users, err := l.srv.getExchangeToken(ctx, accessToken)
	if err != nil {
		return provider.Step{}, err
	}
	var ut userToken
	for _, u := range users {
		if u.ProfileType == 0 { // NORMAL (not an EDU/Sferum profile)
			ut = u
			break
		}
	}
	if ut.UserID == 0 && len(users) > 0 {
		ut = users[0]
	}
	if ut.UserID == 0 {
		return provider.Step{}, errors.New("vk: signed in but no user account came back")
	}
	id := strconv.FormatInt(ut.UserID, 10)
	name := "VK " + id
	if self, err := l.srv.usersGet(ctx, accessToken, id); err == nil && len(self) > 0 && self[0].name() != "" {
		name = self[0].name()
	}
	params := map[string]string{
		keyAccessToken: accessToken,
		keyCommonToken: ut.CommonToken,
		keyUserID:      id,
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: id, Name: name, Params: params,
	}}, nil
}

func (l *login) Close() {}
