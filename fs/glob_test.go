package fs_test

import (
	"testing"
	"testing/fstest"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var (
	_ core.TestCase = globMatchTestCase{}
	_ core.TestCase = globCompileErrorTestCase{}
	_ core.TestCase = globTestCase{}
)

// globMatchTestCase compiles a pattern through GlobCompile and checks
// whether its Matcher matches a path.
type globMatchTestCase struct {
	name    string
	pattern string
	path    string
	want    bool
}

func newGlobMatchTestCase(name, pattern, path string,
	want bool) globMatchTestCase {
	return globMatchTestCase{
		name:    name,
		pattern: pattern,
		path:    path,
		want:    want,
	}
}

// newGlobMatchesTestCase is a row whose pattern matches its path.
func newGlobMatchesTestCase(name, pattern, path string) globMatchTestCase {
	return newGlobMatchTestCase(name, pattern, path, true)
}

// newGlobMissesTestCase is a row whose pattern does not match its path.
func newGlobMissesTestCase(name, pattern, path string) globMatchTestCase {
	return newGlobMatchTestCase(name, pattern, path, false)
}

func (tc globMatchTestCase) Name() string {
	return tc.name
}

func (tc globMatchTestCase) Test(t *testing.T) {
	t.Helper()

	globs, err := fs.GlobCompile(tc.pattern)
	core.AssertMustNoError(t, err, "GlobCompile")
	core.AssertMustEqual(t, 1, len(globs), "matchers")
	core.AssertEqual(t, tc.want, globs[0].Match(tc.path),
		"%q matches %q", tc.pattern, tc.path)
}

func globMatchTestCases() []globMatchTestCase {
	return []globMatchTestCase{
		// Literals.
		newGlobMatchesTestCase("literal", "a/b", "a/b"),
		newGlobMissesTestCase("literal prefix", "a/b", "a/bc"),
		// A star stays within one component.
		newGlobMatchesTestCase("star", "*.txt", "a.txt"),
		newGlobMissesTestCase("star across separator", "*.txt", "dir/a.txt"),
		newGlobMatchesTestCase("lone star", "*", "a"),
		newGlobMissesTestCase("lone star across separator", "*", "a/b"),
		// A question mark is one character other than the separator.
		newGlobMatchesTestCase("question mark", "?.go", "a.go"),
		newGlobMissesTestCase("question mark, two characters", "?.go",
			"ab.go"),
		newGlobMissesTestCase("question mark on separator", "a?b", "a/b"),
		// A double star crosses separators.
		newGlobMatchesTestCase("lone double star", "**", "a/b/c"),
		newGlobMatchesTestCase("double star, one level", "**/*.go",
			"a/b.go"),
		newGlobMatchesTestCase("double star, two levels", "**/*.go",
			"a/b/c.go"),
		newGlobMissesTestCase("double star, top level", "**/*.go", "b.go"),
		newGlobMatchesTestCase("inner double star", "data/**/*.json",
			"data/x/y.json"),
		newGlobMissesTestCase("inner double star, no level between",
			"data/**/*.json", "data/y.json"),
		newGlobMissesTestCase("inner double star, other root",
			"data/**/*.json", "other/x/y.json"),
		// Braces list alternatives.
		newGlobMatchesTestCase("brace, first", "*.{go,md}", "x.go"),
		newGlobMatchesTestCase("brace, second", "*.{go,md}", "x.md"),
		newGlobMissesTestCase("brace, neither", "*.{go,md}", "x.txt"),
		// An unclosed brace reads as closed.
		newGlobMatchesTestCase("unclosed brace", "{a,b", "a"),
		// Character classes.
		newGlobMatchesTestCase("class", "[ab].go", "a.go"),
		newGlobMissesTestCase("class, outside", "[ab].go", "c.go"),
		newGlobMatchesTestCase("negated class", "[!ab].go", "c.go"),
		newGlobMissesTestCase("negated class, inside", "[!ab].go", "a.go"),
		newGlobMatchesTestCase("range", "[a-c].go", "b.go"),
		newGlobMissesTestCase("range, outside", "[a-c].go", "d.go"),
	}
}

// TestGlobCompileMatch covers what a Matcher from GlobCompile matches,
// with '/' as the separator.
func TestGlobCompileMatch(t *testing.T) {
	core.RunTestCases(t, globMatchTestCases())
}

// globCompileErrorTestCase checks GlobCompile rejects patterns it cannot
// compile, returning no matchers at all.
type globCompileErrorTestCase struct {
	name     string
	patterns []string
}

func newGlobCompileErrorTestCase(name string,
	patterns ...string) globCompileErrorTestCase {
	return globCompileErrorTestCase{
		name:     name,
		patterns: patterns,
	}
}

func (tc globCompileErrorTestCase) Name() string {
	return tc.name
}

func (tc globCompileErrorTestCase) Test(t *testing.T) {
	t.Helper()

	globs, err := fs.GlobCompile(tc.patterns...)
	core.AssertError(t, err, "GlobCompile")
	core.AssertNil(t, globs, "matchers")
}

// TestGlobCompileError covers patterns GlobCompile rejects.
func TestGlobCompileError(t *testing.T) {
	core.RunTestCases(t, []globCompileErrorTestCase{
		newGlobCompileErrorTestCase("unclosed class", "[ab"),
		newGlobCompileErrorTestCase("one invalid among valid",
			"*.go", "[ab"),
	})
}

// globTestCase checks the entries Glob finds in newGlobFS for a list
// of patterns.
type globTestCase struct {
	name     string
	patterns []string
	want     []string
}

func newGlobTestCase(name string, patterns, want []string) globTestCase {
	return globTestCase{
		name:     name,
		patterns: patterns,
		want:     want,
	}
}

func (tc globTestCase) Name() string {
	return tc.name
}

func (tc globTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := fs.Glob(newGlobFS(), tc.patterns...)
	core.AssertMustNoError(t, err, "Glob")
	core.AssertSliceEqual(t, tc.want, got, "Glob")
}

// newGlobFS returns the tree the Glob rows search. fstest.MapFS
// provides the directories its files imply.
func newGlobFS() fstest.MapFS {
	return fstest.MapFS{
		"a.txt":        {},
		"b.go":         {},
		"dir/c.go":     {},
		"dir/e.json":   {},
		"dir/sub/d.go": {},
	}
}

func globTestCases() []globTestCase {
	return []globTestCase{
		newGlobTestCase("one pattern", core.S("*.go"), core.S("b.go")),
		newGlobTestCase("any of several", core.S("*.txt", "*.go"),
			core.S("a.txt", "b.go")),
		newGlobTestCase("any depth", core.S("**/*.go"),
			core.S("dir/c.go", "dir/sub/d.go")),
		newGlobTestCase("directories too", core.S("dir/*"),
			core.S("dir/c.go", "dir/e.json", "dir/sub")),
		newGlobTestCase("no match", core.S("*.md"), nil),
		newGlobTestCase("no patterns", nil, core.S(
			".",
			"a.txt",
			"b.go",
			"dir",
			"dir/c.go",
			"dir/e.json",
			"dir/sub",
			"dir/sub/d.go",
		)),
	}
}

// TestGlob covers Glob walking a tree for entries matching any of its
// patterns, or every entry when given none.
func TestGlob(t *testing.T) {
	core.RunTestCases(t, globTestCases())
}
