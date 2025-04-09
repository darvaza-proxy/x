package semaphore

import (
	"context"
	"errors"

	"darvaza.org/core"
	"darvaza.org/x/sync/mutex"
	"darvaza.org/x/sync/utils"
)

// Semaphores is a slice of RWMutexContext that provides synchronized access control
// with read-write mutex capabilities across multiple mutex instances.
type Semaphores []mutex.RWMutexContext

// Acquire attempts to acquire an exclusive lock on all semaphores in the collection using the provided context.
// If the context is cancelled or an error occurs during locking, all locks are released and the aggregated errors returned.
func (s Semaphores) Acquire(ctx context.Context) error {
	if len(s) == 0 {
		return nil
	}

	return s.doLockContext(ctx, 1)
}

// LockContext attempts to acquire an exclusive lock on all semaphores in the collection using the provided context.
// If the context is cancelled or an error occurs during locking, all locks are released and the aggregated errors returned.
func (s Semaphores) LockContext(ctx context.Context) error {
	if len(s) == 0 {
		return nil
	}

	return s.doLockContext(ctx, 1)
}

func (s Semaphores) doLockContext(ctx context.Context, skip int) error {
	if ctx == nil {
		return newNilContextError(skip + 1)
	}

	lock := utils.NewLockContext[mutex.RWMutexContext](ctx)
	unlock := utils.NewUnlock()

	_, err := s.doLockLoop(lock, unlock)
	return err
}

// Lock acquires an exclusive lock on all semaphores in the collection.
// If an error occurs during locking, all locks are released and the aggregated errors emitted as a panic.
func (s Semaphores) Lock() {
	if len(s) > 0 {
		if err := s.doLock(); err != nil {
			core.Panic(err)
		}
	}
}

func (s Semaphores) doLock() error {
	lock := func(mu mutex.RWMutexContext) (bool, error) {
		mu.Lock()
		return true, nil
	}

	unlock := func(mu mutex.RWMutexContext) error {
		mu.Unlock()
		return nil
	}

	_, err := s.doLockLoop(lock, unlock)
	return err
}

// TryLock attempts to acquire an exclusive lock on all semaphores in the collection without blocking.
// If an error occurs during the attempt, it releases all locks and panics emitting the aggregated error.
// Returns true if the lock was successfully acquired, false otherwise.
func (s Semaphores) TryLock() bool {
	ok, err := s.doTryLock()
	if err != nil {
		core.Panic(err)
	}
	return ok
}

// RLockContext attempts to acquire a shared read lock on all semaphores in the collection using the provided context.
// If the context is cancelled or an error occurs during locking, all locks are released and the aggregated errors returned.
func (s Semaphores) RLockContext(ctx context.Context) error {
	if len(s) == 0 {
		return nil
	} else if ctx == nil {
		return newNilContextError(1)
	}

	lock := func(mu mutex.RWMutexContext) (bool, error) {
		if err := mu.RLockContext(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	unlock := func(mu mutex.RWMutexContext) error {
		mu.RUnlock()
		return nil
	}

	_, err := s.doLockLoop(lock, unlock)
	return err
}

// RLock acquires a shared read lock on all semaphores in the collection.
// If an error occurs during locking, all locks are released and the aggregated errors emitted as a panic.
func (s Semaphores) RLock() {
	if len(s) > 0 {
		if err := s.doRLock(1); err != nil {
			core.Panic(err)
		}
	}
}

func (s Semaphores) doRLock(skip int) error {
	lock := func(mu mutex.RWMutexContext) (bool, error) {
		mu.RLock()
		return true, nil
	}

	unlock := func(mu mutex.RWMutexContext) error {
		mu.RUnlock()
		return nil
	}

	_, err := s.doLockLoop(lock, unlock)
	return err
}

// TryLock attempts to acquire a shared lock on all semaphores in the collection without blocking.
// If an error occurs during the attempt, it releases all locks and panics emitting the aggregated error.
// Returns true if the lock was successfully acquired, false otherwise.
func (s Semaphores) TryRLock() bool {
	ok, err := s.doRTryLock()
	if err != nil {
		core.Panic(err)
	}
	return ok
}

func (s Semaphores) doTryLock() (bool, error) {
	return s.doLockLoop(lock, unlock)
}

func (s Semaphores) doRTryLock() (bool, error) {
	return s.doLockLoop(lock, unlock)
}

func (s Semaphores) doLockLoop(lock func(mutex.MutexContext) (bool, error), unlock func(mutex.MutexContext) error) (bool, error) {
	for i, mu := range s {
		ok, err := s.doLockPass(mu, s[:i], lock, unlock)
		if !ok || err != nil {
			return ok, err
		}
	}

	return true, nil
}

func (s Semaphores) doLockPass(mu mutex.RWMutexContext, previous []mutex.RWMutexContext,
	lock func(mutex.RWMutexContext) (bool, error),
	unlock func(mutex.RWMutexContext) error) (bool, error) {
	//
	var err error
	var ok bool

	if mu != nil {
		err = core.Catch(func() error {
			ok, err = lock(mu)
			return err
		})
	} else {
		err = errors.New("nil mutex not allowed")
	}

	if !ok || err != nil {
		// failed, release all previously locked mutexes
		var errs core.CompoundError
		errs.AppendError(err)

		if err := reverseUnlockFn(previous, unlock); err != nil {
			errs.AppendError(err)
		}

		return false, errs.AsError()
	}

	return true, nil
}

func reverseUnlockFn(previous []mutex.RWMutexContext, unlock func(mutex.RWMutexContext) error) error {
	var errs core.CompoundError

	for i := len(previous) - 1; i >= 0; i-- {
		if err := core.Catch(func() error {
			return unlock(previous[i])
		}); err != nil {
			errs.AppendError(err)
		}
	}

	return errs.AsError()
}

func (s Semaphores) Release() error {
	return s.doUnlock()
}

func (s Semaphores) Unlock() {
	if err := s.doUnlock(); err != nil {
		core.Panic(err)
	}
}

func (s Semaphores) RUnlock() {
	if err := s.doRUnlock(); err != nil {
		core.Panic(err)
	}
}

// Compile-time check to ensure Semaphores implements the RWMutexContext interface
var _ mutex.RWMutexContext = Semaphores{}
