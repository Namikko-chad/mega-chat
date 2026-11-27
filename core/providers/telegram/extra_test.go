package telegram

import (
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

func TestMediaReactionsComments(t *testing.T) {
	photo := &tg.Photo{ID: 1, Sizes: []tg.PhotoSizeClass{
		&tg.PhotoStrippedSize{Type: "i"},
		&tg.PhotoSize{Type: "s", W: 90, H: 60, Size: 900},
		&tg.PhotoSize{Type: "m", W: 320, H: 213, Size: 9000},
		&tg.PhotoSizeProgressive{Type: "y", W: 1280, H: 853, Sizes: []int{100, 90000}},
	}}
	full, thumb, w, h, size := photoSizes(photo.Sizes)
	if full != "y" || thumb != "m" || w != 1280 || h != 853 || size != 90000 {
		t.Fatalf("sizes: %s %s %d %d %d", full, thumb, w, h, size)
	}
	rs := tg.MessageReactions{Results: []tg.ReactionCount{
		{Reaction: &tg.ReactionEmoji{Emoticon: "👍"}, Count: 3},
		{Reaction: &tg.ReactionCustomEmoji{DocumentID: 5}, Count: 1},
		{Reaction: &tg.ReactionEmoji{Emoticon: "🔥"}, Count: 1},
	}}
	rs.Results[2].SetChosenOrder(0)
	msg := &tg.Message{ID: 10, PeerID: &tg.PeerChannel{ChannelID: 9}, Date: 100, Message: "post", Media: &tg.MessageMediaPhoto{Photo: photo}}
	msg.SetReactions(rs)
	msg.SetReplies(tg.MessageReplies{Comments: true, Replies: 4, RecentRepliers: []tg.PeerClass{&tg.PeerUser{UserID: 7}}})
	m, ok := convertMessage(dir{}, msg)
	if !ok {
		t.Fatal("not converted")
	}
	a := m.Attachments[0]
	if a.Kind != model.AttachImage || a.Ref != "m:c9:10" || a.Thumb != "t:c9:10" || a.Width != 1280 {
		t.Fatalf("attachment: %+v", a)
	}
	if len(m.Reactions) != 2 || m.Reactions[0] != (model.Reaction{Emoji: "👍", Count: 3}) || !m.Reactions[1].Mine {
		t.Fatalf("reactions: %+v", m.Reactions)
	}
	if m.Thread == nil || m.Thread.ID != "10" || m.Thread.Replies != 4 || m.Thread.Users[0].Name != "Ann" {
		t.Fatalf("thread: %+v", m.Thread)
	}
	loc, _, _, err := fileLocation(msg.Media, true)
	if l, ok := loc.(*tg.InputPhotoFileLocation); err != nil || !ok || l.ThumbSize != "m" {
		t.Fatalf("thumb location: %#v %v", loc, err)
	}
}

func TestAccess(t *testing.T) {
	ch := &tg.Channel{Broadcast: true}
	if channelAccess(ch) != model.AccessRead {
		t.Fatal("someone's channel is writable")
	}
	ch.SetAdminRights(tg.ChatAdminRights{PostMessages: true})
	if channelAccess(ch) != model.AccessWrite {
		t.Fatal("admin of a channel cannot post")
	}
	g := &tg.Channel{Megagroup: true}
	g.SetDefaultBannedRights(tg.ChatBannedRights{SendMessages: true})
	if channelAccess(g) != model.AccessRead {
		t.Fatal("banned writing is writable")
	}
	if chatAccess(&tg.Chat{}) != model.AccessWrite {
		t.Fatal("plain group is read-only")
	}
}

// A channel's comment group the account is not in is not one of its chats,
// though Telegram pushes its messages after the comments were read.
func TestCommentGroupIsNotAChat(t *testing.T) {
	sink := &recSink{ch: make(chan model.Event, 16)}
	s := &session{
		users: map[string]model.User{}, titles: map[string]string{}, kinds: map[string]model.ChatKind{},
		msgChat: map[int]string{}, photos: map[string]int64{}, access: map[string]model.ChatAccess{},
		forums: map[string]bool{}, outside: map[string]bool{},
		peers: &peerStore{m: map[string]int64{"c9": 5, "c8": 6}},
	}
	s.learn(nil, []tg.ChatClass{
		&tg.Channel{ID: 9, AccessHash: 5, Megagroup: true, Left: true, Title: "comments"},
		&tg.Channel{ID: 8, AccessHash: 6, Megagroup: true, Title: "joined"},
	})
	msg := func(ch int64, id int) *tg.Message {
		return &tg.Message{ID: id, PeerID: &tg.PeerChannel{ChannelID: ch}, Date: 100, Message: "hi"}
	}
	s.emitMessage(sink, model.EvMessageNew, msg(9, 1))
	s.emitMessage(sink, model.EvMessageNew, msg(9, 2))
	s.emitMessage(sink, model.EvMessageNew, msg(8, 3))
	var got []model.Event
	for len(sink.ch) > 0 {
		got = append(got, <-sink.ch)
	}
	if len(got) != 2 || got[0].Kind != model.EvChatRemoved || got[0].Chat.ID != "c9" ||
		got[1].Kind != model.EvMessageNew || got[1].Message.ChatID != "c8" {
		t.Fatalf("events: %+v", got)
	}
	// Joined later: its messages come through.
	s.learn(nil, []tg.ChatClass{&tg.Channel{ID: 9, AccessHash: 5, Megagroup: true, Title: "comments"}})
	s.emitMessage(sink, model.EvMessageNew, msg(9, 4))
	if ev := <-sink.ch; ev.Kind != model.EvMessageNew || ev.Message.ChatID != "c9" {
		t.Fatalf("after joining: %+v", ev)
	}
}

// Users' statuses become the presence of their direct chats, reported
// once per change and only while Run runs.
func TestPresence(t *testing.T) {
	sink := &recSink{ch: make(chan model.Event, 16)}
	s := &session{
		users: map[string]model.User{}, online: map[string]bool{},
		peers: &peerStore{kv: provider.NewMemKV(), m: map[string]int64{}},
	}
	soon := int(time.Now().Add(time.Minute).Unix())
	users := []tg.UserClass{
		&tg.User{ID: 1, AccessHash: 1, FirstName: "On", Status: &tg.UserStatusOnline{Expires: soon}},
		&tg.User{ID: 2, AccessHash: 2, FirstName: "Off", Status: &tg.UserStatusOffline{WasOnline: 1}},
		&tg.User{ID: 3, AccessHash: 3, FirstName: "Me", Self: true, Status: &tg.UserStatusOnline{Expires: soon}},
	}
	for _, u := range users {
		u.(*tg.User).SetFlags()
	}
	s.learn(users, nil)
	if len(sink.ch) != 0 {
		t.Fatal("presence reported without a running session")
	}
	s.sink = sink
	s.learn(users, nil)
	s.learn(users, nil) // nothing changed
	s.setPresence(presence{"u1", isOnline(&tg.UserStatusOffline{})})
	var got []model.Event
	for len(sink.ch) > 0 {
		got = append(got, <-sink.ch)
	}
	if len(got) != 2 || got[0].Kind != model.EvPresence || got[0].Chat.ID != "u1" || !got[0].Online ||
		got[1].Chat.ID != "u1" || got[1].Online {
		t.Fatalf("events: %+v", got)
	}
	if isOnline(&tg.UserStatusOnline{Expires: 1}) {
		t.Fatal("an expired online status is online")
	}
}
