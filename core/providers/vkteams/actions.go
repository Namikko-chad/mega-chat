package vkteams

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

var (
	_ provider.FileSender = (*session)(nil)
	_ provider.Editor     = (*session)(nil)
	_ provider.Deleter    = (*session)(nil)
	_ provider.Forwarder  = (*session)(nil)
)

// SendFiles uploads every file first, so a failed upload sends nothing, then
// sends one message per file. The messages come back as events.
func (s *session) SendFiles(ctx context.Context, chatID string, o model.Outgoing) ([]model.Message, error) {
	if len(o.Files) == 0 {
		return nil, errors.New("vk teams: no files")
	}
	aimsid, err := s.live()
	if err != nil {
		return nil, err
	}
	to := chatID
	if o.Thread != "" {
		if to, err = s.threadTarget(ctx, aimsid, chatID, o.Thread); err != nil {
			return nil, err
		}
	}
	links := make([]string, len(o.Files))
	for i, f := range o.Files {
		if links[i], err = s.srv.uploadFile(ctx, aimsid, s.acct.Params["dev_id"], f.Name, f.Data); err != nil {
			return nil, err
		}
	}
	caption, replyTo := outText(o), o.ReplyTo
	// A caption cannot carry a reply: the reply goes as its own message.
	if replyTo != "" && caption != "" {
		_, native, err := s.sendQuoted(ctx, aimsid, chatID, to, o)
		if err != nil {
			return nil, err
		}
		if !native {
			if _, err := s.srv.sendMessage(ctx, aimsid, to, caption, replyTo); err != nil {
				return nil, err
			}
		}
		caption, replyTo = "", ""
	}
	for i, link := range links {
		switch {
		case i == 0 && caption != "":
			_, err = s.srv.sendIM(ctx, aimsid, to, "", []any{captionPart(caption, link)}, "")
		case i == 0 && replyTo != "":
			_, err = s.srv.sendMessage(ctx, aimsid, to, link, replyTo)
		default:
			_, err = s.srv.sendMessage(ctx, aimsid, to, link, "")
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// authorOf is the id of who wrote m in chat.
func (s *session) authorOf(chat string, m wMessage) string {
	switch {
	case m.Outgoing:
		return s.selfUser().ID
	case m.Chat.Sender != "":
		return m.Chat.Sender
	}
	return chat
}

// sendQuoted sends o's text under a native quote of o.Quote. It tells false,
// having sent nothing, when there is no fragment or its message cannot be
// read: the caller then sends the "> " lines.
func (s *session) sendQuoted(ctx context.Context, aimsid, chatID, to string, o model.Outgoing) (string, bool, error) {
	fragment := strings.TrimSpace(o.Quote)
	if fragment == "" || o.ReplyTo == "" {
		return "", false, nil
	}
	home := s.homeOf(chatID, o.ReplyTo)
	src, err := s.srv.getMessage(ctx, aimsid, home, o.ReplyTo)
	if err != nil {
		s.env.Logf("warn", "vk teams: quote as text: %v", err)
		return "", false, nil
	}
	parts := []any{quotePart(home, s.authorOf(home, src), src, fragment)}
	body := o
	body.Quote = ""
	if text := outText(body); text != "" {
		parts = append(parts, map[string]any{"mediaType": "text", "text": text})
	}
	id, err := s.srv.sendIM(ctx, aimsid, to, "", parts, "")
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

// Edit replaces a message's text, keeping its quotes and its file. The hub
// fills in the new text; the server's edit event brings the whole message.
func (s *session) Edit(ctx context.Context, chatID, msgID, text string) (model.Message, error) {
	aimsid, err := s.live()
	if err != nil {
		return model.Message{}, err
	}
	home := s.homeOf(chatID, msgID)
	m, err := s.srv.getMessage(ctx, aimsid, home, msgID)
	if err != nil {
		return model.Message{}, err
	}
	parts, err := editParts(m, text)
	if err != nil {
		return model.Message{}, err
	}
	if _, err := s.srv.sendIM(ctx, aimsid, home, text, parts, msgID); err != nil {
		return model.Message{}, err
	}
	return model.Message{ID: msgID, Edited: time.Now()}, nil
}

// Delete deletes messages, each from the dialog it lives in (a reply: its
// thread).
func (s *session) Delete(ctx context.Context, chatID string, msgIDs []string, forAll bool) error {
	aimsid, err := s.live()
	if err != nil {
		return err
	}
	byHome := map[string][]int64{}
	for _, id := range msgIDs {
		n, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return fmt.Errorf("vk teams: bad message id")
		}
		home := s.homeOf(chatID, id)
		byHome[home] = append(byHome[home], n)
	}
	for home, ids := range byHome {
		if err := s.srv.deleteMessages(ctx, aimsid, home, ids, forAll); err != nil {
			return err
		}
	}
	return nil
}

// DeletesForMe: the server deletes for this account too ("shared": false).
func (s *session) DeletesForMe() bool { return true }

// Forward sends the messages as one message of "forward" parts. The copy
// comes back as an event.
func (s *session) Forward(ctx context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error) {
	aimsid, err := s.live()
	if err != nil {
		return nil, err
	}
	var parts []any
	for _, id := range msgIDs {
		m, err := s.srv.getMessage(ctx, aimsid, fromChat, id)
		if err != nil {
			return nil, err
		}
		sender := s.authorOf(fromChat, m)
		for _, f := range m.Filesharing {
			if f.OriginalURL == "" { // better an error than a copy without its file
				return nil, fmt.Errorf("vk teams: message %s: the file's link is unknown", id)
			}
		}
		got := forwardParts(fromChat, sender, m, m.content(s.personName).text)
		if len(got) == 0 {
			return nil, fmt.Errorf("vk teams: message %s has nothing to forward", id)
		}
		parts = append(parts, got...)
	}
	if len(parts) == 0 {
		return nil, errors.New("vk teams: no messages to forward")
	}
	if _, err := s.srv.sendIM(ctx, aimsid, toChat, "", parts, ""); err != nil {
		return nil, err
	}
	return nil, nil
}
