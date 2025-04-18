package workgroup

//revive:disable:cognitive-complexity

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// Basic benchmarks comparing Runner and sync.WaitGroup

func BenchmarkBasicUsage(b *testing.B) {
	sizes := []int{1, 10, 100, 1000}

	for _, size := range sizes {
		name := "Runner-" + strconv.Itoa(size)
		b.Run(name, func(b *testing.B) {
			benchRunnerBasic(b, size)
		})

		name = "StdWaitGroup-" + strconv.Itoa(size)
		b.Run(name, func(b *testing.B) {
			benchStdWaitGroupBasic(b, size)
		})
	}
}

func benchRunnerBasic(b *testing.B, numGoroutines int) {
	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		for j := 0; j < numGoroutines; j++ {
			err := runner.Go(func() {})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := runner.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
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
				benchRunnerWithWork(b, size, duration)
			})

			name = "StdWaitGroup-" + strconv.Itoa(size) + "-" + duration.String()
			b.Run(name, func(b *testing.B) {
				benchStdWaitGroupWithWork(b, size, duration)
			})
		}
	}
}

func benchRunnerWithWork(b *testing.B, numGoroutines int, workDuration time.Duration) {
	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		for j := 0; j < numGoroutines; j++ {
			err := runner.Go(func() {
				time.Sleep(workDuration)
			})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := runner.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
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
	b.Run("Runner", benchRunnerNested)
	b.Run("StdWaitGroup", benchStdWaitGroupNested)
}

func benchRunnerNested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		// Spawn 10 parent goroutines, each spawning 10 children
		for j := 0; j < 10; j++ {
			err := runner.Go(func() {
				for k := 0; k < 10; k++ {
					_ = runner.Go(func() {
						time.Sleep(5 * time.Microsecond)
					})
				}
			})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		if err := runner.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
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

// Close behaviour benchmark (Runner only)

func BenchmarkClose(b *testing.B) {
	b.Run("CloseWithPendingWork", benchRunnerClose)
}

func benchRunnerClose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		// Add 100 goroutines
		for j := 0; j < 100; j++ {
			err := runner.Go(func() {
				time.Sleep(5 * time.Microsecond)
			})
			if err != nil {
				b.Fatalf("Failed to add goroutine: %v", err)
			}
		}

		// Close and wait
		if err := runner.Close(); err != nil {
			b.Fatalf("Close failed: %v", err)
		}

		if err := runner.Wait(); err != nil {
			b.Fatalf("Wait failed: %v", err)
		}
	}
}

// Fan-out pattern benchmarks

func BenchmarkFanOut(b *testing.B) {
	b.Run("Runner", benchRunnerFanOut)
	b.Run("StdWaitGroup", benchStdWaitGroupFanOut)
}

func benchRunnerFanOut(b *testing.B) {
	const (
		workers    = 5
		tasksPerWG = 100
	)

	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		// Fan out to workers
		for w := 0; w < workers; w++ {
			err := runner.Go(func() {
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

		_ = runner.Wait()
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
	b.Run("Runner", benchRunnerDynamic)
	b.Run("StdWaitGroup", benchStdWaitGroupDynamic)
}

func benchRunnerDynamic(b *testing.B) {
	const initialTasks = 10

	for i := 0; i < b.N; i++ {
		runner := NewRunner()

		// Add initial tasks
		for t := 0; t < initialTasks; t++ {
			taskID := t
			err := runner.Go(func() {
				// Each task may add more tasks depending on its ID
				if taskID < 5 {
					for j := 0; j < taskID; j++ {
						_ = runner.Go(func() {
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

		_ = runner.Wait()
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
