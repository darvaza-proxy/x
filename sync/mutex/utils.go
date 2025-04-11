package mutex

import (
	"context"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
)

// SafeLock attempts to acquire a lock on the provided mutex safely.
// It handles nil mutexes and catches panics from underlying lock operations.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred during locking
func SafeLock[T sync.Locker](mu T) (bool, error) {
	switch any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	default:
		err := core.Catch(func() error {
			mu.Lock()
			return nil
		})

		return err == nil, err
	}
}

// SafeTryLock attempts to acquire a lock without blocking.
// It handles nil mutexes and catches panics from underlying TryLock operations.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, nil) if the lock could not be acquired without blocking
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred during the attempt
func SafeTryLock[T Mutex](mu T) (bool, error) {
	switch any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	default:
		var ok bool

		err := core.Catch(func() error {
			ok = mu.TryLock()
			return nil
		})
		return ok, err
	}
}

// SafeUnlock releases a lock safely.
// It handles nil mutexes and catches panics from underlying operations.
// When used with ReverseUnlock or other multi-mutex operations, all mutexes
// will be unlocked even if some operations fail.
//
// Returns:
//   - nil if the unlock operation was successful
//   - ErrNilMutex if the mutex is nil
//   - err if a panic occurred during unlocking
func SafeUnlock[T sync.Locker](mu T) error {
	switch any(mu).(type) {
	case nil:
		return errors.ErrNilMutex
	default:
		return core.Catch(func() error {
			mu.Unlock()
			return nil
		})
	}
}

// SafeRLock acquires a read lock safely.
// For RWMutex, it acquires a read lock; otherwise, it acquires an exclusive lock.
// It handles nil mutexes and catches panics from underlying operations.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred during locking
func SafeRLock[T sync.Locker](mu T) (bool, error) {
	var lock func() error

	switch r := any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	case RWMutex:
		lock = func() error {
			r.RLock()
			return nil
		}
	default:
		lock = func() error {
			mu.Lock()
			return nil
		}
	}

	err := core.Catch(lock)
	return err == nil, err
}

// SafeTryRLock attempts to acquire a read lock without blocking.
// For RWMutex, it attempts a read lock; otherwise, an exclusive lock.
// It handles nil mutexes and catches panics.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, nil) if the lock could not be acquired without blocking
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred during the attempt
func SafeTryRLock[T Mutex](mu T) (bool, error) {
	var lock func() error
	var ok bool

	switch r := any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	case RWMutex:
		lock = func() error {
			ok = r.TryRLock()
			return nil
		}
	default:
		lock = func() error {
			ok = mu.TryLock()
			return nil
		}
	}

	err := core.Catch(lock)
	return ok, err
}

// SafeRUnlock releases a read lock safely.
// For RWMutex, it releases a read lock; otherwise, it releases an exclusive lock.
// It handles nil mutexes and catches panics.
// When used with ReverseUnlock, all mutexes will be unlocked even if some
// operations fail.
//
// Returns:
//   - nil if the unlock operation was successful
//   - ErrNilMutex if the mutex is nil
//   - err if a panic occurred during unlocking
func SafeRUnlock[T sync.Locker](mu T) error {
	var unlock func() error

	switch r := any(mu).(type) {
	case nil:
		return errors.ErrNilMutex
	case RWMutex:
		unlock = func() error {
			r.RUnlock()
			return nil
		}
	default:
		unlock = func() error {
			mu.Unlock()
			return nil
		}
	}

	return core.Catch(unlock)
}

// NewSafeLockContext creates a context-aware locking function.
// The returned function acquires a lock respecting context cancellation
// or timeouts.
//
// Parameters:
//   - ctx: The context for cancellation or timeout control
//
// Returns:
//   - A function that takes a mutex and returns acquisition status and any error
func NewSafeLockContext[T MutexContext](ctx context.Context) func(mu T) (bool, error) {
	return func(mu T) (bool, error) {
		return SafeLockContext[T](ctx, mu)
	}
}

// NewSafeRLockContext creates a context-aware read locking function.
// The returned function acquires a read lock respecting context cancellation
// or timeouts.
//
// Parameters:
//   - ctx: The context for cancellation or timeout control
//
// Returns:
//   - A function that takes a mutex and returns acquisition status and any error
func NewSafeRLockContext[T MutexContext](ctx context.Context) func(mu T) (bool, error) {
	return func(mu T) (bool, error) {
		return SafeRLockContext[T](ctx, mu)
	}
}

// SafeLockContext implements context-aware locking with error handling.
// It handles nil mutexes, nil contexts, and catches panics.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, ErrNilContext) if the context is nil
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred or the context expired during locking
func SafeLockContext[T MutexContext](ctx context.Context, mu T) (bool, error) {
	if ctx == nil {
		return false, errors.ErrNilContext
	}

	switch any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	default:
		err := core.Catch(func() error {
			return mu.LockContext(ctx)
		})
		return err == nil, err
	}
}

// SafeRLockContext implements context-aware read locking with error handling.
// For RWMutexContext, it acquires a read lock; otherwise, an exclusive lock.
// It handles nil mutexes, nil contexts, and catches panics.
//
// Returns:
//   - (true, nil) if the lock was successfully acquired
//   - (false, ErrNilContext) if the context is nil
//   - (false, ErrNilMutex) if the mutex is nil
//   - (false, err) if a panic occurred or the context expired during locking
func SafeRLockContext[T MutexContext](ctx context.Context, mu T) (bool, error) {
	var lock func() error

	if ctx == nil {
		return false, errors.ErrNilContext
	}

	switch r := any(mu).(type) {
	case nil:
		return false, errors.ErrNilMutex
	case RWMutexContext:
		lock = func() error {
			return r.RLockContext(ctx)
		}
	default:
		lock = func() error {
			return mu.LockContext(ctx)
		}
	}

	err := core.Catch(lock)
	return err == nil, err
}

// ReverseUnlock releases previously acquired locks in reverse order,
// collecting any possible panic. It's used when a lock request fails
// to prevent deadlocks.
// This function will attempt to unlock all provided locks even if some
// operations fail. Errors are aggregated and returned as a single error.
//
// This is a critical safety feature that prevents resource leaks by ensuring
// that unlock attempts are made on all locks, even after encountering failures.
func ReverseUnlock[T Mutex](unlock func(T) error, locks ...T) error {
	switch {
	case unlock == nil:
		return errors.New("unlock function is nil")
	case len(locks) == 0:
		return nil
	default:
		return doReverseUnlock(unlock, locks)
	}
}
