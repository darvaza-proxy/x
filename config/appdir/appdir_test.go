package appdir_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = userDirTestCase{}
var _ core.TestCase = userDirErrTestCase{}
var _ core.TestCase = sysDirTestCase{}
var _ core.TestCase = sysUserModeTestCase{}
var _ core.TestCase = newPrefixTestCase{}
var _ core.TestCase = setSysPrefixTestCase{}

// userDirTestCase tests the UserFooDir functions honouring their
// XDG environment variable override.
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
	setXDGEnv(t, tc.kind, tc.envValue)

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

// newUserCacheDirTestCase declares a row for [appdir.UserCacheDir].
func newUserCacheDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, KindCache, envValue, sub, want)
}

// newUserConfigDirTestCase declares a row for [appdir.UserConfigDir].
func newUserConfigDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, KindConfig, envValue, sub, want)
}

// newUserDataDirTestCase declares a row for [appdir.UserDataDir].
func newUserDataDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, KindData, envValue, sub, want)
}

// newUserRuntimeDirTestCase declares a row for [appdir.UserRuntimeDir].
func newUserRuntimeDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, KindRuntime, envValue, sub, want)
}

func TestUserDir(t *testing.T) {
	testCases := core.S(
		newUserCacheDirTestCase("cache", "/custom/cache",
			core.S("app"), "/custom/cache/app"),
		newUserConfigDirTestCase("config", "/custom/config",
			core.S("app"), "/custom/config/app"),
		newUserDataDirTestCase("data", "/custom/share",
			core.S("app"), "/custom/share/app"),
		newUserRuntimeDirTestCase("runtime", "/custom/run",
			core.S("app"), "/custom/run/app"),
		newUserRuntimeDirTestCase("runtime multipart",
			"/custom/run", core.S("app/state"),
			"/custom/run/app/state"),
	)

	core.RunTestCases(t, testCases)
}

// userDirErrTestCase tests the UserFooDir functions failing when
// both their XDG environment variable and HOME are unset.
type userDirErrTestCase struct {
	name string
	kind Kind
}

func (tc userDirErrTestCase) Name() string {
	return tc.name
}

func (tc userDirErrTestCase) Test(t *testing.T) {
	t.Helper()
	setXDGEnv(t, tc.kind, "")
	t.Setenv("HOME", "")

	_, err := callUserDirFunc(t, tc.kind, "app")
	core.AssertError(t, err, "%s dir", tc.kind)
}

func newUserDirErrTestCase(name string, kind Kind) userDirErrTestCase {
	return userDirErrTestCase{
		kind: kind,
		name: name,
	}
}

// newUserCacheDirErrTestCase declares a row where
// [appdir.UserCacheDir] is expected to fail.
func newUserCacheDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, KindCache)
}

// newUserConfigDirErrTestCase declares a row where
// [appdir.UserConfigDir] is expected to fail.
func newUserConfigDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, KindConfig)
}

// newUserDataDirErrTestCase declares a row where
// [appdir.UserDataDir] is expected to fail.
func newUserDataDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, KindData)
}

func TestUserDirErr(t *testing.T) {
	testCases := core.S(
		newUserCacheDirErrTestCase("cache"),
		newUserConfigDirErrTestCase("config"),
		newUserDataDirErrTestCase("data"),
	)

	core.RunTestCases(t, testCases)
}

func TestUserRuntimeDirFallback(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")

	got, err := appdir.UserRuntimeDir()
	core.AssertMustNoError(t, err, "user runtime dir")

	ok := strings.HasPrefix(got, "/run/user/") ||
		strings.HasPrefix(got, "/tmp/runtime-")
	core.AssertTrue(t, ok, "fallback %q", got)
}

func TestUserDataDirFallback(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/home/test")

	got, err := appdir.UserDataDir("app")
	core.AssertMustNoError(t, err, "user data dir")
	core.AssertEqual(t, "/home/test/.local/share/app", got, "dir")
}

// sysDirTestCase tests the [appdir.Prefix] FooDir methods.
type sysDirTestCase struct {
	prefix  appdir.Prefix
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

	got, err := callPrefixDirFunc(t, tc.prefix, tc.kind, tc.sub...)
	if tc.wantErr {
		core.AssertError(t, err, "%s dir", tc.kind)
		return
	}

	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

// newSysDirTestCase declares a row expected to succeed.
func newSysDirTestCase(name string, kind Kind, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return sysDirTestCase{
		kind:   kind,
		prefix: prefix,
		name:   name,
		want:   want,
		sub:    sub,
	}
}

// newSysCacheDirTestCase declares a row for
// [appdir.Prefix.CacheDir].
func newSysCacheDirTestCase(name string, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return newSysDirTestCase(name, KindCache, prefix, sub, want)
}

// newSysConfigDirTestCase declares a row for
// [appdir.Prefix.ConfigDir].
func newSysConfigDirTestCase(name string, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return newSysDirTestCase(name, KindConfig, prefix, sub, want)
}

// newSysDataDirTestCase declares a row for [appdir.Prefix.DataDir].
func newSysDataDirTestCase(name string, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return newSysDirTestCase(name, KindData, prefix, sub, want)
}

// newSysRuntimeDirTestCase declares a row for
// [appdir.Prefix.RuntimeDir].
func newSysRuntimeDirTestCase(name string, prefix appdir.Prefix,
	sub []string, want string) sysDirTestCase {
	return newSysDirTestCase(name, KindRuntime, prefix, sub, want)
}

// newSysConfigDirTestCaseErr declares a row where
// [appdir.Prefix.ConfigDir] is expected to fail.
func newSysConfigDirTestCaseErr(name string, prefix appdir.Prefix,
	sub []string) sysDirTestCase {
	return sysDirTestCase{
		kind:    KindConfig,
		prefix:  prefix,
		name:    name,
		sub:     sub,
		wantErr: true,
	}
}

func sysDirTestCases() []sysDirTestCase {
	return core.S(
		// PrefixSystem
		newSysCacheDirTestCase("cache system",
			appdir.PrefixSystem, core.S("app"), "/var/cache/app"),
		newSysConfigDirTestCase("config system",
			appdir.PrefixSystem, core.S("app"), "/etc/app"),
		newSysDataDirTestCase("data system",
			appdir.PrefixSystem, core.S("app"), "/var/lib/app"),
		newSysRuntimeDirTestCase("runtime system",
			appdir.PrefixSystem, core.S("app"), "/var/run/app"),
		// PrefixLocal
		newSysConfigDirTestCase("config local",
			appdir.PrefixLocal, core.S("app"),
			"/usr/local/etc/app"),
		newSysDataDirTestCase("data local",
			appdir.PrefixLocal, core.S("app"),
			"/usr/local/var/lib/app"),
		// custom prefix
		newSysConfigDirTestCase("config custom",
			"/srv/pods", core.S("app"), "/srv/pods/etc/app"),
		// PrefixOptional swaps app name and category
		newSysCacheDirTestCase("cache opt",
			appdir.PrefixOptional, core.S("app"), "/opt/app/cache"),
		newSysConfigDirTestCase("config opt",
			appdir.PrefixOptional, core.S("app"), "/opt/app/etc"),
		newSysDataDirTestCase("data opt",
			appdir.PrefixOptional, core.S("app"), "/opt/app/share"),
		newSysRuntimeDirTestCase("runtime opt",
			appdir.PrefixOptional, core.S("app"), "/opt/app/run"),
		newSysDataDirTestCase("data opt multiple sub",
			appdir.PrefixOptional, core.S("app", "models"),
			"/opt/app/share/models"),
		newSysDataDirTestCase("data opt slash sub",
			appdir.PrefixOptional, core.S("app/models"),
			"/opt/app/share/models"),
		newSysConfigDirTestCaseErr("config opt without app name",
			appdir.PrefixOptional, nil),
	)
}

func TestSysDir(t *testing.T) {
	core.RunTestCases(t, sysDirTestCases())
}

// sysUserModeTestCase tests the SysFooDir functions falling through
// to their UserFooDir counterparts under [appdir.PrefixUser].
type sysUserModeTestCase struct {
	envValue string
	name     string
	want     string
	sub      []string
	kind     Kind
}

func (tc sysUserModeTestCase) Name() string {
	return tc.name
}

func (tc sysUserModeTestCase) Test(t *testing.T) {
	t.Helper()
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))
	setXDGEnv(t, tc.kind, tc.envValue)

	got, err := callSysDirFunc(t, tc.kind, tc.sub...)
	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

func newSysUserModeTestCase(name string, kind Kind, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return sysUserModeTestCase{
		kind:     kind,
		envValue: envValue,
		name:     name,
		want:     want,
		sub:      sub,
	}
}

// newSysCacheDirUserModeTestCase declares a row for
// [appdir.SysCacheDir] in user mode.
func newSysCacheDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, KindCache, envValue, sub, want)
}

// newSysConfigDirUserModeTestCase declares a row for
// [appdir.SysConfigDir] in user mode.
func newSysConfigDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, KindConfig, envValue, sub, want)
}

// newSysDataDirUserModeTestCase declares a row for
// [appdir.SysDataDir] in user mode.
func newSysDataDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, KindData, envValue, sub, want)
}

// newSysRuntimeDirUserModeTestCase declares a row for
// [appdir.SysRuntimeDir] in user mode.
func newSysRuntimeDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, KindRuntime, envValue, sub, want)
}

func TestSysDirUserMode(t *testing.T) {
	testCases := core.S(
		newSysCacheDirUserModeTestCase("cache", "/custom/cache",
			core.S("app"), "/custom/cache/app"),
		newSysConfigDirUserModeTestCase("config", "/custom/config",
			core.S("app"), "/custom/config/app"),
		newSysDataDirUserModeTestCase("data", "/custom/share",
			core.S("app"), "/custom/share/app"),
		newSysRuntimeDirUserModeTestCase("runtime", "/custom/run",
			core.S("app"), "/custom/run/app"),
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
	return core.S(
		newNewPrefixTestCase("user mode", "~", appdir.PrefixUser),
		newNewPrefixTestCase("existing dir", tmp,
			appdir.Prefix(tmp)),
		newNewPrefixTestCase("relative path", ".",
			appdir.Prefix(cwd)),
		newNewPrefixTestCaseErr("missing path",
			filepath.Join(tmp, "missing"), fs.ErrNotExist),
		newNewPrefixTestCaseErr("regular file", file,
			syscall.ENOTDIR),
	)
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

// TestNewPrefixAbsError pins [appdir.NewPrefix] propagating the
// filepath.Abs failure resolving a relative argument when the
// working directory no longer exists.
func TestNewPrefixAbsError(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	err := os.Mkdir(gone, 0o750)
	core.AssertMustNoError(t, err, "mkdir")

	t.Chdir(gone)
	err = os.Remove(gone)
	core.AssertMustNoError(t, err, "remove")

	got, err := appdir.NewPrefix(".")
	core.AssertErrorIs(t, err, fs.ErrNotExist, "new prefix")
	core.AssertEqual(t, appdir.Prefix(""), got, "prefix")
}

// setSysPrefixTestCase tests [appdir.SetSysPrefix] validation and its
// effect on the SysFooDir functions.
type setSysPrefixTestCase struct {
	wantErrIs error
	dir       string
	name      string
	want      string
}

func (tc setSysPrefixTestCase) Name() string {
	return tc.name
}

func (tc setSysPrefixTestCase) Test(t *testing.T) {
	t.Helper()
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))

	err := appdir.SetSysPrefix(tc.dir)
	if tc.wantErrIs != nil {
		core.AssertErrorIs(t, err, tc.wantErrIs, "set prefix")
		return
	}

	core.AssertMustNoError(t, err, "set prefix")

	got, err := appdir.SysConfigDir("app")
	core.AssertMustNoError(t, err, "sys config dir")
	core.AssertEqual(t, tc.want, got, "dir")
}

// newSetSysPrefixTestCase declares a row expected to succeed, with
// want holding the resulting SysConfigDir("app") path.
func newSetSysPrefixTestCase(name, dir, want string) setSysPrefixTestCase {
	return setSysPrefixTestCase{
		dir:  dir,
		name: name,
		want: want,
	}
}

// newSetSysPrefixTestCaseErr declares a row expected to fail.
func newSetSysPrefixTestCaseErr(name, dir string,
	wantErrIs error) setSysPrefixTestCase {
	return setSysPrefixTestCase{
		wantErrIs: wantErrIs,
		dir:       dir,
		name:      name,
	}
}

func setSysPrefixTestCases(tmp, file, cwd string) []setSysPrefixTestCase {
	return core.S(
		newSetSysPrefixTestCase("existing dir", tmp,
			filepath.Join(tmp, "etc", "app")),
		newSetSysPrefixTestCase("relative path", ".",
			filepath.Join(cwd, "etc", "app")),
		newSetSysPrefixTestCaseErr("missing path",
			filepath.Join(tmp, "missing"), fs.ErrNotExist),
		newSetSysPrefixTestCaseErr("regular file", file,
			syscall.ENOTDIR),
	)
}

func TestSetSysPrefix(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file")
	err := os.WriteFile(file, []byte("x"), 0o600)
	core.AssertMustNoError(t, err, "write file")

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	core.RunTestCases(t, setSysPrefixTestCases(tmp, file, cwd))
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

// TestSetSysPrefixUserMode pins "~" returning the package to user
// mode after a system-mode Prefix was in effect.
func TestSetSysPrefixUserMode(t *testing.T) {
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixSystem))
	t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")

	err := appdir.SetSysPrefix("~")
	core.AssertMustNoError(t, err, "set prefix")

	got, err := appdir.SysConfigDir("app")
	core.AssertMustNoError(t, err, "sys config dir")
	core.AssertEqual(t, "/home/test/.config/app", got, "dir")
}

func TestAllConfigDir(t *testing.T) {
	t.Run("user mode", runTestAllConfigDirUserMode)
	t.Run("system mode", runTestAllConfigDirSystemMode)
	t.Run("multiple sub", runTestAllConfigDirMultiSub)
}

func runTestAllConfigDirUserMode(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	dirs := appdir.AllConfigDir("app")
	core.AssertSliceEqual(t,
		core.S(cwd, "/home/test/.config/app"), dirs, "dirs")
}

func runTestAllConfigDirSystemMode(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	dirs := appdir.PrefixSystem.AllConfigDir("app")
	core.AssertSliceEqual(t,
		core.S(cwd, "/home/test/.config/app", "/etc/app"),
		dirs, "dirs")
}

// runTestAllConfigDirMultiSub pins the working directory entry
// dropping the application name: ./app/conf becomes ./conf.
func runTestAllConfigDirMultiSub(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	dirs := appdir.AllConfigDir("app", "conf")
	core.AssertSliceEqual(t,
		core.S(filepath.Join(cwd, "conf"),
			"/home/test/.config/app/conf"),
		dirs, "dirs")
}
