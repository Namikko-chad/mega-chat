package embed

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"megachat/core/rpc/pb"
)

func TestStartServesAndStops(t *testing.T) {
	dir := t.TempDir()
	// Unix socket paths are short; t.TempDir may be long.
	sockDir, err := os.MkdirTemp("", "mc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sockDir)
	o := Options{
		Dir: dir, Listen: "unix:" + filepath.Join(sockDir, "s"),
		Endpoint: filepath.Join(dir, "endpoint.json"), LogFile: filepath.Join(dir, "megachat.log"),
		NoCloudSync: true, // never the user's own CloudSync
	}
	c, err := Start(o)
	if err != nil {
		t.Fatal(err)
	}
	addr, token, ok := ReadEndpoint(o.Endpoint)
	if !ok || addr != c.Address || token != c.Token {
		t.Fatalf("endpoint %q %q %v, core %q", addr, token, ok, c.Address)
	}

	conn, err := grpc.NewClient("unix:"+strings.TrimPrefix(addr, "unix://"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	api := pb.NewMegaChatClient(conn)
	if _, err := api.GetState(context.Background(), &pb.Empty{}); err == nil {
		t.Fatal("a call without the token was served")
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
	st, err := api.GetState(ctx, &pb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if st.VaultExists || st.VaultPath != filepath.Join(dir, "vault.mc") {
		t.Fatalf("state %v", st)
	}

	// A second instance finds the first one through the endpoint.
	if _, err := Start(o); !errors.Is(err, ErrRunning) {
		t.Fatalf("second start: %v", err)
	}
	o2 := o
	o2.Endpoint = ""
	if _, err := Start(o2); !errors.Is(err, ErrRunning) {
		t.Fatalf("second start on the same socket: %v", err)
	}

	c.Stop()
	if _, err := os.Stat(o.Endpoint); !os.IsNotExist(err) {
		t.Fatalf("endpoint left after Stop: %v", err)
	}
	// The socket is free again.
	c2, err := Start(o)
	if err != nil {
		t.Fatal(err)
	}
	c2.Stop()
}
