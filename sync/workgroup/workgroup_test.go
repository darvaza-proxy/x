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

// TestGroup_OnCancel tests the OnCancel handler.
func TestGroup_OnCancel(t *testing.T) {
	t.Run("HandlerFiresOnCancel", runTestOnCancelHandlerFires)
	t.Run("HandlerReceivesLiveContext", runTestOnCancelReceivesLiveContext)
	t.Run("HandlerContextRetainsValues", runTestOnCancelContextRetainsValues)
	t.Run("HandlerReceivesCustomCause", runTestOnCancelReceivesCause)
	t.Run("HandlerReceivesContextCanceledForNilCause",
		runTestOnCancelReceivesContextCanceledForNilCause)
	t.Run("HandlerNotCalledWithoutCancel", runTestOnCancelNotCalledWithoutCancel)
	t.Run("HandlerCalledOnceOnRepeatedCancel", runTestOnCancelOnceOnRepeatedCancel)
	t.Run("HandlerPanicContained", runTestOnCancelHandlerPanicContained)
	t.Run("HandlerFiresOnClose", runTestOnCancelFiresOnClose)
	t.Run("HandlerFiresOnParentCancel", runTestOnCancelFiresOnParentCancel)
	t.Run("HandlerReceivesParentCancelCause", runTestOnCancelReceivesParentCancelCause)
	t.Run("HandlerFiresOnceAcrossParentAndCancel",
		runTestOnCancelOnceAcrossParentAndCancel)
}

func runTestOnCancelHandlerFires(t *testing.T) {
	t.Helper()
	handlerRan := make(chan struct{})
	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerRan)
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, handlerRan, 100*time.Millisecond,
		"handler fired")
	core.AssertNoError(t, wg.Wait(), "wait")
}

// runTestOnCancelReceivesLiveContext pins that the handler's context is
// detached from the Group's cancellation: it is live (ctx.Err() is nil)
// rather than the already-cancelled group context.
func runTestOnCancelReceivesLiveContext(t *testing.T) {
	t.Helper()
	var observedErr error
	receivedCh := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(ctx context.Context, _ error) {
		observedErr = ctx.Err()
		close(receivedCh)
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran")
	core.AssertNoError(t, observedErr, "ctx.Err() inside handler")
	core.AssertNoError(t, wg.Wait(), "wait")
}

// runTestOnCancelContextRetainsValues pins that detaching the handler's
// context preserves the Group context's values, so cleanup work keeps any
// logger or trace identifiers carried through it.
func runTestOnCancelContextRetainsValues(t *testing.T) {
	t.Helper()
	type key string
	testKey := key("trace-id")
	testVal := "abc123"

	var received any
	receivedCh := make(chan struct{})

	parent := context.WithValue(context.Background(), testKey, testVal)
	wg := workgroup.New(parent)
	wg.OnCancel = func(ctx context.Context, _ error) {
		received = ctx.Value(testKey)
		close(receivedCh)
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran")
	if got, ok := core.AssertTypeIs[string](t, received, "value type"); ok {
		core.AssertEqual(t, testVal, got, "value carried into handler")
	}
	core.AssertNoError(t, wg.Wait(), "wait")
}

func runTestOnCancelReceivesCause(t *testing.T) {
	t.Helper()
	customErr := errors.New("custom cancel cause")
	var received error
	receivedCh := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, cause error) {
		received = cause
		close(receivedCh)
	}

	wg.Cancel(customErr)
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran")
	core.AssertErrorIs(t, received, customErr, "cause")
	core.AssertErrorIs(t, wg.Wait(), customErr, "wait err")
}

func runTestOnCancelReceivesContextCanceledForNilCause(t *testing.T) {
	t.Helper()
	var received error
	receivedCh := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, cause error) {
		received = cause
		close(receivedCh)
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran")
	core.AssertErrorIs(t, received, context.Canceled, "cause")
	core.AssertNoError(t, wg.Wait(), "wait")
}

func runTestOnCancelNotCalledWithoutCancel(t *testing.T) {
	t.Helper()
	handlerRan := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerRan)
	}

	core.AssertNoError(t, wg.Wait(), "wait")
	synctesting.AssertOpen(t, handlerRan, 50*time.Millisecond,
		"handler not called")
}

func runTestOnCancelOnceOnRepeatedCancel(t *testing.T) {
	t.Helper()
	var count atomic.Int32

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		count.Add(1)
	}

	core.AssertTrue(t, wg.Cancel(nil), "first cancel")
	core.AssertFalse(t, wg.Cancel(nil), "second cancel")
	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertEqual(t, int32(1), count.Load(), "handler call count")
}

// runTestOnCancelHandlerPanicContained pins that a panicking OnCancel handler
// is recovered. The handler runs as a detached, counted task; without
// containment its panic would crash the whole test binary and leave Wait
// blocked forever on the tasks counter. The defer'd tasks.Dec still runs, so
// Wait must return.
func runTestOnCancelHandlerPanicContained(t *testing.T) {
	t.Helper()
	handlerRan := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerRan)
		panic("boom in OnCancel handler")
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, handlerRan, 100*time.Millisecond,
		"handler ran")
	core.AssertNoError(t, wg.Wait(), "wait returns after contained panic")
}

func runTestOnCancelFiresOnClose(t *testing.T) {
	t.Helper()
	handlerRan := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerRan)
	}

	core.AssertNoError(t, wg.Close(), "close")
	synctesting.AssertMustClosed(t, handlerRan, 100*time.Millisecond,
		"handler fired before Close returned")
}

// runTestOnCancelFiresOnParentCancel pins that cancelling the parent
// context — with no explicit Cancel or Close — still drives the handler.
func runTestOnCancelFiresOnParentCancel(t *testing.T) {
	t.Helper()
	var received error
	receivedCh := make(chan struct{})

	parentCtx, cancel := context.WithCancel(context.Background())
	wg := workgroup.New(parentCtx)
	wg.OnCancel = func(_ context.Context, cause error) {
		received = cause
		close(receivedCh)
	}

	cancel()
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran on parent cancel")
	core.AssertErrorIs(t, received, context.Canceled, "cause")
	core.AssertNoError(t, wg.Wait(), "wait")
}

// runTestOnCancelReceivesParentCancelCause pins that the handler receives
// the parent's cancellation cause, not a flattened context.Canceled.
func runTestOnCancelReceivesParentCancelCause(t *testing.T) {
	t.Helper()
	customErr := errors.New("parent cancel cause")
	var received error
	receivedCh := make(chan struct{})

	parentCtx, cancel := context.WithCancelCause(context.Background())
	wg := workgroup.New(parentCtx)
	wg.OnCancel = func(_ context.Context, cause error) {
		received = cause
		close(receivedCh)
	}

	cancel(customErr)
	synctesting.AssertMustClosed(t, receivedCh, 100*time.Millisecond,
		"handler ran on parent cancel")
	core.AssertErrorIs(t, received, customErr, "cause")
	core.AssertErrorIs(t, wg.Wait(), customErr, "wait err")
}

// runTestOnCancelOnceAcrossParentAndCancel pins that a parent cancellation
// racing an explicit Cancel still fires the handler exactly once: whichever
// path wins the transition, the loser is deduped by the cancelled flag.
func runTestOnCancelOnceAcrossParentAndCancel(t *testing.T) {
	t.Helper()
	var count atomic.Int32

	parentCtx, cancel := context.WithCancel(context.Background())
	wg := workgroup.New(parentCtx)
	wg.OnCancel = func(_ context.Context, _ error) {
		count.Add(1)
	}

	go cancel()
	wg.Cancel(nil)

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertEqual(t, int32(1), count.Load(), "handler call count")
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

const (
	concurrentCancelIterations = 200
	concurrentCancellers       = 8
)

// TestGroup_ConcurrentCancelOnceOnly fires several Cancel calls at the same
// Group simultaneously and asserts exactly one wins. doCancel guards the
// cancelled flag under mu: a caller that loses the race observes the flag
// already set once it acquires the lock and returns false without spawning
// a second OnCancel handler or overwriting the cause. Each iteration
// releases all callers from a shared barrier to maximise the overlap.
func TestGroup_ConcurrentCancelOnceOnly(t *testing.T) {
	for i := range concurrentCancelIterations {
		runOneConcurrentCancelIteration(t, i)
	}
}

func runOneConcurrentCancelIteration(t *testing.T, i int) {
	t.Helper()
	wg := workgroup.New(context.Background())

	var trueCount atomic.Int32
	var start, end sync.WaitGroup
	start.Add(concurrentCancellers)
	end.Add(concurrentCancellers)

	for range concurrentCancellers {
		go func() {
			defer end.Done()
			start.Done()
			start.Wait()
			if wg.Cancel(context.Canceled) {
				trueCount.Add(1)
			}
		}()
	}
	end.Wait()

	core.AssertEqual(t, int32(1), trueCount.Load(),
		"exactly one Cancel wins iter %d", i)
	core.AssertNoError(t, wg.Close(), "close iter %d", i)
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

const (
	doneGenCycles   = 20
	doneGenFetchers = 6
)

// TestGroup_DoneGenerations pins the full Done() generation lifecycle on a
// reused Group. Each cycle enrols a gated task so the watcher generation
// stays observable, then asserts the four invariants of doDone: a fresh
// generation hands back a new channel (the prior watcher reset wg.doneCh to
// nil before closing), the channel is open while the task runs, concurrent
// Done() callers all share that one channel, and it closes once the task
// drains. Carrying the previous cycle's channel forward is what catches a
// watcher that closes but fails to reset — it would hand back the cached,
// already-closed channel, tripping AssertNotSame. AssertMustClosed on the
// generation channel is the synchronisation point: close is the watcher's
// outermost defer, so it cannot fire until the reset defer already has,
// which makes the next cycle's freshness assertion deterministic.
func TestGroup_DoneGenerations(t *testing.T) {
	wg := workgroup.New(context.Background())

	var prev <-chan struct{}
	for i := range doneGenCycles {
		prev = runOneDoneGeneration(t, wg, i, prev)
	}

	core.AssertNoError(t, wg.Close(), "close")
}

func runOneDoneGeneration(t *testing.T, wg *workgroup.Group, i int,
	prev <-chan struct{}) <-chan struct{} {
	t.Helper()

	release := make(chan struct{})
	core.AssertNoError(t, wg.Go(func(_ context.Context) {
		<-release
	}), "go iter %d", i)

	ch := wg.Done()
	if prev != nil {
		core.AssertNotSame(t, prev, ch, "fresh generation iter %d", i)
	}
	synctesting.AssertMustOpen(t, ch, 10*time.Millisecond,
		"open while task runs iter %d", i)

	for k, got := range concurrentFetchDone(wg, doneGenFetchers) {
		core.AssertSame(t, ch, got, "shared channel iter %d.%d", i, k)
	}

	close(release)
	synctesting.AssertMustClosed(t, ch, 100*time.Millisecond,
		"closed on drain iter %d", i)
	core.AssertNoError(t, wg.Wait(), "wait iter %d", i)
	return ch
}

// concurrentFetchDone fetches wg.Done() from n goroutines released together,
// returning the channels they observed so the caller can assert the
// generation is shared. The start barrier maximises overlap on the read path.
func concurrentFetchDone(wg *workgroup.Group, n int) []<-chan struct{} {
	got := make([]<-chan struct{}, n)
	var start, end sync.WaitGroup
	start.Add(n)
	end.Add(n)

	for k := range n {
		go func(k int) {
			defer end.Done()
			start.Done()
			start.Wait()
			got[k] = wg.Done()
		}(k)
	}

	end.Wait()
	return got
}

const (
	doneReuseCycles   = 300
	doneReuseFetchers = 6
)

// TestGroup_DoneReuseRaceStress overlaps the Done() drain/close/reset/recreate
// boundary under concurrency, the one window TestGroup_DoneGenerations
// serialises away with its gate. Each cycle enrols a task, takes a census
// fetch of the current generation, then races several fetchers against the
// task's drain so the watcher's reset of wg.doneCh runs alongside their fresh
// Done() calls. Run under -race this guards the lock around wg.doneCh and
// catches a double-close (the watcher would panic). Because the boundary is
// racy, the only assertions it can make are liveness — every Wait returns —
// and a census: across many cycles the watcher must recreate the channel, so
// more than one distinct instance is observed. A watcher that never reset
// would yield exactly one, failing the census.
func TestGroup_DoneReuseRaceStress(t *testing.T) {
	wg := workgroup.New(context.Background())

	seen := make(map[<-chan struct{}]struct{})
	for i := range doneReuseCycles {
		seen[runOneDoneReuseCycle(t, wg, i)] = struct{}{}
	}

	core.AssertTrue(t, len(seen) > 1,
		"watcher recreated the channel across %d cycles", doneReuseCycles)
	core.AssertNoError(t, wg.Close(), "close")
}

func runOneDoneReuseCycle(t *testing.T, wg *workgroup.Group, i int) <-chan struct{} {
	t.Helper()

	release := make(chan struct{})
	core.AssertNoError(t, wg.Go(func(_ context.Context) {
		<-release
	}), "go iter %d", i)

	// Census fetch while the task is still blocked: the generation this cycle
	// opened, distinct each time the watcher reset and recreated.
	census := wg.Done()

	var end sync.WaitGroup
	end.Add(doneReuseFetchers)
	for range doneReuseFetchers {
		go func() {
			defer end.Done()
			<-wg.Done()
		}()
	}

	// Release after the fetchers are in flight so the watcher's
	// drain/close/reset overlaps their concurrent Done() calls.
	close(release)
	end.Wait()
	core.AssertNoError(t, wg.Wait(), "wait iter %d", i)
	return census
}

// TestGroup_DrainRaceRegression stresses the two historical
// Add-after-drain panic sites. When the inner task counter has just
// dropped to zero and a Wait is in flight, a concurrent Cancel with
// OnCancel set (Site 1) or a concurrent Go (Site 2) used to trip
// sync.WaitGroup's "Add(positive) on zero counter with Wait in
// flight → panic" rule. Each subtest races the two operations across
// the drain boundary many times; the bug would surface as a panic on
// at least one iteration. The OnCancel and task flags also verify the
// drain-then-fence contract by ensuring no scheduled work is silently
// dropped.
func TestGroup_DrainRaceRegression(t *testing.T) {
	t.Run("CancelRacingWait", runTestDrainRaceCancelVsWait)
	t.Run("GoRacingWait", runTestDrainRaceGoVsWait)
}

const drainRaceIterations = 200

func runTestDrainRaceCancelVsWait(t *testing.T) {
	t.Helper()
	for i := range drainRaceIterations {
		var handlerRan atomic.Bool
		wg := workgroup.New(context.Background())
		wg.OnCancel = func(_ context.Context, _ error) {
			handlerRan.Store(true)
		}

		_ = wg.Go(func(_ context.Context) {})

		var start sync.WaitGroup
		start.Add(2)
		done := make(chan struct{}, 2)
		go func() {
			start.Done()
			start.Wait()
			_ = wg.Wait()
			done <- struct{}{}
		}()
		go func() {
			start.Done()
			start.Wait()
			wg.Cancel(nil)
			done <- struct{}{}
		}()
		<-done
		<-done
		core.AssertNoError(t, wg.Close(), "close iter %d", i)
		core.AssertTrue(t, handlerRan.Load(),
			"OnCancel must have run by Close return iter %d", i)
	}
}

func runTestDrainRaceGoVsWait(t *testing.T) {
	t.Helper()
	for i := range drainRaceIterations {
		runOneGoVsWaitIteration(t, i)
	}
}

// runOneGoVsWaitIteration exercises one drain-race between Wait and a
// concurrent second Go. D2 contract: if Go returned nil, the task is
// enrolled and must complete before Close returns; if Go returned
// ErrClosed, no goroutine was spawned. No in-between state.
func runOneGoVsWaitIteration(t *testing.T, i int) {
	t.Helper()
	var secondTaskRan atomic.Bool
	var goErr atomic.Pointer[error]

	wg := workgroup.New(context.Background())
	_ = wg.Go(func(_ context.Context) {})

	var start sync.WaitGroup
	start.Add(2)
	done := make(chan struct{}, 2)
	go func() {
		start.Done()
		start.Wait()
		_ = wg.Wait()
		done <- struct{}{}
	}()
	go func() {
		start.Done()
		start.Wait()
		err := wg.Go(func(_ context.Context) {
			secondTaskRan.Store(true)
		})
		goErr.Store(&err)
		done <- struct{}{}
	}()
	<-done
	<-done

	core.AssertNoError(t, wg.Close(), "close iter %d", i)

	goPtr := goErr.Load()
	core.AssertMustNotNil(t, goPtr, "Go result recorded iter %d", i)
	if *goPtr == nil {
		// Go enrolled the task; Close drained it, so it must have run.
		core.AssertTrue(t, secondTaskRan.Load(),
			"task must have run when Go returned nil iter %d", i)
	} else {
		// Go refused: no goroutine was spawned, so the task never ran.
		core.AssertErrorIs(t, *goPtr, errors.ErrClosed,
			"Go error is ErrClosed iter %d", i)
		core.AssertFalse(t, secondTaskRan.Load(),
			"task must not have run when Go returned ErrClosed iter %d", i)
	}
}

// TestGroup_FenceSemantics pins the drain-then-fence contract
// deterministically (no race iteration loops). These tests fail if the
// fence between Wait/Done watcher and doCancel/doGo is removed.
func TestGroup_FenceSemantics(t *testing.T) {
	t.Run("CancelBeforeWait_WaitWaitsForHandler",
		runTestFenceCancelBeforeWaitWaitsForHandler)
	t.Run("CancelBeforeWait_DoneWaitsForHandler",
		runTestFenceCancelBeforeWaitDoneWaitsForHandler)
	t.Run("GoAfterCancelReturnsErrClosed",
		runTestFenceGoAfterCancelReturnsErrClosed)
}

func runTestFenceCancelBeforeWaitWaitsForHandler(t *testing.T) {
	t.Helper()
	release := make(chan struct{})
	handlerEntered := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerEntered)
		<-release
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, handlerEntered, 100*time.Millisecond,
		"handler entered")

	waitDone := make(chan struct{})
	go func() {
		_ = wg.Wait()
		close(waitDone)
	}()

	// Wait must be blocked while the handler is held in `release`.
	synctesting.AssertOpen(t, waitDone, 50*time.Millisecond,
		"wait blocked on handler")

	close(release)
	synctesting.AssertMustClosed(t, waitDone, 100*time.Millisecond,
		"wait completed after handler")
}

func runTestFenceCancelBeforeWaitDoneWaitsForHandler(t *testing.T) {
	t.Helper()
	release := make(chan struct{})
	handlerEntered := make(chan struct{})

	wg := workgroup.New(context.Background())
	wg.OnCancel = func(_ context.Context, _ error) {
		close(handlerEntered)
		<-release
	}

	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, handlerEntered, 100*time.Millisecond,
		"handler entered")

	doneCh := wg.Done()

	// Done channel must remain open while the handler is held.
	synctesting.AssertOpen(t, doneCh, 50*time.Millisecond,
		"done channel blocked on handler")

	close(release)
	synctesting.AssertMustClosed(t, doneCh, 100*time.Millisecond,
		"done channel closed after handler")
}

func runTestFenceGoAfterCancelReturnsErrClosed(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)
	core.AssertNoError(t, wg.Wait(), "wait")

	err := wg.Go(func(_ context.Context) {})
	core.AssertErrorIs(t, err, errors.ErrClosed, "go after cancel")
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
		newNilReceiverErrorCase("GoShutdown",
			func(wg *workgroup.Group) error {
				return wg.GoShutdown(
					func(_ context.Context) error { return nil }, nil, 0)
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

// TestGroup_GoShutdown tests the GoShutdown method: an fn paired with a
// shutdown handler that signals it to end when the Group is cancelled
// while fn is still running.
func TestGroup_GoShutdown(t *testing.T) {
	t.Run("ServerPattern", runTestGoShutdownServerPattern)
	t.Run("CloseSignalsShutdown", runTestGoShutdownCloseSignals)
	t.Run("CleanReturnSkipsShutdown", runTestGoShutdownCleanReturnSkipsShutdown)
	t.Run("FnErrorSkipsShutdown", runTestGoShutdownFnErrorSkipsShutdown)
	t.Run("FnPanicSkipsShutdown", runTestGoShutdownFnPanicSkipsShutdown)
	t.Run("FnErrorSignalsPeer", runTestGoShutdownFnErrorSignalsPeer)
	t.Run("ShutdownContextBounded", runTestGoShutdownContextBounded)
	t.Run("ShutdownContextUnbounded", runTestGoShutdownContextUnbounded)
	t.Run("ShutdownPanicContained", runTestGoShutdownPanicContained)
	t.Run("WatcherMode", runTestGoShutdownWatcherMode)
	t.Run("NilShutdownRunsFn", runTestGoShutdownNilShutdownSuccess)
	t.Run("NilShutdownSupervisesFn", runTestGoShutdownNilShutdownError)
	t.Run("BothNilNoop", runTestGoShutdownBothNil)
	t.Run("EnrolmentWhenCancelled", runTestGoShutdownEnrolmentWhenCancelled)
}

// runTestGoShutdownServerPattern models the http.Server pattern: fn ignores
// the context and runs until the shutdown handler stops it. On cancellation
// the handler closes the server, fn returns, and the paired task releases
// the Group.
func runTestGoShutdownServerPattern(t *testing.T) {
	t.Helper()
	serverStop := make(chan struct{})
	var served atomic.Bool

	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error {
			served.Store(true)
			<-serverStop // ListenAndServe: runs until Shutdown stops it
			return nil
		},
		func(_ context.Context) { close(serverStop) }, // Shutdown
		time.Second,
	), "GoShutdown")

	synctesting.AssertMustEventually(t, served.Load, 100*time.Millisecond,
		"server started")
	wg.Cancel(nil)
	core.AssertNoError(t, wg.Wait(), "wait after shutdown stopped fn")
	core.AssertTrue(t, served.Load(), "served")
}

// runTestGoShutdownCloseSignals pins the defer wg.Close() idiom: Close
// cancels the Group, so it signals the running worker and waits for the
// pair to end before returning.
func runTestGoShutdownCloseSignals(t *testing.T) {
	t.Helper()
	serverStop := make(chan struct{})
	var shutdownRan atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { <-serverStop; return nil },
		func(_ context.Context) {
			shutdownRan.Store(true)
			close(serverStop)
		},
		time.Second,
	), "GoShutdown")

	core.AssertNoError(t, wg.Close(), "close")
	core.AssertTrue(t, shutdownRan.Load(), "shutdown fired by Close")
}

// runTestGoShutdownCleanReturnSkipsShutdown pins that an fn returning
// without the Group being cancelled ends the pair; there is nobody to
// signal, so the shutdown handler never runs.
func runTestGoShutdownCleanReturnSkipsShutdown(t *testing.T) {
	t.Helper()
	var shutdownRan atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { return nil },
		func(_ context.Context) { shutdownRan.Store(true) },
		time.Second,
	), "GoShutdown")

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertFalse(t, shutdownRan.Load(), "shutdown skipped on clean return")
}

// runTestGoShutdownFnErrorSkipsShutdown pins that fn's own error cancels
// the Group but does not invoke its own shutdown handler: the worker has
// already ended, so there is nobody left to signal.
func runTestGoShutdownFnErrorSkipsShutdown(t *testing.T) {
	t.Helper()
	testErr := errors.New("fn failed")
	var shutdownRan atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { return testErr },
		func(_ context.Context) { shutdownRan.Store(true) },
		time.Second,
	), "GoShutdown")

	core.AssertErrorIs(t, wg.Wait(), testErr, "wait err")
	core.AssertFalse(t, shutdownRan.Load(), "shutdown skipped on fn error")
}

// runTestGoShutdownFnPanicSkipsShutdown pins that fn's panic is supervised
// into a PanicError cause without invoking its own shutdown handler.
func runTestGoShutdownFnPanicSkipsShutdown(t *testing.T) {
	t.Helper()
	var shutdownRan atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { panic("boom in fn") },
		func(_ context.Context) { shutdownRan.Store(true) },
		time.Second,
	), "GoShutdown")

	err := wg.Wait()
	core.AssertFalse(t, shutdownRan.Load(), "shutdown skipped on fn panic")
	if panicErr, ok := core.AssertTypeIs[*core.PanicError](t, err,
		"panic type"); ok {
		core.AssertContains(t, panicErr.Error(), "boom in fn", "panic msg")
	}
}

// runTestGoShutdownFnErrorSignalsPeer pins both sides of the rule at once:
// a failing worker cancels the Group without its own shutdown running,
// while the still-running peer is signalled to end. Wait returning at all
// proves the peer's shutdown fired, as nothing else unblocks it.
func runTestGoShutdownFnErrorSignalsPeer(t *testing.T) {
	t.Helper()
	testErr := errors.New("fn failed")
	peerStop := make(chan struct{})
	var ownShutdown atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { <-peerStop; return nil },
		func(_ context.Context) { close(peerStop) },
		time.Second,
	), "peer GoShutdown")
	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { return testErr },
		func(_ context.Context) { ownShutdown.Store(true) },
		time.Second,
	), "failing GoShutdown")

	core.AssertErrorIs(t, wg.Wait(), testErr, "wait err")
	core.AssertFalse(t, ownShutdown.Load(), "own shutdown skipped")
}

// shutdownObservation captures what the shutdown handler observes about the
// context it receives, so a caller can assert its detachment, deadline, and
// values.
type shutdownObservation struct {
	err         error
	value       string
	hasDeadline bool
}

// shutdownCtxKey / shutdownCtxValue seed a value into the parent context so
// the shutdown handler can prove the detached context still carries it.
var shutdownCtxKey = core.NewContextKey[string]("workgroup-test-trace")

const shutdownCtxValue = "trace-xyz"

// observeShutdownContext enrols a server-pattern GoShutdown with the given
// grace period, cancels the Group, and returns what the shutdown handler saw
// of its context. Wait returning proves the handler ran: nothing else
// unblocks fn.
func observeShutdownContext(t *testing.T, grace time.Duration) shutdownObservation {
	t.Helper()
	var obs shutdownObservation
	stop := make(chan struct{})
	parent := shutdownCtxKey.WithValue(context.Background(), shutdownCtxValue)

	wg := workgroup.New(parent)
	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { <-stop; return nil },
		func(ctx context.Context) {
			obs.err = ctx.Err()
			_, obs.hasDeadline = ctx.Deadline()
			if v, ok := ctx.Value(shutdownCtxKey).(string); ok {
				obs.value = v
			}
			close(stop)
		},
		grace,
	), "GoShutdown")

	wg.Cancel(nil)
	core.AssertNoError(t, wg.Wait(), "wait")
	return obs
}

// runTestGoShutdownContextBounded pins that a positive grace period hands
// the shutdown handler a live, value-bearing context with a deadline.
func runTestGoShutdownContextBounded(t *testing.T) {
	t.Helper()
	obs := observeShutdownContext(t, time.Minute)
	core.AssertNoError(t, obs.err, "shutdown context live")
	core.AssertTrue(t, obs.hasDeadline, "deadline present")
	core.AssertEqual(t, shutdownCtxValue, obs.value, "value retained")
}

// runTestGoShutdownContextUnbounded pins that a non-positive grace period
// leaves the shutdown context unbounded — live and value-bearing, no
// deadline.
func runTestGoShutdownContextUnbounded(t *testing.T) {
	t.Helper()
	obs := observeShutdownContext(t, 0)
	core.AssertNoError(t, obs.err, "shutdown context live")
	core.AssertFalse(t, obs.hasDeadline, "no deadline when unbounded")
	core.AssertEqual(t, shutdownCtxValue, obs.value, "value retained")
}

// runTestGoShutdownPanicContained pins that a panicking shutdown handler is
// recovered and discarded, so the signal still lands and Wait returns. The
// handler stops the server before panicking; Wait returning proves both.
func runTestGoShutdownPanicContained(t *testing.T) {
	t.Helper()
	stop := make(chan struct{})
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { <-stop; return nil },
		func(_ context.Context) {
			close(stop)
			panic("boom in shutdown")
		},
		time.Second,
	), "GoShutdown")

	wg.Cancel(nil)
	core.AssertNoError(t, wg.Wait(), "wait returns after contained panic")
}

// runTestGoShutdownWatcherMode pins the nil-fn form: the shutdown handler is
// enrolled alone as a watcher that holds the Group open until cancellation
// signals it.
func runTestGoShutdownWatcherMode(t *testing.T) {
	t.Helper()
	shutdownRan := make(chan struct{})
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(nil,
		func(_ context.Context) { close(shutdownRan) }, time.Second),
		"GoShutdown")

	synctesting.AssertMustOpen(t, shutdownRan, 50*time.Millisecond,
		"shutdown waits for cancellation")
	wg.Cancel(nil)
	synctesting.AssertMustClosed(t, shutdownRan, 100*time.Millisecond,
		"shutdown ran on cancel")
	core.AssertNoError(t, wg.Wait(), "wait")
}

// runTestGoShutdownNilShutdownSuccess pins that a nil shutdown degrades to a
// plain supervised task that runs fn to completion.
func runTestGoShutdownNilShutdownSuccess(t *testing.T) {
	t.Helper()
	var executed atomic.Bool
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { executed.Store(true); return nil },
		nil, time.Second,
	), "GoShutdown")

	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertTrue(t, executed.Load(), "fn executed")
}

// runTestGoShutdownNilShutdownError pins that the nil-shutdown degrade still
// supervises fn: its error cancels the Group, as with GoCatch.
func runTestGoShutdownNilShutdownError(t *testing.T) {
	t.Helper()
	testErr := errors.New("fn failed")
	wg := workgroup.New(context.Background())

	core.AssertNoError(t, wg.GoShutdown(
		func(_ context.Context) error { return testErr }, nil, time.Second,
	), "GoShutdown")

	core.AssertErrorIs(t, wg.Wait(), testErr, "wait err")
}

// runTestGoShutdownBothNil pins that a nil fn and nil shutdown enrol nothing
// and Wait returns at once.
func runTestGoShutdownBothNil(t *testing.T) {
	t.Helper()
	wg := workgroup.New(context.Background())
	core.AssertNoError(t, wg.GoShutdown(nil, nil, time.Second), "both nil")

	done := make(chan struct{})
	go func() { _ = wg.Wait(); close(done) }()
	synctesting.AssertClosed(t, done, 100*time.Millisecond, "wait immediate")
}

// goShutdownCancelledCase exercises one GoShutdown enrolment arm against
// an already cancelled Group: an arm with a handler to enrol refuses with
// ErrClosed, while the both-nil form has nothing to enrol and returns nil.
type goShutdownCancelledCase struct {
	wantErr      error
	name         string
	withFn       bool
	withShutdown bool
}

func (tc goShutdownCancelledCase) Name() string { return tc.name }

func (tc goShutdownCancelledCase) Test(t *testing.T) {
	t.Helper()
	var fnRan, shutdownRan atomic.Bool
	wg := workgroup.New(context.Background())
	wg.Cancel(nil)

	err := wg.GoShutdown(tc.fn(&fnRan), tc.shutdown(&shutdownRan), time.Second)
	if tc.wantErr == nil {
		core.AssertNoError(t, err, "enrol")
	} else {
		core.AssertErrorIs(t, err, tc.wantErr, "enrol")
	}
	core.AssertNoError(t, wg.Wait(), "wait")
	core.AssertFalse(t, fnRan.Load(), "fn not enrolled")
	core.AssertFalse(t, shutdownRan.Load(), "shutdown not enrolled")
}

// fn builds the arm's worker handler, nil when the arm leaves fn unset.
func (tc goShutdownCancelledCase) fn(
	ran *atomic.Bool,
) func(context.Context) error {
	if !tc.withFn {
		return nil
	}
	return func(_ context.Context) error {
		ran.Store(true)
		return nil
	}
}

// shutdown builds the arm's shutdown handler, nil when the arm leaves
// shutdown unset.
func (tc goShutdownCancelledCase) shutdown(
	ran *atomic.Bool,
) func(context.Context) {
	if !tc.withShutdown {
		return nil
	}
	return func(_ context.Context) {
		ran.Store(true)
	}
}

var _ core.TestCase = goShutdownCancelledCase{}

func newGoShutdownCancelledCase(name string, withFn, withShutdown bool,
	wantErr error) goShutdownCancelledCase {
	return goShutdownCancelledCase{
		wantErr:      wantErr,
		name:         name,
		withFn:       withFn,
		withShutdown: withShutdown,
	}
}

// runTestGoShutdownEnrolmentWhenCancelled pins each enrolment arm against
// an already cancelled Group.
func runTestGoShutdownEnrolmentWhenCancelled(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, core.S(
		newGoShutdownCancelledCase("BothHandlers", true, true, errors.ErrClosed),
		newGoShutdownCancelledCase("NilShutdown", true, false, errors.ErrClosed),
		newGoShutdownCancelledCase("Watcher", false, true, errors.ErrClosed),
		newGoShutdownCancelledCase("BothNil", false, false, nil),
	))
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
