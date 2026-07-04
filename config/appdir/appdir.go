// Package appdir determines where an application should keep its
// files — cache, configuration, persistent data, and run-time
// data — following the XDG Base Directory Specification in user
// mode and the Filesystem Hierarchy Standard in system mode,
// under a configurable [Prefix].
package appdir

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"darvaza.org/core"
)

// Prefix is a filesystem prefix under which the system-mode
// directories are composed. The special value [PrefixUser]
// makes the FooDir methods return the same as their
// UserFooDir counterparts. Any other value must be one of the
// well-known prefixes or an absolute path to an existing
// directory — see [Prefix.Validate] — and the FooDir methods
// reject malformed values, including the zero value, with an
// error.
type Prefix string

// PrefixUser is the Prefix indicating user mode, where the
// FooDir methods return the same as UserFooDir().
const PrefixUser Prefix = "~"

// prefix is the default Prefix used by the top-level
// SysFooDir functions.
var prefix = PrefixUser

// NewPrefix returns a Prefix for the given directory, resolved
// to an absolute path and validated via [Prefix.Validate]. The
// special value "~" returns [PrefixUser] instead.
func NewPrefix(dir string) (Prefix, error) {
	if Prefix(dir) == PrefixUser {
		return PrefixUser, nil
	}

	s, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	p := Prefix(s)
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}

// Validate reports whether the Prefix is usable: one of the
// well-known prefixes — [PrefixUser], [PrefixSystem],
// [PrefixLocal] or [PrefixOptional] — or an absolute path to an
// existing directory. Anything else — including the zero value,
// which carries no root — fails with [fs.ErrInvalid], the stat
// error, or [syscall.ENOTDIR].
func (p Prefix) Validate() error {
	switch p {
	case PrefixUser, PrefixSystem, PrefixLocal, PrefixOptional:
		return nil
	default:
		return p.validateDir()
	}
}

// validateDir requires the Prefix to be an absolute path to an
// existing directory.
func (p Prefix) validateDir() error {
	if !filepath.IsAbs(string(p)) {
		return core.Wrapf(fs.ErrInvalid, "invalid prefix %q",
			string(p))
	}

	st, err := os.Stat(string(p))
	switch {
	case err != nil:
		return err
	case !st.IsDir():
		return &fs.PathError{
			Path: string(p),
			Op:   "stat",
			Err:  syscall.ENOTDIR,
		}
	default:
		return nil
	}
}

// SetSysPrefix specifies what filesystem prefix to use
// when generating SysFooDir() strings. The special value "~"
// selects user mode ([PrefixUser]), the default.
func SetSysPrefix(dir string) error {
	p, err := NewPrefix(dir)
	if err != nil {
		return err
	}

	prefix = p
	return nil
}

// SysPrefix returns the default Prefix used by the top-level
// SysFooDir functions.
func SysPrefix() Prefix {
	return prefix
}

// UserCacheDir returns where to store application cache
// when run in user mode.
// ${XDG_CACHE_HOME}/...
func UserCacheDir(sub ...string) (string, error) {
	return joinFn(os.UserCacheDir, sub...)
}

// UserConfigDir returns where to store application configuration,
// when run in user mode.
// ${XDG_CONFIG_HOME}/...
func UserConfigDir(sub ...string) (string, error) {
	return joinFn(os.UserConfigDir, sub...)
}

// UserDataDir returns where to store application persistent
// data, when run in user mode.
// ${XDG_DATA_HOME}/...
func UserDataDir(sub ...string) (string, error) {
	return joinFn(getUserDataDir, sub...)
}

// UserRuntimeDir returns where to store application run-time
// variable data, when run in user mode.
// ${XDG_RUNTIME_DIR}/...
//
// When ${XDG_RUNTIME_DIR} is unset and the systemd
// /run/user/<uid> directory doesn't exist, it falls back to
// ${TMPDIR:-/tmp}/runtime-<user> without creating it. Callers
// must create the fallback with 0700 permissions to honour the
// XDG trust requirements.
func UserRuntimeDir(sub ...string) (string, error) {
	return joinFn(getUserRuntimeDir, sub...)
}

// CacheDir returns where to store application cache under
// this Prefix. Under [PrefixUser] it returns the same as
// [UserCacheDir]. Under [PrefixOptional] the application name
// is required, and its absence is an error.
func (p Prefix) CacheDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserCacheDir(sub...)
	}

	return p.sysCacheDir(sub...)
}

// ConfigDir returns where to store application configuration
// data under this Prefix. Under [PrefixUser] it returns the
// same as [UserConfigDir]. Under [PrefixOptional] the
// application name is required, and its absence is an error.
func (p Prefix) ConfigDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserConfigDir(sub...)
	}

	return p.sysConfigDir(sub...)
}

// DataDir returns where to store application persistent data
// under this Prefix. Under [PrefixUser] it returns the same
// as [UserDataDir]. Under [PrefixOptional] the application
// name is required, and its absence is an error.
func (p Prefix) DataDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserDataDir(sub...)
	}

	return p.sysDataDir(sub...)
}

// RuntimeDir returns where to store application run-time
// variable data under this Prefix. Under [PrefixUser] it
// returns the same as [UserRuntimeDir]. Under [PrefixOptional]
// the application name is required, and its absence is an
// error.
func (p Prefix) RuntimeDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserRuntimeDir(sub...)
	}

	return p.sysRuntimeDir(sub...)
}

// SysCacheDir returns where to store application cache,
// when run in system mode.
func SysCacheDir(sub ...string) (string, error) {
	return prefix.CacheDir(sub...)
}

// SysConfigDir returns where to store application configuration
// data, when run in system mode.
func SysConfigDir(sub ...string) (string, error) {
	return prefix.ConfigDir(sub...)
}

// SysDataDir returns where to store application persistent
// data, when run in system mode.
func SysDataDir(sub ...string) (string, error) {
	return prefix.DataDir(sub...)
}

// SysRuntimeDir returns where to store application run-time
// variable data, when run in system mode.
func SysRuntimeDir(sub ...string) (string, error) {
	return prefix.RuntimeDir(sub...)
}

// AllConfigDir returns a slice containing the application
// configuration path on the current working directory, user
// mode, and this Prefix's system mode.
func (p Prefix) AllConfigDir(sub ...string) []string {
	var u string

	out := make([]string, 0, 3)

	// .
	if s, _ := os.Getwd(); s != "" {
		parts := partsFromSlash(s, sub...)
		if len(parts) > 1 {
			// ./app/foo -> ./foo
			parts[1] = parts[0]
			parts = parts[1:]
		}

		out = append(out, filepath.Join(parts...))
	}

	// ~
	if u, _ = UserConfigDir(sub...); u != "" {
		out = append(out, u)
	}

	// /
	if s, _ := p.ConfigDir(sub...); s != "" && s != u {
		out = append(out, s)
	}

	return out
}

// AllConfigDir returns a slice containing the application
// configuration path on the current working directory, user mode,
// and system mode.
func AllConfigDir(sub ...string) []string {
	return prefix.AllConfigDir(sub...)
}
