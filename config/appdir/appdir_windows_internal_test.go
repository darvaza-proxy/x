//go:build windows

// cspell:words PROGRAMDATA
package appdir

import (
	"os"
	"path/filepath"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = getSysRootTestCase{}
var _ core.TestCase = getTempDirTestCase{}

// getSysRootTestCase tests Prefix.getSysRoot reporting the root to
// compose system directories under, and false under PrefixUser and
// the zero value, which have none.
type getSysRootTestCase struct {
	p      Prefix
	name   string
	want   string
	wantOK bool
}

func (tc getSysRootTestCase) Name() string {
	return tc.name
}

func (tc getSysRootTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("PROGRAMDATA", `C:\ProgramData`)

	got, ok := tc.p.getSysRoot()
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	core.AssertEqual(t, tc.want, got, "root")
}

func newGetSysRootTestCase(name string, p Prefix, want string,
	wantOK bool) getSysRootTestCase {
	return getSysRootTestCase{
		p:      p,
		name:   name,
		want:   want,
		wantOK: wantOK,
	}
}

func TestGetSysRoot(t *testing.T) {
	testCases := []getSysRootTestCase{
		newGetSysRootTestCase("system", PrefixSystem,
			`C:\ProgramData`, true),
		newGetSysRootTestCase("user", PrefixUser, "", false),
		newGetSysRootTestCase("zero value", "", "", false),
		newGetSysRootTestCase("custom", `C:\srv`, `C:\srv`, true),
	}

	core.RunTestCases(t, testCases)
}

// getTempDirTestCase tests getTempDir honouring a %TMP% or %TEMP%
// redirection, resolving a relative value to an absolute path.
// %TMP% takes precedence; the empty %TMP%/%TEMP% fallback is
// covered externally by TestUserRuntimeDirFallback.
type getTempDirTestCase struct {
	tmp  string
	temp string
	name string
	want string
}

func (tc getTempDirTestCase) Name() string {
	return tc.name
}

func (tc getTempDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("TMP", tc.tmp)
	t.Setenv("TEMP", tc.temp)

	got, err := getTempDir()
	core.AssertNoError(t, err, "temp dir")
	core.AssertEqual(t, tc.want, got, "dir")
}

func newGetTempDirTestCase(name, tmp, temp,
	want string) getTempDirTestCase {
	return getTempDirTestCase{
		tmp:  tmp,
		temp: temp,
		name: name,
		want: want,
	}
}

// getTempDirTestCases builds the rows, resolving the relative
// row's expected value against cwd exactly as getTempDir does.
func getTempDirTestCases(cwd string) []getTempDirTestCase {
	return []getTempDirTestCase{
		newGetTempDirTestCase("tmp absolute", `C:\custom\tmp`, "",
			`C:\custom\tmp`),
		newGetTempDirTestCase("tmp relative", `rel\tmp`, "",
			filepath.Join(cwd, `rel\tmp`)),
		newGetTempDirTestCase("temp fallback", "", `C:\custom\temp`,
			`C:\custom\temp`),
	}
}

func TestGetTempDir(t *testing.T) {
	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	core.RunTestCases(t, getTempDirTestCases(cwd))
}
