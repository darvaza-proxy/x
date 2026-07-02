package x509utils_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = nameAsIPTestCase{}
	_ core.TestCase = nameAsSuffixTestCase{}
)

// nameAsIPTestCase exercises NameAsIP: a parseable address is bracketed and
// accepted, anything else is rejected. Acceptance and a non-empty result are
// inseparable, so ok is asserted as the invariant expected != "" rather than
// declared as a redundant column.
type nameAsIPTestCase struct {
	input    string
	expected string
	name     string
}

func (tc nameAsIPTestCase) Name() string { return tc.name }

func (tc nameAsIPTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := x509utils.NameAsIP(tc.input)
	core.AssertEqual(t, tc.expected, got, "addr")
	core.AssertEqual(t, tc.expected != "", ok, "ok")
}

func newNameAsIPTestCase(name, input, expected string) nameAsIPTestCase {
	return nameAsIPTestCase{
		input:    input,
		expected: expected,
		name:     name,
	}
}

func nameAsIPTestCases() []nameAsIPTestCase {
	return core.S(
		newNameAsIPTestCase("hostname", "a.b.c", ""),
		newNameAsIPTestCase("bare zero", "0", "[0.0.0.0]"),
		newNameAsIPTestCase("ipv4", "1.2.3.4", "[1.2.3.4]"),
		newNameAsIPTestCase("overflow octet", "1.2.3.400", ""),
		newNameAsIPTestCase("ipv6 unspecified", "::", "[::]"),
		newNameAsIPTestCase("fqdn", "foo.example.org", ""),
	)
}

func TestNameAsIP(t *testing.T) {
	core.RunTestCases(t, nameAsIPTestCases())
}

// nameAsSuffixTestCase exercises NameAsSuffix: the name is reduced to the
// dotted suffix after its first label, accepted only when that suffix is
// longer than a lone dot. ok is therefore the invariant len(expected) > 1, not
// expected != "": "a." yields "." yet is rejected.
type nameAsSuffixTestCase struct {
	input    string
	expected string
	name     string
}

func (tc nameAsSuffixTestCase) Name() string { return tc.name }

func (tc nameAsSuffixTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := x509utils.NameAsSuffix(tc.input)
	core.AssertEqual(t, tc.expected, got, "suffix")
	core.AssertEqual(t, len(tc.expected) > 1, ok, "ok")
}

func newNameAsSuffixTestCase(name, input,
	expected string) nameAsSuffixTestCase {
	return nameAsSuffixTestCase{
		input:    input,
		expected: expected,
		name:     name,
	}
}

func nameAsSuffixTestCases() []nameAsSuffixTestCase {
	return core.S(
		newNameAsSuffixTestCase("dotted host", "foo.example.com",
			".example.com"),
		newNameAsSuffixTestCase("leading dot", ".example.com", ""),
		newNameAsSuffixTestCase("two labels", "a.b.c", ".b.c"),
		newNameAsSuffixTestCase("leading dot two labels", ".b.c", ""),
		newNameAsSuffixTestCase("single inner label", "b.c", ".c"),
		newNameAsSuffixTestCase("leading dot one label", ".c", ""),
		newNameAsSuffixTestCase("bare label", "c", ""),
		// idx>0 yields a lone dot, which fails the len>1 gate: a non-empty
		// result with ok=false, the row that makes the invariant load-bearing.
		newNameAsSuffixTestCase("trailing dot", "a.", "."),
		newNameAsSuffixTestCase("empty", "", ""),
	)
}

func TestNameAsSuffix(t *testing.T) {
	core.RunTestCases(t, nameAsSuffixTestCases())
}
