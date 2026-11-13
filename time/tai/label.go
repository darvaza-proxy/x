package tai

import (
	"encoding"
	"math"
	"strconv"

	"darvaza.org/core"

	"darvaza.org/x/time/num"
)

const (
	labelSecondsDigits  = 2 * TAILength
	labelFractionDigits = 2 * (TAINLength - TAILength)
	labelSecondsEnd     = 1 + labelSecondsDigits
	labelNanoEnd        = labelSecondsEnd + labelFractionDigits
	labelHexDigits      = "0123456789abcdef"
	labelHexBase        = len(labelHexDigits)
	labelTextPrefix     = '@'

	// TAITextLength is the length of an @-prefixed hexadecimal TAI64 label.
	TAITextLength = 1 + labelSecondsDigits

	// TAINTextLength is the length of an @-prefixed hexadecimal TAI64N label.
	TAINTextLength = TAITextLength + labelFractionDigits

	// TAIATextLength is the length of an @-prefixed hexadecimal TAI64NA label.
	TAIATextLength = TAINTextLength + labelFractionDigits
)

var (
	maxLabelSeconds = num.NewInt128(0, math.MaxUint64)

	_ encoding.TextMarshaler   = TAIA{}
	_ encoding.TextAppender    = TAIA{}
	_ encoding.TextUnmarshaler = (*TAIA)(nil)
)

// ParseTAIA reads @ followed by 16 hexadecimal seconds digits, eight
// nanoseconds digits and eight attoseconds digits. Both uppercase and lowercase
// are accepted. The seconds are the libtai label, with no epoch adjustment or
// UTC conversion. Each fractional field must be below 1e9.
// It returns zero and an error matching ErrInvalidTAIALength,
// ErrInvalidTAIAFormat or ErrTAIARange for an invalid label. Each also
// matches core.ErrInvalid.
//
//revive:disable-next-line:exported The parser is named for its result type.
func ParseTAIA(text string) (TAIA, error) {
	t, err := parseTAIAText([]byte(text))
	return t, core.Wrap(err, "ParseTAIA")
}

// MarshalText implements encoding.TextMarshaler, writing the label ParseTAIA
// reads, with lowercase hexadecimal digits. It fails with ErrTAIARange when
// the seconds do not fit uint64. The full unsigned seconds field is preserved,
// including the range the TAI64 specification reserves for future extensions.
func (t TAIA) MarshalText() ([]byte, error) {
	return t.AppendText(nil)
}

// AppendText implements encoding.TextAppender, appending the hexadecimal
// label MarshalText writes. An error leaves b unchanged.
func (t TAIA) AppendText(b []byte) ([]byte, error) {
	s, n, a := t.Split()
	b, err := appendLabelText(b, s, n, a, TAIATextLength)
	return b, core.Wrap(err, "TAIA.AppendText")
}

// appendLabelText writes the fields through the requested label length. Its
// callers supply normalised fractions and one of the three text lengths.
func appendLabelText(b []byte, s, n, a num.Int128, length int) ([]byte, error) {
	if s.IsNegative() || s.Cmp(maxLabelSeconds) > 0 {
		return b, ErrTAIARange
	}
	// Int64 keeps the low word even above MaxInt64. The range check above
	// establishes that this word is the whole unsigned seconds field.
	seconds, _ := s.Int64()
	nano, _ := n.Int64()
	atto, _ := a.Int64()
	b = append(b, labelTextPrefix)
	b = appendLabelHex(b, uint64(seconds), labelSecondsDigits)
	if length >= TAINTextLength {
		b = appendLabelHex(b, uint64(nano), labelFractionDigits)
	}
	if length == TAIATextLength {
		b = appendLabelHex(b, uint64(atto), labelFractionDigits)
	}
	return b, nil
}

// appendLabelHex appends v in lowercase hexadecimal, zero-filled to width.
// The caller ensures v fits the field and width is at most labelSecondsDigits.
func appendLabelHex(b []byte, v uint64, width int) []byte {
	var buf [labelSecondsDigits]byte
	for i := width - 1; i >= 0; i-- {
		buf[i] = labelHexDigits[v%uint64(labelHexBase)]
		v /= uint64(labelHexBase)
	}
	return append(b, buf[:width]...)
}

// UnmarshalText implements encoding.TextUnmarshaler, reading the label
// ParseTAIA accepts. A failure leaves t unchanged; a valid label on a nil
// receiver fails with core.ErrNilReceiver.
func (t *TAIA) UnmarshalText(text []byte) error {
	v, err := parseTAIAText(text)
	switch {
	case err != nil:
		return core.Wrap(err, "TAIA.UnmarshalText")
	case t == nil:
		return core.Wrap(core.ErrNilReceiver, "TAIA.UnmarshalText")
	default:
		*t = v
		return nil
	}
}

// parseTAIAText validates a fixed-width hexadecimal label and combines its
// unsigned seconds and fractional fields. Invalid input returns a zero
// timestamp and the corresponding label error sentinel.
func parseTAIAText(text []byte) (TAIA, error) {
	switch {
	case len(text) != TAIATextLength:
		return TAIA{}, ErrInvalidTAIALength
	case text[0] != labelTextPrefix:
		return TAIA{}, ErrInvalidTAIAFormat
	}
	seconds, err := strconv.ParseUint(string(text[1:labelSecondsEnd]), labelHexBase, 64)
	if err != nil {
		return TAIA{}, ErrInvalidTAIAFormat
	}
	nano, err := strconv.ParseUint(string(text[labelSecondsEnd:labelNanoEnd]), labelHexBase, 32)
	if err != nil {
		return TAIA{}, ErrInvalidTAIAFormat
	}
	atto, err := strconv.ParseUint(string(text[labelNanoEnd:]), labelHexBase, 32)
	if err != nil {
		return TAIA{}, ErrInvalidTAIAFormat
	}
	if nano >= uint64(AttosPerNano) || atto >= uint64(AttosPerNano) {
		return TAIA{}, ErrTAIARange
	}
	return Combine(num.NewInt128(0, seconds), num.AsInt128(int64(nano)), num.AsInt128(int64(atto))), nil
}
