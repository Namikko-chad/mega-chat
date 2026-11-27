package vkteams

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// A VK Teams thread is a hidden chat of its own ("<n>@chat.agent") whose
// parentTopic names the message it hangs under. Here it is what every other
// messenger's thread is: not in the chat list, its id the root message's id,
// its replies under that message. Wire names come from the open im-desktop
// client (thread/add, thread/feed/get, threadUpdate, parentTopic).

// threadRef is the message a thread hangs under.
type threadRef struct {
	Chat string `json:"chat"`
	Msg  string `json:"msg"`
}

const maxFeedPages = 5

func msgKey(chat, msg string) string { return chat + "\x00" + msg }

// linkThread records that sn is the thread of a message and takes it out
// of the chat list (once).
func (s *session) linkThread(sink provider.Sink, sn string, t *wTopic) {
	if !t.valid() || sn == "" || sn == t.ChatID {
		return
	}
	ref := threadRef{Chat: t.ChatID, Msg: t.MessageID.String()}
	s.mu.Lock()
	d := s.dialogs[sn]
	if d == nil {
		d = &dialog{SN: sn}
		s.dialogs[sn] = d
	}
	first := d.Parent == nil
	d.Parent = &ref
	s.threadOf[msgKey(ref.Chat, ref.Msg)] = sn
	s.mu.Unlock()
	if first {
		s.saveDialogs()
		if sink != nil {
			sink.Emit(model.Event{Kind: model.EvChatRemoved, Chat: &model.Chat{ID: sn}})
		}
	}
}

func (s *session) parentOf(sn string) *threadRef {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d := s.dialogs[sn]; d != nil {
		return d.Parent
	}
	return nil
}

// classify asks the server whether a group is a thread; true if it is. An
// answer is remembered, a failure is asked again later.
func (s *session) classify(ctx context.Context, sink provider.Sink, sn string) bool {
	if !isGroup(sn) {
		return false
	}
	s.mu.Lock()
	d := s.dialogs[sn]
	if d != nil && d.Parent != nil {
		s.mu.Unlock()
		return true
	}
	done := s.notThread[sn]
	s.mu.Unlock()
	if done {
		return false
	}
	aimsid, online := s.current()
	if !online || aimsid == "" {
		return false
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r, err := s.srv.getHistory(cctx, aimsid, sn, "-1", 1)
	if err != nil {
		return false
	}
	if r.ParentTopic.valid() {
		s.linkThread(sink, sn, r.ParentTopic)
		return true
	}
	s.mu.Lock()
	s.notThread[sn] = true
	s.mu.Unlock()
	return false
}

// handleThreadState passes a thread's new messages on as thread replies.
func (s *session) handleThreadState(ctx context.Context, sink provider.Sink, st dlgState, ref threadRef) {
	s.mu.Lock()
	if d := s.dialogs[st.SN]; d != nil {
		d.Unread = st.UnreadCnt
		if st.LastMsgID != "" {
			d.LastID = st.LastMsgID.String()
		}
	}
	s.mu.Unlock()
	rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	reactions := s.reactionsOf(rctx, st.SN, st.Messages)
	cancel()
	for _, m := range st.Messages {
		msg := s.reply(st.SN, ref, m)
		msg.Reactions = reactions[msg.ID]
		for _, p := range m.UpdatePatch {
			if p.Type == "delete" {
				msg = model.Message{ChatID: ref.Chat, ID: msg.ID, InThread: ref.Msg, Deleted: true}
			}
		}
		sink.Emit(model.Event{Kind: model.EvThreadMessage, Message: &msg})
	}
}

// reply converts a message of thread sn into a reply under ref.
func (s *session) reply(sn string, ref threadRef, m wMessage) model.Message {
	msg := s.convert(sn, m)
	msg.ChatID, msg.InThread, msg.Thread = ref.Chat, ref.Msg, nil
	s.mu.Lock()
	s.replyHome[msgKey(ref.Chat, msg.ID)] = sn
	s.mu.Unlock()
	return msg
}

// handleThreadUpdate keeps a thread's reply count and refreshes its root.
func (s *session) handleThreadUpdate(ctx context.Context, sink provider.Sink, raw json.RawMessage) {
	var u wThread
	if json.Unmarshal(raw, &u) != nil || u.ThreadID == "" || !u.ParentTopic.valid() {
		return
	}
	s.linkThread(sink, u.ThreadID, u.ParentTopic)
	s.mu.Lock()
	changed := s.replies[u.ThreadID] != int(u.RepliesCount)
	s.replies[u.ThreadID] = int(u.RepliesCount)
	s.mu.Unlock()
	if changed {
		go s.refreshRoot(ctx, sink, u.ParentTopic.ChatID, u.ParentTopic.MessageID.String())
	}
}

// refreshRoot sends the root message again, with its new count.
func (s *session) refreshRoot(ctx context.Context, sink provider.Sink, chat, msgID string) {
	aimsid, online := s.current()
	id, err := strconv.ParseInt(msgID, 10, 64)
	if !online || aimsid == "" || err != nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// The answer is what is older than fromMsgId, maybe with it.
	r, err := s.srv.getHistory(cctx, aimsid, chat, strconv.FormatInt(id+1, 10), 2)
	if err != nil {
		return
	}
	s.learn(r.Persons)
	reactions := s.reactionsOf(cctx, chat, r.Messages)
	for _, m := range r.Messages {
		if m.MsgID.String() == msgID {
			msg := s.convert(chat, m)
			msg.Reactions = reactions[msg.ID]
			sink.Emit(model.Event{Kind: model.EvMessageEdit, Message: &msg})
			return
		}
	}
}

func sinkCtx(s *session) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runCtx != nil {
		return s.runCtx
	}
	return context.Background()
}

// refreshReply sends a reply again, with its new reactions.
func (s *session) refreshReply(ctx context.Context, sink provider.Sink, sn string, ref threadRef, msgID string) {
	aimsid, online := s.current()
	id, err := strconv.ParseInt(msgID, 10, 64)
	if !online || aimsid == "" || err != nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r, err := s.srv.getHistory(cctx, aimsid, sn, strconv.FormatInt(id+1, 10), 2)
	if err != nil {
		return
	}
	s.learn(r.Persons)
	reactions := s.reactionsOf(cctx, sn, r.Messages)
	for _, m := range r.Messages {
		if m.MsgID.String() == msgID {
			msg := s.reply(sn, ref, m)
			msg.Reactions = reactions[msg.ID]
			sink.Emit(model.Event{Kind: model.EvThreadMessage, Message: &msg})
			return
		}
	}
}

// loadFeed reads the threads the account follows: their roots and counts.
func (s *session) loadFeed(ctx context.Context, sink provider.Sink) {
	aimsid, online := s.current()
	if !online || aimsid == "" {
		return
	}
	cursor := ""
	for range maxFeedPages {
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		r, err := s.srv.threadsFeed(cctx, aimsid, cursor)
		cancel()
		if err != nil {
			if ctx.Err() == nil {
				s.env.Logf("warn", "vk teams: threads feed: %v", err)
			}
			return
		}
		s.learn(r.Persons)
		for i := range r.Threads {
			t := &r.Threads[i]
			s.linkThread(sink, t.ThreadID, t.ParentTopic)
			if t.ParentTopic.valid() {
				s.mu.Lock()
				s.replies[t.ThreadID] = int(t.RepliesCount)
				s.mu.Unlock()
			}
		}
		if r.Cursor == "" || r.Cursor == cursor || len(r.Threads) == 0 {
			return
		}
		cursor = r.Cursor
	}
}

// Thread returns the replies under a message, newest first. A message
// without a thread has none yet: one starts with the first reply.
func (s *session) Thread(ctx context.Context, chatID, threadID, beforeID string, limit int) ([]model.Message, bool, error) {
	aimsid, err := s.live()
	if err != nil {
		return nil, false, err
	}
	s.mu.Lock()
	sn := s.threadOf[msgKey(chatID, threadID)]
	s.mu.Unlock()
	if sn == "" {
		return nil, false, nil
	}
	from, n := "-1", limit
	var before int64
	if beforeID != "" {
		from, n = beforeID, limit+1
		before, _ = strconv.ParseInt(beforeID, 10, 64)
	}
	r, err := s.srv.getHistory(ctx, aimsid, sn, from, n)
	if err != nil {
		return nil, false, err
	}
	s.learn(r.Persons)
	reactions := s.reactionsOf(ctx, sn, r.Messages)
	ref := threadRef{Chat: chatID, Msg: threadID}
	out := make([]model.Message, 0, len(r.Messages))
	for _, m := range r.Messages {
		msg := s.reply(sn, ref, m)
		msg.Reactions = reactions[msg.ID]
		if before != 0 && msg.Seq >= before {
			continue
		}
		out = append(out, msg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq > out[j].Seq })
	if len(out) > limit {
		out = out[:limit]
	}
	// Opening a thread reads it, as in the messenger's own client.
	if beforeID == "" && len(out) > 0 {
		if err := s.srv.setDlgState(ctx, aimsid, sn, out[0].ID); err != nil {
			s.env.Logf("warn", "vk teams: mark thread read: %v", err)
		}
	}
	return out, len(r.Messages) >= n, nil
}

// threadTarget is the hidden chat of a message's thread, made on the first
// reply.
func (s *session) threadTarget(ctx context.Context, aimsid, chatID, threadID string) (string, error) {
	s.mu.Lock()
	sn := s.threadOf[msgKey(chatID, threadID)]
	sink := s.sink
	s.mu.Unlock()
	if sn != "" {
		return sn, nil
	}
	sn, err := s.srv.threadAdd(ctx, aimsid, chatID, threadID)
	if err != nil {
		return "", err
	}
	s.linkThread(sink, sn, &wTopic{Type: "message", ChatID: chatID, MessageID: json.Number(threadID)})
	return sn, nil
}

// homeOf is the chat a message really lives in: its thread for a reply.
func (s *session) homeOf(chatID, msgID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sn := s.replyHome[msgKey(chatID, msgID)]; sn != "" {
		return sn
	}
	return chatID
}
