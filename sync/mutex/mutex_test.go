package mutex

import (
	"sync"
	"testing"
	"time"

	"darvaza.org/core"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentLocking tests concurrent locking behaviour
func TestConcurrentLocking(t *testing.T) {
	m1 := &sync.Mutex{}
	m2 := &sync.Mutex{}
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			// Try to lock both mutexes
			if TryLock(m1, m2) {
				// Hold the locks briefly
				time.Sleep(1 * time.Millisecond)
				// Release the locks
				Unlock(m1, m2)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify both mutexes are unlocked at the end
	assert.True(t, m1.TryLock(), "m1 should be unlocked after concurrent operations")
	m1.Unlock()

	assert.True(t, m2.TryLock(), "m2 should be unlocked after concurrent operations")
	m2.Unlock()
}

// Helper mutexes that panic on different operations
type panicOnLockMutex struct{}

func (*panicOnLockMutex) Lock()         { panic("intentional panic on lock") }
func (*panicOnLockMutex) Unlock()       {}
func (*panicOnLockMutex) TryLock() bool { return true }

type panicOnTryLockMutex struct{}

func (*panicOnTryLockMutex) Lock()         {}
func (*panicOnTryLockMutex) Unlock()       {}
func (*panicOnTryLockMutex) TryLock() bool { panic("intentional panic on trylock") }

// testMutex is a custom implementation of Mutex for testing unlock order
type testMutex struct {
	id                int
	unlockOrder       chan<- int
	locked            bool
	shouldFailTryLock bool
}

var _ Mutex = (*testMutex)(nil)

func (m *testMutex) Lock() {
	m.locked = true
}

func (m *testMutex) Unlock() {
	if !m.locked {
		panic("unlock of unlocked mutex")
	}
	m.unlockOrder <- m.id
	m.locked = false
}

func (m *testMutex) TryLock() bool {
	if m.shouldFailTryLock {
		return false
	}
	if m.locked {
		return false
	}
	m.locked = true
	return true
}

// TestNilImplementations tests interface implementations for nil values
func TestNilImplementations(t *testing.T) {
	// Test behaviour of Safe* functions with nil mutexes
	nilMu := (*sync.Mutex)(nil)

	ok, err := SafeLock(nilMu)
	assert.False(t, ok, "SafeLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeLock should not return ErrNilMutex for typed nil pointer")

	ok, err = SafeTryLock(nilMu)
	assert.False(t, ok, "SafeTryLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeTryLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeTryLock should not return ErrNilMutex for typed nil pointer")

	err = SafeUnlock(nilMu)
	assert.IsType(t, &core.PanicError{}, err, "SafeUnlock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeUnlock should not return ErrNilMutex for typed nil pointer")

	nilRWMu := (*sync.RWMutex)(nil)

	ok, err = SafeRLock(nilRWMu)
	assert.False(t, ok, "SafeRLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeRLock should not return ErrNilMutex for typed nil pointer")

	ok, err = SafeTryRLock(nilRWMu)
	assert.False(t, ok, "SafeTryRLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeTryRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeTryRLock should not return ErrNilMutex for typed nil pointer")

	err = SafeRUnlock(nilRWMu)
	assert.IsType(t, &core.PanicError{}, err, "SafeRUnlock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeRUnlock should not return ErrNilMutex for typed nil pointer")

	// Verify that the original functions still panic with nil mutexes
	assert.Panics(t, func() {
		Lock[Mutex](nil)
	}, "Lock(nil) should panic")

	assert.Panics(t, func() {
		RLock[Mutex](nil)
	}, "RLock(nil) should panic")

	assert.Panics(t, func() {
		Unlock[Mutex](nil)
	}, "Unlock(nil) should panic")

	assert.Panics(t, func() {
		RUnlock[Mutex](nil)
	}, "RUnlock(nil) should panic")
}
