package fs_test

import (
	"path/filepath"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var (
	_ core.TestCase = cleanOSPathTestCase{}
	_ core.TestCase = cleanOSPathUnchangedTestCase{}
	_ core.TestCase = cleanOSPathInvalidTestCase{}
)

type cleanOSPathTestCase struct {
	name string
	path string
	out  string
}

func (tc cleanOSPathTestCase) Name() string {
	return tc.name
}

func (tc cleanOSPathTestCase) Test(t *testing.T) {
	t.Helper()

	out, err := fs.CleanOSPath(tc.path)
	core.AssertNoError(t, err, "clean")
	core.AssertEqual(t, tc.out, out, "cleaned")
}

func newCleanOSPathTestCase(name, path, out string) cleanOSPathTestCase {
	return cleanOSPathTestCase{
		name: name,
		path: path,
		out:  out,
	}
}

// newCleanOSPathSlashCase makes a row written with '/', expecting out
// with the platform's separator.
func newCleanOSPathSlashCase(name, path, out string) cleanOSPathTestCase {
	return newCleanOSPathTestCase(name, path, filepath.FromSlash(out))
}

func cleanOSPathTestCases() []cleanOSPathTestCase {
	return append([]cleanOSPathTestCase{
		// Nothing left after cleaning.
		newCleanOSPathSlashCase("dot", ".", "."),
		newCleanOSPathSlashCase("dot and separator", "./", "."),
		newCleanOSPathSlashCase("root", "/", "/"),
		newCleanOSPathSlashCase("root and dot", "/./", "/"),
		// Separators and dot segments.
		newCleanOSPathSlashCase("single component", "a", "a"),
		newCleanOSPathSlashCase("already clean", "/srv/www/index.html", "/srv/www/index.html"),
		newCleanOSPathSlashCase("doubled separator", "a//b", "a/b"),
		newCleanOSPathSlashCase("leading dot", "./a", "a"),
		newCleanOSPathSlashCase("inner dot", "a/./b", "a/b"),
		// A dot dot goes with the segment before it, and stays without one.
		newCleanOSPathSlashCase("dot dot", "..", ".."),
		newCleanOSPathSlashCase("leading dot dot", "../a", "../a"),
		newCleanOSPathSlashCase("dot then dot dot", "./../a", "../a"),
		newCleanOSPathSlashCase("two leading dot dots", "../..", "../.."),
		newCleanOSPathSlashCase("inner dot dot", "a/../b", "b"),
		newCleanOSPathSlashCase("dot dot past the start", "a/../../b", "../b"),
		newCleanOSPathSlashCase("dot dot after a leading dot dot", "../a/..", ".."),
		newCleanOSPathSlashCase("dot dot after a changed dot dot", "./../..", "../.."),
		// A dot dot after the root goes, the root being its own parent.
		newCleanOSPathSlashCase("rooted dot dot", "/../a", "/a"),
		newCleanOSPathSlashCase("root and dot dot", "/..", "/"),
		newCleanOSPathSlashCase("rooted dot dot past a name", "/a/../../b", "/b"),
		// A trailing separator, or a dot or dot dot after one, stays as one
		// separator, unless the path ends in a dot or dot dot.
		newCleanOSPathSlashCase("trailing separator", "a/", "a/"),
		newCleanOSPathSlashCase("doubled trailing separator", "a//", "a/"),
		newCleanOSPathSlashCase("trailing dot", "a/.", "a/"),
		newCleanOSPathSlashCase("trailing dot and separator", "a/./", "a/"),
		newCleanOSPathSlashCase("trailing dot dot", "a/b/..", "a/"),
		newCleanOSPathSlashCase("trailing dot dot and separator", "a/b/../", "a/"),
		newCleanOSPathSlashCase("trailing dot dot after two names", "../a/b/..", "../a/"),
		newCleanOSPathSlashCase("three dots", "a/.../..", "a/"),
		newCleanOSPathSlashCase("trailing dot dot after a changed path", "a//bc/..", "a/"),
		newCleanOSPathSlashCase("dot dot to nothing", "a/..", "."),
		newCleanOSPathSlashCase("dot dot to nothing and separator", "a/../", "."),
		newCleanOSPathSlashCase("dot dot and separator", "../", ".."),
		newCleanOSPathSlashCase("dot dots and separator", "../../", "../.."),
		newCleanOSPathSlashCase("rooted dot dot and separator", "/a/../", "/"),
	}, cleanOSPathPlatformCases()...)
}

func TestCleanOSPath(t *testing.T) {
	core.RunTestCases(t, cleanOSPathTestCases())
}

// cleanOSPathUnchangedTestCase is a clean path that CleanOSPath gives
// back unchanged and without allocating.
type cleanOSPathUnchangedTestCase struct {
	name string
	path string
}

func (tc cleanOSPathUnchangedTestCase) Name() string {
	return tc.name
}

func (tc cleanOSPathUnchangedTestCase) Test(t *testing.T) {
	t.Helper()

	var out string
	var err error
	allocs := testing.AllocsPerRun(10, func() {
		out, err = fs.CleanOSPath(tc.path)
	})
	core.AssertNoError(t, err, "clean")
	core.AssertEqual(t, tc.path, out, "cleaned")
	core.AssertEqual(t, 0.0, allocs, "allocations")
}

func newCleanOSPathUnchangedTestCase(name, path string) cleanOSPathUnchangedTestCase {
	return cleanOSPathUnchangedTestCase{
		name: name,
		path: path,
	}
}

// newCleanOSPathUnchangedSlashCase makes a row written with '/', for a
// path in the platform's separator.
func newCleanOSPathUnchangedSlashCase(name, path string) cleanOSPathUnchangedTestCase {
	return newCleanOSPathUnchangedTestCase(name, filepath.FromSlash(path))
}

func cleanOSPathUnchangedTestCases() []cleanOSPathUnchangedTestCase {
	return append([]cleanOSPathUnchangedTestCase{
		newCleanOSPathUnchangedSlashCase("dot", "."),
		newCleanOSPathUnchangedSlashCase("dot dot", ".."),
		newCleanOSPathUnchangedSlashCase("root", "/"),
		newCleanOSPathUnchangedSlashCase("single component", "a"),
		newCleanOSPathUnchangedSlashCase("rooted", "/srv/www/index.html"),
		newCleanOSPathUnchangedSlashCase("leading dot dots", "../../a"),
		newCleanOSPathUnchangedSlashCase("trailing separator", "a/"),
	}, cleanOSPathUnchangedPlatformCases()...)
}

func TestCleanOSPathUnchanged(t *testing.T) {
	core.RunTestCases(t, cleanOSPathUnchangedTestCases())
}

// cleanOSPathInvalidTestCase is a path the os functions refuse, which
// CleanOSPath refuses as a PathError wrapping ErrInvalid.
type cleanOSPathInvalidTestCase struct {
	name string
	path string
}

func (tc cleanOSPathInvalidTestCase) Name() string {
	return tc.name
}

func (tc cleanOSPathInvalidTestCase) Test(t *testing.T) {
	t.Helper()

	out, err := fs.CleanOSPath(tc.path)
	core.AssertEqual(t, "", out, "cleaned")
	core.AssertErrorIs(t, err, fs.ErrInvalid, "error")
	pathErr := core.AssertMustErrorAs[*fs.PathError](t, err, "path error")
	core.AssertEqual(t, "clean", pathErr.Op, "op")
	core.AssertEqual(t, tc.path, pathErr.Path, "path")
}

func newCleanOSPathInvalidTestCase(name, path string) cleanOSPathInvalidTestCase {
	return cleanOSPathInvalidTestCase{
		name: name,
		path: path,
	}
}

func TestCleanOSPathInvalid(t *testing.T) {
	core.RunTestCases(t, append([]cleanOSPathInvalidTestCase{
		newCleanOSPathInvalidTestCase("empty", ""),
		newCleanOSPathInvalidTestCase("NUL alone", "\x00"),
		newCleanOSPathInvalidTestCase("NUL in a name", "a\x00b"),
		newCleanOSPathInvalidTestCase("NUL after a dot segment", "./\x00"),
		newCleanOSPathInvalidTestCase("NUL in a share name", `\\host\sh`+"\x00"+`re\a`),
		newCleanOSPathInvalidTestCase(`NUL after \\?\`, `\\?\C:\a`+"\x00"),
	}, cleanOSPathInvalidPlatformCases()...))
}
