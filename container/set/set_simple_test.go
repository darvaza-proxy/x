package set_test

import (
	"fmt"
	"sync"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func doTestPush(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	_, err := s.Push(testItem{ID: 1, Name: "test"})
	core.AssertMustNoError(t, err, "push")
}

func doTestContains(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	core.AssertTrue(t, s.Contains(1), "contains")
}

func doTestGet(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	v, err := s.Get(1)
	core.AssertNoError(t, err, "get")
	core.AssertEqual(t, 1, v.ID, "id")
}

func doTestPop(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	_, err := s.Pop(1)
	core.AssertNoError(t, err, "pop")
}

func doTestVerifyRemoved(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	core.AssertFalse(t, s.Contains(1), "removed")
}

// TestSetBasicOperations exercises the push/contains/get/pop lifecycle.
func TestSetBasicOperations(t *testing.T) {
	s := testConfig().Must()

	doTestPush(t, s)
	doTestContains(t, s)
	doTestGet(t, s)
	doTestPop(t, s)
	doTestVerifyRemoved(t, s)
}

func runConcurrentAdds(s *set.Set[int, int, testItem], wg *sync.WaitGroup, base, itemsPerGoroutine int) {
	defer wg.Done()
	for j := range itemsPerGoroutine {
		id := base*itemsPerGoroutine + j
		_, _ = s.Push(testItem{ID: id, Name: "test", Value: "value"})
	}
}

func runConcurrentRemoves(s *set.Set[int, int, testItem], wg *sync.WaitGroup, base, itemsPerGoroutine int) {
	defer wg.Done()
	for j := range itemsPerGoroutine {
		id := base*itemsPerGoroutine + j
		_ = s.Contains(id)
		_, _ = s.Get(id)
		_, _ = s.Pop(id)
	}
}

// TestSetConcurrency tests thread safety of set operations.
func TestSetConcurrency(t *testing.T) {
	s := testConfig().Must()

	const numGoroutines = 10
	const itemsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // half adding, half removing

	for i := range numGoroutines {
		go runConcurrentAdds(s, &wg, i, itemsPerGoroutine)
	}
	for i := range numGoroutines {
		go runConcurrentRemoves(s, &wg, i, itemsPerGoroutine)
	}

	wg.Wait()

	// the set must remain iterable without panicking.
	core.AssertNoPanic(t, func() { s.ForEach(func(testItem) bool { return true }) }, "iterable")
}

func makeFilledSet(b *testing.B, size int) *set.Set[int, int, testItem] {
	s := testConfig().Must()
	for i := range size {
		if _, err := s.Push(testItem{ID: i}); err != nil {
			b.Fatalf("Failed to populate set: %v", err)
		}
	}
	return s
}

func doBenchmarkSetPush(b *testing.B, size int) {
	s := makeFilledSet(b, size)
	b.ResetTimer()
	for i := range b.N {
		_, err := s.Push(testItem{ID: size + i})
		if err != nil && err != set.ErrExist {
			b.Fatalf("Push failed: %v", err)
		}
		_, _ = s.Pop(size + i) // clean up to maintain size
	}
}

func doBenchmarkSetPop(b *testing.B, size int) {
	s := makeFilledSet(b, size)
	b.ResetTimer()
	for i := range b.N {
		id := i % size
		_, _ = s.Pop(id)
		if _, err := s.Push(testItem{ID: id}); err != nil { // re-add to maintain set
			b.Fatalf("Re-add failed: %v", err)
		}
	}
}

func doBenchmarkSetContains(b *testing.B, size int) {
	s := makeFilledSet(b, size)
	b.ResetTimer()
	for i := range b.N {
		s.Contains(i % size)
	}
}

// Benchmarks

func BenchmarkSetPush(b *testing.B) {
	for _, size := range core.S(10, 1000, 10000) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetPush(b, size)
		})
	}
}

func BenchmarkSetPop(b *testing.B) {
	for _, size := range core.S(10, 1000, 10000) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetPop(b, size)
		})
	}
}

func BenchmarkSetContains(b *testing.B) {
	for _, size := range core.S(10, 1000, 10000) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetContains(b, size)
		})
	}
}
