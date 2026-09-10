package num

import (
	"strconv"

	"darvaza.org/core"
)

// ErrDivZero is the value panicked on division by zero. It wraps
// [core.ErrInvalid], so a caller recovering from Div, Mod, DivMod or
// MulDivMod can match it with errors.Is against either ErrDivZero or
// ErrInvalid.
var ErrDivZero = core.QuietWrap(core.ErrInvalid, "num: division by zero")

// ErrSyntax is the error a Parse function or UnmarshalText returns for
// text that is not a number in the accepted form. It matches both
// [strconv.ErrSyntax] and [core.ErrInvalid] under errors.Is.
var ErrSyntax = core.QuietWrap(
	core.NewCompoundError(strconv.ErrSyntax, core.ErrInvalid),
	"num: invalid syntax")

// ErrRange is the error a Parse function or UnmarshalText returns for
// a well-formed number outside the type's range. It matches both
// [strconv.ErrRange] and [core.ErrInvalid] under errors.Is.
var ErrRange = core.QuietWrap(
	core.NewCompoundError(strconv.ErrRange, core.ErrInvalid),
	"num: value out of range")

// ParseError reports a failed parse in the shape of a [strconv.NumError]:
// the function that failed, the input it was given and the sentinel,
// [ErrSyntax] or [ErrRange], which Unwrap returns so errors.Is reaches
// it. Match it with errors.As against *ParseError; the strconv type is
// not in the chain.
type ParseError strconv.NumError

func (e *ParseError) Error() string {
	return e.Func + ": parsing " + strconv.Quote(e.Num) + ": " + e.Err.Error()
}

// Unwrap returns the sentinel.
func (e *ParseError) Unwrap() error {
	return e.Err
}
