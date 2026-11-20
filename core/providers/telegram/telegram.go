// Package telegram is the Telegram provider: a user account over MTProto
// (gotd/td), signed in with phone, code and the 2FA password.
//
// Chat ids: "u<user id>" for dialogs, "g<chat id>" for basic groups,
// "c<channel id>" for supergroups and channels. Message ids are Telegram's.
// Access hashes (needed to address users and channels) are kept in the
// account's KV, as is the MTProto session and the updates state, so a
// restart resumes with getDifference instead of losing messages.
package telegram

import (
	"context"
	"encoding/json"
	"runtime"
	"strconv"
	"sync"

	tdsession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"

	"megachat/core/provider"
)

// Type is the provider type.
const Type = "telegram"

// DefaultAppID and DefaultAppHash are the app credentials from
// my.telegram.org, set at build time (-ldflags -X). When empty, the login
// asks for them.
var (
	DefaultAppID   string
	DefaultAppHash string
	// AppVersion is reported to Telegram as the client version.
	AppVersion = "dev"
)

const (
	keySession = "session"
	keyUpdates = "updates"
	keyPeers   = "peers"
)

// Provider is the Telegram provider.
type Provider struct{}

// New returns the provider.
func New() *Provider { return &Provider{} }

func (*Provider) Info() provider.Info {
	return provider.Info{
		Type: Type, Name: "Telegram",
		Methods: []provider.AuthMethod{
			{ID: methodQR, Name: "QR code", Description: "Scan with Telegram on your phone: Settings → Devices → Link Desktop Device"},
			{ID: methodPhone, Name: "Phone number", Description: "Code from Telegram, then the cloud password if set"},
		},
		Caps: []provider.Capability{provider.CapMessages},
	}
}

// appCredentials returns the app id/hash of an account: its own, if it
// was signed in with custom ones, else the built-in ones.
func appCredentials(params map[string]string) (int, string, bool) {
	idStr, hash := params["api_id"], params["api_hash"]
	if idStr == "" || hash == "" {
		idStr, hash = DefaultAppID, DefaultAppHash
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 || hash == "" {
		return 0, "", false
	}
	return id, hash, true
}

func newClient(appID int, appHash string, st tdsession.Storage, h telegram.UpdateHandler) *telegram.Client {
	return telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: st,
		UpdateHandler:  h,
		NoUpdates:      h == nil,
		Device: telegram.DeviceConfig{
			DeviceModel:   "MegaChat",
			SystemVersion: runtime.GOOS,
			AppVersion:    AppVersion,
		},
	})
}

// kvSession keeps the MTProto session (auth key, DC) in a KV.
type kvSession struct{ kv provider.KV }

func (s kvSession) LoadSession(context.Context) ([]byte, error) {
	b, err := s.kv.Get(keySession)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, tdsession.ErrNotFound
	}
	return b, nil
}

func (s kvSession) StoreSession(_ context.Context, data []byte) error {
	return s.kv.Put(keySession, data)
}

// stateStore is the updates manager's state (pts, qts, date, seq and the
// pts of every channel), saved as one record on every change.
type stateStore struct {
	mu   sync.Mutex
	kv   provider.KV
	data struct {
		Has      bool          `json:"has"`
		State    updates.State `json:"state"`
		Channels map[int64]int `json:"channels"`
	}
}

func newStateStore(kv provider.KV) (*stateStore, error) {
	s := &stateStore{kv: kv}
	b, err := kv.Get(keyUpdates)
	if err != nil {
		return nil, err
	}
	if b != nil {
		_ = json.Unmarshal(b, &s.data) // a broken record: start from the server state
	}
	if s.data.Channels == nil {
		s.data.Channels = map[int64]int{}
	}
	return s, nil
}

func (s *stateStore) change(fn func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
	b, err := json.Marshal(&s.data)
	if err != nil {
		return err
	}
	return s.kv.Put(keyUpdates, b)
}

func (s *stateStore) GetState(context.Context, int64) (updates.State, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.State, s.data.Has, nil
}

func (s *stateStore) SetState(_ context.Context, _ int64, st updates.State) error {
	return s.change(func() { s.data.State, s.data.Has = st, true })
}

func (s *stateStore) SetPts(_ context.Context, _ int64, v int) error {
	return s.change(func() { s.data.State.Pts = v })
}

func (s *stateStore) SetQts(_ context.Context, _ int64, v int) error {
	return s.change(func() { s.data.State.Qts = v })
}

func (s *stateStore) SetDate(_ context.Context, _ int64, v int) error {
	return s.change(func() { s.data.State.Date = v })
}

func (s *stateStore) SetSeq(_ context.Context, _ int64, v int) error {
	return s.change(func() { s.data.State.Seq = v })
}

func (s *stateStore) SetDateSeq(_ context.Context, _ int64, date, seq int) error {
	return s.change(func() { s.data.State.Date, s.data.State.Seq = date, seq })
}

func (s *stateStore) GetChannelPts(_ context.Context, _, ch int64) (int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data.Channels[ch]
	return v, ok, nil
}

func (s *stateStore) SetChannelPts(_ context.Context, _, ch int64, pts int) error {
	return s.change(func() { s.data.Channels[ch] = pts })
}

func (s *stateStore) ForEachChannels(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
	s.mu.Lock()
	chans := make(map[int64]int, len(s.data.Channels))
	for k, v := range s.data.Channels {
		chans[k] = v
	}
	s.mu.Unlock()
	for ch, pts := range chans {
		if err := f(ctx, ch, pts); err != nil {
			return err
		}
	}
	return nil
}

// peerStore keeps access hashes ("u<id>", "c<id>" → hash). It is saved
// in batches (Save) rather than per peer: a chat list brings hundreds.
type peerStore struct {
	mu    sync.Mutex
	kv    provider.KV
	m     map[string]int64
	dirty bool
}

func newPeerStore(kv provider.KV) (*peerStore, error) {
	p := &peerStore{kv: kv, m: map[string]int64{}}
	b, err := kv.Get(keyPeers)
	if err != nil {
		return nil, err
	}
	if b != nil {
		_ = json.Unmarshal(b, &p.m)
	}
	return p, nil
}

func (p *peerStore) set(key string, hash int64) {
	p.mu.Lock()
	if p.m[key] != hash {
		p.m[key], p.dirty = hash, true
	}
	p.mu.Unlock()
}

func (p *peerStore) get(key string) (int64, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	h, ok := p.m[key]
	return h, ok
}

// Save writes the hashes if they changed.
func (p *peerStore) Save() error {
	p.mu.Lock()
	if !p.dirty {
		p.mu.Unlock()
		return nil
	}
	b, err := json.Marshal(p.m)
	p.dirty = false
	p.mu.Unlock()
	if err != nil {
		return err
	}
	return p.kv.Put(keyPeers, b)
}

func (p *peerStore) SetChannelAccessHash(_ context.Context, _, ch, hash int64) error {
	p.set("c"+strconv.FormatInt(ch, 10), hash)
	return p.Save()
}

func (p *peerStore) GetChannelAccessHash(_ context.Context, _, ch int64) (int64, bool, error) {
	h, ok := p.get("c" + strconv.FormatInt(ch, 10))
	return h, ok, nil
}

func (p *peerStore) SetUserAccessHash(_ context.Context, _, u, hash int64) error {
	p.set("u"+strconv.FormatInt(u, 10), hash)
	return p.Save()
}

func (p *peerStore) GetUserAccessHash(_ context.Context, _, u int64) (int64, bool, error) {
	h, ok := p.get("u" + strconv.FormatInt(u, 10))
	return h, ok, nil
}
