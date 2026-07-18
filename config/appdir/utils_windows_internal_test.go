//go:build windows

// cspell:words USERPROFILE
package appdir

import (
	"testing"

	"darvaza.org/core"
)

// setTestHome points the user's home directory at the fixture the
// platform rows expect.
func setTestHome(t *testing.T) {
	t.Helper()
	t.Setenv("USERPROFILE", `C:\Users\test`)
}

func TestGetEnvDir(t *testing.T) {
	testCases := []getEnvDirTestCase{
		newGetEnvDirTestCase("absolute", `C:\custom\dir`,
			`C:\custom\dir`),
		newGetEnvDirTestCase("relative", `custom\dir`, ""),
		// rooted but volume-less values are not absolute on
		// Windows
		newGetEnvDirTestCase("no volume", "/custom/dir", ""),
		newGetEnvDirTestCase("empty", "", ""),
		newGetEnvDirTestCaseUnset("unset"),
	}

	core.RunTestCases(t, testCases)
}

func TestGetEnvHomeDir(t *testing.T) {
	testCases := []getEnvHomeDirTestCase{
		newGetEnvHomeDirTestCase("absolute", `C:\custom\dir`,
			"AppData/Local", `C:\custom\dir`),
		newGetEnvHomeDirTestCase("relative ignored", `custom\dir`,
			"AppData/Local", `C:\Users\test\AppData\Local`),
		newGetEnvHomeDirTestCase("unset", "",
			"AppData/Local", `C:\Users\test\AppData\Local`),
	}

	core.RunTestCases(t, testCases)
}
