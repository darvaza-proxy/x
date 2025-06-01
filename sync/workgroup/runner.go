package workgroup

import (
	"runtime"
	"sync"
	"sync/atomic"

	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
)

var _ WaitGroup = (*Runner)(nil)

// Runner implements the WaitGroup interface, providing a mechanism to spawn
// and coordinate multiple goroutines.
//
// Once closed, a Runner will not accept new goroutines. Any resources
// associated with it will be released when all active goroutines complete.
type Runner struct {
	mu     sync.Mutex
	count  cond.CountZero
	closed atomic.Bool
}

// check verifies if the Runner is properly initialised.
func (g *Runner) check() error {
	switch {
	case g == nil:
		return errors.ErrNilReceiver
	case g.count.IsNil():
		return errors.ErrNotInitialised
	default:
		return nil
	}
}

// IsNil reports whether the Runner is nil or uninitialised.
func (g *Runner) IsNil() bool {
	return g.check() != nil
}

// IsClosed reports whether the Runner has been closed.
func (g *Runner) IsClosed() bool {
	if g.check() != nil {
		return true
	}
	return g.closed.Load()
}

// Count returns the number of active goroutines.
func (g *Runner) Count() int {
	if g.check() != nil {
		return 0
	}
	return g.count.Value()
}

// Wait blocks until all goroutines complete.
// Returns an error if the Runner is not properly initialised.
func (g *Runner) Wait() error {
	if err := g.check(); err != nil {
		return err
	}

	return g.count.Wait()
}

// Close prevents spawning new goroutines and waits for all existing
// goroutines to complete.
// Returns an error if the Runner is not properly initialised or already closed.
func (g *Runner) Close() error {
	err := g.check()
	switch {
	case err != nil:
		return err
	case !g.closed.CompareAndSwap(false, true):
		return errors.ErrClosed
	case g.count.Value() == 0:
		// empty, clean-up now
		return g.count.Close()
	default:
		// clean-up handled when the last finishes
		return nil
	}
}

// Go spawns a new goroutine to execute the provided function.
// Returns an error if the Runner is closed or otherwise invalid.
// A nil function will be ignored without error.
func (g *Runner) Go(fn func()) error {
	err := g.check()
	switch {
	case err != nil:
		return err
	case fn == nil:
		return nil
	default:
		return g.doGo(fn)
	}
}

// doGo handles the actual spawning of a new goroutine.
func (g *Runner) doGo(fn func()) error {
	g.count.Inc()

	if g.closed.Load() {
		// closed
		g.count.Dec()
		return errors.ErrClosed
	}

	go func() {
		defer g.doGoDone()
		fn()
	}()

	return nil
}

// doGoDone decrements the active goroutine counter and performs cleanup
// if it was the last goroutine and the Runner is closed.
func (g *Runner) doGoDone() {
	n := g.count.Dec()
	if n == 0 && g.closed.Load() {
		// last after closing
		g.count.Close()
	}
}

// Init initialises an uninitialised Runner.
// Returns an error if the Runner is nil or already initialised.
func (g *Runner) Init() error {
	if g == nil {
		return errors.ErrNilReceiver
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.count.IsNil() {
		return errors.ErrAlreadyInitialised
	}

	return g.doInit()
}

// init initialises the cond.CountZero field.
func (g *Runner) doInit() error {
	return g.count.Init(0)
}

// NewRunner creates and initialises a new Runner.
func NewRunner() *Runner {
	g := new(Runner)
	_ = g.doInit()

	runtime.SetFinalizer(g, func(g *Runner) {
		_ = g.Close()
	})

	return g
}
