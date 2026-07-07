package x509utils_test

import (
	"bytes"
	"encoding/pem"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
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
