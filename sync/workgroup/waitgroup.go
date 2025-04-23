package workgroup

// WaitGroup manages a collection of goroutines and provides synchronisation
// primitives to coordinate their execution.
//
// It allows spawning new goroutines, waiting for their completion,
// and closing the group to prevent adding new tasks.
type WaitGroup interface {
	// IsClosed reports whether the WaitGroup has been closed.
	IsClosed() bool

	// Go spawns a new goroutine to execute the provided function.
	// Returns an error if the WaitGroup is closed or otherwise invalid.
	Go(func()) error

	// Count returns the number of active goroutines.
	Count() int

	// Wait blocks until all goroutines complete.
	Wait() error

	// Close prevents adding new goroutines and optionally waits for
	// existing ones to complete.
	Close() error
}
