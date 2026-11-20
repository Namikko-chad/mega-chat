// Package mailru is the Mail.ru provider: the mailbox over IMAP and SMTP
// and the calendar over CalDAV. It serves Mail.ru mailboxes and VK
// WorkSpace (VK Teams' suite: mail at biz.mail.ru uses the same servers),
// and, with the servers entered by hand, an on-premises VK WorkSpace or any
// IMAP/SMTP/CalDAV server.
//
// Sign-in takes the address and an app password ("password for external
// apps" in the mailbox's security settings): the password is what IMAP,
// SMTP and CalDAV take on every connection, so it is kept in the vault.
//
// Mail ids are "<uidvalidity>:<uid>:<folder>"; calendar ids are CalDAV
// collection paths; event ids are object paths, with "#<start>" for an
// occurrence of a repeating event.
package mailru

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // TZIDs of events on devices without a zone database

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"

	"megachat/core/model"
	"megachat/core/provider"
)

// Type is the provider type.
const Type = "mailru"

const (
	methodPassword = "password"
	methodServer   = "server"
)

// Mail.ru's servers (also VK WorkSpace in the cloud). The CalDAV address
// and the absence of a copy in Sent after SMTP follow Mail.ru's help pages
// and are not checked against the live service by the tests.
const (
	defaultIMAP   = "imap.mail.ru:993"
	defaultSMTP   = "smtp.mail.ru:465"
	defaultCalDAV = "https://calendar.mail.ru/"
)

func init() {
	imap.CharsetReader = charset.Reader
}

// Provider is the Mail.ru provider.
type Provider struct {
	// Poll is how often the inbox is checked; calendars every fifth time.
	Poll time.Duration
	// TLS is used for IMAP and SMTP (tests trust their own servers).
	TLS *tls.Config
	// Plain connects without TLS (tests only).
	Plain bool
}

// New returns the provider.
func New() *Provider { return &Provider{Poll: time.Minute} }

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "Mail.ru / VK WorkSpace",
		Methods: []provider.AuthMethod{
			{ID: methodPassword, Name: "Mail.ru or VK WorkSpace", Description: "Address and an app password; mail and calendar"},
			{ID: methodServer, Name: "Own servers", Description: "On-premises VK WorkSpace or any IMAP/SMTP/CalDAV server"},
		},
		Caps: []provider.Capability{provider.CapMail, provider.CapCalendar},
	}
}

// config is an account's connection settings (its vault params).
type config struct {
	Email, Password string
	IMAP, SMTP      string // host:port
	CalDAV          string // base URL; empty: no calendar
	CalHome         string // calendar home set, found at sign-in
	SentCopy        bool   // append sent mail to Sent (the server does not)
}

func configOf(p map[string]string) config {
	return config{
		Email: p["email"], Password: p["password"], IMAP: p["imap"], SMTP: p["smtp"],
		CalDAV: p["caldav"], CalHome: p["caldav_home"], SentCopy: p["sent_copy"] == "1",
	}
}

func (c config) params(method string) map[string]string {
	p := map[string]string{
		"method": method, "email": c.Email, "password": c.Password, "imap": c.IMAP, "smtp": c.SMTP,
		"caldav": c.CalDAV, "caldav_home": c.CalHome,
	}
	if c.SentCopy {
		p["sent_copy"] = "1"
	}
	return p
}

// withPort adds the default port to a host given without one.
func withPort(host, port string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}
	return net.JoinHostPort(host, port)
}

// --- sign-in -----------------------------------------------------------------

type login struct {
	p      *Provider
	http   *http.Client
	method string
	prev   map[string]string
}

func (p *Provider) Login(_ context.Context, method string, acct *provider.Account, env provider.Env) (provider.Login, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	if method == "" {
		method = methodPassword
	}
	if method != methodPassword && method != methodServer {
		return nil, fmt.Errorf("mailru: unknown sign-in method %q", method)
	}
	prev := map[string]string{}
	if acct != nil {
		prev = acct.Params
	}
	return &login{p: p, http: hc, method: method, prev: prev}, nil
}

func (l *login) step(errMsg string) provider.Step {
	st := provider.Step{
		Title: "Mail.ru / VK WorkSpace",
		Message: "Use an app password, not the account's own: in the mailbox open Settings → Security → " +
			"Passwords for external apps, create one with access to mail and calendar and paste it here.",
		Error: errMsg,
		Fields: []provider.Field{
			{Key: "email", Label: "E-mail", Default: l.prev["email"]},
			{Key: "password", Label: "App password", Secret: true},
		},
	}
	if l.method == methodServer {
		st.Title = "Own mail servers"
		st.Message = "The servers of an on-premises VK WorkSpace or any other mail service. Ports default to 993 (IMAP over TLS) and 465 (SMTP over TLS); 587 uses STARTTLS."
		st.Fields = append(st.Fields,
			provider.Field{Key: "imap", Label: "IMAP server", Help: "host or host:port", Default: l.prev["imap"]},
			provider.Field{Key: "smtp", Label: "SMTP server", Help: "host or host:port", Default: l.prev["smtp"]},
			provider.Field{Key: "caldav", Label: "CalDAV address", Optional: true, Help: "https://…; empty: no calendar", Default: l.prev["caldav"]},
			provider.Field{Key: "sent_copy", Label: "Save sent mail to Sent (1 or empty)", Optional: true, Help: "When the server does not do it itself", Default: l.prev["sent_copy"]},
		)
	}
	return st
}

func (l *login) Next(ctx context.Context, values map[string]string) (provider.Step, error) {
	if values == nil {
		return l.step(""), nil
	}
	c := config{
		Email:    strings.ToLower(strings.TrimSpace(values["email"])),
		Password: strings.TrimSpace(values["password"]),
		IMAP:     defaultIMAP, SMTP: defaultSMTP, CalDAV: defaultCalDAV,
	}
	if l.method == methodServer {
		c.IMAP, c.SMTP = withPort(values["imap"], "993"), withPort(values["smtp"], "465")
		c.CalDAV = strings.TrimSpace(values["caldav"])
		c.SentCopy = strings.TrimSpace(values["sent_copy"]) == "1"
		if c.IMAP == "" || c.SMTP == "" {
			return l.step("Enter the IMAP and SMTP servers"), nil
		}
	}
	if !strings.Contains(c.Email, "@") || c.Password == "" {
		return l.step("Enter the address and the app password"), nil
	}
	cl, err := l.p.dialIMAP(c)
	if err != nil {
		if errors.Is(err, provider.ErrAuth) {
			return l.step("The server did not accept the address and password. Is it an app password?"), nil
		}
		return l.step(err.Error()), nil
	}
	cl.Logout()
	if c.CalDAV != "" {
		home, err := discoverCalendars(ctx, l.http, c)
		if err != nil {
			if l.method == methodServer {
				return l.step("CalDAV: " + err.Error()), nil
			}
			c.CalDAV = "" // mail works; the calendar is not offered
		}
		c.CalHome = home
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: c.Email, Name: c.Email, Params: c.params(l.method),
	}}, nil
}

func (l *login) Close() {}

// --- session -----------------------------------------------------------------

type session struct {
	p   *Provider
	cfg config
	env provider.Env
	cal *calClient // nil: no calendar

	imu sync.Mutex // one IMAP command at a time
	c   *client.Client

	mu       sync.Mutex
	inbox    *imap.MailboxStatus // last seen
	calState string
}

// calSession adds the calendar to a session whose account has one.
type calSession struct{ *session }

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	cfg := configOf(acct.Params)
	if cfg.Email == "" || cfg.IMAP == "" {
		return nil, fmt.Errorf("mailru: incomplete account settings")
	}
	s := &session{p: p, cfg: cfg, env: env}
	if cfg.CalDAV == "" {
		return s, nil
	}
	cal, err := newCalClient(hc, cfg)
	if err != nil {
		return nil, err
	}
	s.cal = cal
	return calSession{s}, nil
}

// dialIMAP connects and signs in. A rejected login wraps ErrAuth.
func (p *Provider) dialIMAP(c config) (*client.Client, error) {
	d := &net.Dialer{Timeout: 30 * time.Second}
	var cl *client.Client
	var err error
	if p.Plain {
		cl, err = client.DialWithDialer(d, c.IMAP)
	} else {
		host, _, _ := net.SplitHostPort(c.IMAP)
		cfg := &tls.Config{ServerName: host}
		if p.TLS != nil {
			cfg = p.TLS.Clone()
		}
		cl, err = client.DialWithDialerTLS(d, c.IMAP, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("IMAP %s: %v", c.IMAP, err)
	}
	cl.Timeout = 2 * time.Minute
	if err := cl.Login(c.Email, c.Password); err != nil {
		cl.Logout()
		return nil, fmt.Errorf("IMAP login: %w: %v", provider.ErrAuth, err)
	}
	return cl, nil
}

// imapDo runs fn on the session's connection, connecting when needed and
// once more when the connection turned out to be dead.
func (s *session) imapDo(fn func(c *client.Client) error) error {
	s.imu.Lock()
	defer s.imu.Unlock()
	for attempt := 0; ; attempt++ {
		if s.c == nil {
			c, err := s.p.dialIMAP(s.cfg)
			if err != nil {
				return err
			}
			s.c = c
		}
		err := fn(s.c)
		if err == nil || s.c.State() != imap.LogoutState {
			return err
		}
		s.c = nil
		if attempt > 0 {
			return fmt.Errorf("%w: %v", provider.ErrOffline, err)
		}
	}
}

func (s *session) closeIMAP() {
	s.imu.Lock()
	c := s.c
	s.c = nil
	s.imu.Unlock()
	if c != nil {
		c.Logout()
	}
}

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	defer s.closeIMAP()
	err := s.imapDo(func(c *client.Client) error {
		st, err := c.Status(imap.InboxName, []imap.StatusItem{imap.StatusUidNext, imap.StatusUnseen, imap.StatusMessages})
		if err == nil {
			s.mu.Lock()
			s.inbox = st
			s.mu.Unlock()
		}
		return err
	})
	if err != nil {
		return err
	}
	if s.cal != nil {
		if st, err := s.cal.state(ctx); err == nil {
			s.calState = st
		} else {
			s.env.Logf("warn", "calendar: %v", err)
		}
	}
	sink.Connected(model.User{ID: s.cfg.Email, Name: s.cfg.Email, Username: s.cfg.Email})
	tick := time.NewTicker(s.p.Poll)
	defer tick.Stop()
	for n := 1; ; n++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
		if err := s.pollInbox(sink); err != nil {
			return err
		}
		if s.cal != nil && n%5 == 0 {
			st, err := s.cal.state(ctx)
			if err != nil {
				s.env.Logf("warn", "calendar: %v", err)
				continue
			}
			if st != s.calState {
				s.calState = st
				sink.Emit(model.Event{Kind: model.EvCalendarChanged})
			}
		}
	}
}

// Calendar methods exist only on sessions of accounts with a calendar.

func (s calSession) Calendars(ctx context.Context) ([]model.Calendar, error) {
	return s.cal.calendars(ctx)
}

func (s calSession) Events(ctx context.Context, cal string, from, to time.Time) ([]model.CalEvent, error) {
	return s.cal.eventsAs(ctx, cal, from, to, s.cfg.Email)
}

func (s calSession) SaveEvent(ctx context.Context, ev model.CalEvent) (model.CalEvent, error) {
	return s.cal.save(ctx, ev, s.cfg.Email)
}

func (s calSession) DeleteEvent(ctx context.Context, cal, id string) error {
	return s.cal.remove(ctx, id)
}
