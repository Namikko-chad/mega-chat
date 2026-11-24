package slack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/richtext"
)

// --- threads -------------------------------------------------------------

// replies returns a whole thread, root first (Slack pages it oldest
// first; threads are short enough to read in full, up to 1000).
func (s *session) replies(ctx context.Context, chatID, ts string, limit int) ([]wMessage, error) {
	var all []wMessage
	cursor := ""
	for {
		q := url.Values{"channel": {chatID}, "ts": {ts}, "limit": {fmt.Sprint(limit)}}
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		var r struct {
			Messages []wMessage `json:"messages"`
			Meta     struct {
				Next string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := s.call(ctx, "conversations.replies", q, &r); err != nil {
			return all, err
		}
		all = append(all, r.Messages...)
		if r.Meta.Next == "" || len(all) >= 1000 || limit == 1 {
			return all, nil
		}
		cursor = r.Meta.Next
	}
}

func (s *session) Thread(ctx context.Context, chatID, threadID, beforeID string, limit int) ([]model.Message, bool, error) {
	raw, err := s.replies(ctx, chatID, threadID, 200)
	if err != nil {
		return nil, false, err
	}
	defer s.saveUsers()
	var all []model.Message // newest first
	for i := len(raw) - 1; i >= 0; i-- {
		m := raw[i]
		if m.TS == threadID || m.TS == "" || m.Hidden {
			continue
		}
		msg := s.convert(ctx, chatID, m)
		msg.InThread, msg.ReplyTo = threadID, ""
		all = append(all, msg)
	}
	if beforeID != "" {
		before := tsTime(beforeID)
		i := sort.Search(len(all), func(i int) bool { return all[i].Time.Before(before) })
		all = all[i:]
	}
	if len(all) > limit {
		return all[:limit], true, nil
	}
	return all, false, nil
}

// emitRoot reports a thread's root again with its new reply count.
func (s *session) emitRoot(ctx context.Context, sink provider.Sink, chatID, ts string) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	raw, err := s.replies(ctx, chatID, ts, 1)
	if err != nil {
		s.env.Logf("warn", "thread root: %v", err)
		return
	}
	for _, m := range raw {
		if m.TS == ts && !skipped(m) {
			msg := s.convert(ctx, chatID, m)
			sink.Emit(model.Event{Kind: model.EvMessageEdit, Message: &msg})
		}
	}
}

// --- reactions -----------------------------------------------------------

// reactionsAt returns the reactions of a message.
func (s *session) reactionsAt(ctx context.Context, chatID, ts string) ([]model.Reaction, error) {
	var r struct {
		Message wMessage `json:"message"`
	}
	q := url.Values{"channel": {chatID}, "timestamp": {ts}, "full": {"true"}}
	if err := s.call(ctx, "reactions.get", q, &r); err != nil {
		return nil, err
	}
	return s.reactionsOf(r.Message.Reactions), nil
}

func (s *session) emitReactions(ctx context.Context, sink provider.Sink, chatID, ts string) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	rs, err := s.reactionsAt(ctx, chatID, ts)
	if err != nil {
		s.env.Logf("warn", "reactions: %v", err)
		return
	}
	sink.Emit(model.Event{Kind: model.EvReactions, Message: &model.Message{ChatID: chatID, ID: ts, Reactions: rs}})
}

func (s *session) React(ctx context.Context, chatID, msgID, emoji string, on bool) ([]model.Reaction, error) {
	method := "reactions.add"
	if !on {
		method = "reactions.remove"
	}
	q := url.Values{"channel": {chatID}, "timestamp": {msgID}, "name": {emojiName(emoji)}}
	if err := s.call(ctx, method, q, nil); err != nil {
		if ae, ok := err.(*apiError); !ok || (ae.Code != "already_reacted" && ae.Code != "no_reaction") {
			return nil, err
		}
	}
	return s.reactionsAt(ctx, chatID, msgID)
}

// --- media ---------------------------------------------------------------

// mediaHost tells whether a URL may be fetched and with the account's
// credentials: they go only to Slack's own hosts (the workspace,
// files.slack.com); avatars and previews on Slack's CDNs are public.
func (s *session) mediaHost(u *url.URL) (ok, creds bool) {
	base, _ := url.Parse(s.client().base)
	h := strings.ToLower(u.Hostname())
	switch {
	case base != nil && h == strings.ToLower(base.Hostname()):
		return true, true
	case u.Scheme != "https":
		return false, false
	case h == "slack.com" || strings.HasSuffix(h, ".slack.com"):
		return true, true
	}
	for _, cdn := range []string{"slack-edge.com", "slack-files.com", "slack-imgs.com", "gravatar.com"} {
		if h == cdn || strings.HasSuffix(h, "."+cdn) {
			return true, false
		}
	}
	return false, false
}

func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return nil, "", fmt.Errorf("slack: bad media ref")
	}
	ok, creds := s.mediaHost(u)
	if !ok {
		return nil, "", fmt.Errorf("slack: media from %s is not fetched", u.Hostname())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	if err != nil {
		return nil, "", err
	}
	a := s.client()
	if creds {
		if a.token != "" {
			req.Header.Set("Authorization", "Bearer "+a.token)
		}
		if a.cookie != "" {
			req.Header.Set("Cookie", "d="+url.QueryEscape(a.cookie))
		}
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("slack: media: HTTP %d", resp.StatusCode)
	}
	mime := resp.Header.Get("Content-Type")
	if strings.HasPrefix(mime, "text/html") {
		// Slack answers a file request without access with its sign-in page.
		return nil, "", fmt.Errorf("slack: media: no access to the file")
	}
	if resp.ContentLength > limit {
		return nil, "", provider.ErrTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		return nil, "", provider.ErrTooLarge
	}
	return data, mime, nil
}

// Delete deletes the account's messages for everyone (chat.delete); Slack
// has no "only for me". The delete comes back over RTM too.
func (s *session) Delete(ctx context.Context, chatID string, msgIDs []string, forAll bool) error {
	if !forAll {
		return provider.ErrUnsupported
	}
	for _, id := range msgIDs {
		if err := s.call(ctx, "chat.delete", url.Values{"channel": {chatID}, "ts": {id}}, nil); err != nil {
			return err
		}
	}
	return nil
}

// DeletesForMe: Slack deletes for everyone only.
func (s *session) DeletesForMe() bool { return false }

// Edit changes the text of one of the account's messages (chat.update).
// The full message comes back over RTM as message_changed.
func (s *session) Edit(ctx context.Context, chatID, msgID, text string) (model.Message, error) {
	var r struct {
		TS string `json:"ts"`
	}
	head, err := s.quoteHead(ctx, chatID, msgID)
	if err != nil {
		return model.Message{}, err
	}
	d := richtext.Parse(text, richtext.Markdown)
	if err := s.call(ctx, "chat.update", url.Values{"channel": {chatID}, "ts": {msgID}, "text": {head + richtext.ToSlack(d, nil)}}, &r); err != nil {
		return model.Message{}, err
	}
	return model.Message{ChatID: chatID, ID: msgID, Text: d.Text, Format: d.Spans, Edited: time.Now(), Outgoing: true}, nil
}
