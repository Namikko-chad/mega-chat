package vk

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

var _ provider.Mentioner = (*session)(nil)

// Members lists people of a chat to mention: a group chat's members, or the
// person of a direct chat. Communities are left out (they are not mentioned
// by "@").
func (s *session) Members(ctx context.Context, chatID, query string, limit int) ([]model.User, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return nil, err
	}
	var all []model.User
	if peer >= chatPeerBase {
		m, err := apiCall(ctx, s, func(tok string) (convMembers, error) {
			return s.srv.getConversationMembers(ctx, tok, peer)
		})
		if err != nil {
			return nil, err
		}
		s.learn(m.Profiles, m.Groups)
		for _, mem := range m.Items {
			if mem.MemberID <= 0 { // skip community members
				continue
			}
			all = append(all, s.user(mem.MemberID))
		}
	} else {
		all = append(all, s.user(peer))
	}
	self := s.self.ID
	var out []model.User
	for _, u := range all {
		if u.ID != self && model.MatchUser(u, query) {
			out = append(out, u)
			if limit > 0 && len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

// outText is the text to send: the quote fragment as "> " lines (VK has no
// native fragment quote), mentions as VK's "[id<n>|name]" / "[club<n>|name]".
func outText(o model.Outgoing) string {
	head := ""
	if o.Quote != "" {
		q := o
		q.Text, q.Mentions = "", nil
		head = q.QuotedText()
	}
	body := model.MentionedText(o.Text, o.Mentions, func(m model.Mention, span string) string {
		return vkMention(m.User.ID, span)
	}, nil)
	return head + body
}

// vkMention formats one mention. A negative id is a community (club).
func vkMention(id, text string) string {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return text
	}
	if n < 0 {
		return "[club" + strconv.FormatInt(-n, 10) + "|" + text + "]"
	}
	return "[id" + strconv.FormatInt(n, 10) + "|" + text + "]"
}

// parseVKText turns VK's inline markup into visible text plus spans: mention
// tokens "[id123|Name]" / "[club12|Name]" become the name linked to the
// profile. Offsets are UTF-16 units, as model.TextSpan wants.
func parseVKText(raw string) (string, []model.TextSpan) {
	if !strings.Contains(raw, "[") {
		return raw, nil
	}
	var b strings.Builder
	var spans []model.TextSpan
	u16 := 0 // visible length so far, in UTF-16 units
	for i := 0; i < len(raw); {
		if raw[i] == '[' {
			if id, name, rest, ok := parseMentionToken(raw[i:]); ok {
				start := u16
				b.WriteString(name)
				u16 += utf16Len(name)
				spans = append(spans, model.TextSpan{
					Kind: model.SpanLink, Offset: start, Length: u16 - start,
					URL: "https://vk.com/" + id,
				})
				i = len(raw) - len(rest)
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(raw[i:])
		b.WriteString(raw[i : i+size])
		u16 += utf16Len(string(r))
		i += size
	}
	return b.String(), spans
}

// utf16Len is the number of UTF-16 code units in s.
func utf16Len(s string) int { return len(utf16.Encode([]rune(s))) }

// parseMentionToken parses a leading "[id123|Name]" or "[club12|Name]".
func parseMentionToken(s string) (id, name, rest string, ok bool) {
	if !strings.HasPrefix(s, "[") {
		return "", "", "", false
	}
	end := strings.IndexByte(s, ']')
	bar := strings.IndexByte(s, '|')
	if end < 0 || bar < 0 || bar > end {
		return "", "", "", false
	}
	tag := s[1:bar]
	name = s[bar+1 : end]
	switch {
	case strings.HasPrefix(tag, "id") && allDigits(tag[2:]):
		id = tag
	case strings.HasPrefix(tag, "club") && allDigits(tag[4:]):
		id = tag
	default:
		return "", "", "", false
	}
	if name == "" {
		return "", "", "", false
	}
	return id, name, s[end+1:], true
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
