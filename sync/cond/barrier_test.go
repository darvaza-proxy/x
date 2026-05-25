package cond_test

import (
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/internal/synctesting"
)

// barrierTestTimeout caps each foreground synchronisation step. Generous
// enough to absorb scheduler jitter on loaded CI workers, tight enough
// that a hung waiter fails the run instead of stalling.
//
// barrierOpenGuard bounds negative-case assertions ("did not signal"):
// long enough to catch a spurious early close, short enough to keep the
// test responsive.
const (
	barrierTestTimeout = time.Second

	barrierOpenGuard = 20 * time.Millisecond
)

// barrierStateTestCase exercises IsNil and IsClosed for the same Barrier
// state. Both methods read the same lifecycle, so the row carries both
// expectations to keep the state matrix legible.
type barrierStateTestCase struct {
	name string

	setup func() *cond.Barrier

	wantNil    bool
	wantClosed bool
}

func newBarrierStateTestCase(name string, setup func() *cond.Barrier,
	wantNil, wantClosed bool) barrierStateTestCase {
	return barrierStateTestCase{
		name:       name,
		setup:      setup,
		wantNil:    wantNil,
		wantClosed: wantClosed,
	}
}

func (tc barrierStateTestCase) Name() string { return tc.name }

func (tc barrierStateTestCase) Test(t *testing.T) {
	t.Helper()
	b := tc.setup()
	core.AssertEqual(t, tc.wantNil, b.IsNil(), "IsNil")
	core.AssertEqual(t, tc.wantClosed, b.IsClosed(), "IsClosed")
}

var _ core.TestCase = barrierStateTestCase{}

func barrierStateTestCases() []barrierStateTestCase {
	return []barrierStateTestCase{
		newBarrierStateTestCase("nil pointer",
			func() *cond.Barrier { return nil },
			true, true),
		newBarrierStateTestCase("uninitialised",
			func() *cond.Barrier { return &cond.Barrier{} },
			true, true),
		newBarrierStateTestCase("initialised",
			func() *cond.Barrier { return cond.NewBarrier() },
			false, false),
		newBarrierStateTestCase("closed",
			func() *cond.Barrier {
				b := cond.NewBarrier()
				_ = b.Close()
				return b
			},
			false, true),
		newBarrierStateTestCase("after broadcast",
			func() *cond.Barrier {
				b := cond.NewBarrier()
				b.Broadcast()
				return b
			},
			false, false),
	}
}

// TestBarrierState verifies IsNil and IsClosed across every reachable
// lifecycle state.
func TestBarrierState(t *testing.T) {
	core.RunTestCases(t, barrierStateTestCases())
}

// barrierInitTestCase exercises Init across receiver and prior-state
// conditions.
type barrierInitTestCase struct {
	name string

	setup func() *cond.Barrier

	wantErr error
}

func newBarrierInitTestCase(name string, setup func() *cond.Barrier,
	wantErr error) barrierInitTestCase {
	return barrierInitTestCase{name: name, setup: setup, wantErr: wantErr}
}

func (tc barrierInitTestCase) Name() string { return tc.name }

func (tc barrierInitTestCase) Test(t *testing.T) {
	t.Helper()
	b := tc.setup()
	err := b.Init()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Init")
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Init error")
}

var _ core.TestCase = barrierInitTestCase{}

func barrierInitTestCases() []barrierInitTestCase {
	return []barrierInitTestCase{
		newBarrierInitTestCase("nil receiver",
			func() *cond.Barrier { return nil },
			errors.ErrNilReceiver),
		newBarrierInitTestCase("already initialised",
			func() *cond.Barrier {
				b := &cond.Barrier{}
				_ = b.Init()
				return b
			},
			errors.ErrAlreadyInitialised),
		newBarrierInitTestCase("fresh succeeds",
			func() *cond.Barrier { return &cond.Barrier{} },
			nil),
	}
}

// TestBarrierInit verifies Init's receiver and prior-state error paths
// alongside the happy path.
func TestBarrierInit(t *testing.T) {
	core.RunTestCases(t, barrierInitTestCases())
}

// barrierCloseTestCase exercises Close across receiver and prior-state
// conditions.
type barrierCloseTestCase struct {
	name string

	setup func() *cond.Barrier

	wantErr error
}

func newBarrierCloseTestCase(name string, setup func() *cond.Barrier,
	wantErr error) barrierCloseTestCase {
	return barrierCloseTestCase{name: name, setup: setup, wantErr: wantErr}
}

func (tc barrierCloseTestCase) Name() string { return tc.name }

func (tc barrierCloseTestCase) Test(t *testing.T) {
	t.Helper()
	b := tc.setup()
	err := b.Close()
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "Close")
		return
	}
	core.AssertErrorIs(t, err, tc.wantErr, "Close error")
}

var _ core.TestCase = barrierCloseTestCase{}

func barrierCloseTestCases() []barrierCloseTestCase {
	return []barrierCloseTestCase{
		newBarrierCloseTestCase("nil receiver",
			func() *cond.Barrier { return nil },
			errors.ErrNilReceiver),
		newBarrierCloseTestCase("uninitialised",
			func() *cond.Barrier { return &cond.Barrier{} },
			errors.ErrNotInitialised),
		newBarrierCloseTestCase("happy path",
			func() *cond.Barrier { return cond.NewBarrier() },
			nil),
		newBarrierCloseTestCase("double close",
			func() *cond.Barrier {
				b := cond.NewBarrier()
				_ = b.Close()
				return b
			},
			errors.ErrClosed),
	}
}

// TestBarrierClose verifies Close's receiver and prior-state error paths
// alongside the happy path. The internal `!ok` branch (channel closed
// between the fast-path Load and the receive) is driven by concurrent
// Close in TestBarrierConcurrentClose.
func TestBarrierClose(t *testing.T) {
	core.RunTestCases(t, barrierCloseTestCases())
}

// barrierTryAcquireTestCase exercises TryAcquire across the three
// observable token states: available, held by another acquirer, and
// already closed. The closed state takes the channel-receive branch with
// ok=false, distinct from the default branch taken when the channel is
// open but empty.
type barrierTryAcquireTestCase struct {
	name string

	setup func() *cond.Barrier

	wantOk bool
}

func newBarrierTryAcquireTestCase(name string, setup func() *cond.Barrier,
	wantOk bool) barrierTryAcquireTestCase {
	return barrierTryAcquireTestCase{name: name, setup: setup, wantOk: wantOk}
}

func (tc barrierTryAcquireTestCase) Name() string { return tc.name }

func (tc barrierTryAcquireTestCase) Test(t *testing.T) {
	t.Helper()
	b := tc.setup()
	token, ok := b.TryAcquire()
	core.AssertEqual(t, tc.wantOk, ok, "ok")
	if tc.wantOk {
		core.AssertNotNil(t, token, "token")
		b.Release(token)
		return
	}
	core.AssertNil(t, token, "token")
}

var _ core.TestCase = barrierTryAcquireTestCase{}

func barrierTryAcquireTestCases() []barrierTryAcquireTestCase {
	return []barrierTryAcquireTestCase{
		newBarrierTryAcquireTestCase("fresh barrier yields token",
			func() *cond.Barrier { return cond.NewBarrier() },
			true),
		// The Barrier returned here is intentionally left in a
		// permanently-held state: the token is received but never
		// released, so the Barrier cannot be Close()d cleanly. Acceptable
		// because each row owns a fresh Barrier that the GC reclaims
		// after the subtest.
		newBarrierTryAcquireTestCase("held barrier reports unavailable",
			func() *cond.Barrier {
				b := cond.NewBarrier()
				<-b.Acquire()
				return b
			},
			false),
		newBarrierTryAcquireTestCase("closed barrier reports unavailable",
			func() *cond.Barrier {
				b := cond.NewBarrier()
				_ = b.Close()
				return b
			},
			false),
	}
}

// TestBarrierTryAcquire verifies TryAcquire across every observable
// state.
func TestBarrierTryAcquire(t *testing.T) {
	core.RunTestCases(t, barrierTryAcquireTestCases())
}

// TestBarrierAcquireRelease verifies that an outstanding acquisition
// blocks a second acquirer until the first releases.
func TestBarrierAcquireRelease(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	token := <-b.Acquire()
	core.AssertMustNotNil(t, token, "first acquire")

	synctesting.AssertOpen(t, b.Acquire(), barrierOpenGuard,
		"second acquire blocked while held")

	b.Release(token)

	got, ok := b.TryAcquire()
	core.AssertTrue(t, ok, "TryAcquire after release")
	core.AssertNotNil(t, got, "TryAcquire token after release")
	b.Release(got)
}

// TestBarrierTokenNonConsuming verifies Token returns the live token
// without draining it from the barrier: acquisition still succeeds
// afterwards.
func TestBarrierTokenNonConsuming(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	core.AssertMustNotNil(t, b.Token(), "Token")

	token, ok := b.TryAcquire()
	core.AssertTrue(t, ok, "TryAcquire after Token")
	core.AssertNotNil(t, token, "token after Token")
	b.Release(token)
}

// TestBarrierBroadcast verifies Broadcast closes the current Token and
// installs a fresh one.
func TestBarrierBroadcast(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	original := b.Token()
	core.AssertMustNotNil(t, original, "original token")

	b.Broadcast()

	synctesting.AssertClosed(t, original, barrierTestTimeout,
		"original token closed after Broadcast")

	replacement := b.Token()
	core.AssertNotNil(t, replacement, "replacement token")
	core.AssertNotSame(t, original, replacement,
		"replacement is a distinct channel")
}

// TestBarrierSignaled verifies the Barrier-level Signaled accessor
// returns the live Token channel, which closes on Broadcast.
func TestBarrierSignaled(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	ch := b.Signaled()
	core.AssertMustNotNil(t, ch, "Signaled channel")

	synctesting.AssertOpen(t, ch, barrierOpenGuard,
		"Signaled open before Broadcast")

	b.Broadcast()

	synctesting.AssertClosed(t, ch, barrierTestTimeout,
		"Signaled closed after Broadcast")
}

// TestBarrierWait verifies the Barrier-level Wait blocks until
// Broadcast.
func TestBarrierWait(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	done := make(chan struct{})
	go func() {
		b.Wait()
		close(done)
	}()

	synctesting.AssertOpen(t, done, barrierOpenGuard,
		"Wait blocks before Broadcast")

	b.Broadcast()

	synctesting.AssertClosed(t, done, barrierTestTimeout,
		"Wait returns after Broadcast")
}

// TestBarrierSignal verifies the three observable Signal outcomes:
// false with no waiter, true once a waiter has parked on the Token
// receive, and false on a closed Barrier. The polling loop bounds the
// inherent race between starting the waiter goroutine and the Token
// receive.
func TestBarrierSignal(t *testing.T) {
	b := cond.NewBarrier()

	core.AssertEqual(t, false, b.Signal(), "Signal without waiter")

	waited := make(chan struct{})
	go func() {
		b.Wait()
		close(waited)
	}()

	signaled := synctesting.WaitForCond(b.Signal, barrierTestTimeout,
		synctesting.PollStep)
	core.AssertTrue(t, signaled, "Signal eventually succeeds with waiter")

	synctesting.AssertClosed(t, waited, barrierTestTimeout,
		"waiter released after Signal")

	_ = b.Close()
	core.AssertEqual(t, false, b.Signal(), "Signal after Close")
}

// TestBarrierTokenAfterClose verifies Token returns nil once the
// underlying channel is closed.
func TestBarrierTokenAfterClose(t *testing.T) {
	b := cond.NewBarrier()
	_ = b.Close()

	core.AssertNil(t, b.Token(), "Token after Close")
}

// TestBarrierAcquireAfterClose verifies the Acquire channel yields a
// nil token immediately once the Barrier is closed — the channel-based
// shutdown signal for any code parked on `<-b.Acquire()`.
func TestBarrierAcquireAfterClose(t *testing.T) {
	b := cond.NewBarrier()
	_ = b.Close()

	select {
	case token := <-b.Acquire():
		core.AssertNil(t, token, "post-close acquire token")
	case <-time.After(barrierTestTimeout):
		t.Fatal("Acquire on closed barrier blocked")
	}
}

// TestBarrierBroadcastAfterClose verifies Broadcast is a safe no-op on
// a closed Barrier. The `ok` check on the channel receive skips the
// close-and-reseat work without panicking.
func TestBarrierBroadcastAfterClose(t *testing.T) {
	b := cond.NewBarrier()
	_ = b.Close()

	b.Broadcast()

	token, ok := b.TryAcquire()
	core.AssertEqual(t, false, ok, "TryAcquire after Broadcast on closed")
	core.AssertNil(t, token, "token after Broadcast on closed")
}

// TestBarrierWaitAfterClose pins the current design: Wait on a closed
// Barrier blocks indefinitely because Token() returns nil and `<-nil`
// never unblocks. The waiter goroutine is intentionally leaked for the
// remainder of the test process — any future change that makes Wait
// return cleanly on closed Barriers will be visible here.
func TestBarrierWaitAfterClose(t *testing.T) {
	b := cond.NewBarrier()
	_ = b.Close()

	done := make(chan struct{})
	go func() {
		b.Wait()
		close(done)
	}()

	synctesting.AssertOpen(t, done, barrierOpenGuard,
		"Wait on closed Barrier blocks (current design)")
}

// TestBarrierReleaseAfterClosePanics pins the holder-partition contract
// after the closed-state guard was dropped from Release: returning a
// non-nil Token to a closed Barrier sends on a closed channel and panics.
// Under correct usage this is unreachable — Close drains the live token
// before closing bs.b — but a caller that caches a stale Token and Releases
// it post-close fails loudly rather than silently dropping it.
func TestBarrierReleaseAfterClosePanics(t *testing.T) {
	b := cond.NewBarrier()

	// Acquire and return the live token so Close can drain and proceed.
	token, ok := b.TryAcquire()
	core.AssertMustTrue(t, ok, "TryAcquire before Close")
	b.Release(token)
	core.AssertMustNoError(t, b.Close(), "Close")

	// A caller holding a stale non-nil Token Releases after Close: the
	// send on the closed channel must panic.
	stale := make(cond.Token)
	err := core.Catch(func() error {
		b.Release(stale)
		return nil
	})
	core.AssertError(t, err, "Release on closed Barrier panics")
}

// TestBarrierConcurrent verifies the acquire-release cycle survives N
// concurrent goroutines and leaves the Barrier in a usable state.
func TestBarrierConcurrent(t *testing.T) {
	b := cond.NewBarrier()
	defer func() { _ = b.Close() }()

	const n = 10
	done := make(chan struct{}, n)

	for range n {
		go func() {
			token := <-b.Acquire()
			b.Release(token)
			done <- struct{}{}
		}()
	}

	synctesting.AssertReadersReady(t, done, n, barrierTestTimeout,
		"all goroutines acquire-release")

	token, ok := b.TryAcquire()
	core.AssertTrue(t, ok, "TryAcquire after concurrent ops")
	core.AssertNotNil(t, token, "post-concurrent token")
	b.Release(token)
}

// TestBarrierConcurrentClose pins the queued-Close design: when two
// goroutines call Close concurrently with both parked at <-bs.b,
// the winner closes bs.b and the loser wakes from its blocked
// receive with ok=false, returning ErrClosed via the !ok branch
// rather than the fast-path check.
//
// Holding the token across the spawn keeps both Closers past the
// fast-path Load with no Store possible. The pre-Release sleep
// gives both goroutines time to park at <-bs.b before the winner
// is chosen; should one not have parked yet, it loses via the
// fast-path Load instead and the assertions still hold — that run
// merely leaves the !ok branch uncovered.
func TestBarrierConcurrentClose(t *testing.T) {
	b := cond.NewBarrier()

	token, ok := b.TryAcquire()
	core.AssertMustTrue(t, ok, "initial TryAcquire")

	const goroutines = 2
	started := make(chan struct{}, goroutines)
	results := make(chan error, goroutines)

	for range goroutines {
		go func() {
			started <- struct{}{}
			results <- b.Close()
		}()
	}

	synctesting.AssertMustReadersReady(t, started, goroutines,
		barrierTestTimeout, "Close goroutines started")

	// Both goroutines have signalled started; give them time to
	// enter Close and park on <-bs.b before we Release the token.
	time.Sleep(barrierOpenGuard)

	b.Release(token)

	synctesting.AssertMustEventually(t, func() bool {
		return len(results) == goroutines
	}, barrierTestTimeout, "both Close goroutines return")

	var nilCount, closedCount int
	for range goroutines {
		err := <-results
		if err == nil {
			nilCount++
			continue
		}
		core.AssertErrorIs(t, err, errors.ErrClosed,
			"loser returns ErrClosed")
		closedCount++
	}

	core.AssertEqual(t, 1, nilCount, "exactly one winner")
	core.AssertEqual(t, 1, closedCount, "exactly one loser")
}

// TestTokenWait verifies the Token-level Wait blocks until the channel
// is closed.
func TestTokenWait(t *testing.T) {
	token := make(cond.Token)
	done := make(chan struct{})

	go func() {
		token.Wait()
		close(done)
	}()

	synctesting.AssertOpen(t, done, barrierOpenGuard,
		"Wait blocks before close")

	close(token)

	synctesting.AssertClosed(t, done, barrierTestTimeout,
		"Wait returns after close")
}
