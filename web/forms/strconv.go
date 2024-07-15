package forms

import (
	"reflect"
	"strconv"
	"strings"

	"darvaza.org/core"
)

// ParseSigned converts a string to a signed integer value
// of the correct size for the type using [strconv.ParseInt].
func ParseSigned[T core.Signed](s string) (v T, err error) {
	n, err := strconv.ParseInt(s, 0, bitSize(v))
	return T(n), err
}

// ParseUnsigned converts a string to an unsigned integer value
// of the correct size for the type using [strconv.ParseUint].
func ParseUnsigned[T core.Unsigned](s string) (v T, err error) {
	n, err := strconv.ParseUint(s, 0, bitSize(v))
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

func bitSize[T any](v T) int {
	return int(reflect.TypeOf(v).Size()) * 8
}
