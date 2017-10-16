package shorty

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

var (
	bucketKeyB  = []byte("shorty")
	counterKeyB = []byte("_counter_")
	be          = binary.BigEndian
)

type boltStore struct {
	db *bolt.DB
}

// NewBoltStore returns a new bolt-based store with the specified bolt.DB.
func NewBoltStore(db *bolt.DB) (Store, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketKeyB)
		return err
	}); err != nil {
		return nil, err
	}
	return &boltStore{db}, nil
}

var _ Store = (*boltStore)(nil)

func (bs *boltStore) Get(id string) (*Data, error) {
	var d Data
	if err := bs.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketKeyB).Get([]byte(id))
		return json.Unmarshal(v, &d)
	}); err != nil {
		return nil, err
	}
	return &d, nil
}

func (bs *boltStore) Put(genIDFn func(counter uint32) ID, value *Data) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyB)
		if !value.ID.Valid() {
			value.ID = genIDFn(bs.incCounter(b))
		}
		return bs.putJSON(b, value.ID.String(), value)
	})
}

func (bs *boltStore) Delete(ids ...string) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKeyB)
		for _, id := range ids {
			if err := b.Delete([]byte(id)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bs *boltStore) ForEach(fn func(id string, value *Data) error) error {
	return bs.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketKeyB).ForEach(func(k []byte, v []byte) error {
			if k[0] == '_' {
				return nil
			}
			var d Data
			if err := json.Unmarshal(v, &d); err != nil {
				return fmt.Errorf("error unmarshalling %q (%q): %v", k, v, err)
			}
			return fn(string(k), &d)
		})
	})
}

// Close is a no-op because the db owner should handle closing it.
func (bs *boltStore) Close() error { return nil }

func (bs *boltStore) putJSON(b *bolt.Bucket, key string, v interface{}) error {
	j, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), j)
}

func (bs *boltStore) incCounter(b *bolt.Bucket) (cnt uint32) {
	if v := b.Get(counterKeyB); len(v) == 4 {
		cnt = be.Uint32(v)
	}
	var buf [4]byte
	be.PutUint32(buf[:], cnt+1)
	b.Put(counterKeyB, buf[:])
	return
}
