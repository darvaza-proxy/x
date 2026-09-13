package num

import (
	"strings"
	"testing"

	"darvaza.org/core"
)

// namedString and namedBytes are texts of a named type over each kind
// stringOrBytes admits.
type (
	namedString string
	namedBytes  []byte
)

// longText returns a text that fails to parse and runs past the 32
// bytes the runtime converts from bytes on the stack, so every copy of
// it allocates.
func longText() string {
	return strings.Repeat("0", 40) + "x"
}

func TestNewParseError(t *testing.T) {
	t.Run("string", runTestNewParseErrorOf[string])
	t.Run("bytes", runTestNewParseErrorOf[[]byte])
	t.Run("named string", runTestNewParseErrorOf[namedString])
	t.Run("named bytes", runTestNewParseErrorOf[namedBytes])
	t.Run("allocs", runTestNewParseErrorAllocs)
}

// runTestNewParseErrorOf checks that the report keeps a copy of a text
// of type S, which then holds when the bytes it was read from change.
func runTestNewParseErrorOf[S stringOrBytes](t *testing.T) {
	t.Helper()
	text := longText()
	src := []byte(text)
	err := newParseError("ParseInt32", S(src), ErrSyntax)
	clear(src)

	pe := core.AssertMustErrorAs[*ParseError](t, err, "report")
	core.AssertEqual(t, "ParseInt32", pe.Func, "function")
	core.AssertEqual(t, text, pe.Num, "text")
	core.AssertErrorIs(t, err, ErrSyntax, "cause")
}

// runTestNewParseErrorAllocs checks that a report over bytes costs no
// more than one over a string, both copying the text once.
func runTestNewParseErrorAllocs(t *testing.T) {
	t.Helper()
	text := longText()
	src := []byte(text)

	var err error
	fromString := testing.AllocsPerRun(100, func() {
		err = newParseError("ParseInt32", text, ErrSyntax)
	})
	fromBytes := testing.AllocsPerRun(100, func() {
		err = newParseError("ParseInt32", src, ErrSyntax)
	})
	core.AssertError(t, err, "report")
	core.AssertEqual(t, fromString, fromBytes, "allocations over bytes")
}
