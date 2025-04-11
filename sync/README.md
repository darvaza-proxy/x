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
* Lightweight spinlock implementation for low-contention scenarios
* Semaphore implementation supporting both exclusive and shared access patterns

## Package Structure

* [`darvaza.org/x/sync`][sync-link]: The main package namespace.
  * [`errors`][sync-errors-link]: Contains error types and helpers for
    implementing common synchronisation primitives.
  * [`mutex`][sync-mutex-link]: Contains interfaces and utilities for mutex
    operations.
  * [`semaphore`][sync-semaphore-link]: Provides a cancellable read-write mutex
    implementation with counting semaphore algorithms.
  * [`spinlock`][sync-spinlock-link]: Contains a lightweight spinlock
    implementation of `mutex.Mutex`.

[sync-link]: https://pkg.go.dev/darvaza.org/x/sync
[sync-errors-link]: https://pkg.go.dev/darvaza.org/x/sync/errors
[sync-mutex-link]: https://pkg.go.dev/darvaza.org/x/sync/mutex
[sync-semaphore-link]: https://pkg.go.dev/darvaza.org/x/sync/semaphore
[sync-spinlock-link]: https://pkg.go.dev/darvaza.org/x/sync/spinlock

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

## Semaphore

The `semaphore` package provides a `Semaphore` type that implements both
exclusive and shared access patterns with proper coordination between readers
and writers.

```go
type Semaphore struct{}
```

The `Semaphore` implements all mutex interfaces:

* `sync.Locker`
* `mutex.Mutex`
* `mutex.MutexContext`
* `mutex.RWMutex`
* `mutex.RWMutexContext`

### Writer Starvation Prevention

The `Semaphore` implementation uses the `Cond` type internally to track
waiting writers and prevent their starvation:

* Writers register in a wait queue using the `Cond` counter
* Readers check for waiting writers before acquiring read locks
* When writers are waiting, new readers pause until writers are serviced
* This strategy ensures writers can make progress even during heavy read
traffic

Without this mechanism, continuous reader acquisition could indefinitely
block writers from accessing the shared resource.

### Locking Methods

**Exclusive Access:**

* `Acquire(ctx context.Context) error`: Acquires exclusive lock with context
  support.
* `Lock()`: Acquires exclusive lock, panics on error.
* `LockContext(ctx context.Context) error`: Acquires exclusive lock with
  context cancellation support.
* `TryLock() bool`: Attempts non-blocking acquisition of exclusive lock.
* `TryLockContext(ctx context.Context) (bool, error)`: Non-blocking attempt
  with context support.
* `Release() error`: Releases exclusive lock with error reporting.
* `Unlock()`: Releases exclusive lock, panics on error.

**Shared Access:**

* `RLock()`: Acquires a read lock, panics on error.
* `RLockContext(ctx context.Context) error`: Acquires read lock with context
  cancellation support.
* `TryRLock() bool`: Attempts non-blocking acquisition of read lock.
* `TryRLockContext(ctx context.Context) (bool, error)`: Non-blocking attempt
  with context support.
* `RUnlock()`: Releases a read lock, panics on error.

The semaphore provides advanced synchronisation combining features of both
mutexes and traditional semaphores, with added context-awareness for
cancellation and timeout handling.

### Cond Type

The semaphore package also provides a `Cond` type, which combines features of a condition variable and an atomic counter:

```go
type Cond struct{}
```

`Cond` allows goroutines to coordinate and wait for specific conditions to be met on a shared atomic counter.

#### Key features

* **Atomic Counter**: Maintains an internal int32 value that can be manipulated atomically
* **Condition-based Waiting**: Allows goroutines to wait until a specific condition is satisfied
* **Context Support**: Provides context-aware waiting operations with cancellation and timeout handling
* **Signaling**: Supports both single-waiter signaling and broadcasting to all waiters
* **Zero-State Broadcasting**: Automatically broadcasts to waiters when counter reaches zero

#### Core Methods

**Counter Operations:**

* `Add(n int) int`: Atomically adds n to the counter and returns the new value. Automatically broadcasts only when the counter becomes zero.
* `Inc() int`: Increments the counter by 1 and returns the new value. No automatic broadcast occurs.
* `Dec() int`: Decrements the counter by 1 and returns the new value. Automatically broadcasts only when the counter becomes zero.
* `Value() int`: Returns the current counter value
* `IsZero() bool`: Checks if the counter value is zero

**Waiting Operations:**

* `WaitZero()`: Blocks until the counter value becomes zero
* `WaitFn(until CondFunc)`: Blocks until the provided condition function returns true
* `WaitFnAbort(abort <-chan struct{}, until CondFunc) error`: Waits with the ability to abort via a channel
* `WaitFnContext(ctx context.Context, until CondFunc) error`: Waits with context cancellation support

**Signaling Operations:**

* `Signal() bool`: Wakes up a single waiting goroutine
* `Broadcast() bool`: Wakes up all waiting goroutines

#### CondFunc

The `Cond` type works with condition functions of type:

```go
type CondFunc func(int32) bool
```

These functions define the conditions under which waiting goroutines should be woken up.

#### Example Usage

```go
// Create a condition variable with initial value 5
cond := semaphore.NewCond(5)

// In one goroutine, wait for the value to become zero
go func() {
    cond.WaitZero()
    // Do something when counter reaches zero
}()

// In another goroutine, wait for a custom condition
go func() {
    // Wait until the value is greater than 10
    cond.WaitFn(func(v int32) bool {
        return v > 10
    })
    // Do something when condition is met
}()

// With context cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Wait with a timeout
err := cond.WaitFnContext(ctx, func(v int32) bool {
    return v == 3
})
if err == context.DeadlineExceeded {
    // Handle timeout
}

// Manipulate the counter
newValue := cond.Add(-2) // Decrements by 2, broadcasts only if newValue == 0
cond.Inc()   // Increments by 1, no automatic broadcast

// If the counter reaches a specific value other than zero
// that should wake up waiters, manually broadcast
if cond.Dec() == 3 { // Decrements by 1, broadcasts only if result == 0
    cond.Broadcast() // Manual broadcast for non-zero conditions
}

// Signal waiters
cond.Signal()     // Wake up one waiter
cond.Broadcast()  // Wake up all waiters
```

#### Implementation Details

* Uses atomic operations for counter manipulation
* Efficiently manages waiters using channels and a map
* Handles nil receivers and other edge cases with proper error reporting
* Provides both blocking and non-blocking operations
* Implements clean cancellation handling for context-aware operations
* Automatically broadcasts to all waiters only when counter reaches zero
* For non-zero conditions, the caller must manually check and broadcast as needed

`Cond` is particularly useful for scenarios where threads need to coordinate based on a numeric value reaching a certain state, with the added benefit of context cancellation support.

## SpinLock

The `spinlock` package provides a lightweight spinlock implementation for
scenarios where locks are held for very brief periods.

```go
type SpinLock uint32
```

SpinLock is a mutual exclusion primitive that uses active spinning
(busy-waiting) instead of parking goroutines. It implements both the
`sync.Locker` and `mutex.Mutex` interfaces.

### Key characteristics

* **Zero value**: An unlocked spinlock ready for use
* **Memory footprint**: Minimal (just a uint32)
* **CPU usage**: Consumes CPU cycles while waiting (unlike traditional mutexes)
* **Target use cases**: Low-contention scenarios with briefly held locks

### Methods

* `Lock()`: Acquires the lock, spinning until successful
* `TryLock() bool`: Attempts to acquire the lock without blocking
* `Unlock()`: Releases the lock

### When to use

* Use when lock contention is rare and locks are held for minimal duration
* Avoid when locks might be held for extended periods
* Best for performance-critical code paths where context switching would be
  costly

### Example usage

```go
var lock spinlock.SpinLock

// In concurrent code:
lock.Lock()
defer lock.Unlock()
// Critical section here (keep it very brief)
```

### Implementation details

* Uses atomic operations for lock state management
* Calls `runtime.Gosched()` while spinning to yield the processor
* Panics with appropriate errors for nil receivers or unlocking unlocked
  spinlocks
* Provides internal methods that return errors rather than panicking for
  composability

### Performance characteristics

Benchmark testing shows SpinLock is efficient for operations that complete
quickly, as it avoids the overhead of parking and unparking goroutines or
using channels.

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
* `ErrNilMutex`: Returned when a Mutex was expected but none provided.
* `ErrNilReceiver`: Returned when methods are called on a nil receiver.

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
