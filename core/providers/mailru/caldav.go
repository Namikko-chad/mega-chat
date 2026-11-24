package mailru

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/teambition/rrule-go"

	"megachat/core/crypto"
	"megachat/core/model"
	"megachat/core/provider"
)

// ErrRepeating: occurrences of a repeating CalDAV event cannot be changed
// one by one yet.
var ErrRepeating = fmt.Errorf("a repeating event can only be changed in its own calendar app: %w", provider.ErrUnsupported)

type calClient struct {
	c    *caldav.Client
	home string
}

func davClient(hc *http.Client, cfg config, base string) (*caldav.Client, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	return caldav.NewClient(keepSlash{webdav.HTTPClientWithBasicAuth(hc, cfg.Email, cfg.Password), u.Path}, base)
}

// keepSlash puts back the trailing slash of the base path: go-webdav drops
// it when it asks the base itself, and Mail.ru answers /principals with 404
// (only /principals/ exists).
type keepSlash struct {
	webdav.HTTPClient
	base string
}

func (k keepSlash) Do(req *http.Request) (*http.Response, error) {
	if len(k.base) > 1 && strings.HasSuffix(k.base, "/") && req.URL.Path == strings.TrimSuffix(k.base, "/") {
		req = req.Clone(req.Context())
		req.URL.Path, req.URL.RawPath = k.base, ""
	}
	return k.HTTPClient.Do(req)
}

// discoverCalendars finds the account's calendar home set: the principal
// at the address given, else through /.well-known/caldav.
func discoverCalendars(ctx context.Context, hc *http.Client, cfg config) (string, error) {
	var firstErr error
	for _, base := range []string{cfg.CalDAV, strings.TrimSuffix(cfg.CalDAV, "/") + "/.well-known/caldav"} {
		c, err := davClient(hc, cfg, base)
		if err != nil {
			return "", err
		}
		principal, err := c.FindCurrentUserPrincipal(ctx)
		if err == nil {
			var home string
			if home, err = c.FindCalendarHomeSet(ctx, principal); err == nil {
				if _, err = c.FindCalendars(ctx, home); err == nil {
					return home, nil
				}
			}
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	// The address may name the principal itself.
	if c, err := davClient(hc, cfg, cfg.CalDAV); err == nil {
		if u, err := url.Parse(cfg.CalDAV); err == nil {
			if home, err := c.FindCalendarHomeSet(ctx, u.Path); err == nil {
				if _, err = c.FindCalendars(ctx, home); err == nil {
					return home, nil
				}
			}
		}
	}
	return "", firstErr
}

func newCalClient(hc *http.Client, cfg config) (*calClient, error) {
	c, err := davClient(hc, cfg, cfg.CalDAV)
	if err != nil {
		return nil, err
	}
	return &calClient{c: c, home: cfg.CalHome}, nil
}

func (k *calClient) homeSet(ctx context.Context) (string, error) {
	if k.home != "" {
		return k.home, nil
	}
	p, err := k.c.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}
	k.home, err = k.c.FindCalendarHomeSet(ctx, p)
	return k.home, err
}

func (k *calClient) calendars(ctx context.Context) ([]model.Calendar, error) {
	home, err := k.homeSet(ctx)
	if err != nil {
		return nil, err
	}
	cs, err := k.c.FindCalendars(ctx, home)
	if err != nil {
		return nil, err
	}
	var out []model.Calendar
	for _, c := range cs {
		if len(c.SupportedComponentSet) > 0 && !containsFold(c.SupportedComponentSet, ical.CompEvent) {
			continue // tasks only
		}
		name := c.Name
		if name == "" {
			name = strings.Trim(c.Path[strings.LastIndex(strings.TrimSuffix(c.Path, "/"), "/")+1:], "/")
		}
		out = append(out, model.Calendar{ID: c.Path, Name: name, Primary: len(out) == 0})
	}
	return out, nil
}

func containsFold(l []string, s string) bool {
	for _, x := range l {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

// state is a digest of the etags of every calendar's objects around now:
// it changes when an event does.
func (k *calClient) state(ctx context.Context) (string, error) {
	cals, err := k.calendars(ctx)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, c := range cals {
		now := time.Now()
		q := &caldav.CalendarQuery{
			CompRequest: caldav.CalendarCompRequest{Name: ical.CompCalendar, Props: []string{ical.PropUID}},
			CompFilter: caldav.CompFilter{Name: ical.CompCalendar, Comps: []caldav.CompFilter{
				{Name: ical.CompEvent, Start: now.AddDate(0, -1, 0).UTC(), End: now.AddDate(0, 6, 0).UTC()},
			}},
		}
		objs, err := k.c.QueryCalendar(ctx, c.ID, q)
		if err != nil {
			return "", err
		}
		for _, o := range objs {
			fmt.Fprintf(h, "%s\x00%s\x00", o.Path, o.ETag)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// propTime reads a DTSTART-like property: a date (all day, as midnight
// UTC), a UTC time, a time in its TZID, or a floating time (local).
func propTime(p *ical.Prop) (t time.Time, allDay bool, err error) {
	if p == nil {
		return time.Time{}, false, errors.New("no time")
	}
	v := p.Value
	if p.ValueType() == ical.ValueDate || len(v) == len("20060102") {
		t, err = time.ParseInLocation("20060102", v, time.UTC)
		return t, true, err
	}
	if strings.HasSuffix(v, "Z") {
		t, err = time.ParseInLocation("20060102T150405Z", v, time.UTC)
		return t, false, err
	}
	loc := time.Local
	if tz := p.Params.Get(ical.PropTimezoneID); tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	t, err = time.ParseInLocation("20060102T150405", v, loc)
	return t, false, err
}

// propTimes reads a list property (EXDATE, RDATE), several values each.
func propTimes(props []ical.Prop) []time.Time {
	var out []time.Time
	for _, p := range props {
		for _, v := range strings.Split(p.Value, ",") {
			q := p
			q.Value = strings.TrimSpace(v)
			if t, _, err := propTime(&q); err == nil {
				out = append(out, t)
			}
		}
	}
	return out
}

func text(c *ical.Component, name string) string {
	s, _ := c.Props.Text(name)
	return s
}

var partstat = map[string]model.RSVP{
	"NEEDS-ACTION": model.RSVPNeedsAction, "ACCEPTED": model.RSVPAccepted,
	"TENTATIVE": model.RSVPTentative, "DECLINED": model.RSVPDeclined,
}

func person(p ical.Prop) model.Address {
	return model.Address{Name: p.Params.Get(ical.ParamCommonName), Email: strings.TrimPrefix(strings.TrimPrefix(p.Value, "mailto:"), "MAILTO:")}
}

// eventOf maps a VEVENT with its start and end.
func eventOf(c *ical.Component, id string, start, end time.Time, allDay bool, self string) model.CalEvent {
	e := model.CalEvent{
		ID: id, Title: text(c, ical.PropSummary), Location: text(c, ical.PropLocation),
		Description: text(c, ical.PropDescription), Start: start, End: end, AllDay: allDay,
		Cancelled: strings.EqualFold(text(c, ical.PropStatus), "CANCELLED"),
	}
	if e.Title == "" {
		e.Title = "(no title)"
	}
	if u, err := c.Props.URI(ical.PropURL); err == nil && u != nil {
		e.Link = u.String()
	}
	if o := c.Props.Get(ical.PropOrganizer); o != nil {
		e.Organizer = person(*o)
	}
	for _, a := range c.Props.Values(ical.PropAttendee) {
		p := person(a)
		e.Attendees = append(e.Attendees, model.Attendee{
			Address: p, Status: partstat[strings.ToUpper(a.Params.Get(ical.ParamParticipationStatus))],
			Self: strings.EqualFold(p.Email, self),
		})
	}
	return e
}

// span returns a VEVENT's start, end and whether it lasts whole days.
func span(c *ical.Component) (time.Time, time.Time, bool, error) {
	start, allDay, err := propTime(c.Props.Get(ical.PropDateTimeStart))
	if err != nil {
		return time.Time{}, time.Time{}, false, err
	}
	end := start
	if p := c.Props.Get(ical.PropDateTimeEnd); p != nil {
		end, _, err = propTime(p)
	} else if p := c.Props.Get(ical.PropDuration); p != nil {
		var d time.Duration
		d, err = p.Duration()
		end = start.Add(d)
	} else if allDay {
		end = start.AddDate(0, 0, 1)
	}
	return start, end, allDay, err
}

const occurrenceFormat = "20060102T150405Z"

// expand turns the VEVENTs of one calendar object into the events (and
// occurrences of a repeating one) that overlap [from, to).
func expand(path string, cal *ical.Calendar, from, to time.Time, self string) []model.CalEvent {
	var master *ical.Component
	overrides := map[int64]*ical.Component{}
	for _, c := range cal.Children {
		if c.Name != ical.CompEvent {
			continue
		}
		if rid := c.Props.Get(ical.PropRecurrenceID); rid != nil {
			if t, _, err := propTime(rid); err == nil {
				overrides[t.Unix()] = c
			}
			continue
		}
		master = c
	}
	var out []model.CalEvent
	add := func(c *ical.Component, id string, recurring bool) {
		start, end, allDay, err := span(c)
		if err != nil || !end.After(from) || !start.Before(to) {
			return
		}
		e := eventOf(c, id, start, end, allDay, self)
		e.Recurring = recurring
		out = append(out, e)
	}
	if master == nil {
		for t, c := range overrides { // orphan occurrences
			add(c, path+"#"+time.Unix(t, 0).UTC().Format(occurrenceFormat), true)
		}
		return out
	}
	opt, err := master.Props.RecurrenceRule()
	if err != nil || opt == nil {
		add(master, path, false)
		return out
	}
	start, end, allDay, err := span(master)
	if err != nil {
		return out
	}
	opt.Dtstart = start
	rule, err := rrule.NewRRule(*opt)
	if err != nil {
		add(master, path, false)
		return out
	}
	set := &rrule.Set{}
	set.RRule(rule)
	for _, t := range propTimes(master.Props.Values(ical.PropRecurrenceDates)) {
		set.RDate(t)
	}
	for _, t := range propTimes(master.Props.Values(ical.PropExceptionDates)) {
		set.ExDate(t)
	}
	dur := end.Sub(start)
	for _, t := range set.Between(from.Add(-dur), to, true) {
		id := path + "#" + t.UTC().Format(occurrenceFormat)
		if c, ok := overrides[t.Unix()]; ok {
			add(c, id, true)
			continue
		}
		if !t.Add(dur).After(from) {
			continue
		}
		e := eventOf(master, id, t, t.Add(dur), allDay, self)
		e.Recurring = true
		out = append(out, e)
	}
	return out
}

func (k *calClient) events(ctx context.Context, calPath string, from, to time.Time) ([]model.CalEvent, error) {
	return k.eventsAs(ctx, calPath, from, to, "")
}

func (k *calClient) eventsAs(ctx context.Context, calPath string, from, to time.Time, self string) ([]model.CalEvent, error) {
	q := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{Name: ical.CompCalendar, AllProps: true, AllComps: true},
		CompFilter: caldav.CompFilter{Name: ical.CompCalendar, Comps: []caldav.CompFilter{
			{Name: ical.CompEvent, Start: from.UTC(), End: to.UTC()},
		}},
	}
	objs, err := k.c.QueryCalendar(ctx, calPath, q)
	if err != nil {
		return nil, err
	}
	var out []model.CalEvent
	for _, o := range objs {
		if o.Data != nil {
			out = append(out, expand(o.Path, o.Data, from, to, self)...)
		}
	}
	return out, nil
}

func setTime(props ical.Props, name string, t time.Time, allDay bool) {
	p := ical.NewProp(name)
	if allDay {
		p.SetDate(t.UTC())
	} else {
		p.SetDateTime(t.UTC())
	}
	props.Set(p)
}

// fill writes what the app edits into a VEVENT, keeping everything else.
func fill(c *ical.Component, ev model.CalEvent) {
	c.Props.SetText(ical.PropSummary, ev.Title)
	for name, v := range map[string]string{ical.PropLocation: ev.Location, ical.PropDescription: ev.Description} {
		if v == "" {
			c.Props.Del(name)
		} else {
			c.Props.SetText(name, v)
		}
	}
	setTime(c.Props, ical.PropDateTimeStart, ev.Start, ev.AllDay)
	setTime(c.Props, ical.PropDateTimeEnd, ev.End, ev.AllDay)
	c.Props.Del(ical.PropDuration)
	c.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	if ev.Attendees != nil {
		c.Props.Del(ical.PropAttendee)
		for _, a := range ev.Attendees {
			p := ical.NewProp(ical.PropAttendee)
			p.Value = "mailto:" + a.Email
			if a.Name != "" {
				p.Params.Set(ical.ParamCommonName, a.Name)
			}
			p.Params.Set(ical.ParamParticipationStatus, "NEEDS-ACTION")
			c.Props.Add(p)
		}
	}
}

func (k *calClient) save(ctx context.Context, ev model.CalEvent, self string) (model.CalEvent, error) {
	if strings.Contains(ev.ID, "#") {
		return model.CalEvent{}, ErrRepeating
	}
	var cal *ical.Calendar
	path := ev.ID
	if path == "" {
		uid := crypto.RandomID(16) + "@megachat"
		path = strings.TrimSuffix(ev.CalendarID, "/") + "/" + url.PathEscape(uid) + ".ics"
		cal = ical.NewCalendar()
		cal.Props.SetText(ical.PropVersion, "2.0")
		cal.Props.SetText(ical.PropProductID, "-//MegaChat//EN")
		e := ical.NewEvent()
		e.Props.SetText(ical.PropUID, uid)
		if len(ev.Attendees) > 0 && self != "" {
			o := ical.NewProp(ical.PropOrganizer)
			o.Value = "mailto:" + self
			e.Props.Set(o)
		}
		fill(e.Component, ev)
		cal.Children = append(cal.Children, e.Component)
	} else {
		obj, err := k.c.GetCalendarObject(ctx, path)
		if err != nil {
			return model.CalEvent{}, err
		}
		cal = obj.Data
		var master *ical.Component
		for _, c := range cal.Children {
			if c.Name == ical.CompEvent && c.Props.Get(ical.PropRecurrenceID) == nil {
				master = c
			}
		}
		if master == nil {
			return model.CalEvent{}, fmt.Errorf("mailru: no event in %s", path)
		}
		if master.Props.Get(ical.PropRecurrenceRule) != nil {
			return model.CalEvent{}, ErrRepeating
		}
		fill(master, ev)
	}
	if _, err := k.c.PutCalendarObject(ctx, path, cal); err != nil {
		return model.CalEvent{}, err
	}
	ev.ID = path
	return ev, nil
}

func (k *calClient) remove(ctx context.Context, id string) error {
	if strings.Contains(id, "#") {
		return ErrRepeating
	}
	return k.c.RemoveAll(ctx, id)
}
