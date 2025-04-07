package mutex

import "sync"

// Mutexes is a collection of Mutex objects that can be locked and unlocked
// together. For shared operations, any Mutex not implementing the RWMutex
// interface will use exclusive locking as fallback.
type Mutexes []Mutex

// Lock acquires all locks in a deadlock-safe manner.
// If any acquisition fails, all previously acquired locks are released
// to prevent partial locking states.
func (ms Mutexes) Lock() {
	if err := doLock(ms); err != nil {
		panic(err)
	}
}

// Unlock releases all locks in the collection.
// Attempts to release all locks even if some operations fail.
// All errors encountered are aggregated into a single panic.
func (ms Mutexes) Unlock() {
	if err := doUnlock(ms); err != nil {
		panic(err)
	}
}

// TryLock attempts to acquire all locks without blocking.
// Returns true if all locks were acquired, false otherwise.
// Any partial acquisitions are automatically reversed.
func (ms Mutexes) TryLock() bool {
	ok, err := doTryLock(ms)
	if err != nil {
		panic(err)
	}
	return ok
}

// RLock acquires shared (read) locks on all Mutexes in the collection.
// Falls back to exclusive locks for Mutex objects not implementing RWMutex.
func (ms Mutexes) RLock() {
	if err := doRLock(ms); err != nil {
		panic(err)
	}
}

// RUnlock releases shared (read) locks on all Mutexes in the collection.
// Also releases exclusive locks acquired as fallbacks during RLock.
// All errors are aggregated into a single panic.
func (ms Mutexes) RUnlock() {
	if err := doRUnlock(ms); err != nil {
		panic(err)
	}
}

// TryRLock attempts to acquire shared (read) locks without blocking.
// Returns true if successful, false otherwise.
// Falls back to exclusive locks for Mutex objects not implementing RWMutex.
func (ms Mutexes) TryRLock() bool {
	ok, err := doTryRLock(ms)
	if err != nil {
		panic(err)
	}
	return ok
}

// Type assertions to ensure interface compliance
var _ sync.Locker = Mutexes{}
var _ Mutex = Mutexes{}
var _ RWMutex = Mutexes{}
