package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"

	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/richtext"
)

const keyUsers = "users"

// pollEvery is how often an OAuth account without Socket Mode looks for
// new messages in its most active chats.
var pollEvery = 20 * time.Second

type session struct {
	p    *Provider
	acct provider.Account
	env  provider.Env

	mu     sync.Mutex
	params map[string]string
	api    *api
	users  map[string]model.User
	dirty  bool
	chats  map[string]wChannel
	latest map[string]string // chat → newest ts seen (polling)
	self   model.User
	rtm    *websocket.Conn // while runRTM is connected: presence_sub goes here
	// presence: user → last known presence, for the "@" list (see fillPresence).
	presence map[string]seenPresence
	// foldersDue: a reload of the sidebar sections is on its way.
	foldersDue bool
}

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	params := map[string]string{}
	for k, v := range acct.Params {
		params[k] = v
	}
	s := &session{
		p: p, acct: acct, env: env, params: params, api: p.client(params, hc),
		users: map[string]model.User{}, chats: map[string]wChannel{}, latest: map[string]string{},
	}
	if b, err := env.State.Get(keyUsers); err == nil && b != nil {
		_ = json.Unmarshal(b, &s.users)
	}
	s.self = model.User{ID: params["user_id"], Name: s.users[params["user_id"]].Name}
	return s, nil
}

func (s *session) client() *api {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.api
}

// call is api.call that refreshes a rotating OAuth token once.
func (s *session) call(ctx context.Context, method string, params url.Values, out any) error {
	err := s.client().call(ctx, method, params, out)
	var ae *apiError
	if errors.As(err, &ae) && ae.Code == "token_expired" && s.params["refresh_token"] != "" {
		if rerr := s.refreshToken(ctx); rerr != nil {
			return wrapAuth(err)
		}
		err = s.client().call(ctx, method, params, out)
	}
	return wrapAuth(err)
}

func (s *session) refreshToken(ctx context.Context) error {
	s.mu.Lock()
	params := s.params
	s.mu.Unlock()
	var acc struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	a := &api{base: s.p.SlackURL + "api/", http: s.api.http}
	if err := a.call(ctx, "oauth.v2.access", url.Values{
		"client_id": {params["client_id"]}, "client_secret": {params["client_secret"]},
		"grant_type": {"refresh_token"}, "refresh_token": {params["refresh_token"]},
	}, &acc); err != nil {
		return err
	}
	next := map[string]string{}
	for k, v := range params {
		next[k] = v
	}
	next["token"] = acc.AccessToken
	setRefresh(next, acc.RefreshToken, acc.ExpiresIn)
	s.mu.Lock()
	s.params, s.api = next, s.p.client(next, s.api.http)
	s.mu.Unlock()
	return nil
}

// --- users ---------------------------------------------------------------

func (s *session) userName(id string) string { return s.userOf(context.Background(), id).Name }

// userOf returns a user, asking Slack (and remembering) when unknown.
func (s *session) userOf(ctx context.Context, id string) model.User {
	if id == "" {
		return model.User{}
	}
	s.mu.Lock()
	u, ok := s.users[id]
	s.mu.Unlock()
	if ok {
		return u
	}
	var r struct {
		User wUser `json:"user"`
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.call(ctx, "users.info", url.Values{"user": {id}}, &r); err != nil {
		return model.User{ID: id, Name: id}
	}
	u = model.User{ID: id, Name: r.User.display(), Username: r.User.Name, Avatar: r.User.Profile.Image72}
	s.mu.Lock()
	s.users[id], s.dirty = u, true
	s.mu.Unlock()
	return u
}

func (s *session) saveUsers() {
	s.mu.Lock()
	if !s.dirty {
		s.mu.Unlock()
		return
	}
	b, err := json.Marshal(s.users)
	s.dirty = false
	s.mu.Unlock()
	if err == nil {
		err = s.env.State.Put(keyUsers, b)
	}
	if err != nil {
		s.env.Logf("error", "save users: %v", err)
	}
}

// --- conversion ----------------------------------------------------------

func fileKind(f wFile) model.AttachmentKind {
	switch {
	case f.Subtype == "slack_audio" || strings.HasPrefix(f.Mimetype, "audio/"):
		return model.AttachAudio
	case strings.HasPrefix(f.Mimetype, "image/"):
		return model.AttachImage
	case strings.HasPrefix(f.Mimetype, "video/"):
		return model.AttachVideo
	}
	return model.AttachFile
}

// skipped reports messages that are not shown in the chat itself.
func skipped(m wMessage) bool {
	if m.Hidden {
		return true
	}
	// Thread replies live in their thread, except those also sent to the channel.
	if m.ThreadTS != "" && m.ThreadTS != m.TS && m.Subtype != "thread_broadcast" {
		return true
	}
	switch m.Subtype {
	case "message_changed", "message_deleted", "message_replied":
		return true
	}
	return false
}

var serviceSubtypes = map[string]bool{
	"channel_join": true, "channel_leave": true, "channel_topic": true, "channel_purpose": true,
	"channel_name": true, "channel_archive": true, "channel_unarchive": true, "group_join": true,
	"group_leave": true, "group_topic": true, "group_purpose": true, "group_name": true, "pinned_item": true,
}

func (s *session) convert(ctx context.Context, chat string, m wMessage) model.Message {
	out := model.Message{
		ChatID: chat, ID: m.TS, Time: tsTime(m.TS),
		Text:    plainText(m.Text, s.userName),
		Service: serviceSubtypes[m.Subtype],
	}
	switch {
	case m.User != "":
		out.Sender = s.userOf(ctx, m.User)
	case m.Username != "":
		out.Sender = model.User{ID: m.BotID, Name: m.Username}
	default:
		out.Sender = model.User{ID: m.BotID, Name: "bot"}
	}
	out.Outgoing = m.User != "" && m.User == s.params["user_id"]
	out.Mentioned = !out.Outgoing && mentions(m.Text, s.params["user_id"])
	if m.Edited != nil {
		out.Edited = tsTime(m.Edited.TS)
	}
	if m.Subtype == "thread_broadcast" {
		out.ReplyTo = m.ThreadTS
	}
	if m.ThreadTS != "" && m.ThreadTS != m.TS && m.Subtype != "thread_broadcast" {
		out.InThread = m.ThreadTS
	}
	// An app's blocks carry the body; its text is only the fallback. A
	// user's rich text says what the text does.
	if bs := parseBlocks(m.Blocks); blocksCarryBody(bs) || (strings.TrimSpace(m.Text) == "" && len(bs) > 0) {
		if t := blocksText(bs, s.userName); t != "" {
			out.Text = t
		}
	}
	for _, f := range m.Files {
		out.Attachments = append(out.Attachments, attachmentOf(f))
	}
	// An app's legacy attachments (not link unfurls) add to the body.
	if m.BotID != "" || m.Subtype == "bot_message" {
		for _, a := range m.Attachments {
			if a.IsShare || a.IsMsgUnfurl || a.FromURL != "" || a.OriginalURL != "" {
				continue
			}
			if t := attachmentText(a, s.userName); t != "" && t != out.Text {
				out.Text = strings.TrimSpace(out.Text + "\n" + t)
			}
		}
	}
	// A shared message (Share message, a pasted permalink): its author,
	// text and files instead of the bare link.
	for _, a := range m.Attachments {
		if !a.IsShare && !a.IsMsgUnfurl {
			continue
		}
		if !out.Forwarded {
			out.Forwarded = true
			out.ForwardedFrom = a.AuthorSubname
			if out.ForwardedFrom == "" {
				out.ForwardedFrom = a.AuthorName
			}
			if strings.TrimSpace(strings.Trim(m.Text, "<>")) == a.FromURL || strings.Trim(m.Text, "<>") == a.OriginalURL {
				out.Text = ""
			}
		}
		if t := plainText(a.Text, s.userName); t != "" {
			out.Text = strings.TrimSpace(out.Text + "\n" + t)
		}
		for _, f := range a.Files {
			out.Attachments = append(out.Attachments, attachmentOf(f))
		}
	}
	// Everything above is mrkdwn: its marks become styles.
	if !out.Service {
		// A reply shown in the chat: its leading "> " lines are its quote.
		if out.ReplyTo != "" {
			if lines, rest := splitQuote(out.Text); len(lines) > 0 {
				out.Quote, out.Text = richtext.Parse(strings.Join(lines, "\n"), richtext.Slack).Text, rest
			}
		}
		d := richtext.Parse(out.Text, richtext.Slack)
		out.Text, out.Format = d.Text, d.Spans
	}
	out.Reactions = s.reactionsOf(m.Reactions)
	// Any message of the chat can start a thread: Thread with no replies
	// says so.
	if !out.Service && m.TS != "" && (m.ThreadTS == "" || m.ThreadTS == m.TS) {
		t := &model.Thread{ID: m.TS, Replies: m.ReplyCount}
		if m.LatestRepl != "" {
			t.Last = tsTime(m.LatestRepl)
		}
		for _, u := range m.ReplyUsers[:min(len(m.ReplyUsers), 3)] {
			t.Users = append(t.Users, s.userOf(ctx, u))
		}
		out.Thread = t
	}
	return out
}

func attachmentOf(f wFile) model.Attachment {
	name := f.Name
	if name == "" {
		name = f.Title
	}
	a := model.Attachment{
		Kind: fileKind(f), Name: name, MIME: f.Mimetype, Size: f.Size,
		Ref: f.URLPrivateDownload, Width: f.OriginalW, Height: f.OriginalH,
		Duration: time.Duration(f.DurationMS) * time.Millisecond,
	}
	if a.Ref == "" {
		a.Ref = f.URLPrivate
	}
	switch {
	case f.Thumb480 != "":
		a.Thumb = f.Thumb480
	case f.Thumb360 != "":
		a.Thumb = f.Thumb360
	case f.ThumbVideo != "":
		a.Thumb = f.ThumbVideo
	}
	if a.Width == 0 && f.Thumb360W > 0 {
		a.Width, a.Height = f.Thumb360W, f.Thumb360H
	}
	return a
}

func (s *session) reactionsOf(rs []wReaction) []model.Reaction {
	self := s.params["user_id"]
	var out []model.Reaction
	for _, r := range rs {
		mine := false
		for _, u := range r.Users {
			if u == self {
				mine = true
			}
		}
		out = append(out, model.Reaction{Emoji: emojiOf(r.Name), Count: r.Count, Mine: mine})
	}
	return out
}

func (s *session) chatOf(ctx context.Context, c wChannel) model.Chat {
	out := model.Chat{ID: c.ID, Unread: -1}
	switch {
	case c.IsIM:
		out.Kind = model.ChatDirect
		u := s.userOf(ctx, c.User)
		out.Title, out.Avatar = u.Name, u.Avatar
		if c.User == s.params["user_id"] {
			out.Title += " (you)"
			out.Saved = true // the notes to oneself
		}
	case c.IsMPIM:
		out.Kind, out.Title = model.ChatGroup, mpimTitle(c.Name)
	default:
		out.Kind, out.Title = model.ChatGroup, "#"+c.Name
	}
	out.Access = model.AccessWrite
	if c.IsArchived || c.IsReadOnly {
		out.Access = model.AccessRead
	}
	return out
}

// --- Messenger -----------------------------------------------------------

func (s *session) conversations(ctx context.Context) ([]wChannel, error) {
	var all []wChannel
	cursor := ""
	for {
		var r struct {
			Channels []wChannel `json:"channels"`
			Meta     struct {
				Next string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		q := url.Values{"types": {"public_channel,private_channel,mpim,im"}, "exclude_archived": {"true"}, "limit": {"200"}}
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		if err := s.call(ctx, "users.conversations", q, &r); err != nil {
			return all, err
		}
		all = append(all, r.Channels...)
		if r.Meta.Next == "" || len(all) >= 2000 {
			break
		}
		cursor = r.Meta.Next
	}
	s.mu.Lock()
	for _, c := range all {
		s.chats[c.ID] = c
	}
	s.mu.Unlock()
	s.subscribePresence(ctx)
	return all, nil
}

// subscribePresence asks the RTM connection for the presence of the
// people of the direct chats; Slack answers with their presence_change
// now and on every change. Only browser sessions (RTM) have presence:
// the Events API of Socket Mode does not carry it.
func (s *session) subscribePresence(ctx context.Context) {
	s.mu.Lock()
	c := s.rtm
	var ids []string
	for _, ch := range s.chats {
		if ch.IsIM && ch.User != "" && ch.User != s.self.ID && !slices.Contains(ids, ch.User) {
			ids = append(ids, ch.User)
		}
	}
	s.mu.Unlock()
	if c == nil || len(ids) == 0 {
		return
	}
	sort.Strings(ids)
	msg, _ := json.Marshal(map[string]any{"type": "presence_sub", "ids": ids})
	wctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := c.Write(wctx, websocket.MessageText, msg); err != nil {
		s.env.Logf("warn", "presence_sub: %v", err)
	}
}

// emitPresence reports the presence of users ("active" or "away") as the
// presence of their direct chats.
func (s *session) emitPresence(sink provider.Sink, users []string, presence string) {
	s.setPresence(users, presence == "active")
	s.mu.Lock()
	var chats []string
	for _, ch := range s.chats {
		if ch.IsIM && slices.Contains(users, ch.User) {
			chats = append(chats, ch.ID)
		}
	}
	s.mu.Unlock()
	for _, id := range chats {
		sink.Emit(model.Event{Kind: model.EvPresence, Chat: &model.Chat{ID: id}, Online: presence == "active"})
	}
}

// counts is the web client's unread summary; only browser sessions have it.
type counts struct {
	Channels []countEntry `json:"channels"`
	MPIMs    []countEntry `json:"mpims"`
	IMs      []countEntry `json:"ims"`
}

type countEntry struct {
	ID           string `json:"id"`
	Latest       string `json:"latest"`
	HasUnreads   bool   `json:"has_unreads"`
	MentionCount int    `json:"mention_count"`
}

func (s *session) Chats(ctx context.Context, limit int) ([]model.Chat, error) {
	convs, err := s.conversations(ctx)
	if err != nil {
		return nil, err
	}
	if c, ok := s.selfDM(ctx, convs); ok {
		convs = append(convs, c)
	}
	defer s.saveUsers()
	latest := map[string]countEntry{}
	if s.params["method"] != methodOAuth {
		var c counts
		if err := s.call(ctx, "client.counts", nil, &c); err == nil {
			for _, list := range [][]countEntry{c.Channels, c.MPIMs, c.IMs} {
				for _, e := range list {
					latest[e.ID] = e
				}
			}
		}
	}
	out := make([]model.Chat, 0, len(convs))
	for _, c := range convs {
		ch := s.chatOf(ctx, c)
		if e, ok := latest[c.ID]; ok {
			ch.Updated = tsTime(e.Latest)
			ch.Unread = 0
			if e.HasUnreads {
				ch.Unread = max(e.MentionCount, 1)
			}
		}
		out = append(out, ch)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	if len(out) > limit {
		out = out[:limit]
	}
	// Last messages for the top of the list (Slack has no batch call).
	for i := range out[:min(len(out), 40)] {
		msgs, _, err := s.History(ctx, out[i].ID, "", 1)
		if err != nil {
			break // rate limited or offline: the rest comes with events
		}
		if len(msgs) > 0 {
			out[i].Last = &msgs[0]
			if msgs[0].Time.After(out[i].Updated) {
				out[i].Updated = msgs[0].Time
			}
			s.seen(out[i].ID, msgs[0].ID)
		}
	}
	return out, nil
}

// seen remembers the newest ts of a chat (polling starts there).
func (s *session) seen(chat, ts string) {
	s.mu.Lock()
	if tsTime(ts).After(tsTime(s.latest[chat])) {
		s.latest[chat] = ts
	}
	s.mu.Unlock()
}

func (s *session) History(ctx context.Context, chatID, beforeID string, limit int) ([]model.Message, bool, error) {
	q := url.Values{"channel": {chatID}, "limit": {strconv.Itoa(limit)}}
	if beforeID != "" {
		q.Set("latest", beforeID)
		q.Set("inclusive", "false")
	}
	var r struct {
		Messages []wMessage `json:"messages"`
		HasMore  bool       `json:"has_more"`
	}
	if err := s.call(ctx, "conversations.history", q, &r); err != nil {
		return nil, false, err
	}
	defer s.saveUsers()
	out := make([]model.Message, 0, len(r.Messages))
	for _, m := range r.Messages {
		if skipped(m) {
			continue
		}
		out = append(out, s.convert(ctx, chatID, m))
	}
	return out, r.HasMore, nil
}

func (s *session) Send(ctx context.Context, chatID string, o model.Outgoing) (model.Message, error) {
	// A quote is "> " lines (escaped, as Slack itself writes them).
	q := url.Values{"channel": {chatID}, "text": {outText(o)}}
	switch {
	case o.Thread != "":
		q.Set("thread_ts", o.Thread)
	case o.ReplyTo != "":
		q.Set("thread_ts", o.ReplyTo)
		q.Set("reply_broadcast", "true") // show it in the chat too
	}
	var r struct {
		TS      string   `json:"ts"`
		Message wMessage `json:"message"`
	}
	if err := s.call(ctx, "chat.postMessage", q, &r); err != nil {
		return model.Message{}, err
	}
	if r.Message.TS == "" {
		r.Message.TS = r.TS
	}
	if r.Message.User == "" {
		r.Message.User = s.params["user_id"]
	}
	if o.Thread != "" && r.Message.ThreadTS == "" {
		r.Message.ThreadTS = o.Thread
	}
	if o.Thread == "" && o.ReplyTo != "" && r.Message.Subtype == "" {
		// The answer may not say what reply_broadcast made of it.
		r.Message.Subtype, r.Message.ThreadTS = "thread_broadcast", o.ReplyTo
	}
	m := s.convert(ctx, chatID, r.Message)
	m.Outgoing = true
	if o.Thread == "" {
		m.ReplyTo = o.ReplyTo
		s.seen(chatID, m.ID)
	}
	return m, nil
}

// Forward shares messages the way Slack's own "Share message" looks: a
// post with the message's permalink, which Slack unfurls into the original
// (author, text, files) and marks as shared. One post per message.
func (s *session) Forward(ctx context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error) {
	var out []model.Message
	for _, id := range msgIDs {
		var p struct {
			Permalink string `json:"permalink"`
		}
		if err := s.call(ctx, "chat.getPermalink", url.Values{"channel": {fromChat}, "message_ts": {id}}, &p); err != nil {
			return out, err
		}
		if p.Permalink == "" {
			return out, errors.New("slack: no permalink for the message")
		}
		var r struct {
			TS      string   `json:"ts"`
			Message wMessage `json:"message"`
		}
		q := url.Values{"channel": {toChat}, "text": {p.Permalink}, "unfurl_links": {"true"}}
		if err := s.call(ctx, "chat.postMessage", q, &r); err != nil {
			return out, err
		}
		if r.Message.TS == "" {
			r.Message.TS = r.TS
		}
		if r.Message.User == "" {
			r.Message.User = s.params["user_id"]
		}
		m := s.convert(ctx, toChat, r.Message)
		// The unfurl comes later as an edit; until then it is a shared link.
		m.Outgoing, m.Forwarded = true, true
		s.seen(toChat, m.ID)
		out = append(out, m)
	}
	return out, nil
}

func (s *session) MarkRead(ctx context.Context, chatID, msgID string) error {
	if msgID == "" {
		msgs, _, err := s.History(ctx, chatID, "", 1)
		if err != nil || len(msgs) == 0 {
			return err
		}
		msgID = msgs[0].ID
	}
	return s.call(ctx, "conversations.mark", url.Values{"channel": {chatID}, "ts": {msgID}}, nil)
}

// --- live events ---------------------------------------------------------

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	var at authTest
	if err := s.call(ctx, "auth.test", nil, &at); err != nil {
		return err
	}
	s.mu.Lock()
	s.self = model.User{ID: at.UserID, Name: at.User}
	s.mu.Unlock()
	if u := s.userOf(ctx, at.UserID); u.Name != at.UserID {
		s.mu.Lock()
		s.self = u
		s.mu.Unlock()
	}
	defer s.saveParams(sink)
	switch {
	case s.params["method"] != methodOAuth:
		return s.runRTM(ctx, sink)
	case s.params["app_token"] != "":
		return s.runSocket(ctx, sink)
	default:
		return s.runPoll(ctx, sink)
	}
}

// saveParams keeps a refreshed token.
func (s *session) saveParams(sink provider.Sink) {
	s.mu.Lock()
	changed := s.params["token"] != s.acct.Params["token"]
	params := s.params
	s.mu.Unlock()
	if changed {
		if err := sink.SaveParams(params); err != nil {
			s.env.Logf("error", "save refreshed token: %v", err)
		}
	}
}

func (s *session) dial(ctx context.Context, wsURL string, cookie bool) (*websocket.Conn, error) {
	h := http.Header{}
	if cookie && s.params["cookie"] != "" {
		h.Set("Cookie", "d="+url.QueryEscape(s.params["cookie"]))
	}
	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: h, HTTPClient: s.client().http})
	if err != nil {
		return nil, err
	}
	c.SetReadLimit(16 << 20)
	return c, nil
}

// event is an RTM event or the "event" of a Socket Mode envelope. A
// "message" event is itself the message, so it is decoded a second time
// into wMessage (embedding wMessage here would not work: fields of the
// same name in event shadow its ts, user and subtype).
type event struct {
	Type    string   `json:"type"`
	Subtype string   `json:"subtype"`
	Channel string   `json:"channel"`
	User    string   `json:"user"`
	Message wMessage `json:"message"`
	Deleted string   `json:"deleted_ts"`
	Unread  *int     `json:"unread_count_display"`
	// presence_change: one user, or several with batch_presence_aware.
	Users    []string `json:"users"`
	Presence string   `json:"presence"`
	Item     struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	} `json:"item"` // reaction_added, reaction_removed
}

// isReply tells a message inside a thread (not also sent to the chat).
func isReply(m wMessage) bool {
	return m.ThreadTS != "" && m.ThreadTS != m.TS && m.Subtype != "thread_broadcast"
}

func (s *session) handle(ctx context.Context, sink provider.Sink, raw []byte) {
	var ev event
	if json.Unmarshal(raw, &ev) != nil {
		return
	}
	var msg wMessage
	if ev.Type == "message" && json.Unmarshal(raw, &msg) != nil {
		return
	}
	if strings.HasPrefix(ev.Type, "channel_section") { // the sidebar's sections
		s.sectionsChanged(ctx, sink)
		return
	}
	switch ev.Type {
	case "message":
		switch ev.Subtype {
		case "message_changed":
			switch {
			case ev.Message.TS == "" || ev.Message.Hidden:
			case isReply(ev.Message):
				m := s.convert(ctx, ev.Channel, ev.Message)
				sink.Emit(model.Event{Kind: model.EvThreadMessage, Message: &m})
			case !skipped(ev.Message):
				m := s.convert(ctx, ev.Channel, ev.Message)
				sink.Emit(model.Event{Kind: model.EvMessageEdit, Message: &m})
			}
		case "message_replied":
			// The thread's root with its new reply count.
			if ev.Message.TS != "" && !skipped(ev.Message) {
				m := s.convert(ctx, ev.Channel, ev.Message)
				sink.Emit(model.Event{Kind: model.EvMessageEdit, Message: &m})
			}
		case "message_deleted":
			sink.Emit(model.Event{Kind: model.EvMessageDelete, Message: &model.Message{ChatID: ev.Channel, ID: ev.Deleted}})
		default:
			if msg.TS != "" && !msg.Hidden && isReply(msg) {
				m := s.convert(ctx, ev.Channel, msg)
				sink.Emit(model.Event{Kind: model.EvThreadMessage, Message: &m})
				s.emitRoot(ctx, sink, ev.Channel, msg.ThreadTS)
				return
			}
			if skipped(msg) || msg.TS == "" {
				return
			}
			m := s.convert(ctx, ev.Channel, msg)
			s.seen(ev.Channel, m.ID)
			var chat *model.Chat
			s.mu.Lock()
			c, known := s.chats[ev.Channel]
			s.mu.Unlock()
			if known {
				x := s.chatOf(ctx, c)
				chat = &x
			}
			sink.Emit(model.Event{Kind: model.EvMessageNew, Message: &m, Chat: chat})
		}
		s.saveUsers()
	case "channel_marked", "group_marked", "im_marked", "mpim_marked":
		unread := 0
		if ev.Unread != nil {
			unread = *ev.Unread
		}
		sink.Emit(model.Event{Kind: model.EvRead, Chat: &model.Chat{ID: ev.Channel, Unread: unread}})
	case "reaction_added", "reaction_removed":
		if ev.Item.Type == "message" && ev.Item.Channel != "" {
			s.emitReactions(ctx, sink, ev.Item.Channel, ev.Item.TS)
		}
	case "user_typing":
		sink.Emit(model.Event{Kind: model.EvTyping, Message: &model.Message{ChatID: ev.Channel, Sender: s.userOf(ctx, ev.User)}})
	case "channel_left", "group_left", "channel_deleted", "group_archive", "channel_archive":
		sink.Emit(model.Event{Kind: model.EvChatRemoved, Chat: &model.Chat{ID: ev.Channel}})
	case "presence_change":
		users := ev.Users
		if ev.User != "" {
			users = append(users, ev.User)
		}
		s.emitPresence(sink, users, ev.Presence)
	}
}

// runRTM is the browser session's real-time connection.
func (s *session) runRTM(ctx context.Context, sink provider.Sink) error {
	var r struct {
		URL string `json:"url"`
	}
	if err := s.call(ctx, "rtm.connect", nil, &r); err != nil {
		return err
	}
	c, err := s.dial(ctx, r.URL, true)
	if err != nil {
		return fmt.Errorf("slack: rtm: %w", err)
	}
	defer c.CloseNow()
	s.mu.Lock()
	s.rtm = c
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.rtm = nil
		s.mu.Unlock()
	}()
	pingCtx, stopPing := context.WithCancel(ctx)
	defer stopPing()
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for id := 1; ; id++ {
			select {
			case <-pingCtx.Done():
				return
			case <-t.C:
				msg := fmt.Sprintf(`{"id":%d,"type":"ping"}`, id)
				if c.Write(pingCtx, websocket.MessageText, []byte(msg)) != nil {
					c.CloseNow()
					return
				}
			}
		}
	}()
	for {
		rctx, cancel := context.WithTimeout(ctx, 90*time.Second) // pongs come every 30 s
		_, data, err := c.Read(rctx)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("slack: rtm: %w", err)
		}
		var head struct {
			Type string `json:"type"`
		}
		json.Unmarshal(data, &head)
		switch head.Type {
		case "hello":
			sink.Connected(s.selfUser())
			go s.subscribePresence(ctx) // after a reconnect: the chats are known
		case "goodbye":
			return errors.New("slack: server asked to reconnect")
		default:
			s.handle(ctx, sink, data)
		}
	}
}

// runSocket is Socket Mode with the app-level token (OAuth accounts).
func (s *session) runSocket(ctx context.Context, sink provider.Sink) error {
	app := &api{base: s.p.SlackURL + "api/", token: s.params["app_token"], http: s.client().http}
	var r struct {
		URL string `json:"url"`
	}
	if err := app.call(ctx, "apps.connections.open", nil, &r); err != nil {
		if isAuth(err) {
			return fmt.Errorf("%w (app-level token: %v)", provider.ErrAuth, err)
		}
		return err
	}
	c, err := s.dial(ctx, r.URL, false)
	if err != nil {
		return fmt.Errorf("slack: socket mode: %w", err)
	}
	defer c.CloseNow()
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("slack: socket mode: %w", err)
		}
		var env struct {
			Type       string `json:"type"`
			EnvelopeID string `json:"envelope_id"`
			Payload    struct {
				Event json.RawMessage `json:"event"`
			} `json:"payload"`
		}
		if json.Unmarshal(data, &env) != nil {
			continue
		}
		if env.EnvelopeID != "" {
			ack, _ := json.Marshal(map[string]string{"envelope_id": env.EnvelopeID})
			if err := c.Write(ctx, websocket.MessageText, ack); err != nil {
				return err
			}
		}
		switch env.Type {
		case "hello":
			sink.Connected(s.selfUser())
		case "disconnect":
			return errors.New("slack: socket mode asked to reconnect")
		case "events_api":
			s.handle(ctx, sink, env.Payload.Event)
		}
	}
}

// runPoll looks for new messages in the chats with recent activity.
func (s *session) runPoll(ctx context.Context, sink provider.Sink) error {
	sink.Connected(s.selfUser())
	t := time.NewTicker(pollEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
		if err := s.pollOnce(ctx, sink); err != nil {
			if errors.Is(err, provider.ErrAuth) || ctx.Err() != nil {
				return err
			}
			s.env.Logf("warn", "poll: %v", err)
		}
	}
}

func (s *session) pollOnce(ctx context.Context, sink provider.Sink) error {
	s.mu.Lock()
	type chatTS struct{ id, ts string }
	var list []chatTS
	for id, ts := range s.latest {
		list = append(list, chatTS{id, ts})
	}
	s.mu.Unlock()
	sort.Slice(list, func(i, j int) bool { return tsTime(list[i].ts).After(tsTime(list[j].ts)) })
	for _, c := range list[:min(len(list), 20)] {
		var r struct {
			Messages []wMessage `json:"messages"`
		}
		q := url.Values{"channel": {c.id}, "oldest": {c.ts}, "inclusive": {"false"}, "limit": {"50"}}
		if err := s.call(ctx, "conversations.history", q, &r); err != nil {
			return err
		}
		for i := len(r.Messages) - 1; i >= 0; i-- { // oldest first
			m := r.Messages[i]
			if skipped(m) {
				continue
			}
			msg := s.convert(ctx, c.id, m)
			s.seen(c.id, msg.ID)
			sink.Emit(model.Event{Kind: model.EvMessageNew, Message: &msg})
		}
	}
	s.saveUsers()
	return nil
}

func (s *session) selfUser() model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.self
}
