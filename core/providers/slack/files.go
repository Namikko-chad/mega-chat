package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SelfHold/parley/core/model"
)

// SendFiles shares files the way Slack's current API wants it: each file
// gets an upload URL (files.getUploadURLExternal), its bytes go there, and
// files.completeUploadExternal posts them all as one message with the text
// as its comment. Slack does not answer with the message; it comes over
// RTM like any other. A reply goes into the message's thread (an upload
// cannot be broadcast to the channel).
func (s *session) SendFiles(ctx context.Context, chatID string, o model.Outgoing) ([]model.Message, error) {
	type done struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	var files []done
	for _, f := range o.Files {
		var r struct {
			UploadURL string `json:"upload_url"`
			FileID    string `json:"file_id"`
		}
		q := url.Values{"filename": {f.Name}, "length": {strconv.Itoa(len(f.Data))}}
		if err := s.call(ctx, "files.getUploadURLExternal", q, &r); err != nil {
			return nil, err
		}
		if r.UploadURL == "" || r.FileID == "" {
			return nil, errors.New("slack: no upload URL for the file")
		}
		if err := s.upload(ctx, r.UploadURL, f); err != nil {
			return nil, err
		}
		files = append(files, done{ID: r.FileID, Title: f.Name})
	}
	list, _ := json.Marshal(files)
	q := url.Values{"files": {string(list)}, "channel_id": {chatID}}
	if text := outText(o); text != "" {
		q.Set("initial_comment", text)
	}
	switch {
	case o.Thread != "":
		q.Set("thread_ts", o.Thread)
	case o.ReplyTo != "":
		q.Set("thread_ts", o.ReplyTo)
	}
	if err := s.call(ctx, "files.completeUploadExternal", q, nil); err != nil {
		return nil, err
	}
	return nil, nil
}

// upload posts a file's bytes to the URL Slack gave for it. The account's
// credentials go along only to Slack's own hosts (mediaHost).
func (s *session) upload(ctx context.Context, raw string, f model.OutFile) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("slack: bad upload URL")
	}
	ok, creds := s.mediaHost(u)
	if !ok {
		return fmt.Errorf("slack: upload to %s refused", u.Hostname())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, raw, bytes.NewReader(f.Data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", f.MIME)
	a := s.client()
	if creds {
		if a.token != "" {
			req.Header.Set("Authorization", "Bearer "+a.token)
		}
		if a.cookie != "" {
			req.Header.Set("Cookie", "d="+url.QueryEscape(a.cookie))
		}
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: upload: HTTP %d", resp.StatusCode)
	}
	return nil
}
