package slack

import (
	"context"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"megachat/core/model"
)

// --- Profile (a chat's info page) --------------------------------------------

// profileMembers is how many members of a channel a profile lists: each
// unknown one costs a users.info call.
const profileMembers = 20

type wProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsBot   bool   `json:"is_bot"`
	Profile struct {
		DisplayName string `json:"display_name"`
		RealName    string `json:"real_name"`
		Title       string `json:"title"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		StatusText  string `json:"status_text"`
		StatusEmoji string `json:"status_emoji"`
		Image72     string `json:"image_72"`
		Image192    string `json:"image_192"`
	} `json:"profile"`
}

type wConversation struct {
	wChannel
	Topic struct {
		Value string `json:"value"`
	} `json:"topic"`
	NumMembers int `json:"num_members"`
}

func (s *session) Profile(ctx context.Context, chatID string) (model.Profile, error) {
	var c struct {
		Channel wConversation `json:"channel"`
	}
	if err := s.call(ctx, "conversations.info", url.Values{"channel": {chatID}, "include_num_members": {"true"}}, &c); err != nil {
		return model.Profile{}, wrapAuth(err)
	}
	ch := c.Channel
	chat := s.chatOf(ctx, ch.wChannel)
	p := model.Profile{Title: chat.Title, Avatar: chat.Avatar}
	if ch.IsIM {
		var u struct {
			User wProfile `json:"user"`
		}
		if err := s.call(ctx, "users.info", url.Values{"user": {ch.User}}, &u); err != nil {
			return model.Profile{}, wrapAuth(err)
		}
		w := u.User
		p.Username, p.Bot = w.Name, w.IsBot
		p.Phone, p.Email, p.Position = w.Profile.Phone, w.Profile.Email, w.Profile.Title
		p.About = strings.TrimSpace(w.Profile.StatusEmoji + " " + w.Profile.StatusText)
		if w.Profile.Image192 != "" {
			p.Avatar = w.Profile.Image192
		}
		if base := s.params["workspace_url"]; base != "" {
			p.Link = strings.TrimSuffix(base, "/") + "/team/" + url.PathEscape(ch.User)
		}
		return p, nil
	}
	if !ch.IsMPIM {
		p.Username = ch.Name
	}
	p.About = joinNonEmpty(ch.Purpose.Value, ch.Topic.Value)
	p.About = plainText(p.About, s.userName)
	p.Members = ch.NumMembers
	if base := s.params["workspace_url"]; base != "" {
		p.Link = strings.TrimSuffix(base, "/") + "/archives/" + url.PathEscape(ch.ID)
	}
	var m struct {
		Members []string `json:"members"`
	}
	if err := s.call(ctx, "conversations.members", url.Values{"channel": {chatID}, "limit": {strconv.Itoa(profileMembers)}}, &m); err != nil {
		s.env.Logf("info", "members of %s: %v", chatID, err)
	}
	for _, id := range m.Members {
		if len(p.MemberList) == profileMembers {
			break
		}
		p.MemberList = append(p.MemberList, s.userOf(ctx, id))
	}
	s.saveUsers()
	return p, nil
}

// joinNonEmpty joins the non-empty, distinct texts with a blank line.
func joinNonEmpty(texts ...string) string {
	var out []string
	for _, t := range texts {
		if t = strings.TrimSpace(t); t != "" && !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n\n")
}
