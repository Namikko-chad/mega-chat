// Package vault stores the configuration (master key, messenger accounts
// and their credentials) in one file encrypted with a key derived from a
// password. Larger, often changing session state (Telegram auth keys,
// update cursors) lives in the store, sealed with a key derived from the
// master key, so the vault is rewritten only when accounts change.
//
// File layout:
//
//	magic "PARLEYV\x01" | argon2 time u32 | memory KiB u32 | threads u8 | salt [16]
//	| XChaCha20-Poly1305(key, JSON config, ad = everything before it)
package vault

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/SelfHold/parley/core/crypto"
)

var magic = []byte("PARLEYV\x01")

const (
	saltSize   = 16
	headerSize = 8 + 4 + 4 + 1 + saltSize
)

var (
	ErrBadPassword = errors.New("vault: wrong password or corrupted file")
	ErrFormat      = errors.New("vault: not a vault file")
	// ErrSameLogin: an account of this provider already signs in as this user.
	ErrSameLogin = errors.New("vault: this account is already added")
)

// KDFParams are the Argon2id cost parameters.
type KDFParams struct {
	Time    uint32
	Memory  uint32 // KiB
	Threads uint8
}

// DefaultKDF is strong enough for a password and still fine on phones.
var DefaultKDF = KDFParams{Time: 3, Memory: 64 * 1024, Threads: 4}

// Config is the decrypted content of the vault.
type Config struct {
	VaultID   string    `json:"vault_id"`
	Version   uint64    `json:"version"` // bumped on every change
	MasterKey []byte    `json:"master_key"`
	Accounts  []Account `json:"accounts"`
	// Removed are tombstones of removed accounts, so that merging with a
	// copy made before the removal (Shard, merge.go) does not bring
	// them back. Kept for TombstoneTTL.
	Removed []Removed `json:"removed,omitempty"`
	// PasswordChanged is when the password last changed: when copies
	// sealed with different passwords merge, the latest password wins.
	PasswordChanged time.Time `json:"password_changed,omitempty"`
}

// Removed is the tombstone of a removed account.
type Removed struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
}

// Account is one signed-in messenger account. Several accounts may share a
// Provider (two Slack workspaces, a work and a personal Telegram): ID is
// what identifies an account everywhere.
type Account struct {
	ID       string `json:"id"`
	Provider string `json:"provider"` // "telegram", "slack", "vkteams"
	Name     string `json:"name"`     // shown to the user
	// Login is who the account signs in as (phone, workspace/user id), for
	// display and to refuse adding the same account twice.
	Login string `json:"login"`
	// Method is the provider's auth method the account signed in with.
	Method string `json:"method,omitempty"`
	// Params are the provider's credentials and settings (tokens, server).
	Params   map[string]string `json:"params"`
	Disabled bool              `json:"disabled,omitempty"`
	Added    time.Time         `json:"added"`
	// Updated is when the account last changed (set by Update): the
	// newest copy of an account wins a merge.
	Updated time.Time `json:"updated,omitempty"`
}

// Key returns the master key.
func (c *Config) Key() crypto.Key {
	var k crypto.Key
	copy(k[:], c.MasterKey)
	return k
}

// Account returns the account with the given id.
func (c *Config) Account(id string) (Account, bool) {
	for _, a := range c.Accounts {
		if a.ID == id {
			return a, true
		}
	}
	return Account{}, false
}

// SameLogin returns the account of provider that signs in as login, if any
// (login compared as stored; providers normalise it).
func (c *Config) SameLogin(provider, login string) (Account, bool) {
	if login == "" {
		return Account{}, false
	}
	for _, a := range c.Accounts {
		if a.Provider == provider && a.Login == login {
			return a, true
		}
	}
	return Account{}, false
}

func (c *Config) clone() *Config {
	data, _ := json.Marshal(c)
	var out Config
	_ = json.Unmarshal(data, &out)
	return &out
}

func (c *Config) validate() error {
	if len(c.MasterKey) != crypto.KeySize {
		return errors.New("vault: bad master key")
	}
	seen := map[string]bool{}
	logins := map[string]bool{}
	for _, a := range c.Accounts {
		if a.ID == "" || seen[a.ID] {
			return fmt.Errorf("vault: missing or duplicate account id %q", a.ID)
		}
		seen[a.ID] = true
		if a.Provider == "" {
			return fmt.Errorf("vault: account %s has no provider", a.ID)
		}
		if a.Login != "" {
			k := a.Provider + "\x00" + a.Login
			if logins[k] {
				return fmt.Errorf("%w: %s %s", ErrSameLogin, a.Provider, a.Login)
			}
			logins[k] = true
		}
	}
	return nil
}

// New returns a fresh config with a random vault id and master key.
func New() (*Config, error) {
	k, err := crypto.NewKey()
	if err != nil {
		return nil, err
	}
	c := &Config{VaultID: crypto.RandomID(16), Version: 1, MasterKey: k[:]}
	return c, c.validate()
}

// sealer is a password-derived key together with the header it belongs to.
type sealer struct {
	header []byte
	key    crypto.Key
}

func newSealer(password []byte, p KDFParams) (*sealer, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	h := make([]byte, 0, headerSize)
	h = append(h, magic...)
	h = binary.BigEndian.AppendUint32(h, p.Time)
	h = binary.BigEndian.AppendUint32(h, p.Memory)
	h = append(h, p.Threads)
	h = append(h, salt...)
	return &sealer{header: h, key: deriveKey(password, p, salt)}, nil
}

func deriveKey(password []byte, p KDFParams, salt []byte) crypto.Key {
	var k crypto.Key
	copy(k[:], argon2.IDKey(password, salt, p.Time, p.Memory, p.Threads, crypto.KeySize))
	return k
}

func parseHeader(data []byte) (KDFParams, []byte, error) {
	if len(data) < headerSize || !bytes.Equal(data[:8], magic) {
		return KDFParams{}, nil, ErrFormat
	}
	p := KDFParams{
		Time:    binary.BigEndian.Uint32(data[8:12]),
		Memory:  binary.BigEndian.Uint32(data[12:16]),
		Threads: data[16],
	}
	return p, data[17:headerSize], nil
}

func (s *sealer) seal(c *Config) ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	ct, err := crypto.Seal(s.key, js, s.header)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, s.header...), ct...), nil
}

func openWithKey(k crypto.Key, data []byte) (*Config, error) {
	js, err := crypto.Open(k, data[headerSize:], data[:headerSize])
	if err != nil {
		return nil, ErrBadPassword
	}
	var c Config
	if err := json.Unmarshal(js, &c); err != nil {
		return nil, fmt.Errorf("vault: %w", err)
	}
	return &c, c.validate()
}

// Vault is an unlocked vault file. It is safe for concurrent use.
type Vault struct {
	mu     sync.Mutex
	path   string
	sealer *sealer
	cfg    *Config
	data   []byte // the file as last read or written

	hookMu  sync.Mutex
	onWrite func()
}

// OnWrite sets fn to be called after every change this vault writes
// (Update, ChangePassword, Merge; not Adopt, which takes a copy from
// elsewhere). fn runs on the writer's goroutine, outside the vault's lock.
func (v *Vault) OnWrite(fn func()) {
	v.hookMu.Lock()
	v.onWrite = fn
	v.hookMu.Unlock()
}

func (v *Vault) wrote() {
	v.hookMu.Lock()
	fn := v.onWrite
	v.hookMu.Unlock()
	if fn != nil {
		fn()
	}
}

// Bytes returns the vault file as this vault last read or wrote it.
func (v *Vault) Bytes() []byte {
	v.mu.Lock()
	defer v.mu.Unlock()
	return append([]byte{}, v.data...)
}

// Create writes a new vault file. It fails if the file exists.
func Create(path string, password []byte, c *Config, p KDFParams) (*Vault, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("vault: %s already exists", path)
	}
	s, err := newSealer(password, p)
	if err != nil {
		return nil, err
	}
	v := &Vault{path: path, sealer: s, cfg: c.clone()}
	return v, v.write()
}

// Open unlocks an existing vault file.
func Open(path string, password []byte) (*Vault, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, salt, err := parseHeader(data)
	if err != nil {
		return nil, err
	}
	s := &sealer{header: append([]byte{}, data[:headerSize]...), key: deriveKey(password, p, salt)}
	c, err := openWithKey(s.key, data)
	if err != nil {
		return nil, err
	}
	return &Vault{path: path, sealer: s, cfg: c, data: data}, nil
}

// OpenWithKey unlocks an existing vault file with a key from UnlockKey.
// It fails with ErrBadPassword once the password has changed.
func OpenWithKey(path string, k crypto.Key) (*Vault, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if _, _, err := parseHeader(data); err != nil {
		return nil, err
	}
	c, err := openWithKey(k, data)
	if err != nil {
		return nil, err
	}
	s := &sealer{header: append([]byte{}, data[:headerSize]...), key: k}
	return &Vault{path: path, sealer: s, cfg: c, data: data}, nil
}

// UnlockKey is the key derived from the password for the file as it is
// now: OpenWithKey opens the vault with it until the password changes.
// It is what "remember the password on this machine" keeps, so the
// password itself is never stored.
func (v *Vault) UnlockKey() crypto.Key {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.sealer.key
}

// Path returns the vault file path.
func (v *Vault) Path() string { return v.path }

// Config returns a copy of the current config.
func (v *Vault) Config() *Config {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.cfg.clone()
}

// Update applies fn to a copy of the config, bumps the version and saves.
// If fn returns an error nothing changes. Accounts fn added or changed get
// Updated, removed ones a tombstone.
func (v *Vault) Update(fn func(*Config) error) error {
	if err := v.update(fn); err != nil {
		return err
	}
	v.wrote()
	return nil
}

func (v *Vault) update(fn func(*Config) error) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	next := v.cfg.clone()
	if err := fn(next); err != nil {
		return err
	}
	next.Version = v.cfg.Version + 1
	stamp(v.cfg, next, time.Now().UTC())
	if err := next.validate(); err != nil {
		return err
	}
	prev := v.cfg
	v.cfg = next
	if err := v.write(); err != nil {
		v.cfg = prev
		return err
	}
	return nil
}

// stamp sets Updated on the accounts that differ from prev and adds
// tombstones for the ones that are gone.
func stamp(prev, next *Config, now time.Time) {
	old := make(map[string]Account, len(prev.Accounts))
	for _, a := range prev.Accounts {
		old[a.ID] = a
	}
	for i := range next.Accounts {
		a := &next.Accounts[i]
		if o, ok := old[a.ID]; !ok || !sameAccount(o, *a) {
			a.Updated = now
		}
		delete(old, a.ID)
	}
	for id := range old {
		next.Removed = append(next.Removed, Removed{ID: id, At: now})
	}
	slices.SortFunc(next.Removed, func(x, y Removed) int { return strings.Compare(x.ID, y.ID) })
}

// sameAccount compares accounts apart from Updated.
func sameAccount(a, b Account) bool {
	a.Updated, b.Updated = time.Time{}, time.Time{}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return bytes.Equal(ja, jb)
}

// ChangePassword re-encrypts the vault with a new password and fresh salt.
func (v *Vault) ChangePassword(password []byte, p KDFParams) error {
	s, err := newSealer(password, p)
	if err != nil {
		return err
	}
	v.mu.Lock()
	prevSealer, prevCfg := v.sealer, v.cfg
	next := v.cfg.clone()
	next.Version++
	next.PasswordChanged = time.Now().UTC()
	v.sealer, v.cfg = s, next
	if err := v.write(); err != nil {
		v.sealer, v.cfg = prevSealer, prevCfg
		v.mu.Unlock()
		return err
	}
	v.mu.Unlock()
	v.wrote()
	return nil
}

// MoveTo writes the vault to path, keeps it there from now on and removes
// the old file. It fails if path exists; on failure the vault stays put.
func (v *Vault) MoveTo(path string) error {
	if err := v.moveTo(path); err != nil {
		return err
	}
	v.wrote() // sealed anew: the file's bytes changed
	return nil
}

func (v *Vault) moveTo(path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("vault: %s already exists", path)
	}
	old := v.path
	v.path = path
	if err := v.write(); err != nil {
		v.path = old
		return err
	}
	if err := os.Remove(old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("vault: moved, but the old file stays: %w", err)
	}
	return nil
}

// write saves atomically: temp file in the same directory, then rename.
func (v *Vault) write() error {
	data, err := v.sealer.seal(v.cfg)
	if err != nil {
		return err
	}
	if err := writeFile(v.path, data); err != nil {
		return err
	}
	v.data = data
	return nil
}

// writeFile writes data to path atomically, readable by the user only.
func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".vault-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
