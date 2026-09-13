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
