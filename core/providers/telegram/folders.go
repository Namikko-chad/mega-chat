package telegram

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// Chat folders (dialog filters), the archive (folder 1) and the account's
// settings of a chat: mute, pin, archive, leave.

// loadFilters fetches the account's chat folders and keeps them for the
// folder ids of chats (folderIDs).
func (s *session) loadFilters(ctx context.Context) ([]tg.DialogFilterClass, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, err
	}
	res, err := api.MessagesGetDialogFilters(ctx)
	if err != nil {
		return nil, authErr(err)
	}
	s.mu.Lock()
	s.filters, s.filtersOK = res.Filters, true
	s.mu.Unlock()
	return res.Filters, nil
}

// ChatFolders lists the account's chat folders ("All chats" left out).
func (s *session) ChatFolders(ctx context.Context) ([]model.ChatFolder, error) {
	fs, err := s.loadFilters(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]model.ChatFolder, 0, len(fs))
	for _, fc := range fs {
		switch f := fc.(type) {
		case *tg.DialogFilter:
			out = append(out, model.ChatFolder{
				ID: strconv.Itoa(f.ID), Title: f.Title.Text, Emoji: f.Emoticon,
				Always: s.peerKeys(f.PinnedPeers, f.IncludePeers), Pinned: s.peerKeys(f.PinnedPeers),
				ExcludeMuted: f.ExcludeMuted, ExcludeRead: f.ExcludeRead, ExcludeArchived: f.ExcludeArchived,
			})
		case *tg.DialogFilterChatlist:
			out = append(out, model.ChatFolder{
				ID: strconv.Itoa(f.ID), Title: f.Title.Text, Emoji: f.Emoticon,
				Always: s.peerKeys(f.PinnedPeers, f.IncludePeers), Pinned: s.peerKeys(f.PinnedPeers),
			})
		}
	}
	return out, nil
}

// peerKeys are the chat ids of input peers; the caller holds s.mu.
func (s *session) peerKeys(lists ...[]tg.InputPeerClass) []string {
	var out []string
	for _, l := range lists {
		for _, p := range l {
			if k := s.inputPeerKey(p); k != "" && !slices.Contains(out, k) {
				out = append(out, k)
			}
		}
	}
	return out
}

// inputPeerKey is the chat id of an input peer; the caller holds s.mu.
func (s *session) inputPeerKey(p tg.InputPeerClass) string {
	switch p := p.(type) {
	case *tg.InputPeerSelf:
		return "u" + strconv.FormatInt(s.selfID, 10)
	case *tg.InputPeerUser:
		return "u" + strconv.FormatInt(p.UserID, 10)
	case *tg.InputPeerUserFromMessage:
		return "u" + strconv.FormatInt(p.UserID, 10)
	case *tg.InputPeerChat:
		return "g" + strconv.FormatInt(p.ChatID, 10)
	case *tg.InputPeerChannel:
		return "c" + strconv.FormatInt(p.ChannelID, 10)
	case *tg.InputPeerChannelFromMessage:
		return "c" + strconv.FormatInt(p.ChannelID, 10)
	}
	return ""
}

// folderIDs are the folders a chat is in by their rules (people by contact
// and bot, groups, channels), the excluded chats left out; chats added by
// hand are the folder's Always. The caller holds s.mu.
func (s *session) folderIDs(key string) []string {
	out := []string{}
	for _, fc := range s.filters {
		f, ok := fc.(*tg.DialogFilter)
		if !ok {
			continue
		}
		var match bool
		switch {
		case key[0] == 'u' && s.bots[key]:
			match = f.Bots
		case key[0] == 'u':
			match = f.Contacts && s.contact[key] || f.NonContacts && !s.contact[key]
		case s.kinds[key] == model.ChatChannel:
			match = f.Broadcasts
		default:
			match = f.Groups
		}
		if match && !slices.Contains(s.peerKeys(f.ExcludePeers), key) {
			out = append(out, strconv.Itoa(f.ID))
		}
	}
	return out
}

// handleChatSettings reports changes of folders, the archive, pins and
// mutes made anywhere.
func (s *session) handleChatSettings(d tg.UpdateDispatcher, sink provider.Sink) {
	folders := func() {
		s.mu.Lock()
		s.filtersOK = false // reloaded with the chat list the hub asks for
		s.mu.Unlock()
		sink.Emit(model.Event{Kind: model.EvChatFolders})
	}
	d.OnDialogFilter(func(context.Context, tg.Entities, *tg.UpdateDialogFilter) error { folders(); return nil })
	d.OnDialogFilters(func(context.Context, tg.Entities, *tg.UpdateDialogFilters) error { folders(); return nil })
	d.OnDialogFilterOrder(func(context.Context, tg.Entities, *tg.UpdateDialogFilterOrder) error { folders(); return nil })
	d.OnFolderPeers(func(_ context.Context, _ tg.Entities, u *tg.UpdateFolderPeers) error {
		for _, fp := range u.FolderPeers {
			archived := fp.FolderID == 1
			sink.Emit(model.Event{Kind: model.EvChatFlags, Chat: &model.Chat{ID: peerKey(fp.Peer)}, Flags: &model.ChatFlags{Archived: &archived}})
		}
		return nil
	})
	d.OnDialogPinned(func(_ context.Context, _ tg.Entities, u *tg.UpdateDialogPinned) error {
		if p, ok := u.Peer.(*tg.DialogPeer); ok {
			pinned := u.Pinned
			sink.Emit(model.Event{Kind: model.EvChatFlags, Chat: &model.Chat{ID: peerKey(p.Peer)}, Flags: &model.ChatFlags{Pinned: &pinned}})
		}
		return nil
	})
	d.OnNotifySettings(func(_ context.Context, _ tg.Entities, u *tg.UpdateNotifySettings) error {
		if p, ok := u.Peer.(*tg.NotifyPeer); ok {
			until, _ := u.NotifySettings.GetMuteUntil()
			muted := until > int(time.Now().Unix())
			sink.Emit(model.Event{Kind: model.EvChatFlags, Chat: &model.Chat{ID: peerKey(p.Peer)}, Flags: &model.ChatFlags{Muted: &muted}})
		}
		return nil
	})
}

func (s *session) ChatActions() []model.ChatAction {
	return []model.ChatAction{
		model.ChatMute, model.ChatUnmute, model.ChatPin, model.ChatUnpin,
		model.ChatArchive, model.ChatUnarchive, model.ChatLeave,
	}
}

func (s *session) ManageChat(ctx context.Context, chatID string, action model.ChatAction) error {
	api, gaps, err := s.live()
	if err != nil {
		return err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return err
	}
	var res tg.UpdatesClass
	switch action {
	case model.ChatMute, model.ChatUnmute:
		var set tg.InputPeerNotifySettings
		set.SetMuteUntil(0)
		if action == model.ChatMute {
			set.SetMuteUntil(math.MaxInt32) // "forever", as the apps do
		}
		_, err = api.AccountUpdateNotifySettings(ctx, &tg.AccountUpdateNotifySettingsRequest{Peer: &tg.InputNotifyPeer{Peer: peer}, Settings: set})
	case model.ChatPin, model.ChatUnpin:
		_, err = api.MessagesToggleDialogPin(ctx, &tg.MessagesToggleDialogPinRequest{Pinned: action == model.ChatPin, Peer: &tg.InputDialogPeer{Peer: peer}})
	case model.ChatArchive, model.ChatUnarchive:
		folder := 0
		if action == model.ChatArchive {
			folder = 1
		}
		res, err = api.FoldersEditPeerFolders(ctx, []tg.InputFolderPeer{{Peer: peer, FolderID: folder}})
	case model.ChatLeave:
		switch p := peer.(type) {
		case *tg.InputPeerChannel:
			res, err = api.ChannelsLeaveChannel(ctx, &tg.InputChannel{ChannelID: p.ChannelID, AccessHash: p.AccessHash})
		case *tg.InputPeerChat:
			res, err = api.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{ChatID: p.ChatID, UserID: &tg.InputUserSelf{}})
		default:
			return fmt.Errorf("%w: only groups and channels can be left", provider.ErrUnsupported)
		}
	default:
		return provider.ErrUnsupported
	}
	if err != nil {
		return authErr(err)
	}
	if res != nil {
		// Let the updates manager see the pts change.
		if err := gaps.Handle(ctx, res); err != nil {
			s.env.Logf("warn", "updates after a chat change: %v", err)
		}
	}
	return nil
}
