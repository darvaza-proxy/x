//go:build !windows

package appdir

import (
	"os"
	"path/filepath"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = getTempDirTestCase{}

// getTempDirTestCase tests getTempDir honouring a $TMPDIR
// redirection, resolving a relative value to an absolute path
// and defaulting to /tmp.
type getTempDirTestCase struct {
	tmpDir string
	name   string
	want   string
}

func (tc getTempDirTestCase) Name() string {
	return tc.name
}

func (tc getTempDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("TMPDIR", tc.tmpDir)

	got, err := getTempDir()
	core.AssertNoError(t, err, "temp dir")
	core.AssertEqual(t, tc.want, got, "dir")
}

func newGetTempDirTestCase(name, tmpDir,
	want string) getTempDirTestCase {
	return getTempDirTestCase{
		tmpDir: tmpDir,
		name:   name,
		want:   want,
	}
}

// getTempDirTestCases builds the rows, resolving the relative
// row's expected value against cwd exactly as getTempDir does.
func getTempDirTestCases(cwd string) []getTempDirTestCase {
	return []getTempDirTestCase{
		newGetTempDirTestCase("redirected", "/var/tmp", "/var/tmp"),
		newGetTempDirTestCase("relative", "rel/tmp",
			filepath.Join(cwd, "rel/tmp")),
		newGetTempDirTestCase("default", "", "/tmp"),
	}
}

func TestGetTempDir(t *testing.T) {
	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	core.RunTestCases(t, getTempDirTestCases(cwd))
}
