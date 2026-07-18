package appdir_test

import (
	"path/filepath"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/config/appdir"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = joinTestCase{}

// joinTestCase tests [appdir.Join] path composition. want is
// declared slash-style and normalised to the platform separator
// via filepath.FromSlash.
type joinTestCase struct {
	base string
	name string
	want string
	sub  []string
}

func (tc joinTestCase) Name() string {
	return tc.name
}

func (tc joinTestCase) Test(t *testing.T) {
	t.Helper()

	got := appdir.Join(tc.base, tc.sub...)
	core.AssertEqual(t, filepath.FromSlash(tc.want), got, "join")
}

func newJoinTestCase(name, base string, sub []string,
	want string) joinTestCase {
	return joinTestCase{
		base: base,
		name: name,
		want: want,
		sub:  sub,
	}
}

func TestJoin(t *testing.T) {
	testCases := core.S(
		newJoinTestCase("base only", "/base", nil, "/base"),
		newJoinTestCase("single sub", "/base", core.S("app"),
			"/base/app"),
		newJoinTestCase("slash sub", "/base", core.S("app/conf.d"),
			"/base/app/conf.d"),
		newJoinTestCase("multiple sub", "/base", core.S("app", "x/y"),
			"/base/app/x/y"),
		newJoinTestCase("no base", "", core.S("app"), "app"),
		newJoinTestCase("trailing slash", "/base", core.S("app/"),
			"/base/app"),
		newJoinTestCase("empty sub", "/base", core.S(""), "/base"),
	)

	core.RunTestCases(t, testCases)
}
