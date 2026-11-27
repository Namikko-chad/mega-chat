package vk

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

var _ provider.Reactor = (*session)(nil)

// React adds or removes the account's reaction and returns the message's
// reactions afterwards. VK keys reactions by conversation_message_id, which
// is resolved from the global message id first.
//
// UNVERIFIED and partial: VK reactions are the messenger's own images, not
// Unicode, so they are named ":vk-<id>:" here (as the model names other
// non-drawable emoji). Others' reactions are not pushed live — they appear
// in the set React returns. See reactionName / reactionID.
func (s *session) React(ctx context.Context, chatID, msgID, emoji string, on bool) ([]model.Reaction, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(msgID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("vk: bad message id %q", msgID)
	}
	msgs, err := apiCall(ctx, s, func(tok string) ([]vkMessage, error) {
		return s.srv.getByID(ctx, tok, []int64{id})
	})
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 || msgs[0].ConvMsgID == 0 {
		return nil, fmt.Errorf("vk: cannot find message %s to react to", msgID)
	}
	cmid := msgs[0].ConvMsgID

	if on {
		rid, ok := reactionID(emoji)
		if !ok {
			return nil, fmt.Errorf("vk: unknown reaction %q", emoji)
		}
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.sendReaction(ctx, tok, peer, cmid, rid)
		})
	} else {
		_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
			return struct{}{}, s.srv.deleteReaction(ctx, tok, peer, cmid)
		})
	}
	if err != nil {
		return nil, err
	}

	counts, err := apiCall(ctx, s, func(tok string) ([]reactionCounts, error) {
		return s.srv.getMessagesReactions(ctx, tok, peer, []int64{cmid})
	})
	if err != nil {
		return nil, err
	}
	for _, c := range counts {
		if c.CMID == cmid {
			return toReactions(c), nil
		}
	}
	return nil, nil
}

// toReactions maps VK reaction counts to the model's set.
func toReactions(c reactionCounts) []model.Reaction {
	out := make([]model.Reaction, 0, len(c.Counts))
	for _, r := range c.Counts {
		out = append(out, model.Reaction{
			Emoji: reactionName(r.ReactionID),
			Count: r.Count,
			Mine:  r.Reacted || (c.MyReaction != 0 && r.ReactionID == c.MyReaction),
		})
	}
	return out
}

// reactionName names a VK reaction id for the model. VK reactions have no
// Unicode form, so they are ":vk-<id>:" — overridable once verified.
func reactionName(id int) string {
	if e, ok := vkReactionEmoji[id]; ok {
		return e
	}
	return ":vk-" + strconv.Itoa(id) + ":"
}

// reactionID is reactionName's inverse: the VK id for a model reaction name.
func reactionID(emoji string) (int, bool) {
	for id, e := range vkReactionEmoji {
		if e == emoji {
			return id, true
		}
	}
	if s, ok := strings.CutPrefix(emoji, ":vk-"); ok {
		if n, err := strconv.Atoi(strings.TrimSuffix(s, ":")); err == nil {
			return n, true
		}
	}
	if n, err := strconv.Atoi(emoji); err == nil { // a bare id
		return n, true
	}
	return 0, false
}

// vkReactionEmoji maps the few VK reaction ids whose Unicode emoji is known.
// Left empty until verified against messages.getReactionsAssets on a live
// server; everything else falls back to ":vk-<id>:".
var vkReactionEmoji = map[int]string{}
