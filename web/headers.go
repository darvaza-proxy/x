package web

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"darvaza.org/x/web/consts"
)

// SetHeader sets a header value, optionally formatted.
func SetHeader(hdr http.Header, key, value string, args ...any) {
	doSetHeader(hdr, key, value, args...)
}

// SetHeaderUnlessExists sets a header value if it's not already set.
// Optional value formatting supported.
func SetHeaderUnlessExists(hdr http.Header, key, value string, args ...any) {
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
func SetCache(hdr http.Header, d time.Duration) {
	cacheControl := cacheControlValue(d)
	hdr[consts.CacheControl] = []string{cacheControl}
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

// SetNoCache sets the Cache-Control header to "no-cache"
func SetNoCache(hdr http.Header) {
	hdr[consts.CacheControl] = []string{"no-cache"}
}

// SetRetryAfter sets the Retry-After header with a duration in seconds.
// Uses math.Ceil to round up, ensuring non-zero delays result in at least
// 1 second. Negative durations are clamped to 0.
func SetRetryAfter(hdr http.Header, retryAfter time.Duration) {
	var seconds int
	if retryAfter < 0 {
		seconds = 0
	} else {
		// Ceil rounds up - any non-zero delay becomes at least 1 second
		seconds = int(math.Ceil(retryAfter.Seconds()))
	}
	hdr[consts.RetryAfter] = []string{fmt.Sprintf("%d", seconds)}
}
