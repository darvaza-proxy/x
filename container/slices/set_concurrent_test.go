package slices_test

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/slices"
)

// TestCustomSet_ConcurrentSortInvariant verifies sorted order maintained under concurrent access
func TestCustomSet_ConcurrentSortInvariant(t *testing.T) {
	const (
		numWorkers   = 10
		opsPerWorker = 50
	)

	intCmp := func(a, b int) int { return a - b }
	set, err := slices.NewCustomSet(intCmp)
	core.AssertNoError(t, err, "create set")

	// Pre-populate with some data
	for i := range 100 {
		set.Add(i * 2) // Even numbers
	}

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers*2)

	// Concurrent Add operations
	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range opsPerWorker {
				value := workerID*1000 + i
				set.Add(value)

				// Verify sorted order is maintained
				values := set.Export()
				if !sort.IntsAreSorted(values) {
					errors <- fmt.Errorf("worker %d: slice not sorted after adding %d", workerID, value)
				}
			}
		}(w)
	}

	// Concurrent Delete operations
	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := range opsPerWorker / 2 {
				value := i * 2 // Delete some initial even numbers
				set.Remove(value)

				// Verify sorted order is maintained
				values := set.Export()
				if !sort.IntsAreSorted(values) {
					errors <- fmt.Errorf("worker %d: slice not sorted after deleting %d", workerID, value)
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

	// Final verification
	finalValues := set.Export()
	core.AssertTrue(t, sort.IntsAreSorted(finalValues), "final slice is sorted")

	// Verify no duplicates
	seen := make(map[int]bool)
	for _, v := range finalValues {
		if seen[v] {
			t.Errorf("duplicate value found: %d", v)
		}
		seen[v] = true
	}

	// Verify binary search (Contains) works correctly
	for _, v := range finalValues {
		core.AssertTrue(t, set.Contains(v), fmt.Sprintf("contains %d", v))
	}
}

// TestOrderedSet_ConcurrentOperations tests OrderedSet under concurrent access
func TestOrderedSet_ConcurrentOperations(t *testing.T) {
	set := slices.NewOrderedSet[int]()

	const numGoroutines = 20
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	var addCount, deleteCount atomic.Int32

	// Concurrent mixed operations
	for g := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(id)))

			for i := range opsPerGoroutine {
				value := r.Intn(1000)

				switch r.Intn(4) {
				case 0: // Add
					if set.Add(value) > 0 {
						addCount.Add(1)
					}
				case 1: // Delete
					if set.Remove(value) > 0 {
						deleteCount.Add(1)
					}
				case 2: // Contains
					_ = set.Contains(value)
				case 3: // Len
					_ = set.Len()
				}

				// Periodically verify sorted order
				if i%10 == 0 {
					values := set.Export()
					if !sort.IntsAreSorted(values) {
						panic(fmt.Sprintf("goroutine %d: slice not sorted at iteration %d", id, i))
					}
				}
			}
		}(g)
	}

	wg.Wait()

	// Final verification
	finalValues := set.Export()
	core.AssertTrue(t, sort.IntsAreSorted(finalValues), "final slice is sorted")

	// Check no duplicates
	for i := 1; i < len(finalValues); i++ {
		if finalValues[i] == finalValues[i-1] {
			t.Errorf("duplicate found: %d", finalValues[i])
		}
	}

	t.Logf("Operations: %d adds (net), %d deletes, final size: %d",
		addCount.Load(), deleteCount.Load(), set.Len())
}

// TestCustomSet_ConcurrentClone verifies cloning during concurrent modifications
func TestCustomSet_ConcurrentClone(t *testing.T) {
	strCmp := func(a, b string) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}

	original, err := slices.NewCustomSet(strCmp, "alpha", "beta", "gamma")
	core.AssertNoError(t, err, "create original")

	var wg sync.WaitGroup
	clones := make([]slices.Set[string], 5)

	// Concurrent cloning while modifying
	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Clone and modify
			clones[idx] = original.Clone()
			clones[idx].Add(fmt.Sprintf("clone_%d", idx))
		}(i)
	}

	// Modify original concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		original.Add("delta")
		original.Add("epsilon")
	}()

	wg.Wait()

	// Verify clones are independent
	for i, clone := range clones {
		cloneKey := fmt.Sprintf("clone_%d", i)
		core.AssertTrue(t, clone.Contains(cloneKey), fmt.Sprintf("clone %d has its key", i))

		// Other clones shouldn't have this key
		for j, other := range clones {
			if i != j {
				otherKey := fmt.Sprintf("clone_%d", i)
				core.AssertFalse(t, other.Contains(otherKey),
					fmt.Sprintf("clone %d shouldn't have clone %d's key", j, i))
			}
		}
	}
}

// TestCustomSet_StressTest performs intensive concurrent operations
func TestCustomSet_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	intCmp := func(a, b int) int { return a - b }
	set, err := slices.NewCustomSet(intCmp)
	core.AssertNoError(t, err, "create set")

	const numGoroutines = 50
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for g := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(id)))

			for i := range opsPerGoroutine {
				op := r.Intn(10)
				value := r.Intn(10000)

				switch op {
				case 0, 1, 2: // 30% Add
					set.Add(value)
				case 3, 4: // 20% Delete
					set.Remove(value)
				case 5, 6: // 20% Contains
					_ = set.Contains(value)
				case 7: // 10% Values
					values := set.Export()
					if !sort.IntsAreSorted(values) {
						errors <- fmt.Errorf("goroutine %d: not sorted at op %d", id, i)
						return
					}
				case 8: // 10% Len
					_ = set.Len()
				case 9: // 10% Clone
					clone := set.Clone()
					if clone == nil {
						errors <- fmt.Errorf("goroutine %d: clone returned nil", id)
						return
					}
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
		if errorCount > 10 {
			t.Fatal("too many errors, stopping")
		}
	}

	// Final validation
	finalValues := set.Export()
	core.AssertTrue(t, sort.IntsAreSorted(finalValues), "final values sorted")

	// No duplicates
	for i := 1; i < len(finalValues); i++ {
		if finalValues[i] == finalValues[i-1] {
			t.Errorf("duplicate in final state: %d", finalValues[i])
		}
	}

	t.Logf("Stress test completed: final set size = %d", set.Len())
}
