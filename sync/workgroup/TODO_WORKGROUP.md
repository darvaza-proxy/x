# TODO: Workgroup Done() Race Condition Fix

## Context

This issue was discovered on 2025-01-11 while attempting to use
`workgroup.Group` to fix race conditions in `darvaza.org/resolver/pool.go`.
The resolver's DNS exchange pool needs to:

1. Create a group of concurrent DNS queries
2. Set up a channel that closes when all queries complete
3. Spawn queries dynamically based on strategy (once, sequential, interval)

The natural usage pattern was:

```go
eg := &groupEx{
    wg: workgroup.New(ctx),
    ch: make(chan *poolEx),
}

// Set up completion watcher immediately
go func() {
    <-eg.wg.Done()  // This starts a Wait() before any tasks exist!
    close(eg.ch)
}()

// Later, spawn tasks...
eg.wg.Go(func(ctx context.Context) { ... })  // RACE!
```

After extensive investigation and multiple attempted fixes (reference counts,
atomic counters, boolean flags), we realized the issue is fundamental to how
`sync.WaitGroup` works and discovered a clean solution that could benefit
workgroup itself.

## Problem Statement

The current `Done()` implementation has a race condition when called before
any tasks are added via `Go()`. This violates the `sync.WaitGroup` contract
that all `Add()` calls must happen-before any `Wait()` call.

### Current Issue

```go
// This pattern causes a race:
wg := workgroup.New(ctx)
<-wg.Done()  // Starts goroutine that calls Wait()
wg.Go(fn)    // Calls Add() - RACE with Wait()!
```

### Root Cause

The `doDone()` method immediately starts a goroutine that calls
`wg.wg.Wait()`, but if no tasks have been added yet, this races with
subsequent `Go()` calls that do `wg.wg.Add(1)`.

### Race Detector Output

```text
WARNING: DATA RACE
Write at 0x00c0003b5be8 by goroutine 343:
  darvaza.org/x/sync/workgroup.(*Group).doDone.func1()
      workgroup.go:182 +0xe8  // wg.wg.Wait()

Previous read at 0x00c0003b5be8 by goroutine 335:
  darvaza.org/x/sync/workgroup.(*Group).doGo()
      workgroup.go:341 +0x64  // wg.wg.Add(1)
```

## Failed Attempts

Before arriving at the clean solution, we tried several approaches that all
proved fragile:

### 1. Reference Count Pattern

Added a dummy reference count on init that would be released on first task
or Done():

- **Problem**: Imposed workflow requirements (must call Done/Wait/Close)
- **Complexity**: Required careful tracking of who removes the dummy

### 2. Atomic Task Counter

Tracked task count separately from WaitGroup:

- **Problem**: Race between checking counter and WaitGroup operations
- **Complexity**: Polling loops, sleep delays, goroutine leaks

### 3. Boolean State Flags

Used `started` and `hasDummy` atomic booleans:

- **Problem**: Complex state machine with edge cases
- **Complexity**: Multiple CAS operations, hard to reason about

All these attempts were trying to work around the fundamental issue instead
of fixing it.

## Proposed Solution

Separate channel creation from watcher start - the same pattern discovered
while fixing the resolver pool race.

### Implementation

```go
type Group struct {
    // ... existing fields ...
    doneCh    chan struct{}
    doneOnce  sync.Once  // Ensures watcher starts exactly once
}

func (wg *Group) doDone() <-chan struct{} {
    wg.mu.Lock()
    if ch := wg.doneCh; ch != nil {
        // Reuse existing channel
        wg.mu.Unlock()
        return ch
    }

    // Create channel immediately (safe, no goroutine yet)
    ch := make(chan struct{})
    wg.doneCh = ch
    wg.mu.Unlock()

    // DON'T start watcher here - wait for first task
    return ch
}

func (wg *Group) doGo(fn func(context.Context)) error {
    switch {
    case fn == nil:
        return nil
    case wg.cancelled.Load():
        return errors.ErrClosed
    default:
        // Start done watcher on first task (if Done() was called)
        wg.mu.RLock()
        needsWatcher := wg.doneCh != nil
        wg.mu.RUnlock()

        if needsWatcher {
            wg.doneOnce.Do(func() {
                go func() {
                    wg.wg.Wait()  // Now safe: Add() happens first

                    wg.mu.Lock()
                    if ch := wg.doneCh; ch != nil {
                        close(ch)
                        wg.doneCh = nil
                    }
                    wg.mu.Unlock()
                }()
            })
        }

        wg.wg.Add(1)  // This now happens-before Wait()
        go func() {
            defer wg.wg.Done()
            fn(wg.ctx)
        }()
        return nil
    }
}
```

### Alternative: Reset on Reuse

If we want the Done() channel to work across multiple uses of the group:

```go
func (wg *Group) doDone() <-chan struct{} {
    wg.mu.Lock()
    defer wg.mu.Unlock()

    // If we have a closed channel from previous use, make a new one
    if ch := wg.doneCh; ch != nil {
        select {
        case <-ch:
            // Channel was closed, create new one for reuse
            wg.doneCh = make(chan struct{})
            wg.doneOnce = sync.Once{} // Reset the once
        default:
            // Channel still open, reuse it
            return ch
        }
    } else {
        // First call to Done()
        wg.doneCh = make(chan struct{})
    }

    return wg.doneCh
}
```

## Benefits

1. **Eliminates Race Condition**: Guarantees Add() happens-before Wait()
2. **No Workflow Requirements**: Works regardless of call order
3. **Backwards Compatible**: Existing correct usage continues to work
4. **Zero Overhead**: No watcher goroutine if Done() never called
5. **Lazy Initialization**: Watcher only starts when actually needed

## Testing

```go
func TestDoneBeforeGo(t *testing.T) {
    ctx := context.Background()
    wg := workgroup.New(ctx)

    // This should not race
    done := wg.Done()

    // Add a task after Done()
    wg.Go(func(ctx context.Context) {
        time.Sleep(10 * time.Millisecond)
    })

    // Should complete without race
    select {
    case <-done:
        // Success
    case <-time.After(100 * time.Millisecond):
        t.Fatal("Done() channel didn't close")
    }
}

func TestDoneWithNoTasks(t *testing.T) {
    ctx := context.Background()
    wg := workgroup.New(ctx)

    done := wg.Done()

    // No tasks added - channel should remain open
    select {
    case <-done:
        t.Fatal("Done() channel closed with no tasks")
    case <-time.After(10 * time.Millisecond):
        // Expected - channel stays open
    }

    // Cancel should close it
    wg.Cancel(nil)

    select {
    case <-done:
        // Expected after cancel
    case <-time.After(10 * time.Millisecond):
        t.Fatal("Done() channel didn't close after cancel")
    }
}
```

## Implementation Notes

1. The `doneOnce` ensures the watcher goroutine starts exactly once
2. The mutex protects the `doneCh` field during concurrent access
3. The watcher cleans up by setting `doneCh = nil` after closing
4. For reusable groups, we'd need to reset `doneOnce` when creating a new
   channel

## Why This Matters

The current limitation forces users into specific patterns:

1. Must add at least one task before calling Done()
2. Can't set up completion monitoring before knowing if there will be work
3. Makes it hard to compose workgroup with other patterns (like our DNS pool)

With this fix, workgroup becomes much more flexible and composable,
allowing natural patterns like:

- Setting up result collectors before spawning workers
- Monitoring completion from the start
- Zero-task groups that complete immediately

## Impact

This is a **backwards-compatible** improvement:

- Existing correct usage continues to work
- Previously racy usage becomes safe
- No performance impact (actually slightly better - no watcher if Done()
  never called)

## References

- Original issue discovered: 2025-01-11 during resolver pool testing
- Race detector output: See darvaza.org/resolver/RACE.md
- Similar pattern used in: darvaza.org/resolver/pool.go (groupEx)
- Full investigation: darvaza.org/resolver/RACE.md Section "Workgroup
  Integration Investigation"
