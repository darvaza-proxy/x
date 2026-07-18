// Package appdir determines where an application should keep its
// files — cache, configuration, persistent data, and run-time
// data — following the XDG Base Directory Specification in user
// mode and the Filesystem Hierarchy Standard in system mode,
// under a configurable [Prefix]. On macOS the XDG variables still
// take precedence, but the user-mode fallbacks are the
// Apple-native locations under the user's home — Library/Caches
// and Library/Application Support. On Windows the same interface
// maps to the application data directories: %AppData% and
// %LocalAppData% in user mode, %ProgramData% in system mode.
package appdir

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
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
//
// The predefined constants are platform-specific: unix-like
// systems follow the FHS, with [PrefixSystem] at the root and
// [PrefixLocal] and [PrefixOptional] alongside it, while Windows
// defines [PrefixSystem] alone — the machine-wide %ProgramData%
// directory, resolved when composing.
type Prefix string

// prefix is the default Prefix used by the top-level SysFooDir
// functions. It is stored atomically so [SetSysPrefix] is safe
// against concurrent readers, and initialised to [PrefixUser].
var prefix atomic.Pointer[Prefix]

func init() {
	u := PrefixUser
	prefix.Store(&u)
}

// loadPrefix returns the current default Prefix.
func loadPrefix() Prefix {
	return *prefix.Load()
}

// NewPrefix returns a Prefix for the given directory. A
// well-known hint — [PrefixUser] or one of the system
// prefixes [PrefixSystem], [PrefixLocal] and [PrefixOptional] —
// is symbolic, expanded by the per-OS resolver at composition
// time, so it is returned unchanged. Any other value is treated
// as a path: resolved to absolute and validated via
// [Prefix.Validate]. The empty string is rejected rather than
// silently resolving to the working directory.
func NewPrefix[T string | Prefix](dir T) (Prefix, error) {
	p := Prefix(dir)
	if p.isWellKnown() {
		return p, nil
	}

	if dir == "" {
		// filepath.Abs("") resolves to the working directory,
		// which would anchor the system tree at cwd; reject it
		// as the zero value Validate already does.
		return "", p.Validate()
	}

	s, err := filepath.Abs(string(dir))
	if err != nil {
		return "", err
	}

	p = Prefix(s)
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
	if p.isWellKnown() {
		return nil
	}

	return p.validateDir()
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
// when generating SysFooDir() strings. The well-known
// [PrefixUser] selects user mode, the default. It is safe to
// call concurrently with the SysFooDir readers.
func SetSysPrefix[T string | Prefix](dir T) error {
	p, err := NewPrefix(string(dir))
	if err != nil {
		return err
	}

	prefix.Store(&p)
	return nil
}

// SysPrefix returns the default Prefix used by the top-level
// SysFooDir functions.
func SysPrefix() Prefix {
	return loadPrefix()
}

// UserCacheDir returns where to store application cache
// when run in user mode.
// ${XDG_CACHE_HOME}/... (%LocalAppData% on Windows).
func UserCacheDir(sub ...string) (string, error) {
	return joinFn(getUserCacheDir, sub...)
}

// UserConfigDir returns where to store application configuration,
// when run in user mode.
// ${XDG_CONFIG_HOME}/... (%AppData% on Windows).
func UserConfigDir(sub ...string) (string, error) {
	return joinFn(getUserConfigDir, sub...)
}

// UserDataDir returns where to store application persistent
// data, when run in user mode.
// ${XDG_DATA_HOME}/... (%AppData% on Windows, shared with
// configuration).
func UserDataDir(sub ...string) (string, error) {
	return joinFn(getUserDataDir, sub...)
}

// UserRuntimeDir returns where to store application run-time
// variable data, when run in user mode.
// ${XDG_RUNTIME_DIR}/...
//
// When ${XDG_RUNTIME_DIR} is unset — and, outside macOS, the
// systemd /run/user/<uid> directory doesn't exist — it falls
// back to ${TMPDIR:-/tmp}/runtime-<user> without creating it.
// Callers must create the fallback with 0700 permissions to
// honour the XDG trust requirements. On Windows a
// runtime-<user> directory under the user's temporary directory
// is used instead: %TMP% (or %TEMP%) when defined, otherwise
// the default %LocalAppData%\Temp.
func UserRuntimeDir(sub ...string) (string, error) {
	return joinFn(getUserRuntimeDir, sub...)
}

// CacheDir returns where to store application cache under
// this Prefix. Under [PrefixUser] it returns the same as
// [UserCacheDir]. Under [PrefixOptional] — and on Windows under
// every system-mode Prefix — the application name is required,
// and its absence is an error.
func (p Prefix) CacheDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserCacheDir(sub...)
	}

	return p.sysCacheDir(sub...)
}

// ConfigDir returns where to store application configuration
// data under this Prefix. Under [PrefixUser] it returns the
// same as [UserConfigDir]. Under [PrefixOptional] — and on
// Windows under every system-mode Prefix — the application
// name is required, and its absence is an error.
func (p Prefix) ConfigDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserConfigDir(sub...)
	}

	return p.sysConfigDir(sub...)
}

// DataDir returns where to store application persistent data
// under this Prefix. Under [PrefixUser] it returns the same
// as [UserDataDir]. Under [PrefixOptional] — and on Windows
// under every system-mode Prefix — the application name is
// required, and its absence is an error.
func (p Prefix) DataDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserDataDir(sub...)
	}

	return p.sysDataDir(sub...)
}

// RuntimeDir returns where to store application run-time
// variable data under this Prefix. Under [PrefixUser] it
// returns the same as [UserRuntimeDir]. Under [PrefixOptional]
// — and on Windows under every system-mode Prefix — the
// application name is required, and its absence is an error.
func (p Prefix) RuntimeDir(sub ...string) (string, error) {
	if p == PrefixUser {
		return UserRuntimeDir(sub...)
	}

	return p.sysRuntimeDir(sub...)
}

// SysCacheDir returns where to store application cache,
// when run in system mode.
func SysCacheDir(sub ...string) (string, error) {
	return loadPrefix().CacheDir(sub...)
}

// SysConfigDir returns where to store application configuration
// data, when run in system mode.
func SysConfigDir(sub ...string) (string, error) {
	return loadPrefix().ConfigDir(sub...)
}

// SysDataDir returns where to store application persistent
// data, when run in system mode.
func SysDataDir(sub ...string) (string, error) {
	return loadPrefix().DataDir(sub...)
}

// SysRuntimeDir returns where to store application run-time
// variable data, when run in system mode.
func SysRuntimeDir(sub ...string) (string, error) {
	return loadPrefix().RuntimeDir(sub...)
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
	return loadPrefix().AllConfigDir(sub...)
}
