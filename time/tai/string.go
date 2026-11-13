package tai

import (
	"fmt"

	"darvaza.org/x/time/num"
)

const nanosPerSec = AttosPerSec / AttosPerNano

var (
	_ fmt.Stringer = TAI{}
	_ fmt.Stringer = TAIN{}
	_ fmt.Stringer = TAIA{}
)

// String returns the canonical lowercase TAI64 label, @ and 16 hexadecimal
// digits. The internal count can exceed the label range: seconds outside
// uint64 produce TAI(<count>s), a diagnostic string that is not a label.
func (t TAI) String() string {
	if label, ok := labelString(t.x, num.ZeroInt128, num.ZeroInt128, TAITextLength); ok {
		return label
	}
	return "TAI(" + t.x.String() + "s)"
}

// String returns the canonical lowercase TAI64N label, @ and 24 hexadecimal
// digits. The internal count can exceed the label range: seconds outside
// uint64 produce TAIN(<count>ns), a diagnostic string that is not a label.
func (t TAIN) String() string {
	s, n := num.EuclideanDivMod(t.x, num.AsInt128(nanosPerSec))
	if label, ok := labelString(s, n, num.ZeroInt128, TAINTextLength); ok {
		return label
	}
	return "TAIN(" + t.x.String() + "ns)"
}

// String returns the canonical lowercase TAI64NA label, @ and 32 hexadecimal
// digits. The internal count can exceed the label range: seconds outside
// uint64 produce TAIA(<count>as), a diagnostic string that is not a label.
// Use MarshalText when an out-of-range value must return an error.
func (t TAIA) String() string {
	s, n, a := t.Split()
	if label, ok := labelString(s, n, a, TAIATextLength); ok {
		return label
	}
	return "TAIA(" + t.x.String() + "as)"
}

func labelString(s, n, a num.Int128, length int) (string, bool) {
	var buf [TAIATextLength]byte
	b, err := appendLabelText(buf[:0], s, n, a, length)
	return string(b), err == nil
}
