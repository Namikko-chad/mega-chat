package slack

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"megachat/core/model"
	"megachat/core/provider"
)

// selfDM opens the direct chat with oneself (Slack's notes to self) when
// the account's conversations lack it; ok false when it is there already
// or cannot be opened.
func (s *session) selfDM(ctx context.Context, convs []wChannel) (wChannel, bool) {
	me := s.params["user_id"]
	if me == "" || slices.ContainsFunc(convs, func(c wChannel) bool { return c.IsIM && c.User == me }) {
		return wChannel{}, false
	}
	var r struct {
		Channel wChannel `json:"channel"`
	}
	if err := s.call(ctx, "conversations.open", url.Values{"users": {me}, "return_im": {"true"}}, &r); err != nil || r.Channel.ID == "" {
		if err != nil {
			s.env.Logf("info", "slack: the chat with yourself: %v", err)
		}
		return wChannel{}, false
	}
	c := r.Channel
	c.IsIM, c.User = true, me
	s.mu.Lock()
	s.chats[c.ID] = c
	s.mu.Unlock()
	return c, true
}

// ChatActions: Slack keeps mute and pins in the user's preferences, which
// only the web client writes; leaving works everywhere.
func (s *session) ChatActions() []model.ChatAction {
	return []model.ChatAction{model.ChatLeave}
}

// ManageChat leaves a channel (conversations.leave) or closes a group DM
// (conversations.close: a group DM cannot be left).
func (s *session) ManageChat(ctx context.Context, chatID string, action model.ChatAction) error {
	if action != model.ChatLeave {
		return provider.ErrUnsupported
	}
	s.mu.Lock()
	c, ok := s.chats[chatID]
	s.mu.Unlock()
	method := "conversations.leave"
	switch {
	case ok && c.IsIM:
		return fmt.Errorf("%w: a direct chat cannot be left", provider.ErrUnsupported)
	case ok && c.IsMPIM:
		method = "conversations.close"
	}
	if err := s.call(ctx, method, url.Values{"channel": {chatID}}, nil); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.chats, chatID)
	s.mu.Unlock()
	return nil
}
