// Package num provides the fixed-width integer types used by the time
// packages: Uint128 and Int128, unsigned and signed 128-bit integers.
//
// Arithmetic wraps on overflow, matching Go's built-in integer
// operators, so Add, Sub and Mul never panic. Division by zero panics
// with [ErrDivZero]. Signed division truncates towards zero, with the
// remainder taking the sign of the dividend.
package num
