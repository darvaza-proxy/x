package web

import (
	"fmt"
	"net/http"
	"strings"

	"darvaza.org/x/web/consts"
)

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

// newRedirect builds an HTTPError carrying a Location header
// for the given redirect status code. When args is non-empty,
// dest is Sprintf-formatted; otherwise it passes through
// verbatim so URL percent-escapes survive unchanged. The
// composed dest is then normalised through CleanURL: scheme
// passes verbatim, the host is lowercased with IDN labels
// converted to ASCII punycode and IPv6 canonicalised, the
// default port is stripped when it matches the scheme, and
// the path is reduced through Clean.
//
// When CleanURL rejects the composed dest — url.Parse,
// core.SplitHostPort, or the DNS label-length check —
// newRedirect wraps the failure as a 500 via
// NewStatusInternalServerError rather than emit a broken
// Location.
//
// newRedirect does not validate origin. Callers passing
// untrusted input must allowlist scheme or origin themselves.
func newRedirect(code int, dest string, args ...any) *HTTPError {
	if len(args) > 0 {
		dest = fmt.Sprintf(dest, args...)
	}
	dest, err := CleanURL(dest)
	if err != nil {
		return NewStatusInternalServerError(err)
	}
	return &HTTPError{
		Code: code,
		Hdr: http.Header{
			consts.Location: []string{dest},
		},
	}
}
