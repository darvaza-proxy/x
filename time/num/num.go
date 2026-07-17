package num

// Unsigned is the common surface of the fixed-width unsigned integer
// types in this package: addition, subtraction, multiplication,
// division and comparison over a magnitude that never carries a sign.
//
// Div returns the quotient, Mod the remainder, and DivMod both, so
// that v*q + r recovers the dividend with r < v.
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
