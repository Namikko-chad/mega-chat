package store

import (
	"encoding/binary"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"

	"megachat/core/crypto"
)

// Media cache: downloaded avatars, previews and small files, per account,
// sealed like everything else. When the cache grows over MediaCap the
// least recently used entries go.

// MediaCap is the size the media cache is kept under.
var MediaCap int64 = 256 << 20

type mediaIdx struct {
	Size int64 `json:"s"`
	Used int64 `json:"u"` // unix ns
}

func (s *Store) mediaAD(a []byte) []byte { return append([]byte("media\x00"), a...) }

// encodeMedia packs the MIME type before the content.
func encodeMedia(data []byte, mime string) []byte {
	out := binary.AppendUvarint(make([]byte, 0, len(data)+len(mime)+2), uint64(len(mime)))
	return append(append(out, mime...), data...)
}

func decodeMedia(b []byte) ([]byte, string, bool) {
	n, k := binary.Uvarint(b)
	if k <= 0 || uint64(len(b)-k) < n {
		return nil, "", false
	}
	return b[k+int(n):], string(b[k : k+int(n)]), true
}

// Media returns a cached media file of an account.
func (s *Store) Media(acct, ref string) (data []byte, mime string, ok bool, err error) {
	a := s.acctName(acct)
	key := s.keyName("media", ref)
	err = s.db.View(func(tx *bolt.Tx) error {
		b, err := sub(tx, false, bMedia, a)
		if err != nil || b == nil {
			return err
		}
		ct := b.Get(key)
		if ct == nil {
			return nil
		}
		ad := s.mediaAD(a)
		plain, err := crypto.Open(s.seal, ct, append(append([]byte{}, ad...), key...))
		if err != nil {
			return err
		}
		data, mime, ok = decodeMedia(plain)
		return nil
	})
	if ok {
		// Last use is approximate; a failed touch costs nothing.
		_ = s.db.Update(func(tx *bolt.Tx) error {
			idx, err := sub(tx, true, bMIdx, a)
			if err != nil {
				return err
			}
			return s.put(idx, s.mediaAD(a), key, mediaIdx{Size: int64(len(data)), Used: time.Now().UnixNano()})
		})
	}
	return data, mime, ok, err
}

// PutMedia caches a media file of an account.
func (s *Store) PutMedia(acct, ref string, data []byte, mime string) error {
	a := s.acctName(acct)
	key := s.keyName("media", ref)
	ad := s.mediaAD(a)
	ct, err := crypto.Seal(s.seal, encodeMedia(data, mime), append(append([]byte{}, ad...), key...))
	if err != nil {
		return err
	}
	var total int64
	s.mediaMu.Lock()
	defer s.mediaMu.Unlock()
	err = s.db.Update(func(tx *bolt.Tx) error {
		b, err := sub(tx, true, bMedia, a)
		if err != nil {
			return err
		}
		idx, err := sub(tx, true, bMIdx, a)
		if err != nil {
			return err
		}
		if s.mediaTotal < 0 {
			if s.mediaTotal, err = s.mediaSize(tx); err != nil {
				return err
			}
		}
		var old mediaIdx
		if ok, _ := s.get(idx, ad, key, &old); ok {
			s.mediaTotal -= old.Size
		}
		if err := b.Put(key, ct); err != nil {
			return err
		}
		if err := s.put(idx, ad, key, mediaIdx{Size: int64(len(data)), Used: time.Now().UnixNano()}); err != nil {
			return err
		}
		s.mediaTotal += int64(len(data))
		total = s.mediaTotal
		return nil
	})
	if err != nil {
		s.mediaTotal = -1
		return err
	}
	if total <= MediaCap {
		return nil
	}
	err = s.pruneMedia(MediaCap * 3 / 4)
	s.mediaTotal = -1
	return err
}

type mediaEntry struct {
	acct, key []byte
	mediaIdx
}

func (s *Store) mediaEntries(tx *bolt.Tx) ([]mediaEntry, error) {
	var out []mediaEntry
	root := tx.Bucket(bMIdx)
	err := root.ForEachBucket(func(a []byte) error {
		b := root.Bucket(a)
		ad := s.mediaAD(a)
		return b.ForEach(func(k, v []byte) error {
			var e mediaIdx
			if s.decode(ad, k, v, &e) != nil {
				e = mediaIdx{} // broken: oldest, goes first
			}
			out = append(out, mediaEntry{acct: append([]byte(nil), a...), key: append([]byte(nil), k...), mediaIdx: e})
			return nil
		})
	})
	return out, err
}

func (s *Store) mediaSize(tx *bolt.Tx) (int64, error) {
	es, err := s.mediaEntries(tx)
	var n int64
	for _, e := range es {
		n += e.Size
	}
	return n, err
}

// pruneMedia drops the least recently used media until the cache is
// under keep bytes.
func (s *Store) pruneMedia(keep int64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		es, err := s.mediaEntries(tx)
		if err != nil {
			return err
		}
		var total int64
		for _, e := range es {
			total += e.Size
		}
		sort.Slice(es, func(i, j int) bool { return es[i].Used < es[j].Used })
		for _, e := range es {
			if total <= keep {
				break
			}
			if b := tx.Bucket(bMedia).Bucket(e.acct); b != nil {
				if err := b.Delete(e.key); err != nil {
					return err
				}
			}
			if err := tx.Bucket(bMIdx).Bucket(e.acct).Delete(e.key); err != nil {
				return err
			}
			total -= e.Size
		}
		return nil
	})
}
