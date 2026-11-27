package telegram

import (
	"context"
	"strings"
	"testing"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

type dir struct{}

func (dir) user(key string) (model.User, bool) {
	if key == "u7" {
		return model.User{ID: "u7", Name: "Ann"}, true
	}
	return model.User{}, false
}
func (dir) title(key string) string { return map[string]string{"c9": "News"}[key] }
func (dir) selfUser() model.User    { return model.User{ID: "u1", Name: "Me"} }

func TestKeys(t *testing.T) {
	for _, c := range []struct {
		p   tg.PeerClass
		key string
	}{{&tg.PeerUser{UserID: 7}, "u7"}, {&tg.PeerChat{ChatID: 8}, "g8"}, {&tg.PeerChannel{ChannelID: 9}, "c9"}} {
		if k := peerKey(c.p); k != c.key {
			t.Fatalf("%T: %s", c.p, k)
		}
		kind, id, ok := parseKey(c.key)
		if !ok || kind != c.key[0] || id == 0 {
			t.Fatalf("parse %s", c.key)
		}
	}
	if _, _, ok := parseKey("x1"); ok {
		t.Fatal("bad kind accepted")
	}
}

func TestConvertMessage(t *testing.T) {
	// Private incoming: no FromID, the peer is the sender.
	m, ok := convertMessage(dir{}, &tg.Message{ID: 5, PeerID: &tg.PeerUser{UserID: 7}, Date: 100, Message: "hi",
		ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 4}})
	if !ok || m.ChatID != "u7" || m.Sender.Name != "Ann" || m.ReplyTo != "4" || m.Seq != 5 || m.Time.Unix() != 100 {
		t.Fatalf("%+v", m)
	}
	// Outgoing.
	m, _ = convertMessage(dir{}, &tg.Message{ID: 6, Out: true, PeerID: &tg.PeerUser{UserID: 7}})
	if !m.Outgoing || m.Sender.Name != "Me" || m.Group != "" {
		t.Fatalf("out: %+v", m)
	}
	// A part of an album.
	m, _ = convertMessage(dir{}, &tg.Message{ID: 7, PeerID: &tg.PeerUser{UserID: 7}, GroupedID: 1234567890123})
	if m.Group != "1234567890123" {
		t.Fatalf("album: %+v", m)
	}
	// Channel post with a voice note.
	m, _ = convertMessage(dir{}, &tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 9},
		Media: &tg.MessageMediaDocument{Voice: true, Document: &tg.Document{MimeType: "audio/ogg", Size: 10}}})
	if m.Sender.Name != "News" || len(m.Attachments) != 1 || m.Attachments[0].Kind != model.AttachVoice {
		t.Fatalf("channel: %+v", m)
	}
	// Service message.
	m, _ = convertMessage(dir{}, &tg.MessageService{ID: 2, PeerID: &tg.PeerChat{ChatID: 8}, FromID: &tg.PeerUser{UserID: 7},
		Action: &tg.MessageActionChatEditTitle{Title: "X"}})
	if !m.Service || m.Text != "changed the title to «X»" || m.Sender.Name != "Ann" {
		t.Fatalf("service: %+v", m)
	}
	// Forwarded from a known user, and from a hidden one.
	fw := &tg.Message{ID: 8, PeerID: &tg.PeerUser{UserID: 7}, Message: "x"}
	fw.SetFwdFrom(tg.MessageFwdHeader{FromID: &tg.PeerUser{UserID: 7}})
	if m, _ = convertMessage(dir{}, fw); !m.Forwarded || m.ForwardedFrom != "Ann" {
		t.Fatalf("forwarded: %+v", m)
	}
	fw.SetFwdFrom(tg.MessageFwdHeader{FromName: "Hidden"})
	if m, _ = convertMessage(dir{}, fw); !m.Forwarded || m.ForwardedFrom != "Hidden" {
		t.Fatalf("forwarded, hidden: %+v", m)
	}
	// A reply quoting a fragment.
	rh := &tg.MessageReplyHeader{ReplyToMsgID: 4}
	rh.SetQuoteText("frag")
	if m, _ = convertMessage(dir{}, &tg.Message{ID: 9, PeerID: &tg.PeerUser{UserID: 7}, ReplyTo: rh}); m.ReplyTo != "4" || m.Quote != "frag" || m.Forwarded {
		t.Fatalf("quote: %+v", m)
	}
	if _, ok := convertMessage(dir{}, &tg.MessageEmpty{ID: 3}); ok {
		t.Fatal("empty message converted")
	}
}

func TestNormPhone(t *testing.T) {
	if p := normPhone(" +7 (999) 123-45-67 "); p != "+79991234567" {
		t.Fatal(p)
	}
	if normPhone("abc") != "" {
		t.Fatal("no digits")
	}
}

func TestExtractCode(t *testing.T) {
	for in, want := range map[string]string{
		"12345":   "12345",
		" 12 345": "12345",
		"123-45":  "12345",
		"Login code: 54321. Do not give this code to anyone, even if they say they are from Telegram!": "54321",
		"Telegram code 98765\n\nafYeJbBf-dQ": "98765",
		"afYeJbBf-dQ":                        "",
		"code 12":                            "",
	} {
		if got := extractCode(in); got != want {
			t.Errorf("extractCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCodeStepOffersResend(t *testing.T) {
	l := &login{phone: "+7999"}
	l.applySent(&tg.AuthSentCode{Type: &tg.AuthSentCodeTypeApp{Length: 5}, PhoneCodeHash: "h"})
	st := l.codeStep("")
	if st.Fields[0].Optional || !strings.Contains(st.Message, "«Telegram»") || !strings.Contains(st.Fields[0].Help, "5 digits") {
		t.Fatalf("no next type: %+v", st)
	}
	s := &tg.AuthSentCode{Type: &tg.AuthSentCodeTypeApp{Length: 5}, PhoneCodeHash: "h"}
	s.SetNextType(&tg.AuthCodeTypeSMS{})
	l.applySent(s)
	st = l.codeStep("")
	if !st.Fields[0].Optional || !strings.Contains(st.Message, "again by SMS") {
		t.Fatalf("with SMS next: %+v", st)
	}
}

func TestQRLoginAsksForAppCredentials(t *testing.T) {
	DefaultAppID, DefaultAppHash = "", ""
	p := New()
	if m := p.Info().Methods; len(m) != 2 || m[0].ID != methodQR {
		t.Fatalf("methods: %+v", m)
	}
	l, _ := p.Login(context.Background(), methodQR, nil, provider.Env{})
	st, err := l.Next(context.Background(), nil)
	if err != nil || st.QR != "" || len(st.Fields) != 2 || st.Fields[0].Key != "api_id" {
		t.Fatalf("first step: %+v %v", st, err)
	}
	st, err = l.Next(context.Background(), map[string]string{"api_id": "x", "api_hash": "y"})
	if err != nil || st.Error == "" {
		t.Fatalf("bad id: %+v %v", st, err)
	}
	l.Close()
}

func TestForumTopicOfMessage(t *testing.T) {
	peer := &tg.PeerChannel{ChannelID: 9}
	// In a topic: the header names the topic, which is not a reply.
	in := &tg.Message{ID: 20, PeerID: peer, ReplyTo: &tg.MessageReplyHeader{ForumTopic: true, ReplyToMsgID: 5}}
	if m, _ := convertMessage(dir{}, in); m.ReplyTo != "" || topicOf(in) != 5 {
		t.Fatalf("in topic: reply %q, topic %d", m.ReplyTo, topicOf(in))
	}
	// A reply inside a topic.
	re := &tg.Message{ID: 21, PeerID: peer, ReplyTo: &tg.MessageReplyHeader{ForumTopic: true, ReplyToMsgID: 20, ReplyToTopID: 5}}
	if m, _ := convertMessage(dir{}, re); m.ReplyTo != "20" || topicOf(re) != 5 {
		t.Fatalf("reply in topic: reply %q, topic %d", m.ReplyTo, topicOf(re))
	}
	// The general topic: no header, or a plain reply.
	if topicOf(&tg.Message{ID: 22, PeerID: peer}) != generalTopic ||
		topicOf(&tg.Message{ID: 23, PeerID: peer, ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 22}}) != generalTopic {
		t.Fatal("general topic")
	}
	// The message that opened a topic belongs to it.
	open := &tg.MessageService{ID: 5, PeerID: peer, Action: &tg.MessageActionTopicCreate{Title: "Design"}}
	if m, _ := convertMessage(dir{}, open); topicOf(open) != 5 || !strings.Contains(m.Text, "Design") {
		t.Fatalf("topic created: %q %d", m.Text, topicOf(open))
	}
}

func TestForumTopic(t *testing.T) {
	last := map[int]model.Message{30: {ID: "30", Seq: 30, Text: "latest"}}
	x := forumTopic(&tg.ForumTopic{ID: 5, Title: "Design", TopMessage: 30, UnreadCount: 3, Pinned: true, Date: 10, IconColor: 0x6fb9f0}, last)
	if x.ID != "5" || x.Title != "Design" || x.Unread != 3 || !x.Pinned || x.Color != 0x6fb9f0 || x.Last == nil || x.Last.InThread != "5" {
		t.Fatalf("%+v", x)
	}
}
