// Package appdir provides helpers to compose filenames
// to be used by applications.
package appdir

import (
	"io/fs"
	"os"
	"path/filepath"
)

// PrefixUser is the prefix used on
// [SetSysPrefix] to indicate SysFooDir() will return
// the same as UserFooDir()
const PrefixUser = "~"

var prefix = PrefixUser

// SetSysPrefix specifies what filesystem prefix to use
// when generating SysFooDir() strings.
// Default is "~"
func SetSysPrefix(dir string) error {
	s, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	st, err := os.Stat(s)
	switch {
	case err != nil:
		return err
	case !st.IsDir():
		err = &fs.PathError{
			Path: s,
			Op:   "readdir",
			Err:  fs.ErrInvalid,
		}
		return err
	default:
		prefix = s
		return nil
	}
}

// UserCacheDir returns where to store application cache
// when run in user mode.
// ${XDG_CACHE_HOME}/...
func UserCacheDir(sub ...string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return Join(dir, sub...), nil
}

// UserConfigDir returns where to store application configuration,
// when run in user mode.
// ${XDG_CONFIG_HOME}/...
func UserConfigDir(sub ...string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return Join(dir, sub...), nil
}

// UserDataDir returns where to store application persistent
// data, when run in user mode.
// ${XDG_DATA_HOME}/...
func UserDataDir(sub ...string) (string, error) {
	dir, err := getUserDataDir()
	if err != nil {
		return "", err
	}
	return Join(dir, sub...), nil
}

// UserRuntimeDir returns where to store application run-time
// variable data, when run in user mode.
// ${XDG_RUNTIME_DIR}/...
func UserRuntimeDir(sub ...string) (string, error) {
	dir := getUserRuntimeDir()
	return Join(dir, sub...), nil
}

// SysCacheDir returns where to store application cache,
// when run in system mode.
func SysCacheDir(sub ...string) (string, error) {
	if prefix == PrefixUser {
		return UserCacheDir(sub...)
	}

	dir, err := getSysCacheDir()
	if err != nil {
		return "", err
	}

	return Join(dir, sub...), nil
}

// SysConfigDir returns where to store application configuration
// data, when run in system mode.
func SysConfigDir(sub ...string) (string, error) {
	if prefix == PrefixUser {
		return UserConfigDir(sub...)
	}

	dir, err := getSysConfigDir()
	if err != nil {
		return "", err
	}

	return Join(dir, sub...), nil
}

// SysDataDir returns where to store application persistent
// data, when run in system mode.
func SysDataDir(sub ...string) (string, error) {
	if prefix == PrefixUser {
		return UserDataDir(sub...)
	}

	dir, err := getSysDataDir()
	if err != nil {
		return "", err
	}

	return Join(dir, sub...), nil
}

// SysRuntimeDir returns where to store application run-time
// variable data, when run in system mode.
func SysRuntimeDir(sub ...string) (string, error) {
	if prefix == PrefixUser {
		return UserRuntimeDir(sub...)
	}

	dir, err := getSysRuntimeDir()
	if err != nil {
		return "", err
	}

	return Join(dir, sub...), nil
}

// Join combines file path parts in a OS specific way.
// parts are allowed to be multipart themselves, using `/`
// as delimiter.
func Join(base string, sub ...string) string {
	a := partsFromSlash(base, sub...)
	return filepath.Join(a...)
}

func partsFromSlash(base string, sub ...string) []string {
	var n int

	pp := make([][]string, len(sub))
	for i, s := range sub {
		pp[i] = splitFromSlash(s)
		n += len(pp[i])
	}

	if base != "" {
		n++
	}

	a := make([]string, 0, n)
	if base != "" {
		a = append(a, base)
	}
	for _, p := range pp {
		a = append(a, p...)
	}

	return a
}

func splitFromSlash(s string) []string {
	return filepath.SplitList(filepath.FromSlash(s))
}
