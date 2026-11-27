// Package fakecs is an in-memory CloudSync for tests: the Connector API's
// records with revisions and concurrent versions, pairing and Watch.
package fakecs

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"megachat/core/cloudsync"
	"megachat/core/cloudsync/pb"
)

// Server is one CloudSync vault shared by every device of a test.
type Server struct {
	mu      sync.Mutex
	State   pb.VaultState
	VaultID string
	// Answer is the answer to pairing requests (APPROVED by default).
	Answer  pb.PairingStatus
	tokens  map[string]bool
	records map[string]*record
	seq     int
	watches map[int]chan *pb.Change
	nextW   int
	down    map[string]bool // devices where CloudSync is not running
}

// SetDown stops (or starts again) CloudSync on a device: its calls fail
// and its watches end.
func (s *Server) SetDown(device string, down bool) {
	s.mu.Lock()
	s.down[device] = down
	s.mu.Unlock()
	if down {
		s.broadcast(nil)
	}
}

func (s *Server) isDown(device string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.down[device]
}

type record struct {
	rev      string
	versions []*pb.Version
}

// New returns a READY server.
func New() *Server {
	return &Server{
		State: pb.VaultState_READY, VaultID: "csv1", Answer: pb.PairingStatus_APPROVED,
		tokens: map[string]bool{}, records: map[string]*record{}, watches: map[int]chan *pb.Change{},
		down: map[string]bool{},
	}
}

// Dialer is a device's way to this server.
func (s *Server) Dialer(device string) cloudsync.Dialer { return dialer{s, device} }

// Versions returns the record's current values.
func (s *Server) Versions(key string) [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.records[key]
	if r == nil {
		return nil
	}
	var out [][]byte
	for _, v := range r.versions {
		out = append(out, append([]byte{}, v.Data...))
	}
	return out
}

// Concurrent adds data as one more version of key, as if a device wrote
// it offline from the revision before the current one.
func (s *Server) Concurrent(key string, data []byte, device string) {
	s.mu.Lock()
	r := s.records[key]
	if r == nil {
		r = &record{}
		s.records[key] = r
	}
	s.seq++
	r.versions = append(r.versions, &pb.Version{Id: fmt.Sprintf("w%03d", s.seq), Data: data, Size: int64(len(data)), DeviceName: device, TimeUnixMs: int64(s.seq)})
	r.rev = fmt.Sprintf("r%03d", s.seq)
	ch := &pb.Change{Kind: pb.Change_RECORD, Key: key, Revision: r.rev}
	s.mu.Unlock()
	s.broadcast(ch)
}

// Revoke drops every pairing.
func (s *Server) Revoke() {
	s.mu.Lock()
	s.tokens = map[string]bool{}
	s.mu.Unlock()
	s.broadcast(&pb.Change{Kind: pb.Change_UNPAIRED})
}

func (s *Server) broadcast(ch *pb.Change) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, w := range s.watches {
		select {
		case w <- ch:
		default:
		}
	}
}

type dialer struct {
	s      *Server
	device string
}

func (dialer) UsesToken() bool { return true }
func (dialer) CanOpen() bool   { return false }

func (d dialer) Dial(_ context.Context, token string) (cloudsync.Client, error) {
	d.s.mu.Lock()
	defer d.s.mu.Unlock()
	if d.s.down[d.device] {
		return nil, cloudsync.ErrNotRunning
	}
	return &client{s: d.s, device: d.device, token: token}, nil
}

type client struct {
	s      *Server
	device string
	mu     sync.Mutex
	token  string
}

func (c *client) SetToken(t string) { c.mu.Lock(); c.token = t; c.mu.Unlock() }
func (c *client) Close() error      { return nil }

func (c *client) paired() bool {
	c.mu.Lock()
	t := c.token
	c.mu.Unlock()
	return t != "" && c.s.tokens[t]
}

// auth checks the token and the vault state; called with s.mu.
func (c *client) auth(write bool) error {
	if !c.paired() {
		return status.Error(codes.Unauthenticated, "not paired")
	}
	switch c.s.State {
	case pb.VaultState_LOCKED:
		return status.Error(codes.FailedPrecondition, "locked")
	case pb.VaultState_OFFLINE:
		if write {
			return status.Error(codes.Unavailable, "offline")
		}
	}
	return nil
}

func (c *client) Hello(_ context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	if c.s.isDown(c.device) {
		return nil, cloudsync.ErrNotRunning
	}
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	return &pb.HelloReply{ApiVersion: 1, State: c.s.State, Paired: c.paired(), AppId: r.AppId, VaultId: c.s.VaultID, DeviceName: c.device}, nil
}

func (c *client) RequestPairing(context.Context, *pb.PairingRequest) (*pb.PairingTicket, error) {
	return &pb.PairingTicket{Ticket: "t-" + c.device, Code: "123456"}, nil
}

func (c *client) AwaitPairing(_ context.Context, r *pb.AwaitPairingRequest) (*pb.PairingResult, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	if c.s.Answer != pb.PairingStatus_APPROVED {
		return &pb.PairingResult{Status: c.s.Answer}, nil
	}
	c.s.seq++
	t := fmt.Sprintf("tok-%s-%d", c.device, c.s.seq)
	c.s.tokens[t] = true
	return &pb.PairingResult{Status: pb.PairingStatus_APPROVED, Token: t, AppId: cloudsync.AppID}, nil
}

func (c *client) Get(_ context.Context, r *pb.GetRequest) (*pb.Item, error) {
	if c.s.isDown(c.device) {
		return nil, cloudsync.ErrNotRunning
	}
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	if err := c.auth(false); err != nil {
		return nil, err
	}
	out := &pb.Item{Key: r.Key}
	if rec := c.s.records[r.Key]; rec != nil {
		out.Revision = rec.rev
		for _, v := range rec.versions {
			out.Versions = append(out.Versions, proto.Clone(v).(*pb.Version))
		}
	}
	return out, nil
}

func (c *client) write(key, base string, data []byte, del bool) (*pb.Item, error) {
	if c.s.isDown(c.device) {
		return nil, cloudsync.ErrNotRunning
	}
	c.s.mu.Lock()
	if err := c.auth(true); err != nil {
		c.s.mu.Unlock()
		return nil, err
	}
	rec := c.s.records[key]
	if rec == nil {
		rec = &record{}
		c.s.records[key] = rec
	}
	if rec.rev != base {
		c.s.mu.Unlock()
		return nil, status.Error(codes.Aborted, "base is not the current revision")
	}
	c.s.seq++
	rec.rev = fmt.Sprintf("r%03d", c.s.seq)
	rec.versions = nil
	if !del {
		rec.versions = []*pb.Version{{Id: fmt.Sprintf("w%03d", c.s.seq), Data: append([]byte{}, data...), Size: int64(len(data)), DeviceName: c.device, TimeUnixMs: int64(c.s.seq)}}
	}
	out := &pb.Item{Key: key, Revision: rec.rev}
	ch := &pb.Change{Kind: pb.Change_RECORD, Key: key, Revision: rec.rev, Deleted: del}
	c.s.mu.Unlock()
	c.s.broadcast(ch)
	return out, nil
}

func (c *client) Put(_ context.Context, r *pb.PutRequest) (*pb.Item, error) {
	return c.write(r.Key, r.Base, r.Data, false)
}

func (c *client) Delete(_ context.Context, r *pb.DeleteRequest) (*pb.Item, error) {
	return c.write(r.Key, r.Base, nil, true)
}

func (c *client) Unpair(context.Context, *pb.UnpairRequest) (*pb.UnpairReply, error) {
	c.s.mu.Lock()
	defer c.s.mu.Unlock()
	c.mu.Lock()
	delete(c.s.tokens, c.token)
	c.mu.Unlock()
	return &pb.UnpairReply{}, nil
}

func (c *client) Watch(ctx context.Context) (func() (*pb.Change, error), error) {
	c.s.mu.Lock()
	if !c.paired() {
		c.s.mu.Unlock()
		return nil, status.Error(codes.Unauthenticated, "not paired")
	}
	ch := make(chan *pb.Change, 64)
	id := c.s.nextW
	c.s.nextW++
	c.s.watches[id] = ch
	c.s.mu.Unlock()
	return func() (*pb.Change, error) {
		select {
		case x := <-ch:
			if x == nil || c.s.isDown(c.device) {
				return nil, cloudsync.ErrNotRunning
			}
			return x, nil
		case <-ctx.Done():
			c.s.mu.Lock()
			delete(c.s.watches, id)
			c.s.mu.Unlock()
			return nil, ctx.Err()
		}
	}, nil
}
