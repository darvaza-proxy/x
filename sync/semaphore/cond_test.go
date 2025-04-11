package semaphore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCond verifies the proper initialisation of Cond objects
// with different initial values.
func TestNewCond(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{"Zero", 0},
		{"Positive", 42},
		{"Negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewCond(tt.value)
			require.NotNil(t, cond)
			assert.Equal(t, tt.value, cond.Value())
			// Verify the internal waiters map is properly initialised
			require.NotNil(t, cond.waiters)
		})
	}
}

// TestCond_AtomicOperations confirms atomic operations
// (Add, Inc, Dec, Value) work correctly and maintain consistency.
func TestCond_AtomicOperations(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		cond := NewCond(10)
		assert.Equal(t, 15, cond.Add(5))
		assert.Equal(t, 5, cond.Add(-10))
		assert.Equal(t, 5, cond.Value())

		// Test that Add(0) maintains the same value and doesn't affect
		// subsequent operations
		assert.Equal(t, 5, cond.Add(0), "Add(0) should return current value")
		assert.Equal(t, 5, cond.Value(), "Value shouldn't change after Add(0)")
		assert.Equal(t, 8, cond.Add(3), "Operations after Add(0) should work correctly")
		assert.Equal(t, 8, cond.Value())
	})

	t.Run("Inc", func(t *testing.T) {
		cond := NewCond(0)
		assert.Equal(t, 1, cond.Inc())
		assert.Equal(t, 2, cond.Inc())
		assert.Equal(t, 2, cond.Value())
	})

	t.Run("Dec", func(t *testing.T) {
		cond := NewCond(5)
		assert.Equal(t, 4, cond.Dec())
		assert.Equal(t, 3, cond.Dec())
		assert.Equal(t, 3, cond.Value())
	})

	t.Run("Value", func(t *testing.T) {
		cond := NewCond(42)
		assert.Equal(t, 42, cond.Value())
		cond.Add(10)
		assert.Equal(t, 52, cond.Value())
	})

	// Test multiple operations in sequence
	t.Run("SequentialOperations", func(t *testing.T) {
		cond := NewCond(0)
		cond.Inc()
		cond.Inc()
		cond.Add(3)
		assert.Equal(t, 5, cond.Value())
		cond.Dec()
		assert.Equal(t, 4, cond.Value())
		cond.Add(-4)
		assert.Equal(t, 0, cond.Value())
	})
}

// TestCond_PanicsOnNilReceiver ensures all methods correctly panic
// when called on a nil Cond receiver.
//
//revive:disable-next-line:cognitive-complexity
func TestCond_PanicsOnNilReceiver(t *testing.T) {
	var nilCond *Cond

	t.Run("Add", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Add(5) })
	})

	t.Run("Inc", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Inc() })
	})

	t.Run("Dec", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Dec() })
	})

	t.Run("Value", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Value() })
	})

	t.Run("WaitZero", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.WaitZero() })
	})

	t.Run("WaitFn", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.WaitFn(func(int32) bool { return true }) })
	})

	t.Run("Match", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Match(func(int32) bool { return true }) })
	})

	t.Run("IsZero", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.IsZero() })
	})

	t.Run("Signal", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Signal() })
	})

	t.Run("Broadcast", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Broadcast() })
	})
}

// TestCond_IsZero tests the IsZero method across different scenarios.
func TestCond_IsZero(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		cond := NewCond(0)
		assert.True(t, cond.IsZero())
	})

	t.Run("NonZero", func(t *testing.T) {
		cond := NewCond(1)
		assert.False(t, cond.IsZero())
	})

	t.Run("BecomeZero", func(t *testing.T) {
		cond := NewCond(5)
		assert.False(t, cond.IsZero())
		cond.Add(-5)
		assert.True(t, cond.IsZero())
	})
}

// TestCond_Match tests the Match method with various condition functions.
func TestCond_Match(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		checkFn  CondFunc
		expected bool
	}{
		{"IsZero", 0, nil, true},                                     // nil function checks for zero
		{"NotZero", 1, nil, false},                                   // nil function checks for zero
		{"IsPositive", 5, func(v int32) bool { return v > 0 }, true}, // custom condition
		{"IsNegative", -3, func(v int32) bool { return v < 0 }, true},
		{"IsEven", 4, func(v int32) bool { return v%2 == 0 }, true},
		{"IsOdd", 3, func(v int32) bool { return v%2 == 0 }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewCond(tt.value)
			assert.Equal(t, tt.expected, cond.Match(tt.checkFn))
		})
	}
}

// TestCond_WaitMethods tests all waiting-related methods of Cond.
//
//revive:disable-next-line:cognitive-complexity
func TestCond_WaitMethods(t *testing.T) {
	t.Run("WaitZero", func(t *testing.T) {
		cond := NewCond(1)

		// Start a goroutine that will decrement after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cond.Dec()
		}()

		// This should not block once the value reaches zero
		cond.WaitZero()
		assert.Equal(t, 0, cond.Value())
	})

	t.Run("WaitFn", func(t *testing.T) {
		cond := NewCond(10)

		// Start a goroutine that will decrement until the value is 5
		go func() {
			for i := 0; i < 5; i++ {
				time.Sleep(20 * time.Millisecond)
				cond.Dec()
			}
		}()

		// Wait until the value is <= 5
		cond.WaitFn(func(v int32) bool { return v <= 5 })
		assert.LessOrEqual(t, cond.Value(), 5)
	})

	t.Run("WaitFnAbort", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			cond := NewCond(3)
			abort := make(chan struct{})

			go func() {
				time.Sleep(50 * time.Millisecond)
				cond.Dec()
			}()

			err := cond.WaitFnAbort(abort, func(v int32) bool { return v < 3 })
			assert.NoError(t, err)
			assert.Equal(t, 2, cond.Value())
		})

		t.Run("Aborted", func(t *testing.T) {
			cond := NewCond(5)
			abort := make(chan struct{})

			go func() {
				time.Sleep(50 * time.Millisecond)
				close(abort)
			}()

			err := cond.WaitFnAbort(abort, func(v int32) bool { return v == 0 })
			assert.Error(t, err)
			assert.Equal(t, context.Canceled, err)
			assert.Equal(t, 5, cond.Value()) // Value should remain unchanged
		})

		t.Run("NilReceiver", func(t *testing.T) {
			var nilCond *Cond
			abort := make(chan struct{})

			err := nilCond.WaitFnAbort(abort, func(_ int32) bool { return true })
			assert.Error(t, err)
		})

		t.Run("AlreadyAborted", func(t *testing.T) {
			cond := NewCond(5)
			abort := make(chan struct{})
			close(abort) // Close before calling WaitFnAbort

			err := cond.WaitFnAbort(abort, func(v int32) bool { return v == 0 })
			assert.Error(t, err)
			assert.Equal(t, context.Canceled, err)
		})
	})

	t.Run("WaitFnContext", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			cond := NewCond(2)
			ctx := context.Background()

			go func() {
				time.Sleep(50 * time.Millisecond)
				cond.Dec()
				cond.Dec()
			}()

			err := cond.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
			assert.NoError(t, err)
			assert.Equal(t, 0, cond.Value())
		})

		t.Run("Canceled", func(t *testing.T) {
			cond := NewCond(10)
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := cond.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
			assert.Error(t, err)
			assert.Equal(t, context.DeadlineExceeded, err)
			assert.Equal(t, 10, cond.Value()) // Value should remain unchanged
		})

		t.Run("NilContext", func(t *testing.T) {
			var nilCtx context.Context
			cond := NewCond(1)

			err := cond.WaitFnContext(nilCtx, func(_ int32) bool { return true })
			assert.Error(t, err)
		})

		t.Run("NilReceiver", func(t *testing.T) {
			var nilCond *Cond
			ctx := context.Background()

			err := nilCond.WaitFnContext(ctx, func(_ int32) bool { return true })
			assert.Error(t, err)
		})

		t.Run("ConditionAlreadyMet", func(t *testing.T) {
			cond := NewCond(0)
			ctx := context.Background()

			// Should return immediately since condition is already met
			err := cond.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
			assert.NoError(t, err)
		})
	})
}

// TestCond_SignalAndBroadcast tests the signalling mechanisms for waiters.
//
//revive:disable-next-line:cognitive-complexity
func TestCond_SignalAndBroadcast(t *testing.T) {
	t.Run("Signal_NoWaiters", func(t *testing.T) {
		cond := NewCond(0)
		// No waiters, should return false
		assert.False(t, cond.Signal())
	})

	t.Run("Broadcast_NoWaiters", func(t *testing.T) {
		cond := NewCond(0)
		// No waiters, should return false
		assert.False(t, cond.Broadcast())
	})

	t.Run("Signal_WithWaiters", func(t *testing.T) {
		cond := NewCond(5)

		var wg sync.WaitGroup
		wg.Add(1)

		// Start a goroutine that will wait for the value to become even
		go func() {
			defer wg.Done()
			cond.WaitFn(func(v int32) bool { return v%2 == 0 })
		}()

		// Give time for the goroutine to register as a waiter
		time.Sleep(50 * time.Millisecond)

		// Make value even but don't signal
		cond.Add(1)

		// Explicitly signal the waiter
		signaled := cond.Signal()
		assert.True(t, signaled)

		// Wait for waiter to finish
		wg.Wait()
	})

	t.Run("Broadcast_WithWaiters", func(t *testing.T) {
		cond := NewCond(5)

		var wg sync.WaitGroup
		const numWaiters = 3
		wg.Add(numWaiters)

		// Start multiple goroutines waiting for different conditions
		for i := 0; i < numWaiters; i++ {
			go func(i int) {
				defer wg.Done()
				targetValue := 5 + i
				cond.WaitFn(func(v int32) bool { return v >= int32(targetValue) })
			}(i)
		}

		// Give time for goroutines to register as waiters
		time.Sleep(50 * time.Millisecond)

		// Make value satisfy all conditions
		cond.Add(3)

		// Broadcast to all waiters
		signaled := cond.Broadcast()
		assert.True(t, signaled)

		// Wait for all waiters to finish
		wg.Wait()
	})

	t.Run("BroadcastTriggersAllWaiters", func(t *testing.T) {
		cond := NewCond(0)

		const numWaiters = 5
		finished := make([]bool, numWaiters)
		var mu sync.Mutex

		// Start multiple goroutines waiting for different conditions
		for i := 0; i < numWaiters; i++ {
			go func(i int) {
				// All waiters will wait for any value change
				cond.WaitFn(func(v int32) bool { return v > 0 })

				mu.Lock()
				finished[i] = true
				mu.Unlock()
			}(i)
		}

		// Give time for goroutines to register as waiters
		time.Sleep(50 * time.Millisecond)

		// Change value and broadcast
		cond.Inc()
		cond.Broadcast()

		// Wait for all waiters to process the broadcast
		time.Sleep(50 * time.Millisecond)

		// Verify all waiters were triggered
		mu.Lock()
		for i := 0; i < numWaiters; i++ {
			assert.True(t, finished[i], "Waiter %d was not triggered by broadcast", i)
		}
		mu.Unlock()
	})
}

// TestCond_EdgeCases tests various edge cases and error conditions.
func TestCond_EdgeCases(t *testing.T) {
	t.Run("LazyInitialisation", func(t *testing.T) {
		// Create a Cond object manually without using NewCond
		cond := &Cond{value: 42}

		// Call a method that should trigger lazy initialisation
		assert.Equal(t, 42, cond.Value())

		// Verify internal structures are now initialised
		assert.NotNil(t, cond.waiters)
	})

	t.Run("CancelledWaitWithoutBlock", func(t *testing.T) {
		cond := NewCond(0)
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel before waiting
		cancel()

		// Should return immediately with error
		err := cond.WaitFnContext(ctx, func(v int32) bool { return v > 0 })
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("WaitForNegativeValue", func(t *testing.T) {
		cond := NewCond(10)

		// Start a goroutine that makes the value negative
		go func() {
			time.Sleep(50 * time.Millisecond)
			cond.Add(-15) // 10 -> -5
		}()

		// Wait for negative value
		cond.WaitFn(func(v int32) bool { return v < 0 })
		assert.Less(t, cond.Value(), 0)
	})

	t.Run("NoOpConditionFunction", func(t *testing.T) {
		cond := NewCond(5)

		// Create a condition function that always returns true
		alwaysTrue := func(int32) bool { return true }

		// WaitFn should return immediately
		cond.WaitFn(alwaysTrue)

		// Try the same with context
		err := cond.WaitFnContext(context.Background(), alwaysTrue)
		assert.NoError(t, err)

		// And with abort channel
		abort := make(chan struct{})
		err = cond.WaitFnAbort(abort, alwaysTrue)
		assert.NoError(t, err)
	})
}

// TestCond_Concurrency tests the Cond type under concurrent usage.
func TestCond_Concurrency(t *testing.T) {
	const (
		goroutines = 10
		iterations = 100
	)

	cond := NewCond(0)

	// Launch multiple goroutines to increment the counter
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				cond.Inc()
			}
		}()
	}

	// Wait for the final expected value
	cond.WaitFn(func(v int32) bool {
		return v == goroutines*iterations
	})

	assert.Equal(t, goroutines*iterations, cond.Value())
}

// TestCond_ComplexConcurrency tests complex concurrent usage patterns.
//
//revive:disable-next-line:cognitive-complexity
func TestCond_ComplexConcurrency(t *testing.T) {
	cond := NewCond(0)

	const numInc = 5
	const numDec = 3
	const incOps = 20
	const decOps = 10

	var wg sync.WaitGroup
	wg.Add(numInc + numDec)

	// Start incrementer routines
	for i := 0; i < numInc; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incOps; j++ {
				cond.Inc()
				time.Sleep(1 * time.Millisecond) // Small delay to increase contention
			}
		}()
	}

	// Start decrementer routines
	for i := 0; i < numDec; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < decOps; j++ {
				// Wait until there's something to decrement
				cond.WaitFn(func(v int32) bool { return v > 0 })
				cond.Dec()
				time.Sleep(2 * time.Millisecond) // Small delay to increase contention
			}
		}()
	}

	// Wait for all operations to complete
	wg.Wait()

	// Calculate expected final value
	expectedValue := (numInc * incOps) - (numDec * decOps)
	assert.Equal(t, expectedValue, cond.Value())
}

// Benchmarks for Cond operations

// BenchmarkCond_AtomicOperations benchmarks the atomic operations
func BenchmarkCond_AtomicOperations(b *testing.B) {
	b.Run("Add", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Add(1)
		}
	})

	b.Run("Inc", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Inc()
		}
	})

	b.Run("Dec", func(b *testing.B) {
		cond := NewCond(b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Dec()
		}
	})

	b.Run("Value", func(b *testing.B) {
		cond := NewCond(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cond.Value()
		}
	})
}

// BenchmarkCond_ConditionChecking benchmarks the condition checking methods
func BenchmarkCond_ConditionChecking(b *testing.B) {
	b.Run("IsZero", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.IsZero()
		}
	})

	b.Run("Match_Zero", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Match(nil)
		}
	})

	b.Run("Match_Custom", func(b *testing.B) {
		cond := NewCond(42)
		isEven := func(v int32) bool { return v%2 == 0 }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Match(isEven)
		}
	})
}

// BenchmarkCond_Signaling benchmarks the signaling methods
func BenchmarkCond_Signaling(b *testing.B) {
	b.Run("Signal_NoWaiters", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Signal()
		}
	})

	b.Run("Broadcast_NoWaiters", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.Broadcast()
		}
	})

	b.Run("Signal_WithWaiters", func(b *testing.B) {
		cond := NewCond(1)
		const waiters = 10

		b.ResetTimer()
		b.StopTimer()

		for i := 0; i < b.N; i++ {
			// Setup waiters
			var wg sync.WaitGroup
			wg.Add(waiters)

			for j := 0; j < waiters; j++ {
				go func() {
					defer wg.Done()
					cond.WaitFn(func(v int32) bool { return v == 0 })
				}()
			}

			// Give time for waiters to register
			time.Sleep(1 * time.Millisecond)

			b.StartTimer()
			cond.Dec() // makes value 0
			cond.Signal()
			b.StopTimer()

			// Wait for all waiters to finish
			wg.Wait()

			// Reset for next iteration
			cond.Add(1)
		}
	})

	b.Run("Broadcast_WithWaiters", func(b *testing.B) {
		cond := NewCond(1)
		const waiters = 10

		b.ResetTimer()
		b.StopTimer()

		for i := 0; i < b.N; i++ {
			// Setup waiters
			var wg sync.WaitGroup
			wg.Add(waiters)

			for j := 0; j < waiters; j++ {
				go func() {
					defer wg.Done()
					cond.WaitFn(func(v int32) bool { return v == 0 })
				}()
			}

			// Give time for waiters to register
			time.Sleep(1 * time.Millisecond)

			b.StartTimer()
			cond.Dec() // makes value 0
			cond.Broadcast()
			b.StopTimer()

			// Wait for all waiters to finish
			wg.Wait()

			// Reset for next iteration
			cond.Add(1)
		}
	})
}

// BenchmarkCond_WaitOperations benchmarks the waiting operations
func BenchmarkCond_WaitOperations(b *testing.B) {
	b.Run("WaitFn_ConditionAlreadyTrue", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cond.WaitFn(func(v int32) bool { return v == 0 })
		}
	})

	b.Run("WaitFnContext_ConditionAlreadyTrue", func(b *testing.B) {
		cond := NewCond(0)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cond.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
		}
	})

	b.Run("WaitFnAbort_ConditionAlreadyTrue", func(b *testing.B) {
		cond := NewCond(0)
		abort := make(chan struct{})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cond.WaitFnAbort(abort, func(v int32) bool { return v == 0 })
		}
	})
}

// BenchmarkCond_ConcurrentOperations benchmarks concurrent operations on Cond
func BenchmarkCond_ConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentIncDec", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				cond.Inc()
				cond.Dec()
			}
		})
	})

	b.Run("ConcurrentRead", func(b *testing.B) {
		cond := NewCond(42)
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = cond.Value()
			}
		})
	})

	b.Run("ConcurrentAddWithSignal", func(b *testing.B) {
		cond := NewCond(0)
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				cond.Add(1)
				cond.Signal()
			}
		})
	})
}

// BenchmarkCond_WaitAndSignal benchmarks wait-and-signal patterns
func BenchmarkCond_WaitAndSignal(b *testing.B) {
	b.Run("ProducerConsumer", func(b *testing.B) {
		cond := NewCond(0)
		const maxVal = 10
		done := make(chan struct{})

		// Start consumer goroutine
		go func() {
			for {
				cond.WaitFn(func(v int32) bool { return v > 0 })
				cond.Dec()

				select {
				case <-done:
					return
				default:
				}
			}
		}()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// producer
			for cond.Value() >= maxVal {
				// Wait for consumer to catch up
				time.Sleep(10 * time.Microsecond)
			}
			cond.Inc()
		}

		close(done)
	})

	b.Run("MultipleProducersConsumers", func(b *testing.B) {
		if b.N < 1000 {
			b.Skip("Skipping small benchmark run")
		}

		cond := NewCond(0)
		itemsPerGoroutine := b.N / 10
		if itemsPerGoroutine < 1 {
			itemsPerGoroutine = 1
		}

		var wg sync.WaitGroup
		wg.Add(10) // 5 producers, 5 consumers

		// Start producer goroutines
		for i := 0; i < 5; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					cond.Inc()
					cond.Signal()
				}
			}()
		}

		// Start consumer goroutines
		for i := 0; i < 5; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					cond.WaitFn(func(v int32) bool { return v > 0 })
					cond.Dec()
				}
			}()
		}

		wg.Wait()
	})
}
