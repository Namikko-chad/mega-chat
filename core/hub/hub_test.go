package hub

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"megachat/core/eventlog"
	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/providers/fake"
	"megachat/core/store"
	"megachat/core/vault"
)

type env struct {
	h   *Hub
	srv *fake.Server
	st  *store.Store

	mu     sync.Mutex
	events []Event
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := t.TempDir()
	c, _ := vault.New()
	v, err := vault.Create(filepath.Join(dir, "v"), []byte("pw"), c, vault.KDFParams{Time: 1, Memory: 1024, Threads: 1})
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "cache.db"), c.Key())
	if err != nil {
		t.Fatal(err)
	}
	srv := fake.New()
	e := &env{srv: srv, st: st}
	e.h = New(v, st, provider.NewRegistry(srv), eventlog.New(100, nil))
	e.h.MaxBackoff = 50 * time.Millisecond
	e.h.Subscribe(func(ev Event) {
		e.mu.Lock()
		e.events = append(e.events, ev)
		e.mu.Unlock()
	})
	e.h.Start()
	t.Cleanup(func() { e.h.Close(); st.Close() })
	return e
}

func (e *env) login(t *testing.T, name string) AccountInfo {
	t.Helper()
	ctx := context.Background()
	res, err := e.h.StartLogin(ctx, fake.Type, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Step.Fields) != 1 || res.Step.Done {
		t.Fatalf("first step: %+v", res.Step)
	}
	res2, err := e.h.LoginStep(ctx, res.ID, map[string]string{"name": ""})
	if err != nil || res2.Step.Error == "" {
		t.Fatalf("empty name must be rejected in the step: %+v %v", res2.Step, err)
	}
	res3, err := e.h.LoginStep(ctx, res.ID, map[string]string{"name": name})
	if err != nil {
		t.Fatal(err)
	}
	if !res3.Step.Done || res3.Account == nil {
		t.Fatalf("not done: %+v", res3)
	}
	return *res3.Account
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (e *env) status(id string) Status {
	a, _ := e.h.Account(id)
	return a.Status
}

func TestLoginSessionAndEvents(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.login(t, "alice")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	waitFor(t, "chat list", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })

	// Same login again is refused.
	res, _ := e.h.StartLogin(ctx, fake.Type, "")
	if _, err := e.h.LoginStep(ctx, res.ID, map[string]string{"name": "alice"}); !errors.Is(err, vault.ErrSameLogin) {
		t.Fatalf("duplicate: %v", err)
	}
	// A second account of the same provider is fine.
	b := e.login(t, "bob")
	waitFor(t, "both lists", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 20 })
	if cs, _ := e.h.Chats([]string{b.ID}, 0); len(cs) != 10 || cs[0].AccountID != b.ID {
		t.Fatalf("bob's chats: %d", len(cs))
	}

	// Incoming message: stored, unread counted, subscribers told.
	in := e.srv.Inject("alice", "c1", "hi alice")
	waitFor(t, "unread", func() bool { info, _ := e.h.Account(a.ID); return info.Unread == 1 })
	cs, _ := e.h.Chats([]string{a.ID}, 1)
	if cs[0].ID != "c1" || cs[0].Last == nil || cs[0].Last.Text != "hi alice" {
		t.Fatalf("top chat: %+v", cs[0])
	}
	e.mu.Lock()
	var gotNew bool
	for _, ev := range e.events {
		gotNew = gotNew || (ev.Kind == EvMessageNew && ev.Message.ID == in.ID && ev.AccountID == a.ID)
	}
	e.mu.Unlock()
	if !gotNew {
		t.Fatal("no new-message event")
	}

	// History: the cache has only the live message; the rest comes from
	// the provider, then from the cache.
	msgs, more, err := e.h.History(ctx, a.ID, "c1", "", 10)
	if err != nil || len(msgs) != 10 || !more || msgs[0].ID != in.ID {
		t.Fatalf("history: %d %v %v", len(msgs), more, err)
	}
	all := e.srv.Messages("alice", "c1")
	for i, m := range msgs {
		if want := all[len(all)-1-i].ID; m.ID != want {
			t.Fatalf("history[%d] = %s, want %s", i, m.ID, want)
		}
	}
	rest, more, err := e.h.History(ctx, a.ID, "c1", msgs[9].ID, 100)
	if err != nil || more || len(rest) != len(all)-10 {
		t.Fatalf("rest: %d %v %v", len(rest), more, err)
	}
	if again, _, _ := e.h.History(ctx, a.ID, "c1", "", 31); len(again) != 31 {
		t.Fatalf("cached again: %d", len(again))
	}

	// Send records an outgoing message; read clears unread.
	sent, err := e.h.Send(ctx, a.ID, "c1", model.Outgoing{Text: "hello"})
	if err != nil || !sent.Outgoing {
		t.Fatal(sent, err)
	}
	if err := e.h.MarkRead(ctx, a.ID, "c1", ""); err != nil {
		t.Fatal(err)
	}
	if info, _ := e.h.Account(a.ID); info.Unread != 0 {
		t.Fatalf("unread after read: %d", info.Unread)
	}
	if top, _, _ := e.h.History(ctx, a.ID, "c1", "", 1); top[0].ID != sent.ID {
		t.Fatal("sent message is not the latest")
	}

	// Disable, enable, remove.
	if info, _ := e.h.SetEnabled(b.ID, false); info.Status != StatusDisabled {
		t.Fatalf("disabled: %v", info.Status)
	}
	if _, err := e.h.Send(ctx, b.ID, "c1", model.Outgoing{Text: "x"}); !errors.Is(err, provider.ErrOffline) {
		t.Fatalf("send while disabled: %v", err)
	}
	e.h.SetEnabled(b.ID, true)
	waitFor(t, "bob online", func() bool { return e.status(b.ID) == StatusOnline })
	if err := e.h.Remove(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if cs, _ := e.h.Chats(nil, 0); len(cs) != 10 || len(e.h.Accounts()) != 1 {
		t.Fatalf("after remove: %d chats, %d accounts", len(cs), len(e.h.Accounts()))
	}
}

func TestReorderAccounts(t *testing.T) {
	e := newEnv(t)
	a, b, c := e.login(t, "a").ID, e.login(t, "b").ID, e.login(t, "c").ID
	if err := e.h.Reorder([]string{c, "unknown", a}); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, x := range e.h.Accounts() {
		got = append(got, x.ID)
	}
	if want := []string{c, a, b}; !slices.Equal(got, want) {
		t.Fatalf("order %v, want %v", got, want)
	}
}

func TestNeedsLoginAndRelogin(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	e.srv.FailAuth["carol"] = true
	a := e.login(t, "carol")
	waitFor(t, "needs login", func() bool { return e.status(a.ID) == StatusNeedsLogin })

	e.srv.FailAuth["carol"] = false
	res, err := e.h.Relogin(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.h.LoginStep(ctx, res.ID, map[string]string{"name": "dave"}); err == nil {
		t.Fatal("re-login as someone else must fail")
	}
	res, _ = e.h.Relogin(ctx, a.ID)
	done, err := e.h.LoginStep(ctx, res.ID, map[string]string{"name": "carol"})
	if err != nil || done.Account.ID != a.ID {
		t.Fatalf("relogin: %+v %v", done, err)
	}
	waitFor(t, "online again", func() bool { return e.status(a.ID) == StatusOnline })
	if len(e.h.Accounts()) != 1 {
		t.Fatal("relogin must not add an account")
	}
}

func TestOfflineHistoryFromCache(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.login(t, "erin")
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	if _, _, err := e.h.History(ctx, a.ID, "c2", "", 10); err != nil {
		t.Fatal(err)
	}
	e.h.SetEnabled(a.ID, false)
	msgs, more, err := e.h.History(ctx, a.ID, "c2", "", 20)
	if err != nil || len(msgs) != 10 || !more {
		t.Fatalf("offline: %d %v %v", len(msgs), more, err)
	}
}

func TestMissedEventsFilledAfterReconnect(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.login(t, "frank")
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	if _, _, err := e.h.History(ctx, a.ID, "c1", "", 10); err != nil {
		t.Fatal(err)
	}
	missed := e.srv.Miss("frank", "c1", "lost while down")
	e.h.SetEnabled(a.ID, false)
	e.h.SetEnabled(a.ID, true)
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	msgs, _, err := e.h.History(ctx, a.ID, "c1", "", 10)
	if err != nil || len(msgs) != 10 || msgs[0].ID != missed.ID {
		t.Fatalf("after reconnect: %d %v %+v", len(msgs), err, msgs)
	}
	all := e.srv.Messages("frank", "c1")
	for i, m := range msgs {
		if want := all[len(all)-1-i].ID; m.ID != want {
			t.Fatalf("history[%d] = %s, want %s", i, m.ID, want)
		}
	}
	if cs, _ := e.h.Chats([]string{a.ID}, 1); cs[0].Last == nil || cs[0].Last.ID != missed.ID {
		t.Fatalf("chat's last message: %+v", cs[0].Last)
	}
}

func TestBrokenMessagesDropped(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.login(t, "gina")
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	if _, _, err := e.h.History(ctx, a.ID, "c2", "", 10); err != nil {
		t.Fatal(err)
	}
	// What the Slack bug left: a message without an id.
	e.h.apply(a.ID, model.Event{Kind: model.EvMessageNew, Message: &model.Message{ChatID: "c2", Text: "x"}})
	if _, ok, _ := e.st.Message(a.ID, "c2", ""); ok {
		t.Fatal("a message without an id was stored")
	}
	if _, err := e.st.PutMessage(model.Message{AccountID: a.ID, ChatID: "c2", Text: "x"}); err != nil {
		t.Fatal(err)
	}
	e.h.dropBroken()
	if _, ok, _ := e.st.Message(a.ID, "c2", ""); ok {
		t.Fatal("broken message kept")
	}
	if msgs, _, err := e.h.History(ctx, a.ID, "c2", "", 10); err != nil || len(msgs) != 10 {
		t.Fatalf("history after cleanup: %d %v", len(msgs), err)
	}
}

func TestMediaThreadsReactions(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "ann")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	has := map[provider.Capability]bool{}
	for _, c := range info.Caps {
		has[c] = true
	}
	if !has[provider.CapMedia] || !has[provider.CapThreads] || !has[provider.CapReactions] {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()

	// Media comes from the provider once, then from the cache.
	data, mime, err := e.h.Media(ctx, a.ID, "avatar:c1", 0)
	if err != nil || mime != "image/png" || len(data) == 0 {
		t.Fatalf("media: %d %q %v", len(data), mime, err)
	}
	if _, _, ok, _ := e.st.Media(a.ID, "avatar:c1"); !ok {
		t.Fatal("media not cached")
	}
	if _, _, err := e.h.Media(ctx, a.ID, "avatar:c1", 10); !errors.Is(err, provider.ErrTooLarge) {
		t.Fatalf("limit: %v", err)
	}

	// Channel "News 5" is read-only.
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	cs, _ := e.h.Chats(nil, 0)
	for _, c := range cs {
		if (c.Kind == model.ChatChannel) != (c.Access == model.AccessRead) {
			t.Fatalf("access of %s: %v", c.Title, c.Access)
		}
	}

	// A thread in the group "Team 3".
	msgs, _, err := e.h.History(ctx, a.ID, "c3", "", 50)
	if err != nil {
		t.Fatal(err)
	}
	var root *model.Message
	for i := range msgs {
		if msgs[i].Thread != nil {
			root = &msgs[i]
			break
		}
	}
	if root == nil {
		t.Fatal("no thread in the history")
	}
	replies, more, err := e.h.Thread(ctx, a.ID, "c3", root.Thread.ID, "", 2)
	if err != nil || len(replies) != 2 || !more || replies[0].InThread != root.ID || replies[0].Time.Before(replies[1].Time) {
		t.Fatalf("thread: %+v %v %v", replies, more, err)
	}
	sent, err := e.h.Send(ctx, a.ID, "c3", model.Outgoing{Text: "in thread", Thread: root.ID})
	if err != nil || sent.InThread != root.ID {
		t.Fatalf("send in thread: %+v %v", sent, err)
	}
	waitFor(t, "thread event", func() bool { return e.has(EvThreadMessage, sent.ID) })
	waitFor(t, "root edit", func() bool {
		m, _, _ := e.st.Message(a.ID, "c3", root.ID)
		return m.Thread != nil && m.Thread.Replies == 4
	})
	if m, _, _ := e.st.Message(a.ID, "c3", sent.ID); m.ID != "" {
		t.Fatal("a thread reply went into the chat's history")
	}

	// Reactions.
	m, err := e.h.React(ctx, a.ID, "c3", root.ID, "🎉", true)
	if err != nil || len(m.Reactions) == 0 || m.Reactions[len(m.Reactions)-1] != (model.Reaction{Emoji: "🎉", Count: 1, Mine: true}) {
		t.Fatalf("react: %+v %v", m.Reactions, err)
	}
	if st, _, _ := e.st.Message(a.ID, "c3", root.ID); len(st.Reactions) != len(m.Reactions) {
		t.Fatal("reaction not stored")
	}
	e.h.apply(a.ID, model.Event{Kind: model.EvReactions, Message: &model.Message{ChatID: "c3", ID: root.ID}})
	if st, _, _ := e.st.Message(a.ID, "c3", root.ID); len(st.Reactions) != 0 {
		t.Fatal("reaction event not applied")
	}
}

func (e *env) has(kind EventKind, msgID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ev := range e.events {
		if ev.Kind == kind && ev.Message != nil && ev.Message.ID == msgID {
			return true
		}
	}
	return false
}

func TestForumTopics(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "fay")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapTopics) {
		t.Fatalf("caps: %v", info.Caps)
	}
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	cs, _ := e.h.Chats(nil, 0)
	forum := ""
	for _, c := range cs {
		if c.Forum {
			forum = c.ID
		}
	}
	if forum != "c8" {
		t.Fatalf("forum: %q", forum)
	}
	ctx := context.Background()
	topics, err := e.h.Topics(ctx, a.ID, forum)
	if err != nil || len(topics) != 4 || !topics[0].Pinned || topics[0].Title != "General" {
		t.Fatalf("topics: %+v %v", topics, err)
	}
	for _, tp := range topics {
		if tp.Last == nil || tp.Last.AccountID != a.ID || tp.Last.InThread != tp.ID {
			t.Fatalf("last of %s: %+v", tp.Title, tp.Last)
		}
	}
	design := topics[1].ID
	msgs, _, err := e.h.Thread(ctx, a.ID, forum, design, "", 50)
	if err != nil || len(msgs) == 0 {
		t.Fatalf("topic messages: %d %v", len(msgs), err)
	}
	for _, m := range msgs {
		if m.InThread != design || !strings.Contains(m.Text, topics[1].Title) {
			t.Fatalf("message of another topic: %+v", m)
		}
	}
	sent, err := e.h.Send(ctx, a.ID, forum, model.Outgoing{Text: "to the topic", Thread: design})
	if err != nil || sent.InThread != design {
		t.Fatalf("send to topic: %+v %v", sent, err)
	}
	waitFor(t, "topic event", func() bool { return e.has(EvThreadMessage, sent.ID) })
	// A topic's message is in the chat's history too.
	waitFor(t, "in history", func() bool { m, _, _ := e.st.Message(a.ID, forum, sent.ID); return m.ID == sent.ID })
	if err := e.h.ReadTopic(ctx, a.ID, forum, design, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := e.h.Topics(ctx, a.ID, "c1"); err == nil {
		t.Fatal("topics of a plain chat")
	}
}

func (e *env) calls(id string) []model.Call {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []model.Call
	for _, ev := range e.events {
		if ev.Kind == EvCall && ev.Call.ID == id {
			out = append(out, *ev.Call)
		}
	}
	return out
}

func TestCalls(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "ann")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapCalls) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	if _, err := e.h.StartCall(ctx, a.ID, "c3", false); !errors.Is(err, provider.ErrUnsupported) {
		t.Fatalf("group call: %v", err)
	}
	c, err := e.h.StartCall(ctx, a.ID, "c1", true)
	if err != nil || c.State != model.CallWaiting || !c.Video {
		t.Fatalf("StartCall: %+v %v", c, err)
	}
	if _, err := e.h.StartCall(ctx, a.ID, "c2", false); !errors.Is(err, ErrInCall) {
		t.Fatalf("second call: %v", err)
	}
	if got := e.h.Calls(); len(got) != 1 || got[0].Call.ID != c.ID {
		t.Fatalf("Calls: %+v", got)
	}

	// Signals are numbered per call and kept for an app that reconnects.
	for range 3 {
		e.h.applySignal(a.ID, c.ID, model.CallSignal{Kind: model.SigCandidate, SDP: "candidate:x"})
	}
	if s := e.h.CallSignals(a.ID, c.ID, 1); len(s) != 2 || s[0].Seq != 2 || s[1].Seq != 3 {
		t.Fatalf("CallSignals: %+v", s)
	}

	// The contact calls while this call is live: turned down as busy,
	// the app never hears of it.
	if _, err := e.h.Send(ctx, a.ID, "c2", model.Outgoing{Text: "/call"}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2500 * time.Millisecond)
	e.mu.Lock()
	for _, ev := range e.events {
		if ev.Kind == EvCall && ev.Call.ID != c.ID {
			e.mu.Unlock()
			t.Fatalf("busy call reached the app: %+v", ev.Call)
		}
	}
	e.mu.Unlock()

	if err := e.h.EndCall(ctx, a.ID, c.ID, model.EndHangup); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "ended", func() bool {
		cs := e.calls(c.ID)
		return len(cs) > 0 && cs[len(cs)-1].State == model.CallEnded
	})
	if len(e.h.Calls()) != 0 || e.h.CallSignals(a.ID, c.ID, 0) != nil {
		t.Fatal("call kept after it ended")
	}
}

// Reading up to a message leaves the newer incoming ones unread; the
// cached history comes without the network, gaps and all.
func TestPartialReadAndCachedHistory(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.login(t, "alice")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	// The chat list of the connect merges the provider's counts (0): it
	// must be done before the counts here are made.
	waitFor(t, "chat list", func() bool {
		cs, _ := e.h.Chats([]string{a.ID}, 0)
		e.h.mu.Lock()
		r := e.h.runs[a.ID]
		e.h.mu.Unlock()
		return len(cs) == 10 && r != nil && !r.refreshing.Load()
	})
	chatUnread := func() int {
		c, _, _ := e.h.st.Chat(a.ID, "c2")
		return c.Unread
	}
	m1 := e.srv.Inject("alice", "c2", "one")
	e.srv.Inject("alice", "c2", "two")
	m3 := e.srv.Inject("alice", "c2", "three")
	waitFor(t, "unread", func() bool { return chatUnread() == 3 })

	cached, err := e.h.CachedHistory(a.ID, "c2", 10)
	if err != nil || len(cached) < 3 || cached[0].ID != m3.ID || cached[2].ID != m1.ID {
		t.Fatalf("cached: %v %v", cached, err)
	}
	if err := e.h.MarkRead(ctx, a.ID, "c2", m1.ID); err != nil {
		t.Fatal(err)
	}
	if n := chatUnread(); n != 2 {
		t.Fatalf("unread after reading the first: %d", n)
	}
	// A message the cache does not know keeps the count.
	if err := e.h.MarkRead(ctx, a.ID, "c2", "nope"); err != nil {
		t.Fatal(err)
	}
	if n := chatUnread(); n != 2 {
		t.Fatalf("unread after an unknown id: %d", n)
	}
	if err := e.h.MarkRead(ctx, a.ID, "c2", m3.ID); err != nil {
		t.Fatal(err)
	}
	if n := chatUnread(); n != 0 {
		t.Fatalf("unread after reading the last: %d", n)
	}
}

// A chat attached to a live message (Telegram's chatInfo) does not tell
// muted or pinned: it must not unmute the chat, or the account's unread
// count jumps between the list's count and the one with muted chats.
func TestMessageEventKeepsMuted(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "hank")
	waitFor(t, "chats", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	e.h.apply(a.ID, model.Event{Kind: model.EvChat, Chat: &model.Chat{ID: "c2", Title: "two", Unread: 5, Muted: true, Pinned: true}})
	before, _ := e.h.Account(a.ID)
	m := model.Message{ChatID: "c2", ID: "live-1", Text: "hi", Time: time.Now(), Sender: model.User{ID: "u9", Name: "x"}}
	e.h.apply(a.ID, model.Event{Kind: model.EvMessageNew, Message: &m, Chat: &model.Chat{ID: "c2", Title: "two", Unread: -1}})
	c, ok, err := e.st.Chat(a.ID, "c2")
	if err != nil || !ok || !c.Muted || !c.Pinned || c.Unread != 6 {
		t.Fatalf("chat after a live message: %+v %v %v", c, ok, err)
	}
	if after, _ := e.h.Account(a.ID); after.Unread != before.Unread {
		t.Fatalf("account unread %d → %d from a muted chat", before.Unread, after.Unread)
	}
}

func TestNotify(t *testing.T) {
	in := model.Message{ID: "1", ChatID: "c"}
	mention := in
	mention.Mentioned = true
	out := in
	out.Outgoing = true
	svc := in
	svc.Service = true
	for _, tc := range []struct {
		m     model.Message
		muted bool
		want  bool
	}{
		{in, false, true},
		{in, true, false},
		{mention, true, true},
		{out, false, false},
		{svc, false, false},
	} {
		if got := notifies(tc.m, model.Chat{Muted: tc.muted}); got != tc.want {
			t.Errorf("notifies(%+v, muted %v) = %v", tc.m, tc.muted, got)
		}
	}

	e := newEnv(t)
	a := e.login(t, "alice")
	waitFor(t, "chat list", func() bool { cs, _ := e.h.Chats([]string{a.ID}, 0); return len(cs) == 10 })
	plain := e.srv.Inject("alice", "c1", "hello")
	at := e.srv.Inject("alice", "c1", "hey @alice")
	var gotPlain, gotAt *Event
	waitFor(t, "both messages", func() bool {
		e.mu.Lock()
		defer e.mu.Unlock()
		for i, ev := range e.events {
			if ev.Kind == EvMessageNew && ev.Message.ID == plain.ID {
				gotPlain = &e.events[i]
			}
			if ev.Kind == EvMessageNew && ev.Message.ID == at.ID {
				gotAt = &e.events[i]
			}
		}
		return gotPlain != nil && gotAt != nil
	})
	if !gotPlain.Notify || gotPlain.Message.Mentioned || !gotAt.Notify || !gotAt.Message.Mentioned {
		t.Fatalf("plain %+v / mention %+v", gotPlain, gotAt)
	}

	// A mention in a thread notifies once, not again when edited.
	th := model.Message{AccountID: a.ID, ChatID: "c1", ID: "t1", InThread: "1", Mentioned: true}
	if !e.h.notifyThread(th) || e.h.notifyThread(th) {
		t.Fatal("thread mention must notify exactly once")
	}
	th.ID, th.Mentioned = "t2", false
	if e.h.notifyThread(th) {
		t.Fatal("a thread reply without a mention must not notify")
	}
}

func TestForwardAndQuote(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "ann")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapForward) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	msgs, _, err := e.h.History(ctx, a.ID, "c3", "", 50)
	if err != nil {
		t.Fatal(err)
	}
	var src []model.Message
	for _, m := range msgs {
		if !m.Outgoing && !m.Service && m.Text != "" && len(src) < 2 {
			src = append(src, m)
		}
	}
	if len(src) != 2 {
		t.Fatalf("no messages to forward: %d", len(src))
	}
	ids := []string{src[1].ID, src[0].ID} // oldest first
	out, err := e.h.Forward(ctx, a.ID, "c3", ids, "c1")
	if err != nil || len(out) != 2 {
		t.Fatalf("forward: %+v %v", out, err)
	}
	for i, m := range out {
		want := src[1-i]
		if !m.Forwarded || m.ForwardedFrom != want.Sender.Name || m.Text != want.Text || !m.Outgoing || m.ChatID != "c1" || m.AccountID != a.ID {
			t.Fatalf("copy %d: %+v (of %+v)", i, m, want)
		}
		if st, ok, _ := e.st.Message(a.ID, "c1", m.ID); !ok || !st.Forwarded {
			t.Fatalf("copy %d not stored: %+v", i, st)
		}
	}
	if _, err := e.h.Forward(ctx, a.ID, "c3", nil, "c1"); err == nil {
		t.Fatal("forwarding nothing must fail")
	}

	// A reply quoting a fragment.
	sent, err := e.h.Send(ctx, a.ID, "c3", model.Outgoing{Text: "yes", ReplyTo: src[0].ID, Quote: "frag"})
	if err != nil || sent.Quote != "frag" || sent.Text != "yes" {
		t.Fatalf("quote: %+v %v", sent, err)
	}
}

func TestPresence(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "alice")
	waitFor(t, "chat list", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 10 })
	online := func(id string) bool {
		cs, _ := e.h.Chats([]string{a.ID}, 0)
		for _, c := range cs {
			if c.ID == id {
				return c.Online
			}
		}
		t.Fatalf("no chat %s", id)
		return false
	}
	// The demo has c1 online from the start; the chat list carries it.
	waitFor(t, "c1 online", func() bool { return online("c1") })
	if online("c2") {
		t.Fatal("c2 is not online")
	}
	told := func(id string, on bool) func() bool {
		return func() bool {
			e.mu.Lock()
			defer e.mu.Unlock()
			for _, ev := range e.events {
				if ev.Kind == EvChat && ev.Chat.ID == id && ev.Chat.Online == on {
					return true
				}
			}
			return false
		}
	}
	e.srv.SetOnline("alice", "c2", true)
	waitFor(t, "c2 online event", told("c2", true))
	if !online("c2") {
		t.Fatal("c2 online not in the chat list")
	}
	e.srv.SetOnline("alice", "c1", false)
	waitFor(t, "c1 offline event", told("c1", false))
	// A session that goes down forgets who is online.
	if _, err := e.h.SetEnabled(a.ID, false); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "c2 offline when disabled", func() bool { return !online("c2") })
}

func TestSendFiles(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "ann")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapFiles) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	var png bytes.Buffer
	if err := pngEncode(&png); err != nil {
		t.Fatal(err)
	}
	out := model.Outgoing{Text: "look", Files: []model.OutFile{
		{Name: `C:\Users\me\shot`, Data: png.Bytes()}, // a path and no type: base name, type from the content
		{Name: "notes.txt", Data: []byte("hello")},
	}}
	last, err := e.h.Send(ctx, a.ID, "c1", out)
	if err != nil || last.ID == "" || last.AccountID != a.ID || len(last.Attachments) != 1 {
		t.Fatalf("send: %+v %v", last, err)
	}
	if f := last.Attachments[0]; f.Name != "notes.txt" || !strings.HasPrefix(f.MIME, "text/plain") || f.Size != 5 {
		t.Fatalf("second file: %+v", f)
	}
	msgs, _, err := e.h.History(ctx, a.ID, "c1", "", 2)
	if err != nil || len(msgs) != 2 {
		t.Fatalf("history: %+v %v", msgs, err)
	}
	pic := msgs[1] // newest first
	if f := pic.Attachments[0]; pic.Text != "look" || f.Name != "shot.png" || f.MIME != "image/png" || f.Kind != model.AttachImage || f.Width != 4 || pic.Group == "" || pic.Group != msgs[0].Group {
		t.Fatalf("album: %+v / %+v", pic, msgs[0])
	}
	data, mime, err := e.h.Media(ctx, a.ID, pic.Attachments[0].Ref, 1<<20)
	if err != nil || mime != "image/png" || !bytes.Equal(data, png.Bytes()) {
		t.Fatalf("media: %q %v", mime, err)
	}
	if _, err := e.h.Send(ctx, a.ID, "c1", model.Outgoing{Files: []model.OutFile{{Name: "x"}}}); err == nil {
		t.Fatal("an empty file must fail")
	}
	big := model.Outgoing{Files: []model.OutFile{{Name: "big", Data: make([]byte, MaxUpload+1)}}}
	if _, err := e.h.Send(ctx, a.ID, "c1", big); !errors.Is(err, provider.ErrTooLarge) {
		t.Fatalf("too large: %v", err)
	}
}

func pngEncode(w io.Writer) error {
	return png.Encode(w, image.NewRGBA(image.Rect(0, 0, 4, 3)))
}

func TestProfile(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "pat")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapProfiles) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	// c1's contact is online by presence; c3 is a group.
	waitFor(t, "presence", func() bool { p, err := e.h.Profile(ctx, a.ID, "c1"); return err == nil && p.Online })
	p, err := e.h.Profile(ctx, a.ID, "c3")
	if err != nil || p.Title != "Team 3" || p.Members == 0 || len(p.MemberList) != 3 || p.Online {
		t.Fatalf("group: %+v %v", p, err)
	}
	if _, err := e.h.Profile(ctx, a.ID, "nope"); err == nil {
		t.Fatal("a missing chat has a profile")
	}
}

func TestEdit(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "eve")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapEdit) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	if _, _, err := e.h.History(ctx, a.ID, "c1", "", 10); err != nil {
		t.Fatal(err)
	}
	sent, err := e.h.Send(ctx, a.ID, "c1", model.Outgoing{Text: "helo"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.h.Edit(ctx, a.ID, "c1", "", sent.ID, "hello")
	if err != nil || got.Text != "hello" || got.Edited.IsZero() || got.Time.IsZero() {
		t.Fatalf("edit: %+v %v", got, err)
	}
	if !e.has(EvMessageEdit, sent.ID) {
		t.Fatal("no edit event")
	}
	top, _, _ := e.h.History(ctx, a.ID, "c1", "", 1)
	if top[0].ID != sent.ID || top[0].Text != "hello" || top[0].Edited.IsZero() {
		t.Fatalf("cached: %+v", top[0])
	}
	if c, _ := e.h.Chats(nil, 0); !slices.ContainsFunc(c, func(c model.Chat) bool { return c.ID == "c1" && c.Last != nil && c.Last.Text == "hello" }) {
		t.Fatal("the chat's last message was not edited")
	}
	// Someone else's message, an empty text.
	if _, err := e.h.Edit(ctx, a.ID, "c1", "", top[0].ID, ""); err == nil {
		t.Fatal("an empty text was accepted")
	}
	older, _, _ := e.h.History(ctx, a.ID, "c1", sent.ID, 1)
	if _, err := e.h.Edit(ctx, a.ID, "c1", "", older[0].ID, "mine now"); err == nil {
		t.Fatal("someone else's message was edited")
	}
}

func TestDelete(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "eve")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapDelete) || !slices.Contains(info.Caps, provider.CapDeleteForMe) {
		t.Fatalf("caps: %v", info.Caps)
	}
	ctx := context.Background()
	if _, _, err := e.h.History(ctx, a.ID, "c1", "", 10); err != nil {
		t.Fatal(err)
	}
	sent, err := e.h.Send(ctx, a.ID, "c1", model.Outgoing{Text: "oops"})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.h.Delete(ctx, a.ID, "c1", "", nil, true); err == nil {
		t.Fatal("nothing to delete was accepted")
	}
	if err := e.h.Delete(ctx, a.ID, "c1", "", []string{sent.ID}, false); err != nil {
		t.Fatal(err)
	}
	if !e.has(EvMessageDelete, sent.ID) {
		t.Fatal("no delete event")
	}
	top, _, _ := e.h.History(ctx, a.ID, "c1", "", 1)
	if top[0].ID != sent.ID || !top[0].Deleted || top[0].Text != "" {
		t.Fatalf("cached: %+v", top[0])
	}
	older, _, _ := e.h.History(ctx, a.ID, "c1", sent.ID, 1)
	if err := e.h.Delete(ctx, a.ID, "c1", "", []string{older[0].ID}, true); err == nil {
		t.Fatal("someone else's message was deleted")
	}
}

func TestMembers(t *testing.T) {
	e := newEnv(t)
	a := e.login(t, "eve")
	waitFor(t, "online", func() bool { return e.status(a.ID) == StatusOnline })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapMentions) {
		t.Fatalf("caps: %v", info.Caps)
	}
	chats, _ := e.h.Chats(nil, 0)
	var group string
	for _, c := range chats {
		if c.AccountID == a.ID && c.Kind == model.ChatGroup {
			group = c.ID
		}
	}
	us, err := e.h.Members(context.Background(), a.ID, group, "bo", 0)
	if err != nil || len(us) != 1 || us[0].Name != "Bob Stone" || us[0].Online {
		t.Fatalf("%+v %v", us, err)
	}
	us, err = e.h.Members(context.Background(), a.ID, group, "alice", 0)
	if err != nil || len(us) != 1 || !us[0].Online {
		t.Fatalf("online: %+v %v", us, err)
	}
}
