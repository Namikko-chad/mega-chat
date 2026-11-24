package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"megachat/core/crypto"
	"megachat/core/model"
)

func open(t *testing.T) (*Store, string) {
	t.Helper()
	k, _ := crypto.NewKey()
	path := filepath.Join(t.TempDir(), "cache.db")
	s, err := Open(path, k)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, path
}

func msg(i int) model.Message {
	return model.Message{AccountID: "a", ChatID: "c", ID: fmt.Sprint(i), Seq: int64(i), Text: fmt.Sprintf("secret text %d", i),
		Time: time.Unix(1_700_000_000, 0).Add(time.Duration(i/2) * time.Second)}
}

// page returns messages hi-1 .. lo, newest first.
func page(lo, hi int) []model.Message {
	var out []model.Message
	for i := hi - 1; i >= lo; i-- {
		out = append(out, msg(i))
	}
	return out
}

func ids(ms []model.Message) string {
	var b bytes.Buffer
	for _, m := range ms {
		fmt.Fprint(&b, m.ID, " ")
	}
	return b.String()
}

func TestHistoryRunsAndGaps(t *testing.T) {
	s, _ := open(t)
	// Live message 20 first: alone, not linked to anything.
	if _, err := s.PutMessage(msg(20)); err != nil {
		t.Fatal(err)
	}
	got, end, _ := s.History("a", "c", "", 5)
	if ids(got) != "20 " || end != EndGap {
		t.Fatalf("latest: %q %v", ids(got), end)
	}
	// Fetched page before 20: 19..15, more exist.
	if err := s.PutHistory("a", "c", "20", page(15, 20), true); err != nil {
		t.Fatal(err)
	}
	got, end, _ = s.History("a", "c", "", 4)
	if ids(got) != "20 19 18 17 " || end != EndLimit {
		t.Fatalf("after page: %q %v", ids(got), end)
	}
	got, end, _ = s.History("a", "c", "17", 10)
	if ids(got) != "16 15 " || end != EndGap {
		t.Fatalf("to gap: %q %v", ids(got), end)
	}
	// An old message cached out of band must not join the run.
	s.PutMessage(msg(3))
	got, end, _ = s.History("a", "c", "15", 10)
	if len(got) != 0 || end != EndGap {
		t.Fatalf("before unlinked oldest: %q %v", ids(got), end)
	}
	// The rest, down to the first message.
	if err := s.PutHistory("a", "c", "15", page(0, 15), false); err != nil {
		t.Fatal(err)
	}
	got, end, _ = s.History("a", "c", "5", 10)
	if ids(got) != "4 3 2 1 0 " || end != EndFirst {
		t.Fatalf("to first: %q %v", ids(got), end)
	}
	got, end, _ = s.History("a", "c", "0", 10)
	if len(got) != 0 || end != EndFirst {
		t.Fatalf("before first: %q %v", ids(got), end)
	}
	// An edit keeps the place in the run.
	m := msg(18)
	m.Text = "edited"
	prev, _ := s.PutMessage(m)
	if prev == nil || prev.Text != "secret text 18" {
		t.Fatalf("prev: %+v", prev)
	}
	got, _, _ = s.History("a", "c", "", 21)
	if len(got) != 21 || got[2].Text != "edited" {
		t.Fatalf("full: %q", ids(got))
	}
}

func TestEmptyPageMarksFirst(t *testing.T) {
	s, _ := open(t)
	s.PutMessage(msg(1))
	s.PutHistory("a", "c", "1", nil, false)
	got, end, _ := s.History("a", "c", "", 10)
	if ids(got) != "1 " || end != EndFirst {
		t.Fatalf("%q %v", ids(got), end)
	}
}

func TestChatsKVAndDelete(t *testing.T) {
	s, path := open(t)
	s.PutChat(model.Chat{AccountID: "a", ID: "old", Title: "Old chat", Updated: time.Unix(1, 0)})
	s.PutChat(model.Chat{AccountID: "a", ID: "new", Title: "New chat", Updated: time.Unix(2, 0)})
	s.PutChat(model.Chat{AccountID: "b", ID: "x", Title: "Other account"})
	cs, _ := s.Chats("a")
	if len(cs) != 2 || cs[0].ID != "new" {
		t.Fatalf("chats: %+v", cs)
	}
	kv := s.KV("a")
	kv.Put("session", []byte("auth key"))
	if v, _ := kv.Get("session"); string(v) != "auth key" {
		t.Fatalf("kv: %q", v)
	}
	if v, _ := s.KV("b").Get("session"); v != nil {
		t.Fatal("kv must be per account")
	}
	s.PutMessage(model.Message{AccountID: "a", ChatID: "new", ID: "m1", Text: "secret text"})
	if err := s.DeleteAccount("a"); err != nil {
		t.Fatal(err)
	}
	if cs, _ := s.Chats("a"); len(cs) != 0 {
		t.Fatal("account not deleted")
	}
	if cs, _ := s.Chats("b"); len(cs) != 1 {
		t.Fatal("other account lost")
	}
	s.PutMessage(msg(1))
	s.Close()
	raw, _ := os.ReadFile(path)
	for _, leak := range []string{"secret text", "Other account", "auth key"} {
		if bytes.Contains(raw, []byte(leak)) {
			t.Fatalf("plaintext %q in the file", leak)
		}
	}
}

func TestSecondOpenLocked(t *testing.T) {
	_, path := open(t)
	k, _ := crypto.NewKey()
	if _, err := Open(path, k); err != ErrLocked {
		t.Fatalf("second open: %v", err)
	}
}

func TestDeleteMessageUnlinksNext(t *testing.T) {
	s, _ := open(t)
	if err := s.PutHistory("a", "c", "", page(0, 6), false); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.DeleteMessage("a", "c", "2"); !ok || err != nil {
		t.Fatalf("delete: %v %v", ok, err)
	}
	if _, ok, _ := s.Message("a", "c", "2"); ok {
		t.Fatal("still there")
	}
	got, end, _ := s.History("a", "c", "", 10)
	if ids(got) != "5 4 3 " || end != EndGap {
		t.Fatalf("after delete: %q %v", ids(got), end)
	}
	if ok, err := s.DeleteMessage("a", "c", "missing"); ok || err != nil {
		t.Fatalf("missing: %v %v", ok, err)
	}
}

func TestBlankKeepsStoredText(t *testing.T) {
	s, _ := open(t)
	if err := s.PutHistory("a", "c", "", page(1, 3), false); err != nil {
		t.Fatal(err)
	}
	blank := func(i int) model.Message { m := msg(i); m.Text = ""; return m }
	if err := s.PutHistory("a", "c", "", []model.Message{blank(2), blank(1)}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.PutMessage(blank(2)); err != nil {
		t.Fatal(err)
	}
	got, _, _ := s.History("a", "c", "", 10)
	if len(got) != 2 || got[0].Text != msg(2).Text || got[1].Text != msg(1).Text {
		t.Fatalf("stored text lost: %+v", got)
	}
	gone := msg(2)
	gone.Text, gone.Deleted = "", true
	if _, err := s.PutMessage(gone); err != nil {
		t.Fatal(err)
	}
	if got, _, _ = s.History("a", "c", "", 10); !got[0].Deleted || got[0].Text != "" {
		t.Fatalf("a deletion must still apply: %+v", got[0])
	}
}
