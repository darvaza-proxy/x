//go:build windows

package appdir_test

// osNewPrefixTestCases supplies the platform-specific well-known
// prefixes exercised by [appdir.NewPrefix]. Windows recognises no
// FHS-style prefixes beyond the portable [appdir.PrefixUser] and
// [appdir.PrefixSystem], so it contributes none.
func osNewPrefixTestCases() []newPrefixTestCase {
	return nil
}
