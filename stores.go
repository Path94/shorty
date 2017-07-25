package shorty

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/boltdb/bolt"
)

var (
	bucketKey  = []byte("shorty")
	counterKey = []byte("_counter_")
	be         = binary.BigEndian
)

// Store represents a simple interface for storing and loading shorty's data
type Store interface {
	// Get returns the data associated with that id.
	Get(id string) (*Data, error)

	// Put puts/updates value, if value has a valid id it will be used to update the old value,
	// otherwise it will call genIDFn with the next counter value and assign the new ID to value.ID.
	Put(genIDFn func(counter uint32) ID, value *Data) error

	// Delete deletes the specific id from the store.
	Delete(id string) error

	// ForEach loops over all the valid ids in the store, returning an error will break early.
	ForEach(fn func(id string, v *Data) error) error
}

type boltStore struct {
	db *bolt.DB
}

// NewBoltStore returns a new bolt-based store with the specified bolt.DB.
func NewBoltStore(db *bolt.DB) (Store, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketKey)
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
		v := tx.Bucket(bucketKey).Get([]byte(id))
		return json.Unmarshal(v, &d)
	}); err != nil {
		return nil, err
	}
	return &d, nil
}

func (bs *boltStore) Put(genIDFn func(counter uint32) ID, value *Data) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKey)
		if !value.ID.Valid() {
			value.ID = genIDFn(incCounter(b))
		}
		return putJSON(b, value.ID.String(), value)
	})
}

func (bs *boltStore) Delete(id string) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketKey)
		return b.Delete([]byte(id))
	})
}

func (bs *boltStore) ForEach(fn func(id string, value *Data) error) error {
	return bs.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketKey).ForEach(func(k []byte, v []byte) error {
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

// boltStore implements a store based on bolt.
type memStore struct {
	s       map[string]Data
	counter uint32
	mux     sync.RWMutex
}

// NewMemStore returns a new in-memory store.
func NewMemStore() (Store, error) {
	return &memStore{
		s: map[string]Data{},
	}, nil
}

var _ Store = (*memStore)(nil)

func (ms *memStore) Get(id string) (*Data, error) {
	ms.mux.RLock()
	d, ok := ms.s[id]
	ms.mux.RUnlock()
	if !ok {
		return nil, nil
	}
	return &d, nil
}

func (ms *memStore) Put(genIDFn func(counter uint32) ID, value *Data) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	if !value.ID.Valid() {
		value.ID = genIDFn(ms.counter)
		ms.counter++
	}
	ms.s[value.ID.String()] = *value // copy
	return nil
}

func (ms *memStore) Delete(id string) error {
	ms.mux.Lock()
	delete(ms.s, id)
	ms.mux.Unlock()
	return nil
}

func (ms *memStore) ForEach(fn func(id string, value *Data) error) error {
	ms.mux.RLock()
	ids := make([]string, 0, len(ms.s))
	for id := range ms.s {
		ids = append(ids, id)
	}
	ms.mux.RUnlock()
	sort.Strings(ids)
	for _, id := range ids {
		v, _ := ms.Get(id)
		if v == nil { // id got deleted
			continue
		}
		if err := fn(id, v); err != nil {
			return err
		}
	}
	return nil
}

func putJSON(b *bolt.Bucket, key string, v interface{}) error {
	j, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), j)
}

func incCounter(b *bolt.Bucket) (cnt uint32) {
	if v := b.Get(counterKey); len(v) > 7 {
		cnt = be.Uint32(v)
	}
	var buf [4]byte
	be.PutUint32(buf[:], cnt+1)
	b.Put(counterKey, buf[:])
	return
}
