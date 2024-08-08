package web

import (
	"net/http"
	"strings"

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
	hdr := make(http.Header)

	if s := strings.Join(allowed, ", "); s != "" {
		hdr[consts.Allow] = []string{s}
	}

	return &HTTPError{
		Code: http.StatusMethodNotAllowed,
		Hdr: http.Header{
			consts.Allow: allowed,
		},
	}
}
