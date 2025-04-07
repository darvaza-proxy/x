// Package mutex provides interfaces and utilities for mutual exclusion and
// synchronisation primitives.
package mutex

import "sync"

// Mutex defines a standard interface for mutual exclusion locking mechanisms
// that support basic locking, unlocking, and non-blocking attempts.
//
// Standard library types like sync.Mutex and sync.RWMutex implement this
// interface.
type Mutex interface {
	// Lock acquires the mutex, blocking until it is available.
	Lock()

	// TryLock attempts to acquire the mutex without blocking.
	// Returns true if successful, false otherwise.
	TryLock() bool

	// Unlock releases the mutex.
	// Calling Unlock on an unlocked mutex will panic.
	Unlock()
}

// RWMutex extends the Mutex interface with read-locking capabilities,
// allowing multiple readers or a single writer to access a shared resource.
//
// When a write lock is held, read lock attempts will block until the write
// lock is released. Multiple read locks can be held simultaneously.
//
// Standard library type sync.RWMutex implements this interface.
type RWMutex interface {
	Mutex

	// RLock acquires a read lock, blocking until available if necessary.
	RLock()

	// TryRLock attempts to acquire a read lock without blocking.
	// Returns true if successful, false otherwise.
	TryRLock() bool

	// RUnlock releases a read lock.
	// Calling RUnlock without holding a read lock will panic.
	RUnlock()
}

// interface assertions
var _ Mutex = (*sync.Mutex)(nil)
var _ Mutex = (*sync.RWMutex)(nil)
var _ RWMutex = (*sync.RWMutex)(nil)
