// Package semaphore provides synchronisation primitives for controlling
// access to shared resources.
package semaphore

import (
	"context"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/mutex"
)

const (
	exclusiveLock = true
	readerLock    = false
)

// Semaphore provides a synchronisation primitive for controlling access to
// shared resources using a spinlock mechanism. It supports both exclusive and
// read locks with context-aware and blocking acquisition methods.
type Semaphore struct {
	mu sync.RWMutex

	// barrier controls access to the semaphore state.
	barrier cond.Barrier

	// global holds the state of the semaphore.
	// true if an exclusive lock is held,
	// false if a reader lock is held.
	global chan bool
	// readers holds the count of readers unless it's the first
	readers chan int
	// writers tracks if there are writers waiting
	writers cond.CountZero
	// active tracks locks sessions and waiters
	active cond.CountZero
}

func (s *Semaphore) lazyInit() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	// RO
	s.mu.RLock()
	if !s.barrier.IsNil() {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	// RW
	s.mu.Lock()
	if s.barrier.IsNil() {
		s.init()
	}
	s.mu.Unlock()
	return nil
}

func (s *Semaphore) init() {
	_ = s.active.Init(0)
	_ = s.writers.Init(0)

	s.global = make(chan bool, 1)
	s.readers = make(chan int, 1)
}

//revive:disable-next-line:flag-parameter
func (s *Semaphore) acquire(allowClosed bool) (cond.Token, error) {
	if err := s.lazyInit(); err != nil {
		return nil, err
	}

	t, ok := <-s.barrier.Acquire()
	if ok || allowClosed {
		return t, nil
	}
	return nil, errors.ErrClosed
}

//revive:disable-next-line:flag-parameter
func (s *Semaphore) acquireContext(ctx context.Context, allowClosedOrCancelled bool) (cond.Token, error) {
	err := s.lazyInit()
	switch {
	case err != nil:
		// invalid
		return nil, err
	case ctx == nil:
		// context is required
		return nil, errors.ErrNilContext
	case allowClosedOrCancelled:
		// no further checks
		t := <-s.barrier.Acquire()
		return t, nil
	default:
		t, ok := <-s.barrier.Acquire()
		if !ok {
			// closed
			return nil, errors.ErrClosed
		} else if err := ctx.Err(); err != nil {
			// cancelled
			s.barrier.Release(t)
			return nil, err
		}

		return t, nil
	}
}

// Close releases any resources associated with the Semaphore.
// Returns an error if the Semaphore cannot be properly closed.
// After calling Close, the Semaphore cannot be used again.
func (s *Semaphore) Close() error {
	_, err := s.acquire(false)
	if err != nil {
		return err
	}

	// close barrier
	s.barrier.Close()

	if s.active.Value() > 0 {
		// wait until everyone is done to free resources
		go func() {
			_ = s.active.Wait()
			defer s.doClose()
		}()
	} else {
		// free resources
		s.doClose()
	}
	return nil
}

func (s *Semaphore) doClose() {
	close(s.global)
	close(s.readers)
	s.writers.Close()
}

// LockContext attempts to acquire an exclusive lock with a context.
// Blocks until the lock is acquired or the context is cancelled.
// Returns an error if the context is cancelled before acquisition.
func (s *Semaphore) LockContext(ctx context.Context) error {
	return s.doLockContext(ctx)
}

// Lock acquires an exclusive lock.
// Blocks until the lock is acquired and cannot be cancelled.
// Panics if the lock cannot be acquired.
func (s *Semaphore) Lock() {
	if err := s.doLock(); err != nil {
		core.Panic(err)
	}
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns immediately with a boolean indicating success.
// Returns true if the lock was successfully acquired, false otherwise.
func (s *Semaphore) TryLock() bool {
	ok, err := s.doTryLock()
	if err != nil {
		if errors.Is(err, errors.ErrClosed) {
			return false
		}
		core.Panic(err)
	}
	return ok
}

// RLockContext attempts to acquire a read lock with a context.
// Blocks until the lock is acquired or the context is cancelled.
// Returns an error if the context is cancelled before acquisition.
func (s *Semaphore) RLockContext(ctx context.Context) error {
	return s.doRLockContext(ctx)
}

// RLock acquires a read lock.
// Blocks until the lock is acquired and cannot be cancelled.
// Panics if the semaphore is nil.
func (s *Semaphore) RLock() {
	if err := s.doRLock(); err != nil {
		core.Panic(err)
	}
}

// TryRLock attempts to acquire a read lock without blocking.
// Returns immediately with a boolean indicating success.
// Returns true if the lock was successfully acquired, false otherwise.
func (s *Semaphore) TryRLock() bool {
	ok, err := s.doTryRLock()
	if err != nil {
		if errors.Is(err, errors.ErrClosed) {
			return false
		}
		core.Panic(err)
	}
	return ok
}

// Unlock releases an exclusive lock, allowing other writers or readers to
// acquire the lock. Panics if the lock is not held or cannot be released.
func (s *Semaphore) Unlock() {
	if err := s.doUnlock(); err != nil {
		core.Panic(err)
	}
}

// RUnlock releases a read lock, allowing other readers or writers to
// acquire the lock. Panics if the lock is not held or cannot be released.
func (s *Semaphore) RUnlock() {
	if err := s.doRUnlock(); err != nil {
		core.Panic(err)
	}
}

func (s *Semaphore) doLockContext(ctx context.Context) error {
	var locked bool

	// acquire barrier, unless closed.
	b, err := s.acquireContext(ctx, false)
	if err != nil {
		return err
	}

	// increment active counter, decrease on error.
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// increment writers counter, decrease on error.
	s.writers.Inc()
	defer s.writers.Dec()

	// release token
	s.barrier.Release(b)

	select {
	case s.global <- exclusiveLock:
		// success, remains active.
		locked = true
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) doLock() error {
	var locked bool

	// acquire barrier, unless closed.
	b, err := s.acquire(false)
	if err != nil {
		return err
	}

	// increment active counter, decrease on error
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// increment writers counter, decrease on error
	s.writers.Inc()
	defer s.writers.Dec()

	// release barrier
	s.barrier.Release(b)

	// and wait
	s.global <- exclusiveLock
	locked = true
	return nil
}

func (s *Semaphore) doTryLock() (bool, error) {
	var locked bool

	// acquire barrier, unless closed.
	b, err := s.acquire(false)
	if err != nil {
		return false, err
	}

	// increment active counter, decrease on failure
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// attempt non-blocking exclusive lock
	select {
	case s.global <- exclusiveLock:
		locked = true
	default:
		locked = false
	}

	// release barrier
	s.barrier.Release(b)

	return locked, nil
}

func (s *Semaphore) doUnlock() error {
	var errMsg string

	// acquire barrier, don't mind closed.
	b, err := s.acquire(true)
	if err != nil {
		return err
	}
	// release barrier, don't mind if nil
	s.barrier.Release(b)

	select {
	case exclusive, ok := <-s.global:
		switch {
		case !ok:
			errMsg = "unlock of unlocked mutex"
		case !exclusive:
			errMsg = "unlock of read-locked mutex"
		default:
			// success. decrement active counter
			s.active.Dec()
			return nil
		}
	default:
		errMsg = "unlock of unlocked mutex"
	}

	// bad developer, die. now.
	core.Panic(core.NewPanicError(2, errMsg))

	return core.ErrUnreachable
}

func (s *Semaphore) doRLockContext(ctx context.Context) error {
	var locked bool

	// acquire barrier, unless closed
	b, err := s.acquireContext(ctx, false)
	if err != nil {
		return err
	}

	// increment active counter, decrease on error
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// release barrier
	s.barrier.Release(b)

	// lock
	err = s.unsafeRLock(ctx.Done())
	switch {
	case err == nil:
		locked = true
		return nil
	case err == context.Canceled:
		return ctx.Err()
	default:
		return err
	}
}

func (s *Semaphore) doRLock() error {
	var locked bool

	// acquire barrier, unless closed
	b, err := s.acquire(false)
	if err != nil {
		return err
	}

	// increment active counter, decrease on failure
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// release barrier
	s.barrier.Release(b)

	// lock
	err = s.unsafeRLock(nil)
	if err == nil {
		locked = true
	}
	return err
}

//revive:disable-next-line:cognitive-complexity
func (s *Semaphore) unsafeRLock(abort <-chan struct{}) error {
	var readers int
	var ok bool

	// Check if there are writers waiting before attempting to acquire a read lock
	if err := s.writers.WaitAbort(abort); err != nil {
		// cancelled
		return err
	}

	select {
	case s.global <- readerLock:
		// first reader!

		if isCancelled(abort) {
			// but cancelled. release lock and fail
			<-s.global
			return context.Canceled
		}
	case readers, ok = <-s.readers:
		if !ok {
			return errors.ErrClosed
		}

		// another reader.

		if isCancelled(abort) {
			// cancelled. put back what we took from readers channel
			s.readers <- readers
			return context.Canceled
		}
	case <-abort:
		// cancelled
		return context.Canceled
	}

	// increase readers count and let the next reader in.
	readers++
	s.readers <- readers
	return nil
}

func isCancelled(abort <-chan struct{}) bool {
	select {
	case <-abort: // nil channels are ignored
		return true
	default:
		return false
	}
}

func (s *Semaphore) doTryRLock() (bool, error) {
	var readers int
	var locked bool

	// acquire barrier, unless closed.
	b, err := s.acquire(false)
	if err != nil {
		return false, err
	}

	// increase active counter, decrease on failure
	s.active.Inc()
	defer func() {
		if !locked {
			s.active.Dec()
		}
	}()

	// release barrier
	s.barrier.Release(b)

	// Don't try to acquire read lock if writers are waiting
	if s.writers.Value() > 0 {
		return false, nil
	}

	select {
	case s.global <- readerLock:
		// first reader!
	case readers = <-s.readers:
		// another reader.
	default:
		// not this time.
		return false, nil
	}

	// increase readers count and let the next reader in.
	readers++
	s.readers <- readers

	// success
	locked = true
	return true, nil
}

func (s *Semaphore) doRUnlock() error {
	// acquire barrier, don't mind closed.
	b, err := s.acquire(true)
	if err != nil {
		return err
	}
	// release barrier, don't mind if nil
	s.barrier.Release(b)

	readers, ok := <-s.readers
	if !ok {
		return errors.ErrClosed
	}

	// decrement readers count
	readers--

	if readers == 0 {
		// last. release global lock
		<-s.global
	} else {
		// update readers count, and let the next reader in.
		s.readers <- readers
	}

	// success. decrement active counter
	s.active.Dec()
	return nil
}

var (
	_ sync.Locker          = (*Semaphore)(nil)
	_ mutex.Mutex          = (*Semaphore)(nil)
	_ mutex.MutexContext   = (*Semaphore)(nil)
	_ mutex.RWMutex        = (*Semaphore)(nil)
	_ mutex.RWMutexContext = (*Semaphore)(nil)
)
