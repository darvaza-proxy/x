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

	// global holds the state of the semaphore.
	// true if an exclusive lock is held,
	// false if a reader lock is held.
	global chan bool
	// readers holds the count of readers unless it's the first
	readers chan int
	// writers tracks if there are writers waiting
	writers cond.CountZero
}

func (s *Semaphore) lazyInit() error {
	if s == nil {
		return core.ErrNilReceiver
	}

	// RO
	s.mu.RLock()
	if s.global != nil {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	// RW
	s.mu.Lock()
	if s.global == nil {
		s.init()
	}
	s.mu.Unlock()
	return nil
}

func (s *Semaphore) init() {
	_ = s.writers.Init(0)

	s.global = make(chan bool, 1)
	s.readers = make(chan int, 1)
}

func (s *Semaphore) checkContext(ctx context.Context) error {
	err := s.lazyInit()
	switch {
	case err != nil:
		return err
	case ctx == nil:
		return errors.ErrNilContext
	default:
		return nil
	}
}

// TODO: implement Semaphore.Close

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
	if err := s.checkContext(ctx); err != nil {
		// invalid
		return err
	}

	s.writers.Inc()
	defer s.writers.Dec()

	select {
	case s.global <- exclusiveLock:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) doLock() error {
	if err := s.lazyInit(); err != nil {
		// invalid
		return err
	}

	s.writers.Inc()
	defer s.writers.Dec()

	s.global <- exclusiveLock
	return nil
}

func (s *Semaphore) doTryLock() (bool, error) {
	if err := s.lazyInit(); err != nil {
		return false, err
	}

	select {
	case s.global <- exclusiveLock:
		return true, nil
	default:
		return false, nil
	}
}

func (s *Semaphore) doUnlock() error {
	var errMsg string

	if err := s.lazyInit(); err != nil {
		// invalid
		return err
	}

	select {
	case exclusive := <-s.global:
		if exclusive {
			// success
			return nil
		}

		errMsg = "unlock of read-locked mutex"
	default:
		errMsg = "unlock of unlocked mutex"
	}

	// bad developer, die. now.
	core.Panic(core.NewPanicError(2, errMsg))

	return core.ErrUnreachable
}

func (s *Semaphore) doRLockContext(ctx context.Context) error {
	err := s.checkContext(ctx)
	switch {
	case err != nil:
		// invalid
		return err
	case s.unsafeRLock(ctx.Done()):
		// cancelled
		return ctx.Err()
	default:
		// success
		return nil
	}
}

func (s *Semaphore) doRLock() error {
	if err := s.lazyInit(); err != nil {
		// invalid
		return err
	}

	s.unsafeRLock(nil) // nil means not abort
	return nil
}

func (s *Semaphore) unsafeRLock(abort <-chan struct{}) (cancelled bool) {
	var readers int

	// Check if there are writers waiting before attempting to acquire a read lock
	if err := s.writers.WaitAbort(abort); err != nil {
		// cancelled
		return true
	}

	select {
	case s.global <- readerLock:
		// first reader!

		if isCancelled(abort) {
			// but cancelled. release lock and fail
			<-s.global
			return true
		}
	case readers = <-s.readers:
		// another reader.

		if isCancelled(abort) {
			// cancelled. put back what we took from readers channel
			s.readers <- readers
			return true
		}
	case <-abort: // nil channels are ignored
		// cancelled.
		return true
	}

	// increase readers count and let the next reader in.
	readers++
	s.readers <- readers
	return false
}

func (s *Semaphore) doTryRLock() (bool, error) {
	var readers int

	if err := s.lazyInit(); err != nil {
		// invalid
		return false, err
	}

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
	return true, nil
}

func isCancelled(abort <-chan struct{}) bool {
	select {
	case <-abort: // nil channels are ignored
		return true
	default:
		return false
	}
}

func (s *Semaphore) doRUnlock() error {
	var readers int

	if err := s.lazyInit(); err != nil {
		return err
	}

	select {
	case s.global <- readerLock:
		// it wasn't locked. wtf
		// release unwanted lock
		<-s.global

		// bad developer, die. now.
		err := core.NewPanicError(2, "unlock of unlocked mutex")
		core.Panic(err)
	case readers = <-s.readers:
		// decrement
		readers--
	}

	if readers == 0 {
		// last. release global lock
		<-s.global
	} else {
		// update readers count, and let the next reader in.
		s.readers <- readers
	}
	return nil
}

var (
	_ sync.Locker          = (*Semaphore)(nil)
	_ mutex.Mutex          = (*Semaphore)(nil)
	_ mutex.MutexContext   = (*Semaphore)(nil)
	_ mutex.RWMutex        = (*Semaphore)(nil)
	_ mutex.RWMutexContext = (*Semaphore)(nil)
)
