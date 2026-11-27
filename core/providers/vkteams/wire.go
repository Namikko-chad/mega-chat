package vkteams

// Everything about the VK Teams client protocol lives in this file.
//
// VK Teams (ex Myteam, built on ICQ New) has no public client API. What is
// known: the client talks to "/api/vNN/rapi/<method>" with JSON bodies and
// answers {"status":{"code":20000},"results":{...}} (seen in VK WorkSpace
// client logs, methods getHistory, getChatInfo, sendIM, setDlgState,
// startSession, clientLogin, ...), and the session token is "aimsid". The
// rest follows the ICQ/AIM WIM design VK Teams descends from: clientLogin
// → token + sessionSecret, a signed startSession → aimsid + fetchBaseURL,
// long-polling fetchEvents, events "myInfo", "histDlgState", "typing".
//
// UNVERIFIED against a live server: paths under "wim/", the event and
// field names below, RAPI parameter names and status codes other than
// 20000. Check them against the traffic of the desktop or web client and
// fix them here; tests in wire_test.go pin this file's assumptions.

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/richtext"
)

const (
	// What the VK WorkSpace desktop client (2026) talks to.
	defaultHost    = "u.myteam.vmailru.net"
	defaultVersion = "v132"
	fetchTimeout   = 25 * time.Second
	// startEvents are the event kinds a session subscribes to.
	startEvents = "myInfo,presence,buddylist,typing,hiddenChat,hist,mchat,sentIM,imState,dataIM,offlineIM,userAddedToBuddyList,service,lifestream,apps,permitDeny,diff,webrtcMsg,reactions,threadUpdate"
)

// server is where an account's server lives.
type server struct {
	scheme  string // "https"; tests use "http"
	host    string // "u.myteam.vmailru.net" or the company's own host
	version string // "v132"
	http    *http.Client
	reqSeq  atomic.Int64
}

func (s *server) base() string { return s.scheme + "://" + s.host + "/api/" + s.version + "/" }
func (s *server) wim() string  { return s.base() + "wim/" }
func (s *server) rapi() string { return s.base() + "rapi/" }

func (s *server) reqID() string {
	return strconv.FormatInt(s.reqSeq.Add(1), 10) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
}

// errSession: the aimsid is no longer valid (a new session is needed).
type errSession struct{ detail string }

func (e *errSession) Error() string { return "vk teams: session expired: " + e.detail }

// errLogin: the credentials (token, password) were rejected.
type errLogin struct{ detail string }

func (e *errLogin) Error() string { return "vk teams: sign-in rejected: " + e.detail }

// noURL drops the request URL from a transport error: it carries the aimsid.
func noURL(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		host := ""
		if u, perr := url.Parse(ue.URL); perr == nil {
			host = u.Host
		}
		return fmt.Errorf("vk teams: %s %s: %w", ue.Op, host, ue.Err)
	}
	return err
}

// --- WIM (form requests, {"response":{statusCode,data}} answers) -----------

type wimResponse struct {
	Response struct {
		StatusCode int             `json:"statusCode"`
		StatusText string          `json:"statusText"`
		Data       json.RawMessage `json:"data"`
	} `json:"response"`
}

// wimSessionCodes mean "this aimsid/token is dead".
var wimSessionCodes = map[int]bool{401: true, 460: true, 462: true, 607: true}

func (s *server) wimDo(ctx context.Context, method, rawURL string, form url.Values, out any) error {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return err
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return noURL(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vk teams: HTTP %d", resp.StatusCode)
	}
	var w wimResponse
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("vk teams: bad answer: %w", err)
	}
	switch code := w.Response.StatusCode; {
	case code == 200:
	case wimSessionCodes[code]:
		return &errSession{detail: fmt.Sprintf("%d %s", code, w.Response.StatusText)}
	case code == 330 || code == 408:
		return &errLogin{detail: fmt.Sprintf("%d %s", code, w.Response.StatusText)}
	default:
		return fmt.Errorf("vk teams: %d %s", code, w.Response.StatusText)
	}
	if out != nil && len(w.Response.Data) > 0 {
		return json.Unmarshal(w.Response.Data, out)
	}
	return nil
}

// loginResult is clientLogin's data.
type loginResult struct {
	Token struct {
		A         string `json:"a"`
		ExpiresIn int64  `json:"expiresIn"`
	} `json:"token"`
	SessionSecret string `json:"sessionSecret"`
	HostTime      int64  `json:"hostTime"`
	LoginID       string `json:"loginId"`
}

// clientLogin exchanges login and password for a token.
func (s *server) clientLogin(ctx context.Context, devID, login, password string) (loginResult, error) {
	var r loginResult
	form := url.Values{
		"f": {"json"}, "devId": {devID}, "s": {login}, "pwd": {password},
		"clientName": {"Parley"}, "idType": {"ICQ"}, "tokenType": {"longTerm"},
	}
	err := s.wimDo(ctx, http.MethodPost, s.wim()+"auth/clientLogin", form, &r)
	if err == nil && r.Token.A == "" {
		err = &errLogin{detail: "no token in the answer"}
	}
	return r, err
}

// sessionKey is derived from the password and the login's secret; it
// signs startSession, so the password itself is never stored.
func sessionKey(password, secret string) string {
	m := hmac.New(sha256.New, []byte(password))
	m.Write([]byte(secret))
	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

// sign adds ts and sig_sha256: HMAC-SHA256(sessionKey, "POST&url&params"),
// parameters sorted and percent-encoded as in OAuth 1.0.
func sign(method, rawURL string, params url.Values, key string, ts int64) {
	params.Set("ts", strconv.FormatInt(ts, 10))
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sig_sha256" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var q []string
	for _, k := range keys {
		for _, v := range params[k] {
			q = append(q, oauthEscape(k)+"="+oauthEscape(v))
		}
	}
	base := method + "&" + oauthEscape(rawURL) + "&" + oauthEscape(strings.Join(q, "&"))
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(base))
	params.Set("sig_sha256", base64.StdEncoding.EncodeToString(m.Sum(nil)))
}

func oauthEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// sessionInfo is startSession's data.
type sessionInfo struct {
	AimSID       string `json:"aimsid"`
	FetchBaseURL string `json:"fetchBaseURL"`
	MyInfo       person `json:"myInfo"`
}

// startSession opens a session with a token. timeOffset is the server's
// clock minus ours (from clientLogin's hostTime).
func (s *server) startSession(ctx context.Context, devID, token, key string, timeOffset int64) (sessionInfo, error) {
	u := s.wim() + "aim/startSession"
	p := url.Values{
		"a": {token}, "f": {"json"}, "k": {devID}, "events": {startEvents},
		"clientName": {"Parley"}, "imf": {"plain"}, "invisible": {"false"}, "view": {"online"},
		"includePresenceFields": {"aimId,displayId,friendly,state,userType"},
		"sessionTimeout":        {"2592000"}, "nonce": {s.reqID()},
	}
	sign(http.MethodPost, u, p, key, time.Now().Unix()+timeOffset)
	var r sessionInfo
	err := s.wimDo(ctx, http.MethodPost, u, p, &r)
	if err == nil && r.AimSID == "" {
		err = &errLogin{detail: "no aimsid in the answer"}
	}
	var se *errSession
	if errors.As(err, &se) {
		err = &errLogin{detail: se.detail} // a dead token, not a dead session
	}
	return r, err
}

// firstFetchURL is where polling starts for a session known only by its
// aimsid (pasted from the web client). A guess: the real client polls
// "/api/vNN/bos/<node>/aim/fetchEvents", and the node is known only from
// startSession or a fetchEvents URL (see parseFetchURL).
func (s *server) firstFetchURL(aimsid string) string {
	return s.wim() + "fetchEvents?aimsid=" + url.QueryEscape(aimsid) + "&first=1"
}

// fetchQueryDrop are fetchEvents parameters set per request, not part of
// the session's place.
var fetchQueryDrop = []string{"r", "rnd", "timeout", "f", "peek", "hidden"}

// parseFetchURL takes a fetchEvents URL copied from a client's traffic
// ("https://host/api/v132/bos/<node>/aim/fetchEvents?aimsid=…&seqNum=…") and
// returns its host, API version, aimsid and the URL to poll next. ok is
// false when raw is not such a URL.
func parseFetchURL(raw string) (host, version, aimsid, next string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || !strings.HasSuffix(u.Path, "/fetchEvents") {
		return "", "", "", "", false
	}
	q := u.Query()
	aimsid = q.Get("aimsid")
	if aimsid == "" {
		return "", "", "", "", false
	}
	for _, k := range fetchQueryDrop {
		q.Del(k)
	}
	u.RawQuery = q.Encode()
	u.Fragment = ""
	if parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/"); len(parts) > 1 && parts[0] == "api" {
		version = parts[1]
	}
	return u.Host, version, aimsid, u.String(), true
}

// aimsidLogin is the account an aimsid belongs to: aimsids end with
// ":<login>" ("005.1234567890.1234567890:user@corp.ru").
func aimsidLogin(aimsid string) string {
	if i := strings.LastIndexByte(aimsid, ':'); i >= 0 {
		return aimsid[i+1:]
	}
	return ""
}

// normVersion makes "150", "/api/v150/" and "v150" all "v150".
func normVersion(s string) string {
	s = strings.Trim(strings.TrimSpace(s), "/")
	s = strings.TrimPrefix(s, "api/")
	if s != "" && s[0] >= '0' && s[0] <= '9' {
		s = "v" + s
	}
	return s
}

// rawEvent is one fetchEvents event.
type rawEvent struct {
	Type      string          `json:"type"`
	SeqNum    int64           `json:"seqNum"`
	EventData json.RawMessage `json:"eventData"`
}

type fetchResult struct {
	FetchBaseURL    string     `json:"fetchBaseURL"`
	TimeToNextFetch int64      `json:"timeToNextFetch"`
	Events          []rawEvent `json:"events"`
}

// fetchEvents long-polls once; the answer's FetchBaseURL is the next URL.
func (s *server) fetchEvents(ctx context.Context, fetchURL string, timeout time.Duration) (fetchResult, error) {
	u, err := url.Parse(fetchURL)
	if err != nil {
		return fetchResult{}, err
	}
	q := u.Query()
	q.Set("f", "json")
	q.Set("timeout", strconv.FormatInt(timeout.Milliseconds(), 10))
	q.Set("r", s.reqID())
	u.RawQuery = q.Encode()
	ctx, cancel := context.WithTimeout(ctx, timeout+20*time.Second)
	defer cancel()
	var r fetchResult
	err = s.wimDo(ctx, http.MethodGet, u.String(), nil, &r)
	return r, err
}

// person is a user as events and RAPI describe them.
type person struct {
	AimID     string `json:"aimId"`
	SN        string `json:"sn"`
	Friendly  string `json:"friendly"`
	Nick      string `json:"nick"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// State is the presence: "online", "mobile", "occupied", "away",
	// "offline" (the presence event and buddylist; unverified).
	State string `json:"state"`
}

// online tells whether p is online; known is false when p has no state.
func (p person) online() (on, known bool) {
	switch p.State {
	case "":
		return false, false
	case "online", "mobile", "occupied":
		return true, true
	}
	return false, true
}

func (p person) id() string {
	if p.AimID != "" {
		return p.AimID
	}
	return p.SN
}

// name is p's display name. VK WorkSpace often sends the login (an email)
// as "friendly" and the real name only in firstName/lastName, so a field
// that just repeats the id does not count.
func (p person) name() string {
	id := p.id()
	full := strings.TrimSpace(p.FirstName + " " + p.LastName)
	for _, n := range []string{p.Friendly, full, p.Nick} {
		if n = strings.TrimSpace(n); n != "" && !strings.EqualFold(n, id) {
			return n
		}
	}
	return id
}

// wInt is a number the server sends either bare or quoted (seen:
// filesharing size "12345"); empty, null or unparsable is 0.
type wInt int64

func (n *wInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr != nil {
			*n = 0
			return nil
		}
		v = int64(f)
	}
	*n = wInt(v)
	return nil
}

// wMessage is a message in histDlgState and getHistory.
type wMessage struct {
	MsgID json.Number `json:"msgId"`
	Time  int64       `json:"time"`
	Text  string      `json:"text"`
	// Parts holds the text when "text" is empty, and quotes and forwards.
	Parts json.RawMessage `json:"parts"`
	// Format: native styles of Text, as in the Bot API (unverified here).
	Format   json.RawMessage `json:"format"`
	Outgoing bool            `json:"outgoing"`
	Chat     struct {
		Sender string `json:"sender"`
		Name   string `json:"name"`
		// What a service message tells (see serviceText).
		MemberEvent  json.RawMessage `json:"memberEvent"`
		ModifiedInfo json.RawMessage `json:"modifiedInfo"`
	} `json:"chat"`
	// Thread: a discussion hangs under this message.
	Thread      *wThread        `json:"thread"`
	Reactions   json.RawMessage `json:"reactions"` // {"exist":true}: ask reaction/get
	Class       string          `json:"class"`     // "event" for service messages
	UpdatePatch []struct {
		Type string `json:"type"` // "delete", "edit"
	} `json:"updatePatch"`
	Filesharing []struct {
		Name        string `json:"name"`
		Mime        string `json:"mime"`
		Size        wInt   `json:"size"`
		ID          string `json:"id"`
		Duration    wInt   `json:"duration"`
		OriginalURL string `json:"original_url"`
	} `json:"filesharing"`
	// shape lists the keys of a message that came without "text" (no
	// values), for the log while this format is unverified.
	shape string
}

func (m *wMessage) UnmarshalJSON(b []byte) error {
	type plain wMessage
	if err := json.Unmarshal(b, (*plain)(m)); err != nil {
		return err
	}
	if m.Text == "" || len(m.Filesharing) > 0 {
		m.shape = shapeOf(b, 4)
	}
	return nil
}

// wPart is one part of a message: its own text, or a quoted or forwarded
// message with its author (sn) and the chat it came from.
type wPart struct {
	MediaType string          `json:"mediaType"` // "text", "quote", "forward"
	Text      json.RawMessage `json:"text"`
	Format    json.RawMessage `json:"format"`
	MsgID     json.RawMessage `json:"msgId"`
	SN        string          `json:"sn"`
	Chat      struct {
		Name string `json:"name"`
		SN   string `json:"sn"`
	} `json:"chat"`
}

// text reads the part's text: a string, or {"plain":…} as message/send has it.
func (p wPart) text() string {
	var s string
	if json.Unmarshal(p.Text, &s) == nil {
		return s
	}
	var o struct {
		Plain string `json:"plain"`
	}
	if json.Unmarshal(p.Text, &o) == nil {
		return o.Plain
	}
	return ""
}

// origin names where a forwarded part comes from: its author, and the group
// when there is one. name resolves an id; it returns the id when unknown.
func (p wPart) origin(name func(string) string) string {
	who := ""
	if p.SN != "" {
		if who = name(p.SN); who == p.SN && p.Chat.Name != "" {
			who = ""
		}
	}
	switch {
	case who == "":
		return p.Chat.Name
	case isGroup(p.Chat.SN) && p.Chat.Name != "" && p.Chat.Name != who:
		return who + " · " + p.Chat.Name
	}
	return who
}

// msgContent is what a message's "text" and parts say together.
type msgContent struct {
	text      string
	format    []model.TextSpan // native styles of text, if it has any
	forwarded bool
	from      string // the one origin of what is forwarded; empty if several
	replyTo   string
	quote     string
}

func readParts(raw json.RawMessage) []wPart {
	var list []wPart
	if json.Unmarshal(raw, &list) == nil {
		return list
	}
	var o struct {
		MainPart wPart `json:"mainPart"`
	}
	if json.Unmarshal(raw, &o) == nil && len(o.MainPart.Text) > 0 {
		return []wPart{o.MainPart}
	}
	return nil
}

// content puts the message together. A "text" of its own wins over the
// parts' texts, which it then repeats. Several forwarded messages from
// different people each get their author's name above them.
func (m *wMessage) content(name func(string) string) msgContent {
	var c msgContent
	parts := readParts(m.Parts)
	origins := map[string]bool{}
	for _, p := range parts {
		if p.MediaType == "forward" {
			c.forwarded = true
			origins[p.origin(name)] = true
		}
	}
	if len(origins) == 1 {
		for o := range origins {
			c.from = o
		}
	}
	var out []string
	for _, p := range parts {
		t := p.text()
		switch p.MediaType {
		case "quote":
			if c.replyTo == "" {
				c.replyTo, c.quote = strings.Trim(string(p.MsgID), `"`), t
			}
		case "forward":
			if len(origins) > 1 {
				if o := p.origin(name); o != "" {
					t = o + ":\n" + t
				}
			}
			if t != "" {
				out = append(out, t)
			}
		case "", "text":
			if t != "" {
				out = append(out, t)
			}
		}
	}
	c.text = m.Text
	if c.text != "" {
		c.format = readFormat(m.Format)
	} else {
		c.text = strings.Join(out, "\n\n")
		// Offsets hold only when one part is the whole text.
		if len(out) == 1 {
			for _, p := range parts {
				if (p.MediaType == "" || p.MediaType == "text") && p.text() == c.text {
					c.format = readFormat(p.Format)
				}
			}
		}
	}
	// The text of a file message may be just the file's link.
	for _, f := range m.Filesharing {
		if f.OriginalURL != "" {
			if t := strings.TrimSpace(strings.ReplaceAll(c.text, f.OriginalURL, "")); t != c.text {
				c.text, c.format = t, nil
			}
		}
	}
	return c
}

// readFormat reads native styles in the Bot API's shape: {"bold":[{"offset":
// 0,"length":4}],"link":[{…,"url":…}],"pre":[{…,"code":…}]}. Unverified for
// this protocol: anything else reads as no styles.
func readFormat(raw json.RawMessage) []model.TextSpan {
	if len(raw) == 0 {
		return nil
	}
	var f map[string][]struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		URL    string `json:"url"`
		Code   string `json:"code"`
	}
	if json.Unmarshal(raw, &f) != nil {
		return nil
	}
	kinds := map[string]model.SpanKind{
		"bold": model.SpanBold, "italic": model.SpanItalic, "underline": model.SpanUnderline,
		"strikethrough": model.SpanStrike, "inline_code": model.SpanCode, "pre": model.SpanPre,
		"link": model.SpanLink, "quote": model.SpanQuote,
	}
	var out []model.TextSpan
	for name, list := range f {
		k, ok := kinds[name]
		if !ok {
			continue
		}
		for _, r := range list {
			if r.Length <= 0 || r.Offset < 0 || (k == model.SpanLink && !richtext.SafeURL(r.URL)) {
				continue
			}
			out = append(out, model.TextSpan{Kind: k, Offset: r.Offset, Length: r.Length, URL: r.URL, Lang: r.Code})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Offset != out[j].Offset {
			return out[i].Offset < out[j].Offset
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

// isService tells a chat event from a message.
func (m *wMessage) isService() bool {
	return m.Class == "event" || len(m.Chat.MemberEvent) > 0 || len(m.Chat.ModifiedInfo) > 0
}

func (m *wMessage) hasReactions() bool {
	var r struct {
		Exist bool `json:"exist"`
	}
	return json.Unmarshal(m.Reactions, &r) == nil && r.Exist
}

// serviceText describes a chat event, as the desktop client does from
// chat.memberEvent {type, members, requestedBy} and chat.modifiedInfo; ""
// for an event it does not know.
func (m *wMessage) serviceText(name func(string) string) string {
	actor := func(verb string) string {
		if m.Chat.Sender == "" {
			return verb
		}
		return name(m.Chat.Sender) + " " + verb
	}
	var ev struct {
		Type        string `json:"type"`
		Members     []any  `json:"members"`
		RequestedBy string `json:"requestedBy"`
	}
	if len(m.Chat.MemberEvent) > 0 && json.Unmarshal(m.Chat.MemberEvent, &ev) == nil {
		var names []string
		for _, v := range ev.Members {
			if sn, ok := v.(string); ok && sn != "" {
				names = append(names, name(sn))
			}
		}
		who := strings.Join(names, ", ")
		switch ev.Type {
		case "addMembers":
			return actor("added " + who)
		case "invite":
			return actor("invited " + who)
		case "leave":
			if who != "" {
				return who + " left the chat"
			}
			return actor("left the chat")
		case "delMembers", "kicked":
			return actor("removed " + who)
		case "admGranted":
			return actor("made " + who + " an admin")
		case "admRevoked":
			return actor("removed admin rights from " + who)
		case "allowedToWrite":
			return actor("allowed " + who + " to write")
		case "disallowedToWrite":
			return actor("disallowed " + who + " to write")
		case "waitingForApprove":
			if who != "" {
				return who + " asked to join"
			}
			return actor("asked to join")
		case "joiningApproved":
			return actor("approved " + who)
		case "joiningRejected":
			return actor("rejected " + who)
		case "waitingCanceled":
			if who != "" {
				return who + " canceled the join request"
			}
			return actor("canceled the join request")
		}
	}
	var mod map[string]json.RawMessage
	if len(m.Chat.ModifiedInfo) > 0 && json.Unmarshal(m.Chat.ModifiedInfo, &mod) == nil {
		if raw, ok := mod["name"]; ok {
			var title string
			if json.Unmarshal(raw, &title) == nil && title != "" {
				return actor("changed the title to «" + title + "»")
			}
		}
		for _, c := range []struct{ key, text string }{
			{"avatarLastModified", "changed the photo"},
			{"about", "changed the description"},
			{"rules", "changed the rules"},
			{"stamp", "changed the chat link"},
		} {
			if _, ok := mod[c.key]; ok {
				return actor(c.text)
			}
		}
		return actor("changed the chat settings")
	}
	return ""
}

// body is the message's text.
func (m *wMessage) body() string {
	return m.content(func(id string) string { return id }).text
}

// dlgState is the histDlgState event: a dialog's news.
type dlgState struct {
	SN        string      `json:"sn"`
	UnreadCnt int         `json:"unreadCnt"`
	LastMsgID json.Number `json:"lastMsgId"`
	Messages  []wMessage  `json:"messages"`
	Persons   []person    `json:"persons"`
	Yours     struct {
		LastRead json.Number `json:"lastRead"`
	} `json:"yours"`
	DelUpTo json.Number `json:"delUpto"`
	// ParentTopic: this dialog is the thread of that message.
	ParentTopic *wTopic `json:"parentTopic"`
}

// wTopic is what a thread hangs under; only "message" topics are threads of
// a chat ("task" ones belong to tasks).
type wTopic struct {
	Type      string      `json:"type"`
	ChatID    string      `json:"chatId"`
	MessageID json.Number `json:"messageId"`
}

func (t *wTopic) valid() bool {
	return t != nil && t.Type == "message" && t.ChatID != "" && t.MessageID != "" && t.MessageID != "0"
}

// wThread is a thread in a message, thread/feed/get and threadUpdate.
type wThread struct {
	ThreadID     string  `json:"threadId"`
	ParentTopic  *wTopic `json:"parentTopic"`
	RepliesCount wInt    `json:"repliesCount"`
}

type typingEvent struct {
	AimID        string `json:"aimId"`
	ChatID       string `json:"chatId"`
	TypingStatus string `json:"typingStatus"` // "typing", "typed", "none"
	Sender       string `json:"sender"`
}

// --- RAPI (JSON requests, {"status":{code},"results"} answers) ------------

// rapiOK is the success status; 40100-40199 mean the aimsid is dead.
const rapiOK = 20000

func (s *server) rapiCall(ctx context.Context, aimsid, method string, params any, out any) error {
	body, err := json.Marshal(map[string]any{"reqId": s.reqID(), "aimsid": aimsid, "params": params})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.rapi()+method, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return noURL(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vk teams: %s: HTTP %d", method, resp.StatusCode)
	}
	var r struct {
		Status struct {
			Code   int    `json:"code"`
			Reason string `json:"reason"`
		} `json:"status"`
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("vk teams: %s: %w", method, err)
	}
	switch c := r.Status.Code; {
	case c == rapiOK:
	case c >= 40100 && c < 40200:
		return &errSession{detail: fmt.Sprintf("%s: %d %s", method, c, r.Status.Reason)}
	default:
		return fmt.Errorf("vk teams: %s: %d %s", method, c, r.Status.Reason)
	}
	if out != nil && len(r.Results) > 0 {
		return json.Unmarshal(r.Results, out)
	}
	return nil
}

type historyResult struct {
	Messages    []wMessage `json:"messages"`
	Persons     []person   `json:"persons"`
	ParentTopic *wTopic    `json:"parentTopic"` // the dialog is a thread
}

type threadsFeed struct {
	Threads []wThread `json:"threads"`
	Cursor  string    `json:"cursor"`
	Persons []person  `json:"persons"`
}

func (s *server) threadsFeed(ctx context.Context, aimsid, cursor string) (threadsFeed, error) {
	var r threadsFeed
	err := s.rapiCall(ctx, aimsid, "thread/feed/get", map[string]any{
		"pageSize": 10, "cursor": cursor, "filter": map[string]any{"type": "all"},
	}, &r)
	return r, err
}

// threadAdd makes the thread of a message and returns its id.
func (s *server) threadAdd(ctx context.Context, aimsid, chat, msgID string) (string, error) {
	var r struct {
		ThreadID string `json:"threadId"`
	}
	err := s.rapiCall(ctx, aimsid, "thread/add", map[string]any{
		"parentTopic": map[string]any{"type": "message", "chatId": chat, "messageId": json.Number(msgID)},
	}, &r)
	if err == nil && r.ThreadID == "" {
		err = errors.New("vk teams: thread/add: no thread id")
	}
	return r.ThreadID, err
}

// getHistory returns up to n messages older than from ("-1": the latest),
// as the server orders them.
func (s *server) getHistory(ctx context.Context, aimsid, sn, from string, n int) (historyResult, error) {
	var r historyResult
	err := s.rapiCall(ctx, aimsid, "getHistory", map[string]any{
		"sn": sn, "fromMsgId": from, "count": -n, "patchVersion": "init",
	}, &r)
	return r, err
}

// sendMessage sends a text with RAPI message/send (what the desktop and web
// clients use); a reply quotes the message it answers.
func (s *server) sendMessage(ctx context.Context, aimsid, to, text, replyTo string) (string, error) {
	params := map[string]any{"target": to, "parts": map[string]any{"mainPart": map[string]any{"text": map[string]any{"plain": text}}}}
	if replyTo != "" {
		params["quoteParts"] = []any{map[string]any{"sn": to, "msgId": json.Number(replyTo), "order": 0}}
	}
	var r struct {
		MsgID json.Number `json:"msgId"`
	}
	if err := s.rapiCall(ctx, aimsid, "message/send", params, &r); err != nil {
		return "", err
	}
	return r.MsgID.String(), nil
}

type contactsResult struct {
	Contacts []struct {
		SN string `json:"sn"`
	} `json:"contacts"`
	Persons []person `json:"persons"`
	// Shape lists the answer's keys and their JSON kinds (no values), for
	// the log while this call is unverified.
	Shape string `json:"-"`
}

// shapeOf describes raw JSON by its keys and value kinds, down to depth 2;
// it never includes values.
func shapeOf(raw json.RawMessage, depth int) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return "empty"
	}
	switch raw[0] {
	case '{':
		var m map[string]json.RawMessage
		if json.Unmarshal(raw, &m) != nil {
			return "object?"
		}
		if depth == 0 {
			return fmt.Sprintf("object(%d keys)", len(m))
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + ":" + shapeOf(m[k], depth-1)
		}
		return "{" + strings.Join(parts, ",") + "}"
	case '[':
		var a []json.RawMessage
		if json.Unmarshal(raw, &a) != nil {
			return "array?"
		}
		if len(a) == 0 {
			return "[]"
		}
		return fmt.Sprintf("[%d×%s]", len(a), shapeOf(a[0], depth-1))
	case '"':
		return "string"
	case 't', 'f':
		return "bool"
	case 'n':
		return "null"
	}
	return "number"
}

// contactsGet lists the account's contacts and chats (the web client's
// "contacts/get"; a resumed session gets no buddylist event to learn them).
func (s *server) contactsGet(ctx context.Context, aimsid string) (contactsResult, error) {
	var raw json.RawMessage
	if err := s.rapiCall(ctx, aimsid, "contacts/get", map[string]any{}, &raw); err != nil {
		return contactsResult{}, err
	}
	var r contactsResult
	err := json.Unmarshal(raw, &r)
	r.Shape = shapeOf(raw, 3)
	return r, err
}

// chatName is a group chat's name (getChatInfo).
func (s *server) chatName(ctx context.Context, aimsid, sn string) (string, error) {
	var r struct {
		Name string `json:"name"`
	}
	err := s.rapiCall(ctx, aimsid, "getChatInfo", map[string]any{"sn": sn}, &r)
	return r.Name, err
}

// getUserInfo describes a user (rapi getUserInfo {sn}); the answer has no
// sn of its own, so it is set here.
func (s *server) getUserInfo(ctx context.Context, aimsid, sn string) (person, error) {
	var p person
	err := s.rapiCall(ctx, aimsid, "getUserInfo", map[string]any{"sn": sn}, &p)
	if p.SN == "" && p.AimID == "" {
		p.SN = sn
	}
	return p, err
}

// chatProfile is what getChatInfo tells of a group beyond its name (field
// names follow ICQ New's getChatInfo; unverified).
type chatProfile struct {
	Name         string   `json:"name"`
	About        string   `json:"about"`
	Rules        string   `json:"rules"`
	MembersCount wInt     `json:"membersCount"`
	Members      []person `json:"members"`
}

func (s *server) chatProfile(ctx context.Context, aimsid, sn string) (chatProfile, error) {
	var r chatProfile
	err := s.rapiCall(ctx, aimsid, "getChatInfo", map[string]any{"sn": sn, "memberLimit": 50}, &r)
	return r, err
}

// userProfile is getUserInfo's answer with what a profile shows (about,
// phone, email: unverified).
type userProfile struct {
	person
	About string `json:"about"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Title string `json:"position"`
}

func (s *server) userProfile(ctx context.Context, aimsid, sn string) (userProfile, error) {
	var p userProfile
	err := s.rapiCall(ctx, aimsid, "getUserInfo", map[string]any{"sn": sn}, &p)
	if p.SN == "" && p.AimID == "" {
		p.SN = sn
	}
	return p, err
}

// avatarHost is where avatars live: the desktop client asks "ub.<domain>"
// for a server "u.<domain>", the same host otherwise.
func (s *server) avatarHost() string {
	if strings.HasPrefix(s.host, "u.") {
		return "ub." + s.host[2:]
	}
	return s.host
}

// avatarSize is the avatar edge in pixels asked for.
const avatarSize = 128

// errNoAvatar: the user or chat has no photo (the server says 404).
var errNoAvatar = errors.New("vk teams: no photo")

// avatar downloads sn's photo.
func (s *server) avatar(ctx context.Context, aimsid, sn string, limit int64) ([]byte, string, error) {
	q := url.Values{"aimsid": {aimsid}, "targetSn": {sn}, "size": {strconv.Itoa(avatarSize)}}
	u := s.scheme + "://" + s.avatarHost() + "/api/" + s.version + "/files/avatar/get?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, "", noURL(err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, "", errNoAvatar
	case resp.StatusCode != http.StatusOK:
		return nil, "", fmt.Errorf("vk teams: photo: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, "", errTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		return nil, "", errTooLarge
	}
	mime := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(mime, "image/") {
		mime = http.DetectContentType(data)
		if !strings.HasPrefix(mime, "image/") {
			return nil, "", errNoAvatar // an error page, not a photo
		}
	}
	return data, mime, nil
}

// buddyList is the buddylist event: the contact list in groups.
type buddyList struct {
	Groups []struct {
		Buddies []person `json:"buddies"`
	} `json:"groups"`
}

// setDlgState marks a dialog read up to msgID.
func (s *server) setDlgState(ctx context.Context, aimsid, sn, msgID string) error {
	return s.rapiCall(ctx, aimsid, "setDlgState", map[string]any{"sn": sn, "lastRead": msgID}, nil)
}

// endSession closes the session on the server.
func (s *server) endSession(ctx context.Context, aimsid string) error {
	return s.wimDo(ctx, http.MethodPost, s.wim()+"aim/endSession", url.Values{"aimsid": {aimsid}, "f": {"json"}}, nil)
}

// mentions tells whether message text names user self. The clients write
// a mention as "@[aimId]" (UNVERIFIED; seen in exported chats).
func mentions(text, self string) bool {
	return self != "" && strings.Contains(text, "@["+self+"]")
}

// --- reactions --------------------------------------------------------------
//
// RAPI reaction/add, reaction/remove, reaction/get and the "reactions" event,
// as in the desktop client. An account has one reaction on a message.

// defaultReactions is the desktop client's default set: reaction/add takes
// the list of reactions to count along with the one added.
var defaultReactions = []string{"👍", "👎", "❤️", "🤬", "😮", "🤣"}

// wReactions is a message's reactions in answers and events.
type wReactions struct {
	MsgID     json.Number `json:"msgId"`
	ChatID    string      `json:"chatId"`
	My        string      `json:"myReaction"`
	Reactions []struct {
		Reaction string `json:"reaction"`
		Counter  wInt   `json:"counter"`
	} `json:"reactions"`
}

func (r wReactions) toModel() []model.Reaction {
	out := make([]model.Reaction, 0, len(r.Reactions))
	for _, x := range r.Reactions {
		if x.Reaction != "" && x.Counter > 0 {
			out = append(out, model.Reaction{Emoji: x.Reaction, Count: int(x.Counter), Mine: x.Reaction == r.My})
		}
	}
	return out
}

func (s *server) addReaction(ctx context.Context, aimsid, chat string, msgID int64, emoji string) (wReactions, error) {
	list := append([]string(nil), defaultReactions...)
	if !slices.Contains(list, emoji) {
		list = append(list, emoji)
	}
	var r wReactions
	err := s.rapiCall(ctx, aimsid, "reaction/add", map[string]any{
		"msgId": msgID, "chatId": chat, "reaction": emoji, "reactions": list,
	}, &r)
	return r, err
}

func (s *server) removeReaction(ctx context.Context, aimsid, chat string, msgID int64) error {
	return s.rapiCall(ctx, aimsid, "reaction/remove", map[string]any{"msgId": msgID, "chatId": chat}, nil)
}

func (s *server) getReactions(ctx context.Context, aimsid, chat string, ids []int64) ([]wReactions, error) {
	var r struct {
		Reactions []wReactions `json:"reactions"`
	}
	err := s.rapiCall(ctx, aimsid, "reaction/get", map[string]any{"chatId": chat, "msgIds": ids}, &r)
	return r.Reactions, err
}

// --- shared files ---------------------------------------------------------
//
// The desktop client (25.4) asks <host>/files/info/<id>/?f=json&aimsid=… for
// a file's metadata. The answer's layout is UNVERIFIED: the download link is
// taken from its "dlink" key wherever it is.

const fileInfoLimit = 1 << 20

// fileLink returns the download link of a shared file.
func (s *server) fileLink(ctx context.Context, aimsid, id string) (string, error) {
	q := url.Values{"f": {"json"}, "aimsid": {aimsid}}
	u := s.scheme + "://" + s.avatarHost() + "/api/" + s.version + "/files/info/" + url.PathEscape(id) + "/?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("vk teams: file info: bad request")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return "", noURL(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vk teams: file info: HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, fileInfoLimit))
	if err != nil {
		return "", noURL(err)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", fmt.Errorf("vk teams: file info: not JSON")
	}
	link := findString(v, "dlink")
	if link == "" {
		return "", fmt.Errorf("vk teams: file info: no link in answer %s", shapeOf(raw, 4))
	}
	return link, nil
}

// findString returns the first non-empty string under key, at any depth.
func findString(v any, key string) string {
	switch t := v.(type) {
	case map[string]any:
		if s, ok := t[key].(string); ok && s != "" {
			return s
		}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s := findString(t[k], key); s != "" {
				return s
			}
		}
	case []any:
		for _, e := range t {
			if s := findString(e, key); s != "" {
				return s
			}
		}
	}
	return ""
}

// sameDomain tells that host belongs to the messenger's own domain.
func (s *server) sameDomain(host string) bool {
	own := strings.ToLower((&url.URL{Host: s.host}).Hostname())
	host = strings.ToLower(host)
	if host == own {
		return true
	}
	labels := strings.Split(own, ".")
	if len(labels) < 2 || net.ParseIP(own) != nil {
		return false
	}
	return strings.HasSuffix(host, "."+strings.Join(labels[len(labels)-2:], "."))
}

// fileGet downloads a file from its link, without the session's credentials
// and only from the messenger's own domain.
func (s *server) fileGet(ctx context.Context, link string, limit int64) ([]byte, string, error) {
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || !s.sameDomain(u.Hostname()) {
		return nil, "", fmt.Errorf("vk teams: file link outside the messenger's domain")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, "", fmt.Errorf("vk teams: file: bad link")
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, "", noURL(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("vk teams: file: HTTP %d", resp.StatusCode)
	}
	if !s.sameDomain(resp.Request.URL.Hostname()) { // after redirects
		return nil, "", fmt.Errorf("vk teams: file link outside the messenger's domain")
	}
	if resp.ContentLength > limit {
		return nil, "", errTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", noURL(err)
	}
	if int64(len(data)) > limit {
		return nil, "", errTooLarge
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	return data, mime, nil
}

// --- sending files, editing, deleting, forwarding --------------------------
//
// From the open im-desktop client (core/connections/wim/packets: get_gateway,
// send_file, send_message, del_message_batch); UNVERIFIED on v132. A file is
// uploaded in ranges and then sent as a message holding its link.

// uploadBlock is the size of one uploaded range, as in the desktop client.
const uploadBlock = 1 << 20

// uploadFile uploads a file and returns its link. The session id goes only
// to an upload host of the messenger's own domain, and never over a redirect.
func (s *server) uploadFile(ctx context.Context, aimsid, devID, name string, data []byte) (string, error) {
	q := url.Values{"aimsid": {aimsid}, "f": {"json"}, "size": {strconv.Itoa(len(data))},
		"filename": {name}, "client": {"parley"}, "r": {s.reqID()}}
	if devID != "" {
		q.Set("k", devID)
	}
	hc := *s.http
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.base()+"files/init?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("vk teams: upload: bad request")
	}
	var gate struct {
		Status struct {
			Code int `json:"code"`
		} `json:"status"`
		Result struct {
			Host string `json:"host"`
			URL  string `json:"url"`
		} `json:"result"`
	}
	if err := uploadAnswer(&hc, req, &gate); err != nil {
		return "", err
	}
	if gate.Status.Code == http.StatusRequestEntityTooLarge {
		return "", errTooLarge
	}
	if gate.Status.Code != http.StatusOK {
		return "", fmt.Errorf("vk teams: upload: status %d", gate.Status.Code)
	}
	target, err := url.Parse(s.scheme + "://" + gate.Result.Host + gate.Result.URL)
	if err != nil || target.User != nil || target.Host != gate.Result.Host ||
		!strings.HasPrefix(gate.Result.URL, "/") || !s.sameDomain(target.Hostname()) {
		return "", fmt.Errorf("vk teams: upload host outside the messenger's domain")
	}
	session := strconv.FormatInt(time.Now().Unix(), 10)
	safeName := strings.NewReplacer(`"`, "_", `\`, "_").Replace(name)
	for from := 0; from < len(data); from += uploadBlock {
		to := min(from+uploadBlock, len(data))
		cq := url.Values{"aimsid": {aimsid}, "f": {"json"}, "r": {s.reqID()}}
		if devID != "" {
			cq.Set("k", devID)
		}
		u := *target
		u.RawQuery = cq.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data[from:to]))
		if err != nil {
			return "", fmt.Errorf("vk teams: upload: bad request")
		}
		req.Header.Set("Session-ID", session)
		req.Header.Set("Content-Disposition", `attachment; filename="`+safeName+`"`)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", from, to-1, len(data)))
		var done struct {
			Status struct {
				Code int `json:"code"`
			} `json:"status"`
			Result struct {
				StaticURL string `json:"static_url"`
			} `json:"result"`
		}
		if err := uploadAnswer(&hc, req, &done); err != nil {
			return "", err
		}
		if to == len(data) {
			if done.Status.Code != http.StatusOK || done.Result.StaticURL == "" {
				return "", fmt.Errorf("vk teams: upload: status %d, no link", done.Status.Code)
			}
			return done.Result.StaticURL, nil
		}
	}
	return "", errors.New("vk teams: upload: an empty file")
}

// uploadAnswer does an upload request and reads its JSON answer; a range
// still in progress may answer with no body.
func uploadAnswer(hc *http.Client, req *http.Request, out any) error {
	resp, err := hc.Do(req)
	if err != nil {
		return noURL(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
	case http.StatusRequestEntityTooLarge:
		return errTooLarge
	default:
		return fmt.Errorf("vk teams: upload: HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, fileInfoLimit))
	if err != nil {
		return noURL(err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("vk teams: upload: not JSON")
	}
	return nil
}

// sendIM is WIM im/sendIM: a plain message, or parts when they are given;
// updateID makes it an edit of that message.
func (s *server) sendIM(ctx context.Context, aimsid, to, text string, parts []any, updateID string) (string, error) {
	form := url.Values{"f": {"json"}, "aimsid": {aimsid}, "t": {to}, "r": {s.reqID()},
		"offlineIM": {"1"}, "notifyDelivery": {"true"}}
	if updateID != "" {
		form.Set("updateMsgId", updateID)
	}
	if len(parts) > 0 {
		b, err := json.Marshal(parts)
		if err != nil {
			return "", err
		}
		form.Set("parts", string(b))
	} else {
		form.Set("message", text)
	}
	var r struct {
		HistMsgID json.Number `json:"histMsgId"`
	}
	if err := s.wimDo(ctx, http.MethodPost, s.wim()+"im/sendIM", form, &r); err != nil {
		return "", err
	}
	return r.HistMsgID.String(), nil
}

// captionPart is a file's link with a caption.
func captionPart(caption, link string) map[string]any {
	return map[string]any{"mediaType": "text", "text": "",
		"captionedContent": map[string]any{"caption": caption, "url": link}}
}

// editParts is what an edit of m to text sends: nil for a plain message,
// else the parts that keep its quotes and its file. A message that cannot be
// rewritten without losing something is refused.
func editParts(m wMessage, text string) ([]any, error) {
	var kept []any
	var list []map[string]json.RawMessage
	if json.Unmarshal(m.Parts, &list) == nil {
		for _, p := range list {
			var kind string
			_ = json.Unmarshal(p["mediaType"], &kind)
			switch kind {
			case "quote":
				kept = append(kept, p)
			case "forward":
				return nil, errors.New("vk teams: a forwarded message cannot be edited")
			}
		}
	}
	switch len(m.Filesharing) {
	case 0:
		if len(kept) == 0 {
			return nil, nil
		}
		return append(kept, map[string]any{"mediaType": "text", "text": text}), nil
	case 1:
		link := m.Filesharing[0].OriginalURL
		if link == "" {
			return nil, errors.New("vk teams: the file's link is unknown, the caption cannot be edited")
		}
		return append(kept, captionPart(text, link)), nil
	}
	return nil, errors.New("vk teams: a message with several files cannot be edited")
}

// quotePart quotes a fragment of m, a message of chat by sender.
func quotePart(chat, sender string, m wMessage, fragment string) map[string]any {
	p := map[string]any{"mediaType": "quote", "text": fragment, "sn": sender, "msgId": m.MsgID, "time": m.Time}
	if isGroup(chat) {
		p["chat"] = map[string]any{"sn": chat}
	}
	return p
}

// forwardParts is m as "forward" parts; sender is its author's id.
func forwardParts(chat, sender string, m wMessage, text string) []any {
	part := func() map[string]any {
		p := map[string]any{"mediaType": "forward", "sn": sender, "msgId": m.MsgID, "time": m.Time}
		if isGroup(chat) {
			p["chat"] = map[string]any{"sn": chat}
		}
		return p
	}
	var out []any
	for _, f := range m.Filesharing {
		if f.OriginalURL == "" {
			continue
		}
		p := part()
		if text != "" {
			p["text"] = ""
			p["captionedContent"] = map[string]any{"caption": text, "url": f.OriginalURL}
			text = ""
		} else {
			p["text"] = f.OriginalURL
		}
		out = append(out, p)
	}
	if text != "" {
		p := part()
		p["text"] = text
		out = append(out, p)
	}
	return out
}

// getMessage reads one message. The history starts just above it, because
// the server may or may not include the id it starts from.
func (s *server) getMessage(ctx context.Context, aimsid, sn, id string) (wMessage, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return wMessage{}, fmt.Errorf("vk teams: bad message id")
	}
	r, err := s.getHistory(ctx, aimsid, sn, strconv.FormatInt(n+1, 10), 2)
	if err != nil {
		return wMessage{}, err
	}
	for _, m := range r.Messages {
		if m.MsgID.String() == id {
			return m, nil
		}
	}
	return wMessage{}, errors.New("vk teams: message not found")
}

// deleteMessages is RAPI delMsgBatch; shared deletes for everyone.
func (s *server) deleteMessages(ctx context.Context, aimsid, sn string, ids []int64, shared bool) error {
	return s.rapiCall(ctx, aimsid, "delMsgBatch", map[string]any{"sn": sn, "msgIds": ids, "shared": shared}, nil)
}

// --- the account's settings of a chat -------------------------------------
//
// UNVERIFIED: the WIM methods of ICQ New's desktop client (buddylist/mute,
// buddylist/hideChat, buddylist/removeBuddy) and their parameters; VK Teams
// descends from it. The "hiddenChat" event ({"sn", "lastMsgId"}) tells a
// chat was archived (hidden) on any device; a new message brings it back.

// muteForever is the timeout (seconds) a mute "forever" sends.
const muteForever = 0x7fffffff

// muteChat mutes a chat (or unmutes it: on false).
func (s *server) muteChat(ctx context.Context, aimsid, sn string, on bool) error {
	timeout := "0"
	if on {
		timeout = strconv.Itoa(muteForever)
	}
	return s.wimDo(ctx, http.MethodPost, s.wim()+"buddylist/mute",
		url.Values{"aimsid": {aimsid}, "f": {"json"}, "buddy": {sn}, "timeout": {timeout}}, nil)
}

// hideChat archives a chat up to its latest message.
func (s *server) hideChat(ctx context.Context, aimsid, sn, lastMsgID string) error {
	form := url.Values{"aimsid": {aimsid}, "f": {"json"}, "buddy": {sn}}
	if lastMsgID != "" {
		form.Set("lastMsgId", lastMsgID)
	}
	return s.wimDo(ctx, http.MethodPost, s.wim()+"buddylist/hideChat", form, nil)
}

// leaveChat leaves a group chat (removes it from every buddylist group).
func (s *server) leaveChat(ctx context.Context, aimsid, sn string) error {
	return s.wimDo(ctx, http.MethodPost, s.wim()+"buddylist/removeBuddy",
		url.Values{"aimsid": {aimsid}, "f": {"json"}, "buddy": {sn}, "allGroups": {"1"}}, nil)
}

// hiddenChat is the data of a "hiddenChat" event.
type hiddenChat struct {
	SN        string      `json:"sn"`
	LastMsgID json.Number `json:"lastMsgId"`
}
