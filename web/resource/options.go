package resource

import (
	"io/fs"
	"strings"

	"darvaza.org/core"
)

// An OptionFunc configures the [Resource] during a [New]
// call. They might fail if used after the initialization.
type OptionFunc[T any] func(*Resource[T]) error

// WithChecker will force a specific [CheckerFunc] when initializing
// the [Resource]. If the argument is `nil` [DefaultChecker] will
// be used as if the `Check` function didn't exist.
func WithChecker[T any](fn CheckerFunc[T]) OptionFunc[T] {
	if fn == nil {
		fn = DefaultChecker
	}
	return func(r *Resource[T]) error {
		r.check = fn
		return nil
	}
}

// WithMethod sets a custom method handler during the [New] call.
func WithMethod[T any](method string, fn HandlerFunc[T]) OptionFunc[T] {
	s := strings.ToUpper(strings.TrimSpace(method))
	return func(r *Resource[T]) error {
		var err error
		switch {
		case len(r.methods) > 0:
			err = core.QuietWrap(fs.ErrPermission, "resource already initialized")
		case s == "":
			err = core.QuietWrap(core.ErrInvalid, "no method specified")
		case fn == nil:
			delete(r.h, s)
		default:
			r.h[s] = fn
		}
		return err
	}
}
