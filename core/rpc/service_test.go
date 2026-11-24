package rpc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"megachat/core/keychain"
	"megachat/core/provider"
	"megachat/core/providers/fake"
	"megachat/core/rpc/pb"
	"megachat/core/vault"
)

func serve(t *testing.T) (pb.MegaChatClient, *fake.Server, context.Context) {
	return serveWith(t, nil)
}

func serveWith(t *testing.T, setup func(*Service)) (pb.MegaChatClient, *fake.Server, context.Context) {
	t.Helper()
	dir := t.TempDir()
	srv := fake.New()
	svc := NewService(filepath.Join(dir, "vault.mc"), dir, "test", provider.NewRegistry(srv))
	svc.KDF = vault.KDFParams{Time: 1, Memory: 1024, Threads: 1}
	if setup != nil {
		setup(svc)
	}
	lis := bufconn.Listen(1 << 20)
	g := grpc.NewServer(TokenAuth("tok")...)
	pb.RegisterMegaChatServer(g, svc)
	go g.Serve(lis)
	conn, err := grpc.NewClient("passthrough:///buf",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(); svc.Close(); g.Stop() })
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer tok")
	return pb.NewMegaChatClient(conn), srv, ctx
}

func TestServiceFlow(t *testing.T) {
	c, srv, ctx := serve(t)
	if _, err := c.GetState(context.Background(), &pb.Empty{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("no token: %v", err)
	}
	st, err := c.GetState(ctx, &pb.Empty{})
	if err != nil || st.VaultExists || st.Unlocked {
		t.Fatalf("fresh: %v %v", st, err)
	}
	if _, err := c.Accounts(ctx, &pb.Empty{}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("locked: %v", err)
	}
	if st, err = c.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil || !st.Unlocked {
		t.Fatalf("create: %v %v", st, err)
	}

	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	watch, err := c.Watch(wctx, &pb.Empty{})
	if err != nil {
		t.Fatal(err)
	}

	step, err := c.StartLogin(ctx, &pb.StartLoginRequest{Provider: fake.Type})
	if err != nil || len(step.Fields) != 1 {
		t.Fatalf("start login: %v %v", step, err)
	}
	done, err := c.LoginStep(ctx, &pb.LoginStepRequest{LoginId: step.LoginId, Values: map[string]string{"name": "zoe"}})
	if err != nil || !done.Done || done.Account.GetId() == "" {
		t.Fatalf("login: %v %v", done, err)
	}
	acct := done.Account.Id

	deadline := time.Now().Add(5 * time.Second)
	for {
		cs, _ := c.Chats(ctx, &pb.ChatsRequest{})
		if len(cs.GetChats()) == 10 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("chats never arrived")
		}
		time.Sleep(20 * time.Millisecond)
	}
	in := srv.Inject("zoe", "c3", "ping")
	for {
		ev, err := watch.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if ev.Kind == pb.Event_KIND_MESSAGE_NEW {
			if ev.Message.Id != in.ID || ev.AccountId != acct || ev.Chat.GetTitle() == "" {
				t.Fatalf("event: %v", ev)
			}
			break
		}
	}
	h, err := c.History(ctx, &pb.HistoryRequest{AccountId: acct, ChatId: "c3", Limit: 5})
	if err != nil || len(h.Messages) != 5 || h.Messages[0].Text != "ping" || !h.More {
		t.Fatalf("history: %v %v", h, err)
	}
	if _, err := c.Send(ctx, &pb.SendRequest{AccountId: acct, ChatId: "c3", Text: "pong"}); err != nil {
		t.Fatal(err)
	}
	// A file larger than gRPC's default 4 MB message limit.
	big := make([]byte, 6<<20)
	sent, err := c.Send(ctx, &pb.SendRequest{AccountId: acct, ChatId: "c1", Files: []*pb.OutFile{{Name: "big.bin", Data: big}}})
	if err != nil || len(sent.Attachments) != 1 || sent.Attachments[0].Size != int64(len(big)) || sent.Attachments[0].Name != "big.bin" {
		t.Fatalf("send a file: %v %v", sent, err)
	}
	if _, err := c.History(ctx, &pb.HistoryRequest{AccountId: "nope", ChatId: "c3"}); status.Code(err) != codes.NotFound {
		t.Fatalf("unknown account: %v", err)
	}

	if _, err := c.LockVault(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Unlock(ctx, &pb.PasswordRequest{Password: "bad"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("bad password: %v", err)
	}
	if _, err := c.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	as, err := c.Accounts(ctx, &pb.Empty{})
	if err != nil || len(as.Accounts) != 1 || as.Accounts[0].Name != "Demo zoe" {
		t.Fatalf("accounts after unlock: %v %v", as, err)
	}
	// The cache survived the lock: history is there before any provider call.
	h, err = c.History(ctx, &pb.HistoryRequest{AccountId: acct, ChatId: "c3", Limit: 2})
	if err != nil || len(h.Messages) != 2 || h.Messages[0].Text != "pong" {
		t.Fatalf("cached history: %v %v", h, err)
	}
}

func TestRememberPassword(t *testing.T) {
	c, _, ctx := serve(t)
	if st, _ := c.GetState(ctx, &pb.Empty{}); st.CanRemember {
		t.Fatal("no keychain: can_remember")
	}
	if _, err := c.UnlockRemembered(ctx, &pb.Empty{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("no keychain: %v", err)
	}

	kc := keychain.NewMemory()
	var svc *Service
	c, _, ctx = serveWith(t, func(s *Service) { svc, s.Keychain = s, kc })
	kept := func() []byte { b, _ := kc.Get(svc.keychainName()); return b }
	st, err := c.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"})
	if err != nil || !st.CanRemember || st.Remembered {
		t.Fatalf("create: %v %v", st, err)
	}
	if _, err := c.RememberPassword(ctx, &pb.PasswordRequest{Password: "bad"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("remember a wrong password: %v", err)
	}
	if st, err = c.RememberPassword(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil || !st.Remembered {
		t.Fatalf("remember: %v %v", st, err)
	}
	if v := kept(); len(v) != 32 {
		t.Fatalf("kept %q: the derived key, not the password", v)
	}
	lockAndUnlock := func() (*pb.State, error) {
		if _, err := c.LockVault(ctx, &pb.Empty{}); err != nil {
			t.Fatal(err)
		}
		return c.UnlockRemembered(ctx, &pb.Empty{})
	}
	if st, err = lockAndUnlock(); err != nil || !st.Unlocked {
		t.Fatalf("unlock remembered: %v %v", st, err)
	}

	// A password change keeps the remembered key working.
	if _, err := c.ChangePassword(ctx, &pb.ChangePasswordRequest{OldPassword: "pw", NewPassword: "pw2"}); err != nil {
		t.Fatal(err)
	}
	if st, err = lockAndUnlock(); err != nil || !st.Unlocked || !st.Remembered {
		t.Fatalf("after a password change: %v %v", st, err)
	}

	// A stale key (the vault changed behind our back) is forgotten.
	kc.Set(svc.keychainName(), make([]byte, 32))
	if _, err := lockAndUnlock(); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("stale key: %v", err)
	}
	if st, _ = c.GetState(ctx, &pb.Empty{}); st.Remembered || kept() != nil {
		t.Fatalf("stale key kept: %v", st)
	}
	if _, err := c.UnlockRemembered(ctx, &pb.Empty{}); status.Code(err) != codes.NotFound {
		t.Fatalf("nothing remembered: %v", err)
	}

	if _, err := c.Unlock(ctx, &pb.PasswordRequest{Password: "pw2"}); err != nil {
		t.Fatal(err)
	}
	if _, err = c.RememberPassword(ctx, &pb.PasswordRequest{Password: "pw2"}); err != nil {
		t.Fatal(err)
	}
	if st, err = c.ForgetPassword(ctx, &pb.Empty{}); err != nil || st.Remembered || kept() != nil {
		t.Fatalf("forget: %v %v", st, err)
	}
}

func TestSetVaultPath(t *testing.T) {
	kc := keychain.NewMemory()
	var svc *Service
	pointer := filepath.Join(t.TempDir(), "vault.path")
	c, _, ctx := serveWith(t, func(s *Service) { svc, s.Keychain, s.VaultPointer = s, kc, pointer })
	old := svc.vaultPath()
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: t.TempDir(), Move: true}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("move while locked: %v", err)
	}
	if _, err := c.CreateVault(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.RememberPassword(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: "rel/vault.mc", Move: true}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("relative path: %v", err)
	}
	dir := t.TempDir()
	st, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: dir, Move: true}) // a directory: vault.mc in it
	moved := filepath.Join(dir, "vault.mc")
	if err != nil || st.VaultPath != moved || !st.Unlocked || !st.Remembered {
		t.Fatalf("move: %v %v", st, err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old file stays: %v", err)
	}
	if ReadVaultPointer(pointer) != moved {
		t.Fatalf("pointer: %q", ReadVaultPointer(pointer))
	}
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: moved, Move: false}); err != nil {
		t.Fatalf("same place: %v", err)
	}
	// The remembered key followed the file.
	if _, err := c.LockVault(ctx, &pb.Empty{}); err != nil {
		t.Fatal(err)
	}
	if st, err = c.UnlockRemembered(ctx, &pb.Empty{}); err != nil || !st.Unlocked {
		t.Fatalf("unlock remembered after move: %v %v", st, err)
	}
	// Moving onto an existing file is refused.
	other := filepath.Join(t.TempDir(), "other.mc")
	if err := os.WriteFile(other, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: other, Move: true}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("move onto a file: %v", err)
	}
	// Using another file needs a locked vault and an existing file.
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: moved + ".x"}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("use while unlocked: %v", err)
	}
	c.LockVault(ctx, &pb.Empty{})
	if _, err := c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: moved + ".x"}); status.Code(err) != codes.NotFound {
		t.Fatalf("use a missing file: %v", err)
	}
	back := filepath.Join(t.TempDir(), "copy.mc")
	data, _ := os.ReadFile(moved)
	os.WriteFile(back, data, 0o600)
	if st, err = c.SetVaultPath(ctx, &pb.VaultPathRequest{Path: back}); err != nil || st.VaultPath != back || st.Unlocked {
		t.Fatalf("use: %v %v", st, err)
	}
	if st, err = c.Unlock(ctx, &pb.PasswordRequest{Password: "pw"}); err != nil || !st.Unlocked {
		t.Fatalf("unlock the copy: %v %v", st, err)
	}
}
