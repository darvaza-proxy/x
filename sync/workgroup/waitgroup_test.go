package workgroup

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"darvaza.org/x/sync/errors"
)

// testWaitGroupBehaviours runs common tests on any WaitGroup implementation
func testWaitGroupBehaviours(t *testing.T, factory func() WaitGroup) {
	t.Run("Go operations", func(t *testing.T) {
		t.Run("Single goroutine", func(t *testing.T) {
			testSingleGoroutine(t, factory())
		})
		t.Run("Multiple goroutines", func(t *testing.T) {
			testMultipleGoroutines(t, factory())
		})
		t.Run("Nil function", func(t *testing.T) {
			testNilFunction(t, factory())
		})
	})

	t.Run("Close operations", func(t *testing.T) {
		t.Run("Close empty", func(t *testing.T) {
			testCloseEmpty(t, factory())
		})
		t.Run("Close with active goroutines", func(t *testing.T) {
			testCloseWithActive(t, factory())
		})
	})

	t.Run("Wait operations", func(t *testing.T) {
		t.Run("Wait on empty", func(t *testing.T) {
			testWaitEmpty(t, factory())
		})
	})

	t.Run("Concurrency", func(t *testing.T) {
		t.Run("Concurrent Go and Close", func(t *testing.T) {
			testConcurrentCloseAndGo(t, factory())
		})
	})

	t.Run("Interface methods", func(t *testing.T) {
		testInterfaceMethods(t, factory())
	})
}

// Common test helpers
func testSingleGoroutine(t *testing.T, wg WaitGroup) {
	defer wg.Close()

	var counter int32
	err := wg.Go(func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	})

	require.NoError(t, err)
	assert.Equal(t, 1, wg.Count())

	err = wg.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(1), counter)
	assert.Equal(t, 0, wg.Count())
}

func testMultipleGoroutines(t *testing.T, wg WaitGroup) {
	defer wg.Close()

	var counter int32
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		err := wg.Go(func() {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	err := wg.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, wg.Count())
}

func testNilFunction(t *testing.T, wg WaitGroup) {
	defer wg.Close()

	// Providing nil function should be a no-op
	err := wg.Go(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, wg.Count())
}

func testCloseEmpty(t *testing.T, wg WaitGroup) {
	err := wg.Close()
	assert.NoError(t, err)
	assert.True(t, wg.IsClosed())

	// Second close should fail
	err = wg.Close()
	assert.Equal(t, errors.ErrClosed, err)
}

func testCloseWithActive(t *testing.T, wg WaitGroup) {
	var counter int32
	numGoroutines := 5

	// Add some delay to ensure goroutines are still running when we close
	for i := 0; i < numGoroutines; i++ {
		err := wg.Go(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	// Close should succeed immediately
	err := wg.Close()
	assert.NoError(t, err)
	assert.True(t, wg.IsClosed())

	// Adding more goroutines should fail
	err = wg.Go(func() {})
	assert.Equal(t, errors.ErrClosed, err)

	// Wait for completion
	err = wg.Wait()
	assert.NoError(t, err)

	// All goroutines should have completed
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, wg.Count())
}

func testWaitEmpty(t *testing.T, wg WaitGroup) {
	defer wg.Close()

	err := wg.Wait()
	assert.NoError(t, err)
}

//revive:disable-next-line:cognitive-complexity
func testConcurrentCloseAndGo(t *testing.T, wg WaitGroup) {
	// Start goroutines that will try to add tasks
	for range 10 {
		go func() {
			for range 10 {
				_ = wg.Go(func() {
					time.Sleep(10 * time.Millisecond)
				})
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	// Close in the middle of operations
	time.Sleep(20 * time.Millisecond)
	err := wg.Close()
	assert.NoError(t, err)

	// Wait should succeed
	err = wg.Wait()
	assert.NoError(t, err)
	assert.Equal(t, 0, wg.Count())
}

func testInterfaceMethods(t *testing.T, wg WaitGroup) {
	assert.False(t, wg.IsClosed())
	assert.Equal(t, 0, wg.Count())

	err := wg.Go(func() {})
	assert.NoError(t, err)

	err = wg.Wait()
	assert.NoError(t, err)

	err = wg.Close()
	assert.NoError(t, err)

	// Adding more after close should fail
	err = wg.Go(func() {})
	assert.Equal(t, errors.ErrClosed, err)
}
