// Package tai provides TAI (International Atomic Time) and TAIN timestamps
// that follow the Go time package API conventions.
package tai

import (
	"math"
	"math/bits"
	"time"
)

func toNanos128(sec uint64, nano uint32) (hi, lo uint64) {
	hi, lo = bits.Mul64(sec, nsPerSec)
	var carry uint64
	lo, carry = bits.Add64(lo, uint64(nano), 0)
	hi += carry
	return hi, lo
}

func fromNanos128(hi, lo uint64) (sec uint64, nano uint32) {
	// Guard against bits.Div64 overflow panic.
	if hi >= nsPerSec {
		// Treat this as a saturation/overflow limit, or handle as an error condition
		return math.MaxUint64, uint32(nsPerSec - 1)
	}

	q, r := bits.Div64(hi, lo, nsPerSec)
	return q, uint32(r)
}

func sub128(hi1, lo1, hi2, lo2 uint64) (hi, lo uint64) {
	var borrow uint64
	lo, borrow = bits.Sub64(lo1, lo2, 0)
	hi, _ = bits.Sub64(hi1, hi2, borrow)
	return hi, lo
}

func add128(hi, lo, y uint64) (rHi, rLo uint64) {
	var carry uint64
	rLo, carry = bits.Add64(lo, y, 0)
	rHi = hi + carry
	return rHi, rLo
}

func remainder128(hi, lo, y uint64) uint64 {
	_, rem := bits.Div64(hi%y, lo, y)
	return rem
}

func durationFromSigned128(hi int64, lo uint64) time.Duration {
	// Safe positive bounds: hi is 0, lo fits in positive int64
	if hi == 0 && int64(lo) >= 0 {
		return time.Duration(lo)
	}
	// Safe negative bounds: hi is -1, lo represents a negative int64
	if hi == -1 && int64(lo) < 0 {
		return time.Duration(lo)
	}

	// Out of bounds handler
	if hi < 0 {
		return minDuration
	}
	return maxDuration
}

func safeAddUint64(a, b uint64) (uint64, error) {
	sum, carry := bits.Add64(a, b, 0)
	if carry != 0 {
		return 0, ErrOverflow
	}
	return sum, nil
}

func safeSubUint64(a, b uint64) (uint64, error) {
	diff, borrow := bits.Sub64(a, b, 0)
	if borrow != 0 {
		return 0, ErrUnderflow
	}
	return diff, nil
}

// floorDivMod computes the floor-division quotient and non-negative
// remainder of a / b for a positive divisor b, i.e. a == q*b + r with
// 0 <= r < b. Unlike Go's built-in / and %, which truncate toward zero,
// this normalizes negative attosecond carries without negating a first —
// negation overflows when a == math.MinInt64.
func floorDivMod(a, b int64) (q, r int64) {
	q = a / b
	r = a % b
	if r != 0 && (r < 0) != (b < 0) {
		q--
		r += b
	}
	return q, r
}
