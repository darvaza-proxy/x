//go:build !windows

package appdir_test

import (
	"testing"

	"darvaza.org/core"
)

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
