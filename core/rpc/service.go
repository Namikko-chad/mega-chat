// Package rpc serves the Parley gRPC API over a hub. The same service
// runs inside the app's process on desktops and on Android (core/embed).
package rpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/SelfHold/parley/core/crypto"
	"github.com/SelfHold/parley/core/eventlog"
	"github.com/SelfHold/parley/core/hub"
	"github.com/SelfHold/parley/core/keychain"
	"github.com/SelfHold/parley/core/linkpreview"
	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/rpc/pb"
	"github.com/SelfHold/parley/core/shard"
	"github.com/SelfHold/parley/core/store"
	"github.com/SelfHold/parley/core/vault"
)

// Service implements pb.ParleyServer.
type Service struct {
	pb.UnimplementedParleyServer
	VaultPath string // read through vaultPath(): SetVaultPath changes it
	// VaultPointer keeps a vault path set by SetVaultPath for the next run;
	// "" keeps it for this run only.
	VaultPointer string
	DataRoot     string // the cache lives here
	Platform     string
	Registry     *provider.Registry
	// Log keeps recent events for the app (connection failures, ...).
	Log *eventlog.Log
	// KDF for new vaults and password changes.
	KDF vault.KDFParams
	// Keychain remembers the vault's unlock key on this machine; nil: the
	// platform has none (can_remember is false).
	Keychain keychain.Keychain
	// Sync keeps the vault in Shard (shard.go); nil: not here.
	Sync *shard.Syncer

	pathMu sync.Mutex
	// vaultOp keeps a vault taken from Shard from meeting a lock.
	vaultOp sync.Mutex

	previewOnce sync.Once
	previews    *linkpreview.Fetcher

	mu    sync.Mutex
	v     *vault.Vault
	st    *store.Store
	h     *hub.Hub
	unsub func()

	subMu sync.Mutex
	subs  map[*subscriber]struct{}
}

type subscriber struct {
	ch   chan *pb.Event
	full chan struct{} // closed when ch overflowed: the client must reload
	once sync.Once
}

// NewService returns a locked service for the vault file at vaultPath.
func NewService(vaultPath, dataRoot, platform string, reg *provider.Registry) *Service {
	return &Service{
		VaultPath: vaultPath, DataRoot: dataRoot, Platform: platform, Registry: reg,
		Log: eventlog.New(1000, nil), KDF: vault.DefaultKDF,
		subs: map[*subscriber]struct{}{},
	}
}

// TokenAuth returns interceptors that require metadata "authorization:
// Bearer <token>", so only processes that can read the token file connect.
func TokenAuth(token string) []grpc.ServerOption {
	check := func(ctx context.Context) error {
		md, _ := metadata.FromIncomingContext(ctx)
		for _, v := range md.Get("authorization") {
			if v == "Bearer "+token {
				return nil
			}
		}
		return status.Error(codes.Unauthenticated, "bad token")
	}
	return []grpc.ServerOption{
		// A Send with files carries them whole (hub.MaxUpload).
		grpc.MaxRecvMsgSize(hub.MaxUpload + 1<<20),
		grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
			if err := check(ctx); err != nil {
				return nil, err
			}
			return h(ctx, req)
		}),
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, h grpc.StreamHandler) error {
			if err := check(ss.Context()); err != nil {
				return err
			}
			return h(srv, ss)
		}),
	}
}

// Subscribe calls fn for every event until the returned function is
// called (the Android binding uses it for notifications).
func (s *Service) Subscribe(fn func(*pb.Event)) (cancel func()) {
	sub := s.addSub()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case ev := <-sub.ch:
				fn(ev)
			case <-done:
				return
			}
		}
	}()
	return func() {
		s.removeSub(sub)
		close(done)
	}
}

func (s *Service) addSub() *subscriber {
	sub := &subscriber{ch: make(chan *pb.Event, 1024), full: make(chan struct{})}
	s.subMu.Lock()
	s.subs[sub] = struct{}{}
	s.subMu.Unlock()
	return sub
}

func (s *Service) removeSub(sub *subscriber) {
	s.subMu.Lock()
	delete(s.subs, sub)
	s.subMu.Unlock()
}

// publish never blocks: a client too slow to keep up is cut off and
// reloads when it reconnects, rather than silently missing events.
func (s *Service) publish(ev *pb.Event) {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	for sub := range s.subs {
		select {
		case sub.ch <- ev:
		default:
			sub.once.Do(func() { close(sub.full) })
		}
	}
}

func (s *Service) hub() (*hub.Hub, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.h == nil {
		return nil, status.Error(codes.FailedPrecondition, "vault is locked")
	}
	return s.h, nil
}

func toStatus(err error) error {
	switch {
	case err == nil:
		return nil
	case status.Code(err) != codes.Unknown:
		return err
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, hub.ErrNoAccount), errors.Is(err, hub.ErrNoLogin), errors.Is(err, hub.ErrNoProvider):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, vault.ErrSameLogin), errors.Is(err, fs.ErrExist):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, vault.ErrBadPassword):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, provider.ErrOffline):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, provider.ErrAuth), errors.Is(err, store.ErrLocked):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, provider.ErrUnsupported):
		return status.Error(codes.Unimplemented, err.Error())
	case errors.Is(err, hub.ErrInCall):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, provider.ErrTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

// --- vault lifecycle -----------------------------------------------------

func (s *Service) state() *pb.State {
	_, err := os.Stat(s.vaultPath())
	st := &pb.State{
		VaultExists: err == nil, VaultPath: s.vaultPath(), Platform: s.Platform,
		CoreVersion: Version, CoreBuild: Build,
		CanRemember: s.Keychain != nil,
		Shard:       s.syncState(),
	}
	if s.Keychain != nil {
		_, err := os.Stat(s.rememberedMark())
		st.Remembered = err == nil
	}
	if s.h != nil {
		st.Unlocked = true
		on, total := s.h.Online()
		st.Accounts, st.AccountsOnline = int32(total), int32(on)
	}
	return st
}

func (s *Service) GetState(context.Context, *pb.Empty) (*pb.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state(), nil
}

func (s *Service) CreateVault(_ context.Context, r *pb.PasswordRequest) (*pb.State, error) {
	if r.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "empty password")
	}
	c, err := vault.New()
	if err != nil {
		return nil, toStatus(err)
	}
	v, err := vault.Create(s.vaultPath(), []byte(r.Password), c, s.KDF)
	if err != nil {
		return nil, toStatus(err)
	}
	s.forget() // left from a vault that was here before
	if s.Sync != nil {
		s.Sync.VaultSwitched()
	}
	return s.open(v)
}

func (s *Service) Unlock(_ context.Context, r *pb.PasswordRequest) (*pb.State, error) {
	s.mu.Lock()
	if s.h != nil {
		defer s.mu.Unlock()
		return s.state(), nil
	}
	s.mu.Unlock()
	v, err := vault.Open(s.vaultPath(), []byte(r.Password))
	if err != nil {
		return nil, toStatus(err)
	}
	return s.open(v)
}

func (s *Service) open(v *vault.Vault) (*pb.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.h != nil {
		return s.state(), nil
	}
	cfg := v.Config()
	st, err := store.Open(filepath.Join(s.DataRoot, "cache-"+cfg.VaultID[:8]+".db"), cfg.Key())
	if err != nil {
		return nil, toStatus(err)
	}
	h := hub.New(v, st, s.Registry, s.Log)
	s.v, s.st, s.h = v, st, h
	v.OnWrite(s.syncKick)
	defer s.syncKick()
	s.unsub = h.Subscribe(func(ev hub.Event) { s.publish(eventToPB(ev)) })
	h.Start()
	s.Log.Add(eventlog.Info, "vault", "", "unlocked")
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return s.state(), nil
}

func (s *Service) LockVault(context.Context, *pb.Empty) (*pb.State, error) {
	s.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state(), nil
}

// Close locks the vault: sessions stop, the cache closes.
func (s *Service) Close() {
	s.vaultOp.Lock()
	defer s.vaultOp.Unlock()
	s.mu.Lock()
	h, st, unsub := s.h, s.st, s.unsub
	s.h, s.st, s.v, s.unsub = nil, nil, nil, nil
	s.mu.Unlock()
	if h == nil {
		return
	}
	h.Close()
	unsub()
	st.Close()
	s.Log.Add(eventlog.Info, "vault", "", "locked")
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
}

func (s *Service) ChangePassword(_ context.Context, r *pb.ChangePasswordRequest) (*pb.Empty, error) {
	s.mu.Lock()
	v := s.v
	s.mu.Unlock()
	if v == nil {
		return nil, status.Error(codes.FailedPrecondition, "vault is locked")
	}
	if r.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "empty password")
	}
	if _, err := vault.Open(s.vaultPath(), []byte(r.OldPassword)); err != nil {
		return nil, toStatus(err)
	}
	if err := v.ChangePassword([]byte(r.NewPassword), s.KDF); err != nil {
		return nil, toStatus(err)
	}
	// The remembered key opened the old password only.
	if _, err := os.Stat(s.rememberedMark()); err == nil && s.Keychain != nil {
		if err := s.remember(v.UnlockKey()); err != nil {
			s.Log.Addf(eventlog.Warn, "vault", "", "the new password is not remembered: %v", err)
			s.forget()
		}
	}
	return &pb.Empty{}, nil
}

func (s *Service) CheckPassword(_ context.Context, r *pb.PasswordRequest) (*pb.Empty, error) {
	if _, err := vault.Open(s.vaultPath(), []byte(r.Password)); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

// --- remembered password --------------------------------------------------

// rememberedMark is a file that says a key is kept in the keychain, so
// GetState does not ask the keychain (it may prompt to unlock itself).
// It is this device's state: it lives in the data dir, never beside a
// vault moved elsewhere (a synced folder would lose it or hand it to
// other devices). The default vault keeps its old name.
func (s *Service) rememberedMark() string {
	p := s.vaultPath()
	if s.DataRoot == "" {
		return p + ".remembered"
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	if def, err := filepath.Abs(filepath.Join(s.DataRoot, "vault.parley")); err == nil && def == abs {
		return filepath.Join(s.DataRoot, "vault.parley.remembered")
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(s.DataRoot, "remembered-"+hex.EncodeToString(sum[:8]))
}

// keychainName names the vault's key in the keychain: one per vault file.
func (s *Service) keychainName() string {
	p, err := filepath.Abs(s.vaultPath())
	if err != nil {
		p = s.vaultPath()
	}
	return "vault:" + p
}

func (s *Service) remember(k crypto.Key) error {
	if err := s.Keychain.Set(s.keychainName(), k[:]); err != nil {
		return err
	}
	return os.WriteFile(s.rememberedMark(), nil, 0o600)
}

// forget drops the remembered key, if any.
func (s *Service) forget() error {
	if s.Keychain == nil {
		return nil
	}
	if err := os.Remove(s.rememberedMark()); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return s.Keychain.Delete(s.keychainName())
}

func (s *Service) noKeychain() error {
	return status.Error(codes.Unimplemented, "this machine has no keychain to remember the password in")
}

func (s *Service) RememberPassword(_ context.Context, r *pb.PasswordRequest) (*pb.State, error) {
	if s.Keychain == nil {
		return nil, s.noKeychain()
	}
	v, err := vault.Open(s.vaultPath(), []byte(r.Password))
	if err != nil {
		return nil, toStatus(err)
	}
	if err := s.remember(v.UnlockKey()); err != nil {
		return nil, status.Errorf(codes.Unavailable, "keychain: %v", err)
	}
	s.Log.Add(eventlog.Info, "vault", "", "password remembered on this machine")
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return s.GetState(context.Background(), &pb.Empty{})
}

func (s *Service) ForgetPassword(context.Context, *pb.Empty) (*pb.State, error) {
	if s.Keychain == nil {
		return nil, s.noKeychain()
	}
	if err := s.forget(); err != nil {
		return nil, status.Errorf(codes.Unavailable, "keychain: %v", err)
	}
	s.Log.Add(eventlog.Info, "vault", "", "remembered password forgotten")
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return s.GetState(context.Background(), &pb.Empty{})
}

func (s *Service) UnlockRemembered(context.Context, *pb.Empty) (*pb.State, error) {
	s.mu.Lock()
	if s.h != nil {
		defer s.mu.Unlock()
		return s.state(), nil
	}
	s.mu.Unlock()
	if s.Keychain == nil {
		return nil, s.noKeychain()
	}
	if _, err := os.Stat(s.rememberedMark()); err != nil {
		return nil, status.Error(codes.NotFound, "no remembered password")
	}
	b, err := s.Keychain.Get(s.keychainName())
	if errors.Is(err, keychain.ErrNotFound) || err == nil && len(b) != crypto.KeySize {
		os.Remove(s.rememberedMark())
		return nil, status.Error(codes.NotFound, "no remembered password")
	}
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "keychain: %v", err)
	}
	var k crypto.Key
	copy(k[:], b)
	v, err := vault.OpenWithKey(s.vaultPath(), k)
	if errors.Is(err, vault.ErrBadPassword) {
		s.forget()
		return nil, status.Error(codes.PermissionDenied, "the remembered password no longer opens the vault")
	}
	if err != nil {
		return nil, toStatus(err)
	}
	return s.open(v)
}

// --- where the vault lives -------------------------------------------------

func (s *Service) vaultPath() string {
	s.pathMu.Lock()
	defer s.pathMu.Unlock()
	return s.VaultPath
}

// setVaultPath switches to path and keeps it for the next run.
func (s *Service) setVaultPath(path string) error {
	s.pathMu.Lock()
	s.VaultPath = path
	s.pathMu.Unlock()
	if s.VaultPointer == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.VaultPointer), 0o700); err != nil {
		return err
	}
	tmp := s.VaultPointer + ".tmp"
	if err := os.WriteFile(tmp, []byte(path+"\n"), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.VaultPointer)
}

// ReadVaultPointer returns the vault path kept by SetVaultPath at pointer,
// or "" when there is none.
func ReadVaultPointer(pointer string) string {
	b, err := os.ReadFile(pointer)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func (s *Service) SetVaultPath(_ context.Context, r *pb.VaultPathRequest) (*pb.State, error) {
	if r.Path == "" || !filepath.IsAbs(r.Path) {
		return nil, status.Error(codes.InvalidArgument, "the vault path must be absolute")
	}
	path := filepath.Clean(r.Path)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = filepath.Join(path, "vault.parley")
	}
	if old, err := filepath.Abs(s.vaultPath()); err == nil && old == path {
		return s.GetState(context.Background(), &pb.Empty{})
	}
	if r.Move {
		if err := s.moveVault(path); err != nil {
			return nil, err
		}
	} else if err := s.useVault(path); err != nil {
		return nil, err
	}
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return s.GetState(context.Background(), &pb.Empty{})
}

// moveVault moves the open vault to path; a remembered key follows it
// (the keychain names it by the file's path).
func (s *Service) moveVault(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.v == nil {
		return status.Error(codes.FailedPrecondition, "vault is locked")
	}
	if _, err := os.Stat(path); err == nil {
		return status.Errorf(codes.AlreadyExists, "%s already exists", path)
	}
	_, err := os.Stat(s.rememberedMark())
	remembered := err == nil && s.Keychain != nil
	if err := s.v.MoveTo(path); err != nil {
		return toStatus(err)
	}
	old := s.vaultPath()
	var forgetErr error
	if remembered {
		forgetErr = s.forget()
	}
	if err := s.setVaultPath(path); err != nil {
		s.Log.Addf(eventlog.Warn, "vault", "", "the new vault place is not kept for the next start: %v", err)
	}
	if remembered {
		if err := s.remember(s.v.UnlockKey()); err != nil {
			s.Log.Addf(eventlog.Warn, "vault", "", "the password is not remembered at the new place: %v", err)
		}
	}
	if forgetErr != nil {
		s.Log.Addf(eventlog.Warn, "vault", "", "the key of the old place stays in the keychain: %v", forgetErr)
	}
	s.Log.Addf(eventlog.Info, "vault", "", "moved from %s to %s", old, path)
	return nil
}

// useVault opens the vault at path from now on (locked only).
func (s *Service) useVault(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.h != nil {
		return status.Error(codes.FailedPrecondition, "lock the vault first")
	}
	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return status.Errorf(codes.NotFound, "no vault file at %s", path)
	}
	if err := s.setVaultPath(path); err != nil {
		return status.Errorf(codes.Internal, "keep the vault place: %v", err)
	}
	s.Log.Addf(eventlog.Info, "vault", "", "using the vault at %s", path)
	if s.Sync != nil {
		s.Sync.VaultSwitched()
	}
	return nil
}

// --- providers and accounts ----------------------------------------------

func (s *Service) Providers(context.Context, *pb.Empty) (*pb.ProvidersReply, error) {
	out := &pb.ProvidersReply{}
	for _, p := range s.Registry.All() {
		out.Providers = append(out.Providers, providerToPB(p.Info()))
	}
	return out, nil
}

func (s *Service) StartLogin(ctx context.Context, r *pb.StartLoginRequest) (*pb.LoginStepReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	res, err := h.StartLogin(ctx, r.Provider, r.Method)
	if err != nil {
		return nil, toStatus(err)
	}
	return loginToPB(res), nil
}

func (s *Service) LoginStep(ctx context.Context, r *pb.LoginStepRequest) (*pb.LoginStepReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	res, err := h.LoginStep(ctx, r.LoginId, r.Values)
	if err != nil {
		return nil, toStatus(err)
	}
	return loginToPB(res), nil
}

func (s *Service) CancelLogin(_ context.Context, r *pb.IdRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	h.CancelLogin(r.Id)
	return &pb.Empty{}, nil
}

func (s *Service) Relogin(ctx context.Context, r *pb.IdRequest) (*pb.LoginStepReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	res, err := h.Relogin(ctx, r.Id)
	if err != nil {
		return nil, toStatus(err)
	}
	return loginToPB(res), nil
}

func (s *Service) Accounts(context.Context, *pb.Empty) (*pb.AccountsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	out := &pb.AccountsReply{}
	for _, a := range h.Accounts() {
		out.Accounts = append(out.Accounts, accountToPB(a))
	}
	return out, nil
}

func (s *Service) RenameAccount(_ context.Context, r *pb.RenameRequest) (*pb.Account, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	a, err := h.Rename(r.Id, r.Name)
	if err != nil {
		return nil, toStatus(err)
	}
	return accountToPB(a), nil
}

func (s *Service) SetAccountEnabled(_ context.Context, r *pb.SetEnabledRequest) (*pb.Account, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	a, err := h.SetEnabled(r.Id, r.Enabled)
	if err != nil {
		return nil, toStatus(err)
	}
	return accountToPB(a), nil
}

func (s *Service) RemoveAccount(ctx context.Context, r *pb.IdRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, toStatus(h.Remove(ctx, r.Id))
}

func (s *Service) ReorderAccounts(_ context.Context, r *pb.ReorderAccountsRequest) (*pb.AccountsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if err := h.Reorder(r.Ids); err != nil {
		return nil, toStatus(err)
	}
	out := &pb.AccountsReply{}
	for _, a := range h.Accounts() {
		out.Accounts = append(out.Accounts, accountToPB(a))
	}
	return out, nil
}

// --- chats and messages --------------------------------------------------

func (s *Service) Chats(_ context.Context, r *pb.ChatsRequest) (*pb.ChatsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	cs, err := h.Chats(r.AccountIds, int(r.Limit))
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.ChatsReply{Chats: make([]*pb.Chat, len(cs))}
	for i := range cs {
		out.Chats[i] = chatToPB(&cs[i])
	}
	return out, nil
}

func (s *Service) RefreshChats(ctx context.Context, r *pb.ChatsRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, toStatus(h.RefreshChats(ctx, r.AccountIds))
}

func (s *Service) History(ctx context.Context, r *pb.HistoryRequest) (*pb.HistoryReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	var ms []model.Message
	more := true
	if r.CachedOnly {
		ms, err = h.CachedHistory(r.AccountId, r.ChatId, int(r.Limit))
	} else {
		ms, more, err = h.History(ctx, r.AccountId, r.ChatId, r.BeforeId, int(r.Limit))
	}
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.HistoryReply{More: more, Messages: make([]*pb.Message, len(ms))}
	for i := range ms {
		out.Messages[i] = messageToPB(&ms[i])
	}
	return out, nil
}

func (s *Service) Send(ctx context.Context, r *pb.SendRequest) (*pb.Message, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if r.Text == "" && len(r.Files) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty message")
	}
	m, err := h.Send(ctx, r.AccountId, r.ChatId, outgoing(r))
	if err != nil {
		return nil, toStatus(err)
	}
	return messageToPB(&m), nil
}

func (s *Service) Forward(ctx context.Context, r *pb.ForwardRequest) (*pb.ForwardReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if len(r.MessageIds) == 0 || r.ToChatId == "" {
		return nil, status.Error(codes.InvalidArgument, "nothing to forward")
	}
	msgs, err := h.Forward(ctx, r.AccountId, r.ChatId, r.MessageIds, r.ToChatId)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.ForwardReply{}
	for i := range msgs {
		out.Messages = append(out.Messages, messageToPB(&msgs[i]))
	}
	return out, nil
}

func (s *Service) MarkRead(ctx context.Context, r *pb.MarkReadRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if r.ThreadId != "" {
		return &pb.Empty{}, toStatus(h.ReadTopic(ctx, r.AccountId, r.ChatId, r.ThreadId, r.MessageId))
	}
	return &pb.Empty{}, toStatus(h.MarkRead(ctx, r.AccountId, r.ChatId, r.MessageId))
}

func (s *Service) Topics(ctx context.Context, r *pb.TopicsRequest) (*pb.TopicsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	ts, err := h.Topics(ctx, r.AccountId, r.ChatId)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.TopicsReply{Topics: make([]*pb.Topic, len(ts))}
	for i := range ts {
		out.Topics[i] = topicToPB(&ts[i])
	}
	return out, nil
}

func (s *Service) Edit(ctx context.Context, r *pb.EditRequest) (*pb.Message, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	m, err := h.Edit(ctx, r.AccountId, r.ChatId, r.ThreadId, r.MessageId, r.Text)
	if err != nil {
		return nil, toStatus(err)
	}
	return messageToPB(&m), nil
}

func (s *Service) Delete(ctx context.Context, r *pb.DeleteRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if err := h.Delete(ctx, r.AccountId, r.ChatId, r.ThreadId, r.MessageIds, r.ForAll); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

func (s *Service) Members(ctx context.Context, r *pb.MembersRequest) (*pb.MembersReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	us, err := h.Members(ctx, r.AccountId, r.ChatId, r.Query, int(r.Limit))
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.MembersReply{}
	for _, u := range us {
		out.Users = append(out.Users, userToPB(u))
	}
	return out, nil
}

func (s *Service) LinkPreview(ctx context.Context, r *pb.LinkPreviewRequest) (*pb.LinkPreviewReply, error) {
	if _, err := s.hub(); err != nil {
		return nil, err
	}
	s.previewOnce.Do(func() { s.previews = linkpreview.New() })
	p, err := s.previews.Fetch(ctx, r.Url)
	if err != nil {
		// Not an error of the app: the page just has no preview.
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.LinkPreviewReply{
		Url: p.URL, SiteName: p.SiteName, Title: p.Title, Description: p.Description,
		Image: p.Image, ImageMime: p.ImageMIME,
	}, nil
}

func (s *Service) Profile(ctx context.Context, r *pb.ProfileRequest) (*pb.ChatProfile, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	p, err := h.Profile(ctx, r.AccountId, r.ChatId)
	if err != nil {
		return nil, toStatus(err)
	}
	return profileToPB(&p), nil
}

func (s *Service) Thread(ctx context.Context, r *pb.ThreadRequest) (*pb.HistoryReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	ms, more, err := h.Thread(ctx, r.AccountId, r.ChatId, r.ThreadId, r.BeforeId, int(r.Limit))
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.HistoryReply{More: more, Messages: make([]*pb.Message, len(ms))}
	for i := range ms {
		out.Messages[i] = messageToPB(&ms[i])
	}
	return out, nil
}

func (s *Service) React(ctx context.Context, r *pb.ReactRequest) (*pb.Message, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if r.Emoji == "" {
		return nil, status.Error(codes.InvalidArgument, "no emoji")
	}
	m, err := h.React(ctx, r.AccountId, r.ChatId, r.MessageId, r.Emoji, r.On)
	if err != nil {
		return nil, toStatus(err)
	}
	return messageToPB(&m), nil
}

func (s *Service) Media(ctx context.Context, r *pb.MediaRequest) (*pb.MediaReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	data, mime, err := h.Media(ctx, r.AccountId, r.Ref, r.MaxBytes)
	if err != nil {
		return nil, toStatus(err)
	}
	if r.Prefetch {
		return &pb.MediaReply{Mime: mime}, nil
	}
	return &pb.MediaReply{Data: data, Mime: mime}, nil
}

func (s *Service) StartCall(ctx context.Context, r *pb.StartCallRequest) (*pb.Call, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	c, err := h.StartCall(ctx, r.AccountId, r.ChatId, r.Video)
	if err != nil {
		return nil, toStatus(err)
	}
	return callToPB(r.AccountId, c), nil
}

func (s *Service) AcceptCall(ctx context.Context, r *pb.CallRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if err := h.AcceptCall(ctx, r.AccountId, r.CallId, r.Video); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

func (s *Service) EndCall(ctx context.Context, r *pb.EndCallRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if err := h.EndCall(ctx, r.AccountId, r.CallId, model.CallEnd(r.Reason)); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

func (s *Service) CallSignal(ctx context.Context, r *pb.CallSignalRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if r.Signal == nil {
		return nil, status.Error(codes.InvalidArgument, "no signal")
	}
	if err := h.CallSignal(ctx, r.AccountId, r.CallId, signal(r.Signal)); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

func (s *Service) SetCallMedia(ctx context.Context, r *pb.CallMediaRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if err := h.SetCallMedia(ctx, r.AccountId, r.CallId, callMedia(r.Media)); err != nil {
		return nil, toStatus(err)
	}
	return &pb.Empty{}, nil
}

func (s *Service) Calls(context.Context, *pb.Empty) (*pb.CallsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	out := &pb.CallsReply{}
	for _, c := range h.Calls() {
		out.Calls = append(out.Calls, callToPB(c.AccountID, c.Call))
	}
	return out, nil
}

func (s *Service) CallSignals(_ context.Context, r *pb.CallSignalsRequest) (*pb.CallSignalsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	out := &pb.CallSignalsReply{}
	for _, sig := range h.CallSignals(r.AccountId, r.CallId, r.After) {
		out.Signals = append(out.Signals, signalToPB(sig))
	}
	return out, nil
}

func (s *Service) Watch(_ *pb.Empty, stream grpc.ServerStreamingServer[pb.Event]) error {
	sub := s.addSub()
	defer s.removeSub(sub)
	for {
		select {
		case ev := <-sub.ch:
			if err := stream.Send(ev); err != nil {
				return err
			}
		case <-sub.full:
			return status.Error(codes.ResourceExhausted, "too many events: reload")
		case <-stream.Context().Done():
			return nil
		}
	}
}

// --- logs ----------------------------------------------------------------

func (s *Service) Logs(_ context.Context, r *pb.LogsRequest) (*pb.LogsReply, error) {
	out := &pb.LogsReply{}
	for _, e := range s.Log.Entries(r.AccountId, int(r.Limit)) {
		out.Entries = append(out.Entries, &pb.LogEntry{
			Seq: e.Seq, TimeMs: e.Time.UnixMilli(), FirstMs: e.First.UnixMilli(),
			Level: e.Level, Source: e.Source, AccountId: e.AccountID, Message: e.Message, Count: int32(e.Count),
		})
	}
	return out, nil
}

func (s *Service) ClearLogs(context.Context, *pb.Empty) (*pb.Empty, error) {
	s.Log.Clear()
	return &pb.Empty{}, nil
}
