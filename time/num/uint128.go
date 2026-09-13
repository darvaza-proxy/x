package num

import (
	"fmt"
	"math/bits"
	"strconv"
)

var _ Number[Uint128] = Uint128{}

// Uint128 is an unsigned 128-bit integer stored as a high and low
// 64-bit word. The zero value is numeric zero.
type Uint128 struct {
	hi, lo uint64
}

// one returns the multiplicative unit, the [EuclideanDivMod] quotient
// step.
func (Uint128) one() Uint128 {
	return Uint128{lo: 1}
}

// ulp returns the smallest positive value, the [EuclideanMulDivMod]
// quotient step; for an integer it equals one.
func (Uint128) ulp() Uint128 {
	return Uint128{lo: 1}
}

// NewUint128 assembles a Uint128 from its high and low 64-bit words,
// so the value is hi*2^64 + lo.
func NewUint128(hi, lo uint64) Uint128 {
	return Uint128{hi: hi, lo: lo}
}

// AsUint128 zero-extends an unsigned 64-bit value into a Uint128.
func AsUint128(x uint64) Uint128 {
	return Uint128{lo: x}
}

// wide returns u as a count of whole units, on the way to another
// type of the family. The bits are read as an Int128, which holds
// them while the top bit is clear.
func (u Uint128) wide() wide {
	v := Int128(u)
	return wide{v: v, scale: unitScale128, ok: !v.IsNegative()}
}

// Int32 returns u as an Int32 and whether it fits; the low 32 bits
// stay when it does not.
func (u Uint128) Int32() (Int32, bool) {
	return u.wide().int32()
}

// Int64 returns u as an Int64 and whether it fits; the low 64 bits
// stay when it does not.
func (u Uint128) Int64() (Int64, bool) {
	return u.wide().int64()
}

// Int128 returns u as an Int128 and whether it fits, which it does
// while the top bit is clear; the bit pattern stays when it does not.
func (u Uint128) Int128() (Int128, bool) {
	return u.wide().int128()
}

// Uint128 returns u unchanged, the conversion to its own type, which
// always fits; it stays off the wide path, which reads a set top bit
// as a sign.
func (u Uint128) Uint128() (Uint128, bool) {
	return u, true
}

// Milli32 returns u as whole units of a Milli32 and whether it fits;
// the low 32 bits of the milli count stay when it does not.
func (u Uint128) Milli32() (Milli32, bool) {
	return u.wide().milli32()
}

// Milli64 returns u as whole units of a Milli64 and whether it fits;
// the low 64 bits of the milli count stay when it does not.
func (u Uint128) Milli64() (Milli64, bool) {
	return u.wide().milli64()
}

// Atto128 returns u as whole units of an Atto128 and whether it fits;
// the low 128 bits of the atto count stay when it does not.
func (u Uint128) Atto128() (Atto128, bool) {
	return u.wide().atto128()
}

// GoString returns the constructor call that rebuilds u for %#v:
// AsUint128 over the low word while the high word is zero, and
// NewUint128 over both words in hex otherwise.
func (u Uint128) GoString() string {
	if u.hi == 0 {
		return fmt.Sprintf("num.AsUint128(%d)", u.lo)
	}
	return fmt.Sprintf("num.NewUint128(%#x, %#x)", u.hi, u.lo)
}

// String returns u in decimal, the text %v prints.
func (u Uint128) String() string {
	return string(u.doAppendText(nil))
}

// AppendText implements [encoding.TextAppender], appending u in
// decimal, the text %v prints, to b. It allocates only when b lacks
// the room, and the error is always nil.
func (u Uint128) AppendText(b []byte) ([]byte, error) {
	return u.doAppendText(b), nil
}

// MarshalText implements [encoding.TextMarshaler], returning the
// AppendText text.
func (u Uint128) MarshalText() ([]byte, error) {
	return u.AppendText(nil)
}

// MarshalJSON implements [json.Marshaler], returning the MarshalText
// text as a JSON number while the value is at most 2^53, safe for a
// float64 consumer, and as a JSON string beyond it.
func (u Uint128) MarshalJSON() ([]byte, error) {
	if isJSONSafeInt(u) {
		return u.MarshalText()
	}
	return jsonString(u)
}

// Format implements [fmt.Formatter] with the verbs d, v and s for
// decimal, x and X for hex, o and O for octal and b for binary, under
// the flags, width and precision fmt gives its own integers; %#v prints
// the GoString form. Any other verb prints as %!verb(num.Uint128=value).
func (u Uint128) Format(s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('#'):
		writeGoString(s, u.GoString())
	case !isFormatVerb(verb):
		writeBadVerb(s, verb, "num.Uint128", func() {
			writeNumber(s, 'd', false, u.doAppendText(nil))
		})
	default:
		writeNumber(s, verb, false, u.appendDigits(nil, verb))
	}
}

// appendDigits appends the digits of u in the base verb names.
func (u Uint128) appendDigits(dst []byte, verb rune) []byte {
	shift := verbShift(verb)
	if shift == 0 {
		return u.doAppendText(dst)
	}
	return u.appendBits(dst, shift, verbDigits(verb))
}

// doAppendText writes the base-10 digits of u to dst and returns the
// extended buffer. A value below 2^64 formats in a single pass; a wider
// one is peeled into base-decChunk groups, most significant first, with
// the trailing groups zero-padded to decChunkDigits. The 128-bit range
// spans at most 39 digits, so three groups always suffice.
func (u Uint128) doAppendText(dst []byte) []byte {
	if u.hi == 0 {
		return strconv.AppendUint(dst, u.lo, 10)
	}
	div := Uint128{lo: decChunk}
	var chunk [3]uint64
	n := 0
	for rest := u; !rest.IsZero(); {
		var r Uint128
		rest, r = rest.DivMod(div)
		chunk[n] = r.lo
		n++
	}
	dst = strconv.AppendUint(dst, chunk[n-1], 10)
	for i := n - 2; i >= 0; i-- {
		dst = appendPadded(dst, chunk[i], decChunkDigits)
	}
	return dst
}

// appendBits writes the digits of u in the power-of-two base with shift
// bits per digit, drawn from the digits alphabet. The digits fall out
// least significant first, so they are gathered from the end of a
// buffer wide enough for 128 binary digits.
func (u Uint128) appendBits(dst []byte, shift uint, digits string) []byte {
	var buf [128]byte
	i := len(buf)
	mask := uint64(1)<<shift - 1
	rest := u
	for {
		i--
		buf[i] = digits[rest.lo&mask]
		rest = rest.shr(shift)
		if rest.IsZero() {
			return append(dst, buf[i:]...)
		}
	}
}

// shr returns u shifted right by n bits, for 0 < n < 64.
func (u Uint128) shr(n uint) Uint128 {
	return Uint128{
		hi: u.hi >> n,
		lo: u.lo>>n | u.hi<<(64-n),
	}
}

// IsZero reports whether u is zero.
func (u Uint128) IsZero() bool {
	return u.hi == 0 && u.lo == 0
}

// Equal reports whether u and v are equal.
func (u Uint128) Equal(v Uint128) bool {
	return u == v
}

// IsNegative reports whether u is less than zero, which is never the
// case for an unsigned integer.
func (Uint128) IsNegative() bool {
	return false
}

// Abs returns u unchanged, an unsigned integer being its own absolute
// value.
func (u Uint128) Abs() Uint128 {
	return u
}

// Add returns u+v, wrapping on overflow.
func (u Uint128) Add(v Uint128) Uint128 {
	lo, carry := bits.Add64(u.lo, v.lo, 0)
	hi, _ := bits.Add64(u.hi, v.hi, carry)
	return Uint128{hi: hi, lo: lo}
}

// Sub returns u-v, wrapping on underflow.
func (u Uint128) Sub(v Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v.lo, 0)
	hi, _ := bits.Sub64(u.hi, v.hi, borrow)
	return Uint128{hi: hi, lo: lo}
}

// Mul returns the low 128 bits of the product, wrapping on overflow.
func (u Uint128) Mul(v Uint128) Uint128 {
	hi, lo := bits.Mul64(u.lo, v.lo)
	hi += u.hi*v.lo + u.lo*v.hi
	return Uint128{hi: hi, lo: lo}
}

// Div returns u/v, truncated. It panics with [ErrDivZero] when v is
// zero.
func (u Uint128) Div(v Uint128) Uint128 {
	q, _ := u.DivMod(v)
	return q
}

// Mod returns the remainder of u/v. It panics with [ErrDivZero] when v
// is zero.
func (u Uint128) Mod(v Uint128) Uint128 {
	_, r := u.DivMod(v)
	return r
}

// DivMod returns the quotient and remainder of u/v, so that
// u == q*v + r with r < v. It panics with [ErrDivZero] when v is zero.
func (u Uint128) DivMod(v Uint128) (q, r Uint128) {
	if v.IsZero() {
		panic(ErrDivZero)
	}
	// promote to a 256-bit numerator with a zero high half and reuse
	// the wide divider; the quotient fits in 128 bits.
	return u256{lo: u}.divMod128(v)
}

// MulDivMod returns the quotient and remainder of u*v/d, forming the
// product in a 256-bit intermediate so it cannot overflow before the
// division. The quotient wraps if it exceeds the 128-bit range and the
// remainder is always less than d. It panics with [ErrDivZero] when d
// is zero.
func (u Uint128) MulDivMod(v, d Uint128) (q, r Uint128) {
	if d.IsZero() {
		panic(ErrDivZero)
	}
	return mul256(u, v).divMod128(d)
}

// Cmp returns -1, 0 or +1 as u is less than, equal to or greater
// than v.
func (u Uint128) Cmp(v Uint128) int {
	switch {
	case u.hi > v.hi:
		return 1
	case u.hi < v.hi:
		return -1
	case u.lo > v.lo:
		return 1
	case u.lo < v.lo:
		return -1
	default:
		return 0
	}
}

// bitLen returns the number of bits needed to represent u, zero for
// zero.
func (u Uint128) bitLen() int {
	if u.hi != 0 {
		return 64 + bits.Len64(u.hi)
	}
	return bits.Len64(u.lo)
}

// shl1 returns u shifted left by one bit with in as the new bit 0.
func (u Uint128) shl1(in uint64) Uint128 {
	return Uint128{
		hi: u.hi<<1 | u.lo>>63,
		lo: u.lo<<1 | in&1,
	}
}

// setBit returns u with bit i set when i is within the low 128 bits;
// higher bits are dropped, matching the wrapping policy of Add and Mul.
func (u Uint128) setBit(i int) Uint128 {
	switch {
	case i >= 128:
		return u
	case i >= 64:
		u.hi |= 1 << (i - 64)
	default:
		u.lo |= 1 << i
	}
	return u
}
