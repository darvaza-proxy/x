package cond

import "darvaza.org/core"

// sanitiseFuncs returns a copy of funcs with nil entries removed, preserving
// the order of the surviving predicates.
func sanitiseFuncs[T any](funcs []func(T) bool) []func(T) bool {
	return core.SliceCopyFn(funcs,
		func(_ []func(T) bool, fn func(T) bool) (func(T) bool, bool) {
			return fn, fn != nil
		})
}

// makeAnyOf returns a predicate that evaluates funcs in order and returns
// true on the first match. Caller must ensure funcs has at least one entry
// and no nil entries; use [makeAnyMatch] for the validating dispatcher.
func makeAnyOf[T any](funcs []func(T) bool) func(T) bool {
	return func(v T) bool {
		for _, fn := range funcs {
			if fn(v) {
				return true
			}
		}
		return false
	}
}

// makeAnyMatch returns a predicate that evaluates funcs in order and stops
// at the first match. Nil entries are stripped first to avoid invoking them.
// If no non-nil entries remain (empty input or all-nil), the returned
// predicate always returns true — the "unconstrained" case Count relies on
// when constructed without broadcast filters.
func makeAnyMatch[T any](funcs []func(T) bool) func(T) bool {
	funcs = sanitiseFuncs(funcs)

	switch len(funcs) {
	case 0:
		return func(_ T) bool { return true }
	case 1:
		return funcs[0]
	default:
		return makeAnyOf(funcs)
	}
}

// isCancelled reports whether a receive from abort would proceed without
// blocking — the channel is closed or has a pending value. Nil and
// open-empty channels report false. cond uses close-only signals at every
// call site, but the helper itself does not assume that.
func isCancelled(abort <-chan struct{}) bool {
	select {
	case <-abort:
		return true
	default:
		return false
	}
}
