package cond

//revive:disable:cognitive-complexity

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountInitialisation tests the various ways to initialise a Count.
func TestCountInitialisation(t *testing.T) {
	t.Run("NewCount", runTestNewCount)
	t.Run("Init", runTestInit)
	t.Run("Nil receiver", runTestNilReceiver)
}

func runTestNewCount(t *testing.T) {
	c := NewCount(42)
	require.NotNil(t, c)
	assert.Equal(t, 42, c.Value())
	assert.False(t, c.IsNil())
	require.NoError(t, c.Close())
}

func runTestInit(t *testing.T) {
	c := new(Count)
	err := c.Init(21)
	require.NoError(t, err)
	assert.Equal(t, 21, c.Value())
	assert.False(t, c.IsNil())
	require.NoError(t, c.Close())
}

func runTestNilReceiver(t *testing.T) {
	var c *Count
	assert.True(t, c.IsNil())
	assert.Error(t, c.Init(0))
	assert.Error(t, c.Close())
}

// TestCountOperations tests the basic operations on the counter.
func TestCountOperations(t *testing.T) {
	t.Run("Add", runTestAdd)
	t.Run("Inc", runTestInc)
	t.Run("Dec", runTestDec)
	t.Run("Value", runTestValue)
}

func runTestAdd(t *testing.T) {
	c := NewCount(10)
	defer c.Close()

	assert.Equal(t, 15, c.Add(5))
	assert.Equal(t, 15, c.Value())
	assert.Equal(t, 8, c.Add(-7))
	assert.Equal(t, 8, c.Value())

	// Add with zero doesn't broadcast
	assert.Equal(t, 8, c.Add(0))
}

func runTestInc(t *testing.T) {
	c := NewCount(41)
	defer c.Close()

	assert.Equal(t, 42, c.Inc())
	assert.Equal(t, 42, c.Value())
}

func runTestDec(t *testing.T) {
	c := NewCount(43)
	defer c.Close()

	assert.Equal(t, 42, c.Dec())
	assert.Equal(t, 42, c.Value())
}

func runTestValue(t *testing.T) {
	c := NewCount(42)
	defer c.Close()

	assert.Equal(t, 42, c.Value())
	c.Inc()
	assert.Equal(t, 43, c.Value())
}

// TestCountConditions tests the condition matching functionality.
func TestCountConditions(t *testing.T) {
	t.Run("IsZero", runTestIsZero)
	t.Run("Match", runTestMatch)
	t.Run("BroadcastCondition", runTestBroadcastCondition)
}

func runTestIsZero(t *testing.T) {
	c := NewCount(0)
	defer c.Close()

	assert.True(t, c.IsZero())
	c.Inc()
	assert.False(t, c.IsZero())
	c.Dec()
	assert.True(t, c.IsZero())
}

func runTestMatch(t *testing.T) {
	c := NewCount(42)
	defer c.Close()

	// Test with explicit function
	isFortyTwo := func(v int32) bool { return v == 42 }
	assert.True(t, c.Match(isFortyTwo))

	// Test after modifying value
	c.Inc()
	assert.False(t, c.Match(isFortyTwo))

	// Test with nil function (equivalent to IsZero)
	_ = c.Reset(0)
	assert.True(t, c.Match(nil))
}

func runTestBroadcastCondition(t *testing.T) {
	// Only broadcast when value is divisible by 10
	broadcastOn10 := func(v int32) bool { return v%10 == 0 }
	c := NewCount(0, broadcastOn10)
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// The waiter will only be woken when counter becomes 10
	ready := make(chan struct{})
	go func() {
		defer wg.Done()
		close(ready)
		c.WaitFn(func(v int32) bool { return v >= 5 })
	}()

	// Give time for goroutine to start waiting
	<-ready
	time.Sleep(10 * time.Millisecond)

	// This increment shouldn't wake up the waiter
	c.Inc() // 1
	time.Sleep(10 * time.Millisecond)

	// Continue incrementing to reach the broadcast condition (10)
	c.Add(9) // 10

	// Wait with timeout to ensure the goroutine completes
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		// Success
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Waiter was not notified when broadcast condition was met")
	}
}

// TestCountWait tests the basic Wait functionality.
func TestCountWait(t *testing.T) {
	t.Run("Wait", runTestWait)
}

func runTestWait(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start a goroutine that waits until counter becomes zero
	go func() {
		defer wg.Done()
		c.Wait() // Will block until value becomes 0
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Decrement counter to trigger Wait completion
	c.Dec()

	// Ensure the Wait function completes
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		// Success
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Wait did not complete after counter became zero")
	}
}

// TestCountWaitFn tests the WaitFn functionality.
func TestCountWaitFn(t *testing.T) {
	t.Run("WaitFn", runTestWaitFn)
}

func runTestWaitFn(t *testing.T) {
	c := NewCount(5)
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start a goroutine that waits for a specific condition
	go func() {
		defer wg.Done()
		c.WaitFn(func(v int32) bool { return v >= 10 })
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Update counter to satisfy condition
	c.Add(5) // Not enough to satisfy condition
	time.Sleep(10 * time.Millisecond)
	c.Add(5) // Now the condition is satisfied

	// Ensure the WaitFn function completes
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		// Success
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "WaitFn did not complete when condition was met")
	}
}

// TestCountWaitFnContext tests the WaitFnContext functionality.
func TestCountWaitFnContext(t *testing.T) {
	t.Run("WaitFnContext", runTestWaitFnContextEmpty)
	t.Run("WaitFnContext_Success", runTestWaitFnContextSuccess)
	t.Run("WaitFnContext_Cancelled", runTestWaitFnContextCancelled)
}

func runTestWaitFnContextEmpty(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	// Create a context that won't time out during our test
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use a condition that's immediately satisfied (the count is already 1)
	err := c.WaitFnContext(ctx, func(v int32) bool { return v == 1 })

	// Since the condition is immediately satisfied, this should return without error
	assert.NoError(t, err, "WaitFnContext should not error when condition is immediately satisfied")

	// Test with nil condition function (should check for zero)
	err = c.Reset(0)
	require.NoError(t, err)

	// This should also return immediately without error since counter is 0
	err = c.WaitFnContext(ctx, nil)
	assert.NoError(t, err, "WaitFnContext with nil function should not error when count is zero")
}

func runTestWaitFnContextSuccess(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		errCh <- c.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Decrement to satisfy the condition
	c.Dec()

	// Ensure the WaitFnContext completes
	wg.Wait()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	default:
		assert.Fail(t, "No result from WaitFnContext")
	}
}

func runTestWaitFnContextCancelled(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		errCh <- c.WaitFnContext(ctx, nil) // Wait for zero
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	cancel()

	// Ensure the WaitFnContext completes with cancellation error
	wg.Wait()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")
	default:
		assert.Fail(t, "No result from WaitFnContext")
	}
}

// TestCountWaitFnAbort tests the WaitFnAbort functionality.
func TestCountWaitFnAbort(t *testing.T) {
	t.Run("WaitFnAbort_Success", runTestWaitFnAbortSuccess)
	t.Run("WaitFnAbort_Aborted", runTestWaitFnAbortAborted)
}

func runTestWaitFnAbortSuccess(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	abortCh := make(chan struct{})
	defer close(abortCh)

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		errCh <- c.WaitFnAbort(abortCh, func(v int32) bool { return v == 0 })
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Decrement to satisfy the condition
	c.Dec()

	// Ensure the WaitFnAbort completes
	wg.Wait()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	default:
		assert.Fail(t, "No result from WaitFnAbort")
	}
}

func runTestWaitFnAbortAborted(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	abortCh := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	errCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		errCh <- c.WaitFnAbort(abortCh, nil) // Wait for zero
	}()

	// Allow time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Close the abort channel
	close(abortCh)

	// Ensure the WaitFnAbort completes with cancellation error
	wg.Wait()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	default:
		assert.Fail(t, "No result from WaitFnAbort")
	}
}

// TestCountSignalBroadcast tests the Signal and Broadcast methods.
func TestCountSignalBroadcast(t *testing.T) {
	t.Run("Signal", runTestSignal)
	t.Run("Broadcast", runTestBroadcast)
}

func runTestSignal(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	// No waiters, signal should return false
	assert.False(t, c.Signal())

	var wg sync.WaitGroup
	wg.Add(1)

	ready := make(chan struct{})
	done := make(chan struct{})
	var doneClosed sync.Once

	// Start a waiter goroutine
	go func() {
		defer wg.Done()

		// Signal that the goroutine is about to wait
		close(ready)

		// This condition won't be matched, but the signal will wake it up
		c.WaitFn(func(_ int32) bool {
			doneClosed.Do(func() { close(done) })
			return false // Will keep waiting after being signaled
		})
	}()

	// Wait for goroutine to signal it's ready
	<-ready

	// Give time for goroutine to enter wait state
	time.Sleep(10 * time.Millisecond)

	// Signal should wake up the goroutine
	assert.True(t, c.Signal())

	// Wait for the goroutine to handle the signal
	select {
	case <-done:
		// Success - our waiter's condition function was called
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Waiter was not woken up by Signal")
	}
}

func runTestBroadcast(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	const numWaiters = 3
	// Create a bitmask with 'numWaiters' bits
	const allNotifiedMask = (1 << numWaiters) - 1

	// Use atomic value to track which waiters were notified via bitmask
	var notifiedMask atomic.Uint32

	// Use a channel to signal when all waiters have been notified
	allNotified := make(chan struct{})

	// Use a channel to ensure all waiters are ready before broadcast
	ready := make(chan struct{})

	// Start multiple waiter goroutines
	var wg sync.WaitGroup
	wg.Add(numWaiters)

	for i := range numWaiters {
		go func(waiterID int) {
			defer wg.Done()

			// Signal this waiter is ready
			ready <- struct{}{}

			c.WaitFn(func(_ int32) bool {
				// Set this waiter's bit in the bitmask
				if newMask, changed := atomicMaskOr(&notifiedMask, 1<<waiterID); changed {
					if newMask == allNotifiedMask {
						close(allNotified)
					}
				}

				return false // Continue waiting
			})
		}(i)
	}

	// Wait for all goroutines to enter wait state
	for range numWaiters {
		<-ready
	}

	// Small delay to ensure all goroutines are waiting
	time.Sleep(10 * time.Millisecond)

	// Broadcast should wake up all waiters
	c.Broadcast()

	// Wait for all waiters to be notified with timeout
	select {
	case <-allNotified:
		// Success - all waiters were notified
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Not all waiters were notified by Broadcast")
	}

	// Verify the bitmask has all bits set
	assert.Equal(t, uint32(allNotifiedMask), notifiedMask.Load(),
		"Bitmask should have all waiters' bits set")
}

// TestCountErrorHandling tests error handling and edge cases.
func TestCountErrorHandling(t *testing.T) {
	t.Run("DoubleInit", runTestDoubleInit)
	t.Run("NilContext", runTestNilContext)
	t.Run("Uninitialised", runTestUninitialised)
	t.Run("PanicRecovery", runTestPanicRecovery)
}

func runTestDoubleInit(t *testing.T) {
	c := new(Count)

	err := c.Init(0)
	require.NoError(t, err)
	defer c.Close()

	// Second Init should fail
	err = c.Init(0)
	assert.Error(t, err)
}

func runTestNilContext(t *testing.T) {
	var nilCtx context.Context

	c := NewCount(0)
	defer c.Close()

	err := c.WaitFnContext(nilCtx, nil)
	assert.Error(t, err)
}

func runTestUninitialised(t *testing.T) {
	c := new(Count)

	// Ensure operations on uninitialised Count return errors
	assert.Error(t, c.check())
	assert.True(t, c.IsNil())

	// These should be safe to call but return errors
	assert.Error(t, c.Close())

	// Context-based waits should return errors
	ctx := context.Background()
	assert.Error(t, c.WaitFnContext(ctx, nil))

	abortCh := make(chan struct{})
	assert.Error(t, c.WaitFnAbort(abortCh, nil))
}

func runTestPanicRecovery(t *testing.T) {
	var c *Count

	// These methods should panic for nil receiver
	assert.Panics(t, func() { c.Wait() })
	assert.Panics(t, func() { c.WaitFn(nil) })
	assert.Panics(t, func() { c.IsZero() })
	assert.Panics(t, func() { c.Match(nil) })
	assert.Panics(t, func() { c.Signal() })
	assert.Panics(t, func() { c.Broadcast() })

	// These don't panic but return errors
	assert.Error(t, c.Init(0))
	assert.Error(t, c.Close())

	// For an initialised Count, these methods also panic if uninitialised
	c = new(Count)
	assert.Panics(t, func() { c.Wait() })
	assert.Panics(t, func() { c.WaitFn(nil) })
}

// TestConcurrentAccess tests concurrent access to Count.
func TestConcurrentAccess(t *testing.T) {
	t.Run("ConcurrentIncrementDecrement", runTestConcurrentIncrementDecrement)
}

func runTestConcurrentIncrementDecrement(t *testing.T) {
	c := NewCount(0)
	defer c.Close()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Half incrementing, half decrementing

	// Start goroutines that increment the counter
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range numOperations {
				c.Inc()
			}
		}()
	}

	// Start goroutines that decrement the counter
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				c.Dec()
			}
		}()
	}

	// Wait for all operations to complete
	wg.Wait()

	// Counter should be back to 0
	assert.Equal(t, 0, c.Value(),
		"Counter should return to 0 after equal increments and decrements")
}

// TestCountWithTimeout tests using Count with various timeout patterns.
func TestCountWithTimeout(t *testing.T) {
	t.Run("TimeoutContext", runTestTimeoutContext)
	t.Run("EnsureNoRace", runTestEnsureNoRace)
}

func runTestTimeoutContext(t *testing.T) {
	c := NewCount(1)
	defer c.Close()

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Wait for a condition that won't be satisfied within the timeout
	start := time.Now()
	err := c.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
	elapsed := time.Since(start)

	// Should fail with timeout error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline")

	// Elapsed time should be approximately the timeout duration
	assert.True(t, elapsed >= 45*time.Millisecond,
		"Should wait at least most of the timeout period")
}

func runTestEnsureNoRace(_ *testing.T) {
	c := NewCount(0)
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		// Repeatedly access the counter
		for range 100 {
			c.Inc()
			c.Value()
			c.Dec()
		}
	}()

	// Simultaneously access from this goroutine
	for range 100 {
		c.Value()
		c.Match(func(v int32) bool { return v >= 0 })
		c.Signal()
	}

	<-done
	// If we reach here without the race detector triggering, success
}

// TestCountStress performs stress testing on the Count implementation.
func TestCountStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	t.Run("MultipleWaitersAndUpdaters", runTestMultipleWaitersAndUpdaters)
}

func runTestMultipleWaitersAndUpdaters(t *testing.T) {
	c := NewCount(0)
	defer c.Close()

	const (
		numWaiters  = 20
		numUpdaters = 10
		numUpdates  = 1000
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(numWaiters + numUpdaters)

	// Start waiter goroutines
	for i := 0; i < numWaiters; i++ {
		go func(id int) {
			defer wg.Done()

			// Each waiter watches for a different condition
			target := int32((id % 10) * 100)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Wait until counter reaches or exceeds target
					err := c.WaitFnContext(ctx, func(v int32) bool {
						return v >= target
					})
					if err != nil {
						return // Context cancelled
					}

					// Reset target to a higher value for next wait
					target += 100
				}
			}
		}(i)
	}

	// Start updater goroutines
	for i := 0; i < numUpdaters; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < numUpdates; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					c.Inc()
					time.Sleep(time.Microsecond) // Small delay to reduce contention
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Final value should be numUpdaters * numUpdates
	finalValue := c.Value()
	expectedValue := numUpdaters * numUpdates
	assert.Equal(t, expectedValue, finalValue,
		"Expected final value %d, got %d", expectedValue, finalValue)
}

// TestCountReset tests the Reset functionality.
func TestCountReset(t *testing.T) {
	t.Run("BasicReset", runTestBasicReset)
	t.Run("ResetUninitialized", runTestResetUninitialized)
	t.Run("ResetNil", runTestResetNil)
	t.Run("ResetWithWaiters", runTestResetWithWaiters)
	t.Run("ResetClosed", runTestResetClosed)
}

func runTestBasicReset(t *testing.T) {
	c := NewCount(10)
	defer c.Close()

	assert.Equal(t, 10, c.Value())

	err := c.Reset(42)
	require.NoError(t, err)

	assert.Equal(t, 42, c.Value())
}

func runTestResetUninitialized(t *testing.T) {
	c := new(Count)

	err := c.Reset(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialised")
}

func runTestResetNil(t *testing.T) {
	var c *Count

	err := c.Reset(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil receiver")
}

func runTestResetWithWaiters(t *testing.T) {
	c := NewCount(5)
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start a goroutine that waits for value to be zero
	go func() {
		defer wg.Done()
		c.WaitFn(func(v int32) bool { return v == 0 })
	}()

	// Give time for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Reset to 0, which should wake up the waiter
	err := c.Reset(0)
	require.NoError(t, err)

	// Ensure the waiter completes
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		// Success - waiter was notified
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Waiter was not notified when Reset was called")
	}
}

func runTestResetClosed(t *testing.T) {
	c := NewCount(10)

	// Close the count
	require.NoError(t, c.Close())

	// Reset after close should fail
	err := c.Reset(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

// TestCountCustomInitialisation tests initialising Count with various
// custom broadcast conditions.
func TestCountCustomInitialisation(t *testing.T) {
	// Test with multiple broadcast conditions
	t.Run("MultipleBroadcastConditions", runTestMultipleBroadcastConditions)

	// Test with no broadcast conditions (special case)
	t.Run("NoBroadcastConditions", runTestNoBroadcastConditions)
}

func runTestMultipleBroadcastConditions(t *testing.T) {
	isZero := func(v int32) bool { return v == 0 }
	isPositive := func(v int32) bool { return v > 0 }
	isNegative := func(v int32) bool { return v < 0 }

	c := NewCount(0, isZero, isPositive, isNegative)

	var notified atomic.Uint32

	signalCounter := func() {
		notified.Add(1)
	}

	// Test each condition triggers broadcast
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.WaitFn(func(_ int32) bool {
			signalCounter()
			return false // Keep waiting
		})
	}()
	time.Sleep(10 * time.Millisecond)

	// Value is zero - should broadcast
	prevCount := notified.Load()
	c.Add(-1) // Should broadcast due to isZero
	time.Sleep(10 * time.Millisecond)
	assert.Greater(t, notified.Load(), prevCount,
		"Notification count should increase when zero broadcast condition is met")

	// Change to positive - should broadcast
	prevCount = notified.Load()
	c.Inc() // Now 1
	assert.Eventually(t, func() bool {
		return notified.Load() > prevCount
	}, 20*time.Millisecond, 1*time.Millisecond,
		"Notification count should increase when positive broadcast condition is met")

	// Change to negative - should broadcast
	prevCount = notified.Load()
	c.Add(-2) // Now -1
	assert.Eventually(t, func() bool {
		return notified.Load() > prevCount
	}, 20*time.Millisecond, 1*time.Millisecond,
		"Notification count should increase when negative broadcast condition is met")

	// Back to zero - should broadcast
	prevCount = notified.Load()
	c.Inc() // Now 0
	assert.Eventually(t, func() bool {
		return notified.Load() > prevCount
	}, 20*time.Millisecond, 1*time.Millisecond,
		"Notification count should increase when zero broadcast condition is met again")

	// Clean up goroutine
	_ = c.Close()
	wg.Wait()
}

func runTestNoBroadcastConditions(t *testing.T) {
	// Create a Count with an explicit empty broadcast condition list
	c := NewCount(0, []func(int32) bool{}...)

	// Every value change should broadcast
	var wg sync.WaitGroup
	var mu sync.Mutex
	var notified int

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.WaitFn(func(_ int32) bool {
			mu.Lock()
			notified++
			mu.Unlock()
			return false // Keep waiting
		})
	}()

	// Wait for goroutine to start waiting
	time.Sleep(10 * time.Millisecond)

	// Every operation should broadcast
	mu.Lock()
	prevCount := notified
	mu.Unlock()
	c.Inc()
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	currentCount := notified
	mu.Unlock()
	assert.Greater(t, currentCount, prevCount,
		"Notification count should increase after Inc() with no broadcast conditions")

	// Similar pattern for all other checks
	// (add mutex protection for all accesses to notified)

	// Clean up goroutine
	_ = c.Close()
	wg.Wait()
}

// TestWaitResult tests the waitResult type functionality.
func TestWaitResult(t *testing.T) {
	// Test IsCancelled method
	t.Run("IsCancelled", func(t *testing.T) {
		assert.False(t, waitContinue.IsCancelled())
		assert.False(t, waitSuccess.IsCancelled())
		assert.True(t, waitCancelled.IsCancelled())
	})

	// Test IsContinue method
	t.Run("IsContinue", func(t *testing.T) {
		assert.True(t, waitContinue.IsContinue())
		assert.False(t, waitSuccess.IsContinue())
		assert.False(t, waitCancelled.IsContinue())
	})
}

// TestCountClosed tests the IsClosed functionality.
func TestCountClosed(t *testing.T) {
	t.Run("BeforeClose", runTestBeforeClose)
	t.Run("AfterClose", runTestAfterClose)
	t.Run("NilCount", runTestNilCountIsClosed)
	t.Run("UnitialisedCount", runTestUnitialisedCountIsClosed)
}

func runTestBeforeClose(t *testing.T) {
	c := NewCount(42)
	require.NotNil(t, c)

	// A newly created Count should not be closed
	assert.False(t, c.IsClosed())

	// Performing operations should not affect closed state
	c.Inc()
	assert.False(t, c.IsClosed())

	c.Dec()
	assert.False(t, c.IsClosed())

	c.Add(10)
	assert.False(t, c.IsClosed())

	// Clean up
	require.NoError(t, c.Close())
}

func runTestAfterClose(t *testing.T) {
	c := NewCount(42)
	require.NotNil(t, c)

	// Close the Count
	require.NoError(t, c.Close())

	// After closing, IsClosed should return true
	assert.True(t, c.IsClosed())

	// Operations on a closed Count should return errors
	assert.Error(t, c.Reset(0))

	// IsClosed should remain true
	assert.True(t, c.IsClosed())
}

func runTestNilCountIsClosed(t *testing.T) {
	var c *Count

	// A nil Count should report as closed
	assert.True(t, c.IsClosed())
}

func runTestUnitialisedCountIsClosed(t *testing.T) {
	c := new(Count)

	// An uninitialised Count should report as closed
	assert.True(t, c.IsClosed())

	// After initialisation, it should no longer be closed
	err := c.Init(0)
	require.NoError(t, err)
	assert.False(t, c.IsClosed())

	// Clean up
	require.NoError(t, c.Close())

	// After closing, it should report as closed again
	assert.True(t, c.IsClosed())
}

// atomicMaskOr performs a bitwise OR operation on an atomic uint32 value
// using a compare-and-swap loop.
// It attempts to set the specified mask bits atomically and returns
// the new mask value.
// The second return value indicates whether the mask was modified (true)
// or was already set (false).
func atomicMaskOr(ptr *atomic.Uint32, mask uint32) (uint32, bool) {
	for {
		oldMask := ptr.Load()
		newMask := oldMask | mask
		if oldMask == newMask {
			return newMask, false
		} else if ptr.CompareAndSwap(oldMask, newMask) {
			return newMask, true
		}
	}
}
