// Package embed runs the core inside the app's own process: the Flutter
// app on desktops (desktopbind, linked into the executable) and on Android
// (mobilebind). The app talks to it over the gRPC API, through a private
// unix socket (localhost tcp on Windows) and a bearer token. There is no
// separate daemon: the accounts stay connected while the app runs.
package embed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc"

	"megachat/core/crypto"
	"megachat/core/eventlog"
	"megachat/core/keychain"
	"megachat/core/providers/all"
	"megachat/core/rpc"
	"megachat/core/rpc/pb"
)

// Options say where the core keeps its files and listens.
type Options struct {
	Dir      string // vault.mc and cache-*.db
	Platform string // runtime.GOOS when empty
	Listen   string // unix:PATH or tcp:HOST:PORT
	// Endpoint, when set, is where address and token are written for the
	// mchat CLI ({"address": "unix:///…", "token": "…"}, readable only by
	// the user). An endpoint there that still answers means another app
	// instance runs the core: Start returns ErrRunning.
	Endpoint string
	LogFile  string // also log here (empty: Log only)
	Log      io.Writer
	Demo     bool // add the Demo messenger
}

// Core is a running core.
type Core struct {
	Service *rpc.Service
	Address string // as in endpoint.json: unix:///path or host:port
	Token   string

	srv  *grpc.Server
	lis  net.Listener
	opts Options
	logf io.Closer
}

// ErrRunning: another process of this user already runs the core (the
// cache is locked to one process).
var ErrRunning = errors.New("MegaChat is already running")

// Start opens the socket and serves the API. The vault stays locked until
// the app unlocks it.
func Start(o Options) (*Core, error) {
	if o.Platform == "" {
		o.Platform = runtime.GOOS
	}
	if err := os.MkdirAll(o.Dir, 0o700); err != nil {
		return nil, err
	}
	if o.Endpoint != "" {
		if _, _, ok := ReadEndpoint(o.Endpoint); ok {
			return nil, ErrRunning
		}
	}
	out := o.Log
	var logf *os.File
	if o.LogFile != "" {
		if f, err := openLog(o.LogFile); err == nil {
			logf = f
			if out != nil {
				out = io.MultiWriter(out, f)
			} else {
				out = f
			}
		}
	}
	lis, err := listenOn(o.Listen)
	if err != nil {
		if logf != nil {
			logf.Close()
		}
		return nil, err
	}
	svc := rpc.NewService(filepath.Join(o.Dir, "vault.mc"), o.Dir, o.Platform, all.Registry(o.Demo))
	if out != nil {
		svc.Log = eventlog.New(1000, out)
	}
	svc.Keychain = keychain.System()
	c := &Core{Service: svc, Token: crypto.RandomID(32), lis: lis, opts: o}
	if logf != nil {
		c.logf = logf
	}
	c.Address = lis.Addr().String()
	if lis.Addr().Network() == "unix" {
		c.Address = "unix://" + c.Address
	}
	if o.Endpoint != "" {
		if err := writeEndpoint(o.Endpoint, c.Address, c.Token); err != nil {
			lis.Close()
			c.closeLog()
			return nil, err
		}
	}
	c.srv = grpc.NewServer(rpc.TokenAuth(c.Token)...)
	pb.RegisterMegaChatServer(c.srv, svc)
	go c.srv.Serve(lis)
	return c, nil
}

// Stop locks the vault (the sessions close, the cache is flushed) and
// stops the server.
func (c *Core) Stop() {
	c.Service.Close()
	// Open Watch streams would hold GracefulStop forever.
	stopped := make(chan struct{})
	go func() { c.srv.GracefulStop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		c.srv.Stop()
	}
	if c.opts.Endpoint != "" {
		if addr, _, _ := readEndpointFile(c.opts.Endpoint); addr == c.Address {
			os.Remove(c.opts.Endpoint)
		}
	}
	if strings.HasPrefix(c.opts.Listen, "unix:") {
		os.Remove(strings.TrimPrefix(c.opts.Listen, "unix:"))
	}
	c.closeLog()
}

func (c *Core) closeLog() {
	if c.logf != nil {
		c.logf.Close()
	}
}

// DefaultDir is where the desktop core keeps its files: the user's config
// dir/megachat (~/.config/megachat, %AppData%\megachat).
func DefaultDir() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		cfg = "."
	}
	return filepath.Join(cfg, "megachat")
}

// DefaultListen is the desktop socket: $XDG_RUNTIME_DIR/megachat.sock, or
// a localhost port on Windows.
func DefaultListen(dir string) string {
	if runtime.GOOS == "windows" {
		return "tcp:127.0.0.1:0"
	}
	d := os.Getenv("XDG_RUNTIME_DIR")
	if d == "" {
		d = dir
	}
	return "unix:" + filepath.Join(d, "megachat.sock")
}

// ReadEndpoint reads an endpoint file and says whether the core there
// answers.
func ReadEndpoint(path string) (address, token string, ok bool) {
	address, token, err := readEndpointFile(path)
	if err != nil {
		return "", "", false
	}
	network, addr := "tcp", address
	if strings.HasPrefix(address, "unix://") {
		network, addr = "unix", strings.TrimPrefix(address, "unix://")
	}
	conn, err := net.DialTimeout(network, addr, 2*time.Second)
	if err != nil {
		return address, token, false
	}
	conn.Close()
	return address, token, true
}

func readEndpointFile(path string) (address, token string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var ep struct{ Address, Token string }
	if err := json.Unmarshal(data, &ep); err != nil {
		return "", "", err
	}
	return ep.Address, ep.Token, nil
}

func listenOn(addr string) (net.Listener, error) {
	switch {
	case strings.HasPrefix(addr, "unix:"):
		p := strings.TrimPrefix(addr, "unix:")
		if len(p) > 100 {
			return nil, fmt.Errorf("socket path too long for unix sockets (%d bytes): %s", len(p), p)
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			return nil, err
		}
		if c, err := net.Dial("unix", p); err == nil {
			c.Close()
			return nil, ErrRunning
		}
		os.Remove(p)
		l, err := net.Listen("unix", p)
		if err == nil {
			os.Chmod(p, 0o600)
		}
		return l, err
	case strings.HasPrefix(addr, "tcp:"):
		return net.Listen("tcp", strings.TrimPrefix(addr, "tcp:"))
	}
	return nil, fmt.Errorf("bad listen address %q", addr)
}

func writeEndpoint(path, address, token string) error {
	data, _ := json.MarshalIndent(map[string]string{"address": address, "token": token}, "", "  ")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// openLog opens the log for appending; a log over 5 MB starts over.
func openLog(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil && fi.Size() > 5<<20 {
		os.Rename(path, path+".1")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}
