package mutex_test

import (
	"context"
	"sync"
	"testing"

	"darvaza.org/x/sync/mutex"
)

// newLocks returns n new values of the lock type T.
func newLocks[T any](n int) []*T {
	locks := make([]*T, n)
	for i := range locks {
		locks[i] = new(T)
	}
	return locks
}

// BenchmarkLock measures taking and releasing sync.Mutex values through
// Lock and Unlock, one and three at a time.
func BenchmarkLock(b *testing.B) {
	b.Run("One", func(b *testing.B) { runBenchmarkLock(b, newLocks[sync.Mutex](1)) })
	b.Run("Three", func(b *testing.B) { runBenchmarkLock(b, newLocks[sync.Mutex](3)) })
}

func runBenchmarkLock(b *testing.B, locks []*sync.Mutex) {
	for b.Loop() {
		mutex.Lock(locks...)
		mutex.Unlock(locks...)
	}
}

// BenchmarkTryLock measures taking free sync.Mutex values through
// TryLock, one and three at a time, and releasing them through Unlock.
func BenchmarkTryLock(b *testing.B) {
	b.Run("One", func(b *testing.B) { runBenchmarkTryLock(b, newLocks[sync.Mutex](1)) })
	b.Run("Three", func(b *testing.B) { runBenchmarkTryLock(b, newLocks[sync.Mutex](3)) })
}

func runBenchmarkTryLock(b *testing.B, locks []*sync.Mutex) {
	for b.Loop() {
		if !mutex.TryLock(locks...) {
			b.Fatal("TryLock failed on free mutexes")
		}
		mutex.Unlock(locks...)
	}
}

// BenchmarkRLock measures taking and releasing sync.RWMutex values for
// reading through RLock and RUnlock, one and three at a time.
func BenchmarkRLock(b *testing.B) {
	b.Run("One", func(b *testing.B) { runBenchmarkRLock(b, newLocks[sync.RWMutex](1)) })
	b.Run("Three", func(b *testing.B) { runBenchmarkRLock(b, newLocks[sync.RWMutex](3)) })
}

func runBenchmarkRLock(b *testing.B, locks []*sync.RWMutex) {
	for b.Loop() {
		mutex.RLock(locks...)
		mutex.RUnlock(locks...)
	}
}

// BenchmarkTryRLock measures taking free sync.RWMutex values for reading
// through TryRLock, one and three at a time, and releasing them through
// RUnlock.
func BenchmarkTryRLock(b *testing.B) {
	b.Run("One", func(b *testing.B) { runBenchmarkTryRLock(b, newLocks[sync.RWMutex](1)) })
	b.Run("Three", func(b *testing.B) { runBenchmarkTryRLock(b, newLocks[sync.RWMutex](3)) })
}

func runBenchmarkTryRLock(b *testing.B, locks []*sync.RWMutex) {
	for b.Loop() {
		if !mutex.TryRLock(locks...) {
			b.Fatal("TryRLock failed on free mutexes")
		}
		mutex.RUnlock(locks...)
	}
}

// BenchmarkSafeLockContext measures SafeLockContext and SafeRLockContext
// taking a free lock, released through SafeUnlock and SafeRUnlock. The
// locks are test doubles that set a flag rather than lock, so the figures
// are mostly the helpers' own cost.
func BenchmarkSafeLockContext(b *testing.B) {
	b.Run("Lock", runBenchmarkSafeLockContext)
	b.Run("RLock", runBenchmarkSafeRLockContext)
}

func runBenchmarkSafeLockContext(b *testing.B) {
	ctx := context.Background()
	mu := new(testMutexContext)

	for b.Loop() {
		if _, err := mutex.SafeLockContext(ctx, mu); err != nil {
			b.Fatalf("SafeLockContext returned error: %v", err)
		}
		if err := mutex.SafeUnlock(mu); err != nil {
			b.Fatalf("SafeUnlock returned error: %v", err)
		}
	}
}

func runBenchmarkSafeRLockContext(b *testing.B) {
	ctx := context.Background()
	mu := new(testRWMutexContext)

	for b.Loop() {
		if _, err := mutex.SafeRLockContext(ctx, mu); err != nil {
			b.Fatalf("SafeRLockContext returned error: %v", err)
		}
		if err := mutex.SafeRUnlock(mu); err != nil {
			b.Fatalf("SafeRUnlock returned error: %v", err)
		}
	}
}

// BenchmarkReverseUnlock measures ReverseUnlock releasing three
// sync.Mutex values through SafeUnlock.
func BenchmarkReverseUnlock(b *testing.B) {
	locks := newLocks[sync.Mutex](3)

	for b.Loop() {
		for _, mu := range locks {
			mu.Lock()
		}
		if err := mutex.ReverseUnlock(mutex.SafeUnlock, locks...); err != nil {
			b.Fatalf("ReverseUnlock returned error: %v", err)
		}
	}
}
