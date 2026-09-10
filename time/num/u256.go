package num

import "math/bits"

// u256 is an unexported 256-bit unsigned integer providing the wide
// intermediates for 128-bit arithmetic: mul256 forms the full product
// of two Uint128 values, while divMod128 promotes a numerator to 256
// bits so the shift-subtract loop has room to work.
type u256 struct {
	hi, lo Uint128
}

// mul256 returns the full 256-bit product of two Uint128 values.
func mul256(a, b Uint128) u256 {
	ll1, ll0 := bits.Mul64(a.lo, b.lo)
	lh1, lh0 := bits.Mul64(a.lo, b.hi)
	hl1, hl0 := bits.Mul64(a.hi, b.lo)
	hh1, hh0 := bits.Mul64(a.hi, b.hi)

	// word 0 is ll0; word 1 gathers the low halves of the cross terms,
	// tracking each carry separately since a single Add64 carry is one bit.
	w1, c1a := bits.Add64(ll1, lh0, 0)
	w1, c1b := bits.Add64(w1, hl0, 0)

	// word 2 gathers the high halves of the cross terms and the low half of
	// the top product. The word-1 carries sum to 0, 1 or 2, so they are
	// folded in as an addend rather than a carry bit, which Add64 caps at one.
	w2, c2a := bits.Add64(lh1, hl1, 0)
	w2, c2b := bits.Add64(w2, hh0, 0)
	w2, c2c := bits.Add64(w2, c1a+c1b, 0)

	// word 3 takes the top product half plus the word-2 carries; it cannot
	// overflow because the full product fits in 256 bits.
	w3 := hh1 + c2a + c2b + c2c

	return u256{
		hi: Uint128{hi: w3, lo: w2},
		lo: Uint128{hi: w1, lo: ll0},
	}
}

// bitLen returns the number of bits needed to represent x, zero for
// zero.
func (x u256) bitLen() int {
	if !x.hi.IsZero() {
		return 128 + x.hi.bitLen()
	}
	return x.lo.bitLen()
}

// divMod128 divides by a 128-bit value. It returns the low 128 bits of
// the quotient (wrapping if the true quotient exceeds 128 bits) and the
// remainder, which is always less than v. The divisor must be non-zero.
// A divisor that fits one word takes the word-wise path; otherwise
// shift-subtract long division runs from the numerator's top set bit.
func (x u256) divMod128(v Uint128) (q, r Uint128) {
	if v.hi == 0 {
		return x.divMod64(v.lo)
	}
	for i := x.bitLen() - 1; i >= 0; i-- {
		// carry is the bit shifted out of r's top. When set, the true
		// partial remainder is at least 2^128, hence larger than any
		// 128-bit v, so a subtraction is due even when the truncated r
		// compares below v.
		carry := r.hi >> 63
		r = r.shl1(x.bit(i))
		if carry != 0 || r.Cmp(v) >= 0 {
			r = r.Sub(v)
			q = q.setBit(i)
		}
	}
	return q, r
}

// divMod64 divides by a single non-zero word, one bits.Div64 step per
// word from the top, each carrying its remainder into the next. The
// quotient words above the low 128 bits are dropped, matching the
// wrapping of divMod128.
func (x u256) divMod64(y uint64) (q, r Uint128) {
	_, rem := bits.Div64(0, x.hi.hi, y)
	_, rem = bits.Div64(rem, x.hi.lo, y)
	q.hi, rem = bits.Div64(rem, x.lo.hi, y)
	q.lo, rem = bits.Div64(rem, x.lo.lo, y)
	return q, Uint128{lo: rem}
}

// bit reports bit i (0 = least significant) of the 256-bit value.
func (x u256) bit(i int) uint64 {
	switch {
	case i >= 192:
		return (x.hi.hi >> (i - 192)) & 1
	case i >= 128:
		return (x.hi.lo >> (i - 128)) & 1
	case i >= 64:
		return (x.lo.hi >> (i - 64)) & 1
	default:
		return (x.lo.lo >> i) & 1
	}
}
