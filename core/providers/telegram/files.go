package telegram

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"megachat/core/model"
	"megachat/core/richtext"
)

// maxPhoto is the largest picture Telegram takes as a photo; a larger one
// goes as a file.
const maxPhoto = 10 << 20

// maxAlbum is how many media one album holds.
const maxAlbum = 10

// outMedia is one uploaded file to send.
type outMedia struct {
	media tg.InputMediaClass
	photo bool
}

// SendFiles uploads the files and sends them like Telegram's own apps:
// JPEG and PNG pictures as photos, everything else as files; several of a
// kind in a row make an album (photos and files are never mixed in one).
// The caption, the reply and the quote go with the first message.
func (s *session) SendFiles(ctx context.Context, chatID string, o model.Outgoing) ([]model.Message, error) {
	api, gaps, err := s.live()
	if err != nil {
		return nil, err
	}
	t, err := s.target(ctx, api, chatID, o)
	if err != nil {
		return nil, err
	}
	up := uploader.NewUploader(api)
	items := make([]outMedia, len(o.Files))
	for i, f := range o.Files {
		file, err := up.FromBytes(ctx, f.Name, f.Data)
		if err != nil {
			return nil, authErr(fmt.Errorf("telegram: upload %s: %w", f.Name, err))
		}
		items[i] = inputMedia(file, f)
	}
	var out []model.Message
	first := true
	for len(items) > 0 {
		n := 1
		for n < len(items) && n < maxAlbum && items[n].photo == items[0].photo {
			n++
		}
		batch := items[:n]
		items = items[n:]
		msgs, err := s.sendBatch(ctx, api, gaps, t, o, batch, first)
		if err != nil {
			return out, err
		}
		first = false
		out = append(out, msgs...)
	}
	if t.chat != chatID {
		// A comment: it belongs to the post's thread, not the group.
		for i := range out {
			out[i].ChatID, out[i].InThread = chatID, o.Thread
		}
	}
	return out, nil
}

// sendBatch sends one message (one item) or an album. The first batch
// carries o's caption, reply and quote; a quote Telegram rejects is sent
// in the caption instead.
func (s *session) sendBatch(ctx context.Context, api *tg.Client, gaps *updates.Manager, t sendTarget, o model.Outgoing, batch []outMedia, first bool) ([]model.Message, error) {
	quote := ""
	if first {
		quote = t.quote(o)
	}
	send := func(quote string) (tg.UpdatesClass, error) {
		text := ""
		var ents []tg.MessageEntityClass
		var reply tg.InputReplyToClass
		if first {
			var d richtext.Doc
			d, ents = s.outgoingText(o, caption(o, quote))
			text, reply = d.Text, t.reply(quote)
		} else {
			// Later albums stay in the topic or the comments.
			reply = sendTarget{replyTo: t.root, top: t.top}.reply("")
		}
		if len(batch) == 1 {
			req := &tg.MessagesSendMediaRequest{Peer: t.peer, Media: batch[0].media, Message: text, RandomID: randomID()}
			if len(ents) > 0 {
				req.SetEntities(ents)
			}
			if reply != nil {
				req.SetReplyTo(reply)
			}
			return api.MessagesSendMedia(ctx, req)
		}
		req := &tg.MessagesSendMultiMediaRequest{Peer: t.peer}
		for i, it := range batch {
			m, err := s.uploaded(ctx, api, t.peer, it)
			if err != nil {
				return nil, err
			}
			single := tg.InputSingleMedia{Media: m, RandomID: randomID()}
			if i == 0 {
				single.Message = text
				if len(ents) > 0 {
					single.SetEntities(ents)
				}
			}
			req.MultiMedia = append(req.MultiMedia, single)
		}
		if reply != nil {
			req.SetReplyTo(reply)
		}
		return api.MessagesSendMultiMedia(ctx, req)
	}
	res, err := send(quote)
	if quote != "" && tgerr.Is(err, "QUOTE_TEXT_INVALID") {
		res, err = send("")
	}
	if err != nil {
		return nil, authErr(err)
	}
	out := s.sent(res, t.chat)
	// Let the updates manager see the pts change.
	if err := gaps.Handle(ctx, res); err != nil {
		s.env.Logf("warn", "updates after sending files: %v", err)
	}
	return out, nil
}

// uploaded turns an uploaded file into media an album can refer to
// (messages.sendMultiMedia takes only media already on the server).
func (s *session) uploaded(ctx context.Context, api *tg.Client, peer tg.InputPeerClass, it outMedia) (tg.InputMediaClass, error) {
	res, err := api.MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{Peer: peer, Media: it.media})
	if err != nil {
		return nil, err
	}
	switch m := res.(type) {
	case *tg.MessageMediaPhoto:
		if p, ok := m.Photo.(*tg.Photo); ok {
			return &tg.InputMediaPhoto{ID: &tg.InputPhoto{ID: p.ID, AccessHash: p.AccessHash, FileReference: p.FileReference}}, nil
		}
	case *tg.MessageMediaDocument:
		if d, ok := m.Document.(*tg.Document); ok {
			return &tg.InputMediaDocument{ID: &tg.InputDocument{ID: d.ID, AccessHash: d.AccessHash, FileReference: d.FileReference}}, nil
		}
	}
	return nil, errors.New("telegram: the uploaded media came back empty")
}

// inputMedia is how an uploaded file is sent.
func inputMedia(file tg.InputFileClass, f model.OutFile) outMedia {
	if (f.MIME == "image/jpeg" || f.MIME == "image/png") && len(f.Data) <= maxPhoto {
		return outMedia{media: &tg.InputMediaUploadedPhoto{File: file}, photo: true}
	}
	doc := &tg.InputMediaUploadedDocument{
		File:       file,
		MimeType:   f.MIME,
		Attributes: []tg.DocumentAttributeClass{&tg.DocumentAttributeFilename{FileName: f.Name}},
	}
	if f.Kind() == model.AttachImage {
		// A large or animated picture: keep it a file, as sent.
		doc.SetForceFile(true)
	}
	return outMedia{media: doc}
}
