package num

import (
	"encoding"
	"fmt"
	"strconv"

	"darvaza.org/core"
)

var (
	_ Signed[Int32]    = Int32(0)
	_ Euclidean[Int32] = Int32(0)

	_ fmt.Stringer             = Int32(0)
	_ encoding.TextMarshaler   = Int32(0)
	_ encoding.TextUnmarshaler = (*Int32)(nil)
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

// Div returns v/w, truncated towards zero. It panics when w is zero.
func (v Int32) Div(w Int32) Int32 {
	q, _ := v.DivMod(w)
	return q
}

// Mod returns the remainder of v/w, taking the sign of v. It panics
// when w is zero.
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

// String returns v in base 10.
func (v Int32) String() string {
	return strconv.FormatInt(int64(v), 10)
}

// MarshalText renders v as its signed base-10 digits.
func (v Int32) MarshalText() ([]byte, error) {
	return strconv.AppendInt(nil, int64(v), 10), nil
}

// UnmarshalText parses a signed base-10 integer from b into v. It rejects
// malformed input with [ErrSyntax] and values outside the int32 range
// with [ErrRange].
func (v *Int32) UnmarshalText(b []byte) error {
	x, err := strconv.ParseInt(string(b), 10, 32)
	switch {
	case err != nil:
		return parseIntError(b, err)
	case v == nil:
		return core.ErrNilReceiver
	default:
		*v = Int32(x)
		return nil
	}
}
