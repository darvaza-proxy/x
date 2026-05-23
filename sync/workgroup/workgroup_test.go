package workgroup_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/errors"
	"darvaza.org/x/sync/internal/synctesting"
	"darvaza.org/x/sync/workgroup"
)

// TestNew verifies Group initialisation.
func TestNew(t *testing.T) {
	t.Run("WithContext", runTestNewWithContext)
	t.Run("WithNilContext", runTestNewWithNilContext)
}

func runTestNewWithContext(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	wg := workgroup.New(ctx)

	core.AssertNotNil(t, wg, "group")
	core.AssertEqual(t, ctx, wg.Parent, "parent")
	core.AssertNotNil(t, wg.Context(), "context")
	core.AssertFalse(t, wg.IsCancelled(), "cancelled")
}

func runTestNewWithNilContext(t *testing.T) {
	t.Helper()
	var nilCtx context.Context
	wg := workgroup.New(nilCtx)

	core.AssertNotNil(t, wg, "group")
	core.AssertEqual(t, context.Background(), wg.Parent, "parent")
	core.AssertNotNil(t, wg.Context(), "context")
	core.AssertFalse(t, wg.IsCancelled(), "cancelled")
}

// TestGroup_Context tests the Context method.
func TestGroup_Context(t *testing.T) {
	t.Run("ValidContext", runTestContextValid)
	t.Run("PanicsOnNilReceiver", runTestContextPanicsOnNil)
	t.Run("CancelledWhenGroupCancelled", runTestContextCancelledByGroup)
	t.Run("CancelledWhenParentCancelled", runTestContextCancelledByParent)
}

func runTestContextValid(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	ctx := wg.Context()
	core.AssertNotNil(t, ctx, "context")
	core.AssertNoError(t, ctx.Err(), "ctx err")
}

func runTestContextPanicsOnNil(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { wg.Context() }, errors.ErrNilReceiver, "panic")
}

func runTestContextCancelledByGroup(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	ctx := wg.Context()

	wg.Cancel(nil)
	core.AssertErrorIs(t, ctx.Err(), context.Canceled, "ctx err")
}

func runTestContextCancelledByParent(t *testing.T) {
	t.Helper()
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := workgroup.New(parentCtx)
	ctx := wg.Context()

	cancel()
	synctesting.AssertEventually(t,
		func() bool { return errors.Is(ctx.Err(), context.Canceled) },
		100*time.Millisecond, "ctx cancelled")
}

// TestGroup_Err tests the Err method.
func TestGroup_Err(t *testing.T) {
	t.Run("NilWhenActive", runTestErrNilWhenActive)
	t.Run("ReturnsErrOnNilReceiver", runTestErrNilReceiver)
	t.Run("ReturnsCustomCancelError", runTestErrCustomCancelError)
	t.Run("ReturnsCanceledWhenNoCustomError", runTestErrCanceledDefault)
}

func runTestErrNilWhenActive(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.Err(), "err")
}

func runTestErrNilReceiver(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertErrorIs(t, wg.Err(), errors.ErrNilReceiver, "err")
}

func runTestErrCustomCancelError(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	customErr := errors.New("custom cancel error")
	wg.Cancel(customErr)
	core.AssertErrorIs(t, wg.Err(), customErr, "err")
}

func runTestErrCanceledDefault(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)
	core.AssertErrorIs(t, wg.Err(), context.Canceled, "err")
}

// TestGroup_IsCancelled tests the IsCancelled method.
func TestGroup_IsCancelled(t *testing.T) {
	t.Run("FalseWhenActive", runTestIsCancelledFalseWhenActive)
	t.Run("PanicsOnNilReceiver", runTestIsCancelledPanicsOnNil)
	t.Run("TrueAfterCancel", runTestIsCancelledTrueAfterCancel)
	t.Run("TrueAfterParentContextCancelled", runTestIsCancelledTrueAfterParent)
}

func runTestIsCancelledFalseWhenActive(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertFalse(t, wg.IsCancelled(), "cancelled")
}

func runTestIsCancelledPanicsOnNil(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { wg.IsCancelled() },
		errors.ErrNilReceiver, "panic")
}

func runTestIsCancelledTrueAfterCancel(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)
	core.AssertTrue(t, wg.IsCancelled(), "cancelled")
}

func runTestIsCancelledTrueAfterParent(t *testing.T) {
	t.Helper()
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := workgroup.New(parentCtx)
	cancel()
	synctesting.AssertEventually(t, wg.IsCancelled, 100*time.Millisecond, "cancelled")
}

// TestGroup_Cancelled tests the Cancelled method.
func TestGroup_Cancelled(t *testing.T) {
	t.Run("ReturnsChannel", runTestCancelledReturnsChannel)
	t.Run("PanicsOnNilReceiver", runTestCancelledPanicsOnNil)
	t.Run("ChannelClosedAfterCancel", runTestCancelledChannelClosedAfterCancel)
}

func runTestCancelledReturnsChannel(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	ch := wg.Cancelled()
	core.AssertNotNil(t, ch, "channel")
	synctesting.AssertOpen(t, ch, 10*time.Millisecond, "initial open")
}

func runTestCancelledPanicsOnNil(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { wg.Cancelled() },
		errors.ErrNilReceiver, "panic")
}

func runTestCancelledChannelClosedAfterCancel(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	ch := wg.Cancelled()

	wg.Cancel(nil)
	synctesting.AssertClosed(t, ch, 100*time.Millisecond, "closed after cancel")
}

// TestGroup_Done tests the Done method.
func TestGroup_Done(t *testing.T) {
	t.Run("ChannelCreated", runTestDoneChannelCreated)
	t.Run("PanicsOnNilReceiver", runTestDonePanicsOnNil)
	t.Run("ChannelClosedWhenNoTasks", runTestDoneChannelClosedWhenNoTasks)
	t.Run("ChannelClosedAfterTasksComplete", runTestDoneChannelClosedAfterTasks)
	t.Run("MultipleCallsReturnSameChannel", runTestDoneMultipleCallsSameChannel)
	t.Run("NewChannelAfterCompletion", runTestDoneNewChannelAfterCompletion)
}

func runTestDoneChannelCreated(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNotNil(t, wg.Done(), "channel")
}

func runTestDonePanicsOnNil(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { wg.Done() },
		errors.ErrNilReceiver, "panic")
}

func runTestDoneChannelClosedWhenNoTasks(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	synctesting.AssertClosed(t, wg.Done(), 100*time.Millisecond, "no tasks")
}

func runTestDoneChannelClosedAfterTasks(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	_ = wg.Go(func(_ context.Context) {
		time.Sleep(50 * time.Millisecond)
	})

	ch := wg.Done()
	synctesting.AssertMustOpen(t, ch, 10*time.Millisecond, "open while task runs")
	synctesting.AssertClosed(t, ch, 200*time.Millisecond, "closed after task")
}

func runTestDoneMultipleCallsSameChannel(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	// Pin the watcher channel by keeping a task running so Done()
	// returns the same channel on both calls.
	taskDone := make(chan struct{})
	_ = wg.Go(func(_ context.Context) { <-taskDone })
	defer close(taskDone)

	ch1 := wg.Done()
	ch2 := wg.Done()
	core.AssertSame(t, ch1, ch2, "channel")
}

func runTestDoneNewChannelAfterCompletion(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	ch1 := wg.Done()
	<-ch1

	_ = wg.Go(func(_ context.Context) {
		time.Sleep(10 * time.Millisecond)
	})

	ch2 := wg.Done()
	core.AssertNotSame(t, ch1, ch2, "channel")
}

// TestGroup_Wait tests the Wait method.
func TestGroup_Wait(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", runTestWaitNilReceiver)
	t.Run("ReturnsNilForActiveGroup", runTestWaitNilForActive)
	t.Run("WaitsForAllTasks", runTestWaitForAllTasks)
	t.Run("ReturnsNilOnContextCanceled", runTestWaitNilOnContextCanceled)
	t.Run("ReturnsCustomErrorWhenCancelled", runTestWaitCustomErrorWhenCancelled)
}

func runTestWaitNilReceiver(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertErrorIs(t, wg.Wait(), errors.ErrNilReceiver, "err")
}

func runTestWaitNilForActive(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.Wait(), "err")
}

func runTestWaitForAllTasks(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	counter := atomic.Int32{}
	numTasks := int32(5)

	for range numTasks {
		_ = wg.Go(func(_ context.Context) {
			time.Sleep(10 * time.Millisecond)
			counter.Add(1)
		})
	}

	core.AssertNoError(t, wg.Wait(), "err")
	core.AssertEqual(t, numTasks, counter.Load(), "completed")
}

func runTestWaitNilOnContextCanceled(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	_ = wg.Go(func(ctx context.Context) { <-ctx.Done() })

	wg.Cancel(nil)
	core.AssertNoError(t, wg.Wait(), "err")
}

func runTestWaitCustomErrorWhenCancelled(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	customErr := errors.New("custom cancel reason")

	_ = wg.Go(func(ctx context.Context) { <-ctx.Done() })

	wg.Cancel(customErr)
	core.AssertErrorIs(t, wg.Wait(), customErr, "err")
}

// TestGroup_Cancel tests the Cancel method.
func TestGroup_Cancel(t *testing.T) {
	t.Run("PanicsOnNilReceiver", runTestCancelPanicsOnNil)
	t.Run("ReturnsTrueOnFirstCancel", runTestCancelTrueOnFirst)
	t.Run("ReturnsFalseOnSubsequentCancel", runTestCancelFalseOnSubsequent)
	t.Run("CancelsAllTasks", runTestCancelAllTasks)
	t.Run("PropagatesCustomError", runTestCancelPropagatesCustom)
}

func runTestCancelPanicsOnNil(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { wg.Cancel(nil) },
		errors.ErrNilReceiver, "panic")
}

func runTestCancelTrueOnFirst(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertTrue(t, wg.Cancel(nil), "first cancel")
}

func runTestCancelFalseOnSubsequent(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)
	core.AssertFalse(t, wg.Cancel(nil), "second cancel")
}

func runTestCancelAllTasks(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	cancelled := atomic.Int32{}
	numTasks := int32(5)

	for range numTasks {
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
			cancelled.Add(1)
		})
	}

	wg.Cancel(nil)
	_ = wg.Wait()
	core.AssertEqual(t, numTasks, cancelled.Load(), "cancelled")
}

func runTestCancelPropagatesCustom(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	customErr := errors.New("custom error")

	var receivedErr error
	_ = wg.Go(func(ctx context.Context) {
		<-ctx.Done()
		receivedErr = context.Cause(ctx)
	})

	wg.Cancel(customErr)
	core.AssertErrorIs(t, wg.Wait(), customErr, "wait err")
	core.AssertErrorIs(t, receivedErr, customErr, "task err")
}

// TestGroup_Close tests the Close method.
func TestGroup_Close(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", runTestCloseNilReceiver)
	t.Run("CancelsAndWaitsForTasks", runTestCloseCancelsAndWaits)
	t.Run("CanBeCalledMultipleTimes", runTestCloseMultipleTimes)
}

func runTestCloseNilReceiver(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertErrorIs(t, wg.Close(), errors.ErrNilReceiver, "err")
}

func runTestCloseCancelsAndWaits(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	cancelled := atomic.Int32{}
	numTasks := int32(5)

	for range numTasks {
		_ = wg.Go(func(ctx context.Context) {
			<-ctx.Done()
			time.Sleep(10 * time.Millisecond)
			cancelled.Add(1)
		})
	}

	core.AssertNoError(t, wg.Close(), "err")
	core.AssertEqual(t, numTasks, cancelled.Load(), "cancelled")
}

func runTestCloseMultipleTimes(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	_ = wg.Go(func(ctx context.Context) { <-ctx.Done() })

	core.AssertNoError(t, wg.Close(), "first close")
	core.AssertNoError(t, wg.Close(), "second close")
}

// TestGroup_Go tests the Go method.
func TestGroup_Go(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", runTestGoNilReceiver)
	t.Run("DoesNothingWithNilFunc", runTestGoNilFunc)
	t.Run("ExecutesTask", runTestGoExecutesTask)
	t.Run("ExecutesMultipleTasks", runTestGoExecutesMultiple)
	t.Run("TasksReceiveGroupContext", runTestGoTaskReceivesContext)
	t.Run("ReturnsErrorWhenGroupCancelled", runTestGoErrorWhenCancelled)
}

func runTestGoNilReceiver(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	err := wg.Go(func(_ context.Context) {})
	core.AssertErrorIs(t, err, errors.ErrNilReceiver, "err")
}

func runTestGoNilFunc(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.Go(nil), "Go(nil)")

	done := make(chan struct{})
	go func() {
		_ = wg.Wait()
		close(done)
	}()
	synctesting.AssertClosed(t, done, 100*time.Millisecond, "wait immediate")
}

func runTestGoExecutesTask(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	executed := atomic.Bool{}
	core.AssertNoError(t, wg.Go(func(_ context.Context) {
		executed.Store(true)
	}), "Go")
	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertTrue(t, executed.Load(), "executed")
}

func runTestGoExecutesMultiple(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	counter := atomic.Int32{}
	numTasks := int32(10)

	for range numTasks {
		core.AssertNoError(t, wg.Go(func(_ context.Context) {
			counter.Add(1)
		}), "Go")
	}

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertEqual(t, numTasks, counter.Load(), "count")
}

func runTestGoTaskReceivesContext(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	var received context.Context
	core.AssertNoError(t, wg.Go(func(ctx context.Context) {
		received = ctx
	}), "Go")

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertEqual(t, wg.Context(), received, "ctx")
}

func runTestGoErrorWhenCancelled(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)

	err := wg.Go(func(_ context.Context) {})
	core.AssertErrorIs(t, err, errors.ErrClosed, "err")
}

// TestWorkgroup_Integration tests integration scenarios.
func TestWorkgroup_Integration(t *testing.T) {
	t.Run("ReuseAfterCompletion", runTestIntegrationReuse)
	t.Run("ConcurrentTasksAndCancel", runTestIntegrationConcurrentCancel)
	t.Run("CascadingCancellation", runTestIntegrationCascadingCancel)
}

func runTestIntegrationReuse(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	counter1 := atomic.Int32{}
	for range 3 {
		core.AssertNoError(t, wg.Go(func(_ context.Context) {
			counter1.Add(1)
		}), "Go batch 1")
	}

	core.AssertNoError(t, wg.Wait(), "wait 1")
	core.AssertEqual(t, int32(3), counter1.Load(), "batch 1")

	counter2 := atomic.Int32{}
	for range 5 {
		core.AssertNoError(t, wg.Go(func(_ context.Context) {
			counter2.Add(1)
		}), "Go batch 2")
	}

	core.AssertNoError(t, wg.Wait(), "wait 2")
	core.AssertEqual(t, int32(5), counter2.Load(), "batch 2")
}

func runTestIntegrationConcurrentCancel(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	numTasks := int32(100)
	started := atomic.Int32{}
	cancelled := atomic.Int32{}

	for range numTasks {
		core.AssertNoError(t, wg.Go(func(ctx context.Context) {
			started.Add(1)
			<-ctx.Done()
			cancelled.Add(1)
		}), "Go")
	}

	synctesting.AssertMustEventually(t,
		func() bool { return started.Load() == numTasks },
		200*time.Millisecond, "all started")

	wg.Cancel(errors.New("cancel all tasks"))
	core.AssertError(t, wg.Wait(), "wait")
	core.AssertEqual(t, numTasks, cancelled.Load(), "cancelled")
}

func runTestIntegrationCascadingCancel(t *testing.T) {
	t.Helper()
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	parentWg := workgroup.New(rootCtx)
	childWg := workgroup.New(parentWg.Context())
	grandchildWg := workgroup.New(childWg.Context())

	var parentCancelled, childCancelled, grandchildCancelled atomic.Bool

	core.AssertNoError(t, parentWg.Go(func(ctx context.Context) {
		<-ctx.Done()
		parentCancelled.Store(true)
	}), "parent Go")
	core.AssertNoError(t, childWg.Go(func(ctx context.Context) {
		<-ctx.Done()
		childCancelled.Store(true)
	}), "child Go")
	core.AssertNoError(t, grandchildWg.Go(func(ctx context.Context) {
		<-ctx.Done()
		grandchildCancelled.Store(true)
	}), "grandchild Go")

	rootCancel()
	_ = parentWg.Wait()
	_ = childWg.Wait()
	_ = grandchildWg.Wait()

	core.AssertTrue(t, parentCancelled.Load(), "parent")
	core.AssertTrue(t, childCancelled.Load(), "child")
	core.AssertTrue(t, grandchildCancelled.Load(), "grandchild")
}

// TestGroup_LazyInit tests lazyInit through the public surface.
func TestGroup_LazyInit(t *testing.T) {
	t.Run("InitialisesEmptyGroup", runTestLazyInitEmptyGroup)
	t.Run("InitialisesWithCustomParent", runTestLazyInitCustomParent)
}

func runTestLazyInitEmptyGroup(t *testing.T) {
	t.Helper()
	wg := &workgroup.Group{}

	ctx := wg.Context()
	core.AssertNotNil(t, ctx, "context")
	core.AssertEqual(t, context.Background(), wg.Parent, "parent default")
}

func runTestLazyInitCustomParent(t *testing.T) {
	t.Helper()
	ctxKey := core.NewContextKey[string]("key")
	parentCtx := ctxKey.WithValue(context.Background(), "value")
	wg := &workgroup.Group{Parent: parentCtx}

	ctx := wg.Context()
	core.AssertNotNil(t, ctx, "context")
	core.AssertEqual(t, "value", ctx.Value(ctxKey), "propagated value")
}

// TestGroup_Timeout verifies cancellation when the parent context expires.
func TestGroup_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(),
		50*time.Millisecond)
	defer cancel()

	wg := workgroup.New(ctx)
	taskCancelled := atomic.Bool{}
	_ = wg.Go(func(ctx context.Context) {
		<-ctx.Done()
		taskCancelled.Store(true)
	})

	start := time.Now()
	err := wg.Wait()
	elapsed := time.Since(start)

	core.AssertTrue(t, elapsed >= 50*time.Millisecond, "wait blocks until timeout")
	core.AssertTrue(t, elapsed < 200*time.Millisecond, "wait near timeout")
	core.AssertErrorIs(t, err, context.DeadlineExceeded, "err")
	core.AssertTrue(t, taskCancelled.Load(), "task cancelled")
}

// TestGroup_Concurrent tests concurrent behaviour.
func TestGroup_Concurrent(t *testing.T) {
	t.Run("ConcurrentGoAndCancel", runTestConcurrentGoAndCancel)
	t.Run("ConcurrentDoneChannels", runTestConcurrentDoneChannels)
}

// concurrentCounts groups the atomic counters threaded through the
// concurrent Go/Cancel stress test.
type concurrentCounts struct {
	cancelled, completed atomic.Int32
}

func runTestConcurrentGoAndCancel(t *testing.T) {
	t.Helper()
	const numRoutines = 10
	const tasksPerRoutine = 100

	wg := workgroup.New(context.Background())
	counts := &concurrentCounts{}

	var startWg, endWg sync.WaitGroup
	startWg.Add(numRoutines + 1)
	endWg.Add(numRoutines + 1)

	for range numRoutines {
		go concurrentGoAdder(&startWg, &endWg, wg, tasksPerRoutine, counts)
	}
	go concurrentGoCanceller(&startWg, &endWg, wg)

	endWg.Wait()
	_ = wg.Wait()

	core.AssertTrue(t, counts.cancelled.Load() > 0, "some cancelled")
	t.Logf("completed: %d, cancelled: %d",
		counts.completed.Load(), counts.cancelled.Load())
}

func concurrentGoAdder(start, end *sync.WaitGroup, wg *workgroup.Group,
	n int, counts *concurrentCounts) {
	defer end.Done()
	start.Done()
	start.Wait()

	for range n {
		_ = wg.Go(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				counts.cancelled.Add(1)
			case <-time.After(10 * time.Millisecond):
				counts.completed.Add(1)
			}
		})
		time.Sleep(time.Millisecond)
	}
}

func concurrentGoCanceller(start, end *sync.WaitGroup, wg *workgroup.Group) {
	defer end.Done()
	start.Done()
	start.Wait()

	time.Sleep(50 * time.Millisecond)
	wg.Cancel(errors.New("concurrent cancellation"))
}

func runTestConcurrentDoneChannels(t *testing.T) {
	t.Helper()
	const numGoroutines = 10
	wg := workgroup.New(context.Background())

	_ = wg.Go(func(_ context.Context) {
		time.Sleep(100 * time.Millisecond)
	})

	channels := make([]<-chan struct{}, numGoroutines)
	var startWg sync.WaitGroup
	startWg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(idx int) {
			defer startWg.Done()
			channels[idx] = wg.Done()
		}(i)
	}
	startWg.Wait()

	for i := 1; i < numGoroutines; i++ {
		core.AssertSame(t, channels[0], channels[i], "channel %d", i)
	}

	synctesting.AssertMustClosed(t, channels[0], 200*time.Millisecond,
		"shared Done across %d goroutines", numGoroutines)
	core.AssertNoError(t, wg.Wait(), "wait")
}

// nilReceiverErrorCase asserts a *Group method returns ErrNilReceiver
// when invoked on a nil receiver.
type nilReceiverErrorCase struct {
	op   func(*workgroup.Group) error
	name string
}

func (tc nilReceiverErrorCase) Name() string { return tc.name }

func (tc nilReceiverErrorCase) Test(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertErrorIs(t, tc.op(wg), errors.ErrNilReceiver, tc.name)
}

var _ core.TestCase = nilReceiverErrorCase{}

func newNilReceiverErrorCase(name string,
	op func(*workgroup.Group) error) nilReceiverErrorCase {
	return nilReceiverErrorCase{name: name, op: op}
}

// nilReceiverPanicCase asserts a *Group method panics with ErrNilReceiver
// when invoked on a nil receiver.
type nilReceiverPanicCase struct {
	op   func(*workgroup.Group)
	name string
}

func (tc nilReceiverPanicCase) Name() string { return tc.name }

func (tc nilReceiverPanicCase) Test(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	core.AssertPanic(t, func() { tc.op(wg) },
		errors.ErrNilReceiver, tc.name)
}

var _ core.TestCase = nilReceiverPanicCase{}

func newNilReceiverPanicCase(name string,
	op func(*workgroup.Group)) nilReceiverPanicCase {
	return nilReceiverPanicCase{name: name, op: op}
}

// TestGroup_NilReceiver exercises every public method against a nil
// receiver. Methods returning error must surface ErrNilReceiver; methods
// returning a value or channel must panic with ErrNilReceiver.
func TestGroup_NilReceiver(t *testing.T) {
	core.RunTestCases(t, []nilReceiverErrorCase{
		newNilReceiverErrorCase("Wait",
			func(wg *workgroup.Group) error { return wg.Wait() }),
		newNilReceiverErrorCase("Close",
			func(wg *workgroup.Group) error { return wg.Close() }),
		newNilReceiverErrorCase("Err",
			func(wg *workgroup.Group) error { return wg.Err() }),
		newNilReceiverErrorCase("Go",
			func(wg *workgroup.Group) error {
				return wg.Go(func(_ context.Context) {})
			}),
		newNilReceiverErrorCase("GoCatch",
			func(wg *workgroup.Group) error {
				return wg.GoCatch(
					func(_ context.Context) error { return nil },
					func(_ context.Context, e error) error { return e },
				)
			}),
	})

	core.RunTestCases(t, []nilReceiverPanicCase{
		newNilReceiverPanicCase("Context",
			func(wg *workgroup.Group) { wg.Context() }),
		newNilReceiverPanicCase("IsCancelled",
			func(wg *workgroup.Group) { wg.IsCancelled() }),
		newNilReceiverPanicCase("Cancelled",
			func(wg *workgroup.Group) { wg.Cancelled() }),
		newNilReceiverPanicCase("Done",
			func(wg *workgroup.Group) { wg.Done() }),
		newNilReceiverPanicCase("Cancel",
			func(wg *workgroup.Group) { wg.Cancel(nil) }),
	})
}

// TestGroup_ErrorHandling tests error propagation patterns.
func TestGroup_ErrorHandling(t *testing.T) {
	t.Run("CustomErrorTypes", runTestErrorHandlingCustomTypes)
}

func runTestErrorHandlingCustomTypes(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	myErr := errors.New("custom error type")

	var receivedErr error
	core.AssertNoError(t, wg.Go(func(ctx context.Context) {
		<-ctx.Done()
		receivedErr = context.Cause(ctx)
	}), "Go")

	wg.Cancel(myErr)
	_ = wg.Wait()
	core.AssertErrorIs(t, receivedErr, myErr, "task err")
}

// TestWorkgroup_ContextValues verifies context value propagation into tasks.
func TestWorkgroup_ContextValues(t *testing.T) {
	type key string
	testKey := key("test-key")
	testVal := "test-value"

	ctx := context.WithValue(context.Background(), testKey, testVal)
	wg := workgroup.New(ctx)

	var receivedVal any
	_ = wg.Go(func(ctx context.Context) {
		receivedVal = ctx.Value(testKey)
	})

	core.AssertNoError(t, wg.Wait(), "wait")
	if got, ok := core.AssertTypeIs[string](t, receivedVal, "value type"); ok {
		core.AssertEqual(t, testVal, got, "value")
	}
}

// TestGroup_Reuse verifies a Group can be reused after Wait returns.
func TestGroup_Reuse(t *testing.T) {
	wg := workgroup.New(context.Background())

	counter1 := atomic.Int32{}
	_ = wg.Go(func(_ context.Context) { counter1.Add(1) })
	core.AssertNoError(t, wg.Wait(), "wait 1")
	core.AssertEqual(t, int32(1), counter1.Load(), "batch 1")

	counter2 := atomic.Int32{}
	_ = wg.Go(func(_ context.Context) { counter2.Add(1) })
	core.AssertNoError(t, wg.Wait(), "wait 2")
	core.AssertEqual(t, int32(1), counter2.Load(), "batch 2")
}

// TestGroup_DeadlineExceeded verifies behaviour under context deadline.
func TestGroup_DeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(),
		10*time.Millisecond)
	defer cancel()

	wg := workgroup.New(ctx)
	interrupted := atomic.Bool{}
	_ = wg.Go(func(ctx context.Context) {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			interrupted.Store(true)
		}
	})

	err := wg.Wait()
	core.AssertErrorIs(t, err, context.DeadlineExceeded, "err")
	core.AssertTrue(t, interrupted.Load(), "interrupted")
}

// TestGroup_GoCatch tests the GoCatch method.
func TestGroup_GoCatch(t *testing.T) {
	t.Run("ReturnsErrorOnNilReceiver", runTestGoCatchNilReceiver)
	t.Run("DoesNothingWithNilFunc", runTestGoCatchNilFunc)
	t.Run("ExecutesSuccessfulTask", runTestGoCatchSuccessfulTask)
	t.Run("TasksReceiveGroupContext", runTestGoCatchTaskReceivesContext)
	t.Run("CancelsGroupOnError", runTestGoCatchCancelsOnError)
	t.Run("CatchHandlerProcessesError", runTestGoCatchHandlerProcesses)
	t.Run("CatchHandlerCanPreventCancellation", runTestGoCatchHandlerPrevents)
	t.Run("RecoversPanics", runTestGoCatchRecoversPanics)
	t.Run("ReturnsErrorWhenGroupCancelled", runTestGoCatchErrorWhenCancelled)
	t.Run("CatchHandlerReceivesContext", runTestGoCatchHandlerReceivesContext)
}

func runTestGoCatchNilReceiver(t *testing.T) {
	t.Helper()
	var wg *workgroup.Group
	err := wg.GoCatch(
		func(_ context.Context) error { return nil },
		func(_ context.Context, err error) error { return err },
	)
	core.AssertErrorIs(t, err, errors.ErrNilReceiver, "err")
}

func runTestGoCatchNilFunc(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.GoCatch(nil, nil), "GoCatch(nil)")

	done := make(chan struct{})
	go func() {
		_ = wg.Wait()
		close(done)
	}()
	synctesting.AssertClosed(t, done, 100*time.Millisecond, "wait immediate")
}

func runTestGoCatchSuccessfulTask(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	executed := atomic.Bool{}

	core.AssertNoError(t, wg.GoCatch(
		func(_ context.Context) error {
			executed.Store(true)
			return nil
		}, nil), "GoCatch")
	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertTrue(t, executed.Load(), "executed")
}

func runTestGoCatchTaskReceivesContext(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	var received context.Context
	core.AssertNoError(t, wg.GoCatch(
		func(ctx context.Context) error {
			received = ctx
			return nil
		}, nil), "GoCatch")

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertEqual(t, wg.Context(), received, "ctx")
}

func runTestGoCatchCancelsOnError(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	testErr := errors.New("test error")

	core.AssertNoError(t, wg.GoCatch(
		func(_ context.Context) error { return testErr },
		nil), "GoCatch")

	core.AssertErrorIs(t, wg.Wait(), testErr, "wait err")
}

func runTestGoCatchHandlerProcesses(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	testErr := errors.New("original error")
	processedErr := errors.New("processed error")

	core.AssertNoError(t, wg.GoCatch(
		func(_ context.Context) error { return testErr },
		func(_ context.Context, err error) error {
			core.AssertErrorIs(t, err, testErr, "received")
			return processedErr
		}), "GoCatch")

	err := wg.Wait()
	core.AssertErrorIs(t, err, processedErr, "wait err")
	core.AssertFalse(t, errors.Is(err, testErr), "no testErr")
}

func runTestGoCatchHandlerPrevents(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	testErr := errors.New("non-critical error")

	core.AssertNoError(t, wg.GoCatch(
		func(_ context.Context) error { return testErr },
		func(_ context.Context, _ error) error { return nil },
	), "GoCatch")

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertFalse(t, wg.IsCancelled(), "not cancelled")
}

func runTestGoCatchRecoversPanics(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	panicMsg := "deliberate panic"

	core.AssertNoError(t, wg.GoCatch(
		func(_ context.Context) error { panic(panicMsg) },
		func(_ context.Context, err error) error { return err },
	), "GoCatch")

	err := wg.Wait()
	core.AssertError(t, err, "wait err")
	panicErr, ok := core.AssertTypeIs[*core.PanicError](t, err, "panic type")
	if ok {
		core.AssertContains(t, panicErr.Error(), panicMsg, "panic msg")
	}
}

func runTestGoCatchErrorWhenCancelled(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)

	err := wg.GoCatch(func(_ context.Context) error { return nil }, nil)
	core.AssertErrorIs(t, err, errors.ErrClosed, "err")
}

func runTestGoCatchHandlerReceivesContext(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())

	var taskCtx, catchCtx context.Context
	testErr := errors.New("test error")

	core.AssertNoError(t, wg.GoCatch(
		func(ctx context.Context) error {
			taskCtx = ctx
			return testErr
		},
		func(ctx context.Context, err error) error {
			catchCtx = ctx
			return err
		},
	), "GoCatch")

	_ = wg.Wait()
	core.AssertEqual(t, wg.Context(), taskCtx, "task ctx")
	core.AssertEqual(t, wg.Context(), catchCtx, "catch ctx")
}

// BenchmarkWorkgroup measures workgroup operation performance.
func BenchmarkWorkgroup(b *testing.B) {
	b.Run("TaskCreation", runBenchmarkTaskCreation)
	b.Run("CancelAndWait", runBenchmarkCancelAndWait)
}

func runBenchmarkTaskCreation(b *testing.B) {
	wg := workgroup.New(context.Background())

	b.ResetTimer()
	for range b.N {
		if err := wg.Go(func(_ context.Context) {}); err != nil {
			b.Fatalf("Go returned error: %v", err)
		}
	}
	core.AssertNoError(b, wg.Wait(), "wait")
}

func runBenchmarkCancelAndWait(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		wg := workgroup.New(context.Background())
		if err := wg.Go(func(ctx context.Context) { <-ctx.Done() }); err != nil {
			b.Fatalf("Go returned error: %v", err)
		}
		wg.Cancel(nil)
		core.AssertNoError(b, wg.Wait(), "wait")
	}
}
