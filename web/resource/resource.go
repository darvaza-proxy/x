// Package resource implements a RESTful resource handler
package resource

import (
	"io/fs"
	"net/http"
	"strings"

	"darvaza.org/core"
	"darvaza.org/x/web"
	"darvaza.org/x/web/qlist"
)

var (
	_ http.Handler = (*Resource[any])(nil)
	_ web.Handler  = (*Resource[any])(nil)
)

// Resource is an http.Handler built around a given object
// and a data type.
type Resource[T any] struct {
	check    CheckerFunc[T]
	methods  []string
	identity string

	h  map[string]HandlerFunc[T]
	r  map[string]HandlerFunc[T]
	ql []qlist.QualityValue
}

// ServeHTTP handles the request initially using the TryServeHTTP method, and
// then calling [web.HandleError] if there is any problem.
func (r *Resource[T]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := r.TryServeHTTP(rw, req); err != nil {
		web.HandleError(rw, req, err)
	}
}

// TryServeHTTP attempts to handle the request, and returns an error in the case of
// problems instead of handling it locally.
// This method converts panics into errors in case they occur.
func (r *Resource[T]) TryServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	var c core.Catcher

	return c.Do(func() error {
		return r.doServeHTTP(rw, req)
	})
}

func (r *Resource[T]) doServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	// Path
	req, data, err := r.checkRequest(req)
	if err != nil {
		// 404 or 400
		return err
	}

	// Method
	h, method, err := r.getMethodHandler(req)
	if err != nil {
		return err
	}

	// store sanitized method and call method handler
	req.Method = method
	return h(rw, req, data)
}

func (r *Resource[T]) getMethodHandler(req *http.Request) (HandlerFunc[T], string, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		// Bad request
		err := r.wrap400(fs.ErrInvalid, "%q: invalid method", req.Method)
		return nil, "", err
	}

	// Handler
	if h := r.h[method]; h != nil {
		return h, method, nil
	}

	// Method not supported
	return nil, "", r.err405()
}

func (r *Resource[T]) checkRequest(req *http.Request) (*http.Request, *T, error) {
	req2, data, err := r.check(req)
	switch {
	case err != nil:
		return nil, nil, web.AsError(err)
	case req2 != nil:
		return req2, data, nil
	default:
		return req, data, nil
	}
}

// Methods returns a list of all supported HTTP Methods
func (r *Resource[T]) Methods() []string {
	l := len(r.methods)
	if l == 0 {
		// early call.
		return sortedKeys(r.h)
	}

	// return copy
	out := make([]string, l)
	copy(out, r.methods)
	return out
}

func (r *Resource[T]) serveOptions(rw http.ResponseWriter, _ *http.Request, _ *T) error {
	rw.Header()["Allow"] = r.methods
	rw.WriteHeader(http.StatusNoContent)
	return nil
}

// New creates a [Resource] using the provided handler object and the specified
// data type.
func New[T any](x any, options ...OptionFunc[T]) (*Resource[T], error) {
	h := newResource[T](x)

	for _, opt := range options {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	if err := h.setDefaults(); err != nil {
		return nil, err
	}

	return h, nil
}

func newResource[T any](x any) *Resource[T] {
	h := &Resource[T]{
		h: make(map[string]HandlerFunc[T]),
		r: make(map[string]HandlerFunc[T]),
	}

	if x != nil {
		// Checker
		if fn, ok := checkerOf[T](x); ok {
			h.check = fn
		}

		// Methods
		addHandlers[T](h, x)
	}

	return h
}

func (r *Resource[T]) setDefaults() error {
	// Check
	if r.check == nil {
		r.check = DefaultChecker
	}

	// HEAD
	if r.h[HEAD] == nil && r.h[GET] != nil {
		r.h[HEAD] = r.h[GET]
	}

	// OPTIONS
	if r.h[OPTIONS] == nil {
		r.h[OPTIONS] = r.serveOptions
	}

	r.methods = sortedKeys(r.h)
	return nil
}

// Must creates a [Resource] just like [New], but panics if there is an error.
func Must[T any](x any, options ...OptionFunc[T]) *Resource[T] {
	h, err := New[T](x, options...)
	if err != nil {
		panic(err)
	}
	return h
}
