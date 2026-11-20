package fake

import (
	"bytes"
	"context"
	"testing"
)

func TestMediaTone(t *testing.T) {
	var s session
	b, mime, err := s.Media(context.Background(), "wav:1:2", 1<<20)
	if err != nil || mime != "audio/wav" {
		t.Fatalf("Media: %v %q", err, mime)
	}
	if !bytes.HasPrefix(b, []byte("RIFF")) || len(b) != 44+2*2*8000 {
		t.Fatalf("bad wav: %d bytes", len(b))
	}
	if _, _, err := s.Media(context.Background(), "wav:1:2", 100); err == nil {
		t.Fatal("no limit error")
	}
}
