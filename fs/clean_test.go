package fs_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs"
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

	s, ok := fs.Clean(tc.path)
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
		// Relative paths that stay within their root.
		newCleanTestCase("empty", "", ".", true),
		newCleanTestCase("dot", ".", ".", true),
		newCleanTestCase("single component", "a", "a", true),
		newCleanTestCase("two components", "a/b", "a/b", true),
		newCleanTestCase("double slash", "a//b", "a/b", true),
		newCleanTestCase("self-cancel", "a/..", ".", true),
		newCleanTestCase("cancel and continue", "a/../b", "b", true),
		newCleanTestCase("cancel last", "a/b/..", "a", true),
		newCleanTestCase("cancel last and continue", "a/b/../foo", "a/foo", true),
		// Relative paths that escape their root.
		newCleanTestCase("dotdot", "..", "..", false),
		newCleanTestCase("dotdot trailing slash", "../", "..", false),
		newCleanTestCase("dotdot then dot", "../.", "..", false),
		newCleanTestCase("dotdot then component", "../a", "../a", false),
		// Rooted paths, never valid for fs.FS.
		newCleanTestCase("root", "/", "/", false),
		newCleanTestCase("rooted single", "/a", "/a", false),
		newCleanTestCase("rooted cancel to root", "/a/..", "/", false),
		newCleanTestCase("rooted escape twice", "/../..//", "/../..", false),
	}
}

func TestClean(t *testing.T) {
	core.RunTestCases(t, cleanTestCases())
}
