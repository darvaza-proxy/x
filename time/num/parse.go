package num

import (
	"errors"
	"math/bits"
	"strconv"
	"strings"

	"darvaza.org/core"
)

// A type reads its own text back through a function named for it,
// ParseInt32 for Int32, in the shape of the strconv parsers: the text
// is the whole input, base 10, with no underscores, prefixes or
// spaces, and a failure comes back as a [ParseError] naming the
// function, quoting the text and carrying [ErrSyntax] or [ErrRange],
// the value then zero or the nearest bound as strconv has it.
// UnmarshalText on the pointer side reads the same grammar and stores
// the value, wrapping any failure in its own name. A Decimal
// instantiation reads through [ParseDecimal], and through the function
// named for it where the package defines one, in the shape of
// [strconv.ParseFloat] instead: digits on at least one side of an
// optional point, the fraction below the resolution dropped, and the
// text read for its syntax before its size.

// AsParseError returns the error of a parse as a [ParseError] naming
// fn over the text s, for a parser built on this package or on strconv
// to report in the same shape. A [strconv.NumError] or a ParseError in
// is opened: its cause is reported, [core.ErrInvalid] when it carries
// none, and its function and text stand in for an empty fn or s, so
// strconv's own report passes through with nothing added, and a
// ParseError is re-reported under a new name. The cause comes out as
// the package's own: [strconv.ErrSyntax] and [strconv.ErrRange] as
// [ErrSyntax] and [ErrRange], an error already matching
// [core.ErrInvalid] as it is, and anything else compounded with
// core.ErrInvalid so that both stay reachable. A nil error, a typed
// nil included, stays nil.
func AsParseError(fn, s string, err error) error {
	if e, ok := asNumError(err); ok {
		if e == nil {
			return nil
		}
		fn = core.Coalesce(fn, e.Func)
		s = core.Coalesce(s, e.Num)
		err = core.CoalesceError(e.Err, core.ErrInvalid)
	}
	if err = parseCause(err); err == nil {
		return nil
	}
	return &ParseError{Func: fn, Num: s, Err: err}
}

// asNumError returns err as a *strconv.NumError when it is one, or a
// *ParseError, which shares the shape, and whether it was. A typed nil
// of either comes back as a nil pointer.
func asNumError(err error) (*strconv.NumError, bool) {
	switch e := err.(type) {
	case *strconv.NumError:
		return e, true
	case *ParseError:
		return (*strconv.NumError)(e), true
	default:
		return nil, false
	}
}

// parseCause returns the cause a ParseError carries for err: strconv's
// sentinels as the package's own, an error already matching
// [core.ErrInvalid] at any depth as it is, the package's own included,
// and anything else compounded with core.ErrInvalid so that both stay
// reachable through errors.Is. Nil stays nil.
func parseCause(err error) error {
	switch {
	case err == nil, errors.Is(err, core.ErrInvalid):
		return err
	case err == strconv.ErrRange:
		return ErrRange
	case err == strconv.ErrSyntax:
		return ErrSyntax
	default:
		return core.NewCompoundError(err, core.ErrInvalid)
	}
}

// splitSign cuts a leading sign off s, reporting whether it was a
// minus. The rest is the magnitude, empty for a sign on its own, which
// the digit reader then refuses.
func splitSign(s string) (neg bool, mag string) {
	switch {
	case len(s) > 0 && s[0] == '-':
		return true, s[1:]
	case len(s) > 0 && s[0] == '+':
		return false, s[1:]
	default:
		return false, s
	}
}

// splitDigits cuts a magnitude at its decimal point into the whole
// digits and the fraction digits, reporting whether it is a magnitude
// at all: digits, and nothing else, on at least one side of at most one
// point, so 1. and .5 read as [strconv.ParseFloat] reads them and a
// bare point does not. An absent whole part stands for zero.
func splitDigits(mag string) (whole, frac string, ok bool) {
	whole, frac, _ = strings.Cut(mag, ".")
	if whole == "" && frac == "" {
		return "", "", false
	}
	return whole, frac, allDigits(whole) && allDigits(frac)
}

// allDigits reports whether s is decimal digits and nothing else. An
// empty s passes, since the point may carry digits on one side only.
func allDigits(s string) bool {
	for _, c := range []byte(s) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// parseFraction returns the fraction digits as a count of sub-units at
// a resolution of width digits: the first width of them, zero-filled
// when the text holds fewer, and those below the resolution dropped
// towards zero. The digits are known to be digits. A width of up to
// decGroupDigits is read as one group; a wider one, up to the 38
// digits of a scale below 2^127, as a leading group and a full one,
// joined at the scale of decGroup.
func parseFraction(frac string, width int) Uint128 {
	if width <= decGroupDigits {
		return Uint128{lo: fracGroup(frac, 0, width)}
	}
	cut := width - decGroupDigits
	hi := fracGroup(frac, 0, cut)
	lo := fracGroup(frac, cut, width)
	// hi holds at most decGroupDigits digits, so hi*decGroup + lo is
	// below 10^38 and the carry fits the top word.
	top, low := bits.Mul64(hi, decGroup)
	low, carry := bits.Add64(low, lo, 0)
	return Uint128{hi: top + carry, lo: low}
}

// fracGroup reads the fraction digits from index from up to index to
// as a count, a digit past the end of frac reading as zero. The span
// is at most decGroupDigits long, so the count fits a uint64.
func fracGroup(frac string, from, to int) uint64 {
	var d uint64
	for i := from; i < to; i++ {
		d *= 10
		if i < len(frac) {
			d += uint64(frac[i] - '0')
		}
	}
	return d
}

// boundSign returns the magnitude u under a sign as an Int128 and
// whether it fitted. As strconv.ParseInt has it over ParseUint, the
// magnitude is read at the full unsigned width and then held to the
// sign's bound: Int128 says whether it fits below 2^127, and a negative
// value reaches one further, to 2^127 itself, the bits of MinInt128,
// which Neg leaves as they are. A magnitude past the bound comes back
// as the bound on its own side.
func boundSign(u Uint128, neg bool) (Int128, bool) {
	v, ok := u.Int128()
	switch {
	case ok && neg:
		return v.Neg(), true
	case ok:
		return v, true
	case neg && u == MinInt128.bits():
		return MinInt128, true
	case neg:
		return MinInt128, false
	default:
		return MaxInt128, false
	}
}

// parserName returns the name a [ParseError] reports for the parser of
// a Decimal instantiation: Parse in front of the scaler's type name
// with its package qualifier stripped, so num.Milli32 gives
// ParseMilli32, and an instantiation of another package is named from
// its own scaler the same way.
func parserName(typeName string) string {
	return "Parse" + typeName[strings.LastIndexByte(typeName, '.')+1:]
}

// splitGroup cuts the leading group of digits off s: decGroupDigits of
// them, or fewer when the length is not a multiple of that, so every
// group after the first is full and is added on at the one scale,
// decGroup. An empty s gives an empty head, for strconv to refuse.
func splitGroup(s string) (head, tail string) {
	n := len(s) % decGroupDigits
	if n == 0 && s != "" {
		n = decGroupDigits
	}
	return s[:n], s[n:]
}

// addGroup adds a group of digits d to the magnitude read so far,
// u*scale + d, and reports whether the sum fits 128 bits; MaxUint128
// stands in when it does not.
func addGroup(u, scale, d Uint128) (Uint128, bool) {
	p := mul256(u, scale)
	sum := p.lo.Add(d)
	if !p.hi.IsZero() || sum.Cmp(p.lo) < 0 {
		return MaxUint128, false
	}
	return sum, true
}

// leadingDigits reads the run of decimal digits leading s, returning
// its value and its length. s is a group strconv refused, so the run
// is at most decGroupDigits-1 long and its value fits a uint64.
func leadingDigits(s string) (d uint64, n int) {
	for n < len(s) && '0' <= s[n] && s[n] <= '9' {
		d = d*10 + uint64(s[n]-'0')
		n++
	}
	return d, n
}

// refuseGroup answers for a group strconv refused with err: a group of
// at most decGroupDigits bytes fits a uint64, so it is empty or holds a
// byte that is not a digit. strconv reads one digit at a time and
// reports the failure it meets first, so the answer is
// [strconv.ErrRange] with MaxUint128 when the digits before that byte
// already take the magnitude past 128 bits, and err itself with zero
// otherwise.
func refuseGroup(u Uint128, group string, err error) (Uint128, error) {
	d, n := leadingDigits(group)
	if _, ok := addGroup(u, Pow10(n), Uint128{lo: d}); !ok {
		return MaxUint128, strconv.ErrRange
	}
	return Uint128{}, err
}

// unmarshalInto finishes an UnmarshalText method: it stores the parsed
// value x in p when the parse succeeded and p is not nil, and
// otherwise returns the parse failure, or [core.ErrNilReceiver] once
// the text parsed, behind the method's name. A failure leaves p as it
// was.
func unmarshalInto[T any](p *T, x T, err error, method string) error {
	switch {
	case err != nil:
		return core.Wrap(err, method)
	case p == nil:
		return core.Wrap(core.ErrNilReceiver, method)
	default:
		*p = x
		return nil
	}
}
