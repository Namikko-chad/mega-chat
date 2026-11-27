package vkteams

import (
	"context"

	"github.com/SelfHold/parley/core/model"
)

// Members returns people of a chat to mention: a group's members (as
// getChatInfo lists them), or the person of a direct chat.
func (s *session) Members(ctx context.Context, chatID, query string, limit int) ([]model.User, error) {
	var all []model.User
	if isGroup(chatID) {
		aimsid, err := s.live()
		if err != nil {
			return nil, err
		}
		c, err := s.srv.chatProfile(ctx, aimsid, chatID)
		if err != nil {
			return nil, err
		}
		s.learn(c.Members)
		for _, m := range c.Members {
			if id := m.id(); id != "" {
				on, _ := m.online()
				all = append(all, model.User{ID: id, Name: s.personName(id), Avatar: avatarRef(id), Online: on})
			}
		}
	} else {
		all = append(all, model.User{ID: chatID, Name: s.personName(chatID), Avatar: avatarRef(chatID)})
	}
	self := s.selfUser().ID
	var out []model.User
	for _, u := range all {
		if u.ID != self && model.MatchUser(u, query) {
			out = append(out, u)
			if len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

// outText is the text to send: a quote as "> " lines, a mention as
// "@[aimId]" (UNVERIFIED wire, see mentions in wire.go).
func outText(o model.Outgoing) string {
	head := ""
	if o.Quote != "" {
		q := o
		q.Text = ""
		head = q.QuotedText()
	}
	return head + model.MentionedText(o.Text, o.Mentions, func(m model.Mention, _ string) string { return "@[" + m.User.ID + "]" }, nil)
}
