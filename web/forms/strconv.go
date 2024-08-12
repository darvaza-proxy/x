package forms

import (
	"reflect"
	"strconv"
	"strings"

	"darvaza.org/core"
)

// ParseSigned converts a string to a signed integer value
// of the correct size for the type using [strconv.ParseInt].
func ParseSigned[T core.Signed](s string, base int) (v T, err error) {
	n, err := strconv.ParseInt(s, base, bitSize(v))
	return T(n), err
}

// ParseUnsigned converts a string to an unsigned integer value
// of the correct size for the type using [strconv.ParseUint].
func ParseUnsigned[T core.Unsigned](s string, base int) (v T, err error) {
	n, err := strconv.ParseUint(s, base, bitSize(v))
	return T(n), err
}

// ParseFloat converts a string to a floating point value
// of the correct size for the type using [strconv.ParseFloat].
func ParseFloat[T core.Float](s string) (v T, err error) {
	f, err := strconv.ParseFloat(s, bitSize(v))
	return T(f), err
}

// ParseBool extends the standard [strconv.ParseBool] making
// the comparison case-insensitive and allowing y/yes/n/no
// options.
func ParseBool[T core.Bool](s string) (T, error) {
	s = strings.ToLower(s)
	switch s {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		v, err := strconv.ParseBool(s)
		return T(v), err
	}
}

// FormatSigned extends [strconv.FormatInt] working with
// any [Signed] type.
func FormatSigned[T core.Signed](v T, base int) string {
	return strconv.FormatInt(int64(v), base)
}

// FormatUnsigned extends [strconv.FormatUint] working with
// any [Unsigned] type.
func FormatUnsigned[T core.Unsigned](v T, base int) string {
	return strconv.FormatUint(uint64(v), base)
}

// FormatFloat extends [strconv.FormatFloat] working with
// any [Float] type.
func FormatFloat[T core.Float](v T, fmt byte, prec int) string {
	return strconv.FormatFloat(float64(v), fmt, prec, bitSize(v))
}

func bitSize[T any](v T) int {
	return int(reflect.TypeOf(v).Size()) * 8
}
