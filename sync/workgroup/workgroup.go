// Package workgroup provides concurrent task management and synchronisation
// for coordinating multiple operations within a shared lifecycle.
//
// The workgroup package is useful for scenarios where you need to:
//   - Manage a collection of goroutines that should be treated as a unit
//   - Propagate cancellation signals to all concurrent tasks
//   - Coordinate graceful shutdown of concurrent operations
//   - Track completion of multiple concurrent tasks
//   - Handle errors across concurrent operations
//
// Unlike sync.WaitGroup, this implementation provides context integration,
// cancellation propagation, and lifecycle management for concurrent operations.
package workgroup

import (
	"context"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
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

	ctx context.Context

	// OnCancel is invoked once when the Group transitions to the
	// cancelled state, whether through an explicit Cancel or Close or
	// through cancellation of the parent context. The handler runs in
	// its own goroutine and is tracked as a task: Wait, Done, and Close
	// all block until it returns.
	//
	// Assign OnCancel before the Group can be cancelled — typically
	// right after New and before launching tasks. A handler installed
	// after the Group's context is already done is not guaranteed to
	// run, as the cancellation watcher may have fired already.
	//
	// When the handler runs, the Group's context is already
	// cancelled — ctx.Err() is non-nil and context.Cause(ctx)
	// returns the cancellation cause. For cleanup work that needs
	// a live context, derive from wg.Parent or detach via
	// context.WithoutCancel(ctx); contexts derived from ctx are
	// born cancelled.
	//
	// cause carries the error passed to Cancel (context.Canceled for a
	// nil cause), or context.Cause(parent) when the parent context is
	// cancelled. The handler is invoked at most once per Group lifetime
	// regardless of how many cancellation signals arrive.
	OnCancel func(context.Context, error)

	cancel context.CancelCauseFunc
	doneCh chan struct{}
	tasks  cond.CountZero

	// non-pointer fields kept last so the GC pointer scan stops early
	mu        sync.RWMutex
	cancelled atomic.Bool
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

// Done returns a channel that is closed when all tasks in the Group
// have completed, including any OnCancel handler. This method will
// panic if called on a nil Group.
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

		wg.waitTasks()
	}()
	return ch
}

// Wait blocks until all tasks in the Group have completed.
// If the Group is nil, it returns ErrNilReceiver.
//
// "All tasks" includes the OnCancel handler. If a concurrent
// cancellation — an explicit Cancel or the parent context — returns
// before Wait returns, the OnCancel handler is guaranteed to have
// completed by the time Wait returns. A cancellation that begins after
// Wait has already returned is not awaited; callers that need the
// handler to drain in that case should use Close().
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

	wg.waitTasks()

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
// If OnCancel is set, the handler is enrolled as a task and Cancel
// blocks only until the handler goroutine has started — it does not
// wait for the handler to finish. Wait, Done, and Close await the
// handler's completion.
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
	if cause == nil {
		cause = context.Canceled
	}

	return wg.markCancelled(cause)
}

// onContextDone is registered via context.AfterFunc and fires once the
// Group's context is done by any cause. For an explicit Cancel or Close the
// transition has already happened and this is a no-op; for a parent-context
// cancellation it performs the same transition so the OnCancel handler runs
// regardless of the cancellation source. The cause is read from the
// now-cancelled context.
func (wg *Group) onContextDone() {
	wg.markCancelled(context.Cause(wg.ctx))
}

// markCancelled performs the one-shot transition to the cancelled state: it
// records the cancellation, cancels the context (a no-op once a parent
// cancellation already closed it), and spawns the OnCancel handler, blocking
// until that handler's goroutine has started. It returns true only for the
// caller that won the transition.
func (wg *Group) markCancelled(cause error) bool {
	if wg.cancelled.Load() {
		return false
	}

	// RW
	wg.mu.Lock()

	// Re-check under mu: between the lock-free check above and acquiring
	// the lock a concurrent Cancel — or the AfterFunc watcher reacting to
	// the same context cancellation — may have won and stored cancelled.
	// Without this guard a losing caller would store cancelled again,
	// overwrite the cause, and spawn a second OnCancel handler.
	// TestGroup_ConcurrentCancelOnceOnly exercises this loser path.
	if wg.cancelled.Load() {
		wg.mu.Unlock()
		return false
	}

	// Propagate cancellation before spawning the OnCancel handler so
	// the handler observes wg.ctx in its cancelled state. doGo's
	// cancelled-check + Inc fence is unaffected — both stores
	// complete before this Lock releases.
	wg.cancelled.Store(true)
	wg.cancel(cause)

	// Spawn the handler with the context's authoritative cause rather
	// than the requested one. When a parent cancellation already closed
	// the context, wg.cancel above was a no-op and the requested cause
	// never took effect; reading it back keeps the handler's cause
	// consistent with context.Cause(wg.ctx) regardless of which path —
	// this caller or a task cancelling on its own error — won the
	// transition.
	ready := wg.spawnCancelHandler(context.Cause(wg.ctx))

	wg.mu.Unlock()

	if ready != nil {
		// wait until onCancel has started
		<-ready
	}

	return true
}

// spawnCancelHandler enrols the OnCancel handler as a counted task and runs
// it detached, returning a channel closed once the handler goroutine has
// started, or nil when no handler is set. The caller must hold wg.mu.
func (wg *Group) spawnCancelHandler(cause error) chan struct{} {
	fn := wg.OnCancel
	if fn == nil {
		return nil
	}

	ready := make(chan struct{})
	wg.tasks.Inc()
	go func() {
		defer wg.tasks.Dec()
		close(ready)
		// Contain a panicking handler. Without this, a panic unwinds
		// the detached goroutine with nothing to recover it: it crashes
		// the process while Wait blocks forever on the tasks counter.
		// run() wraps ordinary tasks the same way. The caught error is
		// deliberately dropped: the group is already cancelled, so
		// routing it back through Cancel would be a no-op. Surfacing it
		// instead would require threading a cancellation cause out of
		// the handler, which the Group does not yet support.
		_ = core.Catch(func() error {
			fn(wg.ctx, cause)
			return nil
		})
	}()
	return ready
}

// Close cancels the Group and waits for all tasks to complete,
// including any OnCancel handler. It returns an error if called on
// a nil Group.
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

	// doCancel has set cancelled before returning, so doGo can no longer
	// enrol new tasks: a plain drain suffices here and the waitTasks fence
	// (which guards against a concurrent enrol) is unnecessary.
	wg.doCancel(context.Canceled)
	// tasks.Wait returns context.Canceled if a prior Close already closed
	// the counter; the count is 0 either way, so the discard is safe.
	_ = wg.tasks.Wait()
	// tasks.Close returns ErrClosed on the second and later calls; that is
	// expected and safe to ignore.
	_ = wg.tasks.Close()
	return nil
}

// Go spawns a new goroutine to execute the provided function.
//
// The function receives the Group's context, which will be cancelled
// when the Group is cancelled. This allows the function to respond to
// cancellation appropriately.
//
// If fn is nil, no goroutine is started and Go returns nil.
//
// Go is decisive against concurrent cancellation: if it returns nil,
// the goroutine has been tracked and will be awaited by Wait/Close/Done;
// if it returns ErrClosed, no goroutine was started. There is no
// in-between state where Go spawns a goroutine that subsequent waits
// fail to observe.
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
	if fn == nil {
		return nil
	}

	// RO: serialise the cancelled-check + Inc against doCancel's Lock so
	// the bump establishes a happens-before relationship with any
	// concurrent Wait fence. doGo calls remain mutually concurrent.
	wg.mu.RLock()
	if wg.cancelled.Load() {
		wg.mu.RUnlock()
		return errors.ErrClosed
	}
	wg.tasks.Inc()
	wg.mu.RUnlock()

	go func() {
		defer wg.tasks.Dec()

		fn(wg.ctx)
	}()
	return nil
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

// waitTasks blocks until the tasks counter has stably reached zero across
// a serialisation point. doCancel and doGo call tasks.Inc while holding
// wg.mu (Lock and RLock respectively); the Lock acquired here therefore
// observes any in-flight Inc. If a concurrent doCancel or doGo bumped
// the counter while we were draining, the post-Lock Value read is
// non-zero and we loop back to drain again. Without this fence, Wait
// could return after the counter momentarily hit zero but before a
// concurrent Cancel scheduled OnCancel or a concurrent Go enrolled a
// new task.
func (wg *Group) waitTasks() {
	for {
		// tasks.Wait returns context.Canceled only once the underlying
		// Barrier is closed, which happens solely via tasks.Close in
		// Close; the subsequent Value read is 0 and the loop exits.
		_ = wg.tasks.Wait()
		wg.mu.Lock()
		v := wg.tasks.Value()
		wg.mu.Unlock()
		if v == 0 {
			return
		}
	}
}

// init initialises the Group with a context and cancel function.
// If Parent is nil, it uses context.Background() as the default parent.
func (wg *Group) init() {
	if wg.Parent == nil {
		wg.Parent = context.Background()
	}

	wg.ctx, wg.cancel = context.WithCancelCause(wg.Parent)
	core.MustNoError(wg.tasks.Init(0))

	// Drive the OnCancel transition on any context cancellation, the
	// parent's included, not only an explicit Cancel or Close. The stop
	// func is unneeded: the watcher fires at most once and deregisters
	// itself once the context is done.
	context.AfterFunc(wg.ctx, wg.onContextDone)
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
		wg.init()
	}

	return nil
}

// New creates a new Group with the given parent context.
// The Group is initialised and ready for use.
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
	wg := &Group{Parent: ctx}
	wg.init()
	return wg
}
