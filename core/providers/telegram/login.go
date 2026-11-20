package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"megachat/core/provider"
)

type stage int

const (
	stPhone stage = iota
	stCode
	stPassword
	stQR
)

const (
	methodQR    = "qr"
	methodPhone = "phone"
)

// login drives Telegram's sign-in. A client runs in the background for the
// whole dialog: the code is bound to the connection's auth key, which then
// becomes the account's session.
type login struct {
	relogin *provider.Account
	kv      *provider.MemKV

	stage            stage
	custom           bool // the user typed app credentials
	appID            int
	appHash          string
	phone, codeHash  string
	sentVia, pwdHint string
	nextVia          string // how a resend would come ("" : no other way)
	codeLen          int
	inApp            bool // the code went to Telegram on other devices

	// QR login: the flow runs in its own goroutine; qrChanged signals a
	// new code, qrDone its end.
	qrChanged chan struct{}
	qrDone    chan error
	qrShown   string // the code the last step carried

	mu        sync.Mutex
	cl        *telegram.Client
	cancel    context.CancelFunc
	runCtx    context.Context
	loggedIn  qrlogin.LoggedIn
	ready     chan struct{}
	done      chan error
	qrURL     string
	qrRunning bool
}

func (p *Provider) Login(_ context.Context, method string, acct *provider.Account, _ provider.Env) (provider.Login, error) {
	l := &login{relogin: acct, kv: provider.NewMemKV(), qrChanged: make(chan struct{}, 1), qrDone: make(chan error, 1)}
	if method == methodQR {
		l.stage = stQR
	}
	params := map[string]string{}
	if acct != nil {
		params = acct.Params
	}
	l.appID, l.appHash, _ = appCredentials(params)
	l.custom = params["api_id"] != ""
	return l, nil
}

func normPhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return ""
	}
	return "+" + b.String()
}

func (l *login) phoneStep(errMsg string) provider.Step {
	st := provider.Step{
		Title:   "Telegram",
		Message: "Enter the phone number of your Telegram account. Telegram sends a code to your other devices or by SMS.",
		Error:   errMsg,
		Fields:  []provider.Field{{Key: "phone", Label: "Phone number", Help: "International format, e.g. +7 999 123-45-67", Default: l.phone}},
	}
	if l.phone == "" && l.relogin != nil {
		st.Fields[0].Default = l.relogin.Params["phone"]
	}
	st.Fields = append(st.Fields, l.appFields()...)
	return st
}

func (l *login) codeStep(errMsg string) provider.Step {
	msg := fmt.Sprintf("Enter the code sent %s to %s.", l.sentVia, l.phone)
	if l.inApp {
		msg += " It comes from the service chat «Telegram» (blue check mark); it may be muted or archived."
	}
	field := provider.Field{Key: "code", Label: "Code", Help: "You can paste the whole message"}
	if l.codeLen > 0 {
		field.Help = fmt.Sprintf("%d digits; you can paste the whole message", l.codeLen)
	}
	if l.nextVia != "" {
		msg += fmt.Sprintf(" No code? Leave the field empty and press Continue: Telegram sends it again %s.", l.nextVia)
		field.Optional = true
	}
	return provider.Step{Title: "Telegram", Message: msg, Error: errMsg, Fields: []provider.Field{field}}
}

func (l *login) passwordStep(errMsg string) provider.Step {
	help := ""
	if l.pwdHint != "" {
		help = "Hint: " + l.pwdHint
	}
	return provider.Step{
		Title:   "Telegram",
		Message: "This account is protected with a cloud password (two-step verification).",
		Error:   errMsg,
		Fields:  []provider.Field{{Key: "password", Label: "Cloud password", Secret: true, Help: help}},
	}
}

// appFields asks for the app credentials when the build has none.
func (l *login) appFields() []provider.Field {
	if l.appID != 0 && !l.custom {
		return nil
	}
	def := ""
	if l.appID != 0 {
		def = strconv.Itoa(l.appID)
	}
	return []provider.Field{
		{Key: "api_id", Label: "API ID", Help: "From my.telegram.org → API development tools", Default: def},
		{Key: "api_hash", Label: "API hash", Secret: true, Default: l.appHash},
	}
}

// takeApp reads typed app credentials; "" means they are fine.
func (l *login) takeApp(values map[string]string) string {
	id, hash := strings.TrimSpace(values["api_id"]), strings.TrimSpace(values["api_hash"])
	if id == "" && hash == "" {
		if l.appID == 0 {
			return "This build has no Telegram app credentials: enter API ID and hash"
		}
		return ""
	}
	n, err := strconv.Atoi(id)
	if err != nil || n <= 0 || hash == "" {
		return "API ID must be a number and API hash must be set"
	}
	if n != l.appID || hash != l.appHash {
		l.stop()
	}
	l.appID, l.appHash, l.custom = n, hash, true
	return ""
}

func (l *login) qrStep(errMsg string) provider.Step {
	if l.appID == 0 {
		return provider.Step{
			Title: "Telegram", Error: errMsg, Fields: l.appFields(),
			Message: "Enter the app credentials from my.telegram.org; then a QR code to scan appears.",
		}
	}
	l.mu.Lock()
	url := l.qrURL
	l.mu.Unlock()
	return provider.Step{
		Title: "Telegram", Error: errMsg, QR: url,
		Message: "On your phone open Telegram → Settings → Devices → Link Desktop Device and scan this code. " +
			"It renews itself every 30 seconds. No code is sent anywhere.",
	}
}

func (l *login) current(errMsg string) provider.Step {
	switch l.stage {
	case stQR:
		return l.qrStep(errMsg)
	case stCode:
		return l.codeStep(errMsg)
	case stPassword:
		return l.passwordStep(errMsg)
	}
	return l.phoneStep(errMsg)
}

// client starts the background client (again, if the app id changed).
func (l *login) client(ctx context.Context) (*telegram.Client, error) {
	l.mu.Lock()
	if l.cl == nil {
		cctx, cancel := context.WithCancel(context.Background())
		// Before sign-in the server still sends updateLoginToken, which
		// tells that the QR code was scanned.
		d := tg.NewUpdateDispatcher()
		l.loggedIn = qrlogin.OnLoginToken(d)
		cl := newClient(l.appID, l.appHash, kvSession{l.kv}, d)
		ready, done := make(chan struct{}), make(chan error, 1)
		go func() {
			done <- cl.Run(cctx, func(ctx context.Context) error {
				close(ready)
				<-ctx.Done()
				return nil
			})
		}()
		l.cl, l.cancel, l.runCtx, l.ready, l.done = cl, cancel, cctx, ready, done
	}
	cl, ready, done := l.cl, l.ready, l.done
	l.mu.Unlock()
	select {
	case <-ready:
		return cl, nil
	case err := <-done:
		l.stop()
		return nil, fmt.Errorf("telegram: cannot connect: %w", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *login) stop() {
	l.mu.Lock()
	cancel, done := l.cancel, l.done
	l.cl, l.cancel, l.qrURL = nil, nil, ""
	l.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
}

func (l *login) Close() { l.stop() }

func (l *login) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if l.stage == stQR {
		if l.appID == 0 || values["api_id"] != "" {
			if values == nil {
				return l.qrStep(""), nil
			}
			if msg := l.takeApp(values); msg != "" {
				return l.qrStep(msg), nil
			}
		}
		return l.waitQR(ctx)
	}
	if values == nil {
		return l.current(""), nil
	}
	switch l.stage {
	case stPhone:
		return l.sendCode(ctx, values)
	case stCode:
		raw := strings.TrimSpace(values["code"])
		if raw == "" && l.nextVia != "" {
			return l.resend(ctx)
		}
		return l.signIn(ctx, raw)
	default:
		return l.password(ctx, values["password"])
	}
}

func (l *login) sendCode(ctx context.Context, values map[string]string) (provider.Step, error) {
	l.phone = normPhone(values["phone"])
	if l.phone == "" {
		return l.phoneStep("Enter a phone number"), nil
	}
	if msg := l.takeApp(values); msg != "" {
		return l.phoneStep(msg), nil
	}
	cl, err := l.client(ctx)
	if err != nil {
		return provider.Step{}, err
	}
	sent, err := cl.Auth().SendCode(ctx, l.phone, auth.SendCodeOptions{})
	switch {
	case tgerr.Is(err, "PHONE_NUMBER_INVALID", "PHONE_NUMBER_BANNED", "PHONE_NUMBER_FLOOD"):
		return l.phoneStep(tgMessage(err)), nil
	case tgerr.Is(err, "API_ID_INVALID", "API_ID_PUBLISHED_FLOOD"):
		l.stop()
		return l.phoneStep("Telegram rejected the app credentials (API ID / hash)"), nil
	case err != nil:
		if d, ok := tgerr.AsFloodWait(err); ok {
			return l.phoneStep(fmt.Sprintf("Too many attempts: try again in %s", d)), nil
		}
		return provider.Step{}, fmt.Errorf("telegram: %w", err)
	}
	switch s := sent.(type) {
	case *tg.AuthSentCode:
		if _, ok := s.Type.(*tg.AuthSentCodeTypeSetUpEmailRequired); ok {
			l.stop()
			return l.phoneStep("Telegram asks this account to set up a login e-mail first. " +
				"Do it in the official Telegram app (Settings → Privacy and Security → Login e-mail), then try again."), nil
		}
		l.applySent(s)
	case *tg.AuthSentCodeSuccess:
		return l.finish(ctx, cl)
	default:
		return provider.Step{}, fmt.Errorf("telegram: unexpected answer %T", sent)
	}
	l.stage = stCode
	return l.codeStep(""), nil
}

// applySent records how the code was sent and how a resend would come.
func (l *login) applySent(s *tg.AuthSentCode) {
	l.codeHash, l.sentVia, l.codeLen = s.PhoneCodeHash, codeVia(s.Type), codeLength(s.Type)
	_, l.inApp = s.Type.(*tg.AuthSentCodeTypeApp)
	l.nextVia = ""
	if next, ok := s.GetNextType(); ok {
		l.nextVia = nextVia(next)
	}
}

// resend asks Telegram to send the code again, the next way it offers.
func (l *login) resend(ctx context.Context) (provider.Step, error) {
	cl, err := l.client(ctx)
	if err != nil {
		return provider.Step{}, err
	}
	sent, err := cl.Auth().ResendCode(ctx, l.phone, l.codeHash)
	switch {
	case tgerr.Is(err, "PHONE_CODE_EXPIRED"):
		l.stage = stPhone
		return l.phoneStep("The code expired: request a new one"), nil
	case tgerr.Is(err, "SEND_CODE_UNAVAILABLE"):
		l.nextVia = ""
		return l.codeStep("Telegram has no other way to send the code: wait for the message, or start over later"), nil
	case err != nil:
		if d, ok := tgerr.AsFloodWait(err); ok {
			return l.codeStep(fmt.Sprintf("Too many attempts: try again in %s", d)), nil
		}
		return provider.Step{}, fmt.Errorf("telegram: %w", err)
	}
	s, ok := sent.(*tg.AuthSentCode)
	if !ok {
		return provider.Step{}, fmt.Errorf("telegram: unexpected answer %T", sent)
	}
	l.applySent(s)
	return l.codeStep(""), nil
}

// extractCode finds the code in what was typed or pasted: the input
// itself when it is only digits (with spaces or dashes), else the first
// run of 4–8 digits ("Login code: 12345. Do not give this code…").
func extractCode(s string) string {
	plain := strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	if plain != "" && strings.Trim(plain, "0123456789") == "" {
		return plain
	}
	run := ""
	for _, r := range s + " " {
		if r >= '0' && r <= '9' {
			run += string(r)
			continue
		}
		if len(run) >= 4 && len(run) <= 8 {
			return run
		}
		run = ""
	}
	return ""
}

func codeLength(t tg.AuthSentCodeTypeClass) int {
	switch t := t.(type) {
	case *tg.AuthSentCodeTypeApp:
		return t.Length
	case *tg.AuthSentCodeTypeSMS:
		return t.Length
	case *tg.AuthSentCodeTypeCall:
		return t.Length
	case *tg.AuthSentCodeTypeEmailCode:
		return t.Length
	}
	return 0
}

func nextVia(t tg.AuthCodeTypeClass) string {
	switch t.(type) {
	case *tg.AuthCodeTypeSMS:
		return "by SMS"
	case *tg.AuthCodeTypeFragmentSMS:
		return "by SMS (Fragment)"
	case *tg.AuthCodeTypeCall:
		return "by a phone call"
	case *tg.AuthCodeTypeFlashCall, *tg.AuthCodeTypeMissedCall:
		return "by a call (the code is in the caller's number)"
	}
	return ""
}

func codeVia(t tg.AuthSentCodeTypeClass) string {
	switch t.(type) {
	case *tg.AuthSentCodeTypeApp:
		return "in Telegram on your other devices"
	case *tg.AuthSentCodeTypeSMS, *tg.AuthSentCodeTypeSMSWord, *tg.AuthSentCodeTypeSMSPhrase, *tg.AuthSentCodeTypeFragmentSMS, *tg.AuthSentCodeTypeFirebaseSMS:
		return "by SMS"
	case *tg.AuthSentCodeTypeCall:
		return "by a phone call"
	case *tg.AuthSentCodeTypeEmailCode:
		return "by e-mail"
	}
	return ""
}

func tgMessage(err error) string {
	if e, ok := tgerr.As(err); ok {
		return strings.ReplaceAll(strings.ToLower(e.Type), "_", " ")
	}
	return err.Error()
}

func (l *login) signIn(ctx context.Context, raw string) (provider.Step, error) {
	if raw == "" {
		return l.codeStep("Enter the code"), nil
	}
	code := extractCode(raw)
	if code == "" {
		return l.codeStep(fmt.Sprintf("«%s» is not a login code: the code is a few digits. "+
			"Lines like «afYeJbBf-dQ» at the end of an SMS are for Android, not the code.", raw)), nil
	}
	cl, err := l.client(ctx)
	if err != nil {
		return provider.Step{}, err
	}
	_, err = cl.Auth().SignIn(ctx, l.phone, code, l.codeHash)
	var signUp *auth.SignUpRequired
	switch {
	case errors.Is(err, auth.ErrPasswordAuthNeeded):
		if pw, err := cl.API().AccountGetPassword(ctx); err == nil {
			l.pwdHint = pw.Hint
		}
		l.stage = stPassword
		return l.passwordStep(""), nil
	case tgerr.Is(err, "PHONE_CODE_INVALID", "PHONE_CODE_EMPTY"):
		return l.codeStep("Wrong code"), nil
	case tgerr.Is(err, "PHONE_CODE_EXPIRED"):
		l.stage = stPhone
		return l.phoneStep("The code expired: request a new one"), nil
	case errors.As(err, &signUp):
		return provider.Step{}, errors.New("telegram: this phone has no Telegram account yet; sign up in the official app first")
	case err != nil:
		return provider.Step{}, fmt.Errorf("telegram: %w", err)
	}
	return l.finish(ctx, cl)
}

func (l *login) password(ctx context.Context, pw string) (provider.Step, error) {
	if pw == "" {
		return l.passwordStep("Enter the password"), nil
	}
	cl, err := l.client(ctx)
	if err != nil {
		return provider.Step{}, err
	}
	_, err = cl.Auth().Password(ctx, pw)
	switch {
	case errors.Is(err, auth.ErrPasswordInvalid):
		return l.passwordStep("Wrong password"), nil
	case err != nil:
		if d, ok := tgerr.AsFloodWait(err); ok {
			return l.passwordStep(fmt.Sprintf("Too many attempts: try again in %s", d)), nil
		}
		return provider.Step{}, fmt.Errorf("telegram: %w", err)
	}
	return l.finish(ctx, cl)
}

func (l *login) finish(ctx context.Context, cl *telegram.Client) (provider.Step, error) {
	self, err := cl.Self(ctx)
	if err != nil {
		return provider.Step{}, fmt.Errorf("telegram: %w", err)
	}
	sess, err := l.kv.Get(keySession)
	if err != nil || sess == nil {
		return provider.Step{}, errors.New("telegram: no session after sign-in")
	}
	if l.phone == "" && self.Phone != "" {
		l.phone = "+" + self.Phone // QR login: nobody typed it
	}
	params := map[string]string{"phone": l.phone, "user_id": strconv.FormatInt(self.ID, 10)}
	if l.custom {
		params["api_id"], params["api_hash"] = strconv.Itoa(l.appID), l.appHash
	}
	name := userName(self)
	if name == "" {
		name = l.phone
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: strconv.FormatInt(self.ID, 10), Name: name, Params: params,
		State: map[string][]byte{keySession: sess},
	}}, nil
}

// waitQR returns the QR step as soon as there is a new code, the login is
// over, or ctx ends (then the step is unchanged and the app asks again).
func (l *login) waitQR(ctx context.Context) (provider.Step, error) {
	if err := l.startQR(ctx); err != nil {
		return provider.Step{}, err
	}
	for {
		l.mu.Lock()
		url := l.qrURL
		l.mu.Unlock()
		if url != "" && url != l.qrShown {
			l.qrShown = url
			return l.qrStep(""), nil
		}
		select {
		case <-l.qrChanged:
		case err := <-l.qrDone:
			return l.qrFinished(ctx, err)
		case <-ctx.Done():
			return l.qrStep(""), nil
		}
	}
}

// startQR runs gotd's QR flow once: it exports a login token, shows it,
// renews it when it expires and imports the authorization once scanned.
func (l *login) startQR(ctx context.Context) error {
	l.mu.Lock()
	running := l.qrRunning
	l.mu.Unlock()
	if running {
		return nil
	}
	cl, err := l.client(ctx)
	if err != nil {
		return err
	}
	l.mu.Lock()
	l.qrRunning = true
	rctx, loggedIn := l.runCtx, l.loggedIn
	l.mu.Unlock()
	go func() {
		_, err := cl.QR().Auth(rctx, loggedIn, func(_ context.Context, t qrlogin.Token) error {
			l.mu.Lock()
			l.qrURL = t.URL()
			l.mu.Unlock()
			select {
			case l.qrChanged <- struct{}{}:
			default:
			}
			return nil
		})
		l.qrDone <- err
	}()
	return nil
}

func (l *login) qrFinished(ctx context.Context, err error) (provider.Step, error) {
	l.mu.Lock()
	l.qrRunning = false
	l.mu.Unlock()
	cl, cerr := l.client(ctx)
	switch {
	case cerr != nil:
		return provider.Step{}, cerr
	case err == nil:
		return l.finish(ctx, cl)
	case tgerr.Is(err, "SESSION_PASSWORD_NEEDED"):
		// Scanned; the account has a cloud password.
		if pw, err := cl.API().AccountGetPassword(ctx); err == nil {
			l.pwdHint = pw.Hint
		}
		l.stage = stPassword
		return l.passwordStep(""), nil
	}
	return provider.Step{}, fmt.Errorf("telegram: QR login: %w", err)
}
