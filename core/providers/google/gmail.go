package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/providers/mailutil"
)

const gmail = "gmail/v1/users/me/"

// archiveID is the pseudo folder of all mail (no label).
const archiveID = "ALL"

var systemFolders = []struct {
	id, name string
	role     model.FolderRole
}{
	{"INBOX", "Inbox", model.FolderInbox},
	{"STARRED", "Starred", model.FolderOther},
	{"SENT", "Sent", model.FolderSent},
	{"DRAFT", "Drafts", model.FolderDrafts},
	{archiveID, "All mail", model.FolderArchive},
	{"SPAM", "Spam", model.FolderSpam},
	{"TRASH", "Trash", model.FolderTrash},
}

type label struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	MessagesTotal  int    `json:"messagesTotal"`
	MessagesUnread int    `json:"messagesUnread"`
}

// parallel runs fn for every index, a few at a time, and returns the
// first error.
func parallel(n int, fn func(i int) error) error {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
		sem   = make(chan struct{}, 8)
	)
	for i := range n {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := fn(i); err != nil {
				mu.Lock()
				if first == nil {
					first = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return first
}

func (s *session) Folders(ctx context.Context) ([]model.Folder, error) {
	var ls struct {
		Labels []label `json:"labels"`
	}
	if err := s.c.do(ctx, http.MethodGet, gmail+"labels", nil, &ls); err != nil {
		return nil, err
	}
	var out []model.Folder
	for _, f := range systemFolders {
		out = append(out, model.Folder{ID: f.id, Name: f.name, Role: f.role})
	}
	var user []model.Folder
	for _, l := range ls.Labels {
		if l.Type == "user" {
			user = append(user, model.Folder{ID: l.ID, Name: l.Name, Role: model.FolderOther})
		}
	}
	sort.Slice(user, func(i, j int) bool { return strings.ToLower(user[i].Name) < strings.ToLower(user[j].Name) })
	out = append(out, user...)
	err := parallel(len(out), func(i int) error {
		if out[i].ID == archiveID {
			return nil
		}
		var l label
		if err := s.c.do(ctx, http.MethodGet, gmail+"labels/"+url.PathEscape(out[i].ID), nil, &l); err != nil {
			return err
		}
		out[i].Unread, out[i].Total = l.MessagesUnread, l.MessagesTotal
		return nil
	})
	return out, err
}

type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type part struct {
	PartID   string   `json:"partId"`
	MimeType string   `json:"mimeType"`
	Filename string   `json:"filename"`
	Headers  []header `json:"headers"`
	Body     struct {
		AttachmentID string `json:"attachmentId"`
		Size         int64  `json:"size"`
		Data         string `json:"data"`
	} `json:"body"`
	Parts []part `json:"parts"`
}

type message struct {
	ID           string   `json:"id"`
	ThreadID     string   `json:"threadId"`
	LabelIDs     []string `json:"labelIds"`
	Snippet      string   `json:"snippet"`
	InternalDate string   `json:"internalDate"`
	Payload      part     `json:"payload"`
}

func (p part) header(name string) string {
	for _, h := range p.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func (m message) time() time.Time {
	ms, _ := strconv.ParseInt(m.InternalDate, 10, 64)
	return time.UnixMilli(ms)
}

// summary maps a message fetched with format=metadata (or full).
func (m message) summary() model.Mail {
	h := m.Payload
	out := model.Mail{
		ID: m.ID, ThreadID: m.ThreadID,
		From: mailutil.Address(h.header("From")), To: mailutil.Addresses(h.header("To")),
		Cc: mailutil.Addresses(h.header("Cc")), ReplyTo: mailutil.Addresses(h.header("Reply-To")),
		Subject: mailutil.DecodeHeader(h.header("Subject")), Snippet: html.UnescapeString(m.Snippet),
		Time:   m.time(),
		Unread: slices.Contains(m.LabelIDs, "UNREAD"), Starred: slices.Contains(m.LabelIDs, "STARRED"),
		HasAttachments: h.MimeType == "multipart/mixed",
		MessageID:      h.header("Message-ID"), References: h.header("References"),
	}
	return out
}

var metadataQuery = "?format=metadata&metadataHeaders=From&metadataHeaders=To&metadataHeaders=Cc&metadataHeaders=Subject" +
	"&metadataHeaders=Reply-To&metadataHeaders=Message-ID&metadataHeaders=References"

func (s *session) metadata(ctx context.Context, id string) (message, error) {
	var m message
	err := s.c.do(ctx, http.MethodGet, gmail+"messages/"+url.PathEscape(id)+metadataQuery, nil, &m)
	return m, err
}

func (s *session) Mail(ctx context.Context, folder string, before time.Time, limit int) ([]model.Mail, bool, error) {
	q := url.Values{"maxResults": {strconv.Itoa(limit)}}
	if folder != archiveID {
		q.Set("labelIds", folder)
	}
	if folder == "SPAM" || folder == "TRASH" {
		q.Set("includeSpamTrash", "true")
	}
	if !before.IsZero() {
		// before: takes seconds and is exclusive; the exact cut is below.
		q.Set("q", "before:"+strconv.FormatInt(before.Unix()+1, 10))
	}
	var list struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
		NextPageToken string `json:"nextPageToken"`
	}
	if err := s.c.do(ctx, http.MethodGet, gmail+"messages?"+q.Encode(), nil, &list); err != nil {
		return nil, false, err
	}
	msgs := make([]message, len(list.Messages))
	err := parallel(len(msgs), func(i int) error {
		m, err := s.metadata(ctx, list.Messages[i].ID)
		msgs[i] = m
		if isStatus(err, http.StatusNotFound) {
			return nil // deleted meanwhile
		}
		return err
	})
	if err != nil {
		return nil, false, err
	}
	var out []model.Mail
	for _, m := range msgs {
		if m.ID == "" || (!before.IsZero() && m.time().After(before)) {
			continue
		}
		x := m.summary()
		x.FolderID = folder
		out = append(out, x)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	return out, list.NextPageToken != "", nil
}

func decodeData(s string) ([]byte, error) {
	s = strings.TrimRight(s, "=")
	return base64.RawURLEncoding.DecodeString(s)
}

func (s *session) MailBody(ctx context.Context, id string) (model.Mail, error) {
	var m message
	if err := s.c.do(ctx, http.MethodGet, gmail+"messages/"+url.PathEscape(id)+"?format=full", nil, &m); err != nil {
		return model.Mail{}, err
	}
	out := m.summary()
	var walk func(p part)
	walk = func(p part) {
		switch {
		case p.Filename != "":
			ref := ""
			if p.Body.AttachmentID != "" {
				ref = "att:" + m.ID + ":" + p.Body.AttachmentID
			} else if p.Body.Data != "" {
				ref = "att:" + m.ID + ":part=" + p.PartID
			}
			out.Attachments = append(out.Attachments, model.Attachment{
				Kind: mailutil.AttachmentKind(p.MimeType), Name: p.Filename, MIME: p.MimeType, Size: p.Body.Size, Ref: ref,
			})
		case p.MimeType == "text/plain" && out.Text == "":
			if b, err := decodeData(p.Body.Data); err == nil {
				out.Text = string(b)
			}
		case p.MimeType == "text/html" && out.HTML == "":
			if b, err := decodeData(p.Body.Data); err == nil {
				out.HTML = string(b)
			}
		}
		for _, c := range p.Parts {
			walk(c)
		}
	}
	walk(m.Payload)
	if strings.TrimSpace(out.Text) == "" && out.HTML != "" {
		out.Text = mailutil.HTMLToText(out.HTML)
	}
	out.HasAttachments = len(out.Attachments) > 0
	return out, nil
}

// attachment fetches "<message>:<attachment id>" or "<message>:part=<id>".
func (s *session) attachment(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	msg, att, ok := strings.Cut(ref, ":")
	if !ok {
		return nil, "", fmt.Errorf("google: bad attachment ref")
	}
	if partID, ok := strings.CutPrefix(att, "part="); ok {
		var m message
		if err := s.c.do(ctx, http.MethodGet, gmail+"messages/"+url.PathEscape(msg)+"?format=full", nil, &m); err != nil {
			return nil, "", err
		}
		var found *part
		var walk func(p *part)
		walk = func(p *part) {
			if p.PartID == partID {
				found = p
			}
			for i := range p.Parts {
				walk(&p.Parts[i])
			}
		}
		walk(&m.Payload)
		if found == nil {
			return nil, "", fmt.Errorf("google: no such part")
		}
		data, err := decodeData(found.Body.Data)
		if int64(len(data)) > limit {
			return nil, "", provider.ErrTooLarge
		}
		return data, found.MimeType, err
	}
	var a struct {
		Size int64  `json:"size"`
		Data string `json:"data"`
	}
	if err := s.c.do(ctx, http.MethodGet, gmail+"messages/"+url.PathEscape(msg)+"/attachments/"+url.PathEscape(att), nil, &a); err != nil {
		return nil, "", err
	}
	if a.Size > limit {
		return nil, "", provider.ErrTooLarge
	}
	data, err := decodeData(a.Data)
	return data, "", err
}

func (s *session) SendMail(ctx context.Context, out model.OutMail) error {
	s.mu.Lock()
	from := s.self
	s.mu.Unlock()
	body := map[string]string{}
	var reply *mailutil.Reply
	if out.ReplyTo != "" {
		m, err := s.metadata(ctx, out.ReplyTo)
		if err != nil {
			return err
		}
		reply = &mailutil.Reply{MessageID: m.Payload.header("Message-ID"), References: m.Payload.header("References")}
		body["threadId"] = m.ThreadID
	}
	raw := mailutil.Compose(from, out, reply, time.Now())
	if len(out.Bcc) > 0 {
		// Gmail takes the recipients from the headers: Bcc goes in and
		// Gmail strips it from what the others receive.
		hdr := "Bcc: "
		for i, a := range out.Bcc {
			if i > 0 {
				hdr += ", "
			}
			hdr += a.Email
		}
		raw = append([]byte(hdr+"\r\n"), raw...)
	}
	body["raw"] = base64.RawURLEncoding.EncodeToString(raw)
	return s.c.do(ctx, http.MethodPost, gmail+"messages/send", body, nil)
}

func (s *session) UpdateMail(ctx context.Context, ids []string, action model.MailAction) error {
	var add, remove []string
	switch action {
	case model.MailRead:
		remove = []string{"UNREAD"}
	case model.MailUnread:
		add = []string{"UNREAD"}
	case model.MailStar:
		add = []string{"STARRED"}
	case model.MailUnstar:
		remove = []string{"STARRED"}
	case model.MailTrash:
		return parallel(len(ids), func(i int) error {
			return s.c.do(ctx, http.MethodPost, gmail+"messages/"+url.PathEscape(ids[i])+"/trash", nil, nil)
		})
	default:
		return provider.ErrUnsupported
	}
	body := map[string]any{"ids": ids}
	if add != nil {
		body["addLabelIds"] = add
	}
	if remove != nil {
		body["removeLabelIds"] = remove
	}
	return s.c.do(ctx, http.MethodPost, gmail+"messages/batchModify", body, nil)
}

// MoveMail moves messages between labels: the target's label goes on, the
// listed folder's comes off. All mail is no label (moving there archives);
// Sent, Drafts and Starred are not places a message leaves by moving.
func (s *session) MoveMail(ctx context.Context, ids []string, from, to string) error {
	if to == "TRASH" {
		return s.UpdateMail(ctx, ids, model.MailTrash)
	}
	if from == "TRASH" {
		if err := parallel(len(ids), func(i int) error {
			return s.c.do(ctx, http.MethodPost, gmail+"messages/"+url.PathEscape(ids[i])+"/untrash", nil, nil)
		}); err != nil {
			return err
		}
	}
	body := map[string]any{"ids": ids}
	if to != archiveID {
		body["addLabelIds"] = []string{to}
	}
	switch from {
	case "", archiveID, "TRASH", "SENT", "DRAFT", "STARRED", "IMPORTANT":
	default:
		body["removeLabelIds"] = []string{from}
	}
	if body["addLabelIds"] == nil && body["removeLabelIds"] == nil {
		return nil
	}
	return s.c.do(ctx, http.MethodPost, gmail+"messages/batchModify", body, nil)
}

// resetHistory starts watching for changes from now.
func (s *session) resetHistory(ctx context.Context) error {
	var p struct {
		EmailAddress string `json:"emailAddress"`
		HistoryID    string `json:"historyId"`
	}
	if err := s.c.do(ctx, http.MethodGet, gmail+"profile", nil, &p); err != nil {
		return err
	}
	s.mu.Lock()
	s.historyID = p.HistoryID
	s.mu.Unlock()
	return nil
}

// pollMail reads the mailbox history since the last poll: messages that
// came into the inbox unread are reported as new, anything else as a
// change.
func (s *session) pollMail(ctx context.Context, sink provider.Sink) error {
	s.mu.Lock()
	start := s.historyID
	s.mu.Unlock()
	var added []string
	changed := false
	page := ""
	latest := start
	for {
		q := url.Values{"startHistoryId": {start}}
		if page != "" {
			q.Set("pageToken", page)
		}
		var h struct {
			History []struct {
				MessagesAdded []struct {
					Message struct {
						ID       string   `json:"id"`
						LabelIDs []string `json:"labelIds"`
					} `json:"message"`
				} `json:"messagesAdded"`
			} `json:"history"`
			HistoryID     string `json:"historyId"`
			NextPageToken string `json:"nextPageToken"`
		}
		err := s.c.do(ctx, http.MethodGet, gmail+"history?"+q.Encode(), nil, &h)
		if isStatus(err, http.StatusNotFound) {
			// Too old: start again and let the app reload.
			if err := s.resetHistory(ctx); err != nil {
				return err
			}
			sink.Emit(model.Event{Kind: model.EvMailChanged})
			return nil
		}
		if err != nil {
			return err
		}
		if len(h.History) > 0 {
			changed = true
		}
		for _, e := range h.History {
			for _, a := range e.MessagesAdded {
				l := a.Message.LabelIDs
				if slices.Contains(l, "INBOX") && slices.Contains(l, "UNREAD") && !slices.Contains(l, "SENT") {
					added = append(added, a.Message.ID)
				}
			}
		}
		if h.HistoryID != "" {
			latest = h.HistoryID
		}
		if h.NextPageToken == "" {
			break
		}
		page = h.NextPageToken
	}
	s.mu.Lock()
	s.historyID = latest
	s.mu.Unlock()
	for _, id := range added {
		m, err := s.metadata(ctx, id)
		if err != nil {
			continue
		}
		x := m.summary()
		x.FolderID = "INBOX"
		sink.Emit(model.Event{Kind: model.EvMailNew, Mail: &x})
	}
	if changed && len(added) == 0 {
		sink.Emit(model.Event{Kind: model.EvMailChanged})
	}
	return nil
}
