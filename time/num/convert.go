package num

import (
	"errors"
	"math"
)

// The scale factors as Int128 counts, the resolutions a conversion
// moves between.
var (
	unitScale128  = Int128{lo: 1}
	milliScale128 = Int128{lo: milliScale}
	attoScale128  = Int128{lo: attoScale}
)

// NewFromInt128 returns whole units as a T, any type of the family: the
// way into a type that generic code names only by its type argument,
// and the transpose of the Int128 row of the conversion matrix. A value
// that fits is the one the conversion method named for T gives,
// [Int128.Int32] for an Int32 or [Int128.Milli32] for a Milli32, the
// fraction of a Decimal zero. One that does not fit at the resolution
// of T is the nearest bound, math.MaxInt32 or math.MinInt32 for an
// Int32 and zero for a negative into Uint128, with [ErrRange], the
// value strconv returns on a range failure. T is one of the seven
// types of the package; a Decimal over a scaler of another package is
// [errors.ErrUnsupported] here, and [NewDecimalFromInt128] builds it.
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
		err = errors.ErrUnsupported
	}
	return out, err
}

// NewDecimalFromInt128 returns whole units as a Decimal over T and S:
// the way into an instantiation named by its type arguments, the one a
// Decimal over a scaler of another package takes, and the constructor
// behind the Decimal arms of [NewFromInt128]. The units are scaled to
// the resolution of S and the count narrowed into T; a value past the
// bound of T at that resolution is that bound with [ErrRange], as
// NewFromInt128 has it.
func NewDecimalFromInt128[T SignedNumber[T], S DecimalScaler[T]](v Int128) (Decimal[T, S], error) {
	var s S
	w := v.wide().at(widen(s.Scale()))
	if !w.ok {
		// the count wrapped past 128 bits, so it is past T as well.
		// The Int128 bound on the side of v narrows into the bound of
		// T, with ErrRange, or is that bound already when T is Int128,
		// so the outcome is ErrRange either way.
		c, _ := NewFromInt128[T](bound(v, MinInt128, MaxInt128))
		return AsDecimal[T, S](c), ErrRange
	}
	c, err := NewFromInt128[T](w.v)
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
func bound[T Number[T]](v Int128, lo, hi T) T {
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
