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
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	// What the VK WorkSpace desktop client (2026) talks to.
	defaultHost    = "u.myteam.vmailru.net"
	defaultVersion = "v132"
	fetchTimeout   = 25 * time.Second
	// startEvents are the event kinds a session subscribes to.
	startEvents = "myInfo,presence,buddylist,typing,hiddenChat,hist,mchat,sentIM,imState,dataIM,offlineIM,userAddedToBuddyList,service,lifestream,apps,permitDeny,diff,webrtcMsg"
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
		"clientName": {"MegaChat"}, "idType": {"ICQ"}, "tokenType": {"longTerm"},
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
		"clientName": {"MegaChat"}, "imf": {"plain"}, "invisible": {"false"}, "view": {"online"},
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

func (p person) name() string {
	full := strings.TrimSpace(p.FirstName + " " + p.LastName)
	for _, n := range []string{p.Friendly, full, p.Nick, p.id()} {
		if n != "" {
			return n
		}
	}
	return ""
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
	MsgID    json.Number `json:"msgId"`
	Time     int64       `json:"time"`
	Text     string      `json:"text"`
	Outgoing bool        `json:"outgoing"`
	Chat     struct {
		Sender string `json:"sender"`
		Name   string `json:"name"`
	} `json:"chat"`
	Class       string `json:"class"` // "event" for service messages
	UpdatePatch []struct {
		Type string `json:"type"` // "delete", "edit"
	} `json:"updatePatch"`
	Filesharing []struct {
		Name     string `json:"name"`
		Mime     string `json:"mime"`
		Size     wInt   `json:"size"`
		ID       string `json:"id"`
		Duration wInt   `json:"duration"`
	} `json:"filesharing"`
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
	Messages []wMessage `json:"messages"`
	Persons  []person   `json:"persons"`
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
