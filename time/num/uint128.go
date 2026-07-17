package num

import "math/bits"

var _ Unsigned[Uint128] = Uint128{}

// Uint128 is an unsigned 128-bit integer stored as a high and low
// 64-bit word. The zero value is numeric zero.
type Uint128 struct {
	hi, lo uint64
}

// NewUint128 assembles a Uint128 from its high and low 64-bit words,
// so the value is hi*2^64 + lo.
func NewUint128(hi, lo uint64) Uint128 {
	return Uint128{hi: hi, lo: lo}
}

// IsZero reports whether u is zero.
func (u Uint128) IsZero() bool {
	return u.hi == 0 && u.lo == 0
}

// Equal reports whether u and v are equal.
func (u Uint128) Equal(v Uint128) bool {
	return u == v
}

// Add returns u+v, wrapping on overflow.
func (u Uint128) Add(v Uint128) Uint128 {
	lo, carry := bits.Add64(u.lo, v.lo, 0)
	hi, _ := bits.Add64(u.hi, v.hi, carry)
	return Uint128{hi: hi, lo: lo}
}

// Sub returns u-v, wrapping on underflow.
func (u Uint128) Sub(v Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v.lo, 0)
	hi, _ := bits.Sub64(u.hi, v.hi, borrow)
	return Uint128{hi: hi, lo: lo}
}

// Mul returns the low 128 bits of the product, wrapping on overflow.
func (u Uint128) Mul(v Uint128) Uint128 {
	hi, lo := bits.Mul64(u.lo, v.lo)
	hi += u.hi*v.lo + u.lo*v.hi
	return Uint128{hi: hi, lo: lo}
}

// Div returns u/v, truncated. It panics when v is zero.
func (u Uint128) Div(v Uint128) Uint128 {
	q, _ := u.DivMod(v)
	return q
}

// Mod returns the remainder of u/v. It panics when v is zero.
func (u Uint128) Mod(v Uint128) Uint128 {
	_, r := u.DivMod(v)
	return r
}

// DivMod returns the quotient and remainder of u/v, so that
// u == q*v + r with r < v. It panics when v is zero.
func (u Uint128) DivMod(v Uint128) (q, r Uint128) {
	if v.IsZero() {
		panic(ErrDivZero)
	}
	// promote to a 256-bit numerator with a zero high half and reuse
	// the shift-subtract divider; the quotient fits in 128 bits.
	return u256{lo: u}.divMod128(v)
}

// MulDivMod returns the quotient and remainder of u*v/d, forming the
// product in a 256-bit intermediate so it cannot overflow before the
// division. The quotient wraps if it exceeds the 128-bit range and the
// remainder is always less than d. It panics with [ErrDivZero] when d
// is zero.
func (u Uint128) MulDivMod(v, d Uint128) (q, r Uint128) {
	if d.IsZero() {
		panic(ErrDivZero)
	}
	return mul256(u, v).divMod128(d)
}

// Cmp returns -1, 0 or +1 as u is less than, equal to or greater
// than v.
func (u Uint128) Cmp(v Uint128) int {
	switch {
	case u.hi > v.hi:
		return 1
	case u.hi < v.hi:
		return -1
	case u.lo > v.lo:
		return 1
	case u.lo < v.lo:
		return -1
	default:
		return 0
	}
}

// shl1 returns u shifted left by one bit with in as the new bit 0.
func (u Uint128) shl1(in uint64) Uint128 {
	return Uint128{
		hi: u.hi<<1 | u.lo>>63,
		lo: u.lo<<1 | in&1,
	}
}

// setBit returns u with bit i set when i is within the low 128 bits;
// higher bits are dropped, matching the wrapping policy of Add and Mul.
func (u Uint128) setBit(i int) Uint128 {
	switch {
	case i >= 128:
		return u
	case i >= 64:
		u.hi |= 1 << (i - 64)
	default:
		u.lo |= 1 << i
	}
	return u
}
