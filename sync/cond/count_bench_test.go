package cond

import (
	"testing"
)

// BenchmarkCount provides performance measurements for the Count struct.
//
//revive:disable-next-line:cognitive-complexity
func BenchmarkCount(b *testing.B) {
	// Benchmark Add operation
	b.Run("Add", func(b *testing.B) {
		c := NewCount(0)
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Add(1)
		}
	})

	// Benchmark Inc operation
	b.Run("Inc", func(b *testing.B) {
		c := NewCount(0)
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Inc()
		}
	})

	// Benchmark Value operation
	b.Run("Value", func(b *testing.B) {
		c := NewCount(42)
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = c.Value()
		}
	})

	// Benchmark Signal with no waiters
	b.Run("Signal_NoWaiters", func(b *testing.B) {
		c := NewCount(0)
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Signal()
		}
	})

	// Benchmark Broadcast with no waiters
	b.Run("Broadcast_NoWaiters", func(b *testing.B) {
		c := NewCount(0)
		defer c.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Broadcast()
		}
	})

	// Benchmark condition matching
	b.Run("Match", func(b *testing.B) {
		c := NewCount(42)
		defer c.Close()

		isFortyTwo := func(v int32) bool { return v == 42 }

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Match(isFortyTwo)
		}
	})
}
