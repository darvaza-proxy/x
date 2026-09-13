package num

import (
	"strings"

	"darvaza.org/core"
)

// A type that parses reads its own text back through a function named
// for it, ParseInt32 for Int32, in the shape of the strconv parsers:
// the text is the whole input, base 10, with no underscores, prefixes
// or spaces, and a failure comes back as a [ParseError] naming the
// function, quoting the text and carrying [ErrSyntax] or [ErrRange],
// the value then zero or the nearest bound as strconv has it.
// UnmarshalText on the pointer side reads the same grammar from its
// bytes as given, without converting them to a string, and stores the
// value, wrapping any failure in its own name.

// newParseError returns the report of the parser fn failing on the
// text s with err, [ErrSyntax] or [ErrRange]. The report keeps a copy
// of s, as strconv's does, so that s does not escape through it, and
// a caller converting a short text from bytes keeps the conversion on
// the stack.
func newParseError[S stringOrBytes](fn string, s S, err error) error {
	return &ParseError{Func: fn, Num: cloneText(s), Err: err}
}

// cloneText returns a copy of s as a string, copying the text once:
// the conversion copies bytes, and a string is cloned. A named type
// over either is converted and then cloned, which copies bytes twice.
func cloneText[S stringOrBytes](s S) string {
	switch v := any(s).(type) {
	case []byte:
		return string(v)
	case string:
		return strings.Clone(v)
	default:
		return strings.Clone(string(s))
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
