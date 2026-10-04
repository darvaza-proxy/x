package cond_test

import (
	"testing"

	"darvaza.org/x/sync/cond"
)

// BenchmarkBarrier measures taking and giving back a Barrier's token,
// alone and under RunParallel, and reading, signalling and broadcasting
// it with no waiters.
func BenchmarkBarrier(b *testing.B) {
	b.Run("AcquireRelease", runBenchmarkBarrierAcquireRelease)
	b.Run("AcquireRelease_Parallel", runBenchmarkBarrierAcquireReleaseParallel)
	b.Run("TryAcquire", runBenchmarkBarrierTryAcquire)
	b.Run("Token", runBenchmarkBarrierToken)
	b.Run("Signal_NoWaiters", runBenchmarkBarrierSignalNoWaiters)
	b.Run("Broadcast_NoWaiters", runBenchmarkBarrierBroadcastNoWaiters)
}

func runBenchmarkBarrierAcquireRelease(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	for b.Loop() {
		tok := <-bs.Acquire()
		bs.Release(tok)
	}
}

func runBenchmarkBarrierAcquireReleaseParallel(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tok := <-bs.Acquire()
			bs.Release(tok)
		}
	})
}

func runBenchmarkBarrierTryAcquire(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	for b.Loop() {
		tok, ok := bs.TryAcquire()
		if !ok {
			b.Fatal("TryAcquire failed on a free Barrier")
		}
		bs.Release(tok)
	}
}

func runBenchmarkBarrierToken(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	for b.Loop() {
		bs.Token()
	}
}

func runBenchmarkBarrierSignalNoWaiters(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	for b.Loop() {
		bs.Signal()
	}
}

func runBenchmarkBarrierBroadcastNoWaiters(b *testing.B) {
	bs := cond.NewBarrier()
	defer bs.Close()

	for b.Loop() {
		bs.Broadcast()
	}
}
