// Package mutex provides interfaces and utilities for mutual exclusion and synchronization primitives.
package mutex

import "sync"

// Mutex defines a standard interface for mutual exclusion locking mechanisms
// that support basic locking, unlocking, and non-blocking lock attempts.
//
// This interface is implemented by standard library types like sync.Mutex
// and sync.RWMutex.
type Mutex interface {
	// Lock acquires the mutex, blocking until it is available.
	Lock()

	// TryLock attempts to acquire the mutex without blocking.
	// Returns true if the lock was acquired, false otherwise.
	TryLock() bool

	// Unlock releases the mutex.
	// Calling Unlock on an unlocked mutex is expected to panic.
	Unlock()
}

// RWMutex extends the Mutex interface with read-locking capabilities,
// allowing multiple readers or a single writer to access a shared resource.
//
// This interface is implemented by standard library types like sync.RWMutex.
// When a write lock is held, all read lock attempts will block until the write
// lock is released. Multiple read locks can be held simultaneously.
type RWMutex interface {
	Mutex

	// RLock acquires a read lock on the mutex, blocking until it is available
	// if necessary.
	RLock()

	// TryRLock attempts to acquire a read lock without blocking.
	// Returns true if the lock was acquired, false otherwise.
	TryRLock() bool

	// RUnlock releases a read lock on the mutex.
	// Calling RUnlock on a mutex not holding a read lock is expected to panic.
	RUnlock()
}

// interface assertions
var _ Mutex = (*sync.Mutex)(nil)
var _ Mutex = (*sync.RWMutex)(nil)
var _ RWMutex = (*sync.RWMutex)(nil)
