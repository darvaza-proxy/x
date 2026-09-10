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

// ErrPow10Range is the value [Pow10] panics with for an exponent
// outside the powers of ten a Uint128 holds, a caller's mistake.
var ErrPow10Range = core.QuietWrap(core.ErrInvalid,
	"num: Pow10 exponent out of range")

// ErrSyntax is the cause a [ParseError] carries for text that is not a
// number in the accepted form. It matches both [strconv.ErrSyntax] and
// [core.ErrInvalid] under errors.Is.
var ErrSyntax = core.QuietWrap(
	core.NewCompoundError(strconv.ErrSyntax, core.ErrInvalid),
	"num: invalid syntax")

// ErrRange is the cause a [ParseError] carries for a well-formed number
// outside the type's range. It matches both [strconv.ErrRange] and
// [core.ErrInvalid] under errors.Is.
var ErrRange = core.QuietWrap(
	core.NewCompoundError(strconv.ErrRange, core.ErrInvalid),
	"num: value out of range")

// ParseError reports a failed parse in the shape of a [strconv.NumError]:
// the function that failed, the input it was given and the sentinel,
// [ErrSyntax] or [ErrRange], which Unwrap returns so errors.Is reaches
// it. Match it with errors.As against *ParseError; the strconv type is
// not in the chain.
type ParseError strconv.NumError

// Error returns the report in the shape strconv gives it, without the
// sentinel when it carries none, and nothing for a nil report.
func (e *ParseError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Err == nil:
		return e.Func + ": parsing " + strconv.Quote(e.Num)
	default:
		return e.Func + ": parsing " + strconv.Quote(e.Num) + ": " + e.Err.Error()
	}
}

// Unwrap returns the sentinel, nil for a nil report.
func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
