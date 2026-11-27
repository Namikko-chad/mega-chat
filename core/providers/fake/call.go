package fake

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// Demo calls: the contact is a WebRTC peer in the core that sends back
// what it gets (your voice and camera, a moment later). Calling a demo
// contact works like a real call; writing "/call" or "/videocall" in a
// direct demo chat makes the contact call you.

// Timings of the demo peer.
var (
	ringAfter   = 700 * time.Millisecond  // outgoing: the contact's device rings
	answerAfter = 2500 * time.Millisecond // outgoing: the contact answers
	callAfter   = 1500 * time.Millisecond // "/call": the contact calls
	ringFor     = 45 * time.Second        // nobody answers: missed
)

var errNoCall = errors.New("fake: no such call")

type fakeCall struct {
	mu      sync.Mutex
	call    model.Call
	pc      *webrtc.PeerConnection
	pending []webrtc.ICECandidateInit // came before the remote description
	timer   *time.Timer
}

type calls struct {
	mu  sync.Mutex
	by  map[string]*fakeCall // by login + "/" + id
	seq int
}

func (c *session) callKey(id string) string { return c.login + "/" + id }

func (c *session) getCall(id string) (*fakeCall, error) {
	c.s.calls.mu.Lock()
	defer c.s.calls.mu.Unlock()
	fc := c.s.calls.by[c.callKey(id)]
	if fc == nil {
		return nil, errNoCall
	}
	return fc, nil
}

func (c *session) newCall(chatID string, video, outgoing bool) (*fakeCall, error) {
	c.s.mu.Lock()
	ch := c.s.user(c.login).chats[chatID]
	var peer model.User
	if ch != nil {
		peer = model.User{ID: "p" + chatID, Name: ch.Title, Avatar: ch.Avatar}
	}
	c.s.mu.Unlock()
	if ch == nil {
		return nil, fmt.Errorf("fake: no chat %s", chatID)
	}
	if ch.Kind != model.ChatDirect {
		return nil, provider.ErrUnsupported
	}
	c.s.calls.mu.Lock()
	defer c.s.calls.mu.Unlock()
	if c.s.calls.by == nil {
		c.s.calls.by = map[string]*fakeCall{}
	}
	c.s.calls.seq++
	id := "call" + strconv.Itoa(c.s.calls.seq)
	state := model.CallWaiting
	if !outgoing {
		state = model.CallRinging
	}
	fc := &fakeCall{call: model.Call{
		ID: id, ChatID: chatID, Peer: peer, Outgoing: outgoing, Video: video, State: state,
		Verification: []string{"🐼", "🍋", "🚀", "🎸"},
		// The side that offers first is polite and opens the data channel,
		// as in Telegram calls.
		Polite: outgoing, Offer: outgoing, Channel: "data", OpenChannel: outgoing,
	}}
	c.s.calls.by[c.callKey(id)] = fc
	return fc, nil
}

func (c *session) emitCall(fc *fakeCall) {
	fc.mu.Lock()
	call := fc.call
	fc.mu.Unlock()
	c.emit(model.Event{Kind: model.EvCall, Call: &call})
}

func (c *session) emitSignal(fc *fakeCall, sig model.CallSignal) {
	fc.mu.Lock()
	ids := model.Call{ID: fc.call.ID, ChatID: fc.call.ChatID}
	fc.mu.Unlock()
	c.emit(model.Event{Kind: model.EvCallSignal, Call: &ids, Signal: &sig})
}

// setState moves a live call on; false when it is over already.
func (c *session) setState(fc *fakeCall, s model.CallState, end model.CallEnd, why string) bool {
	fc.mu.Lock()
	if !fc.call.Live() {
		fc.mu.Unlock()
		return false
	}
	fc.call.State = s
	if s == model.CallActive && fc.call.Started.IsZero() {
		fc.call.Started = time.Now()
	}
	var pc *webrtc.PeerConnection
	if s == model.CallEnded {
		fc.call.End, fc.call.Error = end, why
		pc, fc.pc = fc.pc, nil
		if fc.timer != nil {
			fc.timer.Stop()
		}
	}
	fc.mu.Unlock()
	if pc != nil {
		_ = pc.Close()
	}
	if s == model.CallEnded {
		c.s.calls.mu.Lock()
		delete(c.s.calls.by, c.callKey(fc.call.ID))
		c.s.calls.mu.Unlock()
	}
	c.emitCall(fc)
	return true
}

func (c *session) StartCall(_ context.Context, chatID string, video bool) (model.Call, error) {
	fc, err := c.newCall(chatID, video, true)
	if err != nil {
		return model.Call{}, err
	}
	go func() {
		time.Sleep(ringAfter)
		if !c.setState(fc, model.CallRinging, 0, "") {
			return
		}
		time.Sleep(answerAfter)
		if _, err := c.peer(fc); err != nil {
			c.setState(fc, model.CallEnded, model.EndFailed, err.Error())
			return
		}
		c.setState(fc, model.CallConnecting, 0, "")
	}()
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.call, nil
}

// ring makes the contact of a direct chat call this account.
func (c *session) ring(chatID string, video bool) {
	time.Sleep(callAfter)
	fc, err := c.newCall(chatID, video, false)
	if err != nil {
		return
	}
	fc.mu.Lock()
	fc.timer = time.AfterFunc(ringFor, func() {
		fc.mu.Lock()
		ringing := fc.call.State == model.CallRinging
		fc.mu.Unlock()
		if ringing {
			c.setState(fc, model.CallEnded, model.EndMissed, "")
		}
	})
	fc.mu.Unlock()
	c.emitCall(fc)
}

func (c *session) AcceptCall(_ context.Context, id string, _ bool) error {
	fc, err := c.getCall(id)
	if err != nil {
		return err
	}
	fc.mu.Lock()
	if fc.call.Outgoing || fc.call.State != model.CallRinging {
		fc.mu.Unlock()
		return fmt.Errorf("fake: call %s is not ringing", id)
	}
	if fc.timer != nil {
		fc.timer.Stop()
	}
	fc.mu.Unlock()
	pc, err := c.peer(fc)
	if err != nil {
		return err
	}
	if !c.setState(fc, model.CallConnecting, 0, "") {
		return errNoCall
	}
	// The contact offers a data channel like Telegram's, audio and video
	// (the app sends video once its camera is on).
	if _, err := pc.CreateDataChannel("data", nil); err != nil {
		return err
	}
	return c.offer(fc, pc)
}

func (c *session) offer(fc *fakeCall, pc *webrtc.PeerConnection) error {
	o, err := pc.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := pc.SetLocalDescription(o); err != nil {
		return err
	}
	c.emitSignal(fc, model.CallSignal{Kind: model.SigOffer, SDP: o.SDP})
	return nil
}

func (c *session) EndCall(_ context.Context, id string, reason model.CallEnd) error {
	fc, err := c.getCall(id)
	if err != nil {
		return err
	}
	c.setState(fc, model.CallEnded, reason, "")
	return nil
}

func (c *session) SetCallMedia(_ context.Context, id string, m model.CallMedia) error {
	fc, err := c.getCall(id)
	if err != nil {
		return err
	}
	// The contact sends back what it gets: it "turns its camera on" with
	// yours.
	fc.mu.Lock()
	fc.call.Remote = model.CallMedia{Video: m.Video || m.Screen}
	fc.mu.Unlock()
	c.emitCall(fc)
	return nil
}

func (c *session) CallSignal(_ context.Context, id string, sig model.CallSignal) error {
	fc, err := c.getCall(id)
	if err != nil {
		return err
	}
	fc.mu.Lock()
	pc := fc.pc
	fc.mu.Unlock()
	switch sig.Kind {
	case model.SigConnected:
		c.setState(fc, model.CallActive, 0, "")
		return nil
	case model.SigFailed:
		c.setState(fc, model.CallEnded, model.EndFailed, sig.Data)
		return nil
	case model.SigData:
		return nil
	}
	if pc == nil {
		return errNoCall
	}
	switch sig.Kind {
	case model.SigOffer:
		if pc.SignalingState() == webrtc.SignalingStateHaveLocalOffer {
			// The contact is the impolite one: it keeps its own offer.
			return nil
		}
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sig.SDP}); err != nil {
			return err
		}
		c.flushCandidates(fc, pc)
		a, err := pc.CreateAnswer(nil)
		if err != nil {
			return err
		}
		if err := pc.SetLocalDescription(a); err != nil {
			return err
		}
		c.emitSignal(fc, model.CallSignal{Kind: model.SigAnswer, SDP: a.SDP})
	case model.SigAnswer:
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: sig.SDP}); err != nil {
			return err
		}
		c.flushCandidates(fc, pc)
	case model.SigCandidate:
		if sig.SDP == "" {
			return nil
		}
		mline := uint16(sig.MLine)
		cand := webrtc.ICECandidateInit{Candidate: sig.SDP, SDPMid: &sig.Mid, SDPMLineIndex: &mline}
		if pc.RemoteDescription() == nil {
			fc.mu.Lock()
			fc.pending = append(fc.pending, cand)
			fc.mu.Unlock()
			return nil
		}
		return pc.AddICECandidate(cand)
	}
	return nil
}

func (c *session) flushCandidates(fc *fakeCall, pc *webrtc.PeerConnection) {
	fc.mu.Lock()
	pending := fc.pending
	fc.pending = nil
	fc.mu.Unlock()
	for _, cand := range pending {
		_ = pc.AddICECandidate(cand)
	}
}

// peer makes the contact's peer connection: it sends back the audio and
// video it receives.
func (c *session) peer(fc *fakeCall) (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	// Only codecs that can be echoed as they are: what the app sends is
	// what it gets.
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1"},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, RTCPFeedback: []webrtc.RTCPFeedback{{Type: "nack"}, {Type: "nack", Parameter: "pli"}}},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}
	pc, err := webrtc.NewAPI(webrtc.WithMediaEngine(m)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}
	audio, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}, "audio", "echo")
	if err != nil {
		return nil, err
	}
	video, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000}, "video", "echo")
	if err != nil {
		return nil, err
	}
	for _, t := range []webrtc.TrackLocal{audio, video} {
		s, err := pc.AddTrack(t)
		if err != nil {
			_ = pc.Close()
			return nil, err
		}
		go drainRTCP(s)
	}
	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		dst := audio
		if t.Kind() == webrtc.RTPCodecTypeVideo {
			dst = video
			// Ask for key frames now and then, so the echo starts soon.
			go func() {
				for range time.Tick(2 * time.Second) {
					if pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(t.SSRC())}}) != nil {
						return
					}
				}
			}()
		}
		for {
			p, _, err := t.ReadRTP()
			if err != nil {
				return
			}
			_ = dst.WriteRTP(p) // nobody to send to yet: dropped
		}
	})
	pc.OnICECandidate(func(cand *webrtc.ICECandidate) {
		if cand == nil {
			return
		}
		j := cand.ToJSON()
		sig := model.CallSignal{Kind: model.SigCandidate, SDP: j.Candidate}
		if j.SDPMid != nil {
			sig.Mid = *j.SDPMid
		}
		if j.SDPMLineIndex != nil {
			sig.MLine = int(*j.SDPMLineIndex)
		}
		c.emitSignal(fc, sig)
	})
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateFailed {
			c.setState(fc, model.CallEnded, model.EndFailed, "the connection failed")
		}
	})
	fc.mu.Lock()
	fc.pc = pc
	fc.mu.Unlock()
	return pc, nil
}

func drainRTCP(s *webrtc.RTPSender) {
	buf := make([]byte, 1500)
	for {
		if _, _, err := s.Read(buf); err != nil {
			return
		}
	}
}

// callCommand reports whether text asks the demo contact to call.
func callCommand(text string) (call, video bool) {
	switch strings.TrimSpace(text) {
	case "/call":
		return true, false
	case "/videocall":
		return true, true
	}
	return false, false
}
