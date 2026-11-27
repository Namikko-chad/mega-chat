package vault

import (
	"errors"
	"path/filepath"
	"testing"
)

var fastKDF = KDFParams{Time: 1, Memory: 1024, Threads: 1}

func TestCreateOpenUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.parley")
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	v, err := Create(path, []byte("pw"), c, fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Create(path, []byte("pw"), c, fastKDF); err == nil {
		t.Fatal("second create must fail")
	}
	err = v.Update(func(c *Config) error {
		c.Accounts = append(c.Accounts, Account{ID: "a1", Provider: "slack", Login: "T1:U1", Params: map[string]string{"token": "x"}})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path, []byte("bad")); !errors.Is(err, ErrBadPassword) {
		t.Fatalf("wrong password: %v", err)
	}
	v2, err := Open(path, []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	a, ok := v2.Config().Account("a1")
	if !ok || a.Params["token"] != "x" || v2.Config().Version != 2 {
		t.Fatalf("reopened: %+v %v", a, v2.Config().Version)
	}
	k := v2.UnlockKey()
	if v3, err := OpenWithKey(path, k); err != nil || v3.Config().Version != 2 {
		t.Fatalf("open with key: %v", err)
	}
	if err := v2.ChangePassword([]byte("new"), fastKDF); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenWithKey(path, k); !errors.Is(err, ErrBadPassword) {
		t.Fatalf("old key after a password change: %v", err)
	}
	if v4, err := OpenWithKey(path, v2.UnlockKey()); err != nil || v4.Config().Version != 3 {
		t.Fatalf("new key: %v", err)
	}
}

func TestSameLoginRefused(t *testing.T) {
	c, _ := New()
	v, err := Create(filepath.Join(t.TempDir(), "v"), []byte("pw"), c, fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	add := func(id, prov, login string) error {
		return v.Update(func(c *Config) error {
			c.Accounts = append(c.Accounts, Account{ID: id, Provider: prov, Login: login})
			return nil
		})
	}
	if err := add("a", "telegram", "+1"); err != nil {
		t.Fatal(err)
	}
	if err := add("b", "slack", "+1"); err != nil {
		t.Fatal("same login of another provider is fine:", err)
	}
	if err := add("c", "telegram", "+1"); !errors.Is(err, ErrSameLogin) {
		t.Fatalf("duplicate login: %v", err)
	}
	if len(v.Config().Accounts) != 2 {
		t.Fatal("failed update must not change the vault")
	}
}
