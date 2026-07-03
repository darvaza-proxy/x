package appdir_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = joinTestCase{}
var _ core.TestCase = userDirTestCase{}
var _ core.TestCase = userDirErrTestCase{}
var _ core.TestCase = sysDirTestCase{}
var _ core.TestCase = sysUserModeTestCase{}
var _ core.TestCase = setSysPrefixTestCase{}

// directory categories used by the Foo{Cache,Config,Data,Runtime}Dir
// test case dispatchers.
const (
	kindCache   = "cache"
	kindConfig  = "config"
	kindData    = "data"
	kindRuntime = "runtime"
)

// xdgEnvKey returns the XDG basedir environment variable overriding
// the given directory category.
func xdgEnvKey(kind string) string {
	switch kind {
	case kindCache:
		return "XDG_CACHE_HOME"
	case kindConfig:
		return "XDG_CONFIG_HOME"
	case kindData:
		return "XDG_DATA_HOME"
	default:
		return "XDG_RUNTIME_DIR"
	}
}

// userDirFn returns the UserFooDir function for the given category.
func userDirFn(t *testing.T, kind string) func(...string) (string, error) {
	t.Helper()

	switch kind {
	case kindCache:
		return appdir.UserCacheDir
	case kindConfig:
		return appdir.UserConfigDir
	case kindData:
		return appdir.UserDataDir
	case kindRuntime:
		return appdir.UserRuntimeDir
	default:
		t.Fatalf("unknown kind %q", kind)
		return nil
	}
}

// sysDirFn returns the SysFooDir function for the given category.
func sysDirFn(t *testing.T, kind string) func(...string) (string, error) {
	t.Helper()

	switch kind {
	case kindCache:
		return appdir.SysCacheDir
	case kindConfig:
		return appdir.SysConfigDir
	case kindData:
		return appdir.SysDataDir
	case kindRuntime:
		return appdir.SysRuntimeDir
	default:
		t.Fatalf("unknown kind %q", kind)
		return nil
	}
}

// joinTestCase tests [appdir.Join] path composition.
type joinTestCase struct {
	base string
	name string
	want string
	sub  []string
}

func (tc joinTestCase) Name() string {
	return tc.name
}

func (tc joinTestCase) Test(t *testing.T) {
	t.Helper()

	got := appdir.Join(tc.base, tc.sub...)
	core.AssertEqual(t, tc.want, got, "join")
}

func newJoinTestCase(name, base string, sub []string,
	want string) joinTestCase {
	return joinTestCase{
		base: base,
		name: name,
		want: want,
		sub:  sub,
	}
}

func TestJoin(t *testing.T) {
	testCases := []joinTestCase{
		newJoinTestCase("base only", "/base", nil, "/base"),
		newJoinTestCase("single sub", "/base", core.S("app"),
			"/base/app"),
		newJoinTestCase("slash sub", "/base", core.S("app/conf.d"),
			"/base/app/conf.d"),
		newJoinTestCase("multiple sub", "/base", core.S("app", "x/y"),
			"/base/app/x/y"),
		newJoinTestCase("no base", "", core.S("app"), "app"),
		newJoinTestCase("trailing slash", "/base", core.S("app/"),
			"/base/app"),
		newJoinTestCase("empty sub", "/base", core.S(""), "/base"),
	}

	core.RunTestCases(t, testCases)
}

// userDirTestCase tests the UserFooDir functions honouring their
// XDG environment variable override.
type userDirTestCase struct {
	kind     string
	envValue string
	name     string
	want     string
	sub      []string
}

func (tc userDirTestCase) Name() string {
	return tc.name
}

func (tc userDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv(xdgEnvKey(tc.kind), tc.envValue)

	got, err := userDirFn(t, tc.kind)(tc.sub...)
	core.AssertNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

func newUserDirTestCase(name, kind, envValue string, sub []string,
	want string) userDirTestCase {
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
	return newUserDirTestCase(name, kindCache, envValue, sub, want)
}

// newUserConfigDirTestCase declares a row for [appdir.UserConfigDir].
func newUserConfigDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, kindConfig, envValue, sub, want)
}

// newUserDataDirTestCase declares a row for [appdir.UserDataDir].
func newUserDataDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, kindData, envValue, sub, want)
}

// newUserRuntimeDirTestCase declares a row for [appdir.UserRuntimeDir].
func newUserRuntimeDirTestCase(name, envValue string, sub []string,
	want string) userDirTestCase {
	return newUserDirTestCase(name, kindRuntime, envValue, sub, want)
}

func TestUserDir(t *testing.T) {
	testCases := []userDirTestCase{
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
	}

	core.RunTestCases(t, testCases)
}

// userDirErrTestCase tests the UserFooDir functions failing when
// both their XDG environment variable and HOME are unset.
type userDirErrTestCase struct {
	kind string
	name string
}

func (tc userDirErrTestCase) Name() string {
	return tc.name
}

func (tc userDirErrTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv(xdgEnvKey(tc.kind), "")
	t.Setenv("HOME", "")

	_, err := userDirFn(t, tc.kind)("app")
	core.AssertError(t, err, "%s dir", tc.kind)
}

func newUserDirErrTestCase(name, kind string) userDirErrTestCase {
	return userDirErrTestCase{
		kind: kind,
		name: name,
	}
}

// newUserCacheDirErrTestCase declares a row where
// [appdir.UserCacheDir] is expected to fail.
func newUserCacheDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, kindCache)
}

// newUserConfigDirErrTestCase declares a row where
// [appdir.UserConfigDir] is expected to fail.
func newUserConfigDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, kindConfig)
}

// newUserDataDirErrTestCase declares a row where
// [appdir.UserDataDir] is expected to fail.
func newUserDataDirErrTestCase(name string) userDirErrTestCase {
	return newUserDirErrTestCase(name, kindData)
}

func TestUserDirErr(t *testing.T) {
	testCases := []userDirErrTestCase{
		newUserCacheDirErrTestCase("cache"),
		newUserConfigDirErrTestCase("config"),
		newUserDataDirErrTestCase("data"),
	}

	core.RunTestCases(t, testCases)
}

func TestUserRuntimeDirFallback(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")

	got, err := appdir.UserRuntimeDir()
	core.AssertNoError(t, err, "user runtime dir")

	ok := strings.HasPrefix(got, "/run/user/") ||
		strings.HasPrefix(got, "/tmp/runtime-")
	core.AssertTrue(t, ok, "fallback %q", got)
}

func TestUserDataDirFallback(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/home/test")

	got, err := appdir.UserDataDir("app")
	core.AssertNoError(t, err, "user data dir")
	core.AssertEqual(t, "/home/test/.local/share/app", got, "dir")
}

// sysDirTestCase tests the SysFooDir functions under a stubbed
// system prefix.
type sysDirTestCase struct {
	kind    string
	prefix  string
	name    string
	want    string
	sub     []string
	wantErr bool
}

func (tc sysDirTestCase) Name() string {
	return tc.name
}

func (tc sysDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Cleanup(appdir.StubSysPrefix(tc.prefix))

	got, err := sysDirFn(t, tc.kind)(tc.sub...)
	if tc.wantErr {
		core.AssertError(t, err, "%s dir", tc.kind)
		return
	}

	core.AssertNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

// newSysDirTestCase declares a row expected to succeed.
func newSysDirTestCase(name, kind, prefix string, sub []string,
	want string) sysDirTestCase {
	return sysDirTestCase{
		kind:   kind,
		prefix: prefix,
		name:   name,
		want:   want,
		sub:    sub,
	}
}

// newSysCacheDirTestCase declares a row for [appdir.SysCacheDir].
func newSysCacheDirTestCase(name, prefix string, sub []string,
	want string) sysDirTestCase {
	return newSysDirTestCase(name, kindCache, prefix, sub, want)
}

// newSysConfigDirTestCase declares a row for [appdir.SysConfigDir].
func newSysConfigDirTestCase(name, prefix string, sub []string,
	want string) sysDirTestCase {
	return newSysDirTestCase(name, kindConfig, prefix, sub, want)
}

// newSysDataDirTestCase declares a row for [appdir.SysDataDir].
func newSysDataDirTestCase(name, prefix string, sub []string,
	want string) sysDirTestCase {
	return newSysDirTestCase(name, kindData, prefix, sub, want)
}

// newSysRuntimeDirTestCase declares a row for [appdir.SysRuntimeDir].
func newSysRuntimeDirTestCase(name, prefix string, sub []string,
	want string) sysDirTestCase {
	return newSysDirTestCase(name, kindRuntime, prefix, sub, want)
}

// newSysConfigDirTestCaseErr declares a row where
// [appdir.SysConfigDir] is expected to fail.
func newSysConfigDirTestCaseErr(name, prefix string,
	sub []string) sysDirTestCase {
	return sysDirTestCase{
		kind:    kindConfig,
		prefix:  prefix,
		name:    name,
		sub:     sub,
		wantErr: true,
	}
}

func sysDirTestCases() []sysDirTestCase {
	return []sysDirTestCase{
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
	}
}

func TestSysDir(t *testing.T) {
	core.RunTestCases(t, sysDirTestCases())
}

// sysUserModeTestCase tests the SysFooDir functions falling through
// to their UserFooDir counterparts under [appdir.PrefixUser].
type sysUserModeTestCase struct {
	kind     string
	envValue string
	name     string
	want     string
	sub      []string
}

func (tc sysUserModeTestCase) Name() string {
	return tc.name
}

func (tc sysUserModeTestCase) Test(t *testing.T) {
	t.Helper()
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixUser))
	t.Setenv(xdgEnvKey(tc.kind), tc.envValue)

	got, err := sysDirFn(t, tc.kind)(tc.sub...)
	core.AssertNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, tc.want, got, "dir")
}

func newSysUserModeTestCase(name, kind, envValue string, sub []string,
	want string) sysUserModeTestCase {
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
	return newSysUserModeTestCase(name, kindCache, envValue, sub, want)
}

// newSysConfigDirUserModeTestCase declares a row for
// [appdir.SysConfigDir] in user mode.
func newSysConfigDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, kindConfig, envValue, sub, want)
}

// newSysDataDirUserModeTestCase declares a row for
// [appdir.SysDataDir] in user mode.
func newSysDataDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, kindData, envValue, sub, want)
}

// newSysRuntimeDirUserModeTestCase declares a row for
// [appdir.SysRuntimeDir] in user mode.
func newSysRuntimeDirUserModeTestCase(name, envValue string,
	sub []string, want string) sysUserModeTestCase {
	return newSysUserModeTestCase(name, kindRuntime, envValue, sub, want)
}

func TestSysDirUserMode(t *testing.T) {
	testCases := []sysUserModeTestCase{
		newSysCacheDirUserModeTestCase("cache", "/custom/cache",
			core.S("app"), "/custom/cache/app"),
		newSysConfigDirUserModeTestCase("config", "/custom/config",
			core.S("app"), "/custom/config/app"),
		newSysDataDirUserModeTestCase("data", "/custom/share",
			core.S("app"), "/custom/share/app"),
		newSysRuntimeDirUserModeTestCase("runtime", "/custom/run",
			core.S("app"), "/custom/run/app"),
	}

	core.RunTestCases(t, testCases)
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

	core.AssertNoError(t, err, "set prefix")

	got, err := appdir.SysConfigDir("app")
	core.AssertNoError(t, err, "sys config dir")
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
	return []setSysPrefixTestCase{
		newSetSysPrefixTestCase("existing dir", tmp,
			filepath.Join(tmp, "etc", "app")),
		newSetSysPrefixTestCase("relative path", ".",
			filepath.Join(cwd, "etc", "app")),
		newSetSysPrefixTestCaseErr("missing path",
			filepath.Join(tmp, "missing"), fs.ErrNotExist),
		newSetSysPrefixTestCaseErr("regular file", file,
			fs.ErrInvalid),
	}
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
	t.Cleanup(appdir.StubSysPrefix(appdir.PrefixSystem))

	cwd, err := os.Getwd()
	core.AssertMustNoError(t, err, "getwd")

	dirs := appdir.AllConfigDir("app")
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
