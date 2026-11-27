package vk

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"megachat/core/model"
	"megachat/core/provider"
)

var (
	_ provider.FileSender = (*session)(nil)
	_ provider.Editor     = (*session)(nil)
	_ provider.Deleter    = (*session)(nil)
	_ provider.Forwarder  = (*session)(nil)
)

// SendFiles uploads every file first (a failed upload sends nothing), then
// sends one message carrying them all with the text as its caption. The
// message comes back over the long poll, so the result is empty.
func (s *session) SendFiles(ctx context.Context, chatID string, o model.Outgoing) ([]model.Message, error) {
	if len(o.Files) == 0 {
		return nil, errors.New("vk: no files")
	}
	peer, err := parsePeer(chatID)
	if err != nil {
		return nil, err
	}
	refs := make([]string, 0, len(o.Files))
	for _, f := range o.Files {
		var ref string
		if f.Kind() == model.AttachImage {
			ref, err = apiCall(ctx, s, func(tok string) (string, error) {
				return s.srv.uploadImage(ctx, tok, peer, f.Name, f.Data)
			})
		} else {
			kind := docKind(f)
			ref, err = apiCall(ctx, s, func(tok string) (string, error) {
				return s.srv.uploadDoc(ctx, tok, peer, kind, f.Name, f.Data)
			})
		}
		if err != nil {
			return nil, fmt.Errorf("vk: upload %q: %w", f.Name, err)
		}
		refs = append(refs, ref)
	}
	_, err = apiCall(ctx, s, func(tok string) (int64, error) {
		return s.srv.send(ctx, tok, peer, outText(o), sendOpts{
			replyTo: o.ReplyTo, attachment: strings.Join(refs, ","),
		})
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Edit replaces the text of the account's own message; the whole message
// comes back over the long poll as an edit.
func (s *session) Edit(ctx context.Context, chatID, msgID, text string) (model.Message, error) {
	peer, err := parsePeer(chatID)
	if err != nil {
		return model.Message{}, err
	}
	id, err := strconv.ParseInt(msgID, 10, 64)
	if err != nil {
		return model.Message{}, fmt.Errorf("vk: bad message id %q", msgID)
	}
	_, err = apiCall(ctx, s, func(tok string) (struct{}, error) {
		return struct{}{}, s.srv.edit(ctx, tok, peer, id, text)
	})
	if err != nil {
		return model.Message{}, err
	}
	return model.Message{ID: msgID, Text: text, Edited: time.Now()}, nil
}

// Delete deletes the account's own messages, for everyone (forAll) or only
// for this account.
func (s *session) Delete(ctx context.Context, chatID string, msgIDs []string, forAll bool) error {
	if len(msgIDs) == 0 {
		return nil
	}
	_, err := apiCall(ctx, s, func(tok string) (struct{}, error) {
		return struct{}{}, s.srv.del(ctx, tok, msgIDs, forAll)
	})
	return err
}

// DeletesForMe: VK can delete a message for this account alone.
func (s *session) DeletesForMe() bool { return true }

// Forward copies messages of one chat into another chat of the account. The
// copy comes back over the long poll, so the result is empty.
func (s *session) Forward(ctx context.Context, fromChat string, msgIDs []string, toChat string) ([]model.Message, error) {
	if len(msgIDs) == 0 {
		return nil, errors.New("vk: nothing to forward")
	}
	fromPeer, err := parsePeer(fromChat)
	if err != nil {
		return nil, err
	}
	toPeer, err := parsePeer(toChat)
	if err != nil {
		return nil, err
	}
	_, err = apiCall(ctx, s, func(tok string) (int64, error) {
		return s.srv.send(ctx, tok, toPeer, "", sendOpts{forwardFrom: fromPeer, forwardIDs: msgIDs})
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
