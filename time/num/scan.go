package num

import (
	"math"
	"math/bits"
)

// stringOrBytes is the text a scanner reads: a string, or the bytes
// UnmarshalText is given, read where they are rather than converted.
type stringOrBytes interface {
	~string | ~[]byte
}

// scanner reads the text of a number a byte at a time, in the style of
// text/lexer's Cursor cut down to what a number needs: bytes rather
// than runes, since every byte a number's text takes is ASCII, and no
// emit buffer, since the parsers build values rather than text. It
// reads a string or the bytes as given, so UnmarshalText converts
// nothing. The grammar is read by its methods, each called directly in
// the order of the grammar, so the scanner stays on the stack, and all
// but the one reading a magnitude are small enough to inline.
type scanner[S stringOrBytes] struct {
	src S
	pos int
}

// done reports whether the scanner is at the end of the text.
func (sc *scanner[S]) done() bool {
	return sc.pos >= len(sc.src)
}

// accept advances past the next byte when it is c, and reports
// whether it was.
func (sc *scanner[S]) accept(c byte) bool {
	if sc.pos < len(sc.src) && sc.src[sc.pos] == c {
		sc.pos++
		return true
	}
	return false
}

// sign reads an optional sign, + or -, and reports whether it was a
// minus.
func (sc *scanner[S]) sign() bool {
	if sc.accept('-') {
		return true
	}
	sc.accept('+')
	return false
}

// digits reads up to most decimal digits, stopping at the first byte
// that is not one, and returns their value and how many it read. A
// most of up to decGroupDigits keeps the value within a uint64. The
// loop runs on a copy of the text and the position, which stay in
// registers, rather than through the scanner.
func (sc *scanner[S]) digits(most int) (d uint64, n int) {
	s, i := sc.src, sc.pos
	end := min(len(s), i+most)
	for ; i < end; i++ {
		c := s[i] - '0'
		if c > 9 {
			break
		}
		d = d*10 + uint64(c)
	}
	n, sc.pos = i-sc.pos, i
	return d, n
}

// magnitude64 reads the run of decimal digits as a magnitude held to
// bound, a group of decGroupDigits at a time, and returns it, how many
// digits the run holds and whether the magnitude stays within bound. A
// magnitude past bound comes back as bound, its run read to the end
// all the same, so the scanner stops at the first byte that is not a
// digit either way.
func (sc *scanner[S]) magnitude64(bound uint64) (u uint64, n int, ok bool) {
	u, n = sc.digits(decGroupDigits)
	ok = true
	for {
		d, k := sc.digits(decGroupDigits)
		if k == 0 {
			break
		}
		hi, lo := bits.Mul64(u, pow10Table[k].lo)
		lo, carry := bits.Add64(lo, d, 0)
		u, n, ok = lo, n+k, ok && hi|carry == 0
	}
	if !ok || u > bound {
		return bound, n, false
	}
	return u, n, true
}

// doParseInt reads the text of a signed integer width bits wide as
// strconv.ParseInt reads it in base 10: an optional sign and the
// digits, nothing else. It returns the value; for text in any other
// form, zero and [ErrSyntax]; and for a number past the range, the
// nearest bound and [ErrRange]. As in strconv, the magnitude is read
// at the unsigned width and then held to the sign's bound, and the
// text is read from the left, so a magnitude past the unsigned width
// fails on its range before a stray byte after its digits fails on
// syntax.
func doParseInt[S stringOrBytes](s S, width int) (int64, error) {
	sc := scanner[S]{src: s}
	neg := sc.sign()
	u, n, ok := sc.magnitude64(math.MaxUint64 >> (64 - width))
	switch {
	case n == 0, ok && !sc.done():
		return 0, ErrSyntax
	default:
		return boundInt(u, neg, width)
	}
}

// boundInt returns the magnitude u under a sign as a signed integer
// width bits wide: a positive value fits below 2^(width-1), and a
// negative one reaches one further, to 2^(width-1) itself. A magnitude
// past the bound comes back as the bound on its own side, with
// [ErrRange].
func boundInt(u uint64, neg bool, width int) (int64, error) {
	limit := uint64(1) << (width - 1)
	switch {
	case neg && u > limit:
		return -int64(limit), ErrRange
	case neg:
		return -int64(u), nil
	case u >= limit:
		return int64(limit - 1), ErrRange
	default:
		return int64(u), nil
	}
}
