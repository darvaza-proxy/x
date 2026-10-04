package num

import "darvaza.org/core"

// ErrDivZero is the value panicked on division by zero. It wraps
// [core.ErrInvalid], so a caller recovering from Div, Mod, DivMod or
// MulDivMod can match it with errors.Is against either ErrDivZero or
// ErrInvalid.
var ErrDivZero = core.QuietWrap(core.ErrInvalid, "num: division by zero")

// ErrPow10Range is the value [Pow10] panics with for an exponent
// outside the powers of ten a Uint128 holds, a caller's mistake.
var ErrPow10Range = core.QuietWrap(core.ErrInvalid,
	"num: Pow10 exponent out of range")
