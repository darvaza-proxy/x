package cond_test

import (
	"runtime"
	"sync"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/sync/cond"
	"darvaza.org/x/sync/errors"
)

// The once*Target types initialise themselves on first use through
// lazyInit, which checks the receiver.

// onceSyncTarget initialises through a sync.Once, whose Do returns no
// error.
type onceSyncTarget struct {
	ch   chan struct{}
	once sync.Once
}

func (o *onceSyncTarget) lazyInit() error {
	if o == nil {
		return errors.ErrNilReceiver
	}
	o.once.Do(o.init)
	return nil
}

func (o *onceSyncTarget) init() {
	o.ch = make(chan struct{}, 1)
}

// onceCondTarget initialises through a cond.Once.
type onceCondTarget struct {
	ch   chan struct{}
	once cond.Once
}

func (o *onceCondTarget) lazyInit() error {
	if o == nil {
		return errors.ErrNilReceiver
	}
	return o.once.Do(o.init)
}

func (o *onceCondTarget) init() error {
	o.ch = make(chan struct{}, 1)
	return nil
}

// BenchmarkOnce measures initialising on first use through cond.Once
// against sync.Once: lazyInit once initialised, alone and under
// RunParallel; the first call on a new value; and that first call raced
// by GOMAXPROCS goroutines.
func BenchmarkOnce(b *testing.B) {
	b.Run("sync", runBenchmarkOnceSync)
	b.Run("cond", runBenchmarkOnceCond)
}

func runBenchmarkOnceSync(b *testing.B) {
	newLazyInit := func() func() error { return new(onceSyncTarget).lazyInit }

	b.Run("done", runBenchmarkOnceSyncDone)
	b.Run("done_parallel", runBenchmarkOnceSyncDoneParallel)
	b.Run("first", func(b *testing.B) { runBenchmarkOnceFirst(b, newLazyInit) })
	b.Run("contended", func(b *testing.B) { runBenchmarkOnceContended(b, newLazyInit) })
}

func runBenchmarkOnceCond(b *testing.B) {
	newLazyInit := func() func() error { return new(onceCondTarget).lazyInit }

	b.Run("done", runBenchmarkOnceCondDone)
	b.Run("done_parallel", runBenchmarkOnceCondDoneParallel)
	b.Run("first", func(b *testing.B) { runBenchmarkOnceFirst(b, newLazyInit) })
	b.Run("contended", func(b *testing.B) { runBenchmarkOnceContended(b, newLazyInit) })
}

// The done benchmarks are written out per type, so lazyInit is called
// directly rather than through a function value.

func runBenchmarkOnceSyncDone(b *testing.B) {
	var o onceSyncTarget
	core.MustNoError(o.lazyInit())

	for b.Loop() {
		if err := o.lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

func runBenchmarkOnceSyncDoneParallel(b *testing.B) {
	var o onceSyncTarget
	core.MustNoError(o.lazyInit())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := o.lazyInit(); err != nil {
				b.Errorf("lazyInit returned error: %v", err)
				return
			}
		}
	})
}

func runBenchmarkOnceCondDone(b *testing.B) {
	var o onceCondTarget
	core.MustNoError(o.lazyInit())

	for b.Loop() {
		if err := o.lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

func runBenchmarkOnceCondDoneParallel(b *testing.B) {
	var o onceCondTarget
	core.MustNoError(o.lazyInit())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := o.lazyInit(); err != nil {
				b.Errorf("lazyInit returned error: %v", err)
				return
			}
		}
	})
}

// runBenchmarkOnceFirst makes the first call on a new value each
// iteration.
func runBenchmarkOnceFirst(b *testing.B, newLazyInit func() func() error) {
	for b.Loop() {
		lazyInit := newLazyInit()
		if err := lazyInit(); err != nil {
			b.Fatalf("lazyInit returned error: %v", err)
		}
	}
}

// runBenchmarkOnceContended releases GOMAXPROCS goroutines at once into
// the first call on a new value each iteration.
func runBenchmarkOnceContended(b *testing.B, newLazyInit func() func() error) {
	n := runtime.GOMAXPROCS(0)

	for b.Loop() {
		lazyInit := newLazyInit()
		start := make(chan struct{})

		var wg sync.WaitGroup
		for range n {
			wg.Go(func() {
				<-start
				core.MustNoError(lazyInit())
			})
		}
		close(start)
		wg.Wait()
	}
}
