//go:build !windows

package fs_test

// cleanOSPathPlatformCases are the rows for '/' as the only separator,
// where the backslash is an ordinary character in a file name.
func cleanOSPathPlatformCases() []cleanOSPathTestCase {
	return []cleanOSPathTestCase{
		newCleanOSPathTestCase("leading doubled separator", "//a", "/a"),
		newCleanOSPathTestCase("holds a backslash", `a\b`, `a\b`),
		newCleanOSPathTestCase("ends in a backslash", `a\`, `a\`),
		newCleanOSPathTestCase("dot and backslash", `.\a`, `.\a`),
		newCleanOSPathTestCase("dot dots and backslashes", `..\a\..`, `..\a\..`),
		newCleanOSPathTestCase(`\\?\, a drive and a dot`, `\\?\C:\a\.\b`, `\\?\C:\a\.\b`),
		newCleanOSPathTestCase("dot dot after a name holding a backslash", `a\b/..`, `.`),
		newCleanOSPathTestCase("question mark after doubled separator", `//?/a/./b`, `/?/a/b`),
		newCleanOSPathTestCase("question marks after separator", `/??/a/./b`, `/??/a/b`),
		newCleanOSPathTestCase("question mark after two backslashes", `\\?/a/./b`, `\\?/a/b`),
		newCleanOSPathTestCase("question marks after a backslash", `\??/a/./b`, `\??/a/b`),
	}
}

// cleanOSPathInvalidPlatformCases are none, as the paths Windows refuses
// after \\?\ hold ordinary names here.
func cleanOSPathInvalidPlatformCases() []cleanOSPathInvalidTestCase {
	return nil
}

// cleanOSPathUnchangedPlatformCases are the clean paths holding a
// backslash, an ordinary character in a file name.
func cleanOSPathUnchangedPlatformCases() []cleanOSPathUnchangedTestCase {
	return []cleanOSPathUnchangedTestCase{
		newCleanOSPathUnchangedTestCase("holds a backslash", `a\b`),
		newCleanOSPathUnchangedTestCase("dot and backslash", `.\a`),
	}
}
