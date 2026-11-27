package vk

// Uploading files to send. VK takes two round trips: ask the API for an
// upload server, POST the bytes there (multipart), then save the result
// back through the API to get an attachment string for messages.send.
//   - images  → photos.getMessagesUploadServer + photos.saveMessagesPhoto
//   - the rest → docs.getMessagesUploadServer(type=doc) + docs.save
// UNVERIFIED against a live server beyond the public VK API shape.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/SelfHold/parley/core/model"
)

// uploadImage uploads one image and returns its "photo<owner>_<id>" ref.
func (s *server) uploadImage(ctx context.Context, accessToken string, peerID int64, name string, data []byte) (string, error) {
	raw, err := s.method(ctx, "photos.getMessagesUploadServer", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10),
	})
	if err != nil {
		return "", err
	}
	var srv struct {
		UploadURL string `json:"upload_url"`
	}
	if err := json.Unmarshal(raw, &srv); err != nil || srv.UploadURL == "" {
		return "", fmt.Errorf("vk: photos upload server: %w", err)
	}
	up, err := s.uploadTo(ctx, srv.UploadURL, "photo", name, data)
	if err != nil {
		return "", err
	}
	var ur struct {
		Server int64  `json:"server"`
		Photo  string `json:"photo"`
		Hash   string `json:"hash"`
	}
	if err := json.Unmarshal(up, &ur); err != nil {
		return "", fmt.Errorf("vk: photo upload reply: %w", err)
	}
	saved, err := s.method(ctx, "photos.saveMessagesPhoto", accessToken, map[string]string{
		"server": strconv.FormatInt(ur.Server, 10), "photo": ur.Photo, "hash": ur.Hash,
	})
	if err != nil {
		return "", err
	}
	var photos []struct {
		ID        int64  `json:"id"`
		OwnerID   int64  `json:"owner_id"`
		AccessKey string `json:"access_key"`
	}
	if err := json.Unmarshal(saved, &photos); err != nil || len(photos) == 0 {
		return "", fmt.Errorf("vk: saveMessagesPhoto: %w", err)
	}
	p := photos[0]
	ref := fmt.Sprintf("photo%d_%d", p.OwnerID, p.ID)
	if p.AccessKey != "" {
		ref += "_" + p.AccessKey
	}
	return ref, nil
}

// uploadDoc uploads one file as a document and returns its "doc<owner>_<id>"
// ref. kind picks the doc type ("doc" or "audio_message").
func (s *server) uploadDoc(ctx context.Context, accessToken string, peerID int64, kind, name string, data []byte) (string, error) {
	raw, err := s.method(ctx, "docs.getMessagesUploadServer", accessToken, map[string]string{
		"peer_id": strconv.FormatInt(peerID, 10), "type": kind,
	})
	if err != nil {
		return "", err
	}
	var srv struct {
		UploadURL string `json:"upload_url"`
	}
	if err := json.Unmarshal(raw, &srv); err != nil || srv.UploadURL == "" {
		return "", fmt.Errorf("vk: docs upload server: %w", err)
	}
	up, err := s.uploadTo(ctx, srv.UploadURL, "file", name, data)
	if err != nil {
		return "", err
	}
	var ur struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(up, &ur); err != nil || ur.File == "" {
		return "", fmt.Errorf("vk: doc upload reply: %w", err)
	}
	saved, err := s.method(ctx, "docs.save", accessToken, map[string]string{
		"file": ur.File, "title": name,
	})
	if err != nil {
		return "", err
	}
	// docs.save answers {"type":"doc","doc":{...}} or {"type":"audio_message",
	// "audio_message":{...}}.
	var sv struct {
		Type         string          `json:"type"`
		Doc          *docObj         `json:"doc"`
		AudioMessage *docObj         `json:"audio_message"`
		Extra        json.RawMessage `json:"-"`
	}
	if err := json.Unmarshal(saved, &sv); err != nil {
		return "", fmt.Errorf("vk: docs.save: %w", err)
	}
	obj := sv.Doc
	prefix := "doc"
	if sv.Type == "audio_message" && sv.AudioMessage != nil {
		obj, prefix = sv.AudioMessage, "audio_message"
	}
	if obj == nil {
		return "", fmt.Errorf("vk: docs.save: empty result")
	}
	return fmt.Sprintf("%s%d_%d", prefix, obj.OwnerID, obj.ID), nil
}

type docObj struct {
	ID      int64 `json:"id"`
	OwnerID int64 `json:"owner_id"`
}

// uploadTo POSTs one file as multipart/form-data to a VK upload URL.
func (s *server) uploadTo(ctx context.Context, uploadURL, field, name string, data []byte) (json.RawMessage, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if name == "" {
		name = "file"
	}
	fw, err := mw.CreateFormFile(field, name)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(data); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vk: upload: HTTP %d", resp.StatusCode)
	}
	return out, nil
}

// docKind picks the VK doc type for a file.
func docKind(f model.OutFile) string {
	if f.Kind() == model.AttachVoice || strings.HasPrefix(f.MIME, "audio/ogg") {
		return "audio_message"
	}
	return "doc"
}
