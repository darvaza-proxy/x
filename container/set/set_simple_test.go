package set_test

import (
	"fmt"
	"sync"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func runTestPush(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	item := testItem{ID: 1, Name: "test"}
	_, err := s.Push(item)
	core.AssertMustNoError(t, err, "Push")
}

func runTestContains(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	core.AssertTrue(t, s.Contains(1), "Contains(1)")
}

func runTestGet(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	v, err := s.Get(1)
	core.AssertMustNoError(t, err, "Get")
	core.AssertEqual(t, 1, v.ID, "Get item ID")
}

func runTestPop(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	_, err := s.Pop(1)
	core.AssertMustNoError(t, err, "Pop")
}

func runTestVerifyRemoved(t *testing.T, s *set.Set[int, int, testItem]) {
	t.Helper()
	core.AssertFalse(t, s.Contains(1), "item should be removed")
}

// Simple test to ensure basic functionality works
func TestSetBasicOperations(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	runTestPush(t, s)
	runTestContains(t, s)
	runTestGet(t, s)
	runTestPop(t, s)
	runTestVerifyRemoved(t, s)
}

func runConcurrentAdds(s *set.Set[int, int, testItem], wg *sync.WaitGroup, base, itemsPerGoroutine int) {
	defer wg.Done()
	for j := range itemsPerGoroutine {
		id := base*itemsPerGoroutine + j
		item := testItem{
			ID:    id,
			Name:  "test",
			Value: "value",
		}
		_, _ = s.Push(item)
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

// TestSetConcurrency tests thread safety of set operations
func TestSetConcurrency(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	const numGoroutines = 10
	const itemsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Half adding, half removing

	// Start goroutines that add items
	for i := range numGoroutines {
		go runConcurrentAdds(s, &wg, i, itemsPerGoroutine)
	}

	// Start goroutines that remove items
	for i := range numGoroutines {
		go runConcurrentRemoves(s, &wg, i, itemsPerGoroutine)
	}

	wg.Wait()

	// Verify the set is in a consistent state
	// Just check that we can iterate without panic
	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return false
	})
	// Set operations completed successfully
}

func makeFilledSet(b *testing.B, size int) *set.Set[int, int, testItem] {
	cfg := testConfig()
	s, _ := cfg.New()
	for i := range size {
		_, err := s.Push(testItem{ID: i})
		if err != nil {
			b.Fatalf("Failed to populate set: %v", err)
		}
	}
	return s
}

func doBenchmarkSetPush(b *testing.B, size int) {
	s := makeFilledSet(b, size)
	b.ResetTimer()
	for i := range b.N {
		item := testItem{ID: size + i}
		_, err := s.Push(item)
		if err != nil && err != set.ErrExist {
			b.Fatalf("Push failed: %v", err)
		}
		_, _ = s.Pop(size + i) // Clean up to maintain size
	}
}

func doBenchmarkSetPop(b *testing.B, size int) {
	s := makeFilledSet(b, size)
	b.ResetTimer()
	for i := range b.N {
		id := i % size
		_, _ = s.Pop(id)
		_, err := s.Push(testItem{ID: id}) // Re-add to maintain set
		if err != nil {
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
	sizes := []int{10, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetPush(b, size)
		})
	}
}

func BenchmarkSetPop(b *testing.B) {
	sizes := []int{10, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetPop(b, size)
		})
	}
}

func BenchmarkSetContains(b *testing.B) {
	sizes := []int{10, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			doBenchmarkSetContains(b, size)
		})
	}
}
