package num

import (
	"fmt"
	"strings"
)

// DecimalScaler is the scale parameter of [Decimal]: it yields a
// fixed-point resolution, the number of sub-units in one whole unit, as
// a value of the backing integer type T. The unexported methods serve
// the text forms and are met only by the scalers of this package, so
// Milli32, Milli64 and Atto128 are the only instantiations.
type DecimalScaler[T any] interface {
	Scale() T

	// name returns the instantiation's type name, which its
	// constructors carry as a suffix.
	name() string
	// asInt64 returns a backing value as a native int64 and whether
	// it fits.
	asInt64(v T) (int64, bool)
	// asInt128 returns a backing value widened to an Int128.
	asInt128(v T) Int128
	// doAppendText writes a backing value in decimal to dst and returns
	// the extended buffer.
	doAppendText(dst []byte, v T) []byte
}

// Decimal is a signed fixed-point number: a count of sub-units at a
// metric resolution, backed by one of the signed integer types. The
// backing type T fixes the width and the scale type S fixes the
// resolution, so Milli32, Milli64 and Atto128 are all instantiations of
// it. The zero value is numeric zero.
//
// The operations that do not move the point, Add, Sub, Neg, Abs and the
// comparisons, delegate straight to the backing integer. Mul and Div
// carry the scale through MulDivMod so the intermediate product cannot
// overflow before the scale is applied; this caps the family at the
// Int128 backing, whose product fits a 256-bit intermediate.
type Decimal[T SignedEuclidean[T], S DecimalScaler[T]] struct {
	v T
}

// newDecimal builds a Decimal from a whole-unit count and a sub-unit
// fraction. The magnitudes combine as |whole|*scale + |frac|, with the
// sign taken from whole, or from frac when whole is zero. frac need not
// stay below one whole unit: it carries. The combined magnitude wraps if
// it exceeds the backing width.
func newDecimal[T SignedEuclidean[T], S DecimalScaler[T]](whole,
	frac T) Decimal[T, S] {
	var s S
	neg := whole.IsNegative() || (whole.IsZero() && frac.IsNegative())
	mag := whole.Abs().Mul(s.Scale()).Add(frac.Abs())
	if neg {
		mag = mag.Neg()
	}
	return Decimal[T, S]{mag}
}

// one returns the multiplicative unit, one whole at the resolution:
// the step between consecutive DivMod quotients.
func (Decimal[T, S]) one() Decimal[T, S] {
	var s S
	return Decimal[T, S]{s.Scale()}
}

// ulp returns the smallest positive value, one sub-unit at the
// resolution: the step between consecutive MulDivMod quotients.
func (Decimal[T, S]) ulp() Decimal[T, S] {
	var z T
	return Decimal[T, S]{z.ulp()}
}

// one and ulp are reached only through the [Euclidean] interface, as w.one()
// and d.ulp() inside EuclideanDivMod and EuclideanMulDivMod, which
// staticcheck's unused checker cannot follow from a type-parameter
// constraint back to the generic method, so it reports them unused even at
// full coverage. Naming them on a concrete instantiation marks them used;
// a var initialiser is not an instrumented statement, so it costs no
// coverage.
var _ = Atto128{}.one().Add(Atto128{}.ulp())

// wide returns d as its count at its resolution, on the way to another
// type of the family.
func (d Decimal[T, S]) wide() wide {
	var s S
	return wide{v: s.asInt128(d.v), scale: s.asInt128(s.Scale()), ok: true}
}

// count returns the backing count of d read at unit scale, so the
// narrowing of wide checks its size without rescaling it.
func (d Decimal[T, S]) count() wide {
	var s S
	return wide{v: s.asInt128(d.v), scale: unitScale128, ok: true}
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

// AsInt128 returns the count of d, its backing integer, as an Int128,
// which always fits: the inverse of the As constructor.
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
// New form over the whole-unit and sub-unit counts while the whole
// count fits an int64, and the As form over the backing integer's own
// %#v otherwise. Both counts carry the sign of d, so the call
// rebuilds it whichever of them the constructor reads the sign from,
// and a fraction of four digits or more is grouped in thousands.
func (d Decimal[T, S]) GoString() string {
	var s S
	whole, frac := d.parts()
	w, ok := s.asInt64(whole)
	if !ok {
		return fmt.Sprintf("num.As%s(%#v)", s.name(), d.v)
	}
	// frac is below the scale, which fits an int64 at every resolution.
	f, _ := s.asInt64(frac)
	return fmt.Sprintf("num.New%s(%d, %s)", s.name(), w, groupDigits(f))
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
		writeBadVerb(s, verb, "num."+sc.name(), func() {
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

// fracWidth returns the fraction digits of the resolution, one fewer
// than the digits of the scale.
func (Decimal[T, S]) fracWidth() int {
	var sc S
	// every scale is a power of ten below 2^63, so it fits an int64.
	scale, _ := sc.asInt64(sc.Scale())
	width := 0
	for scale > 1 {
		scale /= 10
		width++
	}
	return width
}

// appendFixed appends the magnitude of d with prec fraction digits,
// zero-filled past the resolution and rounded half away from zero
// below it, with a carry out of the fraction reaching the whole count.
// The parts are taken as magnitudes one at a time, since the whole
// count is always far from the backing's minimum even when d is not.
func (d Decimal[T, S]) appendFixed(dst []byte, prec int) []byte {
	var sc S
	whole, frac := d.parts()
	whole = whole.Abs()
	// the remainder is below the scale, so it fits an int64.
	f, _ := sc.asInt64(frac)
	if f < 0 {
		f = -f
	}
	width := d.fracWidth()
	if prec < width {
		unit := pow10(width - prec)
		q, r := f/unit, f%unit
		if 2*r >= unit {
			q++
		}
		if q == pow10(prec) {
			q, whole = 0, whole.Add(whole.one())
		}
		f, width = q, prec
	}
	dst = sc.doAppendText(dst, whole)
	if prec == 0 {
		return dst
	}
	dst = append(dst, '.')
	dst = appendPadded(dst, uint64(f), width)
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
