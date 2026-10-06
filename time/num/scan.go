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
// the order of the grammar, so that the scanner can stay on the stack.
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
// loop runs on local copies of the text and the position, the position
// written back once, rather than through the scanner.
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

// magnitude128 reads the run of decimal digits as a 128-bit magnitude,
// a group of decGroupDigits at a time as magnitude64 does, and returns
// it, how many digits the run holds and whether the magnitude fits 128
// bits. Each later group joins the magnitude in two 64-bit multiplies,
// one per word. A magnitude past 128 bits comes back as MaxUint128,
// its run read to the end all the same, so the scanner stops at the
// first byte that is not a digit either way.
func (sc *scanner[S]) magnitude128() (u Uint128, n int, ok bool) {
	var hi, lo uint64
	lo, n = sc.digits(decGroupDigits)
	ok = true
	for {
		d, k := sc.digits(decGroupDigits)
		if k == 0 {
			break
		}
		// (hi, lo)*10^k + d: lo*10^k spills its high word into hi,
		// and past 128 bits go hi*10^k's high word, over, and the
		// carry out of hi.
		m := pow10Table[k].lo
		over, hm := bits.Mul64(hi, m)
		spill, lm := bits.Mul64(lo, m)
		var carry uint64
		lo, carry = bits.Add64(lm, d, 0)
		hi, carry = bits.Add64(hm, spill, carry)
		n, ok = n+k, ok && over|carry == 0
	}
	if !ok {
		return MaxUint128, n, false
	}
	return Uint128{hi: hi, lo: lo}, n, true
}

// fraction reads the run of decimal digits after a point as a count of
// sub-units at a resolution of width digits, and returns it and how
// many digits the run holds: the first width of them, zero-filled when
// the run is shorter, and those below the resolution read and dropped
// towards zero. A width of up to the 38 digits of a scale below 2^127
// is read as two groups, a leading one of the digits past
// decGroupDigits, none at all for a narrower width, and a full one;
// when the leading group stops short, the run has ended and the full
// one reads nothing.
func (sc *scanner[S]) fraction(width int) (u Uint128, n int) {
	cut := max(width-decGroupDigits, 0)
	hi, k := sc.digits(cut)
	lo, m := sc.digits(width - cut)
	hi *= pow10Table[cut-k].lo
	lo *= pow10Table[width-cut-m].lo
	// hi is below 10^cut and lo below 10^(width-cut), so
	// hi*10^(width-cut) + lo is below 10^width, at most 10^38, and the
	// carry fits the top word.
	top, low := bits.Mul64(hi, pow10Table[width-cut].lo)
	low, carry := bits.Add64(low, lo, 0)
	return Uint128{hi: top + carry, lo: low}, k + m + sc.skip()
}

// skip reads past the run of decimal digits and returns how many it
// held. The loop runs on local copies of the text and the position,
// the position written back once, as in digits.
func (sc *scanner[S]) skip() (n int) {
	s, i := sc.src, sc.pos
	for i < len(s) && s[i]-'0' <= 9 {
		i++
	}
	n, sc.pos = i-sc.pos, i
	return n
}

// doParseInt reads the text of a signed integer bitSize bits wide as
// strconv.ParseInt reads it in base 10: an optional sign and the
// digits, nothing else. It returns the value; for text in any other
// form, zero and [ErrSyntax]; and for a number past the range, the
// nearest bound and [ErrRange]. As in strconv, the magnitude is read
// at the unsigned width and then held to the sign's bound, and the
// text is read from the left, so a magnitude past the unsigned width
// fails on its range before a stray byte after its digits fails on
// syntax.
func doParseInt[S stringOrBytes](s S, bitSize int) (int64, error) {
	sc := scanner[S]{src: s}
	neg := sc.sign()
	u, n, ok := sc.magnitude64(math.MaxUint64 >> (64 - bitSize))
	switch {
	case n == 0, ok && !sc.done():
		return 0, ErrSyntax
	default:
		return boundInt(u, neg, bitSize)
	}
}

// boundInt returns the magnitude u under a sign as a signed integer
// bitSize bits wide: a positive value fits below 2^(bitSize-1), and a
// negative one reaches one further, to 2^(bitSize-1) itself. A
// magnitude past the bound comes back as the bound on its own side,
// with [ErrRange].
func boundInt(u uint64, neg bool, bitSize int) (int64, error) {
	limit := uint64(1) << (bitSize - 1)
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

// boundInt128 returns the magnitude u under a sign as an Int128, as
// boundInt does at 64 bits: u.Int128 says whether it fits below 2^127,
// and a negative value reaches one further, to 2^127 itself, the bits
// of MinInt128. A magnitude past the bound comes back as the bound on
// its own side, with [ErrRange].
func boundInt128(u Uint128, neg bool) (Int128, error) {
	v, ok := u.Int128()
	switch {
	case ok && neg:
		return v.Neg(), nil
	case ok:
		return v, nil
	case neg && u == MinInt128.bits():
		return MinInt128, nil
	case neg:
		return MinInt128, ErrRange
	default:
		return MaxInt128, ErrRange
	}
}

// doParseCount reads the text of a fixed-point number as
// strconv.ParseFloat reads a float without its exponent: an optional
// sign, then digits on at least one side of an optional point, nothing
// else. It returns the count of sub-units at a resolution of width
// fraction digits, those below the resolution dropped towards zero,
// held to the sign's bound at 128 bits by boundInt128. As in
// ParseFloat, the whole text is read for its syntax before its size, so
// text in any other form returns zero and [ErrSyntax] even where it
// passed the range, and only then does a count past the range return
// the nearest bound and [ErrRange].
func doParseCount[S stringOrBytes](s S, width int) (Int128, error) {
	sc := scanner[S]{src: s}
	neg := sc.sign()
	// a whole past 128 bits comes back as MaxUint128, which atScale
	// keeps past the bound, so whether it fitted is not asked here.
	whole, n, _ := sc.magnitude128()
	var frac Uint128
	if sc.accept('.') {
		var k int
		frac, k = sc.fraction(width)
		n += k
	}
	if n == 0 || !sc.done() {
		return Int128{}, ErrSyntax
	}
	return boundInt128(atScale(whole, frac, width), neg)
}

// atScale returns the count whole and frac make at a resolution of
// width fraction digits, whole*10^width + frac, for a frac below
// 10^width. A whole past pow10Bound puts the count past 2^127, beyond
// the bound of an Int128 under either sign, and comes back as
// MaxUint128, which boundInt128 holds to the bound. At unit resolution
// the whole is the count, with nothing to multiply.
func atScale(whole, frac Uint128, width int) Uint128 {
	switch {
	case width == unitExp:
		return whole
	case whole.Cmp(pow10Bound[width]) > 0:
		return MaxUint128
	default:
		return whole.Mul(pow10Table[width]).Add(frac)
	}
}

// doParseCount64 reads the text of a fixed-point number as
// doParseCount does, at a resolution of width fraction digits, and
// returns the count held to the sign's bound at bitSize bits by
// boundInt rather than at 128 bits. A width is at most 18, the digits
// of a scale within 64 bits, so the fraction is one group of digits:
// the first width of the run, zero-filled when it is shorter, and
// those below the resolution read and dropped towards zero. As in
// doParseCount, the whole text is read for its syntax before its size.
func doParseCount64[S stringOrBytes](s S, width, bitSize int) (int64, error) {
	sc := scanner[S]{src: s}
	neg := sc.sign()
	// a whole past 64 bits comes back as MaxUint64, which atScale64
	// keeps past the bound, so whether it fitted is not asked here.
	whole, n, _ := sc.magnitude64(math.MaxUint64)
	var frac uint64
	if sc.accept('.') {
		d, k := sc.digits(width)
		frac = d * pow10Table[width-k].lo
		n += k + sc.skip()
	}
	switch {
	case n == 0, !sc.done():
		return 0, ErrSyntax
	default:
		return boundInt(atScale64(whole, frac, width), neg, bitSize)
	}
}

// atScale64 returns the count whole and frac make at a resolution of
// width fraction digits, whole*10^width + frac, for a frac below
// 10^width, as atScale does at 128 bits. A count past 64 bits comes
// back as MaxUint64, past the bound of a signed integer of up to 64
// bits under either sign, which boundInt holds to the bound.
func atScale64(whole, frac uint64, width int) uint64 {
	hi, lo := bits.Mul64(whole, pow10Table[width].lo)
	lo, carry := bits.Add64(lo, frac, 0)
	if hi|carry != 0 {
		return math.MaxUint64
	}
	return lo
}
