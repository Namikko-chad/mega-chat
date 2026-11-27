package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

func folderSession(t *testing.T, method string, h http.HandlerFunc) *session {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	params := map[string]string{"method": method, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{
		State: provider.NewMemKV(), HTTP: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return sess.(*session)
}

func TestChatFoldersFollowSidebar(t *testing.T) {
	s := folderSession(t, methodSession, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "channel_sections": []map[string]any{
			{"channel_section_id": "S2", "name": "Team", "type": "standard", "emoji": "rocket", "channel_ids_page": map[string]any{"channel_ids": []string{"C2"}, "count": 1}},
			{"channel_section_id": "S0", "name": "", "type": "channels", "next_channel_section_id": "S1", "channel_ids_page": map[string]any{"channel_ids": []string{"C9"}}},
			{"channel_section_id": "S1", "name": "", "type": "stars", "next_channel_section_id": "S2", "channel_ids_page": map[string]any{"channel_ids": []string{"C1"}, "count": 1}},
			{"channel_section_id": "S3", "name": "Empty", "type": "standard"},
		}})
	})
	got, err := s.ChatFolders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "S1" || got[0].Title != "Starred" || got[1].ID != "S2" ||
		got[1].Emoji != "🚀" || len(got[1].Always) != 1 || got[1].Always[0] != "C2" {
		t.Fatalf("folders: %+v", got)
	}
}

func TestChatFoldersRefused(t *testing.T) {
	s := folderSession(t, methodSession, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "unknown_method"})
	})
	if got, err := s.ChatFolders(context.Background()); err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
	o := folderSession(t, methodOAuth, func(w http.ResponseWriter, r *http.Request) {
		t.Error("OAuth tokens cannot call this method")
	})
	if got, err := o.ChatFolders(context.Background()); err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
}

type countSink struct {
	mu sync.Mutex
	n  int
}

func (c *countSink) Emit(ev model.Event) {
	if ev.Kind == model.EvChatFolders {
		c.mu.Lock()
		c.n++
		c.mu.Unlock()
	}
}
func (c *countSink) Connected(model.User)               {}
func (c *countSink) SaveParams(map[string]string) error { return nil }

func TestSectionEventsGather(t *testing.T) {
	old := foldersDelay
	foldersDelay = 20 * time.Millisecond
	defer func() { foldersDelay = old }()
	s := folderSession(t, methodSession, func(http.ResponseWriter, *http.Request) {})
	c := &countSink{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for range 3 {
		s.handle(ctx, c, []byte(`{"type":"channel_section_upserted"}`))
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		n := c.n
		c.mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.n != 1 {
		t.Fatalf("%d reloads", c.n)
	}
}

func TestReplyQuoteReadAndKeptOnEdit(t *testing.T) {
	const raw = "&gt; old words\nmy answer"
	var mu sync.Mutex
	var updated url.Values
	s := folderSession(t, methodSession, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		switch r.URL.Path {
		case "/api/chat.postMessage":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "ts": "5.0", "message": map[string]any{"ts": "5.0", "user": "U0", "text": r.Form.Get("text")}})
		case "/api/conversations.history":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "messages": []map[string]any{
				{"ts": "5.0", "user": "U0", "subtype": "thread_broadcast", "thread_ts": "1.0", "text": raw},
			}})
		case "/api/chat.update":
			mu.Lock()
			updated = r.PostForm
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "ts": "5.0"})
		case "/api/users.info":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": map[string]any{"id": "U0", "name": "me"}})
		default:
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "unknown_method"})
		}
	})
	ctx := context.Background()
	m, err := s.Send(ctx, "C1", model.Outgoing{Text: "my answer", ReplyTo: "1.0", Quote: "old words"})
	if err != nil {
		t.Fatal(err)
	}
	if m.Quote != "old words" || m.Text != "my answer" || m.ReplyTo != "1.0" {
		t.Fatalf("sent: %+v", m)
	}
	if _, err := s.Edit(ctx, "C1", "5.0", "better answer"); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if got := updated.Get("text"); got != "&gt; old words\nbetter answer" {
		t.Fatalf("update: %q", got)
	}
	if lines, rest := splitQuote("> only a quote"); lines != nil || rest != "> only a quote" {
		t.Fatalf("a bare quote: %v %q", lines, rest)
	}
}

type eventSink struct {
	mu  sync.Mutex
	evs []model.Event
}

func (c *eventSink) Emit(ev model.Event) {
	c.mu.Lock()
	c.evs = append(c.evs, ev)
	c.mu.Unlock()
}
func (c *eventSink) Connected(model.User)               {}
func (c *eventSink) SaveParams(map[string]string) error { return nil }

func TestThreadReplyFollowed(t *testing.T) {
	roots := map[string]map[string]any{
		"1.0": {"ts": "1.0", "user": "U1", "text": "root", "reply_users": []string{"U2", "U0"}}, // replied in it
		"2.0": {"ts": "2.0", "user": "U1", "text": "root", "reply_users": []string{"U2"}},       // not in it
		"3.0": {"ts": "3.0", "user": "U1", "text": "root", "subscribed": true},                  // followed by hand
		"4.0": {"ts": "4.0", "user": "U0", "text": "root", "subscribed": false},                 // own, unfollowed
	}
	s := folderSession(t, methodSession, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		switch r.URL.Path {
		case "/api/conversations.replies":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "messages": []map[string]any{roots[r.Form.Get("ts")]}})
		case "/api/users.info":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": map[string]any{"id": r.Form.Get("user"), "name": "x"}})
		default:
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "unknown_method"})
		}
	})
	for root, want := range map[string]bool{"1.0": true, "2.0": false, "3.0": true, "4.0": false} {
		sink := &eventSink{}
		s.handle(context.Background(), sink, []byte(`{"type":"message","channel":"C1","user":"U2","ts":"9.0","thread_ts":"`+root+`","text":"reply"}`))
		if len(sink.evs) != 2 || sink.evs[0].Kind != model.EvThreadMessage || sink.evs[1].Kind != model.EvMessageEdit {
			t.Fatalf("%s: %+v", root, sink.evs)
		}
		if got := sink.evs[0].Message.Followed; got != want {
			t.Fatalf("%s: followed %v", root, got)
		}
	}
}
