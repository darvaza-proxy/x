package blob

import (
	"net/http"
	"path"
	"strings"

	"darvaza.org/x/web"
)

// CleanPath works like [path.Clean] except that
// it only accepts absolute paths and preserves trailing
// slashes.
func CleanPath(s string) (string, bool) {
	switch {
	case s == "" || s[0] != '/':
		return "", false
	case s == "/":
		return s, true
	}

	s0, trailing := strings.CutSuffix(s, "/")
	s1 := path.Clean(s0)

	switch {
	case s1 == "/":
		// root after cleaning, done
		return s1, true
	case s1 == ".":
		// invalid
		return "", false
	case trailing:
		// preserve trailing slash, done
		return s1 + "/", true
	default:
		// done
		return s1, true
	}
}

func joinNameSuffix(a, b string) string {
	var s string
	switch {
	case b == "":
		s = a
	case a == "":
		s = b
	default:
		s = a + b
	}

	if s == "" {
		return "."
	}
	return s
}

func newRedirect(location string) *web.HTTPError {
	h := &web.HTTPError{
		Code: http.StatusPermanentRedirect,
	}
	h.AddHeader("Location", location)
	return h
}
