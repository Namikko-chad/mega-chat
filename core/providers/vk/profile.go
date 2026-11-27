package vk

import (
	"context"
	"strconv"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

var _ provider.Profiles = (*session)(nil)

// Profile describes a chat: a person (users.get), a group chat
// (messages.getConversationsById + a members sample) or a community
// (groups.getById). Nothing here is cached.
func (s *session) Profile(ctx context.Context, chatID string) (model.Profile, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return model.Profile{}, err
	}
	switch {
	case peer >= chatPeerBase:
		return s.groupChatProfile(ctx, peer)
	case peer < 0:
		return s.communityProfile(ctx, -peer)
	default:
		return s.userProfile(ctx, peer)
	}
}

func (s *session) userProfile(ctx context.Context, peer int64) (model.Profile, error) {
	users, err := apiCall(ctx, s, func(tok string) ([]vkFullUser, error) {
		return s.srv.usersGetFull(ctx, tok, strconv.FormatInt(peer, 10))
	})
	if err != nil {
		return model.Profile{}, err
	}
	if len(users) == 0 {
		return model.Profile{}, provider.ErrUnsupported
	}
	u := users[0]
	p := model.Profile{
		Title:    u.name(),
		Username: u.ScreenneM,
		About:    u.Status,
		Phone:    u.Mobile,
		Online:   u.Online == 1,
		Bot:      u.Bot == 1,
	}
	if u.Photo200 != "" {
		p.Avatar = u.Photo200
	} else {
		p.Avatar = u.Photo
	}
	if u.City != nil {
		p.Position = u.City.Title
	}
	if u.LastSeen != nil && u.LastSeen.Time > 0 {
		p.LastSeen = time.Unix(u.LastSeen.Time, 0)
	}
	if u.ScreenneM != "" {
		p.Link = "https://vk.com/" + u.ScreenneM
	} else {
		p.Link = "https://vk.com/id" + strconv.FormatInt(peer, 10)
	}
	return p, nil
}

func (s *session) groupChatProfile(ctx context.Context, peer int64) (model.Profile, error) {
	conv, err := apiCall(ctx, s, func(tok string) (vkConversation, error) {
		return s.srv.getConversationByID(ctx, tok, peer)
	})
	if err != nil {
		return model.Profile{}, err
	}
	p := model.Profile{}
	if conv.ChatSettings != nil {
		p.Title = conv.ChatSettings.Title
		p.Members = conv.ChatSettings.MembersC
		if conv.ChatSettings.Photo != nil {
			p.Avatar = conv.ChatSettings.Photo.Photo100
		}
	}
	// A capped sample of members for the avatars and the "@" list.
	if m, err := apiCall(ctx, s, func(tok string) (convMembers, error) {
		return s.srv.getConversationMembers(ctx, tok, peer)
	}); err == nil {
		s.learn(m.Profiles, m.Groups)
		if p.Members == 0 {
			p.Members = m.Count
		}
		for i, mem := range m.Items {
			if i >= 50 {
				break
			}
			p.MemberList = append(p.MemberList, s.user(mem.MemberID))
		}
	}
	return p, nil
}

func (s *session) communityProfile(ctx context.Context, groupID int64) (model.Profile, error) {
	g, err := apiCall(ctx, s, func(tok string) (vkFullGroup, error) {
		return s.srv.groupsGetByID(ctx, tok, groupID)
	})
	if err != nil {
		return model.Profile{}, err
	}
	p := model.Profile{
		Title:    g.Name,
		Username: g.ScreenName,
		About:    g.Description,
		Avatar:   g.Photo200,
		Members:  g.MembersC,
	}
	if g.ScreenName != "" {
		p.Link = "https://vk.com/" + g.ScreenName
	} else {
		p.Link = "https://vk.com/club" + strconv.FormatInt(groupID, 10)
	}
	return p, nil
}
