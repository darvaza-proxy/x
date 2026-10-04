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
// conversion widens its receiver into one, rescales it to the target's
// resolution and narrows it into the target, so the seven methods of
// each type share one path.
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

// narrow32 returns the count as an Int32, keeping the low 32 bits
// when it does not fit.
func (w wide) narrow32() (Int32, bool) {
	x := Int32(w.v.lo)
	return x, w.ok && x.asInt128() == w.v
}

// narrow64 returns the count as an Int64, keeping the low 64 bits
// when it does not fit.
func (w wide) narrow64() (Int64, bool) {
	x, ok := w.v.asInt64()
	return Int64(x), w.ok && ok
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
	r := w.at(unitExp)
	return r.v, r.ok
}

// uint128 returns the bits of the whole units as a Uint128, which fit
// when the count is not negative.
func (w wide) uint128() (Uint128, bool) {
	r := w.at(unitExp)
	return r.v.bits(), r.ok && !r.v.IsNegative()
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
	r := w.at(attoExp)
	return AsAtto128(r.v), r.ok
}
