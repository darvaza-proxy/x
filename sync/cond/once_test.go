package cond_test

import (
	"runtime"
	"testing"
	"unsafe"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
)

// errOnceTest is what the initialisers under test fail or panic with.
const errOnceTest core.StringError = "once test"

func succeedInit() error { return nil }
func failInit() error    { return errOnceTest }
func panicInit() error   { panic(errOnceTest) }

// countedInit returns an initialiser that counts its calls in calls and
// then runs fn.
func countedInit(calls *int, fn func() error) func() error {
	return func() error {
		*calls++
		return fn()
	}
}

// blockedInit returns an initialiser that closes started, then waits for
// release before failing.
func blockedInit(started, release chan struct{}) func() error {
	return func() error {
		close(started)
		<-release
		return errOnceTest
	}
}

// goexitInit returns an initialiser that closes started, then leaves
// through runtime.Goexit without returning.
func goexitInit(started chan struct{}) func() error {
	return func() error {
		close(started)
		runtime.Goexit()
		return nil
	}
}

// onceDoTestCase runs an initialiser through Do twice. The first call
// returns its result, and the second returns the same without running
// the initialiser again.
type onceDoTestCase struct {
	fn   func() error
	want error
	name string

	panics bool
}

func newOnceDoTestCase(name string, fn func() error, want error) onceDoTestCase {
	return onceDoTestCase{
		name: name,
		fn:   fn,
		want: want,
	}
}

// newOnceDoPanicTestCase builds a case whose initialiser panics with
// payload, which Do returns as a [core.PanicError].
func newOnceDoPanicTestCase(name string, fn func() error, payload error) onceDoTestCase {
	return onceDoTestCase{
		name:   name,
		fn:     fn,
		want:   payload,
		panics: true,
	}
}

func (tc onceDoTestCase) Name() string { return tc.name }

func (tc onceDoTestCase) Test(t *testing.T) {
	t.Helper()
	var once cond.Once
	var calls int

	core.AssertFalse(t, once.Done(), "Done before Do")
	first := once.Do(countedInit(&calls, tc.fn))
	tc.assertResult(t, first)
	core.AssertEqual(t, first == nil, once.Done(), "Done")

	second := once.Do(countedInit(&calls, tc.fn))
	core.AssertEqual(t, first, second, "second Do")
	core.AssertEqual(t, 1, calls, "calls")
}

func (tc onceDoTestCase) assertResult(t *testing.T, err error) {
	t.Helper()
	if !tc.panics {
		core.AssertEqual(t, tc.want, err, "Do")
		return
	}

	pe, ok := core.AssertTypeIs[*core.PanicError](t, err, "Do")
	if ok {
		core.AssertEqual(t, any(tc.want), pe.Recovered(), "recovered")
	}
}

var _ core.TestCase = onceDoTestCase{}

func onceDoTestCases() []onceDoTestCase {
	return core.S(
		newOnceDoTestCase("success", succeedInit, nil),
		newOnceDoTestCase("error", failInit, errOnceTest),
		newOnceDoPanicTestCase("panic", panicInit, errOnceTest),
	)
}

func TestOnceDo(t *testing.T) {
	core.RunTestCases(t, onceDoTestCases())
}

func TestOnceNilFunction(t *testing.T) {
	var once cond.Once
	var calls int

	core.AssertEqual(t, nil, once.Do(nil), "Do(nil)")
	core.AssertTrue(t, once.Done(), "Done")
	core.AssertEqual(t, nil, once.Do(countedInit(&calls, failInit)), "second Do")
	core.AssertEqual(t, 0, calls, "calls")
}

func TestOnceNilReceiver(t *testing.T) {
	var once *cond.Once
	var calls int

	err := once.Do(countedInit(&calls, succeedInit))
	core.AssertErrorIs(t, err, errors.ErrNilReceiver, "Do")
	core.AssertEqual(t, 0, calls, "calls")
	core.AssertFalse(t, once.Done(), "Done")
}

func TestOnceWaitsForFirst(t *testing.T) {
	var once cond.Once
	var calls int
	started := make(chan struct{})
	release := make(chan struct{})

	first := goWait(func() error { return once.Do(blockedInit(started, release)) })
	core.AssertMustClosed(t, started, turnstileTestTimeout, "first initialiser")
	second := goWait(func() error { return once.Do(countedInit(&calls, succeedInit)) })

	assertWaiting(t, second, "second Do")
	core.AssertFalse(t, once.Done(), "Done while running")
	close(release)
	assertReturns(t, first, errOnceTest, "first Do")
	assertReturns(t, second, errOnceTest, "second Do")
	core.AssertEqual(t, 0, calls, "second initialiser calls")
}

// TestOnceGoexit states that an initialiser leaving through
// runtime.Goexit is remembered as not having initialised, and the next
// caller's initialiser does not run.
func TestOnceGoexit(t *testing.T) {
	var once cond.Once
	var calls int
	started := make(chan struct{})

	goWait(func() error { return once.Do(goexitInit(started)) })
	core.AssertMustClosed(t, started, turnstileTestTimeout, "first initialiser")
	second := goWait(func() error { return once.Do(countedInit(&calls, succeedInit)) })

	assertReturns(t, second, errors.ErrNotInitialised, "second Do")
	core.AssertEqual(t, 0, calls, "second initialiser calls")
	core.AssertFalse(t, once.Done(), "Done")
}

func TestOncePanicStack(t *testing.T) {
	var once cond.Once
	core.AssertTopFrame(t, once.Do(panicInit), "panicInit", "Do")
}

// TestOnceSize pins Once at one pointer, the copy guard adding nothing.
func TestOnceSize(t *testing.T) {
	var once cond.Once
	core.AssertEqual(t, unsafe.Sizeof(unsafe.Pointer(nil)), unsafe.Sizeof(once), "Once")
}
