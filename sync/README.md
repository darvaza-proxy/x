# `darvaza.org/x/sync`

[![Go Reference][godoc-badge]][godoc-link]
[![Go Report Card][goreportcard-badge]][goreportcard-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/sync.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/sync
[goreportcard-badge]: https://goreportcard.com/badge/darvaza.org/x/sync
[goreportcard-link]: https://goreportcard.com/report/darvaza.org/x/sync

Package `sync` provides interfaces and utilities for synchronisation
primitives.

## Overview

This package defines standardised interfaces for synchronisation
primitives that complement the standard library while providing
additional functionality.

The primitives handle panic situations that may arise from underlying mutexes.
When panics occur (typically indicating development mistakes rather than
runtime errors), they are aggregated while enabling proper clean-up of other
locks. This approach ensures resources are properly released rather than
leaked, even during panic scenarios.

### Features

* Standardised `Mutex` and `RWMutex` interfaces
* Context-aware mutex interfaces for cancellation and timeout support
* Functions for operating on multiple locks simultaneously
* Safe lock/unlock operations with proper error handling

## Package Structure

* [`darvaza.org/x/sync`][sync-link]: The main package namespace.
  * [`errors`][sync-errors-link]: Contains error types and helpers for
    implementing common synchronisation primitives.
  * [`mutex`][sync-mutex-link]: Contains interfaces and utilities for mutex
    operations.

[sync-link]: https://pkg.go.dev/darvaza.org/x/sync
[sync-errors-link]: https://pkg.go.dev/darvaza.org/x/sync/errors
[sync-mutex-link]: https://pkg.go.dev/darvaza.org/x/sync/mutex

## Interfaces

### Mutex

The core interface for mutual exclusion operations:

```go
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
```

The standard library's `sync.Mutex{}` and `sync.RWMutex{}` implement this
interface.

### RWMutex

Extends the Mutex interface with read-locking capabilities:

```go
type RWMutex interface {
    Mutex

    // RLock acquires a read lock on the mutex, blocking until available.
    RLock()

    // TryRLock attempts to acquire a read lock without blocking.
    // Returns true if the lock was acquired, false otherwise.
    TryRLock() bool

    // RUnlock releases a read lock on the mutex.
    // Calling RUnlock without holding a read lock is expected to panic.
    RUnlock()
}
```

The standard library's `sync.RWMutex{}` implements this interface.

### Context-Aware Interfaces

The package provides two context-aware interfaces that extend the basic mutex
interfaces:

```go
type MutexContext interface {
    Mutex

    // LockContext acquires the mutex with context awareness.
    // It blocks until the lock is acquired or the context is done.
    // Returns an error if the context is cancelled or times out.
    LockContext(context.Context) error
}

type RWMutexContext interface {
    RWMutex
    MutexContext

    // RLockContext acquires a read lock with context awareness.
    // It blocks until the read lock is acquired or the context is done.
    // Returns an error if the context is cancelled or times out.
    RLockContext(context.Context) error
}
```

These interfaces serve primarily as extension points for implementers.
While standard library mutex types don't implement them directly, the package
provides helper functions to work with these interfaces. Custom mutex
implementations can adopt these interfaces to provide context-aware locking
capabilities that respect cancellation and timeouts.

## Utility Functions

The package provides functions that make mutex operations safer by handling
edge cases and panic conditions:

### Single Mutex Operations

* `mutex.SafeLock[T sync.Locker](mu T) (bool, error)`: Safely acquires an
  exclusive lock, handling nil mutexes and panics
* `mutex.SafeRLock[T sync.Locker](mu T) (bool, error)`: Safely acquires a
  read lock (or normal lock if not RWMutex)
* `mutex.SafeTryLock[T mutex.Mutex](mu T) (bool, error)`: Non-blocking
  attempt to acquire a lock with nil and panic handling
* `mutex.SafeTryRLock[T mutex.Mutex](mu T) (bool, error)`: Non-blocking
  attempt to acquire a read lock safely
* `mutex.SafeUnlock[T sync.Locker](mu T) error`: Safely releases a lock,
  handling nil mutexes and panics
* `mutex.SafeRUnlock[T sync.Locker](mu T) error`: Safely releases a read
  lock (or normal lock if not RWMutex)

### Multiple Mutex Operations

* `mutex.Lock[T mutex.Mutex](locks ...T)`: Acquires multiple locks in order
* `mutex.TryLock[T mutex.Mutex](locks ...T) bool`: Non-blocking attempt to
  acquire multiple locks
* `mutex.RLock[T mutex.Mutex](locks ...T)`: Acquires multiple read locks
  when possible
* `mutex.TryRLock[T mutex.Mutex](locks ...T) bool`: Non-blocking attempt to
  acquire multiple read locks
* `mutex.Unlock(locks ...mutex.Mutex)`: Releases multiple locks
* `mutex.RUnlock(locks ...mutex.Mutex)`: Releases multiple read locks
* `mutex.ReverseUnlock[T Mutex](unlock func(T) error, locks ...T) error`:
  Releases locks in reverse order, collecting any panics

### Context-Aware Operations

* `NewSafeLockContext[T MutexContext](ctx context.Context)
  func(mu T) (bool, error)`: Creates a function for context-aware locking
* `NewSafeRLockContext[T MutexContext](ctx context.Context)
  func(mu T) (bool, error)`: Creates a function for context-aware read locking
* `SafeLockContext[T MutexContext](ctx context.Context, mu T) (bool, error)`:
  Acquires a lock with context cancellation/timeout support
* `SafeRLockContext[T MutexContext](ctx context.Context, mu T) (bool, error)`:
  Acquires a read lock with context cancellation/timeout support

These utility functions provide:

1. **Error handling**: Returns explicit success/failure indicators
2. **Type safety**: Uses generics to work with any compatible type
3. **Panic protection**: Catches and converts panics into regular errors
4. **Nil handling**: Checks for nil mutexes, returning appropriate errors
5. **Interface detection**: Automatically uses RWMutex implementations when
   available

## Error Handling

The [`errors`][sync-errors-link] sub-package defines error types for common
synchronisation issues:

* `ErrNilContext`: Returned when a nil context is encountered in
  context-aware operations
* `ErrNilMutex`: Returned when a Mutex was expected but none provided
* `ErrNilReceiver`: Returned when methods are called on a nil receiver

The package uses `core.CompoundError` to collect and combine multiple errors
that may occur during operations on multiple mutexes. This allows for
unlocking all mutexes even when some operations fail, while still providing
comprehensive error reporting.

`core.Catch()` and `core.PanicError` convert panics into regular errors with
stack traces.

This package follows specific patterns for handling error conditions:

### Panic Propagation

* Operations panic when encountering nil mutexes or when underlying mutex
  operations panic.
* Unlocking an unlocked mutex will panic as per standard Go mutex behaviour.
* When operating on multiple locks, panics from individual mutex operations
  are aggregated.
* During failures, the package ensures proper clean-up by releasing any
  successfully acquired locks before propagating the panic.
* Unlock operations attempt to unlock all provided mutexes even if some
  fail, collecting and aggregating errors during the process.

### Design Philosophy

Panic conditions indicate programming mistakes rather than runtime conditions
that should be handled differently. This approach was chosen because:

* Mutex misuse (e.g., unlocking an unlocked mutex) represents a development
  error.
* Aggregating panics allows for proper clean-up while still communicating the
  underlying issue.
* The interfaces lack error returns for lock/unlock operations, as these
  should not fail under normal circumstances.
* Unlock operations always attempt to release all locks, even if some fail,
  to prevent resource leaks.

For context-aware operations (using `MutexContext` and `RWMutexContext`
interfaces), explicit error handling is provided to manage cancellation and
timeout scenarios, which are considered valid runtime conditions rather than
programming errors.

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## Licence

MIT. See `LICENCE.txt` in the `x/sync` directory of the repository for
details.
