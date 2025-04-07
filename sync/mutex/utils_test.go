package mutex

import (
	"context"
	"sync"
	"testing"

	"darvaza.org/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafeLock tests the SafeLock function
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
	assert.ErrorIs(t, err, ErrNilMutex, "SafeLock should return ErrNilMutex for nil mutex")

	// Test with custom mutex that panics
	panicMu := &panicOnLockMutex{}
	ok, err = SafeLock(panicMu)
	assert.False(t, ok, "SafeLock should fail when lock panics")
	assert.Error(t, err, "SafeLock should return error when lock panics")
}

// TestSafeTryLock tests the SafeTryLock function
func TestSafeTryLock(t *testing.T) {
	// Test with normal mutex
	mu := &sync.Mutex{}
	ok, err := SafeTryLock(mu)
	assert.True(t, ok, "SafeTryLock should succeed on unlocked mutex")
	assert.Nil(t, err, "SafeTryLock should not return error on success")

	// Mutex should be locked now, so another try should fail
	ok, err = SafeTryLock(mu)
	assert.False(t, ok, "SafeTryLock should fail on already locked mutex")
	assert.Nil(t, err, "SafeTryLock should not return error when lock unavailable")

	// Release the lock
	err = SafeUnlock(mu)
	assert.Nil(t, err, "SafeUnlock should not return error")

	// Test with nil mutex
	var nilMu Mutex
	ok, err = SafeTryLock(nilMu)
	assert.False(t, ok, "SafeTryLock should fail on nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeTryLock should return ErrNilMutex for nil mutex")

	// Test with custom mutex that panics
	panicMu := &panicOnTryLockMutex{}
	ok, err = SafeTryLock(panicMu)
	assert.False(t, ok, "SafeTryLock should fail when TryLock panics")
	assert.Error(t, err, "SafeTryLock should return error when TryLock panics")
}

// TestSafeUnlock tests the SafeUnlock function
func TestSafeUnlock(t *testing.T) {
	// Test with normal mutex
	mu := &sync.Mutex{}
	mu.Lock() // Lock it first

	err := SafeUnlock(mu)
	assert.Nil(t, err, "SafeUnlock should not return error on success")

	// Test with nil mutex
	var nilMu Mutex
	err = SafeUnlock(nilMu)
	assert.ErrorIs(t, err, ErrNilMutex, "SafeUnlock should return ErrNilMutex for nil mutex")

	// Test with custom mutex that panics
	panicMu := &panicOnUnlockMutex{}
	err = SafeUnlock(panicMu)
	assert.Error(t, err, "SafeUnlock should return error when Unlock panics")
}

// TestSafeRLock tests the SafeRLock function
func TestSafeRLock(t *testing.T) {
	// Test with RWMutex - read locks can be acquired multiple times
	rwMu := &sync.RWMutex{}
	ok, err := SafeRLock(rwMu)
	assert.True(t, ok, "SafeRLock should succeed on unlocked RWMutex")
	assert.Nil(t, err, "SafeRLock should not return error on success")

	// Should be able to acquire another read lock
	ok, err = SafeRLock(rwMu)
	assert.True(t, ok, "SafeRLock should succeed when read lock already held")
	assert.Nil(t, err, "SafeRLock should not return error when acquiring second read lock")

	// Release both locks
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")
	err = SafeRUnlock(rwMu)
	assert.Nil(t, err, "SafeRUnlock should not return error")

	// Test with regular Mutex (should behave like a regular Lock)
	regMu := &sync.Mutex{}
	ok, err = SafeRLock(regMu)
	assert.True(t, ok, "SafeRLock should succeed on unlocked regular Mutex")
	assert.Nil(t, err, "SafeRLock should not return error when using regular mutex")

	// Release the lock before trying another operation
	err = SafeRUnlock(regMu)
	assert.Nil(t, err, "SafeRUnlock should not return error on regular Mutex")

	// Verify that on a regular mutex, SafeRLock behaves like a regular lock by:
	// 1. Lock the mutex first
	regMu.Lock()

	// 2. Try to acquire it using SafeTryRLock (which shouldn't block)
	ok, err = SafeTryRLock(regMu)
	assert.False(t, ok, "SafeTryRLock should fail on already locked regular Mutex")
	assert.Nil(t, err, "SafeTryRLock should not return error when lock unavailable")

	// Unlock for cleanup
	regMu.Unlock()

	// Test with nil mutex, but using an interface instead of a typed nil. Otherwise
	// I can't be nil tested.
	var nilMu RWMutex
	ok, err = SafeRLock(nilMu)
	assert.False(t, ok, "SafeRLock should fail on nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeRLock should return ErrNilMutex for nil mutex")

	// Test with typed nil pointer (should cause a panic that gets converted to error)
	typedNilMu := (*sync.RWMutex)(nil)
	ok, err = SafeRLock(typedNilMu)
	assert.False(t, ok, "SafeRLock should fail on typed nil pointer")
	assert.IsType(t, &core.PanicError{}, err, "SafeRLock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeRLock should not return ErrNilMutex for typed nil pointer")
}

// TestSafeTryRLock tests the SafeTryRLock function
func TestSafeTryRLock(t *testing.T) {
	// Test with RWMutex
	rwMu := &sync.RWMutex{}
	ok, err := SafeTryRLock(rwMu)
	assert.True(t, ok, "SafeTryRLock should succeed on unlocked RWMutex")
	assert.Nil(t, err, "SafeTryRLock should not return error on success")

	// Should be able to acquire another read lock
	ok, err = SafeTryRLock(rwMu)
	assert.True(t, ok, "SafeTryRLock should succeed when read lock already held")
	assert.Nil(t, err, "SafeTryRLock should not return error when acquiring second read lock")

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
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeTryRLock should not return ErrNilMutex for typed nil pointer")
}

// TestSafeRUnlock tests the SafeRUnlock function
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
	assert.Nil(t, err, "SafeRUnlock should work with regular mutex (fallback to Unlock)")

	// Test with nil mutex (typed nil will cause panic)
	nilMu := (*sync.RWMutex)(nil)
	err = SafeRUnlock(nilMu)
	assert.IsType(t, &core.PanicError{}, err, "SafeRUnlock should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeRUnlock should not return ErrNilMutex for typed nil pointer")
}

// TestSafeLockContext tests the SafeLockContext function
func TestSafeLockContext(t *testing.T) {
	// Test with normal mutex and context
	ctx := context.Background()
	mu := &testMutexContext{}

	ok, err := SafeLockContext(ctx, mu)
	assert.True(t, ok, "SafeLockContext should succeed with valid context and mutex")
	assert.Nil(t, err, "SafeLockContext should not return error on success")

	// Test with canceled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mu2 := &testMutexContext{}
	ok, err = SafeLockContext(cancelCtx, mu2)
	assert.False(t, ok, "SafeLockContext should fail with canceled context")
	assert.Error(t, err, "SafeLockContext should return error with canceled context")

	// Test with nil context
	nilCtx := context.Context(nil)
	ok, err = SafeLockContext(nilCtx, mu)
	assert.False(t, ok, "SafeLockContext should fail with nil context")
	assert.ErrorIs(t, err, ErrNilContext, "SafeLockContext should return ErrNilContext for nil context")

	// Test with typed nil mutex (will cause panic)
	nilMu := (*testMutexContext)(nil)
	ok, err = SafeLockContext(ctx, nilMu)
	assert.False(t, ok, "SafeLockContext should fail with typed nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeLockContext should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeLockContext should not return ErrNilMutex for typed nil pointer")

	// Test with mutex that panics
	panicMu := &panicOnLockContextMutex{}
	ok, err = SafeLockContext(ctx, panicMu)
	assert.False(t, ok, "SafeLockContext should fail when LockContext panics")
	assert.Error(t, err, "SafeLockContext should return error when LockContext panics")
}

// TestSafeRLockContext tests the SafeRLockContext function
func TestSafeRLockContext(t *testing.T) {
	// Test with RWMutexContext
	ctx := context.Background()
	rwMu := &testRWMutexContext{}

	ok, err := SafeRLockContext(ctx, rwMu)
	assert.True(t, ok, "SafeRLockContext should succeed with valid context and rwmutex")
	assert.Nil(t, err, "SafeRLockContext should not return error on success")

	// Should be able to acquire another read lock
	ok, err = SafeRLockContext(ctx, rwMu)
	assert.True(t, ok, "SafeRLockContext should succeed when read lock already held")
	assert.Nil(t, err, "SafeRLockContext should not return error when acquiring second read lock")

	// Test with regular MutexContext (fallback to regular lock)
	regMu := &testMutexContext{}
	ok, err = SafeRLockContext(ctx, regMu)
	assert.True(t, ok, "SafeRLockContext should succeed with regular mutex")
	assert.Nil(t, err, "SafeRLockContext should not return error with regular mutex")

	// Test with nil context
	nilCtx := context.Context(nil)
	ok, err = SafeRLockContext(nilCtx, rwMu)
	assert.False(t, ok, "SafeRLockContext should fail with nil context")
	assert.ErrorIs(t, err, ErrNilContext, "SafeRLockContext should return ErrNilContext for nil context")

	// Test with typed nil mutex (will cause panic)
	nilMu := (*testRWMutexContext)(nil)
	ok, err = SafeRLockContext(ctx, nilMu)
	assert.False(t, ok, "SafeRLockContext should fail with typed nil mutex")
	assert.IsType(t, &core.PanicError{}, err, "SafeRLockContext should return core.PanicError for typed nil pointer")
	assert.NotErrorIs(t, err, ErrNilMutex, "SafeRLockContext should not return ErrNilMutex for typed nil pointer")

	// Test with mutex that panics
	panicMu := &panicOnRLockContextMutex{}
	ok, err = SafeRLockContext(ctx, panicMu)
	assert.False(t, ok, "SafeRLockContext should fail when RLockContext panics")
	assert.Error(t, err, "SafeRLockContext should return error when RLockContext panics")
}

// TestNewSafeLockContext tests the NewSafeLockContext function
func TestNewSafeLockContext(t *testing.T) {
	ctx := context.Background()
	lockFn := NewSafeLockContext[MutexContext](ctx)
	require.NotNil(t, lockFn, "NewSafeLockContext should return a non-nil function")

	mu := &testMutexContext{}
	ok, err := lockFn(mu)
	assert.True(t, ok, "Function from NewSafeLockContext should succeed with valid mutex")
	assert.Nil(t, err, "Function from NewSafeLockContext should not return error on success")
}

// TestNewSafeRLockContext tests the NewSafeRLockContext function
func TestNewSafeRLockContext(t *testing.T) {
	ctx := context.Background()
	rlockFn := NewSafeRLockContext[RWMutexContext](ctx)
	require.NotNil(t, rlockFn, "NewSafeRLockContext should return a non-nil function")

	rwMu := &testRWMutexContext{}
	ok, err := rlockFn(rwMu)
	assert.True(t, ok, "Function from NewSafeRLockContext should succeed with valid rwmutex")
	assert.Nil(t, err, "Function from NewSafeRLockContext should not return error on success")
}

// TestReverseUnlock tests the ReverseUnlock function
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
	assert.Equal(t, []int{3, 2, 1}, unlockOrder, "ReverseUnlock should unlock in reverse order")

	// Test with nil unlock function
	err = ReverseUnlock(nil, m1, m2, m3)
	assert.Error(t, err, "ReverseUnlock should return error with nil unlock function")

	// Test with unlock function that returns errors
	errorFn := func(_ Mutex) error {
		return ErrNilMutex // Just using this as a convenient error
	}

	err = ReverseUnlock[Mutex](errorFn, m1, m2, m3)
	assert.Error(t, err, "ReverseUnlock should return error when unlock function returns errors")
}

// TestErrorHandling tests error cases when locking/unlocking mutexes
func TestErrorHandling(t *testing.T) {
	// Test SafeLock with nil mutex
	ok, err := SafeLock[Mutex](nil)
	assert.False(t, ok, "SafeLock should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeLock should return ErrNilMutex with nil mutex")

	// Test SafeUnlock with nil mutex
	err = SafeUnlock[Mutex](nil)
	assert.ErrorIs(t, err, ErrNilMutex, "SafeUnlock should return ErrNilMutex with nil mutex")

	// Test SafeTryLock with nil mutex
	ok, err = SafeTryLock[Mutex](nil)
	assert.False(t, ok, "SafeTryLock should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeTryLock should return ErrNilMutex with nil mutex")

	// Test SafeRLock with nil mutex
	ok, err = SafeRLock[Mutex](nil)
	assert.False(t, ok, "SafeRLock should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeRLock should return ErrNilMutex with nil mutex")

	// Test SafeTryRLock with nil mutex
	ok, err = SafeTryRLock[RWMutex](nil)
	assert.False(t, ok, "SafeTryRLock should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeTryRLock should return ErrNilMutex with nil mutex")

	// Test SafeRUnlock with nil mutex
	err = SafeRUnlock[RWMutex](nil)
	assert.ErrorIs(t, err, ErrNilMutex, "SafeRUnlock should return ErrNilMutex with nil mutex")

	// Test context functions with nil context and mutex
	ctx := context.Background()
	var nilMu MutexContext
	var nilCtx context.Context

	ok, err = SafeLockContext(nilCtx, &testMutexContext{})
	assert.False(t, ok, "SafeLockContext should return false with nil context")
	assert.ErrorIs(t, err, ErrNilContext, "SafeLockContext should return ErrNilContext")

	ok, err = SafeLockContext(ctx, nilMu)
	assert.False(t, ok, "SafeLockContext should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeLockContext should return ErrNilMutex")

	ok, err = SafeRLockContext(nilCtx, &testRWMutexContext{})
	assert.False(t, ok, "SafeRLockContext should return false with nil context")
	assert.ErrorIs(t, err, ErrNilContext, "SafeRLockContext should return ErrNilContext")

	ok, err = SafeRLockContext(ctx, nilMu)
	assert.False(t, ok, "SafeRLockContext should return false with nil mutex")
	assert.ErrorIs(t, err, ErrNilMutex, "SafeRLockContext should return ErrNilMutex")
}

// TestOriginalReverseUnlock tests that locks are released in reverse order on failure
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
	assert.False(t, TryLock(customM1, customM2, customM3), "TryLock should fail when one mutex is configured to fail")

	// Check the unlock order (should be 2, 1 - reverse order)
	// We should have exactly 2 unlocks (not 3, since m3's lock failed)
	assert.Equal(t, 2, len(unlockOrder), "Expected 2 unlocks")

	order1 := <-unlockOrder
	assert.Equal(t, 2, order1, "Expected first unlock to be mutex 2")

	order2 := <-unlockOrder
	assert.Equal(t, 1, order2, "Expected second unlock to be mutex 1")
}

// TestSafeReverseUnlock tests the behaviour of ReverseUnlock with various error conditions
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
	assert.Contains(t, err.Error(), "unlock function is nil", "ReverseUnlock should return specific error message")

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
		return ErrNilMutex // Just reusing a convenient error
	}

	err = ReverseUnlock[Mutex](errFn, m1, m2)
	assert.Error(t, err, "ReverseUnlock should return error when unlock function returns errors")
}

// panicOnUnlockMutex panics when Unlock is called
type panicOnUnlockMutex struct{}

func (*panicOnUnlockMutex) Lock()         {}
func (*panicOnUnlockMutex) Unlock()       { panic("intentional panic on unlock") }
func (*panicOnUnlockMutex) TryLock() bool { return true }
