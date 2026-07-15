package appdir_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = prefixValidateTestCase{}
var _ core.TestCase = prefixUserModeTestCase{}
var _ core.TestCase = newPrefixTestCase{}

// prefixValidateTestCase tests [appdir.Prefix.Validate] pinning the
// usability contract — a well-known prefix or an absolute path to
// an existing directory — and the identity of each rejection.
type prefixValidateTestCase struct {
	wantErrIs error
	p         appdir.Prefix
	name      string
}

func (tc prefixValidateTestCase) Name() string {
	return tc.name
}

func (tc prefixValidateTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.p.Validate()
	if tc.wantErrIs != nil {
		core.AssertErrorIs(t, err, tc.wantErrIs, "validate")
		return
	}

	core.AssertNoError(t, err, "validate")
}

// newPrefixValidateTestCase declares a row expected to pass
// validation.
func newPrefixValidateTestCase(name string,
	p appdir.Prefix) prefixValidateTestCase {
	return prefixValidateTestCase{
		p:    p,
		name: name,
	}
}

// newPrefixValidateTestCaseErr declares a row expected to fail
// validation.
func newPrefixValidateTestCaseErr(name string, p appdir.Prefix,
	wantErrIs error) prefixValidateTestCase {
	return prefixValidateTestCase{
		wantErrIs: wantErrIs,
		p:         p,
		name:      name,
	}
}

func prefixValidateTestCases(tmp,
	file string) []prefixValidateTestCase {
	return core.S(
		newPrefixValidateTestCase("user mode", appdir.PrefixUser),
		newPrefixValidateTestCase("existing dir",
			appdir.Prefix(tmp)),
		newPrefixValidateTestCaseErr("zero value", "",
			fs.ErrInvalid),
		newPrefixValidateTestCaseErr("relative", "srv/pods",
			fs.ErrInvalid),
		newPrefixValidateTestCaseErr("missing path",
			appdir.Prefix(filepath.Join(tmp, "missing")),
			fs.ErrNotExist),
		newPrefixValidateTestCaseErr("regular file",
			appdir.Prefix(file), syscall.ENOTDIR),
	)
}

func TestPrefixValidate(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file")
	err := os.WriteFile(file, []byte("x"), 0o600)
	core.AssertMustNoError(t, err, "write file")

	core.RunTestCases(t, prefixValidateTestCases(tmp, file))
}

// prefixUserModeTestCase pins the user-mode invariant: under
// [appdir.PrefixUser] every FooDir method returns the same as
// its UserFooDir counterpart, on every platform.
type prefixUserModeTestCase struct {
	name string
	kind Kind
}

func (tc prefixUserModeTestCase) Name() string {
	return tc.name
}

func (tc prefixUserModeTestCase) Test(t *testing.T) {
	t.Helper()

	want, err := callUserDirFunc(t, tc.kind, "app")
	core.AssertMustNoError(t, err, "user %s dir", tc.kind)

	got, err := callPrefixDirFunc(t, appdir.PrefixUser, tc.kind, "app")
	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, want, got, "dir")
}

func newPrefixUserModeTestCase(name string,
	kind Kind) prefixUserModeTestCase {
	return prefixUserModeTestCase{
		kind: kind,
		name: name,
	}
}

func TestPrefixUserMode(t *testing.T) {
	testCases := core.S(
		newPrefixUserModeTestCase("cache", KindCache),
		newPrefixUserModeTestCase("config", KindConfig),
		newPrefixUserModeTestCase("data", KindData),
		newPrefixUserModeTestCase("runtime", KindRuntime),
	)

	core.RunTestCases(t, testCases)
}

// newPrefixTestCase tests [appdir.NewPrefix] validation and
// path resolution.
type newPrefixTestCase struct {
	wantErrIs error
	dir       string
	name      string
	want      appdir.Prefix
}

func (tc newPrefixTestCase) Name() string {
	return tc.name
}

func (tc newPrefixTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := appdir.NewPrefix(tc.dir)
	if tc.wantErrIs != nil {
		core.AssertErrorIs(t, err, tc.wantErrIs, "new prefix")
		return
	}

	core.AssertMustNoError(t, err, "new prefix")
	core.AssertEqual(t, tc.want, got, "prefix")
}

// newNewPrefixTestCase declares a row expected to succeed, with
// want holding the resulting Prefix value.
func newNewPrefixTestCase(name, dir string,
	want appdir.Prefix) newPrefixTestCase {
	return newPrefixTestCase{
		dir:  dir,
		name: name,
		want: want,
	}
}

// newNewPrefixTestCaseErr declares a row expected to fail.
func newNewPrefixTestCaseErr(name, dir string,
	wantErrIs error) newPrefixTestCase {
	return newPrefixTestCase{
		wantErrIs: wantErrIs,
		dir:       dir,
		name:      name,
	}
}

func newPrefixTestCases(tmp, file, cwd string) []newPrefixTestCase {
	cases := core.S(
		newNewPrefixTestCase("user mode",
			string(appdir.PrefixUser), appdir.PrefixUser),
		newNewPrefixTestCase("system prefix",
			string(appdir.PrefixSystem), appdir.PrefixSystem),
		newNewPrefixTestCase("existing dir", tmp,
			appdir.Prefix(tmp)),
		newNewPrefixTestCase("relative path", ".",
			appdir.Prefix(cwd)),
		newNewPrefixTestCaseErr("empty", "", fs.ErrInvalid),
		newNewPrefixTestCaseErr("missing path",
			filepath.Join(tmp, "missing"), fs.ErrNotExist),
		newNewPrefixTestCaseErr("regular file", file,
			syscall.ENOTDIR),
	)
	return append(cases, osNewPrefixTestCases()...)
}

func TestNewPrefix(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file")
	err := os.WriteFile(file, []byte("x"), 0o600)
	core.AssertMustNoError(t, err, "write file")

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	core.RunTestCases(t, newPrefixTestCases(tmp, file, cwd))
}

// TestSysPrefix pins the getter reflecting the current default
// Prefix.
func TestSysPrefix(t *testing.T) {
	core.AssertEqual(t, appdir.PrefixUser, appdir.SysPrefix(),
		"default")

	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixSystem))
	core.AssertEqual(t, appdir.PrefixSystem, appdir.SysPrefix(),
		"stubbed")
}

// TestSysPrefixConcurrent exercises the atomic default under
// concurrent writers and readers. Run with -race it guards
// against a data race on the package-level prefix.
func TestSysPrefixConcurrent(t *testing.T) {
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))

	dirs := core.S(appdir.PrefixUser, appdir.PrefixSystem)

	var wg sync.WaitGroup
	for i := range 16 {
		dir := dirs[i%len(dirs)]
		wg.Go(func() {
			_ = appdir.SetSysPrefix(dir)
		})
		wg.Go(func() {
			_, _ = appdir.SysCacheDir("app")
		})
	}
	wg.Wait()
}
