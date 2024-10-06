package certpool

import (
	"context"
)

func copyMap[K comparable, V any](m map[K]V, fn func(V) (V, bool)) map[K]V {
	if fn == nil {
		fn = copyFn
	}

	out := make(map[K]V, len(m))
	for k, v := range m {
		if v, ok := fn(v); ok {
			out[k] = v
		}
	}
	return out
}

func copyMapList[K comparable, V any](m map[K]*List[V], fn func(V) (V, bool)) map[K]*List[V] {
	if fn == nil {
		fn = copyFn
	}

	out := make(map[K]*List[V], len(m))
	for k, l1 := range m {
		if l2 := l1.Copy(fn); l2.Len() > 0 {
			out[k] = l2
		}
	}
	return out
}

func copyFn[V any](v V) (V, bool) {
	return v, true
}

func sliceForEach[T any](ctx context.Context, fn func(context.Context, T) bool, values []T) {
	for _, v := range values {
		switch {
		case ctx.Err() != nil:
			// cancelled
			return
		case !fn(ctx, v):
			// aborted
			return
		}
	}
}

func appendMapList[K, V comparable](m map[K]*List[V], key K, value V) {
	l, ok := m[key]
	if !ok {
		l = new(List[V])
		m[key] = l
	}
	l.PushBack(value)
}

func deleteMapListMatchFn[K, V comparable](m map[K]*List[V], keys []K, eq func(v V) bool) {
	for _, key := range keys {
		if l, ok := m[key]; ok {
			l.DeleteMatchFn(eq)
		}
	}
}
