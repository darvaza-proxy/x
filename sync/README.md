# `darvaza.org/x/sync`

Package `sync` provides interfaces and utilities for synchronization primitives.

## Overview

This package defines standardized interfaces and provides utilities for working
with synchronization primitives in tandem with the standard library, while
providing additional functionality.

## Package Structure

* [`darvaza.org/x/sync`][sync-link]: The main package namespace.
  * [`mutex`][sync-mutex-link]: Contains interfaces and utilities for mutex
    operations.

[sync-link]: https://pkg.go.dev/darvaza.org/x/sync
[sync-mutex-link]: https://pkg.go.dev/darvaza.org/x/sync/mutex

## Features

- Standardized `Mutex` and `RWMutex` interfaces

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

## Dependencies

This package only depends on the standard library and [`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## License

MIT. See `LICENCE.txt` in the `x/sync` directory of the repository for details.
