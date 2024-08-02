package resource

import (
	"net/http"
	"time"

	"darvaza.org/x/web"
)

// SetHeader sets a header value, optionally formatted.
func SetHeader(rw http.ResponseWriter, key, value string, args ...any) {
	web.SetHeader(rw.Header(), key, value, args...)
}

// SetHeaderUnlessExists sets a header value if it's not already set.
// Optional value formatting supported.
func SetHeaderUnlessExists(rw http.ResponseWriter, key, value string, args ...any) {
	web.SetHeaderUnlessExists(rw.Header(), key, value, args...)
}

// SetCache sets the Cache-Control header. a Negative
// value is translated in "private", values of a second
// or longer translated to "max-age", and otherwise
// "no-cache"
func SetCache(rw http.ResponseWriter, d time.Duration) {
	web.SetCache(rw.Header(), d)
}
