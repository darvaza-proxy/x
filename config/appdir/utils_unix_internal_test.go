//go:build !windows

package appdir

import (
	"testing"

	"darvaza.org/core"
)

// setTestHome points the user's home directory at the fixture the
// platform rows expect.
func setTestHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "/home/test")
}

func TestGetEnvDir(t *testing.T) {
	testCases := []getEnvDirTestCase{
		newGetEnvDirTestCase("absolute", "/custom/dir",
			"/custom/dir"),
		newGetEnvDirTestCase("relative", "relative/dir", ""),
		newGetEnvDirTestCase("empty", "", ""),
		newGetEnvDirTestCaseUnset("unset"),
	}

	core.RunTestCases(t, testCases)
}

func TestGetEnvHomeDir(t *testing.T) {
	testCases := []getEnvHomeDirTestCase{
		newGetEnvHomeDirTestCase("absolute", "/custom/dir",
			".cache", "/custom/dir"),
		newGetEnvHomeDirTestCase("relative ignored", "relative/dir",
			".cache", "/home/test/.cache"),
		newGetEnvHomeDirTestCase("empty", "",
			".cache", "/home/test/.cache"),
		newGetEnvHomeDirTestCase("multipart fallback", "",
			".local/share", "/home/test/.local/share"),
	}

	core.RunTestCases(t, testCases)
}
