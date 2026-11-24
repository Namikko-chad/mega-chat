package fake

import (
	"context"
	"fmt"

	"megachat/core/model"
	"megachat/core/provider"
)

// Demo chat folders: people and groups, as Telegram's rules would sort them.
var demoChatFolders = []model.ChatFolder{
	{ID: "people", Title: "People", Emoji: "👤"},
	{ID: "teams", Title: "Teams", Emoji: "👥", ExcludeArchived: true},
}

// folderIDs are the demo folders a chat is in by rule.
func folderIDs(c model.Chat) []string {
	switch c.Kind {
	case model.ChatDirect:
		if c.Saved {
			return []string{}
		}
		return []string{"people"}
	case model.ChatGroup:
		return []string{"teams"}
	}
	return []string{}
}

// SeedSpecial adds the saved chat and archives a chat of login (the demo).
func (s *Server) SeedSpecial(login string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.user(login)
	u.chats["saved"] = &chat{Chat: model.Chat{ID: "saved", Kind: model.ChatDirect, Title: login, Saved: true, Access: model.AccessWrite}, threads: map[string][]model.Message{}}
	if c := u.chats["c10"]; c != nil {
		c.Archived = true
	}
	if c := u.chats["c3"]; c != nil {
		c.Pinned = true
	}
}

func (c *session) ChatFolders(context.Context) ([]model.ChatFolder, error) {
	return append([]model.ChatFolder(nil), demoChatFolders...), nil
}

func (c *session) ChatActions() []model.ChatAction {
	return []model.ChatAction{
		model.ChatMute, model.ChatUnmute, model.ChatPin, model.ChatUnpin,
		model.ChatArchive, model.ChatUnarchive, model.ChatLeave,
	}
}

func (c *session) ManageChat(_ context.Context, chatID string, action model.ChatAction) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	u := c.s.user(c.login)
	ch := u.chats[chatID]
	if ch == nil {
		return fmt.Errorf("fake: no chat %q", chatID)
	}
	if action == model.ChatLeave {
		if ch.Kind == model.ChatDirect {
			return fmt.Errorf("%w: a direct chat cannot be left", provider.ErrUnsupported)
		}
		delete(u.chats, chatID)
		return nil
	}
	f, ok := model.FlagsOf(action)
	if !ok {
		return provider.ErrUnsupported
	}
	f.Apply(&ch.Chat)
	return nil
}
