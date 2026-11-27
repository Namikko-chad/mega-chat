package rpc

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/SelfHold/parley/core/keychain"
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/providers/fake"
	"github.com/SelfHold/parley/core/rpc/pb"
	"github.com/SelfHold/parley/core/shard"
	"github.com/SelfHold/parley/core/shard/fakeshard"
	"github.com/SelfHold/parley/core/vault"
)

// device is one Parley with sync, on a Shard shared by the test.
func device(t *testing.T, cs *fakeshard.Server, name string) *Service {
	t.Helper()
	dir := t.TempDir()
	svc := NewService(filepath.Join(dir, "vault.parley"), dir, "test", provider.NewRegistry(fake.New()))
	svc.KDF = vault.KDFParams{Time: 1, Memory: 1024, Threads: 1}
	svc.Keychain = keychain.NewMemory()
	sy := shard.New(shard.Options{
		Dialer: cs.Dialer(name), Host: svc.SyncHost(), Dir: dir, Keychain: svc.Keychain,
		AppName: "Parley", OnChange: svc.SyncChanged,
	})
	sy.RetryDelay, sy.ProbeDelay, sy.Heartbeat = 50*time.Millisecond, 50*time.Millisecond, 100*time.Millisecond
	svc.Sync = sy
	sy.Start()
	t.Cleanup(func() { sy.Close(); svc.Close() })
	return svc
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out: %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func addAccount(t *testing.T, s *Service, name string) string {
	t.Helper()
	ctx := context.Background()
	step, err := s.StartLogin(ctx, &pb.StartLoginRequest{Provider: fake.Type})
	if err != nil {
		t.Fatal(err)
	}
	done, err := s.LoginStep(ctx, &pb.LoginStepRequest{LoginId: step.LoginId, Values: map[string]string{"name": name}})
	if err != nil || done.Account == nil {
		t.Fatalf("login %s: %v %v", name, done, err)
	}
	return done.Account.Id
}

func names(s *Service) []string {
	s.mu.Lock()
	v := s.v
	s.mu.Unlock()
	if v == nil {
		return nil
	}
	var out []string
	for _, a := range v.Config().Accounts {
		out = append(out, a.Name)
	}
	slices.Sort(out)
	return out
}

func syncStatus(s *Service) shard.Status { return s.Sync.State().Status }

func TestShardTwoDevices(t *testing.T) {
	ctx := context.Background()
	cs := fakeshard.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "phone")

	if _, err := a.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	addAccount(t, a, "zoe")
	if _, err := a.EnableShard(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a seeds Shard", func() bool {
		return syncStatus(a) == shard.StatusSynced && len(cs.Versions(shard.RecordKey)) == 1
	})

	// The phone has no vault: it takes the synced one and opens it with
	// the same password.
	if _, err := b.EnableShard(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b gets the vault", func() bool {
		st, _ := b.GetState(ctx, &pb.Empty{})
		return st.VaultExists && st.Shard.GetStatus() == pb.ShardState_SYNCED
	})
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if got := names(b); !slices.Equal(got, []string{"Demo zoe"}) {
		t.Fatalf("b accounts: %v", got)
	}

	// A change on either device reaches the other.
	addAccount(t, b, "max")
	eventually(t, "a sees max", func() bool { return slices.Equal(names(a), []string{"Demo max", "Demo zoe"}) })
	if _, err := a.RenameAccount(ctx, &pb.RenameRequest{Id: a.v.Config().Accounts[0].ID, Name: "zoe2"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b sees the rename", func() bool { return slices.Equal(names(b), []string{"Demo max", "zoe2"}) })

	// A removal is not undone by a copy made before it (offline device):
	// the merge keeps the tombstone.
	old := b.v.Bytes()
	maxID := ""
	for _, x := range a.v.Config().Accounts {
		if x.Name == "Demo max" {
			maxID = x.ID
		}
	}
	if _, err := a.RemoveAccount(ctx, &pb.IdRequest{Id: maxID}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b drops max", func() bool { return slices.Equal(names(b), []string{"zoe2"}) })
	cs.Concurrent(shard.RecordKey, old, "tablet")
	eventually(t, "the copies are merged", func() bool {
		return len(cs.Versions(shard.RecordKey)) == 1 && syncStatus(a) == shard.StatusSynced && syncStatus(b) == shard.StatusSynced
	})
	eventually(t, "max stays removed", func() bool {
		return slices.Equal(names(a), []string{"zoe2"}) && slices.Equal(names(b), []string{"zoe2"})
	})

	// Switching off keeps the vault; removing the synced copy turns the
	// other device's sync off too.
	if _, err := b.DisableShard(ctx, &pb.DisableShardRequest{RemoveSynced: true}); err != nil {
		t.Fatal(err)
	}
	if got := names(b); !slices.Equal(got, []string{"zoe2"}) {
		t.Fatalf("b after off: %v", got)
	}
	eventually(t, "a turns off", func() bool { return !a.Sync.Enabled() })
}

func TestShardConflictOnEnable(t *testing.T) {
	ctx := context.Background()
	cs := fakeshard.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "desktop")
	for _, s := range []*Service{a, b} {
		if _, err := s.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
			t.Fatal(err)
		}
	}
	addAccount(t, a, "zoe")
	addAccount(t, b, "max")
	if _, err := a.EnableShard(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a seeds", func() bool { return syncStatus(a) == shard.StatusSynced })
	if _, err := b.EnableShard(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b sees a conflict", func() bool { return syncStatus(b) == shard.StatusConflict })
	if b.Sync.State().CanMerge {
		t.Fatal("two vaults made apart cannot merge")
	}
	if _, err := b.ResolveShard(ctx, &pb.ResolveShardRequest{Choice: pb.ResolveShardRequest_MERGE}); err == nil {
		t.Fatal("merge must be refused")
	}
	if _, err := b.ResolveShard(ctx, &pb.ResolveShardRequest{Choice: pb.ResolveShardRequest_USE_SYNCED}); err != nil {
		t.Fatal(err)
	}
	// Another vault: b locks and opens the synced one with its password.
	eventually(t, "b takes the synced vault", func() bool {
		st, _ := b.GetState(ctx, &pb.Empty{})
		return !st.Unlocked && syncStatus(b) == shard.StatusSynced
	})
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if got := names(b); !slices.Equal(got, []string{"Demo zoe"}) {
		t.Fatalf("b: %v", got)
	}
	m, _ := filepath.Glob(b.vaultPath() + ".backup-*")
	if len(m) != 1 {
		t.Fatalf("backup: %v", m)
	}
}

func TestShardOtherPassword(t *testing.T) {
	ctx := context.Background()
	cs := fakeshard.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "phone")
	if _, err := a.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	a.EnableShard(ctx, &pb.Empty{})
	eventually(t, "a seeds", func() bool { return syncStatus(a) == shard.StatusSynced })
	b.EnableShard(ctx, &pb.Empty{})
	eventually(t, "b gets it", func() bool { st, _ := b.GetState(ctx, &pb.Empty{}); return st.VaultExists })
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b synced", func() bool { return syncStatus(b) == shard.StatusSynced })

	// a changes the password while b is offline and b adds an account.
	cs.SetDown("phone", true)
	eventually(t, "b sees Shard gone", func() bool { return syncStatus(b) == shard.StatusNotRunning })
	if _, err := a.ChangePassword(ctx, &pb.ChangePasswordRequest{OldPassword: "pw", NewPassword: "new"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a pushes", func() bool { return syncStatus(a) == shard.StatusSynced && a.Sync.State().SyncedMs > 0 })
	time.Sleep(100 * time.Millisecond)
	addAccount(t, b, "max")
	cs.SetDown("phone", false)
	eventually(t, "b needs the new password", func() bool { return syncStatus(b) == shard.StatusNeedsPassword })
	if _, err := b.ShardPassword(ctx, &pb.PasswordRequest{Password: "wrong"}); err == nil {
		t.Fatal("a wrong password must fail")
	}
	if _, err := b.ShardPassword(ctx, &pb.PasswordRequest{Password: "new"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "both have max", func() bool {
		return slices.Equal(names(a), []string{"Demo max"}) && slices.Equal(names(b), []string{"Demo max"})
	})
	// The latest password wins on both.
	b.Close()
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "new"}); err != nil {
		t.Fatalf("b with the new password: %v", err)
	}
}
