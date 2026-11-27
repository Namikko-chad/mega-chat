package vk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

// chatPeerBase is the offset VK adds to a group chat's local id.
const chatPeerBase = 2000000000

type session struct {
	p    *Provider
	acct provider.Account
	env  provider.Env
	srv  *server

	mu     sync.Mutex
	token  string
	anonym anonymToken
	self   model.User
	people map[int64]model.User // user/community id → user (community ids negated)
	sink   provider.Sink        // of the current Run, for SaveParams
}

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	s := &session{
		p: p, acct: acct, env: env, srv: p.server(hc),
		token:  acct.Params[keyAccessToken],
		people: map[int64]model.User{},
	}
	id, _ := strconv.ParseInt(acct.Params[keyUserID], 10, 64)
	s.self = model.User{ID: acct.Params[keyUserID], Name: "VK " + acct.Params[keyUserID]}
	if id != 0 {
		s.people[id] = s.self
	}
	return s, nil
}

func (s *session) accessToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.token
}

// refresh gets a fresh access token from the stored exchange (common)
// token, the silent-token path the official client uses. It returns an
// error wrapping provider.ErrAuth when the exchange token no longer works.
func (s *session) refresh(ctx context.Context) error {
	common := s.acct.Params[keyCommonToken]
	if common == "" {
		return fmt.Errorf("vk: %w (no exchange token; sign in again)", provider.ErrAuth)
	}
	silent, err := s.srv.authByExchangeToken(ctx, common)
	if err != nil {
		if authFailed(err) {
			return fmt.Errorf("vk: %w (%v)", provider.ErrAuth, err)
		}
		return err
	}
	if silent.SilentToken == "" {
		return fmt.Errorf("vk: %w (exchange token rejected)", provider.ErrAuth)
	}
	anonym, err := s.freshAnonym(ctx, false)
	if err != nil {
		return err
	}
	access, err := s.srv.exchangeSilentAuthToken(ctx, anonym.Token, silent.SilentToken, silent.SilentTokenUID)
	if errors.Is(err, errAnonymExpired) {
		if anonym, err = s.freshAnonym(ctx, true); err != nil {
			return err
		}
		access, err = s.srv.exchangeSilentAuthToken(ctx, anonym.Token, silent.SilentToken, silent.SilentTokenUID)
	}
	if err != nil {
		return err
	}
	params := map[string]string{}
	for k, v := range s.acct.Params {
		params[k] = v
	}
	params[keyAccessToken] = access
	s.mu.Lock()
	s.token = access
	s.acct.Params = params
	sink := s.sink
	s.mu.Unlock()
	if sink != nil {
		return sink.SaveParams(params)
	}
	return nil
}

func (s *session) freshAnonym(ctx context.Context, force bool) (anonymToken, error) {
	s.mu.Lock()
	cur := s.anonym
	s.mu.Unlock()
	if !force && cur.valid() {
		return cur, nil
	}
	t, err := s.srv.getAnonymToken(ctx)
	if err != nil {
		return anonymToken{}, err
	}
	s.mu.Lock()
	s.anonym = t
	s.mu.Unlock()
	return t, nil
}

// apiCall runs an API call with the current token and, if it fails because
// the token expired, refreshes once and retries.
func apiCall[T any](ctx context.Context, s *session, fn func(token string) (T, error)) (T, error) {
	v, err := fn(s.accessToken())
	if authFailed(err) {
		if rerr := s.refresh(ctx); rerr != nil {
			var zero T
			return zero, rerr
		}
		return fn(s.accessToken())
	}
	return v, err
}

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	s.mu.Lock()
	s.sink = sink
	s.mu.Unlock()
	// Learn who we are, then report connected.
	if self, err := apiCall(ctx, s, func(tok string) ([]vkUser, error) {
		return s.srv.usersGet(ctx, tok, s.acct.Params[keyUserID])
	}); err == nil && len(self) > 0 {
		u := toUser(self[0])
		s.mu.Lock()
		s.self = u
		s.people[self[0].ID] = u
		s.mu.Unlock()
	} else if err != nil && errors.Is(err, provider.ErrAuth) {
		return err
	}
	s.mu.Lock()
	self := s.self
	s.mu.Unlock()
	sink.Connected(self)

	lp, err := apiCall(ctx, s, func(tok string) (longPollServer, error) {
		return s.srv.getLongPollServer(ctx, tok)
	})
	if err != nil {
		return err
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		reply, err := s.srv.poll(ctx, lp, 25)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		switch reply.Failed {
		case 0:
			if n, perr := reply.TS.Int64(); perr == nil {
				lp.TS = n
			}
			for _, u := range reply.Updates {
				s.handleUpdate(ctx, sink, u)
			}
		case 1: // the ts is stale; the new one is in the reply
			if n, perr := reply.TS.Int64(); perr == nil {
				lp.TS = n
			}
		default: // 2, 3, 4: key/ts died or version unsupported → new server
			lp, err = apiCall(ctx, s, func(tok string) (longPollServer, error) {
				return s.srv.getLongPollServer(ctx, tok)
			})
			if err != nil {
				return err
			}
		}
	}
}

// handleUpdate decodes one long-poll update array and reports it.
func (s *session) handleUpdate(ctx context.Context, sink provider.Sink, raw json.RawMessage) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return
	}
	var code int
	if err := json.Unmarshal(arr[0], &code); err != nil {
		return
	}
	switch code {
	case 4, 5: // new / edited message
		id := rawInt(arr, 1)
		if id == 0 {
			return
		}
		msgs, err := apiCall(ctx, s, func(tok string) ([]vkMessage, error) {
			return s.srv.getByID(ctx, tok, []int64{id})
		})
		if err != nil || len(msgs) == 0 {
			return
		}
		kind := model.EvMessageNew
		if code == 5 {
			kind = model.EvMessageEdit
		}
		m := s.toMessage(msgs[0])
		sink.Emit(model.Event{Kind: kind, Message: &m})
	case 2: // set message flags: a 128 bit means deleted
		id, mask, peer := rawInt(arr, 1), rawInt(arr, 2), rawInt(arr, 3)
		if id != 0 && mask&128 != 0 {
			sink.Emit(model.Event{Kind: model.EvMessageDelete, Message: &model.Message{
				ChatID: strconv.FormatInt(peer, 10), ID: strconv.FormatInt(id, 10),
			}})
		}
	case 6, 7: // read inbound (6) / outbound (7) up to a local id
		peer := rawInt(arr, 1)
		if peer != 0 {
			sink.Emit(model.Event{Kind: model.EvRead, Chat: &model.Chat{
				ID: strconv.FormatInt(peer, 10), Unread: 0,
			}})
		}
	case 8, 9: // a user came online (8) / went offline (9); id is negated
		uid := -rawInt(arr, 1)
		if uid > 0 {
			sink.Emit(model.Event{Kind: model.EvPresence, Online: code == 8, Chat: &model.Chat{
				ID: strconv.FormatInt(uid, 10),
			}})
		}
	}
}

func rawInt(arr []json.RawMessage, i int) int64 {
	if i >= len(arr) {
		return 0
	}
	var n int64
	if err := json.Unmarshal(arr[i], &n); err != nil {
		return 0
	}
	return n
}

func (s *session) Chats(ctx context.Context, limit int) ([]model.Chat, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	c, err := apiCall(ctx, s, func(tok string) (conversations, error) {
		return s.srv.getConversations(ctx, tok, limit, 0)
	})
	if err != nil {
		return nil, err
	}
	s.learn(c.Profiles, c.Groups)
	out := make([]model.Chat, 0, len(c.Items))
	for _, it := range c.Items {
		chat := s.toChat(it.Conversation)
		if it.LastMessage != nil {
			m := s.toMessage(*it.LastMessage)
			chat.Last = &m
			chat.Updated = m.Time
		}
		out = append(out, chat)
	}
	return out, nil
}

func (s *session) History(ctx context.Context, chatID, beforeID string, limit int) ([]model.Message, bool, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return nil, false, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var start int64
	if beforeID != "" {
		// getHistory with start_message_id returns messages at or before it;
		// skip the first (the message itself) with offset 1.
		start, _ = strconv.ParseInt(beforeID, 10, 64)
	}
	h, err := apiCall(ctx, s, func(tok string) (history, error) {
		off := 0
		if start > 0 {
			off = 1
		}
		return s.srv.getHistory(ctx, tok, peer, start, limit, off)
	})
	if err != nil {
		return nil, false, err
	}
	s.learn(h.Profiles, h.Groups)
	// getHistory returns newest first already.
	out := make([]model.Message, 0, len(h.Items))
	for _, vm := range h.Items {
		out = append(out, s.toMessage(vm))
	}
	more := len(h.Items) == limit
	return out, more, nil
}

func (s *session) Send(ctx context.Context, chatID string, o model.Outgoing) (model.Message, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return model.Message{}, err
	}
	id, err := apiCall(ctx, s, func(tok string) (int64, error) {
		return s.srv.send(ctx, tok, peer, outText(o), sendOpts{replyTo: o.ReplyTo})
	})
	if err != nil {
		return model.Message{}, err
	}
	s.mu.Lock()
	self := s.self
	s.mu.Unlock()
	return model.Message{
		ChatID: chatID, ID: strconv.FormatInt(id, 10), Sender: self,
		Text: o.Text, Time: time.Now(), Outgoing: true, ReplyTo: o.ReplyTo,
	}, nil
}

func (s *session) MarkRead(ctx context.Context, chatID, msgID string) error {
	peer, err := parsePeer(chatID)
	if err != nil {
		return err
	}
	var start int64
	if msgID != "" {
		start, _ = strconv.ParseInt(msgID, 10, 64)
	}
	_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
		return struct{}{}, s.srv.markAsRead(ctx, tok, peer, start)
	})
	return err
}

// Media fetches a direct VK CDN URL (photos, docs, voice). Refs are the
// URLs themselves; they carry their own signature and need no credentials.
func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	if !strings.HasPrefix(ref, "http://") && !strings.HasPrefix(ref, "https://") {
		return nil, "", fmt.Errorf("vk: bad media ref")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.srv.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("vk: media: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, "", provider.ErrTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		return nil, "", provider.ErrTooLarge
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// --- Conversion ----------------------------------------------------------

func (s *session) learn(profiles []vkUser, groups []vkGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range profiles {
		s.people[u.ID] = toUser(u)
	}
	for _, g := range groups {
		s.people[-g.ID] = model.User{
			ID: strconv.FormatInt(-g.ID, 10), Name: g.Name,
			Username: g.ScreenName, Avatar: g.Photo,
		}
	}
}

func toUser(u vkUser) model.User {
	return model.User{
		ID: strconv.FormatInt(u.ID, 10), Name: u.name(),
		Username: u.ScreenneM, Avatar: u.Photo, Online: u.Online == 1,
	}
}

func (s *session) user(id int64) model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.people[id]; ok {
		return u
	}
	return model.User{ID: strconv.FormatInt(id, 10), Name: strconv.FormatInt(id, 10)}
}

func (s *session) toChat(c vkConversation) model.Chat {
	peer := c.Peer.ID
	chat := model.Chat{
		ID: strconv.FormatInt(peer, 10), Unread: c.UnreadCount,
	}
	if c.PushSettings != nil && (c.PushSettings.DisabledForever || c.PushSettings.DisabledUntil > time.Now().Unix()) {
		chat.Muted = true
	}
	if !c.CanWrite.Allowed {
		chat.Access = model.AccessRead
	} else {
		chat.Access = model.AccessWrite
	}
	switch c.Peer.Type {
	case "chat":
		chat.Kind = model.ChatGroup
		if c.ChatSettings != nil {
			chat.Title = c.ChatSettings.Title
			if c.ChatSettings.Photo != nil {
				chat.Avatar = c.ChatSettings.Photo.Photo100
			}
		}
	case "group":
		chat.Kind = model.ChatChannel
		u := s.user(peer)
		chat.Title, chat.Avatar = u.Name, u.Avatar
	default: // user
		chat.Kind = model.ChatDirect
		u := s.user(peer)
		chat.Title, chat.Avatar = u.Name, u.Avatar
	}
	return chat
}

func (s *session) toMessage(vm vkMessage) model.Message {
	peer := vm.PeerID
	text, spans := parseVKText(vm.Text)
	m := model.Message{
		ChatID: strconv.FormatInt(peer, 10),
		ID:     strconv.FormatInt(vm.ID, 10),
		Sender: s.user(vm.FromID),
		Text:   text,
		Format: spans,
		Time:   time.Unix(vm.Date, 0),
		Seq:    vm.ID,
	}
	if vm.ID == 0 && vm.ConvMsgID != 0 {
		m.ID = strconv.FormatInt(vm.ConvMsgID, 10)
	}
	if vm.Out == 1 {
		m.Outgoing = true
	}
	selfLink := "https://vk.com/id" + s.self.ID
	for _, sp := range spans {
		if sp.Kind == model.SpanLink && sp.URL == selfLink {
			m.Mentioned = true
			break
		}
	}
	if vm.UpdateTime > 0 {
		m.Edited = time.Unix(vm.UpdateTime, 0)
	}
	if vm.ReplyMsg != nil {
		m.ReplyTo = strconv.FormatInt(vm.ReplyMsg.ID, 10)
	}
	if vm.Action != nil {
		m.Service = true
		if m.Text == "" {
			m.Text = vm.Action.Text
		}
	}
	if len(vm.Fwd) > 0 {
		m.Forwarded = true
		if f := s.user(vm.Fwd[0].FromID); f.Name != "" {
			m.ForwardedFrom = f.Name
		}
	}
	for _, a := range vm.Attach {
		if att, ok := toAttachment(a); ok {
			m.Attachments = append(m.Attachments, att)
		}
	}
	return m
}

func toAttachment(a vkAttachment) (model.Attachment, bool) {
	switch a.Type {
	case "photo":
		if a.Photo == nil || len(a.Photo.Sizes) == 0 {
			return model.Attachment{}, false
		}
		big := a.Photo.Sizes[0]
		small := a.Photo.Sizes[0]
		for _, sz := range a.Photo.Sizes {
			if sz.Width > big.Width {
				big = sz
			}
			if sz.Width < small.Width {
				small = sz
			}
		}
		return model.Attachment{
			Kind: model.AttachImage, MIME: "image/jpeg",
			Ref: big.URL, Thumb: small.URL, Width: big.Width, Height: big.Height,
		}, true
	case "doc":
		if a.Doc == nil {
			return model.Attachment{}, false
		}
		name := a.Doc.Title
		if a.Doc.Ext != "" && !strings.HasSuffix(strings.ToLower(name), "."+strings.ToLower(a.Doc.Ext)) {
			name += "." + a.Doc.Ext
		}
		return model.Attachment{
			Kind: model.AttachFile, Name: name, Size: a.Doc.Size, Ref: a.Doc.URL,
		}, true
	case "audio_message":
		if a.AudioMsg == nil {
			return model.Attachment{}, false
		}
		return model.Attachment{
			Kind: model.AttachVoice, MIME: "audio/mpeg", Ref: a.AudioMsg.LinkMP3,
			Duration: time.Duration(a.AudioMsg.Duration) * time.Second,
		}, true
	case "sticker":
		if a.Sticker == nil || len(a.Sticker.Images) == 0 {
			return model.Attachment{}, false
		}
		img := a.Sticker.Images[len(a.Sticker.Images)-1]
		return model.Attachment{Kind: model.AttachSticker, Ref: img.URL, MIME: "image/png"}, true
	}
	return model.Attachment{}, false
}

// parsePeer turns a chat id string into a VK peer id.
func parsePeer(chatID string) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(chatID), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("vk: bad chat id %q", chatID)
	}
	return n, nil
}
