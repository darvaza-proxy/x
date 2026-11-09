package tai

import "darvaza.org/x/time/num"

// NewTAIForTest returns the TAI holding count seconds.
func NewTAIForTest(count num.Int128) TAI { return TAI{x: count} }

// NewTAINForTest returns the TAIN holding count nanoseconds.
func NewTAINForTest(count num.Int128) TAIN { return TAIN{x: count} }

// NewTAIAForTest returns the TAIA holding count attoseconds.
func NewTAIAForTest(count num.Int128) TAIA { return TAIA{x: count} }
