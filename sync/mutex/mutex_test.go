package mutex_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/sync/mutex"
)

// ----- test doubles -----

// panicOnLockMutex panics when Lock is called, exercising the panic-catching
// behaviour of SafeLock.
type panicOnLockMutex struct{}

func (*panicOnLockMutex) Lock()         { panic("intentional panic on lock") }
func (*panicOnLockMutex) Unlock()       {}
func (*panicOnLockMutex) TryLock() bool { return true }

// panicOnTryLockMutex panics when TryLock is called.
type panicOnTryLockMutex struct{}

func (*panicOnTryLockMutex) Lock()         {}
func (*panicOnTryLockMutex) Unlock()       {}
func (*panicOnTryLockMutex) TryLock() bool { panic("intentional panic on trylock") }

// panicOnUnlockMutex panics when Unlock is called.
type panicOnUnlockMutex struct{}

func (*panicOnUnlockMutex) Lock()         {}
func (*panicOnUnlockMutex) Unlock()       { panic("intentional panic on unlock") }
func (*panicOnUnlockMutex) TryLock() bool { return true }

// testMutexContext is a minimal MutexContext used to exercise the context
// helpers without real blocking.
type testMutexContext struct {
	locked bool
}

func (m *testMutexContext) Lock()   { m.locked = true }
func (m *testMutexContext) Unlock() { m.locked = false }

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

// testRWMutexContext is a minimal RWMutexContext used to exercise the
// read-side context helpers.
type testRWMutexContext struct {
	writeLocked bool
	readCount   int
}

func (m *testRWMutexContext) Lock()    { m.writeLocked = true }
func (m *testRWMutexContext) Unlock()  { m.writeLocked = false }
func (m *testRWMutexContext) RLock()   { m.readCount++ }
func (m *testRWMutexContext) RUnlock() { m.readCount-- }

func (m *testRWMutexContext) TryLock() bool {
	if m.writeLocked || m.readCount > 0 {
		return false
	}
	m.writeLocked = true
	return true
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

// panicOnLockContextMutex panics when LockContext is called.
type panicOnLockContextMutex struct{}

func (*panicOnLockContextMutex) Lock()         {}
func (*panicOnLockContextMutex) Unlock()       {}
func (*panicOnLockContextMutex) TryLock() bool { return true }
func (*panicOnLockContextMutex) LockContext(context.Context) error {
	panic("intentional panic on LockContext")
}

// panicOnRLockContextMutex panics when RLockContext is called.
type panicOnRLockContextMutex struct{}

func (*panicOnRLockContextMutex) Lock()                             {}
func (*panicOnRLockContextMutex) Unlock()                           {}
func (*panicOnRLockContextMutex) TryLock() bool                     { return true }
func (*panicOnRLockContextMutex) LockContext(context.Context) error { return nil }
func (*panicOnRLockContextMutex) RLock()                            {}
func (*panicOnRLockContextMutex) RUnlock()                          {}
func (*panicOnRLockContextMutex) TryRLock() bool                    { return true }
func (*panicOnRLockContextMutex) RLockContext(context.Context) error {
	panic("intentional panic on RLockContext")
}

// testMutex records the order in which it is unlocked, so reverse-order
// release can be asserted.
type testMutex struct {
	unlockOrder       chan<- int
	id                int
	locked            bool
	shouldFailTryLock bool
}

var _ mutex.Mutex = (*testMutex)(nil)

func (m *testMutex) Lock() { m.locked = true }

func (m *testMutex) Unlock() {
	if !m.locked {
		panic("unlock of unlocked mutex")
	}
	m.unlockOrder <- m.id
	m.locked = false
}

func (m *testMutex) TryLock() bool {
	if m.shouldFailTryLock || m.locked {
		return false
	}
	m.locked = true
	return true
}

// ----- variadic multi-lock API -----

func TestConcurrentLocking(t *testing.T) {
	m1 := &sync.Mutex{}
	m2 := &sync.Mutex{}
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			if mutex.TryLock(m1, m2) {
				time.Sleep(time.Millisecond)
				mutex.Unlock(m1, m2)
			}
		}()
	}
	wg.Wait()

	core.AssertTrue(t, m1.TryLock(), "m1 free after concurrency")
	m1.Unlock()
	core.AssertTrue(t, m2.TryLock(), "m2 free after concurrency")
	m2.Unlock()
}

func TestMultiLockAcquire(t *testing.T) {
	t.Run("exclusive lock and unlock", runTestMultiExclusive)
	t.Run("try lock and unlock", runTestMultiTryLock)
	t.Run("read lock and unlock", runTestMultiReadLock)
	t.Run("try read lock and unlock", runTestMultiTryReadLock)
}

func runTestMultiExclusive(t *testing.T) {
	t.Helper()
	m1, m2 := &sync.Mutex{}, &sync.Mutex{}
	mutex.Lock(m1, m2)
	mutex.Unlock(m1, m2)
	core.AssertTrue(t, m1.TryLock(), "m1 free after unlock")
	m1.Unlock()
}

func runTestMultiTryLock(t *testing.T) {
	t.Helper()
	m1, m2 := &sync.Mutex{}, &sync.Mutex{}
	core.AssertTrue(t, mutex.TryLock(m1, m2), "TryLock both")
	mutex.Unlock(m1, m2)
}

func runTestMultiReadLock(t *testing.T) {
	t.Helper()
	rw1, rw2 := &sync.RWMutex{}, &sync.RWMutex{}
	mutex.RLock(rw1, rw2)
	mutex.RUnlock(rw1, rw2)
	core.AssertTrue(t, rw1.TryLock(), "rw1 free after RUnlock")
	rw1.Unlock()
}

func runTestMultiTryReadLock(t *testing.T) {
	t.Helper()
	rw1, rw2 := &sync.RWMutex{}, &sync.RWMutex{}
	core.AssertTrue(t, mutex.TryRLock(rw1, rw2), "TryRLock both")
	mutex.RUnlock(rw1, rw2)
}

// TestMultiLockReverseUnlock verifies that a partial acquisition failure
// releases the already-held locks in reverse order.
func TestMultiLockReverseUnlock(t *testing.T) {
	unlockOrder := make(chan int, 3)
	m1 := &testMutex{id: 1, unlockOrder: unlockOrder}
	m2 := &testMutex{id: 2, unlockOrder: unlockOrder}
	m3 := &testMutex{id: 3, unlockOrder: unlockOrder, shouldFailTryLock: true}

	core.AssertFalse(t, mutex.TryLock[mutex.Mutex](m1, m2, m3),
		"TryLock fails when m3 refuses")

	core.AssertMustEqual(t, 2, len(unlockOrder), "unlock count")
	core.AssertEqual(t, 2, <-unlockOrder, "first reverse unlock")
	core.AssertEqual(t, 1, <-unlockOrder, "second reverse unlock")
}

// TestMultiLockCleanupError covers doLockLoop's branch where the reverse
// unlock triggered by a mid-acquisition failure itself errors: the first
// mutex acquires but panics on Unlock, the second refuses acquisition, so
// the rollback failure is aggregated and surfaced as a panic.
func TestMultiLockCleanupError(t *testing.T) {
	held := &panicOnUnlockMutex{}
	refuses := &testMutex{shouldFailTryLock: true}
	core.AssertPanic(t, func() {
		mutex.TryLock[mutex.Mutex](held, refuses)
	}, nil, "rollback failure surfaces as panic")
}

// TestMultiLockEmpty verifies every variadic entry point is a no-op when
// given no mutexes.
func TestMultiLockEmpty(t *testing.T) {
	core.AssertNoPanic(t, func() { mutex.Lock[mutex.Mutex]() }, "Lock")
	core.AssertNoPanic(t, func() { mutex.RLock[mutex.Mutex]() }, "RLock")
	core.AssertNoPanic(t, func() { mutex.Unlock[mutex.Mutex]() }, "Unlock")
	core.AssertNoPanic(t, func() { mutex.RUnlock[mutex.Mutex]() }, "RUnlock")
	core.AssertTrue(t, mutex.TryLock[mutex.Mutex](), "TryLock empty")
	core.AssertTrue(t, mutex.TryRLock[mutex.Mutex](), "TryRLock empty")
}

// multiLockPanicTestCase verifies the variadic entry points panic when handed
// a nil mutex. op selects the entry point under the shared assertion path.
type multiLockPanicTestCase struct {
	op   func()
	name string
}

func newMultiLockPanicTestCase(name string, op func()) multiLockPanicTestCase {
	return multiLockPanicTestCase{name: name, op: op}
}

func (tc multiLockPanicTestCase) Name() string { return tc.name }

func (tc multiLockPanicTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertPanic(t, tc.op, nil, "panic")
}

var _ core.TestCase = multiLockPanicTestCase{}

func opLockNil()     { mutex.Lock[mutex.Mutex](nil) }
func opRLockNil()    { mutex.RLock[mutex.Mutex](nil) }
func opUnlockNil()   { mutex.Unlock[mutex.Mutex](nil) }
func opRUnlockNil()  { mutex.RUnlock[mutex.Mutex](nil) }
func opTryLockNil()  { mutex.TryLock[mutex.Mutex](nil) }
func opTryRLockNil() { mutex.TryRLock[mutex.Mutex](nil) }

func TestMultiLockNilPanic(t *testing.T) {
	core.RunTestCases(t, []multiLockPanicTestCase{
		newMultiLockPanicTestCase("Lock nil", opLockNil),
		newMultiLockPanicTestCase("RLock nil", opRLockNil),
		newMultiLockPanicTestCase("Unlock nil", opUnlockNil),
		newMultiLockPanicTestCase("RUnlock nil", opRUnlockNil),
		newMultiLockPanicTestCase("TryLock nil", opTryLockNil),
		newMultiLockPanicTestCase("TryRLock nil", opTryRLockNil),
	})
}
