package recur

import (
	"testing"
	"time"
)

func TestEnding(t *testing.T) {
	at := time.Date(2026, 11, 25, 13, 0, 0, 0, time.UTC)
	if got := Ending("RRULE:FREQ=WEEKLY;COUNT=10;BYDAY=WE", at, false); got != "FREQ=WEEKLY;BYDAY=WE;UNTIL=20261125T125959Z" {
		t.Fatal(got)
	}
	if got := Ending("FREQ=DAILY;UNTIL=20270101", at, true); got != "FREQ=DAILY;UNTIL=20261124" {
		t.Fatal(got)
	}
}

func TestRemaining(t *testing.T) {
	start := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	occ := start.AddDate(0, 0, 3)
	if got := Remaining("FREQ=DAILY;COUNT=5", start, occ); got != "FREQ=DAILY;COUNT=2" {
		t.Fatal(got)
	}
	if got := Remaining("FREQ=DAILY;INTERVAL=2", start, occ); got != "FREQ=DAILY;INTERVAL=2" {
		t.Fatal(got)
	}
}

func TestBetweenAndValid(t *testing.T) {
	start := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC) // a Monday
	got, err := Between("FREQ=WEEKLY;BYDAY=MO,WE", start, start, start.AddDate(0, 0, 7))
	if err != nil || len(got) != 2 || got[1].Weekday() != time.Wednesday {
		t.Fatal(got, err)
	}
	if Valid("FREQ=WEEKLY") != nil || Valid("nonsense") == nil || Valid("") == nil {
		t.Fatal("valid")
	}
}
