package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

func TestText(t *testing.T) {
	name := func(id string) string { return map[string]string{"U1": "ann"}[id] }
	got := plainText("hi <@U1>, see <#C2|general> and <https://x.io|docs> &amp; <!here>", name)
	if got != "hi @ann, see #general and docs (https://x.io) & @here" {
		t.Fatal(got)
	}
	if escapeText("a<b>&c") != "a&lt;b&gt;&amp;c" {
		t.Fatal("escape")
	}
	if mpimTitle("mpdm-ann--bob--carl-1") != "ann, bob, carl" {
		t.Fatal(mpimTitle("mpdm-ann--bob--carl-1"))
	}
	if ts := tsTime("1700000000.000100"); ts.Unix() != 1700000000 || ts.Nanosecond() != 100000 {
		t.Fatal(ts)
	}
	if u, err := workspaceURL("acme"); err != nil || u != "https://acme.slack.com/" {
		t.Fatal(u, err)
	}
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

func (s *sink) wait(t *testing.T, kind model.EventKind) model.Event {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		for _, ev := range s.events {
			if ev.Kind == kind {
				s.mu.Unlock()
				return ev
			}
		}
		s.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no event %v", kind)
	return model.Event{}
}

// fakeSlack answers the Web API and an RTM websocket.
// uploads is what fakeSlack got for a file upload.
var uploads struct {
	sync.Mutex
	body, mime string
	complete   url.Values
}

func fakeSlack(t *testing.T) *httptest.Server {
	var srv *httptest.Server
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, v any) { json.NewEncoder(w).Encode(v) }
	auth := func(r *http.Request) bool {
		return r.Header.Get("Authorization") == "Bearer xoxc-1" && strings.Contains(r.Header.Get("Cookie"), "d=xoxd-1")
	}
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			reply(w, map[string]any{"ok": false, "error": "invalid_auth"})
			return
		}
		r.ParseForm()
		switch strings.TrimPrefix(r.URL.Path, "/api/") {
		case "auth.test":
			reply(w, map[string]any{"ok": true, "team": "Acme", "user": "me", "team_id": "T1", "user_id": "U0"})
		case "users.info":
			names := map[string]string{"U0": "Me Myself", "U1": "Ann"}
			reply(w, map[string]any{"ok": true, "user": map[string]any{"id": r.Form.Get("user"), "name": "ann", "profile": map[string]any{
				"display_name": names[r.Form.Get("user")], "title": "Engineer", "email": "ann@acme.test", "status_text": "On holiday",
			}}})
		case "users.getPresence":
			reply(w, map[string]any{"ok": true, "presence": "active"})
		case "conversations.info":
			chans := map[string]map[string]any{
				"D1": {"id": "D1", "is_im": true, "user": "U1"},
				"C1": {"id": "C1", "name": "general", "purpose": map[string]any{"value": "Talk <https://acme.test|here>"}, "topic": map[string]any{"value": "Q4"}, "num_members": 2},
			}
			reply(w, map[string]any{"ok": true, "channel": chans[r.Form.Get("channel")]})
		case "conversations.members":
			reply(w, map[string]any{"ok": true, "members": []string{"U0", "U1"}})
		case "users.conversations":
			reply(w, map[string]any{"ok": true, "channels": []map[string]any{
				{"id": "D1", "is_im": true, "user": "U1"}, {"id": "C1", "name": "general"},
			}})
		case "client.counts":
			reply(w, map[string]any{"ok": true, "ims": []map[string]any{{"id": "D1", "latest": "1700000100.000000", "has_unreads": true, "mention_count": 2}},
				"channels": []map[string]any{{"id": "C1", "latest": "1700000000.000000"}}})
		case "conversations.history":
			if r.Form.Get("channel") == "C9" {
				w.Write([]byte(appHistory))
				return
			}
			msgs := []map[string]any{
				{"type": "message", "ts": "1700000100.000000", "user": "U1", "text": "hello <@U0>"},
				{"type": "message", "ts": "1700000090.000000", "user": "U1", "text": "in thread", "thread_ts": "1700000050.000000"},
				{"type": "message", "ts": "1700000080.000000", "user": "U0", "text": "earlier", "files": []map[string]any{{"name": "a.png", "mimetype": "image/png", "size": 3}}},
			}
			if r.Form.Get("limit") == "1" {
				msgs = msgs[:1]
			}
			reply(w, map[string]any{"ok": true, "messages": msgs, "has_more": true})
		case "chat.postMessage":
			reply(w, map[string]any{"ok": true, "ts": "1700000200.000000", "message": map[string]any{"ts": "1700000200.000000", "user": "U0", "text": r.Form.Get("text")}})
		case "chat.getPermalink":
			reply(w, map[string]any{"ok": true, "permalink": "https://acme.slack.com/archives/" + r.Form.Get("channel") + "/p" + strings.ReplaceAll(r.Form.Get("message_ts"), ".", "")})
		case "chat.update":
			ok := r.Form.Get("ts") == "1700000200.000000" && r.Form.Get("text") == "a &lt;b&gt;"
			reply(w, map[string]any{"ok": ok, "ts": r.Form.Get("ts"), "error": "cant_update_message"})
		case "chat.delete":
			reply(w, map[string]any{"ok": r.Form.Get("ts") == "1700000200.000000" && r.Form.Get("channel") != "", "error": "message_not_found"})
		case "conversations.mark":
			reply(w, map[string]any{"ok": r.Form.Get("ts") != ""})
		case "files.getUploadURLExternal":
			reply(w, map[string]any{"ok": true, "upload_url": srv.URL + "/upload/F1?len=" + r.Form.Get("length"), "file_id": "F1"})
		case "files.completeUploadExternal":
			uploads.Lock()
			uploads.complete = r.Form
			uploads.Unlock()
			reply(w, map[string]any{"ok": true, "files": []map[string]any{{"id": "F1"}}})
		case "rtm.connect":
			reply(w, map[string]any{"ok": true, "url": "ws" + strings.TrimPrefix(srv.URL, "http") + "/rtm"})
		default:
			reply(w, map[string]any{"ok": false, "error": "unknown_method"})
		}
	})
	mux.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		uploads.Lock()
		uploads.body, uploads.mime = string(body), r.Header.Get("Content-Type")
		uploads.Unlock()
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/rtm", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Cookie"), "d=xoxd-1") {
			http.Error(w, "no cookie", http.StatusForbidden)
			return
		}
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.CloseNow()
		ctx := r.Context()
		for _, ev := range []string{
			`{"type":"hello"}`,
			`{"type":"message","channel":"D1","user":"U1","ts":"1700000300.000000","text":"live"}`,
			`{"type":"message","subtype":"message_changed","channel":"D1","message":{"type":"message","user":"U1","ts":"1700000300.000000","text":"live!","edited":{"ts":"1700000301.000000"}}}`,
			`{"type":"im_marked","channel":"D1","ts":"1700000300.000000","unread_count_display":0}`,
		} {
			c.Write(ctx, websocket.MessageText, []byte(ev))
		}
		for { // until the client goes away
			_, data, err := c.Read(ctx)
			if err != nil {
				return
			}
			if strings.Contains(string(data), `"presence_sub"`) && strings.Contains(string(data), `"U1"`) {
				c.Write(ctx, websocket.MessageText, []byte(`{"type":"presence_change","user":"U1","presence":"active"}`))
			}
		}
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSessionMode(t *testing.T) {
	srv := fakeSlack(t)
	p := New()
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	env := provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()}
	sess, err := p.Open(provider.Account{ID: "a", Params: params}, env)
	if err != nil {
		t.Fatal(err)
	}
	m := sess.(provider.Messenger)
	ctx := context.Background()

	chats, err := m.Chats(ctx, 10)
	if err != nil || len(chats) != 2 {
		t.Fatalf("chats: %+v %v", chats, err)
	}
	if c := chats[0]; c.ID != "D1" || c.Title != "Ann" || c.Kind != model.ChatDirect || c.Unread != 2 || c.Last == nil || c.Last.Text != "hello @Me Myself" {
		t.Fatalf("first chat: %+v", c)
	}
	if chats[1].Title != "#general" {
		t.Fatalf("channel: %+v", chats[1])
	}
	msgs, more, err := m.History(ctx, "D1", "", 50)
	if err != nil || !more || len(msgs) != 2 {
		t.Fatalf("history (thread reply skipped): %+v %v", msgs, err)
	}
	if !msgs[1].Outgoing || len(msgs[1].Attachments) != 1 || msgs[1].Attachments[0].Kind != model.AttachImage {
		t.Fatalf("own message: %+v", msgs[1])
	}
	sent, err := m.Send(ctx, "D1", model.Outgoing{Text: "a<b"})
	if err != nil || sent.ID != "1700000200.000000" || sent.Text != "a<b" || !sent.Outgoing {
		t.Fatalf("send: %+v %v", sent, err)
	}
	if err := m.MarkRead(ctx, "D1", ""); err != nil {
		t.Fatal(err)
	}
	quoted, err := m.Send(ctx, "D1", model.Outgoing{Text: "yes", ReplyTo: "1700000100.000000", Quote: "hello"})
	if err != nil || quoted.Text != "yes" || quoted.Quote != "hello" || len(quoted.Format) != 0 {
		t.Fatalf("quote: %+v %v", quoted, err)
	}
	fwd, err := sess.(provider.Forwarder).Forward(ctx, "D1", []string{"1700000100.000000"}, "C1")
	if err != nil || len(fwd) != 1 || !fwd[0].Forwarded || fwd[0].Text != "https://acme.slack.com/archives/D1/p1700000100000000" {
		t.Fatalf("forward: %+v %v", fwd, err)
	}

	files, err := sess.(provider.FileSender).SendFiles(ctx, "C1", model.Outgoing{Text: "see <this>", Thread: "1700000050.000000",
		Files: []model.OutFile{{Name: "a.txt", MIME: "text/plain", Data: []byte("hello")}}})
	uploads.Lock()
	if err != nil || len(files) != 0 || uploads.body != "hello" || uploads.mime != "text/plain" ||
		uploads.complete.Get("channel_id") != "C1" || uploads.complete.Get("thread_ts") != "1700000050.000000" ||
		uploads.complete.Get("initial_comment") != "see &lt;this&gt;" || uploads.complete.Get("files") != `[{"id":"F1","title":"a.txt"}]` {
		t.Fatalf("files: %v %v %+v", files, err, uploads.complete)
	}
	uploads.Unlock()

	k := &sink{connected: make(chan model.User, 1)}
	rctx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- sess.Run(rctx, k) }()
	if u := <-k.connected; u.Name != "Me Myself" {
		t.Fatalf("self: %+v", u)
	}
	if ev := k.wait(t, model.EvMessageNew); ev.Message.Text != "live" || ev.Message.ID != "1700000300.000000" ||
		ev.Message.Time.Unix() != 1700000300 || ev.Message.Sender.ID != "U1" || ev.Chat == nil || ev.Chat.Title != "Ann" {
		t.Fatalf("new: %+v", ev.Message)
	}
	if ev := k.wait(t, model.EvMessageEdit); ev.Message.Text != "live!" || ev.Message.ID != "1700000300.000000" || ev.Message.Edited.IsZero() {
		t.Fatalf("edit: %+v", ev.Message)
	}
	if ev := k.wait(t, model.EvRead); ev.Chat.ID != "D1" || ev.Chat.Unread != 0 {
		t.Fatalf("read: %+v", ev.Chat)
	}
	if ev := k.wait(t, model.EvPresence); ev.Chat.ID != "D1" || !ev.Online {
		t.Fatalf("presence: %+v %v", ev.Chat, ev.Online)
	}
	cancel()
	<-done

	bad := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-2", "cookie": "xoxd-1"}
	s2, _ := p.Open(provider.Account{ID: "b", Params: bad}, env)
	if err := s2.Run(ctx, k); !isAuthErr(err) {
		t.Fatalf("bad token: %v", err)
	}
}

func isAuthErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), provider.ErrAuth.Error())
}

func TestMentions(t *testing.T) {
	for text, want := range map[string]bool{
		"hi <@U1>":           true,
		"hi <@U1|ann> ok":    true,
		"hi <@U2>":           false,
		"<!here> deploy":     true,
		"<!channel|channel>": true,
		"plain @U1":          false,
	} {
		if got := mentions(text, "U1"); got != want {
			t.Errorf("mentions(%q) = %v", text, got)
		}
	}
}

func TestSharedMessage(t *testing.T) {
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	s := sess.(*session)
	link := "https://acme.slack.com/archives/C1/p1700000100000000"
	m := s.convert(context.Background(), "C2", wMessage{TS: "1700000200.000000", User: "U1", Text: "<" + link + ">", Attachments: []wAttachment{
		{IsMsgUnfurl: true, AuthorSubname: "Ann", AuthorName: "ann", Text: "the original", FromURL: link, Files: []wFile{{Name: "a.png", Mimetype: "image/png"}}},
		{Text: "an ordinary unfurl"},
	}})
	if !m.Forwarded || m.ForwardedFrom != "Ann" || m.Text != "the original" || len(m.Attachments) != 1 {
		t.Fatalf("shared: %+v", m)
	}
}

func TestProfile(t *testing.T) {
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	pr := sess.(provider.Profiles)
	ctx := context.Background()
	p, err := pr.Profile(ctx, "D1")
	if err != nil || p.Title != "Ann" || p.Username != "ann" || p.Position != "Engineer" || p.Email != "ann@acme.test" || p.About != "On holiday" || p.Link != srv.URL+"/team/U1" {
		t.Fatalf("direct: %+v %v", p, err)
	}
	p, err = pr.Profile(ctx, "C1")
	if err != nil || p.Title != "#general" || p.Members != 2 || len(p.MemberList) != 2 || p.MemberList[1].Name != "Ann" || p.Link != srv.URL+"/archives/C1" {
		t.Fatalf("channel: %+v %v", p, err)
	}
	if !strings.Contains(p.About, "https://acme.test") || !strings.Contains(p.About, "Q4") {
		t.Fatalf("about: %q", p.About)
	}
}

func TestEdit(t *testing.T) {
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	ed := sess.(provider.Editor)
	m, err := ed.Edit(context.Background(), "D1", "1700000200.000000", "a <b>")
	if err != nil || m.ID != "1700000200.000000" || m.Text != "a <b>" || m.Edited.IsZero() {
		t.Fatalf("edit: %+v %v", m, err)
	}
	if _, err := ed.Edit(context.Background(), "D1", "1700000100.000000", "x"); err == nil {
		t.Fatal("no error from Slack")
	}
}

func TestDelete(t *testing.T) {
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	d := sess.(provider.Deleter)
	if d.DeletesForMe() {
		t.Fatal("Slack deletes for everyone only")
	}
	if err := d.Delete(context.Background(), "D1", []string{"1700000200.000000"}, true); err != nil {
		t.Fatal(err)
	}
	if err := d.Delete(context.Background(), "D1", []string{"1700000200.000000"}, false); !errors.Is(err, provider.ErrUnsupported) {
		t.Fatalf("for me: %v", err)
	}
	if err := d.Delete(context.Background(), "D1", []string{"1700000100.000000"}, true); err == nil {
		t.Fatal("no error from Slack")
	}
}

// appHistory is a channel with an app's Block Kit message, a webhook's
// legacy attachment and a user's ordinary rich text message.
const appHistory = `{"ok":true,"has_more":false,"messages":[
{"type":"message","ts":"1700000300.000000","bot_id":"B1","username":"SwitchCase","text":"New Website Form Submission","blocks":[
 {"type":"header","text":{"type":"plain_text","text":"New Website Form Submission","emoji":true}},
 {"type":"section","fields":[{"type":"mrkdwn","text":"*Name:*\nJohn Doe"},{"type":"mrkdwn","text":"*Email:*\n<mailto:john@x.io|john@x.io>"}]},
 {"type":"divider"},
 {"type":"section","text":{"type":"mrkdwn","text":"*Comment:*\nCall me &amp; <@U1>"},"accessory":{"type":"image","image_url":"https://x.io/a.png","alt_text":"a"}},
 {"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Reply"},"style":"primary","value":"r"}]},
 {"type":"context","elements":[{"type":"image","image_url":"https://x.io/i.png","alt_text":"i"},{"type":"mrkdwn","text":"via *site*"}]}
]},
{"type":"message","subtype":"bot_message","ts":"1700000200.000000","bot_id":"B2","username":"hook","text":"","attachments":[
 {"fallback":"Form: Jane","pretext":"New lead","title":"Jane Roe","title_link":"https://crm.test/1","fields":[{"title":"Specialty","value":"Cardiology","short":true},{"title":"Phone","value":"123","short":true}]},
 {"fallback":"only fallback"}
]},
{"type":"message","ts":"1700000100.000000","user":"U1","text":"hi <@U0>, see <https://x.io|docs> *now*","blocks":[
 {"type":"rich_text","elements":[{"type":"rich_text_section","elements":[{"type":"text","text":"hi "},{"type":"user","user_id":"U0"},{"type":"text","text":", see "},{"type":"link","url":"https://x.io","text":"docs"},{"type":"text","text":" "},{"type":"text","text":"now","style":{"bold":true}}]}]}
]}
]}`

func TestAppMessages(t *testing.T) {
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	msgs, _, err := sess.(provider.Messenger).History(context.Background(), "C9", "", 50)
	if err != nil || len(msgs) != 3 {
		t.Fatalf("history: %+v %v", msgs, err)
	}
	want := "New Website Form Submission\nName:\nJohn Doe\nEmail:\njohn@x.io (mailto:john@x.io)\n\nComment:\nCall me & @Ann\nvia site"
	bold := 0
	for _, f := range msgs[0].Format {
		if f.Kind == model.SpanBold {
			bold++
		}
	}
	if m := msgs[0]; m.Text != want || m.Sender.Name != "SwitchCase" || bold != 4 {
		t.Fatalf("blocks: %q", m.Text)
	}
	want = "New lead\nJane Roe (https://crm.test/1)\nSpecialty: Cardiology\nPhone: 123\nonly fallback"
	if m := msgs[1]; m.Text != want {
		t.Fatalf("attachments: %q", m.Text)
	}
	if m := msgs[2]; m.Text != "hi @Me Myself, see docs (https://x.io) now" || !m.Mentioned || len(m.Format) != 1 || m.Format[0].Kind != model.SpanBold {
		t.Fatalf("user message: %+v", m)
	}
	// The rich text itself renders the same way when there is no text.
	bs := parseBlocks(json.RawMessage(`[{"type":"rich_text","elements":[{"type":"rich_text_section","elements":[{"type":"text","text":"a "},{"type":"text","text":"b","style":{"bold":true}}]},{"type":"rich_text_list","style":"ordered","elements":[{"type":"rich_text_section","elements":[{"type":"text","text":"one"}]},{"type":"rich_text_section","elements":[{"type":"emoji","name":"+1"}]}]},{"type":"rich_text_quote","elements":[{"type":"text","text":"q"}]},{"type":"rich_text_preformatted","elements":[{"type":"text","text":"x := 1"}]}]}]`))
	if got := blocksText(bs, func(string) string { return "" }); got != "a *b*\n1. one\n2. 👍\n> q\n```\nx := 1\n```" {
		t.Fatalf("rich text: %q", got)
	}
}

func TestMentionMembers(t *testing.T) {
	o := model.Outgoing{Text: "hi @Ann <3", Quote: "a<b", Mentions: []model.Mention{{Offset: 3, Length: 4, User: model.User{ID: "U1"}}}}
	if got := outText(o); got != "&gt; a&lt;b\nhi <@U1> &lt;3" {
		t.Fatalf("%q", got)
	}
	srv := fakeSlack(t)
	params := map[string]string{"method": methodSession, "workspace_url": srv.URL + "/", "token": "xoxc-1", "cookie": "xoxd-1", "user_id": "U0"}
	sess, err := New().Open(provider.Account{ID: "a", Params: params}, provider.Env{State: provider.NewMemKV(), HTTP: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	us, err := sess.(provider.Mentioner).Members(context.Background(), "C1", "an", 10)
	if err != nil || len(us) != 1 || us[0].ID != "U1" || !us[0].Online {
		t.Fatalf("members: %+v %v", us, err)
	}
	// What RTM reports wins over the answer kept from users.getPresence.
	sess.(*session).setPresence([]string{"U1"}, false)
	if us, _ = sess.(provider.Mentioner).Members(context.Background(), "C1", "an", 10); len(us) != 1 || us[0].Online {
		t.Fatalf("went away: %+v", us)
	}
	if us, _ = sess.(provider.Mentioner).Members(context.Background(), "C1", "zz", 10); len(us) != 0 {
		t.Fatalf("no match: %+v", us)
	}
}
