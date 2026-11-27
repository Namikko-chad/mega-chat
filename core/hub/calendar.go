package hub

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"megachat/core/eventlog"
	"megachat/core/model"
	"megachat/core/recur"
)

// calendarTTL: how long a calendar list is used before it is fetched again.
const calendarTTL = 10 * time.Minute

// maxRange caps one Events read.
const maxRange = 400 * 24 * time.Hour

func (h *Hub) calendarList(ctx context.Context, acct string) ([]model.Calendar, error) {
	h.mail.mu.Lock()
	e, ok := h.mail.calendars[acct]
	h.mail.mu.Unlock()
	if ok && time.Since(e.at) < calendarTTL {
		return e.list, nil
	}
	c, err := h.calendarsOf(acct)
	if err != nil {
		return nil, err
	}
	list, err := c.Calendars(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].AccountID = acct
	}
	h.mail.mu.Lock()
	if h.mail.calendars == nil {
		h.mail.calendars = map[string]calEntry{}
	}
	h.mail.calendars[acct] = calEntry{list: list, at: time.Now()}
	h.mail.mu.Unlock()
	return list, nil
}

// Calendars returns the calendars of the given accounts (all calendar
// accounts when empty), account by account, primary first.
func (h *Hub) Calendars(ctx context.Context, accounts []string) ([]model.Calendar, []AccountError) {
	accounts = h.withCap(accounts, isCalendars)
	var mu sync.Mutex
	got := map[string][]model.Calendar{}
	errs := fanOut(ctx, accounts, func(ctx context.Context, a string) error {
		cs, err := h.calendarList(ctx, a)
		if err == nil {
			mu.Lock()
			got[a] = cs
			mu.Unlock()
		}
		return err
	})
	var out []model.Calendar
	for _, a := range accounts {
		cs := slices.Clone(got[a])
		slices.SortStableFunc(cs, func(x, y model.Calendar) int {
			if x.Primary != y.Primary {
				if x.Primary {
					return -1
				}
				return 1
			}
			return cmp.Compare(x.Name, y.Name)
		})
		out = append(out, cs...)
	}
	return out, errs
}

// EventsQuery asks for the events of a time range.
type EventsQuery struct {
	Accounts []string
	// Calendars limits the read to these calendars (of any account, as
	// "account/calendar" keys); empty: all of them.
	Calendars []CalendarKey
	From, To  time.Time
}

// CalendarKey names a calendar across accounts.
type CalendarKey struct{ Account, Calendar string }

// Events returns the events of all calendars of the given accounts that
// overlap [from, to), merged and sorted by start.
func (h *Hub) Events(ctx context.Context, q EventsQuery) ([]model.CalEvent, []AccountError, error) {
	if !q.To.After(q.From) {
		return nil, nil, errors.New("empty time range")
	}
	if q.To.Sub(q.From) > maxRange {
		return nil, nil, errors.New("time range too long")
	}
	accounts := q.Accounts
	if len(accounts) == 0 && len(q.Calendars) > 0 {
		for _, k := range q.Calendars {
			if !slices.Contains(accounts, k.Account) {
				accounts = append(accounts, k.Account)
			}
		}
	}
	accounts = h.withCap(accounts, isCalendars)
	var mu sync.Mutex
	var out []model.CalEvent
	errs := fanOut(ctx, accounts, func(ctx context.Context, a string) error {
		c, err := h.calendarsOf(a)
		if err != nil {
			return err
		}
		cals, err := h.calendarList(ctx, a)
		if err != nil {
			return err
		}
		var firstErr error
		for _, cal := range cals {
			if len(q.Calendars) > 0 && !slices.Contains(q.Calendars, CalendarKey{a, cal.ID}) {
				continue
			}
			evs, err := c.Events(ctx, cal.ID, q.From, q.To)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			mu.Lock()
			for _, e := range evs {
				if e.Cancelled || !e.End.After(q.From) || !e.Start.Before(q.To) {
					continue
				}
				e.AccountID, e.CalendarID = a, cal.ID
				out = append(out, e)
			}
			mu.Unlock()
		}
		return firstErr
	})
	slices.SortFunc(out, func(x, y model.CalEvent) int {
		if c := x.Start.Compare(y.Start); c != 0 {
			return c
		}
		if x.AllDay != y.AllDay {
			if x.AllDay {
				return -1
			}
			return 1
		}
		if c := cmp.Compare(x.AccountID, y.AccountID); c != 0 {
			return c
		}
		return cmp.Compare(x.ID, y.ID)
	})
	return out, errs, nil
}

// SaveEvent creates or changes an event; scope says what of a repeating
// event an occurrence's change touches.
func (h *Hub) SaveEvent(ctx context.Context, ev model.CalEvent, scope model.RecurScope) (model.CalEvent, error) {
	if ev.Title == "" {
		return model.CalEvent{}, errors.New("an event needs a title")
	}
	if !ev.End.After(ev.Start) {
		return model.CalEvent{}, errors.New("an event must end after it starts")
	}
	if ev.Recurrence = recur.Clean(ev.Recurrence); ev.Recurrence != "" {
		if err := recur.Valid(ev.Recurrence); err != nil {
			return model.CalEvent{}, err
		}
	}
	c, err := h.calendarsOf(ev.AccountID)
	if err != nil {
		return model.CalEvent{}, err
	}
	acct := ev.AccountID
	out, err := c.SaveEvent(ctx, ev, scope)
	if err != nil {
		return model.CalEvent{}, err
	}
	out.AccountID = acct
	if out.CalendarID == "" {
		out.CalendarID = ev.CalendarID
	}
	h.logf(eventlog.Info, "calendar", acct, "saved %q", ev.Title)
	h.emit(Event{Kind: EvCalendar, AccountID: acct})
	return out, nil
}

// DeleteEvent deletes an event; scope says what of a repeating event goes
// with an occurrence.
func (h *Hub) DeleteEvent(ctx context.Context, acct, calendar, id string, scope model.RecurScope) error {
	c, err := h.calendarsOf(acct)
	if err != nil {
		return err
	}
	if err := c.DeleteEvent(ctx, calendar, id, scope); err != nil {
		return err
	}
	h.emit(Event{Kind: EvCalendar, AccountID: acct})
	return nil
}
