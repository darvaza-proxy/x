package web

import (
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = cleanTestCase{}

type cleanTestCase struct {
	name string
	path string
	out  string
	ok   bool
}

func (tc cleanTestCase) Name() string {
	return tc.name
}

func (tc cleanTestCase) Test(t *testing.T) {
	t.Helper()

	s, ok := Clean(tc.path)
	core.AssertEqual(t, tc.out, s, "cleaned")
	core.AssertEqual(t, tc.ok, ok, "ok")
}

func newCleanTestCase(name, path, out string, ok bool) cleanTestCase {
	return cleanTestCase{
		name: name,
		path: path,
		out:  out,
		ok:   ok,
	}
}

func cleanTestCases() []cleanTestCase {
	return []cleanTestCase{
		// fs.Clean parity for ordinary paths.
		newCleanTestCase("empty", "", ".", true),
		newCleanTestCase("dot", ".", ".", true),
		newCleanTestCase("single component", "a", "a", true),
		newCleanTestCase("two components", "a/b", "a/b", true),
		newCleanTestCase("double slash", "a//b", "a/b", true),
		newCleanTestCase("self-cancel", "a/..", ".", true),
		newCleanTestCase("cancel and continue", "a/../b", "b", true),
		newCleanTestCase("root", "/", "/", true),
		newCleanTestCase("rooted single", "/a", "/a", true),
		newCleanTestCase("rooted trailing slash", "/a/", "/a/", true),
		newCleanTestCase("rooted cancel to root", "/a/..", "/", true),
		newCleanTestCase("rooted two", "/a/b", "/a/b", true),
		newCleanTestCase("rooted cancel", "/a/b/..", "/a", true),
		newCleanTestCase("rooted cancel and continue", "/a/b/../foo", "/a/foo", true),
		// WHATWG backslash normalisation: \ becomes / before fs.Clean.
		newCleanTestCase("backslash root", `/\`, "/", true),
		newCleanTestCase("protocol-relative backslash", `/\evil.com`, "/evil.com", true),
		newCleanTestCase("protocol-relative backslash with path",
			`/\evil.com/path`, "/evil.com/path", true),
		newCleanTestCase("dot-backslash bypass", `/./\evil.com`, "/evil.com", true),
		newCleanTestCase("dotdot-backslash bypass",
			`/a/../\evil.com`, "/evil.com", true),
		// Protocol-relative shapes flatten via fs.Clean's empty-leading drop.
		newCleanTestCase("double slash root", "//", "/", true),
		newCleanTestCase("protocol-relative slash", "//evil.com", "/evil.com", true),
		newCleanTestCase("protocol-relative slash with path",
			"//evil.com/path", "/evil.com/path", true),
		// /.. escape stripping — ok=false signals discarded blocks.
		newCleanTestCase("escape root", "/..", "/", false),
		newCleanTestCase("escape root trailing", "/../", "/", false),
		newCleanTestCase("escape root twice", "/../..", "/", false),
		newCleanTestCase("escape then path", "/../evil.com", "/evil.com", false),
		newCleanTestCase("escape then path with trailing",
			"/../foo/", "/foo/", false),
		newCleanTestCase("escape twice then path",
			"/../../evil.com", "/evil.com", false),
		newCleanTestCase("escape three deep",
			"/../../../x", "/x", false),
		newCleanTestCase("escape with trailing double slash",
			"/../..//", "/", false),
		newCleanTestCase("dotdot prefix of filename",
			"/..foo", "/..foo", true),
		// Relative escape: fs.Clean leaves it; we don't strip (not /..).
		newCleanTestCase("relative dotdot", "..", "..", true),
		newCleanTestCase("relative dotdot trailing", "../", "../", true),
		newCleanTestCase("relative dotdot with path", "../a", "../a", true),
		// ok invariant: non-/.. reductions keep ok=true even
		// when the output differs from the input.
		newCleanTestCase("ok-invariant backslash normalisation",
			`/a\b`, "/a/b", true),
		newCleanTestCase("ok-invariant fs.Clean reduction",
			"/a/b/../c", "/a/c", true),
		// Pathological inputs: documented, not guarded against.
		newCleanTestCase("carriage return passes through",
			"/a\rb", "/a\rb", true),
		newCleanTestCase("newline passes through",
			"/a\nb", "/a\nb", true),
		newCleanTestCase("null byte passes through",
			"/a\x00b", "/a\x00b", true),
		newCleanTestCase("url-encoded dotdot not decoded",
			"/%2e%2e/evil.com", "/%2e%2e/evil.com", true),
		newCleanTestCase("full URL input mangled",
			"http://evil.com/foo", "http:/evil.com/foo", true),
	}
}

func TestClean(t *testing.T) {
	core.RunTestCases(t, cleanTestCases())
}
