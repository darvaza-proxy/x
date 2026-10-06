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

// Semaphore is a read-write lock whose waits can be cancelled through a
// context: it admits one writer, or any number of readers, at a time. The
// zero value is ready for use, and a Semaphore must not be copied after
// first use.
//
// A writer holds a turnstile while it waits for the lock, and readers pass
// that turnstile before taking a read lock, so readers arriving after a
// waiting writer wait behind it. The writer releases the turnstile once
// it holds the lock. As with [sync.RWMutex], code that holds a read lock
// while it waits for another to be taken, in the same goroutine or
// another, deadlocks if a writer arrives in between.
type Semaphore struct {
	// global holds the state of the semaphore.
	// true if an exclusive lock is held,
	// false if a reader lock is held.
	global chan bool
	// readers holds the number of readers while read-locked, except
	// while a reader changes it.
	readers chan int

	// once runs init, which creates global and readers.
	once cond.Once

	// turn orders writers and readers: a writer holds it while it waits
	// for global, and readers pass it before taking a read lock.
	turn cond.Turnstile
}

// lazyInit initialises the Semaphore on first use. After that it costs
// an atomic load.
func (s *Semaphore) lazyInit() error {
	switch {
	case s == nil:
		return core.ErrNilReceiver
	case s.once.Done():
		return nil
	default:
		return s.once.Do(s.init)
	}
}

func (s *Semaphore) init() error {
	s.global = make(chan bool, 1)
	s.readers = make(chan int, 1)
	return nil
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

// LockContext attempts to acquire an exclusive lock with a context.
// Blocks until the lock is acquired or the context is cancelled.
// Returns the context's error if it is cancelled before acquisition,
// [errors.ErrNilContext] if ctx is nil, and [core.ErrNilReceiver] if the
// semaphore is nil.
func (s *Semaphore) LockContext(ctx context.Context) error {
	return s.doLockContext(ctx)
}

// Lock acquires an exclusive lock.
// Blocks until the lock is acquired and cannot be cancelled.
// Panics with [core.ErrNilReceiver] if the semaphore is nil.
func (s *Semaphore) Lock() {
	if err := s.doLock(); err != nil {
		core.PanicFrom(1, err)
	}
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns immediately with a boolean indicating success.
// Returns true if the lock was successfully acquired, false otherwise.
// It fails while the lock is held or another writer waits for it, and
// can also fail while another goroutine is part-way through taking the
// lock or releasing it.
// Panics with [core.ErrNilReceiver] if the semaphore is nil.
func (s *Semaphore) TryLock() bool {
	ok, err := s.doTryLock()
	if err != nil {
		core.PanicFrom(1, err)
	}
	return ok
}

// RLockContext attempts to acquire a read lock with a context.
// Blocks until the lock is acquired or the context is cancelled, waiting
// while a writer holds the lock or waits for it.
// Returns the context's error if it is cancelled before acquisition,
// [errors.ErrNilContext] if ctx is nil, and [core.ErrNilReceiver] if the
// semaphore is nil.
func (s *Semaphore) RLockContext(ctx context.Context) error {
	return s.doRLockContext(ctx)
}

// RLock acquires a read lock.
// Blocks until the lock is acquired and cannot be cancelled, waiting
// while a writer holds the lock or waits for it.
// Panics with [core.ErrNilReceiver] if the semaphore is nil.
func (s *Semaphore) RLock() {
	if err := s.doRLock(); err != nil {
		core.PanicFrom(1, err)
	}
}

// TryRLock attempts to acquire a read lock without blocking.
// Returns immediately with a boolean indicating success.
// Returns true if the lock was successfully acquired, false otherwise.
// It fails while a writer holds the lock or waits for it, and can also
// fail while another goroutine is part-way through taking the lock or
// releasing it.
// Panics with [core.ErrNilReceiver] if the semaphore is nil.
func (s *Semaphore) TryRLock() bool {
	ok, err := s.doTryRLock()
	if err != nil {
		core.PanicFrom(1, err)
	}
	return ok
}

// Unlock releases an exclusive lock, allowing other writers or readers to
// acquire the lock. It panics with [errors.ErrNotLocked] if the semaphore
// is unlocked, with [errors.ErrReadLocked] if it is read-locked, and with
// [core.ErrNilReceiver] if the semaphore is nil.
//
// Calling it on a read-locked semaphore while a reader is taking or
// releasing its lock still panics, but may not put back what it took,
// which can let a writer waiting for the lock take it beside the readers.
func (s *Semaphore) Unlock() {
	if err := s.doUnlock(); err != nil {
		core.PanicFrom(1, err)
	}
}

// RUnlock releases a read lock, allowing other readers or writers to
// acquire the lock. It panics with [errors.ErrNotLocked] if the semaphore
// is unlocked, and with [core.ErrNilReceiver] if the semaphore is nil.
// Calling it while a writer holds the lock is a bug the semaphore does
// not reliably detect.
func (s *Semaphore) RUnlock() {
	if err := s.doRUnlock(); err != nil {
		core.PanicFrom(1, err)
	}
}

func (s *Semaphore) doLockContext(ctx context.Context) error {
	if err := s.checkContext(ctx); err != nil {
		return err
	}

	// hold the turnstile until the lock is ours, so readers arriving
	// meanwhile wait behind this writer. checkContext has checked ctx
	// and Semaphore does not close turn, so an error here comes from the
	// context.
	if err := s.turn.LockContext(ctx); err != nil {
		return err
	}
	defer s.turn.Unlock()

	select {
	case s.global <- exclusiveLock:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) doLock() error {
	if err := s.lazyInit(); err != nil {
		return err
	}

	// hold the turnstile until the lock is ours, so readers arriving
	// meanwhile wait behind this writer.
	s.turn.Lock()
	s.global <- exclusiveLock
	s.turn.Unlock()
	return nil
}

func (s *Semaphore) doTryLock() (bool, error) {
	if err := s.lazyInit(); err != nil {
		return false, err
	}

	// take the turnstile too, so TryLock does not go ahead of a writer
	// that has taken it but not yet reached global.
	if !s.turn.TryLock() {
		// a writer is waiting, or another goroutine is at the
		// turnstile.
		return false, nil
	}
	defer s.turn.Unlock()

	select {
	case s.global <- exclusiveLock:
		return true, nil
	default:
		return false, nil
	}
}

func (s *Semaphore) doUnlock() error {
	if err := s.lazyInit(); err != nil {
		return err
	}

	select {
	case readers := <-s.readers:
		// read-locked, a misuse. put the count back, leaving global
		// alone. nothing else fills readers meanwhile: other readers
		// wait for the count, and a first reader for global, which
		// holds the readers' value unless another misuse took it.
		s.readers <- readers
		return errors.ErrReadLocked
	default:
		return s.unlockGlobal()
	}
}

func (s *Semaphore) unlockGlobal() error {
	select {
	case exclusive := <-s.global:
		if exclusive {
			// success
			return nil
		}

		// read-locked, a misuse, with a reader changing the count.
		// put the readers' value back, unless a goroutine waiting at
		// global took the slot in the same receive.
		select {
		case s.global <- exclusive:
		default:
		}
		return errors.ErrReadLocked
	default:
		// unlocked, a misuse.
		return errors.ErrNotLocked
	}
}

func (s *Semaphore) doRLockContext(ctx context.Context) error {
	if err := s.checkContext(ctx); err != nil {
		// invalid
		return err
	}

	// wait behind any writer waiting for the lock. checkContext has
	// checked ctx and Semaphore does not close turn, so an error here
	// comes from the context.
	if err := s.turn.PassContext(ctx); err != nil {
		// cancelled
		return err
	}

	if s.unsafeRLock(ctx.Done()) {
		// cancelled
		return ctx.Err()
	}
	return nil
}

func (s *Semaphore) doRLock() error {
	if err := s.lazyInit(); err != nil {
		return err
	}

	s.turn.Pass()      // wait behind any writer waiting for the lock
	s.unsafeRLock(nil) // nil means not abort
	return nil
}

func (s *Semaphore) unsafeRLock(abort <-chan struct{}) bool {
	var readers int

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
		return false, err
	}

	if !s.turn.TryPass() {
		// a writer is waiting, or another goroutine is at the
		// turnstile.
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

	// called while a writer holds the lock, a misuse left undetected:
	// neither case is ready until the writer unlocks, and a non-blocking
	// check would also catch a reader changing the count. if a reader
	// takes the lock first, this takes one off its count instead of
	// panicking, unlocking the semaphore while that reader holds it, so
	// a writer can take the lock beside it.
	select {
	case s.global <- readerLock:
		// unlocked, a misuse. give back the lock just taken.
		<-s.global
		return errors.ErrNotLocked
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
