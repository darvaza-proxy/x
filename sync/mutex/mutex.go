package mutex

import (
	"sync"

	"darvaza.org/core"
)

// TryLock attempts to acquire locks on multiple mutexes simultaneously without
// blocking. Returns true if all locks were successfully acquired, false otherwise.
// If it can't acquire all locks, it will release any acquired locks in reverse order.
// If problems occur during locking or releasing, errors will be aggregated and
// raised as a panic at the end.
//
// This function will panic in two cases:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during lock/unlock operations (after releasing locks)
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
// It returns true if all read locks are successfully acquired, and false otherwise.
// If it can't acquire all read locks, it releases any previously acquired read locks
// in reverse order.
// If problems occur during locking or releasing, errors will be aggregated and
// raised as a panic at the end.
//
// For RWMutex instances, it uses the TryRLock method instead of TryLock.
// For regular Mutex instances, it falls back to TryLock.
//
// This function will panic in two cases:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during lock/unlock operations (after releasing locks)
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
// It attempts to unlock all provided mutexes even if some operations fail.
// If any unlock operation panics, errors will be aggregated and raised at the end.
// This ensures that all locks are released as much as possible, preventing resource
// leaks.
//
// This function will panic:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during unlock operations
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

// RUnlock releases multiple mutexes simultaneously using read-unlock operations.
// For RWMutex instances, it uses the RUnlock method instead of Unlock.
// For regular Mutex instances, it falls back to Unlock.
//
// It attempts to unlock all provided mutexes even if some operations fail.
// If any unlock operation panics, errors will be aggregated and raised at the end.
// This ensures that all locks are released as much as possible, preventing resource
// leaks.
//
// This function will panic:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during unlock operations
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

// Lock acquires multiple mutexes simultaneously.
// It acquires locks in the order provided and blocks until all locks are acquired.
// If any lock operation panics, previously acquired locks are released
// in reverse order to prevent deadlocks.
// If any unlock operation panics, errors will be aggregated and raised at the end
// together with the initial panic that caused the release.
//
// This function will panic in two cases:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during lock/unlock operations (after releasing locks)
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
// For RWMutex instances, it uses the RLock method to acquire a read lock.
// For regular Mutex instances, it falls back to Lock.
//
// If any lock operation panics, previously acquired locks are released
// in reverse order to prevent deadlocks.
// If any unlock operation panics, errors will be aggregated and raised at the end
// together with the initial panic that caused the release.
//
// This function will panic in two cases:
// 1. If any of the provided mutexes are nil
// 2. If any mutex panics during lock/unlock operations (after releasing locks)
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
// It attempts to lock each mutex in the given slice using the provided lock function.
// If any lock operation fails, it calls doReverseUnlock to release any locks
// that were successfully acquired.
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
			// failed to lock mutex. release all previously acquired locks
			// and fail returning the aggregated error.
			if err := doReverseUnlock(unlock, locks[:i]); err != nil {
				errs.AppendError(err)
			}
			errs.AppendError(err)
			return false, errs.AsError()
		}
	}
	return true, nil
}

// doUnlockLoop implements the unlock operation for multiple mutexes.
// It attempts to unlock all provided mutexes, even if some operations fail.
// Any errors encountered during unlocking are collected and returned as a combined error.
func doUnlockLoop[T Mutex](locks []T, unlock func(T) error) error {
	var errs core.CompoundError

	for _, mu := range locks {
		if err := unlock(mu); err != nil {
			errs.AppendError(err)
		}
	}

	return errs.AsError()
}

// doReverseUnlock implements the reverse-order unlocking used to release
// locks when a lock acquisition fails. It will attempt to unlock all locks
// even if some unlock operations fail, collecting any errors that occur.
func doReverseUnlock[T Mutex](unlock func(T) error, locks []T) error {
	var errs core.CompoundError

	for i := len(locks) - 1; i >= 0; i-- {
		if err := unlock(locks[i]); err != nil {
			errs.AppendError(err)
		}
	}

	return errs.AsError()
}

// interface assertions
var _ Mutex = (*sync.Mutex)(nil)
var _ Mutex = (*sync.RWMutex)(nil)
var _ RWMutex = (*sync.RWMutex)(nil)
