package web

import (
	"net/http"

	"darvaza.org/x/web/consts"
)

// NewStatusBadRequest returns a 400 HTTP error.
func NewStatusBadRequest(err error) *HTTPError {
	return &HTTPError{
		Code: http.StatusBadRequest,
		Err:  err,
	}
}

// NewStatusMethodNotAllowed returns a 405 HTTP error
func NewStatusMethodNotAllowed(allowed ...string) *HTTPError {
	return &HTTPError{
		Code: http.StatusMethodNotAllowed,
		Hdr: http.Header{
			consts.Allow: allowed,
		},
	}
}
