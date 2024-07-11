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
	if fn, _ := ErrorHandler(req.Context()); fn != nil {
		// pass over to the error handler
		fn(rw, req, err)
		return
	}

	code := http.StatusInternalServerError

	h, ok := AsErrorWithCode(err, code).(http.Handler)
	if !ok {
		// naked 500 then
		h = &HTTPError{Code: code}
	}

	h.ServeHTTP(rw, req)
}

// AsError converts a given error into an HTTP-aware one.
// A nil argument will be treated as no error.
// If the status code can't be inferred, a 500 will be assumed.
func AsError(err error) error {
	return AsErrorWithCode(err, http.StatusInternalServerError)
}

// AsErrorWithCode converts a given error into an HTTP-aware one using
// the given status code when none could be inferred.
// If a non-positive code is provided, a 500 will be assumed.
func AsErrorWithCode(err error, code int) error {
	switch err.(type) {
	case nil:
		return nil
	case http.Handler:
		return err
	default:
		return &HTTPError{
			Code: getAsErrorCode(err, code),
			Err:  err,
			Hdr:  getAsErrorHeaders(err),
		}
	}
}

func getAsErrorCode(err error, code int) int {
	if e, ok := err.(Error); ok {
		if c := e.HTTPStatus(); c != 0 {
			code = c
		}
	}

	if code > 0 {
		return code
	}

	return http.StatusInternalServerError
}

func getAsErrorHeaders(err error) http.Header {
	switch v := err.(type) {
	case interface {
		Header() http.Header
	}:
		return v.Header().Clone()
	case interface {
		Headers() http.Header
	}:
		return v.Headers().Clone()
	default:
		return nil
	}
}
