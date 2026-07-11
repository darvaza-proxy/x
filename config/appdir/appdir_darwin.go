package appdir

// The XDG Base Directory variables take precedence when set to
// absolute paths, and the fallbacks are the Apple-native
// locations under the user's home.

func getUserCacheDir() (string, error) {
	return getEnvHomeDir("XDG_CACHE_HOME", "Library/Caches")
}

func getUserConfigDir() (string, error) {
	return getEnvHomeDir("XDG_CONFIG_HOME",
		"Library/Application Support")
}

func getUserDataDir() (string, error) {
	return getEnvHomeDir("XDG_DATA_HOME",
		"Library/Application Support")
}

// getUserRuntimeDir honours ${XDG_RUNTIME_DIR} and otherwise
// falls back to runtime-<user> under the temporary directory;
// $TMPDIR is per-user on macOS, keeping the fallback volatile
// and user-specific.
func getUserRuntimeDir() (string, error) {
	dir := getEnvDir("XDG_RUNTIME_DIR")
	if dir != "" {
		return dir, nil
	}

	return getUserRuntimeTempDir()
}
