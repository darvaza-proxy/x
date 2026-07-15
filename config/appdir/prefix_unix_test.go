//go:build !windows

package appdir_test

import (
	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// osNewPrefixTestCases supplies the platform-specific well-known
// prefixes exercised by [appdir.NewPrefix]. On unix the FHS
// prefixes [appdir.PrefixLocal] and [appdir.PrefixOptional]
// extend the portable [appdir.PrefixUser] and
// [appdir.PrefixSystem] set.
func osNewPrefixTestCases() []newPrefixTestCase {
	return core.S(
		newNewPrefixTestCase("local prefix",
			string(appdir.PrefixLocal), appdir.PrefixLocal),
		newNewPrefixTestCase("optional prefix",
			string(appdir.PrefixOptional), appdir.PrefixOptional),
	)
}
