package appdir_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = prefixValidateTestCase{}
var _ core.TestCase = prefixUserModeTestCase{}

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
