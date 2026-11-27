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

	"github.com/SelfHold/parley/core/crypto"
	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/recur"
)

type calClient struct {
	c    *caldav.Client // discovery
	raw  davRaw         // objects
	home string
	logf func(string, ...any)
}

// davClient talks CalDAV at base; logf (may be nil) is told which request
// a server refused, since its errors carry only the status.
func davClient(hc *http.Client, cfg config, base string, logf func(string, ...any)) (*caldav.Client, error) {
	c, _, err := davClients(hc, cfg, base, logf)
	return c, err
}

// davClients returns go-webdav's client and one for raw requests.
func davClients(hc *http.Client, cfg config, base string, logf func(string, ...any)) (*caldav.Client, davRaw, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, davRaw{}, err
	}
	h := keepSlash{webdav.HTTPClientWithBasicAuth(hc, cfg.Email, cfg.Password), u.Path, logf}
	c, err := caldav.NewClient(h, base)
	return c, davRaw{hc: h, base: u}, err
}

// keepSlash puts back the trailing slash of the base path: go-webdav drops
// it when it asks the base itself, and Mail.ru answers /principals with 404
// (only /principals/ exists).
type keepSlash struct {
	webdav.HTTPClient
	base string
	logf func(string, ...any)
}

func (k keepSlash) Do(req *http.Request) (*http.Response, error) {
	if len(k.base) > 1 && strings.HasSuffix(k.base, "/") && req.URL.Path == strings.TrimSuffix(k.base, "/") {
		req = req.Clone(req.Context())
		req.URL.Path, req.URL.RawPath = k.base, ""
	}
	resp, err := k.HTTPClient.Do(req)
	if err == nil && resp.StatusCode >= 400 && k.logf != nil {
		depth := req.Header.Get("Depth")
		k.logf("caldav: %s %s (depth %q): %s", req.Method, req.URL.Path, depth, resp.Status)
	}
	return resp, err
}

// discoverCalendars finds the account's calendar home set: the principal
// at the address given, else through /.well-known/caldav.
func discoverCalendars(ctx context.Context, hc *http.Client, cfg config) (string, error) {
	var firstErr error
	for _, base := range []string{cfg.CalDAV, strings.TrimSuffix(cfg.CalDAV, "/") + "/.well-known/caldav"} {
		c, err := davClient(hc, cfg, base, nil)
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
	if c, err := davClient(hc, cfg, cfg.CalDAV, nil); err == nil {
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

func newCalClient(hc *http.Client, cfg config, logf func(string, ...any)) (*calClient, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	c, raw, err := davClients(hc, cfg, cfg.CalDAV, logf)
	if err != nil {
		return nil, err
	}
	return &calClient{c: c, raw: raw, home: cfg.CalHome, logf: logf}, nil
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
		objs, err := k.raw.query(ctx, c.ID, now.AddDate(0, -1, 0), now.AddDate(0, 6, 0), false, k.logf)
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
		if recurring && master != nil {
			e.Recurrence = ruleOf(master)
		}
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
		e.Recurrence = ruleOf(master)
		out = append(out, e)
	}
	return out
}

func (k *calClient) events(ctx context.Context, calPath string, from, to time.Time) ([]model.CalEvent, error) {
	return k.eventsAs(ctx, calPath, from, to, "")
}

func (k *calClient) eventsAs(ctx context.Context, calPath string, from, to time.Time, self string) ([]model.CalEvent, error) {
	objs, err := k.raw.query(ctx, calPath, from, to, true, k.logf)
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

// setTime writes a DTSTART-like property: a date, a time in the zone the
// property had (a repeating event keeps its wall clock over DST), or UTC.
func setTime(props ical.Props, name string, t time.Time, allDay bool) {
	p := ical.NewProp(name)
	switch {
	case allDay:
		p.SetDate(t.UTC())
	default:
		if old := props.Get(name); old != nil {
			if tz := old.Params.Get(ical.PropTimezoneID); tz != "" {
				if loc, err := time.LoadLocation(tz); err == nil {
					p.Params.Set(ical.PropTimezoneID, tz)
					p.Value = t.In(loc).Format("20060102T150405")
					break
				}
			}
		}
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

// setRule sets (or, empty, drops) a VEVENT's RRULE.
func setRule(c *ical.Component, rule string) {
	c.Props.Del(ical.PropRecurrenceRule)
	if rule != "" {
		p := ical.NewProp(ical.PropRecurrenceRule)
		p.Value = rule
		c.Props.Set(p)
	}
}

// ruleOf returns a VEVENT's RRULE value, "" when it has none.
func ruleOf(c *ical.Component) string {
	if p := c.Props.Get(ical.PropRecurrenceRule); p != nil {
		return recur.Clean(p.Value)
	}
	return ""
}

// splitID parses an event id: an object path and, for an occurrence of a
// repeating event, the occurrence's (original) start.
func splitID(id string) (path string, occ time.Time, isOcc bool, err error) {
	path, at, isOcc := strings.Cut(id, "#")
	if !isOcc {
		return path, time.Time{}, false, nil
	}
	occ, err = time.ParseInLocation(occurrenceFormat, at, time.UTC)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("mailru: bad event id %q", id)
	}
	return path, occ, true, nil
}

// seriesOf returns the VEVENT of a calendar object that is the series (or
// the single event), and the overrides of single occurrences.
func seriesOf(cal *ical.Calendar) (master *ical.Component, overrides []*ical.Component) {
	for _, c := range cal.Children {
		if c.Name != ical.CompEvent {
			continue
		}
		if c.Props.Get(ical.PropRecurrenceID) != nil {
			overrides = append(overrides, c)
		} else if master == nil {
			master = c
		}
	}
	return master, overrides
}

// recurrenceID reads an override's RECURRENCE-ID.
func recurrenceID(c *ical.Component) (time.Time, bool) {
	t, _, err := propTime(c.Props.Get(ical.PropRecurrenceID))
	return t, err == nil
}

// occurrenceProp makes a RECURRENCE-ID or EXDATE for an occurrence of a
// series whose start is all day or not.
func occurrenceProp(name string, occ time.Time, allDay bool) ical.Prop {
	p := ical.NewProp(name)
	if allDay {
		p.SetDate(occ.UTC())
	} else {
		p.SetDateTime(occ.UTC())
	}
	return *p
}

// dropChildren removes the VEVENTs keep says no to.
func dropChildren(cal *ical.Calendar, keep func(*ical.Component) bool) {
	out := cal.Children[:0]
	for _, c := range cal.Children {
		if c.Name != ical.CompEvent || keep(c) {
			out = append(out, c)
		}
	}
	cal.Children = out
}

// cloneEvent copies a VEVENT's properties (not its alarms and other
// children).
func cloneEvent(c *ical.Component) *ical.Component {
	out := ical.NewComponent(c.Name)
	for name, ps := range c.Props {
		out.Props[name] = append([]ical.Prop(nil), ps...)
	}
	return out
}

func newObjectPath(calendarID string) (path, uid string) {
	uid = crypto.RandomID(16) + "@parley"
	return strings.TrimSuffix(calendarID, "/") + "/" + url.PathEscape(uid) + ".ics", uid
}

func newCalendarObject() *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//Parley//EN")
	return cal
}

func (k *calClient) save(ctx context.Context, ev model.CalEvent, self string, scope model.RecurScope) (model.CalEvent, error) {
	if ev.ID == "" {
		path, uid := newObjectPath(ev.CalendarID)
		cal := newCalendarObject()
		e := ical.NewEvent()
		e.Props.SetText(ical.PropUID, uid)
		if len(ev.Attendees) > 0 && self != "" {
			o := ical.NewProp(ical.PropOrganizer)
			o.Value = "mailto:" + self
			e.Props.Set(o)
		}
		fill(e.Component, ev)
		setRule(e.Component, ev.Recurrence)
		cal.Children = append(cal.Children, e.Component)
		if err := k.raw.put(ctx, path, cal); err != nil {
			return model.CalEvent{}, err
		}
		ev.ID = path
		return ev, nil
	}
	path, occ, isOcc, err := splitID(ev.ID)
	if err != nil {
		return model.CalEvent{}, err
	}
	obj, err := k.raw.get(ctx, path)
	if err != nil {
		return model.CalEvent{}, err
	}
	cal := obj.Data
	master, overrides := seriesOf(cal)
	if master == nil {
		return model.CalEvent{}, fmt.Errorf("mailru: no event in %s", path)
	}
	mStart, mEnd, mAllDay, err := span(master)
	if err != nil {
		return model.CalEvent{}, err
	}
	rule := ruleOf(master)
	if !isOcc || rule == "" {
		// A single event; it repeats from now on when given a rule.
		fill(master, ev)
		if ev.Recurrence != "" {
			setRule(master, ev.Recurrence)
		}
		if err := k.raw.put(ctx, path, cal); err != nil {
			return model.CalEvent{}, err
		}
		ev.ID = path
		return ev, nil
	}
	if scope == model.ScopeFollowing && !occ.After(mStart) {
		scope = model.ScopeAll // from the first one: the whole series
	}
	switch scope {
	case model.ScopeThis:
		var o *ical.Component
		for _, c := range overrides {
			if t, ok := recurrenceID(c); ok && t.Equal(occ) {
				o = c
			}
		}
		if o == nil {
			o = cloneEvent(master)
			for _, name := range []string{ical.PropRecurrenceRule, ical.PropRecurrenceDates, ical.PropExceptionDates} {
				o.Props.Del(name)
			}
			rid := occurrenceProp(ical.PropRecurrenceID, occ, mAllDay)
			o.Props.Set(&rid)
			cal.Children = append(cal.Children, o)
		}
		fill(o, ev)
		ev.ID = path + "#" + occ.UTC().Format(occurrenceFormat)
	case model.ScopeAll:
		delta := ev.Start.Sub(occ)
		moved := ev
		moved.Start = mStart.Add(delta)
		moved.End = moved.Start.Add(ev.End.Sub(ev.Start))
		if ev.AllDay != mAllDay {
			moved.Start, moved.End = ev.Start, ev.End // a new kind of time: from this occurrence
		}
		fill(master, moved)
		if ev.Recurrence != "" {
			setRule(master, ev.Recurrence)
		}
		if delta != 0 || ev.AllDay != mAllDay || moved.End.Sub(moved.Start) != mEnd.Sub(mStart) {
			// The occurrences moved: their own changes no longer fit.
			dropChildren(cal, func(c *ical.Component) bool { return c.Props.Get(ical.PropRecurrenceID) == nil })
			if delta != 0 || ev.AllDay != mAllDay {
				// Deleted occurrences stay deleted where they moved to.
				ex := propTimes(master.Props.Values(ical.PropExceptionDates))
				master.Props.Del(ical.PropExceptionDates)
				if ev.AllDay == mAllDay {
					for _, t := range ex {
						p := occurrenceProp(ical.PropExceptionDates, t.Add(delta), ev.AllDay)
						master.Props.Add(&p)
					}
				}
			}
		}
		ev.ID = path + "#" + moved.Start.UTC().Format(occurrenceFormat)
	case model.ScopeFollowing:
		next := newCalendarObject()
		for _, c := range cal.Children {
			if c.Name != ical.CompEvent { // the time zones
				next.Children = append(next.Children, c)
			}
		}
		n := cloneEvent(master)
		for _, name := range []string{ical.PropRecurrenceDates, ical.PropExceptionDates} {
			n.Props.Del(name)
		}
		newPath, uid := newObjectPath(path[:strings.LastIndex(path, "/")+1])
		n.Props.SetText(ical.PropUID, uid)
		fill(n, ev)
		nr := ev.Recurrence
		if nr == "" || nr == rule {
			nr = recur.Remaining(rule, mStart, occ)
		}
		setRule(n, nr)
		next.Children = append(next.Children, n)
		setRule(master, recur.Ending(rule, occ, mAllDay))
		dropChildren(cal, func(c *ical.Component) bool {
			t, ok := recurrenceID(c)
			return !ok || t.Before(occ)
		})
		if err := k.raw.put(ctx, newPath, next); err != nil {
			return model.CalEvent{}, err
		}
		ev.ID = newPath + "#" + ev.Start.UTC().Format(occurrenceFormat)
	default:
		return model.CalEvent{}, fmt.Errorf("mailru: unknown scope %d", scope)
	}
	if err := k.raw.put(ctx, path, cal); err != nil {
		return model.CalEvent{}, err
	}
	ev.Recurrence = ruleOf(master)
	if scope == model.ScopeFollowing {
		ev.Recurrence = ""
	}
	return ev, nil
}

func (k *calClient) remove(ctx context.Context, id string, scope model.RecurScope) error {
	path, occ, isOcc, err := splitID(id)
	if err != nil {
		return err
	}
	if !isOcc || scope == model.ScopeAll {
		return k.c.RemoveAll(ctx, path)
	}
	obj, err := k.raw.get(ctx, path)
	if err != nil {
		return err
	}
	cal := obj.Data
	master, _ := seriesOf(cal)
	if master == nil {
		return k.c.RemoveAll(ctx, path)
	}
	mStart, _, mAllDay, err := span(master)
	if err != nil {
		return err
	}
	rule := ruleOf(master)
	switch {
	case scope == model.ScopeFollowing && !occ.After(mStart):
		return k.c.RemoveAll(ctx, path)
	case scope == model.ScopeFollowing && rule != "":
		setRule(master, recur.Ending(rule, occ, mAllDay))
		dropChildren(cal, func(c *ical.Component) bool {
			t, ok := recurrenceID(c)
			return !ok || t.Before(occ)
		})
	default: // this occurrence
		ex := occurrenceProp(ical.PropExceptionDates, occ, mAllDay)
		master.Props.Add(&ex)
		dropChildren(cal, func(c *ical.Component) bool {
			t, ok := recurrenceID(c)
			return !ok || !t.Equal(occ)
		})
	}
	return k.raw.put(ctx, path, cal)
}
