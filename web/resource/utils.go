package resource

import (
	"fmt"
	"net/http"
	"time"

	"darvaza.org/x/web/consts"
)

// SetHeader sets a header value, optionally formatted.
func SetHeader(rw http.ResponseWriter, key, value string, args ...any) {
	doSetHeader(rw.Header(), key, value, args...)
}

// SetHeaderUnlessExists sets a header value if it's not already set.
// Optional value formatting supported.
func SetHeaderUnlessExists(rw http.ResponseWriter, key, value string, args ...any) {
	hdr := rw.Header()
	if !headerExists(hdr, key) {
		doSetHeader(hdr, key, value, args...)
	}
}

func doSetHeader(hdr http.Header, key, value string, args ...any) {
	if len(args) > 0 {
		value = fmt.Sprintf(value, args...)
	}

	hdr.Set(key, value)
}

func headerExists(hdr http.Header, key string) bool {
	s, ok := hdr[key]
	if !ok || len(s) == 0 || s[0] == "" {
		return false
	}
	return true
}

// SetCache sets the Cache-Control header. a Negative
// value is translated in "private", values of a second
// or longer translated to "max-age", and otherwise
// "no-cache"
func SetCache(rw http.ResponseWriter, d time.Duration) {
	rw.Header().Set(consts.CacheControl, cacheControlValue(d))
}

func cacheControlValue(d time.Duration) string {
	if d < 0 {
		return "private"
	}

	if sec := d / time.Second; sec > 0 {
		return fmt.Sprintf("max-age=%v", sec)
	}

	return "no-cache"
}
