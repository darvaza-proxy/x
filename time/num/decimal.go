package num

// DecimalScaler is the scale parameter of [Decimal]: it yields a
// fixed-point resolution — the number of sub-units in one whole unit — as
// a value of the backing integer type T. Milli32, Milli64 and Atto128 are
// the instantiations this package provides.
type DecimalScaler[T any] interface {
	Scale() T
}

// Decimal is a signed fixed-point number: a count of sub-units at a
// metric resolution, backed by one of the signed integer types. The
// backing type T fixes the width and the scale type S fixes the
// resolution, so Milli32, Milli64 and Atto128 are all instantiations of
// it. The zero value is numeric zero.
//
// The operations that do not move the point — Add, Sub, Neg, Abs and the
// comparisons — delegate straight to the backing integer. Mul and Div
// carry the scale through MulDivMod so the intermediate product cannot
// overflow before the scale is applied; this caps the family at the
// Int128 backing, whose product fits a U256 intermediate.
type Decimal[T Signed[T], S DecimalScaler[T]] struct {
	v T
}

// newDecimal builds a Decimal from a whole-unit count and a sub-unit
// fraction. The magnitudes combine as |whole|*scale + |frac|, with the
// sign taken from whole, or from frac when whole is zero. frac need not
// stay below one whole unit: it carries. The combined magnitude wraps if
// it exceeds the backing width.
func newDecimal[T Signed[T], S DecimalScaler[T]](whole, frac T) Decimal[T, S] {
	var s S
	neg := whole.IsNegative() || (whole.IsZero() && frac.IsNegative())
	mag := whole.Abs().Mul(s.Scale()).Add(frac.Abs())
	if neg {
		mag = mag.Neg()
	}
	return Decimal[T, S]{mag}
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
// DivMod returns: 4.5/2.1 is 2.142..., not 2. It panics when v is zero.
func (d Decimal[T, S]) Div(v Decimal[T, S]) Decimal[T, S] {
	var s S
	q, _ := d.v.MulDivMod(s.Scale(), v.v)
	return Decimal[T, S]{q}
}

// Mod returns the remainder of reducing d by whole multiples of v,
// taking the sign of d with the result smaller in magnitude than v. It
// panics when v is zero.
func (d Decimal[T, S]) Mod(v Decimal[T, S]) Decimal[T, S] {
	_, r := d.DivMod(v)
	return r
}

// DivMod splits d into whole multiples of v and a remainder, so that
// d == q*v + r with q an integer-valued Decimal and |r| < |v|. The count
// is truncated towards zero and r takes the sign of d. It panics when v
// is zero.
//
// This is ordinary integer division carried to fixed point: just as 5
// DivMod 2 is (2, 1), 4.5 DivMod 2.1 is (2, 0.3). It differs from Div,
// which instead yields the fractional ratio (4.5/2.1 = 2.142...).
func (d Decimal[T, S]) DivMod(v Decimal[T, S]) (q, r Decimal[T, S]) {
	var s S
	count := d.v.Div(v.v)
	r = Decimal[T, S]{d.v.Sub(count.Mul(v.v))}
	q = Decimal[T, S]{count.Mul(s.Scale())}
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
