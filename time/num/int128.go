package num

import (
	"encoding"
	"fmt"
	"math/bits"

	"darvaza.org/core"
)

var (
	_ Signed[Int128]    = Int128{}
	_ Euclidean[Int128] = Int128{}

	_ fmt.Stringer             = Int128{}
	_ encoding.TextMarshaler   = Int128{}
	_ encoding.TextUnmarshaler = (*Int128)(nil)
)

// Int128 is a signed 128-bit integer in two's-complement form,
// sharing the two-word layout of Uint128. The zero value is numeric
// zero.
type Int128 Uint128

func (v Int128) bits() Uint128 {
	return Uint128(v)
}

// one returns the multiplicative unit, the [EuclideanDivMod] quotient
// step.
func (Int128) one() Int128 {
	return Int128{lo: 1}
}

// ulp returns the smallest positive value, the [EuclideanMulDivMod]
// quotient step; for an integer it equals one.
func (Int128) ulp() Int128 {
	return Int128{lo: 1}
}

// NewInt128 sign-extends a signed 64-bit value into an Int128.
func NewInt128(x int64) Int128 {
	var hi uint64
	if x < 0 {
		hi = maxUint64
	}
	return Int128{hi: hi, lo: uint64(x)}
}

// IsZero reports whether v is zero.
func (v Int128) IsZero() bool {
	return v.hi == 0 && v.lo == 0
}

// Equal reports whether v and w are equal.
func (v Int128) Equal(w Int128) bool {
	return v == w
}

// IsNegative reports whether v is less than zero.
func (v Int128) IsNegative() bool {
	return v.hi&(1<<63) != 0
}

// Neg returns -v, wrapping only for MinInt128, which has no
// positive counterpart.
func (v Int128) Neg() Int128 {
	lo, carry := bits.Add64(^v.lo, 1, 0)
	hi, _ := bits.Add64(^v.hi, 0, carry)
	return Int128{hi: hi, lo: lo}
}

// Abs returns the absolute value of v. MinInt128 is returned
// unchanged, having no positive counterpart.
func (v Int128) Abs() Int128 {
	if v.IsNegative() {
		return v.Neg()
	}
	return v
}

// Add returns v+w, wrapping on overflow.
func (v Int128) Add(w Int128) Int128 {
	lo, carry := bits.Add64(v.lo, w.lo, 0)
	hi, _ := bits.Add64(v.hi, w.hi, carry)
	return Int128{hi: hi, lo: lo}
}

// Sub returns v-w, wrapping on overflow.
func (v Int128) Sub(w Int128) Int128 {
	lo, borrow := bits.Sub64(v.lo, w.lo, 0)
	hi, _ := bits.Sub64(v.hi, w.hi, borrow)
	return Int128{hi: hi, lo: lo}
}

// Mul returns the low 128 bits of v*w, wrapping on overflow.
func (v Int128) Mul(w Int128) Int128 {
	// two's-complement multiplication yields the same low 128 bits
	// regardless of sign, so the unsigned product is reused directly.
	return Int128(v.bits().Mul(w.bits()))
}

// Div returns v/w, truncated towards zero. It panics when w is zero.
func (v Int128) Div(w Int128) Int128 {
	q, _ := v.DivMod(w)
	return q
}

// Mod returns the remainder of v/w, taking the sign of v. It panics
// when w is zero.
func (v Int128) Mod(w Int128) Int128 {
	_, r := v.DivMod(w)
	return r
}

// DivMod returns the quotient and remainder of v/w. The quotient is
// truncated towards zero and the remainder takes the sign of v, so
// that v == q*w + r with |r| < |w|. It panics when w is zero.
func (v Int128) DivMod(w Int128) (q, r Int128) {
	neg := v.IsNegative() != w.IsNegative()
	uq, ur := v.Abs().bits().DivMod(w.Abs().bits())
	q, r = Int128(uq), Int128(ur)
	if neg {
		q = q.Neg()
	}
	if v.IsNegative() {
		r = r.Neg()
	}
	return q, r
}

// MulDivMod returns the quotient and remainder of v*w/d, forming the
// product in a 256-bit intermediate so it cannot overflow before the
// division. The quotient is truncated towards zero and wraps if it
// exceeds the 128-bit range. The remainder takes the sign of the
// product v*w with |r| < |d|, so v*w == q*d + r whenever the quotient
// does not wrap. It panics with [ErrDivZero] when d is zero.
func (v Int128) MulDivMod(w, d Int128) (q, r Int128) {
	prodNeg := v.IsNegative() != w.IsNegative()
	uq, ur := v.Abs().bits().MulDivMod(w.Abs().bits(), d.Abs().bits())
	q, r = Int128(uq), Int128(ur)
	if prodNeg {
		r = r.Neg()
	}
	if prodNeg != d.IsNegative() {
		q = q.Neg()
	}
	return q, r
}

// Cmp returns -1, 0 or +1 as v is less than, equal to or greater
// than w.
func (v Int128) Cmp(w Int128) int {
	// the high word carries the sign, so it must be compared as
	// signed; the low word remains an unsigned magnitude.
	sv, sw := int64(v.hi), int64(w.hi)
	switch {
	case sv > sw:
		return 1
	case sv < sw:
		return -1
	case v.lo > w.lo:
		return 1
	case v.lo < w.lo:
		return -1
	default:
		return 0
	}
}

// String returns v in base 10, with a leading minus sign when negative.
func (v Int128) String() string {
	return string(v.appendText(nil))
}

// MarshalText renders v as its signed base-10 digits.
func (v Int128) MarshalText() ([]byte, error) {
	return v.appendText(nil), nil
}

// UnmarshalText parses a signed base-10 integer from b into v. It rejects
// malformed input with [ErrSyntax] and values outside the Int128 range
// with [ErrRange].
func (v *Int128) UnmarshalText(b []byte) error {
	x, err := parseInt128(string(b))
	switch {
	case err != nil:
		return err
	case v == nil:
		return core.ErrNilReceiver
	default:
		*v = x
		return nil
	}
}

// appendText writes v's signed base-10 digits to dst. The magnitude is
// taken as an unsigned 128-bit value, so MinInt128 renders as 2^127
// rather than wrapping.
func (v Int128) appendText(dst []byte) []byte {
	if v.IsNegative() {
		dst = append(dst, '-')
	}
	return v.Abs().bits().appendText(dst)
}
