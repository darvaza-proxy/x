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

#### cond Package

- **`Barrier`**: Token-based coordination primitive.
- **`Count`**: Atomic counter with conditional waiting.
- **`CountZero`**: Specialized counter that signals at zero.
- **`Token`**: Channel-based synchronization mechanism.

#### errors Package

- Standardized error types for synchronization issues.
- Integration with core.CompoundError for multi-error handling.

#### mutex Package

- Utility functions for safe lock operations.
- Multi-mutex operations with proper cleanup.
- Context-aware locking helpers.

#### semaphore Package

- **`Semaphore`**: Counting semaphore with full mutex interface support.
- Context-aware operations for both exclusive and shared access.

#### spinlock Package

- **`SpinLock`**: Lightweight busy-wait mutex for low-contention scenarios.

#### workgroup Package

- **`Group`**: Context-aware task coordination and lifecycle management.

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
[root AGENT.md](../AGENT.md).

## Testing Patterns

Tests focus on:

- Concurrent access patterns.
- Panic handling and recovery.
- Context cancellation scenarios.
- Performance benchmarks (especially for spinlock and count).
- Edge cases (nil receivers, double initialization).

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

### Semaphore Usage

```go
sem := semaphore.New(5) // Max 5 concurrent operations
defer sem.Close()

// Acquire with context
if err := sem.LockContext(ctx); err == nil {
    defer sem.Unlock()
    // Limited resource access
}

// Read locks for shared access
sem.RLock()
defer sem.RUnlock()
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
- **Semaphore**: Higher overhead but supports complex patterns.
- **Count/Barrier**: Efficient channel-based coordination.
- **Workgroup**: Minimal overhead over plain goroutines.

## Error Handling

Standard errors from the errors package:

- `ErrAlreadyInitialised`: Double initialization attempted.
- `ErrNotInitialised`: Operation on uninitialized primitive.
- `ErrClosed`: Operation on closed primitive.
- `ErrNilContext`: Nil context provided.
- `ErrNilMutex`: Nil mutex provided.
- `ErrNilReceiver`: Method called on nil receiver.

## Dependencies

- `darvaza.org/core`: Core utilities and error handling.
- Standard library (sync, context, runtime, atomic).

## See Also

- [Package README](README.md) for comprehensive API documentation.
- [Root AGENT.md](../AGENT.md) for mono-repo overview.
- Individual subpackage documentation for detailed usage.
