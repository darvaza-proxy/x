//go:build !windows

package appdir

import (
	"io/fs"
	"path/filepath"
	"strings"

	"darvaza.org/core"
)

const (
	// PrefixLocal represents services installed outside
	// the scope of the package manager.
	PrefixLocal Prefix = "/usr/local"
	// PrefixSystem represents services installed by
	// the package manager.
	PrefixSystem Prefix = "/"
	// PrefixOptional represents services installed outside
	// the scope of the package manager but requiring
	// a complex hierarchy, usually installed by extracting
	// an archive file.
	PrefixOptional Prefix = "/opt"
)

func (p Prefix) sysCacheDir(sub ...string) (string, error) {
	return p.sysDir("/var/cache", sub...)
}

func (p Prefix) sysConfigDir(sub ...string) (string, error) {
	return p.sysDir("/etc", sub...)
}

func (p Prefix) sysDataDir(sub ...string) (string, error) {
	return p.sysDir("/var/lib", sub...)
}

func (p Prefix) sysRuntimeDir(sub ...string) (string, error) {
	return p.sysDir("/var/run", sub...)
}

// getSysPrefix returns the filesystem prefix to prepend to
// system directories. It reports false under [PrefixUser],
// which has no system prefix.
func (p Prefix) getSysPrefix() (string, bool) {
	switch p {
	case PrefixSystem:
		return "", true
	case PrefixUser:
		return "", false
	default:
		return string(p), true
	}
}

func (p Prefix) sysDir(dir string, sub ...string) (string, error) {
	if p == PrefixOptional {
		return getSysOptDir(dir, sub...)
	}

	// every FooDir method handles PrefixUser before calling
	// sysDir; reaching it in user mode means internal misuse.
	s := core.MustOK(p.getSysPrefix())
	return Join(s+dir, sub...), nil
}

func getSysOptDir(dir string, sub ...string) (string, error) {
	// flatten
	var flat string
	switch dir {
	case "/var/lib":
		flat = "/share"
	default:
		flat, _ = strings.CutPrefix(dir, "/var")
	}

	// split
	parts := partsFromSlash(flat, sub...)

	// convert /foo/app/blah to /opt/app/foo/blah
	if len(parts) > 1 {
		parts[0], parts[1] = parts[1], parts[0]
		parts = append([]string{string(PrefixOptional)}, parts...)
		return filepath.Join(parts...), nil
	}

	// application name not specified
	err := &fs.PathError{
		Path: string(PrefixOptional) + dir,
		Op:   "stat",
		Err:  fs.ErrInvalid,
	}

	return "", err
}
