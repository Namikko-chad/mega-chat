package vkteams

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

// fakeServer implements wire.go's view of the protocol.
type fakeServer struct {
	t       *testing.T
	mu      sync.Mutex
	aimsid  string
	fetches int
	sent    map[string]any
	read    map[string]any
	managed []string // "method buddy" of buddylist calls
}

func (f *fakeServer) wim(w http.ResponseWriter, code int, data any) {
	json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"statusCode": code, "statusText": "x", "data": data}})
}

func (f *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	switch r.URL.Path {
	case "/api/v132/wim/auth/clientLogin":
		if r.Form.Get("pwd") != "secret" || r.Form.Get("devId") != "dev" {
			f.wim(w, 330, nil)
			return
		}
		f.wim(w, 200, map[string]any{"token": map[string]any{"a": "TOKEN"}, "sessionSecret": "SS", "hostTime": time.Now().Unix()})
	case "/api/v132/wim/aim/startSession":
		p := url.Values{}
		for k, v := range r.PostForm {
			p[k] = v
		}
		got := p.Get("sig_sha256")
		sign("POST", "http://"+r.Host+r.URL.Path, p, sessionKey("secret", "SS"), mustInt(p.Get("ts")))
		if got != p.Get("sig_sha256") || r.Form.Get("a") != "TOKEN" {
			f.wim(w, 401, nil)
			return
		}
		f.aimsid = "AIM2"
		f.wim(w, 200, map[string]any{"aimsid": "AIM2", "fetchBaseURL": "http://" + r.Host + "/fetch?aimsid=AIM2&seqNum=1",
			"myInfo": map[string]any{"aimId": "me@corp.ru", "friendly": "Me"}})
	case "/api/v132/wim/fetchEvents", "/api/v150/bos/node1/aim/fetchEvents", "/fetch":
		if r.Form.Get("aimsid") != f.aimsid {
			f.wim(w, 401, nil)
			return
		}
		f.fetches++
		next := "http://" + r.Host + "/fetch?aimsid=" + f.aimsid + "&seqNum=2"
		if f.fetches > 1 {
			f.mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			f.mu.Lock()
			f.wim(w, 200, map[string]any{"fetchBaseURL": next, "events": []any{}})
			return
		}
		f.wim(w, 200, map[string]any{"fetchBaseURL": next, "events": []any{
			map[string]any{"type": "myInfo", "eventData": map[string]any{"aimId": "me@corp.ru", "friendly": "Me"}},
			map[string]any{"type": "histDlgState", "eventData": map[string]any{
				"sn": "ann@corp.ru", "unreadCnt": 1, "lastMsgId": 102,
				"persons":  []any{map[string]any{"sn": "ann@corp.ru", "friendly": "Ann"}},
				"messages": []any{map[string]any{"msgId": 102, "time": 1700000000, "text": "hi"}},
			}},
			map[string]any{"type": "typing", "eventData": map[string]any{"aimId": "ann@corp.ru", "typingStatus": "typing"}},
			map[string]any{"type": "presence", "eventData": map[string]any{"aimId": "ann@corp.ru", "state": "online"}},
		}})
	case "/api/v132/rapi/getHistory":
		var body struct {
			AimSID string         `json:"aimsid"`
			Params map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.AimSID != f.aimsid {
			json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 40101}})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{
			"messages": []any{
				map[string]any{"msgId": 100, "time": 1699999000, "text": "first", "outgoing": true},
				map[string]any{"msgId": 101, "time": 1699999500, "text": "second"},
				map[string]any{"msgId": 102, "time": 1700000000, "text": "hi"},
			}}})
	case "/api/v132/rapi/message/send", "/api/v150/rapi/message/send":
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		f.sent = body
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{"msgId": 103, "state": "accepted"}})
	case "/api/v150/rapi/contacts/get":
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{
			"contacts": []any{map[string]any{"sn": "ann@corp.ru"}, map[string]any{"sn": "42@chat.agent"}, map[string]any{"sn": "bob@corp.ru"}},
			"persons":  []any{map[string]any{"sn": "ann@corp.ru", "friendly": "Ann"}, map[string]any{"sn": "bob@corp.ru", "friendly": "bob@corp.ru"}},
		}})
	case "/api/v132/rapi/getUserInfo", "/api/v150/rapi/getUserInfo":
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		switch body["params"].(map[string]any)["sn"] {
		case "me@corp.ru":
		case "bob@corp.ru": // VK WorkSpace: friendly is the login, the name is in first/lastName
			json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{"firstName": "Bob", "lastName": "Builder", "friendly": "bob@corp.ru"}})
			return
		default:
			json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 40400}})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{"firstName": "Me", "lastName": "Self", "friendly": ""}})
	case "/api/v150/files/avatar/get":
		if r.Form.Get("aimsid") != f.aimsid {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Form.Get("targetSn") != "ann@corp.ru" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("\xff\xd8\xff\xe0JPEG"))
	case "/api/v150/rapi/getChatInfo":
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{"sn": "42@chat.agent", "name": "Team",
			"about": "Our team", "membersCount": "2", "members": []any{map[string]any{"sn": "ann@corp.ru"}, map[string]any{"sn": "me@corp.ru"}}}})
	case "/api/v132/rapi/setDlgState":
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		f.read = body["params"].(map[string]any)
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}})
	case "/api/v132/wim/buddylist/mute", "/api/v132/wim/buddylist/hideChat", "/api/v132/wim/buddylist/removeBuddy":
		if r.Form.Get("aimsid") != f.aimsid {
			f.wim(w, 401, nil)
			return
		}
		f.managed = append(f.managed, strings.TrimPrefix(r.URL.Path, "/api/v132/wim/buddylist/")+" "+r.Form.Get("buddy")+" "+r.Form.Get("timeout")+r.Form.Get("lastMsgId"))
		f.wim(w, 200, nil)
	default:
		http.NotFound(w, r)
	}
}

func mustInt(s string) int64 {
	var n int64
	json.Unmarshal([]byte(s), &n)
	return n
}

type sink struct {
	mu        sync.Mutex
	events    []model.Event
	connected chan model.User
}

func (s *sink) Emit(ev model.Event) {
	s.mu.Lock()
	s.events = append(s.events, ev)
	s.mu.Unlock()
}
func (s *sink) Connected(u model.User)             { s.connected <- u }
func (s *sink) SaveParams(map[string]string) error { return nil }

func (s *sink) has(kind model.EventKind) *model.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.events {
		if s.events[i].Kind == kind {
			return &s.events[i]
		}
	}
	return nil
}

func TestPasswordLoginSessionAndRenew(t *testing.T) {
	fs := &fakeServer{t: t, aimsid: "none"}
	srv := httptest.NewServer(fs)
	defer srv.Close()
	p := &Provider{scheme: "http"}
	host := strings.TrimPrefix(srv.URL, "http://")
	ctx := context.Background()
	env := provider.Env{HTTP: srv.Client()}

	l, _ := p.Login(ctx, methodPassword, nil, env)
	step, _ := l.Next(ctx, nil)
	if len(step.Fields) != 5 {
		t.Fatalf("fields: %+v", step.Fields)
	}
	step, err := l.Next(ctx, map[string]string{"host": host, "login": "me@corp.ru", "password": "bad", "dev_id": "dev"})
	if err != nil || step.Error == "" {
		t.Fatalf("bad password: %+v %v", step, err)
	}
	step, err = l.Next(ctx, map[string]string{"host": "http://" + host + "/", "login": "me@corp.ru", "password": "secret", "dev_id": "dev"})
	if err != nil || !step.Done {
		t.Fatalf("login: %+v %v", step, err)
	}
	cr := step.Result
	if cr.Login != host+"/me@corp.ru" || cr.Name != "Me" || cr.Params["token"] != "TOKEN" || cr.Params["password"] != "" {
		t.Fatalf("credentials: %+v", cr)
	}
	for _, v := range cr.Params {
		if v == "secret" {
			t.Fatal("password stored")
		}
	}

	// The stored aimsid is dead: the session renews it with the token.
	kv := provider.NewMemKV()
	kv.Put(keyAimsid, []byte("DEAD"))
	sess, _ := p.Open(provider.Account{ID: "a", Login: cr.Login, Params: cr.Params}, provider.Env{State: kv, HTTP: srv.Client()})
	k := &sink{connected: make(chan model.User, 1)}
	rctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- sess.Run(rctx, k) }()
	select {
	case u := <-k.connected:
		if u.Name != "Me" {
			t.Fatalf("self: %+v", u)
		}
	case err := <-done:
		t.Fatalf("run: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for k.has(model.EvChat) == nil || k.has(model.EvTyping) == nil || k.has(model.EvPresence) == nil {
		if time.Now().After(deadline) {
			t.Fatal("no chat/typing/presence events")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if ev := k.has(model.EvMessageNew); ev == nil || ev.Message.Sender.Name != "Ann" || ev.Message.Text != "hi" {
		t.Fatalf("message: %+v", ev)
	}
	if ev := k.has(model.EvChat); ev.Chat.Title != "Ann" || ev.Chat.Unread != 1 || ev.Chat.Kind != model.ChatDirect {
		t.Fatalf("chat: %+v", ev.Chat)
	}
	if ev := k.has(model.EvPresence); ev.Chat.ID != "ann@corp.ru" || !ev.Online {
		t.Fatalf("presence: %+v %v", ev.Chat, ev.Online)
	}

	m := sess.(provider.Messenger)
	chats, _ := m.Chats(ctx, 10)
	// Ann's chat and the saved chat with oneself, listed even when empty.
	if len(chats) != 2 || chats[0].Last == nil || chats[1].ID != "me@corp.ru" || !chats[1].Saved {
		t.Fatalf("chats: %+v", chats)
	}
	cm := sess.(provider.ChatManager)
	if err := cm.ManageChat(ctx, "ann@corp.ru", model.ChatMute); err != nil {
		t.Fatal(err)
	}
	if err := cm.ManageChat(ctx, "ann@corp.ru", model.ChatArchive); err != nil {
		t.Fatal(err)
	}
	if err := cm.ManageChat(ctx, "ann@corp.ru", model.ChatLeave); !errors.Is(err, provider.ErrUnsupported) {
		t.Fatalf("leave a direct chat: %v", err)
	}
	if chats, _ = m.Chats(ctx, 10); !chats[0].Muted || !chats[0].Archived {
		t.Fatalf("after mute and archive: %+v", chats[0])
	}
	fs.mu.Lock()
	managed := strings.Join(fs.managed, ";")
	fs.mu.Unlock()
	if managed != "mute ann@corp.ru 2147483647;hideChat ann@corp.ru 102" {
		t.Fatalf("buddylist calls: %s", managed)
	}
	msgs, _, err := m.History(ctx, "ann@corp.ru", "102", 10)
	if err != nil || len(msgs) != 2 || msgs[0].ID != "101" || !msgs[1].Outgoing {
		t.Fatalf("history: %+v %v", msgs, err)
	}
	sent, err := m.Send(ctx, "ann@corp.ru", model.Outgoing{Text: "yo"})
	if err != nil || sent.ID != "103" || fs.sent["aimsid"] != "AIM2" ||
		fs.sent["params"].(map[string]any)["target"] != "ann@corp.ru" {
		t.Fatalf("send: %+v %v %v", sent, err, fs.sent)
	}
	if err := m.MarkRead(ctx, "ann@corp.ru", ""); err != nil || fs.read["lastRead"] != "102" {
		t.Fatalf("read: %v %v", err, fs.read)
	}
	cancel()
	<-done
}

func TestAimsidSessionEnds(t *testing.T) {
	fs := &fakeServer{t: t, aimsid: "AIM1"}
	srv := httptest.NewServer(fs)
	defer srv.Close()
	p := &Provider{scheme: "http"}
	host := strings.TrimPrefix(srv.URL, "http://")
	ctx := context.Background()
	l, _ := p.Login(ctx, methodAimsid, nil, provider.Env{HTTP: srv.Client()})
	step, err := l.Next(ctx, map[string]string{"host": host, "aimsid": "AIM1"})
	if err != nil || !step.Done || step.Result.Login != host+"/me@corp.ru" {
		t.Fatalf("login: %+v %v", step, err)
	}
	// The web session is gone: there is no token to renew it.
	fs.mu.Lock()
	fs.aimsid = "OTHER"
	fs.mu.Unlock()
	sess, _ := p.Open(provider.Account{ID: "a", Login: step.Result.Login, Params: step.Result.Params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err := sess.Run(ctx, &sink{connected: make(chan model.User, 1)}); !errors.Is(err, provider.ErrAuth) {
		t.Fatalf("run: %v", err)
	}
}

func TestLoginWithFetchURL(t *testing.T) {
	const aimsid = "005.1:me@corp.ru"
	fs := &fakeServer{t: t, aimsid: aimsid, fetches: 1} // a resumed poll: no myInfo
	srv := httptest.NewServer(fs)
	defer srv.Close()
	p := &Provider{scheme: "http"}
	host := strings.TrimPrefix(srv.URL, "http://")
	ctx := context.Background()
	l, _ := p.Login(ctx, methodAimsid, nil, provider.Env{HTTP: srv.Client()})
	pasted := "https://" + host + "/api/v150/bos/node1/aim/fetchEvents?aimsid=" + url.QueryEscape(aimsid) +
		"&rnd=1.5&seqNum=7&f=json&r=9&timeout=60000&peek=0&hidden=1"
	// The fields are left as they were: the address wins.
	step, err := l.Next(ctx, map[string]string{"host": "u.example.org", "api_version": "92", "aimsid": pasted})
	if err != nil || !step.Done {
		t.Fatalf("login: %+v %v", step, err)
	}
	c := step.Result
	if c.Login != host+"/me@corp.ru" || c.Name != "Me Self" || c.Params["aimsid"] != aimsid ||
		c.Params["host"] != host || c.Params["api_version"] != "v150" {
		t.Fatalf("credentials: %+v", c)
	}
	if strings.Contains(c.Params["aimsid"], "seqNum") {
		t.Fatalf("aimsid carries the URL: %q", c.Params["aimsid"])
	}
}

func TestLoginRejectsMaskedAimsid(t *testing.T) {
	p := &Provider{scheme: "http"}
	ctx := context.Background()
	l, _ := p.Login(ctx, methodAimsid, nil, provider.Env{HTTP: http.DefaultClient})
	step, err := l.Next(ctx, map[string]string{"host": "127.0.0.1:1", "aimsid": "005.1234567890.**********:me@corp.ru"})
	if err != nil || step.Done || !strings.Contains(step.Error, "masked") {
		t.Fatalf("step: %+v %v", step, err)
	}
}

func TestTransportErrorHasNoAimsid(t *testing.T) {
	p := &Provider{scheme: "http"}
	ctx := context.Background()
	l, _ := p.Login(ctx, methodAimsid, nil, provider.Env{HTTP: http.DefaultClient})
	step, err := l.Next(ctx, map[string]string{"host": "127.0.0.1:1", "aimsid": "SECRETAIMSID"})
	if err != nil || step.Done || step.Error == "" || strings.Contains(step.Error, "SECRETAIMSID") {
		t.Fatalf("step: %+v %v", step, err)
	}
}

func TestNormVersion(t *testing.T) {
	for in, want := range map[string]string{"150": "v150", "v150": "v150", " /api/v132/ ": "v132", "": ""} {
		if got := normVersion(in); got != want {
			t.Errorf("normVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResumedSessionListsChats(t *testing.T) {
	const aimsid = "005.1:me@corp.ru"
	fs := &fakeServer{t: t, aimsid: aimsid, fetches: 1} // a resumed poll: no events
	srv := httptest.NewServer(fs)
	defer srv.Close()
	p := &Provider{scheme: "http"}
	host := strings.TrimPrefix(srv.URL, "http://")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l, _ := p.Login(ctx, methodAimsid, nil, provider.Env{HTTP: srv.Client()})
	step, err := l.Next(ctx, map[string]string{"aimsid": "http://" + host + "/api/v150/bos/node1/aim/fetchEvents?aimsid=" + url.QueryEscape(aimsid) + "&seqNum=7"})
	if err != nil || !step.Done {
		t.Fatalf("login: %+v %v", step, err)
	}
	kv := provider.NewMemKV()
	for k, v := range step.Result.State {
		kv.Put(k, v)
	}
	sess, _ := p.Open(provider.Account{ID: "a", Login: step.Result.Login, Params: step.Result.Params}, provider.Env{State: kv, HTTP: srv.Client()})
	k := &sink{connected: make(chan model.User, 1)}
	done := make(chan error, 1)
	go func() { done <- sess.Run(ctx, k) }()
	select {
	case u := <-k.connected:
		if u.Name != "Me Self" || u.Avatar != "avatar:me@corp.ru" {
			t.Fatalf("self: %+v", u)
		}
	case err := <-done:
		t.Fatalf("run: %v", err)
	}
	md := sess.(provider.Media)
	if data, mime, err := md.Media(ctx, "avatar:ann@corp.ru", 1<<20); err != nil || mime != "image/jpeg" || len(data) == 0 {
		t.Fatalf("avatar: %q %q %v", data, mime, err)
	}
	if _, _, err := md.Media(ctx, "avatar:42@chat.agent", 1<<20); !errors.Is(err, errNoAvatar) {
		t.Fatalf("no avatar: %v", err)
	}
	if _, _, err := md.Media(ctx, "avatar:ann@corp.ru", 3); !errors.Is(err, provider.ErrTooLarge) {
		t.Fatalf("limit: %v", err)
	}
	if _, _, err := md.Media(ctx, "file123", 1<<20); err == nil {
		t.Fatal("a file ref was fetched")
	}
	m := sess.(provider.Messenger)
	chats, err := m.Chats(ctx, 10)
	if err != nil || len(chats) != 4 { // three and the saved chat
		t.Fatalf("chats: %+v %v", chats, err)
	}
	titles := map[string]string{}
	for _, c := range chats {
		titles[c.ID] = c.Title
	}
	if titles["ann@corp.ru"] != "Ann" {
		t.Fatalf("titles: %v", titles)
	}
	// Group names and names of people the contacts gave none come later, as chat updates.
	deadline := time.Now().Add(3 * time.Second)
	for {
		k.mu.Lock()
		group, person := false, false
		for _, ev := range k.events {
			if ev.Kind == model.EvChat && ev.Chat.ID == "42@chat.agent" && ev.Chat.Title == "Team" && ev.Chat.Unread == -1 {
				group = true
			}
			if ev.Kind == model.EvChat && ev.Chat.ID == "bob@corp.ru" && ev.Chat.Title == "Bob Builder" && ev.Chat.Unread == -1 {
				person = true
			}
		}
		k.mu.Unlock()
		if group && person {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("no group or person name: group %v, person %v", group, person)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := m.Send(ctx, "42@chat.agent", model.Outgoing{Text: "yo", ReplyTo: "77"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	q := fs.sent["params"].(map[string]any)["quoteParts"].([]any)[0].(map[string]any)
	if q["msgId"] != float64(77) {
		t.Fatalf("quote: %v", q)
	}
	pr := sess.(provider.Profiles)
	if g, err := pr.Profile(ctx, "42@chat.agent"); err != nil || g.Title != "Team" || g.About != "Our team" || g.Members != 2 || len(g.MemberList) != 2 || g.MemberList[0].Name != "Ann" {
		t.Fatalf("group profile: %+v %v", g, err)
	}
	if u, err := pr.Profile(ctx, "bob@corp.ru"); err != nil || u.Title != "Bob Builder" || u.Email != "bob@corp.ru" {
		t.Fatalf("user profile: %+v %v", u, err)
	}
	cancel()
	<-done
}

func TestMentions(t *testing.T) {
	if !mentions("hi @[ann@corp.ru], look", "ann@corp.ru") || mentions("hi @[bob@corp.ru]", "ann@corp.ru") || mentions("@[]", "") {
		t.Fatal("mentions")
	}
}

func TestAvatarHost(t *testing.T) {
	for host, want := range map[string]string{"u.myteam.vmailru.net": "ub.myteam.vmailru.net", "teams.corp.ru": "teams.corp.ru"} {
		if got := (&server{host: host}).avatarHost(); got != want {
			t.Errorf("avatarHost(%q) = %q, want %q", host, got, want)
		}
	}
}

func TestFilesharingNumbersQuoted(t *testing.T) {
	var m wMessage
	in := `{"msgId":"1","filesharing":[{"name":"a.ogg","mime":"audio/ogg","size":"12345","duration":"7","id":"f1"},{"size":678,"duration":null},{"size":""}]}`
	if err := json.Unmarshal([]byte(in), &m); err != nil {
		t.Fatal(err)
	}
	if f := m.Filesharing; f[0].Size != 12345 || f[0].Duration != 7 || f[1].Size != 678 || f[1].Duration != 0 || f[2].Size != 0 {
		t.Fatalf("filesharing: %+v", f)
	}
}

func TestMessageBody(t *testing.T) {
	for raw, want := range map[string]string{
		`{"msgId":1,"text":"plain"}`:                                                                 "plain",
		`{"msgId":1,"parts":{"mainPart":{"text":{"plain":"main"}}}}`:                                 "main",
		`{"msgId":1,"parts":[{"mediaType":"quote","text":"old"},{"mediaType":"text","text":"new"}]}`: "new",
		`{"msgId":1,"parts":[{"mediaType":"forward","text":"old"}]}`:                                 "old",
		`{"msgId":1}`: "",
	} {
		var m wMessage
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatal(err)
		}
		if got := m.body(); got != want {
			t.Errorf("%s: got %q, want %q", raw, got, want)
		}
	}
	var m wMessage
	if err := json.Unmarshal([]byte(`{"msgId":1,"secret":"value"}`), &m); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(m.shape, "value") || !strings.Contains(m.shape, "secret:string") {
		t.Errorf("shape must list keys only: %s", m.shape)
	}
}

func TestFileDownload(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/files/info/abc/":
			if r.URL.Query().Get("aimsid") != "sid" {
				t.Errorf("info without the session: %s", r.URL.RawQuery)
			}
			json.NewEncoder(w).Encode(map[string]any{"status": 200, "result": map[string]any{"info": map[string]any{"dlink": ts.URL + "/dl/abc"}}})
		case "/dl/abc":
			if r.URL.Query().Get("aimsid") != "" {
				t.Errorf("the session went to the download link")
			}
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("picture"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	srv := &server{scheme: "http", host: u.Host, version: "v1", http: ts.Client()}
	link, err := srv.fileLink(context.Background(), "sid", "abc")
	if err != nil {
		t.Fatal(err)
	}
	data, mime, err := srv.fileGet(context.Background(), link, 100)
	if err != nil || string(data) != "picture" || mime != "image/jpeg" {
		t.Fatalf("%q %q %v", data, mime, err)
	}
	if _, _, err := srv.fileGet(context.Background(), link, 3); !errors.Is(err, errTooLarge) {
		t.Fatalf("limit: %v", err)
	}
	if _, _, err := srv.fileGet(context.Background(), "https://evil.example/x", 100); err == nil {
		t.Fatal("a foreign host must be refused")
	}
	own := &server{host: "u.myteam.vmailru.net"}
	for host, want := range map[string]bool{"files.vmailru.net": true, "u.myteam.vmailru.net": true, "vmailru.net.evil.io": false, "evilvmailru.net": false} {
		if own.sameDomain(host) != want {
			t.Errorf("sameDomain(%s) != %v", host, want)
		}
	}
}

func TestMessageForwardAndQuote(t *testing.T) {
	names := func(id string) string {
		if id == "oleg@x" {
			return "Oleg"
		}
		return id
	}
	read := func(raw string) msgContent {
		t.Helper()
		var m wMessage
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatal(err)
		}
		return m.content(names)
	}
	// The shape seen from the server.
	c := read(`{"msgId":"5","mediaType":"forward","time":1,"parts":[{"chat":{"name":"Front","sn":"1@chat.agent"},"mediaType":"forward","msgId":"3","sn":"oleg@x","text":"hello","time":1}]}`)
	if c.text != "hello" || !c.forwarded || c.from != "Oleg · Front" {
		t.Fatalf("forward: %+v", c)
	}
	c = read(`{"msgId":"5","parts":[{"mediaType":"forward","sn":"oleg@x","text":"a"},{"mediaType":"forward","sn":"ann@x","chat":{"name":"Ann"},"text":"b"}]}`)
	if c.text != "Oleg:\na\n\nAnn:\nb" || c.from != "" {
		t.Fatalf("two authors: %+v", c)
	}
	c = read(`{"msgId":"5","parts":[{"mediaType":"quote","msgId":"3","sn":"oleg@x","text":"old"},{"mediaType":"text","text":"new"}]}`)
	if c.text != "new" || c.replyTo != "3" || c.quote != "old" || c.forwarded {
		t.Fatalf("quote: %+v", c)
	}
	c = read(`{"msgId":"5","parts":[{"mediaType":"text","text":"https://f/get/1"}],"filesharing":[{"id":"1","original_url":"https://f/get/1"}]}`)
	if c.text != "" {
		t.Fatalf("a file's own link is not its text: %+v", c)
	}
}

func TestServiceText(t *testing.T) {
	names := func(id string) string {
		return map[string]string{"admin@x": "Administrator", "a@x": "Valeria", "b@x": "Yuri"}[id]
	}
	for raw, want := range map[string]string{
		`{"msgId":"1","class":"event","chat":{"sender":"admin@x","memberEvent":{"type":"delMembers","members":["a@x"]}}}`:       "Administrator removed Valeria",
		`{"msgId":"1","class":"event","chat":{"sender":"admin@x","memberEvent":{"type":"addMembers","members":["a@x","b@x"]}}}`: "Administrator added Valeria, Yuri",
		`{"msgId":"1","class":"event","chat":{"sender":"a@x","memberEvent":{"type":"leave","members":["a@x"]}}}`:                "Valeria left the chat",
		`{"msgId":"1","class":"event","chat":{"sender":"admin@x","modifiedInfo":{"name":"Front"}}}`:                             "Administrator changed the title to «Front»",
		`{"msgId":"1","class":"event","chat":{"sender":"admin@x","modifiedInfo":{"about":"x"}}}`:                                "Administrator changed the description",
		`{"msgId":"1","class":"event","chat":{"sender":"admin@x","memberEvent":{"type":"somethingNew"}}}`:                       "",
	} {
		var m wMessage
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatal(err)
		}
		if !m.isService() {
			t.Errorf("%s: not a service message", raw)
		}
		if got := m.serviceText(names); got != want {
			t.Errorf("%s: got %q, want %q", raw, got, want)
		}
	}
}

func TestReactions(t *testing.T) {
	var gone bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Params map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		one := map[string]any{"msgId": 7, "chatId": "c@chat.agent", "myReaction": "👍",
			"reactions": []any{map[string]any{"reaction": "👍", "counter": 2}, map[string]any{"reaction": "👎", "counter": 0}}}
		var results any
		switch {
		case strings.HasSuffix(r.URL.Path, "/reaction/add"):
			if body.Params["reaction"] != "🚀" && body.Params["reaction"] != "👍" {
				t.Errorf("add: %v", body.Params)
			}
			if _, ok := body.Params["msgId"].(float64); !ok {
				t.Errorf("msgId must be a number: %v", body.Params["msgId"])
			}
			results = one
		case strings.HasSuffix(r.URL.Path, "/reaction/remove"):
			gone, results = true, map[string]any{}
		case strings.HasSuffix(r.URL.Path, "/reaction/get"):
			if gone {
				one["myReaction"] = ""
				one["reactions"] = []any{map[string]any{"reaction": "👍", "counter": 1}}
			}
			results = map[string]any{"reactions": []any{one}}
		}
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": rapiOK}, "results": results})
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	s := &session{aimsid: "sid", online: true, srv: &server{scheme: "http", host: u.Host, version: "v1", http: ts.Client()}}
	var _ provider.Reactor = s
	ctx := context.Background()
	got, err := s.React(ctx, "c@chat.agent", "7", "👍", true)
	if err != nil || len(got) != 1 || got[0] != (model.Reaction{Emoji: "👍", Count: 2, Mine: true}) {
		t.Fatalf("add: %+v %v", got, err)
	}
	got, err = s.React(ctx, "c@chat.agent", "7", "👍", false)
	if err != nil || len(got) != 1 || got[0] != (model.Reaction{Emoji: "👍", Count: 1}) {
		t.Fatalf("remove: %+v %v", got, err)
	}
}

type threadSink struct {
	provider.Sink
	evs []model.Event
}

func (t *threadSink) Emit(e model.Event) { t.evs = append(t.evs, e) }

func TestThreadIsNotAChat(t *testing.T) {
	sess, err := New().Open(provider.Account{ID: "a", Login: "h/me@corp.ru"}, provider.Env{State: provider.NewMemKV()})
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	s.dialogs["7@chat.agent"] = &dialog{SN: "7@chat.agent"}
	s.dialogs["1@chat.agent"] = &dialog{SN: "1@chat.agent", Title: "Group"}
	sink := &threadSink{}
	s.handle(sink, rawEvent{Type: "histDlgState", EventData: json.RawMessage(`{"sn":"7@chat.agent","parentTopic":{"type":"message","chatId":"1@chat.agent","messageId":42},
		"messages":[{"msgId":"50","time":1700000000,"text":"**reply**","chat":{"sender":"bob@corp.ru"}}]}`)})
	if len(sink.evs) != 2 || sink.evs[0].Kind != model.EvChatRemoved || sink.evs[0].Chat.ID != "7@chat.agent" {
		t.Fatalf("events: %+v", sink.evs)
	}
	r := sink.evs[1]
	if r.Kind != model.EvThreadMessage || r.Message.ChatID != "1@chat.agent" || r.Message.InThread != "42" || r.Message.Text != "reply" {
		t.Fatalf("reply: %+v", r.Message)
	}
	chats, err := s.Chats(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range chats {
		if c.ID == "7@chat.agent" {
			t.Fatal("the thread is in the chat list")
		}
	}
	root := s.convert("1@chat.agent", wMessage{MsgID: "42", Text: "root", Thread: &wThread{ThreadID: "7@chat.agent"}})
	if root.Thread == nil || root.Thread.ID != "42" || s.threadOf[msgKey("1@chat.agent", "42")] != "7@chat.agent" {
		t.Fatalf("root: %+v", root.Thread)
	}
	if got := s.homeOf("1@chat.agent", "50"); got != "7@chat.agent" {
		t.Fatalf("reply lives in %q", got)
	}
}
