package num

import (
	"fmt"
	"strconv"
)

var (
	_ Signed[Int64] = Int64(0)
	_ Number[Int64] = Int64(0)
)

// Int64 is a signed 64-bit integer wrapping the native int64,
// extending the Signed family to 64 bits with the same method surface
// as Int128. The zero value is numeric zero.
type Int64 int64

// sys returns v as its underlying native int64.
func (v Int64) sys() int64 {
	return int64(v)
}

// one returns the multiplicative unit, the [EuclideanDivMod] quotient
// step.
func (Int64) one() Int64 {
	return 1
}

// ulp returns the smallest positive value, the [EuclideanMulDivMod]
// quotient step; for an integer it equals one.
func (Int64) ulp() Int64 {
	return 1
}

// AsInt64 takes a signed 64-bit value as an Int64, the conversion.
//
//revive:disable-next-line:confusing-naming misfiled Decimal method of the same name
func AsInt64(x int64) Int64 {
	return Int64(x)
}

// wide returns v as a count of whole units, on the way to another
// type of the family.
func (v Int64) wide() wide {
	return wide{v: AsInt128(v.sys()), scale: unitScale128, ok: true}
}

// Int32 returns v as an Int32 and whether it fits; the low 32 bits
// stay when it does not.
func (v Int64) Int32() (Int32, bool) {
	return v.wide().int32()
}

// Int64 returns v unchanged, the conversion to its own type, which
// always fits.
func (v Int64) Int64() (Int64, bool) {
	return v.wide().int64()
}

// Int128 returns v as an Int128, which always fits.
func (v Int64) Int128() (Int128, bool) {
	return v.wide().int128()
}

// Uint128 returns v as a Uint128 and whether it fits, which it does
// when not negative; the bit pattern stays when it does not.
func (v Int64) Uint128() (Uint128, bool) {
	return v.wide().uint128()
}

// Milli32 returns v as whole units of a Milli32 and whether it fits;
// the low 32 bits of the milli count stay when it does not.
func (v Int64) Milli32() (Milli32, bool) {
	return v.wide().milli32()
}

// Milli64 returns v as whole units of a Milli64 and whether it fits;
// the low 64 bits of the milli count stay when it does not.
func (v Int64) Milli64() (Milli64, bool) {
	return v.wide().milli64()
}

// Atto128 returns v as whole units of an Atto128, which always fits.
func (v Int64) Atto128() (Atto128, bool) {
	return v.wide().atto128()
}

// Format implements [fmt.Formatter] with the verbs d, v and s for
// decimal, x and X for hex, o and O for octal and b for binary, handed
// to fmt over the native int64 so the flags, width and precision behave
// as they do there; %#v prints the GoString form. Any other verb prints
// as %!verb(num.Int64=value).
func (v Int64) Format(s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('#'):
		writeGoString(s, v.GoString())
	case !isFormatVerb(verb):
		writeBadVerb(s, verb, "num.Int64", func() {
			_, _ = fmt.Fprintf(s, fmt.FormatString(s, 'd'), v.sys())
		})
	default:
		_, _ = fmt.Fprintf(s, fmt.FormatString(s, nativeVerb(verb)), v.sys())
	}
}

// GoString returns the constructor that rebuilds v, num.AsInt64(-5),
// for %#v.
func (v Int64) GoString() string {
	return fmt.Sprintf("num.AsInt64(%d)", v.sys())
}

// String returns v in decimal, the text %v prints.
func (v Int64) String() string {
	return strconv.FormatInt(v.sys(), 10)
}

// AppendText implements [encoding.TextAppender], appending v in
// decimal, the text %v prints, to b. It allocates only when b lacks
// the room, and the error is always nil.
func (v Int64) AppendText(b []byte) ([]byte, error) {
	return v.doAppendText(b), nil
}

// MarshalText implements [encoding.TextMarshaler], returning the
// AppendText text.
func (v Int64) MarshalText() ([]byte, error) {
	return v.AppendText(nil)
}

// doAppendText writes v in decimal to dst and returns the extended
// buffer.
func (v Int64) doAppendText(dst []byte) []byte {
	return strconv.AppendInt(dst, v.sys(), 10)
}

// IsZero reports whether v is zero.
func (v Int64) IsZero() bool {
	return v == 0
}

// Equal reports whether v and w are equal.
func (v Int64) Equal(w Int64) bool {
	return v == w
}

// IsNegative reports whether v is less than zero.
func (v Int64) IsNegative() bool {
	return v < 0
}

// Neg returns -v, wrapping only for the most negative value, which has
// no positive counterpart.
func (v Int64) Neg() Int64 {
	return -v
}

// Abs returns the absolute value of v. The most negative value is
// returned unchanged, having no positive counterpart.
func (v Int64) Abs() Int64 {
	if v < 0 {
		return -v
	}
	return v
}

// Add returns v+w, wrapping on overflow.
func (v Int64) Add(w Int64) Int64 {
	return v + w
}

// Sub returns v-w, wrapping on overflow.
func (v Int64) Sub(w Int64) Int64 {
	return v - w
}

// Mul returns the low 64 bits of v*w, wrapping on overflow.
func (v Int64) Mul(w Int64) Int64 {
	return v * w
}

// Div returns v/w, truncated towards zero. It panics with [ErrDivZero]
// when w is zero.
func (v Int64) Div(w Int64) Int64 {
	q, _ := v.DivMod(w)
	return q
}

// Mod returns the remainder of v/w, taking the sign of v. It panics
// with [ErrDivZero] when w is zero.
func (v Int64) Mod(w Int64) Int64 {
	_, r := v.DivMod(w)
	return r
}

// DivMod returns the quotient and remainder of v/w. The quotient is
// truncated towards zero and the remainder takes the sign of v, so
// that v == q*w + r with |r| < |w|. It panics with [ErrDivZero] when w
// is zero.
func (v Int64) DivMod(w Int64) (q, r Int64) {
	if w == 0 {
		panic(ErrDivZero)
	}
	return v / w, v % w
}

// MulDivMod returns the quotient and remainder of v*w/d, forming the
// product in a 128-bit intermediate so it cannot overflow before the
// division. The quotient is truncated towards zero and wraps if it
// exceeds the 64-bit range. The remainder takes the sign of the product
// v*w with |r| < |d|, so v*w == q*d + r whenever the quotient does not
// wrap. It panics with [ErrDivZero] when d is zero.
func (v Int64) MulDivMod(w, d Int64) (q, r Int64) {
	if d == 0 {
		panic(ErrDivZero)
	}
	v128, w128, d128 := AsInt128(v.sys()), AsInt128(w.sys()), AsInt128(d.sys())
	q128, r128 := v128.MulDivMod(w128, d128)
	return Int64(q128.lo), Int64(r128.lo)
}

// Cmp returns -1, 0 or +1 as v is less than, equal to or greater
// than w.
func (v Int64) Cmp(w Int64) int {
	switch {
	case v > w:
		return 1
	case v < w:
		return -1
	default:
		return 0
	}
}
