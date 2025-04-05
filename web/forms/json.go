package forms

import (
	"encoding/json"
	"fmt"
	"net/url"

	"darvaza.org/core"
)

// UnmarshalFormJSON attempts to convert a JSON document
// into a url.Values map.
func UnmarshalFormJSON(b []byte) (url.Values, error) {
	// parse
	data := make(map[string]any)
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	// flatten
	values := make(url.Values)
	for k, v := range data {
		jsonValue(values, k, v)
	}

	return values, nil
}

//revive:disable-next-line:cognitive-complexity
func jsonValue(out url.Values, key string, v any) {
	switch s := v.(type) {
	case nil:
		// Null
		if _, ok := out[key]; !ok {
			out[key] = []string{}
		}
	case string:
		// String
		out.Add(key, s)
	case []any:
		// Array
		for _, v := range s {
			jsonValue(out, key, v)
		}
	case map[string]any:
		// Object
		for k, v := range s {
			jsonValue(out, key+"."+k, v)
		}
	case bool:
		// Boolean]
		out.Add(key, core.IIf(s, "true", "false"))
	default: // Number
		out.Add(key, fmt.Sprint(v))
	}
}
