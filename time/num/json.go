package num

import "encoding"

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
