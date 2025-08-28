package workgroup

//revive:disable:cognitive-complexity

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"

	"github.com/stretchr/testify/assert"
)

// TestNew verifies the Group initialisation functionality
func TestNew(t *testing.T) {
	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		wg := New(ctx)

		assert.NotNil(t, wg)
		assert.Equal(t, ctx, wg.Parent)
		assert.NotNil(t, wg.ctx)
		assert.NotNil(t, wg.cancel)
		assert.False(t, wg.cancelled.Load())
	})

	t.Run("WithNilContext", func(t *testing.T) {
		var nilCtx context.Context
		wg := New(nilCtx)

		assert.NotNil(t, wg)
		assert.Equal(t, context.Background(), wg.Parent)
		assert.NotNil(t, wg.ctx)
		assert.NotNil(t, wg.cancel)
		assert.False(t, wg.cancelled.Load())
	})
}

// TestGroup_Context tests the Context method
func TestGroup_Context(t *testing.T) {
	t.Run("ValidContext", func(t *testing.T) {
		wg := New(context.Background())
		ctx := wg.Context()

		assert.NotNil(t, ctx)
		assert.Nil(t, ctx.Err())
	})

	t.Run("PanicsOnNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Panics(t, func() {
			wg.Context()
		})
	})

	t.Run("CancelledWhenGroupCancelled", func(t *testing.T) {
		wg := New(context.Background())
		ctx := wg.Context()

		wg.Cancel(nil)
		assert.NotNil(t, ctx.Err())
		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("CancelledWhenParentCancelled", func(t *testing.T) {
		parentCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := New(parentCtx)
		ctx := wg.Context()

		cancel()
		assert.Eventually(t, func() bool {
			return ctx.Err() == context.Canceled
		}, 10*time.Millisecond, 1*time.Millisecond)
	})
}

// TestGroup_Err tests the Err method
func TestGroup_Err(t *testing.T) {
	t.Run("NilWhenActive", func(t *testing.T) {
		wg := New(context.Background())
		assert.Nil(t, wg.Err())
	})

	t.Run("ReturnsErrOnNilReceiver", func(t *testing.T) {
		var wg *Group
		err := wg.Err()
		assert.Error(t, err)
	})

	t.Run("ReturnsCustomCancelError", func(t *testing.T) {
		wg := New(context.Background())
		customErr := errors.New("custom cancel error")
		wg.Cancel(customErr)

		err := wg.Err()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, customErr))
	})

	t.Run("ReturnsCanceledWhenNoCustomError", func(t *testing.T) {
		wg := New(context.Background())
		wg.Cancel(nil)

		err := wg.Err()
		assert.Equal(t, context.Canceled, err)
	})
}

// TestGroup_IsCancelled tests the IsCancelled method
func TestGroup_IsCancelled(t *testing.T) {
	t.Run("FalseWhenActive", func(t *testing.T) {
		wg := New(context.Background())
		assert.False(t, wg.IsCancelled())
	})

	t.Run("PanicsOnNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Panics(t, func() {
			wg.IsCancelled()
		})
	})

	t.Run("TrueAfterCancel", func(t *testing.T) {
		wg := New(context.Background())
		wg.Cancel(nil)
		assert.True(t, wg.IsCancelled())
	})

	t.Run("TrueAfterParentContextCancelled", func(t *testing.T) {
		parentCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := New(parentCtx)

		cancel()
		assert.Eventually(t, func() bool {
			return wg.IsCancelled()
		}, 10*time.Millisecond, 1*time.Millisecond)
	})
}

// TestGroup_Cancelled tests the Cancelled method
func TestGroup_Cancelled(t *testing.T) {
	t.Run("ReturnsChannel", func(t *testing.T) {
		wg := New(context.Background())
		ch := wg.Cancelled()

		assert.NotNil(t, ch)

		// Channel should not be closed initially
		select {
		case <-ch:
			assert.Fail(t, "channel should not be closed")
		default:
			// This is expected
		}
	})

	t.Run("PanicsOnNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Panics(t, func() {
			wg.Cancelled()
		})
	})

	t.Run("ChannelClosedAfterCancel", func(t *testing.T) {
		wg := New(context.Background())
		ch := wg.Cancelled()

		wg.Cancel(nil)

		// Channel should be closed now
		select {
		case <-ch:
			// This is expected
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "channel should be closed after cancellation")
		}
	})
}

// TestGroup_Done tests the Done method
func TestGroup_Done(t *testing.T) {
	t.Run("ChannelCreated", func(t *testing.T) {
		wg := New(context.Background())
		ch := wg.Done()
		assert.NotNil(t, ch)
	})

	t.Run("PanicsOnNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Panics(t, func() {
			wg.Done()
		})
	})

	t.Run("ChannelClosedWhenNoTasks", func(t *testing.T) {
		wg := New(context.Background())
		ch := wg.Done()

		// Channel should close quickly as there are no tasks
		select {
		case <-ch:
			// Expected behaviour
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Done channel should close when no tasks")
		}
	})

	t.Run("ChannelClosedAfterTasksComplete", func(t *testing.T) {
		wg := New(context.Background())

		// Add a task that sleeps briefly
		_ = wg.Go(func(_ context.Context) {
			time.Sleep(50 * time.Millisecond)
		})

		ch := wg.Done()

		// Channel shouldn't be closed immediately
		select {
		case <-ch:
			assert.Fail(t, "Done channel should not close before tasks complete")
		case <-time.After(10 * time.Millisecond):
			// Expected behaviour
		}

		// Channel should be closed after task completes
		select {
		case <-ch:
			// Expected behaviour
		case <-time.After(200 * time.Millisecond):
			assert.Fail(t, "Done channel should close after tasks complete")
		}
	})

	t.Run("MultipleCallsReturnSameChannel", func(t *testing.T) {
		wg := New(context.Background())
		ch1 := wg.Done()
		ch2 := wg.Done()

		// Both channels should be the same
		assert.Equal(t, ch1, ch2)
	})

	t.Run("NewChannelAfterCompletion", func(t *testing.T) {
		wg := New(context.Background())
		ch1 := wg.Done()

		// Wait for the Done channel to close
		<-ch1

		// Add a new task and get a new Done channel
		_ = wg.Go(func(_ context.Context) {
			time.Sleep(10 * time.Millisecond)
		})

		ch2 := wg.Done()

		// Should be a different channel
		assert.NotEqual(t, ch1, ch2)
	})
}

func TestGroup_Wait(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", func(t *testing.T) {
		var wg *Group
		err := wg.Wait()
		assert.Error(t, err)
	})

	t.Run("ReturnsNilForActiveGroup", func(t *testing.T) {
		wg := New(context.Background())
		err := wg.Wait()
		assert.Nil(t, err)
	})

	t.Run("WaitsForAllTasks", func(t *testing.T) {
		wg := New(context.Background())

		counter := atomic.Int32{}
		numTasks := 5

		for i := 0; i < numTasks; i++ {
			_ = wg.Go(func(_ context.Context) {
				time.Sleep(10 * time.Millisecond)
				counter.Add(1)
			})
		}

		err := wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, int32(numTasks), counter.Load())
	})

	t.Run("ReturnsNilOnContextCanceled", func(t *testing.T) {
		wg := New(context.Background())

		// Add a task that should never complete on its own
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
		})

		// Cancel the group and then wait
		wg.Cancel(nil)
		err := wg.Wait()

		assert.Nil(t, err)
	})

	t.Run("ReturnsCustomErrorWhenCancelled", func(t *testing.T) {
		wg := New(context.Background())
		customErr := errors.New("custom cancel reason")

		// Add a task that should never complete on its own
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
		})

		// Cancel with custom error and then wait
		wg.Cancel(customErr)
		err := wg.Wait()

		assert.Error(t, err)
		assert.True(t, errors.Is(err, customErr))
	})
}

func TestGroup_Cancel(t *testing.T) {
	t.Run("PanicsOnNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Panics(t, func() {
			wg.Cancel(nil)
		})
	})

	t.Run("ReturnsTrueOnFirstCancel", func(t *testing.T) {
		wg := New(context.Background())
		result := wg.Cancel(nil)
		assert.True(t, result)
	})

	t.Run("ReturnsFalseOnSubsequentCancel", func(t *testing.T) {
		wg := New(context.Background())
		wg.Cancel(nil)
		result := wg.Cancel(nil)
		assert.False(t, result)
	})

	t.Run("CancelsAllTasks", func(t *testing.T) {
		wg := New(context.Background())

		tasksCancelled := atomic.Int32{}
		numTasks := 5

		for i := 0; i < numTasks; i++ {
			_ = wg.Go(func(ctx context.Context) {
				<-ctx.Done()
				tasksCancelled.Add(1)
			})
		}

		wg.Cancel(nil)
		_ = wg.Wait()

		assert.Equal(t, int32(numTasks), tasksCancelled.Load())
	})

	t.Run("PropagatesCustomError", func(t *testing.T) {
		wg := New(context.Background())
		customErr := errors.New("custom error")

		var receivedErr error

		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
			receivedErr = context.Cause(ctx)
		})

		wg.Cancel(customErr)
		err := wg.Wait()

		assert.Equal(t, customErr, receivedErr)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, customErr))
	})
}

func TestGroup_Close(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", func(t *testing.T) {
		var wg *Group
		err := wg.Close()
		assert.Error(t, err)
	})

	t.Run("CancelsAndWaitsForTasks", func(t *testing.T) {
		wg := New(context.Background())

		tasksCancelled := atomic.Int32{}
		numTasks := 5

		for i := 0; i < numTasks; i++ {
			_ = wg.Go(func(ctx context.Context) {
				<-ctx.Done()
				time.Sleep(10 * time.Millisecond) // Simulate cleanup work
				tasksCancelled.Add(1)
			})
		}

		err := wg.Close()
		assert.Nil(t, err)
		assert.Equal(t, int32(numTasks), tasksCancelled.Load())
	})

	t.Run("CanBeCalledMultipleTimes", func(t *testing.T) {
		wg := New(context.Background())

		// Add a task
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
		})

		// Should not error on multiple closes
		err1 := wg.Close()
		err2 := wg.Close()

		assert.Nil(t, err1)
		assert.Nil(t, err2)
	})
}

func TestGroup_Go(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", func(t *testing.T) {
		var wg *Group
		err := wg.Go(func(_ context.Context) {})
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("DoesNothingWithNilFunc", func(t *testing.T) {
		wg := New(context.Background())
		err := wg.Go(nil)
		assert.Nil(t, err, "Go with nil func should return nil error")

		// Wait should return immediately as no task was added
		done := make(chan struct{})
		go func() {
			_ = wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Expected - should return quickly
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Wait should return immediately with no tasks")
		}
	})

	t.Run("ExecutesTask", func(t *testing.T) {
		wg := New(context.Background())

		executed := atomic.Bool{}
		err := wg.Go(func(_ context.Context) {
			executed.Store(true)
		})
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Nil(t, err)
		assert.True(t, executed.Load())
	})

	t.Run("ExecutesMultipleTasks", func(t *testing.T) {
		wg := New(context.Background())

		counter := atomic.Int32{}
		numTasks := 10

		for range numTasks {
			err := wg.Go(func(_ context.Context) {
				counter.Add(1)
			})
			assert.Nil(t, err)
		}

		err := wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, int32(numTasks), counter.Load())
	})

	t.Run("TasksReceiveGroupContext", func(t *testing.T) {
		wg := New(context.Background())

		var receivedCtx context.Context

		err := wg.Go(func(ctx context.Context) {
			receivedCtx = ctx
		})
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, wg.Context(), receivedCtx)
	})

	t.Run("ReturnsErrorWhenGroupCancelled", func(t *testing.T) {
		wg := New(context.Background())
		wg.Cancel(nil)

		err := wg.Go(func(_ context.Context) {
			// This should not execute
		})

		assert.Error(t, err)
		assert.Equal(t, errors.ErrClosed, err)
	})
}

// TestWorkgroup_Integration tests more complex integration scenarios
func TestWorkgroup_Integration(t *testing.T) {
	t.Run("ReuseAfterCompletion", func(t *testing.T) {
		wg := New(context.Background())

		// First batch of tasks
		counter1 := atomic.Int32{}
		for range 3 {
			err := wg.Go(func(_ context.Context) {
				counter1.Add(1)
			})
			assert.Nil(t, err)
		}

		err := wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, int32(3), counter1.Load())

		// Second batch of tasks with the same workgroup
		counter2 := atomic.Int32{}
		for range 5 {
			err := wg.Go(func(_ context.Context) {
				counter2.Add(1)
			})
			assert.Nil(t, err)
		}

		err = wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, int32(5), counter2.Load())
	})

	t.Run("ConcurrentTasksAndCancel", func(t *testing.T) {
		wg := New(context.Background())

		// Start a bunch of tasks that wait to be cancelled
		numTasks := 100
		tasksStarted := atomic.Int32{}
		tasksCancelled := atomic.Int32{}

		for range numTasks {
			err := wg.Go(func(ctx context.Context) {
				tasksStarted.Add(1)
				<-ctx.Done()
				tasksCancelled.Add(1)
			})
			assert.Nil(t, err)
		}

		// Wait until all tasks have started
		assert.Eventually(t,
			func() bool { return tasksStarted.Load() == int32(numTasks) },
			100*time.Millisecond,
			1*time.Millisecond)

		// Cancel the group
		wg.Cancel(errors.New("cancel all tasks"))
		err := wg.Wait()
		assert.NotNil(t, err)

		assert.Equal(t, int32(numTasks), tasksCancelled.Load())
	})

	t.Run("CascadingCancellation", func(t *testing.T) {
		// Create a hierarchy of contexts and groups
		rootCtx, rootCancel := context.WithCancel(context.Background())
		defer rootCancel()

		parentWg := New(rootCtx)
		childWg := New(parentWg.Context())
		grandchildWg := New(childWg.Context())

		// Add tasks at each level
		var parentCancelled, childCancelled, grandchildCancelled atomic.Bool

		err := parentWg.Go(func(ctx context.Context) {
			<-ctx.Done()
			parentCancelled.Store(true)
		})
		assert.Nil(t, err)

		err = childWg.Go(func(ctx context.Context) {
			<-ctx.Done()
			childCancelled.Store(true)
		})
		assert.Nil(t, err)

		err = grandchildWg.Go(func(ctx context.Context) {
			<-ctx.Done()
			grandchildCancelled.Store(true)
		})
		assert.Nil(t, err)

		// Cancel the root context
		rootCancel()

		// All workgroups should be cancelled
		_ = parentWg.Wait()
		_ = childWg.Wait()
		_ = grandchildWg.Wait()

		assert.True(t, parentCancelled.Load())
		assert.True(t, childCancelled.Load())
		assert.True(t, grandchildCancelled.Load())
	})
}

// TestGroup_LazyInit tests the lazyInit behaviour indirectly
func TestGroup_LazyInit(t *testing.T) {
	t.Run("InitialisesEmptyGroup", func(t *testing.T) {
		// Create a group but don't initialise it with New
		wg := &Group{}

		// This should trigger lazyInit
		ctx := wg.Context()

		// Should now be initialised
		assert.NotNil(t, ctx)
		assert.NotNil(t, wg.ctx)
		assert.NotNil(t, wg.cancel)
		assert.Equal(t, context.Background(), wg.Parent)
	})

	t.Run("InitialisesWithCustomParent", func(t *testing.T) {
		ctxKey := core.NewContextKey[string]("key")
		parentCtx := ctxKey.WithValue(context.Background(), "value")

		// Create a group with parent but don't initialise
		wg := &Group{Parent: parentCtx}

		// This should trigger lazyInit
		ctx := wg.Context()

		// Should now be initialised with the custom parent
		assert.NotNil(t, ctx)
		assert.Equal(t, "value", ctx.Value(ctxKey))
	})
}

// TestGroup_Timeout tests behaviour with timeouts
func TestGroup_Timeout(t *testing.T) {
	t.Run("GroupCancelledWhenTimeoutExpires", func(t *testing.T) {
		// Create context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		wg := New(ctx)

		// Add a task that would run forever if not cancelled
		taskCancelled := atomic.Bool{}
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
			taskCancelled.Store(true)
		})

		// Wait should return when the timeout expires
		start := time.Now()
		err := wg.Wait()
		elapsed := time.Since(start)

		assert.True(t, elapsed >= 50*time.Millisecond,
			"Wait should block until timeout")
		assert.True(t, elapsed < 200*time.Millisecond,
			"Wait shouldn't block much longer than timeout")
		assert.NotNil(t, err)
		assert.True(t, taskCancelled.Load(), "Task should be cancelled")
	})
}

// TestGroup_Concurrent tests concurrent behaviour
func TestGroup_Concurrent(t *testing.T) {
	t.Run("ConcurrentGoAndCancel", func(t *testing.T) {
		wg := New(context.Background())

		// Launch a bunch of goroutines that add tasks
		const numRoutines = 10
		const tasksPerRoutine = 100

		var startWg, endWg sync.WaitGroup
		startWg.Add(numRoutines + 1) // +1 for the cancelling goroutine
		endWg.Add(numRoutines + 1)

		// Track number of tasks that were cancelled vs completed
		cancelled := atomic.Int32{}
		completed := atomic.Int32{}

		// Launch goroutines that add tasks
		for range numRoutines {
			go func() {
				defer endWg.Done()
				startWg.Done()
				startWg.Wait() // Wait for all goroutines to be ready

				for range tasksPerRoutine {
					_ = wg.Go(func(ctx context.Context) {
						// Each task either completes or gets cancelled
						select {
						case <-ctx.Done():
							cancelled.Add(1)
						case <-time.After(10 * time.Millisecond):
							completed.Add(1)
						}
					})

					// Small sleep to interleave with cancellation
					time.Sleep(time.Millisecond)
				}
			}()
		}

		// Launch a goroutine that cancels after some tasks are added
		go func() {
			defer endWg.Done()
			startWg.Done()
			startWg.Wait() // Wait for all goroutines to be ready

			// Wait a bit to let some tasks be added
			time.Sleep(50 * time.Millisecond)

			// Cancel the group
			wg.Cancel(errors.New("concurrent cancellation"))
		}()

		// Wait for all the task-adding goroutines to finish
		endWg.Wait()

		// Wait for all tasks to complete or be cancelled
		_ = wg.Wait()

		// Some tasks should have been cancelled
		assert.True(t, cancelled.Load() > 0, "Some tasks should be cancelled")
		t.Logf("Tasks completed: %d, cancelled: %d", completed.Load(),
			cancelled.Load())
	})

	t.Run("ConcurrentDoneChannels", func(t *testing.T) {
		wg := New(context.Background())

		// Add a long-running task
		_ = wg.Go(func(_ context.Context) {
			time.Sleep(100 * time.Millisecond)
		})

		// Get Done channels concurrently
		const numGoroutines = 10
		channels := make([]<-chan struct{}, numGoroutines)

		var startWg sync.WaitGroup
		startWg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				defer startWg.Done()
				channels[idx] = wg.Done()
			}(i)
		}

		startWg.Wait()

		// All channels should be the same
		for i := 1; i < numGoroutines; i++ {
			assert.Equal(t, channels[0], channels[i],
				"All Done channels should be the same instance")
		}

		// All channels should close when the task completes
		<-channels[0]

		// Check that the task is really done
		err := wg.Wait()
		assert.Nil(t, err)
	})
}

// testNilReceiverBehaviour is a helper function to test behaviour with nil Group receivers
func testNilReceiverBehaviour(t *testing.T) {
	t.Helper()
	var wg *Group

	// Methods that return errors
	err1 := wg.Wait()
	err2 := wg.Close()
	err3 := wg.Err()
	err4 := wg.Go(func(_ context.Context) {})

	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Error(t, err3)
	assert.Error(t, err4)
	assert.Equal(t, errors.ErrNilReceiver, err4)

	// Methods that panic
	assert.Panics(t, func() { wg.Context() })
	assert.Panics(t, func() { wg.IsCancelled() })
	assert.Panics(t, func() { wg.Cancelled() })
	assert.Panics(t, func() { wg.Done() })
	assert.Panics(t, func() { wg.Cancel(nil) })
}

// TestGroup_NilReceiver tests behaviour with nil receiver
func TestGroup_NilReceiver(t *testing.T) {
	testNilReceiverBehaviour(t)
}

// TestGroup_ErrorHandling tests various error patterns
func TestGroup_ErrorHandling(t *testing.T) {
	t.Run("NilGroups", func(t *testing.T) {
		testNilReceiverBehaviour(t)
	})

	t.Run("CustomErrorTypes", func(t *testing.T) {
		// Test with custom error types including wrapped errors
		wg := New(context.Background())

		// Create a custom error
		myErr := errors.New("custom error type")

		// Add a task that waits for cancellation
		var receivedErr error
		err := wg.Go(func(ctx context.Context) {
			<-ctx.Done()
			receivedErr = context.Cause(ctx)
		})
		assert.Nil(t, err)

		// Cancel with the custom error
		wg.Cancel(myErr)
		_ = wg.Wait()

		// The error should be propagated correctly
		assert.Equal(t, myErr, receivedErr)
	})
}

// TestWorkgroup_ContextValues tests passing context values
func TestWorkgroup_ContextValues(t *testing.T) {
	// Create a context with a value
	type key string
	testKey := key("test-key")
	testVal := "test-value"

	ctx := context.WithValue(context.Background(), testKey, testVal)
	wg := New(ctx)

	// Verify the value is accessible in the task
	var receivedVal any
	_ = wg.Go(func(ctx context.Context) {
		receivedVal = ctx.Value(testKey)
	})

	err := wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, testVal, receivedVal)
}

// TestGroup_Reuse tests that a Group can be reused
func TestGroup_Reuse(t *testing.T) {
	wg := New(context.Background())

	// First use
	counter1 := atomic.Int32{}
	_ = wg.Go(func(_ context.Context) {
		counter1.Add(1)
	})

	err := wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, int32(1), counter1.Load())

	// Second use
	counter2 := atomic.Int32{}
	_ = wg.Go(func(_ context.Context) {
		counter2.Add(1)
	})

	err = wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, int32(1), counter2.Load())
}

// TestGroup_DeadlineExceeded tests context deadline behaviour
func TestGroup_DeadlineExceeded(t *testing.T) {
	// Create a context with a very short deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	wg := New(ctx)

	// Add a task that should be interrupted
	taskInterrupted := atomic.Bool{}
	_ = wg.Go(func(ctx context.Context) {
		select {
		case <-time.After(200 * time.Millisecond):
			// This should not happen
		case <-ctx.Done():
			taskInterrupted.Store(true)
		}
	})

	// Wait for the deadline to trigger
	err := wg.Wait()

	// The error should be context.DeadlineExceeded
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.True(t, taskInterrupted.Load(), "Task should be interrupted by deadline")
}

// TestGroup_GoCatch tests the GoCatch method
func TestGroup_GoCatch(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", func(t *testing.T) {
		var wg *Group
		err := wg.GoCatch(
			func(_ context.Context) error { return nil },
			func(_ context.Context, err error) error { return err },
		)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("DoesNothingWithNilFunc", func(t *testing.T) {
		wg := New(context.Background())
		err := wg.GoCatch(nil, nil)
		assert.Nil(t, err, "GoCatch with nil func should return nil error")

		// Wait should return immediately as no task was added
		done := make(chan struct{})
		go func() {
			_ = wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Expected - should return quickly
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Wait should return immediately with no tasks")
		}
	})

	t.Run("ExecutesSuccessfulTask", func(t *testing.T) {
		wg := New(context.Background())

		executed := atomic.Bool{}
		err := wg.GoCatch(
			func(_ context.Context) error {
				executed.Store(true)
				return nil
			},
			nil,
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Nil(t, err)
		assert.True(t, executed.Load())
	})

	t.Run("TasksReceiveGroupContext", func(t *testing.T) {
		wg := New(context.Background())

		var receivedCtx context.Context

		err := wg.GoCatch(
			func(ctx context.Context) error {
				receivedCtx = ctx
				return nil
			},
			nil,
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Nil(t, err)
		assert.Equal(t, wg.Context(), receivedCtx)
	})

	t.Run("CancelsGroupOnError", func(t *testing.T) {
		wg := New(context.Background())

		testErr := errors.New("test error")
		err := wg.GoCatch(
			func(_ context.Context) error {
				return testErr
			},
			nil,
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, testErr))
	})

	t.Run("CatchHandlerProcessesError", func(t *testing.T) {
		wg := New(context.Background())

		testErr := errors.New("original error")
		processedErr := errors.New("processed error")

		err := wg.GoCatch(
			func(_ context.Context) error {
				return testErr
			},
			func(_ context.Context, err error) error {
				assert.Equal(t, testErr, err)
				return processedErr
			},
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, processedErr))
		assert.False(t, errors.Is(err, testErr))
	})

	t.Run("CatchHandlerCanPreventCancellation", func(t *testing.T) {
		wg := New(context.Background())

		testErr := errors.New("non-critical error")

		err := wg.GoCatch(
			func(_ context.Context) error {
				return testErr
			},
			func(_ context.Context, err error) error {
				assert.Equal(t, testErr, err)
				// Return nil to prevent group cancellation
				return nil
			},
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Nil(t, err)
		assert.False(t, wg.IsCancelled())
	})

	t.Run("RecoversPanics", func(t *testing.T) {
		wg := New(context.Background())

		panicMsg := "deliberate panic"

		err := wg.GoCatch(
			func(_ context.Context) error {
				panic(panicMsg)
			},
			func(_ context.Context, err error) error {
				return err
			},
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Error(t, err)

		// Should be a PanicError
		var panicErr *core.PanicError
		assert.True(t, errors.As(err, &panicErr))
		assert.Contains(t, err.Error(), panicMsg)
	})

	t.Run("ReturnsErrorWhenGroupCancelled", func(t *testing.T) {
		wg := New(context.Background())
		wg.Cancel(nil)

		err := wg.GoCatch(
			func(_ context.Context) error {
				return nil
			},
			nil,
		)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrClosed, err)
	})

	t.Run("CatchHandlerReceivesContext", func(t *testing.T) {
		wg := New(context.Background())

		var taskCtx, catchCtx context.Context
		testErr := errors.New("test error")

		err := wg.GoCatch(
			func(ctx context.Context) error {
				taskCtx = ctx
				return testErr
			},
			func(ctx context.Context, err error) error {
				catchCtx = ctx
				return err
			},
		)
		assert.Nil(t, err)

		err = wg.Wait()
		assert.Error(t, err)
		assert.Equal(t, taskCtx, wg.Context())
		assert.Equal(t, catchCtx, wg.Context())
	})
}

// TestGroup_Count tests the Count method
func TestGroup_Count(t *testing.T) {
	t.Run("ReportsActiveTaskCount", func(t *testing.T) {
		wg := New(context.Background())

		assert.Equal(t, 0, wg.Count())

		taskStarted := make(chan struct{})
		taskDone := make(chan struct{})

		_ = wg.Go(func(_ context.Context) {
			close(taskStarted)
			<-taskDone
		})

		<-taskStarted
		assert.Equal(t, 1, wg.Count())

		_ = wg.Go(func(_ context.Context) {
			<-taskDone
		})

		assert.Equal(t, 2, wg.Count())

		close(taskDone)
		time.Sleep(19 * time.Millisecond) // Allow tasks to complete
		assert.Equal(t, 0, wg.Count())
	})

	t.Run("ReportsZeroForNilReceiver", func(t *testing.T) {
		var wg *Group
		assert.Equal(t, 0, wg.Count())
	})
}

// BenchmarkWorkgroup measures the performance of workgroup operations
func BenchmarkWorkgroup(b *testing.B) {
	b.Run("TaskCreation", func(b *testing.B) {
		wg := New(context.Background())

		b.ResetTimer()
		for range b.N {
			err := wg.Go(func(_ context.Context) {})
			if err != nil {
				b.Fatalf("Go returned error: %v", err)
			}
		}
		err := wg.Wait()
		assert.Nil(b, err)
	})

	b.Run("CancelAndWait", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			wg := New(context.Background())

			// Add a task that waits for cancellation
			err := wg.Go(func(ctx context.Context) {
				<-ctx.Done()
			})
			if err != nil {
				b.Fatalf("Go returned error: %v", err)
			}

			wg.Cancel(nil)
			err = wg.Wait()
			assert.Nil(b, err)
		}
	})
}
