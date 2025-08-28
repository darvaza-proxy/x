package fs

import (
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = joinRunesTestCase{}
var _ core.TestCase = unsafeCutRootTestCase{}

// joinRunes test cases
type joinRunesTestCase struct {
	before   []rune
	after    []rune
	expected []rune
	name     string
}

func (tc joinRunesTestCase) Name() string {
	return tc.name
}

func (tc joinRunesTestCase) Test(t *testing.T) {
	t.Helper()

	result := joinRunes(tc.before, tc.after)
	core.AssertSliceEqual(t, tc.expected, result, "joined runes")
}

func newJoinRunesTestCase(name string, before, after, expected []rune) joinRunesTestCase {
	return joinRunesTestCase{
		name:     name,
		before:   before,
		after:    after,
		expected: expected,
	}
}

func joinRunesTestCases() []joinRunesTestCase {
	return []joinRunesTestCase{
		newJoinRunesTestCase("both empty", []rune{}, []rune{}, []rune{}),
		newJoinRunesTestCase("empty before", []rune{}, []rune("after"), []rune("after")),
		newJoinRunesTestCase("empty after", []rune("before"), []rune{}, []rune("before")),
		newJoinRunesTestCase("both non-empty", []rune("before"), []rune("after"), []rune("before/after")),
		newJoinRunesTestCase("single char before", []rune("a"), []rune("b"), []rune("a/b")),
		newJoinRunesTestCase("single char after", []rune("before"), []rune("x"), []rune("before/x")),
		newJoinRunesTestCase("unicode characters", []rune("café"), []rune("naïve"), []rune("café/naïve")),
		newJoinRunesTestCase("path components", []rune("usr/bin"), []rune("ls"), []rune("usr/bin/ls")),
	}
}

func TestJoinRunes(t *testing.T) {
	core.RunTestCases(t, joinRunesTestCases())
}

// unsafeCutRoot test cases
type unsafeCutRootTestCase struct {
	name     string
	root     string
	expected string
	wantOK   bool
	testName string
}

func (tc unsafeCutRootTestCase) Name() string {
	return tc.testName
}

func (tc unsafeCutRootTestCase) Test(t *testing.T) {
	t.Helper()

	result, ok := unsafeCutRoot(tc.name, tc.root)

	core.AssertEqual(t, tc.wantOK, ok, "cut success")
	if tc.wantOK {
		core.AssertEqual(t, tc.expected, result, "cut result")
	}
}

func newUnsafeCutRootTestCase(testName, name, root, expected string, wantOK bool) unsafeCutRootTestCase {
	return unsafeCutRootTestCase{
		testName: testName,
		name:     name,
		root:     root,
		expected: expected,
		wantOK:   wantOK,
	}
}

func unsafeCutRootTestCases() []unsafeCutRootTestCase {
	return []unsafeCutRootTestCase{
		// Exact match cases - core functionality
		newUnsafeCutRootTestCase("exact match", "path", "path", ".", true),
		newUnsafeCutRootTestCase("single char match", "a", "a", ".", true),

		// Dot root cases - special handling for current directory
		newUnsafeCutRootTestCase("dot root with path", "path/to/file", ".", "path/to/file", true),
		newUnsafeCutRootTestCase("dot root with dot name", ".", ".", ".", true),

		// Valid prefix matching - expected usage patterns
		newUnsafeCutRootTestCase("simple prefix", "usr/bin/ls", "usr", "bin/ls", true),
		newUnsafeCutRootTestCase("nested prefix", "home/user/docs/file.txt", "home/user", "docs/file.txt", true),
		newUnsafeCutRootTestCase("root with trailing content", "usr/bin", "usr", "bin", true),
		newUnsafeCutRootTestCase("deep nested", "a/b/c/d/e", "a/b", "c/d/e", true),

		// Edge case where path equals root with slash
		newUnsafeCutRootTestCase("root equals name plus slash", "usr/", "usr", ".", true),

		// Non-matching cases - function should return false
		newUnsafeCutRootTestCase("no prefix match", "usr/bin", "home", "", false),
		newUnsafeCutRootTestCase("partial string match", "usr-data", "usr", "", false),
		newUnsafeCutRootTestCase("shorter name than root", "usr", "usr/bin", "", false),
		newUnsafeCutRootTestCase("empty name non-empty root", "", "root", "", false),
	}
}

func TestUnsafeCutRoot(t *testing.T) {
	core.RunTestCases(t, unsafeCutRootTestCases())
}
