package slack

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

// Chat folders are the sidebar's sections. UNVERIFIED: Slack does not document
// users.channelSections.list or the "channel_section*" events; only a browser
// session (xoxc) can call them. Names below are what the web client is known
// to use. An account where the call is refused simply has no folders.

var _ provider.ChatFolders = (*session)(nil)

// foldersDelay gathers a burst of section events into one reload (a reload
// reads the whole chat list).
var foldersDelay = 3 * time.Second

type wSection struct {
	ID    string `json:"channel_section_id"`
	Name  string `json:"name"`
	Type  string `json:"type"` // "standard" is a section the user made
	Emoji string `json:"emoji"`
	Next  string `json:"next_channel_section_id"`
	Page  struct {
		IDs   []string `json:"channel_ids"`
		Count int      `json:"count"`
	} `json:"channel_ids_page"`
}

// ChatFolders returns the sections that list chats by hand, in sidebar order.
// The catch-all ones (Channels, Direct messages, Apps) are the chat list itself.
func (s *session) ChatFolders(ctx context.Context) ([]model.ChatFolder, error) {
	if s.params["method"] == methodOAuth {
		return nil, nil
	}
	var r struct {
		Sections []wSection `json:"channel_sections"`
	}
	var keys map[string]json.RawMessage
	var raw json.RawMessage
	if err := s.call(ctx, "users.channelSections.list", nil, &raw); err != nil {
		var ae *apiError
		if errors.As(err, &ae) && !isAuth(err) {
			s.env.Logf("warn", "slack: sidebar sections: %s", ae.Code)
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	if len(r.Sections) == 0 {
		if json.Unmarshal(raw, &keys) == nil {
			names := make([]string, 0, len(keys))
			for k := range keys {
				names = append(names, k)
			}
			sort.Strings(names)
			s.env.Logf("info", "slack: no sidebar sections; answer keys: %s", strings.Join(names, ","))
		}
		return nil, nil
	}
	var out []model.ChatFolder
	for _, c := range sidebarOrder(r.Sections) {
		title := strings.TrimSpace(c.Name)
		if c.Type == "stars" && title == "" {
			title = "Starred"
		}
		if (c.Type != "standard" && c.Type != "stars") || title == "" || len(c.Page.IDs) == 0 {
			continue
		}
		if c.Page.Count > len(c.Page.IDs) {
			s.env.Logf("warn", "slack: a sidebar section lists %d of its %d chats", len(c.Page.IDs), c.Page.Count)
		}
		f := model.ChatFolder{ID: c.ID, Title: title, Always: c.Page.IDs}
		base, _, _ := strings.Cut(strings.Trim(c.Emoji, ":"), "::")
		f.Emoji = emojiNames[base] // a custom emoji has no character
		out = append(out, f)
	}
	return out, nil
}

// sidebarOrder follows the sections' "next" links from the first one; any
// section the links miss goes last, in the order given.
func sidebarOrder(list []wSection) []wSection {
	byID := make(map[string]wSection, len(list))
	pointed := make(map[string]bool, len(list))
	for _, c := range list {
		byID[c.ID] = c
		if c.Next != "" {
			pointed[c.Next] = true
		}
	}
	out := make([]wSection, 0, len(list))
	seen := make(map[string]bool, len(list))
	for _, c := range list {
		if pointed[c.ID] {
			continue
		}
		for ok := true; ok && !seen[c.ID]; c, ok = byID[c.Next] {
			seen[c.ID] = true
			out = append(out, c)
		}
	}
	for _, c := range list {
		if !seen[c.ID] {
			seen[c.ID] = true
			out = append(out, c)
		}
	}
	return out
}

// sectionsChanged reports a change of the sections, once per burst.
func (s *session) sectionsChanged(ctx context.Context, sink provider.Sink) {
	s.mu.Lock()
	due := s.foldersDue
	s.foldersDue = true
	s.mu.Unlock()
	if due {
		return
	}
	go func() {
		select {
		case <-time.After(foldersDelay):
		case <-ctx.Done():
		}
		s.mu.Lock()
		s.foldersDue = false
		s.mu.Unlock()
		if ctx.Err() == nil {
			sink.Emit(model.Event{Kind: model.EvChatFolders})
		}
	}()
}

// --- quotes ----------------------------------------------------------------

// splitQuote cuts the leading "> " lines (raw "&gt; " too) off a text. With
// nothing after them the text stays whole: it is a quote, not a reply to one.
func splitQuote(text string) (quote []string, rest string) {
	lines := strings.Split(text, "\n")
	n := 0
	for n < len(lines) {
		l, ok := strings.CutPrefix(lines[n], "&gt; ")
		if !ok {
			l, ok = strings.CutPrefix(lines[n], "> ")
		}
		if !ok {
			break
		}
		quote = append(quote, l)
		n++
	}
	rest = strings.Join(lines[n:], "\n")
	if n == 0 || strings.TrimSpace(rest) == "" {
		return nil, text
	}
	return quote, rest
}

// quoteHead is the raw quote a reply shown in the chat starts with: the app
// edits only the text after it, and chat.update replaces the whole text.
func (s *session) quoteHead(ctx context.Context, chatID, ts string) (string, error) {
	var r struct {
		Messages []wMessage `json:"messages"`
	}
	q := map[string][]string{"channel": {chatID}, "latest": {ts}, "inclusive": {"true"}, "limit": {"1"}}
	if err := s.call(ctx, "conversations.history", q, &r); err != nil {
		return "", err
	}
	for _, m := range r.Messages {
		if m.TS != ts || m.Subtype != "thread_broadcast" {
			continue
		}
		lines, rest := splitQuote(m.Text)
		if len(lines) == 0 {
			return "", nil
		}
		return m.Text[:len(m.Text)-len(rest)], nil
	}
	return "", nil
}
