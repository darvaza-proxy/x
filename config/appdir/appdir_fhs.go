//go:build !windows

package appdir

import (
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	// PrefixLocal represent services installed outside
	// the scope of the package manager
	PrefixLocal = "/usr/local"
	// PrefixSystem represent services installed by
	// the package manager
	PrefixSystem = "/"
	// PrefixOptional represent services installed outside
	// the scope of the package manager but require
	// a complex hierarchy, usually installed extracting
	// an archive file.
	PrefixOptional = "/opt"
)

func getSysCacheDir(sub ...string) (string, error) {
	return getSysDir("/var/cache", sub...)
}

func getSysConfigDir(sub ...string) (string, error) {
	return getSysDir("/etc", sub...)
}

func getSysDataDir(sub ...string) (string, error) {
	return getSysDir("/var/lib", sub...)
}

func getSysRuntimeDir(sub ...string) (string, error) {
	return getSysDir("/var/run", sub...)
}

func getSysDir(dir string, sub ...string) (string, error) {
	switch prefix {
	case PrefixOptional:
		return getSysOptDir(dir, sub...)
	case PrefixSystem:
		// ready
	case PrefixUser:
		panic("unreachable")
	default:
		dir = prefix + dir
	}

	return Join(dir, sub...), nil
}

func getSysOptDir(dir string, sub ...string) (string, error) {
	// flatten
	switch dir {
	case "/var/lib":
		dir = "/share"
	default:
		dir, _ = strings.CutPrefix(dir, "/var")
	}

	// split
	parts := partsFromSlash(dir, sub...)

	// convert /foo/app/blah to /opt/app/foo/blah
	if len(parts) > 1 {
		parts[0], parts[1] = parts[1], parts[0]
		dir = PrefixOptional + filepath.Join(parts...)
		return dir, nil
	}

	// application name not specified
	err := &fs.PathError{
		Path: "/opt" + dir,
		Op:   "stat",
		Err:  fs.ErrInvalid,
	}

	return "", err
}
