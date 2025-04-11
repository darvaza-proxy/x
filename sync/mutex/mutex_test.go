package mutex

import (
	"context"
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentLocking verifies the behaviour of simultaneous locking
// operations across multiple goroutines.
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

// Helper mutex types that panic on specific operations

type panicOnLockMutex struct{}

func (*panicOnLockMutex) Lock()         { panic("intentional panic on lock") }
func (*panicOnLockMutex) Unlock()       {}
func (*panicOnLockMutex) TryLock() bool { return true }

type panicOnTryLockMutex struct{}

func (*panicOnTryLockMutex) Lock()         {}
func (*panicOnTryLockMutex) Unlock()       {}
func (*panicOnTryLockMutex) TryLock() bool { panic("intentional panic on trylock") }

// Context-aware mutex implementations for testing
type testMutexContext struct {
	locked bool
}

func (m *testMutexContext) Lock() {
	m.locked = true
}

func (m *testMutexContext) Unlock() {
	m.locked = false
}

func (m *testMutexContext) TryLock() bool {
	if m.locked {
		return false
	}
	m.locked = true
	return true
}

func (m *testMutexContext) LockContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.Lock()
		return nil
	}
}

type testRWMutexContext struct {
	writeLocked bool
	readCount   int
}

func (m *testRWMutexContext) Lock() {
	m.writeLocked = true
}

func (m *testRWMutexContext) Unlock() {
	m.writeLocked = false
}

func (m *testRWMutexContext) TryLock() bool {
	if m.writeLocked || m.readCount > 0 {
		return false
	}
	m.writeLocked = true
	return true
}

func (m *testRWMutexContext) RLock() {
	m.readCount++
}

func (m *testRWMutexContext) RUnlock() {
	m.readCount--
}

func (m *testRWMutexContext) TryRLock() bool {
	if m.writeLocked {
		return false
	}
	m.readCount++
	return true
}

func (m *testRWMutexContext) LockContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.Lock()
		return nil
	}
}

func (m *testRWMutexContext) RLockContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.RLock()
		return nil
	}
}

// Panic mutexes for context-aware testing
type panicOnLockContextMutex struct{}

func (*panicOnLockContextMutex) Lock()         {}
func (*panicOnLockContextMutex) Unlock()       {}
func (*panicOnLockContextMutex) TryLock() bool { return true }
func (*panicOnLockContextMutex) LockContext(_ context.Context) error {
	panic("intentional panic on LockContext")
}

type panicOnRLockContextMutex struct{}

func (*panicOnRLockContextMutex) Lock()                               {}
func (*panicOnRLockContextMutex) Unlock()                             {}
func (*panicOnRLockContextMutex) TryLock() bool                       { return true }
func (*panicOnRLockContextMutex) LockContext(_ context.Context) error { return nil }
func (*panicOnRLockContextMutex) RLock()                              {}
func (*panicOnRLockContextMutex) RUnlock()                            {}
func (*panicOnRLockContextMutex) TryRLock() bool                      { return true }
func (*panicOnRLockContextMutex) RLockContext(_ context.Context) error {
	panic("intentional panic on RLockContext")
}

// testMutex implements the Mutex interface for testing unlock order
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

// TestNilImplementations verifies how interface implementations handle
// nil values, ensuring they either panic or return appropriate errors.
func TestNilImplementations(t *testing.T) {
	// Test Safe* functions with nil mutexes
	nilMu := (*sync.Mutex)(nil)

	ok, err := SafeLock(nilMu)
	assert.False(t, ok, "SafeLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err,
		"SafeLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeLock should not return ErrNilMutex for typed nil pointer")

	ok, err = SafeTryLock(nilMu)
	assert.False(t, ok, "SafeTryLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err,
		"SafeTryLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeTryLock should not return ErrNilMutex for typed nil pointer")

	err = SafeUnlock(nilMu)
	assert.IsType(t, &core.PanicError{}, err,
		"SafeUnlock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeUnlock should not return ErrNilMutex for typed nil pointer")

	nilRWMu := (*sync.RWMutex)(nil)

	ok, err = SafeRLock(nilRWMu)
	assert.False(t, ok, "SafeRLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err,
		"SafeRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeRLock should not return ErrNilMutex for typed nil pointer")

	ok, err = SafeTryRLock(nilRWMu)
	assert.False(t, ok, "SafeTryRLock should return false with nil mutex")
	assert.IsType(t, &core.PanicError{}, err,
		"SafeTryRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeTryRLock should not return ErrNilMutex for typed nil pointer")

	err = SafeRUnlock(nilRWMu)
	assert.IsType(t, &core.PanicError{}, err,
		"SafeRUnlock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex,
		"SafeRUnlock should not return ErrNilMutex for typed nil pointer")

	// Verify that original functions panic with nil mutexes
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
