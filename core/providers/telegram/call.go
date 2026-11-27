package telegram

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"megachat/core/model"
	"megachat/core/provider"
)

// 1:1 calls. The core does the key exchange, the call's state and the
// signalling of library version 11.0.0 (tgcalls InstanceV2ReferenceImpl,
// a stock WebRTC peer connection on the other side); the app runs the
// peer connection. Only 11.0.0 is offered, so the server picks it for
// every peer that has it (all current official apps).

const callVersion = "11.0.0"

// Timeouts when help.getConfig gives none.
const (
	defReceiveTimeout = 20 * time.Second // outgoing: until the peer's device rings
	defRingTimeout    = 90 * time.Second
	defConnectTimeout = 30 * time.Second
)

const (
	maxEarlyCalls   = 32  // unknown call ids kept for a while
	maxEarlyPackets = 64  // signalling packets kept per call before its channel
	maxQueuedSigs   = 256 // app signals kept before the channel exists
	earlyFor        = 30 * time.Second
)

func callProtocol() tg.PhoneCallProtocol {
	// Layers as tdesktop: kMinLayer 65, tgcalls Meta::MaxLayer() 92.
	return tg.PhoneCallProtocol{UDPP2P: true, UDPReflector: true, MinLayer: 65, MaxLayer: 92, LibraryVersions: []string{callVersion}}
}

var errNoCall = errors.New("telegram: no such call")

// phoneAPI is what calls need of the client (*tg.Client); tests fake it.
type phoneAPI interface {
	MessagesGetDhConfig(ctx context.Context, request *tg.MessagesGetDhConfigRequest) (tg.MessagesDhConfigClass, error)
	HelpGetConfig(ctx context.Context) (*tg.Config, error)
	PhoneRequestCall(ctx context.Context, request *tg.PhoneRequestCallRequest) (*tg.PhonePhoneCall, error)
	PhoneReceivedCall(ctx context.Context, peer tg.InputPhoneCall) (bool, error)
	PhoneAcceptCall(ctx context.Context, request *tg.PhoneAcceptCallRequest) (*tg.PhonePhoneCall, error)
	PhoneConfirmCall(ctx context.Context, request *tg.PhoneConfirmCallRequest) (*tg.PhonePhoneCall, error)
	PhoneDiscardCall(ctx context.Context, request *tg.PhoneDiscardCallRequest) (tg.UpdatesClass, error)
	PhoneSendSignalingData(ctx context.Context, request *tg.PhoneSendSignalingDataRequest) (bool, error)
}

type callTimes struct{ receive, ring, connect time.Duration }

// earlyCall is what came for a call id the session does not know (yet).
type earlyCall struct {
	at    time.Time
	calls []tg.PhoneCallClass
	data  [][]byte
}

// phone is the session's call state.
type phone struct {
	api phoneAPI // tests; nil: the live client

	mu    sync.Mutex
	sink  provider.Sink
	calls map[int64]*tgCall
	early map[int64]*earlyCall
	dh    *dhParams
	times *callTimes
}

type tgCall struct {
	id, access, user int64

	mu       sync.Mutex
	call     model.Call
	dh       dhParams
	secret   []byte // a or b
	ga       []byte
	gaHash   []byte // callee: SHA256(g_a) from the request
	key      []byte // the shared key, once known
	fp       int64
	accepted bool // callee: answered; caller: the peer answered
	sig      *signaling
	early    [][]byte           // signalling packets before sig
	queued   []model.CallSignal // app signals before sig
	timer    *time.Timer
	gen      int // invalidates a superseded timer
	diag     callDiag
}

// callDiag counts the signals of a call by kind, for the log when it does
// not connect: offers, answers and candidates by type, never their content.
type callDiag struct{ app, peer map[string]int }

func (d *callDiag) note(fromApp bool, what string) {
	m := &d.peer
	if fromApp {
		m = &d.app
	}
	if *m == nil {
		*m = map[string]int{}
	}
	(*m)[what]++
}

func (d *callDiag) String() string {
	list := func(m map[string]int) string {
		if len(m) == 0 {
			return "nothing"
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = fmt.Sprintf("%s ×%d", k, m[k])
		}
		return strings.Join(parts, ", ")
	}
	return "from the app: " + list(d.app) + "; from the peer: " + list(d.peer)
}

func (c *tgCall) diagLocked() string {
	out := c.diag.String()
	if c.sig != nil {
		out += "; " + c.sig.stats()
	} else {
		out += "; no signalling channel"
	}
	return out
}

func (c *tgCall) input() tg.InputPhoneCall { return tg.InputPhoneCall{ID: c.id, AccessHash: c.access} }

func (p *phone) get(id int64) *tgCall {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls[id]
}

// add registers a call and hands back what came for it before.
func (p *phone) add(c *tgCall) *earlyCall {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.calls == nil {
		p.calls = map[int64]*tgCall{}
	}
	p.calls[c.id] = c
	e := p.early[c.id]
	delete(p.early, c.id)
	return e
}

func (p *phone) remove(id int64) {
	p.mu.Lock()
	delete(p.calls, id)
	p.mu.Unlock()
}

// stash keeps an update or a packet of an unknown call for a while: it
// can come before the reply that makes the call known.
func (p *phone) stash(id int64, pc tg.PhoneCallClass, data []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for k, e := range p.early {
		if now.Sub(e.at) > earlyFor {
			delete(p.early, k)
		}
	}
	e := p.early[id]
	if e == nil {
		if len(p.early) >= maxEarlyCalls {
			return
		}
		if p.early == nil {
			p.early = map[int64]*earlyCall{}
		}
		e = &earlyCall{at: now}
		p.early[id] = e
	}
	if pc != nil && len(e.calls) < 8 {
		e.calls = append(e.calls, pc)
	}
	if data != nil && len(e.data) < maxEarlyPackets {
		e.data = append(e.data, data)
	}
}

func (p *phone) cachedTimes() callTimes {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.times != nil {
		return *p.times
	}
	return callTimes{defReceiveTimeout, defRingTimeout, defConnectTimeout}
}

// --- session plumbing ----------------------------------------------------

func (s *session) phoneLive() (phoneAPI, *updates.Manager, error) {
	if s.ph.api != nil {
		return s.ph.api, nil, nil
	}
	api, gaps, err := s.live()
	if err != nil {
		return nil, nil, err
	}
	return api, gaps, nil
}

func (s *session) callEmit(ev model.Event) {
	s.ph.mu.Lock()
	sink := s.ph.sink
	s.ph.mu.Unlock()
	if sink != nil {
		sink.Emit(ev)
	}
}

// emitLocked reports the call; it runs under c.mu so that the events of a
// call keep their order.
func (s *session) emitLocked(c *tgCall) {
	call := c.call
	call.ICE = slices.Clone(call.ICE)
	call.Verification = slices.Clone(call.Verification)
	s.callEmit(model.Event{Kind: model.EvCall, Call: &call})
}

func (s *session) emitSignalLocked(c *tgCall, sig model.CallSignal) {
	s.callEmit(model.Event{Kind: model.EvCallSignal, Call: &model.Call{ID: c.call.ID, ChatID: c.call.ChatID}, Signal: &sig})
}

// armLocked starts the call's one timer; fn runs under c.mu unless the
// timer was replaced or stopped meanwhile.
func (s *session) armLocked(c *tgCall, d time.Duration, fn func()) {
	s.stopLocked(c)
	gen := c.gen
	c.timer = time.AfterFunc(d, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.gen == gen && c.call.Live() {
			fn()
		}
	})
}

func (s *session) stopLocked(c *tgCall) {
	c.gen++
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

// finishLocked ends the call here; false when it was over already.
func (s *session) finishLocked(c *tgCall, end model.CallEnd, why string) bool {
	if !c.call.Live() {
		return false
	}
	if end == model.EndFailed && c.key != nil {
		s.env.Logf("info", "call %d: failed while %s: %s", c.id, stateName(c.call.State), c.diagLocked())
	}
	c.call.State, c.call.End, c.call.Error = model.CallEnded, end, why
	s.stopLocked(c)
	if c.sig != nil {
		c.sig.close()
		c.sig = nil
	}
	c.early, c.queued = nil, nil
	s.ph.remove(c.id)
	s.emitLocked(c)
	return true
}

// failLocked ends the call and tells the server, in the background.
func (s *session) failLocked(c *tgCall, end model.CallEnd, why string, reason tg.PhoneCallDiscardReasonClass) {
	if why != "" {
		s.env.Logf("warn", "call %d: %s", c.id, why)
	}
	if s.finishLocked(c, end, why) {
		req := s.discardRequestLocked(c, reason)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_ = s.discard(ctx, req)
		}()
	}
}

func (s *session) discardRequestLocked(c *tgCall, reason tg.PhoneCallDiscardReasonClass) *tg.PhoneDiscardCallRequest {
	dur := 0
	if !c.call.Started.IsZero() {
		dur = int(time.Since(c.call.Started).Seconds())
	}
	return &tg.PhoneDiscardCallRequest{Video: c.call.Video, Peer: c.input(), Duration: dur, Reason: reason}
}

func (s *session) discard(ctx context.Context, req *tg.PhoneDiscardCallRequest) error {
	api, gaps, err := s.phoneLive()
	if err != nil {
		return err
	}
	res, err := api.PhoneDiscardCall(ctx, req)
	if err != nil {
		if tgerr.Is(err, "CALL_ALREADY_DECLINED", "CALL_ALREADY_ACCEPTED", "CALL_PEER_INVALID") {
			return nil // over on the server already
		}
		s.env.Logf("warn", "call %d: discard: %v", req.Peer.ID, err)
		return authErr(err)
	}
	if gaps != nil {
		if err := gaps.Handle(ctx, res); err != nil {
			s.env.Logf("warn", "updates after discard: %v", err)
		}
	}
	return nil
}

func (s *session) dhConfig(ctx context.Context, api phoneAPI) (dhParams, []byte, error) {
	s.ph.mu.Lock()
	cur := s.ph.dh
	s.ph.mu.Unlock()
	ver := 0
	if cur != nil {
		ver = cur.version
	}
	res, err := api.MessagesGetDhConfig(ctx, &tg.MessagesGetDhConfigRequest{Version: ver, RandomLength: callKeySize})
	if err != nil {
		return dhParams{}, nil, authErr(err)
	}
	switch r := res.(type) {
	case *tg.MessagesDhConfig:
		p := new(big.Int).SetBytes(r.P)
		if err := crypto.CheckDH(r.G, p); err != nil {
			return dhParams{}, nil, fmt.Errorf("telegram: bad DH parameters: %w", err)
		}
		d := dhParams{g: r.G, p: p, version: r.Version}
		s.ph.mu.Lock()
		s.ph.dh = &d
		s.ph.mu.Unlock()
		return d, r.Random, nil
	case *tg.MessagesDhConfigNotModified:
		if cur == nil {
			return dhParams{}, nil, errors.New("telegram: no DH parameters")
		}
		return *cur, r.Random, nil
	}
	return dhParams{}, nil, errors.New("telegram: no DH parameters")
}

func (s *session) callTimes(ctx context.Context, api phoneAPI) callTimes {
	s.ph.mu.Lock()
	t := s.ph.times
	s.ph.mu.Unlock()
	if t != nil {
		return *t
	}
	out := callTimes{defReceiveTimeout, defRingTimeout, defConnectTimeout}
	cfg, err := api.HelpGetConfig(ctx)
	if err != nil {
		return out
	}
	ms := func(v int, d *time.Duration) {
		if v > 0 {
			*d = time.Duration(v) * time.Millisecond
		}
	}
	ms(cfg.CallReceiveTimeoutMs, &out.receive)
	ms(cfg.CallRingTimeoutMs, &out.ring)
	ms(cfg.CallConnectTimeoutMs, &out.connect)
	s.ph.mu.Lock()
	s.ph.times = &out
	s.ph.mu.Unlock()
	return out
}

func newCall(id, access, user int64, chatID string, peer model.User, outgoing, video bool, state model.CallState) *tgCall {
	// Roles as tgcalls: the caller offers first, opens the "data" channel
	// and is the polite peer (the callee ignores colliding offers).
	return &tgCall{id: id, access: access, user: user, call: model.Call{
		ID: strconv.FormatInt(id, 10), ChatID: chatID, Peer: peer, Outgoing: outgoing, Video: video, State: state,
		Polite: outgoing, Offer: outgoing, Channel: "data", OpenChannel: outgoing,
	}}
}

func (s *session) replay(c *tgCall, e *earlyCall) {
	if e == nil {
		return
	}
	c.mu.Lock()
	c.early = append(c.early, e.data...)
	c.mu.Unlock()
	for _, pc := range e.calls {
		s.applyPhoneCall(c, pc)
	}
}

// --- updates -------------------------------------------------------------

// dropCalls ends the calls when the session stops (the hub ends them for
// the app); there is no one to tell, the server ends them by timeout.
func (s *session) dropCalls() {
	s.ph.mu.Lock()
	s.ph.sink = nil
	calls := make([]*tgCall, 0, len(s.ph.calls))
	for _, c := range s.ph.calls {
		calls = append(calls, c)
	}
	s.ph.mu.Unlock()
	for _, c := range calls {
		c.mu.Lock()
		s.finishLocked(c, model.EndFailed, "The connection to Telegram was lost")
		c.mu.Unlock()
	}
}

func (s *session) handleCalls(d tg.UpdateDispatcher, sink provider.Sink) {
	s.ph.mu.Lock()
	s.ph.sink = sink
	s.ph.mu.Unlock()
	d.OnPhoneCall(func(_ context.Context, e tg.Entities, u *tg.UpdatePhoneCall) error {
		s.learnEntities(e)
		s.onPhoneCall(u.PhoneCall)
		return nil
	})
	d.OnPhoneCallSignalingData(func(_ context.Context, _ tg.Entities, u *tg.UpdatePhoneCallSignalingData) error {
		s.onSignalingData(u.PhoneCallID, u.Data)
		return nil
	})
}

func (s *session) onPhoneCall(pc tg.PhoneCallClass) {
	if r, ok := pc.(*tg.PhoneCallRequested); ok {
		s.incoming(r)
		return
	}
	c := s.ph.get(pc.GetID())
	if c == nil {
		s.ph.stash(pc.GetID(), pc, nil)
		return
	}
	s.applyPhoneCall(c, pc)
}

func (s *session) onSignalingData(id int64, data []byte) {
	c := s.ph.get(id)
	if c == nil {
		s.ph.stash(id, nil, data)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	switch {
	case c.sig != nil:
		c.sig.receive(data)
	case c.call.Live() && len(c.early) < maxEarlyPackets:
		c.early = append(c.early, data)
	}
}

func (s *session) applyPhoneCall(c *tgCall, pc tg.PhoneCallClass) {
	switch p := pc.(type) {
	case *tg.PhoneCallWaiting:
		c.mu.Lock()
		if c.call.Outgoing && c.call.State == model.CallWaiting && p.ReceiveDate != 0 {
			c.call.State = model.CallRinging
			s.armLocked(c, s.ph.cachedTimes().ring, func() {
				s.failLocked(c, model.EndMissed, "", &tg.PhoneCallDiscardReasonMissed{})
			})
			s.emitLocked(c)
		}
		c.mu.Unlock()
	case *tg.PhoneCallAccepted:
		s.accepted(c, p)
	case *tg.PhoneCall:
		s.confirmed(c, p)
	case *tg.PhoneCallDiscarded:
		s.discarded(c, p)
	case *tg.PhoneCallEmpty:
		c.mu.Lock()
		s.finishLocked(c, model.EndFailed, "The call is no longer available")
		c.mu.Unlock()
	}
}

// incoming: someone calls. The caller's device shows "ringing" once
// phone.receivedCall is sent.
func (s *session) incoming(r *tg.PhoneCallRequested) {
	if r.ParticipantID != s.selfID || len(r.GAHash) != sha256.Size || s.ph.get(r.ID) != nil {
		return
	}
	if time.Unix(int64(r.Date), 0).Add(s.ph.cachedTimes().ring).Before(time.Now()) {
		return // an old request (the session was offline)
	}
	key := "u" + strconv.FormatInt(r.AdminID, 10)
	c := newCall(r.ID, r.AccessHash, r.AdminID, key, senderOf(s, key), false, r.Video, model.CallRinging)
	c.gaHash = bytes.Clone(r.GAHash)
	e := s.ph.add(c)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		api, _, err := s.phoneLive()
		if err == nil {
			s.callTimes(ctx, api)
			_, err = api.PhoneReceivedCall(ctx, c.input())
		}
		c.mu.Lock()
		if !c.call.Live() {
			c.mu.Unlock()
			return
		}
		if err != nil {
			s.env.Logf("warn", "call %d: received: %v", c.id, err)
			s.finishLocked(c, model.EndFailed, "The call could not be received")
			c.mu.Unlock()
			return
		}
		s.armLocked(c, s.ph.cachedTimes().ring, func() {
			s.failLocked(c, model.EndMissed, "", &tg.PhoneCallDiscardReasonMissed{})
		})
		s.emitLocked(c)
		c.mu.Unlock()
		s.replay(c, e)
	}()
}

// accepted: the callee answered; the caller computes the key and confirms.
func (s *session) accepted(c *tgCall, p *tg.PhoneCallAccepted) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.call.Outgoing || c.accepted || !c.call.Live() {
		return
	}
	c.accepted = true
	if p.ParticipantID != c.user {
		s.failLocked(c, model.EndFailed, "The call went to someone else", &tg.PhoneCallDiscardReasonDisconnect{})
		return
	}
	key, err := dhShared(c.dh, p.GB, c.secret)
	if err != nil {
		s.failLocked(c, model.EndFailed, "The key exchange failed", &tg.PhoneCallDiscardReasonDisconnect{})
		return
	}
	c.key, c.fp = key, keyFingerprint(key)
	s.armLocked(c, s.ph.cachedTimes().connect, func() {
		s.failLocked(c, model.EndFailed, "The call could not connect", &tg.PhoneCallDiscardReasonDisconnect{})
	})
	req := &tg.PhoneConfirmCallRequest{Peer: c.input(), GA: c.ga, KeyFingerprint: c.fp, Protocol: callProtocol()}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		api, _, err := s.phoneLive()
		var res *tg.PhonePhoneCall
		if err == nil {
			res, err = api.PhoneConfirmCall(ctx, req)
		}
		if err != nil {
			c.mu.Lock()
			s.failLocked(c, model.EndFailed, "The call could not be confirmed", &tg.PhoneCallDiscardReasonDisconnect{})
			c.mu.Unlock()
			return
		}
		s.learn(res.Users, nil)
		s.applyPhoneCall(c, res.PhoneCall)
	}()
}

// confirmed: both sides have the key; the signalling starts.
func (s *session) confirmed(c *tgCall, p *tg.PhoneCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sig != nil || !c.call.Live() {
		return
	}
	disconnect := &tg.PhoneCallDiscardReasonDisconnect{}
	if !c.call.Outgoing {
		if !c.accepted {
			s.finishLocked(c, model.EndHangup, "") // answered on another device
			return
		}
		if c.secret == nil {
			return
		}
		if h := sha256.Sum256(p.GAOrB); !bytes.Equal(h[:], c.gaHash) {
			s.failLocked(c, model.EndFailed, "The key exchange failed", disconnect)
			return
		}
		key, err := dhShared(c.dh, p.GAOrB, c.secret)
		if err != nil {
			s.failLocked(c, model.EndFailed, "The key exchange failed", disconnect)
			return
		}
		c.ga, c.key, c.fp = bytes.Clone(p.GAOrB), key, keyFingerprint(key)
	}
	if c.key == nil {
		return
	}
	if p.KeyFingerprint != c.fp {
		s.failLocked(c, model.EndFailed, "The encryption keys do not match", disconnect)
		return
	}
	if v := p.Protocol.LibraryVersions; len(v) == 0 || v[0] != callVersion {
		got := "none"
		if len(v) > 0 {
			got = v[0]
		}
		s.failLocked(c, model.EndFailed, "The other app speaks call version "+got+"; MegaChat speaks only "+callVersion, disconnect)
		return
	}
	if customFlag(p.CustomParameters.Data, "network_use_mtproto") {
		// The peer would carry media over MTProto framing instead of
		// DTLS, which a stock peer connection cannot speak.
		s.failLocked(c, model.EndFailed, "Telegram asked for a call transport MegaChat does not support", disconnect)
		return
	}
	c.call.ICE = iceServers(p.Connections)
	c.call.RelayOnly = !p.P2PAllowed
	s.env.Logf("info", "call %d: connecting (outgoing %v, video %v): %s; p2p %v; peer library %v; custom %s",
		c.id, c.call.Outgoing, c.call.Video, connectionsSummary(p.Connections), p.P2PAllowed, p.Protocol.LibraryVersions, clip(p.CustomParameters.Data, 300))
	c.call.Verification = verificationEmoji(c.key, c.ga)
	c.call.State = model.CallConnecting
	s.armLocked(c, s.ph.cachedTimes().connect, func() {
		s.failLocked(c, model.EndFailed, "The call could not connect", disconnect)
	})
	input := c.input()
	send := func(ctx context.Context, pkt []byte) error {
		api, _, err := s.phoneLive()
		if err != nil {
			return err
		}
		_, err = api.PhoneSendSignalingData(ctx, &tg.PhoneSendSignalingDataRequest{Peer: input, Data: pkt})
		return err
	}
	c.sig = newSignaling(c.key, c.call.Outgoing, send, func(m sigMessage) { s.fromPeer(c, m) }, s.env.Logf)
	// The app gets the call with its servers before any signal for it.
	s.emitLocked(c)
	mode := sigRespond
	if c.call.Outgoing {
		mode = sigInitiate
	}
	c.sig.start(mode)
	for _, pkt := range c.early {
		c.sig.receive(pkt)
	}
	for _, sig := range c.queued {
		s.toPeerLocked(c, sig)
	}
	c.early, c.queued = nil, nil
}

// connectionsSummary counts the call's servers by kind (no addresses or
// credentials).
func connectionsSummary(conns []tg.PhoneConnectionClass) string {
	var refl, reflTCP, stun, turn, turnNoCreds int
	for _, cc := range conns {
		switch c := cc.(type) {
		case *tg.PhoneConnection:
			if c.TCP {
				reflTCP++
			} else {
				refl++
			}
		case *tg.PhoneConnectionWebrtc:
			if c.Stun {
				stun++
			}
			if c.Turn {
				if c.Username != "" && c.Password != "" {
					turn++
				} else {
					turnNoCreds++
				}
			}
		}
	}
	return fmt.Sprintf("servers: %d reflectors (+%d tcp), %d stun, %d turn (+%d without credentials)", refl, reflTCP, stun, turn, turnNoCreds)
}

func stateName(st model.CallState) string {
	switch st {
	case model.CallWaiting:
		return "waiting"
	case model.CallRinging:
		return "ringing"
	case model.CallConnecting:
		return "connecting"
	case model.CallActive:
		return "active"
	}
	return "state " + strconv.Itoa(int(st))
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func customFlag(js, name string) bool {
	var m map[string]any
	if json.Unmarshal([]byte(js), &m) != nil {
		return false
	}
	v, _ := m[name].(bool)
	return v
}

func (s *session) discarded(c *tgCall, p *tg.PhoneCallDiscarded) {
	c.mu.Lock()
	defer c.mu.Unlock()
	answered := c.accepted || c.key != nil
	end, why := model.EndHangup, ""
	switch p.Reason.(type) {
	case *tg.PhoneCallDiscardReasonBusy:
		end = model.EndBusy
	case *tg.PhoneCallDiscardReasonMissed:
		end = model.EndMissed
	case *tg.PhoneCallDiscardReasonDisconnect:
		end, why = model.EndFailed, "The connection was lost"
	case *tg.PhoneCallDiscardReasonHangup:
		if !answered {
			end = model.EndMissed // the caller gave up
			if c.call.Outgoing {
				end = model.EndDeclined
			}
		}
	default: // none (answered on another device) or a group call migration
		if !answered && c.call.Outgoing {
			end = model.EndDeclined
		}
	}
	if c.call.Started.IsZero() && p.Duration > 0 {
		c.call.Started = time.Now().Add(-time.Duration(p.Duration) * time.Second)
	}
	s.finishLocked(c, end, why)
}

// fromPeer: a signalling message from the other side for the app.
func (s *session) fromPeer(c *tgCall, m sigMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.call.Live() {
		return
	}
	if m.Type == "candidate" {
		c.diag.note(false, "candidate "+candidateKind(m.SDP))
	} else {
		c.diag.note(false, m.Type)
	}
	switch m.Type {
	case "offer":
		s.emitSignalLocked(c, model.CallSignal{Kind: model.SigOffer, SDP: m.SDP})
	case "answer":
		s.emitSignalLocked(c, model.CallSignal{Kind: model.SigAnswer, SDP: m.SDP})
	case "candidate":
		if m.SDP == "" || reflectorCandidate(m.SDP) {
			return
		}
		s.emitSignalLocked(c, model.CallSignal{Kind: model.SigCandidate, SDP: m.SDP, Mid: m.Mid, MLine: m.MLine})
	}
}

func (s *session) toPeerLocked(c *tgCall, sig model.CallSignal) {
	var err error
	switch sig.Kind {
	case model.SigOffer:
		err = c.sig.sendJSON(sdpMessage{Type: "offer", SDP: sig.SDP})
	case model.SigAnswer:
		err = c.sig.sendJSON(sdpMessage{Type: "answer", SDP: sig.SDP})
	case model.SigCandidate:
		if sig.SDP == "" {
			return // end of candidates: tgcalls sends none
		}
		err = c.sig.sendJSON(candidateMessage{Type: "candidate", Mid: sig.Mid, MLine: sig.MLine, SDP: strings.TrimPrefix(sig.SDP, "a=")})
	}
	if err != nil {
		s.env.Logf("warn", "call %d: signal: %v", c.id, err)
	}
}

// --- Caller --------------------------------------------------------------

func (s *session) lookupCall(id string) (*tgCall, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errNoCall
	}
	c := s.ph.get(n)
	if c == nil {
		return nil, errNoCall
	}
	return c, nil
}

func requestErr(err error, peer string) error {
	switch {
	case tgerr.Is(err, "USER_PRIVACY_RESTRICTED"):
		return fmt.Errorf("telegram: %s does not accept calls", peer)
	case tgerr.Is(err, "PARTICIPANT_VERSION_OUTDATED", "CALL_PROTOCOL_LAYER_INVALID"):
		return fmt.Errorf("telegram: %s's app cannot take this call", peer)
	case tgerr.Is(err, "USER_IS_BLOCKED"):
		return fmt.Errorf("telegram: %s cannot be called", peer)
	}
	return authErr(err)
}

func (s *session) StartCall(ctx context.Context, chatID string, video bool) (model.Call, error) {
	kind, uid, ok := parseKey(chatID)
	if !ok || kind != 'u' {
		return model.Call{}, provider.ErrUnsupported
	}
	if uid == s.selfID {
		return model.Call{}, errors.New("telegram: you cannot call yourself")
	}
	hash, ok := s.peers.get(chatID)
	if !ok {
		return model.Call{}, fmt.Errorf("telegram: unknown user %s (refresh the chat list)", chatID)
	}
	api, _, err := s.phoneLive()
	if err != nil {
		return model.Call{}, err
	}
	times := s.callTimes(ctx, api)
	dh, random, err := s.dhConfig(ctx, api)
	if err != nil {
		return model.Call{}, err
	}
	a := dhSecret(random)
	ga, err := dhPublic(dh, a)
	if err != nil {
		return model.Call{}, err
	}
	gaHash := sha256.Sum256(ga)
	res, err := api.PhoneRequestCall(ctx, &tg.PhoneRequestCallRequest{
		Video: video, UserID: &tg.InputUser{UserID: uid, AccessHash: hash},
		RandomID: int(int32(randomID())), GAHash: gaHash[:], Protocol: callProtocol(),
	})
	peer := senderOf(s, chatID)
	if err != nil {
		return model.Call{}, requestErr(err, peer.Name)
	}
	s.learn(res.Users, nil)
	w, ok := res.PhoneCall.(*tg.PhoneCallWaiting)
	if !ok {
		return model.Call{}, errors.New("telegram: the call was not placed")
	}
	c := newCall(w.ID, w.AccessHash, uid, chatID, senderOf(s, chatID), true, video, model.CallWaiting)
	c.dh, c.secret, c.ga = dh, a, ga
	e := s.ph.add(c)
	c.mu.Lock()
	s.armLocked(c, times.receive, func() {
		s.failLocked(c, model.EndMissed, "", &tg.PhoneCallDiscardReasonMissed{})
	})
	s.emitLocked(c)
	c.mu.Unlock()
	s.applyPhoneCall(c, w)
	s.replay(c, e)
	c.mu.Lock()
	out := c.call
	c.mu.Unlock()
	return out, nil
}

func (s *session) AcceptCall(ctx context.Context, callID string, video bool) error {
	c, err := s.lookupCall(callID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	if c.call.Outgoing || c.call.State != model.CallRinging || c.accepted {
		c.mu.Unlock()
		return errors.New("telegram: the call cannot be answered now")
	}
	c.accepted = true
	c.mu.Unlock()
	fail := func(err error) error {
		c.mu.Lock()
		s.failLocked(c, model.EndFailed, "The call could not be answered", &tg.PhoneCallDiscardReasonDisconnect{})
		c.mu.Unlock()
		return err
	}
	api, _, err := s.phoneLive()
	if err != nil {
		return fail(err)
	}
	dh, random, err := s.dhConfig(ctx, api)
	if err != nil {
		return fail(err)
	}
	b := dhSecret(random)
	gb, err := dhPublic(dh, b)
	if err != nil {
		return fail(err)
	}
	c.mu.Lock()
	if !c.call.Live() {
		c.mu.Unlock()
		return errNoCall
	}
	c.dh, c.secret = dh, b
	// Now the caller confirms; phoneCall comes as an update.
	s.armLocked(c, s.ph.cachedTimes().connect, func() {
		s.failLocked(c, model.EndFailed, "The call could not connect", &tg.PhoneCallDiscardReasonDisconnect{})
	})
	c.mu.Unlock()
	res, err := api.PhoneAcceptCall(ctx, &tg.PhoneAcceptCallRequest{Peer: c.input(), GB: gb, Protocol: callProtocol()})
	if err != nil {
		if tgerr.Is(err, "CALL_ALREADY_ACCEPTED", "CALL_ALREADY_DECLINED") {
			c.mu.Lock()
			s.finishLocked(c, model.EndHangup, "")
			c.mu.Unlock()
			return nil
		}
		return fail(authErr(err))
	}
	s.learn(res.Users, nil)
	s.applyPhoneCall(c, res.PhoneCall)
	return nil
}

func (s *session) EndCall(ctx context.Context, callID string, reason model.CallEnd) error {
	c, err := s.lookupCall(callID)
	if err != nil {
		return nil // over already
	}
	c.mu.Lock()
	var (
		why tg.PhoneCallDiscardReasonClass = &tg.PhoneCallDiscardReasonHangup{}
		end                                = model.EndHangup
	)
	switch {
	case !c.call.Outgoing && !c.accepted:
		// Declining is "busy" for the caller, as in Telegram's own apps.
		why, end = &tg.PhoneCallDiscardReasonBusy{}, model.EndDeclined
		if reason == model.EndBusy {
			end = model.EndBusy
		}
	case c.call.Outgoing && !c.accepted:
		// Cancelled before the answer: the peer sees a missed call.
		why, end = &tg.PhoneCallDiscardReasonMissed{}, model.EndDeclined
	}
	req := s.discardRequestLocked(c, why)
	ok := s.finishLocked(c, end, "")
	c.mu.Unlock()
	if !ok {
		return nil
	}
	return s.discard(ctx, req)
}

func (s *session) CallSignal(ctx context.Context, callID string, sig model.CallSignal) error {
	c, err := s.lookupCall(callID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.call.Live() {
		return errNoCall
	}
	switch sig.Kind {
	case model.SigOffer:
		c.diag.note(true, "offer")
	case model.SigAnswer:
		c.diag.note(true, "answer")
	case model.SigCandidate:
		if sig.SDP != "" {
			c.diag.note(true, "candidate "+candidateKind(sig.SDP))
		}
	}
	switch sig.Kind {
	case model.SigOffer, model.SigAnswer, model.SigCandidate:
		if c.sig == nil {
			if len(c.queued) < maxQueuedSigs {
				c.queued = append(c.queued, sig)
			}
			return nil
		}
		s.toPeerLocked(c, sig)
	case model.SigData:
		if m, ok := parseMediaState(sig.Data); ok && m != c.call.Remote {
			c.call.Remote = m
			s.emitLocked(c)
		}
	case model.SigConnected:
		if c.call.State == model.CallConnecting {
			s.env.Logf("info", "call %d: connected: %s", c.id, c.diagLocked())
			c.call.State, c.call.Started = model.CallActive, time.Now()
			s.stopLocked(c)
			s.emitLocked(c)
		}
	case model.SigFailed:
		s.env.Logf("info", "call %d: the app reports: %s", c.id, clip(sig.Data, 300))
		s.failLocked(c, model.EndFailed, "The call could not connect", &tg.PhoneCallDiscardReasonDisconnect{})
	}
	return nil
}

func (s *session) SetCallMedia(ctx context.Context, callID string, m model.CallMedia) error {
	c, err := s.lookupCall(callID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.call.Live() {
		return errNoCall
	}
	// The app sends it on the data channel (tgcalls sends its MediaState
	// there when the channel opens and on every change).
	s.emitSignalLocked(c, model.CallSignal{Kind: model.SigData, Data: mediaStateJSON(m)})
	return nil
}

var _ provider.Caller = (*session)(nil)
