package vk

// messages.* calls and the JSON shapes VK returns for them. These follow
// the public VK API; peer ids are VK's: a positive id below 2e9 is a user,
// 2000000000+<n> a group chat, a negative id a community.

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// vkGroup is one community in a messages.* "groups" list.
type vkGroup struct {
	ID         int64  `json:"id"` // positive here; peer ids negate it
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	Photo      string `json:"photo_100"`
	IsAdmin    int    `json:"is_admin"`
}

// vkAttachment is one attachment of a message (only the kinds we map).
type vkAttachment struct {
	Type  string `json:"type"`
	Photo *struct {
		Sizes []struct {
			Type   string `json:"type"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"sizes"`
	} `json:"photo,omitempty"`
	Doc *struct {
		Title string `json:"title"`
		Size  int64  `json:"size"`
		Ext   string `json:"ext"`
		URL   string `json:"url"`
	} `json:"doc,omitempty"`
	AudioMsg *struct {
		Duration int    `json:"duration"`
		LinkMP3  string `json:"link_mp3"`
	} `json:"audio_message,omitempty"`
	Sticker *struct {
		Images []struct {
			URL   string `json:"url"`
			Width int    `json:"width"`
		} `json:"images"`
	} `json:"sticker,omitempty"`
}

// vkMessage is a message object.
type vkMessage struct {
	ID         int64          `json:"id"`
	ConvMsgID  int64          `json:"conversation_message_id"`
	Date       int64          `json:"date"`
	UpdateTime int64          `json:"update_time"`
	PeerID     int64          `json:"peer_id"`
	FromID     int64          `json:"from_id"`
	Text       string         `json:"text"`
	Out        int            `json:"out"`
	ReplyMsg   *vkMessage     `json:"reply_message,omitempty"`
	Fwd        []vkMessage    `json:"fwd_messages,omitempty"`
	Attach     []vkAttachment `json:"attachments,omitempty"`
	Action     *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"action,omitempty"`
}

// vkConversation is the state of one chat.
type vkConversation struct {
	Peer struct {
		ID   int64  `json:"id"`
		Type string `json:"type"` // user, chat, group
	} `json:"peer"`
	LastMsgID   int64 `json:"last_message_id"`
	InRead      int64 `json:"in_read"`
	OutRead     int64 `json:"out_read"`
	UnreadCount int   `json:"unread_count"`
	CanWrite    struct {
		Allowed bool `json:"allowed"`
	} `json:"can_write"`
	SortID struct {
		MajorID int64 `json:"major_id"`
	} `json:"sort_id"`
	PushSettings *struct {
		DisabledForever bool  `json:"disabled_forever"`
		DisabledUntil   int64 `json:"disabled_until"`
	} `json:"push_settings,omitempty"`
	ChatSettings *struct {
		Title    string `json:"title"`
		MembersC int    `json:"members_count"`
		Photo    *struct {
			Photo100 string `json:"photo_100"`
		} `json:"photo,omitempty"`
	} `json:"chat_settings,omitempty"`
}

// conversations is messages.getConversations' response.
type conversations struct {
	Count int `json:"count"`
	Items []struct {
		Conversation vkConversation `json:"conversation"`
		LastMessage  *vkMessage     `json:"last_message,omitempty"`
	} `json:"items"`
	Profiles []vkUser  `json:"profiles"`
	Groups   []vkGroup `json:"groups"`
}

func (s *server) getConversations(ctx context.Context, accessToken string, count, offset int) (conversations, error) {
	raw, err := s.method(ctx, "messages.getConversations", accessToken, map[string]string{
		"count": strconv.Itoa(count), "offset": strconv.Itoa(offset),
		"extended": "1", "fields": "photo_100,screen_name,online",
	})
	if err != nil {
		return conversations{}, err
	}
	var c conversations
	if err := json.Unmarshal(raw, &c); err != nil {
		return conversations{}, fmt.Errorf("vk: getConversations: %w", err)
	}
	return c, nil
}

// history is messages.getHistory's response.
type history struct {
	Count    int         `json:"count"`
	Items    []vkMessage `json:"items"`
	Profiles []vkUser    `json:"profiles"`
	Groups   []vkGroup   `json:"groups"`
}

func (s *server) getHistory(ctx context.Context, accessToken string, peerID int64, startMsgID int64, count, offset int) (history, error) {
	params := map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10),
		"count":   strconv.Itoa(count), "offset": strconv.Itoa(offset),
		"extended": "1", "fields": "photo_100,screen_name,online",
	}
	if startMsgID > 0 {
		params["start_message_id"] = strconv.FormatInt(startMsgID, 10)
	}
	raw, err := s.method(ctx, "messages.getHistory", accessToken, params)
	if err != nil {
		return history{}, err
	}
	var h history
	if err := json.Unmarshal(raw, &h); err != nil {
		return history{}, fmt.Errorf("vk: getHistory: %w", err)
	}
	return h, nil
}

// getByID fetches full messages by their global ids (for long-poll events).
func (s *server) getByID(ctx context.Context, accessToken string, ids []int64) ([]vkMessage, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatInt(id, 10)
	}
	raw, err := s.method(ctx, "messages.getById", accessToken, map[string]string{
		"message_ids": strings.Join(strs, ","), "extended": "1",
		"fields": "photo_100,screen_name,online",
	})
	if err != nil {
		return nil, err
	}
	var r struct {
		Items []vkMessage `json:"items"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("vk: getById: %w", err)
	}
	return r.Items, nil
}

// sendOpts are the optional parts of messages.send.
type sendOpts struct {
	replyTo     string
	attachment  string // comma-separated "photo…"/"doc…" refs
	forwardFrom int64  // forward the forwardIDs from this peer
	forwardIDs  []string
}

// send posts a message and returns the new message's global id. Without
// forward_messages/peer_ids VK answers with the bare message id.
func (s *server) send(ctx context.Context, accessToken string, peerID int64, text string, opt sendOpts) (int64, error) {
	params := map[string]string{
		"peer_id":   strconv.FormatInt(peerID, 10),
		"message":   text,
		"random_id": randomID(),
	}
	if opt.replyTo != "" {
		params["reply_to"] = opt.replyTo
	}
	if opt.attachment != "" {
		params["attachment"] = opt.attachment
	}
	if len(opt.forwardIDs) > 0 {
		// forward_messages is simplest for messages of one chat; VK accepts
		// global ids here and copies them into peerID.
		params["forward_messages"] = strings.Join(opt.forwardIDs, ",")
	}
	raw, err := s.method(ctx, "messages.send", accessToken, params)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := json.Unmarshal(raw, &id); err != nil {
		return 0, fmt.Errorf("vk: send: %w", err)
	}
	return id, nil
}

// edit replaces the text of a message (global id) in a peer.
func (s *server) edit(ctx context.Context, accessToken string, peerID, msgID int64, text string) error {
	_, err := s.method(ctx, "messages.edit", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "message_id": strconv.FormatInt(msgID, 10),
		"message": text, "keep_forward_messages": "1", "keep_snippets": "1",
	})
	return err
}

// del deletes messages (global ids). forAll revokes them for everyone.
func (s *server) del(ctx context.Context, accessToken string, msgIDs []string, forAll bool) error {
	p := map[string]string{"message_ids": strings.Join(msgIDs, ",")}
	if forAll {
		p["delete_for_all"] = "1"
	}
	_, err := s.method(ctx, "messages.delete", accessToken, p)
	return err
}

// member is one entry of messages.getConversationMembers.
type member struct {
	MemberID int64 `json:"member_id"`
	IsAdmin  bool  `json:"is_admin"`
}

// convMembers is messages.getConversationMembers' response.
type convMembers struct {
	Count    int       `json:"count"`
	Items    []member  `json:"items"`
	Profiles []vkUser  `json:"profiles"`
	Groups   []vkGroup `json:"groups"`
}

func (s *server) getConversationMembers(ctx context.Context, accessToken string, peerID int64) (convMembers, error) {
	raw, err := s.method(ctx, "messages.getConversationMembers", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "fields": "photo_100,screen_name,online",
	})
	if err != nil {
		return convMembers{}, err
	}
	var m convMembers
	if err := json.Unmarshal(raw, &m); err != nil {
		return convMembers{}, fmt.Errorf("vk: getConversationMembers: %w", err)
	}
	return m, nil
}

// convByID is messages.getConversationsById (extended) for a group chat.
type convByID struct {
	Items    []vkConversation `json:"items"`
	Profiles []vkUser         `json:"profiles"`
	Groups   []vkGroup        `json:"groups"`
}

func (s *server) getConversationByID(ctx context.Context, accessToken string, peerID int64) (vkConversation, error) {
	raw, err := s.method(ctx, "messages.getConversationsById", accessToken, map[string]string{
		"peer_ids": strconv.FormatInt(peerID, 10), "extended": "1", "fields": "photo_100,screen_name,online",
	})
	if err != nil {
		return vkConversation{}, err
	}
	var r convByID
	if err := json.Unmarshal(raw, &r); err != nil {
		return vkConversation{}, fmt.Errorf("vk: getConversationsById: %w", err)
	}
	if len(r.Items) == 0 {
		return vkConversation{}, fmt.Errorf("vk: getConversationsById: no such chat")
	}
	return r.Items[0], nil
}

// vkFullUser adds profile fields to vkUser for messages.getConversationsById
// users and users.get with fields.
type vkFullUser struct {
	vkUser
	Status   string `json:"status"`
	Photo200 string `json:"photo_200"`
	LastSeen *struct {
		Time int64 `json:"time"`
	} `json:"last_seen,omitempty"`
	City *struct {
		Title string `json:"title"`
	} `json:"city,omitempty"`
	Mobile string `json:"mobile_phone"`
	Site   string `json:"site"`
	Bot    int    `json:"is_bot"`
}

func (s *server) usersGetFull(ctx context.Context, accessToken, ids string) ([]vkFullUser, error) {
	raw, err := s.method(ctx, "users.get", accessToken, map[string]string{
		"user_ids": ids,
		"fields":   "photo_100,photo_200,screen_name,online,last_seen,status,city,contacts,site",
	})
	if err != nil {
		return nil, err
	}
	var users []vkFullUser
	if err := json.Unmarshal(raw, &users); err != nil {
		return nil, fmt.Errorf("vk: users.get: %w", err)
	}
	return users, nil
}

// vkFullGroup is groups.getById's shape.
type vkFullGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ScreenName  string `json:"screen_name"`
	Photo200    string `json:"photo_200"`
	Description string `json:"description"`
	MembersC    int    `json:"members_count"`
}

// --- Chat management -----------------------------------------------------

// setSilence mutes (time -1 forever, sound off) or unmutes a conversation.
func (s *server) setSilence(ctx context.Context, accessToken string, peerID int64, mute bool) error {
	p := map[string]string{"peer_id": strconv.FormatInt(peerID, 10)}
	if mute {
		p["time"], p["sound"] = "-1", "0"
	} else {
		p["time"], p["sound"] = "0", "1"
	}
	_, err := s.method(ctx, "account.setSilenceMode", accessToken, p)
	return err
}

func (s *server) pinConversation(ctx context.Context, accessToken string, peerID int64, pin bool) error {
	name := "messages.unpinConversation"
	if pin {
		name = "messages.pinConversation"
	}
	_, err := s.method(ctx, name, accessToken, map[string]string{"peer_id": strconv.FormatInt(peerID, 10)})
	return err
}

func (s *server) archiveConversation(ctx context.Context, accessToken string, peerID int64, archive bool) error {
	name := "messages.unarchiveConversation"
	if archive {
		name = "messages.archiveConversation"
	}
	_, err := s.method(ctx, name, accessToken, map[string]string{"peer_id": strconv.FormatInt(peerID, 10)})
	return err
}

// removeChatUser drops a member (here: the account itself) from a group
// chat. chatLocalID is the peer id minus chatPeerBase.
func (s *server) removeChatUser(ctx context.Context, accessToken string, chatLocalID, userID int64) error {
	_, err := s.method(ctx, "messages.removeChatUser", accessToken, map[string]string{
		"chat_id": strconv.FormatInt(chatLocalID, 10), "user_id": strconv.FormatInt(userID, 10),
	})
	return err
}

// --- Reactions (UNVERIFIED) ----------------------------------------------
// VK reactions are keyed by conversation_message_id (cmid), not the global
// message id this provider uses elsewhere; the caller resolves the cmid
// first. Reaction ids are VK's own numeric ids (the emoji are VK images, not
// Unicode): see reactionName / reactionID for how they map to model names.

func (s *server) sendReaction(ctx context.Context, accessToken string, peerID, cmid int64, reactionID int) error {
	_, err := s.method(ctx, "messages.sendReaction", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "cmid": strconv.FormatInt(cmid, 10),
		"reaction_id": strconv.Itoa(reactionID),
	})
	return err
}

func (s *server) deleteReaction(ctx context.Context, accessToken string, peerID, cmid int64) error {
	_, err := s.method(ctx, "messages.deleteReaction", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "cmid": strconv.FormatInt(cmid, 10),
	})
	return err
}

// reactionCounts is one message's reactions in messages.getMessagesReactions.
type reactionCounts struct {
	CMID   int64 `json:"conversation_message_id"`
	Counts []struct {
		ReactionID int  `json:"reaction_id"`
		Count      int  `json:"count"`
		Reacted    bool `json:"reacted"`
	} `json:"counts"`
	MyReaction int `json:"reaction_id"` // the account's own, when extended
}

func (s *server) getMessagesReactions(ctx context.Context, accessToken string, peerID int64, cmids []int64) ([]reactionCounts, error) {
	strs := make([]string, len(cmids))
	for i, c := range cmids {
		strs[i] = strconv.FormatInt(c, 10)
	}
	raw, err := s.method(ctx, "messages.getMessagesReactions", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "cmids": strings.Join(strs, ","), "extended": "1",
	})
	if err != nil {
		return nil, err
	}
	var r struct {
		Items []reactionCounts `json:"items"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("vk: getMessagesReactions: %w", err)
	}
	return r.Items, nil
}

func (s *server) groupsGetByID(ctx context.Context, accessToken string, groupID int64) (vkFullGroup, error) {
	raw, err := s.method(ctx, "groups.getById", accessToken, map[string]string{
		"group_id": strconv.FormatInt(groupID, 10),
		"fields":   "description,members_count,photo_200,screen_name",
	})
	if err != nil {
		return vkFullGroup{}, err
	}
	// Newer API wraps it in {"groups":[...]}; older returns a bare array.
	var wrapped struct {
		Groups []vkFullGroup `json:"groups"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Groups) > 0 {
		return wrapped.Groups[0], nil
	}
	var arr []vkFullGroup
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return vkFullGroup{}, fmt.Errorf("vk: groups.getById: no group")
	}
	return arr[0], nil
}

// markAsRead marks a peer read up to and including startMsgID (0: all).
func (s *server) markAsRead(ctx context.Context, accessToken string, peerID, startMsgID int64) error {
	params := map[string]string{"peer_id": strconv.FormatInt(peerID, 10)}
	if startMsgID > 0 {
		params["start_message_id"] = strconv.FormatInt(startMsgID, 10)
	}
	_, err := s.method(ctx, "messages.markAsRead", accessToken, params)
	return err
}
