package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
)

var (
	_ Error            = (*HTTPError)(nil)
	_ core.Unwrappable = (*HTTPError)(nil)
	_ http.Handler     = (*HTTPError)(nil)
)

// Error is an error that knows its HTTP Status Code
type Error interface {
	Error() string
	HTTPStatus() int
}

// HTTPError extends [core.WrappedError] with HTTP Status Code
type HTTPError struct {
	Err  error
	Code int
	Hdr  http.Header
}

// HTTPStatus returns the HTTP status code associated with the Error
func (err *HTTPError) HTTPStatus() int {
	switch {
	case err.Code == 0:
		return http.StatusOK
	case err.Code < 0:
		return http.StatusInternalServerError
	default:
		return err.Code
	}
}

// Header returns a [http.Header] attached to this error for custom fields
func (err *HTTPError) Header() http.Header {
	if err.Hdr == nil {
		err.Hdr = make(http.Header)
	}
	return err.Hdr
}

// AddHeader appends a value to an HTTP header entry of the HTTPError
func (err *HTTPError) AddHeader(key, value string) {
	err.Header().Add(key, value)
}

// SetHeader sets the value of a header key of the HTTPError
func (err *HTTPError) SetHeader(key, value string) {
	err.Header().Set(key, value)
}

// DeleteHeader removes a header key from the HTTPError if present
func (err *HTTPError) DeleteHeader(key string) {
	err.Header().Del(key)
}

// ServeHTTP is a very primitive handler that will try to pass the error
// to a [middleware.ErrorHandlerFunc] provided via the request's context.Context.
// if it exists, otherwise it will invoke the [Render] method to serve it.
func (err *HTTPError) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h, ok := ErrorHandler(req.Context()); ok {
		// pass over to the error handler
		h(rw, req, err)
		return
	}

	err.Render(rw, req)
}

// Render will serve the error ignoring the [ErrorHandler].
func (err *HTTPError) Render(rw http.ResponseWriter, req *http.Request) {
	code, hdr := err.prepareHeaders(rw)

	if req.Method == consts.HEAD || code < http.StatusBadRequest {
		// no content
		delete(hdr, consts.ContentType)

		rw.WriteHeader(code)
		return
	}

	// override media type
	hdr[consts.ContentType] = []string{consts.TXT}

	rw.WriteHeader(code)

	_, _ = fmt.Fprintln(rw, ErrorText(code))

	if err.Err != nil {
		err.renderPayload(rw)
	}
}

func (err *HTTPError) renderPayload(rw io.Writer) {
	msg := err.Err.Error()

	if msg != "" {
		_, _ = fmt.Fprint(rw, "\n", msg)
	}

	if e, ok := err.Err.(core.CallStacker); ok {
		stack := e.CallStack()
		if n := len(stack); n > 0 {
			if msg != "" {
				_, _ = fmt.Fprint(rw, "\n\n----\n")
			}

			err.renderStack(rw, stack)
		}
	}
}

func (*HTTPError) renderStack(rw io.Writer, stack core.Stack) {
	for i, f := range stack {
		pkgName, funcName := f.SplitName()
		fileName := path.Base(f.File())
		line := f.Line()

		_, _ = fmt.Fprintf(rw, "[%v/%v]: %s.%s\n\t%s/%s:%v\n", i, len(stack),
			pkgName, funcName,
			pkgName, fileName, line)
	}
}

func (err *HTTPError) prepareHeaders(rw http.ResponseWriter) (int, http.Header) {
	// HTTP Status Code
	code := err.HTTPStatus()
	if code == http.StatusOK {
		code = http.StatusNoContent
	}

	// Extend Headers
	hdr := rw.Header()
	for k, s := range err.Hdr {
		hdr[k] = append(hdr[k], s...)
	}

	// sanitize header
	delete(hdr, consts.ContentLength)
	delete(hdr, consts.ContentEncoding)

	return code, hdr
}

func (err *HTTPError) Error() string {
	var msg string

	text := ErrorText(err.HTTPStatus())
	if err.Err != nil {
		msg = err.Err.Error()
	}
	if msg == "" {
		return text
	}

	return fmt.Sprintf("%s: %s", text, msg)
}

func (err *HTTPError) Unwrap() error {
	return err.Err
}

// NewHTTPError creates a new HTTPError with a given StatusCode
// and optional cause and annotation
func NewHTTPError(code int, err error, note string) *HTTPError {
	switch {
	case err != nil:
		err = core.Wrap(err, note)
	case note != "":
		err = errors.New(note)
	}

	return &HTTPError{Err: err, Code: code}
}

// NewHTTPErrorf creates a new HTTPError with a given StatusCode
// and optional cause and formatted annotation
func NewHTTPErrorf(code int, err error, format string, args ...any) *HTTPError {
	switch {
	case err != nil:
		err = core.Wrapf(err, format, args...)
	default:
		err = fmt.Errorf(format, args...)
		if err.Error() == "" {
			err = nil
		}
	}

	return &HTTPError{Err: err, Code: code}
}

// ErrorText returns the title corresponding to
// a given HTTP Status code
func ErrorText(code int) string {
	text := http.StatusText(code)

	switch {
	case text == "":
		text = fmt.Sprintf("Unknown Error %d", code)
	case code >= 400:
		text = fmt.Sprintf("%s (Error %d)", text, code)
	}

	return text
}
