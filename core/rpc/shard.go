package rpc

// "Sync via Shard" (CONNECTOR.md): the rpc service is the vault side
// of the shard.Syncer. The vault file is synced as it is (sealed with
// the password), so Shard's copy opens only with that password.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SelfHold/parley/core/crypto"
	"github.com/SelfHold/parley/core/eventlog"
	"github.com/SelfHold/parley/core/rpc/pb"
	"github.com/SelfHold/parley/core/shard"
	"github.com/SelfHold/parley/core/vault"
)

// SyncHost returns the vault side of a shard.Syncer for this service;
// set the syncer as Sync afterwards.
func (s *Service) SyncHost() shard.Host { return syncHost{s} }

// syncChanged publishes the sync state.
func (s *Service) syncChanged() { s.publish(&pb.Event{Kind: pb.Event_KIND_SHARD}) }

// SyncChanged is the syncer's OnChange.
func (s *Service) SyncChanged() { s.syncChanged() }

func (s *Service) syncKick() {
	if s.Sync != nil {
		s.Sync.Kick()
	}
}

func shardToPB(st shard.State) *pb.ShardState {
	return &pb.ShardState{
		Status: pb.ShardState_Status(st.Status), Enabled: st.Enabled, Available: st.Available,
		PairingCode: st.PairingCode, Message: st.Message, SyncedMs: st.SyncedMs, DeviceName: st.DeviceName,
		CanMerge: st.CanMerge, Unpaired: st.Unpaired, CanOpen: st.CanOpen,
	}
}

func (s *Service) syncState() *pb.ShardState {
	if s.Sync == nil {
		return nil
	}
	return shardToPB(s.Sync.State())
}

func (s *Service) syncer() (*shard.Syncer, error) {
	if s.Sync == nil {
		return nil, status.Error(codes.Unimplemented, "Shard is not available here")
	}
	return s.Sync, nil
}

func (s *Service) Shard(context.Context, *pb.Empty) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	return shardToPB(sy.State()), nil
}

func (s *Service) EnableShard(context.Context, *pb.Empty) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	st, err := sy.Enable()
	if err != nil {
		return nil, err
	}
	return shardToPB(st), nil
}

func (s *Service) CancelShardPairing(context.Context, *pb.Empty) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	return shardToPB(sy.CancelPairing()), nil
}

func (s *Service) ResolveShard(_ context.Context, r *pb.ResolveShardRequest) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	st, err := sy.Resolve(shard.Choice(r.Choice))
	if err != nil {
		return nil, err
	}
	return shardToPB(st), nil
}

func (s *Service) ShardPassword(ctx context.Context, r *pb.PasswordRequest) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	st, err := sy.Password(ctx, r.Password)
	if err != nil {
		return nil, err
	}
	return shardToPB(st), nil
}

func (s *Service) DisableShard(ctx context.Context, r *pb.DisableShardRequest) (*pb.ShardState, error) {
	sy, err := s.syncer()
	if err != nil {
		return nil, err
	}
	st, err := sy.Disable(ctx, r.RemoveSynced)
	if err != nil {
		return nil, toStatus(err)
	}
	return shardToPB(st), nil
}

// --- the vault side ----------------------------------------------------------

type syncHost struct{ s *Service }

func (h syncHost) open() (*vault.Vault, bool) {
	h.s.mu.Lock()
	defer h.s.mu.Unlock()
	return h.s.v, h.s.v != nil
}

func (h syncHost) LocalVault() ([]byte, error) {
	if v, ok := h.open(); ok {
		return v.Bytes(), nil
	}
	b, err := os.ReadFile(h.s.vaultPath())
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return b, err
}

func (h syncHost) CanEnable() error {
	h.s.mu.Lock()
	unlocked := h.s.v != nil
	h.s.mu.Unlock()
	if unlocked {
		return nil
	}
	if _, err := os.Stat(h.s.vaultPath()); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return status.Error(codes.FailedPrecondition, "unlock the vault first")
}

func syncErr(err error) error {
	switch {
	case errors.Is(err, vault.ErrChanged):
		return shard.ErrLocalChanged
	case errors.Is(err, vault.ErrBadPassword):
		return shard.ErrNeedPassword
	case errors.Is(err, vault.ErrOtherVault):
		return shard.ErrOtherVault
	}
	return err
}

func (h syncHost) AdoptVault(data []byte, expect string, passwords [][]byte) error {
	s := h.s
	s.vaultOp.Lock()
	defer s.vaultOp.Unlock()
	s.mu.Lock()
	v, hb := s.v, s.h
	s.mu.Unlock()
	if v == nil {
		if err := vault.ReplaceFile(s.vaultPath(), data, expect); err != nil {
			return syncErr(err)
		}
		s.publish(&pb.Event{Kind: pb.Event_KIND_STATE}) // perhaps a vault where there was none
		return nil
	}
	prev, key := v.Config(), v.UnlockKey()
	if err := v.Adopt(data, expect, passwords); err != nil {
		return syncErr(err)
	}
	s.applied(v, key)
	hb.Reload(prev)
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return nil
}

func (h syncHost) MergeVault(copies [][]byte, passwords [][]byte) ([]byte, error) {
	s := h.s
	s.vaultOp.Lock()
	defer s.vaultOp.Unlock()
	s.mu.Lock()
	v, hb := s.v, s.h
	s.mu.Unlock()
	if v == nil {
		return nil, shard.ErrLocked
	}
	prev, key := v.Config(), v.UnlockKey()
	data, err := v.MergeIn(copies, passwords)
	if err != nil {
		return nil, syncErr(err)
	}
	s.applied(v, key)
	hb.Reload(prev)
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return data, nil
}

func (h syncHost) CompareVault(data []byte, passwords [][]byte) (bool, bool, error) {
	v, ok := h.open()
	if !ok {
		return false, false, shard.ErrLocked
	}
	same, sameVault, err := v.Compare(data, passwords)
	return same, sameVault, syncErr(err)
}

func (h syncHost) UseSynced(data []byte, expect string, passwords [][]byte) error {
	s := h.s
	if err := vault.CheckFile(data); err != nil {
		return err
	}
	if local, err := h.LocalVault(); err == nil && local != nil {
		backup := fmt.Sprintf("%s.backup-%s", s.vaultPath(), time.Now().Format("20060102-150405"))
		if err := os.WriteFile(backup, local, 0o600); err != nil {
			return fmt.Errorf("keep a backup of the vault: %w", err)
		}
		s.Log.Addf(eventlog.Info, "shard", "", "this device's vault is kept in %s", backup)
	}
	err := h.AdoptVault(data, expect, passwords)
	if !errors.Is(err, shard.ErrOtherVault) && !errors.Is(err, shard.ErrNeedPassword) {
		return err
	}
	// Another vault, or another password: lock and take the file as it
	// is; it opens with its own password.
	s.Close()
	s.vaultOp.Lock()
	defer s.vaultOp.Unlock()
	if err := vault.ReplaceFile(s.vaultPath(), data, expect); err != nil {
		return syncErr(err)
	}
	s.Log.Add(eventlog.Info, "shard", "", "took the synced vault: unlock it with its password")
	s.publish(&pb.Event{Kind: pb.Event_KIND_STATE})
	return nil
}

// applied follows a vault that changed from elsewhere: a remembered key
// is kept up to date when the password changed with it.
func (s *Service) applied(v *vault.Vault, key crypto.Key) {
	if v.UnlockKey() == key || s.Keychain == nil {
		return
	}
	if _, err := os.Stat(s.rememberedMark()); err != nil {
		return
	}
	if err := s.remember(v.UnlockKey()); err != nil {
		s.Log.Addf(eventlog.Warn, "vault", "", "the synced password is not remembered: %v", err)
		s.forget()
	}
}
