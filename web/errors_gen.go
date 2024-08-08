package web

// Code generated by ./errors_gen.sh; DO NOT EDIT

//go:generate ./errors_gen.sh

import (
	"fmt"
	"net/http"
	"strings"

	"darvaza.org/x/fs"
	"darvaza.org/x/web/consts"
)

// NewStatusMovedPermanently returns a 301 redirect error.
func NewStatusMovedPermanently(dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}

	trailing := strings.HasSuffix(dest, "/")
	dest, _ = fs.Clean(dest)
	if trailing && !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	return &HTTPError{
		Code: http.StatusMovedPermanently,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}

// NewStatusFound returns a 302 redirect error.
func NewStatusFound(dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}

	trailing := strings.HasSuffix(dest, "/")
	dest, _ = fs.Clean(dest)
	if trailing && !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	return &HTTPError{
		Code: http.StatusFound,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}

// NewStatusSeeOther returns a 303 redirect error.
func NewStatusSeeOther(dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}

	trailing := strings.HasSuffix(dest, "/")
	dest, _ = fs.Clean(dest)
	if trailing && !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	return &HTTPError{
		Code: http.StatusSeeOther,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}

// NewStatusTemporaryRedirect returns a 307 redirect error.
func NewStatusTemporaryRedirect(dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}

	trailing := strings.HasSuffix(dest, "/")
	dest, _ = fs.Clean(dest)
	if trailing && !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	return &HTTPError{
		Code: http.StatusTemporaryRedirect,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}

// NewStatusPermanentRedirect returns a 308 redirect error.
func NewStatusPermanentRedirect(dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}

	trailing := strings.HasSuffix(dest, "/")
	dest, _ = fs.Clean(dest)
	if trailing && !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	return &HTTPError{
		Code: http.StatusPermanentRedirect,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}

// NewStatusBadRequest returns a 400 HTTP error.
func NewStatusBadRequest(err error) *HTTPError {
	return &HTTPError{
		Code: http.StatusBadRequest,
		Err:  err,
	}
}

// NewStatusInternalServerError returns a 500 HTTP error.
func NewStatusInternalServerError(err error) *HTTPError {
	return &HTTPError{
		Code: http.StatusInternalServerError,
		Err:  err,
	}
}

// NewStatusNotModified returns a 304 HTTP error.
func NewStatusNotModified() *HTTPError {
	return &HTTPError{
		Code: http.StatusNotModified,
	}
}

// NewStatusNotFound returns a 404 HTTP error.
func NewStatusNotFound() *HTTPError {
	return &HTTPError{
		Code: http.StatusNotFound,
	}
}

// NewStatusNotAcceptable returns a 406 HTTP error.
func NewStatusNotAcceptable() *HTTPError {
	return &HTTPError{
		Code: http.StatusNotAcceptable,
	}
}
