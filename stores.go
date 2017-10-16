package shorty

import (
	"sort"
	"sync"

	"github.com/missionMeteora/toolkit/errors"
)

// Store represents a simple interface for storing and loading shorty's data
type Store interface {
	// Get returns the data associated with that id.
	Get(id string) (*Data, error)

	// Put puts/updates value, if value has a valid id it will be used to update the old value,
	// otherwise it will call genIDFn with the next counter value and assign the new ID to value.ID.
	Put(genIDFn func(counter uint32) ID, value *Data) error

	// Delete deletes the specific id from the store.
	Delete(ids ...string) error

	// ForEach loops over all the valid ids in the store, returning an error will break early.
	ForEach(fn func(id string, v *Data) error) error

	// Close closes the underlying store if applicable.
	Close() error
}

// memStore implements a store using map[string]interface{}
type memStore struct {
	s       map[string]Data
	counter uint32
	mux     sync.RWMutex
}

// NewMemStore returns a new in-memory store.
func NewMemStore() Store {
	return &memStore{
		s: map[string]Data{},
	}
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

func (ms *memStore) Delete(ids ...string) error {
	ms.mux.Lock()
	for _, id := range ids {
		delete(ms.s, id)
	}
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

func (ms *memStore) Close() error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	if ms.s == nil {
		return errors.ErrIsClosed
	}
	ms.s = nil
	return nil
}
