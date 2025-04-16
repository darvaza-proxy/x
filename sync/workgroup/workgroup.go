// Package workgroup provides concurrent task management and synchronisation
// for coordinating multiple operations within a shared lifecycle.
//
// The workgroup package is useful for scenarios where you need to:
//   - Manage a collection of goroutines that should be treated as a unit
//   - Propagate cancellation signals to all concurrent tasks
//   - Coordinate graceful shutdown of concurrent operations
//   - Track completion of multiple concurrent tasks
//   - Handle errors across concurrent operations
//   - Control the maximum number of concurrent tasks
//
// Unlike sync.WaitGroup, this implementation provides context integration,
// cancellation propagation, and lifecycle management for concurrent operations.
package workgroup

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// Group manages a collection of concurrent tasks with cancellation and
// synchronisation. It coordinates and controls multiple concurrent operations
// within a shared lifecycle.
//
// A Group is safe for concurrent use, allowing tasks to be added, monitored,
// and cancelled from multiple goroutines simultaneously.
//
// Typical usage pattern:
//
//	wg := workgroup.New(ctx)
//	defer wg.Close()
//
//	wg.Go(func(ctx context.Context) {
//	    // Task 1 implementation with cancellation handling
//	})
//
//	wg.Go(func(ctx context.Context) {
//	    // Task 2 implementation with cancellation handling
//	})
//
//	// Wait for all tasks to complete or context to be cancelled
//	if err := wg.Wait(); err != nil {
//	    // Handle error
//	}
type Group struct {
	// Parent is the parent context for the group. If nil during initialisation,
	// context.Background() will be used as the default parent.
	Parent context.Context

	// Limit specifies the maximum number of concurrent tasks that can be
	// executed simultaneously in the Group.
	// If set to zero or a negative value, no limit is applied, allowing
	// unlimited concurrent tasks. This is useful for controlling resource
	// usage or preventing excessive concurrency.
	// The Limit value is fixed after the Group is created via New() or
	// NewLimited() and cannot be changed later.
	Limit int

	// OnCancel is called when the Group is cancelled if defined.
	OnCancel func(context.Context, error)

	ctx       context.Context
	cancel    context.CancelCauseFunc
	cancelled atomic.Bool
	mu        sync.RWMutex
	wg        WaitGroup
	doneCh    chan struct{}
}

// Context returns the context associated with the Group.
// This context is used for cancellation and deadline management.
//
// The returned context will be cancelled when the Group is cancelled
// either explicitly via Cancel() or through its parent context.
//
// This method will panic if called on a nil Group.
//
// Tasks launched via Go() should use this context to respond to cancellation
// signals and respect the Group's lifecycle.
func (wg *Group) Context() context.Context {
	if err := wg.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return wg.ctx
}

// Count returns the number of active tasks in the Group.
// If initialisation fails or the receiver is nil, it returns 0.
// This method is safe to call on a Group and provides a snapshot
// of current task count.
func (wg *Group) Count() int {
	if err := wg.lazyInit(); err != nil {
		return 0
	}

	return wg.wg.Count()
}

// Err returns the cancellation cause, if any.
// If the Group is nil, it returns ErrNilReceiver.
//
// For a cancelled Group, it returns the error provided to Cancel,
// or context.Canceled if no specific error was provided.
// For an active Group, it returns nil.
//
// This method helps determine whether the Group was cancelled and why.
func (wg *Group) Err() error {
	if err := wg.lazyInit(); err != nil {
		return err
	}

	return context.Cause(wg.ctx)
}

// IsCancelled reports whether the Group has been cancelled.
// It returns true if Cancel has been called or the parent context
// has been cancelled, otherwise false.
//
// This method will panic if called on a nil Group.
//
// This is a convenience method that checks the context's error state
// and is equivalent to checking whether Context().Err() is non-nil.
func (wg *Group) IsCancelled() bool {
	if err := wg.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return wg.ctx.Err() != nil
}

// Cancelled returns a channel that is closed when the Group is cancelled.
// This allows waiting for or detecting cancellation using a select statement.
//
// This method will panic if called on a nil Group.
//
// Typical usage in a select statement:
//
//	select {
//	case <-wg.Cancelled():
//	    // Group was cancelled, take appropriate action
//	case <-otherChannel:
//	    // Handle other event
//	}
func (wg *Group) Cancelled() <-chan struct{} {
	if err := wg.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return wg.ctx.Done()
}

// Done returns a channel that is closed when all tasks in the Group have
// completed. This method will panic if called on a nil Group.
//
// The channel is created once and reused until the Group is emptied.
// If not cancelled, the Group can be reused after being emptied and a new
// Done channel will be created.
//
// This method is useful for waiting on task completion with select statements:
//
//	select {
//	case <-wg.Done():
//	    // All tasks have completed
//	case <-timeout:
//	    // Timed out waiting for tasks
//	    wg.Cancel(ErrTimeout)
//	}
func (wg *Group) Done() <-chan struct{} {
	if err := wg.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return wg.doDone()
}

func (wg *Group) doDone() <-chan struct{} {
	wg.mu.Lock()
	if ch := wg.doneCh; ch != nil {
		// reused
		wg.mu.Unlock()
		return ch
	}

	// fresh watcher
	ch := make(chan struct{})
	wg.doneCh = ch
	wg.mu.Unlock()

	go func() {
		defer close(ch)

		defer func() {
			// forget
			wg.mu.Lock()
			wg.doneCh = nil
			wg.mu.Unlock()
		}()

		_ = wg.wg.Wait()
	}()
	return ch
}

// Wait blocks until all tasks in the Group have completed.
// If the Group is nil, it returns ErrNilReceiver.
//
// Wait returns an error only if the Group was cancelled with a cause other
// than context.Canceled. If cancelled with context.Canceled or with Cancel()
// without a specific error, Wait returns nil.
//
// This method provides a synchronous alternative to using the Done channel.
//
// Note that Wait() only reports errors from cancellation, not from the tasks
// themselves. For error collection from tasks, implement your own mechanism.
func (wg *Group) Wait() error {
	if err := wg.lazyInit(); err != nil {
		return err
	}

	_ = wg.wg.Wait()

	err := context.Cause(wg.ctx)
	if err == context.Canceled {
		err = nil
	}
	return err
}

// Cancel attempts to cancel the Group with an optional error cause.
// If called on a nil Group, it will panic.
//
// Cancel propagates the cancellation to all tasks in the Group
// through the context returned by the Context() method.
//
// It returns true if this was the first cancellation, false if the Group
// was already cancelled.
//
// If cause is nil, context.Canceled will be used instead.
//
// Example:
//
//	if someCondition {
//	    wg.Cancel(fmt.Errorf("operation failed: %w", someError))
//	}
func (wg *Group) Cancel(cause error) bool {
	if err := wg.lazyInit(); err != nil {
		core.Panic(core.NewPanicError(1, err))
	}

	return wg.doCancel(cause)
}

func (wg *Group) doCancel(cause error) bool {
	var ready chan struct{}

	if cause == nil {
		cause = context.Canceled
	}

	if wg.cancelled.Load() {
		return false
	}

	// RW
	wg.mu.Lock()

	if wg.cancelled.Load() {
		wg.mu.Unlock()
		return false
	}

	// call the OnCancel function if defined
	if fn := wg.OnCancel; fn != nil {
		ready = make(chan struct{})

		// Offload the OnCancel function to a goroutine
		// to avoid blocking the critical section.
		// It skips the queue if a limiter was set.
		_ = wg.wg.Go(func() {
			close(ready)
			fn(wg.ctx, cause)
		})
	}

	wg.cancelled.Store(true)
	wg.cancel(cause)
	wg.mu.Unlock()

	if ready != nil {
		// wait until onCancel has started
		<-ready
	}

	_ = wg.wg.Close()

	return true
}

// Close cancels the Group and waits for all tasks to complete.
// It returns an error if called on a nil Group.
//
// Close ensures all resources are freed by first cancelling any
// running tasks and then waiting for their completion. This is useful
// in defer statements to ensure proper cleanup.
//
// Even if the Group was already cancelled by a parent context or
// via Cancel(), Close() ensures all tasks have completed.
//
// Recommended usage in defer statements:
//
//	wg := workgroup.New(ctx)
//	defer wg.Close()
func (wg *Group) Close() error {
	if err := wg.lazyInit(); err != nil {
		return err
	}

	wg.doCancel(context.Canceled)
	_ = wg.wg.Wait()
	return nil
}

// Go spawns a new goroutine to execute the provided function.
//
// The function receives the Group's context, which will be cancelled
// when the Group is cancelled. This allows the function to respond to
// cancellation appropriately.
//
// If fn is nil, no goroutine is started.
//
// The Group automatically tracks the lifetime of the spawned goroutine
// and will wait for its completion in Wait(), Close(), or through the
// Done() channel.
//
// Tasks should monitor the provided context for cancellation:
//
//	wg.Go(func(ctx context.Context) {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            // Clean up and return when cancelled
//	            return
//	        case <-ticker.C:
//	            // Do periodic work
//	        }
//	    }
//	})
func (wg *Group) Go(fn func(context.Context)) error {
	if err := wg.lazyInit(); err != nil {
		return err
	}
	return wg.doGo(fn)
}

func (wg *Group) doGo(fn func(context.Context)) error {
	switch {
	case fn == nil:
		return nil
	case wg.cancelled.Load():
		return errors.ErrClosed
	default:
		return wg.wg.Go(func() {
			fn(wg.ctx)
		})
	}
}

// GoCatch spawns a new goroutine with error handling and supervision.
// If called on a nil Group, it will panic.
//
// The function receives the Group's context and can return an error. Panics
// occurring in fn are captured and converted to core.PanicError{} values.
//
// The catch handler, if provided, can process, transform, or filter the error.
// If the final error (after catch) is non-nil, the entire Group will be
// cancelled with that error. To prevent Group cancellation, the catch handler
// can return nil.
//
// If fn is nil, no goroutine is started.
//
// The Group tracks the lifetime of the spawned goroutine and waits for its
// completion in Wait(), Close(), or through the Done() channel.
//
// Example usage with error handling:
//
//	wg.GoCatch(
//	    func(ctx context.Context) error {
//	        // Task implementation that may return errors
//	        return doSomething(ctx)
//	    },
//	    func(ctx context.Context, err error) error {
//	        // Process the error
//	        if isRecoverable(err) {
//	            logError(err)
//	            return nil // Prevent Group cancellation for recoverable errors
//	        }
//	        return fmt.Errorf("critical task failure: %w", err)
//	    },
//	)
func (wg *Group) GoCatch(fn func(context.Context) error, catch func(context.Context, error) error) error {
	err := wg.lazyInit()
	switch {
	case err != nil:
		return err
	case fn == nil:
		return nil
	default:
		return wg.doGo(func(_ context.Context) {
			wg.run(fn, catch)
		})
	}
}

func (wg *Group) run(fn func(context.Context) error, catch func(context.Context, error) error) {
	err := core.Catch(func() error {
		// execute the function
		return fn(wg.ctx)
	})

	if catch != nil {
		// process the exit condition
		err = core.Catch(func() error {
			return catch(wg.ctx, err)
		})
	}

	// cancel the group if the resulting error is non-nil
	if err != nil {
		wg.Cancel(err)
	}
}

// init initialises the Group with a context and cancel function.
// If Parent is nil, it uses context.Background() as the default parent.
func (wg *Group) init() error {
	if wg.Parent == nil {
		wg.Parent = context.Background()
	}

	wg.ctx, wg.cancel = context.WithCancelCause(wg.Parent)

	if wg.Limit > 0 {
		wg.wg = MustLimiter(wg.Limit)
	} else {
		wg.wg = NewRunner()
	}

	return nil
}

// lazyInit ensures the Group is properly initialised before use.
// Returns an error if the receiver is nil, otherwise initialises
// the Group if needed and returns nil.
func (wg *Group) lazyInit() error {
	if wg == nil {
		return errors.ErrNilReceiver
	}

	// RO
	wg.mu.RLock()
	ready := wg.ctx != nil
	wg.mu.RUnlock()

	if ready {
		return nil
	}

	// RW
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.ctx == nil {
		return wg.init()
	}

	return nil
}

// New creates a new Group with the given parent context and no concurrency
// limit. This is equivalent to calling NewLimited(ctx, 0).
//
// If ctx is nil, context.Background() will be used as the default parent.
//
// The returned Group can be used to spawn and manage concurrent tasks
// that share the same lifecycle and cancellation mechanism.
//
// Example:
//
//	// Create a workgroup with a timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	wg := workgroup.New(ctx)
//	defer wg.Close()
//
//	// Add tasks to the workgroup
//	wg.Go(func(ctx context.Context) { ... })
//	wg.Go(func(ctx context.Context) { ... })
func New(ctx context.Context) *Group {
	return NewLimited(ctx, 0)
}

// NewLimited creates a new Group with the given parent context and concurrency
// limit. The limit parameter sets the maximum number of concurrent tasks,
// with a value of zero or less meaning unlimited concurrency.
//
// If ctx is nil, context.Background() will be used as the default parent.
//
// The limit is fixed upon creation and cannot be changed afterwards.
//
// Example:
//
//	// Create a workgroup with a timeout and concurrency limit of 10
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	wg := workgroup.NewLimited(ctx, 10)
//	defer wg.Close()
//
//	// Add tasks to the workgroup
//	wg.Go(func(ctx context.Context) { ... })
func NewLimited(ctx context.Context, limit int) *Group {
	wg := &Group{Parent: ctx, Limit: limit}
	_ = wg.init()

	runtime.SetFinalizer(wg, func(wg *Group) {
		_ = wg.Close()
	})

	return wg
}
