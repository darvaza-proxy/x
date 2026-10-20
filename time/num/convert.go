package num

// The resolutions a conversion moves between, as the exponents of
// their scales.
const (
	unitExp  = 0
	milliExp = 3
	attoExp  = 18
)

// wide is a value in transit between two types of the family: its
// count of sub-units widened to an Int128, the exponent of the scale
// of that count, and whether the value has fitted so far. Every
// conversion widens its receiver into one and narrows it into the
// target, so the seven methods of each type share one path. On the way
// the count is rescaled to the target's resolution, except between two
// integers, both counted in whole units already.
type wide struct {
	v   Int128
	exp int
	ok  bool
}

// at returns w rescaled to the resolution of exponent exp. Moving to a
// finer resolution multiplies the count, and the value stops fitting
// when its magnitude passes the bound pow10Bound gives, where the
// product wraps; moving to a coarser one divides it, dropping the
// fraction digits below the resolution towards zero, which always
// fits.
func (w wide) at(exp int) wide {
	switch {
	case exp > w.exp:
		k := exp - w.exp
		w.ok = w.ok && w.v.Abs().bits().Cmp(pow10Bound[k]) <= 0
		w.v = w.v.Mul(Int128(Pow10(k)))
	case exp < w.exp:
		w.v = w.v.Div(Int128(Pow10(w.exp - exp)))
	default:
		// the same resolution, nothing to move.
	}
	w.exp = exp
	return w
}

// narrow32 returns the count as an Int32 and whether the value fits:
// it must have fitted so far, and the count must fit 32 bits. The low
// 32 bits stay when it does not.
func (w wide) narrow32() (Int32, bool) {
	x, ok := w.v.asInt64()
	return Int32(x), w.ok && ok && int64(int32(x)) == x
}

// narrow64 returns the count as an Int64 and whether the value fits:
// it must have fitted so far, and the count must fit 64 bits. The low
// 64 bits stay when it does not.
func (w wide) narrow64() (Int64, bool) {
	x, ok := w.v.asInt64()
	return Int64(x), w.ok && ok
}

// narrow128 returns the count as an Int128, which holds it whole, and
// whether the value has fitted so far.
func (w wide) narrow128() (Int128, bool) {
	return w.v, w.ok
}

// narrowUint128 returns the bits of the count as a Uint128 and whether
// the value fits: it must have fitted so far, and the count must not
// be negative.
func (w wide) narrowUint128() (Uint128, bool) {
	return w.v.bits(), w.ok && !w.v.IsNegative()
}

// int32 returns the whole units as an Int32.
func (w wide) int32() (Int32, bool) {
	return w.at(unitExp).narrow32()
}

// int64 returns the whole units as an Int64.
func (w wide) int64() (Int64, bool) {
	return w.at(unitExp).narrow64()
}

// int128 returns the whole units as an Int128.
func (w wide) int128() (Int128, bool) {
	return w.at(unitExp).narrow128()
}

// uint128 returns the bits of the whole units as a Uint128.
func (w wide) uint128() (Uint128, bool) {
	return w.at(unitExp).narrowUint128()
}

// milli32 returns the count at milli resolution as a Milli32.
func (w wide) milli32() (Milli32, bool) {
	c, ok := w.at(milliExp).narrow32()
	return AsMilli32(c), ok
}

// milli64 returns the count at milli resolution as a Milli64.
func (w wide) milli64() (Milli64, bool) {
	c, ok := w.at(milliExp).narrow64()
	return AsMilli64(c), ok
}

// atto128 returns the count at atto resolution as an Atto128.
func (w wide) atto128() (Atto128, bool) {
	c, ok := w.at(attoExp).narrow128()
	return AsAtto128(c), ok
}
