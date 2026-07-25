package num

import (
	"encoding"
	"fmt"
	"strconv"

	"darvaza.org/core"
)

var (
	_ Signed[Int64]    = Int64(0)
	_ Euclidean[Int64] = Int64(0)

	_ fmt.Stringer             = Int64(0)
	_ encoding.TextMarshaler   = Int64(0)
	_ encoding.TextUnmarshaler = (*Int64)(nil)
)

// Int64 is a signed 64-bit integer wrapping the native int64,
// extending the Signed family to 64 bits with the same method surface
// as Int128. The zero value is numeric zero.
type Int64 int64

// sys returns v as its underlying native int64.
func (v Int64) sys() int64 {
	return int64(v)
}

// one returns the multiplicative unit, the [EuclideanDivMod] quotient
// step.
func (Int64) one() Int64 {
	return 1
}

// ulp returns the smallest positive value, the [EuclideanMulDivMod]
// quotient step; for an integer it equals one.
func (Int64) ulp() Int64 {
	return 1
}

// IsZero reports whether v is zero.
func (v Int64) IsZero() bool {
	return v == 0
}

// Equal reports whether v and w are equal.
func (v Int64) Equal(w Int64) bool {
	return v == w
}

// IsNegative reports whether v is less than zero.
func (v Int64) IsNegative() bool {
	return v < 0
}

// Neg returns -v, wrapping only for the most negative value, which has
// no positive counterpart.
func (v Int64) Neg() Int64 {
	return -v
}

// Abs returns the absolute value of v. The most negative value is
// returned unchanged, having no positive counterpart.
func (v Int64) Abs() Int64 {
	if v < 0 {
		return -v
	}
	return v
}

// Add returns v+w, wrapping on overflow.
func (v Int64) Add(w Int64) Int64 {
	return v + w
}

// Sub returns v-w, wrapping on overflow.
func (v Int64) Sub(w Int64) Int64 {
	return v - w
}

// Mul returns the low 64 bits of v*w, wrapping on overflow.
func (v Int64) Mul(w Int64) Int64 {
	return v * w
}

// Div returns v/w, truncated towards zero. It panics when w is zero.
func (v Int64) Div(w Int64) Int64 {
	q, _ := v.DivMod(w)
	return q
}

// Mod returns the remainder of v/w, taking the sign of v. It panics
// when w is zero.
func (v Int64) Mod(w Int64) Int64 {
	_, r := v.DivMod(w)
	return r
}

// DivMod returns the quotient and remainder of v/w. The quotient is
// truncated towards zero and the remainder takes the sign of v, so
// that v == q*w + r with |r| < |w|. It panics with [ErrDivZero] when w
// is zero.
func (v Int64) DivMod(w Int64) (q, r Int64) {
	if w == 0 {
		panic(ErrDivZero)
	}
	return v / w, v % w
}

// MulDivMod returns the quotient and remainder of v*w/d, forming the
// product in a 128-bit intermediate so it cannot overflow before the
// division. The quotient is truncated towards zero and wraps if it
// exceeds the 64-bit range. The remainder takes the sign of the product
// v*w with |r| < |d|, so v*w == q*d + r whenever the quotient does not
// wrap. It panics with [ErrDivZero] when d is zero.
func (v Int64) MulDivMod(w, d Int64) (q, r Int64) {
	if d == 0 {
		panic(ErrDivZero)
	}
	v128, w128, d128 := NewInt128(v.sys()), NewInt128(w.sys()), NewInt128(d.sys())
	q128, r128 := v128.MulDivMod(w128, d128)
	return Int64(q128.lo), Int64(r128.lo)
}

// Cmp returns -1, 0 or +1 as v is less than, equal to or greater
// than w.
func (v Int64) Cmp(w Int64) int {
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
func (v Int64) String() string {
	return strconv.FormatInt(v.sys(), 10)
}

// MarshalText renders v as its signed base-10 digits.
func (v Int64) MarshalText() ([]byte, error) {
	return strconv.AppendInt(nil, v.sys(), 10), nil
}

// UnmarshalText parses a signed base-10 integer from b into v. It rejects
// malformed input with [ErrSyntax] and values outside the int64 range
// with [ErrRange].
func (v *Int64) UnmarshalText(b []byte) error {
	x, err := strconv.ParseInt(string(b), 10, 64)
	switch {
	case err != nil:
		return parseIntError(b, err)
	case v == nil:
		return core.ErrNilReceiver
	default:
		*v = Int64(x)
		return nil
	}
}
