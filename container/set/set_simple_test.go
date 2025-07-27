package set

import (
	"fmt"
	"sync"
	"testing"
)

func doTestPush(t *testing.T, s *Set[int, int, testItem]) {
	item := testItem{ID: 1, Name: "test"}
	if _, err := s.Push(item); err != nil {
		t.Fatalf("Push failed: %v", err)
	}
}

func doTestContains(t *testing.T, s *Set[int, int, testItem]) {
	if !s.Contains(1) {
		t.Error("Contains should return true")
	}
}

func doTestGet(t *testing.T, s *Set[int, int, testItem]) {
	if v, err := s.Get(1); err != nil {
		t.Errorf("Get failed: %v", err)
	} else if v.ID != 1 {
		t.Errorf("Get returned wrong item: %+v", v)
	}
}

func doTestPop(t *testing.T, s *Set[int, int, testItem]) {
	if _, err := s.Pop(1); err != nil {
		t.Errorf("Pop failed: %v", err)
	}
}

func doTestVerifyRemoved(t *testing.T, s *Set[int, int, testItem]) {
	if s.Contains(1) {
		t.Error("Item should be removed")
	}
}

// Simple test to ensure basic functionality works
func TestSetBasicOperations(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	doTestPush(t, s)
	doTestContains(t, s)
	doTestGet(t, s)
	doTestPop(t, s)
	doTestVerifyRemoved(t, s)
}

func runConcurrentAdds(s *Set[int, int, testItem], wg *sync.WaitGroup, base, itemsPerGoroutine int) {
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

func runConcurrentRemoves(s *Set[int, int, testItem], wg *sync.WaitGroup, base, itemsPerGoroutine int) {
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
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

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
	t.Logf("Set contains %d items after concurrent operations", count)
}

func makeFilledSet(b *testing.B, size int) *Set[int, int, testItem] {
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
		if err != nil && err != ErrExist {
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
