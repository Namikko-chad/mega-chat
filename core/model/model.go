// Package model holds the provider-neutral types every messenger is mapped
// to. Ids are the provider's own: a chat or message id is unique only
// inside one account, so anything that crosses accounts carries AccountID.
package model

import (
	"strings"
	"time"
)

// ChatKind tells a 1:1 dialog from a group and a broadcast channel.
type ChatKind int

const (
	ChatUnknown ChatKind = iota
	ChatDirect
	ChatGroup
	ChatChannel
)

// User is a person or bot as a provider shows them.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
	// Avatar is a media ref of the user's photo (see Media); empty: none.
	Avatar string `json:"avatar,omitempty"`
}

// Chat is a dialog, group or channel of one account.
type Chat struct {
	AccountID string   `json:"account_id"`
	ID        string   `json:"id"`
	Kind      ChatKind `json:"kind"`
	Title     string   `json:"title"`
	// Avatar is a media ref of the chat's photo (see Media).
	Avatar  string    `json:"avatar,omitempty"`
	Unread  int       `json:"unread"`
	Muted   bool      `json:"muted,omitempty"`
	Pinned  bool      `json:"pinned,omitempty"`
	Last    *Message  `json:"last,omitempty"`
	Updated time.Time `json:"updated"` // last activity, for ordering
	// Access tells whether the account may write here. AccessUnknown in
	// an event keeps the stored value.
	Access ChatAccess `json:"access,omitempty"`
	// Forum: the chat is split into topics (Telegram groups with topics),
	// listed with provider.Topics. Its history holds every topic.
	Forum bool `json:"forum,omitempty"`
	// Online: the other person of a direct chat is online now. Live state
	// from the hub (EvPresence), never stored.
	Online bool `json:"-"`
}

// Topic is one topic of a forum chat. Its messages are a thread of the
// chat with the topic's id (Threads.Thread, Outgoing.Thread).
type Topic struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Unread int    `json:"unread"`
	// Color of the topic's icon as 0xRRGGBB (0: none).
	Color   int       `json:"color,omitempty"`
	Pinned  bool      `json:"pinned,omitempty"`
	Closed  bool      `json:"closed,omitempty"` // only admins may write
	Hidden  bool      `json:"hidden,omitempty"` // the general topic, collapsed
	Last    *Message  `json:"last,omitempty"`
	Updated time.Time `json:"updated"`
}

// ChatAccess is what the account may do in a chat.
type ChatAccess int

const (
	AccessUnknown ChatAccess = iota
	AccessWrite
	// AccessRead: the chat can only be read (a broadcast channel of
	// someone else, a group where writing is banned, an archived channel).
	AccessRead
)

// AttachmentKind is what an attachment is, for display.
type AttachmentKind int

const (
	AttachUnknown AttachmentKind = iota
	AttachFile
	AttachImage
	AttachVideo
	AttachAudio
	AttachVoice
	AttachSticker
)

// Attachment describes a file sent with a message. Content is not
// downloaded yet: Ref (the file) and Thumb (a small preview image) are
// media refs, fetched with the session's Media on demand.
type Attachment struct {
	Kind  AttachmentKind `json:"kind"`
	Name  string         `json:"name,omitempty"`
	MIME  string         `json:"mime,omitempty"`
	Size  int64          `json:"size,omitempty"`
	Ref   string         `json:"ref,omitempty"`
	Thumb string         `json:"thumb,omitempty"`
	// Width and Height of an image or video in pixels (0: unknown);
	// Duration of audio and video.
	Width    int           `json:"w,omitempty"`
	Height   int           `json:"h,omitempty"`
	Duration time.Duration `json:"dur,omitempty"`
}

// Reaction is one emoji under a message with how many people chose it.
type Reaction struct {
	// Emoji is the emoji itself ("👍") or, for one the app cannot draw
	// (Slack's custom ones), its name in colons (":party-parrot:").
	Emoji string `json:"e"`
	Count int    `json:"n"`
	Mine  bool   `json:"me,omitempty"` // the account reacted with it
}

// Thread is the discussion under a message: Slack's thread, the comments
// of a Telegram channel post. One with no replies yet means one can be
// started.
type Thread struct {
	// ID names the thread for Threads.Thread (the root's id in Slack, the
	// post's id in Telegram).
	ID      string    `json:"id"`
	Replies int       `json:"n"`
	Last    time.Time `json:"last,omitzero"`
	// Who wrote the latest replies (up to a few), for the avatars under
	// the message.
	Users []User `json:"users,omitempty"`
}

// Message is one message of a chat.
type Message struct {
	AccountID string    `json:"account_id"`
	ChatID    string    `json:"chat_id"`
	ID        string    `json:"id"`
	Sender    User      `json:"sender"`
	Text      string    `json:"text"`
	Time      time.Time `json:"time"`
	// Seq orders messages with the same Time (Telegram/VK message ids
	// grow within a chat); 0 when Time alone is precise enough.
	Seq         int64        `json:"seq,omitempty"`
	Edited      time.Time    `json:"edited,omitzero"`
	Outgoing    bool         `json:"outgoing,omitempty"`
	ReplyTo     string       `json:"reply_to,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Deleted     bool         `json:"deleted,omitempty"`
	Service     bool         `json:"service,omitempty"`
	Reactions   []Reaction   `json:"reactions,omitempty"`
	// Thread: a discussion hangs under this message.
	Thread *Thread `json:"thread,omitempty"`
	// InThread: this message is a reply inside the thread with that id
	// (only in Threads.Thread results and EvThreadMessage events).
	InThread string `json:"in_thread,omitempty"`
	// Group: messages sent together as one album (Telegram's grouped_id)
	// share it; the app shows neighbours with the same Group as one message.
	Group string `json:"group,omitempty"`
	// Mentioned: the message names this account's user (Slack <@me>,
	// @here/@channel; Telegram's mention flag, which replies to us set
	// too; VK Teams @[me]). Muted chats still notify about these.
	Mentioned bool `json:"mentioned,omitempty"`
	// Forwarded: a copy of someone's message; ForwardedFrom names the
	// original's author (empty when the messenger hides them).
	Forwarded     bool   `json:"forwarded,omitempty"`
	ForwardedFrom string `json:"forwarded_from,omitempty"`
	// Quote: the fragment of ReplyTo's text this reply quotes.
	Quote string `json:"quote,omitempty"`
}

// EventKind says what changed.
type EventKind int

const (
	EvUnknown EventKind = iota
	// EvMessageNew: Message is new (incoming or sent from another client).
	EvMessageNew
	// EvMessageEdit: Message replaces the stored one with the same id.
	EvMessageEdit
	// EvMessageDelete: Message carries ChatID and ID only.
	EvMessageDelete
	// EvChat: Chat is new or changed. Empty Title/Avatar, a nil Last and
	// Unread < 0 mean "unknown" and keep the stored values.
	EvChat
	// EvChatRemoved: the account left the chat; Chat carries the ID.
	EvChatRemoved
	// EvRead: the chat was read up to Message.ID (on any device);
	// Chat.Unread is the new count (-1 when unknown).
	EvRead
	// EvTyping: Message.Sender is typing in Message.ChatID.
	EvTyping
	// EvReactions: the reactions of Message.ID (in Message.ChatID) are
	// now Message.Reactions; the rest of Message is empty.
	EvReactions
	// EvThreadMessage: Message is new or edited inside the thread
	// Message.InThread. It is not part of the chat's history.
	EvThreadMessage
	// EvCall: a call was offered, changed state or ended (Event.Call).
	EvCall
	// EvCallSignal: something for the app's peer connection of a call
	// (Event.Call ids, Event.Signal).
	EvCallSignal
	// EvMailNew: Event.Mail is a new message in the inbox (summary only).
	EvMailNew
	// EvMailChanged: folders or messages changed (flags, moves, counts).
	EvMailChanged
	// EvCalendarChanged: events changed; reload the shown range.
	EvCalendarChanged
	// EvPresence: the other person of the direct chat Chat.ID came online
	// (Event.Online) or went away. A session reports what it knows after
	// connecting and then changes; the hub forgets it all on disconnect.
	EvPresence
)

// Event is something a session reports. The account fields are filled by
// the hub; providers leave AccountID empty.
type Event struct {
	Kind    EventKind
	Chat    *Chat
	Message *Message
	// EvCall: the call's new state. EvCallSignal: Call carries the ids
	// only (ID, ChatID) and Signal what to give the app's peer connection.
	Call   *Call
	Signal *CallSignal
	// EvMailNew: the new message.
	Mail *Mail
	// EvPresence: whether the peer of Chat is online.
	Online bool
}

// Outgoing is a message to send.
type Outgoing struct {
	Text    string
	ReplyTo string
	// Thread sends the message into the thread with that id.
	Thread string
	// Quote is a fragment of ReplyTo's text the reply quotes. A provider
	// without native quotes sends QuotedText instead of Text.
	Quote string
}

// QuotedText is Text with Quote before it as "> " lines (Markdown-style
// quoting), for messengers without native quotes.
func (o Outgoing) QuotedText() string {
	q := strings.TrimSpace(o.Quote)
	if q == "" {
		return o.Text
	}
	lines := strings.Split(q, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n") + "\n" + o.Text
}

// CallState is where a 1:1 call is.
type CallState int

const (
	CallUnknown CallState = iota
	// CallWaiting: outgoing, sent, the peer's device does not ring yet.
	CallWaiting
	// CallRinging: outgoing, the peer's device rings; incoming, waiting
	// for the user to answer.
	CallRinging
	// CallConnecting: answered; the media connects.
	CallConnecting
	// CallActive: media flows.
	CallActive
	// CallEnded: over; Call.End says why.
	CallEnded
)

// CallEnd is why a call ended.
type CallEnd int

const (
	EndNone     CallEnd = iota
	EndHangup           // one side hung up after it was answered
	EndDeclined         // the callee (or the caller, cancelling) declined
	EndMissed           // nobody answered in time
	EndBusy             // the peer is in another call
	EndFailed           // it could not connect; Call.Error says why
)

// IceServer is a STUN/TURN server for the call's peer connection. TURN
// credentials are issued per call by the messenger and are not account
// credentials.
type IceServer struct {
	URLs       []string
	Username   string
	Credential string
}

// CallMedia is what one side sends.
type CallMedia struct {
	Muted  bool
	Video  bool // camera on
	Screen bool // sharing the screen
}

// Call is a 1:1 audio or video call. Media runs in the app as a standard
// WebRTC peer connection; the provider carries its offer, answer and
// candidates over the messenger's own signalling.
type Call struct {
	ID       string
	ChatID   string
	Peer     User
	Outgoing bool
	Video    bool // started as a video call
	State    CallState
	End      CallEnd
	Error    string    // EndFailed: why, for the user
	Started  time.Time // when it became active
	// Verification are the emoji both sides see when the call is
	// end-to-end encrypted (empty when not).
	Verification []string
	ICE          []IceServer
	RelayOnly    bool // use only relay (TURN) candidates
	// Polite is the perfect-negotiation role: a polite peer rolls back its
	// own offer when offers collide, an impolite one ignores the other's.
	Polite bool
	// Channel is the label of the data channel the provider uses (empty:
	// none); OpenChannel says this side creates it before its first offer.
	Channel     string
	OpenChannel bool
	// Offer says this side sends the first offer.
	Offer  bool
	Remote CallMedia
}

// Live reports whether the call is not over.
func (c Call) Live() bool { return c.State != CallEnded && c.State != CallUnknown }

// SignalKind is what a CallSignal carries.
type SignalKind int

const (
	SigUnknown   SignalKind = iota
	SigOffer                // SDP offer
	SigAnswer               // SDP answer
	SigCandidate            // ICE candidate (SDP, Mid, MLine); empty SDP: end of candidates
	SigData                 // a text message on the call's data channel
	// From the app only: the peer connection's state.
	SigConnected // media flows
	SigFailed    // it cannot connect (Data: why)
)

// CallSignal goes between the app's peer connection and the provider.
type CallSignal struct {
	// Seq numbers the signals of a call to the app, from 1 (set by the hub).
	Seq   int64
	Kind  SignalKind
	SDP   string
	Mid   string
	MLine int
	Data  string
}
