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
	// Online: set only on the people Members returns. Live, never stored.
	Online bool `json:"-"`
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
	// Saved: the account's own notes (Telegram's Saved Messages, the chat
	// with oneself in VK Teams and Slack). The app names it itself.
	Saved bool `json:"saved,omitempty"`
	// Archived: the chat is in the messenger's archive (Telegram's archive
	// folder, a hidden VK Teams chat, an archived Slack channel).
	Archived bool `json:"archived,omitempty"`
	// Folders are the ids of the account's chat folders (ChatFolder) the
	// chat belongs to by the folders' rules; nil: unknown, keep the stored.
	Folders []string `json:"folders,omitempty"`
}

// ChatFolder is a folder of an account's chat list (Telegram's chat
// folders). A chat is in it when its id is in Always, or Chat.Folders
// names the folder and none of the Exclude flags applies.
type ChatFolder struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Emoji  string   `json:"emoji,omitempty"`
	Always []string `json:"always,omitempty"` // chats added by hand
	Pinned []string `json:"pinned,omitempty"` // shown first, in this order (also in Always)
	// Exclude chats that are muted, read or archived (unless in Always).
	ExcludeMuted    bool `json:"exclude_muted,omitempty"`
	ExcludeRead     bool `json:"exclude_read,omitempty"`
	ExcludeArchived bool `json:"exclude_archived,omitempty"`
}

// Contains tells whether c is in the folder.
func (f ChatFolder) Contains(c Chat) bool {
	for _, id := range f.Always {
		if id == c.ID {
			return true
		}
	}
	in := false
	for _, id := range c.Folders {
		in = in || id == f.ID
	}
	return in && !(f.ExcludeMuted && c.Muted) && !(f.ExcludeRead && c.Unread == 0) && !(f.ExcludeArchived && c.Archived)
}

// ChatAction is something the account does to one of its chats.
type ChatAction int

const (
	ChatActionUnknown ChatAction = iota
	ChatMute
	ChatUnmute
	ChatPin
	ChatUnpin
	ChatArchive
	ChatUnarchive
	// ChatLeave: leave a group or channel; it disappears from the list.
	ChatLeave
)

// ChatFlags are the account's own settings of a chat that changed (nil:
// unchanged), for EvChatFlags.
type ChatFlags struct {
	Muted, Pinned, Archived *bool
}

// Apply sets what f carries on c.
func (f ChatFlags) Apply(c *Chat) {
	if f.Muted != nil {
		c.Muted = *f.Muted
	}
	if f.Pinned != nil {
		c.Pinned = *f.Pinned
	}
	if f.Archived != nil {
		c.Archived = *f.Archived
	}
}

// FlagsOf is what action changes, ok false for an action that changes no
// flag (leaving).
func FlagsOf(action ChatAction) (f ChatFlags, ok bool) {
	t, n := true, false
	switch action {
	case ChatMute:
		f.Muted = &t
	case ChatUnmute:
		f.Muted = &n
	case ChatPin:
		f.Pinned = &t
	case ChatUnpin:
		f.Pinned = &n
	case ChatArchive:
		f.Archived = &t
	case ChatUnarchive:
		f.Archived = &n
	default:
		return f, false
	}
	return f, true
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

// Profile describes a chat for its info page: the other person of a direct
// chat, a group or a channel. Empty fields are unknown or hidden.
type Profile struct {
	Title    string
	Avatar   string // a media ref (Media), larger when the messenger has one
	Username string // @handle without "@", Slack display name, VK nick
	About    string // bio, description, channel purpose and topic
	Phone    string
	Email    string
	Position string // job title, status text
	// Link is a public address of the chat or person (https://t.me/x).
	Link string
	// Members is how many members a group or channel has (0: unknown).
	Members int
	// MemberList is some of the members (may be empty or partial).
	MemberList []User
	// LastSeen is when the person of a direct chat was last online (zero:
	// unknown or hidden); Online tells they are online now.
	LastSeen time.Time
	Online   bool
	Bot      bool
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
	// Followed: a thread reply (InThread) in a thread this account's user
	// follows (Slack: started, replied to or subscribed to it). It notifies
	// like a message of the chat; unfollowed replies only when Mentioned.
	Followed bool `json:"followed,omitempty"`
	// Forwarded: a copy of someone's message; ForwardedFrom names the
	// original's author (empty when the messenger hides them).
	Forwarded     bool   `json:"forwarded,omitempty"`
	ForwardedFrom string `json:"forwarded_from,omitempty"`
	// Quote: the fragment of ReplyTo's text this reply quotes.
	Quote string `json:"quote,omitempty"`
	// Format styles Text, which holds no markup itself (core/richtext).
	Format []TextSpan `json:"format,omitempty"`
}

// SpanKind is a style of a part of a text.
type SpanKind int

const (
	SpanUnknown SpanKind = iota
	SpanBold
	SpanItalic
	SpanUnderline
	SpanStrike
	SpanSpoiler
	SpanCode
	SpanPre
	SpanLink
	SpanQuote
	SpanHeading
)

// TextSpan styles a part of a text; offsets are UTF-16 units, as in
// Telegram and in the app's strings. Spans may nest and overlap.
type TextSpan struct {
	Kind   SpanKind `json:"k"`
	Offset int      `json:"o"`
	Length int      `json:"l"`
	URL    string   `json:"u,omitempty"`  // SpanLink: http(s) or mailto
	Lang   string   `json:"lg,omitempty"` // SpanPre: the code's language
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
	// EvChatFlags: the account's settings of Chat.ID (mute, pin, archive)
	// changed, maybe on another device; Event.Flags says which.
	EvChatFlags
	// EvChatFolders: the account's chat folders changed (ChatFolders); the
	// chats' Folders come with the next chat list.
	EvChatFolders
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
	// EvChatFlags: what changed.
	Flags *ChatFlags
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
	// Files are sent with the message, Text being their caption. They go
	// to the session's FileSender, never to Send.
	Files []OutFile
	// Mentions are people picked in the composer: spans of Text (byte
	// offsets, sorted, not overlapping) the messenger turns into its own
	// mentions.
	Mentions []Mention
}

// Mention is a span of an outgoing text that names a person.
type Mention struct {
	Offset, Length int // bytes of Outgoing.Text
	User           User
}

// MentionedText is text with every mention span replaced by mention(m,
// span) and the rest passed through plain (nil: as is).
func MentionedText(text string, ms []Mention, mention func(m Mention, span string) string, plain func(string) string) string {
	if plain == nil {
		plain = func(s string) string { return s }
	}
	var b strings.Builder
	at := 0
	for _, m := range ms {
		if m.Offset < at || m.Offset+m.Length > len(text) {
			continue
		}
		b.WriteString(plain(text[at:m.Offset]))
		b.WriteString(mention(m, text[m.Offset:m.Offset+m.Length]))
		at = m.Offset + m.Length
	}
	b.WriteString(plain(text[at:]))
	return b.String()
}

// MatchUser tells whether a person fits what was typed after "@": a word
// of the name or the username starts with query (any case); "" fits all.
func MatchUser(u User, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	if strings.HasPrefix(strings.ToLower(u.Username), q) || strings.HasPrefix(strings.ToLower(u.Name), q) {
		return true
	}
	for _, w := range strings.FieldsFunc(strings.ToLower(u.Name), func(r rune) bool { return r == ' ' || r == '.' || r == '-' || r == '_' }) {
		if strings.HasPrefix(w, q) {
			return true
		}
	}
	return false
}

// OutFile is a file to send. The hub fills in Name and MIME when the app
// left them empty.
type OutFile struct {
	Name string
	MIME string
	Data []byte
}

// Kind is how a messenger should show the file.
func (f OutFile) Kind() AttachmentKind {
	switch {
	case f.MIME == "image/jpeg" || f.MIME == "image/png" || f.MIME == "image/gif" || f.MIME == "image/webp":
		return AttachImage
	case strings.HasPrefix(f.MIME, "video/"):
		return AttachVideo
	case strings.HasPrefix(f.MIME, "audio/"):
		return AttachAudio
	}
	return AttachFile
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
