package cond

import (
	"context"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/sync/atomic"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/mutex"
)

// Turnstile is a lock with a second way through. Holding it takes it
// exclusively, as a mutex. Passing it waits until it is free and goes
// through without keeping it. Holders and passers wait on the same
// channel and are served in arrival order, so whoever arrives after a
// holder waits behind it.
//
// The zero value is ready for use, and a Turnstile must not be copied
// after first use. It uses a [Barrier] as a token lock. Close sets a
// flag and closes the token, which wakes the waiters.
type Turnstile struct {
	held Token
	done Token
	once Once
	b    Barrier

	closed atomic.Bool
}

// lazyInit initialises the Turnstile on first use. After that it costs
// an atomic load.
func (t *Turnstile) lazyInit() error {
	switch {
	case t == nil:
		return errors.ErrNilReceiver
	case t.once.Done():
		return nil
	default:
		return t.once.Do(t.init)
	}
}

func (t *Turnstile) init() error {
	// Init rejects a nil receiver and a Barrier already initialised.
	// t.b belongs to a non-nil t, and once runs this a single time.
	core.MustNoError(t.b.Init())
	t.done = t.b.Token()
	return nil
}

// acquire waits for the token until abort closes. It returns
// [errors.ErrClosed] once the Turnstile is closed, and
// [context.Canceled] if abort closes first. A nil abort waits
// indefinitely.
func (t *Turnstile) acquire(abort <-chan struct{}) (Token, error) {
	switch {
	case t.closed.Load():
		return nil, errors.ErrClosed
	case isCancelled(abort):
		return nil, context.Canceled
	default:
		// a free token means nobody is waiting, so taking it
		// straight away keeps the arrival order.
		if tok, ok := t.tryAcquire(); ok {
			return tok, nil
		}
		return t.wait(abort)
	}
}

// wait blocks until the token is free, the Turnstile closes, or abort
// closes.
func (t *Turnstile) wait(abort <-chan struct{}) (Token, error) {
	select {
	case tok := <-t.b.Acquire():
		return t.keep(tok, abort)
	case <-t.done.Signaled():
		return nil, errors.ErrClosed
	case <-abort:
		return nil, context.Canceled
	}
}

// tryAcquire takes the token if it is free and the Turnstile is open.
func (t *Turnstile) tryAcquire() (Token, bool) {
	tok, ok := t.b.TryAcquire()
	if !ok {
		return nil, false
	}

	tok, err := t.keep(tok, nil)
	return tok, err == nil
}

// keep returns a token just taken, unless the Turnstile or abort closed
// while it was being taken. In that case it gives the token back and
// returns [errors.ErrClosed] or [context.Canceled].
func (t *Turnstile) keep(tok Token, abort <-chan struct{}) (Token, error) {
	var err error
	switch {
	case t.closed.Load():
		err = errors.ErrClosed
	case isCancelled(abort):
		err = context.Canceled
	default:
		return tok, nil
	}

	t.b.Release(tok)
	return nil, err
}

func (t *Turnstile) doLock(abort <-chan struct{}) error {
	if err := t.lazyInit(); err != nil {
		return err
	}

	tok, err := t.acquire(abort)
	if err != nil {
		return err
	}

	t.held = tok
	return nil
}

func (t *Turnstile) doPass(abort <-chan struct{}) error {
	if err := t.lazyInit(); err != nil {
		return err
	}

	tok, err := t.acquire(abort)
	if err != nil {
		return err
	}

	t.b.Release(tok)
	return nil
}

// checkContext checks the receiver, then the context.
func (t *Turnstile) checkContext(ctx context.Context) error {
	switch {
	case t == nil:
		return errors.ErrNilReceiver
	case ctx == nil:
		return errors.ErrNilContext
	default:
		return nil
	}
}

// Close closes the Turnstile. Calls waiting to hold or pass it, and
// later ones, fail with [errors.ErrClosed]: LockContext, PassContext and
// PassAbort return it, and Lock and Pass panic with it, so code that may
// still be waiting when Close runs uses the forms that return an error.
// TryLock and TryPass report false. A current holder keeps it until
// Unlock. Closing a Turnstile that was never used initialises it first.
// Close returns [errors.ErrClosed] if the Turnstile is already closed,
// and [errors.ErrNilReceiver] if it is nil.
func (t *Turnstile) Close() error {
	if err := t.lazyInit(); err != nil {
		return err
	}
	if !t.closed.CompareAndSwap(false, true) {
		return errors.ErrClosed
	}
	close(t.done)
	return nil
}

// Lock holds the Turnstile, waiting until it is free.
// It panics if the Turnstile is nil or closed.
func (t *Turnstile) Lock() {
	if err := t.doLock(nil); err != nil {
		core.PanicFrom(1, err)
	}
}

// LockContext holds the Turnstile, waiting until it is free or the
// context is cancelled. It returns the context's error if cancelled,
// [errors.ErrClosed] if the Turnstile is closed, [errors.ErrNilContext]
// if ctx is nil, and [errors.ErrNilReceiver] if the Turnstile is nil.
func (t *Turnstile) LockContext(ctx context.Context) error {
	if err := t.checkContext(ctx); err != nil {
		return err
	}

	err := t.doLock(ctx.Done())
	if errors.Is(err, context.Canceled) {
		return ctx.Err()
	}
	return err
}

// TryLock holds the Turnstile if it is free, and reports whether it
// did. It returns false if the Turnstile is closed, and panics if it
// is nil.
func (t *Turnstile) TryLock() bool {
	if err := t.lazyInit(); err != nil {
		core.PanicFrom(1, err)
	}

	tok, ok := t.tryAcquire()
	if ok {
		t.held = tok
	}
	return ok
}

// Unlock releases a Turnstile held by Lock, LockContext or TryLock,
// closed or not. It panics with [errors.ErrNotLocked] if the Turnstile is
// not held, and with [errors.ErrNilReceiver] if it is nil.
func (t *Turnstile) Unlock() {
	if err := t.lazyInit(); err != nil {
		core.PanicFrom(1, err)
	}

	tok := t.held
	if tok == nil {
		core.PanicFrom(1, errors.ErrNotLocked)
	}

	t.held = nil
	t.b.Release(tok)
}

// Pass waits until the Turnstile is free and goes through without
// holding it. It panics if the Turnstile is nil or closed.
func (t *Turnstile) Pass() {
	if err := t.doPass(nil); err != nil {
		core.PanicFrom(1, err)
	}
}

// PassAbort waits until the Turnstile is free or abort closes, and goes
// through without holding it. It returns [context.Canceled] if aborted,
// [errors.ErrClosed] if the Turnstile is closed, and
// [errors.ErrNilReceiver] if it is nil. A nil abort waits indefinitely.
func (t *Turnstile) PassAbort(abort <-chan struct{}) error {
	return t.doPass(abort)
}

// PassContext waits until the Turnstile is free or the context is
// cancelled, and goes through without holding it. It returns the
// context's error if cancelled, [errors.ErrClosed] if the Turnstile is
// closed, [errors.ErrNilContext] if ctx is nil, and
// [errors.ErrNilReceiver] if the Turnstile is nil.
func (t *Turnstile) PassContext(ctx context.Context) error {
	if err := t.checkContext(ctx); err != nil {
		return err
	}

	err := t.doPass(ctx.Done())
	if errors.Is(err, context.Canceled) {
		return ctx.Err()
	}
	return err
}

// TryPass goes through the Turnstile if it is free, and reports whether
// it did. It returns false if the Turnstile is closed, and panics if it
// is nil.
func (t *Turnstile) TryPass() bool {
	if err := t.lazyInit(); err != nil {
		core.PanicFrom(1, err)
	}

	tok, ok := t.tryAcquire()
	if ok {
		t.b.Release(tok)
	}
	return ok
}

var (
	_ sync.Locker        = (*Turnstile)(nil)
	_ mutex.Mutex        = (*Turnstile)(nil)
	_ mutex.MutexContext = (*Turnstile)(nil)
)
