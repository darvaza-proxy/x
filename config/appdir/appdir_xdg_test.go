//go:build !windows && !darwin

package appdir_test

import (
	"testing"

	"darvaza.org/core"
)

func TestUserDirFallback(t *testing.T) {
	testCases := core.S(
		newUserDirFallbackTestCase("cache empty", KindCache,
			"", "/home/test/.cache/app"),
		newUserDirFallbackTestCase("config empty", KindConfig,
			"", "/home/test/.config/app"),
		newUserDirFallbackTestCase("data empty", KindData,
			"", "/home/test/.local/share/app"),
		newUserDirFallbackTestCase("cache relative", KindCache,
			"relative/cache", "/home/test/.cache/app"),
		newUserDirFallbackTestCase("config relative", KindConfig,
			"relative/config", "/home/test/.config/app"),
		newUserDirFallbackTestCase("data relative", KindData,
			"relative/share", "/home/test/.local/share/app"),
	)

	core.RunTestCases(t, testCases)
}
