package telegram

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"

	"megachat/core/model"
	"megachat/core/provider"
)

// --- Media ---------------------------------------------------------------

// limitWriter fails once more than n bytes were written.
type limitWriter struct {
	buf bytes.Buffer
	n   int64
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if int64(w.buf.Len()+len(p)) > w.n {
		return 0, provider.ErrTooLarge
	}
	return w.buf.Write(p)
}

func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, "", err
	}
	kind, rest, _ := strings.Cut(ref, ":")
	i := strings.LastIndexByte(rest, ':')
	if i < 0 {
		return nil, "", fmt.Errorf("telegram: bad media ref %q", ref)
	}
	chat, num := rest[:i], rest[i+1:]
	var (
		loc  tg.InputFileLocationClass
		mime string
		size int64
	)
	switch kind {
	case "a":
		photoID, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("telegram: bad media ref %q", ref)
		}
		peer, err := s.inputPeer(chat)
		if err != nil {
			return nil, "", err
		}
		loc, mime = &tg.InputPeerPhotoFileLocation{Peer: peer, PhotoID: photoID}, "image/jpeg"
	case "m", "t":
		id, err := strconv.Atoi(num)
		if err != nil {
			return nil, "", fmt.Errorf("telegram: bad media ref %q", ref)
		}
		media, err := s.messageMedia(ctx, api, chat, id)
		if err != nil {
			return nil, "", err
		}
		if loc, mime, size, err = fileLocation(media, kind == "t"); err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("telegram: bad media ref %q", ref)
	}
	if size > limit {
		return nil, "", provider.ErrTooLarge
	}
	w := &limitWriter{n: limit}
	if _, err := downloader.NewDownloader().Download(api, loc).Stream(ctx, w); err != nil {
		return nil, "", authErr(err)
	}
	data := w.buf.Bytes()
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	return data, mime, nil
}

// messageMedia fetches a message again for a fresh file reference.
func (s *session) messageMedia(ctx context.Context, api *tg.Client, chat string, id int) (tg.MessageMediaClass, error) {
	ids := []tg.InputMessageClass{&tg.InputMessageID{ID: id}}
	var res tg.MessagesMessagesClass
	peer, err := s.inputPeer(chat)
	if err != nil {
		return nil, err
	}
	if ch, ok := peer.(*tg.InputPeerChannel); ok {
		res, err = api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash}, ID: ids,
		})
	} else {
		res, err = api.MessagesGetMessages(ctx, ids)
	}
	if err != nil {
		return nil, authErr(err)
	}
	page, ok := res.AsModified()
	if !ok {
		return nil, errors.New("telegram: message not found")
	}
	for _, mc := range page.GetMessages() {
		if m, ok := mc.(*tg.Message); ok && m.ID == id && m.Media != nil {
			return m.Media, nil
		}
	}
	return nil, errors.New("telegram: the message has no media (deleted?)")
}

// fileLocation is where the file (or its preview) of a message's media is.
func fileLocation(media tg.MessageMediaClass, thumb bool) (tg.InputFileLocationClass, string, int64, error) {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		p, ok := m.Photo.(*tg.Photo)
		if !ok {
			return nil, "", 0, errors.New("telegram: photo is gone")
		}
		full, small, _, _, size := photoSizes(p.Sizes)
		typ := full
		if thumb {
			typ, size = small, 0
		}
		return &tg.InputPhotoFileLocation{ID: p.ID, AccessHash: p.AccessHash, FileReference: p.FileReference, ThumbSize: typ}, "image/jpeg", size, nil
	case *tg.MessageMediaDocument:
		d, ok := m.Document.(*tg.Document)
		if !ok {
			return nil, "", 0, errors.New("telegram: file is gone")
		}
		loc := &tg.InputDocumentFileLocation{ID: d.ID, AccessHash: d.AccessHash, FileReference: d.FileReference}
		if !thumb {
			return loc, d.MimeType, d.Size, nil
		}
		if _, small, _, _, _ := photoSizes(d.Thumbs); small != "" {
			loc.ThumbSize = small
			return loc, "", 0, nil
		}
		return loc, d.MimeType, d.Size, nil // no preview: the file itself (a small image)
	}
	return nil, "", 0, errors.New("telegram: no file in this message")
}

// --- Threads (channel comments) --------------------------------------------

func (s *session) Thread(ctx context.Context, chatID, threadID, beforeID string, limit int) ([]model.Message, bool, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, false, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return nil, false, err
	}
	post, err := strconv.Atoi(threadID)
	if err != nil {
		return nil, false, fmt.Errorf("telegram: bad thread id %q", threadID)
	}
	offset := 0
	if beforeID != "" {
		if offset, err = strconv.Atoi(beforeID); err != nil {
			return nil, false, fmt.Errorf("telegram: bad message id %q", beforeID)
		}
	}
	res, err := api.MessagesGetReplies(ctx, &tg.MessagesGetRepliesRequest{Peer: peer, MsgID: post, OffsetID: offset, Limit: limit})
	if err != nil && post == generalTopic && s.forum(chatID) {
		// The general topic has no message that opened it.
		return s.generalTopic(ctx, api, peer, offset, limit)
	}
	if err != nil {
		return nil, false, authErr(err)
	}
	page, ok := res.AsModified()
	if !ok {
		return nil, false, nil
	}
	s.learn(page.GetUsers(), page.GetChats())
	raw := page.GetMessages()
	out := make([]model.Message, 0, len(raw))
	for _, mc := range raw {
		if m, ok := s.convert(mc); ok {
			m.Thread = nil
			out = append(out, m)
		}
	}
	return out, len(raw) >= limit, nil
}

// generalTopic reads a forum's general topic from the chat's history.
func (s *session) generalTopic(ctx context.Context, api *tg.Client, peer tg.InputPeerClass, offset, limit int) ([]model.Message, bool, error) {
	var out []model.Message
	for range 5 {
		res, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: peer, OffsetID: offset, Limit: 100})
		if err != nil {
			return out, len(out) > 0, authErr(err)
		}
		page, ok := res.AsModified()
		if !ok {
			return out, false, nil
		}
		s.learn(page.GetUsers(), page.GetChats())
		raw := page.GetMessages()
		for _, mc := range raw {
			if topicOf(mc) != generalTopic {
				continue
			}
			if m, ok := s.convert(mc); ok {
				out = append(out, m)
				if len(out) == limit {
					return out, true, nil
				}
			}
		}
		if len(raw) < 100 {
			return out, false, nil
		}
		offset = raw[len(raw)-1].GetID()
	}
	return out, true, nil
}

// discussion returns the comment group of a channel post and the id of
// the post's copy there, which comments reply to.
func (s *session) discussion(ctx context.Context, api *tg.Client, chatID string, post int) (string, int, error) {
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return "", 0, err
	}
	res, err := api.MessagesGetDiscussionMessage(ctx, &tg.MessagesGetDiscussionMessageRequest{Peer: peer, MsgID: post})
	if err != nil {
		return "", 0, authErr(err)
	}
	s.learn(res.Users, res.Chats)
	for _, mc := range res.Messages {
		if m, ok := mc.(*tg.Message); ok {
			return peerKey(m.PeerID), m.ID, nil
		}
	}
	return "", 0, errors.New("telegram: comments are off for this post")
}

// --- Reactions -------------------------------------------------------------

func (s *session) React(ctx context.Context, chatID, msgID, emoji string, on bool) ([]model.Reaction, error) {
	api, gaps, err := s.live()
	if err != nil {
		return nil, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(msgID)
	if err != nil {
		return nil, fmt.Errorf("telegram: bad message id %q", msgID)
	}
	// One reaction per message (what any account may set): choosing one
	// replaces the previous, taking it back leaves none.
	req := &tg.MessagesSendReactionRequest{Peer: peer, MsgID: id, Reaction: []tg.ReactionClass{}}
	if on {
		req.Reaction = []tg.ReactionClass{&tg.ReactionEmoji{Emoticon: emoji}}
	}
	res, err := api.MessagesSendReaction(ctx, req)
	if err != nil {
		return nil, authErr(err)
	}
	var out []model.Reaction
	found := false
	if u, ok := res.(*tg.Updates); ok {
		for _, x := range u.Updates {
			if r, ok := x.(*tg.UpdateMessageReactions); ok && r.MsgID == id {
				out, found = reactionsOf(r.Reactions), true
			}
		}
	}
	if err := gaps.Handle(ctx, res); err != nil {
		s.env.Logf("warn", "updates after reaction: %v", err)
	}
	if !found {
		return nil, errors.New("telegram: the reaction was sent, but the server did not return the result")
	}
	return out, nil
}
