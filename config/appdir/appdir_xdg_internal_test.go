//go:build !windows

package appdir

import (
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = getTempDirTestCase{}

// getTempDirTestCase tests getTempDir honouring a $TMPDIR
// redirection and defaulting to /tmp.
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

func TestGetTempDir(t *testing.T) {
	testCases := []getTempDirTestCase{
		newGetTempDirTestCase("redirected", "/var/tmp", "/var/tmp"),
		newGetTempDirTestCase("default", "", "/tmp"),
	}

	core.RunTestCases(t, testCases)
}
