package spinlock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			assert.Equal(t, int32(tt.value), cond.Value())
		})
	}
}

func TestCond_AtomicOperations(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		cond := NewCond(10)
		assert.Equal(t, int32(15), cond.Add(5))
		assert.Equal(t, int32(5), cond.Add(-10))
		assert.Equal(t, int32(5), cond.Value())
	})

	t.Run("Inc", func(t *testing.T) {
		cond := NewCond(0)
		assert.Equal(t, int32(1), cond.Inc())
		assert.Equal(t, int32(2), cond.Inc())
		assert.Equal(t, int32(2), cond.Value())
	})

	t.Run("Dec", func(t *testing.T) {
		cond := NewCond(5)
		assert.Equal(t, int32(4), cond.Dec())
		assert.Equal(t, int32(3), cond.Dec())
		assert.Equal(t, int32(3), cond.Value())
	})

	t.Run("Value", func(t *testing.T) {
		cond := NewCond(42)
		assert.Equal(t, int32(42), cond.Value())
		cond.Add(10)
		assert.Equal(t, int32(52), cond.Value())
	})
}

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

	t.Run("Is", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.Is(func(int32) bool { return true }) })
	})

	t.Run("IsZero", func(t *testing.T) {
		assert.Panics(t, func() { nilCond.IsZero() })
	})
}

func TestCond_IsZero(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		cond := NewCond(0)
		assert.True(t, cond.IsZero())
	})

	t.Run("NonZero", func(t *testing.T) {
		cond := NewCond(1)
		assert.False(t, cond.IsZero())
	})
}

func TestCond_Is(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		checkFn  CondFunc
		expected bool
	}{
		{"IsZero", 0, nil, true},
		{"NotZero", 1, nil, false},
		{"IsPositive", 5, func(v int32) bool { return v > 0 }, true},
		{"IsNegative", -3, func(v int32) bool { return v < 0 }, true},
		{"IsEven", 4, func(v int32) bool { return v%2 == 0 }, true},
		{"IsOdd", 3, func(v int32) bool { return v%2 == 0 }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewCond(tt.value)
			assert.Equal(t, tt.expected, cond.Is(tt.checkFn))
		})
	}
}

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
		assert.Equal(t, int32(0), cond.Value())
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
		assert.LessOrEqual(t, cond.Value(), int32(5))
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
			assert.Equal(t, int32(2), cond.Value())
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
			assert.Equal(t, int32(5), cond.Value()) // Value should remain unchanged
		})

		t.Run("NilReceiver", func(t *testing.T) {
			var nilCond *Cond
			abort := make(chan struct{})

			err := nilCond.WaitFnAbort(abort, func(_ int32) bool { return true })
			assert.Error(t, err)
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
			assert.Equal(t, int32(0), cond.Value())
		})

		t.Run("Canceled", func(t *testing.T) {
			cond := NewCond(10)
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := cond.WaitFnContext(ctx, func(v int32) bool { return v == 0 })
			assert.Error(t, err)
			assert.Equal(t, context.DeadlineExceeded, err)
			assert.Equal(t, int32(10), cond.Value()) // Value should remain unchanged
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
	})
}

// TestCond_Concurrency tests the Cond type under concurrent usage
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

	assert.Equal(t, int32(goroutines*iterations), cond.Value())
}
