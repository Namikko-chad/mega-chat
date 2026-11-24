package slack

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"megachat/core/model"
	"megachat/core/richtext"
)

// memberIDs caches conversations.members per chat for a while: typing "@"
// asks on every key.
type memberIDs struct {
	mu   sync.Mutex
	byCh map[string]memberList
}

type memberList struct {
	ids []string
	at  time.Time
}

const (
	membersTTL   = 5 * time.Minute
	membersFetch = 1000 // members asked per chat
	membersNamed = 300  // members whose names are looked up (users.info)

	presenceTTL  = time.Minute
	presenceWait = 2 * time.Second
	presenceAsks = 8 // users.getPresence at once
)

var members = memberIDs{byCh: map[string]memberList{}}

// Members returns people of a channel or direct chat to mention.
func (s *session) Members(ctx context.Context, chatID, query string, limit int) ([]model.User, error) {
	key := s.acct.ID + "\x00" + chatID
	members.mu.Lock()
	l, ok := members.byCh[key]
	members.mu.Unlock()
	if !ok || time.Since(l.at) > membersTTL {
		var r struct {
			Members []string `json:"members"`
		}
		if err := s.call(ctx, "conversations.members", url.Values{"channel": {chatID}, "limit": {"1000"}}, &r); err != nil {
			return nil, err
		}
		l = memberList{ids: r.Members, at: time.Now()}
		if len(l.ids) > membersFetch {
			l.ids = l.ids[:membersFetch]
		}
		members.mu.Lock()
		members.byCh[key] = l
		members.mu.Unlock()
	}
	self := s.params["user_id"]
	var out []model.User
	for i, id := range l.ids {
		if id == self {
			continue
		}
		var u model.User
		if i < membersNamed {
			u = s.userOf(ctx, id)
		} else {
			s.mu.Lock()
			u, ok = s.users[id]
			s.mu.Unlock()
			if !ok {
				continue
			}
		}
		if model.MatchUser(u, query) {
			out = append(out, u)
			if len(out) == limit {
				break
			}
		}
	}
	s.saveUsers()
	s.fillPresence(ctx, out)
	return out, nil
}

// seenPresence is what is known of a person's presence, and since when.
type seenPresence struct {
	on bool
	at time.Time
}

// setPresence remembers the presence of users.
func (s *session) setPresence(users []string, on bool) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.presence == nil {
		s.presence = map[string]seenPresence{}
	}
	for _, id := range users {
		s.presence[id] = seenPresence{on: on, at: now}
	}
}

// fillPresence sets Online of us. RTM reports only the people of direct
// chats, so the others are asked for (users.getPresence, rate-limited:
// kept for presenceTTL). The list never waits longer than presenceWait; a
// failed answer leaves the last known state until the next try.
func (s *session) fillPresence(ctx context.Context, us []model.User) {
	ctx, cancel := context.WithTimeout(ctx, presenceWait)
	defer cancel()
	var wg sync.WaitGroup
	slots := make(chan struct{}, presenceAsks)
	for i := range us {
		u := &us[i]
		s.mu.Lock()
		p, ok := s.presence[u.ID]
		s.mu.Unlock()
		u.Online = p.on
		if ok && time.Since(p.at) < presenceTTL {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-ctx.Done():
				return
			}
			var r struct {
				Presence string `json:"presence"`
			}
			on := p.on
			if err := s.call(ctx, "users.getPresence", url.Values{"user": {u.ID}}, &r); err == nil {
				on = r.Presence == "active"
			}
			s.setPresence([]string{u.ID}, on)
			u.Online = on
		}()
	}
	wg.Wait()
}

// outText is what Slack gets for an outgoing text: a quote as "> " lines,
// mentions as <@id>, Markdown as mrkdwn, the rest escaped.
func outText(o model.Outgoing) string {
	head := ""
	if strings.TrimSpace(o.Quote) != "" {
		q := o
		q.Text = ""
		head = escapeText(q.QuotedText())
	}
	keep := make([]richtext.Range, len(o.Mentions))
	for i, m := range o.Mentions {
		keep[i] = richtext.Range{Offset: m.Offset, Length: m.Length}
	}
	d := richtext.Parse(o.Text, richtext.Markdown, keep...)
	return head + richtext.ToSlack(d, func(i int) string { return "<@" + o.Mentions[i].User.ID + ">" })
}
