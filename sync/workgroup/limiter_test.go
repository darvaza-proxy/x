package workgroup

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"darvaza.org/x/sync/errors"
)

// TestLimiter runs all tests for Limiter
func TestLimiter(t *testing.T) {
	t.Run("Creation", TestLimiterCreation)
	t.Run("Nil receiver", TestLimiterNilReceiver)

	// Run common WaitGroup tests
	t.Run("Common WaitGroup behaviour", func(t *testing.T) {
		testWaitGroupBehaviours(t, func() WaitGroup {
			lim, err := NewLimiter(5)
			require.NoError(t, err)
			return lim
		})
	})

	// Limiter-specific tests
	t.Run("Limiting behaviour", TestLimiterSpecific)
}

// TestLimiterCreation tests the creation and initialisation of a Limiter
func TestLimiterCreation(t *testing.T) {
	t.Run("NewLimiter", runTestNewLimiter)
	t.Run("MustLimiter", runTestMustLimiter)
	t.Run("Manual initialisation", runTestLimiterManualInit)
	t.Run("Invalid limit", runTestLimiterInvalidLimit)
}

func runTestNewLimiter(t *testing.T) {
	limiter, err := NewLimiter(2)
	require.NoError(t, err)
	assert.NotNil(t, limiter)
	assert.False(t, limiter.IsNil())
	assert.False(t, limiter.IsClosed())
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 2, limiter.Size())
}

func runTestMustLimiter(t *testing.T) {
	assert.NotPanics(t, func() {
		limiter := MustLimiter(2)
		assert.NotNil(t, limiter)
		assert.Equal(t, 2, limiter.Size())
	})

	assert.Panics(t, func() {
		_ = MustLimiter(0)
	})
}

func runTestLimiterManualInit(t *testing.T) {
	limiter := new(Limiter)
	assert.True(t, limiter.IsNil())

	err := limiter.Init(3)
	assert.NoError(t, err)
	assert.False(t, limiter.IsNil())
	assert.False(t, limiter.IsClosed())
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 3, limiter.Size())

	// Init should fail on already initialised limiter
	err = limiter.Init(5)
	assert.Equal(t, errors.ErrAlreadyInitialised, err)
}

func runTestLimiterInvalidLimit(t *testing.T) {
	_, err := NewLimiter(0)
	assert.Error(t, err)

	limiter := new(Limiter)
	err = limiter.Init(0)
	assert.Error(t, err)
}

// TestLimiterNilReceiver tests operations on nil Limiter
func TestLimiterNilReceiver(t *testing.T) {
	var limiter *Limiter

	t.Run("IsNil", func(t *testing.T) {
		assert.True(t, limiter.IsNil())
	})

	t.Run("IsClosed", func(t *testing.T) {
		assert.True(t, limiter.IsClosed())
	})

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 0, limiter.Count())
	})

	t.Run("Size", func(t *testing.T) {
		assert.Equal(t, 0, limiter.Size())
	})

	t.Run("Len", func(t *testing.T) {
		assert.Equal(t, 0, limiter.Len())
	})

	t.Run("Init", func(t *testing.T) {
		err := limiter.Init(2)
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Go", func(t *testing.T) {
		err := limiter.Go(func() {})
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Wait", func(t *testing.T) {
		err := limiter.Wait()
		assert.Equal(t, errors.ErrNilReceiver, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := limiter.Close()
		assert.Equal(t, errors.ErrNilReceiver, err)
	})
}

// TestLimiterSpecific tests the limiting functionality specific to Limiter
func TestLimiterSpecific(t *testing.T) {
	t.Run("Multiple goroutines within limit", runTestLimiterMultipleWithinLimit)
	t.Run("Multiple goroutines exceeding limit", runTestLimiterExceedingLimit)
	t.Run("Close with queued goroutines", runTestLimiterCloseWithQueued)
	t.Run("Wait with active and queued", runTestLimiterWaitActiveAndQueued)
	t.Run("Concurrent Go with limit enforcement", runTestLimiterConcurrentWithLimit)
}

func runTestLimiterMultipleWithinLimit(t *testing.T) {
	limiter, err := NewLimiter(5)
	require.NoError(t, err)
	defer limiter.Close()

	var counter int32
	numGoroutines := 3

	for range numGoroutines {
		err := limiter.Go(func() {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	// All should be running since they're within the limit
	assert.LessOrEqual(t, limiter.Count(), 5) // May be less due to timing
	assert.Equal(t, numGoroutines, limiter.Len())

	err = limiter.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 0, limiter.Len())
}

//revive:disable-next-line:cognitive-complexity
func runTestLimiterExceedingLimit(t *testing.T) {
	limit := 2
	limiter, err := NewLimiter(limit)
	require.NoError(t, err)
	defer limiter.Close()

	var counter int32
	var running int32
	var maxRunning int32
	numGoroutines := 6

	for range numGoroutines {
		err := limiter.Go(func() {
			current := atomic.AddInt32(&running, 1)
			// Track max concurrent runners
			for {
				maxValue := atomic.LoadInt32(&maxRunning)
				if current <= maxValue || atomic.CompareAndSwapInt32(&maxRunning, maxValue, current) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&running, -1)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	// Wait for all tasks to be either running or queued
	time.Sleep(10 * time.Millisecond)

	// Some should be queued
	assert.LessOrEqual(t, limiter.Count(), limit)
	assert.Equal(t, numGoroutines, limiter.Len())

	err = limiter.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 0, limiter.Len())

	// Verify concurrency limit was respected
	assert.LessOrEqual(t, maxRunning, int32(limit))
}

func runTestLimiterCloseWithQueued(t *testing.T) {
	limiter, err := NewLimiter(1)
	require.NoError(t, err)

	var counter int32
	numGoroutines := 3

	// Add some goroutines - one will run, others will be queued
	for range numGoroutines {
		err := limiter.Go(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	// Wait for the first goroutine to start and others to be queued
	time.Sleep(50 * time.Millisecond)

	// Close should succeed immediately
	err = limiter.Close()
	assert.NoError(t, err)
	assert.True(t, limiter.IsClosed())

	// Wait for completion
	err = limiter.Wait()
	assert.NoError(t, err)

	// All goroutines should have completed, even the queued ones
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 0, limiter.Len())
}

func runTestLimiterWaitActiveAndQueued(t *testing.T) {
	limiter, err := NewLimiter(2)
	require.NoError(t, err)
	defer limiter.Close()

	var counter int32
	numGoroutines := 5

	// Add goroutines - some will run immediately, others will be queued
	for range numGoroutines {
		err := limiter.Go(func() {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
		require.NoError(t, err)
	}

	// Wait should block until all goroutines complete, including queued ones
	err = limiter.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(numGoroutines), counter)
	assert.Equal(t, 0, limiter.Count())
	assert.Equal(t, 0, limiter.Len())
}

//revive:disable-next-line:cognitive-complexity
func runTestLimiterConcurrentWithLimit(t *testing.T) {
	limiter, err := NewLimiter(5)
	require.NoError(t, err)
	defer limiter.Close()

	var counter int32
	var running int32
	var maxRunning int32
	numGoroutines := 100

	for range numGoroutines {
		go func() {
			err := limiter.Go(func() {
				current := atomic.AddInt32(&running, 1)
				// Track max concurrent runners
				for {
					maxValue := atomic.LoadInt32(&maxRunning)
					if current <= maxValue || atomic.CompareAndSwapInt32(&maxRunning, maxValue, current) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&running, -1)
				atomic.AddInt32(&counter, 1)
			})
			if err != nil {
				t.Logf("Got error: %v", err)
			}
		}()
	}

	// Wait for all goroutines to be scheduled
	time.Sleep(50 * time.Millisecond)

	err = limiter.Wait()
	assert.NoError(t, err)

	// We can't assert exact count due to race conditions with Close
	// but we can check we have at least some increments
	assert.Greater(t, int(counter), 0)
	// Verify the concurrency limit was enforced
	assert.LessOrEqual(t, int(maxRunning), 5)
}
