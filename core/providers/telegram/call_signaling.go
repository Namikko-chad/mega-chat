package telegram

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/tg"
	"github.com/pion/logging"
	"github.com/pion/sctp"

	"megachat/core/model"
)

// Signalling of library version 11.0.0 (tgcalls InstanceV2ReferenceImpl,
// signaling protocol V2): every phone.sendSignalingData blob is one raw
// SCTP packet (ports 5000/5000, no DTLS, no UDP); each SCTP message on
// stream 0 (PPID 53, no DCEP) is callCipher.seal(gzip(JSON)). The caller
// starts the association.

const (
	sigMaxMessage = 262144 // DcSctpTransport::Start(5000, 5000, 262144)
	sigMaxPlain   = 2 << 20
	sigMTU        = 1191 // dcsctp's default MTU; the packets are small anyway
	sigRTOMax     = 3000 // ms: a packet sent before the peer's engine exists is lost
)

// packetConn is the packet net.Conn the association runs over: a Write is
// one packet for Telegram, a Read one packet that came from it.
type packetConn struct {
	in    chan []byte
	out   func([]byte) // must not block
	done  chan struct{}
	first chan struct{} // closed when the first packet came
	once  sync.Once
	fonce sync.Once
}

func newPacketConn(out func([]byte)) *packetConn {
	return &packetConn{in: make(chan []byte, 64), out: out, done: make(chan struct{}), first: make(chan struct{})}
}

// deliver hands a packet from the peer to the association; when it falls
// behind, the packet is dropped and SCTP sends it again.
func (c *packetConn) deliver(p []byte) {
	c.fonce.Do(func() { close(c.first) })
	select {
	case c.in <- p:
	case <-c.done:
	default:
	}
}

func (c *packetConn) Read(b []byte) (int, error) {
	select {
	case p := <-c.in:
		return copy(b, p), nil
	case <-c.done:
		return 0, io.EOF
	}
}

func (c *packetConn) Write(b []byte) (int, error) {
	select {
	case <-c.done:
		return 0, net.ErrClosed
	default:
	}
	c.out(bytes.Clone(b))
	return len(b), nil
}

func (c *packetConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

type sigAddr struct{}

func (sigAddr) Network() string { return "telegram" }
func (sigAddr) String() string  { return "signaling" }

func (c *packetConn) LocalAddr() net.Addr              { return sigAddr{} }
func (c *packetConn) RemoteAddr() net.Addr             { return sigAddr{} }
func (c *packetConn) SetDeadline(time.Time) error      { return nil }
func (c *packetConn) SetReadDeadline(time.Time) error  { return nil }
func (c *packetConn) SetWriteDeadline(time.Time) error { return nil }

// sigMode is how this side starts the association.
type sigMode int

const (
	sigRespond  sigMode = iota // wait for the peer's INIT (callee)
	sigInitiate                // send INIT at once (caller)
	// sigInitiateLate sends INIT when the first packet came, as a tgcalls
	// callee does (its transport becomes writable then): both INITs cross.
	sigInitiateLate
)

// sigMessage is a message of the signalling channel.
type sigMessage struct {
	Type  string `json:"@type"`
	SDP   string `json:"sdp"`
	Mid   string `json:"mid"`
	MLine int    `json:"mline"`
}

// signaling is one call's encrypted signalling channel.
type signaling struct {
	ciph  *callCipher
	conn  *packetConn
	send  func(context.Context, []byte) error // a packet to the peer
	onMsg func(sigMessage)
	logf  func(level, format string, args ...any)

	ctx    context.Context
	cancel context.CancelFunc
	outq   chan []byte

	// What went through, for the log when a call does not connect.
	pktIn, pktOut, sendErrs, dropped atomic.Int32
	up                               atomic.Bool // the association is established

	mu      sync.Mutex
	assoc   *sctp.Association
	stream  *sctp.Stream
	pending [][]byte // sealed before the stream was open
}

func newSignaling(key []byte, outgoing bool, send func(context.Context, []byte) error, onMsg func(sigMessage), logf func(string, string, ...any)) *signaling {
	g := &signaling{ciph: newCallCipher(key, outgoing), send: send, onMsg: onMsg, logf: logf, outq: make(chan []byte, 256)}
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.conn = newPacketConn(func(p []byte) {
		select {
		case g.outq <- p:
		default: // SCTP sends it again
			g.dropped.Add(1)
		}
	})
	return g
}

func (g *signaling) start(mode sigMode) {
	go g.sendLoop()
	go g.run(mode)
}

// sendLoop sends the packets in order, one request at a time.
func (g *signaling) sendLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case p := <-g.outq:
			ctx, cancel := context.WithTimeout(g.ctx, 15*time.Second)
			if err := g.send(ctx, p); err != nil {
				if g.ctx.Err() == nil {
					g.sendErrs.Add(1)
					g.logf("warn", "call signalling: send: %v", err)
				}
			} else {
				g.pktOut.Add(1)
			}
			cancel()
		}
	}
}

func (g *signaling) run(mode sigMode) {
	opts := []sctp.AssociationOption{
		sctp.WithNetConn(g.conn), sctp.WithName("call"), sctp.WithLoggerFactory(sctpLog{g.logf}),
		sctp.WithMTU(sigMTU), sctp.WithMaxMessageSize(sigMaxMessage), sctp.WithRTOMax(sigRTOMax),
		// As WebRTC's dcsctp: plain DATA chunks, CRC32c checksums.
		sctp.WithEnableInterleaving(false), sctp.WithEnableZeroChecksum(false),
	}
	var (
		a   *sctp.Association
		err error
	)
	switch mode {
	case sigRespond:
		so := make([]sctp.ServerOption, len(opts))
		for i, o := range opts {
			so[i] = o
		}
		a, err = sctp.ServerWithOptions(so...)
	default:
		if mode == sigInitiateLate {
			select {
			case <-g.conn.first:
			case <-g.ctx.Done():
				return
			}
		}
		co := make([]sctp.ClientOption, len(opts))
		for i, o := range opts {
			co[i] = o
		}
		a, err = sctp.ClientContext(g.ctx, co...)
	}
	if err != nil {
		if g.ctx.Err() == nil {
			g.logf("warn", "call signalling: association: %v", err)
		}
		return
	}
	st, err := a.OpenStream(0, sctp.PayloadTypeWebRTCBinary)
	if err != nil {
		_ = a.Close()
		return
	}
	g.mu.Lock()
	if g.ctx.Err() != nil {
		g.mu.Unlock()
		_ = a.Close()
		return
	}
	g.assoc, g.stream = a, st
	g.up.Store(true)
	pending := g.pending
	g.pending = nil
	g.mu.Unlock()
	for _, m := range pending {
		g.write(st, m)
	}
	buf := make([]byte, sigMaxMessage+1024)
	for {
		n, _, err := st.ReadSCTP(buf)
		if err != nil {
			if !errors.Is(err, io.ErrShortBuffer) {
				return
			}
			continue
		}
		g.handle(buf[:n])
	}
}

func (g *signaling) write(st *sctp.Stream, sealed []byte) {
	if _, err := st.WriteSCTP(sealed, sctp.PayloadTypeWebRTCBinary); err != nil && g.ctx.Err() == nil {
		g.logf("warn", "call signalling: write: %v", err)
	}
}

func (g *signaling) handle(sealed []byte) {
	plain, err := g.ciph.open(sealed)
	if err != nil {
		return // forged, repeated or too old
	}
	if len(plain) >= 2 && plain[0] == 0x1f && plain[1] == 0x8b {
		if plain, err = gunzip(plain, sigMaxPlain); err != nil {
			g.logf("warn", "call signalling: bad gzip")
			return
		}
	}
	var m sigMessage
	if err := json.Unmarshal(plain, &m); err != nil || m.Type == "" {
		g.logf("warn", "call signalling: not a message")
		return
	}
	g.onMsg(m)
}

// receive takes a packet that came in updatePhoneCallSignalingData.
func (g *signaling) receive(p []byte) {
	g.pktIn.Add(1)
	g.conn.deliver(p)
}

// stats says what went through the channel (no payload).
func (g *signaling) stats() string {
	st := "not established"
	if g.up.Load() {
		st = "established"
	}
	return fmt.Sprintf("sctp %s, packets %d sent / %d received, %d send errors, %d dropped",
		st, g.pktOut.Load(), g.pktIn.Load(), g.sendErrs.Load(), g.dropped.Load())
}

// sendJSON sends a message; before the association is up it is queued.
func (g *signaling) sendJSON(v any) error {
	plain, err := marshalJSON(v)
	if err != nil {
		return err
	}
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	_, _ = zw.Write(plain)
	if err := zw.Close(); err != nil {
		return err
	}
	sealed := g.ciph.seal(z.Bytes())
	if len(sealed) > sigMaxMessage {
		return errors.New("telegram: call signalling message too large")
	}
	g.mu.Lock()
	st := g.stream
	if st == nil {
		g.pending = append(g.pending, sealed)
	}
	g.mu.Unlock()
	if st != nil {
		g.write(st, sealed)
	}
	return nil
}

func (g *signaling) close() {
	g.cancel()
	_ = g.conn.Close()
	g.mu.Lock()
	a := g.assoc
	g.assoc, g.stream = nil, nil
	g.mu.Unlock()
	if a != nil {
		go func() { _ = a.Close() }()
	}
}

func marshalJSON(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

func gunzip(b []byte, limit int64) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(io.LimitReader(zr, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(out)) > limit {
		return nil, errors.New("too large")
	}
	return out, nil
}

// --- messages ------------------------------------------------------------

// The fields of the messages are in json11's (sorted) order.

type sdpMessage struct {
	Type string `json:"@type"` // offer, answer
	SDP  string `json:"sdp"`
}

type candidateMessage struct {
	Type  string `json:"@type"` // candidate
	Mid   string `json:"mid"`
	MLine int    `json:"mline"`
	SDP   string `json:"sdp"` // the "candidate:" line
}

// mediaState is what each side sends on the "data" channel when it opens
// and whenever it changes (tgcalls MediaStateMessage).
type mediaState struct {
	Type            string `json:"@type"`
	LowBattery      bool   `json:"lowBattery"`
	Muted           bool   `json:"muted"`
	ScreencastState string `json:"screencastState"`
	VideoRotation   int    `json:"videoRotation"`
	VideoState      string `json:"videoState"`
}

func videoState(on bool) string {
	if on {
		return "active"
	}
	return "inactive"
}

func mediaStateJSON(m model.CallMedia) string {
	b, _ := marshalJSON(mediaState{
		Type: "MediaState", Muted: m.Muted,
		VideoState: videoState(m.Video), ScreencastState: videoState(m.Screen),
	})
	return string(b)
}

// parseMediaState reads a data channel message; ok is false for anything
// but a MediaState.
func parseMediaState(s string) (model.CallMedia, bool) {
	var m mediaState
	if err := json.Unmarshal([]byte(s), &m); err != nil || m.Type != "MediaState" {
		return model.CallMedia{}, false
	}
	return model.CallMedia{Muted: m.Muted, Video: m.VideoState == "active", Screen: m.ScreencastState == "active"}, true
}

// --- ICE -----------------------------------------------------------------

func uriHost(h string) string {
	if strings.Contains(h, ":") {
		return "[" + h + "]"
	}
	return h
}

// iceServers maps the call's servers the way tgcalls hands them to its
// peer connection. Reflectors (phoneConnection) speak Telegram's own
// protocol, which a stock WebRTC stack cannot, so they are left out.
func iceServers(conns []tg.PhoneConnectionClass) []model.IceServer {
	var out []model.IceServer
	for _, cc := range conns {
		c, ok := cc.(*tg.PhoneConnectionWebrtc)
		if !ok {
			continue
		}
		var hosts []string
		for _, h := range []string{c.IP, c.Ipv6} {
			if h != "" {
				hosts = append(hosts, uriHost(h)+":"+strconv.Itoa(c.Port))
			}
		}
		if len(hosts) == 0 {
			continue
		}
		if c.Stun {
			s := model.IceServer{}
			for _, h := range hosts {
				s.URLs = append(s.URLs, "stun:"+h)
			}
			out = append(out, s)
		}
		if c.Turn && c.Username != "" && c.Password != "" {
			s := model.IceServer{Username: c.Username, Credential: c.Password}
			for _, h := range hosts {
				s.URLs = append(s.URLs, "turn:"+h)
			}
			out = append(out, s)
		}
	}
	return out
}

// reflectorCandidate reports a candidate on a Telegram reflector
// ("reflector-<id>-<tag>.reflector"): the app could not resolve or use it.
func reflectorCandidate(sdp string) bool {
	f := strings.Fields(strings.TrimPrefix(sdp, "a="))
	return len(f) >= 5 && strings.HasSuffix(strings.ToLower(f[4]), ".reflector")
}

// candidateKind is a candidate's protocol and type ("udp host", "udp
// relay", "reflector") for the log; it carries no address.
func candidateKind(sdp string) string {
	if reflectorCandidate(sdp) {
		return "reflector"
	}
	f := strings.Fields(strings.TrimPrefix(sdp, "a="))
	typ := "?"
	for i := 0; i+1 < len(f); i++ {
		if f[i] == "typ" {
			typ = f[i+1]
			break
		}
	}
	proto := "?"
	if len(f) >= 3 {
		proto = strings.ToLower(f[2])
	}
	return proto + " " + typ
}

// --- pion logs -----------------------------------------------------------

// sctpLog passes pion's warnings and errors on (they carry no payload).
type sctpLog struct {
	logf func(level, format string, args ...any)
}

func (l sctpLog) NewLogger(string) logging.LeveledLogger { return l }

func (sctpLog) Trace(string)          {}
func (sctpLog) Tracef(string, ...any) {}
func (sctpLog) Debug(string)          {}
func (sctpLog) Debugf(string, ...any) {}
func (sctpLog) Info(string)           {}
func (sctpLog) Infof(string, ...any)  {}
func (l sctpLog) Warn(msg string)     { l.logf("warn", "sctp: %s", msg) }
func (l sctpLog) Warnf(f string, a ...any) {
	l.logf("warn", "sctp: "+f, a...)
}
func (l sctpLog) Error(msg string) { l.logf("warn", "sctp: %s", msg) }
func (l sctpLog) Errorf(f string, a ...any) {
	l.logf("warn", "sctp: "+f, a...)
}
