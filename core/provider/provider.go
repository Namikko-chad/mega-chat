// Package provider is the contract between the core and a messenger
// (Telegram, Slack, VK Teams, ...). A provider describes itself (Info),
// signs accounts in through a short dialog (Login) and opens a Session for
// a signed-in account. What a session can do is told by the capability
// interfaces it implements: Messenger today; calls, mail and calendar will
// be further interfaces next to it, so adding them touches neither the
// existing providers nor the hub's account handling.
package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"megachat/core/model"
)

// Capability is a feature a provider offers. Values match the Capability
// enum in proto/megachat.proto.
type Capability int

const (
	CapMessages   Capability = 1
	CapCalls      Capability = 2 // Caller: 1:1 audio/video calls
	CapGroupCalls Capability = 3 // reserved: conferences
	CapMail       Capability = 4 // Mailer (the hub adds it)
	CapCalendar   Capability = 5 // Calendars (the hub adds it)
	// Told by the session's interfaces (the hub adds them), not by Info.
	CapThreads   Capability = 6  // Threads
	CapReactions Capability = 7  // Reactor
	CapMedia     Capability = 8  // Media
	CapTopics    Capability = 9  // Topics
	CapForward   Capability = 10 // Forwarder
)

var (
	// ErrAuth means the stored credentials no longer work (session
	// revoked, token expired): the account must sign in again. Wrap it.
	ErrAuth = errors.New("sign-in required")
	// ErrOffline means the session is not connected right now.
	ErrOffline = errors.New("account is offline")
	// ErrUnsupported means the provider cannot do this.
	ErrUnsupported = errors.New("not supported by this messenger")
)

// AuthMethod is one way to sign in (a provider may have several).
type AuthMethod struct {
	ID          string
	Name        string
	Description string
}

// Info describes a provider.
type Info struct {
	Type    string // stable id, stored in accounts: "telegram"
	Name    string // shown to the user
	Methods []AuthMethod
	Caps    []Capability
}

// Field is one input of a login step.
type Field struct {
	Key       string
	Label     string
	Help      string
	Secret    bool
	Optional  bool
	Multiline bool
	Default   string
}

// Step is what a login shows next.
type Step struct {
	Title   string
	Message string
	Error   string // the previous answer was rejected
	Fields  []Field
	// OpenURL is a page to open in the browser (OAuth). Next may then
	// finish by itself when called with no values.
	OpenURL string
	// QR is a text to show as a QR code to scan with another device
	// (Telegram's "Link Desktop Device"). Like OpenURL, Next may then
	// finish by itself when called with no values.
	QR     string
	Done   bool
	Result *Credentials // set when Done
}

// Credentials are what a finished login produces.
type Credentials struct {
	// Login identifies who signed in (normalised phone, team:user id), so
	// the same account is not added twice.
	Login string
	// Name is a suggested account name ("Acme Slack", "+7 999…").
	Name string
	// Params go to the vault (tokens, server); State to the account's KV
	// (provider session state such as Telegram's auth key).
	Params map[string]string
	State  map[string][]byte
}

// Login is one sign-in dialog.
type Login interface {
	// Next takes the values for the current step's fields (nil on the
	// first call) and returns the next step. It may wait for something
	// outside (an OAuth redirect) until ctx ends; then it returns the
	// current step again with a nil error.
	Next(ctx context.Context, values map[string]string) (Step, error)
	// Close drops the login (finished or not).
	Close()
}

// KV is an account's private, encrypted key-value state that survives
// restarts. Get returns nil, nil for a missing key.
type KV interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
}

// Env is what the core hands to a provider.
type Env struct {
	AccountID string
	// State is the account's KV. During a login it is a scratch KV whose
	// content is not kept: put what must survive into Credentials.State.
	State KV
	HTTP  *http.Client
	Log   func(level, msg string)
}

// Logf logs through env.Log (which may be nil).
func (e Env) Logf(level, format string, args ...any) {
	if e.Log != nil {
		e.Log(level, fmt.Sprintf(format, args...))
	}
}

// Account is what a session is opened with.
type Account struct {
	ID     string
	Login  string
	Params map[string]string
}

// Sink receives what a running session reports.
type Sink interface {
	// Emit reports a change. Events may come from any goroutine.
	Emit(model.Event)
	// Connected is called whenever the session is live (again).
	Connected(self model.User)
	// SaveParams replaces the account's vault params (refreshed token).
	SaveParams(params map[string]string) error
}

// Provider is one messenger.
type Provider interface {
	Info() Info
	// Login starts signing in with the given method. For a re-login, acct
	// is the existing account (to prefill and keep the same identity);
	// otherwise nil.
	Login(ctx context.Context, method string, acct *Account, env Env) (Login, error)
	// Open returns a session for a signed-in account; it does not connect.
	Open(acct Account, env Env) (Session, error)
}

// Session is a signed-in account.
type Session interface {
	// Run connects and reports events to sink until ctx ends. It returns
	// an error wrapping ErrAuth when the credentials stopped working; other
	// errors mean "try again later" and the hub reconnects with backoff.
	Run(ctx context.Context, sink Sink) error
}

// Messenger is the chat capability. Calls may come before or while Run
// connects; a session that needs its connection returns ErrOffline.
type Messenger interface {
	// Chats returns up to limit chats with the latest activity.
	Chats(ctx context.Context, limit int) ([]model.Chat, error)
	// History returns up to limit messages before beforeID (the latest
	// when empty), newest first; more is false when the start was reached.
	History(ctx context.Context, chatID, beforeID string, limit int) (msgs []model.Message, more bool, err error)
	Send(ctx context.Context, chatID string, out model.Outgoing) (model.Message, error)
	// MarkRead marks the chat read up to msgID (the latest when empty).
	MarkRead(ctx context.Context, chatID, msgID string) error
}

// Media is implemented by sessions that can download what a media ref
// (an avatar, an attachment, its preview) names. A ref is opaque outside
// the provider; it is valid only for the account that produced it.
type Media interface {
	// Media returns the content and its MIME type (may be empty). limit
	// caps the size: a larger file returns ErrTooLarge.
	Media(ctx context.Context, ref string, limit int64) (data []byte, mime string, err error)
}

// ErrTooLarge means a media file is bigger than the caller allowed.
var ErrTooLarge = errors.New("file is too large")

// Threads is implemented by sessions with discussions under messages
// (Slack threads, Telegram channel comments).
type Threads interface {
	// Thread returns up to limit replies of a thread before beforeID (the
	// latest when empty), newest first, like History; more is false when
	// the first reply was reached. The root message is not included.
	// Replies carry InThread = threadID.
	Thread(ctx context.Context, chatID, threadID, beforeID string, limit int) (msgs []model.Message, more bool, err error)
}

// Topics is implemented by sessions with forum chats (Chat.Forum: a
// Telegram group with topics). A topic's messages are the thread with the
// topic's id: read with Threads.Thread, sent with Outgoing.Thread, and live
// ones come as EvThreadMessage too; they also stay in the chat's history.
// A session implementing Topics implements Threads.
type Topics interface {
	// Topics returns the topics of a forum chat, pinned first, then by
	// latest activity.
	Topics(ctx context.Context, chatID string) ([]model.Topic, error)
	// ReadTopic marks a topic read up to msgID (the latest when empty).
	ReadTopic(ctx context.Context, chatID, topicID, msgID string) error
}

// Reactor is implemented by sessions that can react to messages.
type Reactor interface {
	// React adds (on) or takes back (!on) the account's reaction emoji
	// (as model.Reaction.Emoji names it) and returns the message's
	// reactions after the change.
	React(ctx context.Context, chatID, msgID, emoji string, on bool) ([]model.Reaction, error)
}

// Forwarder is implemented by sessions that can forward messages between
// chats of the account natively: attachments go along and the messenger
// marks the copies as forwarded. Forwarding between accounts is not a
// provider's business (the app sends the text).
type Forwarder interface {
	// Forward forwards msgIDs (oldest first) of fromChat to toChat and
	// returns what appeared in toChat, oldest first. The result may be
	// empty when the messenger does not echo the copies (they come as
	// events then).
	Forward(ctx context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error)
}

// LogOuter is implemented by sessions that can end the session on the
// server when the account is removed.
type LogOuter interface {
	LogOut(ctx context.Context) error
}

// Registry holds the known providers by type.
type Registry struct {
	mu sync.RWMutex
	m  map[string]Provider
}

// NewRegistry returns a registry with the given providers.
func NewRegistry(ps ...Provider) *Registry {
	r := &Registry{m: map[string]Provider{}}
	for _, p := range ps {
		r.Register(p)
	}
	return r
}

// Register adds p, replacing a provider of the same type.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	r.m[p.Info().Type] = p
	r.mu.Unlock()
}

// Get returns the provider of a type.
func (r *Registry) Get(typ string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.m[typ]
	return p, ok
}

// All returns the providers sorted by type.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	out := make([]Provider, 0, len(r.m))
	for _, p := range r.m {
		out = append(out, p)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Info().Type < out[j].Info().Type })
	return out
}

// MemKV is a KV in memory (login scratch state, tests).
type MemKV struct {
	mu sync.Mutex
	m  map[string][]byte
}

func NewMemKV() *MemKV { return &MemKV{m: map[string][]byte{}} }

func (k *MemKV) Get(key string) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if v, ok := k.m[key]; ok {
		return append([]byte(nil), v...), nil
	}
	return nil, nil
}

func (k *MemKV) Put(key string, v []byte) error {
	k.mu.Lock()
	k.m[key] = append([]byte(nil), v...)
	k.mu.Unlock()
	return nil
}

func (k *MemKV) Delete(key string) error {
	k.mu.Lock()
	delete(k.m, key)
	k.mu.Unlock()
	return nil
}

// Snapshot returns a copy of all entries.
func (k *MemKV) Snapshot() map[string][]byte {
	k.mu.Lock()
	defer k.mu.Unlock()
	out := make(map[string][]byte, len(k.m))
	for key, v := range k.m {
		out[key] = append([]byte(nil), v...)
	}
	return out
}

// Caller is implemented by sessions with 1:1 audio and video calls. Media
// runs in the app as a standard WebRTC peer connection; the session carries
// its offer, answer and candidates over the messenger's own signalling and
// reports everything with EvCall and EvCallSignal events.
type Caller interface {
	// StartCall calls the peer of a direct chat.
	StartCall(ctx context.Context, chatID string, video bool) (model.Call, error)
	// AcceptCall answers an incoming call.
	AcceptCall(ctx context.Context, callID string, video bool) error
	// EndCall hangs up, cancels or declines (reason EndHangup or
	// EndDeclined), or turns a call down because the device is in another
	// one (EndBusy).
	EndCall(ctx context.Context, callID string, reason model.CallEnd) error
	// CallSignal delivers what the app's peer connection produced.
	CallSignal(ctx context.Context, callID string, sig model.CallSignal) error
	// SetCallMedia tells the peer whether this side is muted, sends video
	// or shares the screen.
	SetCallMedia(ctx context.Context, callID string, m model.CallMedia) error
}
