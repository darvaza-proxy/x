package workgroup

import (
	"container/list"
	"runtime"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
)

var _ WaitGroup = (*Limiter)(nil)

// Limiter implements the WaitGroup interface with an added constraint on
// the maximum number of concurrently executing goroutines.
//
// It maintains a queue of pending functions when the concurrent limit
// is reached. These functions will be executed as running goroutines complete.
// Once closed, a Limiter will not accept new goroutines, and any resources
// associated with it will be released when all active goroutines complete.
type Limiter struct {
	mu     sync.Mutex
	count  cond.CountZero
	limit  int
	closed atomic.Bool

	barrier chan struct{}
	list    list.List
}

// check verifies if the Limiter is properly initialised.
func (g *Limiter) check() error {
	switch {
	case g == nil:
		return errors.ErrNilReceiver
	case g.barrier == nil:
		return errors.ErrNotInitialised
	default:
		return nil
	}
}

// IsNil reports whether the Limiter is nil or uninitialised.
func (g *Limiter) IsNil() bool {
	return g.check() != nil
}

// IsClosed reports whether the Limiter has been closed.
func (g *Limiter) IsClosed() bool {
	if g.check() != nil {
		return true
	}
	return g.closed.Load()
}

// Size returns the maximum number of goroutines that can run concurrently.
func (g *Limiter) Size() int {
	if g.check() != nil {
		return 0
	}
	return g.limit
}

// Count returns the number of currently active goroutines.
func (g *Limiter) Count() int {
	if g.check() != nil {
		return 0
	}
	return g.count.Value()
}

// Len returns the total number of active and queued goroutines.
func (g *Limiter) Len() int {
	if g.check() != nil {
		return 0
	}

	g.mu.Lock()
	n := g.count.Value() + g.list.Len()
	g.mu.Unlock()
	return n
}

// Wait blocks until all goroutines complete.
// Returns an error if the Limiter is not properly initialised.
func (g *Limiter) Wait() error {
	if err := g.check(); err != nil {
		return err
	}

	return g.count.Wait()
}

// Close prevents spawning new goroutines and waits for all existing
// goroutines to complete.
// Returns an error if the Limiter is not properly initialised or already closed.
func (g *Limiter) Close() error {
	err := g.check()
	switch {
	case err != nil:
		return err
	case !g.closed.CompareAndSwap(false, true):
		return errors.ErrClosed
	default:
		return g.doClose()
	}
}

// doClose handles the internal closing logic for the Limiter.
// It cleans up resources when all goroutines have completed.
func (g *Limiter) doClose() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.count.Value()+g.list.Len() == 0 {
		// empty, clean-up now
		g.count.Close()
	}
	close(g.barrier)
	return nil
}

// Go spawns a new goroutine to execute the provided function, respecting
// the concurrency limit.
// If the concurrency limit has been reached, the function is queued for
// later execution.
// Returns an error if the Limiter is closed or otherwise invalid.
// A nil function will be ignored without error.
func (g *Limiter) Go(fn func()) error {
	err := g.check()
	switch {
	case err != nil:
		return err
	case g.closed.Load():
		return errors.ErrClosed
	case fn == nil:
		return nil
	default:
		g.doGo(fn)
		return nil
	}
}

// doGo handles the actual spawning or queueing of a new goroutine.
func (g *Limiter) doGo(fn func()) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.list.Len() == 0 {
		// first
		select {
		case _, ok := <-g.barrier:
			if ok {
				// immediate
				g.count.Inc()
				g.spawn(fn)
				return
			}
		default:
		}
	}

	// enqueue
	g.unsafePush(fn)
}

// spawn starts a new goroutine to execute the provided function.
func (g *Limiter) spawn(fn func()) {
	go func() {
		defer g.spawnDone()
		fn()
	}()
}

// spawnDone handles the completion of a goroutine, potentially starting
// a queued function if one exists.
func (g *Limiter) spawnDone() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if fn, ok := g.unsafePop(); ok {
		// spawn next using the current token
		g.spawn(fn)
		return
	}

	closed := g.closed.Load()

	if !closed {
		// release the token unless closed.
		g.barrier <- struct{}{}
	}

	// and reduce the count
	n := g.count.Dec()
	if n == 0 && closed {
		// last after closing, clean-up
		g.count.Close()
	}
}

// Init initialises an uninitialised Limiter with the specified worker limit.
// Returns an error if the Limiter is nil, already initialised, or if limit is
// less than 1.
func (g *Limiter) Init(limit int) error {
	switch {
	case g == nil:
		return errors.ErrNilReceiver
	case g.barrier != nil:
		return errors.ErrAlreadyInitialised
	default:
		// RW
		g.mu.Lock()
		defer g.mu.Unlock()

		if g.barrier != nil {
			return errors.ErrAlreadyInitialised
		}

		return g.doInit(limit)
	}
}

// init handles the internal initialisation of the Limiter.
func (g *Limiter) doInit(limit int) error {
	if limit < 1 {
		return core.Wrap(core.ErrInvalid, "limit")
	}

	g.limit = limit
	g.barrier = make(chan struct{}, limit)
	for range limit {
		g.barrier <- struct{}{}
	}
	_ = g.count.Init(0)
	return nil
}

// NewLimiter creates and initializes a new Limiter with the specified worker limit.
// It returns the initialized Limiter and any error encountered during initialization.
// If limit is less than 1, it returns nil and an error.
func NewLimiter(limit int) (*Limiter, error) {
	g := new(Limiter)
	if err := g.doInit(limit); err != nil {
		return nil, err
	}

	runtime.SetFinalizer(g, func(g *Limiter) {
		_ = g.Close()
	})

	return g, nil
}

// MustLimiter creates a new Limiter with the specified worker limit and panics
// if initialization fails.
// It is a convenience wrapper around NewLimiter that simplifies creating a Limiter
// when an error is unacceptable.
func MustLimiter(limit int) *Limiter {
	g, err := NewLimiter(limit)
	if err != nil {
		core.Panic(core.NewPanicError(1, err))
	}
	return g
}

// unsafePush adds a function to the queue.
// The caller must hold g.mu.
func (g *Limiter) unsafePush(fn func()) {
	g.list.PushBack(fn)
}

// unsafePop removes and returns the first function from the queue.
// The caller must hold g.mu.
// It returns the function and a boolean indicating success.
func (g *Limiter) unsafePop() (func(), bool) {
	if el := g.list.Front(); el != nil {
		g.list.Remove(el)

		if fn, ok := el.Value.(func()); ok {
			return fn, true
		}
	}

	return nil, false
}
