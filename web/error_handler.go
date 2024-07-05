package web

import (
	"context"
	"net/http"

	"darvaza.org/core"
)

// ErrorHandlerFunc is the signature of a function used as ErrorHandler
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

// WithErrorHandler attaches an ErrorHandler function to a context
// for later retrieval
func WithErrorHandler(ctx context.Context, h ErrorHandlerFunc) context.Context {
	return errCtxKey.WithValue(ctx, h)
}

// ErrorHandler attempts to pull an ErrorHandler from the context.Context
func ErrorHandler(ctx context.Context) (ErrorHandlerFunc, bool) {
	return errCtxKey.Get(ctx)
}

var (
	errCtxKey = core.NewContextKey[ErrorHandlerFunc]("ErrorHandler")
)

// HandleError attempts to call the [ErrorHandler] set for the content,
// otherwise it serves the error directly using its own ServeHTTP if defined
// or [HTTPError] if not.
func HandleError(rw http.ResponseWriter, req *http.Request, err error) {
	var h http.Handler

	if fn, _ := ErrorHandler(req.Context()); fn != nil {
		// pass over to the error handler
		fn(rw, req, err)
		return
	}

	if e, ok := err.(http.Handler); ok {
		// direct
		h = e
	} else {
		// assemble
		var code int

		if e, ok := err.(Error); ok {
			code = e.HTTPStatus()
		}

		h = &HTTPError{
			Code: core.IIf(code > 0, code, http.StatusInternalServerError),
			Err:  err,
		}
	}

	h.ServeHTTP(rw, req)
}
