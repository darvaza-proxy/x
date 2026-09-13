package num

import (
	"encoding"
	"encoding/json"
	"fmt"
)

// Unsigned is the common surface of the fixed-width unsigned integer
// types in this package: addition, subtraction, multiplication,
// division and comparison over a magnitude that never carries a sign.
//
// Div returns the quotient, Mod the remainder, and DivMod both, so
// that v*q + r recovers the dividend with r < v. MulDivMod fuses a
// multiply and a divide, forming the product in an intermediate wide
// enough that it cannot overflow before the division.
type Unsigned[T any] interface {
	IsZero() bool
	Equal(v T) bool
	Cmp(v T) int

	Add(v T) T
	Sub(v T) T
	Mul(v T) T
	Div(v T) T
	Mod(v T) T
	DivMod(v T) (q, r T)
	MulDivMod(v, d T) (q, r T)
}

// Signed extends Unsigned with the sign-aware operations of a
// two's-complement integer: sign inspection, negation and absolute
// value. Division truncates towards zero, so the remainder takes the
// sign of the dividend.
type Signed[T any] interface {
	Unsigned[T]

	IsNegative() bool
	Neg() T
	Abs() T
}

// Number is the whole surface the family shares, the constraint a
// generic consumer names to take any of its seven types: the
// arithmetic of Unsigned with the Euclidean correction surface, which
// closes it to this package, the fmt, encoding and JSON forms, and a
// conversion to every type of the family.
//
// The conversions keep the value, not the count: an integer becomes
// whole units and a Decimal is rescaled, so AsInt32(5).Atto128() is
// 5.0 and NewAtto128(1, 500e15).Int64() is 1, the fraction digits
// below the target's resolution dropped towards zero. The flag is
// false only when the whole units do not fit the target, the result
// then keeping the low bits as a Go conversion does; a negative into
// Uint128 is its bit pattern. The method named for the receiver's own
// type returns it unchanged.
type Number[T any] interface {
	Unsigned[T]
	Euclidean[T]

	encoding.TextAppender
	encoding.TextMarshaler
	json.Marshaler

	fmt.Formatter
	fmt.GoStringer
	fmt.Stringer

	Int32() (Int32, bool)
	Int64() (Int64, bool)
	Int128() (Int128, bool)
	Uint128() (Uint128, bool)

	Milli32() (Milli32, bool)
	Milli64() (Milli64, bool)
	Atto128() (Atto128, bool)
}
