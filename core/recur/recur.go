// Package recur works on the rules of repeating calendar events: RFC 5545
// RRULE values without "RRULE:" ("FREQ=WEEKLY;BYDAY=MO"), as
// model.CalEvent.Recurrence carries them. Providers use it to split a
// series ("this and the following") and to expand one.
package recur

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// ErrRule: a rule that cannot be used.
var ErrRule = errors.New("not a repeat rule")

// Clean takes a rule with or without "RRULE:" and returns it without.
func Clean(r string) string {
	r = strings.TrimSpace(r)
	if len(r) >= 6 && strings.EqualFold(r[:6], "RRULE:") {
		r = r[6:]
	}
	return r
}

// Valid checks a rule.
func Valid(r string) error {
	if _, err := rrule.StrToROption(Clean(r)); err != nil || Clean(r) == "" {
		return fmt.Errorf("%w: %q", ErrRule, r)
	}
	return nil
}

func parts(r string) []string {
	var out []string
	for _, p := range strings.Split(Clean(r), ";") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func key(p string) string {
	k, _, _ := strings.Cut(p, "=")
	return strings.ToUpper(k)
}

// get returns the value of a rule part, "" when absent.
func get(r, k string) string {
	for _, p := range parts(r) {
		if key(p) == k {
			_, v, _ := strings.Cut(p, "=")
			return v
		}
	}
	return ""
}

// set replaces (or drops, v == "") parts of a rule, keeping the order of
// the others; a new part goes at the end.
func set(r string, kv ...string) string {
	ps := parts(r)
	for i := 0; i+1 < len(kv); i += 2 {
		k, v := kv[i], kv[i+1]
		found := false
		out := ps[:0:0]
		for _, p := range ps {
			if key(p) != k {
				out = append(out, p)
				continue
			}
			if v != "" && !found {
				out = append(out, k+"="+v)
			}
			found = true
		}
		if !found && v != "" {
			out = append(out, k+"="+v)
		}
		ps = out
	}
	return strings.Join(ps, ";")
}

// Ending returns rule r cut so its last occurrence starts before t: UNTIL
// just before t (a date for all-day events), COUNT dropped.
func Ending(r string, t time.Time, allDay bool) string {
	until := t.Add(-time.Second).UTC().Format("20060102T150405Z")
	if allDay {
		until = t.UTC().AddDate(0, 0, -1).Format("20060102")
	}
	return set(r, "COUNT", "", "UNTIL", until)
}

// rule builds the rule of a series that starts at start.
func rule(r string, start time.Time) (*rrule.RRule, error) {
	opt, err := rrule.StrToROptionInLocation(Clean(r), start.Location())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRule, err)
	}
	opt.Dtstart = start
	return rrule.NewRRule(*opt)
}

// Remaining returns the rule of a series that goes on from occurrence occ
// of a series that started at start: with COUNT, as many as were left.
func Remaining(r string, start, occ time.Time) string {
	c := get(r, "COUNT")
	if c == "" {
		return Clean(r)
	}
	n, err := strconv.Atoi(c)
	if err != nil {
		return Clean(r)
	}
	if rr, err := rule(r, start); err == nil && occ.After(start) {
		n -= len(rr.Between(start, occ.Add(-time.Second), true))
	}
	return set(r, "COUNT", strconv.Itoa(max(n, 1)))
}

// Between returns the starts of a series' occurrences in [from, to); the
// series starts at start.
func Between(r string, start, from, to time.Time) ([]time.Time, error) {
	rr, err := rule(r, start)
	if err != nil {
		return nil, err
	}
	var out []time.Time
	for _, t := range rr.Between(from, to, true) {
		if t.Before(to) {
			out = append(out, t)
		}
	}
	return out, nil
}
