package num

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Every type prints itself under %#v as the constructor call that
// rebuilds it, so a value dumped from a failing test pastes back into a
// row. The count forms, the As constructors, print the native integer
// in decimal; the words forms, NewUint128 and NewInt128, print the two
// 64-bit words in hex, where the word boundary is visible.
//
// Every other verb goes through Format, in the shape fmt gives its own
// integers: d, v and s for decimal, x and X for hex, o and O for octal
// and b for binary, with the '+', ' ', '#', '-' and '0' flags, width
// and precision meaning what they mean there. An unsupported verb
// prints the %!verb(type=value) form fmt uses.

// groupDigits returns x in decimal with an underscore every three
// digits from the right, so an eighteen-digit atto fraction reads in
// thousands. Fewer than four digits are left alone.
func groupDigits(x int64) string {
	var buf [20]byte
	digits := strconv.AppendInt(buf[:0], x, 10)
	b := make([]byte, 0, len(digits)+len(digits)/3)
	if digits[0] == '-' {
		b = append(b, '-')
		digits = digits[1:]
	}
	for i, c := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b = append(b, '_')
		}
		b = append(b, c)
	}
	return string(b)
}

// decChunk is the largest power of ten fitting in a uint64. Peeling a
// Uint128 by it yields base-10 groups of decChunkDigits digits each,
// which strconv then formats one word at a time.
const decChunk uint64 = 1e19

// decChunkDigits is the number of decimal digits in decChunk.
const decChunkDigits = 19

// appendPadded appends v in base 10 to dst, left-padded with zeros to at
// least width digits.
func appendPadded(dst []byte, v uint64, width int) []byte {
	var tmp [20]byte
	s := strconv.AppendUint(tmp[:0], v, 10)
	for pad := width - len(s); pad > 0; pad-- {
		dst = append(dst, '0')
	}
	return append(dst, s...)
}

// verbShift returns the bits per digit of a power-of-two verb, zero for
// the decimal verbs and for anything Format does not take.
func verbShift(verb rune) uint {
	switch verb {
	case 'b':
		return 1
	case 'o', 'O':
		return 3
	case 'x', 'X':
		return 4
	default:
		return 0
	}
}

// verbDigits returns the digit alphabet of a power-of-two verb: upper
// case for X, lower case otherwise.
func verbDigits(verb rune) string {
	if verb == 'X' {
		return "0123456789ABCDEF"
	}
	return "0123456789abcdef"
}

// nativeVerb returns the verb to hand fmt for a native integer: s
// prints decimal, as d does, and the rest are fmt's own, v included, so
// that fmt applies its own flag rules to it rather than this package's
// reading of them.
func nativeVerb(verb rune) rune {
	if verb == 's' {
		return 'd'
	}
	return verb
}

// isFormatVerb reports whether Format takes verb on an integer.
func isFormatVerb(verb rune) bool {
	switch verb {
	case 'd', 'v', 's':
		return true
	default:
		return verbShift(verb) != 0
	}
}

// formatSign returns the sign to print: the minus of a negative value,
// or the '+' or ' ' the flags ask for on a non-negative one. Under the
// v verb fmt hands over the '+' of %+v, which asks for the field names
// of a struct rather than for a sign, so only the space flag signs
// there.
func formatSign(s fmt.State, verb rune, neg bool) string {
	switch {
	case neg:
		return "-"
	case s.Flag('+') && verb != 'v':
		return "+"
	case s.Flag(' '):
		return " "
	default:
		return ""
	}
}

// formatPrefix returns the base prefix of a number already carrying
// zeros leading zeros: the 0o that O always takes, and on top of it the
// prefix '#' asks for on b, o, O, x and X. The octal 0 is not added to
// a leading zero, since that zero is already the prefix, so a precision
// covers it and %#.3o of 1 is 001, as fmt has it.
func formatPrefix(s fmt.State, verb rune, zeros int, digits []byte) string {
	base := ""
	if verb == 'O' {
		base = "0o"
	}
	if !s.Flag('#') {
		return base
	}
	switch verb {
	case 'b':
		return base + "0b"
	case 'o', 'O':
		if zeros > 0 || digits[0] == '0' {
			return base
		}
		return base + "0"
	case 'x':
		return base + "0x"
	case 'X':
		return base + "0X"
	default:
		return base
	}
}

// numField is a number ready for padding: its sign, base prefix, the
// zeros leading its digits, and the digits.
type numField struct {
	sign   string
	prefix string
	digits []byte
	zeros  int
}

// writeTo writes the field to s, padded with spaces to the width s asks
// for, on the left or, under '-', on the right.
func (f numField) writeTo(s fmt.State) {
	width, _ := s.Width()
	pad := max(0, width-len(f.sign)-len(f.prefix)-f.zeros-len(f.digits))
	left, right := " ", ""
	if s.Flag('-') {
		left, right = "", " "
	}
	_, _ = io.WriteString(s, strings.Repeat(left, pad))
	_, _ = io.WriteString(s, f.sign)
	_, _ = io.WriteString(s, f.prefix)
	_, _ = io.WriteString(s, strings.Repeat("0", f.zeros))
	_, _ = s.Write(f.digits)
	_, _ = io.WriteString(s, strings.Repeat(right, pad))
}

// padZeros returns the zeros the '0' flag turns a width into. fmt
// treats them as a precision, so they leave room for the sign but not
// for the base prefix, which then sits outside the width; the '-' flag
// switches them off.
func padZeros(s fmt.State, sign, digits int) int {
	width, ok := s.Width()
	if !ok || !s.Flag('0') || s.Flag('-') {
		return 0
	}
	return max(0, width-sign-digits)
}

// writeNumber writes an integer's sign, prefix and digits to s under
// the flags, width and precision of the verb. Precision is the least
// number of digits, met with leading zeros, and takes the place of the
// '0' flag; a zero under precision zero prints as padding alone. Both
// as fmt does.
func writeNumber(s fmt.State, verb rune, neg bool, digits []byte) {
	zero := len(digits) == 1 && digits[0] == '0'
	prec, hasPrec := s.Precision()
	if hasPrec && prec == 0 && zero {
		numField{}.writeTo(s)
		return
	}
	f := numField{
		digits: digits,
		sign:   formatSign(s, verb, neg),
	}
	switch {
	case hasPrec:
		f.zeros = max(0, prec-len(digits))
	default:
		f.zeros = padZeros(s, len(f.sign), len(digits))
	}
	f.prefix = formatPrefix(s, verb, f.zeros, digits)
	f.writeTo(s)
}

// pow10 returns 10^n for 0 <= n <= 18.
func pow10(n int) int64 {
	p := int64(1)
	for range n {
		p *= 10
	}
	return p
}

// writeGoString writes the GoString form of a value to s under the
// width and precision fmt gives a [fmt.GoStringer], which pads and
// truncates it as it pads and truncates any string. fmt applies them
// itself only when it reaches that path, which a [fmt.Formatter] never
// lets it do, so the verb hands the text straight back to it.
func writeGoString(s fmt.State, gs string) {
	_, _ = fmt.Fprintf(s, fmt.FormatString(s, 's'), gs)
}

// writeBadVerb writes the %!verb(type=value) form fmt prints for a verb
// a type does not take, calling write for the value. fmt prints that
// value under the flags, width and precision the bad verb was given,
// reading them as it reads them for d, since '+' and '#' take the
// meanings v gives them only under v itself; nothing around the value
// is padded.
func writeBadVerb(s fmt.State, verb rune, typeName string, write func()) {
	_, _ = fmt.Fprintf(s, "%%!%c(%s=", verb, typeName)
	write()
	_, _ = io.WriteString(s, ")")
}
