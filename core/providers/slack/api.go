package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"megachat/core/provider"
)

// api calls Slack's Web API. A browser session (xoxc token) must go to the
// workspace's own host with the "d" cookie; an OAuth token goes to
// slack.com.
type api struct {
	base   string // ".../api/"
	token  string
	cookie string
	http   *http.Client
}

// apiError is a Slack "ok": false answer.
type apiError struct {
	Method string
	Code   string
}

func (e *apiError) Error() string { return "slack: " + e.Method + ": " + e.Code }

var authCodes = map[string]bool{
	"invalid_auth": true, "not_authed": true, "token_revoked": true, "token_expired": true,
	"account_inactive": true, "no_permission": false, "missing_scope": false,
}

func isAuth(err error) bool {
	e, ok := err.(*apiError)
	return ok && authCodes[e.Code]
}

func (a *api) call(ctx context.Context, method string, params url.Values, out any) error {
	if params == nil {
		params = url.Values{}
	}
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.base+method, strings.NewReader(params.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if a.token != "" {
			req.Header.Set("Authorization", "Bearer "+a.token)
		}
		if a.cookie != "" {
			req.Header.Set("Cookie", "d="+url.QueryEscape(a.cookie))
		}
		resp, err := a.http.Do(req)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			wait, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			select {
			case <-time.After(time.Duration(max(wait, 1)) * time.Second):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("slack: %s: HTTP %d", method, resp.StatusCode)
		}
		var head struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &head); err != nil {
			return fmt.Errorf("slack: %s: %w", method, err)
		}
		if !head.OK {
			return &apiError{Method: method, Code: head.Error}
		}
		if out != nil {
			return json.Unmarshal(body, out)
		}
		return nil
	}
}

// wrapAuth turns Slack's auth failures into provider.ErrAuth.
func wrapAuth(err error) error {
	if isAuth(err) {
		return fmt.Errorf("%w (%v)", provider.ErrAuth, err)
	}
	return err
}

// workspaceURL turns "acme", "acme.slack.com" or a full URL into
// "https://acme.slack.com/".
func workspaceURL(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("enter the workspace")
	}
	if !strings.Contains(s, "://") {
		if !strings.Contains(s, ".") {
			s += ".slack.com"
		}
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("not a workspace address: %q", s)
	}
	return "https://" + u.Host + "/", nil
}

// --- wire types ----------------------------------------------------------

type wUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Deleted  bool   `json:"deleted"`
	Profile  struct {
		DisplayName string `json:"display_name"`
		RealName    string `json:"real_name"`
		Image72     string `json:"image_72"`
	} `json:"profile"`
}

func (u wUser) display() string {
	for _, n := range []string{u.Profile.DisplayName, u.Profile.RealName, u.RealName, u.Name} {
		if n != "" {
			return n
		}
	}
	return u.ID
}

type wFile struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Mimetype string `json:"mimetype"`
	Size     int64  `json:"size"`
	Subtype  string `json:"subtype"`
	ID       string `json:"id"`

	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`
	Thumb360           string `json:"thumb_360"`
	Thumb480           string `json:"thumb_480"`
	ThumbVideo         string `json:"thumb_video"`
	OriginalW          int    `json:"original_w"`
	OriginalH          int    `json:"original_h"`
	Thumb360W          int    `json:"thumb_360_w"`
	Thumb360H          int    `json:"thumb_360_h"`
	DurationMS         int64  `json:"duration_ms"`
}

type wReaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

type wMessage struct {
	Type     string  `json:"type"`
	Subtype  string  `json:"subtype"`
	TS       string  `json:"ts"`
	ThreadTS string  `json:"thread_ts"`
	User     string  `json:"user"`
	BotID    string  `json:"bot_id"`
	Username string  `json:"username"`
	Text     string  `json:"text"`
	Files    []wFile `json:"files"`
	Edited   *struct {
		TS string `json:"ts"`
	} `json:"edited"`
	Hidden bool `json:"hidden"`

	ReplyCount  int           `json:"reply_count"`
	ReplyUsers  []string      `json:"reply_users"`
	LatestRepl  string        `json:"latest_reply"`
	Reactions   []wReaction   `json:"reactions"`
	Attachments []wAttachment `json:"attachments"`
}

// wAttachment is a legacy attachment; only shared messages are used.
type wAttachment struct {
	IsShare       bool    `json:"is_share"`
	IsMsgUnfurl   bool    `json:"is_msg_unfurl"`
	AuthorName    string  `json:"author_name"`
	AuthorSubname string  `json:"author_subname"`
	Text          string  `json:"text"`
	FromURL       string  `json:"from_url"`
	OriginalURL   string  `json:"original_url"`
	Files         []wFile `json:"files"`
}

type wChannel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsIM       bool   `json:"is_im"`
	IsMPIM     bool   `json:"is_mpim"`
	IsPrivate  bool   `json:"is_private"`
	IsArchived bool   `json:"is_archived"`
	IsReadOnly bool   `json:"is_read_only"`
	User       string `json:"user"` // the other side of an IM
	Purpose    struct {
		Value string `json:"value"`
	} `json:"purpose"`
}

// tsTime parses a Slack timestamp ("1700000000.000100").
func tsTime(ts string) time.Time {
	sec, frac, _ := strings.Cut(ts, ".")
	s, err := strconv.ParseInt(sec, 10, 64)
	if err != nil {
		return time.Time{}
	}
	frac = (frac + "000000")[:6]
	us, _ := strconv.ParseInt(frac, 10, 64)
	return time.Unix(s, us*1000)
}

var (
	reUser    = regexp.MustCompile(`<@([UW][A-Z0-9]+)(?:\|([^>]*))?>`)
	reChannel = regexp.MustCompile(`<#([CG][A-Z0-9]+)(?:\|([^>]*))?>`)
	reSpecial = regexp.MustCompile(`<!(here|channel|everyone)(?:\|[^>]*)?>`)
	reLink    = regexp.MustCompile(`<((?:https?|mailto|tel):[^|>]+)(?:\|([^>]*))?>`)
)

// mentions tells whether raw Slack text names user self or the whole
// channel (@here, @channel, @everyone).
func mentions(text, self string) bool {
	if reSpecial.MatchString(text) {
		return true
	}
	if self == "" {
		return false
	}
	for _, sm := range reUser.FindAllStringSubmatch(text, -1) {
		if sm[1] == self {
			return true
		}
	}
	return false
}

// plainText turns Slack's markup into readable text; name resolves user ids.
func plainText(s string, name func(id string) string) string {
	s = reUser.ReplaceAllStringFunc(s, func(m string) string {
		sm := reUser.FindStringSubmatch(m)
		if sm[2] != "" {
			return "@" + sm[2]
		}
		return "@" + name(sm[1])
	})
	s = reChannel.ReplaceAllStringFunc(s, func(m string) string {
		sm := reChannel.FindStringSubmatch(m)
		if sm[2] != "" {
			return "#" + sm[2]
		}
		return "#" + sm[1]
	})
	s = reSpecial.ReplaceAllString(s, "@$1")
	s = reLink.ReplaceAllStringFunc(s, func(m string) string {
		sm := reLink.FindStringSubmatch(m)
		if sm[2] != "" && sm[2] != sm[1] {
			return sm[2] + " (" + sm[1] + ")"
		}
		return strings.TrimPrefix(sm[1], "mailto:")
	})
	return html.UnescapeString(s)
}

// escapeText escapes what Slack treats as markup in outgoing text.
func escapeText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// mpimTitle makes "mpdm-ann--bob--carl-1" readable.
func mpimTitle(name string) string {
	n := strings.TrimPrefix(name, "mpdm-")
	if i := strings.LastIndex(n, "-"); i > 0 {
		if _, err := strconv.Atoi(n[i+1:]); err == nil {
			n = n[:i]
		}
	}
	return strings.Join(strings.Split(n, "--"), ", ")
}
