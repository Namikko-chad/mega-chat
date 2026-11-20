package telegram

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"math/big"
	"slices"
	"sync"
)

// The key exchange and the signalling encryption of 1:1 calls, as in
// tdesktop (calls_call.cpp, calls_emoji_fingerprint.cpp) and tgcalls
// (CryptoHelper.cpp, EncryptedConnection.cpp, Type::Signaling).

const callKeySize = 256

var errCallCrypto = errors.New("telegram: bad call key exchange")

// dhParams are the server's Diffie-Hellman parameters (messages.getDhConfig).
type dhParams struct {
	g       int
	p       *big.Int
	version int
}

// dhSecret is a fresh a (or b): random bytes mixed with the server's, as
// the docs ask.
func dhSecret(serverRandom []byte) []byte {
	b := make([]byte, callKeySize)
	_, _ = rand.Read(b)
	for i := range min(len(b), len(serverRandom)) {
		b[i] ^= serverRandom[i]
	}
	return b
}

// pad256 writes x big-endian into 256 bytes, left-padded with zeros.
func pad256(x *big.Int) []byte {
	return x.FillBytes(make([]byte, callKeySize))
}

// dhPublic is g^secret mod p, checked as the peer will check it.
func dhPublic(dh dhParams, secret []byte) ([]byte, error) {
	gx := new(big.Int).Exp(big.NewInt(int64(dh.g)), new(big.Int).SetBytes(secret), dh.p)
	if err := dhCheck(dh, gx); err != nil {
		return nil, err
	}
	return pad256(gx), nil
}

// dhCheck: 1 < gx < p-1, and 2^(2048-64) < gx < p - 2^(2048-64).
func dhCheck(dh dhParams, gx *big.Int) error {
	one := big.NewInt(1)
	if gx.Cmp(one) <= 0 || gx.Cmp(new(big.Int).Sub(dh.p, one)) >= 0 {
		return errCallCrypto
	}
	lo := new(big.Int).Lsh(one, 2048-64)
	if gx.Cmp(lo) <= 0 || gx.Cmp(new(big.Int).Sub(dh.p, lo)) >= 0 {
		return errCallCrypto
	}
	return nil
}

// dhShared is the call key: other^secret mod p in 256 bytes.
func dhShared(dh dhParams, other, secret []byte) ([]byte, error) {
	o := new(big.Int).SetBytes(other)
	if err := dhCheck(dh, o); err != nil {
		return nil, err
	}
	return pad256(new(big.Int).Exp(o, new(big.Int).SetBytes(secret), dh.p)), nil
}

// keyFingerprint is the last 8 bytes of SHA1(key), little-endian
// (tdesktop ComputeFingerprint).
func keyFingerprint(key []byte) int64 {
	h := sha1.Sum(key)
	return int64(binary.LittleEndian.Uint64(h[12:20]))
}

// verificationEmoji are the 4 emoji both sides show: SHA256(key || g_a)
// split into 4 big-endian 63-bit numbers, each mod the table size.
func verificationEmoji(key, ga []byte) []string {
	h := sha256.New()
	h.Write(key)
	h.Write(ga)
	sum := h.Sum(nil)
	out := make([]string, 0, 4)
	for i := 0; i < len(sum); i += 8 {
		v := binary.BigEndian.Uint64(sum[i:i+8]) & 0x7FFFFFFFFFFFFFFF
		out = append(out, callEmoji[v%uint64(len(callEmoji))])
	}
	return out
}

// --- signalling encryption -----------------------------------------------

// keepCounters is how many recent incoming counters are remembered
// (kKeepIncomingCountersCount): older or repeated packets are dropped.
const keepCounters = 64

// maxSealedPacket is the largest packet decrypted (kMaxIncomingPacketSize).
const maxSealedPacket = 128 * 1024

// callCipher seals what one side sends and opens what the other sent:
// MTProto 2.0 key derivation with AES-256-CTR, no padding, a 32-bit
// counter in front of the payload.
type callCipher struct {
	key      []byte
	outgoing bool // the caller

	mu      sync.Mutex
	counter uint32
	seen    []uint32 // sorted, the largest last
}

func newCallCipher(key []byte, outgoing bool) *callCipher {
	return &callCipher{key: key, outgoing: outgoing}
}

// x is the key offset: 128 for signalling, +8 for what the callee sends.
func (c *callCipher) x(sending bool) int {
	if sending == c.outgoing {
		return 128
	}
	return 136
}

// callAESKeyIV is tgcalls PrepareAesKeyIv.
func callAESKeyIV(key, msgKey []byte, x int) (aesKey, iv []byte) {
	a := sha256.New()
	a.Write(msgKey)
	a.Write(key[x : x+36])
	sa := a.Sum(nil)
	b := sha256.New()
	b.Write(key[40+x : 40+x+36])
	b.Write(msgKey)
	sb := b.Sum(nil)
	aesKey = slices.Concat(sa[:8], sb[8:24], sa[24:32])
	iv = slices.Concat(sb[:4], sa[8:16], sb[24:28])
	return aesKey, iv
}

func callCTR(aesKey, iv, dst, src []byte) {
	blk, _ := aes.NewCipher(aesKey) // the key is always 32 bytes
	cipher.NewCTR(blk, iv).XORKeyStream(dst, src)
}

// msgKey is SHA256(key[88+x : 120+x] || plain)[8:24].
func callMsgKey(key, plain []byte, x int) []byte {
	h := sha256.New()
	h.Write(key[88+x : 88+x+32])
	h.Write(plain)
	return h.Sum(nil)[8:24]
}

// seal is encryptRawPacket: msg_key || AES-CTR(seq || payload).
func (c *callCipher) seal(payload []byte) []byte {
	c.mu.Lock()
	c.counter++
	seq := c.counter
	c.mu.Unlock()
	plain := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(plain, seq)
	copy(plain[4:], payload)
	x := c.x(true)
	mk := callMsgKey(c.key, plain, x)
	k, iv := callAESKeyIV(c.key, mk, x)
	out := make([]byte, 16+len(plain))
	copy(out, mk)
	callCTR(k, iv, out[16:], plain)
	return out
}

// open is decryptRawPacket; it fails for a forged, repeated or too old
// packet.
func (c *callCipher) open(pkt []byte) ([]byte, error) {
	if len(pkt) < 21 || len(pkt) > maxSealedPacket {
		return nil, errCallCrypto
	}
	x := c.x(false)
	mk := pkt[:16]
	k, iv := callAESKeyIV(c.key, mk, x)
	plain := make([]byte, len(pkt)-16)
	callCTR(k, iv, plain, pkt[16:])
	if subtle.ConstantTimeCompare(callMsgKey(c.key, plain, x), mk) != 1 {
		return nil, errCallCrypto
	}
	// Bits 31 and 30 are flags of the older (V1) framing.
	counter := binary.BigEndian.Uint32(plain) &^ (1<<31 | 1<<30)
	c.mu.Lock()
	ok := c.register(counter)
	c.mu.Unlock()
	if !ok {
		return nil, errCallCrypto
	}
	return plain[4:], nil
}

// register is registerIncomingCounter: false for a counter seen already or
// keepCounters behind the largest one.
func (c *callCipher) register(n uint32) bool {
	pos, found := slices.BinarySearch(c.seen, n)
	if found {
		return false
	}
	if len(c.seen) > 0 && uint64(n)+keepCounters <= uint64(c.seen[len(c.seen)-1]) {
		return false
	}
	drop := 0
	for drop < len(c.seen) && uint64(c.seen[drop])+keepCounters <= uint64(n) {
		drop++
	}
	c.seen = slices.Insert(c.seen[drop:], pos-drop, n)
	return true
}

// callEmoji is the emoji table of tdesktop calls_emoji_fingerprint.cpp
// (Data/Offsets, 333 entries, same order). Text-style symbols and keycaps
// carry U+FE0F so they draw as emoji.
var callEmoji = [...]string{
	"😉", "😍", "😛", "😭", "😱", "😡", "😎", "😴", "😵", "😈", "😬", "😇",
	"😏", "👮", "👷", "💂", "👶", "👨", "👩", "👴", "👵", "😻", "😽", "🙀",
	"👺", "🙈", "🙉", "🙊", "💀", "👽", "💩", "🔥", "💥", "💤", "👂", "👀",
	"👃", "👅", "👄", "👍", "👎", "👌", "👊", "✌️", "✋", "👐", "👆", "👇",
	"👉", "👈", "🙏", "👏", "💪", "🚶", "🏃", "💃", "👫", "👪", "👬", "👭",
	"💅", "🎩", "👑", "👒", "👟", "👞", "👠", "👕", "👗", "👖", "👙", "👜",
	"👓", "🎀", "💄", "💛", "💙", "💜", "💚", "💍", "💎", "🐶", "🐺", "🐱",
	"🐭", "🐹", "🐰", "🐸", "🐯", "🐨", "🐻", "🐷", "🐮", "🐗", "🐴", "🐑",
	"🐘", "🐼", "🐧", "🐥", "🐔", "🐍", "🐢", "🐛", "🐝", "🐜", "🐞", "🐌",
	"🐙", "🐚", "🐟", "🐬", "🐋", "🐐", "🐊", "🐫", "🍀", "🌹", "🌻", "🍁",
	"🌾", "🍄", "🌵", "🌴", "🌳", "🌞", "🌚", "🌙", "🌎", "🌋", "⚡", "☔",
	"❄️", "⛄", "🌀", "🌈", "🌊", "🎓", "🎆", "🎃", "👻", "🎅", "🎄", "🎁",
	"🎈", "🔮", "🎥", "📷", "💿", "💻", "☎️", "📡", "📺", "📻", "🔉", "🔔",
	"⏳", "⏰", "⌚", "🔒", "🔑", "🔎", "💡", "🔦", "🔌", "🔋", "🚿", "🚽",
	"🔧", "🔨", "🚪", "🚬", "💣", "🔫", "🔪", "💊", "💉", "💰", "💵", "💳",
	"✉️", "📫", "📦", "📅", "📁", "✂️", "📌", "📎", "✒️", "✏️", "📐", "📚",
	"🔬", "🔭", "🎨", "🎬", "🎤", "🎧", "🎵", "🎹", "🎻", "🎺", "🎸", "👾",
	"🎮", "🃏", "🎲", "🎯", "🏈", "🏀", "⚽", "⚾", "🎾", "🎱", "🏉", "🎳",
	"🏁", "🏇", "🏆", "🏊", "🏄", "☕", "🍼", "🍺", "🍷", "🍴", "🍕", "🍔",
	"🍟", "🍗", "🍱", "🍚", "🍜", "🍡", "🍳", "🍞", "🍩", "🍦", "🎂", "🍰",
	"🍪", "🍫", "🍭", "🍯", "🍎", "🍏", "🍊", "🍋", "🍒", "🍇", "🍉", "🍓",
	"🍑", "🍌", "🍐", "🍍", "🍆", "🍅", "🌽", "🏡", "🏥", "🏦", "⛪", "🏰",
	"⛺", "🏭", "🗻", "🗽", "🎠", "🎡", "⛲", "🎢", "🚢", "🚤", "⚓", "🚀",
	"✈️", "🚁", "🚂", "🚋", "🚎", "🚌", "🚙", "🚗", "🚕", "🚛", "🚨", "🚔",
	"🚒", "🚑", "🚲", "🚠", "🚜", "🚦", "⚠️", "🚧", "⛽", "🎰", "🗿", "🎪",
	"🎭", "🇯🇵", "🇰🇷", "🇩🇪", "🇨🇳", "🇺🇸", "🇫🇷", "🇪🇸", "🇮🇹", "🇷🇺", "🇬🇧", "1️⃣",
	"2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "0️⃣", "🔟", "❗", "❓",
	"♥️", "♦️", "💯", "🔗", "🔱", "🔴", "🔵", "🔶", "🔷",
}
