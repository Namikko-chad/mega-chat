// Package hub runs the accounts: one session per enabled account, kept
// connected with backoff; what sessions report goes into the store and on
// to subscribers (the app). Reads are served from the store; history the
// store does not have is fetched from the provider and kept.
package hub

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"megachat/core/crypto"
	"megachat/core/eventlog"
	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/store"
	"megachat/core/vault"
)

// Status of an account's session.
type Status int

const (
	StatusDisabled Status = iota + 1
	StatusConnecting
	StatusOnline
	StatusOffline
	StatusNeedsLogin
)

var (
	ErrNoAccount  = errors.New("no such account")
	ErrNoLogin    = errors.New("sign-in expired, start again")
	ErrNoProvider = errors.New("unknown messenger")
)

// AccountInfo is an account with its live state.
type AccountInfo struct {
	vault.Account
	Status    Status
	StatusMsg string
	Self      model.User
	Unread    int
	Caps      []provider.Capability
	// MailUnread: unread messages in the inbox (mail accounts).
	MailUnread int
}

// EventKind says what an Event is about.
type EventKind int

const (
	EvAccount EventKind = iota + 1 // Account is set; nil: removed (AccountID)
	EvChat
	EvChatRemoved
	EvMessageNew
	EvMessageEdit
	EvMessageDelete
	EvTyping
	EvThreadMessage
	EvCall       // Call changed (AccountID)
	EvCallSignal // Signal for the peer connection of Call (ids only)
)

// Event goes to subscribers.
type Event struct {
	Kind      EventKind
	AccountID string
	Account   *AccountInfo
	Chat      *model.Chat
	Message   *model.Message
	Call      *model.Call
	Signal    *model.CallSignal
	Mail      *model.Mail // EvMailNew
	// Notify: EvMessageNew or EvThreadMessage the user should be told
	// about (see notifies); the app adds whether it is on screen.
	Notify bool
}

// Hub is the running core of an unlocked vault.
type Hub struct {
	v    *vault.Vault
	st   *store.Store
	reg  *provider.Registry
	log  *eventlog.Log
	http *http.Client

	// MaxBackoff caps the reconnect delay.
	MaxBackoff time.Duration
	// LoginWait caps how long one LoginStep waits for the provider.
	LoginWait time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.Mutex
	runs   map[string]*run
	logins map[string]*pendingLogin

	subMu  sync.Mutex
	subs   map[int]func(Event)
	nextID int

	callMu sync.Mutex
	calls  map[callKey]*liveCall

	mail mailCache // folder and calendar lists (mail.go)

	// notified: thread messages already notified about (a thread
	// message comes again when edited); cleared when it grows large.
	notified map[string]struct{}
}

type run struct {
	acct   vault.Account
	prov   provider.Provider
	sess   provider.Session
	cancel context.CancelFunc
	done   chan struct{}

	refreshing atomic.Bool

	// guarded by Hub.mu
	status Status
	msg    string
	self   model.User
	// fresh: chats whose latest messages were fetched since the session
	// last connected (see syncHead).
	fresh map[string]bool
	// online: direct chats whose peer is online (EvPresence); dropped
	// when the session goes down.
	online map[string]bool
}

// New returns a hub; Start runs the accounts.
func New(v *vault.Vault, st *store.Store, reg *provider.Registry, log *eventlog.Log) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		v: v, st: st, reg: reg, log: log,
		http:       &http.Client{Timeout: 90 * time.Second},
		MaxBackoff: 5 * time.Minute, LoginWait: 25 * time.Second,
		ctx: ctx, cancel: cancel,
		runs: map[string]*run{}, logins: map[string]*pendingLogin{}, subs: map[int]func(Event){},
		calls: map[callKey]*liveCall{},
	}
}

// Start opens a session for every account.
func (h *Hub) Start() {
	h.dropBroken()
	for _, a := range h.v.Config().Accounts {
		h.startRun(a)
	}
}

// dropBroken removes cached messages without an id, which a bug in the
// Slack provider stored for every live message (all under the id "").
func (h *Hub) dropBroken() {
	for _, a := range h.v.Config().Accounts {
		chats, err := h.st.Chats(a.ID)
		if err != nil {
			h.logf(eventlog.Error, "store", a.ID, "%v", err)
			continue
		}
		for _, c := range chats {
			if _, err := h.st.DeleteMessage(a.ID, c.ID, ""); err != nil {
				h.logf(eventlog.Error, "store", a.ID, "%v", err)
			}
			if c.Last != nil && c.Last.ID == "" {
				_, err = h.st.UpdateChat(a.ID, c.ID, func(c *model.Chat, ok bool) bool {
					c.Last = nil
					return ok
				})
				if err != nil {
					h.logf(eventlog.Error, "store", a.ID, "%v", err)
				}
			}
		}
	}
}

// Close stops every session and login. The store and vault stay open.
func (h *Hub) Close() {
	h.mu.Lock()
	for id, p := range h.logins {
		p.login.Close()
		delete(h.logins, id)
	}
	h.mu.Unlock()
	h.cancel()
	h.wg.Wait()
}

// Subscribe calls fn for every event until cancel is called. fn must not
// block: it runs on the goroutine that produced the event.
func (h *Hub) Subscribe(fn func(Event)) (cancel func()) {
	h.subMu.Lock()
	id := h.nextID
	h.nextID++
	h.subs[id] = fn
	h.subMu.Unlock()
	return func() {
		h.subMu.Lock()
		delete(h.subs, id)
		h.subMu.Unlock()
	}
}

func (h *Hub) emit(ev Event) {
	if ev.Chat != nil {
		acct := ev.Chat.AccountID
		if acct == "" {
			acct = ev.AccountID
		}
		ev.Chat.Online = h.online(acct, ev.Chat.ID)
	}
	h.subMu.Lock()
	fns := make([]func(Event), 0, len(h.subs))
	for _, fn := range h.subs {
		fns = append(fns, fn)
	}
	h.subMu.Unlock()
	for _, fn := range fns {
		fn(ev)
	}
}

func (h *Hub) logf(level, source, acct, format string, args ...any) {
	h.log.Addf(level, source, acct, format, args...)
}

func (h *Hub) env(acctID, source string, kv provider.KV) provider.Env {
	return provider.Env{
		AccountID: acctID, State: kv, HTTP: h.http,
		Log: func(level, msg string) { h.log.Add(level, source, acctID, msg) },
	}
}

// --- sessions ------------------------------------------------------------

// startRun creates the run of an account and connects it unless disabled.
func (h *Hub) startRun(a vault.Account) {
	r := &run{acct: a, status: StatusDisabled, done: make(chan struct{})}
	h.mu.Lock()
	h.runs[a.ID] = r
	h.mu.Unlock()
	p, ok := h.reg.Get(a.Provider)
	if !ok {
		close(r.done)
		h.setStatus(r, StatusOffline, fmt.Sprintf("%v: %s", ErrNoProvider, a.Provider))
		return
	}
	r.prov = p
	sess, err := p.Open(provider.Account{ID: a.ID, Login: a.Login, Params: a.Params}, h.env(a.ID, a.Provider, h.st.KV(a.ID)))
	if err != nil {
		close(r.done)
		h.setStatus(r, StatusOffline, err.Error())
		return
	}
	r.sess = sess
	if a.Disabled {
		close(r.done)
		h.emitAccount(a.ID)
		return
	}
	ctx, cancel := context.WithCancel(h.ctx)
	r.cancel = cancel
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		defer close(r.done)
		h.loop(ctx, r)
	}()
}

// stopRun stops an account's session and waits for it.
func (h *Hub) stopRun(id string) {
	h.mu.Lock()
	r := h.runs[id]
	delete(h.runs, id)
	h.mu.Unlock()
	if r == nil {
		return
	}
	if r.cancel != nil {
		r.cancel()
	}
	<-r.done
	h.dropCalls(id)
	h.forgetMail(id)
	h.mu.Lock()
	gone := r.online
	r.online = nil
	h.mu.Unlock()
	for chat := range gone { // no longer known: shown offline
		h.emitChat(id, chat)
	}
}

func (h *Hub) loop(ctx context.Context, r *run) {
	backoff := time.Second
	for {
		h.setStatus(r, StatusConnecting, "")
		started := time.Now()
		err := r.sess.Run(ctx, &sink{h: h, r: r})
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, provider.ErrAuth) {
			h.logf(eventlog.Warn, r.acct.Provider, r.acct.ID, "sign-in required: %v", err)
			h.setStatus(r, StatusNeedsLogin, err.Error())
			return
		}
		msg := "disconnected"
		if err != nil {
			msg = err.Error()
		}
		h.logf(eventlog.Warn, r.acct.Provider, r.acct.ID, "%s", msg)
		h.setStatus(r, StatusOffline, msg)
		if time.Since(started) > 2*time.Minute {
			backoff = time.Second // it was up for a while: not a crash loop
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, h.MaxBackoff)
	}
}

func (h *Hub) setStatus(r *run, s Status, msg string) {
	h.mu.Lock()
	changed := r.status != s || r.msg != msg
	r.status, r.msg = s, msg
	var gone map[string]bool
	if s != StatusOnline {
		gone, r.online = r.online, nil // presence is not known while down
	}
	h.mu.Unlock()
	if changed {
		h.emitAccount(r.acct.ID)
	}
	for chat := range gone {
		h.emitChat(r.acct.ID, chat)
	}
}

// online tells whether the peer of a direct chat is online.
func (h *Hub) online(acct, chat string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.runs[acct]
	return r != nil && r.online[chat]
}

// applyPresence records whether the peer of a direct chat is online and
// tells subscribers when it changed.
func (h *Hub) applyPresence(acct, chat string, on bool) {
	h.mu.Lock()
	r := h.runs[acct]
	changed := r != nil && r.online[chat] != on
	if changed {
		if on {
			if r.online == nil {
				r.online = map[string]bool{}
			}
			r.online[chat] = true
		} else {
			delete(r.online, chat)
		}
	}
	h.mu.Unlock()
	if changed {
		h.emitChat(acct, chat)
	}
}

// emitChat tells subscribers about a stored chat (nothing when it is not).
func (h *Hub) emitChat(acct, chat string) {
	if c, ok, _ := h.st.Chat(acct, chat); ok {
		h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
	}
}

func (h *Hub) emitAccount(id string) {
	if info, ok := h.Account(id); ok {
		h.emit(Event{Kind: EvAccount, AccountID: id, Account: &info})
	}
}

type sink struct {
	h *Hub
	r *run
}

func (s *sink) Emit(ev model.Event) { s.h.apply(s.r.acct.ID, ev) }

func (s *sink) Connected(self model.User) {
	s.h.mu.Lock()
	s.r.self = self
	s.r.fresh = map[string]bool{} // events may have been missed while down
	s.h.mu.Unlock()
	s.h.setStatus(s.r, StatusOnline, "")
	s.h.logf(eventlog.Info, s.r.acct.Provider, s.r.acct.ID, "connected as %s", self.Name)
	go s.h.refreshRun(s.r)
}

func (s *sink) SaveParams(params map[string]string) error {
	id := s.r.acct.ID
	return s.h.v.Update(func(c *vault.Config) error {
		for i := range c.Accounts {
			if c.Accounts[i].ID == id {
				c.Accounts[i].Params = params
				return nil
			}
		}
		return ErrNoAccount
	})
}

// messenger returns the chat capability of an account's session.
func (h *Hub) messenger(acct string) (provider.Messenger, *run, error) {
	h.mu.Lock()
	r := h.runs[acct]
	h.mu.Unlock()
	if r == nil {
		return nil, nil, ErrNoAccount
	}
	if r.sess == nil || r.acct.Disabled {
		return nil, r, provider.ErrOffline
	}
	m, ok := r.sess.(provider.Messenger)
	if !ok {
		return nil, r, provider.ErrUnsupported
	}
	return m, r, nil
}

// refreshRun reloads the chat list of a connected account (one at a time).
func (h *Hub) refreshRun(r *run) {
	if !r.refreshing.CompareAndSwap(false, true) {
		return
	}
	defer r.refreshing.Store(false)
	// Folders (mail unread counts) load beside the chat list, not before it.
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.refreshMail(r.acct.ID)
	}()
	if err := h.refresh(h.ctx, r.acct.ID); err != nil && h.ctx.Err() == nil {
		h.logf(eventlog.Warn, r.acct.Provider, r.acct.ID, "chat list: %v", err)
	}
}

func (h *Hub) refresh(ctx context.Context, acct string) error {
	m, _, err := h.messenger(acct)
	if errors.Is(err, provider.ErrUnsupported) {
		return nil // a mail or calendar account: no chats
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	chats, err := m.Chats(ctx, 300)
	if err != nil {
		return err
	}
	for i := range chats {
		h.apply(acct, model.Event{Kind: model.EvChat, Chat: &chats[i]})
	}
	return nil
}

// --- events --------------------------------------------------------------

// apply stores what a session reported and tells subscribers.
func (h *Hub) apply(acct string, ev model.Event) {
	if ev.Message != nil {
		ev.Message.AccountID = acct
	}
	if ev.Chat != nil {
		ev.Chat.AccountID = acct
	}
	var err error
	switch ev.Kind {
	case model.EvMessageNew, model.EvMessageEdit:
		if ev.Message == nil {
			return
		}
		if ev.Message.ID == "" || ev.Message.ChatID == "" {
			h.logf(eventlog.Warn, "store", acct, "dropped a message without an id (chat %q)", ev.Message.ChatID)
			return
		}
		err = h.applyMessage(acct, ev)
	case model.EvMessageDelete:
		if ev.Message == nil {
			return
		}
		err = h.applyDelete(acct, *ev.Message)
	case model.EvChat:
		if ev.Chat == nil {
			return
		}
		var c model.Chat
		if c, err = h.mergeChat(*ev.Chat, true); err == nil {
			h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
		}
	case model.EvChatRemoved:
		if ev.Chat == nil {
			return
		}
		if err = h.st.DeleteChat(acct, ev.Chat.ID); err == nil {
			h.emit(Event{Kind: EvChatRemoved, AccountID: acct, Chat: &model.Chat{AccountID: acct, ID: ev.Chat.ID}})
		}
	case model.EvRead:
		if ev.Chat == nil {
			return
		}
		unread := ev.Chat.Unread
		var c model.Chat
		c, err = h.st.UpdateChat(acct, ev.Chat.ID, func(c *model.Chat, ok bool) bool {
			if !ok {
				return false
			}
			if unread >= 0 { // -1: unknown, the stored count stays
				c.Unread = unread
			}
			return true
		})
		if err == nil && c.Title != "" {
			h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
			h.emitAccount(acct)
		}
	case model.EvTyping:
		if ev.Message != nil {
			h.emit(Event{Kind: EvTyping, AccountID: acct, Message: ev.Message})
		}
	case model.EvReactions:
		if ev.Message != nil {
			err = h.applyReactions(acct, *ev.Message)
		}
	case model.EvCall:
		if ev.Call != nil {
			h.applyCall(acct, *ev.Call)
		}
	case model.EvCallSignal:
		if ev.Call != nil && ev.Signal != nil {
			h.applySignal(acct, ev.Call.ID, *ev.Signal)
		}
	case model.EvMailNew, model.EvMailChanged, model.EvCalendarChanged:
		h.applyMail(acct, ev)
	case model.EvPresence:
		if ev.Chat != nil && ev.Chat.ID != "" {
			h.applyPresence(acct, ev.Chat.ID, ev.Online)
		}
	case model.EvThreadMessage:
		if ev.Message != nil && ev.Message.InThread != "" {
			ev.Message.AccountID = acct
			out := Event{Kind: EvThreadMessage, AccountID: acct, Message: ev.Message}
			// The chat (title, muted) for a notification about a mention.
			if c, ok, _ := h.st.Chat(acct, ev.Message.ChatID); ok {
				out.Chat = &c
			}
			out.Notify = h.notifyThread(*ev.Message)
			h.emit(out)
		}
	}
	if err != nil {
		h.logf(eventlog.Error, "store", acct, "%v", err)
	}
}

// notifies tells whether a message new in chat c deserves a notification:
// someone else wrote it, and the chat is not muted or the message
// mentions the user.
func notifies(m model.Message, c model.Chat) bool {
	return !m.Outgoing && !m.Service && !m.Deleted && (!c.Muted || m.Mentioned)
}

// notifyThread: replies in threads notify only when they mention the user,
// once per message.
func (h *Hub) notifyThread(m model.Message) bool {
	if m.Outgoing || m.Deleted || !m.Mentioned {
		return false
	}
	key := m.AccountID + "\x00" + m.ChatID + "\x00" + m.ID
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.notified[key]; ok {
		return false
	}
	if h.notified == nil || len(h.notified) >= 1000 {
		h.notified = map[string]struct{}{}
	}
	h.notified[key] = struct{}{}
	return true
}

func sameMessage(a, b model.Message) bool {
	return a.Text == b.Text && a.Edited.Equal(b.Edited) && a.Deleted == b.Deleted &&
		len(a.Attachments) == len(b.Attachments) && sameReactions(a.Reactions, b.Reactions) && sameThread(a.Thread, b.Thread)
}

func sameReactions(a, b []model.Reaction) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameThread(a, b *model.Thread) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID && a.Replies == b.Replies && a.Last.Equal(b.Last) && len(a.Users) == len(b.Users)
}

// applyReactions replaces the reactions of a stored message. A message
// that is not cached concerns only an open chat that shows it.
func (h *Hub) applyReactions(acct string, in model.Message) error {
	m, ok, err := h.st.UpdateMessage(acct, in.ChatID, in.ID, func(m *model.Message) { m.Reactions = in.Reactions })
	if err != nil || !ok {
		return err
	}
	h.emit(Event{Kind: EvMessageEdit, AccountID: acct, Message: &m})
	return nil
}

func (h *Hub) applyMessage(acct string, ev model.Event) error {
	m := *ev.Message
	if ev.Chat != nil {
		// The chat beside a message is what the session knows offhand; it
		// does not tell muted or pinned (Telegram), so those stay.
		if _, err := h.mergeChat(*ev.Chat, false); err != nil {
			return err
		}
	}
	if ev.Kind == model.EvMessageEdit {
		if _, ok, err := h.st.Message(acct, m.ChatID, m.ID); err != nil || !ok {
			// Not cached: nothing to correct but the open chat, if any.
			if err == nil {
				h.emit(Event{Kind: EvMessageEdit, AccountID: acct, Message: &m})
			}
			return err
		}
	}
	prev, err := h.st.PutMessage(m)
	if err != nil {
		return err
	}
	if prev != nil {
		// A repeat (sent here, then echoed by the server) or an edit.
		if sameMessage(*prev, m) {
			return nil
		}
		h.emit(Event{Kind: EvMessageEdit, AccountID: acct, Message: &m})
		c, err := h.st.UpdateChat(acct, m.ChatID, func(c *model.Chat, ok bool) bool {
			if !ok || c.Last == nil || c.Last.ID != m.ID {
				return false
			}
			c.Last = &m
			return true
		})
		if err == nil && c.Last != nil && c.Last.ID == m.ID {
			h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
		}
		return err
	}
	if ev.Kind == model.EvMessageEdit {
		return nil
	}
	known := true
	c, err := h.st.UpdateChat(acct, m.ChatID, func(c *model.Chat, ok bool) bool {
		known = ok
		if c.Last == nil || !m.Time.Before(c.Last.Time) {
			c.Last = &m
		}
		if m.Time.After(c.Updated) {
			c.Updated = m.Time
		}
		if !m.Outgoing {
			c.Unread++
		}
		if !ok && c.Title == "" {
			c.Kind, c.Title = model.ChatDirect, m.Sender.Name // until the chat list says better
		}
		return true
	})
	if err != nil {
		return err
	}
	h.emit(Event{Kind: EvMessageNew, AccountID: acct, Message: &m, Chat: &c, Notify: notifies(m, c)})
	h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
	if !m.Outgoing {
		h.emitAccount(acct)
	}
	if !known && ev.Chat == nil {
		h.mu.Lock()
		r := h.runs[acct]
		h.mu.Unlock()
		if r != nil {
			go h.refreshRun(r)
		}
	}
	return nil
}

func (h *Hub) applyDelete(acct string, m model.Message) error {
	_, _, err := h.st.UpdateMessage(acct, m.ChatID, m.ID, func(s *model.Message) {
		s.Deleted, s.Text, s.Attachments = true, "", nil
	})
	if err != nil {
		return err
	}
	del := model.Message{AccountID: acct, ChatID: m.ChatID, ID: m.ID, Deleted: true}
	h.emit(Event{Kind: EvMessageDelete, AccountID: acct, Message: &del})
	c, err := h.st.UpdateChat(acct, m.ChatID, func(c *model.Chat, ok bool) bool {
		if !ok || c.Last == nil || c.Last.ID != m.ID {
			return false
		}
		c.Last.Deleted, c.Last.Text, c.Last.Attachments = true, "", nil
		return true
	})
	if err == nil && c.Last != nil && c.Last.ID == m.ID {
		h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
	}
	return err
}

// mergeChat folds what a provider knows about a chat into the stored one.
// flags: in carries Muted and Pinned (a chat-list entry or EvChat); a chat
// attached to a message event does not, and the stored ones stay.
func (h *Hub) mergeChat(in model.Chat, flags bool) (model.Chat, error) {
	if in.Last != nil {
		in.Last.AccountID, in.Last.ChatID = in.AccountID, in.ID
		if _, err := h.st.PutMessage(*in.Last); err != nil {
			return model.Chat{}, err
		}
	}
	return h.st.UpdateChat(in.AccountID, in.ID, func(c *model.Chat, ok bool) bool {
		if in.Kind != model.ChatUnknown {
			// Only a provider that knows the chat tells its kind, and
			// with it whether it is a forum.
			c.Kind, c.Forum = in.Kind, in.Forum
		}
		if in.Title != "" {
			c.Title = in.Title
		}
		if in.Avatar != "" {
			c.Avatar = in.Avatar
		}
		if in.Unread >= 0 {
			c.Unread = in.Unread
		}
		if flags {
			c.Muted, c.Pinned = in.Muted, in.Pinned
		}
		if in.Access != model.AccessUnknown {
			c.Access = in.Access
		}
		if in.Last != nil && (c.Last == nil || !in.Last.Time.Before(c.Last.Time)) {
			c.Last = in.Last
		}
		if in.Updated.After(c.Updated) {
			c.Updated = in.Updated
		}
		if c.Last != nil && c.Last.Time.After(c.Updated) {
			c.Updated = c.Last.Time
		}
		return true
	})
}

// --- reads and writes ----------------------------------------------------

// Accounts returns all accounts in the order they were added.
func (h *Hub) Accounts() []AccountInfo {
	var out []AccountInfo
	for _, a := range h.v.Config().Accounts {
		if info, ok := h.Account(a.ID); ok {
			out = append(out, info)
		}
	}
	return out
}

// Account returns one account with its state.
func (h *Hub) Account(id string) (AccountInfo, bool) {
	a, ok := h.v.Config().Account(id)
	if !ok {
		return AccountInfo{}, false
	}
	info := AccountInfo{Account: a, Status: StatusDisabled}
	h.mu.Lock()
	if r := h.runs[id]; r != nil {
		info.Status, info.StatusMsg, info.Self = r.status, r.msg, r.self
		if r.prov != nil {
			info.Caps = append([]provider.Capability(nil), r.prov.Info().Caps...)
		}
		if r.sess != nil {
			if _, ok := r.sess.(provider.Threads); ok {
				info.Caps = append(info.Caps, provider.CapThreads)
			}
			if _, ok := r.sess.(provider.Reactor); ok {
				info.Caps = append(info.Caps, provider.CapReactions)
			}
			if _, ok := r.sess.(provider.Media); ok {
				info.Caps = append(info.Caps, provider.CapMedia)
			}
			if _, ok := r.sess.(provider.Caller); ok {
				info.Caps = append(info.Caps, provider.CapCalls)
			}
			if _, ok := r.sess.(provider.Topics); ok {
				info.Caps = append(info.Caps, provider.CapTopics)
			}
			if _, ok := r.sess.(provider.Forwarder); ok {
				info.Caps = append(info.Caps, provider.CapForward)
			}
			if _, ok := r.sess.(provider.Mailer); ok && !slices.Contains(info.Caps, provider.CapMail) {
				info.Caps = append(info.Caps, provider.CapMail)
			}
			if _, ok := r.sess.(provider.Calendars); ok && !slices.Contains(info.Caps, provider.CapCalendar) {
				info.Caps = append(info.Caps, provider.CapCalendar)
			}
		}
	}
	h.mu.Unlock()
	if a.Disabled {
		info.Status = StatusDisabled
	}
	chats, _ := h.st.Chats(id)
	for _, c := range chats {
		if !c.Muted {
			info.Unread += c.Unread
		}
	}
	info.MailUnread = h.mailUnread(id)
	return info, true
}

// Online counts accounts that are connected.
func (h *Hub) Online() (online, total int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.runs {
		total++
		if r.status == StatusOnline {
			online++
		}
	}
	return online, total
}

// Chats returns the cached chats of the given accounts (all when empty),
// latest activity first; limit 0 means all.
func (h *Hub) Chats(accounts []string, limit int) ([]model.Chat, error) {
	if len(accounts) == 0 {
		for _, a := range h.v.Config().Accounts {
			accounts = append(accounts, a.ID)
		}
	}
	var out []model.Chat
	for _, id := range accounts {
		cs, err := h.st.Chats(id)
		if err != nil {
			return nil, err
		}
		out = append(out, cs...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	h.mu.Lock()
	for i := range out {
		if r := h.runs[out[i].AccountID]; r != nil {
			out[i].Online = r.online[out[i].ID]
		}
	}
	h.mu.Unlock()
	return out, nil
}

// RefreshChats reloads the chat lists from the providers.
func (h *Hub) RefreshChats(ctx context.Context, accounts []string) error {
	if len(accounts) == 0 {
		for _, a := range h.v.Config().Accounts {
			if !a.Disabled {
				accounts = append(accounts, a.ID)
			}
		}
	}
	var errs []error
	for _, id := range accounts {
		if err := h.refresh(ctx, id); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

// History returns up to limit messages before beforeID (latest when
// empty), newest first. Messages the cache lacks come from the provider;
// offline, the cached part is returned.
func (h *Hub) History(ctx context.Context, acct, chat, beforeID string, limit int) ([]model.Message, bool, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if beforeID == "" {
		h.syncHead(ctx, acct, chat, limit)
	}
	cached, end, err := h.st.History(acct, chat, beforeID, limit)
	if err != nil {
		return nil, false, err
	}
	switch end {
	case store.EndLimit:
		return cached, true, nil
	case store.EndFirst:
		return cached, false, nil
	}
	m, _, err := h.messenger(acct)
	if err != nil {
		if len(cached) > 0 && !errors.Is(err, ErrNoAccount) {
			return cached, true, nil
		}
		return nil, false, err
	}
	from := beforeID
	if len(cached) > 0 {
		from = cached[len(cached)-1].ID
	}
	page, more, err := m.History(ctx, chat, from, limit-len(cached))
	if err != nil {
		if len(cached) > 0 {
			h.logf(eventlog.Warn, "history", acct, "%v", err)
			return cached, true, nil
		}
		return nil, false, err
	}
	for i := range page {
		page[i].AccountID, page[i].ChatID = acct, chat
	}
	if err := h.st.PutHistory(acct, chat, from, page, more); err != nil {
		return nil, false, err
	}
	return append(cached, page...), more, nil
}

// CachedHistory returns the latest cached messages of a chat, newest
// first, without asking the messenger: possibly with gaps, to show at once
// while History brings the real page.
func (h *Hub) CachedHistory(acct, chat string, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return h.st.Recent(acct, chat, limit)
}

// syncHead fetches the latest messages of a chat once per connection. The
// cache cannot tell that its newest message is not the chat's newest:
// events missed while the session was down (or never delivered) leave a
// gap above it that nothing else would fill.
func (h *Hub) syncHead(ctx context.Context, acct, chat string, limit int) {
	m, r, err := h.messenger(acct)
	if err != nil {
		return
	}
	h.mu.Lock()
	skip := r.status != StatusOnline || r.fresh == nil || r.fresh[chat]
	h.mu.Unlock()
	if skip {
		return
	}
	page, more, err := m.History(ctx, chat, "", limit)
	if err != nil {
		h.logf(eventlog.Warn, "history", acct, "%v", err)
		return
	}
	for i := range page {
		page[i].AccountID, page[i].ChatID = acct, chat
	}
	if err := h.st.PutHistory(acct, chat, "", page, more); err != nil {
		h.logf(eventlog.Error, "store", acct, "%v", err)
		return
	}
	h.mu.Lock()
	if r.fresh != nil {
		r.fresh[chat] = true
	}
	h.mu.Unlock()
	if len(page) == 0 {
		return
	}
	last, changed := page[0], false
	c, err := h.st.UpdateChat(acct, chat, func(c *model.Chat, ok bool) bool {
		if !ok || (c.Last != nil && !last.Time.After(c.Last.Time)) {
			return false
		}
		c.Last, changed = &last, true
		if last.Time.After(c.Updated) {
			c.Updated = last.Time
		}
		return true
	})
	if err != nil {
		h.logf(eventlog.Error, "store", acct, "%v", err)
	} else if changed {
		h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
	}
}

// Send sends a text message and records it.
func (h *Hub) Send(ctx context.Context, acct, chat string, out model.Outgoing) (model.Message, error) {
	m, _, err := h.messenger(acct)
	if err != nil {
		return model.Message{}, err
	}
	if out.Thread != "" {
		if _, ok := m.(provider.Threads); !ok {
			return model.Message{}, provider.ErrUnsupported
		}
	}
	msg, err := m.Send(ctx, chat, out)
	if err != nil {
		return model.Message{}, err
	}
	msg.AccountID, msg.ChatID, msg.Outgoing = acct, chat, true
	if out.Thread != "" && msg.InThread == "" {
		msg.InThread = out.Thread
	}
	if msg.InThread != "" {
		// A reply inside a thread is not part of the chat's history.
		h.apply(acct, model.Event{Kind: model.EvThreadMessage, Message: &msg})
		return msg, nil
	}
	h.apply(acct, model.Event{Kind: model.EvMessageNew, Message: &msg})
	return msg, nil
}

// Forward forwards messages of a chat to another chat of the same account
// and records the copies.
func (h *Hub) Forward(ctx context.Context, acct, from string, ids []string, to string) ([]model.Message, error) {
	if len(ids) == 0 {
		return nil, errors.New("no messages to forward")
	}
	m, _, err := h.messenger(acct)
	if err != nil {
		return nil, err
	}
	f, ok := m.(provider.Forwarder)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	msgs, err := f.Forward(ctx, from, ids, to)
	if err != nil {
		return nil, err
	}
	for i := range msgs {
		msgs[i].AccountID, msgs[i].ChatID, msgs[i].Outgoing = acct, to, true
		h.apply(acct, model.Event{Kind: model.EvMessageNew, Message: &msgs[i]})
	}
	return msgs, nil
}

// MaxMedia is the default size limit of Media.
const MaxMedia = 32 << 20

// mediaCacheMax: larger files are fetched each time instead of cached.
const mediaCacheMax = 8 << 20

// Media returns the content of a media ref of an account, from the cache
// or the provider.
func (h *Hub) Media(ctx context.Context, acct, ref string, limit int64) ([]byte, string, error) {
	if ref == "" {
		return nil, "", errors.New("empty media ref")
	}
	if limit <= 0 {
		limit = MaxMedia
	}
	if data, mime, ok, err := h.st.Media(acct, ref); err != nil {
		h.logf(eventlog.Error, "store", acct, "media cache: %v", err)
	} else if ok {
		if int64(len(data)) > limit {
			return nil, "", provider.ErrTooLarge
		}
		return data, mime, nil
	}
	h.mu.Lock()
	r := h.runs[acct]
	h.mu.Unlock()
	if r == nil {
		return nil, "", ErrNoAccount
	}
	if r.sess == nil {
		return nil, "", provider.ErrOffline
	}
	md, ok := r.sess.(provider.Media)
	if !ok {
		return nil, "", provider.ErrUnsupported
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	data, mime, err := md.Media(ctx, ref, limit)
	if err != nil {
		return nil, "", err
	}
	if len(data) <= mediaCacheMax {
		if err := h.st.PutMedia(acct, ref, data, mime); err != nil {
			h.logf(eventlog.Error, "store", acct, "media cache: %v", err)
		}
	}
	return data, mime, nil
}

// Thread returns replies of a thread, newest first (not cached).
func (h *Hub) Thread(ctx context.Context, acct, chat, thread, beforeID string, limit int) ([]model.Message, bool, error) {
	m, _, err := h.messenger(acct)
	if err != nil {
		return nil, false, err
	}
	t, ok := m.(provider.Threads)
	if !ok {
		return nil, false, provider.ErrUnsupported
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	msgs, more, err := t.Thread(ctx, chat, thread, beforeID, limit)
	for i := range msgs {
		msgs[i].AccountID, msgs[i].ChatID, msgs[i].InThread = acct, chat, thread
	}
	return msgs, more, err
}

// Topics returns the topics of a forum chat (not cached).
func (h *Hub) Topics(ctx context.Context, acct, chat string) ([]model.Topic, error) {
	m, _, err := h.messenger(acct)
	if err != nil {
		return nil, err
	}
	t, ok := m.(provider.Topics)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	topics, err := t.Topics(ctx, chat)
	for i := range topics {
		if l := topics[i].Last; l != nil {
			l.AccountID, l.ChatID, l.InThread = acct, chat, topics[i].ID
		}
	}
	return topics, err
}

// ReadTopic marks a topic of a forum chat read up to msgID (the latest
// when empty). The chat's unread count follows with the provider's event.
func (h *Hub) ReadTopic(ctx context.Context, acct, chat, topic, msgID string) error {
	m, _, err := h.messenger(acct)
	if err != nil {
		return err
	}
	t, ok := m.(provider.Topics)
	if !ok {
		return provider.ErrUnsupported
	}
	return t.ReadTopic(ctx, chat, topic, msgID)
}

// React adds or takes back the account's reaction to a message.
func (h *Hub) React(ctx context.Context, acct, chat, msgID, emoji string, on bool) (model.Message, error) {
	m, _, err := h.messenger(acct)
	if err != nil {
		return model.Message{}, err
	}
	r, ok := m.(provider.Reactor)
	if !ok {
		return model.Message{}, provider.ErrUnsupported
	}
	rs, err := r.React(ctx, chat, msgID, emoji, on)
	if err != nil {
		return model.Message{}, err
	}
	if err := h.applyReactions(acct, model.Message{ChatID: chat, ID: msgID, Reactions: rs}); err != nil {
		return model.Message{}, err
	}
	msg, ok, err := h.st.Message(acct, chat, msgID)
	if err != nil {
		return model.Message{}, err
	}
	if !ok {
		msg = model.Message{AccountID: acct, ChatID: chat, ID: msgID}
	}
	msg.Reactions = rs
	return msg, nil
}

// MarkRead marks a chat read up to msgID (latest when empty).
func (h *Hub) MarkRead(ctx context.Context, acct, chat, msgID string) error {
	m, _, err := h.messenger(acct)
	if err != nil {
		return err
	}
	if err := m.MarkRead(ctx, chat, msgID); err != nil {
		return err
	}
	h.apply(acct, model.Event{Kind: model.EvRead, Chat: &model.Chat{ID: chat, Unread: h.unreadAfter(acct, chat, msgID)}})
	return nil
}

// unreadAfter is the unread count of a chat read up to msgID: the incoming
// messages after it, as far as the cache knows them, never more than the
// chat's count before (0 when the whole chat was read).
func (h *Hub) unreadAfter(acct, chat, msgID string) int {
	if msgID == "" {
		return 0
	}
	c, ok, err := h.st.Chat(acct, chat)
	if err != nil || !ok || c.Unread == 0 {
		return 0
	}
	if c.Last != nil && c.Last.ID == msgID {
		return 0
	}
	n, found, err := h.st.CountAfter(acct, chat, msgID, c.Unread+200, func(m model.Message) bool {
		return !m.Outgoing && !m.Deleted
	})
	if err != nil || !found {
		return -1 // keep the stored count
	}
	return min(n, c.Unread)
}

// --- account management --------------------------------------------------

// Rename changes the name shown for an account.
func (h *Hub) Rename(id, name string) (AccountInfo, error) {
	if name == "" {
		return AccountInfo{}, errors.New("empty name")
	}
	if err := h.updateAccount(id, func(a *vault.Account) { a.Name = name }); err != nil {
		return AccountInfo{}, err
	}
	h.mu.Lock()
	if r := h.runs[id]; r != nil {
		r.acct.Name = name
	}
	h.mu.Unlock()
	h.emitAccount(id)
	info, _ := h.Account(id)
	return info, nil
}

// SetEnabled connects or disconnects an account.
func (h *Hub) SetEnabled(id string, enabled bool) (AccountInfo, error) {
	if err := h.updateAccount(id, func(a *vault.Account) { a.Disabled = !enabled }); err != nil {
		return AccountInfo{}, err
	}
	h.stopRun(id)
	a, _ := h.v.Config().Account(id)
	h.startRun(a)
	h.emitAccount(id)
	info, _ := h.Account(id)
	return info, nil
}

// Remove signs an account out (best effort) and forgets it.
func (h *Hub) Remove(ctx context.Context, id string) error {
	if _, ok := h.v.Config().Account(id); !ok {
		return ErrNoAccount
	}
	h.mu.Lock()
	r := h.runs[id]
	h.mu.Unlock()
	h.stopRun(id)
	if r != nil && r.sess != nil {
		if lo, ok := r.sess.(provider.LogOuter); ok {
			lctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			if err := lo.LogOut(lctx); err != nil {
				h.logf(eventlog.Warn, r.acct.Provider, id, "sign out: %v", err)
			}
			cancel()
		}
	}
	if err := h.v.Update(func(c *vault.Config) error {
		for i, a := range c.Accounts {
			if a.ID == id {
				c.Accounts = append(c.Accounts[:i], c.Accounts[i+1:]...)
				return nil
			}
		}
		return ErrNoAccount
	}); err != nil {
		return err
	}
	if err := h.st.DeleteAccount(id); err != nil {
		return err
	}
	h.emit(Event{Kind: EvAccount, AccountID: id})
	return nil
}

func (h *Hub) updateAccount(id string, fn func(a *vault.Account)) error {
	return h.v.Update(func(c *vault.Config) error {
		for i := range c.Accounts {
			if c.Accounts[i].ID == id {
				fn(&c.Accounts[i])
				return nil
			}
		}
		return ErrNoAccount
	})
}

// --- sign-in -------------------------------------------------------------

type pendingLogin struct {
	mu      sync.Mutex // one Next at a time
	prov    provider.Provider
	method  string
	login   provider.Login
	relogin string // account id for a re-login
	touched time.Time
}

// LoginResult is a login step as the hub returns it.
type LoginResult struct {
	ID      string
	Step    provider.Step
	Account *AccountInfo // set when done
}

// StartLogin begins signing in a new account.
func (h *Hub) StartLogin(ctx context.Context, providerType, method string) (LoginResult, error) {
	p, ok := h.reg.Get(providerType)
	if !ok {
		return LoginResult{}, ErrNoProvider
	}
	return h.beginLogin(ctx, p, method, nil)
}

// Relogin signs an existing account in again (same identity).
func (h *Hub) Relogin(ctx context.Context, id string) (LoginResult, error) {
	a, ok := h.v.Config().Account(id)
	if !ok {
		return LoginResult{}, ErrNoAccount
	}
	p, ok := h.reg.Get(a.Provider)
	if !ok {
		return LoginResult{}, ErrNoProvider
	}
	return h.beginLogin(ctx, p, a.Method, &a)
}

func (h *Hub) beginLogin(ctx context.Context, p provider.Provider, method string, a *vault.Account) (LoginResult, error) {
	h.expireLogins()
	if method == "" && len(p.Info().Methods) > 0 {
		method = p.Info().Methods[0].ID
	}
	var pa *provider.Account
	relogin := ""
	if a != nil {
		pa = &provider.Account{ID: a.ID, Login: a.Login, Params: a.Params}
		relogin = a.ID
	}
	l, err := p.Login(ctx, method, pa, h.env("", p.Info().Type, provider.NewMemKV()))
	if err != nil {
		return LoginResult{}, err
	}
	pl := &pendingLogin{prov: p, method: method, login: l, relogin: relogin, touched: time.Now()}
	id := crypto.RandomID(12)
	h.mu.Lock()
	h.logins[id] = pl
	h.mu.Unlock()
	return h.next(ctx, id, pl, nil)
}

// LoginStep answers the current step of a login.
func (h *Hub) LoginStep(ctx context.Context, id string, values map[string]string) (LoginResult, error) {
	h.mu.Lock()
	pl := h.logins[id]
	h.mu.Unlock()
	if pl == nil {
		return LoginResult{}, ErrNoLogin
	}
	return h.next(ctx, id, pl, values)
}

// CancelLogin drops a login.
func (h *Hub) CancelLogin(id string) {
	h.mu.Lock()
	pl := h.logins[id]
	delete(h.logins, id)
	h.mu.Unlock()
	if pl != nil {
		pl.login.Close()
	}
}

func (h *Hub) next(ctx context.Context, id string, pl *pendingLogin, values map[string]string) (LoginResult, error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.touched = time.Now()
	wctx, cancel := context.WithTimeout(ctx, h.LoginWait)
	step, err := pl.login.Next(wctx, values)
	cancel()
	if err != nil {
		h.CancelLogin(id)
		return LoginResult{}, err
	}
	res := LoginResult{ID: id, Step: step}
	if !step.Done {
		return res, nil
	}
	h.CancelLogin(id)
	if step.Result == nil {
		return LoginResult{}, errors.New("login finished without credentials")
	}
	info, err := h.finishLogin(pl, *step.Result)
	if err != nil {
		return LoginResult{}, err
	}
	res.Account = &info
	return res, nil
}

func (h *Hub) finishLogin(pl *pendingLogin, cr provider.Credentials) (AccountInfo, error) {
	typ := pl.prov.Info().Type
	id := pl.relogin
	if id != "" {
		a, ok := h.v.Config().Account(id)
		if !ok {
			return AccountInfo{}, ErrNoAccount
		}
		if a.Login != "" && cr.Login != a.Login {
			return AccountInfo{}, fmt.Errorf("signed in as %s, but this account is %s", cr.Login, a.Login)
		}
		h.stopRun(id)
		if err := h.updateAccount(id, func(a *vault.Account) {
			a.Params, a.Method, a.Login = cr.Params, pl.method, cr.Login
		}); err != nil {
			return AccountInfo{}, err
		}
	} else {
		id = crypto.RandomID(8)
		name := cr.Name
		if name == "" {
			name = pl.prov.Info().Name
		}
		err := h.v.Update(func(c *vault.Config) error {
			if a, ok := c.SameLogin(typ, cr.Login); ok {
				return fmt.Errorf("%w as %q", vault.ErrSameLogin, a.Name)
			}
			c.Accounts = append(c.Accounts, vault.Account{
				ID: id, Provider: typ, Name: name, Login: cr.Login, Method: pl.method,
				Params: cr.Params, Added: time.Now().UTC(),
			})
			return nil
		})
		if err != nil {
			return AccountInfo{}, err
		}
	}
	kv := h.st.KV(id)
	for k, v := range cr.State {
		if err := kv.Put(k, v); err != nil {
			return AccountInfo{}, err
		}
	}
	a, _ := h.v.Config().Account(id)
	h.startRun(a)
	h.logf(eventlog.Info, typ, id, "signed in as %s", cr.Login)
	info, _ := h.Account(id)
	return info, nil
}

func (h *Hub) expireLogins() {
	h.mu.Lock()
	var old []*pendingLogin
	for id, pl := range h.logins {
		if pl.mu.TryLock() {
			if time.Since(pl.touched) > 30*time.Minute {
				old = append(old, pl)
				delete(h.logins, id)
			}
			pl.mu.Unlock()
		}
	}
	h.mu.Unlock()
	for _, pl := range old {
		pl.login.Close()
	}
}
