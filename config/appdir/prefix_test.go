package appdir_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = prefixUserModeTestCase{}

// prefixUserModeTestCase pins the user-mode invariant: under
// [appdir.PrefixUser] every FooDir method returns the same as
// its UserFooDir counterpart, on every platform.
type prefixUserModeTestCase struct {
	name string
	kind Kind
}

func (tc prefixUserModeTestCase) Name() string {
	return tc.name
}

func (tc prefixUserModeTestCase) Test(t *testing.T) {
	t.Helper()

	want, err := callUserDirFunc(t, tc.kind, "app")
	core.AssertMustNoError(t, err, "user %s dir", tc.kind)

	got, err := callPrefixDirFunc(t, appdir.PrefixUser, tc.kind, "app")
	core.AssertMustNoError(t, err, "%s dir", tc.kind)
	core.AssertEqual(t, want, got, "dir")
}

func newPrefixUserModeTestCase(name string,
	kind Kind) prefixUserModeTestCase {
	return prefixUserModeTestCase{
		kind: kind,
		name: name,
	}
}

func TestPrefixUserMode(t *testing.T) {
	testCases := core.S(
		newPrefixUserModeTestCase("cache", KindCache),
		newPrefixUserModeTestCase("config", KindConfig),
		newPrefixUserModeTestCase("data", KindData),
		newPrefixUserModeTestCase("runtime", KindRuntime),
	)

	core.RunTestCases(t, testCases)
}
