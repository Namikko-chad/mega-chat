package vkteams

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// actionServer records what the file, edit, delete and forward calls send.
type actionServer struct {
	mu       sync.Mutex
	host     string // the upload host init answers with
	ranges   []string
	uploaded []byte
	sendIM   []url.Values
	sent     []map[string]any
	deleted  []map[string]any
	history  []map[string]any
}

func (a *actionServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	rapi := func(results any) {
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 20000}, "results": results})
	}
	params := func() map[string]any {
		var b struct {
			Params map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&b)
		return b.Params
	}
	switch r.URL.Path {
	case "/api/v1/files/init":
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 200},
			"result": map[string]any{"host": a.host, "url": "/up/1"}})
	case "/up/1":
		body, _ := io.ReadAll(r.Body)
		a.uploaded = append(a.uploaded, body...)
		a.ranges = append(a.ranges, r.Header.Get("Content-Range"))
		if len(a.uploaded) < 1<<20+5 { // the test file's size: more ranges to come
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{"code": 200},
			"result": map[string]any{"static_url": "https://files.example/get/f1", "fileid": "f1"}})
	case "/api/v1/wim/im/sendIM":
		r.ParseForm()
		a.sendIM = append(a.sendIM, r.PostForm)
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"statusCode": 200, "data": map[string]any{"histMsgId": 900}}})
	case "/api/v1/rapi/message/send":
		a.sent = append(a.sent, params())
		rapi(map[string]any{"msgId": 901})
	case "/api/v1/rapi/delMsgBatch":
		a.deleted = append(a.deleted, params())
		rapi(map[string]any{})
	case "/api/v1/rapi/getHistory":
		rapi(map[string]any{"messages": a.history})
	default:
		http.NotFound(w, r)
	}
}

func actionSession(t *testing.T, a *actionServer) *session {
	t.Helper()
	ts := httptest.NewServer(a)
	t.Cleanup(ts.Close)
	u, _ := url.Parse(ts.URL)
	a.host = u.Host
	return &session{
		acct:   provider.Account{Params: map[string]string{"dev_id": "dev"}},
		srv:    &server{scheme: "http", host: u.Host, version: "v1", http: ts.Client()},
		aimsid: "sid", online: true, self: model.User{ID: "me@x"},
		people: map[string]string{}, dialogs: map[string]*dialog{}, threadOf: map[string]string{},
		replyHome: map[string]string{}, replies: map[string]int{}, notThread: map[string]bool{},
	}
}

func TestSendFilesUploadsInRangesWithCaption(t *testing.T) {
	a := &actionServer{}
	s := actionSession(t, a)
	data := strings.Repeat("x", 1<<20+5)
	_, err := s.SendFiles(context.Background(), "bob@x", model.Outgoing{Text: "look",
		Files: []model.OutFile{{Name: "a.bin", MIME: "application/octet-stream", Data: []byte(data)}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.ranges) != 2 || a.ranges[0] != "bytes 0-1048575/1048581" || a.ranges[1] != "bytes 1048576-1048580/1048581" {
		t.Fatalf("ranges: %v", a.ranges)
	}
	if string(a.uploaded) != data {
		t.Fatal("the uploaded bytes differ")
	}
	if len(a.sendIM) != 1 || !strings.Contains(a.sendIM[0].Get("parts"), `"caption":"look"`) ||
		!strings.Contains(a.sendIM[0].Get("parts"), "https://files.example/get/f1") {
		t.Fatalf("sendIM: %v", a.sendIM)
	}
}

func TestUploadRefusesForeignHost(t *testing.T) {
	a := &actionServer{}
	s := actionSession(t, a)
	a.host = "evil.example"
	_, err := s.SendFiles(context.Background(), "bob@x", model.Outgoing{
		Files: []model.OutFile{{Name: "a", Data: []byte("x")}}})
	if err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("got %v", err)
	}
	if len(a.sendIM)+len(a.sent) != 0 {
		t.Fatal("nothing may be sent")
	}
}

func TestEditKeepsFileAndQuote(t *testing.T) {
	a := &actionServer{history: []map[string]any{{
		"msgId": 7, "time": 1, "outgoing": true,
		"parts":       []any{map[string]any{"mediaType": "quote", "text": "old", "sn": "bob@x", "msgId": 3}},
		"filesharing": []any{map[string]any{"id": "f", "original_url": "https://files.example/get/f"}},
	}}}
	s := actionSession(t, a)
	if _, err := s.Edit(context.Background(), "bob@x", "7", "new caption"); err != nil {
		t.Fatal(err)
	}
	f := a.sendIM[0]
	p := f.Get("parts")
	if f.Get("updateMsgId") != "7" || !strings.Contains(p, `"mediaType":"quote"`) ||
		!strings.Contains(p, `"caption":"new caption"`) || !strings.Contains(p, "get/f") {
		t.Fatalf("edit: %v", f)
	}
}

func TestEditPlainAndRefusals(t *testing.T) {
	a := &actionServer{history: []map[string]any{{"msgId": 7, "time": 1, "text": "old", "outgoing": true}}}
	s := actionSession(t, a)
	if _, err := s.Edit(context.Background(), "bob@x", "7", "new"); err != nil {
		t.Fatal(err)
	}
	if f := a.sendIM[0]; f.Get("message") != "new" || f.Get("parts") != "" {
		t.Fatalf("edit: %v", f)
	}
	a.history = []map[string]any{{"msgId": 7, "parts": []any{map[string]any{"mediaType": "forward", "text": "x"}}}}
	if _, err := s.Edit(context.Background(), "bob@x", "7", "new"); err == nil {
		t.Fatal("a forwarded message must not be edited")
	}
	if len(a.sendIM) != 1 {
		t.Fatal("a refused edit must send nothing")
	}
}

func TestDeleteGoesToReplyThread(t *testing.T) {
	a := &actionServer{}
	s := actionSession(t, a)
	s.replyHome[msgKey("g@chat.agent", "5")] = "t@chat.agent"
	if err := s.Delete(context.Background(), "g@chat.agent", []string{"5"}, false); err != nil {
		t.Fatal(err)
	}
	d := a.deleted[0]
	if d["sn"] != "t@chat.agent" || d["shared"] != false {
		t.Fatalf("delete: %v", d)
	}
}

func TestForwardSendsForwardParts(t *testing.T) {
	a := &actionServer{history: []map[string]any{{
		"msgId": 7, "time": 11, "text": "hello", "chat": map[string]any{"sender": "bob@x"},
	}}}
	s := actionSession(t, a)
	if _, err := s.Forward(context.Background(), "g@chat.agent", []string{"7"}, "ann@x"); err != nil {
		t.Fatal(err)
	}
	f := a.sendIM[0]
	var parts []map[string]any
	if err := json.Unmarshal([]byte(f.Get("parts")), &parts); err != nil {
		t.Fatal(err)
	}
	if f.Get("t") != "ann@x" || len(parts) != 1 || parts[0]["mediaType"] != "forward" ||
		parts[0]["sn"] != "bob@x" || parts[0]["text"] != "hello" {
		t.Fatalf("forward: %v", f)
	}
}
