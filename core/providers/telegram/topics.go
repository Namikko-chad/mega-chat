package telegram

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/gotd/td/tg"

	"megachat/core/model"
)

// --- Topics (forum groups) ---------------------------------------------------
//
// A forum is a supergroup split into topics. A topic is named by the id of
// the service message that created it; the general topic has id 1. A
// message names its topic in its reply header (forum_topic), so a topic's
// messages are the replies to that id (messages.getReplies) and sending into
// a topic replies to it.

// generalTopic is the id of a forum's general topic.
const generalTopic = 1

func (s *session) forum(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.forums[key]
}

// topicOf is the topic a message of a forum chat belongs to.
func topicOf(mc tg.MessageClass) int {
	var reply tg.MessageReplyHeaderClass
	switch m := mc.(type) {
	case *tg.Message:
		reply = m.ReplyTo
	case *tg.MessageService:
		if _, ok := m.Action.(*tg.MessageActionTopicCreate); ok {
			return m.ID
		}
		reply = m.ReplyTo
	}
	if r, ok := reply.(*tg.MessageReplyHeader); ok && r.ForumTopic {
		if r.ReplyToTopID != 0 {
			return r.ReplyToTopID
		}
		if r.ReplyToMsgID != 0 {
			return r.ReplyToMsgID
		}
	}
	return generalTopic
}

func (s *session) Topics(ctx context.Context, chatID string) ([]model.Topic, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return nil, err
	}
	var (
		out  []model.Topic
		seen = map[string]bool{}
		req  = &tg.MessagesGetForumTopicsRequest{Peer: peer, Limit: 100}
	)
	for range 5 { // up to 500 topics
		res, err := api.MessagesGetForumTopics(ctx, req)
		if err != nil {
			return out, authErr(err)
		}
		s.learn(res.Users, res.Chats)
		last := map[int]model.Message{}
		for _, mc := range res.Messages {
			if m, ok := s.convert(mc); ok {
				last[int(m.Seq)] = m
			}
		}
		var tail *tg.ForumTopic
		for _, tc := range res.Topics {
			t, ok := tc.(*tg.ForumTopic)
			if !ok {
				continue
			}
			tail = t
			x := forumTopic(t, last)
			if seen[x.ID] {
				continue
			}
			seen[x.ID] = true
			out = append(out, x)
		}
		if tail == nil || len(res.Topics) < req.Limit || len(out) >= res.Count {
			break
		}
		req.OffsetTopic, req.OffsetID, req.OffsetDate = tail.ID, tail.TopMessage, 0
		if m, ok := last[tail.TopMessage]; ok {
			req.OffsetDate = int(m.Time.Unix())
		}
	}
	// Pinned first, keeping their order; then the latest activity first.
	slices.SortStableFunc(out, func(a, b model.Topic) int {
		switch {
		case a.Pinned != b.Pinned:
			if a.Pinned {
				return -1
			}
			return 1
		case a.Pinned:
			return 0
		}
		return b.Updated.Compare(a.Updated)
	})
	return out, nil
}

// forumTopic maps a forum topic; last holds the topics' latest messages.
func forumTopic(t *tg.ForumTopic, last map[int]model.Message) model.Topic {
	x := model.Topic{
		ID: strconv.Itoa(t.ID), Title: t.Title, Unread: t.UnreadCount, Color: t.IconColor,
		Pinned: t.Pinned, Closed: t.Closed, Hidden: t.Hidden, Updated: time.Unix(int64(t.Date), 0),
	}
	if m, ok := last[t.TopMessage]; ok {
		m.InThread = x.ID
		x.Last = &m
		if m.Time.After(x.Updated) {
			x.Updated = m.Time
		}
	}
	return x
}

func (s *session) ReadTopic(ctx context.Context, chatID, topicID, msgID string) error {
	api, _, err := s.live()
	if err != nil {
		return err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return err
	}
	topic, err := strconv.Atoi(topicID)
	if err != nil {
		return fmt.Errorf("telegram: bad topic id %q", topicID)
	}
	maxID := 0
	if msgID != "" {
		if maxID, err = strconv.Atoi(msgID); err != nil {
			return fmt.Errorf("telegram: bad message id %q", msgID)
		}
	} else {
		res, err := api.MessagesGetForumTopicsByID(ctx, &tg.MessagesGetForumTopicsByIDRequest{Peer: peer, Topics: []int{topic}})
		if err != nil {
			return authErr(err)
		}
		for _, tc := range res.Topics {
			if t, ok := tc.(*tg.ForumTopic); ok && t.ID == topic {
				maxID = t.TopMessage
			}
		}
		if maxID == 0 {
			return errors.New("telegram: no such topic")
		}
	}
	if _, err := api.MessagesReadDiscussion(ctx, &tg.MessagesReadDiscussionRequest{Peer: peer, MsgID: topic, ReadMaxID: maxID}); err != nil {
		return authErr(err)
	}
	s.reportUnread(ctx, api, chatID, peer)
	return nil
}

// reportUnread tells the hub the chat's unread count after a topic was
// read (Telegram sends no update for it).
func (s *session) reportUnread(ctx context.Context, api *tg.Client, chatID string, peer tg.InputPeerClass) {
	s.mu.Lock()
	sink := s.sink
	s.mu.Unlock()
	if sink == nil {
		return
	}
	res, err := api.MessagesGetPeerDialogs(ctx, []tg.InputDialogPeerClass{&tg.InputDialogPeer{Peer: peer}})
	if err != nil {
		s.env.Logf("warn", "unread after reading a topic: %v", err)
		return
	}
	for _, dc := range res.Dialogs {
		if d, ok := dc.(*tg.Dialog); ok && peerKey(d.Peer) == chatID {
			sink.Emit(model.Event{Kind: model.EvRead, Chat: &model.Chat{ID: chatID, Unread: d.UnreadCount}})
		}
	}
}
