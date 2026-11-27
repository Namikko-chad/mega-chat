package google

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/recur"
)

const gcal = "calendar/v3/"

func (s *session) Calendars(ctx context.Context) ([]model.Calendar, error) {
	var out []model.Calendar
	page := ""
	for {
		q := url.Values{"maxResults": {"250"}}
		if page != "" {
			q.Set("pageToken", page)
		}
		var r struct {
			Items []struct {
				ID              string `json:"id"`
				Summary         string `json:"summary"`
				SummaryOverride string `json:"summaryOverride"`
				BackgroundColor string `json:"backgroundColor"`
				AccessRole      string `json:"accessRole"`
				Primary         bool   `json:"primary"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := s.c.do(ctx, http.MethodGet, gcal+"users/me/calendarList?"+q.Encode(), nil, &r); err != nil {
			return nil, err
		}
		for _, it := range r.Items {
			name := it.SummaryOverride
			if name == "" {
				name = it.Summary
			}
			c := model.Calendar{
				ID: it.ID, Name: name, Primary: it.Primary,
				ReadOnly: it.AccessRole != "owner" && it.AccessRole != "writer",
			}
			if v, err := strconv.ParseInt(strings.TrimPrefix(it.BackgroundColor, "#"), 16, 32); err == nil {
				c.Color = int(v)
			}
			out = append(out, c)
		}
		if r.NextPageToken == "" {
			return out, nil
		}
		page = r.NextPageToken
	}
}

type gTime struct {
	Date     string `json:"date,omitempty"`
	DateTime string `json:"dateTime,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

type gPerson struct {
	Email          string `json:"email,omitempty"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
	Self           bool   `json:"self,omitempty"`
}

type gEvent struct {
	ID               string   `json:"id,omitempty"`
	Status           string   `json:"status,omitempty"`
	Summary          string   `json:"summary"`
	Description      string   `json:"description,omitempty"`
	Location         string   `json:"location,omitempty"`
	HTMLLink         string   `json:"htmlLink,omitempty"`
	HangoutLink      string   `json:"hangoutLink,omitempty"`
	Start            gTime    `json:"start"`
	End              gTime    `json:"end"`
	RecurringEventID string   `json:"recurringEventId,omitempty"`
	Recurrence       []string `json:"recurrence,omitempty"` // a series: RRULE, EXDATE, RDATE lines
	// OriginalStartTime: an occurrence's start in its series.
	OriginalStartTime *gTime    `json:"originalStartTime,omitempty"`
	Organizer         *gPerson  `json:"organizer,omitempty"`
	Attendees         []gPerson `json:"attendees,omitempty"`
	ConferenceData    *struct {
		EntryPoints []struct {
			EntryPointType string `json:"entryPointType"`
			URI            string `json:"uri"`
		} `json:"entryPoints"`
	} `json:"conferenceData,omitempty"`
}

var rsvpOf = map[string]model.RSVP{
	"needsAction": model.RSVPNeedsAction, "accepted": model.RSVPAccepted,
	"tentative": model.RSVPTentative, "declined": model.RSVPDeclined,
}

func (t gTime) parse() (time.Time, bool, error) {
	if t.Date != "" {
		d, err := time.Parse("2006-01-02", t.Date)
		return d, true, err
	}
	d, err := time.Parse(time.RFC3339, t.DateTime)
	return d, false, err
}

func toTime(t time.Time, allDay bool) gTime {
	if allDay {
		return gTime{Date: t.UTC().Format("2006-01-02")}
	}
	return gTime{DateTime: t.Format(time.RFC3339)}
}

func (g gEvent) model() (model.CalEvent, bool) {
	start, allDay, err := g.Start.parse()
	if err != nil {
		return model.CalEvent{}, false
	}
	end, _, err := g.End.parse()
	if err != nil {
		return model.CalEvent{}, false
	}
	e := model.CalEvent{
		ID: g.ID, Title: g.Summary, Location: g.Location, Description: g.Description,
		Start: start, End: end, AllDay: allDay, Recurring: g.RecurringEventID != "",
		Cancelled: g.Status == "cancelled", Link: g.HangoutLink, Recurrence: ruleOfLines(g.Recurrence),
	}
	if e.Title == "" {
		e.Title = "(no title)"
	}
	if e.Link == "" && g.ConferenceData != nil {
		for _, p := range g.ConferenceData.EntryPoints {
			if p.EntryPointType == "video" {
				e.Link = p.URI
				break
			}
		}
	}
	if e.Link == "" {
		e.Link = g.HTMLLink
	}
	if o := g.Organizer; o != nil {
		e.Organizer = model.Address{Name: o.DisplayName, Email: o.Email}
	}
	for _, a := range g.Attendees {
		e.Attendees = append(e.Attendees, model.Attendee{
			Address: model.Address{Name: a.DisplayName, Email: a.Email}, Status: rsvpOf[a.ResponseStatus], Self: a.Self,
		})
	}
	return e, true
}

func eventsPath(cal string) string { return gcal + "calendars/" + url.PathEscape(cal) + "/events" }

func (s *session) Events(ctx context.Context, cal string, from, to time.Time) ([]model.CalEvent, error) {
	var out []model.CalEvent
	page := ""
	for {
		q := url.Values{
			"timeMin": {from.Format(time.RFC3339)}, "timeMax": {to.Format(time.RFC3339)},
			"singleEvents": {"true"}, "orderBy": {"startTime"}, "maxResults": {"2500"},
		}
		if page != "" {
			q.Set("pageToken", page)
		}
		var r struct {
			Items         []gEvent `json:"items"`
			NextPageToken string   `json:"nextPageToken"`
		}
		if err := s.c.do(ctx, http.MethodGet, eventsPath(cal)+"?"+q.Encode(), nil, &r); err != nil {
			return nil, err
		}
		for _, g := range r.Items {
			if e, ok := g.model(); ok {
				e.Recurrence = g.RecurringEventID // the series, for its rule below
				out = append(out, e)
			}
		}
		if r.NextPageToken == "" {
			break
		}
		page = r.NextPageToken
	}
	// Occurrences come without their series' rule: it is read once a while.
	for i := range out {
		if series := out[i].Recurrence; series != "" {
			out[i].Recurrence = s.seriesRule(ctx, cal, series)
		}
	}
	return out, nil
}

// ruleOfLines returns the RRULE of a series' recurrence lines.
func ruleOfLines(lines []string) string {
	for _, l := range lines {
		if strings.HasPrefix(strings.ToUpper(l), "RRULE:") {
			return recur.Clean(l)
		}
	}
	return ""
}

// withRule returns recurrence lines with the RRULE replaced (or dropped,
// rule ""), keeping the EXDATE and RDATE lines.
func withRule(lines []string, rule string) []string {
	out := []string{}
	if rule != "" {
		out = append(out, "RRULE:"+rule)
	}
	for _, l := range lines {
		if !strings.HasPrefix(strings.ToUpper(l), "RRULE:") && rule != "" {
			out = append(out, l)
		}
	}
	return out
}

const ruleTTL = 10 * time.Minute

type ruleEntry struct {
	rule string
	at   time.Time
}

// seriesRule returns the rule of a series, "" when it cannot be read.
func (s *session) seriesRule(ctx context.Context, cal, series string) string {
	k := cal + "\x00" + series
	s.mu.Lock()
	e, ok := s.rules[k]
	s.mu.Unlock()
	if ok && time.Since(e.at) < ruleTTL {
		return e.rule
	}
	var g gEvent
	if err := s.c.do(ctx, http.MethodGet, eventsPath(cal)+"/"+url.PathEscape(series), nil, &g); err != nil {
		s.env.Logf("warn", "calendar: series %s: %v", series, err)
		return ""
	}
	r := ruleOfLines(g.Recurrence)
	s.mu.Lock()
	if s.rules == nil {
		s.rules = map[string]ruleEntry{}
	}
	s.rules[k] = ruleEntry{r, time.Now()}
	s.mu.Unlock()
	return r
}

// forgetRules drops the remembered rules: the series changed.
func (s *session) forgetRules() {
	s.mu.Lock()
	s.rules = nil
	s.mu.Unlock()
}

// zoned writes a time in a zone; repeating events need one (Google keeps
// their wall clock over DST).
func zoned(t time.Time, allDay bool, tz string) gTime {
	if allDay || tz == "" {
		return toTime(t, allDay)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return toTime(t, allDay)
	}
	return gTime{DateTime: t.In(loc).Format(time.RFC3339), TimeZone: tz}
}

// calendarZone returns a calendar's time zone, "UTC" when unknown.
func (s *session) calendarZone(ctx context.Context, cal string) string {
	var r struct {
		TimeZone string `json:"timeZone"`
	}
	if err := s.c.do(ctx, http.MethodGet, gcal+"calendars/"+url.PathEscape(cal), nil, &r); err != nil || r.TimeZone == "" {
		return "UTC"
	}
	return r.TimeZone
}

// eventBody is what the app edits, for a POST or a PATCH.
func eventBody(ev model.CalEvent, tz string) map[string]any {
	b := map[string]any{
		"summary": ev.Title, "description": ev.Description, "location": ev.Location,
		"start": zoned(ev.Start, ev.AllDay, tz), "end": zoned(ev.End, ev.AllDay, tz),
	}
	if len(ev.Attendees) > 0 {
		var as []gPerson
		for _, a := range ev.Attendees {
			as = append(as, gPerson{Email: a.Email, DisplayName: a.Name})
		}
		b["attendees"] = as
	}
	return b
}

func (s *session) result(r gEvent, ev model.CalEvent) model.CalEvent {
	out, ok := r.model()
	if !ok {
		out = ev
		out.ID = r.ID
	}
	out.CalendarID = ev.CalendarID
	return out
}

func (s *session) getEvent(ctx context.Context, cal, id string) (gEvent, error) {
	var g gEvent
	err := s.c.do(ctx, http.MethodGet, eventsPath(cal)+"/"+url.PathEscape(id), nil, &g)
	return g, err
}

// occurrence reads an occurrence and its series: the series' start and
// whether it is all day, and the occurrence's start in the series.
func (s *session) occurrence(ctx context.Context, cal string, inst gEvent) (series gEvent, mStart time.Time, mAllDay bool, occ time.Time, err error) {
	if series, err = s.getEvent(ctx, cal, inst.RecurringEventID); err != nil {
		return
	}
	if mStart, mAllDay, err = series.Start.parse(); err != nil {
		return
	}
	o := inst.Start
	if inst.OriginalStartTime != nil {
		o = *inst.OriginalStartTime
	}
	occ, _, err = o.parse()
	return
}

func (s *session) SaveEvent(ctx context.Context, ev model.CalEvent, scope model.RecurScope) (model.CalEvent, error) {
	defer s.forgetRules()
	var r gEvent
	if ev.ID == "" {
		tz := ""
		if ev.Recurrence != "" {
			tz = s.calendarZone(ctx, ev.CalendarID)
		}
		b := eventBody(ev, tz)
		if ev.Recurrence != "" {
			b["recurrence"] = []string{"RRULE:" + ev.Recurrence}
		}
		if err := s.c.do(ctx, http.MethodPost, eventsPath(ev.CalendarID), b, &r); err != nil {
			return model.CalEvent{}, err
		}
		return s.result(r, ev), nil
	}
	cur, err := s.getEvent(ctx, ev.CalendarID, ev.ID)
	if err != nil {
		return model.CalEvent{}, err
	}
	patch := func(id string, b map[string]any) error {
		return s.c.do(ctx, http.MethodPatch, eventsPath(ev.CalendarID)+"/"+url.PathEscape(id), b, &r)
	}
	if cur.RecurringEventID == "" || scope == model.ScopeThis {
		tz := cur.Start.TimeZone
		if tz == "" && ev.Recurrence != "" && cur.RecurringEventID == "" {
			tz = s.calendarZone(ctx, ev.CalendarID)
		}
		b := eventBody(ev, tz)
		if cur.RecurringEventID == "" && ev.Recurrence != "" {
			b["recurrence"] = withRule(cur.Recurrence, ev.Recurrence)
		}
		if err := patch(ev.ID, b); err != nil {
			return model.CalEvent{}, err
		}
		return s.result(r, ev), nil
	}
	series, mStart, mAllDay, occ, err := s.occurrence(ctx, ev.CalendarID, cur)
	if err != nil {
		return model.CalEvent{}, err
	}
	rule := ruleOfLines(series.Recurrence)
	if scope == model.ScopeFollowing && !occ.After(mStart) {
		scope = model.ScopeAll
	}
	tz := series.Start.TimeZone
	if tz == "" && !ev.AllDay {
		tz = s.calendarZone(ctx, ev.CalendarID)
	}
	if scope == model.ScopeAll {
		moved := ev
		moved.Start = mStart.Add(ev.Start.Sub(occ))
		moved.End = moved.Start.Add(ev.End.Sub(ev.Start))
		if ev.AllDay != mAllDay {
			moved.Start, moved.End = ev.Start, ev.End
		}
		b := eventBody(moved, tz)
		if ev.Recurrence != "" {
			b["recurrence"] = withRule(series.Recurrence, ev.Recurrence)
		}
		if err := patch(series.ID, b); err != nil {
			return model.CalEvent{}, err
		}
		return s.result(r, ev), nil
	}
	// This and the following: the series ends before, a new one goes on.
	if err := patch(series.ID, map[string]any{"recurrence": withRule(series.Recurrence, recur.Ending(rule, occ, mAllDay))}); err != nil {
		return model.CalEvent{}, err
	}
	next := ev.Recurrence
	if next == "" || next == rule {
		next = recur.Remaining(rule, mStart, occ)
	}
	b := eventBody(ev, tz)
	b["recurrence"] = []string{"RRULE:" + next}
	r = gEvent{}
	if err := s.c.do(ctx, http.MethodPost, eventsPath(ev.CalendarID), b, &r); err != nil {
		return model.CalEvent{}, err
	}
	return s.result(r, ev), nil
}

func (s *session) DeleteEvent(ctx context.Context, cal, id string, scope model.RecurScope) error {
	defer s.forgetRules()
	del := func(id string) error {
		err := s.c.do(ctx, http.MethodDelete, eventsPath(cal)+"/"+url.PathEscape(id), nil, nil)
		if isStatus(err, http.StatusGone) {
			return nil // already deleted
		}
		return err
	}
	if scope == model.ScopeThis {
		return del(id)
	}
	cur, err := s.getEvent(ctx, cal, id)
	if err != nil {
		return err
	}
	if cur.RecurringEventID == "" {
		return del(id)
	}
	series, mStart, mAllDay, occ, err := s.occurrence(ctx, cal, cur)
	if err != nil {
		return err
	}
	if scope == model.ScopeAll || !occ.After(mStart) {
		return del(series.ID)
	}
	rule := ruleOfLines(series.Recurrence)
	return s.c.do(ctx, http.MethodPatch, eventsPath(cal)+"/"+url.PathEscape(series.ID),
		map[string]any{"recurrence": withRule(series.Recurrence, recur.Ending(rule, occ, mAllDay))}, nil)
}

// pollCalendars looks for events changed since the last look.
func (s *session) pollCalendars(ctx context.Context, sink provider.Sink) error {
	s.mu.Lock()
	since := s.calSeen
	s.mu.Unlock()
	now := time.Now()
	cals, err := s.Calendars(ctx)
	if err != nil {
		return err
	}
	for _, c := range cals {
		q := url.Values{"updatedMin": {since.UTC().Format(time.RFC3339)}, "showDeleted": {"true"}, "maxResults": {"1"}}
		var r struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		if err := s.c.do(ctx, http.MethodGet, eventsPath(c.ID)+"?"+q.Encode(), nil, &r); err != nil {
			if isStatus(err, http.StatusGone) || isStatus(err, http.StatusNotFound) || isStatus(err, http.StatusForbidden) {
				continue
			}
			return err
		}
		if len(r.Items) > 0 {
			s.forgetRules()
			sink.Emit(model.Event{Kind: model.EvCalendarChanged})
			break
		}
	}
	s.mu.Lock()
	s.calSeen = now
	s.mu.Unlock()
	return nil
}

var _ provider.Calendars = (*session)(nil)
