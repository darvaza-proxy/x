package cond

import "darvaza.org/core"

// makeAnyMatch creates a function that returns true if any of the provided
// functions return true when applied to a value.
// The returned function evaluates each predicate in sequence and returns early
// on the first match.
//
//revive:disable-next-line:cognitive-complexity
func makeAnyMatch[T any](funcs []func(T) bool) func(T) bool {
	// remove nil entries
	funcs = core.SliceCopyFn(funcs, func(_ []func(T) bool, fn func(T) bool) (func(T) bool, bool) {
		return fn, fn != nil
	})

	switch len(funcs) {
	case 0:
		// unconstrained
		return func(_ T) bool { return true }
	case 1:
		// single case
		return funcs[0]
	default:
		// any of the given conditions
		return func(v T) bool {
			for _, fn := range funcs {
				if fn(v) {
					return true
				}
			}
			return false
		}
	}
}

// isCancelled checks if the abort channel has been closed, indicating
// cancellation. Returns true if the channel is closed, false otherwise.
func isCancelled(abort <-chan struct{}) bool {
	select {
	case <-abort:
		return true
	default:
		return false
	}
}
