package appdir

import "darvaza.org/core"

// The native Windows resolvers are not implemented yet: each
// entry point reports [core.ErrNotImplemented] so the package
// builds and vets on Windows. The real implementation lands with
// the Windows support work, replacing this file in place.

func getUserCacheDir() (string, error) {
	return "", core.ErrNotImplemented
}

func getUserConfigDir() (string, error) {
	return "", core.ErrNotImplemented
}

func getUserDataDir() (string, error) {
	return "", core.ErrNotImplemented
}

func getUserRuntimeDir() (string, error) {
	return "", core.ErrNotImplemented
}

func (Prefix) sysCacheDir(...string) (string, error) {
	return "", core.ErrNotImplemented
}

func (Prefix) sysConfigDir(...string) (string, error) {
	return "", core.ErrNotImplemented
}

func (Prefix) sysDataDir(...string) (string, error) {
	return "", core.ErrNotImplemented
}

func (Prefix) sysRuntimeDir(...string) (string, error) {
	return "", core.ErrNotImplemented
}
