package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/tg"

	"megachat/core/model"
)

// --- Profile (a chat's info page) --------------------------------------------

// profileMembers is how many members a profile lists.
const profileMembers = 50

func (s *session) Profile(ctx context.Context, chatID string) (model.Profile, error) {
	api, _, err := s.live()
	if err != nil {
		return model.Profile{}, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return model.Profile{}, err
	}
	var p model.Profile
	if c := s.chatInfo(chatID); c != nil {
		p.Title, p.Avatar = c.Title, c.Avatar
	}
	switch pr := peer.(type) {
	case *tg.InputPeerSelf, *tg.InputPeerUser:
		var in tg.InputUserClass = &tg.InputUserSelf{}
		if u, ok := pr.(*tg.InputPeerUser); ok {
			in = &tg.InputUser{UserID: u.UserID, AccessHash: u.AccessHash}
		}
		res, err := api.UsersGetFullUser(ctx, in)
		if err != nil {
			return model.Profile{}, authErr(err)
		}
		s.learn(res.Users, res.Chats)
		p.About = res.FullUser.About
		for _, uc := range res.Users {
			if u, ok := uc.(*tg.User); ok && u.ID == res.FullUser.ID {
				userProfile(&p, u)
			}
		}
	case *tg.InputPeerChat:
		res, err := api.MessagesGetFullChat(ctx, pr.ChatID)
		if err != nil {
			return model.Profile{}, authErr(err)
		}
		s.learn(res.Users, res.Chats)
		full, ok := res.FullChat.(*tg.ChatFull)
		if !ok {
			break
		}
		p.About = full.About
		if list, ok := full.Participants.(*tg.ChatParticipants); ok {
			p.Members = len(list.Participants)
			for _, x := range list.Participants {
				if len(p.MemberList) == profileMembers {
					break
				}
				if u, ok := s.user("u" + strconv.FormatInt(x.GetUserID(), 10)); ok {
					p.MemberList = append(p.MemberList, u)
				}
			}
		}
		if inv, ok := full.ExportedInvite.(*tg.ChatInviteExported); ok {
			p.Link = inv.Link
		}
	case *tg.InputPeerChannel:
		ch := &tg.InputChannel{ChannelID: pr.ChannelID, AccessHash: pr.AccessHash}
		res, err := api.ChannelsGetFullChannel(ctx, ch)
		if err != nil {
			return model.Profile{}, authErr(err)
		}
		s.learn(res.Users, res.Chats)
		full, ok := res.FullChat.(*tg.ChannelFull)
		if !ok {
			break
		}
		p.About, p.Members = full.About, full.ParticipantsCount
		if inv, ok := full.ExportedInvite.(*tg.ChatInviteExported); ok {
			p.Link = inv.Link
		}
		broadcast := false
		for _, cc := range res.Chats {
			if c, ok := cc.(*tg.Channel); ok && c.ID == pr.ChannelID {
				broadcast = c.Broadcast
				if name := channelUsername(c); name != "" {
					p.Username, p.Link = name, "https://t.me/"+name
				}
			}
		}
		// A group's members, when the account may see them (a channel's
		// subscribers are for its admins only).
		if !broadcast && full.CanViewParticipants {
			p.MemberList = s.channelMembers(ctx, api, ch)
		}
	default:
		return model.Profile{}, fmt.Errorf("telegram: bad chat id %q", chatID)
	}
	return p, nil
}

// userProfile fills in what a user tells about themselves.
func userProfile(p *model.Profile, u *tg.User) {
	if p.Title == "" {
		p.Title = userName(u)
	}
	if name := userUsername(u); name != "" {
		p.Username, p.Link = name, "https://t.me/"+name
	}
	if u.Phone != "" {
		p.Phone = "+" + u.Phone
	}
	p.Bot = u.Bot
	switch st := u.Status.(type) {
	case *tg.UserStatusOnline:
		p.Online = time.Unix(int64(st.Expires), 0).After(time.Now())
	case *tg.UserStatusOffline:
		p.LastSeen = time.Unix(int64(st.WasOnline), 0)
	}
}

// channelMembers returns the recent members of a supergroup; nothing when
// they are hidden.
func (s *session) channelMembers(ctx context.Context, api *tg.Client, ch *tg.InputChannel) []model.User {
	res, err := api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: ch, Filter: &tg.ChannelParticipantsRecent{}, Limit: profileMembers,
	})
	if err != nil {
		s.env.Logf("info", "members of a channel: %v", err)
		return nil
	}
	list, ok := res.(*tg.ChannelsChannelParticipants)
	if !ok {
		return nil
	}
	s.learn(list.Users, list.Chats)
	var out []model.User
	for _, uc := range list.Users {
		if u, ok := uc.(*tg.User); ok && !u.Deleted {
			out = append(out, userOf(u))
		}
	}
	return out
}

func userUsername(u *tg.User) string {
	if u.Username != "" {
		return u.Username
	}
	for _, n := range u.Usernames {
		if n.Active {
			return n.Username
		}
	}
	return ""
}

func channelUsername(c *tg.Channel) string {
	if c.Username != "" {
		return c.Username
	}
	for _, n := range c.Usernames {
		if n.Active {
			return n.Username
		}
	}
	return ""
}
