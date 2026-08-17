package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

// Pin the aliased mode bits as constants: shamash relies on their
// constant-ness to fold them into a constant expression, so a slip to a
// var would break downstream builds rather than these tests.
const _ fs.FileMode = fs.ModeDir | fs.ModeAppend | fs.ModeExclusive |
	fs.ModeTemporary | fs.ModeSymlink | fs.ModeDevice |
	fs.ModeNamedPipe | fs.ModeSocket | fs.ModeSetuid |
	fs.ModeSetgid | fs.ModeCharDevice | fs.ModeSticky |
	fs.ModeIrregular | fs.ModeType | fs.ModePerm

// Compile-time verification that test case types implement TestCase.
var (
	_ core.TestCase = validPathTestCase{}
	_ core.TestCase = readFileTestCase{}
	_ core.TestCase = statTestCase{}
	_ core.TestCase = readDirTestCase{}
)

// newTreeFS builds a small real directory tree under a temporary root and
// returns it as an [fs.FS] for the proxy tests to walk.
func newTreeFS(t *testing.T) fs.FS {
	t.Helper()
	root := t.TempDir()
	writeTreeFile(t, root, "hello.txt", "hello")
	writeTreeFile(t, filepath.Join(root, "sub"), "world.txt", "world")
	writeTreeFile(t, filepath.Join(root, "sub", "deep"), "leaf.txt", "leaf")
	return os.DirFS(root)
}

// writeTreeFile creates dir (with parents) and writes content into name.
func writeTreeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	core.AssertMustNoError(t, os.MkdirAll(dir, 0o755), "mkdir %q", dir)
	path := filepath.Join(dir, name)
	core.AssertMustNoError(t, os.WriteFile(path, []byte(content), 0o644),
		"write %q", path)
}

// dirEntryNames extracts the entry names in the order given.
func dirEntryNames(entries []fs.DirEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

// validPathTestCase exercises the ValidPath proxy.
type validPathTestCase struct {
	name  string
	path  string
	valid bool
}

func newValidPathTestCase(name, path string, valid bool) validPathTestCase {
	return validPathTestCase{
		name:  name,
		path:  path,
		valid: valid,
	}
}

func (tc validPathTestCase) Name() string { return tc.name }

func (tc validPathTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.valid, fs.ValidPath(tc.path), "ValidPath(%q)", tc.path)
}

func validPathTestCases() []validPathTestCase {
	return []validPathTestCase{
		newValidPathTestCase("dot", ".", true),
		newValidPathTestCase("file", "hello.txt", true),
		newValidPathTestCase("nested", "sub/world.txt", true),
		newValidPathTestCase("empty", "", false),
		newValidPathTestCase("root", "/", false),
		newValidPathTestCase("rooted", "/hello", false),
		newValidPathTestCase("trailing slash", "sub/", false),
		newValidPathTestCase("dot element", "./hello", false),
		newValidPathTestCase("parent element", "../hello", false),
		newValidPathTestCase("empty element", "sub//world", false),
	}
}

func TestValidPath(t *testing.T) {
	core.RunTestCases(t, validPathTestCases())
}

// readFileTestCase exercises the ReadFile proxy against a real tree.
type readFileTestCase struct {
	fsys    fs.FS
	name    string
	path    string
	want    string
	wantErr bool
}

func newReadFileTestCase(fsys fs.FS, name, path, want string,
	wantErr bool) readFileTestCase {
	return readFileTestCase{
		fsys:    fsys,
		name:    name,
		path:    path,
		want:    want,
		wantErr: wantErr,
	}
}

func (tc readFileTestCase) Name() string { return tc.name }

func (tc readFileTestCase) Test(t *testing.T) {
	t.Helper()
	got, err := fs.ReadFile(tc.fsys, tc.path)
	if tc.wantErr {
		core.AssertError(t, err, "ReadFile(%q)", tc.path)
		return
	}
	core.AssertMustNoError(t, err, "ReadFile(%q)", tc.path)
	core.AssertEqual(t, tc.want, string(got), "content of %q", tc.path)
}

func readFileTestCases(fsys fs.FS) []readFileTestCase {
	return []readFileTestCase{
		newReadFileTestCase(fsys, "root file", "hello.txt", "hello", false),
		newReadFileTestCase(fsys, "nested file", "sub/world.txt", "world", false),
		newReadFileTestCase(fsys, "deep file", "sub/deep/leaf.txt", "leaf", false),
		newReadFileTestCase(fsys, "absent", "nope.txt", "", true),
	}
}

func TestReadFile(t *testing.T) {
	core.RunTestCases(t, readFileTestCases(newTreeFS(t)))
}

// statTestCase exercises the Stat proxy against a real tree.
type statTestCase struct {
	fsys     fs.FS
	name     string
	path     string
	wantName string
	wantSize int64
	wantDir  bool
	wantErr  bool
}

func newStatFileTestCase(fsys fs.FS, name, path, wantName string,
	wantSize int64) statTestCase {
	return statTestCase{
		fsys:     fsys,
		name:     name,
		path:     path,
		wantName: wantName,
		wantSize: wantSize,
	}
}

func newStatDirTestCase(fsys fs.FS, name, path, wantName string) statTestCase {
	return statTestCase{
		fsys:     fsys,
		name:     name,
		path:     path,
		wantName: wantName,
		wantDir:  true,
	}
}

func newStatErrorTestCase(fsys fs.FS, name, path string) statTestCase {
	return statTestCase{
		fsys:    fsys,
		name:    name,
		path:    path,
		wantErr: true,
	}
}

func (tc statTestCase) Name() string { return tc.name }

func (tc statTestCase) Test(t *testing.T) {
	t.Helper()
	fi, err := fs.Stat(tc.fsys, tc.path)
	if tc.wantErr {
		core.AssertError(t, err, "Stat(%q)", tc.path)
		return
	}
	core.AssertMustNoError(t, err, "Stat(%q)", tc.path)
	core.AssertEqual(t, tc.wantName, fi.Name(), "name")
	core.AssertEqual(t, tc.wantDir, fi.IsDir(), "isDir")
	if !tc.wantDir {
		core.AssertEqual(t, tc.wantSize, fi.Size(), "size")
	}
}

func statTestCases(fsys fs.FS) []statTestCase {
	return []statTestCase{
		newStatFileTestCase(fsys, "file", "hello.txt", "hello.txt", 5),
		newStatDirTestCase(fsys, "directory", "sub", "sub"),
		newStatDirTestCase(fsys, "deep directory", "sub/deep", "deep"),
		newStatErrorTestCase(fsys, "absent", "nope"),
	}
}

func TestStat(t *testing.T) {
	core.RunTestCases(t, statTestCases(newTreeFS(t)))
}

// readDirTestCase exercises the ReadDir proxy against a real tree.
type readDirTestCase struct {
	fsys    fs.FS
	name    string
	dir     string
	want    []string
	wantErr bool
}

func newReadDirTestCase(fsys fs.FS, name, dir string, want []string,
	wantErr bool) readDirTestCase {
	return readDirTestCase{
		fsys:    fsys,
		name:    name,
		dir:     dir,
		want:    want,
		wantErr: wantErr,
	}
}

func (tc readDirTestCase) Name() string { return tc.name }

func (tc readDirTestCase) Test(t *testing.T) {
	t.Helper()
	entries, err := fs.ReadDir(tc.fsys, tc.dir)
	if tc.wantErr {
		core.AssertError(t, err, "ReadDir(%q)", tc.dir)
		return
	}
	core.AssertMustNoError(t, err, "ReadDir(%q)", tc.dir)
	core.AssertSliceEqual(t, tc.want, dirEntryNames(entries), "names")
}

func readDirTestCases(fsys fs.FS) []readDirTestCase {
	return []readDirTestCase{
		newReadDirTestCase(fsys, "root", ".", core.S("hello.txt", "sub"), false),
		newReadDirTestCase(fsys, "sub", "sub", core.S("deep", "world.txt"), false),
		newReadDirTestCase(fsys, "deep", "sub/deep", core.S("leaf.txt"), false),
		newReadDirTestCase(fsys, "absent", "nope", nil, true),
	}
}

func TestReadDir(t *testing.T) {
	core.RunTestCases(t, readDirTestCases(newTreeFS(t)))
}
