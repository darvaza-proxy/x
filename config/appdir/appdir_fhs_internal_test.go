//go:build !windows

package appdir

import (
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = getSysPrefixTestCase{}

// getSysPrefixTestCase tests Prefix.getSysPrefix reporting the
// string to prepend to system directories, and false under
// PrefixUser, which has none.
type getSysPrefixTestCase struct {
	p      Prefix
	name   string
	want   string
	wantOK bool
}

func (tc getSysPrefixTestCase) Name() string {
	return tc.name
}

func (tc getSysPrefixTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := tc.p.getSysPrefix()
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	core.AssertEqual(t, tc.want, got, "prefix")
}

func newGetSysPrefixTestCase(name string, p Prefix, want string,
	wantOK bool) getSysPrefixTestCase {
	return getSysPrefixTestCase{
		p:      p,
		name:   name,
		want:   want,
		wantOK: wantOK,
	}
}

func TestGetSysPrefix(t *testing.T) {
	testCases := []getSysPrefixTestCase{
		newGetSysPrefixTestCase("system", PrefixSystem, "", true),
		newGetSysPrefixTestCase("user", PrefixUser, "", false),
		newGetSysPrefixTestCase("zero value", "", "", false),
		newGetSysPrefixTestCase("local", PrefixLocal,
			"/usr/local", true),
		newGetSysPrefixTestCase("custom", "/srv/pods",
			"/srv/pods", true),
	}

	core.RunTestCases(t, testCases)
}
