//go:build !windows && !darwin

package appdir

import (
	"os"
	"strconv"
)

func getUserCacheDir() (string, error) {
	return getEnvHomeDir("XDG_CACHE_HOME", ".cache")
}

func getUserConfigDir() (string, error) {
	return getEnvHomeDir("XDG_CONFIG_HOME", ".config")
}

func getUserDataDir() (string, error) {
	return getEnvHomeDir("XDG_DATA_HOME", ".local/share")
}

func getUserRuntimeDir() (string, error) {
	dir := getEnvDir("XDG_RUNTIME_DIR")
	if dir != "" {
		return dir, nil
	}

	// systemd special
	dir = "/run/user/" + strconv.Itoa(os.Getuid())
	st, _ := os.Stat(dir)
	if st != nil && st.IsDir() {
		return dir, nil
	}

	return getUserRuntimeTempDir()
}
