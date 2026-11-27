package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// fakeGoogle serves the parts of the OAuth, Gmail and Calendar APIs the
// provider uses.
type fakeGoogle struct {
	t  *testing.T
	mu sync.Mutex

	refresh  string // the valid refresh token
	modified []map[string]any
	sent     []string
	created  []gEvent
	history  string
	patched  []patch  // calendar events
	deleted  []string // calendar event ids
}

type patch struct {
	id   string
	body map[string]any
}

// The calendar's series "s" (Fridays at 10 in Moscow) and its occurrence
// "e1" on 13 November.
var gSeries = map[string]gEvent{
	"s": {ID: "s", Summary: "Standup", Start: gTime{DateTime: "2026-11-06T10:00:00+03:00", TimeZone: "Europe/Moscow"},
		End: gTime{DateTime: "2026-11-06T10:15:00+03:00", TimeZone: "Europe/Moscow"}, Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=FR", "EXDATE;TZID=Europe/Moscow:20261120T100000"}},
	"e1": {ID: "e1", Summary: "Standup", Start: gTime{DateTime: "2026-11-13T10:00:00+03:00", TimeZone: "Europe/Moscow"},
		End: gTime{DateTime: "2026-11-13T10:15:00+03:00", TimeZone: "Europe/Moscow"}, RecurringEventID: "s",
		OriginalStartTime: &gTime{DateTime: "2026-11-13T10:00:00+03:00", TimeZone: "Europe/Moscow"}},
	"single": {ID: "single", Summary: "Once", Start: gTime{DateTime: "2026-11-13T12:00:00+03:00"}, End: gTime{DateTime: "2026-11-13T13:00:00+03:00"}},
}

func b64(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

func (f *fakeGoogle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	js := func(v any) { w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(v) }
	if r.URL.Path == "/token" {
		r.ParseForm()
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			if r.Form.Get("code") != "the-code" || r.Form.Get("code_verifier") == "" {
				js(map[string]string{"error": "invalid_grant"})
				return
			}
			js(map[string]any{"access_token": "at1", "refresh_token": "rt1", "expires_in": 3600})
		case "refresh_token":
			if r.Form.Get("refresh_token") != f.refresh {
				js(map[string]string{"error": "invalid_grant", "error_description": "revoked"})
				return
			}
			js(map[string]any{"access_token": "at2", "expires_in": 3600})
		}
		return
	}
	if a := r.Header.Get("Authorization"); a != "Bearer at1" && a != "Bearer at2" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	p := r.URL.Path
	switch {
	case p == "/oauth2/v3/userinfo":
		js(userInfo{Sub: "42", Email: "Me@Gmail.com", Name: "Me", Picture: "https://lh3.googleusercontent.com/p"})
	case p == "/gmail/v1/users/me/profile":
		js(map[string]string{"historyId": f.history})
	case p == "/gmail/v1/users/me/labels":
		js(map[string]any{"labels": []label{{ID: "INBOX", Type: "system"}, {ID: "Label_1", Name: "Work", Type: "user"}}})
	case strings.HasPrefix(p, "/gmail/v1/users/me/labels/"):
		js(label{MessagesTotal: 10, MessagesUnread: 3})
	case p == "/gmail/v1/users/me/messages":
		if r.URL.Query().Get("labelIds") != "INBOX" {
			f.t.Errorf("labelIds %q", r.URL.Query().Get("labelIds"))
		}
		ids := []map[string]string{{"id": "a"}, {"id": "b"}, {"id": "c"}}
		next := "more"
		if q := r.URL.Query().Get("q"); q != "" {
			ids, next = []map[string]string{{"id": "c"}}, ""
		}
		js(map[string]any{"messages": ids, "nextPageToken": next})
	case strings.HasPrefix(p, "/gmail/v1/users/me/messages/") && strings.HasSuffix(p, "/attachments/ATT"):
		js(map[string]any{"size": 5, "data": b64("hello")})
	case p == "/gmail/v1/users/me/messages/batchModify":
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		f.modified = append(f.modified, b)
	case p == "/gmail/v1/users/me/messages/send":
		var b map[string]string
		json.NewDecoder(r.Body).Decode(&b)
		raw, _ := base64.RawURLEncoding.DecodeString(b["raw"])
		f.sent = append(f.sent, string(raw)+"\nthread="+b["threadId"])
	case strings.HasPrefix(p, "/gmail/v1/users/me/messages/"):
		id := strings.TrimPrefix(p, "/gmail/v1/users/me/messages/")
		at := map[string]string{"a": "1700000300000", "b": "1700000200000", "c": "1700000100000"}[id]
		m := message{
			ID: id, ThreadID: "t" + id, LabelIDs: []string{"INBOX", "UNREAD"}, Snippet: "Hi &amp; bye", InternalDate: at,
			Payload: part{MimeType: "multipart/mixed", Headers: []header{
				{"From", `"Anna" <anna@example.com>`}, {"To", "me@gmail.com"}, {"Subject", "=?utf-8?q?Hello_=E2=9C=93?="},
				{"Message-ID", "<" + id + "@x>"},
			}},
		}
		if r.URL.Query().Get("format") == "full" {
			alt := part{MimeType: "multipart/alternative", Parts: []part{{MimeType: "text/html"}}}
			alt.Parts[0].Body.Data = b64("<p>Hello <b>there</b></p>")
			file := part{PartID: "2", MimeType: "application/pdf", Filename: "a.pdf"}
			file.Body.AttachmentID, file.Body.Size = "ATT", 5
			m.Payload.Parts = []part{alt, file}
		}
		js(m)
	case p == "/gmail/v1/users/me/history":
		if r.URL.Query().Get("startHistoryId") != f.history {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.history = "101"
		js(map[string]any{"historyId": "101", "history": []any{map[string]any{
			"messagesAdded": []any{map[string]any{"message": map[string]any{"id": "a", "labelIds": []string{"INBOX", "UNREAD"}}}},
		}}})
	case p == "/calendar/v3/users/me/calendarList":
		js(map[string]any{"items": []map[string]any{
			{"id": "me@gmail.com", "summary": "Me", "backgroundColor": "#9fe1e7", "accessRole": "owner", "primary": true},
			{"id": "holidays", "summary": "Holidays", "accessRole": "reader"},
		}})
	case strings.HasSuffix(p, "/events") && r.Method == http.MethodGet:
		js(map[string]any{"items": []gEvent{
			{ID: "e1", Summary: "Standup", Start: gTime{DateTime: "2026-11-13T10:00:00+03:00"}, End: gTime{DateTime: "2026-11-13T10:15:00+03:00"}, RecurringEventID: "s", HangoutLink: "https://meet.google.com/x"},
			{ID: "e2", Summary: "Trip", Start: gTime{Date: "2026-11-14"}, End: gTime{Date: "2026-11-16"},
				Attendees: []gPerson{{Email: "me@gmail.com", ResponseStatus: "accepted", Self: true}}},
		}})
	case strings.Contains(p, "/events/"):
		id := p[strings.LastIndex(p, "/")+1:]
		switch r.Method {
		case http.MethodGet:
			js(gSeries[id])
		case http.MethodPatch:
			var b map[string]any
			json.NewDecoder(r.Body).Decode(&b)
			f.patched = append(f.patched, patch{id, b})
			js(gSeries[id])
		case http.MethodDelete:
			f.deleted = append(f.deleted, id)
			w.WriteHeader(http.StatusNoContent)
		}
	case strings.HasPrefix(p, "/calendar/v3/calendars/") && !strings.Contains(p, "/events"):
		js(map[string]string{"timeZone": "Europe/Moscow"})
	case strings.HasSuffix(p, "/events") && r.Method == http.MethodPost:
		var g gEvent
		json.NewDecoder(r.Body).Decode(&g)
		f.created = append(f.created, g)
		g.ID = "new"
		js(g)
	default:
		f.t.Errorf("unexpected %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}
}

type sinkRec struct {
	mu     sync.Mutex
	events []model.Event
	self   model.User
	up     chan struct{}
}

func (s *sinkRec) Emit(ev model.Event) { s.mu.Lock(); s.events = append(s.events, ev); s.mu.Unlock() }
func (s *sinkRec) Connected(u model.User) {
	s.mu.Lock()
	s.self = u
	s.mu.Unlock()
	close(s.up)
}
func (s *sinkRec) SaveParams(map[string]string) error { return nil }

func setup(t *testing.T) (*Provider, *fakeGoogle) {
	f := &fakeGoogle{t: t, refresh: "rt1", history: "100"}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	p := New()
	p.AuthURL, p.TokenURL, p.API = srv.URL+"/auth", srv.URL+"/token", srv.URL+"/"
	return p, f
}

func TestLogin(t *testing.T) {
	p, _ := setup(t)
	DefaultClientID = ""
	ctx := context.Background()
	l, err := p.Login(ctx, "", nil, provider.Env{HTTP: http.DefaultClient})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	st, _ := l.Next(ctx, nil)
	if len(st.Fields) != 2 || st.OpenURL != "" {
		t.Fatalf("client step %+v", st)
	}
	st, _ = l.Next(ctx, map[string]string{"client_id": "cid", "client_secret": "sec"})
	if st.OpenURL == "" {
		t.Fatalf("no consent URL: %+v", st)
	}
	u, _ := url.Parse(st.OpenURL)
	q := u.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("access_type") != "offline" || !strings.Contains(q.Get("scope"), "gmail.modify") {
		t.Fatalf("auth url %s", st.OpenURL)
	}
	// The browser comes back to the loopback address.
	go func() {
		resp, err := http.Get(q.Get("redirect_uri") + "?state=" + q.Get("state") + "&code=the-code")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	st, err = l.Next(wctx, nil)
	if err != nil || !st.Done {
		t.Fatalf("not done: %+v %v", st, err)
	}
	cr := st.Result
	if cr.Login != "me@gmail.com" || cr.Params["refresh_token"] != "rt1" || cr.Params["client_id"] != "cid" {
		t.Fatalf("credentials %+v", cr)
	}
}

func open(t *testing.T, p *Provider) (*session, *sinkRec, context.CancelFunc) {
	sess, _ := p.Open(provider.Account{ID: "a", Login: "me@gmail.com", Params: map[string]string{"refresh_token": "rt1", "client_id": "cid"}}, provider.Env{HTTP: http.DefaultClient})
	sink := &sinkRec{up: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- sess.Run(ctx, sink) }()
	select {
	case <-sink.up:
	case err := <-done:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("not connected")
	}
	return sess.(*session), sink, cancel
}

func TestGmail(t *testing.T) {
	p, f := setup(t)
	p.Poll = 20 * time.Millisecond
	s, sink, cancel := open(t, p)
	defer cancel()
	ctx := context.Background()
	if sink.self.Avatar != "photo:https://lh3.googleusercontent.com/p" || sink.self.Username != "Me@Gmail.com" {
		t.Fatalf("self %+v", sink.self)
	}

	fs, err := s.Folders(ctx)
	if err != nil || fs[0].Role != model.FolderInbox || fs[0].Unread != 3 || fs[len(fs)-1].Name != "Work" {
		t.Fatalf("folders %+v %v", fs, err)
	}
	ms, more, err := s.Mail(ctx, "INBOX", time.Time{}, 3)
	if err != nil || len(ms) != 3 || !more {
		t.Fatalf("mail %d %v %v", len(ms), more, err)
	}
	m := ms[0]
	if m.ID != "a" || m.Subject != "Hello ✓" || m.From.Name != "Anna" || !m.Unread || m.Snippet != "Hi & bye" || !m.HasAttachments {
		t.Fatalf("summary %+v", m)
	}
	// Older than b: only c (the server filters by seconds, the provider exactly).
	ms, more, _ = s.Mail(ctx, "INBOX", time.UnixMilli(1700000150000), 3)
	if len(ms) != 1 || ms[0].ID != "c" || more {
		t.Fatalf("before: %+v %v", ms, more)
	}

	full, err := s.MailBody(ctx, "a")
	if err != nil || full.Text != "Hello there" || len(full.Attachments) != 1 || full.Attachments[0].Ref != "att:a:ATT" {
		t.Fatalf("body %+v %v", full, err)
	}
	data, _, err := s.Media(ctx, full.Attachments[0].Ref, 100)
	if err != nil || string(data) != "hello" {
		t.Fatalf("attachment %q %v", data, err)
	}
	if _, _, err := s.Media(ctx, "photo:https://evil.example.com/x", 100); err == nil {
		t.Fatal("photo from another host must be refused")
	}

	if err := s.UpdateMail(ctx, []string{"a", "b"}, model.MailRead); err != nil {
		t.Fatal(err)
	}
	if err := s.SendMail(ctx, model.OutMail{To: []model.Address{{Email: "anna@example.com"}}, Bcc: []model.Address{{Email: "bcc@x"}}, Subject: "Re", Text: "ok", ReplyTo: "a"}); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	if len(f.modified) != 1 || f.modified[0]["removeLabelIds"].([]any)[0] != "UNREAD" {
		t.Fatalf("modify %+v", f.modified)
	}
	f.modified = nil
	f.mu.Unlock()
	// Moving: the target label on, the listed folder's off; to All mail
	// only off; out of Starred only on.
	for _, mv := range [][2]string{{"INBOX", "Label_1"}, {"INBOX", archiveID}, {"STARRED", "SPAM"}} {
		if err := s.MoveMail(ctx, []string{"a"}, mv[0], mv[1]); err != nil {
			t.Fatal(err)
		}
	}
	f.mu.Lock()
	if len(f.modified) != 3 || fmt.Sprint(f.modified[0]["addLabelIds"], f.modified[0]["removeLabelIds"]) != "[Label_1] [INBOX]" ||
		f.modified[1]["addLabelIds"] != nil || fmt.Sprint(f.modified[1]["removeLabelIds"]) != "[INBOX]" ||
		fmt.Sprint(f.modified[2]["addLabelIds"]) != "[SPAM]" || f.modified[2]["removeLabelIds"] != nil {
		t.Fatalf("move %+v", f.modified)
	}
	if len(f.sent) != 1 || !strings.Contains(f.sent[0], "In-Reply-To: <a@x>") || !strings.Contains(f.sent[0], "thread=ta") || !strings.Contains(f.sent[0], "Bcc: bcc@x") {
		t.Fatalf("sent %q", f.sent)
	}
	f.mu.Unlock()

	// The history poll reports the new inbox message.
	deadline := time.Now().Add(5 * time.Second)
	for {
		sink.mu.Lock()
		n := len(sink.events)
		var ev model.Event
		if n > 0 {
			ev = sink.events[0]
		}
		sink.mu.Unlock()
		if n > 0 {
			if ev.Kind != model.EvMailNew || ev.Mail.ID != "a" || ev.Mail.FolderID != "INBOX" {
				t.Fatalf("event %+v", ev)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no new-mail event")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCalendar(t *testing.T) {
	p, f := setup(t)
	p.Poll = time.Hour
	s, _, cancel := open(t, p)
	defer cancel()
	ctx := context.Background()
	cs, err := s.Calendars(ctx)
	if err != nil || len(cs) != 2 || !cs[0].Primary || cs[0].Color != 0x9fe1e7 || cs[0].ReadOnly || !cs[1].ReadOnly {
		t.Fatalf("calendars %+v %v", cs, err)
	}
	evs, err := s.Events(ctx, "me@gmail.com", time.Now(), time.Now().Add(time.Hour))
	if err != nil || len(evs) != 2 {
		t.Fatal(evs, err)
	}
	if e := evs[0]; !e.Recurring || e.Link != "https://meet.google.com/x" || e.End.Sub(e.Start) != 15*time.Minute {
		t.Fatalf("event %+v", e)
	}
	if e := evs[1]; !e.AllDay || e.Start != time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC) || e.Attendees[0].Status != model.RSVPAccepted {
		t.Fatalf("all-day %+v", e)
	}
	day := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	out, err := s.SaveEvent(ctx, model.CalEvent{CalendarID: "me@gmail.com", Title: "Off", Start: day, End: day.AddDate(0, 0, 1), AllDay: true}, model.ScopeThis)
	if err != nil || out.ID != "new" || !out.AllDay {
		t.Fatalf("save %+v %v", out, err)
	}
	f.mu.Lock()
	if f.created[0].Start.Date != "2026-12-01" || f.created[0].End.Date != "2026-12-02" {
		t.Fatalf("created %+v", f.created[0])
	}
	f.mu.Unlock()
}

func TestRevoked(t *testing.T) {
	p, f := setup(t)
	f.refresh = "other"
	sess, _ := p.Open(provider.Account{Login: "me@gmail.com", Params: map[string]string{"refresh_token": "rt1"}}, provider.Env{HTTP: http.DefaultClient})
	err := sess.Run(context.Background(), &sinkRec{up: make(chan struct{})})
	if !errors.Is(err, provider.ErrAuth) {
		t.Fatalf("want ErrAuth, got %v", err)
	}
}

func TestRecurringCalendar(t *testing.T) {
	p, f := setup(t)
	p.Poll = time.Hour
	s, _, cancel := open(t, p)
	defer cancel()
	ctx := context.Background()
	evs, err := s.Events(ctx, "me@gmail.com", time.Now(), time.Now().Add(time.Hour))
	if err != nil || evs[0].Recurrence != "FREQ=WEEKLY;BYDAY=FR" || evs[1].Recurrence != "" {
		t.Fatalf("rules %+v %v", evs, err)
	}
	msk, _ := time.LoadLocation("Europe/Moscow")
	occ := model.CalEvent{CalendarID: "me@gmail.com", ID: "e1", Title: "Standup",
		Start: time.Date(2026, 11, 13, 11, 0, 0, 0, msk), End: time.Date(2026, 11, 13, 11, 15, 0, 0, msk)}
	last := func() patch {
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.patched[len(f.patched)-1]
	}

	// This occurrence: the occurrence itself.
	if _, err := s.SaveEvent(ctx, occ, model.ScopeThis); err != nil || last().id != "e1" || last().body["recurrence"] != nil {
		t.Fatalf("this: %+v %v", last(), err)
	}
	// The whole series, an hour later: the series from its own first day.
	if _, err := s.SaveEvent(ctx, occ, model.ScopeAll); err != nil {
		t.Fatal(err)
	}
	b := last()
	start := b.body["start"].(map[string]any)
	if b.id != "s" || start["dateTime"] != "2026-11-06T11:00:00+03:00" || start["timeZone"] != "Europe/Moscow" || b.body["recurrence"] != nil {
		t.Fatalf("all: %+v", b)
	}
	// This and the following: the series ends, a new one starts.
	occ.Title = "New standup"
	if _, err := s.SaveEvent(ctx, occ, model.ScopeFollowing); err != nil {
		t.Fatal(err)
	}
	if b := last(); b.id != "s" || fmt.Sprint(b.body["recurrence"]) != "[RRULE:FREQ=WEEKLY;BYDAY=FR;UNTIL=20261113T065959Z EXDATE;TZID=Europe/Moscow:20261120T100000]" {
		t.Fatalf("following, old: %+v", b)
	}
	f.mu.Lock()
	n := f.created[len(f.created)-1]
	f.mu.Unlock()
	if n.Summary != "New standup" || fmt.Sprint(n.Recurrence) != "[RRULE:FREQ=WEEKLY;BYDAY=FR]" || n.Start.TimeZone != "Europe/Moscow" {
		t.Fatalf("following, new: %+v", n)
	}
	// A single event made repeating.
	one := model.CalEvent{CalendarID: "me@gmail.com", ID: "single", Title: "Weekly now", Start: occ.Start, End: occ.End, Recurrence: "FREQ=WEEKLY"}
	if _, err := s.SaveEvent(ctx, one, model.ScopeThis); err != nil || fmt.Sprint(last().body["recurrence"]) != "[RRULE:FREQ=WEEKLY]" || last().body["start"].(map[string]any)["timeZone"] != "Europe/Moscow" {
		t.Fatalf("made repeating: %+v %v", last(), err)
	}

	// Deleting: the occurrence, from it on, the series.
	if err := s.DeleteEvent(ctx, "me@gmail.com", "e1", model.ScopeThis); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteEvent(ctx, "me@gmail.com", "e1", model.ScopeFollowing); err != nil || !strings.Contains(fmt.Sprint(last().body["recurrence"]), "UNTIL=20261113T065959Z") {
		t.Fatalf("delete following: %+v %v", last(), err)
	}
	if err := s.DeleteEvent(ctx, "me@gmail.com", "e1", model.ScopeAll); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.Equal(f.deleted, []string{"e1", "s"}) {
		t.Fatalf("deleted %v", f.deleted)
	}
}
