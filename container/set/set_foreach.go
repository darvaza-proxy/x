package set

// Values returns all values in the [Set].
func (set *Set[K, H, T]) Values() []T {
	if set == nil {
		return nil
	}

	// RO
	set.mu.RLock()
	if !set.unsafeIsReady() {
		set.mu.RUnlock()
		return nil
	}

	vv, count := set.unsafeValues()
	set.mu.RUnlock()

	return set.doValues(vv, count)
}

func (*Set[K, H, T]) doValues(vv [][]T, count int) []T {
	// join
	out := make([]T, 0, count)
	for _, v := range vv {
		out = append(out, v...)
	}
	return out
}

// ForEach calls a function for each value in the [Set] until it returns
// false.
func (set *Set[K, H, T]) ForEach(fn func(T) bool) {
	if set == nil || fn == nil {
		return
	}

	// RO
	set.mu.RLock()
	if !set.unsafeIsReady() {
		set.mu.RUnlock()
		return
	}

	vv, _ := set.unsafeValues()
	set.mu.RUnlock()

	// iter
	set.doForEach(vv, fn)
}

func (*Set[K, H, T]) doForEach(vv [][]T, fn func(T) bool) {
	for _, v := range vv {
		for _, x := range v {
			if !fn(x) {
				return
			}
		}
	}
}

func (set *Set[K, H, T]) unsafeValues() (vv [][]T, count int) {
	n := len(set.buckets)
	if n == 0 {
		return nil, 0
	}

	vv = make([][]T, 0, n)
	for _, l := range set.buckets {
		v := l.Values()
		if n = len(v); n > 0 {
			count += n
			vv = append(vv, v)
		}
	}

	return vv, count
}
