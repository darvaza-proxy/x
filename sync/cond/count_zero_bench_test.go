package cond_test

import (
	"testing"

	"darvaza.org/x/sync/cond"
)

// BenchmarkCountZero measures a CountZero going up and back down, which
// broadcasts each time it reaches zero, alone and under RunParallel, and
// a Wait that finds it at zero.
func BenchmarkCountZero(b *testing.B) {
	b.Run("IncDec", runBenchmarkCountZeroIncDec)
	b.Run("IncDec_Parallel", runBenchmarkCountZeroIncDecParallel)
	b.Run("Wait_AtZero", runBenchmarkCountZeroWaitAtZero)
}

func runBenchmarkCountZeroIncDec(b *testing.B) {
	c := cond.NewCountZero(0)
	defer c.Close()

	for b.Loop() {
		c.Inc()
		c.Dec()
	}
}

func runBenchmarkCountZeroIncDecParallel(b *testing.B) {
	c := cond.NewCountZero(0)
	defer c.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
			c.Dec()
		}
	})
}

func runBenchmarkCountZeroWaitAtZero(b *testing.B) {
	c := cond.NewCountZero(0)
	defer c.Close()

	for b.Loop() {
		if err := c.Wait(); err != nil {
			b.Fatalf("Wait returned error: %v", err)
		}
	}
}
