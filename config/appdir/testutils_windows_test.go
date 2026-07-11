// cspell:words LOCALAPPDATA
package appdir_test

import (
	"testing"

	"darvaza.org/core"
)

// getWinEnvKey returns the environment variable overriding the
// given user-mode directory category, or false if the category is
// not recognised. Run-time directories compose under the
// temporary directory, so %TMP% overrides them.
func getWinEnvKey(kind Kind) (string, bool) {
	switch kind {
	case KindCache:
		return "LOCALAPPDATA", true
	case KindConfig, KindData:
		return "APPDATA", true
	case KindRuntime:
		return "TMP", true
	default:
		return "", false
	}
}

// setWinEnv sets the environment variable overriding the given
// category to value, failing the test if the category is not
// recognised.
func setWinEnv(t *testing.T, kind Kind, value string) {
	t.Helper()

	key, ok := getWinEnvKey(kind)
	core.AssertMustTrue(t, ok, "%s env key", kind)
	t.Setenv(key, value)
}
