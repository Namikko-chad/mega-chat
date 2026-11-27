package rpc

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"megachat/core/cloudsync"
	"megachat/core/cloudsync/fakecs"
	"megachat/core/keychain"
	"megachat/core/provider"
	"megachat/core/providers/fake"
	"megachat/core/rpc/pb"
	"megachat/core/vault"
)

// device is one MegaChat with sync, on a CloudSync shared by the test.
func device(t *testing.T, cs *fakecs.Server, name string) *Service {
	t.Helper()
	dir := t.TempDir()
	svc := NewService(filepath.Join(dir, "vault.mc"), dir, "test", provider.NewRegistry(fake.New()))
	svc.KDF = vault.KDFParams{Time: 1, Memory: 1024, Threads: 1}
	svc.Keychain = keychain.NewMemory()
	sy := cloudsync.New(cloudsync.Options{
		Dialer: cs.Dialer(name), Host: svc.SyncHost(), Dir: dir, Keychain: svc.Keychain,
		AppName: "MegaChat", OnChange: svc.SyncChanged,
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

func syncStatus(s *Service) cloudsync.Status { return s.Sync.State().Status }

func TestCloudSyncTwoDevices(t *testing.T) {
	ctx := context.Background()
	cs := fakecs.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "phone")

	if _, err := a.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	addAccount(t, a, "zoe")
	if _, err := a.EnableCloudSync(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a seeds CloudSync", func() bool {
		return syncStatus(a) == cloudsync.StatusSynced && len(cs.Versions(cloudsync.RecordKey)) == 1
	})

	// The phone has no vault: it takes the synced one and opens it with
	// the same password.
	if _, err := b.EnableCloudSync(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b gets the vault", func() bool {
		st, _ := b.GetState(ctx, &pb.Empty{})
		return st.VaultExists && st.Cloudsync.GetStatus() == pb.CloudSyncState_SYNCED
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
	cs.Concurrent(cloudsync.RecordKey, old, "tablet")
	eventually(t, "the copies are merged", func() bool {
		return len(cs.Versions(cloudsync.RecordKey)) == 1 && syncStatus(a) == cloudsync.StatusSynced && syncStatus(b) == cloudsync.StatusSynced
	})
	eventually(t, "max stays removed", func() bool {
		return slices.Equal(names(a), []string{"zoe2"}) && slices.Equal(names(b), []string{"zoe2"})
	})

	// Switching off keeps the vault; removing the synced copy turns the
	// other device's sync off too.
	if _, err := b.DisableCloudSync(ctx, &pb.DisableCloudSyncRequest{RemoveSynced: true}); err != nil {
		t.Fatal(err)
	}
	if got := names(b); !slices.Equal(got, []string{"zoe2"}) {
		t.Fatalf("b after off: %v", got)
	}
	eventually(t, "a turns off", func() bool { return !a.Sync.Enabled() })
}

func TestCloudSyncConflictOnEnable(t *testing.T) {
	ctx := context.Background()
	cs := fakecs.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "desktop")
	for _, s := range []*Service{a, b} {
		if _, err := s.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
			t.Fatal(err)
		}
	}
	addAccount(t, a, "zoe")
	addAccount(t, b, "max")
	if _, err := a.EnableCloudSync(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a seeds", func() bool { return syncStatus(a) == cloudsync.StatusSynced })
	if _, err := b.EnableCloudSync(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b sees a conflict", func() bool { return syncStatus(b) == cloudsync.StatusConflict })
	if b.Sync.State().CanMerge {
		t.Fatal("two vaults made apart cannot merge")
	}
	if _, err := b.ResolveCloudSync(ctx, &pb.ResolveCloudSyncRequest{Choice: pb.ResolveCloudSyncRequest_MERGE}); err == nil {
		t.Fatal("merge must be refused")
	}
	if _, err := b.ResolveCloudSync(ctx, &pb.ResolveCloudSyncRequest{Choice: pb.ResolveCloudSyncRequest_USE_SYNCED}); err != nil {
		t.Fatal(err)
	}
	// Another vault: b locks and opens the synced one with its password.
	eventually(t, "b takes the synced vault", func() bool {
		st, _ := b.GetState(ctx, &pb.Empty{})
		return !st.Unlocked && syncStatus(b) == cloudsync.StatusSynced
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

func TestCloudSyncOtherPassword(t *testing.T) {
	ctx := context.Background()
	cs := fakecs.New()
	a := device(t, cs, "laptop")
	b := device(t, cs, "phone")
	if _, err := a.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	a.EnableCloudSync(ctx, &pb.Empty{})
	eventually(t, "a seeds", func() bool { return syncStatus(a) == cloudsync.StatusSynced })
	b.EnableCloudSync(ctx, &pb.Empty{})
	eventually(t, "b gets it", func() bool { st, _ := b.GetState(ctx, &pb.Empty{}); return st.VaultExists })
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b synced", func() bool { return syncStatus(b) == cloudsync.StatusSynced })

	// a changes the password while b is offline and b adds an account.
	cs.SetDown("phone", true)
	eventually(t, "b sees CloudSync gone", func() bool { return syncStatus(b) == cloudsync.StatusNotRunning })
	if _, err := a.ChangePassword(ctx, &pb.ChangePasswordRequest{OldPassword: "pw", NewPassword: "new"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a pushes", func() bool { return syncStatus(a) == cloudsync.StatusSynced && a.Sync.State().SyncedMs > 0 })
	time.Sleep(100 * time.Millisecond)
	addAccount(t, b, "max")
	cs.SetDown("phone", false)
	eventually(t, "b needs the new password", func() bool { return syncStatus(b) == cloudsync.StatusNeedsPassword })
	if _, err := b.CloudSyncPassword(ctx, &pb.PasswordRequest{Password: "wrong"}); err == nil {
		t.Fatal("a wrong password must fail")
	}
	if _, err := b.CloudSyncPassword(ctx, &pb.PasswordRequest{Password: "new"}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "both have max", func() bool { return slices.Equal(names(a), []string{"Demo max"}) && slices.Equal(names(b), []string{"Demo max"}) })
	// The latest password wins on both.
	b.Close()
	if _, err := b.Unlock(ctx, &pb.PasswordRequest{Password: "new"}); err != nil {
		t.Fatalf("b with the new password: %v", err)
	}
}
