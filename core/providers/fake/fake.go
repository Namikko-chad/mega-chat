// Package fake is an in-memory messenger for tests and for working on the
// app without real accounts (PARLEY_DEMO=1). Its "server" keeps chats
// per login; Inject plays the other side of a conversation.
package fake

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/richtext"
)

// Type is the provider type.
const Type = "fake"

// Server is the shared state of all fake accounts.
type Server struct {
	mu    sync.Mutex
	users map[string]*user // by login
	seq   int64
	// FailAuth makes sessions of these logins fail with ErrAuth.
	FailAuth map[string]bool
	// Special: a new login also gets the saved chat, an archived and a
	// pinned chat (SeedSpecial; the demo).
	Special bool

	calls calls
	// Files sent by the accounts, by media ref ("up:<n>").
	files map[string]sentFile
}

type sentFile struct {
	data []byte
	mime string
}

type user struct {
	chats map[string]*chat
	sinks map[*session]provider.Sink
	mail  *mailbox // mail.go; created on first use
}

type chat struct {
	model.Chat
	msgs    []model.Message            // oldest first
	threads map[string][]model.Message // root id → replies, oldest first
	// A forum: its topics, and the topic of each message by id.
	topics []model.Topic
	topic  map[string]string
	online bool // a direct chat's contact is online
}

// demoTopics are the topics of a demo forum.
var demoTopics = []string{"General", "Design", "Releases", "Off-topic"}

// New returns a provider with an empty server.
func New() *Server {
	return &Server{users: map[string]*user{}, FailAuth: map[string]bool{}, files: map[string]sentFile{}}
}

func (s *Server) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "Demo",
		Methods: []provider.AuthMethod{{ID: "name", Name: "Name", Description: "Any name; demo chats are created"}},
		Caps:    []provider.Capability{provider.CapMessages},
	}
}

func (s *Server) user(login string) *user {
	u := s.users[login]
	if u == nil {
		u = &user{chats: map[string]*chat{}, sinks: map[*session]provider.Sink{}}
		s.users[login] = u
	}
	return u
}

// Seed creates n chats with m messages each for login.
func (s *Server) Seed(login string, n, m int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.user(login)
	base := time.Now().Add(-time.Duration(n*m) * time.Minute)
	for i := range n {
		id := fmt.Sprintf("c%d", i+1)
		c := &chat{Chat: model.Chat{ID: id, Kind: model.ChatDirect, Title: fmt.Sprintf("Contact %d", i+1), Access: model.AccessWrite}, threads: map[string][]model.Message{}}
		switch i % 5 {
		case 2:
			c.Kind, c.Title = model.ChatGroup, fmt.Sprintf("Team %d", i+1)
			if i == 7 {
				c.Title, c.Forum, c.topic = "Forum "+strconv.Itoa(i+1), true, map[string]string{}
				for k, t := range demoTopics {
					c.topics = append(c.topics, model.Topic{ID: "t" + strconv.Itoa(k+1), Title: t, Color: []int{0x6fb9f0, 0xffd67e, 0x8eee98, 0xff93b2}[k], Pinned: k == 0})
				}
			}
		case 4:
			c.Kind, c.Title, c.Access = model.ChatChannel, fmt.Sprintf("News %d", i+1), model.AccessRead
		}
		if i%2 == 0 {
			c.Avatar = "avatar:" + id
		}
		c.online = c.Kind == model.ChatDirect && i%3 == 0
		for j := range m {
			s.seq++
			at := base.Add(time.Duration(i*m+j) * time.Minute)
			msg := model.Message{
				ChatID: id, ID: strconv.FormatInt(s.seq, 10), Seq: s.seq,
				Sender: model.User{ID: "p" + id, Name: c.Title, Avatar: c.Avatar},
				Text:   fmt.Sprintf("message %d of %s", j+1, c.Title),
				Time:   at,
			}
			if j%7 == 3 {
				w, h := 640, 400
				if j%2 == 1 {
					w, h = 360, 480
				}
				ref := fmt.Sprintf("img:%d:%d:%d", s.seq, w, h)
				msg.Attachments = []model.Attachment{{Kind: model.AttachImage, MIME: "image/png", Name: "photo.png", Ref: ref, Thumb: ref, Width: w, Height: h, Size: int64(w * h)}}
			}
			if j%7 == 5 {
				secs := 3 + j%4
				msg.Text = ""
				msg.Attachments = []model.Attachment{{Kind: model.AttachVoice, MIME: "audio/wav", Ref: fmt.Sprintf("wav:%d:%d", s.seq, secs), Duration: time.Duration(secs) * time.Second, Size: int64(secs * 16000)}}
			}
			if j%5 == 1 {
				msg.Reactions = []model.Reaction{{Emoji: "👍", Count: 2}, {Emoji: "🔥", Count: 1, Mine: true}}
			}
			if c.Forum {
				t := c.topics[j%len(c.topics)].ID
				c.topic[msg.ID] = t
				msg.Text = fmt.Sprintf("message %d in %s", j+1, demoTopics[j%len(c.topics)])
			}
			if c.Kind != model.ChatDirect && !c.Forum && j%9 == 4 {
				var replies []model.Message
				for k := range 3 {
					s.seq++
					replies = append(replies, model.Message{
						ChatID: id, ID: strconv.FormatInt(s.seq, 10), Seq: s.seq, InThread: msg.ID,
						Sender: model.User{ID: fmt.Sprintf("q%d", k), Name: fmt.Sprintf("Colleague %d", k+1)},
						Text:   fmt.Sprintf("reply %d in the thread", k+1), Time: at.Add(time.Duration(k+1) * time.Second),
					})
				}
				c.threads[msg.ID] = replies
				msg.Thread = threadOf(msg.ID, replies)
			}
			c.msgs = append(c.msgs, msg)
		}
		u.chats[id] = c
	}
}

// SetOnline turns the contact of a direct chat online or offline.
func (s *Server) SetOnline(login, chatID string, on bool) {
	s.mu.Lock()
	u := s.user(login)
	c := u.chats[chatID]
	if c == nil || c.online == on {
		s.mu.Unlock()
		return
	}
	c.online = on
	var sinks []provider.Sink
	for _, k := range u.sinks {
		sinks = append(sinks, k)
	}
	s.mu.Unlock()
	for _, k := range sinks {
		k.Emit(model.Event{Kind: model.EvPresence, Chat: &model.Chat{ID: chatID}, Online: on})
	}
}

func threadOf(id string, replies []model.Message) *model.Thread {
	t := &model.Thread{ID: id, Replies: len(replies)}
	if n := len(replies); n > 0 {
		t.Last = replies[n-1].Time
		for _, r := range replies[max(0, n-3):] {
			t.Users = append(t.Users, r.Sender)
		}
	}
	return t
}

// Inject delivers an incoming message from the other side and returns it.
func (s *Server) Inject(login, chatID, text string) model.Message {
	s.mu.Lock()
	u := s.user(login)
	c := u.chats[chatID]
	if c == nil {
		c = &chat{Chat: model.Chat{ID: chatID, Kind: model.ChatDirect, Title: "New " + chatID}, threads: map[string][]model.Message{}}
		u.chats[chatID] = c
	}
	s.seq++
	m := model.Message{
		ChatID: chatID, ID: strconv.FormatInt(s.seq, 10), Seq: s.seq,
		Sender: model.User{ID: "p" + chatID, Name: c.Title}, Text: text, Time: time.Now(),
		Mentioned: strings.Contains(text, "@"+login),
	}
	c.msgs = append(c.msgs, m)
	sinks := make([]provider.Sink, 0, len(u.sinks))
	for _, k := range u.sinks {
		sinks = append(sinks, k)
	}
	s.mu.Unlock()
	for _, k := range sinks {
		mc := m
		k.Emit(model.Event{Kind: model.EvMessageNew, Message: &mc})
	}
	return m
}

// Miss adds an incoming message without telling the sessions, like an
// event lost while a connection was down.
func (s *Server) Miss(login, chatID, text string) model.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.user(login).chats[chatID]
	s.seq++
	m := model.Message{
		ChatID: chatID, ID: strconv.FormatInt(s.seq, 10), Seq: s.seq,
		Sender: model.User{ID: "p" + chatID, Name: c.Title}, Text: text, Time: time.Now(),
	}
	c.msgs = append(c.msgs, m)
	return m
}

// Messages returns the server-side history of a chat, oldest first.
func (s *Server) Messages(login, chatID string) []model.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c := s.user(login).chats[chatID]; c != nil {
		return slices.Clone(c.msgs)
	}
	return nil
}

// --- login ---------------------------------------------------------------

type login struct{ s *Server }

func (s *Server) Login(_ context.Context, method string, acct *provider.Account, _ provider.Env) (provider.Login, error) {
	return &login{s: s}, nil
}

func (l *login) Next(_ context.Context, values map[string]string) (provider.Step, error) {
	step := provider.Step{
		Title: "Demo account", Message: "Pick a name. Ten demo chats are created for it, one of them a forum with topics.",
		Fields: []provider.Field{{Key: "name", Label: "Name"}},
	}
	if values == nil {
		return step, nil
	}
	name := strings.TrimSpace(values["name"])
	if name == "" {
		step.Error = "Enter a name"
		return step, nil
	}
	l.s.mu.Lock()
	_, exists := l.s.users[name]
	l.s.mu.Unlock()
	if !exists {
		l.s.Seed(name, 10, 30)
		if l.s.Special {
			l.s.SeedSpecial(name)
		}
	}
	return provider.Step{Done: true, Result: &provider.Credentials{
		Login: name, Name: "Demo " + name,
		Params: map[string]string{"name": name},
		State:  map[string][]byte{"joined": []byte(time.Now().Format(time.RFC3339))},
	}}, nil
}

func (l *login) Close() {}

// --- session -------------------------------------------------------------

type session struct {
	s     *Server
	login string
}

func (s *Server) Open(acct provider.Account, _ provider.Env) (provider.Session, error) {
	return &session{s: s, login: acct.Login}, nil
}

func (c *session) Run(ctx context.Context, sink provider.Sink) error {
	c.s.mu.Lock()
	if c.s.FailAuth[c.login] {
		c.s.mu.Unlock()
		return fmt.Errorf("fake: %w", provider.ErrAuth)
	}
	u := c.s.user(c.login)
	u.sinks[c] = sink
	var online []string
	for id, ch := range u.chats {
		if ch.online {
			online = append(online, id)
		}
	}
	c.s.mu.Unlock()
	defer func() {
		c.s.mu.Lock()
		delete(u.sinks, c)
		c.s.mu.Unlock()
	}()
	sink.Connected(model.User{ID: "me", Name: c.login})
	for _, id := range online {
		sink.Emit(model.Event{Kind: model.EvPresence, Chat: &model.Chat{ID: id}, Online: true})
	}
	<-ctx.Done()
	return ctx.Err()
}

func (c *session) Chats(_ context.Context, limit int) ([]model.Chat, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	var out []model.Chat
	for _, ch := range c.s.user(c.login).chats {
		x := ch.Chat
		x.Folders = folderIDs(x)
		if n := len(ch.msgs); n > 0 {
			last := ch.msgs[n-1]
			x.Last, x.Updated = &last, last.Time
		}
		out = append(out, x)
	}
	slices.SortFunc(out, func(a, b model.Chat) int { return b.Updated.Compare(a.Updated) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (c *session) History(_ context.Context, chatID, beforeID string, limit int) ([]model.Message, bool, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return nil, false, fmt.Errorf("fake: no chat %s", chatID)
	}
	end := len(ch.msgs)
	if beforeID != "" {
		end = slices.IndexFunc(ch.msgs, func(m model.Message) bool { return m.ID == beforeID })
		if end < 0 {
			return nil, false, fmt.Errorf("fake: no message %s", beforeID)
		}
	}
	start := max(0, end-limit)
	out := slices.Clone(ch.msgs[start:end])
	slices.Reverse(out)
	return out, start > 0, nil
}

// later delivers an incoming message to chatID after a few seconds.
func (c *session) later(chatID string, text ...string) {
	time.Sleep(PingDelay)
	c.s.Inject(c.login, chatID, strings.Join(text, ""))
}

// PingDelay is how long the demo chats take to answer "/ping" and "/mention".
var PingDelay = 3 * time.Second

func (c *session) Send(_ context.Context, chatID string, o model.Outgoing) (model.Message, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return model.Message{}, fmt.Errorf("fake: no chat %s", chatID)
	}
	c.s.seq++
	m := model.Message{
		ChatID: chatID, ID: strconv.FormatInt(c.s.seq, 10), Seq: c.s.seq,
		Sender: model.User{ID: "me", Name: c.login}, Text: o.Text, ReplyTo: o.ReplyTo,
		Time: time.Now(), Outgoing: true, InThread: o.Thread,
	}
	if o.ReplyTo != "" {
		m.Quote = o.Quote // a native quote, as Telegram has
	} else {
		m.Text = o.QuotedText()
	}
	if o.Quote == "" {
		d := richtext.Parse(m.Text, richtext.Markdown)
		m.Text, m.Format = d.Text, d.Spans
	}
	if o.Thread != "" && ch.Forum {
		if !slices.ContainsFunc(ch.topics, func(t model.Topic) bool { return t.ID == o.Thread }) {
			return model.Message{}, fmt.Errorf("fake: no topic %s", o.Thread)
		}
		ch.msgs = append(ch.msgs, m)
		ch.topic[m.ID] = o.Thread
		// The server's echo puts it into the chat's history too.
		echo := m
		echo.InThread = ""
		go c.emit(model.Event{Kind: model.EvMessageNew, Message: &echo})
		return m, nil
	}
	if o.Thread != "" {
		i := slices.IndexFunc(ch.msgs, func(x model.Message) bool { return x.ID == o.Thread })
		if i < 0 {
			return model.Message{}, fmt.Errorf("fake: no message %s", o.Thread)
		}
		ch.threads[o.Thread] = append(ch.threads[o.Thread], m)
		ch.msgs[i].Thread = threadOf(o.Thread, ch.threads[o.Thread])
		root := ch.msgs[i]
		go c.emit(model.Event{Kind: model.EvMessageEdit, Message: &root})
		return m, nil
	}
	ch.msgs = append(ch.msgs, m)
	if call, video := callCommand(o.Text); call && ch.Kind == model.ChatDirect {
		go c.ring(chatID, video)
	}
	// "/ping" and "/mention": the chat answers a few seconds later, time
	// to switch to another window and see the notification.
	switch strings.TrimSpace(o.Text) {
	case "/ping":
		go c.later(chatID, "Pong")
	case "/mention":
		go c.later(chatID, "@"+c.login, " look at this")
	}
	return m, nil
}

// SendFiles sends one message per file, like Telegram: several make an
// album, the caption goes with the first one. Thread replies are not
// supported (the composer of a thread has no files here).
func (c *session) SendFiles(_ context.Context, chatID string, o model.Outgoing) ([]model.Message, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return nil, fmt.Errorf("fake: no chat %s", chatID)
	}
	if o.Thread != "" {
		return nil, provider.ErrUnsupported
	}
	var out []model.Message
	group := ""
	if len(o.Files) > 1 {
		group = "up" + strconv.FormatInt(c.s.seq+1, 10)
	}
	for i, f := range o.Files {
		c.s.seq++
		ref := "up:" + strconv.FormatInt(c.s.seq, 10)
		c.s.files[ref] = sentFile{data: slices.Clone(f.Data), mime: f.MIME}
		a := model.Attachment{Kind: f.Kind(), Name: f.Name, MIME: f.MIME, Size: int64(len(f.Data)), Ref: ref}
		if a.Kind == model.AttachImage {
			if cfg, _, err := image.DecodeConfig(bytes.NewReader(f.Data)); err == nil {
				a.Width, a.Height, a.Thumb = cfg.Width, cfg.Height, ref
			}
		}
		m := model.Message{
			ChatID: chatID, ID: strconv.FormatInt(c.s.seq, 10), Seq: c.s.seq,
			Sender: model.User{ID: "me", Name: c.login}, Attachments: []model.Attachment{a},
			Time: time.Now(), Outgoing: true, Group: group,
		}
		if i == 0 {
			m.Text, m.ReplyTo = o.Text, o.ReplyTo
			if o.ReplyTo != "" {
				m.Quote = o.Quote
			} else {
				m.Text = o.QuotedText()
			}
		}
		ch.msgs = append(ch.msgs, m)
		out = append(out, m)
	}
	return out, nil
}

// Forward copies messages of a chat into another one, attachments
// included, marked with their original author.
func (c *session) Forward(_ context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	u := c.s.user(c.login)
	from, to := u.chats[fromChat], u.chats[toChat]
	if from == nil || to == nil {
		return nil, fmt.Errorf("fake: no chat %s or %s", fromChat, toChat)
	}
	var out []model.Message
	for _, id := range msgIDs {
		i := slices.IndexFunc(from.msgs, func(m model.Message) bool { return m.ID == id })
		if i < 0 {
			return out, fmt.Errorf("fake: no message %s", id)
		}
		src := from.msgs[i]
		if src.Deleted || src.Service {
			continue
		}
		c.s.seq++
		m := model.Message{
			ChatID: toChat, ID: strconv.FormatInt(c.s.seq, 10), Seq: c.s.seq,
			Sender: model.User{ID: "me", Name: c.login}, Text: src.Text,
			Attachments: slices.Clone(src.Attachments), Time: time.Now(), Outgoing: true,
			Forwarded: true, ForwardedFrom: src.Sender.Name,
		}
		if src.Forwarded {
			m.ForwardedFrom = src.ForwardedFrom
		}
		if src.Group != "" { // an album stays an album

			m.Group = "fwd" + src.Group
		}
		to.msgs = append(to.msgs, m)
		out = append(out, m)
	}
	return out, nil
}

// emit tells the account's sessions about a change.
func (c *session) emit(ev model.Event) {
	c.s.mu.Lock()
	var sinks []provider.Sink
	for _, k := range c.s.user(c.login).sinks {
		sinks = append(sinks, k)
	}
	c.s.mu.Unlock()
	for _, k := range sinks {
		k.Emit(ev)
	}
}

func (c *session) Thread(_ context.Context, chatID, threadID, beforeID string, limit int) ([]model.Message, bool, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return nil, false, fmt.Errorf("fake: no chat %s", chatID)
	}
	all := ch.threads[threadID]
	if ch.Forum {
		all = nil
		for _, m := range ch.msgs {
			if ch.topic[m.ID] == threadID {
				all = append(all, m)
			}
		}
	}
	end := len(all)
	if beforeID != "" {
		end = max(0, slices.IndexFunc(all, func(m model.Message) bool { return m.ID == beforeID }))
	}
	start := max(0, end-limit)
	out := slices.Clone(all[start:end])
	slices.Reverse(out)
	return out, start > 0, nil
}

func (c *session) Topics(_ context.Context, chatID string) ([]model.Topic, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil || !ch.Forum {
		return nil, fmt.Errorf("fake: no forum %s", chatID)
	}
	out := slices.Clone(ch.topics)
	for i := range out {
		for j := len(ch.msgs) - 1; j >= 0; j-- {
			if m := ch.msgs[j]; ch.topic[m.ID] == out[i].ID {
				m.InThread = out[i].ID
				out[i].Last, out[i].Updated = &m, m.Time
				break
			}
		}
	}
	slices.SortStableFunc(out, func(a, b model.Topic) int {
		if a.Pinned != b.Pinned {
			if a.Pinned {
				return -1
			}
			return 1
		}
		return b.Updated.Compare(a.Updated)
	})
	return out, nil
}

// Delete marks the user's own messages deleted (for everyone or not, the
// demo has one reader).
func (c *session) Delete(_ context.Context, chatID string, msgIDs []string, _ bool) error {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return fmt.Errorf("fake: no chat %s", chatID)
	}
	del := func(list []model.Message, id string) bool {
		i := slices.IndexFunc(list, func(x model.Message) bool { return x.ID == id })
		if i < 0 || !list[i].Outgoing {
			return false
		}
		list[i].Deleted, list[i].Text, list[i].Attachments = true, "", nil
		return true
	}
next:
	for _, id := range msgIDs {
		if del(ch.msgs, id) {
			continue
		}
		for _, replies := range ch.threads {
			if del(replies, id) {
				continue next
			}
		}
		return fmt.Errorf("fake: no message of yours %s", id)
	}
	return nil
}

// Members: a direct chat's contact, or a few colleagues of a group.
func (c *session) Members(_ context.Context, chatID, query string, limit int) ([]model.User, error) {
	c.s.mu.Lock()
	ch := c.s.user(c.login).chats[chatID]
	online := ch != nil && ch.online
	c.s.mu.Unlock()
	if ch == nil {
		return nil, fmt.Errorf("fake: no chat %s", chatID)
	}
	var all []model.User
	if ch.Kind == model.ChatDirect {
		all = []model.User{{ID: chatID, Name: ch.Title, Username: strings.ToLower(strings.ReplaceAll(ch.Title, " ", "_")), Avatar: ch.Avatar, Online: online}}
	} else {
		for k, n := range []string{"Alice Cooper", "Bob Stone", "Carol King", "Dave Grohl", "Eve Moneypenny"} {
			all = append(all, model.User{ID: fmt.Sprintf("q%d", k), Name: n, Username: strings.ToLower(strings.Fields(n)[0]), Online: k%2 == 0})
		}
	}
	var out []model.User
	for _, u := range all {
		if model.MatchUser(u, query) && len(out) < limit {
			out = append(out, u)
		}
	}
	return out, nil
}

// DeletesForMe: the demo can.
func (c *session) DeletesForMe() bool { return true }

func (c *session) Edit(_ context.Context, chatID, msgID, text string) (model.Message, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return model.Message{}, fmt.Errorf("fake: no chat %s", chatID)
	}
	edit := func(list []model.Message) (model.Message, bool) {
		i := slices.IndexFunc(list, func(x model.Message) bool { return x.ID == msgID })
		if i < 0 || !list[i].Outgoing || list[i].Deleted {
			return model.Message{}, false
		}
		d := richtext.Parse(text, richtext.Markdown)
		list[i].Text, list[i].Format, list[i].Edited = d.Text, d.Spans, time.Now()
		return list[i], true
	}
	if m, ok := edit(ch.msgs); ok {
		return m, nil
	}
	for _, replies := range ch.threads {
		if m, ok := edit(replies); ok {
			return m, nil
		}
	}
	return model.Message{}, fmt.Errorf("fake: no message of yours %s", msgID)
}

func (c *session) Profile(_ context.Context, chatID string) (model.Profile, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return model.Profile{}, fmt.Errorf("fake: no chat %s", chatID)
	}
	p := model.Profile{Title: ch.Title, Avatar: ch.Avatar, Online: ch.online}
	slug := strings.ToLower(strings.ReplaceAll(ch.Title, " ", "_"))
	switch ch.Kind {
	case model.ChatDirect:
		p.Username, p.Phone, p.Email = slug, "+1 555 0100", slug+"@example.org"
		p.Position = "Demo contact"
		p.About = "A contact of the demo messenger. Docs: https://example.org/demo"
		if !ch.online {
			p.LastSeen = time.Now().Add(-42 * time.Minute)
		}
	default:
		p.Username, p.Link = slug, "https://example.org/"+slug
		p.About = "A " + map[model.ChatKind]string{model.ChatGroup: "group", model.ChatChannel: "channel"}[ch.Kind] + " of the demo messenger.\nRules: https://example.org/rules"
		p.Members = 3 + len(ch.Title)
		if ch.Kind == model.ChatGroup {
			for k := range 3 {
				p.MemberList = append(p.MemberList, model.User{ID: fmt.Sprintf("q%d", k), Name: fmt.Sprintf("Colleague %d", k+1)})
			}
		}
	}
	return p, nil
}

func (c *session) ReadTopic(context.Context, string, string, string) error { return nil }

func (c *session) React(_ context.Context, chatID, msgID, emoji string, on bool) ([]model.Reaction, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	ch := c.s.user(c.login).chats[chatID]
	if ch == nil {
		return nil, fmt.Errorf("fake: no chat %s", chatID)
	}
	i := slices.IndexFunc(ch.msgs, func(m model.Message) bool { return m.ID == msgID })
	if i < 0 {
		return nil, fmt.Errorf("fake: no message %s", msgID)
	}
	rs := slices.Clone(ch.msgs[i].Reactions)
	j := slices.IndexFunc(rs, func(r model.Reaction) bool { return r.Emoji == emoji })
	switch {
	case on && j < 0:
		rs = append(rs, model.Reaction{Emoji: emoji, Count: 1, Mine: true})
	case on && !rs[j].Mine:
		rs[j].Count++
		rs[j].Mine = true
	case !on && j >= 0 && rs[j].Mine:
		rs[j].Count--
		rs[j].Mine = false
		if rs[j].Count == 0 {
			rs = slices.Delete(rs, j, j+1)
		}
	}
	ch.msgs[i].Reactions = rs
	return rs, nil
}

// Media draws avatars ("avatar:<chat>") and pictures ("img:<n>:<w>:<h>"),
// and plays tones ("wav:<n>:<seconds>").
func (c *session) Media(_ context.Context, ref string, limit int64) ([]byte, string, error) {
	var img image.Image
	switch {
	case strings.HasPrefix(ref, "up:"):
		c.s.mu.Lock()
		f, ok := c.s.files[ref]
		c.s.mu.Unlock()
		if !ok {
			return nil, "", fmt.Errorf("fake: no file %q", ref)
		}
		if int64(len(f.data)) > limit {
			return nil, "", provider.ErrTooLarge
		}
		return f.data, f.mime, nil
	case strings.HasPrefix(ref, "wav:"):
		parts := strings.Split(ref, ":")
		if len(parts) != 3 {
			return nil, "", fmt.Errorf("fake: bad ref %q", ref)
		}
		n, _ := strconv.Atoi(parts[1])
		secs, _ := strconv.Atoi(parts[2])
		b := tone(min(max(secs, 1), 30), 330+float64(n%6)*55)
		if int64(len(b)) > limit {
			return nil, "", provider.ErrTooLarge
		}
		return b, "audio/wav", nil
	case strings.HasPrefix(ref, "avatar:"):
		h := 0
		for _, c := range ref {
			h = h*31 + int(c)
		}
		img = picture(96, 96, h)
	case strings.HasPrefix(ref, "img:"):
		parts := strings.Split(ref, ":")
		if len(parts) != 4 {
			return nil, "", fmt.Errorf("fake: bad ref %q", ref)
		}
		n, _ := strconv.Atoi(parts[1])
		w, _ := strconv.Atoi(parts[2])
		h, _ := strconv.Atoi(parts[3])
		img = picture(min(max(w, 1), 2000), min(max(h, 1), 2000), n)
	default:
		return nil, "", fmt.Errorf("fake: bad ref %q", ref)
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, "", err
	}
	if int64(b.Len()) > limit {
		return nil, "", provider.ErrTooLarge
	}
	return b.Bytes(), "image/png", nil
}

// picture is a diagonal gradient whose colours depend on seed.
func picture(w, h, seed int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r0, g0, b0 := uint8(seed*53), uint8(seed*97), uint8(seed*31+120)
	for y := range h {
		for x := range w {
			t := float64(x+y) / float64(w+h)
			img.Set(x, y, color.RGBA{uint8(float64(r0) * (1 - t)), uint8(float64(g0)*t + 40), b0, 255})
		}
	}
	return img
}

func (c *session) MarkRead(context.Context, string, string) error { return nil }

// tone is a mono 8 kHz 16-bit WAV of a soft beeping tone.
func tone(secs int, freq float64) []byte {
	const rate = 8000
	n := secs * rate
	var b bytes.Buffer
	le := binary.LittleEndian
	b.WriteString("RIFF")
	_ = binary.Write(&b, le, uint32(36+2*n))
	b.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(rate), uint32(2 * rate), uint16(2), uint16(16)} {
		_ = binary.Write(&b, le, v)
	}
	b.WriteString("data")
	_ = binary.Write(&b, le, uint32(2*n))
	for i := range n {
		t := float64(i) / rate
		env := 0.5 + 0.5*math.Sin(2*math.Pi*2*t) // two beats a second
		_ = binary.Write(&b, le, int16(6000*env*math.Sin(2*math.Pi*freq*t)))
	}
	return b.Bytes()
}
