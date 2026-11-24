package rpc

import (
	"context"

	"megachat/core/hub"
	"megachat/core/model"
	"megachat/core/rpc/pb"
)

func init() {
	eventKindToPB[hub.EvChatFolders] = pb.Event_KIND_CHAT_FOLDERS
}

// Values of model.ChatAction and pb.ChatAction match.
var _ = [1]struct{}{}[int(model.ChatLeave)-int(pb.ChatAction_CHAT_ACTION_LEAVE)]

func chatActionsToPB(as []model.ChatAction) []pb.ChatAction {
	out := make([]pb.ChatAction, len(as))
	for i, a := range as {
		out[i] = pb.ChatAction(a)
	}
	return out
}

func (s *Service) ManageChat(ctx context.Context, r *pb.ManageChatRequest) (*pb.Chat, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	c, err := h.ManageChat(ctx, r.AccountId, r.ChatId, model.ChatAction(r.Action))
	if err != nil {
		return nil, toStatus(err)
	}
	return chatToPB(&c), nil
}

func (s *Service) ChatFolders(ctx context.Context, r *pb.AccountsRequest) (*pb.ChatFoldersReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	afs, errs := h.ChatFolders(ctx, r.AccountIds)
	out := &pb.ChatFoldersReply{Failed: failedToPB(errs)}
	for _, af := range afs {
		for _, f := range af.Folders {
			out.Folders = append(out.Folders, &pb.ChatFolder{
				AccountId: af.Account, Id: f.ID, Title: f.Title, Emoji: f.Emoji, Always: f.Always, Pinned: f.Pinned,
				ExcludeMuted: f.ExcludeMuted, ExcludeRead: f.ExcludeRead, ExcludeArchived: f.ExcludeArchived,
			})
		}
	}
	return out, nil
}
