package workgroup

//revive:disable:cognitive-complexity

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// Basic benchmarks comparing implementations of WaitGroup interface and sync.WaitGroup

func BenchmarkBasicUsage(b *testing.B) {
	sizes := []int{1, 10, 100, 1000}

	for _, size := range sizes {
		name := "Runner-" + strconv.Itoa(size)
		b.Run(name, func(b *testing.B) {
			benchWaitGroupBasic(b, NewRunner(), size)
		})

		name = "Limiter-" + strconv.Itoa(size)
		b.Run(name, func(b *testing.B) {
			limiter, err := NewLimiter(size)
			if err != nil {
				b.Fatalf("Failed to create limiter: %v", err)
			}
			benchWaitGroupBasic(b, limiter, size)
		})

		name = "StdWaitGroup-" + strconv.Itoa(size)
		b.Run(name, func(b *testing.B) {
			benchStdWaitGroupBasic(b, size)
		})
	}
}

func benchWaitGroupBasic(b *testing.B, wg WaitGroup, numGoroutines int) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < numGoroutines; j++ {
			err := wg.Go(func() {})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := wg.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
		}

		// Reset for reuse in benchmarks
		if wg.IsClosed() {
			// Create new instance if closed
			switch wg.(type) {
			case *Runner:
				wg = NewRunner()
			case *Limiter:
				wg, _ = NewLimiter(numGoroutines)
			}
		}
	}
}

func benchStdWaitGroupBasic(b *testing.B, numGoroutines int) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		wg.Add(numGoroutines)
		for j := 0; j < numGoroutines; j++ {
			go func() {
				defer wg.Done()
			}()
		}

		wg.Wait()
	}
}

// Benchmarks with simulated work in goroutines

func BenchmarkWithWork(b *testing.B) {
	sizes := []int{10, 100}
	durations := []time.Duration{10 * time.Microsecond, 100 * time.Microsecond}

	for _, size := range sizes {
		for _, duration := range durations {
			name := "Runner-" + strconv.Itoa(size) + "-" + duration.String()
			b.Run(name, func(b *testing.B) {
				benchWaitGroupWithWork(b, NewRunner(), size, duration)
			})

			name = "Limiter-" + strconv.Itoa(size) + "-" + duration.String()
			b.Run(name, func(b *testing.B) {
				limiter, err := NewLimiter(size)
				if err != nil {
					b.Fatalf("Failed to create limiter: %v", err)
				}
				benchWaitGroupWithWork(b, limiter, size, duration)
			})

			name = "StdWaitGroup-" + strconv.Itoa(size) + "-" + duration.String()
			b.Run(name, func(b *testing.B) {
				benchStdWaitGroupWithWork(b, size, duration)
			})
		}
	}
}

func benchWaitGroupWithWork(b *testing.B, wg WaitGroup, numGoroutines int, workDuration time.Duration) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < numGoroutines; j++ {
			err := wg.Go(func() {
				time.Sleep(workDuration)
			})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := wg.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
		}

		// Reset for reuse in benchmarks
		if wg.IsClosed() {
			// Create new instance if closed
			switch wg.(type) {
			case *Runner:
				wg = NewRunner()
			case *Limiter:
				wg, _ = NewLimiter(numGoroutines)
			}
		}
	}
}

func benchStdWaitGroupWithWork(b *testing.B, numGoroutines int, workDuration time.Duration) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		wg.Add(numGoroutines)
		for j := 0; j < numGoroutines; j++ {
			go func() {
				defer wg.Done()
				time.Sleep(workDuration)
			}()
		}

		wg.Wait()
	}
}

// Nested goroutines benchmarks

func BenchmarkNestedGoroutines(b *testing.B) {
	b.Run("Runner", func(b *testing.B) {
		benchWaitGroupNested(b, NewRunner())
	})

	b.Run("Limiter", func(b *testing.B) {
		limiter, err := NewLimiter(100) // Reasonable limit for nested goroutines
		if err != nil {
			b.Fatalf("Failed to create limiter: %v", err)
		}
		benchWaitGroupNested(b, limiter)
	})

	b.Run("StdWaitGroup", benchStdWaitGroupNested)
}

func benchWaitGroupNested(b *testing.B, wg WaitGroup) {
	for i := 0; i < b.N; i++ {
		// Spawn 10 parent goroutines, each spawning 10 children
		for j := 0; j < 10; j++ {
			err := wg.Go(func() {
				for k := 0; k < 10; k++ {
					_ = wg.Go(func() {
						time.Sleep(5 * time.Microsecond)
					})
				}
			})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := wg.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
		}

		// Reset for reuse in benchmarks
		if wg.IsClosed() {
			// Create new instance if closed
			switch wg.(type) {
			case *Runner:
				wg = NewRunner()
			case *Limiter:
				wg, _ = NewLimiter(100)
			}
		}
	}
}

func benchStdWaitGroupNested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		// We need to predict total count for sync.WaitGroup
		wg.Add(10) // parents
		for j := 0; j < 10; j++ {
			go func() {
				defer wg.Done()

				// Add children
				var childrenWG sync.WaitGroup
				childrenWG.Add(10)
				for k := 0; k < 10; k++ {
					go func() {
						defer childrenWG.Done()
						time.Sleep(5 * time.Microsecond)
					}()
				}
				childrenWG.Wait() // Parent waits for its children
			}()
		}

		wg.Wait()
	}
}

// Fan-out pattern benchmarks

func BenchmarkFanOut(b *testing.B) {
	b.Run("Runner", func(b *testing.B) {
		benchWaitGroupFanOut(b, NewRunner())
	})

	b.Run("Limiter", func(b *testing.B) {
		limiter, err := NewLimiter(5) // Same as workers count
		if err != nil {
			b.Fatalf("Failed to create limiter: %v", err)
		}
		benchWaitGroupFanOut(b, limiter)
	})

	b.Run("StdWaitGroup", benchStdWaitGroupFanOut)
}

func benchWaitGroupFanOut(b *testing.B, wg WaitGroup) {
	const (
		workers    = 5
		tasksPerWG = 100
	)

	for i := 0; i < b.N; i++ {
		// Fan out to workers
		for w := 0; w < workers; w++ {
			err := wg.Go(func() {
				// Each worker handles multiple tasks
				for t := 0; t < tasksPerWG/workers; t++ {
					// Simulate work
					time.Sleep(1 * time.Microsecond)
				}
			})
			if err != nil {
				b.Fatalf("Failed to add worker: %v", err)
			}
		}

		_ = wg.Wait()

		// Reset for reuse in benchmarks
		if wg.IsClosed() {
			// Create new instance if closed
			switch wg.(type) {
			case *Runner:
				wg = NewRunner()
			case *Limiter:
				wg, _ = NewLimiter(5)
			}
		}
	}
}

func benchStdWaitGroupFanOut(b *testing.B) {
	const (
		workers    = 5
		tasksPerWG = 100
	)

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		// Fan out to workers
		wg.Add(workers)
		for w := 0; w < workers; w++ {
			go func() {
				defer wg.Done()
				// Each worker handles multiple tasks
				for t := 0; t < tasksPerWG/workers; t++ {
					// Simulate work
					time.Sleep(1 * time.Microsecond)
				}
			}()
		}

		wg.Wait()
	}
}

// Dynamic workload benchmarks

func BenchmarkDynamicWorkload(b *testing.B) {
	b.Run("Runner", func(b *testing.B) {
		benchWaitGroupDynamic(b, NewRunner())
	})

	b.Run("Limiter", func(b *testing.B) {
		limiter, err := NewLimiter(10) // Same as initial tasks
		if err != nil {
			b.Fatalf("Failed to create limiter: %v", err)
		}
		benchWaitGroupDynamic(b, limiter)
	})

	b.Run("StdWaitGroup", benchStdWaitGroupDynamic)
}

func benchWaitGroupDynamic(b *testing.B, wg WaitGroup) {
	const initialTasks = 10

	for i := 0; i < b.N; i++ {
		// Add initial tasks
		for t := 0; t < initialTasks; t++ {
			taskID := t
			err := wg.Go(func() {
				// Each task may add more tasks depending on its ID
				if taskID < 5 {
					for j := 0; j < taskID; j++ {
						_ = wg.Go(func() {
							time.Sleep(1 * time.Microsecond)
						})
					}
				}
				time.Sleep(5 * time.Microsecond)
			})
			if err != nil {
				b.Fatalf("Failed to add task: %v", err)
			}
		}

		_ = wg.Wait()

		// Reset for reuse in benchmarks
		if wg.IsClosed() {
			// Create new instance if closed
			switch wg.(type) {
			case *Runner:
				wg = NewRunner()
			case *Limiter:
				wg, _ = NewLimiter(10)
			}
		}
	}
}

func benchStdWaitGroupDynamic(b *testing.B) {
	const initialTasks = 10

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		var mu sync.Mutex

		// Add initial tasks
		wg.Add(initialTasks)
		for t := 0; t < initialTasks; t++ {
			taskID := t
			go func() {
				defer wg.Done()

				// Each task may add more tasks depending on its ID
				if taskID < 5 {
					mu.Lock()
					additionalTasks := taskID
					wg.Add(additionalTasks)
					mu.Unlock()

					for j := 0; j < additionalTasks; j++ {
						go func() {
							defer wg.Done()
							time.Sleep(1 * time.Microsecond)
						}()
					}
				}
				time.Sleep(5 * time.Microsecond)
			}()
		}

		wg.Wait()
	}
}
