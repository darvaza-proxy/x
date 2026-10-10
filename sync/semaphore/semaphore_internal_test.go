package semaphore

import (
	"context"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/sync/errors"
)

// TestSemaphore_Init exercises the unexported lazyInit and the channel
// fields it populates. It lives in the white-box test package because
// neither the function nor the global/readers channels are part of the
// public surface.
func TestSemaphore_Init(t *testing.T) {
	t.Run("nil receiver", runTestInitNilReceiver)
	t.Run("initialise channels", runTestInitChannels)
	t.Run("idempotent", runTestInitIdempotent)
}

func runTestInitNilReceiver(t *testing.T) {
	t.Helper()
	var s *Semaphore
	core.AssertErrorIs(t, s.lazyInit(), core.ErrNilReceiver,
		"lazyInit on nil receiver")
}

func runTestInitChannels(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")
	core.AssertNotNil(t, s.global, "global channel")
	core.AssertNotNil(t, s.readers, "readers channel")
}

func runTestInitIdempotent(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "first lazyInit")

	global, readers := s.global, s.readers

	core.AssertMustNoError(t, s.lazyInit(), "second lazyInit")
	core.AssertSame(t, global, s.global, "global channel unchanged")
	core.AssertSame(t, readers, s.readers, "readers channel unchanged")
}

// BenchmarkLazyInit measures lazyInit on a semaphore that is already
// initialised, the state every call after the first finds it in. The lock
// and unlock methods run it before touching the channels, so this is the
// entry cost each of them pays.
func BenchmarkLazyInit(b *testing.B) {
	b.Run("serial", runBenchmarkLazyInitSerial)
	b.Run("parallel", runBenchmarkLazyInitParallel)
}

func runBenchmarkLazyInitSerial(b *testing.B) {
	s := newInitialisedSemaphore(b)

	for b.Loop() {
		if err := s.lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

func runBenchmarkLazyInitParallel(b *testing.B) {
	s := newInitialisedSemaphore(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := s.lazyInit(); err != nil {
				b.Errorf("lazyInit returned error: %v", err)
				return
			}
		}
	})
}

func newInitialisedSemaphore(b *testing.B) *Semaphore {
	b.Helper()
	s := &Semaphore{}
	if err := s.lazyInit(); err != nil {
		b.Fatalf("lazyInit returned error: %v", err)
	}
	return s
}

// cancelledRLockRounds is the number of cancelled unsafeRLock attempts run
// per scenario. select chooses uniformly between the ready acquire case and
// the already-closed abort, so each post-acquire rollback arm is reached with
// probability 1/2 per round; enough rounds drive the miss probability
// (2^-rounds) to nothing without depending on a single lucky scheduling.
const cancelledRLockRounds = 256

// TestSemaphore_UnsafeRLockCancelled pins the two post-acquire rollback arms
// of unsafeRLock: a reader that wins its slot and only then observes the
// abort must release what it took. These live in the white-box test package
// because unsafeRLock and the channel fields are unexported, and because the
// arms are unreachable from the public RLockContext tests, which always hold
// a lock before cancelling and so never present an idle or reader-occupied
// slot as the ready acquire case.
func TestSemaphore_UnsafeRLockCancelled(t *testing.T) {
	t.Run("first reader", runTestUnsafeRLockCancelledFirst)
	t.Run("subsequent reader", runTestUnsafeRLockCancelledSubsequent)
}

// runTestUnsafeRLockCancelledFirst exercises the "won the global slot as
// first reader, then noticed the abort" arm. On an idle semaphore both the
// global send and the closed abort are ready; whichever wins, unsafeRLock
// reports cancellation and leaves the slot empty, so the receiver stays idle
// across rounds.
func runTestUnsafeRLockCancelledFirst(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")

	abort := closedAbort()
	for range cancelledRLockRounds {
		core.AssertMustEqual(t, true, s.unsafeRLock(abort), "cancelled")
		core.AssertMustEqual(t, 0, len(s.global), "global released")
	}
}

// runTestUnsafeRLockCancelledSubsequent exercises the "took the next reader's
// slot, then noticed the abort" arm. With one reader already holding, the
// ready acquire case is the readers channel rather than the global send, so
// the rollback puts the count back. Each round restores the held reader, so
// the receiver state is stable across rounds.
func runTestUnsafeRLockCancelledSubsequent(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	core.AssertMustNoError(t, s.lazyInit(), "lazyInit")

	// Seed one reader so the readers channel is the ready acquire case.
	core.AssertMustEqual(t, false, s.unsafeRLock(nil), "seed reader")

	abort := closedAbort()
	for range cancelledRLockRounds {
		core.AssertMustEqual(t, true, s.unsafeRLock(abort), "cancelled")
		core.AssertMustEqual(t, 1, len(s.readers), "reader count restored")
		core.AssertMustEqual(t, 1, len(s.global), "global still held")
	}
}

// closedAbort returns an already-cancelled abort channel: every receive on it
// succeeds immediately, so isCancelled reports true.
func closedAbort() <-chan struct{} {
	abort := make(chan struct{})
	close(abort)
	return abort
}

// semaphoreTestTimeout caps each wait for a call to return.
//
// semaphoreOpenGuard is how long a call must stay blocked for the tests
// to accept that it is waiting.
const (
	semaphoreTestTimeout = time.Second

	semaphoreOpenGuard = 20 * time.Millisecond
)

// TestSemaphore_WriterPreference pins the order the turnstile keeps: a
// writer waiting for the readers to leave holds it, so readers arriving
// after that writer wait behind it, until the writer takes the lock or
// gives up. It lives in the white-box test package to see the turnstile
// the writer holds, which shows the writer is waiting.
func TestSemaphore_WriterPreference(t *testing.T) {
	t.Run("reader behind writer", runTestReaderBehindWriter)
	t.Run("RLock behind Lock", runTestRLockBehindLock)
	t.Run("TryRLock behind writer", runTestTryRLockBehindWriter)
	t.Run("RLockContext cancelled behind writer",
		runTestRLockContextBehindWriter)
	t.Run("LockContext cancelled while queued",
		runTestLockContextCancelledQueued)
	t.Run("TryLock while turnstile held", runTestTryLockTurnstileHeld)
}

// newSemaphoreWithQueuedWriter returns a semaphore holding one read lock
// and a writer waiting behind it through lock, and the channel lock
// reports on. It waits until the writer holds the turnstile, as readers
// arriving from then on wait behind it.
func newSemaphoreWithQueuedWriter(t *testing.T, lock func(*Semaphore) error) (
	*Semaphore, <-chan error) {
	t.Helper()
	s := &Semaphore{}
	s.RLock()

	locking := goWait(func() error { return lock(s) })
	core.AssertMustEventually(t, func() bool { return !s.turn.TryPass() },
		semaphoreTestTimeout, "writer holds the turnstile")
	assertWaiting(t, locking, "writer")
	return s, locking
}

// writerLockContext returns a writer's entry that takes the lock through
// LockContext under ctx.
func writerLockContext(ctx context.Context) func(*Semaphore) error {
	return func(s *Semaphore) error { return s.LockContext(ctx) }
}

// writerLock is a writer's entry that takes the lock through Lock.
func writerLock(s *Semaphore) error {
	s.Lock()
	return nil
}

func runTestReaderBehindWriter(t *testing.T) {
	t.Helper()
	lock := writerLockContext(t.Context())
	s, locking := newSemaphoreWithQueuedWriter(t, lock)
	reading := goWait(func() error { return s.RLockContext(t.Context()) })
	assertReaderBehindWriter(t, s, locking, reading)
}

// runTestRLockBehindLock queues the writer through Lock and the reader
// through RLock, the forms that cannot be cancelled.
func runTestRLockBehindLock(t *testing.T) {
	t.Helper()
	s, locking := newSemaphoreWithQueuedWriter(t, writerLock)
	reading := goWait(func() error { s.RLock(); return nil })
	assertReaderBehindWriter(t, s, locking, reading)
}

// assertReaderBehindWriter releases the read lock s holds, then the
// writer's lock, and states that the reader reporting on reading waits
// behind the writer reporting on locking until the writer unlocks.
func assertReaderBehindWriter(t *testing.T, s *Semaphore,
	locking, reading <-chan error) {
	t.Helper()
	assertWaiting(t, reading, "reader behind the writer")

	s.RUnlock()
	assertReturns(t, locking, nil, "writer")
	assertWaiting(t, reading, "reader behind the holder")

	s.Unlock()
	assertReturns(t, reading, nil, "reader")
	s.RUnlock()
}

func runTestTryRLockBehindWriter(t *testing.T) {
	t.Helper()
	lock := writerLockContext(t.Context())
	s, locking := newSemaphoreWithQueuedWriter(t, lock)
	core.AssertMustFalse(t, s.TryRLock(), "TryRLock behind the writer")

	s.RUnlock()
	assertReturns(t, locking, nil, "LockContext")
	s.Unlock()
}

// runTestRLockContextBehindWriter cancels a reader waiting behind the
// writer, and states that it took nothing: the writer gets the lock once
// the first reader leaves.
func runTestRLockContextBehindWriter(t *testing.T) {
	t.Helper()
	lock := writerLockContext(t.Context())
	s, locking := newSemaphoreWithQueuedWriter(t, lock)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	reading := goWait(func() error { return s.RLockContext(ctx) })
	assertWaiting(t, reading, "RLockContext behind the writer")

	cancel()
	assertReturns(t, reading, context.Canceled, "RLockContext")

	s.RUnlock()
	assertReturns(t, locking, nil, "LockContext")
	s.Unlock()
}

// runTestLockContextCancelledQueued cancels a writer waiting behind a
// reader, and states that it gives the turnstile back: another reader
// gets in while the first still holds the lock.
func runTestLockContextCancelledQueued(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	lock := writerLockContext(ctx)
	s, locking := newSemaphoreWithQueuedWriter(t, lock)

	cancel()
	assertReturns(t, locking, context.Canceled, "LockContext")
	core.AssertMustTrue(t, s.TryRLock(), "TryRLock after the cancel")
	s.RUnlock()
	s.RUnlock()
}

// runTestTryLockTurnstileHeld holds the turnstile directly, as a writer
// does between taking it and reaching global, the gap TryLock's own
// turnstile check covers. Behind a writer already waiting on global, the
// lock is taken and TryLock would fail either way.
func runTestTryLockTurnstileHeld(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	s.turn.Lock()
	defer s.turn.Unlock()

	core.AssertFalse(t, s.TryLock(), "TryLock")
}

// TestSemaphore_UnlockReadLocked pins Unlock on a semaphore readers hold:
// it panics without waiting, and puts back what it took. It lives in the
// white-box test package to take the readers' count out, as a reader
// changing it does.
func TestSemaphore_UnlockReadLocked(t *testing.T) {
	t.Run("writer waiting", runTestUnlockReadLockedWriterWaiting)
	t.Run("count in use", runTestUnlockReadLockedCountInUse)
	t.Run("count in use, writer waiting",
		runTestUnlockReadLockedCountInUseWriterWaiting)
}

// runTestUnlockReadLockedWriterWaiting calls Unlock while a reader holds
// the lock and a writer waits for it. The writer stays behind the reader.
func runTestUnlockReadLockedWriterWaiting(t *testing.T) {
	t.Helper()
	lock := writerLockContext(t.Context())
	s, locking := newSemaphoreWithQueuedWriter(t, lock)

	assertReturns(t, goUnlock(s), errors.ErrReadLocked, "Unlock")
	assertWaiting(t, locking, "LockContext")

	s.RUnlock()
	assertReturns(t, locking, nil, "LockContext")
	s.Unlock()
}

// runTestUnlockReadLockedCountInUse calls Unlock while the readers' count
// is out, so Unlock takes the readers' value from global, and puts it
// back.
func runTestUnlockReadLockedCountInUse(t *testing.T) {
	t.Helper()
	s := &Semaphore{}
	s.RLock()
	readers := <-s.readers

	core.AssertPanic(t, s.Unlock, errors.ErrReadLocked, "Unlock")
	core.AssertMustEqual(t, 1, len(s.global), "global still held")

	s.readers <- readers
	s.RUnlock()
	core.AssertMustTrue(t, s.TryLock(), "TryLock after RUnlock")
	s.Unlock()
}

// runTestUnlockReadLockedCountInUseWriterWaiting calls Unlock while the
// readers' count is out and a writer waits. Taking the readers' value
// from global hands the slot to the writer, so Unlock cannot put it back.
func runTestUnlockReadLockedCountInUseWriterWaiting(t *testing.T) {
	t.Helper()
	lock := writerLockContext(t.Context())
	s, locking := newSemaphoreWithQueuedWriter(t, lock)
	<-s.readers

	assertReturns(t, goUnlock(s), errors.ErrReadLocked, "Unlock")
	assertReturns(t, locking, nil, "LockContext")
}

// goWait runs fn in a goroutine and returns the channel its result
// arrives on.
func goWait(fn func() error) <-chan error {
	done := make(chan error, 1)
	go func() { done <- fn() }()
	return done
}

// assertWaiting states that the call behind done is still blocked.
func assertWaiting(t *testing.T, done <-chan error, name string) {
	t.Helper()
	core.AssertMustQuiet(t, done, semaphoreOpenGuard, "%s waiting", name)
}

// assertReturns waits for the call behind done and states that it
// failed with want, or succeeded when want is nil.
func assertReturns(t *testing.T, done <-chan error, want error, name string) {
	t.Helper()
	got := core.AssertMustReceives(t, done, 1, semaphoreTestTimeout, name)
	if want == nil {
		core.AssertNoError(t, got[0], name)
	} else {
		core.AssertErrorIs(t, got[0], want, name)
	}
}

// goUnlock calls Unlock in a goroutine and returns the channel the panic
// it raises arrives on.
func goUnlock(s *Semaphore) <-chan error {
	return goWait(func() error {
		return core.Catch(func() error {
			s.Unlock()
			return nil
		})
	})
}
