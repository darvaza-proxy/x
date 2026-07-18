// Package num provides the fixed-width integer types used by the time
// packages: the signed integers Int32, Int64 and Int128, and the unsigned
// Uint128.
//
// Arithmetic wraps on overflow, matching Go's built-in integer
// operators, so Add, Sub and Mul never panic. Division by zero panics
// with [ErrDivZero]. Signed division truncates towards zero, with the
// remainder taking the sign of the dividend.
package num
