package workgroup

import (
	"context"
	"testing"
)

// BenchmarkLazyInit measures lazyInit on a Group that is already
// initialised, the state every call after the first finds it in. The
// methods that start, wait for and cancel tasks run it first, so this is
// the entry cost each of them pays.
func BenchmarkLazyInit(b *testing.B) {
	b.Run("serial", runBenchmarkLazyInitSerial)
	b.Run("parallel", runBenchmarkLazyInitParallel)
}

func runBenchmarkLazyInitSerial(b *testing.B) {
	wg := New(context.Background())

	for b.Loop() {
		if err := wg.lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

func runBenchmarkLazyInitParallel(b *testing.B) {
	wg := New(context.Background())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := wg.lazyInit(); err != nil {
				b.Errorf("lazyInit returned error: %v", err)
				return
			}
		}
	})
}
