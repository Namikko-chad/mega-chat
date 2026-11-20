package vkteams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

type session struct {
	p    *Provider
	acct provider.Account
	env  provider.Env
	srv  *server

	mu       sync.Mutex
	aimsid   string
	fetchURL string
	self     model.User
	people   map[string]string // sn → name
	dialogs  map[string]*dialog
	online   bool
	listed   bool            // contacts/get was tried in this run
	unnamed  map[string]bool // groups getChatInfo gave no name in this run
	naming   bool            // nameGroups is running
	sink     provider.Sink   // of the current Run
	runCtx   context.Context // of the current Run
}

// dialog is what the session knows of a chat (saved, so the chat list is
// there before the first events).
type dialog struct {
	SN      string         `json:"sn"`
	Title   string         `json:"title"`
	Unread  int            `json:"unread"`
	LastID  string         `json:"last_id"`
	Last    *model.Message `json:"last,omitempty"`
	Updated time.Time      `json:"updated"`
}

func (p *Provider) Open(acct provider.Account, env provider.Env) (provider.Session, error) {
	hc := env.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	s := &session{
		p: p, acct: acct, env: env, srv: p.server(acct.Params, hc),
		people: map[string]string{}, dialogs: map[string]*dialog{}, unnamed: map[string]bool{},
	}
	if b, _ := env.State.Get(keyAimsid); b != nil {
		s.aimsid = string(b)
	} else {
		s.aimsid = acct.Params["aimsid"]
	}
	if b, _ := env.State.Get(keyFetch); b != nil {
		s.fetchURL = string(b)
	}
	if b, _ := env.State.Get(keyDialogs); b != nil {
		_ = json.Unmarshal(b, &s.dialogs)
	}
	for _, d := range s.dialogs {
		if d.Title != "" && !isGroup(d.SN) {
			s.people[d.SN] = d.Title
		}
	}
	_, id, _ := strings.Cut(acct.Login, "/")
	s.self = model.User{ID: id, Name: id}
	return s, nil
}

func isGroup(sn string) bool { return strings.HasSuffix(sn, "@chat.agent") }

func (s *session) current() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.aimsid, s.online
}

// renew starts a new session with the stored token (password accounts).
func (s *session) renew(ctx context.Context) error {
	pr := s.acct.Params
	if pr["method"] != methodPassword || pr["token"] == "" {
		return fmt.Errorf("vk teams: %w (the web session ended)", provider.ErrAuth)
	}
	offset, _ := strconv.ParseInt(pr["time_offset"], 10, 64)
	si, err := s.srv.startSession(ctx, pr["dev_id"], pr["token"], pr["session_key"], offset)
	var le *errLogin
	if errors.As(err, &le) {
		return fmt.Errorf("vk teams: %w (%v)", provider.ErrAuth, err)
	}
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.aimsid, s.fetchURL = si.AimSID, si.FetchBaseURL
	if n := si.MyInfo.name(); n != "" {
		s.self = model.User{ID: si.MyInfo.id(), Name: n}
	}
	s.mu.Unlock()
	s.env.State.Put(keyAimsid, []byte(si.AimSID))
	return s.env.State.Put(keyFetch, []byte(si.FetchBaseURL))
}

func (s *session) Run(ctx context.Context, sink provider.Sink) error {
	s.mu.Lock()
	s.sink, s.runCtx = sink, ctx
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.online = false
		s.mu.Unlock()
	}()
	if s.aimsid == "" {
		if err := s.renew(ctx); err != nil {
			return err
		}
	}
	connected := false
	for {
		s.mu.Lock()
		u := s.fetchURL
		if u == "" {
			u = s.srv.firstFetchURL(s.aimsid)
		}
		s.mu.Unlock()
		r, err := s.srv.fetchEvents(ctx, u, fetchTimeout)
		var se *errSession
		switch {
		case ctx.Err() != nil:
			return ctx.Err()
		case errors.As(err, &se):
			if err := s.renew(ctx); err != nil {
				return err
			}
			continue
		case err != nil:
			return err
		}
		if !connected {
			connected = true
			s.mu.Lock()
			s.online = true
			s.mu.Unlock()
			for _, ev := range r.Events { // learn who we are before telling
				if ev.Type == "myInfo" {
					s.handle(sink, ev)
				}
			}
			s.learnSelf(ctx)
			sink.Connected(s.selfUser())
		}
		for _, ev := range r.Events {
			s.handle(sink, ev)
		}
		if r.FetchBaseURL != "" {
			s.mu.Lock()
			s.fetchURL = r.FetchBaseURL
			s.mu.Unlock()
			// Resume where we stopped after a restart.
			s.env.State.Put(keyFetch, []byte(r.FetchBaseURL))
		}
		if len(r.Events) > 0 {
			s.saveDialogs()
		}
		if r.TimeToNextFetch > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(min(r.TimeToNextFetch, 5000)) * time.Millisecond):
			}
		}
	}
}

func (s *session) selfUser() model.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.self
	u.Avatar = avatarRef(u.ID)
	return u
}

// avatarRef is the media ref of sn's photo.
func avatarRef(sn string) string {
	if sn == "" {
		return ""
	}
	return refAvatar + sn
}

const refAvatar = "avatar:"

// errTooLarge is what Media returns for a file over the limit.
var errTooLarge = provider.ErrTooLarge

// Media fetches what a ref names; only photos ("avatar:<sn>") for now.
func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	sn, ok := strings.CutPrefix(ref, refAvatar)
	if !ok || sn == "" {
		return nil, "", fmt.Errorf("vk teams: files are not supported yet")
	}
	aimsid, err := s.live()
	if err != nil {
		return nil, "", err
	}
	return s.srv.avatar(ctx, aimsid, sn, limit)
}

// learnSelf asks the server who this session is when nothing said it
// (a resumed web session gets no myInfo).
func (s *session) learnSelf(ctx context.Context) {
	s.mu.Lock()
	me, aimsid := s.self, s.aimsid
	s.mu.Unlock()
	if me.ID == "" || (me.Name != "" && me.Name != me.ID) {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	p, err := s.srv.getUserInfo(cctx, aimsid, me.ID)
	if err != nil {
		s.env.Logf("warn", "vk teams: own profile: %v", err)
		return
	}
	if n := p.name(); n != "" && n != me.ID {
		s.mu.Lock()
		s.self.Name = n
		s.people[me.ID] = n
		s.mu.Unlock()
	}
}

func (s *session) saveDialogs() {
	s.mu.Lock()
	b, err := json.Marshal(s.dialogs)
	s.mu.Unlock()
	if err == nil {
		err = s.env.State.Put(keyDialogs, b)
	}
	if err != nil {
		s.env.Logf("error", "save dialogs: %v", err)
	}
}

func (s *session) learn(ps []person) {
	s.mu.Lock()
	for _, p := range ps {
		if id, n := p.id(), p.name(); id != "" && n != "" && n != id {
			s.people[id] = n
		}
	}
	s.mu.Unlock()
}

func (s *session) personName(sn string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := s.people[sn]; n != "" {
		return n
	}
	return sn
}

func (s *session) convert(sn string, m wMessage) model.Message {
	out := model.Message{
		ChatID: sn, ID: m.MsgID.String(), Text: m.Text, Time: time.Unix(m.Time, 0),
		Outgoing: m.Outgoing, Service: m.Class == "event",
	}
	out.Seq, _ = m.MsgID.Int64()
	out.Mentioned = !m.Outgoing && mentions(m.Text, s.selfUser().ID)
	switch {
	case m.Outgoing:
		out.Sender = s.selfUser()
	case m.Chat.Sender != "":
		out.Sender = model.User{ID: m.Chat.Sender, Name: s.personName(m.Chat.Sender), Avatar: avatarRef(m.Chat.Sender)}
	default:
		out.Sender = model.User{ID: sn, Name: s.personName(sn), Avatar: avatarRef(sn)}
	}
	for _, f := range m.Filesharing {
		a := model.Attachment{Kind: model.AttachFile, Name: f.Name, MIME: f.Mime, Size: int64(f.Size), Ref: f.ID}
		switch {
		case strings.HasPrefix(f.Mime, "image/"):
			a.Kind = model.AttachImage
		case strings.HasPrefix(f.Mime, "video/"):
			a.Kind = model.AttachVideo
		case strings.HasPrefix(f.Mime, "audio/") && f.Duration > 0:
			a.Kind = model.AttachVoice
		}
		out.Attachments = append(out.Attachments, a)
	}
	return out
}

// chatOf is d as a chat; the caller holds s.mu.
func (s *session) chatOf(d *dialog) model.Chat {
	c := model.Chat{ID: d.SN, Title: d.Title, Unread: d.Unread, Last: d.Last, Updated: d.Updated, Kind: model.ChatDirect, Avatar: avatarRef(d.SN)}
	if isGroup(d.SN) {
		c.Kind = model.ChatGroup
	}
	if c.Title == "" {
		c.Title = s.people[d.SN] // s.mu is held: not personName
	}
	if c.Title == "" {
		c.Title = d.SN
	}
	return c
}

func (s *session) handle(sink provider.Sink, ev rawEvent) {
	switch ev.Type {
	case "myInfo":
		var me person
		if json.Unmarshal(ev.EventData, &me) == nil && me.id() != "" {
			s.mu.Lock()
			s.self = model.User{ID: me.id(), Name: me.name()}
			s.mu.Unlock()
		}
	case "histDlgState":
		var st dlgState
		if json.Unmarshal(ev.EventData, &st) != nil || st.SN == "" {
			return
		}
		s.learn(st.Persons)
		s.mu.Lock()
		d := s.dialogs[st.SN]
		if d == nil {
			d = &dialog{SN: st.SN}
			s.dialogs[st.SN] = d
		}
		d.Unread = st.UnreadCnt
		if st.LastMsgID != "" {
			d.LastID = st.LastMsgID.String()
		}
		if !isGroup(st.SN) {
			d.Title = s.people[st.SN]
		}
		s.mu.Unlock()
		for _, m := range st.Messages {
			msg := s.convert(st.SN, m)
			kind := model.EvMessageNew
			for _, p := range m.UpdatePatch {
				switch p.Type {
				case "delete":
					kind = model.EvMessageDelete
				case "edit":
					kind = model.EvMessageEdit
				}
			}
			if kind == model.EvMessageNew {
				s.mu.Lock()
				if d.Last == nil || !msg.Time.Before(d.Last.Time) {
					mc := msg
					d.Last, d.Updated = &mc, msg.Time
				}
				s.mu.Unlock()
			}
			sink.Emit(model.Event{Kind: kind, Message: &msg})
		}
		s.mu.Lock()
		c := s.chatOf(d)
		s.mu.Unlock()
		sink.Emit(model.Event{Kind: model.EvChat, Chat: &c})
	case "buddylist":
		var bl buddyList
		if json.Unmarshal(ev.EventData, &bl) != nil {
			return
		}
		var chats []model.Chat
		var seen []model.Event
		s.mu.Lock()
		for _, g := range bl.Groups {
			for _, b := range g.Buddies {
				sn := b.id()
				if sn == "" || strings.HasPrefix(sn, "+") {
					continue
				}
				if on, ok := b.online(); ok && !isGroup(sn) {
					seen = append(seen, presenceEvent(sn, on))
				}
				if n := b.name(); n != sn && !isGroup(sn) {
					s.people[sn] = n
				}
				d := s.dialogs[sn]
				if d == nil {
					d = &dialog{SN: sn}
					s.dialogs[sn] = d
				}
				if n := b.name(); n != sn && d.Title == "" {
					d.Title = n
				}
				chats = append(chats, s.chatOf(d))
			}
		}
		s.mu.Unlock()
		for i := range chats {
			sink.Emit(model.Event{Kind: model.EvChat, Chat: &chats[i]})
		}
		for _, ev := range seen {
			sink.Emit(ev)
		}
	case "presence":
		var p person
		if json.Unmarshal(ev.EventData, &p) != nil || p.id() == "" || isGroup(p.id()) {
			return
		}
		if on, ok := p.online(); ok && p.id() != s.selfUser().ID {
			sink.Emit(presenceEvent(p.id(), on))
		}
	case "typing":
		var t typingEvent
		if json.Unmarshal(ev.EventData, &t) != nil || t.TypingStatus != "typing" {
			return
		}
		chat, who := t.AimID, t.AimID
		if t.ChatID != "" {
			chat, who = t.ChatID, t.Sender
		}
		sink.Emit(model.Event{Kind: model.EvTyping, Message: &model.Message{ChatID: chat, Sender: model.User{ID: who, Name: s.personName(who)}}})
	}
}

// presenceEvent: the person sn (whose direct chat is sn) is online or not.
func presenceEvent(sn string, on bool) model.Event {
	return model.Event{Kind: model.EvPresence, Chat: &model.Chat{ID: sn}, Online: on}
}

// --- Messenger -----------------------------------------------------------

func (s *session) live() (string, error) {
	aimsid, online := s.current()
	if !online || aimsid == "" {
		return "", provider.ErrOffline
	}
	return aimsid, nil
}

// maxNameLookups bounds getChatInfo calls per Chats.
const maxNameLookups = 100

// fill learns the chats of a session that knows none (a pasted web session
// resumes the event stream and gets no buddylist or dialog states) and the
// names of group chats.
func (s *session) fill(ctx context.Context) {
	aimsid, online := s.current()
	if !online || aimsid == "" {
		s.env.Logf("info", "vk teams: chats asked while offline")
		return
	}
	s.mu.Lock()
	empty := len(s.dialogs) == 0 && !s.listed
	known := len(s.dialogs)
	s.listed = true
	s.mu.Unlock()
	s.env.Logf("info", "vk teams: chats: %d known, list from contacts: %v", known, empty)
	changed := false
	if empty {
		r, err := s.srv.contactsGet(ctx, aimsid)
		if err != nil {
			s.env.Logf("warn", "vk teams: contacts: %v (answer %s)", err, r.Shape)
		} else {
			s.env.Logf("info", "vk teams: contacts: %d contacts, %d persons; answer %s", len(r.Contacts), len(r.Persons), r.Shape)
		}
		s.learn(r.Persons)
		s.mu.Lock()
		for _, c := range r.Contacts {
			if c.SN != "" && s.dialogs[c.SN] == nil {
				s.dialogs[c.SN] = &dialog{SN: c.SN, Title: s.people[c.SN]}
				changed = true
			}
		}
		s.mu.Unlock()
	}
	if changed {
		s.saveDialogs()
	}
	s.nameGroups()
}

// nameGroups asks getChatInfo for the names of group chats that have none,
// in the background (a hundred calls must not hold up the chat list), and
// tells the names as chat updates.
func (s *session) nameGroups() {
	s.mu.Lock()
	if s.naming || s.sink == nil || s.runCtx == nil {
		s.mu.Unlock()
		return
	}
	var nameless []string
	for sn, d := range s.dialogs {
		if isGroup(sn) && d.Title == "" && !s.unnamed[sn] {
			nameless = append(nameless, sn)
		}
	}
	if len(nameless) == 0 {
		s.mu.Unlock()
		return
	}
	if len(nameless) > maxNameLookups {
		nameless = nameless[:maxNameLookups]
	}
	s.naming = true
	sink, ctx := s.sink, s.runCtx
	s.mu.Unlock()
	go func() {
		defer func() {
			s.mu.Lock()
			s.naming = false
			s.mu.Unlock()
			s.saveDialogs()
		}()
		work := make(chan string)
		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for sn := range work {
					s.nameGroup(ctx, sink, sn)
				}
			}()
		}
		for _, sn := range nameless {
			select {
			case work <- sn:
			case <-ctx.Done():
			}
			if ctx.Err() != nil {
				break
			}
		}
		close(work)
		wg.Wait()
	}()
}

func (s *session) nameGroup(ctx context.Context, sink provider.Sink, sn string) {
	aimsid, online := s.current()
	if !online || aimsid == "" {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	name, err := s.srv.chatName(cctx, aimsid, sn)
	s.mu.Lock()
	d := s.dialogs[sn]
	if d == nil || err != nil || name == "" {
		s.unnamed[sn] = true // do not ask again in this run
		s.mu.Unlock()
		if err != nil && ctx.Err() == nil {
			s.env.Logf("warn", "vk teams: chat name: %v", err)
		}
		return
	}
	d.Title = name
	c := s.chatOf(d)
	c.Unread = -1 // unknown here: keep the stored count
	s.mu.Unlock()
	sink.Emit(model.Event{Kind: model.EvChat, Chat: &c})
}

func (s *session) Chats(ctx context.Context, limit int) ([]model.Chat, error) {
	s.fill(ctx)
	s.mu.Lock()
	out := make([]model.Chat, 0, len(s.dialogs))
	for _, d := range s.dialogs {
		out = append(out, s.chatOf(d))
	}
	s.mu.Unlock()
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *session) History(ctx context.Context, chatID, beforeID string, limit int) ([]model.Message, bool, error) {
	aimsid, err := s.live()
	if err != nil {
		return nil, false, err
	}
	from, n := "-1", limit
	var before int64
	if beforeID != "" {
		from, n = beforeID, limit+1 // the answer may include fromMsgId itself
		before, _ = strconv.ParseInt(beforeID, 10, 64)
	}
	r, err := s.srv.getHistory(ctx, aimsid, chatID, from, n)
	if err != nil {
		return nil, false, err
	}
	s.learn(r.Persons)
	out := make([]model.Message, 0, len(r.Messages))
	for _, m := range r.Messages {
		msg := s.convert(chatID, m)
		if before != 0 && msg.Seq >= before {
			continue
		}
		out = append(out, msg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq > out[j].Seq })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, len(r.Messages) >= n, nil
}

func (s *session) Send(ctx context.Context, chatID string, o model.Outgoing) (model.Message, error) {
	aimsid, err := s.live()
	if err != nil {
		return model.Message{}, err
	}
	// No forwarding and no native quotes of a fragment here: the wire for
	// them is not known (see wire.go), the fragment goes as "> " lines.
	text := o.QuotedText()
	id, err := s.srv.sendMessage(ctx, aimsid, chatID, text, o.ReplyTo)
	if err != nil {
		return model.Message{}, err
	}
	m := model.Message{ChatID: chatID, ID: id, Text: text, Time: time.Now(), Outgoing: true, Sender: s.selfUser(), ReplyTo: o.ReplyTo}
	m.Seq, _ = strconv.ParseInt(id, 10, 64)
	return m, nil
}

func (s *session) MarkRead(ctx context.Context, chatID, msgID string) error {
	aimsid, err := s.live()
	if err != nil {
		return err
	}
	if msgID == "" {
		s.mu.Lock()
		if d := s.dialogs[chatID]; d != nil {
			msgID = d.LastID
		}
		s.mu.Unlock()
		if msgID == "" {
			return nil
		}
	}
	return s.srv.setDlgState(ctx, aimsid, chatID, msgID)
}

// LogOut ends the server session.
func (s *session) LogOut(ctx context.Context) error {
	aimsid, _ := s.current()
	if aimsid == "" {
		return nil
	}
	return s.srv.endSession(ctx, aimsid)
}
