package forms

import (
	"errors"
	"strconv"
)

// NumError is an alias of the standard string to value conversion error
type NumError = strconv.NumError

// ErrMissing indicates the field was missing from the request.
var ErrMissing = errors.New("not specified")

// ErrSyntax indicates the field isn't of the correct form.
var ErrSyntax = strconv.ErrSyntax

// ErrRange indicates the value of the field is out of range
var ErrRange = strconv.ErrRange

func errRange(fn, s string) error {
	return &strconv.NumError{
		Func: fn,
		Num:  s,
		Err:  strconv.ErrRange,
	}
}

// IsEmptyString tells if the given error was caused by the argument
// being an empty string.
func IsEmptyString(err error) bool {
	switch e := err.(type) {
	case *strconv.NumError:
		return e.Err == strconv.ErrSyntax && e.Num == ""
	default:
		return false
	}
}

// IsNilOrEmptyString tells if the given error was caused by the argument
// being an empty string but it also succeeds when no error is given.
// To be used when testing an error condition.
func IsNilOrEmptyString(err error) bool {
	switch e := err.(type) {
	case nil:
		return true
	case *strconv.NumError:
		return e.Err == strconv.ErrSyntax && e.Num == ""
	default:
		return false
	}
}
