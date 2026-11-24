package hub

import (
	"context"
	"fmt"
	"slices"
	"time"

	"megachat/core/eventlog"
	"megachat/core/model"
	"megachat/core/provider"
)

// EvChatFolders: the chat folders of AccountID changed: reload them.
const EvChatFolders EventKind = 200

// ErrCannotLeave is returned for leaving a direct or saved chat.
var ErrCannotLeave = fmt.Errorf("%w: only groups and channels can be left", provider.ErrUnsupported)

// chatManager returns the ChatManager of an account's session.
func (h *Hub) chatManager(acct string) (provider.ChatManager, error) {
	m, _, err := h.messenger(acct)
	if err != nil {
		return nil, err
	}
	cm, ok := m.(provider.ChatManager)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	return cm, nil
}

// ChatActions lists what ManageChat can do for an account (nil: nothing).
func (h *Hub) ChatActions(acct string) []model.ChatAction {
	h.mu.Lock()
	r := h.runs[acct]
	var sess provider.Session
	if r != nil {
		sess = r.sess
	}
	h.mu.Unlock()
	if cm, ok := sess.(provider.ChatManager); ok {
		return cm.ChatActions()
	}
	return nil
}

// ManageChat mutes, pins, archives or leaves a chat and returns it as it is
// now (after leaving: only the ids; it is removed from the cache).
func (h *Hub) ManageChat(ctx context.Context, acct, chat string, action model.ChatAction) (model.Chat, error) {
	cm, err := h.chatManager(acct)
	if err != nil {
		return model.Chat{}, err
	}
	if !slices.Contains(cm.ChatActions(), action) {
		return model.Chat{}, provider.ErrUnsupported
	}
	if action == model.ChatLeave {
		if c, ok, _ := h.st.Chat(acct, chat); ok && (c.Kind == model.ChatDirect || c.Saved) {
			return model.Chat{}, ErrCannotLeave
		}
	}
	if err := cm.ManageChat(ctx, chat, action); err != nil {
		return model.Chat{}, err
	}
	if action == model.ChatLeave {
		h.apply(acct, model.Event{Kind: model.EvChatRemoved, Chat: &model.Chat{ID: chat}})
		return model.Chat{AccountID: acct, ID: chat}, nil
	}
	f, _ := model.FlagsOf(action)
	c, err := h.applyChatFlags(acct, chat, f)
	if err != nil {
		return model.Chat{}, err
	}
	h.emitAccount(acct) // muting changes the unread count
	return c, nil
}

// applyChatFlags patches a stored chat and tells subscribers.
func (h *Hub) applyChatFlags(acct, chat string, f model.ChatFlags) (model.Chat, error) {
	changed := false
	c, err := h.st.UpdateChat(acct, chat, func(c *model.Chat, ok bool) bool {
		if !ok {
			return false
		}
		was := *c
		f.Apply(c)
		changed = was.Muted != c.Muted || was.Pinned != c.Pinned || was.Archived != c.Archived
		return changed
	})
	if err != nil || c.Title == "" {
		return c, err
	}
	if changed {
		h.mu.Lock()
		if r := h.runs[acct]; r != nil {
			c.Online = r.online[chat]
		}
		h.mu.Unlock()
		h.emit(Event{Kind: EvChat, AccountID: acct, Chat: &c})
	}
	return c, nil
}

// AccountChatFolders are the chat folders of one account.
type AccountChatFolders struct {
	Account string
	Folders []model.ChatFolder
}

// ChatFolders returns the chat folders of the given accounts (all when
// empty), in each account's order. Accounts without folders are left out;
// a failing account is in failed.
func (h *Hub) ChatFolders(ctx context.Context, accounts []string) (out []AccountChatFolders, failed []AccountError) {
	if len(accounts) == 0 {
		for _, a := range h.v.Config().Accounts {
			if !a.Disabled {
				accounts = append(accounts, a.ID)
			}
		}
	}
	for _, id := range accounts {
		fs, err := h.chatFolders(ctx, id)
		if err != nil {
			failed = append(failed, AccountError{Account: id, Err: err})
		} else if len(fs) > 0 {
			out = append(out, AccountChatFolders{Account: id, Folders: fs})
		}
	}
	return out, failed
}

// chatFolders returns an account's folders, from memory when loaded.
func (h *Hub) chatFolders(ctx context.Context, acct string) ([]model.ChatFolder, error) {
	h.mu.Lock()
	r := h.runs[acct]
	var (
		sess   provider.Session
		cached []model.ChatFolder
		loaded bool
	)
	if r != nil {
		sess, cached, loaded = r.sess, r.folders, r.foldersLoaded
	}
	h.mu.Unlock()
	cf, ok := sess.(provider.ChatFolders)
	if !ok {
		return nil, nil
	}
	if loaded {
		return cached, nil
	}
	fs, err := cf.ChatFolders(ctx)
	if err != nil {
		return nil, err
	}
	h.keepFolders(acct, fs)
	return fs, nil
}

// keepFolders remembers an account's folders and tells whether they changed.
func (h *Hub) keepFolders(acct string, fs []model.ChatFolder) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.runs[acct]
	if r == nil {
		return false
	}
	changed := !r.foldersLoaded || !slices.EqualFunc(r.folders, fs, sameFolder)
	r.folders, r.foldersLoaded = fs, true
	return changed
}

func sameFolder(a, b model.ChatFolder) bool {
	return a.ID == b.ID && a.Title == b.Title && a.Emoji == b.Emoji && slices.Equal(a.Always, b.Always) &&
		slices.Equal(a.Pinned, b.Pinned) && a.ExcludeMuted == b.ExcludeMuted && a.ExcludeRead == b.ExcludeRead &&
		a.ExcludeArchived == b.ExcludeArchived
}

// loadChatFolders reloads an account's folders from its session and tells
// subscribers when they changed.
func (h *Hub) loadChatFolders(ctx context.Context, acct string) {
	h.mu.Lock()
	r := h.runs[acct]
	var sess provider.Session
	if r != nil {
		sess = r.sess
	}
	h.mu.Unlock()
	cf, ok := sess.(provider.ChatFolders)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	fs, err := cf.ChatFolders(ctx)
	if err != nil {
		if h.ctx.Err() == nil {
			h.logf(eventlog.Warn, r.acct.Provider, acct, "chat folders: %v", err)
		}
		return
	}
	if h.keepFolders(acct, fs) {
		h.emit(Event{Kind: EvChatFolders, AccountID: acct})
	}
}

// foldersChanged: the session says the folders changed. Their rules decide
// which chats are in them, so the chat list is reloaded too.
func (h *Hub) foldersChanged(acct string) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.loadChatFolders(h.ctx, acct)
		if err := h.refresh(h.ctx, acct); err != nil && h.ctx.Err() == nil {
			h.logf(eventlog.Warn, "hub", acct, "chat list: %v", err)
		}
	}()
}
