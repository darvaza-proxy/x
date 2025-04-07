# `darvaza.org/x/sync`

Package `sync` provides interfaces and utilities for synchronization primitives.

## Overview

This package defines standardized interfaces and provides utilities for working
with synchronization primitives in tandem with the standard library, while
providing additional functionality.

The synchronization primitives in this package are designed to handle panic
situations that may arise from underlying mutexes. When panics occur (which
typically indicate development mistakes rather than runtime errors), they are
aggregated while still enabling the clean-up of other locks. This approach
ensures that even in panic scenarios, resources are properly released rather
than leaking.

## Package Structure

* [`darvaza.org/x/sync`][sync-link]: The main package namespace.
  * [`mutex`][sync-mutex-link]: Contains interfaces and utilities for mutex
    operations.

[sync-link]: https://pkg.go.dev/darvaza.org/x/sync
[sync-mutex-link]: https://pkg.go.dev/darvaza.org/x/sync/mutex

## Features

- Standardized `Mutex` and `RWMutex` interfaces
- Functions for operating on multiple locks simultaneously
- Safe lock/unlock operations with proper error handling
- Collections of mutexes that can be operated on as a group

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

`darvaza.org/x/sync/mutex.Mutex` is implemented by `sync.Mutex{}` and
`sync.RWMutex{}` from the standard library.

### RWMutex

Extends the Mutex interface with read-locking capabilities:

```go
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
```

`darvaza.org/x/sync/mutex.RWMutex` is implemented by `sync.RWMutex{}` from the standard library.

## Functions

The package provides several utility functions that make mutex operations safer
by handling edge cases and panic conditions:

### Single Mutex Lock Operations

- `mutex.SafeLock[T sync.Locker](mu T) (bool, error)`: Safely acquires an exclusive lock,
    handling nil mutexes and panics
- `mutex.SafeRLock[T sync.Locker](mu T) (bool, error)`: Safely acquires a read lock
    (or normal lock if not RWMutex)
- `mutex.SafeTryLock[T mutex.Mutex](mu T) (bool, error)`: Non-blocking attempt to acquire a
    lock with nil and panic handling
- `mutex.SafeTryRLock[T mutex.Mutex](mu T) (bool, error)`: Non-blocking attempt to acquire a
    read lock safely

### Single Mutex Unlock Operations

- `mutex.SafeUnlock[T sync.Locker](mu T) error`: Safely releases a lock, handling nil
    mutexes and panics
- `mutex.SafeRUnlock[T sync.Locker](mu T) error`: Safely releases a read lock (or
    normal lock if not RWMutex)

### Multiple Mutex Lock Operations

- `mutex.Lock[T mutex.Mutex](locks ...T)`: Acquires multiple locks in order.
- `mutex.TryLock[T mutex.Mutex](locks ...T) bool`: Non-blocking attempt to acquire
    multiple locks. Returns true only if all locks were successfully acquired;
    otherwise returns false after releasing any acquired locks.
- `mutex.RLock[T mutex.Mutex](locks ...T)`: Acquires multiple read locks when
    possible.
- `mutex.TryRLock[T mutex.Mutex](locks ...T) bool`: Non-blocking attempt to acquire
    multiple read locks. Returns true only if all locks were successfully acquired;
    otherwise returns false after releasing any acquired locks.

### Multiple Mutex Unlock Operations

- `mutex.Unlock(locks ...mutex.Mutex)`: Releases multiple locks. Will attempt to
    unlock all locks even if some operations fail, aggregating any errors that
    occur.
- `mutex.RUnlock(locks ...mutex.Mutex)`: Releases multiple read locks. Will
    attempt to unlock all locks even if some operations fail, aggregating any
    errors that occur.
- `mutex.ReverseUnlock[T Mutex](unlock func(T) error, locks ...T) error`:
    Releases previously acquired locks in reverse order, collecting any panics
    that occur during unlocking. This function will attempt to unlock all provided
    locks even if some unlock operations fail. Any errors encountered are
    aggregated and returned. This function is used internally when a lock
    acquisition fails to prevent deadlocks by ensuring locks are released in the
    correct order.

These utility functions provide several benefits:

1. **Error handling**: They return explicit success/failure indicators and error
     information
2. **Type safety**: They use generics to work with any type implementing the
     required interfaces
3. **Panic protection**: They catch and convert panics into regular error
     returns
4. **Nil handling**: They check for nil mutexes and contexts, returning
     appropriate errors
5. **Interface detection**: Read operations automatically detect and use RWMutex
     implementations when available

## Error Aggregation

The package uses `core.CompoundError` to collect and combine multiple errors that may occur
during operations on multiple mutexes. This allows the package to attempt unlocking all
mutexes even when some operations fail, while still providing comprehensive error reporting
about what went wrong.

`core.Catch()` and `core.PanicError` are used to catch and convert panics into
regular errors including stack traces.

## Collections

### Mutexes

`darvaza.org/x/sync/mutex.Mutexes` is a slice-based collection of `mutex.Mutex`
objects that can be locked and unlocked together:

```go
type Mutexes []Mutex
```

It implements the `Mutex` and `RWMutex` interfaces, providing both exclusive and
shared locking operations on the collection:

- `Lock()`: Acquires exclusive locks on all mutexes in the collection.
- `RLock()`: Acquires shared (read) locks on all mutexes in the collection if
    they support it.
- `TryLock() bool`: Attempts to acquire exclusive locks without blocking.
- `TryRLock() bool`: Attempts to acquire shared locks without blocking if they
    support it.
- `Unlock()`: Releases exclusive locks on all mutexes in the collection. Will
    attempt to unlock all mutexes even if some fail, aggregating any errors that
    occur.
- `RUnlock()`: Releases shared (read) locks on all mutexes in the collection if
    they support it. Will attempt to unlock all mutexes even if some fail,
    aggregating any errors that occur.

Shared (read) lock operations test if the underlying `Mutex` does also implement
`RWMutex`. Otherwise, an exclusive lock is used instead.

### Example Usage of Mutexes Collection

```go
// Create a collection of mutexes
locks := mutex.Mutexes{
    &sync.Mutex{},
    &sync.RWMutex{},
    customMutex,
}

// Lock all mutexes in the collection
locks.Lock()

// Perform operations requiring all locks
// ...

// Release all locks
locks.Unlock()

// Using read locks where available
locks.RLock()
// Read operations
locks.RUnlock()
```

## Dependencies

This package only depends on the standard library and [`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## License

MIT. See `LICENCE.txt` in the `x/sync` directory of the repository for details.
