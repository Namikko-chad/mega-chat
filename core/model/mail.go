package model

import "time"

// FolderRole tells what a mail folder is for, so the same folder of every
// account (all inboxes, all sent folders) can be shown as one.
type FolderRole int

const (
	FolderOther FolderRole = iota // a user's folder or label
	FolderInbox
	FolderSent
	FolderDrafts
	FolderTrash
	FolderSpam
	// FolderArchive: Gmail's "All mail", an IMAP \Archive or \All folder.
	FolderArchive
)

// Folder is a mail folder (IMAP mailbox, Gmail label) of one account.
type Folder struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Role   FolderRole `json:"role"`
	Unread int        `json:"unread"`
	Total  int        `json:"total"`
}

// Address is a mail address with its display name.
type Address struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// Mail is one message of a mailbox. Lists carry the summary (no Text,
// HTML or Attachments); Mailer.MailBody fills the rest.
type Mail struct {
	AccountID string `json:"account_id"`
	// ID is the provider's id, unique inside the account (it names the
	// folder for IMAP).
	ID       string `json:"id"`
	FolderID string `json:"folder_id,omitempty"`
	ThreadID string `json:"thread_id,omitempty"`

	From    Address   `json:"from"`
	To      []Address `json:"to,omitempty"`
	Cc      []Address `json:"cc,omitempty"`
	ReplyTo []Address `json:"reply_to,omitempty"`
	Subject string    `json:"subject"`
	Snippet string    `json:"snippet,omitempty"`
	Time    time.Time `json:"time"`

	Unread         bool `json:"unread,omitempty"`
	Starred        bool `json:"starred,omitempty"`
	HasAttachments bool `json:"has_attachments,omitempty"`

	// Body. Text is the plain text (or the HTML part turned into text);
	// HTML is the original HTML part, if any. Attachments are media refs.
	Text        string       `json:"text,omitempty"`
	HTML        string       `json:"html,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	// MessageID and References are the RFC 5322 headers, for replies.
	MessageID  string `json:"message_id,omitempty"`
	References string `json:"references,omitempty"`
}

// OutMail is a message to send.
type OutMail struct {
	To, Cc, Bcc []Address
	Subject     string
	Text        string
	// ReplyTo is the id of the message this one answers (threading
	// headers come from it); empty: a new message.
	ReplyTo string
}

// MailAction changes messages.
type MailAction int

const (
	MailUnknownAction MailAction = iota
	MailRead
	MailUnread
	MailStar
	MailUnstar
	MailTrash
)
