package store

import (
	"sync"
	"sync/atomic"

	"github.com/hashicorp/golang-lru"
)

// memStore implements an in-memory Store.
type memStore struct {
	sync.RWMutex
	keys *lru.Cache
	m    map[string]*int64
}

// NewMemStore creates a new MemStore. If maxKeys > 0, the number of different keys
// is restricted to the specified amount. In this case, it uses an LRU algorithm to
// evict older keys to make room for newer ones. If a request is made for a key that
// has been evicted, it will be processed as if its count was 0, possibly allowing requests
// that should be denied.
//
// If maxKeys <= 0, there is no limit on the number of keys, which may use an unbounded amount of
// memory depending on the server's load.
//
// The MemStore is only for single-process rate-limiting. To share the rate limit state
// among multiple instances of the web server, use a database- or key-value-based
// store.
//
func NewMemStore(maxKeys int) (Store, error) {
	var m *memStore

	if maxKeys > 0 {
		keys, err := lru.New(maxKeys)
		if err != nil {
			return nil, err
		}

		m = &memStore{
			keys: keys,
		}
	} else {
		m = &memStore{
			m: make(map[string]*int64),
		}
	}
	return m, nil
}

func (ms *memStore) Get(key string) (int64, error) {
	valP, ok := ms.get(key, false)

	if !ok {
		return 0, ErrNoSuchKey
	}

	return atomic.LoadInt64(valP), nil
}

func (ms *memStore) SetNX(key string, value int64) (bool, error) {
	_, ok := ms.get(key, false)

	if ok {
		return false, nil
	}

	ms.Lock()
	defer ms.Unlock()

	_, ok = ms.get(key, true)

	if ok {
		return false, nil
	}

	// Store a pointer to a new instance so that the caller
	// can't mutate the value after setting
	v := value

	if ms.keys != nil {
		ms.keys.Add(key, &v)
	} else {
		ms.m[key] = &v
	}

	return true, nil
}

func (ms *memStore) CompareAndSwap(key string, old, new int64) (bool, error) {
	valP, ok := ms.get(key, false)

	if !ok {
		return false, ErrNoSuchKey
	}

	return atomic.CompareAndSwapInt64(valP, old, new), nil
}

func (ms *memStore) get(key string, locked bool) (*int64, bool) {
	var valP *int64
	var ok bool

	if ms.keys != nil {
		var valI interface{}

		valI, ok = ms.keys.Get(key)
		if ok {
			valP = valI.(*int64)
		}
	} else {
		if !locked {
			ms.RLock()
			defer ms.RUnlock()
		}
		valP, ok = ms.m[key]
	}

	return valP, ok
}
