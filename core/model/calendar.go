package model

import "time"

// Calendar is one calendar of an account.
type Calendar struct {
	AccountID string `json:"account_id"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	// Color as 0xRRGGBB; 0: none (the app picks one).
	Color    int  `json:"color,omitempty"`
	ReadOnly bool `json:"read_only,omitempty"`
	Primary  bool `json:"primary,omitempty"`
}

// RSVP is an attendee's answer.
type RSVP int

const (
	RSVPUnknown RSVP = iota
	RSVPNeedsAction
	RSVPAccepted
	RSVPTentative
	RSVPDeclined
)

// Attendee is a person invited to an event.
type Attendee struct {
	Address
	Status RSVP `json:"status,omitempty"`
	Self   bool `json:"self,omitempty"` // the account itself
}

// CalEvent is one event (or one occurrence of a repeating event).
type CalEvent struct {
	AccountID  string `json:"account_id"`
	CalendarID string `json:"calendar_id"`
	// ID names the event (occurrence) for SaveEvent and DeleteEvent.
	ID          string `json:"id"`
	Title       string `json:"title"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	// Start and End (exclusive). An all-day event spans whole days: its
	// Start and End are midnights UTC of those dates.
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	AllDay bool      `json:"all_day,omitempty"`
	// Recurring: an occurrence of a repeating event.
	Recurring bool       `json:"recurring,omitempty"`
	Cancelled bool       `json:"cancelled,omitempty"`
	Organizer Address    `json:"organizer,omitzero"`
	Attendees []Attendee `json:"attendees,omitempty"`
	// Link is a page or video meeting of the event (Google Meet, ...).
	Link string `json:"link,omitempty"`
	// Recurrence is the rule of a repeating event: an RFC 5545 RRULE value
	// without "RRULE:" ("FREQ=WEEKLY;BYDAY=MO"); occurrences carry their
	// series' rule, "" when it is not known. See SaveEvent for its use there.
	Recurrence string `json:"recurrence,omitempty"`
}

// RecurScope is what of a repeating event a change or a delete of one of
// its occurrences touches.
type RecurScope int

const (
	ScopeThis      RecurScope = iota // this occurrence only
	ScopeFollowing                   // this and the following occurrences
	ScopeAll                         // the whole series
)
