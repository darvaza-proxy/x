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

### Features

* Standardised `Mutex` and `RWMutex` interfaces

## Package Structure

* [`darvaza.org/x/sync`][sync-link]: The main package namespace.
  * [`mutex`][sync-mutex-link]: Contains interfaces and utilities for mutex
    operations.

[sync-link]: https://pkg.go.dev/darvaza.org/x/sync
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

## Dependencies

This package only depends on the standard library and
[`darvaza.org/core`][core-link].

[core-link]: https://pkg.go.dev/darvaza.org/core

## Licence

MIT. See `LICENCE.txt` in the `x/sync` directory of the repository for
details.
