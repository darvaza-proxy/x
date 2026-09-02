package fs_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var _ core.TestCase = splitTestCase{}

type splitTestCase struct {
	name string
	path string
	dir  string
	file string
}

func (tc splitTestCase) Name() string {
	return tc.name
}

func (tc splitTestCase) Test(t *testing.T) {
	t.Helper()

	dir, file := fs.Split(tc.path)
	core.AssertEqual(t, tc.dir, dir, "dir")
	core.AssertEqual(t, tc.file, file, "file")
}

func newSplitTestCase(name, path, dir, file string) splitTestCase {
	return splitTestCase{
		name: name,
		path: path,
		dir:  dir,
		file: file,
	}
}

func splitTestCases() []splitTestCase {
	return []splitTestCase{
		newSplitTestCase("single component", "a", ".", "a"),
		newSplitTestCase("two components", "a/b", "a", "b"),
		newSplitTestCase("three components", "a/b/c", "a/b", "c"),
		newSplitTestCase("multi-character components", "aa/bb/cc", "aa/bb", "cc"),
		newSplitTestCase("rooted single", "/a", "", "a"),
		newSplitTestCase("rooted with noise", "/a//./b/c", "/a/b", "c"),
		newSplitTestCase("rooted with cancel", "/a/../b/c", "/b", "c"),
	}
}

func TestSplit(t *testing.T) {
	core.RunTestCases(t, splitTestCases())
}
