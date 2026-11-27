package vault

// Several copies of one vault (CloudSync gives every device the vault file
// and keeps concurrent writes as several versions, CONNECTOR.md § 6.2):
// Merge folds them into one config, the same on every device whatever the
// order of the copies; Adopt takes a copy as it is.

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

// TombstoneTTL is how long a removed account's tombstone is kept: a copy
// older than that would bring the account back.
const TombstoneTTL = 90 * 24 * time.Hour

var (
	// ErrOtherVault: the copy is another vault (another id or master key).
	ErrOtherVault = errors.New("vault: the copy is another vault")
	// ErrChanged: the vault file is not the one the caller expected.
	ErrChanged = errors.New("vault: the file changed meanwhile")
)

// Hash names a vault file's bytes (hex SHA-256); "" for no file.
func Hash(data []byte) string {
	if data == nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// CheckFile tells whether data looks like a vault file (its header).
func CheckFile(data []byte) error {
	_, _, err := parseHeader(data)
	return err
}

// ReplaceFile writes data as the vault file at path if the file there is
// still the one with hash expect ("" = no file), else ErrChanged. Only the
// header is checked: the vault is locked here, so data cannot be opened.
func ReplaceFile(path string, data []byte, expect string) error {
	if err := CheckFile(data); err != nil {
		return err
	}
	cur, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if Hash(cur) != expect {
		return ErrChanged
	}
	return writeFile(path, data)
}

// Merge folds copies of one vault into one config. The result does not
// depend on the order of cfgs:
//   - accounts are united by ID, the copy with the latest Updated wins
//     (then the higher Version, then the smaller JSON);
//   - a tombstone removes its account in every copy; tombstones are united
//     and dropped after TombstoneTTL;
//   - the account order is the newest copy's, accounts it lacks follow by
//     Added;
//   - two accounts signing in as the same login (added on two devices):
//     the one added first stays, the other gets a tombstone;
//   - Version is the highest + 1, PasswordChanged the latest.
//
// Copies of different vaults fail with ErrOtherVault.
func Merge(cfgs ...*Config) (*Config, error) {
	if len(cfgs) == 0 {
		return nil, errors.New("vault: nothing to merge")
	}
	type copyOf struct {
		c  *Config
		js []byte
	}
	all := make([]copyOf, 0, len(cfgs))
	for _, c := range cfgs {
		js, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		all = append(all, copyOf{c.clone(), js})
	}
	// Newest first; equal versions by their content.
	slices.SortFunc(all, func(x, y copyOf) int {
		if d := cmp.Compare(y.c.Version, x.c.Version); d != 0 {
			return d
		}
		return bytes.Compare(x.js, y.js)
	})
	base := all[0].c
	for _, x := range all[1:] {
		if x.c.VaultID != base.VaultID || !bytes.Equal(x.c.MasterKey, base.MasterKey) {
			return nil, ErrOtherVault
		}
	}
	now := time.Now().UTC()

	removed := map[string]time.Time{}
	for _, x := range all {
		for _, r := range x.c.Removed {
			if at, ok := removed[r.ID]; !ok || r.At.After(at) {
				removed[r.ID] = r.At
			}
		}
	}

	// The winning copy of every account; rank says which copy it came from
	// (all is sorted newest first, so a lower rank is newer).
	type pick struct {
		a    Account
		rank int
		js   []byte
	}
	best := map[string]pick{}
	for rank, x := range all {
		for _, a := range x.c.Accounts {
			if _, gone := removed[a.ID]; gone {
				continue
			}
			js, _ := json.Marshal(a)
			p, ok := best[a.ID]
			if !ok || newer(a, rank, js, p.a, p.rank, p.js) {
				best[a.ID] = pick{a, rank, js}
			}
		}
	}

	// One account per login: the first added stays.
	byAdded := make([]Account, 0, len(best))
	for _, p := range best {
		byAdded = append(byAdded, p.a)
	}
	slices.SortFunc(byAdded, func(x, y Account) int {
		if d := x.Added.Compare(y.Added); d != 0 {
			return d
		}
		return strings.Compare(x.ID, y.ID)
	})
	logins := map[string]bool{}
	for _, a := range byAdded {
		if a.Login == "" {
			continue
		}
		k := a.Provider + "\x00" + a.Login
		if logins[k] {
			delete(best, a.ID)
			// From the account itself, so every device makes the same.
			removed[a.ID] = a.Added
			if a.Updated.After(a.Added) {
				removed[a.ID] = a.Updated
			}
			continue
		}
		logins[k] = true
	}

	out := &Config{VaultID: base.VaultID, MasterKey: base.MasterKey}
	for _, a := range base.Accounts {
		if p, ok := best[a.ID]; ok {
			out.Accounts = append(out.Accounts, p.a)
			delete(best, a.ID)
		}
	}
	for _, a := range byAdded {
		if p, ok := best[a.ID]; ok {
			out.Accounts = append(out.Accounts, p.a)
		}
	}
	for id, at := range removed {
		if now.Sub(at) < TombstoneTTL {
			out.Removed = append(out.Removed, Removed{ID: id, At: at})
		}
	}
	slices.SortFunc(out.Removed, func(x, y Removed) int { return strings.Compare(x.ID, y.ID) })
	for _, x := range all {
		out.Version = max(out.Version, x.c.Version)
		if x.c.PasswordChanged.After(out.PasswordChanged) {
			out.PasswordChanged = x.c.PasswordChanged
		}
	}
	out.Version++
	return out, out.validate()
}

// newer says whether account a (from the copy of rank) beats b.
func newer(a Account, ra int, ja []byte, b Account, rb int, jb []byte) bool {
	if d := a.Updated.Compare(b.Updated); d != 0 {
		return d > 0
	}
	if ra != rb {
		return ra < rb
	}
	return bytes.Compare(ja, jb) < 0
}

// Same tells whether two configs hold the same vault with the same
// accounts (Version and tombstones aside).
func Same(a, b *Config) bool {
	if a.VaultID != b.VaultID || !bytes.Equal(a.MasterKey, b.MasterKey) || len(a.Accounts) != len(b.Accounts) {
		return false
	}
	for i := range a.Accounts {
		if !sameAccount(a.Accounts[i], b.Accounts[i]) {
			return false
		}
	}
	return true
}

// decode opens a copy of a vault file: with this vault's key when the
// copy has the same header, else with the key each password derives for
// the copy's header. ErrBadPassword when none opens it. Called with v.mu.
func (v *Vault) decode(data []byte, passwords [][]byte) (*sealer, *Config, error) {
	p, salt, err := parseHeader(data)
	if err != nil {
		return nil, nil, err
	}
	header := append([]byte{}, data[:headerSize]...)
	if bytes.Equal(header, v.sealer.header) {
		c, err := openWithKey(v.sealer.key, data)
		if err == nil {
			return v.sealer, c, nil
		}
		if !errors.Is(err, ErrBadPassword) {
			return nil, nil, err
		}
	}
	for _, pw := range passwords {
		k := deriveKey(pw, p, salt)
		c, err := openWithKey(k, data)
		if err == nil {
			return &sealer{header: header, key: k}, c, nil
		}
		if !errors.Is(err, ErrBadPassword) {
			return nil, nil, err
		}
	}
	return nil, nil, ErrBadPassword
}

// Compare opens a copy of the vault file (see Adopt for passwords) and
// tells whether it holds the same accounts (same) and whether it is this
// vault at all (sameVault: then the two can merge).
func (v *Vault) Compare(data []byte, passwords [][]byte) (same, sameVault bool, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	_, c, err := v.decode(data, passwords)
	if err != nil {
		return false, false, err
	}
	sameVault = c.VaultID == v.cfg.VaultID && bytes.Equal(c.MasterKey, v.cfg.MasterKey)
	return Same(c, v.cfg), sameVault, nil
}

// Adopt makes a copy of this vault's file (from another device) the vault,
// byte for byte, if the file is still the one with hash expect (else
// ErrChanged). The copy opens with this vault's key or, sealed with
// another password, with one of passwords (else ErrBadPassword); the vault
// then opens with that password from now on. Another vault fails with
// ErrOtherVault. OnWrite is not called.
func (v *Vault) Adopt(data []byte, expect string, passwords [][]byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if Hash(v.data) != expect {
		return ErrChanged
	}
	s, c, err := v.decode(data, passwords)
	if err != nil {
		return err
	}
	if c.VaultID != v.cfg.VaultID || !bytes.Equal(c.MasterKey, v.cfg.MasterKey) {
		return ErrOtherVault
	}
	if err := writeFile(v.path, data); err != nil {
		return err
	}
	v.sealer, v.cfg, v.data = s, c, append([]byte{}, data...)
	return nil
}

// MergeIn merges copies of this vault's file (see Adopt for passwords)
// with the vault (Merge), saves the result and returns the file. It is
// sealed with the latest password among the copies (PasswordChanged; the
// vault's own on a tie).
func (v *Vault) MergeIn(copies [][]byte, passwords [][]byte) ([]byte, error) {
	v.mu.Lock()
	cfgs := []*Config{v.cfg}
	s, latest := v.sealer, v.cfg.PasswordChanged
	for _, data := range copies {
		cs, c, err := v.decode(data, passwords)
		if err != nil {
			v.mu.Unlock()
			return nil, err
		}
		if c.PasswordChanged.After(latest) {
			s, latest = cs, c.PasswordChanged
		}
		cfgs = append(cfgs, c)
	}
	merged, err := Merge(cfgs...)
	if err != nil {
		v.mu.Unlock()
		return nil, err
	}
	if merged.VaultID != v.cfg.VaultID {
		v.mu.Unlock()
		return nil, ErrOtherVault
	}
	prevSealer, prevCfg := v.sealer, v.cfg
	v.sealer, v.cfg = s, merged
	if err := v.write(); err != nil {
		v.sealer, v.cfg = prevSealer, prevCfg
		v.mu.Unlock()
		return nil, fmt.Errorf("vault: save the merge: %w", err)
	}
	out := append([]byte{}, v.data...)
	v.mu.Unlock()
	return out, nil
}
