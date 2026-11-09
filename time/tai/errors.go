package tai

import "darvaza.org/core"

var (
	// ErrInvalidTAIAFormat reports a TAIA text label with an invalid prefix or digit.
	ErrInvalidTAIAFormat = core.QuietWrap(core.ErrInvalid, "invalid TAIA hexadecimal label")
	// ErrInvalidTAIALength reports a TAIA text label with an invalid length.
	ErrInvalidTAIALength = core.QuietWrap(core.ErrInvalid, "invalid TAIA text length")
	// ErrTAIARange reports seconds outside uint64 or a fractional field outside [0, 1e9).
	ErrTAIARange = core.QuietWrap(core.ErrInvalid, "TAIA label value out of range")
)
