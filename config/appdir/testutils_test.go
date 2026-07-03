package appdir_test

import (
	"fmt"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Kind identifies a category of application directory: cache,
// configuration, data, or runtime state.
type Kind uint

var _ fmt.Stringer = Kind(0)

// The recognised directory categories. Counting from iota+1 leaves
// the zero value invalid, so an unset Kind never resolves to a real
// category by accident.
const (
	KindCache Kind = iota + 1
	KindConfig
	KindData
	KindRuntime
)

// String returns the lowercase category name, or "Kind(n)" for an
// unrecognised value.
func (k Kind) String() string {
	switch k {
	case KindCache:
		return "cache"
	case KindConfig:
		return "config"
	case KindData:
		return "data"
	case KindRuntime:
		return "runtime"
	default:
		return fmt.Sprintf("Kind(%d)", uint(k))
	}
}

// getXDGEnvKey returns the XDG basedir environment variable overriding
// the given directory category, or false if the category is not
// recognised.
func getXDGEnvKey(kind Kind) (string, bool) {
	switch kind {
	case KindCache:
		return "XDG_CACHE_HOME", true
	case KindConfig:
		return "XDG_CONFIG_HOME", true
	case KindData:
		return "XDG_DATA_HOME", true
	case KindRuntime:
		return "XDG_RUNTIME_DIR", true
	default:
		return "", false
	}
}

// setXDGEnv sets the XDG basedir environment variable for the given
// category to value, failing the test if the category is not
// recognised.
func setXDGEnv(t *testing.T, kind Kind, value string) {
	t.Helper()

	key, ok := getXDGEnvKey(kind)
	core.AssertMustTrue(t, ok, "%s env key", kind)
	t.Setenv(key, value)
}

// getUserDirFunc returns the UserFooDir function for the given category,
// or nil if the category is not recognised.
func getUserDirFunc(kind Kind) func(...string) (string, error) {
	switch kind {
	case KindCache:
		return appdir.UserCacheDir
	case KindConfig:
		return appdir.UserConfigDir
	case KindData:
		return appdir.UserDataDir
	case KindRuntime:
		return appdir.UserRuntimeDir
	default:
		return nil
	}
}

// getSysDirFunc returns the SysFooDir function for the given category,
// or nil if the category is not recognised.
func getSysDirFunc(kind Kind) func(...string) (string, error) {
	switch kind {
	case KindCache:
		return appdir.SysCacheDir
	case KindConfig:
		return appdir.SysConfigDir
	case KindData:
		return appdir.SysDataDir
	case KindRuntime:
		return appdir.SysRuntimeDir
	default:
		return nil
	}
}

// getPrefixDirFunc returns the [appdir.Prefix] FooDir method for the
// given category, or nil if the category is not recognised.
func getPrefixDirFunc(p appdir.Prefix,
	kind Kind) func(...string) (string, error) {
	switch kind {
	case KindCache:
		return p.CacheDir
	case KindConfig:
		return p.ConfigDir
	case KindData:
		return p.DataDir
	case KindRuntime:
		return p.RuntimeDir
	default:
		return nil
	}
}

// callUserDirFunc resolves the UserFooDir function for the category,
// asserts it was recognised, and invokes it.
func callUserDirFunc(t *testing.T, kind Kind,
	args ...string) (string, error) {
	t.Helper()

	fn := getUserDirFunc(kind)
	core.AssertMustNotNil(t, fn, "user %s dir fn", kind)
	return fn(args...)
}

// callSysDirFunc resolves the SysFooDir function for the category,
// asserts it was recognised, and invokes it.
func callSysDirFunc(t *testing.T, kind Kind,
	args ...string) (string, error) {
	t.Helper()

	fn := getSysDirFunc(kind)
	core.AssertMustNotNil(t, fn, "sys %s dir fn", kind)
	return fn(args...)
}

// callPrefixDirFunc resolves the FooDir method of the given
// [appdir.Prefix] for the category, asserts it was recognised, and
// invokes it.
func callPrefixDirFunc(t *testing.T, p appdir.Prefix, kind Kind,
	args ...string) (string, error) {
	t.Helper()

	fn := getPrefixDirFunc(p, kind)
	core.AssertMustNotNil(t, fn, "prefix %s dir fn", kind)
	return fn(args...)
}
