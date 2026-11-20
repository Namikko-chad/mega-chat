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
			"contacts": []any{map[string]any{"sn": "ann@corp.ru"}, map[string]any{"sn": "42@chat.agent"}},
			"persons":  []any{map[string]any{"sn": "ann@corp.ru", "friendly": "Ann"}},
		}})
	case "/api/v132/rapi/getUserInfo", "/api/v150/rapi/getUserInfo":
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["params"].(map[string]any)["sn"] != "me@corp.ru" {
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
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": map[string]any{"sn": "42@chat.agent", "name": "Team"}})
	case "/api/v132/rapi/setDlgState":
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		f.read = body["params"].(map[string]any)
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}})
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
	if len(chats) != 1 || chats[0].Last == nil {
		t.Fatalf("chats: %+v", chats)
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
	if err != nil || len(chats) != 2 {
		t.Fatalf("chats: %+v %v", chats, err)
	}
	titles := map[string]string{}
	for _, c := range chats {
		titles[c.ID] = c.Title
	}
	if titles["ann@corp.ru"] != "Ann" {
		t.Fatalf("titles: %v", titles)
	}
	// Group names come later, as chat updates.
	deadline := time.Now().Add(3 * time.Second)
	for {
		k.mu.Lock()
		named := false
		for _, ev := range k.events {
			if ev.Kind == model.EvChat && ev.Chat.ID == "42@chat.agent" && ev.Chat.Title == "Team" && ev.Chat.Unread == -1 {
				named = true
			}
		}
		k.mu.Unlock()
		if named {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no group name")
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
