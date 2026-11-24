package telegram

import (
	"testing"

	"github.com/gotd/td/tg"

	"megachat/core/model"
)

func TestInputMedia(t *testing.T) {
	file := &tg.InputFile{ID: 1, Parts: 1, Name: "x"}
	if m := inputMedia(file, model.OutFile{Name: "a.jpg", MIME: "image/jpeg", Data: make([]byte, 10)}); !m.photo {
		t.Fatalf("a JPEG goes as a photo: %+v", m.media)
	}
	for _, f := range []model.OutFile{
		{Name: "big.png", MIME: "image/png", Data: make([]byte, maxPhoto+1)},
		{Name: "a.gif", MIME: "image/gif", Data: make([]byte, 10)},
	} {
		m := inputMedia(file, f)
		doc, ok := m.media.(*tg.InputMediaUploadedDocument)
		if m.photo || !ok || !doc.ForceFile {
			t.Fatalf("%s goes as a file as it is: %+v", f.Name, m.media)
		}
	}
	m := inputMedia(file, model.OutFile{Name: "r.pdf", MIME: "application/pdf", Data: make([]byte, 10)})
	doc, ok := m.media.(*tg.InputMediaUploadedDocument)
	if !ok || doc.MimeType != "application/pdf" || doc.ForceFile || len(doc.Attributes) != 1 {
		t.Fatalf("pdf: %+v", m.media)
	}
	if a, ok := doc.Attributes[0].(*tg.DocumentAttributeFilename); !ok || a.FileName != "r.pdf" {
		t.Fatalf("name: %+v", doc.Attributes)
	}
}
