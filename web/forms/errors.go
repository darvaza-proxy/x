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
