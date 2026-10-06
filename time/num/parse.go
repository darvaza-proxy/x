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
// value, wrapping any failure in its own name. UnmarshalJSON reads the
// same grammar from a JSON number as given, or from the contents of a
// JSON string, and stores the value as UnmarshalText does. A Decimal
// instantiation reads through [ParseDecimal], and through the function
// named for it where the package defines one, in the shape of
// [strconv.ParseFloat] instead: digits on at least one side of an
// optional point, the fraction below the resolution dropped, and the
// text read for its syntax before its size.

// newParseError returns the report of the parser fn failing on the
// text s with err, [ErrSyntax] or [ErrRange]. The report keeps a copy
// of s, as strconv's does, so that s does not escape through it.
func newParseError[S stringOrBytes](fn string, s S, err error) error {
	return &ParseError{Func: fn, Num: cloneText(s), Err: err}
}

// cloneText returns a copy of s as a string: bytes are converted and a
// string is cloned, each copying the text once. A named type over
// either is converted and then cloned, which copies named bytes twice
// and a named string once.
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

// parserName returns the name a [ParseError] reports for the parser of
// a Decimal instantiation: Parse in front of the scaler's type name
// with its package qualifier stripped, so num.Milli32 gives
// ParseMilli32, and an instantiation of another package is named from
// its own scaler the same way.
func parserName(typeName string) string {
	return "Parse" + typeName[strings.LastIndexByte(typeName, '.')+1:]
}

// unmarshalInto finishes an unmarshal method: it stores the parsed
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
