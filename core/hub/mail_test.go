package hub

import (
	"context"
	"slices"
	"testing"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

func onlineBoth(t *testing.T, e *env) (a, b AccountInfo) {
	t.Helper()
	a, b = e.login(t, "alice"), e.login(t, "bob")
	waitFor(t, "online", func() bool {
		x, _ := e.h.Account(a.ID)
		y, _ := e.h.Account(b.ID)
		return x.Status == StatusOnline && y.Status == StatusOnline
	})
	return a, b
}

func TestMergedInbox(t *testing.T) {
	e := newEnv(t)
	a, b := onlineBoth(t, e)
	ctx := context.Background()

	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapMail) || !slices.Contains(info.Caps, provider.CapCalendar) {
		t.Fatalf("caps %v", info.Caps)
	}
	waitFor(t, "mail unread", func() bool { x, _ := e.h.Account(a.ID); return x.MailUnread > 0 })

	// Page through all inboxes: every message once, newest first.
	seen := map[[2]string]bool{}
	var q MailQuery
	q.Role, q.Limit = model.FolderInbox, 7
	var last model.Mail
	for pages := 0; ; pages++ {
		if pages > 50 {
			t.Fatal("paging does not end")
		}
		p, err := e.h.Mail(ctx, q)
		if err != nil || len(p.Failed) > 0 {
			t.Fatal(err, p.Failed)
		}
		for _, m := range p.Mails {
			k := [2]string{m.AccountID, m.ID}
			if seen[k] {
				t.Fatalf("%v twice", k)
			}
			seen[k] = true
			if last.ID != "" && mailLess(last, m) >= 0 {
				t.Fatalf("out of order: %v then %v", last.Time, m.Time)
			}
			if m.FolderID != "inbox" {
				t.Fatalf("folder %q", m.FolderID)
			}
			last = m
		}
		if !p.More {
			break
		}
		q.After = p.Next
	}
	if len(seen) != 90 {
		t.Fatalf("saw %d of 90", len(seen))
	}
	var fromA, fromB int
	for k := range seen {
		switch k[0] {
		case a.ID:
			fromA++
		case b.ID:
			fromB++
		}
	}
	if fromA != 45 || fromB != 45 {
		t.Fatalf("a %d b %d", fromA, fromB)
	}

	// One folder of one account.
	p, err := e.h.Mail(ctx, MailQuery{Accounts: []string{a.ID}, Folder: "sent", Limit: 50})
	if err != nil || len(p.Mails) != 12 || p.More {
		t.Fatalf("sent: %d %v %v", len(p.Mails), p.More, err)
	}
	if _, err := e.h.Mail(ctx, MailQuery{Folder: "sent"}); err == nil {
		t.Fatal("a folder id without an account must fail")
	}

	// Reading marks it read and the account's count follows.
	before, _ := e.h.Account(a.ID)
	p, _ = e.h.Mail(ctx, MailQuery{Accounts: []string{a.ID}, Role: model.FolderInbox, Limit: 50})
	i := slices.IndexFunc(p.Mails, func(m model.Mail) bool { return m.Unread })
	body, err := e.h.MailBody(ctx, a.ID, p.Mails[i].ID)
	if err != nil || body.Text == "" {
		t.Fatalf("body %+v %v", body, err)
	}
	if err := e.h.UpdateMail(ctx, a.ID, []string{body.ID}, model.MailRead); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "unread down", func() bool { x, _ := e.h.Account(a.ID); return x.MailUnread == before.MailUnread-1 })

	// A new message is reported and counted.
	e.srv.InjectMail("bob", "Hello there")
	waitFor(t, "mail event", func() bool {
		e.mu.Lock()
		defer e.mu.Unlock()
		return slices.ContainsFunc(e.events, func(ev Event) bool {
			return ev.Kind == EvMailNew && ev.AccountID == b.ID && ev.Mail.Subject == "Hello there"
		})
	})

	if err := e.h.SendMail(ctx, a.ID, model.OutMail{Subject: "x"}); err == nil {
		t.Fatal("no recipients must fail")
	}
	if err := e.h.SendMail(ctx, a.ID, model.OutMail{To: []model.Address{{Email: "x@y"}}, Subject: "hi", Text: "t"}); err != nil {
		t.Fatal(err)
	}
	p, _ = e.h.Mail(ctx, MailQuery{Accounts: []string{a.ID}, Folder: "sent", Limit: 50})
	if len(p.Mails) != 13 {
		t.Fatalf("sent after send: %d", len(p.Mails))
	}
}

func TestMergedCalendar(t *testing.T) {
	e := newEnv(t)
	a, b := onlineBoth(t, e)
	ctx := context.Background()

	cals, errs := e.h.Calendars(ctx, nil)
	if len(errs) > 0 || len(cals) != 6 || !cals[0].Primary {
		t.Fatalf("calendars %+v %v", cals, errs)
	}
	from := time.Now().Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, 14)
	evs, errs, err := e.h.Events(ctx, EventsQuery{From: from, To: to})
	if err != nil || len(errs) > 0 || len(evs) == 0 {
		t.Fatal(err, errs, len(evs))
	}
	accts := map[string]bool{}
	for i, ev := range evs {
		accts[ev.AccountID] = true
		if i > 0 && ev.Start.Before(evs[i-1].Start) {
			t.Fatal("not sorted by start")
		}
	}
	if !accts[a.ID] || !accts[b.ID] {
		t.Fatalf("accounts %v", accts)
	}

	only, _, _ := e.h.Events(ctx, EventsQuery{Calendars: []CalendarKey{{a.ID, "work"}}, From: from, To: to})
	for _, ev := range only {
		if ev.AccountID != a.ID || ev.CalendarID != "work" {
			t.Fatalf("filter leaked %+v", ev)
		}
	}

	start := from.Add(36 * time.Hour)
	saved, err := e.h.SaveEvent(ctx, model.CalEvent{AccountID: b.ID, CalendarID: "personal", Title: "Dentist", Start: start, End: start.Add(time.Hour)})
	if err != nil || saved.ID == "" || saved.AccountID != b.ID {
		t.Fatalf("save %+v %v", saved, err)
	}
	if _, err := e.h.SaveEvent(ctx, model.CalEvent{AccountID: b.ID, CalendarID: "personal", Title: "Bad", Start: start, End: start}); err == nil {
		t.Fatal("an empty event must fail")
	}
	has := func() bool {
		evs, _, _ := e.h.Events(ctx, EventsQuery{From: from, To: to})
		return slices.ContainsFunc(evs, func(ev model.CalEvent) bool { return ev.ID == saved.ID && ev.Title == "Dentist" })
	}
	if !has() {
		t.Fatal("saved event not listed")
	}
	if err := e.h.DeleteEvent(ctx, b.ID, "personal", saved.ID); err != nil || has() {
		t.Fatal("not deleted", err)
	}
	if _, _, err := e.h.Events(ctx, EventsQuery{From: to, To: from}); err == nil {
		t.Fatal("an empty range must fail")
	}
}

func TestMoveMail(t *testing.T) {
	e := newEnv(t)
	a, _ := onlineBoth(t, e)
	ctx := context.Background()
	var q MailQuery
	q.Accounts, q.Role, q.Limit = []string{a.ID}, model.FolderInbox, 3
	p, err := e.h.Mail(ctx, q)
	if err != nil || len(p.Mails) != 3 {
		t.Fatalf("inbox: %+v %v", p, err)
	}
	ids := []string{p.Mails[0].ID, p.Mails[1].ID}
	if err := e.h.MoveMail(ctx, a.ID, ids, "inbox", "projects"); err != nil {
		t.Fatal(err)
	}
	q.Role, q.Folder, q.Limit = model.FolderOther, "projects", 100
	p, err = e.h.Mail(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	moved := 0
	for _, m := range p.Mails {
		if slices.Contains(ids, m.ID) {
			moved++
		}
	}
	if moved != 2 {
		t.Fatalf("moved %d of 2", moved)
	}
	if err := e.h.MoveMail(ctx, a.ID, ids, "projects", "nowhere"); err == nil {
		t.Fatal("moved into a missing folder")
	}
}
