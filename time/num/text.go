package num

import (
	"errors"
	"fmt"
	"strconv"

	"darvaza.org/core"
)

// Shared base-10 text helpers for the integer types. Their String,
// MarshalText and UnmarshalText methods live beside each type; this file
// holds the parsing and formatting internals those methods share. Parsing
// rejects malformed input with [ErrSyntax] and out-of-range values with
// [ErrRange], each also wrapping core.ErrInvalid so the failures join the
// rest of the package's error family.

// decChunk is the largest power of ten fitting in a uint64. Peeling a
// Uint128 by it yields base-10 groups of decChunkDigits digits each,
// which strconv then formats one word at a time.
const decChunk uint64 = 1e19

// decChunkDigits is the number of decimal digits in decChunk.
const decChunkDigits = 19

// appendPadded appends v in base 10 to dst, left-padded with zeros to at
// least width digits.
func appendPadded(dst []byte, v uint64, width int) []byte {
	var tmp [20]byte
	s := strconv.AppendUint(tmp[:0], v, 10)
	for pad := width - len(s); pad > 0; pad-- {
		dst = append(dst, '0')
	}
	return append(dst, s...)
}

// parseInt128 parses a signed base-10 integer, accepting an optional
// leading sign. The magnitude bound depends on the sign: MaxInt128 for a
// positive value and 2^127 for a negative one, so MinInt128 parses back
// without wrapping.
func parseInt128(s string) (Int128, error) {
	neg, body := splitSign(s)
	bound := Uint128(MaxInt128)
	if neg {
		bound = MinInt128.bits()
	}
	mag, err := parseUint128(body, bound)
	if err != nil {
		return Int128{}, err
	}
	v := Int128(mag)
	if neg {
		v = v.Neg()
	}
	return v, nil
}

// splitSign strips a leading '+' or '-' from s, reporting whether the
// sign was negative along with the remaining digits.
func splitSign(s string) (neg bool, body string) {
	switch {
	case len(s) > 0 && s[0] == '-':
		return true, s[1:]
	case len(s) > 0 && s[0] == '+':
		return false, s[1:]
	default:
		return false, s
	}
}

// parseUint128 parses the base-10 digits of s into a Uint128 no greater
// than bound. It rejects empty input and non-digit bytes with [ErrSyntax]
// and magnitudes above bound with [ErrRange].
func parseUint128(s string, bound Uint128) (Uint128, error) {
	if s == "" {
		return Uint128{}, numError(ErrSyntax, s)
	}
	ten := Uint128{lo: 10}
	cutOff, cutOffRem := bound.DivMod(ten)
	var u Uint128
	for i := 0; i < len(s); i++ {
		d, ok := decDigit(s[i])
		if !ok {
			return Uint128{}, numError(ErrSyntax, s)
		}
		if overflowsUint128(u, cutOff, d, cutOffRem.lo) {
			return Uint128{}, numError(ErrRange, s)
		}
		u = u.Mul(ten).Add(Uint128{lo: d})
	}
	return u, nil
}

// decDigit maps an ASCII byte to its decimal value, reporting whether it
// was a digit at all.
func decDigit(c byte) (uint64, bool) {
	if c < '0' || c > '9' {
		return 0, false
	}
	return uint64(c - '0'), true
}

// overflowsUint128 reports whether appending digit d to u would exceed the
// bound whose floor division by ten is (cutOff, cutOffDigit).
func overflowsUint128(u, cutOff Uint128, d, cutOffDigit uint64) bool {
	cmp := u.Cmp(cutOff)
	return cmp > 0 || (cmp == 0 && d > cutOffDigit)
}

// parseIntError restates a strconv integer-parse failure through numError,
// so the fixed-width parsers report the same [ErrSyntax] or [ErrRange] over
// core.ErrInvalid as the 128-bit ones.
func parseIntError(b []byte, err error) error {
	sentinel := ErrSyntax
	if errors.Is(err, ErrRange) {
		sentinel = ErrRange
	}
	return numError(sentinel, string(b))
}

// numError reports a base-10 parse failure of s, wrapping the sentinel
// (either [ErrSyntax] or [ErrRange]) alongside core.ErrInvalid so the error
// matches through both.
func numError(sentinel error, s string) error {
	return fmt.Errorf("num: parsing %q: %w (%w)", s, sentinel, core.ErrInvalid)
}
