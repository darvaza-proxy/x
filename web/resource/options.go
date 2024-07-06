package resource

import (
	"fmt"
	"io/fs"
	"strings"

	"darvaza.org/core"
	"darvaza.org/x/web/qlist"
)

var errBusy = core.QuietWrap(fs.ErrPermission, "resource already initialized")

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
			err = errBusy
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

// WithRenderer provides a custom renderer for the specified media type
func WithRenderer[T any](mediaType string, fn HandlerFunc[T]) OptionFunc[T] {
	s := strings.ToLower(strings.TrimSpace(mediaType))
	return func(r *Resource[T]) error {
		if len(r.methods) > 0 {
			return errBusy
		}
		return r.addRenderer(s, fn)
	}
}

func (r *Resource[T]) addRenderer(mediaType string, fn HandlerFunc[T]) error {
	qv, err := qlist.ParseQualityValue(mediaType)
	switch {
	case err != nil, !qv.IsMediaType():
		return fmt.Errorf("%q: invalid media type", mediaType)
	case fn == nil:
		return fmt.Errorf("%q: no renderer specified", mediaType)
	}

	name := qv.String()
	if prev := r.getRenderer(name); prev == nil {
		// new
		r.ql = append(r.ql, qv)
	}

	// set or replace
	r.r[name] = fn
	return nil
}

// WithIdentity specifies the media type to use when nothing is acceptable for the client
func WithIdentity[T any](mediaType string) OptionFunc[T] {
	s := strings.ToLower(strings.TrimSpace(mediaType))
	return func(r *Resource[T]) error {
		if len(r.methods) > 0 {
			return errBusy
		}
		return r.setIdentity(s)
	}
}

func (r *Resource[T]) setIdentity(mediaType string) error {
	if fn := r.getRenderer(mediaType); fn == nil {
		return fmt.Errorf("%q: no renderer specified", mediaType)
	}

	// set
	r.identity = mediaType
	return nil
}
