package x509utils_test

import (
	"bytes"
	"encoding/pem"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = nilOptionTestCase{}
	_ core.TestCase = readDirPEMTestCase{}
	_ core.TestCase = readFilePEMTestCase{}
	_ core.TestCase = readPEMTestCase{}
	_ core.TestCase = readStringPEMTestCase{}
)

// newBlockCounter returns a callback counting the blocks it receives,
// asserting each one is non-nil.
func newBlockCounter(t *testing.T, got *int) x509utils.DecodePEMBlockFunc {
	return func(_ fs.FS, _ string, block *pem.Block) bool {
		core.AssertNotNil(t, block, "block")
		*got++
		return true
	}
}

// checkReadPEMResult asserts the error and delivered-block count of a
// ReadPEM/ReadStringPEM call.
func checkReadPEMResult(t *testing.T, wantErr, err error,
	wantBlocks, got int) {
	t.Helper()

	if wantErr == nil {
		core.AssertNoError(t, err, "read")
	} else {
		core.AssertErrorIs(t, err, wantErr, "error")
	}
	core.AssertEqual(t, wantBlocks, got, "blocks")
}

// readPEMTestCase feeds raw input to ReadPEM and checks the returned
// error and how many blocks reached the callback.
type readPEMTestCase struct {
	wantErr    error
	name       string
	input      string
	wantBlocks int
}

func newReadPEMTestCase(name, input string, wantErr error,
	wantBlocks int) readPEMTestCase {
	return readPEMTestCase{
		name:       name,
		input:      input,
		wantErr:    wantErr,
		wantBlocks: wantBlocks,
	}
}

func (tc readPEMTestCase) Name() string {
	return tc.name
}

func (tc readPEMTestCase) Test(t *testing.T) {
	t.Helper()

	var got int
	err := x509utils.ReadPEM([]byte(tc.input), newBlockCounter(t, &got))
	checkReadPEMResult(t, tc.wantErr, err, tc.wantBlocks, got)
}

func readPEMTestCases() []readPEMTestCase {
	block := string(pem.EncodeToMemory(&pem.Block{
		Type:  "TEST BLOCK",
		Bytes: []byte("test data"),
	}))

	return core.S(
		newReadPEMTestCase("empty input", "", x509utils.ErrEmpty, 0),
		newReadPEMTestCase("garbage only", "no PEM here\n",
			core.ErrInvalid, 0),
		newReadPEMTestCase("single block", block, nil, 1),
		newReadPEMTestCase("trailing blank line", block+"\n", nil, 1),
		newReadPEMTestCase("trailing garbage",
			block+"trailing text\n", nil, 1),
		newReadPEMTestCase("two blocks", block+block, nil, 2),
		newReadPEMTestCase("text between blocks",
			block+"subject=test\n"+block, nil, 2),
	)
}

func TestReadPEM(t *testing.T) {
	core.RunTestCases(t, readPEMTestCases())
}

// readStringPEMTestCase feeds a string to ReadStringPEM and checks the
// returned error and how many blocks reached the callback.
type readStringPEMTestCase struct {
	wantErr    error
	name       string
	input      string
	wantBlocks int
}

func newReadStringPEMTestCase(name, input string, wantErr error,
	wantBlocks int) readStringPEMTestCase {
	return readStringPEMTestCase{
		name:       name,
		input:      input,
		wantErr:    wantErr,
		wantBlocks: wantBlocks,
	}
}

func (tc readStringPEMTestCase) Name() string {
	return tc.name
}

func (tc readStringPEMTestCase) Test(t *testing.T) {
	t.Helper()

	var got int
	err := x509utils.ReadStringPEM(tc.input, newBlockCounter(t, &got))
	checkReadPEMResult(t, tc.wantErr, err, tc.wantBlocks, got)
}

func readStringPEMTestCases(t *testing.T) []readStringPEMTestCase {
	t.Helper()

	block := string(pem.EncodeToMemory(&pem.Block{
		Type:  "TEST BLOCK",
		Bytes: []byte("test data"),
	}))
	missing := filepath.Join(t.TempDir(), "missing.pem")

	return core.S(
		newReadStringPEMTestCase("raw block", block, nil, 1),
		newReadStringPEMTestCase("raw block trailing blank line",
			block+"\n", nil, 1),
		newReadStringPEMTestCase("missing path",
			missing, fs.ErrNotExist, 0),
		// exactly PATH_MAX: the boundary row — a path this long is
		// impossible, so the guard must fire without a stat.
		newReadStringPEMTestCase("overlong non-path",
			strings.Repeat("#", 4096), fs.ErrInvalid, 0),
		newReadStringPEMTestCase("NUL byte non-path",
			"not\x00a-path", fs.ErrInvalid, 0),
	)
}

func TestReadStringPEM(t *testing.T) {
	core.RunTestCases(t, readStringPEMTestCases(t))
}

// TestReadStringPEMNotPathReasons pins the split between the two
// not-a-possible-path guard errors; both satisfy fs.ErrInvalid
// (asserted by the table above) but state different reasons.
func TestReadStringPEMNotPathReasons(t *testing.T) {
	err := x509utils.ReadStringPEM(strings.Repeat("#", 4096), nil)
	core.AssertMustError(t, err, "overlong")
	core.AssertContains(t, err.Error(), "path too long", "overlong reason")
	// the blob is clamped, not echoed whole, into the path error.
	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "overlong path")
	core.AssertContains(t, pe.Path, "… (4096 bytes)", "overlong clamped")
	core.AssertTrue(t, len(pe.Path) < 4096, "overlong bounded")

	err = x509utils.ReadStringPEM("not\x00a-path", nil)
	core.AssertMustError(t, err, "NUL byte")
	core.AssertContains(t, err.Error(), "NUL byte in path", "NUL reason")
}

// TestReadStringPEMDirsDisabled reaches readDirPEM's guard: a path that
// stat resolves to a directory is rejected as invalid, rather than scanned,
// when directory support is turned off.
func TestReadStringPEMDirsDisabled(t *testing.T) {
	fSys := fstest.MapFS{
		"certs/leaf.pem": &fstest.MapFile{Data: []byte("data")},
	}

	err := x509utils.ReadStringPEM("certs", nil,
		x509utils.ReadWithFS(fSys), x509utils.ReadWithoutDirs())

	core.AssertMustError(t, err, "disabled")
	core.AssertErrorIs(t, err, fs.ErrInvalid, "invalid")
	core.AssertContains(t, err.Error(),
		"directories support disabled", "reason")

	// the op stays lowercase, matching fs.PathError convention.
	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "PathError")
	core.AssertEqual(t, "read", pe.Op, "op")
}

// TestReadPEMAbort stops the walk when the callback returns false, leaving the
// remaining blocks undelivered.
func TestReadPEMAbort(t *testing.T) {
	block := pem.EncodeToMemory(&pem.Block{
		Type:  "TEST BLOCK",
		Bytes: []byte("test data"),
	})

	var got int
	cb := func(_ fs.FS, _ string, _ *pem.Block) bool {
		got++
		return false // abort after the first block
	}

	// two identical blocks; the callback must stop the walk at the first.
	err := x509utils.ReadPEM(bytes.Repeat(block, 2), cb)
	core.AssertNoError(t, err, "abort")
	core.AssertEqual(t, 1, got, "blocks")
}

// testPEM encodes a single test PEM block.
func testPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "TEST BLOCK",
		Bytes: []byte("test data"),
	})
}

// pemGoodFS is a small tree of valid PEM files: three blocks in total, one at
// the root and two in a sub-directory, so a recursive read must descend.
func pemGoodFS() fstest.MapFS {
	one := testPEM()
	two := append(testPEM(), testPEM()...)
	return fstest.MapFS{
		"a.pem":       &fstest.MapFile{Data: one},
		"sub/two.pem": &fstest.MapFile{Data: two},
	}
}

// pemMixedFS adds a non-PEM file to the good tree, so a directory read reports
// an error while still delivering the valid blocks.
func pemMixedFS() fstest.MapFS {
	fSys := pemGoodFS()
	fSys["bad.txt"] = &fstest.MapFile{Data: []byte("not pem\n")}
	return fSys
}

// pemBadSubdirFS places the non-PEM file inside the sub-directory, so the error
// surfaces from the recursive descent rather than the top-level files.
func pemBadSubdirFS() fstest.MapFS {
	fSys := pemGoodFS()
	fSys["sub/bad.txt"] = &fstest.MapFile{Data: []byte("not pem\n")}
	return fSys
}

// failReadFS resolves like its embedded MapFS — so stat still reports a file —
// but fails every ReadFile, to exercise the read-error path after a successful
// stat.
type failReadFS struct{ fstest.MapFS }

func (failReadFS) ReadFile(string) ([]byte, error) {
	return nil, fs.ErrPermission
}

// invalidStatFS reports every path as an invalid argument, to drive run's
// normalisation of a stat "invalid" error down to the bare fs sentinel.
type invalidStatFS struct{ fstest.MapFS }

func (invalidStatFS) Stat(string) (fs.FileInfo, error) {
	return nil, &fs.PathError{Op: "stat", Path: "x", Err: fs.ErrInvalid}
}

// invalidPlainFS reports an invalid-wrapping error that is not a PathError, so
// run must leave it intact rather than collapse it to the bare sentinel.
type invalidPlainFS struct{ fstest.MapFS }

func (invalidPlainFS) Stat(string) (fs.FileInfo, error) {
	return nil, core.Wrap(fs.ErrInvalid, "stat quirk")
}

// readFilePEMTestCase feeds a single file to ReadFilePEM and checks the
// delivered block count and error.
type readFilePEMTestCase struct {
	wantErr    error
	fSys       fs.FS
	name       string
	file       string
	wantBlocks int
}

func newReadFilePEMTestCase(name string, fSys fs.FS, file string,
	wantErr error, wantBlocks int) readFilePEMTestCase {
	return readFilePEMTestCase{
		name:       name,
		fSys:       fSys,
		file:       file,
		wantErr:    wantErr,
		wantBlocks: wantBlocks,
	}
}

func (tc readFilePEMTestCase) Name() string { return tc.name }

func (tc readFilePEMTestCase) Test(t *testing.T) {
	t.Helper()

	var got int
	err := x509utils.ReadFilePEM(tc.fSys, tc.file, newBlockCounter(t, &got))
	checkReadPEMResult(t, tc.wantErr, err, tc.wantBlocks, got)
}

func readFilePEMTestCases() []readFilePEMTestCase {
	fSys := pemGoodFS()
	return core.S(
		newReadFilePEMTestCase("single block", fSys, "a.pem", nil, 1),
		newReadFilePEMTestCase("two blocks", fSys, "sub/two.pem", nil, 2),
		newReadFilePEMTestCase("non-PEM file", pemMixedFS(), "bad.txt",
			core.ErrInvalid, 0),
		newReadFilePEMTestCase("missing file", fSys, "missing.pem",
			fs.ErrNotExist, 0),
	)
}

func TestReadFilePEM(t *testing.T) {
	core.RunTestCases(t, readFilePEMTestCases())
}

// TestReadFilePEMDecodeError pins the wrapper ReadFilePEM puts around a decode
// failure: an fs.PathError tagged pem.Decode that still matches core.ErrInvalid.
func TestReadFilePEMDecodeError(t *testing.T) {
	err := x509utils.ReadFilePEM(pemMixedFS(), "bad.txt", nil)

	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "PathError")
	core.AssertEqual(t, "pem.Decode", pe.Op, "op")
	core.AssertErrorIs(t, err, core.ErrInvalid, "invalid")
}

// readDirPEMTestCase feeds a directory to ReadDirPEM and checks the total
// blocks delivered across the tree and the aggregate error.
type readDirPEMTestCase struct {
	wantErr    error
	fSys       fs.FS
	name       string
	dir        string
	wantBlocks int
}

func newReadDirPEMTestCase(name string, fSys fs.FS, dir string,
	wantErr error, wantBlocks int) readDirPEMTestCase {
	return readDirPEMTestCase{
		name:       name,
		fSys:       fSys,
		dir:        dir,
		wantErr:    wantErr,
		wantBlocks: wantBlocks,
	}
}

func (tc readDirPEMTestCase) Name() string { return tc.name }

func (tc readDirPEMTestCase) Test(t *testing.T) {
	t.Helper()

	var got int
	err := x509utils.ReadDirPEM(tc.fSys, tc.dir, newBlockCounter(t, &got))
	checkReadPEMResult(t, tc.wantErr, err, tc.wantBlocks, got)
}

func readDirPEMTestCases() []readDirPEMTestCase {
	return core.S(
		// descends into sub/ to reach the two blocks there.
		newReadDirPEMTestCase("recursive", pemGoodFS(), ".", nil, 3),
		// the non-PEM file fails, yet the valid blocks are still delivered.
		newReadDirPEMTestCase("aggregates errors", pemMixedFS(), ".",
			core.ErrInvalid, 3),
		// same, but the failing file lives in the sub-directory, so the error
		// bubbles up through the recursive descent.
		newReadDirPEMTestCase("aggregates subdir errors", pemBadSubdirFS(), ".",
			core.ErrInvalid, 3),
		newReadDirPEMTestCase("missing dir", pemGoodFS(), "nope",
			fs.ErrNotExist, 0),
	)
}

func TestReadDirPEM(t *testing.T) {
	core.RunTestCases(t, readDirPEMTestCases())
}

// TestReadDirPEMNilCallback covers the short-circuit: with no callback there is
// nothing to run, so the walk is skipped without error.
func TestReadDirPEMNilCallback(t *testing.T) {
	err := x509utils.ReadDirPEM(pemGoodFS(), ".", nil)
	core.AssertNoError(t, err, "nil cb")
}

// TestReadStringPEMFromFS covers the file and directory paths ReadStringPEM
// resolves through a supplied fs.FS.
func TestReadStringPEMFromFS(t *testing.T) {
	fSys := pemGoodFS()

	t.Run("file", func(t *testing.T) {
		var got int
		err := x509utils.ReadStringPEM("a.pem", newBlockCounter(t, &got),
			x509utils.ReadWithFS(fSys))
		core.AssertNoError(t, err, "read")
		core.AssertEqual(t, 1, got, "blocks")
	})

	t.Run("directory", func(t *testing.T) {
		var got int
		err := x509utils.ReadStringPEM(".", newBlockCounter(t, &got),
			x509utils.ReadWithFS(fSys), x509utils.ReadWithDirs())
		core.AssertNoError(t, err, "read")
		core.AssertEqual(t, 3, got, "blocks")
	})
}

// TestReadStringPEMFromDisk covers the os-backed stat, read and directory paths
// taken when no fs.FS is supplied.
func TestReadStringPEMFromDisk(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "leaf.pem"), testPEM(), 0o600)
	core.AssertMustNoError(t, err, "write")

	t.Run("file", func(t *testing.T) {
		var got int
		err := x509utils.ReadStringPEM(filepath.Join(dir, "leaf.pem"),
			newBlockCounter(t, &got))
		core.AssertNoError(t, err, "read")
		core.AssertEqual(t, 1, got, "blocks")
	})

	t.Run("directory", func(t *testing.T) {
		var got int
		err := x509utils.ReadStringPEM(dir, newBlockCounter(t, &got))
		core.AssertNoError(t, err, "read")
		core.AssertEqual(t, 1, got, "blocks")
	})
}

// TestReadStringPEMReadError covers the read-error path: stat resolves the
// string to a file, but reading its contents fails.
func TestReadStringPEMReadError(t *testing.T) {
	fSys := failReadFS{pemGoodFS()}
	err := x509utils.ReadStringPEM("a.pem", nil, x509utils.ReadWithFS(fSys))
	core.AssertErrorIs(t, err, fs.ErrPermission, "read error")
}

// TestReadStringPEMInvalidStat covers run's normalisation: a stat PathError
// that matches fs.ErrInvalid surfaces as the bare sentinel, not the wrapper.
func TestReadStringPEMInvalidStat(t *testing.T) {
	err := x509utils.ReadStringPEM("x", nil, x509utils.ReadWithFS(invalidStatFS{}))
	core.AssertErrorIs(t, err, fs.ErrInvalid, "invalid")
}

// TestReadStringPEMInvalidNonPath pins the other side: a non-PathError error
// wrapping fs.ErrInvalid is passed through intact, not collapsed.
func TestReadStringPEMInvalidNonPath(t *testing.T) {
	err := x509utils.ReadStringPEM("x", nil, x509utils.ReadWithFS(invalidPlainFS{}))
	core.AssertErrorIs(t, err, fs.ErrInvalid, "invalid")
	core.AssertContains(t, err.Error(), "stat quirk", "preserved")
}

// TestReadWithFSNil rejects a nil fs.FS: the option cannot resolve paths, so it
// reports the misconfiguration as invalid.
func TestReadWithFSNil(t *testing.T) {
	err := x509utils.ReadStringPEM("whatever", nil, x509utils.ReadWithFS(nil))
	core.AssertErrorIs(t, err, core.ErrInvalid, "nil fs")
}

// nilOptionTestCase applies a ReadOption to a nil *readOptions to reach the
// shared nil-receiver guard, which is callable as Option()(nil) even though
// ReadStringPEM only ever applies options to a live one.
type nilOptionTestCase struct {
	opt  x509utils.ReadOption
	name string
}

func newNilOptionTestCase(name string,
	opt x509utils.ReadOption) nilOptionTestCase {
	return nilOptionTestCase{name: name, opt: opt}
}

func (tc nilOptionTestCase) Name() string { return tc.name }

func (tc nilOptionTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertErrorIs(t, tc.opt(nil), core.ErrNilReceiver, "nil receiver")
}

func nilOptionTestCases() []nilOptionTestCase {
	return core.S(
		newNilOptionTestCase("ReadWithFS", x509utils.ReadWithFS(fstest.MapFS{})),
		newNilOptionTestCase("ReadWithoutDirs", x509utils.ReadWithoutDirs()),
		newNilOptionTestCase("ReadWithDirs", x509utils.ReadWithDirs()),
	)
}

func TestReadOptionNilReceiver(t *testing.T) {
	core.RunTestCases(t, nilOptionTestCases())
}
