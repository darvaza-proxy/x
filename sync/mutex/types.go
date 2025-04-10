// Package mutex provides interfaces and utilities for mutual exclusion and
// synchronisation primitives.
package mutex

import (
	"context"
	"sync"
)

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

// MutexContext extends Mutex with context-aware locking capabilities,
// allowing lock acquisition to respect context cancellation and timeouts.
// Useful for systems requiring bounded waiting times or cancellable operations.
//
//revive:disable:exported
type MutexContext interface {
	//revive:enable:exported
	Mutex

	// LockContext acquires the mutex with context awareness.
	// Blocks until lock acquisition or context completion.
	// Returns an error if the context is cancelled or times out.
	LockContext(context.Context) error
}

// RWMutexContext combines RWMutex and MutexContext interfaces, providing
// context-aware operations for both read and write locks. This enables
// timeout-bounded or cancellable lock acquisition for both reader and writer
// operations, valuable in systems where deadlock prevention and cancellation
// responsiveness are important.
type RWMutexContext interface {
	RWMutex
	MutexContext

	// RLockContext acquires a read lock with context awareness.
	// Blocks until read lock acquisition or context completion.
	// Returns an error if the context is cancelled or times out.
	RLockContext(context.Context) error
}

// interface assertions
var _ Mutex = (*sync.Mutex)(nil)
var _ Mutex = (*sync.RWMutex)(nil)
var _ RWMutex = (*sync.RWMutex)(nil)
