package telegram

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
)

func TestMediaStateJSON(t *testing.T) {
	// json11 writes the keys sorted; tgcalls reads them in any order.
	got := mediaStateJSON(model.CallMedia{Muted: true, Video: true})
	want := `{"@type":"MediaState","lowBattery":false,"muted":true,"screencastState":"inactive","videoRotation":0,"videoState":"active"}`
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
	m, ok := parseMediaState(`{"videoState":"suspended","@type":"MediaState","muted":false,"lowBattery":true,"videoRotation":90,"screencastState":"active"}`)
	if !ok || m != (model.CallMedia{Screen: true}) {
		t.Fatalf("parsed %+v %v", m, ok)
	}
	if m, ok := parseMediaState(`{"@type":"MediaState","muted":true}`); !ok || m != (model.CallMedia{Muted: true}) {
		t.Fatalf("short: %+v %v", m, ok)
	}
	for _, bad := range []string{`{"@type":"Candidates"}`, `nope`, `{"muted":true}`} {
		if _, ok := parseMediaState(bad); ok {
			t.Fatalf("%s parsed", bad)
		}
	}
}

func TestCallICE(t *testing.T) {
	conns := []tg.PhoneConnectionClass{
		&tg.PhoneConnection{ID: 1, IP: "91.108.9.1", Port: 530, PeerTag: make([]byte, 16)}, // reflector
		&tg.PhoneConnectionWebrtc{ID: 2, Stun: true, IP: "91.108.9.2", Ipv6: "2001:b28:f23f::2", Port: 3478},
		&tg.PhoneConnectionWebrtc{ID: 3, Turn: true, Stun: true, IP: "turn.example.org", Port: 3478, Username: "u", Password: "p"},
		&tg.PhoneConnectionWebrtc{ID: 4, Turn: true, IP: "10.0.0.1", Port: 3478}, // no credentials
	}
	got := iceServers(conns)
	want := []model.IceServer{
		{URLs: []string{"stun:91.108.9.2:3478", "stun:[2001:b28:f23f::2]:3478"}},
		{URLs: []string{"stun:turn.example.org:3478"}},
		{URLs: []string{"turn:turn.example.org:3478"}, Username: "u", Credential: "p"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if strings.Join(got[i].URLs, " ") != strings.Join(want[i].URLs, " ") || got[i].Username != want[i].Username || got[i].Credential != want[i].Credential {
			t.Fatalf("%d: %+v, want %+v", i, got[i], want[i])
		}
	}
	for sdp, refl := range map[string]bool{
		"candidate:1 1 udp 2122260223 192.168.1.2 51000 typ host generation 0":                                      false,
		"candidate:2 1 udp 41885439 reflector-1-a1b2c3d4.reflector 530 typ relay raddr 0.0.0.0 rport 0":             true,
		"a=candidate:2 1 udp 41885439 REFLECTOR-2-00000000.Reflector 530 typ relay":                                 true,
		"candidate:3 1 udp 1686052607 5f0e1b4c-8d55-4a0e-9a11-3b5c2f3e7d10.local 51000 typ host generation 0 ufrag": false,
		"": false,
	} {
		if reflectorCandidate(sdp) != refl {
			t.Fatalf("%q: %v", sdp, !refl)
		}
	}
	for sdp, kind := range map[string]string{
		"candidate:1 1 udp 2122260223 192.168.1.2 51000 typ host generation 0":                          "udp host",
		"a=candidate:4 1 UDP 16777215 1.2.3.4 3478 typ relay raddr 0.0.0.0 rport 0":                     "udp relay",
		"candidate:2 1 udp 41885439 reflector-1-a1b2c3d4.reflector 530 typ relay raddr 0.0.0.0 rport 0": "reflector",
		"": "? ?",
	} {
		if got := candidateKind(sdp); got != kind {
			t.Fatalf("%q: %q, want %q", sdp, got, kind)
		}
	}
	if !customFlag(`{"network_use_mtproto":true,"x":1}`, "network_use_mtproto") || customFlag(`{"network_use_mtproto":false}`, "network_use_mtproto") || customFlag(``, "x") {
		t.Fatal("custom parameters")
	}
}

// relay wires two signalling stacks like Telegram does: each packet one
// sendSignalingData, delivered in order.
type relay struct {
	mu      sync.Mutex
	a, b    *signaling
	drop    atomic.Int32 // packets to lose, from the start
	packets atomic.Int32
	data    atomic.Int32 // DATA chunks on stream 0, PPID 53
	bad     atomic.Value
}

func (r *relay) inspect(p []byte) {
	r.packets.Add(1)
	if len(p) < 12 || binary.BigEndian.Uint16(p) != 5000 || binary.BigEndian.Uint16(p[2:]) != 5000 {
		r.bad.Store("not an SCTP packet for ports 5000/5000")
		return
	}
	for c := p[12:]; len(c) >= 4; {
		n := int(binary.BigEndian.Uint16(c[2:]))
		if n < 4 || n > len(c) {
			r.bad.Store("bad chunk length")
			return
		}
		if c[0] == 64 {
			r.bad.Store("I-DATA: dcsctp has message interleaving off")
		}
		if c[0] == 0 && n >= 16 { // DATA
			if binary.BigEndian.Uint16(c[8:]) != 0 || binary.BigEndian.Uint32(c[12:]) != 53 {
				r.bad.Store("DATA not on stream 0 with PPID 53")
			}
			r.data.Add(1)
		}
		c = c[(n+3)&^3:]
	}
}

func (r *relay) sender(to func() *signaling) func(context.Context, []byte) error {
	return func(_ context.Context, p []byte) error {
		r.inspect(p)
		if r.drop.Add(-1) >= 0 {
			return nil
		}
		r.mu.Lock()
		dst := to()
		r.mu.Unlock()
		dst.receive(p)
		return nil
	}
}

func newPair(t *testing.T) (*relay, chan sigMessage, chan sigMessage) {
	t.Helper()
	key := testKey()
	r := &relay{}
	fromCaller, fromCallee := make(chan sigMessage, 64), make(chan sigMessage, 64)
	logf := func(string, string, ...any) {} // pion may log after the test ended
	r.mu.Lock()
	r.a = newSignaling(key, true, r.sender(func() *signaling { return r.b }), func(m sigMessage) { fromCallee <- m }, logf)
	r.b = newSignaling(key, false, r.sender(func() *signaling { return r.a }), func(m sigMessage) { fromCaller <- m }, logf)
	r.mu.Unlock()
	t.Cleanup(func() { r.a.close(); r.b.close() })
	return r, fromCaller, fromCallee
}

func recv(t *testing.T, ch chan sigMessage) sigMessage {
	t.Helper()
	select {
	case m := <-ch:
		return m
	case <-time.After(15 * time.Second):
		t.Fatal("no message")
	}
	return sigMessage{}
}

func TestSignalingBackToBack(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode sigMode
		drop int32
	}{
		{"responder", sigRespond, 0},
		// A tgcalls callee sends its own INIT when the first packet comes.
		{"crossed INITs", sigInitiateLate, 0},
		// The callee's engine did not exist yet: the first INIT is lost.
		{"lost INIT", sigRespond, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, fromCaller, fromCallee := newPair(t)
			r.drop.Store(tc.drop)
			// The caller's offer is queued before the association is up.
			offer := "v=0\r\no=- 1 2 IN IP4 127.0.0.1\r\n" + strings.Repeat("a=x-padding:<&>\r\n", 400) + "a=x-noise:" + noise(6000) + "\r\n"
			if err := r.a.sendJSON(sdpMessage{Type: "offer", SDP: offer}); err != nil {
				t.Fatal(err)
			}
			r.a.start(sigInitiate)
			r.b.start(tc.mode)
			if m := recv(t, fromCaller); m.Type != "offer" || m.SDP != offer {
				t.Fatalf("offer: %+v", m.Type)
			}
			_ = r.b.sendJSON(sdpMessage{Type: "answer", SDP: "v=0 answer"})
			_ = r.b.sendJSON(candidateMessage{Type: "candidate", Mid: "0", MLine: 0, SDP: "candidate:1 1 udp 1 10.0.0.2 5000 typ host"})
			if m := recv(t, fromCallee); m.Type != "answer" || m.SDP != "v=0 answer" {
				t.Fatalf("answer: %+v", m)
			}
			if m := recv(t, fromCallee); m.Type != "candidate" || m.Mid != "0" || m.MLine != 0 || !strings.HasPrefix(m.SDP, "candidate:1 ") {
				t.Fatalf("candidate: %+v", m)
			}
			_ = r.a.sendJSON(candidateMessage{Type: "candidate", Mid: "1", MLine: 1, SDP: "candidate:2 1 udp 1 10.0.0.1 5000 typ host"})
			if m := recv(t, fromCaller); m.Type != "candidate" || m.Mid != "1" || m.MLine != 1 {
				t.Fatalf("candidate: %+v", m)
			}
			if v := r.bad.Load(); v != nil {
				t.Fatal(v)
			}
			if r.data.Load() < 4 {
				t.Fatalf("%d DATA chunks", r.data.Load())
			}
		})
	}
}

// noise is text gzip cannot shrink much: the message spans several SCTP
// packets.
func noise(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func TestSignalingRawJSON(t *testing.T) {
	// tgcalls takes a message that is not gzipped as is.
	got := make(chan sigMessage, 1)
	g := newSignaling(testKey(), false, nil, func(m sigMessage) { got <- m }, func(string, string, ...any) {})
	peer := newCallCipher(testKey(), true)
	g.handle(peer.seal([]byte(`{"@type":"answer","sdp":"v=0"}`)))
	g.handle(peer.seal([]byte(`not json`)))
	if m := <-got; m.Type != "answer" || m.SDP != "v=0" {
		t.Fatalf("%+v", m)
	}
	if len(got) != 0 {
		t.Fatal("garbage delivered")
	}
}
