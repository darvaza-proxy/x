package resource

import (
	"fmt"
	"net/http"
)

// SetHeader sets a header value, optionally formatted.
func SetHeader(rw http.ResponseWriter, key, value string, args ...any) {
	if len(args) > 0 {
		value = fmt.Sprintf(value, args...)
	}

	rw.Header().Set(key, value)
}
