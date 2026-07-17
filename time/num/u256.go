package num

// u256 is an unexported 256-bit unsigned integer used as the wide
// intermediate for 128-bit division: divMod128 promotes the numerator
// to 256 bits so the shift-subtract loop has room to work.
type u256 struct {
	hi, lo Uint128
}

// divMod128 divides by a 128-bit value using shift-subtract long
// division. It returns the low 128 bits of the quotient (wrapping if
// the true quotient exceeds 128 bits) and the remainder, which is
// always less than v. The divisor must be non-zero.
func (x u256) divMod128(v Uint128) (q, r Uint128) {
	for i := 255; i >= 0; i-- {
		r = r.shl1(x.bit(i))
		if r.Cmp(v) >= 0 {
			r = r.Sub(v)
			q = q.setBit(i)
		}
	}
	return q, r
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
