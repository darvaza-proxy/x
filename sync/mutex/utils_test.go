package mutex

import (
	"sync"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"

	"github.com/stretchr/testify/assert"
)

// TestSafeLock verifies the SafeLock function behaves correctly
func TestSafeLock(t *testing.T) {
	// Test with normal mutex
	mu := &sync.Mutex{}
	ok, err := SafeLock(mu)
	assert.True(t, ok, "SafeLock should succeed on unlocked mutex")
	assert.Nil(t, err, "SafeLock should not return error on success")

	// Release the lock
	err = SafeUnlock(mu)
	assert.Nil(t, err, "SafeUnlock should not return error")

	// Test with nil mutex
	ok, err = SafeLock[Mutex](nil)
	assert.False(t, ok, "SafeLock should fail on nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeLock should return ErrNilMutex")

	// Test with mutex that panics on Lock
	panicMu := &panicOnLockMutex{}
	ok, err = SafeLock(panicMu)
	assert.False(t, ok, "SafeLock should fail when lock panics")
	assert.Error(t, err, "SafeLock should return error when lock panics")
}

// TestSafeTryLock verifies the SafeTryLock function behaves correctly
func TestSafeTryLock(t *testing.T) {
	// Test with normal mutex
	mu := &sync.Mutex{}
	ok, err := SafeTryLock(mu)
	assert.True(t, ok, "SafeTryLock should succeed on unlocked mutex")
	assert.Nil(t, err, "SafeTryLock should not return error on success")

	// Mutex should be locked now, so another try should fail
	ok, err = SafeTryLock(mu)
	assert.False(t, ok, "SafeTryLock should fail on already locked mutex")
	assert.Nil(t, err, "SafeTryLock should not return error when unavailable")

	// Release the lock
	err = SafeUnlock(mu)
	assert.Nil(t, err, "SafeUnlock should not return error")

	// Test with nil mutex
	var nilMu Mutex
	ok, err = SafeTryLock(nilMu)
	assert.False(t, ok, "SafeTryLock should fail on nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeTryLock should return ErrNilMutex")

	// Test with mutex that panics on TryLock
	panicMu := &panicOnTryLockMutex{}
	ok, err = SafeTryLock(panicMu)
	assert.False(t, ok, "SafeTryLock should fail when TryLock panics")
	assert.Error(t, err, "SafeTryLock should return error when TryLock panics")
}

// TestSafeUnlock verifies the SafeUnlock function behaves correctly
func TestSafeUnlock(t *testing.T) {
	// Test with normal mutex
	mu := &sync.Mutex{}
	mu.Lock() // Lock it first

	err := SafeUnlock(mu)
	assert.Nil(t, err, "SafeUnlock should not return error on success")

	// Test with nil mutex
	var nilMu Mutex
	err = SafeUnlock(nilMu)
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeUnlock should return ErrNilMutex")
	// Test with mutex that panics on Unlock
	panicMu := &panicOnUnlockMutex{}
	err = SafeUnlock(panicMu)
	assert.Error(t, err, "SafeUnlock should return error when Unlock panics")
}

// TestSafeRLock verifies the SafeRLock function behaves correctly
func TestSafeRLock(t *testing.T) {
	// Test with RWMutex - read locks can be acquired multiple times
	rwMu := &sync.RWMutex{}
	ok, err := SafeRLock(rwMu)
	assert.True(t, ok, "SafeRLock should succeed on unlocked RWMutex")
	assert.Nil(t, err, "SafeRLock should not return error on success")

	// Should be able to acquire another read lock
	ok, err = SafeRLock(rwMu)
	assert.True(t, ok, "SafeRLock should succeed when read lock already held")
	assert.Nil(t, err, "SafeRLock should not return error for second read lock")

	// Release both locks
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")

	// Test with regular Mutex (should behave like a regular Lock)
	regMu := &sync.Mutex{}
	ok, err = SafeRLock(regMu)
	assert.True(t, ok, "SafeRLock should succeed on unlocked regular Mutex")
	assert.Nil(t, err, "SafeRLock should not return error with regular mutex")

	// Release the lock before trying another operation
	err = SafeRUnlock(regMu)
	assert.Nil(t, err, "SafeRUnlock should not return error on regular Mutex")

	// Verify SafeRLock behaves like regular lock for normal mutex:
	// 1. Lock the mutex first
	regMu.Lock()

	// 2. Try to acquire it using SafeTryRLock (should not block)
	ok, err = SafeTryRLock(regMu)
	assert.False(t, ok, "SafeTryRLock should fail on already locked mutex")
	assert.Nil(t, err, "SafeTryRLock should not return error when unavailable")

	// Unlock for cleanup
	regMu.Unlock()

	// Test with nil mutex through interface
	var nilMu RWMutex
	ok, err = SafeRLock(nilMu)
	assert.False(t, ok, "SafeRLock should fail on nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeRLock should return ErrNilMutex")
	// Test with typed nil pointer (should convert panic to error)
	typedNilMu := (*sync.RWMutex)(nil)
	ok, err = SafeRLock(typedNilMu)
	assert.False(t, ok, "SafeRLock should fail on typed nil pointer")

	assert.IsType(t, &core.PanicError{}, err, "SafeRLock should return PanicError")
	assert.NotErrorIs(t, err, errors.ErrNilMutex, "Error should not be ErrNilMutex")
}

// TestSafeTryRLock verifies the SafeTryRLock function behaves correctly
func TestSafeTryRLock(t *testing.T) {
	// Test with RWMutex
	rwMu := &sync.RWMutex{}
	ok, err := SafeTryRLock(rwMu)
	assert.True(t, ok, "SafeTryRLock should succeed on unlocked RWMutex")
	assert.Nil(t, err, "SafeTryRLock should not return error on success")

	// Should be able to acquire another read lock
	ok, err = SafeTryRLock(rwMu)
	assert.True(t, ok, "SafeTryRLock should succeed with read lock already held")
	assert.Nil(t, err, "SafeTryRLock should not return error for second read lock")

	// Write lock should fail (read locks are held)
	rwMu2 := rwMu // Just to have a different variable
	ok, _ = SafeTryLock(rwMu2)
	assert.False(t, ok, "SafeLock should fail when read locks are held")

	// Release both read locks
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")

	// Test caught panic with typed nil pointer
	typedNilMu := (*sync.RWMutex)(nil)
	ok, err = SafeTryRLock(typedNilMu)
	assert.False(t, ok, "SafeTryRLock should fail on typed nil pointer")
	assert.IsType(t, &core.PanicError{}, err, "SafeTryRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, errors.ErrNilMutex, "SafeTryRLock should not return ErrNilMutex for typed nil pointer")
}

// TestSafeRUnlock verifies the SafeRUnlock function behaves correctly
func TestSafeRUnlock(t *testing.T) {
	// Test with RWMutex
	rwMu := &sync.RWMutex{}
	rwMu.RLock() // Lock it first

	err := SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error on success")

	// Test with regular Mutex
	regMu := &sync.Mutex{}
	regMu.Lock() // Lock it first

	err = SafeRUnlock(regMu)
	assert.Nil(t, err, "SafeRUnlock should work with regular mutex (Unlock)")

	// Test with typed nil pointer
	nilMu := (*sync.RWMutex)(nil)
	err = SafeRUnlock(nilMu)
	assert.IsType(t, &core.PanicError{}, err, "SafeRUnlock should return PanicError")
	assert.NotErrorIs(t, err, errors.ErrNilMutex, "Error should not be ErrNilMutex")
}

// TestReverseUnlock verifies the ReverseUnlock function works correctly
func TestReverseUnlock(t *testing.T) {
	// Create mutexes to test with
	m1 := &sync.Mutex{}
	m2 := &sync.Mutex{}
	m3 := &sync.Mutex{}

	// Lock all three mutexes
	m1.Lock()
	m2.Lock()
	m3.Lock()

	// Record unlock order
	unlockOrder := make([]int, 0, 3)
	unlockFn := func(mu Mutex) error {
		switch mu {
		case m1:
			unlockOrder = append(unlockOrder, 1)
		case m2:
			unlockOrder = append(unlockOrder, 2)
		case m3:
			unlockOrder = append(unlockOrder, 3)
		}
		mu.Unlock()
		return nil
	}

	// Call ReverseUnlock and check result
	err := ReverseUnlock[Mutex](unlockFn, m1, m2, m3)
	assert.Nil(t, err, "ReverseUnlock should not return error with valid mutexes")

	// Verify the unlock order is reversed
	assert.Equal(t, []int{3, 2, 1}, unlockOrder, "Should unlock in reverse order")

	// Test with nil unlock function
	err = ReverseUnlock(nil, m1, m2, m3)
	assert.Error(t, err, "ReverseUnlock should return error with nil function")

	// Test with unlock function that returns errors
	errorFn := func(_ Mutex) error {
		return errors.ErrNilMutex // Using this as a convenient error
	}

	err = ReverseUnlock[Mutex](errorFn, m1, m2, m3)
	assert.Error(t, err, "ReverseUnlock should return error when function fails")
}

// TestErrorHandling verifies error cases when locking/unlocking mutexes
func TestErrorHandling(t *testing.T) {
	// Test nil mutex with various functions
	ok, err := SafeLock[Mutex](nil)
	assert.False(t, ok, "SafeLock should return false with nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeLock should return ErrNilMutex")

	err = SafeUnlock[Mutex](nil)
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeUnlock should return ErrNilMutex")

	ok, err = SafeTryLock[Mutex](nil)
	assert.False(t, ok, "SafeTryLock should return false with nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeTryLock should return ErrNilMutex")

	// Test SafeRLock with nil mutex
	ok, err = SafeRLock[Mutex](nil)
	assert.False(t, ok, "SafeRLock should return false with nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeRLock should return ErrNilMutex with nil mutex")

	// Test SafeTryRLock with nil mutex
	ok, err = SafeTryRLock[RWMutex](nil)
	assert.False(t, ok, "SafeTryRLock should return false with nil mutex")
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeTryRLock should return ErrNilMutex with nil mutex")

	// Test SafeRUnlock with nil mutex
	err = SafeRUnlock[RWMutex](nil)
	assert.ErrorIs(t, err, errors.ErrNilMutex, "SafeRUnlock should return ErrNilMutex with nil mutex")
}

// TestOriginalReverseUnlock verifies locks are released in reverse order
// on failure
func TestOriginalReverseUnlock(t *testing.T) {
	// Create a channel to track unlock order
	unlockOrder := make(chan int, 3)

	// Create custom mutex implementations to track unlock order
	customM1 := &testMutex{id: 1, unlockOrder: unlockOrder}
	customM2 := &testMutex{id: 2, unlockOrder: unlockOrder}
	customM3 := &testMutex{id: 3, unlockOrder: unlockOrder}

	// Set up m3 to fail when TryLock is called
	customM3.shouldFailTryLock = true

	// Try to lock all three (should fail on m3)
	assert.False(t, TryLock(customM1, customM2, customM3),
		"TryLock should fail when one mutex is configured to fail")

	// Check the unlock order (should be 2, 1 - reverse order)
	// We should have exactly 2 unlocks (not 3, since m3's lock failed)
	assert.Equal(t, 2, len(unlockOrder), "Expected 2 unlocks")

	order1 := <-unlockOrder
	assert.Equal(t, 2, order1, "Expected first unlock to be mutex 2")

	order2 := <-unlockOrder
	assert.Equal(t, 1, order2, "Expected second unlock to be mutex 1")
}

// TestSafeReverseUnlock verifies the behaviour of ReverseUnlock with
// various error conditions
func TestSafeReverseUnlock(t *testing.T) {
	m1 := &sync.Mutex{}
	m2 := &sync.Mutex{}

	// Lock mutexes before attempting to unlock them
	m1.Lock()
	m2.Lock()

	// Test with valid mutexes
	err := ReverseUnlock(SafeUnlock, m1, m2)
	assert.Nil(t, err, "ReverseUnlock should not return error with valid mutexes")

	// Test with nil function - no locking needed as function is nil
	err = ReverseUnlock(nil, m1, m2)
	assert.Error(t, err, "ReverseUnlock should return error with nil function")
	assert.Contains(t, err.Error(), "unlock function is nil",
		"ReverseUnlock should return specific error message")

	// LOCK THE MUTEXES AGAIN before the third call
	m1.Lock()
	m2.Lock()

	// Test with a mix of valid and nil mutexes
	err = ReverseUnlock(SafeUnlock, m1, nil, m2)
	assert.Error(t, err, "ReverseUnlock should return error with nil mutex")

	// LOCK THE MUTEXES AGAIN before the fourth call
	m1.Lock()
	m2.Lock()

	// Test with a function that always returns errors
	errFn := func(_ Mutex) error {
		return errors.ErrNilMutex // Just reusing a convenient error
	}

	err = ReverseUnlock[Mutex](errFn, m1, m2)
	assert.Error(t, err,
		"ReverseUnlock should return error when unlock function returns errors")
}

// panicOnUnlockMutex panics when Unlock is called
type panicOnUnlockMutex struct{}

func (*panicOnUnlockMutex) Lock()         {}
func (*panicOnUnlockMutex) Unlock()       { panic("intentional panic on unlock") }
func (*panicOnUnlockMutex) TryLock() bool { return true }
