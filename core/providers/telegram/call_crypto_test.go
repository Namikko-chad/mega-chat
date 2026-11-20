package telegram

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"
)

// The vectors come from a C++ harness built from tgcalls' CryptoHelper.cpp
// with EncryptedConnection's encryptPrepared/decryptRawPacket/
// registerIncomingCounter and tdesktop's ComputeFingerprint/
// ComputeEmojiIndex copied in (OpenSSL AES-CTR, SHA1, SHA256, BN_mod_exp).

func testKey() []byte {
	k := make([]byte, callKeySize)
	for i := range k {
		k[i] = byte(i*7 + 3)
	}
	return k
}

func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestCallFingerprintEmoji(t *testing.T) {
	key := testKey()
	if fp := uint64(keyFingerprint(key)); fp != 0x31c96c19efe6d69e {
		t.Fatalf("fingerprint %x", fp)
	}
	ga := make([]byte, callKeySize)
	for i := range ga {
		ga[i] = byte(255 - i)
	}
	got := verificationEmoji(key, ga)
	want := []string{callEmoji[238], callEmoji[73], callEmoji[220], callEmoji[115]}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("emoji %v, want %v", got, want)
	}
	if len(callEmoji) != 333 || callEmoji[0] != "😉" || callEmoji[332] != "🔷" || callEmoji[43] != "✌️" || callEmoji[311] != "1️⃣" {
		t.Fatalf("emoji table: %d %q %q %q %q", len(callEmoji), callEmoji[0], callEmoji[332], callEmoji[43], callEmoji[311])
	}
}

func TestCallKDF(t *testing.T) {
	key := testKey()
	mk := make([]byte, 16)
	for i := range mk {
		mk[i] = byte(0xa0 + i)
	}
	for _, c := range []struct {
		x       int
		key, iv string
	}{
		{128, "213517c0f5c931193b601b4d32612658fadffa055ccde70d4d96ccedd645dda7", "2b9964cfdd5d36b83f7ef8202f059fd1"},
		{136, "1b91a4f791a1349067a51a578146c81c9bf3476539d0a86c890687484d90c84c", "a81e02c64c869090aaf5478d3c598a67"},
		{0, "95cd2967d58db1d69cfb17bc68dd23a02cd99d8bb7920f4bd39b75f8ff2a9a22", "a4c8f0d4eab41665b9fa121dcf661b94"},
		{8, "5e5863da2c462650c746e946c65552eca0e76f1e4f18510eb7929d4ab85349c2", "9daddc3c5b35ea847795bca33f59af52"},
	} {
		k, iv := callAESKeyIV(key, mk, c.x)
		if hex.EncodeToString(k) != c.key || hex.EncodeToString(iv) != c.iv {
			t.Fatalf("x=%d: %x %x", c.x, k, iv)
		}
	}
}

func TestCallCipherVectors(t *testing.T) {
	key := testKey()
	caller, callee := newCallCipher(key, true), newCallCipher(key, false)
	msgs := []string{"hello, signalling", `{"@type":"offer","sdp":"v=0"}`, "a slightly longer third message that spans more than two AES blocks"}
	want := []string{"d9240cce46c5ddeedc11e007b849c20586ed8912f90be0839bbc096f70818d3158b54ab937", "5f85c61f3d1a14b4c0c751739c0f19e58fe7c6f26216398ac9f2dbd60a95c8982fcaba6df895e753276037372b2ae82605", "8a289afebc9584f522d610fa1a4d1b12f07df590c6519c5cf8ac587bc2c5d9b32ba7193435d1ce3e27cea560b4b90021d8eac3f8edabb89a1b12ffd242e879a5f41d757e59edfc0373195e6886944bd672500a848a1a40"}
	for i, m := range msgs {
		pkt := caller.seal([]byte(m))
		if hex.EncodeToString(pkt) != want[i] {
			t.Fatalf("packet %d: %x", i+1, pkt)
		}
		got, err := callee.open(pkt)
		if err != nil || string(got) != m {
			t.Fatalf("open %d: %q %v", i+1, got, err)
		}
		if _, err := callee.open(pkt); err == nil {
			t.Fatal("a repeated packet was accepted")
		}
	}
	pkt := callee.seal([]byte(msgs[0]))
	if hex.EncodeToString(pkt) != "ff7a1ddf934dbaff8e87272121debe8697ab9a23c2e32549f2161e2e6de42890465437018e" {
		t.Fatalf("callee packet: %x", pkt)
	}
	if got, err := caller.open(pkt); err != nil || string(got) != msgs[0] {
		t.Fatalf("caller open: %q %v", got, err)
	}
	// What a side sends it cannot open itself (other key half), and a
	// changed byte fails the msg_key check.
	if _, err := newCallCipher(key, true).open(caller.seal([]byte("x-x-x-x"))); err == nil {
		t.Fatal("opened own packet")
	}
	bad := unhex(t, want[2])
	bad[30] ^= 1
	if _, err := newCallCipher(key, false).open(bad); err == nil {
		t.Fatal("opened a changed packet")
	}
	if _, err := callee.open(bad[:20]); err == nil {
		t.Fatal("opened a short packet")
	}
}

func TestCallReplayWindow(t *testing.T) {
	// counter:accepted, as registerIncomingCounter decided.
	const want = "1:1,2:1,2:0,5:1,3:1,3:0,100:1,37:1,36:0,99:1,101:1,50:1,60:1,37:0,165:1,101:0,102:1,1000:1,937:1,936:0,999:1,"
	c := newCallCipher(testKey(), false)
	var got strings.Builder
	for _, f := range strings.Split(strings.TrimSuffix(want, ","), ",") {
		n, _, _ := strings.Cut(f, ":")
		var v uint32
		for _, d := range n {
			v = v*10 + uint32(d-'0')
		}
		ok := 0
		if c.register(v) {
			ok = 1
		}
		got.WriteString(n + ":" + string(rune('0'+ok)) + ",")
	}
	if got.String() != want {
		t.Fatalf("window\n got %s\nwant %s", got.String(), want)
	}
	if len(c.seen) > keepCounters {
		t.Fatalf("%d counters kept", len(c.seen))
	}
}

func TestCallDH(t *testing.T) {
	// g^a mod p left-padded to 256 bytes, as OpenSSL computes it.
	p := new(big.Int).SetBytes(testKey())
	p.SetBit(p, 2047, 1).SetBit(p, 0, 1)
	a := make([]byte, callKeySize)
	for i := range a {
		a[i] = byte(i ^ 0x5a)
	}
	ga := pad256(new(big.Int).Exp(big.NewInt(3), new(big.Int).SetBytes(a), p))
	if hex.EncodeToString(ga) != "6cdafdc6efa5b30fd9c147f9e55afa86071d7c231eb50962248341add8bdfadb5aabbb4f0146008512762ee41306dfba2e503566735e0ed2189da9ef53967daf8b30ecfbc581ba878aa91b57c9c060aed4d7dc494cb40aed506e0c5e506eeda27ad77201496d0206748d15a284568c10508fb840428a1e7834d1150a2ca58291bc3d81c3718b85bca68ddc6437037c59f3fce06adc1b557047dd53c7ec0bce7e914855ba0226661068690fc24b200dbb2df1b2a4d3a74468a350c903ae2bf825a567357293a3c776629aef9b869e7d656ac9b6afd7ffda99e8fefe1ef46d9920a0d3d0114bb4b4b6ff2ba06498e3368919b368b56e131c20a3283c205cbc42d5" {
		t.Fatalf("g^a: %x", ga)
	}
	if got := pad256(big.NewInt(5)); len(got) != callKeySize || got[255] != 5 || !bytes.Equal(got[:255], make([]byte, 255)) {
		t.Fatal("padding")
	}

	// A full exchange over Telegram's test prime.
	dh := dhParams{g: 3, p: testPrime()}
	sa, sb := dhSecret(nil), dhSecret([]byte{1, 2, 3})
	gA, err := dhPublic(dh, sa)
	if err != nil {
		t.Fatal(err)
	}
	gB, err := dhPublic(dh, sb)
	if err != nil {
		t.Fatal(err)
	}
	k1, err1 := dhShared(dh, gB, sa)
	k2, err2 := dhShared(dh, gA, sb)
	if err1 != nil || err2 != nil || !bytes.Equal(k1, k2) || len(k1) != callKeySize {
		t.Fatal("keys differ")
	}
	if _, err := dhShared(dh, []byte{1}, sa); err == nil {
		t.Fatal("g_b = 1 accepted")
	}
	pm1 := new(big.Int).Sub(dh.p, big.NewInt(1)).Bytes()
	if _, err := dhShared(dh, pm1, sa); err == nil {
		t.Fatal("g_b = p-1 accepted")
	}
}

// testPrime is the 2048-bit safe prime Telegram's test servers use (g = 3).
func testPrime() *big.Int {
	p, _ := new(big.Int).SetString("C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F"+
		"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37"+
		"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64"+
		"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4"+
		"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754"+
		"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4"+
		"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F"+
		"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B", 16)
	return p
}
