package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/tg"

	"megachat/core/model"
)

// peerKey returns the chat id of a peer.
func peerKey(p tg.PeerClass) string {
	switch p := p.(type) {
	case *tg.PeerUser:
		return "u" + strconv.FormatInt(p.UserID, 10)
	case *tg.PeerChat:
		return "g" + strconv.FormatInt(p.ChatID, 10)
	case *tg.PeerChannel:
		return "c" + strconv.FormatInt(p.ChannelID, 10)
	}
	return ""
}

// parseKey splits a chat id into its kind letter and number.
func parseKey(key string) (byte, int64, bool) {
	if len(key) < 2 {
		return 0, 0, false
	}
	id, err := strconv.ParseInt(key[1:], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	switch key[0] {
	case 'u', 'g', 'c':
		return key[0], id, true
	}
	return 0, 0, false
}

func userName(u *tg.User) string {
	name := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if name == "" {
		name = u.Username
	}
	if name == "" && u.Deleted {
		name = "Deleted account"
	}
	return name
}

func userOf(u *tg.User) model.User {
	key := "u" + strconv.FormatInt(u.ID, 10)
	out := model.User{ID: key, Name: userName(u), Username: u.Username}
	if p, ok := u.Photo.(*tg.UserProfilePhoto); ok && p.PhotoID != 0 {
		out.Avatar = avatarRef(key, p.PhotoID)
	}
	return out
}

// Media refs: "a:<peer>:<photo id>" is a peer's photo; "m:<chat>:<msg>"
// is the media of a message and "t:<chat>:<msg>" its preview. A message's
// file reference expires, so the message is fetched again on download.

func avatarRef(peer string, photoID int64) string {
	return "a:" + peer + ":" + strconv.FormatInt(photoID, 10)
}

func chatPhotoID(p tg.ChatPhotoClass) int64 {
	if p, ok := p.(*tg.ChatPhoto); ok {
		return p.PhotoID
	}
	return 0
}

// photoSizes picks the largest size of a photo (the file) and a preview
// of about 320 px; wh is the largest size's dimensions.
func photoSizes(sizes []tg.PhotoSizeClass) (full, thumb string, w, h int, size int64) {
	type sz struct {
		typ        string
		w, h, size int
	}
	var all []sz
	for _, s := range sizes {
		switch s := s.(type) {
		case *tg.PhotoSize:
			all = append(all, sz{s.Type, s.W, s.H, s.Size})
		case *tg.PhotoSizeProgressive:
			n := 0
			if len(s.Sizes) > 0 {
				n = s.Sizes[len(s.Sizes)-1]
			}
			all = append(all, sz{s.Type, s.W, s.H, n})
		}
	}
	if len(all) == 0 {
		return "", "", 0, 0, 0
	}
	big, small := all[0], all[0]
	for _, s := range all {
		if s.w*s.h > big.w*big.h {
			big = s
		}
	}
	small = big
	for _, s := range all {
		if max(s.w, s.h) >= 320 && s.w*s.h < small.w*small.h {
			small = s
		}
	}
	return big.typ, small.typ, big.w, big.h, int64(big.size)
}

func attachments(chat string, id int, media tg.MessageMediaClass) []model.Attachment {
	ref := ":" + chat + ":" + strconv.Itoa(id)
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		a := model.Attachment{Kind: model.AttachImage, MIME: "image/jpeg", Ref: "m" + ref, Thumb: "t" + ref}
		if p, ok := m.Photo.(*tg.Photo); ok {
			_, _, a.Width, a.Height, a.Size = photoSizes(p.Sizes)
		} else {
			a.Ref, a.Thumb = "", ""
		}
		return []model.Attachment{a}
	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return nil
		}
		a := model.Attachment{Kind: model.AttachFile, MIME: doc.MimeType, Size: doc.Size, Ref: "m" + ref}
		for _, attr := range doc.Attributes {
			switch at := attr.(type) {
			case *tg.DocumentAttributeFilename:
				a.Name = at.FileName
			case *tg.DocumentAttributeAudio:
				a.Kind = model.AttachAudio
				if at.Voice {
					a.Kind = model.AttachVoice
				}
				a.Duration = time.Duration(at.Duration) * time.Second
				if a.Name == "" && at.Title != "" {
					a.Name = strings.TrimSpace(at.Performer + " — " + at.Title)
				}
			case *tg.DocumentAttributeVideo:
				if a.Kind == model.AttachFile || a.Kind == model.AttachImage {
					a.Kind = model.AttachVideo
				}
				a.Width, a.Height = at.W, at.H
				a.Duration = time.Duration(at.Duration * float64(time.Second))
			case *tg.DocumentAttributeSticker:
				a.Kind = model.AttachSticker
			case *tg.DocumentAttributeImageSize:
				a.Width, a.Height = at.W, at.H
				if a.Kind == model.AttachFile && strings.HasPrefix(doc.MimeType, "image/") {
					a.Kind = model.AttachImage
				}
			case *tg.DocumentAttributeAnimated:
				if a.Kind == model.AttachFile {
					a.Kind = model.AttachVideo // a GIF is sent as an MP4
				}
			}
		}
		if m.Voice {
			a.Kind = model.AttachVoice
		}
		if len(doc.Thumbs) > 0 {
			a.Thumb = "t" + ref
		} else if (a.Kind == model.AttachImage || a.Kind == model.AttachSticker) && doc.MimeType != "application/x-tgsticker" && doc.Size < 1<<20 {
			a.Thumb = a.Ref // small enough to be its own preview
		}
		return []model.Attachment{a}
	}
	return nil
}

// reactionsOf maps a message's reactions; custom emoji (which the app
// cannot draw) are left out.
func reactionsOf(r tg.MessageReactions) []model.Reaction {
	var out []model.Reaction
	for _, rc := range r.Results {
		var e string
		switch x := rc.Reaction.(type) {
		case *tg.ReactionEmoji:
			e = x.Emoticon
		case *tg.ReactionPaid:
			e = "⭐"
		}
		if e == "" {
			continue
		}
		_, mine := rc.GetChosenOrder()
		out = append(out, model.Reaction{Emoji: e, Count: rc.Count, Mine: mine})
	}
	return out
}

// mediaText is a text stand-in for media that is not a file.
func mediaText(media tg.MessageMediaClass) string {
	switch m := media.(type) {
	case *tg.MessageMediaGeo, *tg.MessageMediaGeoLive, *tg.MessageMediaVenue:
		return "[location]"
	case *tg.MessageMediaContact:
		return "[contact] " + strings.TrimSpace(m.FirstName+" "+m.LastName+" "+m.PhoneNumber)
	case *tg.MessageMediaPoll:
		return "[poll] " + m.Poll.Question.Text
	case *tg.MessageMediaDice:
		return m.Emoticon
	}
	return ""
}

// callText describes a call in the history: "Outgoing call · 3:12",
// "Missed video call".
func callText(a *tg.MessageActionPhoneCall, out bool) string {
	kind := "call"
	if a.Video {
		kind = "video call"
	}
	switch a.Reason.(type) {
	case *tg.PhoneCallDiscardReasonMissed:
		if out {
			return "Cancelled " + kind
		}
		return "Missed " + kind
	case *tg.PhoneCallDiscardReasonBusy:
		if out {
			return "Busy"
		}
		return "Declined " + kind
	}
	text := "Incoming " + kind
	if out {
		text = "Outgoing " + kind
	}
	if d := a.Duration; d > 0 {
		if d >= 3600 {
			text += fmt.Sprintf(" · %d:%02d:%02d", d/3600, d/60%60, d%60)
		} else {
			text += fmt.Sprintf(" · %d:%02d", d/60, d%60)
		}
	}
	return text
}

// serviceText describes a service message; out: the account sent it.
func serviceText(a tg.MessageActionClass, out bool) string {
	switch a := a.(type) {
	case *tg.MessageActionChatCreate:
		return "created the group «" + a.Title + "»"
	case *tg.MessageActionChannelCreate:
		return "created the channel «" + a.Title + "»"
	case *tg.MessageActionChatEditTitle:
		return "changed the title to «" + a.Title + "»"
	case *tg.MessageActionChatEditPhoto:
		return "changed the photo"
	case *tg.MessageActionChatAddUser, *tg.MessageActionChatJoinedByLink, *tg.MessageActionChatJoinedByRequest:
		return "joined"
	case *tg.MessageActionChatDeleteUser:
		return "left"
	case *tg.MessageActionPinMessage:
		return "pinned a message"
	case *tg.MessageActionPhoneCall:
		return callText(a, out)
	case *tg.MessageActionGroupCall:
		return "group call"
	case *tg.MessageActionTopicCreate:
		return "created the topic «" + a.Title + "»"
	case *tg.MessageActionTopicEdit:
		switch {
		case a.Title != "":
			return "renamed the topic to «" + a.Title + "»"
		case a.Closed:
			return "closed the topic"
		default:
			return "changed the topic"
		}
	case *tg.MessageActionChatMigrateTo, *tg.MessageActionChannelMigrateFrom:
		return "the group was upgraded to a supergroup"
	}
	return "service message"
}

// dirOf answers who is behind an id, from what the session has seen.
type dirOf interface {
	user(key string) (model.User, bool)
	title(key string) string
	selfUser() model.User
}

// convertMessage maps a Telegram message; ok is false for empty ones.
func convertMessage(d dirOf, mc tg.MessageClass) (model.Message, bool) {
	var (
		out    model.Message
		from   tg.PeerClass
		peer   tg.PeerClass
		outgo  bool
		date   int
		reply  tg.MessageReplyHeaderClass
		postAs string
	)
	switch m := mc.(type) {
	case *tg.Message:
		out.ID, out.Seq = strconv.Itoa(m.ID), int64(m.ID)
		from, peer, outgo, date, reply = m.FromID, m.PeerID, m.Out, m.Date, m.ReplyTo
		out.Text, out.Format = incomingText(m.Message, m.Entities)
		if t := mediaText(m.Media); t != "" {
			shift := utf16Len(t) + 1
			for i := range out.Format {
				out.Format[i].Offset += shift
			}
			out.Text = strings.TrimSpace(t + "\n" + out.Text)
		}
		out.Attachments = attachments(peerKey(m.PeerID), m.ID, m.Media)
		if m.EditDate > 0 && !m.EditHide {
			out.Edited = time.Unix(int64(m.EditDate), 0)
		}
		postAs = m.PostAuthor
		if f, ok := m.GetFwdFrom(); ok {
			out.Forwarded = true
			switch {
			case f.FromName != "":
				out.ForwardedFrom = f.FromName
			case f.FromID != nil:
				out.ForwardedFrom = senderOf(d, peerKey(f.FromID)).Name
			}
			if a, ok := f.GetPostAuthor(); ok && a != "" && out.ForwardedFrom == "" {
				out.ForwardedFrom = a
			}
		}
		// Set by the server for @us and for replies to our messages.
		out.Mentioned = m.Mentioned && !m.Out
		if m.GroupedID != 0 {
			out.Group = strconv.FormatInt(m.GroupedID, 10)
		}
		if r, ok := m.GetReactions(); ok {
			out.Reactions = reactionsOf(r)
		}
		if r, ok := m.GetReplies(); ok && r.Comments {
			t := &model.Thread{ID: strconv.Itoa(m.ID), Replies: r.Replies}
			for _, p := range r.RecentRepliers {
				t.Users = append(t.Users, senderOf(d, peerKey(p)))
			}
			out.Thread = t
		}
	case *tg.MessageService:
		out.ID, out.Seq = strconv.Itoa(m.ID), int64(m.ID)
		from, peer, outgo, date, reply = m.FromID, m.PeerID, m.Out, m.Date, m.ReplyTo
		out.Service, out.Text = true, serviceText(m.Action, m.Out)
	default:
		return model.Message{}, false
	}
	out.ChatID = peerKey(peer)
	out.Time = time.Unix(int64(date), 0)
	out.Outgoing = outgo
	if r, ok := reply.(*tg.MessageReplyHeader); ok && r.ReplyToMsgID != 0 && (!r.ForumTopic || r.ReplyToTopID != 0) {
		// In a forum a message names its topic as what it replies to;
		// only one with a separate topic id is a real reply.
		out.ReplyTo = strconv.Itoa(r.ReplyToMsgID)
		out.Quote, _ = r.GetQuoteText()
	}
	switch {
	case outgo:
		out.Sender = d.selfUser()
	case from != nil:
		out.Sender = senderOf(d, peerKey(from))
	default:
		// No sender: a private chat (the peer wrote) or a channel post.
		out.Sender = senderOf(d, out.ChatID)
	}
	if postAs != "" {
		out.Sender.Name = postAs
	}
	return out, true
}

func senderOf(d dirOf, key string) model.User {
	if u, ok := d.user(key); ok {
		return u
	}
	return model.User{ID: key, Name: d.title(key)}
}
