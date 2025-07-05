package mutex

import (
	"sync"

	"darvaza.org/core"
)

// TryLock attempts to acquire locks on multiple mutexes simultaneously without
// blocking. Returns true if all locks were successfully acquired, false
// otherwise. Upon partial acquisition, it releases acquired locks in reverse
// order.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during lock/unlock
func TryLock[T Mutex](locks ...T) bool {
	ok, err := doTryLock(locks)
	if err != nil {
		panic(err)
	}
	return ok
}

func doTryLock[T Mutex](locks []T) (bool, error) {
	if len(locks) > 0 {
		return doLockLoop(locks, SafeTryLock, SafeUnlock)
	}

	return true, nil
}

// TryRLock attempts to acquire read locks on multiple mutexes simultaneously.
// Returns true if all read locks were successfully acquired, false otherwise.
// Upon partial acquisition, it releases acquired locks in reverse order.
//
// For RWMutex instances, it uses TryRLock; for regular Mutex instances, it
// uses TryLock.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during lock/unlock
func TryRLock[T Mutex](locks ...T) bool {
	ok, err := doTryRLock(locks)
	if err != nil {
		panic(err)
	}
	return ok
}

func doTryRLock[T Mutex](locks []T) (bool, error) {
	if len(locks) > 0 {
		return doLockLoop(locks, SafeTryRLock, SafeRUnlock)
	}
	return true, nil
}

// Unlock releases multiple mutexes simultaneously.
// It attempts to unlock all provided mutexes even if some operations fail,
// collecting and reporting errors.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during unlock
func Unlock[T Mutex](locks ...T) {
	if err := doUnlock(locks); err != nil {
		panic(err)
	}
}

func doUnlock[T Mutex](locks []T) error {
	if len(locks) > 0 {
		return doUnlockLoop(locks, SafeUnlock)
	}
	return nil
}

// RUnlock releases multiple mutexes using read-unlock operations.
// For RWMutex instances, it uses RUnlock; for regular Mutex instances, it
// uses Unlock.
//
// All provided mutexes are unlocked even if some operations fail, with errors
// collected and reported.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during unlock
func RUnlock[T Mutex](locks ...T) {
	if err := doRUnlock(locks); err != nil {
		panic(err)
	}
}

func doRUnlock[T Mutex](locks []T) error {
	if len(locks) > 0 {
		return doUnlockLoop(locks, SafeRUnlock)
	}
	return nil
}

// Lock acquires multiple mutexes simultaneously in the order provided.
// It blocks until all locks are acquired. Upon failure, previously acquired
// locks are released in reverse order to prevent deadlocks.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during lock/unlock
func Lock[T Mutex](locks ...T) {
	if err := doLock(locks); err != nil {
		panic(err)
	}
}

func doLock[T Mutex](locks []T) error {
	if len(locks) > 0 {
		_, err := doLockLoop(locks, SafeLock, SafeUnlock)
		return err
	}
	return nil
}

// RLock acquires multiple read locks simultaneously.
// For RWMutex instances, it uses RLock; for regular Mutex instances, it
// uses Lock.
//
// Upon failure, previously acquired locks are released in reverse order to
// prevent deadlocks.
//
// This function will panic if:
// 1. Any of the provided mutexes are nil
// 2. Any mutex operation raises an exception during lock/unlock
func RLock[T Mutex](locks ...T) {
	if err := doRLock(locks); err != nil {
		panic(err)
	}
}

func doRLock[T Mutex](locks []T) error {
	if len(locks) > 0 {
		_, err := doLockLoop(locks, SafeRLock, SafeRUnlock)
		return err
	}
	return nil
}

// doLockLoop implements the generic lock acquisition with clean-up logic.
// It attempts to lock each mutex using the provided lock function.
// If any lock operation fails, it releases any successful locks in reverse
// order.
func doLockLoop[T Mutex](
	locks []T,
	lock func(T) (bool, error),
	unlock func(T) error,
) (bool, error) {
	var errs core.CompoundError
	i := 0

	for ; i < len(locks); i++ {
		mu := locks[i]
		ok, err := lock(mu)
		if !ok || err != nil {
			// Failed to lock mutex. Release all previously acquired locks
			// and fail returning the aggregated error.
			if err := doReverseUnlock(unlock, locks[:i]); err != nil {
				_ = errs.AppendError(err)
			}
			_ = errs.AppendError(err)
			return false, errs.AsError()
		}
	}
	return true, nil
}

// doUnlockLoop implements the unlock operation for multiple mutexes.
// It attempts to unlock all provided mutexes, collecting any errors
// encountered during the process.
func doUnlockLoop[T Mutex](locks []T, unlock func(T) error) error {
	var errs core.CompoundError

	for _, mu := range locks {
		if err := unlock(mu); err != nil {
			_ = errs.AppendError(err)
		}
	}

	return errs.AsError()
}

// doReverseUnlock implements reverse-order unlocking to release locks when
// acquisition fails. It attempts to unlock all locks even if some operations
// fail, collecting any errors that occur.
func doReverseUnlock[T Mutex](unlock func(T) error, locks []T) error {
	var errs core.CompoundError

	for i := len(locks) - 1; i >= 0; i-- {
		if err := unlock(locks[i]); err != nil {
			_ = errs.AppendError(err)
		}
	}

	return errs.AsError()
}

// interface assertions
var _ Mutex = (*sync.Mutex)(nil)
var _ Mutex = (*sync.RWMutex)(nil)
var _ RWMutex = (*sync.RWMutex)(nil)
