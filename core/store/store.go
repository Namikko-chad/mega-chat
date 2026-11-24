// Package store is the local, encrypted cache of the core: provider
// session state (KV), chats and messages of every account. It is a bbolt
// file; bucket and key names are HMACs and values are sealed with
// XChaCha20-Poly1305 bound to their location, so the file shows neither
// chat names nor ids nor text. The file is locked: one process per store.
//
// Messages of a chat are kept in time order. The cache usually holds parts
// of a history (latest messages from live events, pages fetched when the
// user scrolled), so every message records whether the next older cached
// message really is its predecessor (Linked) and whether it is the first
// message of the chat (First). History reads only contiguous runs and
// tells the caller where the run ended.
package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"megachat/core/crypto"
	"megachat/core/model"
	"megachat/core/provider"
)

var (
	bKV    = []byte("kv")
	bChats = []byte("chats")
	bMsgs  = []byte("msgs")
	bMids  = []byte("mids") // message id -> order key
	bMedia = []byte("media")
	bMIdx  = []byte("mediaidx") // media key -> size and last use
)

// ErrLocked means another process has the store open.
var ErrLocked = errors.New("store: in use by another process")

// Store is an open cache.
type Store struct {
	db    *bolt.DB
	seal  crypto.Key
	names crypto.Key

	mediaMu    sync.Mutex
	mediaTotal int64 // bytes in the media cache; -1: not counted yet
}

// Open opens (creating) the store at path with keys derived from master.
func Open(path string, master crypto.Key) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if errors.Is(err, bolt.ErrTimeout) {
		return nil, ErrLocked
	}
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bKV, bChats, bMsgs, bMids, bMedia, bMIdx} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db, seal: master.Derive("store-seal", nil), names: master.Derive("store-names", nil), mediaTotal: -1}, nil
}

// Close closes the file.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) acctName(acct string) []byte {
	return s.names.MAC([]byte("acct\x00"), []byte(acct))[:16]
}

func (s *Store) chatName(acct, chat string) []byte {
	return s.names.MAC([]byte("chat\x00"), []byte(acct), []byte{0}, []byte(chat))[:16]
}

func (s *Store) keyName(parts ...string) []byte {
	m := s.names.NewMAC()
	for _, p := range parts {
		m.Write([]byte(p))
		m.Write([]byte{0})
	}
	return m.Sum(nil)[:16]
}

func (s *Store) put(b *bolt.Bucket, ad []byte, key []byte, v any) error {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ct, err := crypto.Seal(s.seal, js, append(append([]byte{}, ad...), key...))
	if err != nil {
		return err
	}
	return b.Put(key, ct)
}

func (s *Store) get(b *bolt.Bucket, ad []byte, key []byte, v any) (bool, error) {
	ct := b.Get(key)
	if ct == nil {
		return false, nil
	}
	return true, s.decode(ad, key, ct, v)
}

func (s *Store) decode(ad, key, ct []byte, v any) error {
	js, err := crypto.Open(s.seal, ct, append(append([]byte{}, ad...), key...))
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	return json.Unmarshal(js, v)
}

// sub returns the nested bucket root/names..., creating it when create.
func sub(tx *bolt.Tx, create bool, root []byte, names ...[]byte) (*bolt.Bucket, error) {
	b := tx.Bucket(root)
	for _, n := range names {
		if b == nil {
			return nil, nil
		}
		if create {
			var err error
			if b, err = b.CreateBucketIfNotExists(n); err != nil {
				return nil, err
			}
		} else {
			b = b.Bucket(n)
		}
	}
	return b, nil
}

// --- KV ------------------------------------------------------------------

type kv struct {
	s    *Store
	acct []byte
}

// KV returns the provider state of an account.
func (s *Store) KV(accountID string) provider.KV { return &kv{s: s, acct: s.acctName(accountID)} }

func (k *kv) ad() []byte { return append([]byte("kv\x00"), k.acct...) }

func (k *kv) Get(key string) ([]byte, error) {
	var out []byte
	err := k.s.db.View(func(tx *bolt.Tx) error {
		b, _ := sub(tx, false, bKV, k.acct)
		if b == nil {
			return nil
		}
		_, err := k.s.get(b, k.ad(), k.s.keyName(key), &out)
		return err
	})
	return out, err
}

func (k *kv) Put(key string, v []byte) error {
	return k.s.db.Update(func(tx *bolt.Tx) error {
		b, err := sub(tx, true, bKV, k.acct)
		if err != nil {
			return err
		}
		return k.s.put(b, k.ad(), k.s.keyName(key), v)
	})
}

func (k *kv) Delete(key string) error {
	return k.s.db.Update(func(tx *bolt.Tx) error {
		b, _ := sub(tx, false, bKV, k.acct)
		if b == nil {
			return nil
		}
		return b.Delete(k.s.keyName(key))
	})
}

// --- chats ---------------------------------------------------------------

func (s *Store) chatAD(acct []byte) []byte { return append([]byte("chat\x00"), acct...) }

// PutChat stores (replaces) a chat.
func (s *Store) PutChat(c model.Chat) error {
	a := s.acctName(c.AccountID)
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := sub(tx, true, bChats, a)
		if err != nil {
			return err
		}
		return s.put(b, s.chatAD(a), s.chatName(c.AccountID, c.ID), c)
	})
}

// UpdateChat applies fn to the stored chat (ok false if there is none) and
// stores the result unless fn returns false.
func (s *Store) UpdateChat(acct, id string, fn func(c *model.Chat, ok bool) bool) (model.Chat, error) {
	a := s.acctName(acct)
	var c model.Chat
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := sub(tx, true, bChats, a)
		if err != nil {
			return err
		}
		key := s.chatName(acct, id)
		ok, err := s.get(b, s.chatAD(a), key, &c)
		if err != nil {
			return err
		}
		if !ok {
			c = model.Chat{AccountID: acct, ID: id}
		}
		if !fn(&c, ok) {
			return nil
		}
		return s.put(b, s.chatAD(a), key, c)
	})
	return c, err
}

// Chat returns one chat.
func (s *Store) Chat(acct, id string) (model.Chat, bool, error) {
	a := s.acctName(acct)
	var c model.Chat
	var ok bool
	err := s.db.View(func(tx *bolt.Tx) error {
		b, _ := sub(tx, false, bChats, a)
		if b == nil {
			return nil
		}
		var err error
		ok, err = s.get(b, s.chatAD(a), s.chatName(acct, id), &c)
		return err
	})
	return c, ok, err
}

// Chats returns all chats of an account, latest activity first.
func (s *Store) Chats(acct string) ([]model.Chat, error) {
	a := s.acctName(acct)
	var out []model.Chat
	err := s.db.View(func(tx *bolt.Tx) error {
		b, _ := sub(tx, false, bChats, a)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var c model.Chat
			if err := s.decode(s.chatAD(a), k, v, &c); err != nil {
				return err
			}
			out = append(out, c)
			return nil
		})
	})
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	return out, err
}

// DeleteChat forgets a chat and its messages.
func (s *Store) DeleteChat(acct, id string) error {
	a, c := s.acctName(acct), s.chatName(acct, id)
	return s.db.Update(func(tx *bolt.Tx) error {
		if b, _ := sub(tx, false, bChats, a); b != nil {
			if err := b.Delete(c); err != nil {
				return err
			}
		}
		for _, root := range [][]byte{bMsgs, bMids} {
			if b, _ := sub(tx, false, root, a); b != nil && b.Bucket(c) != nil {
				if err := b.DeleteBucket(c); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// DeleteAccount forgets everything of an account.
func (s *Store) DeleteAccount(acct string) error {
	a := s.acctName(acct)
	s.mediaMu.Lock()
	s.mediaTotal = -1
	s.mediaMu.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, root := range [][]byte{bKV, bChats, bMsgs, bMids, bMedia, bMIdx} {
			if b := tx.Bucket(root); b.Bucket(a) != nil {
				if err := b.DeleteBucket(a); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// --- messages ------------------------------------------------------------

// record is a stored message with its place in the history.
type record struct {
	Msg model.Message `json:"m"`
	// Linked: the next older cached message is this one's predecessor.
	Linked bool `json:"l,omitempty"`
	// First: there is no older message in the chat.
	First bool `json:"f,omitempty"`
}

func (s *Store) msgAD(a, c []byte) []byte {
	return append(append([]byte("msg\x00"), a...), c...)
}

// orderKey sorts by time, then Seq, then a keyed hash of the id.
func (s *Store) orderKey(m model.Message) []byte {
	k := make([]byte, 0, 24)
	ns := m.Time.UnixNano()
	if ns < 0 {
		ns = 0
	}
	k = binary.BigEndian.AppendUint64(k, uint64(ns))
	k = binary.BigEndian.AppendUint64(k, uint64(m.Seq))
	return append(k, s.keyName("mid", m.ID)[:8]...)
}

type msgTx struct {
	s          *Store
	msgs, mids *bolt.Bucket
	ad         []byte
}

func (s *Store) msgBuckets(tx *bolt.Tx, create bool, acct, chat string) (*msgTx, error) {
	a, c := s.acctName(acct), s.chatName(acct, chat)
	msgs, err := sub(tx, create, bMsgs, a, c)
	if err != nil || msgs == nil {
		return nil, err
	}
	mids, err := sub(tx, create, bMids, a, c)
	if err != nil || mids == nil {
		return nil, err
	}
	return &msgTx{s: s, msgs: msgs, mids: mids, ad: s.msgAD(a, c)}, nil
}

func (t *msgTx) find(id string) ([]byte, *record, error) {
	key := t.mids.Get(t.s.keyName("mid", id))
	if key == nil {
		return nil, nil, nil
	}
	var r record
	ok, err := t.s.get(t.msgs, t.ad, key, &r)
	if err != nil || !ok {
		return nil, nil, err
	}
	return append([]byte(nil), key...), &r, nil
}

func (t *msgTx) write(r *record) error {
	key, old, err := t.find(r.Msg.ID)
	if err != nil {
		return err
	}
	nk := t.s.orderKey(r.Msg)
	if old != nil && !bytes.Equal(key, nk) {
		if err := t.msgs.Delete(key); err != nil {
			return err
		}
	}
	if err := t.s.put(t.msgs, t.ad, nk, r); err != nil {
		return err
	}
	return t.mids.Put(t.s.keyName("mid", r.Msg.ID), nk)
}

// PutMessage stores a message (new or changed), keeping its place in the
// history. It returns the previously stored version, if any.
func (s *Store) PutMessage(m model.Message) (prev *model.Message, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, true, m.AccountID, m.ChatID)
		if err != nil {
			return err
		}
		_, old, err := t.find(m.ID)
		if err != nil {
			return err
		}
		r := &record{Msg: m}
		if old != nil {
			prev = &old.Msg
			if Blank(m) {
				return nil // never replaces what is stored
			}
			r.Linked, r.First = old.Linked, old.First
		}
		return t.write(r)
	})
	return prev, err
}

// Blank tells that m carries no content at all: a provider that failed to
// read a message, not a real change (deletions come as Deleted).
func Blank(m model.Message) bool {
	return m.Text == "" && len(m.Attachments) == 0 && !m.Deleted && !m.Service
}

// UpdateMessage applies fn to a stored message; ok is false if it is not
// cached (fn is not called).
func (s *Store) UpdateMessage(acct, chat, id string, fn func(m *model.Message)) (m model.Message, ok bool, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		_, r, err := t.find(id)
		if err != nil || r == nil {
			return err
		}
		fn(&r.Msg)
		m, ok = r.Msg, true
		return t.write(r)
	})
	return m, ok, err
}

// DeleteMessage drops a message from the cache (not a tombstone: the
// message must not exist, e.g. a broken record). The next newer message
// loses its link, since its predecessor is no longer known.
func (s *Store) DeleteMessage(acct, chat, id string) (ok bool, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		key, r, err := t.find(id)
		if err != nil || r == nil {
			return err
		}
		if err := t.msgs.Delete(key); err != nil {
			return err
		}
		if err := t.mids.Delete(s.keyName("mid", id)); err != nil {
			return err
		}
		ok = true
		nk, ct := t.msgs.Cursor().Seek(key)
		if nk == nil {
			return nil
		}
		var next record
		if err := s.decode(t.ad, nk, ct, &next); err != nil {
			return err
		}
		if !next.Linked {
			return nil
		}
		next.Linked = false
		return t.write(&next)
	})
	return ok, err
}

// Message returns one cached message.
func (s *Store) Message(acct, chat, id string) (model.Message, bool, error) {
	var out model.Message
	var ok bool
	err := s.db.View(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		_, r, err := t.find(id)
		if r != nil {
			out, ok = r.Msg, true
		}
		return err
	})
	return out, ok, err
}

// End tells why History stopped.
type End int

const (
	EndLimit End = iota // limit reached
	EndGap              // older messages are not cached (fetch before the last one)
	EndFirst            // the chat's first message was reached
)

// History returns up to limit cached messages before beforeID (the latest
// when empty), newest first, as far as they are contiguous. With a
// beforeID that is not cached, or whose predecessor is not, it returns
// nothing and EndGap.
func (s *Store) History(acct, chat, beforeID string, limit int) ([]model.Message, End, error) {
	var out []model.Message
	end := EndGap
	err := s.db.View(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		c := t.msgs.Cursor()
		var k, v []byte
		if beforeID == "" {
			k, v = c.Last()
		} else {
			key, r, err := t.find(beforeID)
			if err != nil || r == nil {
				return err
			}
			if r.First {
				end = EndFirst
				return nil
			}
			if !r.Linked {
				return nil
			}
			c.Seek(key)
			k, v = c.Prev()
		}
		for k != nil {
			if len(out) == limit {
				end = EndLimit
				return nil
			}
			var r record
			if err := s.decode(t.ad, k, v, &r); err != nil {
				return err
			}
			out = append(out, r.Msg)
			if r.First {
				end = EndFirst
				return nil
			}
			if !r.Linked {
				return nil
			}
			k, v = c.Prev()
		}
		return nil
	})
	if len(out) == limit && end == EndGap {
		end = EndLimit
	}
	return out, end, err
}

// PutHistory stores a contiguous page fetched from the provider: msgs are
// the messages right before beforeID (the latest when empty), newest
// first; more is false when the page reaches the chat's first message.
func (s *Store) PutHistory(acct, chat, beforeID string, msgs []model.Message, more bool) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, true, acct, chat)
		if err != nil {
			return err
		}
		if beforeID != "" {
			if _, r, err := t.find(beforeID); err != nil {
				return err
			} else if r != nil {
				if len(msgs) > 0 {
					r.Linked = true
				} else if !more {
					r.First = true
				}
				if err := t.write(r); err != nil {
					return err
				}
			}
		}
		for i, m := range msgs {
			m.AccountID, m.ChatID = acct, chat
			r := &record{Msg: m}
			if _, old, err := t.find(m.ID); err != nil {
				return err
			} else if old != nil {
				r.Linked, r.First = old.Linked, old.First
				if Blank(m) {
					r.Msg.Text, r.Msg.Format, r.Msg.Attachments = old.Msg.Text, old.Msg.Format, old.Msg.Attachments
				}
			}
			if i < len(msgs)-1 {
				r.Linked = true
			} else if !more {
				r.First = true
			}
			if err := t.write(r); err != nil {
				return err
			}
		}
		return nil
	})
}

// Recent returns up to limit cached messages of a chat, newest first,
// whether they are contiguous or not: what was seen of the chat, to show
// while the real history (with its gaps filled) is on its way.
func (s *Store) Recent(acct, chat string, limit int) ([]model.Message, error) {
	var out []model.Message
	err := s.db.View(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		c := t.msgs.Cursor()
		for k, v := c.Last(); k != nil && len(out) < limit; k, v = c.Prev() {
			var r record
			if err := s.decode(t.ad, k, v, &r); err != nil {
				return err
			}
			out = append(out, r.Msg)
		}
		return nil
	})
	return out, err
}

// CountAfter counts the cached messages newer than id that match, looking
// at no more than limit messages. found is false when id was not among them.
func (s *Store) CountAfter(acct, chat, id string, limit int, match func(model.Message) bool) (n int, found bool, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		t, err := s.msgBuckets(tx, false, acct, chat)
		if err != nil || t == nil {
			return err
		}
		c := t.msgs.Cursor()
		seen := 0
		for k, v := c.Last(); k != nil && seen < limit; k, v = c.Prev() {
			var r record
			if err := s.decode(t.ad, k, v, &r); err != nil {
				return err
			}
			if r.Msg.ID == id {
				found = true
				return nil
			}
			if match(r.Msg) {
				n++
			}
			seen++
		}
		return nil
	})
	return n, found, err
}
