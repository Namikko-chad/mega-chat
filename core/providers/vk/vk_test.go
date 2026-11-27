package vk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

// fakeVK implements wire.go's view of the VK API.
type fakeVK struct {
	t    *testing.T
	host string

	mu           sync.Mutex
	goodTokens   map[string]bool // access tokens the API accepts
	checks       int             // auth.checkAuthCode calls
	anonCalls    int
	silentUsed   int // exchangeSilentAuthToken calls
	lpTS         int64
	lastSend     url.Values // the last messages.send form
	lastEdit     url.Values
	lastDelete   url.Values
	lastReaction url.Values
	manage       []string // "<method> <form>" of chat-management calls
	reacted      bool     // whether the self reaction is on
}

func (f *fakeVK) ok(w http.ResponseWriter, response any) {
	json.NewEncoder(w).Encode(map[string]any{"response": response})
}

func (f *fakeVK) fail(w http.ResponseWriter, code int, msg string) {
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"error_code": code, "error_msg": msg}})
}

func (f *fakeVK) authed(r *http.Request) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.goodTokens[r.Form.Get("access_token")]
}

func (f *fakeVK) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	switch r.URL.Path {
	case "/get_anonym_token":
		f.mu.Lock()
		f.anonCalls++
		tok := "ANON" + strconv.Itoa(f.anonCalls)
		f.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"token": tok, "expired_at": time.Now().Add(time.Hour).Unix()})
	case "/auth_by_exchange_token":
		if r.Form.Get("exchange_token") != "COMMON" {
			f.fail(w, 5, "bad exchange token")
			return
		}
		f.ok(w, map[string]any{"silent_token": "SILENT", "silent_token_uuid": "UUID"})
	case "/method/auth.getAuthCode":
		f.ok(w, map[string]any{"auth_url": "https://vk.com/qr?hash=x", "auth_hash": "HASH"})
	case "/method/auth.checkAuthCode":
		f.mu.Lock()
		f.checks++
		n := f.checks
		f.mu.Unlock()
		if n < 2 {
			f.ok(w, map[string]any{"status": qrOpened})
			return
		}
		f.ok(w, map[string]any{"status": qrApproved, "access_token": "ACCESS"})
	case "/method/auth.getExchangeToken":
		if r.Form.Get("access_token") != "ACCESS" {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"users_exchange_tokens": []any{
			map[string]any{"user_id": 42, "common_token": "COMMON", "profile_type": 0},
		}})
	case "/method/auth.exchangeSilentAuthToken":
		f.mu.Lock()
		f.silentUsed++
		n := f.silentUsed
		anon := r.Form.Get("access_token")
		f.mu.Unlock()
		if r.Form.Get("token") != "SILENT" {
			f.fail(w, 5, "bad silent token")
			return
		}
		if n == 1 { // pretend the first anonymous token had expired
			f.fail(w, 1114, "anonymous token expired")
			return
		}
		f.mu.Lock()
		f.goodTokens["ACCESS2"] = true
		f.mu.Unlock()
		_ = anon
		f.ok(w, map[string]any{"access_token": "ACCESS2"})
	case "/method/messages.getConversations":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{
			"count": 1,
			"items": []any{map[string]any{
				"conversation": map[string]any{
					"peer":         map[string]any{"id": 100, "type": "user"},
					"unread_count": 3,
					"can_write":    map[string]any{"allowed": true},
				},
				"last_message": map[string]any{"id": 7, "date": 1700000000, "peer_id": 100, "from_id": 100, "text": "hi", "out": 0},
			}},
			"profiles": []any{map[string]any{"id": 100, "first_name": "Anna", "last_name": "K", "photo_100": "https://vk.com/a.jpg"}},
			"groups":   []any{},
		})
	case "/method/messages.getLongPollServer":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"server": strings.TrimPrefix(f.host, "http://") + "/im", "key": "KEY", "ts": 10})
	case "/im":
		f.mu.Lock()
		f.lpTS++
		first := f.lpTS == 1
		f.mu.Unlock()
		if first {
			json.NewEncoder(w).Encode(map[string]any{"ts": 11, "updates": []any{
				[]any{4, 55, 0, 100, 1700000100, "ping"},
			}})
			return
		}
		// keep the long poll quiet afterwards
		time.Sleep(20 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]any{"ts": 11, "updates": []any{}})
	case "/method/messages.getById":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"items": []any{
			map[string]any{"id": 55, "conversation_message_id": 70, "date": 1700000100, "peer_id": 100, "from_id": 100, "text": "ping", "out": 0},
		}})
	case "/method/messages.send":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.lastSend = cloneForm(r.PostForm)
		f.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"response": 60})
	case "/method/messages.edit":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.lastEdit = cloneForm(r.PostForm)
		f.mu.Unlock()
		f.ok(w, 1)
	case "/method/messages.delete":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.lastDelete = cloneForm(r.PostForm)
		f.mu.Unlock()
		f.ok(w, map[string]any{})
	case "/method/messages.getConversationMembers":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{
			"count": 2,
			"items": []any{
				map[string]any{"member_id": 42, "is_admin": true},
				map[string]any{"member_id": 100},
			},
			"profiles": []any{
				map[string]any{"id": 100, "first_name": "Anna", "last_name": "K", "screen_name": "anna"},
			},
			"groups": []any{},
		})
	case "/method/messages.getConversationsById":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"items": []any{map[string]any{
			"peer":      map[string]any{"id": 2000000001, "type": "chat"},
			"can_write": map[string]any{"allowed": true},
			"chat_settings": map[string]any{
				"title": "Team", "members_count": 3,
				"photo": map[string]any{"photo_100": "https://vk.com/c.jpg"},
			},
		}}, "profiles": []any{}, "groups": []any{}})
	case "/method/users.get":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, []any{map[string]any{
			"id": 42, "first_name": "Ivan", "last_name": "Petrov", "screen_name": "ivan",
			"photo_200": "https://vk.com/i200.jpg", "status": "hello", "online": 1,
		}})
	case "/method/groups.getById":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"groups": []any{map[string]any{
			"id": 777, "name": "Cats", "screen_name": "cats",
			"description": "meow", "members_count": 9000, "photo_200": "https://vk.com/g.jpg",
		}}})
	case "/method/photos.getMessagesUploadServer":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, map[string]any{"upload_url": f.host + "/upload_photo"})
	case "/upload_photo":
		json.NewEncoder(w).Encode(map[string]any{"server": 1, "photo": "PDATA", "hash": "H"})
	case "/method/photos.saveMessagesPhoto":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.ok(w, []any{map[string]any{"id": 5, "owner_id": 42, "access_key": "ak"}})
	case "/method/account.setSilenceMode",
		"/method/messages.pinConversation", "/method/messages.unpinConversation",
		"/method/messages.archiveConversation", "/method/messages.unarchiveConversation",
		"/method/messages.removeChatUser":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.manage = append(f.manage, r.URL.Path[len("/method/"):]+" "+r.PostForm.Encode())
		f.mu.Unlock()
		f.ok(w, 1)
	case "/method/messages.sendReaction":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.lastReaction = cloneForm(r.PostForm)
		f.reacted = true
		f.mu.Unlock()
		f.ok(w, 1)
	case "/method/messages.deleteReaction":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		f.reacted = false
		f.mu.Unlock()
		f.ok(w, 1)
	case "/method/messages.getMessagesReactions":
		if !f.authed(r) {
			f.fail(w, 5, "auth")
			return
		}
		f.mu.Lock()
		reacted := f.reacted
		f.mu.Unlock()
		counts := []any{map[string]any{"reaction_id": 2, "count": 4, "reacted": false}}
		my := 0
		if reacted {
			counts = append(counts, map[string]any{"reaction_id": 5, "count": 1, "reacted": true})
			my = 5
		}
		f.ok(w, map[string]any{"items": []any{map[string]any{
			"conversation_message_id": 70, "reaction_id": my, "counts": counts,
		}}})
	default:
		f.t.Errorf("unexpected path %s", r.URL.Path)
		f.fail(w, 1, "no route")
	}
}

func newFake(t *testing.T) (*fakeVK, *Provider) {
	f := &fakeVK{t: t, goodTokens: map[string]bool{"ACCESS": true}}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	f.host = srv.URL
	host := strings.TrimPrefix(srv.URL, "http://")
	p := &Provider{scheme: "http", apiHost: host, authHost: host, oauthHost: host}
	return f, p
}

type sink struct {
	mu        sync.Mutex
	events    chan model.Event
	connected chan model.User
	saved     map[string]string
}

func newSink() *sink {
	return &sink{events: make(chan model.Event, 16), connected: make(chan model.User, 1)}
}

func (s *sink) Emit(ev model.Event)    { s.events <- ev }
func (s *sink) Connected(u model.User) { s.connected <- u }
func (s *sink) SaveParams(p map[string]string) error {
	s.mu.Lock()
	s.saved = p
	s.mu.Unlock()
	return nil
}

func TestQRLoginAndChats(t *testing.T) {
	_, p := newFake(t)
	env := provider.Env{HTTP: http.DefaultClient, State: provider.NewMemKV()}
	l, err := p.Login(context.Background(), methodQR, nil, env)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	ctx := context.Background()
	step, err := l.Next(ctx, nil) // first call: show the QR
	if err != nil {
		t.Fatal(err)
	}
	if step.QR == "" || step.Done {
		t.Fatalf("expected a QR step, got %+v", step)
	}
	step, err = l.Next(ctx, nil) // poll until approved
	if err != nil {
		t.Fatal(err)
	}
	if !step.Done || step.Result == nil {
		t.Fatalf("expected the login to finish, got %+v", step)
	}
	cr := *step.Result
	if cr.Login != "42" {
		t.Errorf("login id = %q, want 42", cr.Login)
	}
	if cr.Name != "Ivan Petrov" {
		t.Errorf("name = %q, want Ivan Petrov", cr.Name)
	}
	if cr.Params[keyCommonToken] != "COMMON" || cr.Params[keyAccessToken] != "ACCESS" {
		t.Errorf("params = %v", cr.Params)
	}

	// Open a session with those credentials and list chats.
	sess, err := p.Open(provider.Account{ID: "a1", Login: cr.Login, Params: cr.Params}, env)
	if err != nil {
		t.Fatal(err)
	}
	chats, err := sess.(provider.Messenger).Chats(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 1 || chats[0].ID != "100" || chats[0].Kind != model.ChatDirect {
		t.Fatalf("chats = %+v", chats)
	}
	if chats[0].Title != "Anna K" || chats[0].Unread != 3 {
		t.Errorf("chat = %+v", chats[0])
	}
	if chats[0].Last == nil || chats[0].Last.Text != "hi" {
		t.Errorf("last message = %+v", chats[0].Last)
	}
}

func TestTokenRefresh(t *testing.T) {
	f, p := newFake(t)
	env := provider.Env{HTTP: http.DefaultClient, State: provider.NewMemKV()}
	// Start with a stale access token; the good one comes from a refresh.
	params := map[string]string{keyAccessToken: "STALE", keyCommonToken: "COMMON", keyUserID: "42"}
	sess, err := p.Open(provider.Account{ID: "a1", Login: "42", Params: params}, env)
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	sk := newSink()
	s.mu.Lock()
	s.sink = sk
	s.mu.Unlock()

	chats, err := s.Chats(context.Background(), 50)
	if err != nil {
		t.Fatalf("Chats after refresh: %v", err)
	}
	if len(chats) != 1 {
		t.Fatalf("chats = %+v", chats)
	}
	if s.accessToken() != "ACCESS2" {
		t.Errorf("token after refresh = %q, want ACCESS2", s.accessToken())
	}
	sk.mu.Lock()
	saved := sk.saved
	sk.mu.Unlock()
	if saved[keyAccessToken] != "ACCESS2" {
		t.Errorf("SaveParams got %v", saved)
	}
	f.mu.Lock()
	n := f.silentUsed
	f.mu.Unlock()
	if n < 2 {
		t.Errorf("exchangeSilentAuthToken called %d times, want the 1114 retry (>=2)", n)
	}
}

func TestLongPollNewMessage(t *testing.T) {
	_, p := newFake(t)
	env := provider.Env{HTTP: http.DefaultClient, State: provider.NewMemKV()}
	params := map[string]string{keyAccessToken: "ACCESS", keyCommonToken: "COMMON", keyUserID: "42"}
	sess, err := p.Open(provider.Account{ID: "a1", Login: "42", Params: params}, env)
	if err != nil {
		t.Fatal(err)
	}
	sk := newSink()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- sess.Run(ctx, sk) }()

	select {
	case <-sk.connected:
	case <-time.After(2 * time.Second):
		t.Fatal("no Connected")
	}
	select {
	case ev := <-sk.events:
		if ev.Kind != model.EvMessageNew || ev.Message == nil || ev.Message.ID != "55" {
			t.Fatalf("event = %+v", ev)
		}
		if ev.Message.ChatID != "100" || ev.Message.Text != "ping" {
			t.Errorf("message = %+v", ev.Message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message event")
	}
	cancel()
	<-done
}

func cloneForm(v url.Values) url.Values {
	out := url.Values{}
	for k, vs := range v {
		out[k] = append([]string(nil), vs...)
	}
	return out
}

// openSession opens a session with a valid token and its sink set.
func openSession(t *testing.T, p *Provider) (*session, *sink) {
	t.Helper()
	env := provider.Env{HTTP: http.DefaultClient, State: provider.NewMemKV()}
	params := map[string]string{keyAccessToken: "ACCESS", keyCommonToken: "COMMON", keyUserID: "42"}
	sess, err := p.Open(provider.Account{ID: "a1", Login: "42", Params: params}, env)
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	sk := newSink()
	s.mu.Lock()
	s.sink = sk
	s.mu.Unlock()
	return s, sk
}

func TestSendWithMention(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	_, err := s.Send(context.Background(), "100", model.Outgoing{
		Text:     "hi Anna!",
		Mentions: []model.Mention{{Offset: 3, Length: 4, User: model.User{ID: "100"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	got := f.lastSend.Get("message")
	f.mu.Unlock()
	if got != "hi [id100|Anna]!" {
		t.Errorf("sent message = %q", got)
	}
}

func TestEditAndDelete(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	ctx := context.Background()
	m, err := s.Edit(ctx, "100", "55", "fixed")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "55" || m.Text != "fixed" || m.Edited.IsZero() {
		t.Errorf("edit result = %+v", m)
	}
	f.mu.Lock()
	ed := f.lastEdit
	f.mu.Unlock()
	if ed.Get("peer_id") != "100" || ed.Get("message_id") != "55" || ed.Get("message") != "fixed" {
		t.Errorf("edit form = %v", ed)
	}
	if !s.DeletesForMe() {
		t.Error("DeletesForMe = false")
	}
	if err := s.Delete(ctx, "100", []string{"55", "56"}, true); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	del := f.lastDelete
	f.mu.Unlock()
	if del.Get("message_ids") != "55,56" || del.Get("delete_for_all") != "1" {
		t.Errorf("delete form = %v", del)
	}
}

func TestForward(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	if _, err := s.Forward(context.Background(), "100", []string{"7", "8"}, "2000000001"); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	got := f.lastSend
	f.mu.Unlock()
	if got.Get("peer_id") != "2000000001" || got.Get("forward_messages") != "7,8" {
		t.Errorf("forward form = %v", got)
	}
}

func TestSendImageFile(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	_, err := s.SendFiles(context.Background(), "100", model.Outgoing{
		Text:  "pic",
		Files: []model.OutFile{{Name: "a.jpg", MIME: "image/jpeg", Data: []byte("bytes")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	got := f.lastSend
	f.mu.Unlock()
	if got.Get("attachment") != "photo42_5_ak" || got.Get("message") != "pic" {
		t.Errorf("file send form = %v", got)
	}
}

func TestMembers(t *testing.T) {
	_, p := newFake(t)
	s, _ := openSession(t, p)
	ms, err := s.Members(context.Background(), "2000000001", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	// Self (42) is dropped; community members too. Anna (100) remains.
	if len(ms) != 1 || ms[0].ID != "100" || ms[0].Name != "Anna K" {
		t.Fatalf("members = %+v", ms)
	}
}

func TestProfileUser(t *testing.T) {
	_, p := newFake(t)
	s, _ := openSession(t, p)
	pr, err := s.Profile(context.Background(), "100")
	if err != nil {
		t.Fatal(err)
	}
	if pr.Title != "Ivan Petrov" || pr.Username != "ivan" || pr.About != "hello" {
		t.Errorf("profile = %+v", pr)
	}
	if pr.Avatar != "https://vk.com/i200.jpg" || pr.Link != "https://vk.com/ivan" || !pr.Online {
		t.Errorf("profile = %+v", pr)
	}
}

func TestProfileGroupChatAndCommunity(t *testing.T) {
	_, p := newFake(t)
	s, _ := openSession(t, p)
	ctx := context.Background()
	gc, err := s.Profile(ctx, "2000000001")
	if err != nil {
		t.Fatal(err)
	}
	if gc.Title != "Team" || gc.Members != 3 || len(gc.MemberList) == 0 {
		t.Errorf("group chat profile = %+v", gc)
	}
	com, err := s.Profile(ctx, "-777")
	if err != nil {
		t.Fatal(err)
	}
	if com.Title != "Cats" || com.About != "meow" || com.Members != 9000 || com.Link != "https://vk.com/cats" {
		t.Errorf("community profile = %+v", com)
	}
}

func TestParseVKText(t *testing.T) {
	text, spans := parseVKText("hi [id100|Anna] and [club5|Cats] ok")
	if text != "hi Anna and Cats ok" {
		t.Fatalf("text = %q", text)
	}
	if len(spans) != 2 {
		t.Fatalf("spans = %+v", spans)
	}
	if spans[0].Kind != model.SpanLink || spans[0].Offset != 3 || spans[0].Length != 4 || spans[0].URL != "https://vk.com/id100" {
		t.Errorf("span0 = %+v", spans[0])
	}
	if spans[1].URL != "https://vk.com/club5" || spans[1].Offset != 12 {
		t.Errorf("span1 = %+v", spans[1])
	}
	// Non-mention brackets are left untouched.
	if got, _ := parseVKText("array[0] = x"); got != "array[0] = x" {
		t.Errorf("plain text changed: %q", got)
	}
}

func TestChatManager(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	ctx := context.Background()
	acts := s.ChatActions()
	if len(acts) != 7 {
		t.Errorf("actions = %v", acts)
	}
	for _, a := range []model.ChatAction{model.ChatMute, model.ChatPin, model.ChatArchive} {
		if err := s.ManageChat(ctx, "100", a); err != nil {
			t.Fatalf("action %v: %v", a, err)
		}
	}
	// Leave a direct chat is refused; a group chat is allowed.
	if err := s.ManageChat(ctx, "100", model.ChatLeave); err == nil {
		t.Error("leaving a direct chat should fail")
	}
	if err := s.ManageChat(ctx, "2000000005", model.ChatLeave); err != nil {
		t.Fatalf("leave group: %v", err)
	}
	f.mu.Lock()
	got := strings.Join(f.manage, "\n")
	f.mu.Unlock()
	for _, want := range []string{
		"account.setSilenceMode", "peer_id=100", "time=-1",
		"messages.pinConversation", "messages.archiveConversation",
		"messages.removeChatUser", "chat_id=5", "user_id=42",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("manage calls missing %q in:\n%s", want, got)
		}
	}
}

func TestReact(t *testing.T) {
	f, p := newFake(t)
	s, _ := openSession(t, p)
	ctx := context.Background()
	rs, err := s.React(ctx, "100", "55", ":vk-5:", true)
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	rf := f.lastReaction
	f.mu.Unlock()
	if rf.Get("cmid") != "70" || rf.Get("reaction_id") != "5" || rf.Get("peer_id") != "100" {
		t.Errorf("sendReaction form = %v", rf)
	}
	// After reacting: id 2 (count 4, not mine) and id 5 (count 1, mine).
	var mine *model.Reaction
	for i := range rs {
		if rs[i].Mine {
			mine = &rs[i]
		}
	}
	if mine == nil || mine.Emoji != ":vk-5:" || mine.Count != 1 {
		t.Errorf("reactions = %+v", rs)
	}

	rs, err = s.React(ctx, "100", "55", ":vk-5:", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		if r.Mine {
			t.Errorf("still mine after removal: %+v", rs)
		}
	}
}
