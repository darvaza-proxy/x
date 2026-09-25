package num

import (
	"fmt"
	"math/bits"
	"strings"
)

// Decimal is a signed fixed-point number: a count of sub-units at a
// metric resolution, backed by a signed integer. The backing type T
// fixes the width and the scale type S fixes the resolution, so
// Milli32, Milli64 and Atto128 are all instantiations of it, and
// another package instantiates it over a backing and a scaler of its
// own. Decimal reads an outside backing through DivMod by two, one bit
// at a time into an Int128, and sizes the fraction by the scale, so
// its conversions and text read an integer backing of up to 128 bits
// over a power-of-ten scale. The zero value is numeric zero.
//
// The operations that do not move the point, Add, Sub, Neg, Abs and the
// comparisons, delegate straight to the backing integer. Mul and Div
// carry the scale through the backing's MulDivMod so the intermediate
// product cannot overflow before the scale is applied.
type Decimal[T Signed[T], S DecimalScaler[T]] struct {
	v T
}

// NewDecimal builds a Decimal from a whole-unit count and a sub-unit
// fraction. The magnitudes combine as |whole|*scale + |frac|, with the
// sign taken from whole, or from frac when whole is zero. frac need not
// stay below one whole unit: it carries. The combined magnitude wraps
// if it exceeds the backing width.
func NewDecimal[T Signed[T], S DecimalScaler[T]](whole,
	frac T) Decimal[T, S] {
	var s S
	neg := whole.IsNegative() || (whole.IsZero() && frac.IsNegative())
	mag := whole.Abs().Mul(s.Scale()).Add(frac.Abs())
	if neg {
		mag = mag.Neg()
	}
	return Decimal[T, S]{mag}
}

// AsDecimal takes a backing integer as a count of sub-units at the
// resolution of S.
func AsDecimal[T Signed[T], S DecimalScaler[T]](count T) Decimal[T, S] {
	return Decimal[T, S]{count}
}

// One returns the multiplicative unit, one whole at the resolution:
// the step between consecutive DivMod quotients.
func (Decimal[T, S]) One() Decimal[T, S] {
	var s S
	return Decimal[T, S]{s.Scale()}
}

// ULP returns the unit in the last place, the smallest positive
// value, one sub-unit at the resolution: the step between
// consecutive MulDivMod quotients.
func (Decimal[T, S]) ULP() Decimal[T, S] {
	var ulp T
	// The package's own backings answer through their concrete ULP,
	// sparing the division; any other takes the scale divided by
	// itself, one for any integer backing.
	switch p := any(&ulp).(type) {
	case *Int32:
		*p = p.ULP()
	case *Int64:
		*p = p.ULP()
	case *Int128:
		*p = p.ULP()
	default:
		var s S
		ulp = s.Scale().Div(s.Scale())
	}
	return Decimal[T, S]{ulp}
}

// wide returns d as its count at its resolution, on the way to another
// type of the family.
func (d Decimal[T, S]) wide() wide {
	var s S
	return wide{v: d.widen(d.v), scale: d.widen(s.Scale()), ok: true}
}

// count returns the backing count of d read at unit scale, so the
// narrowing of wide checks its size without rescaling it.
func (d Decimal[T, S]) count() wide {
	return wide{v: d.widen(d.v), scale: unitScale128, ok: true}
}

// widen returns a backing value as an Int128.
func (Decimal[T, S]) widen(v T) Int128 {
	switch p := any(&v).(type) {
	case *Int32:
		return p.asInt128()
	case *Int64:
		return p.asInt128()
	case *Int128:
		return *p
	default:
		return widenBits(v)
	}
}

// widenBits returns v as an Int128 read one bit at a time by
// truncating division by two, the two built from v divided by itself,
// which is one for any integer but zero. The remainder takes the sign
// of v, so the minimum reads without wrapping. The reading stops at the
// 128th bit, so bits past it are dropped unread.
func widenBits[T Signed[T]](v T) Int128 {
	if v.IsZero() {
		return ZeroInt128
	}
	one := v.Div(v)
	two := one.Add(one)
	neg := v.IsNegative()
	var mag Uint128
	for i := 0; i < 128 && !v.IsZero(); i++ {
		q, r := v.DivMod(two)
		if !r.IsZero() {
			mag = mag.setBit(i)
		}
		v = q
	}
	if neg {
		return Int128(mag).Neg()
	}
	return Int128(mag)
}

// AsInt32 returns the count of d, its backing integer, as an Int32 and
// whether it fits, the low 32 bits kept when it does not: the inverse
// of the As constructor.
//
//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (d Decimal[T, S]) AsInt32() (Int32, bool) {
	return d.count().narrow32()
}

// AsInt64 returns the count of d, its backing integer, as an Int64 and
// whether it fits, the low 64 bits kept when it does not: the inverse
// of the As constructor.
//
//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (d Decimal[T, S]) AsInt64() (Int64, bool) {
	return d.count().narrow64()
}

// AsInt128 returns the count of d, its backing integer, as an Int128
// and a true flag: the inverse of the As constructor. A backing wider
// than 128 bits keeps its low 128 bits, as Decimal reads it.
//
//revive:disable-next-line:confusing-naming two-parameter receiver misfiled as a function
func (d Decimal[T, S]) AsInt128() (Int128, bool) {
	w := d.count()
	return w.v, w.ok
}

// Int32 returns the whole units of d as an Int32, the fraction dropped
// towards zero, and whether they fit; the low 32 bits stay when they
// do not.
func (d Decimal[T, S]) Int32() (Int32, bool) {
	return d.wide().int32()
}

// Int64 returns the whole units of d as an Int64, the fraction dropped
// towards zero, and whether they fit; the low 64 bits stay when they
// do not.
func (d Decimal[T, S]) Int64() (Int64, bool) {
	return d.wide().int64()
}

// Int128 returns the whole units of d as an Int128, the fraction
// dropped towards zero, which always fit.
func (d Decimal[T, S]) Int128() (Int128, bool) {
	return d.wide().int128()
}

// Uint128 returns the whole units of d as a Uint128, the fraction
// dropped towards zero, and whether they fit, which they do when not
// negative; the bit pattern stays when they do not.
func (d Decimal[T, S]) Uint128() (Uint128, bool) {
	return d.wide().uint128()
}

// Milli32 returns d at milli resolution as a Milli32, the digits below
// it dropped towards zero, and whether the value fits; the low 32 bits
// of the count stay when it does not.
func (d Decimal[T, S]) Milli32() (Milli32, bool) {
	return d.wide().milli32()
}

// Milli64 returns d at milli resolution as a Milli64, the digits below
// it dropped towards zero, and whether the value fits; the low 64 bits
// of the count stay when it does not.
func (d Decimal[T, S]) Milli64() (Milli64, bool) {
	return d.wide().milli64()
}

// Atto128 returns d at atto resolution as an Atto128 and whether the
// value fits; the low 128 bits of the count stay when it does not.
func (d Decimal[T, S]) Atto128() (Atto128, bool) {
	return d.wide().atto128()
}

// parts splits d into its whole-unit count and the sub-unit remainder,
// both in the backing type and both carrying the sign of d, so that
// whole*scale + frac == d.
func (d Decimal[T, S]) parts() (whole, frac T) {
	var s S
	return d.v.DivMod(s.Scale())
}

// GoString returns the constructor call that rebuilds d for %#v: the
// New form over the whole-unit and sub-unit counts while both fit an
// int64, and the As form over the backing integer's own %#v otherwise.
// Both counts carry the sign of d, so the call rebuilds it whichever
// of them the constructor reads the sign from, and a fraction of four
// digits or more is grouped in thousands.
func (d Decimal[T, S]) GoString() string {
	var s S
	whole, frac := d.parts()
	w, wok := d.widen(whole).asInt64()
	f, fok := d.widen(frac).asInt64()
	if !wok || !fok {
		return fmt.Sprintf("%s(%#v)", constructorName(s.Name(), "As"), d.v)
	}
	return fmt.Sprintf("%s(%d, %s)", constructorName(s.Name(), "New"), w,
		groupThousands(f))
}

// String returns d at full resolution, 1.500 for a Milli32, the text
// %v prints.
func (d Decimal[T, S]) String() string {
	return string(d.doAppendText(nil))
}

// AppendText implements [encoding.TextAppender], appending d at full
// resolution, the text %v prints, to b. It allocates only when b lacks
// the room, and the error is always nil.
func (d Decimal[T, S]) AppendText(b []byte) ([]byte, error) {
	return d.doAppendText(b), nil
}

// MarshalText implements [encoding.TextMarshaler], returning the
// AppendText text.
func (d Decimal[T, S]) MarshalText() ([]byte, error) {
	return d.AppendText(nil)
}

// MarshalJSON implements [json.Marshaler], returning the MarshalText
// text as a JSON number while the count and the scale both sit below
// 10^15, safe for a float64 consumer, and as a JSON string otherwise:
// a Milli32 always a number, a Milli64 a number below a count of
// 10^15, an Atto128 always a string.
func (d Decimal[T, S]) MarshalJSON() ([]byte, error) {
	var s S
	count, ok := d.widen(d.v).asInt64()
	scale, sok := d.widen(s.Scale()).asInt64()
	if ok && sok && isJSONSafeDecimal(count, scale) {
		return d.MarshalText()
	}
	return jsonString(d)
}

// doAppendText writes d at full resolution to dst, the sign before the
// magnitude, and returns the extended buffer.
func (d Decimal[T, S]) doAppendText(dst []byte) []byte {
	if d.IsNegative() {
		dst = append(dst, '-')
	}
	return d.appendFixed(dst, d.fracWidth())
}

// Format implements [fmt.Formatter] with the verbs v and s printing the
// value at full resolution, 1.500 for a Milli32, every digit the
// resolution holds and nothing more whatever precision is asked for,
// and f, or F under another name, printing it with the fraction digits
// the precision asks for, six without one, as fmt does for its floats:
// zero-filled past the resolution, and below it rounded half away from
// zero, where a float64 rounds half to even. The ' ', '-' and '0' flags
// and the width apply as for a float, '#' keeps the point of a zero
// precision, and '+' signs the value everywhere but under v, where fmt
// gives it the struct-field meaning; %#v prints the GoString form. Any
// other verb prints as %!verb(type=value).
func (d Decimal[T, S]) Format(s fmt.State, verb rune) {
	if verb == 'F' {
		verb = 'f'
	}
	switch verb {
	case 'v', 's', 'f':
		if verb == 'v' && s.Flag('#') {
			writeGoString(s, d.GoString())
			return
		}
		d.writeFixed(s, verb)
	default:
		var sc S
		// d prints the value as v does, at full resolution, but reads
		// the flags as a bad verb leaves them, so '+' signs there.
		writeBadVerb(s, verb, sc.Name(), func() {
			d.writeFixed(s, 'd')
		})
	}
}

// writeFixed writes d's sign, digits and padding to s under the flags,
// width and precision verb was given.
func (d Decimal[T, S]) writeFixed(s fmt.State, verb rune) {
	f := numField{
		digits: d.appendFormatted(s, verb),
		sign:   formatSign(s, verb, d.IsNegative()),
	}
	f.zeros = padZeros(s, len(f.sign), len(f.digits))
	f.writeTo(s)
}

// appendFormatted returns the magnitude of d with the fraction digits
// verb asks for, keeping the point of a zero precision under '#' as
// fmt does for a float.
func (d Decimal[T, S]) appendFormatted(s fmt.State, verb rune) []byte {
	prec := d.precision(s, verb)
	dst := d.appendFixed(nil, prec)
	if prec == 0 && s.Flag('#') {
		dst = append(dst, '.')
	}
	return dst
}

// precision returns the fraction digits verb prints: the resolution
// for v and s, and for f what the state asks, or six.
func (d Decimal[T, S]) precision(s fmt.State, verb rune) int {
	if verb != 'f' {
		return d.fracWidth()
	}
	if p, ok := s.Precision(); ok {
		return p
	}
	return 6
}

// fracWidth returns the fraction digits of the resolution, the exponent
// of the scale. Every power of ten below 2^128 has a bit length of its
// own, and scaling that length by 1233/4096, a hair below log10(2),
// gives the exponent exactly for each of them.
func (d Decimal[T, S]) fracWidth() int {
	var sc S
	return d.widen(sc.Scale()).bits().bitLen() * 1233 >> 12
}

// appendFixed appends the magnitude of d with prec fraction digits,
// zero-filled past the resolution and rounded half away from zero
// below it, with a carry out of the fraction reaching the whole count.
// The parts are taken as magnitudes one at a time, each as an unsigned
// 128-bit value, which keeps even the backing's minimum where Abs
// would wrap and holds a fraction of any width the scale gives.
func (d Decimal[T, S]) appendFixed(dst []byte, prec int) []byte {
	whole, frac := d.parts()
	mag := d.widen(whole).Abs().bits()
	f := d.widen(frac).Abs().bits()
	width := d.fracWidth()
	if prec < width {
		unit := pow10(width - prec)
		q, r := f.DivMod(unit)
		if r.Add(r).Cmp(unit) >= 0 {
			q = q.Add(q.One())
		}
		if q.Equal(pow10(prec)) {
			q, mag = ZeroUint128, mag.Add(mag.One())
		}
		f, width = q, prec
	}
	dst = mag.doAppendText(dst)
	if prec == 0 {
		return dst
	}
	dst = append(dst, '.')
	// the fraction is below 2^127, so its top word is below decGroup and
	// one word division splits it at the nineteenth digit from the right.
	// Up to nineteen digits wide the upper part is zero and takes no room.
	hi, lo := bits.Div64(f.hi, f.lo, decGroup)
	dst = appendPadded(dst, hi, width-decGroupDigits)
	dst = appendPadded(dst, lo, min(width, decGroupDigits))
	return append(dst, strings.Repeat("0", prec-width)...)
}

// IsZero reports whether d is zero.
func (d Decimal[T, S]) IsZero() bool {
	return d.v.IsZero()
}

// Equal reports whether d and v represent the same value.
func (d Decimal[T, S]) Equal(v Decimal[T, S]) bool {
	return d.v.Equal(v.v)
}

// Cmp returns -1, 0 or +1 as d is less than, equal to or greater
// than v.
func (d Decimal[T, S]) Cmp(v Decimal[T, S]) int {
	return d.v.Cmp(v.v)
}

// IsNegative reports whether d is less than zero.
func (d Decimal[T, S]) IsNegative() bool {
	return d.v.IsNegative()
}

// Neg returns -d.
func (d Decimal[T, S]) Neg() Decimal[T, S] {
	return Decimal[T, S]{d.v.Neg()}
}

// Abs returns the absolute value of d.
func (d Decimal[T, S]) Abs() Decimal[T, S] {
	return Decimal[T, S]{d.v.Abs()}
}

// Add returns d+v, wrapping on overflow.
func (d Decimal[T, S]) Add(v Decimal[T, S]) Decimal[T, S] {
	return Decimal[T, S]{d.v.Add(v.v)}
}

// Sub returns d-v, wrapping on overflow.
func (d Decimal[T, S]) Sub(v Decimal[T, S]) Decimal[T, S] {
	return Decimal[T, S]{d.v.Sub(v.v)}
}

// Mul returns the fixed-point product, dropping any fraction below the
// resolution. It wraps if the result exceeds the backing width.
func (d Decimal[T, S]) Mul(v Decimal[T, S]) Decimal[T, S] {
	var s S
	q, _ := d.v.MulDivMod(v.v, s.Scale())
	return Decimal[T, S]{q}
}

// Div returns the fixed-point ratio, truncated towards zero at the
// resolution. Unlike the integer types this is not the whole count
// DivMod returns: 4.5/2.1 is 2.142..., not 2. It panics with
// [ErrDivZero] when v is zero.
func (d Decimal[T, S]) Div(v Decimal[T, S]) Decimal[T, S] {
	var s S
	q, _ := d.v.MulDivMod(s.Scale(), v.v)
	return Decimal[T, S]{q}
}

// Mod returns the remainder of reducing d by whole multiples of v,
// taking the sign of d with the result smaller in magnitude than v. It
// panics with [ErrDivZero] when v is zero.
func (d Decimal[T, S]) Mod(v Decimal[T, S]) Decimal[T, S] {
	_, r := d.DivMod(v)
	return r
}

// DivMod splits d into whole multiples of v and a remainder, so that
// d == q*v + r with q an integer-valued Decimal and |r| < |v|. The count
// is truncated towards zero and r takes the sign of d; scaling the
// count back up to a Decimal wraps if it exceeds the backing width. It
// panics with [ErrDivZero] when v is zero.
//
// This is ordinary integer division carried to fixed point: just as 5
// DivMod 2 is (2, 1), 4.5 DivMod 2.1 is (2, 0.3). It differs from Div,
// which instead yields the fractional ratio (4.5/2.1 = 2.142...).
func (d Decimal[T, S]) DivMod(v Decimal[T, S]) (q, r Decimal[T, S]) {
	var s S
	count, rem := d.v.DivMod(v.v)
	q = Decimal[T, S]{count.Mul(s.Scale())}
	r = Decimal[T, S]{rem}
	return q, r
}

// MulDivMod returns d*v/w and the leftover remainder, forming the product
// in an intermediate wide enough that it cannot overflow before the
// division. The two scale factors of the product against the one of the
// divisor leave the quotient at the original resolution. The quotient is
// truncated towards zero and wraps if it exceeds the backing width; the
// remainder takes the sign of the product d*v with |r| < |w|. It panics
// with [ErrDivZero] when w is zero.
func (d Decimal[T, S]) MulDivMod(v, w Decimal[T, S]) (q, r Decimal[T, S]) {
	dq, dr := d.v.MulDivMod(v.v, w.v)
	return Decimal[T, S]{dq}, Decimal[T, S]{dr}
}
