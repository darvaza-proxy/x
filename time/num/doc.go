// Package num provides the fixed-width numeric types used by the time
// packages: the signed integers Int32, Int64 and Int128, the unsigned
// Uint128, and the signed fixed-point Decimal, whose Milli32, Milli64
// and Atto128 instantiations count sub-units at milli (10^-3) and atto
// (10^-18) resolution.
//
// Two constructor prefixes cover every type. New builds a value from
// its parts: the two words of a 128-bit integer, or the whole units and
// sub-units of a Decimal. As takes a value that already is the count and
// changes only its type: a native integer extended into a wider one, or
// a backing integer read as a Decimal at its resolution. So
// NewAtto128(1, 500e15) and AsAtto128(AsInt128(1500e15)) are both 1.5.
// AsInt32 and AsInt64 are the conversions of the native types under
// the same name; Int32 and Int64 have no parts, so no New.
//
// Under %#v every type prints as the call that rebuilds it, so a value
// dumped from a failing test pastes back into a row: num.AsInt32(-5),
// num.AsInt128(-42), num.NewMilli32(1, 500). The As form gives way to
// the New form over the two words in hex once a value no longer fits
// the native word, and a Decimal whose whole count no longer fits an
// int64 prints as As over its backing integer,
// num.AsAtto128(num.NewInt128(hi, lo)).
//
// Arithmetic wraps on overflow, matching Go's built-in integer
// operators, so Add, Sub and Mul never panic. Division by zero panics
// with [ErrDivZero]. Signed division truncates towards zero, with the
// remainder taking the sign of the dividend; [EuclideanDivMod] and
// [EuclideanMulDivMod] instead keep the remainder non-negative.
package num
