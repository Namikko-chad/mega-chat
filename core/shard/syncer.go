package shard

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SelfHold/parley/core/keychain"
	"github.com/SelfHold/parley/core/shard/pb"
)

const (
	// AppID is Parley's data in Shard: never change it.
	AppID = "parley"
	// RecordKey is the record holding the vault file's bytes.
	RecordKey = "vault.parley"
	// MaxRecord is Shard's limit for one record.
	MaxRecord = 256 << 10
)

// Status is ShardState.Status of parley.proto (same numbers).
type Status int32

const (
	StatusOff           Status = 1
	StatusNotRunning    Status = 2
	StatusNoVault       Status = 3
	StatusLocked        Status = 4
	StatusOffline       Status = 5
	StatusPairing       Status = 6
	StatusSyncing       Status = 7
	StatusSynced        Status = 8
	StatusConflict      Status = 9
	StatusNeedsPassword Status = 10
	StatusWaitingUnlock Status = 11
	StatusError         Status = 12
)

// Choice resolves a conflict (ResolveShardRequest.Choice).
type Choice int32

const (
	UseSynced Choice = 1
	Upload    Choice = 2
	MergeBoth Choice = 3
)

// State is what the app shows (ShardState of parley.proto).
type State struct {
	Status      Status
	Enabled     bool
	Available   bool
	PairingCode string
	Message     string
	SyncedMs    int64
	DeviceName  string
	CanMerge    bool
	Unpaired    bool
	CanOpen     bool
}

// Errors of a Host.
var (
	// ErrLocked: the vault is locked here, and the step needs its key.
	ErrLocked = errors.New("unlock Parley to finish syncing")
	// ErrNeedPassword: the synced copy is sealed with another password.
	ErrNeedPassword = errors.New("the synced vault has another password")
	// ErrLocalChanged: the vault file changed meanwhile (read it again).
	ErrLocalChanged = errors.New("the vault changed meanwhile")
	// ErrOtherVault: the synced copy is another vault: no merge.
	ErrOtherVault = errors.New("the synced vault is another vault")
)

// Host is the vault side (the rpc service).
type Host interface {
	// LocalVault returns the vault file's bytes; nil when there is none.
	LocalVault() ([]byte, error)
	// AdoptVault makes data (a copy of this vault from another device)
	// the vault file, if the file is still the one with hash expect
	// ("" = no file; else ErrLocalChanged). passwords open a copy sealed
	// with another password (ErrNeedPassword without one that does).
	AdoptVault(data []byte, expect string, passwords [][]byte) error
	// MergeVault merges copies into the open vault (ErrLocked when it is
	// locked), saves it and returns the file.
	MergeVault(copies [][]byte, passwords [][]byte) ([]byte, error)
	// CompareVault tells whether data holds the same accounts as the open
	// vault, and whether it is this vault at all (it can be merged).
	CompareVault(data []byte, passwords [][]byte) (same, sameVault bool, err error)
	// UseSynced replaces the vault (whose file has hash expect) with data,
	// another vault perhaps, keeping a backup of this one. It may lock
	// the vault: data then opens with its own password.
	UseSynced(data []byte, expect string, passwords [][]byte) error
	// CanEnable: sync can start (the vault is unlocked, or there is none).
	CanEnable() error
}

// Options configure a Syncer.
type Options struct {
	Dialer Dialer
	Host   Host
	// Dir keeps this device's sync state (shard.json; the token there
	// too when there is no keychain): never synced.
	Dir string
	// Keychain keeps the pairing token (desktops); nil: a file in Dir.
	Keychain    keychain.Keychain
	AppName     string // shown in Shard, "Parley"
	Description string // "Parley 0.1 on linux"
	Log         func(level, msg string)
	// OnChange is called when State changes (not on its goroutine's lock).
	OnChange func()
}

// saved is this device's sync state (CONNECTOR.md § 7).
type saved struct {
	Enabled bool `json:"enabled"`
	// VaultID is Shard's vault: another one is a new, empty remote.
	VaultID string `json:"vault_id,omitempty"`
	// Base is the record's revision the local file was last synced at,
	// SyncedHash the file's hash then. Both empty: never synced (the
	// first sync compares instead of merging).
	Base       string `json:"base,omitempty"`
	SyncedHash string `json:"synced_hash,omitempty"`
	SyncedMs   int64  `json:"synced_ms,omitempty"`
}

// Syncer keeps the vault file and Shard's record the same.
type Syncer struct {
	o Options

	// Delays (tests shorten them).
	RetryDelay time.Duration // Shard not there / an error
	ProbeDelay time.Duration // sync off: look whether Shard is there
	Heartbeat  time.Duration // reconcile now and then anyway

	kick   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu         sync.Mutex
	sv         saved
	token      *string // loaded lazily
	st         State
	wantPair   bool
	pairCancel context.CancelFunc
	sessCancel context.CancelFunc
	resolve    Choice
	passwords  [][]byte
	pwWait     chan error // ShardPassword waits for the next reconcile
}

// New returns a syncer; Start runs it.
func New(o Options) *Syncer {
	s := &Syncer{
		o:          o,
		RetryDelay: 30 * time.Second, ProbeDelay: time.Minute, Heartbeat: 10 * time.Minute,
		kick: make(chan struct{}, 1), done: make(chan struct{}),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.load()
	s.st = State{Status: StatusNotRunning, Enabled: s.sv.Enabled, SyncedMs: s.sv.SyncedMs, CanOpen: o.Dialer.CanOpen()}
	if !s.sv.Enabled {
		s.st.Status = StatusOff
	}
	return s
}

// Start runs the syncer until Close.
func (s *Syncer) Start() { go s.loop() }

// Close stops it.
func (s *Syncer) Close() {
	s.cancel()
	<-s.done
}

// Kick makes the syncer look at once (the vault changed, unlocked, ...).
func (s *Syncer) Kick() {
	select {
	case s.kick <- struct{}{}:
	default:
	}
}

// State returns what to show.
func (s *Syncer) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.st
}

// Enabled says whether sync is on.
func (s *Syncer) Enabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sv.Enabled
}

// Enable switches sync on: it pairs this device when it is not paired,
// then syncs (the first sync compares the two vaults).
func (s *Syncer) Enable() (State, error) {
	if err := s.o.Host.CanEnable(); err != nil {
		return s.State(), err
	}
	s.mu.Lock()
	if !s.sv.Enabled {
		s.wantPair = true
		s.sv.Base, s.sv.SyncedHash = "", ""
		s.resolve, s.passwords = 0, nil
		s.st.Unpaired, s.st.Message, s.st.CanMerge = false, "", false
		s.st.Status = StatusSyncing
	}
	s.mu.Unlock()
	s.changed()
	s.Kick()
	return s.State(), nil
}

// CancelPairing stops waiting for the user to allow the pairing.
func (s *Syncer) CancelPairing() State {
	s.mu.Lock()
	s.wantPair = false
	if s.pairCancel != nil {
		s.pairCancel()
	}
	if !s.sv.Enabled {
		s.st.Status, s.st.PairingCode, s.st.Message = StatusOff, "", ""
	}
	s.mu.Unlock()
	s.changed()
	s.Kick()
	return s.State()
}

// Resolve answers a conflict.
func (s *Syncer) Resolve(c Choice) (State, error) {
	s.mu.Lock()
	switch {
	case s.st.Status != StatusConflict:
		s.mu.Unlock()
		return s.State(), status.Error(codes.FailedPrecondition, "there is no conflict to resolve")
	case c == MergeBoth && !s.st.CanMerge:
		s.mu.Unlock()
		return s.State(), status.Error(codes.FailedPrecondition, "these vaults cannot be merged")
	case c < UseSynced || c > MergeBoth:
		s.mu.Unlock()
		return s.State(), status.Error(codes.InvalidArgument, "unknown choice")
	}
	s.resolve = c
	s.st.Status, s.st.Message = StatusSyncing, ""
	s.mu.Unlock()
	s.changed()
	s.Kick()
	return s.State(), nil
}

// Password gives the password of the synced copy and waits for the sync
// to try it: PermissionDenied when it does not open the copy.
func (s *Syncer) Password(ctx context.Context, pw string) (State, error) {
	if pw == "" {
		return s.State(), status.Error(codes.InvalidArgument, "empty password")
	}
	wait := make(chan error, 1)
	s.mu.Lock()
	s.passwords = append(s.passwords, []byte(pw))
	s.pwWait = wait
	s.mu.Unlock()
	s.Kick()
	select {
	case err := <-wait:
		if errors.Is(err, ErrNeedPassword) {
			s.mu.Lock()
			s.passwords = slices.DeleteFunc(s.passwords, func(p []byte) bool { return string(p) == pw })
			s.mu.Unlock()
			return s.State(), status.Error(codes.PermissionDenied, "this password does not open the synced vault")
		}
	case <-time.After(2 * time.Minute):
	case <-ctx.Done():
		return s.State(), ctx.Err()
	}
	return s.State(), nil
}

// Disable switches sync off and unpairs; the local vault stays as it is.
// removeSynced also deletes the synced copy from Shard (the other
// devices then stop syncing it).
func (s *Syncer) Disable(ctx context.Context, removeSynced bool) (State, error) {
	s.mu.Lock()
	s.sv = saved{}
	s.wantPair, s.resolve, s.passwords = false, 0, nil
	if s.pairCancel != nil {
		s.pairCancel()
	}
	if s.sessCancel != nil {
		s.sessCancel()
	}
	s.st = State{Status: StatusOff, Available: s.st.Available, CanOpen: s.st.CanOpen}
	s.mu.Unlock()
	err := s.save()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	c, derr := s.o.Dialer.Dial(ctx, s.loadToken())
	if derr == nil {
		defer c.Close()
		if removeSynced {
			item, gerr := c.Get(ctx, &pb.GetRequest{Key: RecordKey})
			if gerr == nil && len(item.Versions) > 0 {
				_, gerr = c.Delete(ctx, &pb.DeleteRequest{Key: RecordKey, Base: item.Revision})
			}
			if gerr != nil {
				err = errors.Join(err, fmt.Errorf("remove the synced vault: %w", gerr))
			}
		}
		if _, uerr := c.Unpair(ctx, &pb.UnpairRequest{}); uerr != nil && status.Code(uerr) != codes.Unauthenticated {
			s.logf("warn", "unpair: %v", uerr)
		}
	} else if removeSynced {
		err = errors.Join(err, status.Error(codes.Unavailable, "Shard is not running: the synced vault stays there"))
	}
	s.clearToken()
	s.logf("info", "sync via Shard is off")
	s.changed()
	s.Kick()
	if err != nil {
		return s.State(), err
	}
	return s.State(), nil
}

// VaultSwitched says the local vault file is another one now (created
// anew, another file chosen): the next sync compares instead of merging.
func (s *Syncer) VaultSwitched() {
	s.mu.Lock()
	s.sv.Base, s.sv.SyncedHash = "", ""
	s.mu.Unlock()
	s.save()
	s.Kick()
}

// --- the loop --------------------------------------------------------------

func (s *Syncer) loop() {
	defer close(s.done)
	for s.ctx.Err() == nil {
		wait := s.session()
		if wait <= 0 {
			continue
		}
		t := time.NewTimer(wait)
		select {
		case <-s.ctx.Done():
		case <-s.kick:
		case <-t.C:
		}
		t.Stop()
	}
}

// errEnd ends a session (sync went off, the pairing is gone, ...).
var errEnd = errors.New("end")

// errConflict: the user has to choose (status CONFLICT is set).
var errConflict = errors.New("conflict")

// session connects, pairs when asked to and syncs until the connection or
// the sync ends. It returns how long to wait before the next one.
func (s *Syncer) session() time.Duration {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	s.mu.Lock()
	s.sessCancel = cancel
	enabled, wantPair := s.sv.Enabled, s.wantPair
	s.mu.Unlock()

	c, err := s.o.Dialer.Dial(ctx, s.loadToken())
	if err != nil {
		s.unreachable(err)
		return s.RetryDelay
	}
	defer c.Close()
	hello, err := c.Hello(ctx, &pb.HelloRequest{AppId: AppID})
	if err != nil {
		s.unreachable(err)
		return s.RetryDelay
	}
	if hello.ApiVersion > APIVersion {
		s.unreachable(ErrTooNew)
		return s.ProbeDelay
	}
	s.update(func(st *State) {
		st.Available = hello.State != pb.VaultState_NO_VAULT
		st.DeviceName = hello.DeviceName
	})
	if hello.State == pb.VaultState_NO_VAULT {
		s.setStatus(StatusNoVault, "Set up Shard first")
		return s.ProbeDelay
	}
	if !enabled && !wantPair {
		s.setStatus(StatusOff, "")
		return s.ProbeDelay
	}
	if !hello.Paired {
		if !wantPair {
			s.unpaired("The pairing with Shard was revoked: switch sync on to pair again.")
			return s.ProbeDelay
		}
		if !s.pair(ctx, c) {
			return s.RetryDelay
		}
	}
	s.mu.Lock()
	if s.wantPair {
		s.wantPair = false
		s.sv.Enabled = true
	}
	if !s.sv.Enabled {
		s.mu.Unlock()
		return 0
	}
	if s.sv.VaultID != "" && s.sv.VaultID != hello.VaultId {
		// Another Shard vault: never merge across vaults.
		s.sv = saved{}
		s.st.Enabled = false
		s.mu.Unlock()
		s.save()
		s.setStatus(StatusOff, "Shard opened another vault, so sync is off: switch it on to sync with that vault.")
		return s.ProbeDelay
	}
	s.sv.VaultID = hello.VaultId
	s.st.Enabled = true
	s.mu.Unlock()
	s.save()
	s.changed()

	next, err := c.Watch(ctx)
	if err != nil {
		s.failed(err)
		return s.RetryDelay
	}
	type change struct {
		c   *pb.Change
		err error
	}
	changes := make(chan change, 16)
	go func() {
		for {
			ch, err := next()
			select {
			case changes <- change{ch, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	state := hello.State
	for {
		retry := s.Heartbeat
		if state == pb.VaultState_LOCKED {
			s.setStatus(StatusLocked, "Unlock Shard to sync.")
		} else {
			err := s.reconcile(ctx, c, state)
			s.answerPassword(err)
			switch d := s.handle(err); {
			case errors.Is(d.err, errEnd):
				return d.wait
			case d.wait > 0:
				retry = d.wait
			}
		}
		t := time.NewTimer(retry)
		select {
		case <-ctx.Done():
			t.Stop()
			return 0
		case <-s.kick:
		case <-t.C:
		case ch := <-changes:
			if ch.err != nil {
				t.Stop()
				if errors.Is(ch.err, ErrNotRunning) || status.Code(ch.err) == codes.Unavailable {
					s.setStatus(StatusNotRunning, "")
				}
				return s.RetryDelay / 6
			}
			switch ch.c.Kind {
			case pb.Change_STATE:
				h, err := c.Hello(ctx, &pb.HelloRequest{AppId: AppID})
				if err != nil {
					t.Stop()
					return s.RetryDelay
				}
				if !h.Paired {
					t.Stop()
					s.unpaired("The pairing with Shard was revoked: switch sync on to pair again.")
					return s.ProbeDelay
				}
				if h.VaultId != hello.VaultId {
					t.Stop()
					return 0 // the next session turns sync off
				}
				state = h.State
			case pb.Change_UNPAIRED:
				t.Stop()
				s.unpaired("The pairing with Shard was revoked: switch sync on to pair again.")
				return s.ProbeDelay
			}
		}
		t.Stop()
	}
}

// pair asks the user to allow Parley and waits for the answer.
func (s *Syncer) pair(ctx context.Context, c Client) bool {
	t, err := c.RequestPairing(ctx, &pb.PairingRequest{AppId: AppID, AppName: s.o.AppName, Description: s.o.Description})
	if err != nil {
		s.mu.Lock()
		s.wantPair = false
		s.mu.Unlock()
		if status.Code(err) == codes.ResourceExhausted {
			s.setStatus(StatusError, "Too many pairing requests are waiting in Shard: try again later.")
		} else {
			s.failed(err)
		}
		return false
	}
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.mu.Lock()
	s.pairCancel = cancel
	s.mu.Unlock()
	s.update(func(st *State) {
		st.Status, st.PairingCode, st.Message, st.Unpaired = StatusPairing, formatCode(t.Code), "", false
	})
	stop := func(st Status, unpaired bool, msg string) bool {
		s.mu.Lock()
		s.wantPair = false
		s.mu.Unlock()
		s.update(func(x *State) {
			x.Status, x.PairingCode, x.Message, x.Unpaired = st, "", msg, unpaired
		})
		return false
	}
	for {
		r, err := c.AwaitPairing(pctx, &pb.AwaitPairingRequest{Ticket: t.Ticket, WaitSeconds: 60})
		if pctx.Err() != nil {
			s.mu.Lock()
			wanted := s.wantPair
			s.mu.Unlock()
			if !wanted {
				return stop(StatusOff, false, "")
			}
			return false
		}
		if err != nil {
			stop(StatusOff, false, "")
			s.failed(err)
			return false
		}
		switch r.Status {
		case pb.PairingStatus_PENDING:
			continue
		case pb.PairingStatus_APPROVED:
			if s.o.Dialer.UsesToken() {
				if err := s.storeToken(r.Token); err != nil {
					return stop(StatusError, false, "The Shard pairing could not be kept on this device: "+err.Error())
				}
				c.SetToken(r.Token)
			}
			s.mu.Lock()
			s.wantPair = false
			s.sv.Enabled = true
			s.mu.Unlock()
			s.save()
			s.update(func(x *State) { x.Status, x.PairingCode, x.Enabled = StatusSyncing, "", true })
			s.logf("info", "paired with Shard")
			return true
		case pb.PairingStatus_DENIED:
			return stop(StatusOff, true, "Shard did not allow Parley.")
		default:
			return stop(StatusOff, false, "The pairing request expired: try again.")
		}
	}
}

func formatCode(code string) string {
	if len(code) == 6 {
		return code[:3] + " " + code[3:]
	}
	return code
}

// --- reconciling -------------------------------------------------------------

func hashOf(data []byte) string {
	if data == nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// reconcile syncs once, again when someone wrote meanwhile.
func (s *Syncer) reconcile(ctx context.Context, c Client, state pb.VaultState) error {
	var err error
	for range 5 {
		err = s.reconcileOnce(ctx, c, state)
		if status.Code(err) != codes.Aborted && !errors.Is(err, ErrLocalChanged) {
			return err
		}
	}
	return err
}

func (s *Syncer) reconcileOnce(ctx context.Context, c Client, state pb.VaultState) error {
	item, err := c.Get(ctx, &pb.GetRequest{Key: RecordKey})
	if err != nil {
		return err
	}
	local, err := s.o.Host.LocalVault()
	if err != nil {
		return err
	}
	lh := hashOf(local)
	vs := slices.Clone(item.Versions)
	slices.SortFunc(vs, func(a, b *pb.Version) int { return strings.Compare(a.Id, b.Id) })
	s.mu.Lock()
	sv, choice, pw := s.sv, s.resolve, slices.Clone(s.passwords)
	conflicted := s.st.Status == StatusConflict
	s.mu.Unlock()

	if sv.Base == "" && sv.SyncedHash == "" {
		return s.first(ctx, c, item, vs, local, lh, choice, pw, conflicted)
	}
	localChanged := local != nil && lh != sv.SyncedHash
	remoteChanged := item.Revision != sv.Base
	switch {
	case len(vs) == 0 && remoteChanged:
		s.turnOff("The synced vault was removed from Shard, so sync is off here. This device keeps its vault.")
		return errEnd
	case len(vs) == 0:
		if localChanged {
			return s.put(ctx, c, local, item.Revision)
		}
	case local == nil:
		return s.adopt(newest(vs).Data, "", item.Revision, pw)
	case len(vs) == 1 && !remoteChanged:
		if localChanged {
			return s.put(ctx, c, local, sv.Base)
		}
	case len(vs) == 1 && !localChanged:
		return s.adopt(vs[0].Data, lh, item.Revision, pw)
	default:
		return s.merge(ctx, c, vs, item.Revision, pw)
	}
	s.synced(state)
	return nil
}

// first is the first sync after switching it on (CONNECTOR.md § 7).
func (s *Syncer) first(ctx context.Context, c Client, item *pb.Item, vs []*pb.Version, local []byte, lh string, choice Choice, pw [][]byte, conflicted bool) error {
	if len(vs) == 0 {
		if local == nil {
			s.turnOff("Shard has no Parley vault yet. Create one here, then switch sync on in Settings.")
			return errEnd
		}
		return s.put(ctx, c, local, item.Revision) // this device seeds it
	}
	if local == nil {
		return s.adopt(newest(vs).Data, "", item.Revision, pw)
	}
	if len(vs) == 1 && bytes.Equal(vs[0].Data, local) {
		s.markSynced(item.Revision, lh)
		return nil
	}
	switch choice {
	case UseSynced:
		data := newest(vs).Data
		if err := s.o.Host.UseSynced(data, lh, pw); err != nil {
			return err
		}
		s.takeChoice()
		s.markSynced(item.Revision, hashOf(data))
		return nil
	case Upload:
		if err := s.put(ctx, c, local, item.Revision); err != nil {
			return err
		}
		s.takeChoice()
		return nil
	case MergeBoth:
		if err := s.merge(ctx, c, vs, item.Revision, pw); err != nil {
			return err
		}
		s.takeChoice()
		return nil
	}
	if conflicted {
		return errConflict
	}
	same, sameVault, err := s.o.Host.CompareVault(vs[0].Data, pw)
	if err != nil && !errors.Is(err, ErrNeedPassword) {
		return err
	}
	if err == nil && same && len(vs) == 1 {
		return s.adopt(vs[0].Data, lh, item.Revision, pw)
	}
	s.update(func(st *State) {
		st.Status, st.CanMerge = StatusConflict, sameVault
		st.Message = "The vault in Shard differs from this device's."
	})
	return errConflict
}

func newest(vs []*pb.Version) *pb.Version {
	best := vs[0]
	for _, v := range vs[1:] {
		if v.TimeUnixMs > best.TimeUnixMs || v.TimeUnixMs == best.TimeUnixMs && v.Id > best.Id {
			best = v
		}
	}
	return best
}

func (s *Syncer) takeChoice() {
	s.mu.Lock()
	s.resolve = 0
	s.st.CanMerge = false
	s.mu.Unlock()
}

func (s *Syncer) put(ctx context.Context, c Client, data []byte, base string) error {
	if len(data) > MaxRecord {
		return fmt.Errorf("the vault (%d KiB) is larger than Shard keeps (%d KiB)", len(data)>>10, MaxRecord>>10)
	}
	s.setStatus(StatusSyncing, "")
	item, err := c.Put(ctx, &pb.PutRequest{Key: RecordKey, Data: data, Base: base})
	if err != nil {
		return err
	}
	s.markSynced(item.Revision, hashOf(data))
	return nil
}

func (s *Syncer) adopt(data []byte, expect, revision string, pw [][]byte) error {
	if err := s.o.Host.AdoptVault(data, expect, pw); err != nil {
		return err
	}
	s.logf("info", "took the vault synced by another device")
	s.markSynced(revision, hashOf(data))
	return nil
}

func (s *Syncer) merge(ctx context.Context, c Client, vs []*pb.Version, revision string, pw [][]byte) error {
	copies := make([][]byte, 0, len(vs))
	for _, v := range vs {
		copies = append(copies, v.Data)
	}
	s.setStatus(StatusSyncing, "")
	merged, err := s.o.Host.MergeVault(copies, pw)
	if err != nil {
		return err
	}
	s.logf("info", "merged the vault with the copies of other devices")
	return s.put(ctx, c, merged, revision)
}

func (s *Syncer) markSynced(revision, hash string) {
	s.mu.Lock()
	s.sv.Base, s.sv.SyncedHash, s.sv.SyncedMs = revision, hash, time.Now().UnixMilli()
	s.passwords = nil
	s.st.Status, s.st.Message, s.st.SyncedMs, s.st.CanMerge = StatusSynced, "", s.sv.SyncedMs, false
	s.mu.Unlock()
	s.save()
	s.changed()
}

func (s *Syncer) synced(state pb.VaultState) {
	if state == pb.VaultState_OFFLINE {
		s.setStatus(StatusOffline, "Shard has no storage yet: changes wait.")
		return
	}
	s.setStatus(StatusSynced, "")
}

// answerPassword tells a waiting Password call how the sync went.
func (s *Syncer) answerPassword(err error) {
	s.mu.Lock()
	w := s.pwWait
	s.pwWait = nil
	s.mu.Unlock()
	if w != nil {
		w <- err
	}
}

type decision struct {
	wait time.Duration
	err  error
}

// handle shows what a reconcile ended with and says when to retry.
func (s *Syncer) handle(err error) decision {
	switch {
	case err == nil:
		return decision{}
	case errors.Is(err, errEnd):
		return decision{s.ProbeDelay, errEnd}
	case errors.Is(err, errConflict):
		return decision{}
	case errors.Is(err, ErrNeedPassword):
		s.setStatus(StatusNeedsPassword, "The vault in Shard has another password: enter it to sync.")
		return decision{}
	case errors.Is(err, ErrLocked):
		s.setStatus(StatusWaitingUnlock, "Unlock Parley to finish syncing.")
		return decision{}
	case errors.Is(err, ErrOtherVault):
		s.update(func(st *State) {
			st.Status, st.CanMerge = StatusConflict, false
			st.Message = "The vault in Shard is another vault."
		})
		return decision{}
	case errors.Is(err, ErrNotRunning):
		s.setStatus(StatusNotRunning, "")
		return decision{s.RetryDelay, errEnd}
	}
	switch status.Code(err) {
	case codes.Unauthenticated:
		s.unpaired("The pairing with Shard was revoked: switch sync on to pair again.")
		return decision{s.ProbeDelay, errEnd}
	case codes.FailedPrecondition:
		s.setStatus(StatusLocked, "Unlock Shard to sync.")
		return decision{s.RetryDelay, nil}
	case codes.Unavailable:
		s.setStatus(StatusOffline, "Shard has no storage yet: changes wait.")
		return decision{s.RetryDelay * 2, nil}
	case codes.Unimplemented:
		s.setStatus(StatusError, "This Shard is too old: update it.")
		return decision{s.ProbeDelay, errEnd}
	case codes.Canceled:
		return decision{s.RetryDelay / 6, errEnd}
	}
	s.logf("error", "sync: %v", err)
	s.setStatus(StatusError, errText(err))
	return decision{s.RetryDelay * 2, nil}
}

func errText(err error) string {
	if st, ok := status.FromError(err); ok {
		return st.Message()
	}
	return err.Error()
}

// unreachable: Shard is not there (or too new).
func (s *Syncer) unreachable(err error) {
	if errors.Is(err, ErrTooNew) {
		s.update(func(st *State) { st.Status, st.Message, st.Available = StatusError, err.Error(), false })
		return
	}
	s.update(func(st *State) {
		st.Status, st.Available, st.PairingCode = StatusNotRunning, false, ""
		st.Message = ""
	})
}

func (s *Syncer) failed(err error) {
	if errors.Is(err, ErrNotRunning) {
		s.unreachable(err)
		return
	}
	s.setStatus(StatusError, errText(err))
}

// unpaired: the pairing is gone; sync goes off, the vault stays.
func (s *Syncer) unpaired(msg string) {
	s.clearToken()
	s.mu.Lock()
	s.sv = saved{}
	s.wantPair = false
	s.st.Enabled, s.st.Unpaired, s.st.PairingCode = false, true, ""
	s.mu.Unlock()
	s.save()
	s.logf("warn", "%s", msg)
	s.setStatus(StatusOff, msg)
}

// turnOff switches sync off but keeps the pairing.
func (s *Syncer) turnOff(msg string) {
	s.mu.Lock()
	s.sv = saved{VaultID: s.sv.VaultID}
	s.st.Enabled = false
	s.mu.Unlock()
	s.save()
	s.logf("info", "%s", msg)
	s.setStatus(StatusOff, msg)
}

// --- state ---------------------------------------------------------------

func (s *Syncer) setStatus(st Status, msg string) {
	s.update(func(x *State) { x.Status, x.Message = st, msg })
}

func (s *Syncer) update(fn func(*State)) {
	s.mu.Lock()
	before := s.st
	fn(&s.st)
	s.st.Enabled = s.sv.Enabled
	after := s.st
	s.mu.Unlock()
	if before != after {
		s.changed()
	}
}

func (s *Syncer) changed() {
	if s.o.OnChange != nil {
		s.o.OnChange()
	}
}

func (s *Syncer) logf(level, format string, args ...any) {
	if s.o.Log != nil {
		s.o.Log(level, fmt.Sprintf(format, args...))
	}
}

func (s *Syncer) statePath() string { return filepath.Join(s.o.Dir, "shard.json") }

func (s *Syncer) load() {
	data, err := os.ReadFile(s.statePath())
	if err != nil {
		return
	}
	var sv saved
	if json.Unmarshal(data, &sv) == nil {
		s.sv = sv
	}
}

func (s *Syncer) save() error {
	s.mu.Lock()
	data, _ := json.MarshalIndent(s.sv, "", "  ")
	s.mu.Unlock()
	if err := os.MkdirAll(s.o.Dir, 0o700); err != nil {
		return err
	}
	tmp := s.statePath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		s.logf("error", "keep the sync state: %v", err)
		return err
	}
	return os.Rename(tmp, s.statePath())
}

// --- the pairing token (desktops): this device only, never synced ----------

func (s *Syncer) tokenName() string {
	dir, err := filepath.Abs(s.o.Dir)
	if err != nil {
		dir = s.o.Dir
	}
	return "shard:" + dir
}

func (s *Syncer) tokenFile() string { return filepath.Join(s.o.Dir, "shard.token") }

func (s *Syncer) loadToken() string {
	if !s.o.Dialer.UsesToken() {
		return ""
	}
	s.mu.Lock()
	if s.token != nil {
		defer s.mu.Unlock()
		return *s.token
	}
	s.mu.Unlock()
	t := ""
	if b, err := os.ReadFile(s.tokenFile()); err == nil {
		t = strings.TrimSpace(string(b))
	} else if s.o.Keychain != nil {
		if b, err := s.o.Keychain.Get(s.tokenName()); err == nil {
			t = string(b)
		} else if !errors.Is(err, keychain.ErrNotFound) {
			s.logf("warn", "keychain: %v", err)
			return "" // asked again next time
		}
	}
	s.mu.Lock()
	s.token = &t
	s.mu.Unlock()
	return t
}

func (s *Syncer) storeToken(t string) error {
	if s.o.Keychain != nil {
		err := s.o.Keychain.Set(s.tokenName(), []byte(t))
		if err == nil {
			os.Remove(s.tokenFile())
			s.mu.Lock()
			s.token = &t
			s.mu.Unlock()
			return nil
		}
		s.logf("warn", "keychain: %v; the Shard pairing is kept in a file", err)
	}
	if err := os.MkdirAll(s.o.Dir, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(s.tokenFile(), []byte(t), 0o600); err != nil {
		return err
	}
	s.mu.Lock()
	s.token = &t
	s.mu.Unlock()
	return nil
}

func (s *Syncer) clearToken() {
	if !s.o.Dialer.UsesToken() {
		return
	}
	os.Remove(s.tokenFile())
	if s.o.Keychain != nil {
		if err := s.o.Keychain.Delete(s.tokenName()); err != nil {
			s.logf("warn", "keychain: %v", err)
		}
	}
	s.mu.Lock()
	empty := ""
	s.token = &empty
	s.mu.Unlock()
}
