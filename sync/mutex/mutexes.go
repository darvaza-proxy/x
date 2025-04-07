package mutex

import "sync"

// Mutexes is a collection of Mutex objects that can be locked and unlocked together.
// When using shared (read) lock operations, any Mutex that doesn't implement the
// RWMutex interface will fall back to exclusive locking.
type Mutexes []Mutex

// Lock acquires all locks in the collection in a deadlock-safe way.
// If any lock acquisition fails, any locks already acquired will be
// automatically released to prevent partial locking states.
func (ms Mutexes) Lock() {
	if err := doLock(ms); err != nil {
		panic(err)
	}
}

// Unlock releases all locks in the collection.
// It will attempt to unlock all locks even if some operations fail.
// Any errors or panics encountered during unlocking will be aggregated and
// returned as a single panic.
func (ms Mutexes) Unlock() {
	if err := doUnlock(ms); err != nil {
		panic(err)
	}
}

// TryLock attempts to acquire all locks without blocking.
// Returns true if all locks were acquired, false otherwise.
// If any lock cannot be acquired, all previously acquired locks
// within this attempt will be automatically released.
func (ms Mutexes) TryLock() bool {
	ok, err := doTryLock(ms)
	if err != nil {
		panic(err)
	}
	return ok
}

// RLock acquires shared (read) locks on all Mutexes in the collection.
// If a Mutex does not implement the RWMutex interface, an exclusive lock
// will be used as a fallback.
func (ms Mutexes) RLock() {
	if err := doRLock(ms); err != nil {
		panic(err)
	}
}

// RUnlock releases shared (read) locks on all Mutexes in the collection.
// If a Mutex does not implement the RWMutex interface, the corresponding
// exclusive lock acquired during RLock will be released.
// It will attempt to unlock all locks even if some operations fail.
// Any errors or panics encountered during unlocking will be aggregated and
// returned as a single panic.
func (ms Mutexes) RUnlock() {
	if err := doRUnlock(ms); err != nil {
		panic(err)
	}
}

// TryRLock attempts to acquire shared (read) locks on all Mutexes without blocking.
// Returns true if all locks were acquired, false otherwise.
// If a Mutex does not implement the RWMutex interface, an exclusive lock
// will be attempted as a fallback.
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
