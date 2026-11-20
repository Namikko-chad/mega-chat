package hub

import (
	"context"
	"errors"
	"time"

	"megachat/core/eventlog"
	"megachat/core/model"
	"megachat/core/provider"
)

// ErrInCall: the device is in another call already.
var ErrInCall = errors.New("already in a call")

// maxSignals bounds the signals kept per call for an app that reconnects.
const maxSignals = 512

type callKey struct{ acct, id string }

type liveCall struct {
	call model.Call // zero ID: only signals came so far
	sigs []model.CallSignal
	seq  int64
}

// caller returns the call capability of an account's session.
func (h *Hub) caller(acct string) (provider.Caller, error) {
	h.mu.Lock()
	r := h.runs[acct]
	h.mu.Unlock()
	if r == nil {
		return nil, ErrNoAccount
	}
	if r.sess == nil || r.acct.Disabled {
		return nil, provider.ErrOffline
	}
	c, ok := r.sess.(provider.Caller)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	return c, nil
}

// busy reports whether a call other than k is live (one call per device,
// over all accounts).
func (h *Hub) busy(k callKey) bool {
	for ok, lc := range h.calls {
		if ok != k && lc.call.ID != "" && lc.call.Live() {
			return true
		}
	}
	return false
}

func (h *Hub) applyCall(acct string, c model.Call) {
	if c.ID == "" {
		return
	}
	k := callKey{acct, c.ID}
	h.callMu.Lock()
	lc := h.calls[k]
	if (lc == nil || lc.call.ID == "") && !c.Outgoing && c.Live() && h.busy(k) {
		delete(h.calls, k)
		h.callMu.Unlock()
		h.logf(eventlog.Info, "call", acct, "declined an incoming call: busy")
		go h.decline(acct, c.ID)
		return
	}
	if lc == nil {
		lc = &liveCall{}
		h.calls[k] = lc
	}
	if !c.Live() && lc.call.ID == "" && c.State == model.CallEnded {
		// Ended before we heard of it (a missed call already over).
		delete(h.calls, k)
		h.callMu.Unlock()
		return
	}
	lc.call = c
	if !c.Live() {
		delete(h.calls, k)
	}
	h.callMu.Unlock()
	h.emit(Event{Kind: EvCall, AccountID: acct, Call: &c})
}

func (h *Hub) decline(acct, id string) {
	cl, err := h.caller(acct)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(h.ctx, 30*time.Second)
	defer cancel()
	if err := cl.EndCall(ctx, id, model.EndBusy); err != nil {
		h.logf(eventlog.Warn, "call", acct, "decline: %v", err)
	}
}

func (h *Hub) applySignal(acct, id string, sig model.CallSignal) {
	k := callKey{acct, id}
	h.callMu.Lock()
	lc := h.calls[k]
	if lc == nil {
		lc = &liveCall{}
		h.calls[k] = lc
	}
	lc.seq++
	sig.Seq = lc.seq
	lc.sigs = append(lc.sigs, sig)
	if len(lc.sigs) > maxSignals {
		lc.sigs = append([]model.CallSignal(nil), lc.sigs[len(lc.sigs)-maxSignals:]...)
	}
	h.callMu.Unlock()
	h.emit(Event{Kind: EvCallSignal, AccountID: acct, Call: &model.Call{ID: id}, Signal: &sig})
}

// dropCalls ends the calls of an account whose session stopped.
func (h *Hub) dropCalls(acct string) {
	var ended []model.Call
	h.callMu.Lock()
	for k, lc := range h.calls {
		if k.acct != acct {
			continue
		}
		delete(h.calls, k)
		if lc.call.ID != "" && lc.call.Live() {
			c := lc.call
			c.State, c.End = model.CallEnded, model.EndFailed
			ended = append(ended, c)
		}
	}
	h.callMu.Unlock()
	for _, c := range ended {
		h.emit(Event{Kind: EvCall, AccountID: acct, Call: &c})
	}
}

// StartCall calls the peer of a direct chat.
func (h *Hub) StartCall(ctx context.Context, acct, chat string, video bool) (model.Call, error) {
	cl, err := h.caller(acct)
	if err != nil {
		return model.Call{}, err
	}
	h.callMu.Lock()
	busy := h.busy(callKey{})
	h.callMu.Unlock()
	if busy {
		return model.Call{}, ErrInCall
	}
	c, err := cl.StartCall(ctx, chat, video)
	if err != nil {
		return model.Call{}, err
	}
	// The session reports it too; whichever comes later wins, both are
	// the same call.
	h.callMu.Lock()
	k := callKey{acct, c.ID}
	if lc := h.calls[k]; lc == nil {
		h.calls[k] = &liveCall{call: c}
	} else if lc.call.ID == "" {
		lc.call = c
	} else {
		c = lc.call
	}
	h.callMu.Unlock()
	return c, nil
}

// AcceptCall answers an incoming call.
func (h *Hub) AcceptCall(ctx context.Context, acct, id string, video bool) error {
	cl, err := h.caller(acct)
	if err != nil {
		return err
	}
	return cl.AcceptCall(ctx, id, video)
}

// EndCall hangs up, cancels or declines.
func (h *Hub) EndCall(ctx context.Context, acct, id string, reason model.CallEnd) error {
	cl, err := h.caller(acct)
	if err != nil {
		return err
	}
	if reason != model.EndDeclined {
		reason = model.EndHangup
	}
	return cl.EndCall(ctx, id, reason)
}

// CallSignal passes what the app's peer connection produced.
func (h *Hub) CallSignal(ctx context.Context, acct, id string, sig model.CallSignal) error {
	cl, err := h.caller(acct)
	if err != nil {
		return err
	}
	return cl.CallSignal(ctx, id, sig)
}

// SetCallMedia tells the peer about this side's microphone, camera, screen.
func (h *Hub) SetCallMedia(ctx context.Context, acct, id string, m model.CallMedia) error {
	cl, err := h.caller(acct)
	if err != nil {
		return err
	}
	return cl.SetCallMedia(ctx, id, m)
}

// LiveCall is a call with its account.
type LiveCall struct {
	AccountID string
	Call      model.Call
}

// Calls returns the live calls.
func (h *Hub) Calls() []LiveCall {
	h.callMu.Lock()
	defer h.callMu.Unlock()
	var out []LiveCall
	for k, lc := range h.calls {
		if lc.call.ID != "" && lc.call.Live() {
			out = append(out, LiveCall{AccountID: k.acct, Call: lc.call})
		}
	}
	return out
}

// CallSignals returns the kept signals of a call after seq.
func (h *Hub) CallSignals(acct, id string, after int64) []model.CallSignal {
	h.callMu.Lock()
	defer h.callMu.Unlock()
	lc := h.calls[callKey{acct, id}]
	if lc == nil {
		return nil
	}
	var out []model.CallSignal
	for _, s := range lc.sigs {
		if s.Seq > after {
			out = append(out, s)
		}
	}
	return out
}
