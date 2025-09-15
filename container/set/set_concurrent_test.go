package set_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
)

// TestSet_ConcurrentDataIntegrity verifies no data corruption under concurrent access
func TestSet_ConcurrentDataIntegrity(t *testing.T) {
	const (
		numWorkers     = 10
		itemsPerWorker = 100
	)

	cfg := set.Config[int, int, int]{
		ItemKey: func(v int) (int, error) { return v, nil },
		Hash:    func(k int) (int, error) { return k % 10, nil }, // Create hash collisions
	}

	s := &set.Set[int, int, int]{}
	err := cfg.Init(s)
	core.AssertNoError(t, err, "init")

	var wg sync.WaitGroup
	var pushCount, popCount atomic.Int32
	errors := make(chan error, numWorkers*3)

	// Concurrent Push operations
	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range itemsPerWorker {
				value := workerID*itemsPerWorker + i
				_, err := s.Push(value)
				if err == nil {
					pushCount.Add(1)
				} else if err != set.ErrExist {
					errors <- fmt.Errorf("worker %d push %d: %w", workerID, value, err)
				}
			}
		}(w)
	}

	// Concurrent Get operations
	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range itemsPerWorker {
				value := workerID*itemsPerWorker + i
				if _, err := s.Get(value); err != nil && err != set.ErrNotExist {
					errors <- fmt.Errorf("worker %d get %d: %w", workerID, value, err)
				}
			}
		}(w)
	}

	// Concurrent Pop operations
	for w := range numWorkers / 2 {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range itemsPerWorker / 2 {
				value := workerID*itemsPerWorker + i
				_, err := s.Pop(value)
				if err == nil {
					popCount.Add(1)
				} else if err != set.ErrNotExist {
					errors <- fmt.Errorf("worker %d pop %d: %w", workerID, value, err)
				}
			}
		}(w)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify final state consistency
	finalLen := 0
	s.ForEach(func(_ int) bool {
		finalLen++
		return true
	})
	expectedRemaining := int(pushCount.Load() - popCount.Load())
	core.AssertEqual(t, expectedRemaining, finalLen, "final set length")

	// Verify all remaining items are retrievable
	count := 0
	var forEachErr error
	s.ForEach(func(v int) bool {
		count++
		// Each item should be retrievable
		retrieved, err := s.Get(v)
		if err != nil {
			forEachErr = fmt.Errorf("cannot retrieve item %d during ForEach: %w", v, err)
			return false
		}
		if retrieved != v {
			forEachErr = fmt.Errorf("retrieved value %d != expected %d", retrieved, v)
			return false
		}
		return true
	})
	core.AssertNoError(t, forEachErr, "ForEach verification")
	core.AssertEqual(t, finalLen, count, "ForEach count matches Len")
}

// TestSet_ConcurrentForEach verifies ForEach is safe during modifications
func TestSet_ConcurrentForEach(t *testing.T) {
	cfg := set.Config[string, string, string]{
		ItemKey: func(v string) (string, error) { return v, nil },
		Hash: func(k string) (string, error) {
			if len(k) > 0 {
				return k[:1], nil // Group by first letter
			}
			return "", nil
		},
	}

	s := &set.Set[string, string, string]{}
	err := cfg.Init(s, "apple", "banana", "cherry", "date", "elderberry")
	core.AssertNoError(t, err, "init")

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent ForEach operations
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			visitedCount := 0
			var forEachErr error
			s.ForEach(func(v string) bool {
				visitedCount++
				if v == "" {
					forEachErr = fmt.Errorf("forEach %d: empty value", id)
					return false
				}
				return true
			})
			if forEachErr != nil {
				errors <- forEachErr
			}
			// Each ForEach should see a consistent snapshot
			if visitedCount == 0 {
				errors <- fmt.Errorf("forEach %d visited no items", id)
			}
		}(i)
	}

	// Concurrent modifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		fruits := []string{"fig", "grape", "honeydew", "kiwi", "lemon"}
		for _, fruit := range fruits {
			_, _ = s.Push(fruit)
		}
	}()

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestSet_ConcurrentClone verifies Clone produces independent copy during concurrent ops
func TestSet_ConcurrentClone(t *testing.T) {
	cfg := set.Config[int, int, int]{
		ItemKey: func(v int) (int, error) { return v, nil },
		Hash:    func(k int) (int, error) { return k, nil },
	}

	original := &set.Set[int, int, int]{}
	err := cfg.Init(original)
	core.AssertNoError(t, err, "init original")

	// Add initial data
	for i := range 100 {
		_, _ = original.Push(i)
	}

	var wg sync.WaitGroup
	clones := make([]*set.Set[int, int, int], 5)

	// Concurrent cloning while modifying original
	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clones[idx] = original.Clone()
			// Modify clone independently
			for j := range 10 {
				_, _ = clones[idx].Push(idx*1000 + 100 + j)
			}
		}(i)
	}

	// Modify original concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 10 {
			_, _ = original.Push(200 + i)
		}
	}()

	wg.Wait()

	// Verify clones are independent
	for i, clone := range clones {
		// Each clone should have its unique additions
		for j := range 10 {
			key := i*1000 + 100 + j
			_, err := clone.Get(key)
			core.AssertNoError(t, err, fmt.Sprintf("clone %d should have %d", i, key))

			// Other clones should not have this key
			for k, otherClone := range clones {
				if k != i {
					_, err := otherClone.Get(key)
					core.AssertError(t, err, fmt.Sprintf("clone %d should not have clone %d's key %d", k, i, key))
				}
			}
		}
	}
}

// TestSet_BucketConsistencyAfterOperations verifies hash bucket distribution remains correct
func TestSet_BucketConsistencyAfterOperations(t *testing.T) {
	// Use a hash function that creates known collisions
	cfg := set.Config[int, int, int]{
		ItemKey: func(v int) (int, error) { return v, nil },
		Hash:    func(k int) (int, error) { return k % 3, nil }, // Only 3 buckets
	}

	s := &set.Set[int, int, int]{}
	err := cfg.Init(s)
	core.AssertNoError(t, err, "init")

	// Add items that will collide in buckets
	itemsAdded := make(map[int]bool)
	for i := range 30 {
		_, err := s.Push(i)
		core.AssertNoError(t, err, fmt.Sprintf("push %d", i))
		itemsAdded[i] = true
	}

	// Remove some items from each bucket
	for i := 5; i < 20; i += 3 {
		_, err := s.Pop(i)
		core.AssertNoError(t, err, fmt.Sprintf("pop %d", i))
		delete(itemsAdded, i)
	}

	// Verify all remaining items are still retrievable
	for k := range itemsAdded {
		retrieved, err := s.Get(k)
		core.AssertNoError(t, err, fmt.Sprintf("get %d after removals", k))
		core.AssertEqual(t, k, retrieved, fmt.Sprintf("retrieved value for key %d", k))
	}

	// Verify ForEach visits each item exactly once
	visited := make(map[int]int)
	s.ForEach(func(v int) bool {
		visited[v]++
		return true
	})

	// Check each item visited exactly once
	for k := range itemsAdded {
		count, found := visited[k]
		core.AssertTrue(t, found, fmt.Sprintf("item %d visited", k))
		core.AssertEqual(t, 1, count, fmt.Sprintf("item %d visit count", k))
	}

	// Verify no extra items
	core.AssertEqual(t, len(itemsAdded), len(visited), "visited count matches expected")
}
