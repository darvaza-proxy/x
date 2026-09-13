package num

// The scale factors as Int128 counts, the resolutions a conversion
// moves between.
var (
	unitScale128  = Int128{lo: 1}
	milliScale128 = Int128{lo: milliScale}
	attoScale128  = Int128{lo: attoScale}
)

// wide is a value in transit between two types of the family: its
// count of sub-units widened to an Int128, the scale of that count,
// and whether the value has fitted so far. Every conversion widens
// its receiver into one, rescales it to the target's resolution and
// narrows it into the target, so the seven methods of each type share
// one path.
type wide struct {
	v     Int128
	scale Int128
	ok    bool
}

// at returns w rescaled to scale. Moving to a finer resolution
// multiplies the count, and the value stops fitting when the product
// wraps; moving to a coarser one divides it, dropping the fraction
// digits below the resolution towards zero, which always fits.
func (w wide) at(scale Int128) wide {
	switch c := scale.Cmp(w.scale); {
	case c > 0:
		ratio := scale.Div(w.scale)
		v := w.v.Mul(ratio)
		// the product wrapped if and only if dividing it back misses
		// w.v: a wrap moves it by a multiple of 2^128, which the ratio,
		// far smaller, cannot divide back into a step below one.
		w.ok = w.ok && v.Div(ratio).Equal(w.v)
		w.v = v
	case c < 0:
		w.v = w.v.Div(w.scale.Div(scale))
	default:
		// the same resolution, nothing to move.
	}
	w.scale = scale
	return w
}

// narrow32 returns the count as an Int32, keeping the low 32 bits
// when it does not fit.
func (w wide) narrow32() (Int32, bool) {
	x := Int32(w.v.lo)
	return x, w.ok && AsInt128(int64(x)) == w.v
}

// narrow64 returns the count as an Int64, keeping the low 64 bits
// when it does not fit.
func (w wide) narrow64() (Int64, bool) {
	x, ok := w.v.asInt64()
	return Int64(x), w.ok && ok
}

// int32 returns the whole units as an Int32.
func (w wide) int32() (Int32, bool) {
	return w.at(unitScale128).narrow32()
}

// int64 returns the whole units as an Int64.
func (w wide) int64() (Int64, bool) {
	return w.at(unitScale128).narrow64()
}

// int128 returns the whole units as an Int128.
func (w wide) int128() (Int128, bool) {
	r := w.at(unitScale128)
	return r.v, r.ok
}

// uint128 returns the bits of the whole units as a Uint128, which fit
// when the count is not negative.
func (w wide) uint128() (Uint128, bool) {
	r := w.at(unitScale128)
	return r.v.bits(), r.ok && !r.v.IsNegative()
}

// milli32 returns the count at milli resolution as a Milli32.
func (w wide) milli32() (Milli32, bool) {
	c, ok := w.at(milliScale128).narrow32()
	return AsMilli32(c), ok
}

// milli64 returns the count at milli resolution as a Milli64.
func (w wide) milli64() (Milli64, bool) {
	c, ok := w.at(milliScale128).narrow64()
	return AsMilli64(c), ok
}

// atto128 returns the count at atto resolution as an Atto128.
func (w wide) atto128() (Atto128, bool) {
	r := w.at(attoScale128)
	return AsAtto128(r.v), r.ok
}
