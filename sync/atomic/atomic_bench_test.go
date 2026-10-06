package atomic_test

import (
	"runtime"
	"testing"

	"darvaza.org/x/sync/atomic"
)

// BenchmarkBitmaskOr measures BitmaskOr setting a bit and finding it
// already set, alone and under RunParallel. To set it each time, the
// serial loop clears it with a store first, and each parallel goroutine
// clears a bit of its own with And, 32 goroutines to a word, so their ORs
// race one another.
func BenchmarkBitmaskOr(b *testing.B) {
	b.Run("Changed", runBenchmarkBitmaskOrChanged)
	b.Run("Changed_Parallel", runBenchmarkBitmaskOrChangedParallel)
	b.Run("Unchanged", runBenchmarkBitmaskOrUnchanged)
	b.Run("Unchanged_Parallel", runBenchmarkBitmaskOrUnchangedParallel)
}

func runBenchmarkBitmaskOrChanged(b *testing.B) {
	var p atomic.Uint32

	for b.Loop() {
		p.Store(0)
		atomic.BitmaskOr(&p, 1)
	}
}

func runBenchmarkBitmaskOrChangedParallel(b *testing.B) {
	// RunParallel starts a goroutine per GOMAXPROCS.
	words := make([]atomic.Uint32, (runtime.GOMAXPROCS(0)+31)/32)
	var next atomic.Uint32

	b.RunParallel(func(pb *testing.PB) {
		n := next.Add(1) - 1
		p := &words[n/32]
		bit := uint32(1) << (n % 32)
		for pb.Next() {
			atomic.BitmaskOr(p, bit)
			p.And(^bit)
		}
	})
}

func runBenchmarkBitmaskOrUnchanged(b *testing.B) {
	var p atomic.Uint32
	p.Store(1)

	for b.Loop() {
		atomic.BitmaskOr(&p, 1)
	}
}

func runBenchmarkBitmaskOrUnchangedParallel(b *testing.B) {
	var p atomic.Uint32
	p.Store(1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.BitmaskOr(&p, 1)
		}
	})
}

// BenchmarkUpdateMax measures UpdateMax raising the value and finding it
// already higher, and goroutines under RunParallel racing to raise it, so
// some calls raise it and others find it higher.
func BenchmarkUpdateMax(b *testing.B) {
	b.Run("Raised", runBenchmarkUpdateMaxRaised)
	b.Run("Unchanged", runBenchmarkUpdateMaxUnchanged)
	b.Run("Parallel", runBenchmarkUpdateMaxParallel)
}

func runBenchmarkUpdateMaxRaised(b *testing.B) {
	var p atomic.Int32
	var v int32

	for b.Loop() {
		v++
		atomic.UpdateMax(&p, v)
	}
}

func runBenchmarkUpdateMaxUnchanged(b *testing.B) {
	var p atomic.Int32
	p.Store(1)

	for b.Loop() {
		atomic.UpdateMax(&p, 0)
	}
}

func runBenchmarkUpdateMaxParallel(b *testing.B) {
	var p atomic.Int32

	b.RunParallel(func(pb *testing.PB) {
		var v int32
		for pb.Next() {
			v++
			atomic.UpdateMax(&p, v)
		}
	})
}
