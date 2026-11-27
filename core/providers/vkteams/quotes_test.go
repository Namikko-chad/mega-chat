package vkteams

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

func quoteEnv() provider.Env {
	return provider.Env{State: provider.NewMemKV()}
}

func TestQuoteIsNative(t *testing.T) {
	a := &actionServer{history: []map[string]any{{
		"msgId": 7, "time": 11, "text": "a long message", "chat": map[string]any{"sender": "bob@x"},
	}}}
	s := actionSession(t, a)
	s.env = quoteEnv()
	m, err := s.Send(context.Background(), "g@chat.agent", model.Outgoing{Text: "agreed", ReplyTo: "7", Quote: "long"})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.sendIM) != 1 || len(a.sent) != 0 {
		t.Fatalf("sendIM %d, message/send %d", len(a.sendIM), len(a.sent))
	}
	var parts []map[string]any
	if err := json.Unmarshal([]byte(a.sendIM[0].Get("parts")), &parts); err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 || parts[0]["mediaType"] != "quote" || parts[0]["text"] != "long" || parts[0]["sn"] != "bob@x" ||
		parts[1]["mediaType"] != "text" || parts[1]["text"] != "agreed" {
		t.Fatalf("parts: %v", parts)
	}
	if m.Text != "agreed" || m.Quote != "long" || m.ReplyTo != "7" {
		t.Fatalf("message: %+v", m)
	}
}

func TestQuoteFallsBackToLines(t *testing.T) {
	a := &actionServer{} // the quoted message cannot be found
	s := actionSession(t, a)
	s.env = quoteEnv()
	if _, err := s.Send(context.Background(), "bob@x", model.Outgoing{Text: "agreed", ReplyTo: "7", Quote: "long"}); err != nil {
		t.Fatal(err)
	}
	if len(a.sendIM) != 0 || len(a.sent) != 1 {
		t.Fatalf("sendIM %d, message/send %d", len(a.sendIM), len(a.sent))
	}
	main := a.sent[0]["parts"].(map[string]any)["mainPart"].(map[string]any)
	if got := main["text"].(map[string]any)["plain"]; got != "> long\nagreed" {
		t.Fatalf("sent: %v", got)
	}
}
