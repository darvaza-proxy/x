package appdir_test

import (
	"testing"

	"darvaza.org/core"
)

func TestUserDirFallback(t *testing.T) {
	testCases := core.S(
		newUserDirFallbackTestCase("cache empty", KindCache,
			"", "/home/test/Library/Caches/app"),
		newUserDirFallbackTestCase("config empty", KindConfig,
			"", "/home/test/Library/Application Support/app"),
		newUserDirFallbackTestCase("data empty", KindData,
			"", "/home/test/Library/Application Support/app"),
		newUserDirFallbackTestCase("cache relative", KindCache,
			"relative/cache", "/home/test/Library/Caches/app"),
		newUserDirFallbackTestCase("config relative", KindConfig,
			"relative/config",
			"/home/test/Library/Application Support/app"),
		newUserDirFallbackTestCase("data relative", KindData,
			"relative/share",
			"/home/test/Library/Application Support/app"),
	)

	core.RunTestCases(t, testCases)
}
