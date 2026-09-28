package num

import (
	"math"

	"darvaza.org/core"
)

// The resolutions a conversion moves between, as the exponents of
// their scales.
const (
	unitExp  = 0
	milliExp = 3
	attoExp  = 18
)

// NewFromInt128 returns whole units as a T, any type of the family: the
// way into a type that generic code names only by its type argument,
// and the transpose of the Int128 row of the conversion matrix. A value
// that fits is the one the conversion method named for T gives,
// [Int128.Int32] for an Int32 or [Int128.Milli32] for a Milli32, a
// Decimal with a zero fraction. One that does not fit at the
// resolution of T is the nearest bound, math.MaxInt32 or math.MinInt32
// for an Int32 and zero for a negative into Uint128, with [ErrRange],
// the value strconv returns on a range failure. T is one of the seven
// types of the package; a Decimal over a scaler of another package is
// refused here, the zero value with [core.ErrUnsupported] naming its
// type, and [NewDecimalFromInt128] builds it.
func NewFromInt128[T Number[T]](v Int128) (T, error) {
	var out T
	var err error
	switch p := any(&out).(type) {
	case *Int32:
		*p, err = rangedInt32(v)
	case *Int64:
		*p, err = rangedInt64(v)
	case *Int128:
		*p = v
	case *Uint128:
		*p, err = rangedUint128(v)
	case *Milli32:
		*p, err = NewDecimalFromInt128[Int32, milli32Scale](v)
	case *Milli64:
		*p, err = NewDecimalFromInt128[Int64, milli64Scale](v)
	case *Atto128:
		*p, err = NewDecimalFromInt128[Int128, atto128Scale](v)
	default:
		err = core.Wrapf(core.ErrUnsupported, "NewFromInt128[%s]",
			core.TypeName[T]())
	}
	return out, err
}

// NewDecimalFromInt128 returns whole units as a Decimal over T and S:
// the way into an instantiation named by its type arguments, the one a
// Decimal over a scaler or a backing of another package takes, and the
// constructor behind the Decimal arms of [NewFromInt128]. The units are
// scaled to the resolution of S and the count narrowed into T; a value
// past the bound of T at that resolution is that bound with
// [ErrRange], as NewFromInt128 has it.
func NewDecimalFromInt128[T Signed[T], S DecimalScaler[T]](v Int128) (Decimal[T, S], error) {
	var d Decimal[T, S]
	w := v.wide().at(d.fracWidth())
	if !w.ok {
		// the count wrapped past 128 bits, so it is past T as well.
		// The Int128 bound on the side of v narrows into the bound of
		// T with ErrRange, or is that bound already when T is Int128,
		// the one answer without an error.
		c, err := d.narrow(bound(v, MinInt128, MaxInt128))
		if err == nil {
			err = ErrRange
		}
		return AsDecimal[T, S](c), err
	}
	c, err := d.narrow(w.v)
	return AsDecimal[T, S](c), err
}

// ranged returns the conversion of v into a target when it fitted, and
// the bound of the target on the side of v with ErrRange when it did
// not.
func ranged[T Number[T]](v Int128, conv func(Int128) (T, bool), lo, hi T) (T, error) {
	if x, ok := conv(v); ok {
		return x, nil
	}
	return bound(v, lo, hi), ErrRange
}

// bound returns the bound of a target on the side of v: lo when v is
// negative, hi otherwise.
func bound[T any](v Int128, lo, hi T) T {
	if v.IsNegative() {
		return lo
	}
	return hi
}

// rangedInt32 returns v as an Int32, or the nearest bound with ErrRange
// when it does not fit.
func rangedInt32(v Int128) (Int32, error) {
	return ranged(v, Int128.Int32, math.MinInt32, math.MaxInt32)
}

// rangedInt64 returns v as an Int64, or the nearest bound with ErrRange
// when it does not fit.
func rangedInt64(v Int128) (Int64, error) {
	return ranged(v, Int128.Int64, math.MinInt64, math.MaxInt64)
}

// rangedUint128 returns v as a Uint128, or zero with ErrRange when v is
// negative, the only side it can miss on.
func rangedUint128(v Int128) (Uint128, error) {
	if x, ok := v.Uint128(); ok {
		return x, nil
	}
	return ZeroUint128, ErrRange
}

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
