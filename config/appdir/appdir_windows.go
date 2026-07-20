package appdir

// cspell:words LOCALAPPDATA USERPROFILE writable

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"

	"darvaza.org/core"
)

// system-mode directory categories, composed as
// <root>\<app>\<category>.
const (
	sysCacheName   = "cache"
	sysConfigName  = "config"
	sysDataName    = "data"
	sysRuntimeName = "run"
)

func (p Prefix) sysCacheDir(sub ...string) (string, error) {
	return p.sysDir(sysCacheName, sub...)
}

func (p Prefix) sysConfigDir(sub ...string) (string, error) {
	return p.sysDir(sysConfigName, sub...)
}

func (p Prefix) sysDataDir(sub ...string) (string, error) {
	return p.sysDir(sysDataName, sub...)
}

func (p Prefix) sysRuntimeDir(sub ...string) (string, error) {
	return p.sysDir(sysRuntimeName, sub...)
}

// getSysRoot returns the filesystem root under which the system
// directories are composed, resolving [PrefixSystem] via
// %ProgramData% — where a non-absolute value is unusable and
// treated as undefined. It reports false under [PrefixUser] and
// the zero value, which have no system root.
func (p Prefix) getSysRoot() (string, bool) {
	switch p {
	case PrefixSystem:
		return getEnvDir("ProgramData"), true
	case PrefixUser, "":
		return "", false
	default:
		return string(p), true
	}
}

func (p Prefix) sysDir(category string, sub ...string) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}

	// every FooDir method handles PrefixUser before calling
	// sysDir; reaching it in user mode means internal misuse.
	root := core.MustOK(p.getSysRoot())
	if root == "" {
		return "", errors.New("%ProgramData% is not defined")
	}

	return sysRootDir(root, category, sub...)
}

// sysRootDir composes <root>\<app>\<category> with any remaining
// sub-path elements appended, requiring the application name.
func sysRootDir(root, category string, sub ...string) (string, error) {
	parts := partsFromSlash("", sub...)
	if len(parts) == 0 {
		// application name not specified
		err := &fs.PathError{
			Path: filepath.Join(root, category),
			Op:   "stat",
			Err:  fs.ErrInvalid,
		}
		return "", err
	}

	// convert app/blah to <root>/app/<category>/blah
	out := make([]string, 0, len(parts)+2)
	out = append(out, root, parts[0], category)
	out = append(out, parts[1:]...)
	return filepath.Join(out...), nil
}

func getUserCacheDir() (string, error) {
	// %LocalAppData% designates the local application data
	// directory; AppData\Local is its default value under the
	// user's profile.
	return getEnvHomeDir("LOCALAPPDATA", "AppData/Local")
}

func getUserConfigDir() (string, error) {
	// %AppData% designates the roaming application data
	// directory; AppData\Roaming is its default value under
	// the user's profile.
	return getEnvHomeDir("APPDATA", "AppData/Roaming")
}

func getUserDataDir() (string, error) {
	// the roaming application data directory holds both
	// configuration and persistent data
	return getUserConfigDir()
}

func getUserRuntimeDir() (string, error) {
	// the temporary directory is volatile but carries no
	// ownership contract, so a user-distinguishing leaf keeps
	// run-time data user-specific by construction.
	name, err := userName()
	if err != nil {
		return "", err
	}

	return joinFn(getTempDir, "runtime-"+name)
}

// userName identifies the current user for path composition:
// %USERNAME%, falling back to the account SID.
// [user.User.Username] is DOMAIN\name, unusable in a path.
func userName() (string, error) {
	if name := os.Getenv("USERNAME"); name != "" {
		return name, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Uid, nil
}

// getTempDir returns the user's volatile temporary directory.
// %TMP% and %TEMP% designate it per-session and a redirection is
// deliberate and honoured, a relative one resolved against the
// working directory so the result stays absolute; otherwise it
// composes their default per-user value, %LocalAppData%\Temp —
// falling back to its own default under %USERPROFILE% — still a
// temporary directory, excluded from back-ups and subject to
// cleaning. There is deliberately no tier below the user's
// profile: the roots os.TempDir would degrade to are either
// precious and backed up (%USERPROFILE% itself) or shared and
// not user-writable (the Windows directory), so a profile-less
// environment is an error instead.
func getTempDir() (string, error) {
	if dir := core.Coalesce(os.Getenv("TMP"), os.Getenv("TEMP")); dir != "" {
		return filepath.Abs(dir)
	}

	return joinFn(getUserCacheDir, "Temp")
}
