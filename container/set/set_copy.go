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

	return set.unsafeClone()
}

func (set *Set[K, H, T]) unsafeClone() *Set[K, H, T] {
	dst := new(Set[K, H, T])
	dst.cfg = set.cfg
	dst.buckets = make(map[H]*list.List[T], len(set.buckets))

	for h, l := range set.buckets {
		dst.buckets[h] = l.Clone()
	}

	return dst
}
