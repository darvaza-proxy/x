package web

import "net/http"

// MiddlewareFunc is a middleware handler that is introduced
// into an HTTP handling chain via [NewMiddleware].
type MiddlewareFunc func(rw http.ResponseWriter, req *http.Request, next http.Handler)

// MiddlewareWithErrorFunc is a middleware handler that is introduced
// into an HTTP handling chain via [NewMiddlewareError].
type MiddlewareWithErrorFunc func(rw http.ResponseWriter, req *http.Request, next http.Handler) error

// NewMiddleware adds a [MiddlewareFunc] in a HTTP handling chain.
func NewMiddleware(h MiddlewareFunc) func(http.Handler) http.Handler {
	if h == nil {
		return NoMiddleware
	}

	return func(next http.Handler) http.Handler {
		if next == nil {
			next = NewStatusNotFound()
		}

		fn := func(rw http.ResponseWriter, req *http.Request) {
			h(rw, req, next)
		}

		return http.HandlerFunc(fn)
	}
}

// NewMiddlewareWithError adds a [MiddlewareWithErrorFunc] in a HTTP handling chain.
func NewMiddlewareWithError(h MiddlewareWithErrorFunc) func(http.Handler) http.Handler {
	if h == nil {
		return NoMiddleware
	}

	return func(next http.Handler) http.Handler {
		if next == nil {
			next = NewStatusNotFound()
		}

		fn := func(rw http.ResponseWriter, req *http.Request) {
			if err := h(rw, req, next); err != nil {
				HandleError(rw, req, err)
			}
		}

		return http.HandlerFunc(fn)
	}
}

// NoMiddleware is a middleware handler that does nothing.
func NoMiddleware(next http.Handler) http.Handler {
	if next == nil {
		next = NewStatusNotFound()
	}
	return next
}
