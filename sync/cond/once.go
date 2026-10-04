package cond

import (
	"runtime"
	"sync/atomic"
	"unsafe"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// onceNil is the result stored for an initialiser that returned nil, so
// a successful run allocates nothing and Do recognises it by address.
//
// onceExited is the result stored for an initialiser that left through
// runtime.Goexit, which returns nothing.
//
// onceRunning marks a Once whose initialiser is running.
var (
	onceNil     error
	onceExited  error = errors.ErrNotInitialised
	onceRunning error
)

// newOnceResult returns the cell to store err in: onceNil for nil, or
// a new one holding err.
func newOnceResult(err error) *error {
	if err == nil {
		return &onceNil
	}

	p := new(error)
	*p = err
	return p
}

// Once runs an initialiser once and remembers its result, keeping
// its state in a single pointer. Callers that arrive while it runs
// yield the processor until it finishes, so the initialiser should be
// brief.
//
// The zero value is ready for use, and a Once must not be copied
// after first use.
type Once struct {
	_ noCopy

	// result holds a *error, through the untyped atomic functions:
	// the generic atomic.Pointer leaves Do at the edge of the
	// inlining budget. load, store and claim convert.
	result unsafe.Pointer
}

// load returns the cell holding the result: nil before the first run,
// and &onceRunning during it.
func (o *Once) load() *error {
	return (*error)(atomic.LoadPointer(&o.result))
}

// store stores the cell holding the result.
func (o *Once) store(result *error) {
	atomic.StorePointer(&o.result, unsafe.Pointer(result))
}

// claim marks the first run as started, and reports whether this call
// did.
func (o *Once) claim() bool {
	return atomic.CompareAndSwapPointer(&o.result, nil, unsafe.Pointer(&onceRunning))
}

// Do calls fn on the first call, and returns that call's result to every
// call: the error fn returned, the [core.PanicError] of its panic, or
// [errors.ErrNotInitialised] if it left through runtime.Goexit. A nil
// fn counts as a success. Callers that arrive while fn runs wait for
// it, and fn must not call Do on the same Once. Do returns
// [errors.ErrNilReceiver] if the Once is nil, without calling fn.
func (o *Once) Do(fn func() error) error {
	if o != nil && o.load() == &onceNil {
		return nil
	}
	return o.doSlow(fn)
}

// Done reports whether the initialiser ran and succeeded, so Do returns
// nil without calling its function. A failed run, a run in progress and
// a nil Once report false.
func (o *Once) Done() bool {
	return o != nil && o.load() == &onceNil
}

// doSlow handles what the fast path leaves: a nil receiver, a
// remembered failure, a run in progress, and the first run. It is
// kept out of Do so Do inlines.
func (o *Once) doSlow(fn func() error) error {
	switch {
	case o == nil:
		return errors.ErrNilReceiver
	case o.claim():
		return o.run(fn)
	default:
		// The claim failed, so the state is set: claim stores
		// &onceRunning, run stores a result cell, and nothing stores
		// nil back.
		for {
			p := o.load()
			if p != &onceRunning {
				return *p
			}
			runtime.Gosched()
		}
	}
}

// run calls fn and stores its result. The deferred function stores
// whatever fn left: its error, the [core.PanicError] of its panic, or
// onceExited when it left through runtime.Goexit.
func (o *Once) run(fn func() error) (err error) {
	result := &onceExited
	defer func() {
		if r := core.AsRecovered(recover()); r != nil {
			err = r
			result = newOnceResult(r)
		}
		o.store(result)
	}()

	if fn != nil {
		err = fn()
	}
	result = newOnceResult(err)
	return err
}
