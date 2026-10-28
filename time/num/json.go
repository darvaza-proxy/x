package num

import (
	"bytes"
	"encoding"
	"encoding/json"

	"darvaza.org/core"
)

// The bounds under which a value is emitted as a JSON number, safe for
// a consumer reading it into a float64: an integer at a magnitude of
// at most 2^53, past which consecutive integers stop being
// representable, and a fixed-point count and scale both below 10^15,
// fifteen significant digits, which a float64 carries through a round
// trip.
const (
	jsonSafeInt     = 1 << 53
	jsonSafeDecimal = 1e15
)

// jsonString returns the text of v as a JSON string. The text holds
// digits, a sign and a point at most, so quoting adds nothing but the
// quotes.
func jsonString(v encoding.TextAppender) ([]byte, error) {
	b, err := v.AppendText([]byte{'"'})
	return append(b, '"'), err
}

// isJSONSafeInt reports whether an integer is emitted as a JSON
// number: it fits an Int64 and its magnitude is at most 2^53.
func isJSONSafeInt(v interface{ Int64() (Int64, bool) }) bool {
	x, ok := v.Int64()
	return ok && x >= -jsonSafeInt && x <= jsonSafeInt
}

// isJSONSafeDecimal reports whether a fixed-point value is emitted as
// a JSON number: its count below 10^15 in magnitude at a scale below
// 10^15, so an Atto128 never is.
func isJSONSafeDecimal(count, scale int64) bool {
	return scale < jsonSafeDecimal &&
		count > -jsonSafeDecimal && count < jsonSafeDecimal
}

// doUnmarshalJSON finishes an UnmarshalJSON method, reading the JSON
// value data through parse, the parser of T over bytes. A null stores
// nothing, the no-op [json.Unmarshaler] asks of it, and on a nil p
// reports [core.ErrNilReceiver] as any other valid value does. A
// string's contents, as jsonStringText reads them, are parsed as
// encoding/json hands them to an UnmarshalText; a string that does not
// end on its closing quote, or does not unquote, is parsed as given,
// and fails on its quote with [ErrSyntax].
// Any other token, a number among them, is parsed as given. The value
// is stored by unmarshalInto, under the method's name.
func doUnmarshalJSON[T any](p *T, data []byte,
	parse func([]byte) (T, error)) error {
	switch {
	case string(data) == "null":
		if p == nil {
			return core.Wrap(core.ErrNilReceiver, "UnmarshalJSON")
		}
		return nil
	case len(data) > 0 && data[0] == '"':
		data = jsonStringText(data)
	default:
		// a number, or any other token, parsed as given.
	}
	x, err := parse(data)
	return unmarshalInto(p, x, err, "UnmarshalJSON")
}

// jsonStringText returns the contents of data, a JSON string from its
// opening quote: data as given when it does not end on its closing
// quote, the bytes between the quotes, in place, when there is no
// backslash, and otherwise the text [json.Unmarshal] unquotes, or data
// as given when it does not unquote.
func jsonStringText(data []byte) []byte {
	n := len(data)
	if n < 2 || data[n-1] != '"' {
		// a string left open, or followed by other bytes.
		return data
	}
	if bytes.IndexByte(data, '\\') < 0 {
		// no escape, so the contents are the bytes between the quotes.
		return data[1 : n-1]
	}
	var s string
	if json.Unmarshal(data, &s) != nil {
		// a string left open, or another that is not valid JSON.
		return data
	}
	return []byte(s)
}
