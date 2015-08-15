// Package store offers in-memory and Redis-based stores for throttled.
package store // import "gopkg.in/throttled/throttled.v0/store"

import (
	"time"
)

// GCRAStore is the interface to implement to store state for a GCRA rate limiter
// TODO: It's a bit funny that we pass a separate TTL when it's actually equal to value - now
// would it be better to make this interface much more GCRA specific?
type GCRAStore interface {
	// Get returns the value of the key if it is in the store or -1 if it does
	// not exist. Also returns the current time at the Store. The time must
	// be representable as a positive int64 of nanoseconds since the epoch.
	//
	// GCRA assumes that all instances sharing the same Store also share the
	// same clock. Using separate clocks will work if the skew is small but
	// not recommended in practice unless you're lucky enough to be hooked up to
	// GPS or atomic clocks.
	GetWithTime(key string) (int64, time.Time, error)

	// SetIfNotExists sets the value of key only if it is not already set in the store
	// it returns whether a new value was set. If the store supports expiring
	// keys and a new value was set, the key will expire after the provided ttl.
	SetIfNotExists(key string, value int64, ttl time.Duration) (bool, error)

	// CompareAndSwap atomically compares the value at key to the old value.
	// If it matches, it sets it to the new value and returns true. Otherwise,
	// it returns false. If the key does not exist in the store, it returns
	// false with no error. If the store supports expiring keys and the swap
	// succeeded, the key will expire after the provided ttl.
	CompareAndSwap(key string, old, new int64, ttl time.Duration) (bool, error)
}