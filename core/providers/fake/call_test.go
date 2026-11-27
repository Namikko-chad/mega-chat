package fake

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

type callSink struct {
	mu     sync.Mutex
	calls  []model.Call
	events chan model.Event
}

func (s *callSink) Emit(ev model.Event) {
	if ev.Kind == model.EvCall || ev.Kind == model.EvCallSignal {
		s.events <- ev
	}
}
func (s *callSink) Connected(model.User)               {}
func (s *callSink) SaveParams(map[string]string) error { return nil }

// The app side of a demo call, played by a second pion peer: it calls the
// contact, negotiates, sends audio and gets it back.
func TestDemoCallEcho(t *testing.T) {
	srv := New()
	srv.Seed("ann", 2, 3)
	sess, _ := srv.Open(provider.Account{ID: "a", Login: "ann"}, provider.Env{})
	sink := &callSink{events: make(chan model.Event, 256)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = sess.Run(ctx, sink) }()
	time.Sleep(50 * time.Millisecond)
	ringAfter, answerAfter = 10*time.Millisecond, 10*time.Millisecond

	cl := sess.(provider.Caller)
	call, err := cl.StartCall(ctx, "c1", false)
	if err != nil || call.State != model.CallWaiting || !call.Offer || !call.OpenChannel {
		t.Fatalf("StartCall: %+v %v", call, err)
	}
	waitState := func(want model.CallState) {
		t.Helper()
		for {
			select {
			case ev := <-sink.events:
				if ev.Kind == model.EvCall && ev.Call.State == want {
					return
				}
				if ev.Kind == model.EvCall && ev.Call.State == model.CallEnded {
					t.Fatalf("ended: %+v", ev.Call)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("no state %v", want)
			}
		}
	}
	waitState(model.CallConnecting)

	app, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	if _, err := app.CreateDataChannel("data", nil); err != nil {
		t.Fatal(err)
	}
	track, _ := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}, "audio", "me")
	if _, err := app.AddTrack(track); err != nil {
		t.Fatal(err)
	}
	echoed := make(chan []byte, 1)
	app.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		p, _, err := tr.ReadRTP()
		if err == nil {
			echoed <- p.Payload
		}
	})
	app.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		j := c.ToJSON()
		_ = cl.CallSignal(ctx, call.ID, model.CallSignal{Kind: model.SigCandidate, SDP: j.Candidate, Mid: *j.SDPMid, MLine: int(*j.SDPMLineIndex)})
	})
	connected := make(chan struct{})
	app.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateConnected {
			close(connected)
		}
	})
	offer, _ := app.CreateOffer(nil)
	if err := app.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	if err := cl.CallSignal(ctx, call.ID, model.CallSignal{Kind: model.SigOffer, SDP: offer.SDP}); err != nil {
		t.Fatal(err)
	}
	// Signals from the contact.
	go func() {
		for ev := range sink.events {
			if ev.Kind != model.EvCallSignal {
				continue
			}
			switch ev.Signal.Kind {
			case model.SigAnswer:
				_ = app.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: ev.Signal.SDP})
			case model.SigCandidate:
				ml := uint16(ev.Signal.MLine)
				_ = app.AddICECandidate(webrtc.ICECandidateInit{Candidate: ev.Signal.SDP, SDPMid: &ev.Signal.Mid, SDPMLineIndex: &ml})
			}
		}
	}()
	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		t.Fatal("not connected")
	}
	payload := []byte{0xf8, 0xff, 0xfe} // an Opus silence frame
	for i := range 100 {
		_ = track.WriteRTP(&rtp.Packet{Header: rtp.Header{Version: 2, PayloadType: 111, SequenceNumber: uint16(i), Timestamp: uint32(i * 960), SSRC: 1}, Payload: payload})
		select {
		case got := <-echoed:
			if string(got) != string(payload) {
				t.Fatalf("echo %x", got)
			}
			if err := cl.EndCall(ctx, call.ID, model.EndHangup); err != nil {
				t.Fatal(err)
			}
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
	t.Fatal("no echo")
}

func TestDemoCallCommand(t *testing.T) {
	if call, video := callCommand(" /videocall "); !call || !video {
		t.Fatal("/videocall")
	}
	if call, _ := callCommand("call me"); call {
		t.Fatal("plain text")
	}
}
