package cond

import (
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// Barrier provides a synchronisation mechanism to coordinate goroutines.
// It manages a reusable token that can be used to signal state changes
// and coordinate access to shared resources.
// Barrier is primarily designed to be used by other synchronisation
// primitives internally.
type Barrier struct {
	_ sync.Mutex // prevent copies

	closed bool

	b chan Token
}

// NewBarrier creates and initialises a new Barrier, panicking if
// initialisation fails. It returns a fully initialised Barrier ready for use.
func NewBarrier() *Barrier {
	c := &Barrier{}
	if err := c.Init(); err != nil {
		core.Panic(err)
	}
	return c
}

// IsNil reports whether the Barrier or its underlying channel is nil.
// This is used for lazy initialisation in higher-level primitives.
func (bs *Barrier) IsNil() bool {
	return bs == nil || bs.b == nil
}

// IsClosed reports whether the Barrier is no longer usable.
func (bs *Barrier) IsClosed() bool {
	return bs == nil || bs.b == nil || bs.closed
}

// Init initialises the barrier by creating a channel with a capacity
// of 1 and placing a new Token in it. This should be called before
// any other methods if the barrier was not properly initialised.
func (bs *Barrier) Init() error {
	switch {
	case bs == nil:
		return errors.ErrNilReceiver
	case bs.b != nil:
		return errors.ErrAlreadyInitialised
	default:
		b := make(chan Token, 1)
		b <- make(Token)
		bs.b = b
		return nil
	}
}

// Close terminates the Barrier by closing its underlying channel and the
// current Token. It returns an error if the Barrier is nil, not initialised
// or already closed. After calling Close, the Barrier cannot be used again.
func (bs *Barrier) Close() error {
	switch {
	case bs == nil:
		return errors.ErrNilReceiver
	case bs.b == nil:
		return errors.ErrNotInitialised
	case bs.closed:
		return errors.ErrClosed
	default:
		// lock
		b, ok := <-bs.b
		if !ok {
			return errors.ErrClosed
		}

		bs.closed = true
		close(b)
		close(bs.b)
		return nil
	}
}

// Broadcast closes the current Token (signalling all waiters) and creates
// a new Token. This is typically called when a watched condition is met,
// to wake up all waiting goroutines and prepare for the next wait cycle.
func (bs *Barrier) Broadcast() {
	t, ok := <-bs.b
	if ok {
		// omit when closed.
		close(t)
		bs.b <- make(Token)
	}
}

// Signal attempts to signal the current Token, waking up a single waiting
// goroutine. It returns true if a goroutine was successfully signalled,
// and false otherwise. The method ensures the Token is returned to the
// Barrier after signalling.
func (bs *Barrier) Signal() bool {
	b, ok := <-bs.b
	if !ok {
		// closed
		return false
	}
	signaled := b.Signal()
	bs.b <- b
	return signaled
}

// Acquire returns a receive-only channel for obtaining the current Token.
// When a Token is received, the caller has exclusive access until it calls
// Release with the same Token.
func (bs *Barrier) Acquire() <-chan Token {
	return bs.b
}

// TryAcquire attempts to acquire the token without blocking.
// Returns the token and true if successful, nil and false otherwise.
func (bs *Barrier) TryAcquire() (Token, bool) {
	select {
	case t, ok := <-bs.b:
		return t, ok
	default:
		return nil, false
	}
}

// Release returns the Token to the barrier, allowing other goroutines
// to acquire it. This should always be called after Acquire to maintain
// proper barrier operation. It can be safely called with a nil Token.
func (bs *Barrier) Release(b Token) {
	if b != nil && !bs.closed {
		bs.b <- b
	}
}

// Token retrieves the current Token without removing it from the barrier.
// This provides access to the Token for waiting on state changes without
// exclusive access, used to provide a select-friendly channel that will be
// closed when a condition is met.
func (bs *Barrier) Token() Token {
	b, ok := <-bs.b
	if !ok {
		// closed
		return nil
	}
	bs.b <- b
	return b
}

// Signaled returns a channel that can be used to wait for the barrier's
// completion. It provides a select-friendly way to wait for a waiter to be
// signalled.
func (bs *Barrier) Signaled() <-chan struct{} {
	return bs.Token()
}

// Wait blocks until the barrier's current condition is signalled.
// It provides a simple way to wait for the barrier to complete its current
// state.
func (bs *Barrier) Wait() {
	<-bs.Signaled()
}

// Token is a channel of empty structs that serves as a signalling mechanism.
// When closed, any goroutines waiting on the Token will be unblocked.
// This type is used for efficient signalling with minimal memory overhead.
type Token chan struct{}

// Signaled returns the Token as a channel that can be used to wait for the
// Token's completion. It provides a select-friendly way to wait for the Token
// to be signalled or closed.
func (t Token) Signaled() <-chan struct{} {
	return t
}

// Wait blocks until the Token is closed, which happens when the condition
// being monitored (like a counter reaching zero) is satisfied.
func (t Token) Wait() {
	<-t
}

// Signal wakes up a single goroutine waiting on the Token, if there is one.
// It returns true if a goroutine was woken up, and false otherwise.
// It does not block if no goroutines are waiting.
func (t Token) Signal() bool {
	select {
	case t <- struct{}{}:
		return true
	default:
		return false
	}
}
