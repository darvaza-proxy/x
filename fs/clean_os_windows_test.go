//go:build windows

package fs_test

// cleanOSPathPlatformCases are the rows for '/' and '\' as separators,
// and for volume names.
func cleanOSPathPlatformCases() []cleanOSPathTestCase {
	return []cleanOSPathTestCase{
		// Separators.
		newCleanOSPathTestCase("slash", `a/b`, `a\b`),
		newCleanOSPathTestCase("doubled backslash", `a\\b`, `a\b`),
		newCleanOSPathTestCase("mixed separators", `a\/b`, `a\b`),
		newCleanOSPathTestCase("ends in a backslash", `a\`, `a\`),
		// A "." stays, or is put first, where a path would read as one
		// with a volume name.
		newCleanOSPathTestCase("dot before a colon", `.\c:a`, `.\c:a`),
		newCleanOSPathTestCase("dot in slashes before a colon", `./c:a`, `.\c:a`),
		newCleanOSPathTestCase("doubled dot before a colon", `.\.\c:a`, `.\c:a`),
		newCleanOSPathTestCase("rooted dot before question marks", `\.\??\c:\a`, `\.\??\c:\a`),
		newCleanOSPathTestCase("drive and dot before a colon", `C:.\d:a`, `C:d:a`),
		newCleanOSPathTestCase("dot dot before a colon", `a\..\c:b`, `.\c:b`),
		newCleanOSPathTestCase("rooted dot dot before question marks", `\a\..\??\b`, `\.\??\b`),
		// Dot dots between doubled backslashes.
		newCleanOSPathTestCase("doubled backslashes and dot dots", `..\\a\\..`, `..`),
		// A rooted path without a volume starts at the current drive's
		// root, whose parent is that root.
		newCleanOSPathTestCase("rooted dot dot without a volume", `\..\a`, `\a`),
		newCleanOSPathTestCase("root and dot dot without a volume", `\..`, `\`),
		// Drives.
		newCleanOSPathTestCase("drive", `C:`, `C:`),
		newCleanOSPathTestCase("drive and dot", `C:.`, `C:.`),
		newCleanOSPathTestCase("drive root", `C:\`, `C:\`),
		newCleanOSPathTestCase("drive relative", `C:a\.\b`, `C:a\b`),
		newCleanOSPathTestCase("drive root dot dot", `C:\..\a`, `C:\a`),
		newCleanOSPathTestCase("drive relative dot dot", `C:..\a`, `C:..\a`),
		newCleanOSPathTestCase("drive relative dot dot to nothing", `C:a\..`, `C:.`),
		newCleanOSPathTestCase("drive, dot and separator", `C:.\`, `C:.`),
		newCleanOSPathTestCase("drive relative trailing dot dot", `C:a\b\..`, `C:a\`),
		// Shares.
		newCleanOSPathTestCase("share", `\\host\share`, `\\host\share`),
		newCleanOSPathTestCase("share and separator", `\\host\share\`, `\\host\share\`),
		newCleanOSPathTestCase("host without a share", `//a`, `\\a`),
		newCleanOSPathTestCase("share and dot", `\\host\share\.\a`, `\\host\share\a`),
		newCleanOSPathTestCase("share in slashes", `//host/share/a`, `\\host\share\a`),
		newCleanOSPathTestCase("share and dot dots", `\\host\share\a\..\..\b`, `\\host\share\b`),
		// A host and share are taken as written, dot dots included.
		newCleanOSPathTestCase("share named dot dot", `\\host\..\a`, `\\host\..\a`),
		newCleanOSPathTestCase("share named dot dot in slashes", `//host/../a`, `\\host\..\a`),
		newCleanOSPathTestCase("share named dot dot alone", `\\host\..`, `\\host\..`),
		newCleanOSPathTestCase("share named dot dot and dot dots", `\\host\..\a\..\..\b`, `\\host\..\b`),
		newCleanOSPathTestCase("host named dot dot", `\\..\share\a`, `\\..\share\a`),
		newCleanOSPathTestCase("host and share named dot dot", `\\..\..\a`, `\\..\..\a`),
		// A separator before a colon starts a rooted path, not a drive.
		newCleanOSPathTestCase("rooted colon and dot dot", `\:a\..\b`, `\b`),
		newCleanOSPathTestCase("rooted colon in slashes", `/:a/../b`, `\b`),
		newCleanOSPathTestCase("rooted colon and dot dot alone", `\:a\..`, `\`),
		newCleanOSPathTestCase("rooted colon alone", `\:`, `\:`),
		newCleanOSPathTestCase("dot dot before a rooted colon", `\a\..\:b`, `\:b`),
		// Paths starting \\?\ or \??\ reach the device as written, and stay.
		newCleanOSPathTestCase(`\\?\ and a drive`, `\\?\C:\a\b`, `\\?\C:\a\b`),
		newCleanOSPathTestCase(`\\?\, a drive and a separator`, `\\?\C:\a\`, `\\?\C:\a\`),
		newCleanOSPathTestCase(`\\?\ and a share`, `\\?\UNC\host\share\a`, `\\?\UNC\host\share\a`),
		newCleanOSPathTestCase(`\??\ and a drive`, `\??\C:\a\b`, `\??\C:\a\b`),
		newCleanOSPathTestCase(`\\?\, a pipe and a dot`, `\\?\pipe\a\.\b`, `\\?\pipe\a\.\b`),
		newCleanOSPathTestCase(`\\?\, a pipe and a slash`, `\\?\pipe\a/b`, `\\?\pipe\a/b`),
		newCleanOSPathTestCase(`\??\, a pipe and a dot dot`, `\??\pipe\a\..\b`, `\??\pipe\a\..\b`),
		newCleanOSPathTestCase(`\\?\, a digit like a drive and a dot`, `\\?\1:\a\.\b`, `\\?\1:\a\.\b`),
		newCleanOSPathTestCase(`\\?\ and a leading dot dot`, `\\?\..\a`, `\\?\..\a`),
		newCleanOSPathTestCase(`\??\ and a leading dot dot`, `\??\..\a`, `\??\..\a`),
		// Spelt with a '/', or alone, \\? starts a device path, and \?? a
		// rooted one.
		newCleanOSPathTestCase(`\\.\ in slashes`, `//./C:/a/./b`, `\\.\C:\a\b`),
		newCleanOSPathTestCase(`\\?\ in slashes`, `//?/C:/a//b`, `\\.\C:\a\b`),
		newCleanOSPathTestCase(`\\?\ in slashes and a dot dot`, `//?/C:/a/../b`, `\\.\C:\b`),
		// In a device path, \\. alone is the root, and a dot dot can remove
		// the device name after it.
		newCleanOSPathTestCase(`\\.\ and a dot dot`, `\\.\C:\..\a`, `\\.\a`),
		newCleanOSPathTestCase(`\\.\ and dot dots`, `\\.\C:\a\..\..\b`, `\\.\b`),
		newCleanOSPathTestCase(`\\.\ and a dot dot past the root`, `\\.\..\a`, `\\.\a`),
		newCleanOSPathTestCase(`\\.\ and a dot dot alone`, `\\.\..`, `\\.\`),
		newCleanOSPathTestCase(`\\.\UNC\ and a dot dot`, `\\.\UNC\host\share\..\a`, `\\.\UNC\host\a`),
		newCleanOSPathTestCase(`\\?\ in slashes and dot dots`, `//?/C:/../a`, `\\.\a`),
		newCleanOSPathTestCase(`\\? and a slash`, `\\?/C:\a\.\b`, `\\.\C:\a\b`),
		newCleanOSPathTestCase(`\\?\ led by a slash`, `/\?\C:\a`, `\\.\C:\a`),
		newCleanOSPathTestCase(`\\?\UNC\ in slashes`, `//?/UNC/host/share/a`, `\\.\UNC\host\share\a`),
		newCleanOSPathTestCase(`\??\ in slashes`, `/??/C:/a`, `\.\??\C:\a`),
		newCleanOSPathTestCase(`\?? and a slash`, `\??/C:\a`, `\.\??\C:\a`),
		newCleanOSPathTestCase(`\\? alone`, `\\?`, `\\.`),
		newCleanOSPathTestCase(`\?? alone`, `\??`, `\.\??`),
		newCleanOSPathTestCase("host named like a prefix", `\\?a\share\.\b`, `\\?a\share\b`),
	}
}

// cleanOSPathInvalidPlatformCases are the paths under a drive or share
// after \\?\ or \??\, which Windows passes on as written, holding a
// segment the file system refuses.
func cleanOSPathInvalidPlatformCases() []cleanOSPathInvalidTestCase {
	return []cleanOSPathInvalidTestCase{
		newCleanOSPathInvalidTestCase(`\\?\, a drive and a dot`, `\\?\C:\a\.\b`),
		newCleanOSPathInvalidTestCase(`\\?\, a drive and a trailing dot`, `\\?\C:\a\.`),
		newCleanOSPathInvalidTestCase(`\\?\, a drive and a dot dot`, `\\?\C:\a\..\b`),
		newCleanOSPathInvalidTestCase(`\\?\, a drive and a slash`, `\\?\C:\a/b`),
		newCleanOSPathInvalidTestCase(`\\?\, a drive and a doubled separator`, `\\?\C:\a\\b`),
		newCleanOSPathInvalidTestCase(`\\?\, a drive root and a dot`, `\\?\C:\.\a`),
		newCleanOSPathInvalidTestCase(`\\?\, a lower case drive and a dot`, `\\?\c:\a\.\b`),
		newCleanOSPathInvalidTestCase(`\\?\, a share and a dot`, `\\?\UNC\host\share\.\a`),
		newCleanOSPathInvalidTestCase(`\\?\, a share and a dot dot`, `\\?\unc\host\..\a`),
		newCleanOSPathInvalidTestCase(`\??\, a drive and a dot`, `\??\C:\a\.\b`),
		newCleanOSPathInvalidTestCase(`\??\, a drive and a slash`, `\??\C:\a/b`),
	}
}

// cleanOSPathUnchangedPlatformCases are the clean paths with volume
// names, and those the kernel takes as written.
func cleanOSPathUnchangedPlatformCases() []cleanOSPathUnchangedTestCase {
	return []cleanOSPathUnchangedTestCase{
		newCleanOSPathUnchangedTestCase("drive", `C:`),
		newCleanOSPathUnchangedTestCase("drive root", `C:\`),
		newCleanOSPathUnchangedTestCase("drive path", `C:\srv\index.html`),
		newCleanOSPathUnchangedTestCase("drive relative", `C:a\b`),
		newCleanOSPathUnchangedTestCase("share", `\\host\share`),
		newCleanOSPathUnchangedTestCase("share path", `\\host\share\a`),
		newCleanOSPathUnchangedTestCase("share named dot dot", `\\host\..\a`),
		newCleanOSPathUnchangedTestCase("share named dot dot alone", `\\host\..`),
		newCleanOSPathUnchangedTestCase("share named dot alone", `\\host\.`),
		newCleanOSPathUnchangedTestCase("rooted colon", `\:a\b`),
		newCleanOSPathUnchangedTestCase(`\\.\ path`, `\\.\C:\a`),
		newCleanOSPathUnchangedTestCase(`\\?\ path`, `\\?\C:\a\b`),
		newCleanOSPathUnchangedTestCase(`\\?\ pipe path`, `\\?\pipe\a/.\b`),
		newCleanOSPathUnchangedTestCase(`\\?\ device named C:/a`, `\\?\C:/a\.\b`),
		newCleanOSPathUnchangedTestCase(`\??\ path`, `\??\C:\a\b`),
	}
}
