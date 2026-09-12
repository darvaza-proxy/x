package num

import (
	"encoding"
	"fmt"
	"strconv"
)

var (
	_ Signed[Int32]    = Int32(0)
	_ Euclidean[Int32] = Int32(0)

	_ fmt.Formatter  = Int32(0)
	_ fmt.GoStringer = Int32(0)
	_ fmt.Stringer   = Int32(0)

	_ encoding.TextAppender  = Int32(0)
	_ encoding.TextMarshaler = Int32(0)
)

// Int32 is a signed 32-bit integer wrapping the native int32,
// extending the Signed family down to 32 bits with the same method
// surface as Int128. The zero value is numeric zero.
type Int32 int32

// one returns the multiplicative unit, the [EuclideanDivMod] quotient
// step.
func (Int32) one() Int32 {
	return 1
}

// ulp returns the smallest positive value, the [EuclideanMulDivMod]
// quotient step; for an integer it equals one.
func (Int32) ulp() Int32 {
	return 1
}

// AsInt32 takes a signed 32-bit value as an Int32, the conversion.
func AsInt32(x int32) Int32 {
	return Int32(x)
}

// Format implements [fmt.Formatter] with the verbs d, v and s for
// decimal, x and X for hex, o and O for octal and b for binary, handed
// to fmt over the native int32 so the flags, width and precision behave
// as they do there; %#v prints the GoString form. Any other verb prints
// as %!verb(num.Int32=value).
func (v Int32) Format(s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('#'):
		writeGoString(s, v.GoString())
	case !isFormatVerb(verb):
		writeBadVerb(s, verb, "num.Int32", func() {
			_, _ = fmt.Fprintf(s, fmt.FormatString(s, 'd'), int32(v))
		})
	default:
		_, _ = fmt.Fprintf(s, fmt.FormatString(s, nativeVerb(verb)), int32(v))
	}
}

// GoString returns the constructor that rebuilds v, num.AsInt32(-5),
// for %#v.
func (v Int32) GoString() string {
	return fmt.Sprintf("num.AsInt32(%d)", int32(v))
}

// String returns v in decimal, the text %v prints.
func (v Int32) String() string {
	return strconv.FormatInt(int64(v), 10)
}

// AppendText implements [encoding.TextAppender], appending v in
// decimal, the text %v prints, to b. It allocates only when b lacks
// the room, and the error is always nil.
func (v Int32) AppendText(b []byte) ([]byte, error) {
	return v.doAppendText(b), nil
}

// MarshalText implements [encoding.TextMarshaler], returning the
// AppendText text.
func (v Int32) MarshalText() ([]byte, error) {
	return v.AppendText(nil)
}

// doAppendText writes v in decimal to dst and returns the extended
// buffer.
func (v Int32) doAppendText(dst []byte) []byte {
	return strconv.AppendInt(dst, int64(v), 10)
}

// IsZero reports whether v is zero.
func (v Int32) IsZero() bool {
	return v == 0
}

// Equal reports whether v and w are equal.
func (v Int32) Equal(w Int32) bool {
	return v == w
}

// IsNegative reports whether v is less than zero.
func (v Int32) IsNegative() bool {
	return v < 0
}

// Neg returns -v, wrapping only for the most negative value, which has
// no positive counterpart.
func (v Int32) Neg() Int32 {
	return -v
}

// Abs returns the absolute value of v. The most negative value is
// returned unchanged, having no positive counterpart.
func (v Int32) Abs() Int32 {
	if v < 0 {
		return -v
	}
	return v
}

// Add returns v+w, wrapping on overflow.
func (v Int32) Add(w Int32) Int32 {
	return v + w
}

// Sub returns v-w, wrapping on overflow.
func (v Int32) Sub(w Int32) Int32 {
	return v - w
}

// Mul returns the low 32 bits of v*w, wrapping on overflow.
func (v Int32) Mul(w Int32) Int32 {
	return v * w
}

// Div returns v/w, truncated towards zero. It panics with [ErrDivZero]
// when w is zero.
func (v Int32) Div(w Int32) Int32 {
	q, _ := v.DivMod(w)
	return q
}

// Mod returns the remainder of v/w, taking the sign of v. It panics
// with [ErrDivZero] when w is zero.
func (v Int32) Mod(w Int32) Int32 {
	_, r := v.DivMod(w)
	return r
}

// DivMod returns the quotient and remainder of v/w. The quotient is
// truncated towards zero and the remainder takes the sign of v, so
// that v == q*w + r with |r| < |w|. It panics with [ErrDivZero] when w
// is zero.
func (v Int32) DivMod(w Int32) (q, r Int32) {
	if w == 0 {
		panic(ErrDivZero)
	}
	return v / w, v % w
}

// MulDivMod returns the quotient and remainder of v*w/d, forming the
// product in a 64-bit intermediate so it cannot overflow before the
// division. The quotient is truncated towards zero and wraps if it
// exceeds the 32-bit range. The remainder takes the sign of the product
// v*w with |r| < |d|, so v*w == q*d + r whenever the quotient does not
// wrap. It panics with [ErrDivZero] when d is zero.
func (v Int32) MulDivMod(w, d Int32) (q, r Int32) {
	if d == 0 {
		panic(ErrDivZero)
	}
	prod := int64(v) * int64(w)
	den := int64(d)
	return Int32(prod / den), Int32(prod % den)
}

// Cmp returns -1, 0 or +1 as v is less than, equal to or greater
// than w.
func (v Int32) Cmp(w Int32) int {
	switch {
	case v > w:
		return 1
	case v < w:
		return -1
	default:
		return 0
	}
}
