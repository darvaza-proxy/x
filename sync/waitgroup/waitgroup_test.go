package waitgroup

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWaitGroupBasics verifies the essential functionality of a WaitGroup.
func TestWaitGroupBasics(t *testing.T) {
	var wg WaitGroup

	// Initial state should be 0
	assert.Equal(t, 0, wg.Count(), "initial count should be zero")

	// Add workers
	wg.Add(3)
	assert.Equal(t, 3, wg.Count(), "count should reflect added workers")

	// Complete some work
	wg.Done()
	assert.Equal(t, 2, wg.Count(), "count should decrease after Done")
	wg.Done()
	assert.Equal(t, 1, wg.Count(), "count should decrease after Done")

	// // a goroutine to wait
	waitFinished := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitFinished)
	}()

	// Wait shouldn't finish yet
	select {
	case <-waitFinished:
		t.Fatal("Wait returned before all workers completed")
	case <-time.After(50 * time.Millisecond):
		// Expected behaviour
	}

	// Complete the last worker
	wg.Done()

	// Wait should finish
	select {
	case <-waitFinished:
		// Expected behaviour
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait didn't return after all workers completed")
	}

	// After Wait, the WaitGroup should be reusable
	assert.Equal(t, 0, wg.Count(), "count should be reset after Wait")
}

// TestWaitGroupWithGoroutines tests coordination of multiple goroutines.
func TestWaitGroupWithGoroutines(t *testing.T) {
	var wg WaitGroup
	const numGoroutines = 10

	var counter int32

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
			// Simulate work
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(numGoroutines), counter,
		"all goroutines should have incremented the counter")
	assert.Equal(t, 0, wg.Count(), "count should be 0 after Wait")
}

// TestNilWaitGroup examines the behaviour of a nil WaitGroup.
func TestNilWaitGroup(t *testing.T) {
	var wg *WaitGroup = nil

	// Count should return 0 for nil receiver
	assert.Equal(t, 0, wg.Count(), "nil waitgroup should return count 0")

	// Add should panic with nil receiver
	assert.Panics(t, func() {
		wg.Add(1)
	}, "Add on nil WaitGroup should panic")

	// Done should panic with nil receiver
	assert.Panics(t, func() {
		wg.Done()
	}, "Done on nil WaitGroup should panic")

	// Wait should panic with nil receiver
	assert.Panics(t, func() {
		wg.Wait()
	}, "Wait on nil WaitGroup should panic")
}

// TestWaitGroupInvalidAdd tests adding invalid values.
func TestWaitGroupInvalidAdd(t *testing.T) {
	var wg WaitGroup

	// Adding zero has no effect
	wg.Add(0)
	assert.Equal(t, 0, wg.Count(), "count should remain 0 after adding 0")

	// Adding negative value should panic
	assert.Panics(t, func() {
		wg.Add(-1)
	}, "adding negative value should panic")
}

// TestWaitGroupDoneWithoutAdd tests calling Done without a preceding Add.
func TestWaitGroupDoneWithoutAdd(t *testing.T) {
	var wg WaitGroup

	// Done without a preceding Add should panic
	assert.Panics(t, func() {
		wg.Done()
	}, "Done without Add should panic")
}

// TestWaitGroupReuse verifies a Wait Wait can be reused after Wait completes.
func TestWaitGroupReuse(t *testing.T) {
	var wg WaitGroup

	// Use it once
	wg.Add(1)
	wg.Done()
	wg.Wait()

	// Use it again
	wg.Add(2)
	assert.Equal(t, 2, wg.Count(), "count should reflect new workers after reuse")

	wg.Done()
	wg.Done()
	wg.Wait()

	assert.Equal(t, 0, wg.Count(), "count should be 0 after second Wait")
}

// TestWaitGroupConcurrentAdd tests adding workers from multiple goroutines.
func TestWaitGroupConcurrentAdd(t *testing.T) {
	var wg WaitGroup
	const numGoroutines = 10

	// Start goroutines that each add one worker
	for i := 0; i < numGoroutines; i++ {
		go func() {
			wg.Add(1)
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		}()
	}

	// Allow time for goroutines to start and add workers
	time.Sleep(30 * time.Millisecond)

	// Wait for all to complete
	wg.Wait()
	assert.Equal(t, 0, wg.Count(),
		"count should be 0 after all goroutines complete")
}

// TestWaitGroupStress runs a high concurrency test.
func TestWaitGroupStress(t *testing.T) {
	var wg WaitGroup
	const numGoroutines = 1000

	// Limit runtime to avoid test timeouts on slower systems
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			runtime.Gosched() // Increase contention
			defer wg.Done()

			// Simulate varied work
			if i%2 == 0 {
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 0, wg.Count(), "count should be 0 after stress test")
}

// TestWaitGroupAddAfterWaitStarts tests adding workers during Wait execution.
func TestWaitGroupAddAfterWaitStarts(t *testing.T) {
	var wg WaitGroup
	wg.Add(1)

	// Start waiting in background
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	// Brief pause to ensure Wait() is running
	time.Sleep(20 * time.Millisecond)

	// Add another worker during Wait execution
	wg.Add(1)
	assert.Equal(t, 2, wg.Count(), "count should include new worker")

	// Done for both workers
	wg.Done()
	wg.Done()

	// Wait should complete
	select {
	case <-waitCh:
		// Expected behaviour
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait didn't complete after all Done calls")
	}
}

// TestWaitGroupMultipleWaiters verifies multiple goroutines can wait.
func TestWaitGroupMultipleWaiters(t *testing.T) {
	var wg WaitGroup
	const numWaiters = 5

	wg.Add(1)

	// Create multiple waiter goroutines
	var waitersFinished int32
	for i := 0; i < numWaiters; i++ {
		go func() {
			wg.Wait()
			atomic.AddInt32(&waitersFinished, 1)
		}()
	}

	// Allow waiters time to start waiting
	time.Sleep(20 * time.Millisecond)

	// Confirm not finished yet
	assert.Equal(t, int32(0), atomic.LoadInt32(&waitersFinished),
		"no waiters should have finished yet")

	// Complete the work
	wg.Done()

	// All waiters should finish
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(numWaiters), atomic.LoadInt32(&waitersFinished),
		"all waiters should have finished")
}

// TestWaitGroupLargeCount tests the WaitGroup with a large count.
func TestWaitGroupLargeCount(t *testing.T) {
	var wg WaitGroup
	const largeCount = 10000

	wg.Add(largeCount)
	assert.Equal(t, largeCount, wg.Count(),
		"count should match large number of workers")

	// Complete all work in batches
	for i := 0; i < largeCount; i++ {
		wg.Done()
	}

	// Wait should complete immediately since all work is done
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// Expected behaviour
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait didn't complete promptly after all workers finished")
	}
}

// TestWaitGroupZeroValue confirms a zero value WaitGroup works correctly.
func TestWaitGroupZeroValue(t *testing.T) {
	var wg WaitGroup

	// Wait on a zero value should return immediately
	wg.Wait()

	// Should be able to add and use normally after Wait
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait shouldn't complete yet
	select {
	case <-done:
		t.Fatal("Wait returned before worker completed")
	case <-time.After(50 * time.Millisecond):
		// Expected behaviour
	}

	wg.Done()

	// Now Wait should complete
	select {
	case <-done:
		// Expected behaviour
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Wait didn't return after worker completed")
	}
}

// TestWaitGroupAddDoneInternal examines the internal doAdd and doDec methods
// by checking the returned errors.
func TestWaitGroupAddDoneInternal(t *testing.T) {
	var wg WaitGroup
	var nilWg *WaitGroup

	// Test doAdd
	err := wg.doAdd(1)
	assert.NoError(t, err, "doAdd with valid value should succeed")

	err = wg.doAdd(0)
	assert.NoError(t, err, "doAdd with zero should succeed")

	err = wg.doAdd(-1)
	assert.Error(t, err, "doAdd with negative value should return error")

	err = nilWg.doAdd(1)
	assert.Error(t, err, "doAdd on nil should return error")

	// Test doDec
	err = wg.doDec()
	assert.NoError(t, err, "doDec on valid count should succeed")

	err = wg.doDec()
	assert.Error(t, err, "doDec beyond zero should return error")

	err = nilWg.doDec()
	assert.Error(t, err, "doDec on nil should return error")
}
