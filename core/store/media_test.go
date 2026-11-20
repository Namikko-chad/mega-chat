package store

import (
	"bytes"
	"os"
	"testing"
)

func TestMediaCache(t *testing.T) {
	s, path := open(t)
	secret := bytes.Repeat([]byte("PICTURE-SECRET"), 100)
	if err := s.PutMedia("a", "ref-1", secret, "image/png"); err != nil {
		t.Fatal(err)
	}
	data, mime, ok, err := s.Media("a", "ref-1")
	if err != nil || !ok || mime != "image/png" || !bytes.Equal(data, secret) {
		t.Fatalf("got %d bytes %q %v %v", len(data), mime, ok, err)
	}
	if _, _, ok, _ := s.Media("b", "ref-1"); ok {
		t.Fatal("media is per account")
	}
	s.db.Sync()
	raw, _ := os.ReadFile(path)
	if bytes.Contains(raw, []byte("PICTURE-SECRET")) || bytes.Contains(raw, []byte("ref-1")) {
		t.Fatal("media is readable on disk")
	}

	// Over the cap the least recently used entries go.
	old := MediaCap
	MediaCap = 3000
	t.Cleanup(func() { MediaCap = old })
	s.mediaTotal = -1
	for _, ref := range []string{"x1", "x2", "x3"} {
		if err := s.PutMedia("a", ref, make([]byte, 1000), ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, ok, _ := s.Media("a", "x3"); !ok {
		t.Fatal("the newest entry was dropped")
	}
	if _, _, ok, _ := s.Media("a", "ref-1"); ok {
		t.Fatal("the oldest entry was kept over the cap")
	}

	if err := s.DeleteAccount("a"); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, _ := s.Media("a", "x3"); ok {
		t.Fatal("media survived the account")
	}
}
