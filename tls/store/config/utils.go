package config

func merge[T any](ss ...[]T) []T {
	switch len(ss) {
	case 0:
		// empty
		return []T{}
	case 1:
		// as-is
		return ss[0]
	default:
		// count
		n := 0
		for _, s := range ss {
			n += len(s)
		}
		// and combine
		out := make([]T, 0, n)
		for _, s := range ss {
			out = append(out, s...)
		}
		return out
	}
}

func insert[T any](before, after []T, middle ...T) []T {
	return merge(before, middle, after)
}
