package num

import "darvaza.org/core"

// ErrDivZero is the value panicked on division by zero. It wraps
// [core.ErrInvalid], so a caller recovering from Div, Mod, DivMod or
// MulDivMod can match it with errors.Is against either ErrDivZero or
// ErrInvalid.
var ErrDivZero = core.QuietWrap(core.ErrInvalid, "num: division by zero")
