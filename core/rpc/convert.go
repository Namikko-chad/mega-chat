package rpc

import (
	"sort"
	"time"

	"megachat/core/hub"
	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/richtext"
	"megachat/core/rpc/pb"
)

func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

func capsToPB(cs []provider.Capability) []pb.Capability {
	out := make([]pb.Capability, len(cs))
	for i, c := range cs {
		out[i] = pb.Capability(c)
	}
	return out
}

func providerToPB(i provider.Info) *pb.ProviderInfo {
	out := &pb.ProviderInfo{Type: i.Type, Name: i.Name, Capabilities: capsToPB(i.Caps)}
	for _, m := range i.Methods {
		out.AuthMethods = append(out.AuthMethods, &pb.AuthMethod{Id: m.ID, Name: m.Name, Description: m.Description})
	}
	return out
}

func loginToPB(r hub.LoginResult) *pb.LoginStepReply {
	out := &pb.LoginStepReply{
		LoginId: r.ID, Title: r.Step.Title, Message: r.Step.Message, Error: r.Step.Error,
		OpenUrl: r.Step.OpenURL, Qr: r.Step.QR, Done: r.Step.Done,
	}
	for _, f := range r.Step.Fields {
		out.Fields = append(out.Fields, &pb.Field{
			Key: f.Key, Label: f.Label, Help: f.Help, Secret: f.Secret, Optional: f.Optional,
			DefaultValue: f.Default, Multiline: f.Multiline,
		})
	}
	if r.Account != nil {
		out.Account = accountToPB(*r.Account)
	}
	return out
}

var statusToPB = map[hub.Status]pb.AccountStatus{
	hub.StatusDisabled:   pb.AccountStatus_ACCOUNT_STATUS_DISABLED,
	hub.StatusConnecting: pb.AccountStatus_ACCOUNT_STATUS_CONNECTING,
	hub.StatusOnline:     pb.AccountStatus_ACCOUNT_STATUS_ONLINE,
	hub.StatusOffline:    pb.AccountStatus_ACCOUNT_STATUS_OFFLINE,
	hub.StatusNeedsLogin: pb.AccountStatus_ACCOUNT_STATUS_NEEDS_LOGIN,
}

// accountToPB never copies Params: credentials do not leave the core.
func accountToPB(a hub.AccountInfo) *pb.Account {
	return &pb.Account{
		Id: a.ID, Provider: a.Provider, Name: a.Name, Login: a.Login,
		Status: statusToPB[a.Status], StatusMessage: a.StatusMsg, Enabled: !a.Disabled,
		Self: userToPB(a.Self), Unread: int32(a.Unread), Capabilities: capsToPB(a.Caps),
		MailUnread: int32(a.MailUnread), ChatActions: chatActionsToPB(a.ChatActions),
	}
}

func userToPB(u model.User) *pb.User {
	return &pb.User{Id: u.ID, Name: u.Name, Username: u.Username, AvatarUrl: u.Avatar, Online: u.Online}
}

func chatToPB(c *model.Chat) *pb.Chat {
	if c == nil {
		return nil
	}
	out := &pb.Chat{
		AccountId: c.AccountID, Id: c.ID, Kind: pb.ChatKind(c.Kind), Title: c.Title, AvatarUrl: c.Avatar,
		Unread: int32(max(c.Unread, 0)), Muted: c.Muted, Pinned: c.Pinned, UpdatedMs: ms(c.Updated),
		ReadOnly: c.Access == model.AccessRead, Forum: c.Forum, Online: c.Online,
		Saved: c.Saved, Archived: c.Archived, Folders: c.Folders,
	}
	if c.Last != nil {
		out.LastMessage = messageToPB(c.Last)
	}
	return out
}

func messageToPB(m *model.Message) *pb.Message {
	if m == nil {
		return nil
	}
	out := &pb.Message{
		AccountId: m.AccountID, ChatId: m.ChatID, Id: m.ID, Sender: userToPB(m.Sender), Text: m.Text,
		TimeMs: ms(m.Time), EditedMs: ms(m.Edited), Outgoing: m.Outgoing, ReplyToId: m.ReplyTo,
		Deleted: m.Deleted, Service: m.Service, InThread: m.InThread, Group: m.Group,
		Mentioned: m.Mentioned, Forwarded: m.Forwarded, ForwardedFrom: m.ForwardedFrom, Quote: m.Quote,
		Format: spansToPB(m.Format),
	}
	if m.Outgoing && !m.Deleted && !m.Service && m.Text != "" {
		if md := richtext.ToMarkdown(m.Text, m.Format); md != m.Text {
			out.Markdown = md
		}
	}
	for _, a := range m.Attachments {
		out.Attachments = append(out.Attachments, &pb.Attachment{
			Kind: pb.AttachmentKind(a.Kind), Name: a.Name, Mime: a.MIME, Size: a.Size,
			Ref: a.Ref, Thumb: a.Thumb, Width: int32(a.Width), Height: int32(a.Height),
			DurationMs: int32(a.Duration.Milliseconds()),
		})
	}
	for _, r := range m.Reactions {
		out.Reactions = append(out.Reactions, &pb.Reaction{Emoji: r.Emoji, Count: int32(r.Count), Mine: r.Mine})
	}
	if t := m.Thread; t != nil {
		out.Thread = &pb.Thread{Id: t.ID, Replies: int32(t.Replies), LastMs: ms(t.Last)}
		for _, u := range t.Users {
			out.Thread.Users = append(out.Thread.Users, userToPB(u))
		}
	}
	return out
}

func spansToPB(in []model.TextSpan) []*pb.TextSpan {
	if len(in) == 0 {
		return nil
	}
	out := make([]*pb.TextSpan, len(in))
	for i, s := range in {
		out[i] = &pb.TextSpan{Style: pb.TextStyle(s.Kind), Offset: int32(s.Offset), Length: int32(s.Length), Url: s.URL, Language: s.Lang}
	}
	return out
}

func topicToPB(t *model.Topic) *pb.Topic {
	return &pb.Topic{
		Id: t.ID, Title: t.Title, Unread: int32(max(t.Unread, 0)), Color: int32(t.Color),
		Pinned: t.Pinned, Closed: t.Closed, Hidden: t.Hidden, LastMessage: messageToPB(t.Last), UpdatedMs: ms(t.Updated),
	}
}

func profileToPB(p *model.Profile) *pb.ChatProfile {
	out := &pb.ChatProfile{
		Title: p.Title, AvatarUrl: p.Avatar, Username: p.Username, About: p.About, Phone: p.Phone,
		Email: p.Email, Position: p.Position, Link: p.Link, Members: int32(max(p.Members, 0)),
		LastSeenMs: ms(p.LastSeen), Online: p.Online, Bot: p.Bot,
	}
	for _, u := range p.MemberList {
		out.MemberList = append(out.MemberList, userToPB(u))
	}
	return out
}

func outgoing(r *pb.SendRequest) model.Outgoing {
	o := model.Outgoing{Text: r.Text, ReplyTo: r.ReplyToId, Thread: r.ThreadId, Quote: r.Quote}
	for _, f := range r.Files {
		o.Files = append(o.Files, model.OutFile{Name: f.Name, MIME: f.Mime, Data: f.Data})
	}
	o.Mentions = mentions(r.Text, r.Mentions)
	return o
}

// mentions turns the app's UTF-16 spans into byte spans of text, sorted;
// a span off the text, empty, without a person or overlapping another is
// dropped (its text goes as it is).
func mentions(text string, in []*pb.Mention) []model.Mention {
	if len(in) == 0 {
		return nil
	}
	// byteAt[i] is the byte offset of UTF-16 unit i (len(text) at the end);
	// -1 inside a surrogate pair.
	byteAt := make([]int, 0, len(text)+1)
	for i, r := range text {
		byteAt = append(byteAt, i)
		if r >= 0x10000 {
			byteAt = append(byteAt, -1)
		}
	}
	byteAt = append(byteAt, len(text))
	var out []model.Mention
	for _, m := range in {
		start, end := int(m.Offset), int(m.Offset)+int(m.Length)
		if m.User == nil || m.User.Id == "" || m.Length <= 0 || start < 0 || end >= len(byteAt) || byteAt[start] < 0 || byteAt[end] < 0 {
			continue
		}
		u := model.User{ID: m.User.Id, Name: m.User.Name, Username: m.User.Username}
		out = append(out, model.Mention{Offset: byteAt[start], Length: byteAt[end] - byteAt[start], User: u})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })
	kept := out[:0]
	for _, m := range out {
		if len(kept) > 0 && m.Offset < kept[len(kept)-1].Offset+kept[len(kept)-1].Length {
			continue
		}
		kept = append(kept, m)
	}
	return kept
}

var eventKindToPB = map[hub.EventKind]pb.Event_Kind{
	hub.EvAccount:       pb.Event_KIND_ACCOUNT,
	hub.EvChat:          pb.Event_KIND_CHAT,
	hub.EvChatRemoved:   pb.Event_KIND_CHAT_REMOVED,
	hub.EvMessageNew:    pb.Event_KIND_MESSAGE_NEW,
	hub.EvMessageEdit:   pb.Event_KIND_MESSAGE_EDIT,
	hub.EvMessageDelete: pb.Event_KIND_MESSAGE_DELETE,
	hub.EvTyping:        pb.Event_KIND_TYPING,
	hub.EvThreadMessage: pb.Event_KIND_THREAD_MESSAGE,
	hub.EvCall:          pb.Event_KIND_CALL,
	hub.EvCallSignal:    pb.Event_KIND_CALL_SIGNAL,
}

func eventToPB(ev hub.Event) *pb.Event {
	out := &pb.Event{Kind: eventKindToPB[ev.Kind], AccountId: ev.AccountID, Chat: chatToPB(ev.Chat), Message: messageToPB(ev.Message), Notify: ev.Notify}
	if ev.Account != nil {
		out.Account = accountToPB(*ev.Account)
	} else if ev.Kind == hub.EvAccount {
		out.Account = &pb.Account{Id: ev.AccountID} // removed
	}
	if ev.Call != nil {
		out.Call = callToPB(ev.AccountID, *ev.Call)
	}
	if ev.Signal != nil {
		out.Signal = signalToPB(*ev.Signal)
	}
	if ev.Mail != nil {
		out.Mail = mailToPB(ev.Mail)
	}
	return out
}

func callToPB(acct string, c model.Call) *pb.Call {
	out := &pb.Call{
		AccountId: acct, Id: c.ID, ChatId: c.ChatID, Outgoing: c.Outgoing, Video: c.Video,
		State: pb.CallState(c.State), End: pb.CallEnd(c.End), Error: c.Error,
		Verification: c.Verification, RelayOnly: c.RelayOnly, Polite: c.Polite,
		Channel: c.Channel, OpenChannel: c.OpenChannel, Offer: c.Offer,
		Remote: callMediaToPB(c.Remote),
	}
	if c.Peer.ID != "" || c.Peer.Name != "" {
		out.Peer = userToPB(c.Peer)
	}
	if !c.Started.IsZero() {
		out.StartedMs = c.Started.UnixMilli()
	}
	for _, s := range c.ICE {
		out.Ice = append(out.Ice, &pb.IceServer{Urls: s.URLs, Username: s.Username, Credential: s.Credential})
	}
	return out
}

func callMediaToPB(m model.CallMedia) *pb.CallMedia {
	return &pb.CallMedia{Muted: m.Muted, Video: m.Video, Screen: m.Screen}
}

func callMedia(m *pb.CallMedia) model.CallMedia {
	return model.CallMedia{Muted: m.GetMuted(), Video: m.GetVideo(), Screen: m.GetScreen()}
}

func signalToPB(s model.CallSignal) *pb.CallSignal {
	return &pb.CallSignal{Kind: pb.SignalKind(s.Kind), Sdp: s.SDP, Mid: s.Mid, Mline: int32(s.MLine), Data: s.Data, Seq: s.Seq}
}

func signal(s *pb.CallSignal) model.CallSignal {
	return model.CallSignal{Kind: model.SignalKind(s.GetKind()), SDP: s.GetSdp(), Mid: s.GetMid(), MLine: int(s.GetMline()), Data: s.GetData()}
}
