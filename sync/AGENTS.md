# Agent Documentation for x/sync

## Overview

The `sync` package provides advanced synchronization primitives that extend
Go's standard sync package. It offers standardized interfaces, context-aware
operations, condition variables, semaphores, spinlocks, and workgroup
management with comprehensive error handling.

For detailed API documentation and usage examples, see [README.md](README.md).

## Key Components

### Core Interfaces

- **`Mutex`**: Standard mutex interface with TryLock support.
- **`RWMutex`**: Read-write mutex interface.
- **`MutexContext`**: Context-aware mutex for cancellation/timeout.
- **`RWMutexContext`**: Context-aware read-write mutex.

### Subpackages

#### atomic Package

- Shadows and re-exports the standard library `sync/atomic` types
  (`Bool`, `Int32`, `Int64`, `Uint32`, `Uint64`, `Uintptr`, `Value`,
  `Pointer[T]`).
- **`BitmaskOr`**: Atomic bit-mask OR returning a per-mask first-writer flag.
- **`UpdateMax`**: Atomic monotonic max update.

#### cond Package

- **`Barrier`**: Token-based coordination primitive.
- **`Count`**: Atomic counter with conditional waiting.
- **`CountZero`**: Specialized counter that signals at zero.
- **`Once`**: Runs an initialiser once and returns its result, error or
  panic included, to every call.
- **`Token`**: Channel-based synchronization mechanism.
- **`Turnstile`**: Lock that can also be passed without holding it,
  serving holders and passers in arrival order.

#### errors Package

- Standardized error types for synchronization issues.
- Integration with core.CompoundError for multi-error handling.

#### mutex Package

- Utility functions for safe lock operations.
- Multi-mutex operations with proper cleanup.
- Context-aware locking helpers.

#### semaphore Package

- **`Semaphore`**: Cancellable read-write lock with full mutex interface
  support.
- Context-aware operations for both exclusive and shared access.
- Writers hold a `cond.Turnstile` while they wait for the lock and readers
  pass it, so readers arriving after a waiting writer wait behind it.

#### spinlock Package

- **`SpinLock`**: Lightweight busy-wait mutex for low-contention scenarios.

#### workgroup Package

- **`Group`**: Context-aware task coordination and lifecycle management.

#### internal/synctesting Package (test-only)

Benchmark reporting for the TryLock-style benchmarks. Consumed by
sibling test suites in `sync/` via the `internal/` visibility boundary;
not part of the public API. Channel and timing assertions come from
`darvaza.org/core`.

- **`ReportTryMetrics`**: report the attempts-per-acquisition,
  acquisitions-per-second and nanoseconds-per-attempt ratios of a
  TryLock benchmark.
- **`MetricReporter`**: the `*testing.B` subset `ReportTryMetrics`
  needs, taken as an interface so a test can record what was emitted.

## Architecture Notes

The package follows several design principles:

1. **Interface Standardization**: Common interfaces for all mutex types.
2. **Panic Safety**: Operations handle and aggregate panics properly.
3. **Context Integration**: Full context.Context support where appropriate.
4. **Type Safety**: Generic functions for compile-time safety.
5. **Resource Cleanup**: Ensures locks are released even during failures.

Key patterns:

- Panics indicate programming errors, not runtime conditions.
- Safe* functions convert panics to errors for composability.
- Multi-lock operations maintain order and cleanup on failure.
- Context operations handle cancellation as normal flow.

## Development Commands

For common development commands and workflow, see the
[root AGENTS.md](../AGENTS.md).

## Testing Patterns

Tests focus on:

- Concurrent access patterns.
- Panic handling and recovery.
- Context cancellation scenarios.
- Performance benchmarks (especially for spinlock and count).
- Edge cases (nil receivers, double initialization).

For tests that wait on channels or synchronisation primitives, use
`core.AssertEventually`, `core.AssertClosed`, `core.AssertQuiet` and
`core.AssertReceives` rather than ad-hoc `time.After` or polling loops.
`AssertQuiet` is the one for a signal that must be withheld: `AssertOpen`
consumes the value and passes.

## Common Usage Patterns

### Basic Mutex Operations

```go
// Safe locking with error handling
locked, err := mutex.SafeLock(mu)
if err != nil {
    // Handle error
}
defer mutex.SafeUnlock(mu)
```

### Multi-Mutex Operations

```go
// Lock multiple mutexes in order
mutex.Lock(mu1, mu2, mu3)
defer mutex.Unlock(mu1, mu2, mu3)

// Try to lock all or none
if mutex.TryLock(mu1, mu2, mu3) {
    defer mutex.Unlock(mu1, mu2, mu3)
    // Critical section
}
```

### Context-Aware Locking

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if locked, err := mutex.SafeLockContext(ctx, mu); err == nil && locked {
    defer mutex.SafeUnlock(mu)
    // Critical section
}
```

### Condition Variables

```go
// Count-based coordination
count := cond.NewCount(10)
defer count.Close()

// Wait for specific condition
count.WaitFn(func(n int32) bool { return n < 5 })

// CountZero for completion tracking
cz := cond.NewCountZero(workers)
for i := 0; i < workers; i++ {
    go func() {
        defer cz.Dec()
        // Do work
    }()
}
cz.Wait() // Wait for all workers
```

### Lazy Initialisation

A type whose zero value is ready for use initialises itself on first use
through a `cond.Once` field:

```go
type Queue struct {
    once  cond.Once
    items chan Item
}

func (q *Queue) lazyInit() error {
    switch {
    case q == nil:
        return errors.ErrNilReceiver
    case q.once.Done():
        return nil
    default:
        return q.once.Do(q.init)
    }
}

func (q *Queue) init() error {
    q.items = make(chan Item, 16)
    return nil
}

func (q *Queue) Push(it Item) error {
    if err := q.lazyInit(); err != nil {
        return err
    }
    q.items <- it
    return nil
}
```

- Check the receiver in `lazyInit`: `q.once` on a nil `q` panics before
  `Do` can return `errors.ErrNilReceiver`.
- Check `Done` before `Do`: passing `q.init` builds a method value on
  every call, which `Done` skips once the initialiser has succeeded.
- Call `lazyInit` before reading anything the initialiser sets. The
  atomic load in `Done` and `Do` orders those reads after the
  initialiser's writes.
- Keep the initialiser brief, as calls arriving during it spin. It must
  not call `Do` on the same `Once`, directly or through a method that
  calls `lazyInit`: that call waits for itself.
- A failure is returned to every later call rather than retried, so
  return an error only for a state that retrying would not fix.
- `Once` holds a pointer, so `fieldalignment` wants it among the pointer
  fields, ahead of any field whose tail holds none.

### Semaphore Usage

```go
var sem semaphore.Semaphore // the zero value is ready for use

func write(ctx context.Context) error {
    // Wait for exclusive access, giving up when ctx is cancelled.
    if err := sem.LockContext(ctx); err != nil {
        return err
    }
    defer sem.Unlock()
    // Write.
    return nil
}

func read() {
    // Wait for shared access, behind any writer already waiting.
    sem.RLock()
    defer sem.RUnlock()
    // Read.
}
```

### SpinLock for Low Contention

```go
var lock spinlock.SpinLock

// Very brief critical section
lock.Lock()
counter++
lock.Unlock()
```

### Workgroup Management

```go
wg := workgroup.New(ctx)
defer wg.Close()

// Spawn workers
for i := 0; i < 10; i++ {
    wg.Go(func(ctx context.Context) {
        select {
        case <-ctx.Done():
            return
        case <-work:
            // Process work
        }
    })
}

// Wait for completion or cancellation
err := wg.Wait()
```

## Performance Characteristics

- **SpinLock**: Best for very brief locks under low contention.
- **Semaphore**: Channel operations dominate its cost, and each lock also
  holds or passes a `Turnstile`, so it costs more than `sync.RWMutex` in
  exchange for cancellable waits.
- **Count/Barrier**: Efficient channel-based coordination.
- **Workgroup**: Minimal overhead over plain goroutines.

## Error Handling

Standard errors from the errors package:

- `ErrAlreadyInitialised`: Double initialisation attempted.
- `ErrNotInitialised`: Operation on uninitialised primitive.
- `ErrClosed`: Operation on closed primitive.
- `ErrNotLocked`: Unlock of a lock not held, raised as a panic.
- `ErrReadLocked`: Exclusive unlock of a read-locked lock, raised as a panic.
- `ErrNilContext`: Nil context provided.
- `ErrNilFunction`: Nil function provided.
- `ErrNilMutex`: Nil mutex provided.
- `ErrNilReceiver`: Method called on nil receiver.

## Dependencies

- `darvaza.org/core`: Core utilities and error handling.
- Standard library (sync, context, runtime, atomic).

## See Also

- [Package README](README.md) for comprehensive API documentation.
- [Root AGENTS.md](../AGENTS.md) for mono-repo overview.
- Individual subpackage documentation for detailed usage.
