// cspell:words LOCALAPPDATA PROGRAMDATA USERPROFILE
package appdir_test

import (
	"path/filepath"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = userDirTestCase{}
var _ core.TestCase = userDirErrTestCase{}
var _ core.TestCase = userDirFallbackTestCase{}
var _ core.TestCase = sysDirTestCase{}

// userDirTestCase tests the UserFooDir functions honouring the
// environment variable providing their base directory.
type userDirTestCase struct {
	envValue string
	name     string
	want     string
	sub      []string
	kind     Kind
}

func (tc userDirTestCase) Name() string {
	return tc.name
}

func (tc userDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("USERNAME", "test")
	setWinEnv(t, tc.kind, tc.envValue)

	got, err := callUserDirFunc(t, tc.kind, tc.sub...)
	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

func newUserDirTestCase(name string, kind Kind, envValue string,
	sub []string, want string) userDirTestCase {
	return userDirTestCase{
		kind:     kind,
		envValue: envValue,
		name:     name,
		want:     want,
		sub:      sub,
	}
}

func TestUserDir(t *testing.T) {
	testCases := core.S(
		newUserDirTestCase("cache", KindCache,
			`C:\custom\local`, core.S("app"),
			`C:\custom\local\app`),
		newUserDirTestCase("config", KindConfig,
			`C:\custom\roaming`, core.S("app"),
			`C:\custom\roaming\app`),
		// data shares the roaming directory with config
		newUserDirTestCase("data", KindData,
			`C:\custom\roaming`, core.S("app"),
			`C:\custom\roaming\app`),
		// run-time data gets a user-distinguishing leaf under
		// the temporary directory
		newUserDirTestCase("runtime", KindRuntime,
			`C:\custom\tmp`, core.S("app"),
			`C:\custom\tmp\runtime-test\app`),
		newUserDirTestCase("runtime multipart", KindRuntime,
			`C:\custom\tmp`, core.S("app/state"),
			`C:\custom\tmp\runtime-test\app\state`),
	)

	core.RunTestCases(t, testCases)
}

// userDirErrTestCase tests the UserFooDir functions failing when
// the profile environment is undefined.
type userDirErrTestCase struct {
	name string
	kind Kind
}

func (tc userDirErrTestCase) Name() string {
	return tc.name
}

func (tc userDirErrTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("APPDATA", "")
	t.Setenv("TMP", "")
	t.Setenv("TEMP", "")
	t.Setenv("USERPROFILE", "")

	_, err := callUserDirFunc(t, tc.kind, "app")
	core.AssertError(t, err, "%s dir", tc.kind)
}

func newUserDirErrTestCase(name string, kind Kind) userDirErrTestCase {
	return userDirErrTestCase{
		kind: kind,
		name: name,
	}
}

func TestUserDirErr(t *testing.T) {
	testCases := core.S(
		newUserDirErrTestCase("cache", KindCache),
		newUserDirErrTestCase("config", KindConfig),
		newUserDirErrTestCase("data", KindData),
		newUserDirErrTestCase("runtime", KindRuntime),
	)

	core.RunTestCases(t, testCases)
}

// TestUserRuntimeDirFallback pins [appdir.UserRuntimeDir]
// composing the default %LocalAppData%\Temp when neither %TMP%
// nor %TEMP% is defined.
func TestUserRuntimeDirFallback(t *testing.T) {
	t.Setenv("TMP", "")
	t.Setenv("TEMP", "")
	t.Setenv("USERNAME", "test")
	t.Setenv("LOCALAPPDATA", `C:\custom\local`)

	got, err := appdir.UserRuntimeDir("app")
	core.AssertNoError(t, err, "runtime dir")
	core.AssertEqual(t, `C:\custom\local\Temp\runtime-test\app`,
		got, "dir")
}

// userDirFallbackTestCase tests the UserFooDir functions composing
// their default under %USERPROFILE% when the profile environment
// variables are undefined.
type userDirFallbackTestCase struct {
	name string
	want string
	kind Kind
}

func (tc userDirFallbackTestCase) Name() string {
	return tc.name
}

func (tc userDirFallbackTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("APPDATA", "")
	t.Setenv("TMP", "")
	t.Setenv("TEMP", "")
	t.Setenv("USERNAME", "test")
	t.Setenv("USERPROFILE", `C:\Users\test`)

	got, err := callUserDirFunc(t, tc.kind, "app")
	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

func newUserDirFallbackTestCase(name string, kind Kind,
	want string) userDirFallbackTestCase {
	return userDirFallbackTestCase{
		kind: kind,
		name: name,
		want: want,
	}
}

func TestUserDirFallback(t *testing.T) {
	testCases := core.S(
		newUserDirFallbackTestCase("cache", KindCache,
			`C:\Users\test\AppData\Local\app`),
		newUserDirFallbackTestCase("config", KindConfig,
			`C:\Users\test\AppData\Roaming\app`),
		// data shares the roaming directory with config
		newUserDirFallbackTestCase("data", KindData,
			`C:\Users\test\AppData\Roaming\app`),
		newUserDirFallbackTestCase("runtime", KindRuntime,
			`C:\Users\test\AppData\Local\Temp\runtime-test\app`),
	)

	core.RunTestCases(t, testCases)
}

// sysDirTestCase tests the [appdir.Prefix] FooDir methods under a
// stubbed %ProgramData%.
type sysDirTestCase struct {
	prefix  appdir.Prefix
	root    string
	name    string
	want    string
	sub     []string
	kind    Kind
	wantErr bool
}

func (tc sysDirTestCase) Name() string {
	return tc.name
}

func (tc sysDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("PROGRAMDATA", tc.root)

	got, err := callPrefixDirFunc(t, tc.prefix, tc.kind, tc.sub...)
	if tc.wantErr {
		core.AssertError(t, err, "%s dir", tc.kind)
		return
	}

	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

// programData is the stubbed %ProgramData% root used by the
// system-mode rows.
const programData = `C:\ProgramData`

// newSysDirTestCase declares a row expected to succeed.
func newSysDirTestCase(name string, kind Kind, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return sysDirTestCase{
		kind:   kind,
		prefix: prefix,
		root:   programData,
		name:   name,
		want:   want,
		sub:    sub,
	}
}

// newSysDirTestCaseErr declares a row expected to fail.
func newSysDirTestCaseErr(name string, kind Kind, prefix appdir.Prefix,
	sub []string) sysDirTestCase {
	return sysDirTestCase{
		kind:    kind,
		prefix:  prefix,
		root:    programData,
		name:    name,
		sub:     sub,
		wantErr: true,
	}
}

// newSysDirTestCaseNoRoot declares a row expected to fail because
// %ProgramData% is not defined.
func newSysDirTestCaseNoRoot(name string, kind Kind,
	prefix appdir.Prefix, sub []string) sysDirTestCase {
	return sysDirTestCase{
		kind:    kind,
		prefix:  prefix,
		name:    name,
		sub:     sub,
		wantErr: true,
	}
}

func sysDirTestCases(tmp string) []sysDirTestCase {
	return core.S(
		// PrefixSystem
		newSysDirTestCase("cache system", KindCache,
			appdir.PrefixSystem, core.S("app"),
			`C:\ProgramData\app\cache`),
		newSysDirTestCase("config system", KindConfig,
			appdir.PrefixSystem, core.S("app"),
			`C:\ProgramData\app\config`),
		newSysDirTestCase("data system", KindData,
			appdir.PrefixSystem, core.S("app"),
			`C:\ProgramData\app\data`),
		newSysDirTestCase("runtime system", KindRuntime,
			appdir.PrefixSystem, core.S("app"),
			`C:\ProgramData\app\run`),
		// custom prefix
		newSysDirTestCase("config custom", KindConfig,
			appdir.Prefix(tmp), core.S("app"),
			filepath.Join(tmp, "app", "config")),
		// sub-paths append after the category
		newSysDirTestCase("data multiple sub", KindData,
			appdir.PrefixSystem, core.S("app", "models"),
			`C:\ProgramData\app\data\models`),
		newSysDirTestCase("config slash sub", KindConfig,
			appdir.PrefixSystem, core.S("app/conf.d"),
			`C:\ProgramData\app\config\conf.d`),
		// the application name is always required
		newSysDirTestCaseErr("config without app name",
			KindConfig, appdir.PrefixSystem, nil),
		// malformed prefixes carry no root and are rejected
		newSysDirTestCaseErr("config zero value", KindConfig,
			"", core.S("app")),
		newSysDirTestCaseErr("data zero value", KindData,
			"", core.S("app")),
		newSysDirTestCaseErr("cache relative", KindCache,
			"srv/pods", core.S("app")),
		// %ProgramData% must be defined for the FHS prefixes
		newSysDirTestCaseNoRoot("config undefined root",
			KindConfig, appdir.PrefixSystem, core.S("app")),
	)
}

func TestSysDir(t *testing.T) {
	core.RunTestCases(t, sysDirTestCases(t.TempDir()))
}
