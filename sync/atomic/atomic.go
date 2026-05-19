// Package atomic shadows the standard library sync/atomic package and adds
// helpers for patterns recurring in synchronisation primitives.
//
// The standard-library atomic types (Bool, Int32, Int64, Uint32, Uint64,
// Uintptr, Value, Pointer) are re-exported here as aliases (see std.go) so
// callers using the type-based API can reach both the stdlib types and the
// extension helpers through a single import of darvaza.org/x/sync/atomic.
//
// The legacy free-function API (AddInt32, LoadInt32, ...) is intentionally
// omitted; new code should use the type-based methods that supersede it
// since Go 1.19. Callers that still need those functions should keep their
// sync/atomic import alongside this one.
//
// All helpers are safe for concurrent use.
package atomic

// BitmaskOr atomically applies a bitwise OR of mask to *p, returning the
// resulting value and whether any bits actually changed.
//
// The boolean signals first-writer semantics for the supplied mask: only
// the goroutine whose OR contributed bits not already set observes
// changed == true. Callers that need to fire a "target reached" notification
// exactly once can gate on changed && result == target, ensuring exactly one
// writer sees the transition that completes the target mask.
//
// A zero mask can change no bits, so BitmaskOr returns (current, false).
func BitmaskOr(p *Uint32, mask uint32) (uint32, bool) {
	for {
		cur := p.Load()
		next := cur | mask
		if cur == next {
			return next, false
		}
		if p.CompareAndSwap(cur, next) {
			return next, true
		}
	}
}

// UpdateMax atomically raises *p to val when val is greater than the current
// value, leaving *p untouched otherwise. Returns the value the caller raised
// *p to, or the value observed when no update was performed; *p is guaranteed
// to be at least the returned value after the call, but may already be higher
// under contention.
func UpdateMax(p *Int32, val int32) int32 {
	for {
		cur := p.Load()
		if val <= cur {
			return cur
		}
		if p.CompareAndSwap(cur, val) {
			return val
		}
	}
}
