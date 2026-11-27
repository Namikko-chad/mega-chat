package telegram

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// maxMsgChats bounds the message id → chat map used to place deletions
// (Telegram reports deleted messages of non-channel chats by id only).
const maxMsgChats = 50000

type session struct {
	acct    provider.Account
	env     provider.Env
	appID   int
	appHash string
	peers   *peerStore
	state   *stateStore

	mu      sync.Mutex
	api     *tg.Client
	gaps    *updates.Manager
	self    model.User
	selfID  int64
	users   map[string]model.User // "u<id>"
	titles  map[string]string     // chat id → title
	kinds   map[string]model.ChatKind
	photos  map[string]int64            // chat id → photo id
	access  map[string]model.ChatAccess // chat id → may the account write
	forums  map[string]bool             // chat id → split into topics
	outside map[string]bool             // channel the account is not in → its chat was removed
	msgChat map[int]string
	sink    provider.Sink   // while Run runs
	online  map[string]bool // "u<id>" → reported online since Run connected
	contact map[string]bool // "u<id>" → in the account's contacts
	bots    map[string]bool // "u<id>" → a bot
	// filters are the account's chat folders (folders.go); filtersOK: loaded.
	filters   []tg.DialogFilterClass
	filtersOK bool

	ph phone // 1:1 calls
}

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	id, hash, ok := appCredentials(acct.Params)
	if !ok {
		return nil, errors.New("telegram: no app credentials (sign in again with API ID and hash)")
	}
	peers, err := newPeerStore(env.State)
	if err != nil {
		return nil, err
	}
	state, err := newStateStore(env.State)
	if err != nil {
		return nil, err
	}
	s := &session{
		acct: acct, env: env, appID: id, appHash: hash, peers: peers, state: state,
		users: map[string]model.User{}, titles: map[string]string{}, kinds: map[string]model.ChatKind{}, msgChat: map[int]string{},
		photos: map[string]int64{}, access: map[string]model.ChatAccess{}, forums: map[string]bool{},
		outside: map[string]bool{}, online: map[string]bool{}, contact: map[string]bool{}, bots: map[string]bool{},
	}
	s.selfID, _ = strconv.ParseInt(acct.Params["user_id"], 10, 64)
	return s, nil
}

// --- what the session knows about peers ----------------------------------

func (s *session) user(key string) (model.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[key]
	return u, ok
}

func (s *session) title(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t := s.titles[key]; t != "" {
		return t
	}
	if u, ok := s.users[key]; ok {
		return u.Name
	}
	return ""
}

func (s *session) selfUser() model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.self
}

// learn records users and chats a response or update brought along.
func (s *session) learn(users []tg.UserClass, chats []tg.ChatClass) {
	s.mu.Lock()
	if s.contact == nil {
		s.contact, s.bots = map[string]bool{}, map[string]bool{}
	}
	var seen []presence
	for _, uc := range users {
		if u, ok := uc.(*tg.User); ok {
			key := "u" + strconv.FormatInt(u.ID, 10)
			s.users[key] = userOf(u)
			if !u.Min {
				s.peers.set(key, u.AccessHash)
				s.contact[key] = u.Contact || u.MutualContact
			}
			s.bots[key] = u.Bot
			if st, ok := u.GetStatus(); ok && !u.Self && !u.Bot {
				seen = append(seen, presence{key, isOnline(st)})
			}
		}
	}
	defer s.setPresence(seen...)
	for _, cc := range chats {
		switch c := cc.(type) {
		case *tg.Chat:
			key := "g" + strconv.FormatInt(c.ID, 10)
			s.titles[key], s.kinds[key] = c.Title, model.ChatGroup
			s.photos[key] = chatPhotoID(c.Photo)
			s.access[key] = chatAccess(c)
		case *tg.ChatForbidden:
			key := "g" + strconv.FormatInt(c.ID, 10)
			s.titles[key], s.kinds[key] = c.Title, model.ChatGroup
			s.access[key] = model.AccessRead
		case *tg.Channel:
			key := "c" + strconv.FormatInt(c.ID, 10)
			s.titles[key], s.kinds[key] = c.Title, model.ChatGroup
			if c.Broadcast {
				s.kinds[key] = model.ChatChannel
			}
			s.photos[key] = chatPhotoID(c.Photo)
			s.forums[key] = c.Forum && !c.Broadcast
			if !c.Min {
				s.peers.set(key, c.AccessHash)
				s.access[key] = channelAccess(c)
				if _, ok := s.outside[key]; c.Left && !ok {
					s.outside[key] = false
				} else if !c.Left {
					delete(s.outside, key)
				}
			}
		case *tg.ChannelForbidden:
			key := "c" + strconv.FormatInt(c.ID, 10)
			s.titles[key] = c.Title
			s.peers.set(key, c.AccessHash)
			s.access[key] = model.AccessRead
		}
	}
	s.mu.Unlock()
	if err := s.peers.Save(); err != nil {
		s.env.Logf("error", "save peers: %v", err)
	}
}

// presence is whether the user of a direct chat ("u<id>") is online.
type presence struct {
	chat   string
	online bool
}

// isOnline tells whether a Telegram status means online now.
func isOnline(st tg.UserStatusClass) bool {
	on, ok := st.(*tg.UserStatusOnline)
	return ok && time.Unix(int64(on.Expires), 0).After(time.Now())
}

// setPresence reports the presence that changed while Run runs.
func (s *session) setPresence(ps ...presence) {
	if len(ps) == 0 {
		return
	}
	s.mu.Lock()
	sink := s.sink
	var changed []presence
	for _, p := range ps {
		if sink != nil && s.online[p.chat] != p.online {
			s.online[p.chat] = p.online
			changed = append(changed, p)
		}
	}
	s.mu.Unlock()
	for _, p := range changed {
		sink.Emit(model.Event{Kind: model.EvPresence, Chat: &model.Chat{ID: p.chat}, Online: p.online})
	}
}

func (s *session) learnEntities(e tg.Entities) {
	users := make([]tg.UserClass, 0, len(e.Users))
	for _, u := range e.Users {
		users = append(users, u)
	}
	chats := make([]tg.ChatClass, 0, len(e.Chats)+len(e.Channels))
	for _, c := range e.Chats {
		chats = append(chats, c)
	}
	for _, c := range e.Channels {
		chats = append(chats, c)
	}
	s.learn(users, chats)
}

func chatAccess(c *tg.Chat) model.ChatAccess {
	if c.Left || c.Deactivated {
		return model.AccessRead
	}
	if c.Creator {
		return model.AccessWrite
	}
	if _, admin := c.GetAdminRights(); admin {
		return model.AccessWrite
	}
	if r, ok := c.GetDefaultBannedRights(); ok && r.SendMessages {
		return model.AccessRead
	}
	return model.AccessWrite
}

func channelAccess(c *tg.Channel) model.ChatAccess {
	if c.Creator {
		return model.AccessWrite
	}
	admin, isAdmin := c.GetAdminRights()
	if c.Broadcast {
		if isAdmin && admin.PostMessages {
			return model.AccessWrite
		}
		return model.AccessRead
	}
	if isAdmin {
		return model.AccessWrite
	}
	if c.Left && !c.JoinToSend {
		// A public group not joined: it can be read; writing joins it,
		// which the app does not do.
		return model.AccessRead
	}
	if r, ok := c.GetBannedRights(); ok && r.SendMessages {
		return model.AccessRead
	}
	if r, ok := c.GetDefaultBannedRights(); ok && r.SendMessages {
		return model.AccessRead
	}
	return model.AccessWrite
}

// chatInfo is what the session knows of a chat, for events (Unread -1:
// unknown, the hub keeps its count).
func (s *session) chatInfo(key string) *model.Chat {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := &model.Chat{ID: key, Unread: -1, Kind: s.kinds[key], Title: s.titles[key], Access: s.access[key], Forum: s.forums[key]}
	if id := s.photos[key]; id != 0 {
		c.Avatar = avatarRef(key, id)
	}
	if key[0] == 'u' {
		c.Kind = model.ChatDirect
		c.Access = model.AccessWrite
		if u, ok := s.users[key]; ok {
			c.Title, c.Avatar = u.Name, u.Avatar
		}
		if s.selfID != 0 && key == "u"+strconv.FormatInt(s.selfID, 10) {
			c.Title, c.Avatar, c.Saved = "Saved Messages", "", true
		}
	}
	if c.Title == "" {
		return nil
	}
	if s.filtersOK {
		c.Folders = s.folderIDs(key)
	}
	return c
}

func (s *session) convert(mc tg.MessageClass) (model.Message, bool) {
	m, ok := convertMessage(s, mc)
	if ok && m.ChatID != "" && m.ChatID[0] != 'c' {
		id, _ := strconv.Atoi(m.ID)
		s.mu.Lock()
		if len(s.msgChat) >= maxMsgChats {
			clear(s.msgChat)
		}
		s.msgChat[id] = m.ChatID
		s.mu.Unlock()
	}
	return m, ok && m.ChatID != ""
}

func (s *session) inputPeer(key string) (tg.InputPeerClass, error) {
	kind, id, ok := parseKey(key)
	if !ok {
		return nil, fmt.Errorf("telegram: bad chat id %q", key)
	}
	switch kind {
	case 'g':
		return &tg.InputPeerChat{ChatID: id}, nil
	case 'u':
		if id == s.selfID {
			return &tg.InputPeerSelf{}, nil
		}
		h, ok := s.peers.get(key)
		if !ok {
			return nil, fmt.Errorf("telegram: unknown user %s (refresh the chat list)", key)
		}
		return &tg.InputPeerUser{UserID: id, AccessHash: h}, nil
	default:
		h, ok := s.peers.get(key)
		if !ok {
			return nil, fmt.Errorf("telegram: unknown channel %s (refresh the chat list)", key)
		}
		return &tg.InputPeerChannel{ChannelID: id, AccessHash: h}, nil
	}
}

func (s *session) live() (*tg.Client, *updates.Manager, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.api == nil {
		return nil, nil, provider.ErrOffline
	}
	return s.api, s.gaps, nil
}

// --- running -------------------------------------------------------------

func authErr(err error) error {
	if err != nil && (auth.IsKeyUnregistered(err) || auth.IsUnauthorized(err)) {
		return fmt.Errorf("telegram: %w (%v)", provider.ErrAuth, err)
	}
	return err
}

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	defer s.dropCalls()
	d := tg.NewUpdateDispatcher()
	s.handle(d, sink)
	gaps := updates.New(updates.Config{
		Handler: d, Storage: s.state, AccessHasher: s.peers, UserAccessHasher: s.peers,
		OnTooLong: func() { s.env.Logf("warn", "too many missed updates; the chat list is reloaded") },
	})
	cl := newClient(s.appID, s.appHash, kvSession{s.env.State}, gaps)
	err := cl.Run(ctx, func(ctx context.Context) error {
		st, err := cl.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !st.Authorized {
			return fmt.Errorf("telegram: %w (session ended)", provider.ErrAuth)
		}
		self := userOf(st.User)
		s.mu.Lock()
		s.api, s.gaps, s.self, s.selfID, s.sink = cl.API(), gaps, self, st.User.ID, sink
		s.users[self.ID] = self
		s.online = map[string]bool{} // the hub forgot it when the session went down
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.api, s.gaps, s.sink = nil, nil, nil
			s.mu.Unlock()
		}()
		return gaps.Run(ctx, cl.API(), st.User.ID, updates.AuthOptions{
			OnStart: func(context.Context) { sink.Connected(self) },
		})
	})
	return authErr(err)
}

func (s *session) emitMessage(sink provider.Sink, kind model.EventKind, mc tg.MessageClass) {
	m, ok := s.convert(mc)
	if !ok {
		return
	}
	if s.notMember(sink, m.ChatID) {
		return
	}
	sink.Emit(model.Event{Kind: kind, Message: &m, Chat: s.chatInfo(m.ChatID)})
	if s.forum(m.ChatID) {
		// The topic shows it too.
		t := m
		t.InThread = strconv.Itoa(topicOf(mc))
		sink.Emit(model.Event{Kind: model.EvThreadMessage, Message: &t})
	}
}

// notMember tells whether chat is a channel or group the account is not in.
// Telegram pushes such a chat's messages for a while after its comments
// were read (a channel's discussion group); they are not a chat of the
// account. The first one also removes the chat in case an earlier version
// listed it.
func (s *session) notMember(sink provider.Sink, chat string) bool {
	s.mu.Lock()
	removed, out := s.outside[chat]
	first := out && !removed
	if first {
		s.outside[chat] = true
	}
	s.mu.Unlock()
	if first {
		sink.Emit(model.Event{Kind: model.EvChatRemoved, Chat: &model.Chat{ID: chat}})
	}
	return out
}

func typing(a tg.SendMessageActionClass) bool {
	_, ok := a.(*tg.SendMessageTypingAction)
	return ok
}

func (s *session) handle(d tg.UpdateDispatcher, sink provider.Sink) {
	s.handleCalls(d, sink)
	d.OnNewMessage(func(_ context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		s.learnEntities(e)
		s.emitMessage(sink, model.EvMessageNew, u.Message)
		return nil
	})
	d.OnNewChannelMessage(func(_ context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
		s.learnEntities(e)
		s.emitMessage(sink, model.EvMessageNew, u.Message)
		return nil
	})
	d.OnEditMessage(func(_ context.Context, e tg.Entities, u *tg.UpdateEditMessage) error {
		s.learnEntities(e)
		s.emitMessage(sink, model.EvMessageEdit, u.Message)
		return nil
	})
	d.OnEditChannelMessage(func(_ context.Context, e tg.Entities, u *tg.UpdateEditChannelMessage) error {
		s.learnEntities(e)
		s.emitMessage(sink, model.EvMessageEdit, u.Message)
		return nil
	})
	d.OnDeleteMessages(func(_ context.Context, _ tg.Entities, u *tg.UpdateDeleteMessages) error {
		for _, id := range u.Messages {
			s.mu.Lock()
			chat := s.msgChat[id]
			s.mu.Unlock()
			if chat != "" {
				sink.Emit(model.Event{Kind: model.EvMessageDelete, Message: &model.Message{ChatID: chat, ID: strconv.Itoa(id)}})
			}
		}
		return nil
	})
	d.OnDeleteChannelMessages(func(_ context.Context, _ tg.Entities, u *tg.UpdateDeleteChannelMessages) error {
		chat := "c" + strconv.FormatInt(u.ChannelID, 10)
		for _, id := range u.Messages {
			sink.Emit(model.Event{Kind: model.EvMessageDelete, Message: &model.Message{ChatID: chat, ID: strconv.Itoa(id)}})
		}
		return nil
	})
	d.OnReadHistoryInbox(func(_ context.Context, _ tg.Entities, u *tg.UpdateReadHistoryInbox) error {
		sink.Emit(model.Event{Kind: model.EvRead, Chat: &model.Chat{ID: peerKey(u.Peer), Unread: u.StillUnreadCount}})
		return nil
	})
	d.OnReadChannelInbox(func(_ context.Context, _ tg.Entities, u *tg.UpdateReadChannelInbox) error {
		key := "c" + strconv.FormatInt(u.ChannelID, 10)
		sink.Emit(model.Event{Kind: model.EvRead, Chat: &model.Chat{ID: key, Unread: u.StillUnreadCount}})
		return nil
	})
	d.OnMessageReactions(func(_ context.Context, _ tg.Entities, u *tg.UpdateMessageReactions) error {
		if u.Reactions.Min {
			return nil // without the account's own choice; the next history load brings it
		}
		sink.Emit(model.Event{Kind: model.EvReactions, Message: &model.Message{
			ChatID: peerKey(u.Peer), ID: strconv.Itoa(u.MsgID), Reactions: reactionsOf(u.Reactions),
		}})
		return nil
	})
	s.handleChatSettings(d, sink)
	d.OnUserStatus(func(_ context.Context, _ tg.Entities, u *tg.UpdateUserStatus) error {
		s.mu.Lock()
		self := u.UserID == s.selfID
		s.mu.Unlock()
		if !self {
			s.setPresence(presence{"u" + strconv.FormatInt(u.UserID, 10), isOnline(u.Status)})
		}
		return nil
	})
	d.OnUserTyping(func(_ context.Context, _ tg.Entities, u *tg.UpdateUserTyping) error {
		if typing(u.Action) {
			key := "u" + strconv.FormatInt(u.UserID, 10)
			sink.Emit(model.Event{Kind: model.EvTyping, Message: &model.Message{ChatID: key, Sender: senderOf(s, key)}})
		}
		return nil
	})
	d.OnChatUserTyping(func(_ context.Context, _ tg.Entities, u *tg.UpdateChatUserTyping) error {
		if typing(u.Action) {
			chat := "g" + strconv.FormatInt(u.ChatID, 10)
			sink.Emit(model.Event{Kind: model.EvTyping, Message: &model.Message{ChatID: chat, Sender: senderOf(s, peerKey(u.FromID))}})
		}
		return nil
	})
	d.OnChannelUserTyping(func(_ context.Context, _ tg.Entities, u *tg.UpdateChannelUserTyping) error {
		if typing(u.Action) {
			chat := "c" + strconv.FormatInt(u.ChannelID, 10)
			sink.Emit(model.Event{Kind: model.EvTyping, Message: &model.Message{ChatID: chat, Sender: senderOf(s, peerKey(u.FromID))}})
		}
		return nil
	})
}

// --- Messenger -----------------------------------------------------------

// Chats lists the main chat list and then the archive (folder 1), and the
// Saved Messages chat when neither has it.
func (s *session) Chats(ctx context.Context, limit int) ([]model.Chat, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, err
	}
	if _, err := s.loadFilters(ctx); err != nil {
		s.env.Logf("warn", "telegram: chat folders: %v", err)
	}
	out, err := s.dialogs(ctx, api, 0, limit)
	if err != nil {
		return out, err
	}
	archived, err := s.dialogs(ctx, api, 1, max(limit/3, 50))
	if err != nil {
		s.env.Logf("warn", "telegram: archive: %v", err)
	}
	for i := range archived {
		archived[i].Archived = true
	}
	out = append(out, archived...)
	if !slices.ContainsFunc(out, func(c model.Chat) bool { return c.Saved }) {
		if c, err := s.savedChat(ctx, api); err != nil {
			s.env.Logf("warn", "telegram: saved messages: %v", err)
		} else if c != nil {
			out = append(out, *c)
		}
	}
	return out, nil
}

// dialogs lists up to limit chats of a folder (0: the main list, 1: the
// archive) with the latest activity.
func (s *session) dialogs(ctx context.Context, api *tg.Client, folder, limit int) ([]model.Chat, error) {
	var (
		out        []model.Chat
		offsetPeer tg.InputPeerClass = &tg.InputPeerEmpty{}
		offsetID   int
		offsetDate int
	)
	for len(out) < limit {
		n := min(100, limit-len(out))
		req := &tg.MessagesGetDialogsRequest{OffsetPeer: offsetPeer, OffsetID: offsetID, OffsetDate: offsetDate, Limit: n}
		req.SetFolderID(folder)
		res, err := api.MessagesGetDialogs(ctx, req)
		if err != nil {
			return out, authErr(err)
		}
		page, ok := res.AsModified()
		if !ok {
			break
		}
		s.learn(page.GetUsers(), page.GetChats())
		top := s.topMessages(page.GetMessages())
		dialogs := page.GetDialogs()
		var last *tg.Dialog
		for _, dc := range dialogs {
			dlg, ok := dc.(*tg.Dialog)
			if !ok {
				continue
			}
			last = dlg
			out = append(out, s.dialogChat(dlg, top))
		}
		if _, slice := res.(*tg.MessagesDialogsSlice); !slice || len(dialogs) < n || last == nil {
			break
		}
		if offsetPeer, err = s.inputPeer(peerKey(last.Peer)); err != nil {
			break
		}
		offsetID = last.TopMessage
		offsetDate = 0
		if m, ok := top[peerKey(last.Peer)+"/"+strconv.Itoa(last.TopMessage)]; ok {
			offsetDate = int(m.Time.Unix())
		}
	}
	return out, nil
}

// topMessages converts the last messages of a dialog page by "chat/id".
func (s *session) topMessages(msgs []tg.MessageClass) map[string]model.Message {
	top := map[string]model.Message{}
	for _, mc := range msgs {
		if m, ok := s.convert(mc); ok {
			top[m.ChatID+"/"+m.ID] = m
		}
	}
	return top
}

// dialogChat is a dialog of a list as a chat.
func (s *session) dialogChat(dlg *tg.Dialog, top map[string]model.Message) model.Chat {
	key := peerKey(dlg.Peer)
	c := s.chatInfo(key)
	if c == nil {
		c = &model.Chat{ID: key, Title: key}
	}
	c.Unread, c.Pinned = dlg.UnreadCount, dlg.Pinned
	c.Muted = dlg.NotifySettings.MuteUntil > int(time.Now().Unix())
	c.Archived = dlg.FolderID == 1
	if m, ok := top[key+"/"+strconv.Itoa(dlg.TopMessage)]; ok {
		c.Last, c.Updated = &m, m.Time
	}
	return *c
}

// savedChat is the Saved Messages dialog (nil: there is none yet).
func (s *session) savedChat(ctx context.Context, api *tg.Client) (*model.Chat, error) {
	res, err := api.MessagesGetPeerDialogs(ctx, []tg.InputDialogPeerClass{&tg.InputDialogPeer{Peer: &tg.InputPeerSelf{}}})
	if err != nil {
		return nil, authErr(err)
	}
	s.learn(res.Users, res.Chats)
	top := s.topMessages(res.Messages)
	for _, dc := range res.Dialogs {
		if dlg, ok := dc.(*tg.Dialog); ok {
			c := s.dialogChat(dlg, top)
			return &c, nil
		}
	}
	return nil, nil
}

func (s *session) History(ctx context.Context, chatID, beforeID string, limit int) ([]model.Message, bool, error) {
	api, _, err := s.live()
	if err != nil {
		return nil, false, err
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return nil, false, err
	}
	offset := 0
	if beforeID != "" {
		if offset, err = strconv.Atoi(beforeID); err != nil {
			return nil, false, fmt.Errorf("telegram: bad message id %q", beforeID)
		}
	}
	res, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: peer, OffsetID: offset, Limit: limit})
	if err != nil {
		return nil, false, authErr(err)
	}
	page, ok := res.AsModified()
	if !ok {
		return nil, false, nil
	}
	s.learn(page.GetUsers(), page.GetChats())
	raw := page.GetMessages()
	out := make([]model.Message, 0, len(raw))
	for _, mc := range raw {
		if m, ok := s.convert(mc); ok {
			out = append(out, m)
		}
	}
	return out, len(raw) >= limit, nil
}

func randomID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]))
}

// sendTarget is where a message goes: a chat, a forum topic of it, or the
// comments of a channel post (the channel's discussion group).
type sendTarget struct {
	peer    tg.InputPeerClass
	chat    string // the chat the message appears in
	replyTo string
	top     int    // the forum topic sent to
	root    string // a comment: the post's copy in the group
}

func (s *session) target(ctx context.Context, api *tg.Client, chatID string, o model.Outgoing) (sendTarget, error) {
	t := sendTarget{chat: chatID, replyTo: o.ReplyTo}
	if o.Thread != "" && s.forum(chatID) {
		top, err := strconv.Atoi(o.Thread)
		if err != nil {
			return t, fmt.Errorf("telegram: bad topic id %q", o.Thread)
		}
		t.top = top
	} else if o.Thread != "" {
		// A comment goes to the channel's discussion group, as a reply to
		// the post's copy there (or to another comment).
		post, err := strconv.Atoi(o.Thread)
		if err != nil {
			return t, fmt.Errorf("telegram: bad thread id %q", o.Thread)
		}
		group, top, err := s.discussion(ctx, api, chatID, post)
		if err != nil {
			return t, err
		}
		t.chat, t.root = group, strconv.Itoa(top)
		if t.replyTo == "" {
			t.replyTo = t.root
		}
	}
	peer, err := s.inputPeer(t.chat)
	if err != nil {
		return t, err
	}
	t.peer = peer
	return t, nil
}

// reply is the reply header of a message to t (nil: none), quoting quote.
func (t sendTarget) reply(quote string) tg.InputReplyToClass {
	if t.replyTo != "" {
		id, err := strconv.Atoi(t.replyTo)
		if err != nil {
			return nil
		}
		r := &tg.InputReplyToMessage{ReplyToMsgID: id}
		if t.top != 0 && t.top != generalTopic {
			r.SetTopMsgID(t.top)
		}
		if quote != "" {
			r.SetQuoteText(quote)
		}
		return r
	}
	if t.top != 0 && t.top != generalTopic {
		// Into a topic: a "reply" to the message that opened it.
		return &tg.InputReplyToMessage{ReplyToMsgID: t.top}
	}
	return nil
}

// quote is the native quote of o (empty: none).
func (t sendTarget) quote(o model.Outgoing) string {
	if t.replyTo == "" {
		return ""
	}
	return strings.TrimSpace(o.Quote)
}

// caption is o's text, with the quote as "> " lines when it is not sent as
// a native one.
func caption(o model.Outgoing, quote string) string {
	if quote == "" && o.Quote != "" {
		return o.QuotedText()
	}
	return o.Text
}

func (s *session) Send(ctx context.Context, chatID string, o model.Outgoing) (model.Message, error) {
	api, gaps, err := s.live()
	if err != nil {
		return model.Message{}, err
	}
	t, err := s.target(ctx, api, chatID, o)
	if err != nil {
		return model.Message{}, err
	}
	quote := t.quote(o)
	var format []model.TextSpan
	build := func(quote string) *tg.MessagesSendMessageRequest {
		d, ents := s.outgoingText(o, caption(o, quote))
		format = d.Spans
		req := &tg.MessagesSendMessageRequest{Peer: t.peer, Message: d.Text, RandomID: randomID()}
		if len(ents) > 0 {
			req.SetEntities(ents)
		}
		if r := t.reply(quote); r != nil {
			req.SetReplyTo(r)
		}
		return req
	}
	req := build(quote)
	res, err := api.MessagesSendMessage(ctx, req)
	if quote != "" && tgerr.Is(err, "QUOTE_TEXT_INVALID") {
		// The fragment is not the message's own text (a caption of an
		// album part, a media note): quote it in the text instead.
		quote = ""
		req = build("")
		res, err = api.MessagesSendMessage(ctx, req)
	}
	if err != nil {
		return model.Message{}, authErr(err)
	}
	m := model.Message{ChatID: chatID, Text: req.Message, Format: format, ReplyTo: o.ReplyTo, Quote: quote, Outgoing: true, Sender: s.selfUser(), Time: time.Now(), InThread: o.Thread}
	switch r := res.(type) {
	case *tg.UpdateShortSentMessage:
		m.ID, m.Seq, m.Time = strconv.Itoa(r.ID), int64(r.ID), time.Unix(int64(r.Date), 0)
	case *tg.Updates:
		for _, u := range r.Updates {
			if mid, ok := u.(*tg.UpdateMessageID); ok && mid.RandomID == req.RandomID {
				m.ID, m.Seq = strconv.Itoa(mid.ID), int64(mid.ID)
			}
		}
	}
	// Let the updates manager see the pts change, or it would think it
	// missed updates and ask for the difference.
	if err := gaps.Handle(ctx, res); err != nil {
		s.env.Logf("warn", "updates after send: %v", err)
	}
	if m.ID == "" {
		return model.Message{}, errors.New("telegram: sent, but the server returned no message id")
	}
	return m, nil
}

// sent returns the new messages of chat in what a send returned, oldest
// first.
func (s *session) sent(res tg.UpdatesClass, chat string) []model.Message {
	var out []model.Message
	if u, ok := res.(interface {
		GetUpdates() []tg.UpdateClass
		GetUsers() []tg.UserClass
		GetChats() []tg.ChatClass
	}); ok {
		s.learn(u.GetUsers(), u.GetChats())
		for _, up := range u.GetUpdates() {
			var mc tg.MessageClass
			switch x := up.(type) {
			case *tg.UpdateNewMessage:
				mc = x.Message
			case *tg.UpdateNewChannelMessage:
				mc = x.Message
			default:
				continue
			}
			if m, ok := s.convert(mc); ok && m.ChatID == chat {
				out = append(out, m)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out
}

// Forward forwards messages with messages.forwardMessages: attachments go
// along, Telegram marks the copies with their original author.
func (s *session) Forward(ctx context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error) {
	api, gaps, err := s.live()
	if err != nil {
		return nil, err
	}
	from, err := s.inputPeer(fromChat)
	if err != nil {
		return nil, err
	}
	to, err := s.inputPeer(toChat)
	if err != nil {
		return nil, err
	}
	req := &tg.MessagesForwardMessagesRequest{FromPeer: from, ToPeer: to}
	for _, id := range msgIDs {
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("telegram: bad message id %q", id)
		}
		req.ID = append(req.ID, n)
		req.RandomID = append(req.RandomID, randomID())
	}
	res, err := api.MessagesForwardMessages(ctx, req)
	if err != nil {
		return nil, authErr(err)
	}
	out := s.sent(res, toChat)
	if err := gaps.Handle(ctx, res); err != nil {
		s.env.Logf("warn", "updates after forward: %v", err)
	}
	return out, nil
}

func (s *session) MarkRead(ctx context.Context, chatID, msgID string) error {
	api, gaps, err := s.live()
	if err != nil {
		return err
	}
	maxID := 0
	if msgID != "" {
		if maxID, err = strconv.Atoi(msgID); err != nil {
			return fmt.Errorf("telegram: bad message id %q", msgID)
		}
	}
	peer, err := s.inputPeer(chatID)
	if err != nil {
		return err
	}
	if ch, ok := peer.(*tg.InputPeerChannel); ok {
		_, err := api.ChannelsReadHistory(ctx, &tg.ChannelsReadHistoryRequest{
			Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash}, MaxID: maxID,
		})
		return authErr(err)
	}
	aff, err := api.MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{Peer: peer, MaxID: maxID})
	if err != nil {
		return authErr(err)
	}
	return gaps.HandleAffected(ctx, 0, aff.Pts, aff.PtsCount)
}

// LogOut ends the session on Telegram's side.
func (s *session) LogOut(ctx context.Context) error {
	if api, _, err := s.live(); err == nil {
		_, err := api.AuthLogOut(ctx)
		return err
	}
	cl := newClient(s.appID, s.appHash, kvSession{s.env.State}, nil)
	return cl.Run(ctx, func(ctx context.Context) error {
		_, err := cl.API().AuthLogOut(ctx)
		return err
	})
}

var _ telegram.UpdateHandler = (*updates.Manager)(nil)
