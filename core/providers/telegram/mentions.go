package telegram

import (
	"context"
	"strconv"
	"unicode/utf16"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
)

// Members returns people of a chat to mention: the person of a direct
// chat, a basic group's members, a supergroup's search by name (a channel's
// subscribers are for its admins only: none then).
func (s *session) Members(ctx context.Context, chatID, query string, limit int) ([]model.User, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return nil, err
	}
	var all []model.User
	switch pr := peer.(type) {
	case *tg.InputPeerUser:
		if u, ok := s.user(chatID); ok {
			all = append(all, u)
		}
	case *tg.InputPeerChat:
		res, err := api.MessagesGetFullChat(ctx, pr.ChatID)
		if err != nil {
			return nil, authErr(err)
		}
		s.learn(res.Users, res.Chats)
		if full, ok := res.FullChat.(*tg.ChatFull); ok {
			if list, ok := full.Participants.(*tg.ChatParticipants); ok {
				for _, x := range list.Participants {
					if u, ok := s.user("u" + strconv.FormatInt(x.GetUserID(), 10)); ok {
						all = append(all, u)
					}
				}
			}
		}
	case *tg.InputPeerChannel:
		res, err := api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: &tg.InputChannel{ChannelID: pr.ChannelID, AccessHash: pr.AccessHash},
			Filter:  &tg.ChannelParticipantsSearch{Q: query}, Limit: limit + 1,
		})
		if err != nil {
			s.env.Logf("info", "members to mention: %v", err)
			return nil, nil
		}
		if list, ok := res.(*tg.ChannelsChannelParticipants); ok {
			s.learn(list.Users, list.Chats)
			for _, uc := range list.Users {
				if u, ok := uc.(*tg.User); ok && !u.Deleted {
					all = append(all, userOf(u))
				}
			}
		}
	}
	out := pickMembers(all, s.selfUser().ID, query, limit)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range out {
		out[i].Online = s.online[out[i].ID] // just refreshed by learn above
	}
	return out, nil
}

// pickMembers keeps the people that match query, not self, up to limit.
func pickMembers(all []model.User, self, query string, limit int) []model.User {
	var out []model.User
	for _, u := range all {
		if u.ID != self && model.MatchUser(u, query) {
			out = append(out, u)
			if len(out) == limit {
				break
			}
		}
	}
	return out
}

// mentionEntities makes the mentions of a text Telegram entities: a span
// that is the person's "@username" is a mention by itself; others name the
// user (inputMessageEntityMentionName, offsets in UTF-16).
func (s *session) mentionEntities(text string, ms []model.Mention) []tg.MessageEntityClass {
	var out []tg.MessageEntityClass
	for _, m := range ms {
		if m.Offset < 0 || m.Offset+m.Length > len(text) {
			continue
		}
		span := text[m.Offset : m.Offset+m.Length]
		if m.User.Username != "" && span == "@"+m.User.Username {
			continue
		}
		peer, err := s.inputPeer(m.User.ID)
		if err != nil {
			continue
		}
		pu, ok := peer.(*tg.InputPeerUser)
		if !ok {
			continue
		}
		out = append(out, &tg.InputMessageEntityMentionName{
			Offset: utf16Len(text[:m.Offset]), Length: utf16Len(span),
			UserID: &tg.InputUser{UserID: pu.UserID, AccessHash: pu.AccessHash},
		})
	}
	return out
}

func utf16Len(s string) int { return len(utf16.Encode([]rune(s))) }
