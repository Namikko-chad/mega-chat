package mailru

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/server"
	"github.com/emersion/go-webdav/caldav"

	"megachat/core/model"
	"megachat/core/provider"
)

// --- local servers -----------------------------------------------------------

func imapServer(t *testing.T) (addr string, be *memory.Backend) {
	be = memory.New()
	u, _ := be.Login(nil, "username", "password")
	for _, name := range []string{"Sent", "Корзина"} {
		if err := u.CreateMailbox(name); err != nil {
			t.Fatal(err)
		}
	}
	inbox, _ := u.GetMailbox("INBOX")
	inbox.(*memory.Mailbox).Messages = nil // drop the backend's sample
	base := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	for i, m := range []string{
		"From: Anna <anna@example.com>\r\nTo: username@mail.ru\r\nSubject: =?utf-8?b?0J/RgNC40LLQtdGC?=\r\nMessage-ID: <one@x>\r\nContent-Type: text/plain; charset=koi8-r\r\n\r\n\xf0\xd2\xc9\xd7\xc5\xd4\r\n",
		"From: Boris <boris@example.com>\r\nSubject: Report\r\nMessage-ID: <two@x>\r\nReferences: <zero@x>\r\nContent-Type: multipart/mixed; boundary=b\r\n\r\n--b\r\nContent-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n<p>See =\r\nthe <b>file</b></p>\r\n--b\r\nContent-Type: application/pdf; name=r.pdf\r\nContent-Disposition: attachment; filename=r.pdf\r\nContent-Transfer-Encoding: base64\r\n\r\naGVs\r\nbG8=\r\n--b--\r\n",
	} {
		if err := inbox.CreateMessage(nil, base.Add(time.Duration(i)*time.Hour), strings.NewReader(m)); err != nil {
			t.Fatal(err)
		}
	}
	s := server.New(be)
	s.AllowInsecureAuth = true
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve(ln)
	t.Cleanup(func() { s.Close() })
	return ln.Addr().String(), be
}

type smtpServer struct {
	mu   sync.Mutex
	from string
	rcpt []string
	data string
}

func (s *smtpServer) serve(t *testing.T) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go s.handle(c)
		}
	}()
	return ln.Addr().String()
}

func (s *smtpServer) handle(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	say := func(l string) { c.Write([]byte(l + "\r\n")) }
	say("220 fake")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			say("250-fake")
			say("250 AUTH PLAIN")
		case strings.HasPrefix(cmd, "AUTH PLAIN"):
			say("235 ok")
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			s.mu.Lock()
			s.from = strings.TrimSpace(line)[10:]
			s.mu.Unlock()
			say("250 ok")
		case strings.HasPrefix(cmd, "RCPT TO:"):
			s.mu.Lock()
			s.rcpt = append(s.rcpt, strings.Trim(strings.TrimSpace(line)[8:], "<>"))
			s.mu.Unlock()
			say("250 ok")
		case cmd == "DATA":
			say("354 go")
			var b strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil || l == ".\r\n" {
					break
				}
				b.WriteString(l)
			}
			s.mu.Lock()
			s.data = b.String()
			s.mu.Unlock()
			say("250 queued")
		case cmd == "QUIT":
			say("221 bye")
			return
		default:
			say("250 ok")
		}
	}
}

// davBackend is a CalDAV server with one calendar in memory.
type davBackend struct {
	mu   sync.Mutex
	objs map[string]*ical.Calendar
}

// go-webdav's server tells resources by depth: principal, home, calendar.
const calPath = "/username/calendars/main/"

func (b *davBackend) CurrentUserPrincipal(context.Context) (string, error) { return "/username/", nil }
func (b *davBackend) CalendarHomeSetPath(context.Context) (string, error) {
	return "/username/calendars/", nil
}
func (b *davBackend) CreateCalendar(context.Context, *caldav.Calendar) error { return nil }
func (b *davBackend) ListCalendars(context.Context) ([]caldav.Calendar, error) {
	return []caldav.Calendar{{Path: calPath, Name: "Main", SupportedComponentSet: []string{"VEVENT"}}}, nil
}
func (b *davBackend) GetCalendar(_ context.Context, p string) (*caldav.Calendar, error) {
	return &caldav.Calendar{Path: calPath, Name: "Main", SupportedComponentSet: []string{"VEVENT"}}, nil
}
func (b *davBackend) list() []caldav.CalendarObject {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []caldav.CalendarObject
	for p, c := range b.objs {
		out = append(out, caldav.CalendarObject{Path: p, ETag: `"` + p + `"`, Data: c})
	}
	return out
}
func (b *davBackend) GetCalendarObject(_ context.Context, p string, _ *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	for _, o := range b.list() {
		if o.Path == p {
			return &o, nil
		}
	}
	return nil, errors.New("not found")
}
func (b *davBackend) ListCalendarObjects(context.Context, string, *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	return b.list(), nil
}

// QueryCalendarObjects returns everything: the client must expand and
// cut by itself, as against a server that does not.
func (b *davBackend) QueryCalendarObjects(context.Context, string, *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	return b.list(), nil
}
func (b *davBackend) PutCalendarObject(_ context.Context, p string, c *ical.Calendar, _ *caldav.PutCalendarObjectOptions) (*caldav.CalendarObject, error) {
	b.mu.Lock()
	b.objs[p] = c
	b.mu.Unlock()
	return &caldav.CalendarObject{Path: p}, nil
}
func (b *davBackend) DeleteCalendarObject(_ context.Context, p string) error {
	b.mu.Lock()
	delete(b.objs, p)
	b.mu.Unlock()
	return nil
}

const weekly = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:test\r\nBEGIN:VEVENT\r\nUID:w\r\nDTSTAMP:20260101T000000Z\r\n" +
	"DTSTART;TZID=Europe/Moscow:20261102T100000\r\nDTEND;TZID=Europe/Moscow:20261102T103000\r\nRRULE:FREQ=WEEKLY;BYDAY=MO,WE\r\n" +
	"EXDATE;TZID=Europe/Moscow:20261111T100000\r\nSUMMARY:Planning\r\nATTENDEE;CN=Me;PARTSTAT=ACCEPTED:mailto:username@mail.ru\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:w\r\nDTSTAMP:20260101T000000Z\r\nRECURRENCE-ID;TZID=Europe/Moscow:20261116T100000\r\n" +
	"DTSTART;TZID=Europe/Moscow:20261116T150000\r\nDTEND;TZID=Europe/Moscow:20261116T153000\r\nSUMMARY:Planning (moved)\r\nEND:VEVENT\r\n" +
	"END:VCALENDAR\r\n"

const allDay = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:test\r\nBEGIN:VEVENT\r\nUID:d\r\nDTSTAMP:20260101T000000Z\r\n" +
	"DTSTART;VALUE=DATE:20261113\r\nSUMMARY:Holiday\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

func davServer(t *testing.T) (string, *davBackend) {
	be := &davBackend{objs: map[string]*ical.Calendar{}}
	for p, s := range map[string]string{calPath + "w.ics": weekly, calPath + "d.ics": allDay} {
		c, err := ical.NewDecoder(strings.NewReader(s)).Decode()
		if err != nil {
			t.Fatal(err)
		}
		be.objs[p] = c
	}
	h := &caldav.Handler{Backend: be}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != "username@mail.ru" || p != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/username/", be
}

// --- tests -------------------------------------------------------------------

type sinkRec struct {
	mu     sync.Mutex
	events []model.Event
	up     chan struct{}
}

func (s *sinkRec) Emit(ev model.Event)                { s.mu.Lock(); s.events = append(s.events, ev); s.mu.Unlock() }
func (s *sinkRec) Connected(model.User)               { close(s.up) }
func (s *sinkRec) SaveParams(map[string]string) error { return nil }

func (s *sinkRec) kinds() []model.EventKind {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []model.EventKind
	for _, e := range s.events {
		out = append(out, e.Kind)
	}
	return out
}

// An account whose calendar discovery failed at sign-in was saved without a
// CalDAV address: a standard one still gets Mail.ru's, own servers do not.
func TestOpenRestoresDefaultCalendar(t *testing.T) {
	cfg := config{Email: "a@mail.ru", Password: "p", IMAP: defaultIMAP, SMTP: defaultSMTP}
	p := New()
	sess, err := p.Open(provider.Account{Params: cfg.params(methodPassword)}, provider.Env{})
	if err != nil {
		t.Fatal(err)
	}
	if cs, ok := sess.(calSession); !ok || cs.cfg.CalDAV != defaultCalDAV {
		t.Fatalf("standard account without CalDAV: %#v", sess)
	}
	sess, err = p.Open(provider.Account{Params: cfg.params(methodServer)}, provider.Env{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sess.(provider.Calendars); ok {
		t.Fatal("own servers without CalDAV must have no calendar")
	}
}

func TestMail(t *testing.T) {
	imapAddr, _ := imapServer(t)
	sm := &smtpServer{}
	smtpAddr := sm.serve(t)
	p := &Provider{Plain: true, Poll: 20 * time.Millisecond}

	l, _ := p.Login(context.Background(), methodServer, nil, provider.Env{})
	st, _ := l.Next(context.Background(), map[string]string{"email": "username", "password": "wrong", "imap": imapAddr, "smtp": smtpAddr})
	if st.Done || st.Error == "" {
		t.Fatalf("a wrong password must be rejected in the step: %+v", st)
	}

	// The memory backend's user has no @: sign in with "username" itself.
	cfg := config{Email: "username", Password: "password", IMAP: imapAddr, SMTP: smtpAddr, SentCopy: true}
	sess, err := p.Open(provider.Account{Login: "username", Params: cfg.params(methodServer)}, provider.Env{})
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	if _, ok := sess.(provider.Calendars); ok {
		t.Fatal("no CalDAV address: no calendar")
	}
	sink := &sinkRec{up: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Run(ctx, sink)
	<-sink.up

	fs, err := s.Folders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	roles := map[model.FolderRole]string{}
	for _, f := range fs {
		roles[f.Role] = f.ID
	}
	if roles[model.FolderInbox] != "INBOX" || roles[model.FolderSent] != "Sent" || roles[model.FolderTrash] != "Корзина" || fs[0].Total != 2 {
		t.Fatalf("folders %+v", fs)
	}

	ms, more, err := s.Mail(ctx, "INBOX", time.Time{}, 10)
	if err != nil || len(ms) != 2 || more {
		t.Fatalf("mail %+v %v %v", ms, more, err)
	}
	if ms[0].Subject != "Report" || !ms[0].HasAttachments || ms[1].Subject != "Привет" || ms[1].From.Name != "Anna" {
		t.Fatalf("summaries %+v", ms)
	}
	older, more, _ := s.Mail(ctx, "INBOX", ms[0].Time.Add(-time.Minute), 10)
	if len(older) != 1 || older[0].ID != ms[1].ID || more {
		t.Fatalf("before: %+v", older)
	}
	page, more, _ := s.Mail(ctx, "INBOX", time.Time{}, 1)
	if len(page) != 1 || !more {
		t.Fatalf("limit: %d %v", len(page), more)
	}

	koi, err := s.MailBody(ctx, ms[1].ID)
	if err != nil || strings.TrimSpace(koi.Text) != "Привет" {
		t.Fatalf("koi8 body %q %v", koi.Text, err)
	}
	rep, err := s.MailBody(ctx, ms[0].ID)
	if err != nil || rep.Text != "See the file" || rep.References != "<zero@x>" || len(rep.Attachments) != 1 {
		t.Fatalf("html body %+v %v", rep, err)
	}
	data, _, err := s.Media(ctx, rep.Attachments[0].Ref, 100)
	if err != nil || string(data) != "hello" {
		t.Fatalf("attachment %q %v", data, err)
	}

	if err := s.UpdateMail(ctx, []string{ms[0].ID}, model.MailRead); err != nil {
		t.Fatal(err)
	}
	// (The memory backend counts no UNSEEN: look at the flags.)
	if ms2, _, _ := s.Mail(ctx, "INBOX", time.Time{}, 10); ms2[0].Unread || !ms2[1].Unread {
		t.Fatalf("flags after read: %+v", ms2)
	}

	err = s.SendMail(ctx, model.OutMail{To: []model.Address{{Email: "anna@example.com"}}, Bcc: []model.Address{{Email: "b@x"}}, Subject: "Re: Report", Text: "ok", ReplyTo: ms[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	sm.mu.Lock()
	if !slices.Equal(sm.rcpt, []string{"anna@example.com", "b@x"}) || !strings.Contains(sm.data, "In-Reply-To: <two@x>") || !strings.Contains(sm.data, "References: <zero@x> <two@x>") || strings.Contains(sm.data, "b@x") {
		t.Fatalf("smtp %v %q", sm.rcpt, sm.data)
	}
	sm.mu.Unlock()
	if sent, _, _ := s.Mail(ctx, "Sent", time.Time{}, 10); len(sent) != 1 {
		t.Fatalf("copy in Sent: %d", len(sent))
	}

	if err := s.UpdateMail(ctx, []string{ms[1].ID}, model.MailTrash); err != nil {
		t.Fatal(err)
	}
	if tr, _, _ := s.Mail(ctx, "Корзина", time.Time{}, 10); len(tr) != 1 {
		t.Fatalf("trash: %d", len(tr))
	}
	// Moving back to the inbox.
	tr, _, _ := s.Mail(ctx, "Корзина", time.Time{}, 10)
	if err := s.MoveMail(ctx, []string{tr[0].ID}, tr[0].FolderID, "INBOX"); err != nil {
		t.Fatal(err)
	}
	if after, _, _ := s.Mail(ctx, "Корзина", time.Time{}, 10); len(after) != 0 {
		t.Fatalf("still in the trash: %d", len(after))
	}
	if in, _, _ := s.Mail(ctx, "INBOX", time.Time{}, 10); len(in) != len(ms) {
		t.Fatalf("inbox: %d, want %d", len(in), len(ms))
	}

	// New mail arrives: the poll reports it.
	// Through the session's own connection: the memory backend is not
	// safe for use from outside the server.
	raw := "From: c@x\r\nSubject: New one\r\n\r\nhi\r\n"
	if err := s.imapDo(func(c *client.Client) error {
		return c.Append("INBOX", nil, time.Now(), literal{strings.NewReader(raw), len(raw)})
	}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for !slices.Contains(sink.kinds(), model.EvMailNew) {
		if time.Now().After(deadline) {
			t.Fatalf("no new-mail event: %v", sink.kinds())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCalDAV(t *testing.T) {
	imapAddr, _ := imapServer(t)
	davURL, be := davServer(t)
	p := &Provider{Plain: true, Poll: time.Hour}

	// The CalDAV server checks the full address; the memory IMAP user is
	// "username", so sign-in is checked with discovery only.
	home, err := discoverCalendars(context.Background(), http.DefaultClient, config{Email: "username@mail.ru", Password: "password", CalDAV: davURL})
	if err != nil || home != "/username/calendars/" {
		t.Fatalf("discovery %q %v", home, err)
	}
	if _, err := discoverCalendars(context.Background(), http.DefaultClient, config{Email: "username@mail.ru", Password: "bad", CalDAV: davURL}); err == nil {
		t.Fatal("a wrong password must fail discovery")
	}
	cfg := config{Email: "username@mail.ru", Password: "password", IMAP: imapAddr, SMTP: "127.0.0.1:1", CalDAV: davURL, CalHome: home}
	sess, err := p.Open(provider.Account{Params: cfg.params(methodServer)}, provider.Env{HTTP: http.DefaultClient})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := sess.(provider.Calendars)
	if !ok {
		t.Fatal("no calendar capability")
	}
	ctx := context.Background()
	cals, err := c.Calendars(ctx)
	if err != nil || len(cals) != 1 || cals[0].ID != calPath || cals[0].Name != "Main" || !cals[0].Primary {
		t.Fatalf("calendars %+v %v", cals, err)
	}
	msk, _ := time.LoadLocation("Europe/Moscow")
	from, to := time.Date(2026, 11, 9, 0, 0, 0, 0, msk), time.Date(2026, 11, 23, 0, 0, 0, 0, msk)
	evs, err := c.Events(ctx, calPath, from, to)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range evs {
		got = append(got, e.Title+"@"+e.Start.In(msk).Format("01-02 15:04"))
	}
	slices.Sort(got)
	// Mondays and Wednesdays at 10, 11 Nov excluded, 16 Nov moved to 15:00.
	want := []string{"Holiday@11-13 03:00", "Planning (moved)@11-16 15:00", "Planning@11-09 10:00", "Planning@11-18 10:00"}
	if !slices.Equal(got, want) {
		t.Fatalf("events\n got %v\nwant %v", got, want)
	}
	for _, e := range evs {
		if e.Title == "Planning" && (!e.Recurring || !strings.Contains(e.ID, "#") || !e.Attendees[0].Self || e.Attendees[0].Status != model.RSVPAccepted) {
			t.Fatalf("occurrence %+v", e)
		}
		if e.Title == "Holiday" && (!e.AllDay || e.End.Sub(e.Start) != 24*time.Hour) {
			t.Fatalf("all-day %+v", e)
		}
		if e.Title == "Planning" && e.Recurrence != "FREQ=WEEKLY;BYDAY=MO,WE" {
			t.Fatalf("rule %q", e.Recurrence)
		}
	}

	start := time.Date(2026, 11, 20, 9, 0, 0, 0, msk)
	saved, err := c.SaveEvent(ctx, model.CalEvent{CalendarID: calPath, Title: "Dentist", Location: "Clinic", Start: start, End: start.Add(time.Hour)}, model.ScopeThis)
	if err != nil || !strings.HasPrefix(saved.ID, calPath) {
		t.Fatalf("save %+v %v", saved, err)
	}
	saved.Title = "Dentist (late)"
	if _, err := c.SaveEvent(ctx, saved, model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	evs, _ = c.Events(ctx, calPath, from, to)
	if !slices.ContainsFunc(evs, func(e model.CalEvent) bool {
		return e.ID == saved.ID && e.Title == "Dentist (late)" && e.Location == "Clinic" && e.Start.Equal(start)
	}) {
		t.Fatalf("edited event not found in %+v", evs)
	}
	before, _ := sess.(calSession).cal.state(ctx)
	if err := c.DeleteEvent(ctx, calPath, saved.ID, model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	if after, _ := sess.(calSession).cal.state(ctx); after == before {
		t.Fatal("state did not change after a delete")
	}
	be.mu.Lock()
	n := len(be.objs)
	be.mu.Unlock()
	if n != 2 {
		t.Fatalf("objects left: %d", n)
	}
}

// Mail.ru has /principals/ but answers /principals with 404.
func TestCalDAVKeepsBaseSlash(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/principals/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusMultiStatus)
		io.WriteString(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:"><d:response><d:href>/principals/</d:href><d:propstat><d:prop><d:current-user-principal><d:href>/principals/example.com/me/</d:href></d:current-user-principal></d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response></d:multistatus>`)
	}))
	defer ts.Close()
	c, err := davClient(ts.Client(), config{Email: "me@example.com", Password: "p"}, ts.URL+"/principals/", nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.FindCurrentUserPrincipal(context.Background())
	if err != nil || got != "/principals/example.com/me/" {
		t.Fatalf("%q %v", got, err)
	}
}

// Mail moved into a folder in bulk arrives in any order: pages go by date,
// not by sequence number.
func TestMailPagesByDate(t *testing.T) {
	imapAddr, _ := imapServer(t)
	p := &Provider{Plain: true, Poll: time.Hour}
	cfg := config{Email: "username", Password: "password", IMAP: imapAddr, SMTP: "127.0.0.1:1"}
	sess, err := p.Open(provider.Account{Login: "username", Params: cfg.params(methodServer)}, provider.Env{})
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	defer s.closeIMAP()
	ctx := context.Background()
	// Newest first, the way a whole inbox lands in the trash.
	base := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for i := 5; i >= 0; i-- {
		raw := fmt.Sprintf("From: a@x\r\nSubject: n%d\r\n\r\nhi\r\n", i)
		if err := s.imapDo(func(c *client.Client) error {
			return c.Append("Корзина", nil, base.AddDate(0, i, 0), literal{strings.NewReader(raw), len(raw)})
		}); err != nil {
			t.Fatal(err)
		}
	}
	subjects := func(ms []model.Mail) (out []string) {
		for _, m := range ms {
			out = append(out, m.Subject)
		}
		return out
	}
	page, more, err := s.Mail(ctx, "Корзина", time.Time{}, 2)
	if err != nil || !more || !slices.Equal(subjects(page), []string{"n5", "n4"}) {
		t.Fatalf("first page %v %v %v", subjects(page), more, err)
	}
	next, more, _ := s.Mail(ctx, "Корзина", page[1].Time.Add(-time.Second), 3)
	if !more || !slices.Equal(subjects(next), []string{"n3", "n2", "n1"}) {
		t.Fatalf("older page %v %v", subjects(next), more)
	}
	last, more, _ := s.Mail(ctx, "Корзина", next[2].Time.Add(-time.Second), 3)
	if more || !slices.Equal(subjects(last), []string{"n0"}) {
		t.Fatalf("last page %v %v", subjects(last), more)
	}
	// A message moved in later shows up: the index follows the folder.
	raw := "From: a@x\r\nSubject: newest\r\n\r\nhi\r\n"
	s.imapDo(func(c *client.Client) error {
		return c.Append("Корзина", nil, base.AddDate(1, 0, 0), literal{strings.NewReader(raw), len(raw)})
	})
	if top, _, _ := s.Mail(ctx, "Корзина", time.Time{}, 1); len(top) != 1 || top[0].Subject != "newest" {
		t.Fatalf("after append %v", subjects(top))
	}
}

// Mail.ru sends bare ETags and calendar-query responses without
// calendar-data: those objects come by multiget or GET, and one that
// cannot be read is left out instead of failing the calendar.
func TestCalDAVLenientObjects(t *testing.T) {
	const ics = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:x\r\nBEGIN:VEVENT\r\nUID:%s\r\nDTSTAMP:20261125T100000Z\r\nDTSTART:20261126T080000Z\r\nDTEND:20261126T090000Z\r\nSUMMARY:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	resp := func(href, etag, data, status string) string {
		var d string
		if data != "" {
			d = "<c:calendar-data>" + data + "</c:calendar-data>"
		}
		return `<d:response><d:href>` + href + `</d:href><d:propstat><d:prop><d:getetag>` + etag + `</d:getetag>` + d + `</d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat>` +
			`<d:propstat><d:prop><c:calendar-data/></d:prop><d:status>HTTP/1.1 ` + status + `</d:status></d:propstat></d:response>`
	}
	var puts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == "REPORT" && strings.Contains(string(body), "calendar-query"):
			w.WriteHeader(http.StatusMultiStatus)
			io.WriteString(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">`+
				`<d:response><d:href>/cal/</d:href><d:propstat><d:prop/><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response>`+
				resp("/cal/a.ics", "abc", fmt.Sprintf(ics, "a", "A"), "200 OK")+
				resp("/cal/b.ics", "bcd", "", "404 Not Found")+
				resp("/cal/c.ics", "cde", "", "404 Not Found")+
				resp("/cal/gone.ics", "x", "", "404 Not Found")+`</d:multistatus>`)
		case r.Method == "REPORT": // multiget: b only
			if !strings.Contains(string(body), "/cal/b.ics") {
				t.Errorf("multiget %s", body)
			}
			w.WriteHeader(http.StatusMultiStatus)
			io.WriteString(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">`+
				resp("/cal/b.ics", "bcd2", fmt.Sprintf(ics, "b", "B"), "200 OK")+`</d:multistatus>`)
		case r.Method == http.MethodGet && r.URL.Path == "/cal/c.ics":
			w.Header().Set("ETag", "cde")
			fmt.Fprintf(w, ics, "c", "C")
		case r.Method == http.MethodPut && r.URL.Path == "/cal/c.ics":
			puts++
			w.Header().Set("ETag", "new")
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	var logged []string
	k, err := newCalClient(ts.Client(), config{Email: "me@example.com", Password: "p", CalDAV: ts.URL + "/principals/"}, func(f string, a ...any) { logged = append(logged, fmt.Sprintf(f, a...)) })
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)
	evs, err := k.events(context.Background(), "/cal/", from, from.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, e := range evs {
		titles = append(titles, e.Title)
	}
	if !slices.Equal(titles, []string{"A", "B", "C"}) || len(logged) != 2 || !strings.Contains(logged[1], "gone.ics: no calendar data") {
		t.Fatalf("events %v, logged %v", titles, logged)
	}
	objs, err := k.raw.query(context.Background(), "/cal/", from, from.AddDate(0, 0, 1), false, k.logf)
	if err != nil || len(objs) != 4 || objs[0].ETag != "abc" {
		t.Fatalf("etags %+v %v", objs, err)
	}
	ev := evs[2]
	ev.Title = "C2"
	if _, err := k.save(context.Background(), ev, "", model.ScopeThis); err != nil || puts != 1 {
		t.Fatalf("save %v, puts %d", err, puts)
	}
}

// Occurrences of a repeating event are changed and deleted one by one, from
// one on, or as a whole series.
func TestCalDAVRecurring(t *testing.T) {
	davURL, be := davServer(t)
	cfg := config{Email: "username@mail.ru", Password: "password", IMAP: "x:1", CalDAV: davURL, CalHome: "/username/calendars/"}
	k, err := newCalClient(http.DefaultClient, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	msk, _ := time.LoadLocation("Europe/Moscow")
	from, to := time.Date(2026, 11, 1, 0, 0, 0, 0, msk), time.Date(2026, 12, 1, 0, 0, 0, 0, msk)
	list := func() (got []string, byDay map[string]model.CalEvent) {
		evs, err := k.events(ctx, calPath, from, to)
		if err != nil {
			t.Fatal(err)
		}
		byDay = map[string]model.CalEvent{}
		for _, e := range evs {
			if e.AllDay {
				continue
			}
			d := e.Start.In(msk).Format("01-02")
			got = append(got, d+" "+e.Start.In(msk).Format("15:04")+" "+e.Title)
			byDay[d] = e
		}
		slices.Sort(got)
		return got, byDay
	}
	check := func(step string, want ...string) map[string]model.CalEvent {
		t.Helper()
		got, byDay := list()
		if !slices.Equal(got, want) {
			t.Fatalf("%s\n got %q\nwant %q", step, got, want)
		}
		return byDay
	}
	days := check("start",
		"11-02 10:00 Planning", "11-04 10:00 Planning", "11-09 10:00 Planning", "11-16 15:00 Planning (moved)",
		"11-18 10:00 Planning", "11-23 10:00 Planning", "11-25 10:00 Planning", "11-30 10:00 Planning")

	// This occurrence: a new override; the 16th's own one is kept.
	e := days["11-18"]
	e.Title = "Planning (18)"
	if _, err := k.save(ctx, e, "", model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	if err := k.remove(ctx, days["11-09"].ID, model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	days = check("this",
		"11-02 10:00 Planning", "11-04 10:00 Planning", "11-16 15:00 Planning (moved)",
		"11-18 10:00 Planning (18)", "11-23 10:00 Planning", "11-25 10:00 Planning", "11-30 10:00 Planning")

	// This and the following: a new series from the 23rd at 11.
	e = days["11-23"]
	e.Title, e.Start, e.End = "Planning v2", e.Start.Add(time.Hour), e.End.Add(time.Hour)
	if _, err := k.save(ctx, e, "", model.ScopeFollowing); err != nil {
		t.Fatal(err)
	}
	days = check("following",
		"11-02 10:00 Planning", "11-04 10:00 Planning", "11-16 15:00 Planning (moved)",
		"11-18 10:00 Planning (18)", "11-23 11:00 Planning v2", "11-25 11:00 Planning v2", "11-30 11:00 Planning v2")
	if r := days["11-02"].Recurrence; !strings.Contains(r, "UNTIL=20261123T065959Z") {
		t.Fatalf("old series' rule %q", r)
	}
	be.mu.Lock()
	n := len(be.objs)
	be.mu.Unlock()
	if n != 3 {
		t.Fatalf("objects: %d", n)
	}

	// The whole (old) series half an hour later: overrides go, the cut stays.
	e = days["11-04"]
	e.Start, e.End = e.Start.Add(30*time.Minute), e.End.Add(30*time.Minute)
	if _, err := k.save(ctx, e, "", model.ScopeAll); err != nil {
		t.Fatal(err)
	}
	days = check("all",
		"11-02 10:30 Planning", "11-04 10:30 Planning", "11-16 10:30 Planning", "11-18 10:30 Planning",
		"11-23 11:00 Planning v2", "11-25 11:00 Planning v2", "11-30 11:00 Planning v2")

	// Deleting from the 25th on, then the old series whole.
	if err := k.remove(ctx, days["11-25"].ID, model.ScopeFollowing); err != nil {
		t.Fatal(err)
	}
	if err := k.remove(ctx, days["11-02"].ID, model.ScopeAll); err != nil {
		t.Fatal(err)
	}
	check("deleted", "11-23 11:00 Planning v2")

	// A new repeating event; then a single one made repeating.
	start := time.Date(2026, 11, 10, 9, 0, 0, 0, msk)
	if _, err := k.save(ctx, model.CalEvent{CalendarID: calPath, Title: "Daily", Start: start, End: start.Add(time.Hour), Recurrence: "FREQ=DAILY;COUNT=2"}, "", model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	check("new", "11-10 09:00 Daily", "11-11 09:00 Daily", "11-23 11:00 Planning v2")
}
