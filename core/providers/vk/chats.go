package vk

import (
	"context"
	"fmt"
	"strconv"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

var _ provider.ChatManager = (*session)(nil)

// ChatActions lists what VK can change for a conversation. The hub keeps
// ChatLeave away from direct and saved chats; VK can leave group chats only.
func (s *session) ChatActions() []model.ChatAction {
	return []model.ChatAction{
		model.ChatMute, model.ChatUnmute,
		model.ChatPin, model.ChatUnpin,
		model.ChatArchive, model.ChatUnarchive,
		model.ChatLeave,
	}
}

// ManageChat applies one action to a conversation. The hub stores the
// resulting flags itself (or removes the chat after ChatLeave).
func (s *session) ManageChat(ctx context.Context, chatID string, action model.ChatAction) error {
	peer, err := parsePeer(chatID)
	if err != nil {
		return err
	}
	switch action {
	case model.ChatMute, model.ChatUnmute:
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.setSilence(ctx, tok, peer, action == model.ChatMute)
		})
	case model.ChatPin, model.ChatUnpin:
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.pinConversation(ctx, tok, peer, action == model.ChatPin)
		})
	case model.ChatArchive, model.ChatUnarchive:
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.archiveConversation(ctx, tok, peer, action == model.ChatArchive)
		})
	case model.ChatLeave:
		if peer < chatPeerBase {
			return fmt.Errorf("%w: only group chats can be left", provider.ErrUnsupported)
		}
		self, _ := strconv.ParseInt(s.self.ID, 10, 64)
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.removeChatUser(ctx, tok, peer-chatPeerBase, self)
		})
	default:
		return provider.ErrUnsupported
	}
	return err
}
