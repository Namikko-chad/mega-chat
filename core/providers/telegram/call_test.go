package telegram

import (
	"context"
	"crypto/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
)

// fakeTG is Telegram's side of 1:1 calls for two sessions, as the docs
// and tdesktop describe it.
type fakeTG struct {
	mu      sync.Mutex
	inbox   map[int64]chan func() // user id → its update queue
	sess    map[int64]*session
	calls   map[int64]*fakeTGCall
	next    int64
	version string // the library version the server picks
	custom  string // custom_parameters
}

type fakeTGCall struct {
	admin, part int64
	access      int64
	video       bool
	gaHash, gb  []byte
}

func newFakeTG() *fakeTG {
	return &fakeTG{inbox: map[int64]chan func(){}, sess: map[int64]*session{}, calls: map[int64]*fakeTGCall{}, next: 1000, version: callVersion, custom: "{}"}
}

// session makes a session of user self that knows the other user.
func (f *fakeTG) session(t *testing.T, self, other int64, name string) (*session, chan model.Event) {
	sink := &recSink{ch: make(chan model.Event, 1024)}
	key := "u" + strconv.FormatInt(other, 10)
	s := &session{
		selfID: self, users: map[string]model.User{key: {ID: key, Name: name}},
		titles: map[string]string{}, kinds: map[string]model.ChatKind{}, msgChat: map[int]string{},
		photos: map[string]int64{}, access: map[string]model.ChatAccess{},
		peers: &peerStore{m: map[string]int64{key: other * 11}},
	}
	s.ph.api = &fakeAPI{f, self}
	s.ph.sink = sink
	q := make(chan func(), 256)
	f.mu.Lock()
	f.inbox[self], f.sess[self] = q, s
	f.mu.Unlock()
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	go func() {
		for {
			select {
			case fn := <-q:
				fn()
			case <-done:
				return
			}
		}
	}()
	return s, sink.ch
}

// push queues an update for a user, in order.
func (f *fakeTG) push(user int64, fn func(s *session)) {
	f.mu.Lock()
	q, s := f.inbox[user], f.sess[user]
	f.mu.Unlock()
	if q != nil {
		q <- func() { fn(s) }
	}
}

func (f *fakeTG) protocol() tg.PhoneCallProtocol {
	p := callProtocol()
	p.LibraryVersions = []string{f.version}
	return p
}

type fakeAPI struct {
	f    *fakeTG
	self int64
}

func (a *fakeAPI) MessagesGetDhConfig(_ context.Context, r *tg.MessagesGetDhConfigRequest) (tg.MessagesDhConfigClass, error) {
	rnd := make([]byte, r.RandomLength)
	_, _ = rand.Read(rnd)
	if r.Version == 1 {
		return &tg.MessagesDhConfigNotModified{Random: rnd}, nil
	}
	return &tg.MessagesDhConfig{G: 3, P: testPrime().Bytes(), Version: 1, Random: rnd}, nil
}

func (a *fakeAPI) HelpGetConfig(context.Context) (*tg.Config, error) {
	return &tg.Config{CallReceiveTimeoutMs: 20000, CallRingTimeoutMs: 90000, CallConnectTimeoutMs: 30000}, nil
}

func (a *fakeAPI) PhoneRequestCall(_ context.Context, r *tg.PhoneRequestCallRequest) (*tg.PhonePhoneCall, error) {
	u := r.UserID.(*tg.InputUser)
	f := a.f
	f.mu.Lock()
	f.next++
	id := f.next
	c := &fakeTGCall{admin: a.self, part: u.UserID, access: id * 7, video: r.Video, gaHash: r.GAHash}
	f.calls[id] = c
	f.mu.Unlock()
	f.push(c.part, func(s *session) {
		s.onPhoneCall(&tg.PhoneCallRequested{ID: id, AccessHash: c.access, Date: int(time.Now().Unix()), AdminID: c.admin, ParticipantID: c.part, GAHash: c.gaHash, Protocol: r.Protocol, Video: r.Video})
	})
	return &tg.PhonePhoneCall{PhoneCall: &tg.PhoneCallWaiting{ID: id, AccessHash: c.access, AdminID: a.self, ParticipantID: c.part, Protocol: r.Protocol}}, nil
}

func (a *fakeAPI) call(id int64) *fakeTGCall {
	a.f.mu.Lock()
	defer a.f.mu.Unlock()
	return a.f.calls[id]
}

func (a *fakeAPI) PhoneReceivedCall(_ context.Context, p tg.InputPhoneCall) (bool, error) {
	c := a.call(p.ID)
	a.f.push(c.admin, func(s *session) {
		s.onPhoneCall(&tg.PhoneCallWaiting{ID: p.ID, AccessHash: c.access, AdminID: c.admin, ParticipantID: c.part, ReceiveDate: int(time.Now().Unix())})
	})
	return true, nil
}

func (a *fakeAPI) PhoneAcceptCall(_ context.Context, r *tg.PhoneAcceptCallRequest) (*tg.PhonePhoneCall, error) {
	c := a.call(r.Peer.ID)
	c.gb = r.GB
	a.f.push(c.admin, func(s *session) {
		s.onPhoneCall(&tg.PhoneCallAccepted{ID: r.Peer.ID, AccessHash: c.access, AdminID: c.admin, ParticipantID: c.part, GB: r.GB, Protocol: a.f.protocol()})
	})
	return &tg.PhonePhoneCall{PhoneCall: &tg.PhoneCallWaiting{ID: r.Peer.ID, AccessHash: c.access, AdminID: c.admin, ParticipantID: c.part, Protocol: a.f.protocol()}}, nil
}

func (a *fakeAPI) PhoneConfirmCall(_ context.Context, r *tg.PhoneConfirmCallRequest) (*tg.PhonePhoneCall, error) {
	c := a.call(r.Peer.ID)
	a.f.mu.Lock()
	custom := a.f.custom
	a.f.mu.Unlock()
	pc := &tg.PhoneCall{
		ID: r.Peer.ID, AccessHash: c.access, AdminID: c.admin, ParticipantID: c.part, GAOrB: r.GA, KeyFingerprint: r.KeyFingerprint,
		Protocol: a.f.protocol(), P2PAllowed: true, StartDate: int(time.Now().Unix()), CustomParameters: tg.DataJSON{Data: custom},
		Connections: []tg.PhoneConnectionClass{
			&tg.PhoneConnection{ID: 1, IP: "91.108.9.1", Port: 530, PeerTag: make([]byte, 16)},
			&tg.PhoneConnectionWebrtc{ID: 2, Turn: true, Stun: true, IP: "91.108.9.2", Port: 3478, Username: "user", Password: "pass"},
		},
	}
	a.f.push(c.part, func(s *session) { s.onPhoneCall(pc) })
	return &tg.PhonePhoneCall{PhoneCall: pc}, nil
}

func (a *fakeAPI) PhoneDiscardCall(_ context.Context, r *tg.PhoneDiscardCallRequest) (tg.UpdatesClass, error) {
	c := a.call(r.Peer.ID)
	other := c.part
	if a.self == c.part {
		other = c.admin
	}
	a.f.push(other, func(s *session) {
		s.onPhoneCall(&tg.PhoneCallDiscarded{ID: r.Peer.ID, Reason: r.Reason, Duration: r.Duration})
	})
	return &tg.Updates{}, nil
}

func (a *fakeAPI) PhoneSendSignalingData(_ context.Context, r *tg.PhoneSendSignalingDataRequest) (bool, error) {
	c := a.call(r.Peer.ID)
	other := c.part
	if a.self == c.part {
		other = c.admin
	}
	a.f.push(other, func(s *session) { s.onSignalingData(r.Peer.ID, r.Data) })
	return true, nil
}

type recSink struct{ ch chan model.Event }

func (r *recSink) Emit(ev model.Event)                { r.ch <- ev }
func (r *recSink) Connected(model.User)               {}
func (r *recSink) SaveParams(map[string]string) error { return nil }

// waitCall waits for the call to reach a state.
func waitCall(t *testing.T, ch chan model.Event, state model.CallState) model.Call {
	t.Helper()
	for {
		select {
		case ev := <-ch:
			if ev.Kind == model.EvCall && ev.Call.State == state {
				return *ev.Call
			}
			if ev.Kind == model.EvCall && ev.Call.State == model.CallEnded {
				t.Fatalf("ended (%d, %q) waiting for state %d", ev.Call.End, ev.Call.Error, state)
			}
		case <-time.After(15 * time.Second):
			t.Fatalf("no state %d", state)
		}
	}
}

func waitEnd(t *testing.T, ch chan model.Event) model.Call {
	t.Helper()
	for {
		select {
		case ev := <-ch:
			if ev.Kind == model.EvCall && ev.Call.State == model.CallEnded {
				return *ev.Call
			}
		case <-time.After(15 * time.Second):
			t.Fatal("not ended")
		}
	}
}

func waitSignal(t *testing.T, ch chan model.Event, kind model.SignalKind) model.CallSignal {
	t.Helper()
	for {
		select {
		case ev := <-ch:
			if ev.Kind == model.EvCallSignal && ev.Signal.Kind == kind {
				return *ev.Signal
			}
			if ev.Kind == model.EvCall && ev.Call.State == model.CallEnded {
				t.Fatalf("ended (%d, %q) waiting for signal %d", ev.Call.End, ev.Call.Error, kind)
			}
		case <-time.After(15 * time.Second):
			t.Fatalf("no signal %d", kind)
		}
	}
}

// ring starts a call from Alice (1) to Bob (2) and waits until it rings.
func ring(t *testing.T, f *fakeTG, video bool) (alice, bob *session, ach, bch chan model.Event, id string) {
	alice, ach = f.session(t, 1, 2, "Bob")
	bob, bch = f.session(t, 2, 1, "Alice")
	ctx := context.Background()
	c, err := alice.StartCall(ctx, "u2", video)
	if err != nil {
		t.Fatal(err)
	}
	if c.State != model.CallWaiting && c.State != model.CallRinging || !c.Outgoing || c.Peer.Name != "Bob" || c.ChatID != "u2" {
		t.Fatalf("started: %+v", c)
	}
	in := waitCall(t, bch, model.CallRinging)
	if in.ID != c.ID || in.Outgoing || in.ChatID != "u1" || in.Peer.Name != "Alice" || in.Video != video {
		t.Fatalf("incoming: %+v", in)
	}
	waitCall(t, ach, model.CallRinging)
	return alice, bob, ach, bch, c.ID
}

func TestCallFlow(t *testing.T) {
	f := newFakeTG()
	alice, bob, ach, bch, id := ring(t, f, true)
	ctx := context.Background()
	if err := bob.AcceptCall(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	ac := waitCall(t, ach, model.CallConnecting)
	bc := waitCall(t, bch, model.CallConnecting)
	if len(ac.Verification) != 4 || strings.Join(ac.Verification, "") != strings.Join(bc.Verification, "") {
		t.Fatalf("emoji %v / %v", ac.Verification, bc.Verification)
	}
	if !ac.Polite || !ac.Offer || !ac.OpenChannel || ac.Channel != "data" || bc.Polite || bc.Offer || bc.OpenChannel || bc.Channel != "data" {
		t.Fatalf("roles %+v / %+v", ac, bc)
	}
	if len(ac.ICE) != 2 || ac.ICE[1].URLs[0] != "turn:91.108.9.2:3478" || ac.ICE[1].Credential != "pass" || ac.RelayOnly {
		t.Fatalf("ice %+v", ac.ICE)
	}

	// Offer, answer and candidates between the two apps.
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(alice.CallSignal(ctx, id, model.CallSignal{Kind: model.SigOffer, SDP: "v=0 offer"}))
	if s := waitSignal(t, bch, model.SigOffer); s.SDP != "v=0 offer" {
		t.Fatalf("offer %+v", s)
	}
	must(bob.CallSignal(ctx, id, model.CallSignal{Kind: model.SigAnswer, SDP: "v=0 answer"}))
	must(bob.CallSignal(ctx, id, model.CallSignal{Kind: model.SigCandidate, SDP: "candidate:9 1 udp 1 reflector-1-00ff00ff.reflector 530 typ relay", Mid: "0"}))
	must(bob.CallSignal(ctx, id, model.CallSignal{Kind: model.SigCandidate, SDP: "a=candidate:1 1 udp 2122260223 10.0.0.2 50000 typ host", Mid: "0", MLine: 0}))
	if s := waitSignal(t, ach, model.SigAnswer); s.SDP != "v=0 answer" {
		t.Fatalf("answer %+v", s)
	}
	if s := waitSignal(t, ach, model.SigCandidate); s.SDP != "candidate:1 1 udp 2122260223 10.0.0.2 50000 typ host" || s.Mid != "0" {
		t.Fatalf("candidate %+v (the reflector one must be dropped)", s)
	}

	// Media state: Alice's goes out for the data channel, Bob's app hands
	// it in.
	must(alice.SetCallMedia(ctx, id, model.CallMedia{Video: true}))
	ms := waitSignal(t, ach, model.SigData)
	must(bob.CallSignal(ctx, id, model.CallSignal{Kind: model.SigData, Data: ms.Data}))
	if r := waitCall(t, bch, model.CallConnecting); r.Remote != (model.CallMedia{Video: true}) {
		t.Fatalf("remote %+v", r.Remote)
	}

	must(alice.CallSignal(ctx, id, model.CallSignal{Kind: model.SigConnected}))
	if a := waitCall(t, ach, model.CallActive); a.Started.IsZero() {
		t.Fatal("no start time")
	}
	must(alice.EndCall(ctx, id, model.EndHangup))
	if e := waitEnd(t, ach); e.End != model.EndHangup {
		t.Fatalf("alice end %d", e.End)
	}
	if e := waitEnd(t, bch); e.End != model.EndHangup {
		t.Fatalf("bob end %d", e.End)
	}
	if alice.ph.get(1001) != nil || bob.ph.get(1001) != nil {
		t.Fatal("calls kept")
	}
}

func TestCallDeclineBusyCancel(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name           string
		callee         bool // who ends it
		reason         model.CallEnd
		aliceEnd, bobE model.CallEnd
	}{
		{"declined", true, model.EndDeclined, model.EndBusy, model.EndDeclined}, // the caller hears "busy"
		{"busy", true, model.EndBusy, model.EndBusy, model.EndBusy},
		{"cancelled", false, model.EndHangup, model.EndDeclined, model.EndMissed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alice, bob, ach, bch, id := ring(t, newFakeTG(), false)
			by := alice
			if tc.callee {
				by = bob
			}
			if err := by.EndCall(ctx, id, tc.reason); err != nil {
				t.Fatal(err)
			}
			if e := waitEnd(t, ach); e.End != tc.aliceEnd {
				t.Fatalf("alice end %d", e.End)
			}
			if e := waitEnd(t, bch); e.End != tc.bobE {
				t.Fatalf("bob end %d", e.End)
			}
		})
	}
}

func TestCallUnsupported(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct{ name, version, custom, want string }{
		{"version", "13.0.0", "{}", "13.0.0"},
		{"mtproto", callVersion, `{"network_use_mtproto":true}`, "transport"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFakeTG()
			f.version, f.custom = tc.version, tc.custom
			_, bob, ach, bch, id := ring(t, f, false)
			if err := bob.AcceptCall(ctx, id, false); err != nil {
				t.Fatal(err)
			}
			e := waitEnd(t, ach)
			if e.End != model.EndFailed || !strings.Contains(e.Error, tc.want) {
				t.Fatalf("alice: %d %q", e.End, e.Error)
			}
			if e := waitEnd(t, bch); e.End != model.EndFailed {
				t.Fatalf("bob: %d %q", e.End, e.Error)
			}
		})
	}
}

func TestCallText(t *testing.T) {
	for _, c := range []struct {
		a    tg.MessageActionPhoneCall
		out  bool
		want string
	}{
		{tg.MessageActionPhoneCall{Duration: 192, Reason: &tg.PhoneCallDiscardReasonHangup{}}, true, "Outgoing call · 3:12"},
		{tg.MessageActionPhoneCall{Duration: 3725, Video: true}, false, "Incoming video call · 1:02:05"},
		{tg.MessageActionPhoneCall{Reason: &tg.PhoneCallDiscardReasonMissed{}}, false, "Missed call"},
		{tg.MessageActionPhoneCall{Reason: &tg.PhoneCallDiscardReasonMissed{}, Video: true}, true, "Cancelled video call"},
		{tg.MessageActionPhoneCall{Reason: &tg.PhoneCallDiscardReasonBusy{}}, false, "Declined call"},
	} {
		if got := callText(&c.a, c.out); got != c.want {
			t.Fatalf("%q, want %q", got, c.want)
		}
	}
	m, ok := convertMessage(dir{}, &tg.MessageService{ID: 5, Out: true, PeerID: &tg.PeerUser{UserID: 2}, Date: 1, Action: &tg.MessageActionPhoneCall{Duration: 61}})
	if !ok || !m.Service || m.Text != "Outgoing call · 1:01" {
		t.Fatalf("%+v", m)
	}
}
