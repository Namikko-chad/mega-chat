package provider

import (
	"context"
	"time"

	"github.com/SelfHold/parley/core/model"
)

// Mailer is implemented by sessions with a mailbox (CapMail). Live changes
// come as model.EvMailNew (a new message in the inbox) and
// model.EvMailChanged (anything else: flags, moves, counts).
type Mailer interface {
	// Folders returns the account's folders with their counts.
	Folders(ctx context.Context) ([]model.Folder, error)
	// Mail returns up to limit messages of a folder received at or before
	// before (the latest when zero), newest first; more is false when the
	// folder holds no older ones. Bodies are not filled.
	Mail(ctx context.Context, folderID string, before time.Time, limit int) (mails []model.Mail, more bool, err error)
	// MailBody returns one message with its body and attachments (media
	// refs, fetched with Media). It does not mark it read.
	MailBody(ctx context.Context, id string) (model.Mail, error)
	SendMail(ctx context.Context, out model.OutMail) error
	UpdateMail(ctx context.Context, ids []string, action model.MailAction) error
}

// MailMover is implemented by mailers that can move messages between
// folders (Gmail: change their labels).
type MailMover interface {
	// MoveMail moves ids, listed in folder from, to folder to. Moving to
	// the trash is MailTrash's business, but works too.
	MoveMail(ctx context.Context, ids []string, from, to string) error
}

// Calendars is implemented by sessions with calendars (CapCalendar). Live
// changes come as model.EvCalendarChanged.
type Calendars interface {
	Calendars(ctx context.Context) ([]model.Calendar, error)
	// Events returns the events of a calendar that overlap [from, to),
	// repeating events expanded into their occurrences.
	Events(ctx context.Context, calendarID string, from, to time.Time) ([]model.CalEvent, error)
	// SaveEvent creates (empty ID) or changes an event and returns it as
	// stored. A new event, or one that did not repeat, repeats when
	// ev.Recurrence is set. For an occurrence of a repeating event scope
	// says what changes: the occurrence, it and the following ones (a new
	// series from it; ev.Recurrence its rule, the old one or "" go on with
	// what was left of a COUNT), or the whole series (moved by as much as
	// the occurrence moved; ev.Recurrence its new rule, "" keeps it).
	SaveEvent(ctx context.Context, ev model.CalEvent, scope model.RecurScope) (model.CalEvent, error)
	// DeleteEvent deletes an event; for an occurrence scope says what goes.
	DeleteEvent(ctx context.Context, calendarID, id string, scope model.RecurScope) error
}
