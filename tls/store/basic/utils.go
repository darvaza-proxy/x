package basic

func join[T any](vv [][]T) []T {
	var count int
	for _, v := range vv {
		count += len(v)
	}

	out := make([]T, 0, count)
	for _, v := range vv {
		out = append(out, v...)
	}
	return out
}
