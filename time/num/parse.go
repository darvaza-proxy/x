package num

import (
	"errors"
	"strconv"

	"darvaza.org/core"
)

// A type reads its own text back through a function named for it,
// ParseInt32 for Int32, in the shape of the strconv parsers: the text
// is the whole input, base 10, with no underscores, prefixes or
// spaces, and a failure comes back as a [ParseError] naming the
// function, quoting the text and carrying [ErrSyntax] or [ErrRange],
// the value then zero or the nearest bound as strconv has it.
// UnmarshalText on the pointer side reads the same grammar and stores
// the value, wrapping any failure in its own name.

// AsParseError returns the error of a parse as a [ParseError] naming
// fn over the text s, for a parser built on this package or on strconv
// to report in the same shape. A [strconv.NumError] or a ParseError in
// is opened: its cause is reported, and its function and text stand in
// for an empty fn or s, so strconv's own report passes through with
// nothing added, and a ParseError is re-reported under a new name. The
// cause comes out as the package's own: [strconv.ErrSyntax] and
// [strconv.ErrRange] as [ErrSyntax] and [ErrRange], an error already
// matching [core.ErrInvalid] as it is, and anything else compounded
// with core.ErrInvalid so that both stay reachable. A nil error, a
// typed nil included, stays nil.
func AsParseError(fn, s string, err error) error {
	if e, ok := asNumError(err); ok {
		if e == nil {
			return nil
		}
		fn = core.Coalesce(fn, e.Func)
		s = core.Coalesce(s, e.Num)
		err = e.Err
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

// addGroup adds a group of digits to the magnitude read so far,
// u*scale + d for the scale the group's length gives, and reports
// whether the sum fits 128 bits; MaxUint128 stands in when it does not.
func addGroup(u Uint128, scale, d uint64) (Uint128, bool) {
	p := mul256(u, Uint128{lo: scale})
	sum := p.lo.Add(Uint128{lo: d})
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
	if _, ok := addGroup(u, uint64(pow10(n)), d); !ok {
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
