package resource

import (
	"fmt"
	"net/http"
	"sort"
)

func sortedKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SetHeader sets a header value, optionally formatted.
func SetHeader(rw http.ResponseWriter, key, value string, args ...any) {
	if len(args) > 0 {
		value = fmt.Sprintf(value, args...)
	}

	rw.Header().Set(key, value)
}
