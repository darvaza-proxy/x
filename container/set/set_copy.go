package set

import "darvaza.org/x/container/list"

// Clone creates a new Set containing the same values.
func (set *Set[K, H, T]) Clone() *Set[K, H, T] {
	if set == nil {
		return nil
	}

	// RO
	set.mu.RLock()
	defer set.mu.RUnlock()

	if !set.unsafeIsReady() {
		return nil
	}

	return set.unsafeClone(nil, nil)
}

func (set *Set[K, H, T]) unsafeClone(dst *Set[K, H, T], cond func(T) bool) *Set[K, H, T] {
	if dst == nil {
		dst = new(Set[K, H, T])
	}

	dst.cfg = set.cfg
	dst.buckets = make(map[H]*list.List[T], len(set.buckets))

	fn := func(v T) (T, bool) {
		return v, cond == nil || cond(v)
	}

	for h, l := range set.buckets {
		dst.buckets[h] = l.Copy(fn)
	}

	return dst
}

// Copy copies the values in the Set to another, optionally
// using a condition function.
//
// If no destination is provided, one will be created.
// If an uninitialized destination is provided, it will be initialized using
// the source's [Config] and values copied in bulk.
func (set *Set[K, H, T]) Copy(dst *Set[K, H, T], cond func(T) bool) *Set[K, H, T] {
	if set == nil || set == dst {
		// no source, or source and destination are the same.
		// nothing to do.
		return dst
	}

	// RO
	set.mu.RLock()
	defer set.mu.RUnlock()

	switch {
	case !set.unsafeIsReady():
		// uninitialized source, nothing to do.
		return dst
	case dst == nil:
		// destination not provided, create new.
		return set.unsafeClone(nil, cond)
	default:
		// RW destination
		dst.mu.Lock()

		switch {
		case !dst.unsafeIsReady():
			// uninitialized destination. externally allocated.
			// keep lock but treat as own.
			defer dst.mu.Unlock()

			return set.unsafeClone(dst, cond)
		default:
			// no optimizations. release lock and Push.
			dst.mu.Unlock()

			return set.unsafeCopy(dst, cond)
		}
	}
}

func (set *Set[K, H, T]) unsafeCopy(dst *Set[K, H, T], cond func(T) bool) *Set[K, H, T] {
	for _, l := range set.buckets {
		l.ForEach(func(v T) bool {
			if cond == nil || cond(v) {
				_, _ = dst.Push(v)
			}
			return true
		})
	}
	return dst
}
