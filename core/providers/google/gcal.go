package google

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
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
	ID               string    `json:"id,omitempty"`
	Status           string    `json:"status,omitempty"`
	Summary          string    `json:"summary"`
	Description      string    `json:"description,omitempty"`
	Location         string    `json:"location,omitempty"`
	HTMLLink         string    `json:"htmlLink,omitempty"`
	HangoutLink      string    `json:"hangoutLink,omitempty"`
	Start            gTime     `json:"start"`
	End              gTime     `json:"end"`
	RecurringEventID string    `json:"recurringEventId,omitempty"`
	Organizer        *gPerson  `json:"organizer,omitempty"`
	Attendees        []gPerson `json:"attendees,omitempty"`
	ConferenceData   *struct {
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
		Cancelled: g.Status == "cancelled", Link: g.HangoutLink,
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
				out = append(out, e)
			}
		}
		if r.NextPageToken == "" {
			return out, nil
		}
		page = r.NextPageToken
	}
}

func (s *session) SaveEvent(ctx context.Context, ev model.CalEvent) (model.CalEvent, error) {
	g := gEvent{
		Summary: ev.Title, Description: ev.Description, Location: ev.Location,
		Start: toTime(ev.Start, ev.AllDay), End: toTime(ev.End, ev.AllDay),
	}
	for _, a := range ev.Attendees {
		g.Attendees = append(g.Attendees, gPerson{Email: a.Email, DisplayName: a.Name})
	}
	var r gEvent
	var err error
	if ev.ID == "" {
		err = s.c.do(ctx, http.MethodPost, eventsPath(ev.CalendarID), g, &r)
	} else {
		err = s.c.do(ctx, http.MethodPatch, eventsPath(ev.CalendarID)+"/"+url.PathEscape(ev.ID), g, &r)
	}
	if err != nil {
		return model.CalEvent{}, err
	}
	out, ok := r.model()
	if !ok {
		out = ev
		out.ID = r.ID
	}
	out.CalendarID = ev.CalendarID
	return out, nil
}

func (s *session) DeleteEvent(ctx context.Context, cal, id string) error {
	err := s.c.do(ctx, http.MethodDelete, eventsPath(cal)+"/"+url.PathEscape(id), nil, nil)
	if isStatus(err, http.StatusGone) {
		return nil // already deleted
	}
	return err
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
			sink.Emit(model.Event{Kind: model.EvCalendarChanged})
			break
		}
	}
	s.mu.Lock()
	s.calSeen = now
	s.mu.Unlock()
	return nil
}
