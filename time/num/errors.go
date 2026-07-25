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

// ErrSyntax is reported by the text parsers for input that is not a base-10
// integer: an empty string or a stray non-digit byte. It is
// [strconv.ErrSyntax], re-exported so callers need not import strconv to
// match it; the parsers additionally wrap [core.ErrInvalid].
var ErrSyntax = strconv.ErrSyntax

// ErrRange is reported by the text parsers for a value outside the target
// type's range. It is [strconv.ErrRange], re-exported likewise, and the
// parsers additionally wrap [core.ErrInvalid].
var ErrRange = strconv.ErrRange
