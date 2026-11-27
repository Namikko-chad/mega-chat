package vault

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func acct(id, login string, added, updated int) Account {
	return Account{ID: id, Provider: "slack", Name: id, Login: login,
		Added: time.Unix(int64(added), 0).UTC(), Updated: time.Unix(int64(updated), 0).UTC()}
}

func ids(c *Config) []string {
	var out []string
	for _, a := range c.Accounts {
		out = append(out, a.ID+"="+a.Name)
	}
	return out
}

func TestMergeIsDeterministic(t *testing.T) {
	base, _ := New()
	now := time.Now().UTC()
	x := base.clone()
	x.Version = 5
	x.Accounts = []Account{acct("a", "L1", 1, 10), acct("b", "L2", 2, 20)}
	y := base.clone()
	y.Version = 5
	ya := acct("a", "L1", 1, 30)
	ya.Name = "renamed"
	y.Accounts = []Account{ya, acct("c", "L3", 3, 3), acct("d", "L2", 4, 4)}
	y.Removed = []Removed{{ID: "b", At: now}}

	m1, err := Merge(x, y)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := Merge(y, x)
	if err != nil {
		t.Fatal(err)
	}
	got := ids(m1)
	if want := []string{"a=renamed", "c=c", "d=d"}; len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("merged: %v", got)
	}
	if !Same(m1, m2) || m1.Version != 6 || m2.Version != 6 {
		t.Fatalf("order matters: %v / %v", ids(m1), ids(m2))
	}

	// Two accounts with one login: the first added stays.
	z := base.clone()
	z.Accounts = []Account{acct("e", "L3", 0, 0)}
	m3, err := Merge(m1, z)
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(m3); len(got) != 3 || got[2] != "e=e" && got[1] != "e=e" {
		t.Fatalf("same login: %v", got)
	}
	for _, a := range m3.Accounts {
		if a.ID == "c" {
			t.Fatal("c signs in as e's login and was added later")
		}
	}

	other, _ := New()
	if _, err := Merge(x, other); !errors.Is(err, ErrOtherVault) {
		t.Fatalf("other vault: %v", err)
	}
}

func TestUpdateStampsAndTombstones(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.parley")
	c, _ := New()
	v, err := Create(path, []byte("pw"), c, fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	writes := 0
	v.OnWrite(func() { writes++ })
	v.Update(func(c *Config) error {
		c.Accounts = append(c.Accounts, Account{ID: "a1", Provider: "slack", Login: "x"})
		return nil
	})
	if a, _ := v.Config().Account("a1"); a.Updated.IsZero() {
		t.Fatal("no Updated")
	}
	v.Update(func(c *Config) error { c.Accounts = nil; return nil })
	if r := v.Config().Removed; len(r) != 1 || r[0].ID != "a1" {
		t.Fatalf("tombstone: %v", r)
	}
	if writes != 2 {
		t.Fatalf("writes: %d", writes)
	}
}

func TestAdoptAndMergeIn(t *testing.T) {
	dir := t.TempDir()
	c, _ := New()
	a, err := Create(filepath.Join(dir, "a"), []byte("pw"), c, fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	// b is a copy of a on another device.
	b, err := Create(filepath.Join(dir, "b"), []byte("pw"), c, fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Adopt(a.Bytes(), Hash(b.Bytes()), nil); !errors.Is(err, ErrBadPassword) {
		t.Fatalf("another salt needs the password: %v", err)
	}
	if err := b.Adopt(a.Bytes(), Hash(b.Bytes()), [][]byte{[]byte("pw")}); err != nil {
		t.Fatal(err)
	}
	if err := b.Adopt(a.Bytes(), "stale", nil); !errors.Is(err, ErrChanged) {
		t.Fatalf("stale: %v", err)
	}
	a.Update(func(c *Config) error {
		c.Accounts = append(c.Accounts, Account{ID: "x", Provider: "slack", Login: "1"})
		return nil
	})
	b.Update(func(c *Config) error {
		c.Accounts = append(c.Accounts, Account{ID: "y", Provider: "slack", Login: "2"})
		return nil
	})
	if err := a.ChangePassword([]byte("new"), fastKDF); err != nil {
		t.Fatal(err)
	}
	if _, err := b.MergeIn([][]byte{a.Bytes()}, nil); !errors.Is(err, ErrBadPassword) {
		t.Fatalf("a's new password: %v", err)
	}
	data, err := b.MergeIn([][]byte{a.Bytes()}, [][]byte{[]byte("new")})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(b.Config().Accounts); n != 2 {
		t.Fatalf("merged accounts: %d", n)
	}
	// The latest password wins.
	if _, err := OpenWithKey(filepath.Join(dir, "b"), a.UnlockKey()); err != nil {
		t.Fatalf("merged file opens with a's key: %v", err)
	}
	if err := a.Adopt(data, Hash(a.Bytes()), nil); err != nil {
		t.Fatalf("a adopts the merge: %v", err)
	}
}
