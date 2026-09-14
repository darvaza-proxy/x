package basic

import "context"

func doForEach[T any](ctx context.Context, fn func(context.Context, T) bool, items []T) {
	for _, item := range items {
		if ctx.Err() != nil {
			// cancelled
			break
		}

		if !fn(ctx, item) {
			break
		}
	}
}

func doForEach2[K comparable, T any](ctx context.Context, fn func(context.Context, K, T) bool, dict map[K]T) {
	for k, v := range dict {
		if ctx.Err() != nil {
			// cancelled
			break
		}

		if !fn(ctx, k, v) {
			break
		}
	}
}
