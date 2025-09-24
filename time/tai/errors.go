// Package tai provides TAI (International Atomic Time) and TAIN timestamps
// that follow the Go time package API conventions.
package tai

import (
	"darvaza.org/core"
)

// Error variables for common error conditions
var (
	ErrInvalidTAIFormat       = core.QuietWrap(core.ErrInvalid, "invalid TAI format: expected @XXXXXXXXXXXXXXXX")
	ErrInvalidTAILength       = core.QuietWrap(core.ErrInvalid, "invalid TAI length")
	ErrInvalidTAIBinaryLength = core.QuietWrap(core.ErrInvalid, "invalid TAI binary data length")
)

// Error variables for TAIN operations
var (
	ErrInvalidTAINFormat = core.QuietWrap(core.ErrInvalid,
		"invalid TAI64N format: expected @XXXXXXXXXXXXXXXXXXXXXXXX")
	ErrInvalidTAINBinaryLength = core.QuietWrap(core.ErrInvalid, "invalid TAIN binary data length")
	ErrInvalidTAINLength       = core.QuietWrap(core.ErrInvalid, "invalid TAIN length")
)

// Error variables for TAINA operations
var (
	ErrInvalidTAINAFormat = core.QuietWrap(core.ErrInvalid,
		"invalid TAI64NA format: expected @XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	ErrInvalidTAINABinaryLength = core.QuietWrap(core.ErrInvalid, "invalid TAINA binary data length")
	ErrInvalidTAINALength       = core.QuietWrap(core.ErrInvalid, "invalid TAINA data length")
	ErrInvalidNanosecondRange   = core.QuietWrap(core.ErrInvalid, "nanoseconds must be in range 0-999999999")
	ErrInvalidAttosecondRange   = core.QuietWrap(core.ErrInvalid, "attoseconds must be in range 0-999999999")
	ErrOverflow                 = core.QuietWrap(core.ErrInvalid, "arithmetic overflow")
	ErrUnderflow                = core.QuietWrap(core.ErrInvalid, "arithmetic underflow")
)
