package cond_test

import (
	"testing"

	"darvaza.org/x/sync/cond"
)

// BenchmarkCount provides performance measurements for the Count struct.
func BenchmarkCount(b *testing.B) {
	b.Run("Add", runBenchmarkCountAdd)
	b.Run("Inc", runBenchmarkCountInc)
	b.Run("Value", runBenchmarkCountValue)
	b.Run("Signal_NoWaiters", runBenchmarkCountSignalNoWaiters)
	b.Run("Broadcast_NoWaiters", runBenchmarkCountBroadcastNoWaiters)
	b.Run("Match", runBenchmarkCountMatch)
}

func runBenchmarkCountAdd(b *testing.B) {
	c := cond.NewCount(0)
	defer c.Close()

	for b.Loop() {
		c.Add(1)
	}
}

func runBenchmarkCountInc(b *testing.B) {
	c := cond.NewCount(0)
	defer c.Close()

	for b.Loop() {
		c.Inc()
	}
}

func runBenchmarkCountValue(b *testing.B) {
	c := cond.NewCount(42)
	defer c.Close()

	for b.Loop() {
		_ = c.Value()
	}
}

func runBenchmarkCountSignalNoWaiters(b *testing.B) {
	c := cond.NewCount(0)
	defer c.Close()

	for b.Loop() {
		c.Signal()
	}
}

func runBenchmarkCountBroadcastNoWaiters(b *testing.B) {
	c := cond.NewCount(0)
	defer c.Close()

	for b.Loop() {
		c.Broadcast()
	}
}

func runBenchmarkCountMatch(b *testing.B) {
	c := cond.NewCount(42)
	defer c.Close()

	isFortyTwo := func(v int32) bool { return v == 42 }

	for b.Loop() {
		c.Match(isFortyTwo)
	}
}
