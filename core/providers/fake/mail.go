package fake

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/recur"
)

// The demo mailbox and calendars of a fake account: a few folders with
// generated messages, and two calendars with a daily stand-up, meetings
// and an all-day event around today.

type mailbox struct {
	mails  []model.Mail // all folders, any order
	seq    int
	events map[string][]model.CalEvent // own events (and series) by calendar
	gone   map[string]bool             // deleted occurrences
	cut    map[string]time.Time        // generated series hidden from then on
}

var demoPeople = []model.Address{
	{Name: "Anna Petrova", Email: "anna@example.com"},
	{Name: "Boris Ivanov", Email: "boris@example.org"},
	{Name: "Clara Schmidt", Email: "clara@example.net"},
	{Name: "GitHub", Email: "noreply@github.com"},
	{Name: "Daniel Kim", Email: "daniel@example.com"},
	{Name: "Travel Agency", Email: "booking@example.travel"},
}

var demoSubjects = []string{
	"Quarterly report draft", "Lunch on Friday?", "Re: release checklist", "[parley] New pull request",
	"Contract for review", "Your booking is confirmed", "Design review notes", "Invoice #%d",
	"Team offsite agenda", "Re: Re: holiday plans",
}

func demoFolders() []model.Folder {
	return []model.Folder{
		{ID: "inbox", Name: "Inbox", Role: model.FolderInbox},
		{ID: "sent", Name: "Sent", Role: model.FolderSent},
		{ID: "drafts", Name: "Drafts", Role: model.FolderDrafts},
		{ID: "archive", Name: "Archive", Role: model.FolderArchive},
		{ID: "spam", Name: "Spam", Role: model.FolderSpam},
		{ID: "trash", Name: "Trash", Role: model.FolderTrash},
		{ID: "projects", Name: "Projects", Role: model.FolderOther},
	}
}

// box returns the account's mailbox, filling it on first use (s.mu held).
func (s *Server) box(login string) *mailbox {
	u := s.user(login)
	if u.mail != nil {
		return u.mail
	}
	b := &mailbox{events: map[string][]model.CalEvent{}, gone: map[string]bool{}, cut: map[string]time.Time{}}
	u.mail = b
	// Logins get different times, so merged inboxes interleave.
	salt := 0
	for _, r := range login {
		salt += int(r)
	}
	now := time.Now().Truncate(time.Minute)
	add := func(folder string, i int, from model.Address, to []model.Address, unread bool) {
		b.seq++
		subj := demoSubjects[(i+salt)%len(demoSubjects)]
		if strings.Contains(subj, "%d") {
			subj = fmt.Sprintf(subj, 1000+i)
		}
		m := model.Mail{
			ID: "m" + strconv.Itoa(b.seq), FolderID: folder, ThreadID: "t" + strconv.Itoa(b.seq),
			From: from, To: to, Subject: subj,
			Time:   now.Add(-time.Duration(i*7+salt%7)*time.Hour - time.Duration(salt%53)*time.Minute),
			Unread: unread, Starred: i%9 == 2, HasAttachments: i%6 == 1,
		}
		m.Text = fmt.Sprintf("Hi,\n\nthis is demo message %d in %s of %s.\n\n> quoted line from an earlier message\n\nBest regards,\n%s", i+1, folder, login, from.Name)
		m.Snippet = strings.ReplaceAll(m.Text, "\n", " ")[:60]
		b.mails = append(b.mails, m)
	}
	me := model.Address{Name: login, Email: strings.ToLower(strings.ReplaceAll(login, " ", ".")) + "@demo.local"}
	for i := range 45 {
		add("inbox", i, demoPeople[(i+salt)%len(demoPeople)], []model.Address{me}, i < 4 || i%11 == 5)
	}
	for i := range 12 {
		add("sent", i, me, []model.Address{demoPeople[i%len(demoPeople)]}, false)
	}
	for i := range 6 {
		add("archive", 50+i, demoPeople[i%len(demoPeople)], []model.Address{me}, false)
	}
	add("drafts", 1, me, []model.Address{demoPeople[2]}, false)
	add("spam", 3, model.Address{Name: "Prize", Email: "win@spam.example"}, []model.Address{me}, true)
	add("trash", 4, demoPeople[1], []model.Address{me}, false)
	add("projects", 2, demoPeople[3], []model.Address{me}, false)
	return b
}

// InjectMail delivers a new message to the inbox of login and returns it.
func (s *Server) InjectMail(login, subject string) model.Mail {
	s.mu.Lock()
	b := s.box(login)
	b.seq++
	m := model.Mail{
		ID: "m" + strconv.Itoa(b.seq), FolderID: "inbox", From: demoPeople[0],
		Subject: subject, Snippet: "A new demo message", Text: "A new demo message.", Time: time.Now(), Unread: true,
	}
	b.mails = append(b.mails, m)
	sinks := s.sinksOf(login)
	s.mu.Unlock()
	for _, k := range sinks {
		mc := m
		mc.Text = ""
		k.Emit(model.Event{Kind: model.EvMailNew, Mail: &mc})
	}
	return m
}

func (s *Server) sinksOf(login string) []provider.Sink {
	var out []provider.Sink
	for _, k := range s.user(login).sinks {
		out = append(out, k)
	}
	return out
}

func (c *session) Folders(context.Context) ([]model.Folder, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	b := c.s.box(c.login)
	fs := demoFolders()
	for i := range fs {
		for _, m := range b.mails {
			if m.FolderID == fs[i].ID {
				fs[i].Total++
				if m.Unread {
					fs[i].Unread++
				}
			}
		}
	}
	return fs, nil
}

func (c *session) Mail(_ context.Context, folder string, before time.Time, limit int) ([]model.Mail, bool, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	var all []model.Mail
	for _, m := range c.s.box(c.login).mails {
		if m.FolderID == folder && (before.IsZero() || !m.Time.After(before)) {
			m.Text = ""
			all = append(all, m)
		}
	}
	slices.SortFunc(all, func(a, b model.Mail) int { return b.Time.Compare(a.Time) })
	if len(all) > limit {
		return all[:limit], true, nil
	}
	return all, false, nil
}

func (c *session) find(id string) (*model.Mail, error) {
	b := c.s.box(c.login)
	for i := range b.mails {
		if b.mails[i].ID == id {
			return &b.mails[i], nil
		}
	}
	return nil, fmt.Errorf("fake: no mail %s", id)
}

func (c *session) MailBody(_ context.Context, id string) (model.Mail, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	m, err := c.find(id)
	if err != nil {
		return model.Mail{}, err
	}
	out := *m
	if out.HasAttachments {
		ref := "img:" + strings.TrimPrefix(id, "m") + ":640:400"
		out.Attachments = []model.Attachment{
			{Kind: model.AttachImage, Name: "photo.png", MIME: "image/png", Ref: ref, Thumb: ref, Width: 640, Height: 400, Size: 640 * 400},
		}
	}
	return out, nil
}

func (c *session) SendMail(_ context.Context, out model.OutMail) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	b := c.s.box(c.login)
	b.seq++
	b.mails = append(b.mails, model.Mail{
		ID: "m" + strconv.Itoa(b.seq), FolderID: "sent", From: model.Address{Name: c.login, Email: "me@demo.local"},
		To: out.To, Cc: out.Cc, Subject: out.Subject, Text: out.Text, Snippet: out.Text, Time: time.Now(),
	})
	return nil
}

func (c *session) UpdateMail(_ context.Context, ids []string, action model.MailAction) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	for _, id := range ids {
		m, err := c.find(id)
		if err != nil {
			return err
		}
		switch action {
		case model.MailRead:
			m.Unread = false
		case model.MailUnread:
			m.Unread = true
		case model.MailStar:
			m.Starred = true
		case model.MailUnstar:
			m.Starred = false
		case model.MailTrash:
			m.FolderID = "trash"
		}
	}
	return nil
}

func (c *session) MoveMail(_ context.Context, ids []string, _, to string) error {
	if !slices.ContainsFunc(demoFolders(), func(f model.Folder) bool { return f.ID == to }) {
		return fmt.Errorf("fake: no folder %s", to)
	}
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	for _, id := range ids {
		m, err := c.find(id)
		if err != nil {
			return err
		}
		m.FolderID = to
	}
	return nil
}

// --- calendars ---------------------------------------------------------------

func (c *session) Calendars(context.Context) ([]model.Calendar, error) {
	return []model.Calendar{
		{ID: "personal", Name: "Personal", Color: 0x10b981, Primary: true},
		{ID: "work", Name: "Work", Color: 0x3b82f6},
		{ID: "holidays", Name: "Holidays", Color: 0xf59e0b, ReadOnly: true},
	}, nil
}

// generated returns the demo events of a calendar in [from, to).
func (c *session) generated(cal string, from, to time.Time) []model.CalEvent {
	var out []model.CalEvent
	day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	for ; day.Before(to); day = day.AddDate(0, 0, 1) {
		wd := day.Weekday()
		n := day.YearDay()
		at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
		key := day.Format("20060102")
		switch cal {
		case "work":
			if wd != time.Saturday && wd != time.Sunday {
				out = append(out, model.CalEvent{
					ID: "standup-" + key, Title: "Stand-up", Start: at(10, 0), End: at(10, 15), Recurring: true, Recurrence: generatedRules["standup"],
					Link: "https://meet.example.com/standup", Organizer: demoPeople[0],
				})
				if n%3 == 0 {
					out = append(out, model.CalEvent{
						ID: "review-" + key, Title: "Design review", Location: "Room 4", Start: at(14, 0), End: at(15, 30),
						Attendees: []model.Attendee{{Address: demoPeople[1], Status: model.RSVPAccepted}, {Address: demoPeople[2], Status: model.RSVPTentative}},
					})
				}
			}
		case "personal":
			if wd == time.Tuesday || wd == time.Thursday {
				out = append(out, model.CalEvent{ID: "gym-" + key, Title: "Gym", Start: at(19, 0), End: at(20, 0), Recurring: true, Recurrence: generatedRules["gym"]})
			}
			if n%10 == 4 {
				out = append(out, model.CalEvent{ID: "dinner-" + key, Title: "Dinner with friends", Location: "Bistro", Start: at(20, 30), End: at(22, 30)})
			}
		case "holidays":
			if n%17 == 0 {
				d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
				out = append(out, model.CalEvent{ID: "hol-" + key, Title: "Demo holiday", Start: d, End: d.AddDate(0, 0, 1), AllDay: true})
			}
		}
	}
	return out
}

// generatedRules are the rules of the generated repeating events.
var generatedRules = map[string]string{"standup": "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", "gym": "FREQ=WEEKLY;BYDAY=TU,TH"}

// generatedSeries returns the series of a generated occurrence's id.
func generatedSeries(id string) (string, bool) {
	name, _, ok := strings.Cut(id, "-")
	_, known := generatedRules[name]
	return name, ok && known
}

// cutOff reports a generated occurrence hidden by a change of its series.
func (b *mailbox) cutOff(e model.CalEvent) bool {
	name, ok := generatedSeries(e.ID)
	if !ok {
		return false
	}
	at, cut := b.cut[name]
	return cut && !e.Start.Before(at)
}

func occurrenceID(series string, t time.Time) string {
	return series + "#" + strconv.FormatInt(t.Unix(), 10)
}

// expand turns own events into what overlaps [from, to): series into
// their occurrences.
func (b *mailbox) expand(evs []model.CalEvent, from, to time.Time) []model.CalEvent {
	var out []model.CalEvent
	for _, e := range evs {
		if e.Recurrence == "" {
			out = append(out, e)
			continue
		}
		dur := e.End.Sub(e.Start)
		starts, err := recur.Between(e.Recurrence, e.Start, from.Add(-dur), to)
		if err != nil {
			continue
		}
		for _, t := range starts {
			o := e
			o.ID, o.Start, o.End, o.Recurring = occurrenceID(e.ID, t), t, t.Add(dur), true
			out = append(out, o)
		}
	}
	return out
}

func (c *session) Events(_ context.Context, cal string, from, to time.Time) ([]model.CalEvent, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	b := c.s.box(c.login)
	var out []model.CalEvent
	for _, e := range append(c.generated(cal, from.AddDate(0, 0, -1), to), b.expand(b.events[cal], from, to)...) {
		if b.gone[e.ID] || b.cutOff(e) || !e.End.After(from) || !e.Start.Before(to) {
			continue
		}
		e.CalendarID = cal
		out = append(out, e)
	}
	return out, nil
}

// own returns the index of an own event (or series).
func (b *mailbox) own(cal, id string) int {
	return slices.IndexFunc(b.events[cal], func(e model.CalEvent) bool { return e.ID == id })
}

func (b *mailbox) add(ev model.CalEvent) model.CalEvent {
	b.seq++
	ev.ID, ev.Recurring = "ev"+strconv.Itoa(b.seq), false
	b.events[ev.CalendarID] = append(b.events[ev.CalendarID], ev)
	return ev
}

func (c *session) SaveEvent(_ context.Context, ev model.CalEvent, scope model.RecurScope) (model.CalEvent, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	if ev.CalendarID == "holidays" {
		return model.CalEvent{}, fmt.Errorf("fake: the calendar is read-only")
	}
	b := c.s.box(c.login)
	if ev.ID == "" {
		return b.add(ev), nil
	}
	series, _, isOcc := strings.Cut(ev.ID, "#")
	if name, ok := generatedSeries(ev.ID); ok && scope != model.ScopeThis {
		// A generated series: hidden from this one (or whole), an own one goes on.
		at := ev.Start
		if scope == model.ScopeAll {
			at = time.Time{}
		}
		b.cut[name] = at
		if ev.Recurrence == "" && scope == model.ScopeFollowing {
			ev.Recurrence = generatedRules[name]
		}
		return b.add(ev), nil
	}
	i := b.own(ev.CalendarID, series)
	if i < 0 { // a generated event or occurrence becomes an own one
		b.gone[ev.ID] = true
		ev.Recurrence = ""
		return b.add(ev), nil
	}
	evs := b.events[ev.CalendarID]
	m := evs[i]
	if !isOcc || m.Recurrence == "" {
		ev.ID = m.ID
		evs[i] = ev
		return ev, nil
	}
	occ, err := c.occurrence(ev.ID)
	if err != nil {
		return model.CalEvent{}, err
	}
	if scope == model.ScopeFollowing && !occ.After(m.Start) {
		scope = model.ScopeAll
	}
	switch scope {
	case model.ScopeThis:
		b.gone[ev.ID] = true
		ev.Recurrence = ""
		return b.add(ev), nil
	case model.ScopeFollowing:
		if ev.Recurrence == "" || ev.Recurrence == m.Recurrence {
			ev.Recurrence = recur.Remaining(m.Recurrence, m.Start, occ)
		}
		evs[i].Recurrence = recur.Ending(m.Recurrence, occ, m.AllDay)
		return b.add(ev), nil
	default:
		moved := ev
		moved.ID = m.ID
		moved.Start = m.Start.Add(ev.Start.Sub(occ))
		moved.End = moved.Start.Add(ev.End.Sub(ev.Start))
		if moved.Recurrence == "" {
			moved.Recurrence = m.Recurrence
		}
		evs[i] = moved
		return moved, nil
	}
}

func (c *session) occurrence(id string) (time.Time, error) {
	_, at, _ := strings.Cut(id, "#")
	n, err := strconv.ParseInt(at, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("fake: bad event id %q", id)
	}
	return time.Unix(n, 0), nil
}

func (c *session) DeleteEvent(_ context.Context, cal, id string, scope model.RecurScope) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	b := c.s.box(c.login)
	if name, ok := generatedSeries(id); ok && scope != model.ScopeThis {
		at := time.Time{}
		if scope == model.ScopeFollowing {
			for _, e := range c.generated(cal, time.Now().AddDate(-1, 0, 0), time.Now().AddDate(1, 0, 0)) {
				if e.ID == id {
					at = e.Start
				}
			}
		}
		b.cut[name] = at
		return nil
	}
	series, _, isOcc := strings.Cut(id, "#")
	i := b.own(cal, series)
	if i < 0 || !isOcc || scope == model.ScopeThis && b.events[cal][i].Recurrence == "" {
		if i >= 0 && !isOcc {
			b.events[cal] = slices.Delete(b.events[cal], i, i+1)
		}
		b.gone[id] = true
		return nil
	}
	m := b.events[cal][i]
	occ, err := c.occurrence(id)
	if err != nil {
		return err
	}
	switch {
	case scope == model.ScopeThis:
		b.gone[id] = true
	case scope == model.ScopeAll || !occ.After(m.Start):
		b.events[cal] = slices.Delete(b.events[cal], i, i+1)
	default:
		b.events[cal][i].Recurrence = recur.Ending(m.Recurrence, occ, m.AllDay)
	}
	return nil
}

var _ provider.Calendars = (*session)(nil)
