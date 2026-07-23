// Package num provides the fixed-width numeric types used by the time
// packages: the signed integers Int32, Int64 and Int128, the unsigned
// Uint128, and the signed fixed-point Decimal, whose Milli32, Milli64
// and Atto128 instantiations count sub-units at milli (10^-3) and atto
// (10^-18) resolution.
//
// Arithmetic wraps on overflow, matching Go's built-in integer
// operators, so Add, Sub and Mul never panic. Division by zero panics
// with [ErrDivZero]. Signed division truncates towards zero, with the
// remainder taking the sign of the dividend; [EuclideanDivMod] and
// [EuclideanMulDivMod] instead keep the remainder non-negative.
package num
