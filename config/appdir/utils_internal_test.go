package appdir

import (
	"errors"
	"os"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = joinFnTestCase{}
var _ core.TestCase = getEnvDirTestCase{}
var _ core.TestCase = getEnvHomeDirTestCase{}
var _ core.TestCase = partsFromSlashTestCase{}
var _ core.TestCase = splitFromSlashTestCase{}

// errStub marks the error path of a stubbed base directory accessor.
var errStub = errors.New("stub")

func stubBaseDir() (string, error) {
	return "/base", nil
}

func stubBaseDirErr() (string, error) {
	return "", errStub
}

// joinFnTestCase tests joinFn composing a directory from a base
// directory accessor, propagating its error.
type joinFnTestCase struct {
	wantErr error
	fn      func() (string, error)
	name    string
	want    string
	sub     []string
}

func (tc joinFnTestCase) Name() string {
	return tc.name
}

func (tc joinFnTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := joinFn(tc.fn, tc.sub...)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, "join fn")
		core.AssertEqual(t, "", got, "dir")
		return
	}

	core.AssertNoError(t, err, "join fn")
	core.AssertEqual(t, tc.want, got, "dir")
}

// newJoinFnTestCase declares a row expected to succeed.
func newJoinFnTestCase(name string, fn func() (string, error),
	sub []string, want string) joinFnTestCase {
	return joinFnTestCase{
		fn:   fn,
		name: name,
		want: want,
		sub:  sub,
	}
}

// newJoinFnTestCaseErr declares a row where fn is expected to fail.
func newJoinFnTestCaseErr(name string, fn func() (string, error),
	wantErr error) joinFnTestCase {
	return joinFnTestCase{
		wantErr: wantErr,
		fn:      fn,
		name:    name,
	}
}

func TestJoinFn(t *testing.T) {
	testCases := []joinFnTestCase{
		newJoinFnTestCase("no sub", stubBaseDir, nil, "/base"),
		newJoinFnTestCase("single sub", stubBaseDir,
			core.S("app"), "/base/app"),
		newJoinFnTestCaseErr("propagates error", stubBaseDirErr,
			errStub),
	}

	core.RunTestCases(t, testCases)
}

// getEnvDirTestCase tests getEnvDir honouring absolute values of
// the environment variable and treating anything else as unset.
// Absolute paths are platform-specific, so the rows are
// platform-gated.
type getEnvDirTestCase struct {
	envValue string
	name     string
	want     string
	unset    bool
}

func (tc getEnvDirTestCase) Name() string {
	return tc.name
}

func (tc getEnvDirTestCase) Test(t *testing.T) {
	t.Helper()
	// t.Setenv registers the restore that os.Unsetenv needs to
	// leave behind when making the variable genuinely absent.
	t.Setenv("APPDIR_TEST_DIR", tc.envValue)
	if tc.unset {
		err := os.Unsetenv("APPDIR_TEST_DIR")
		core.AssertMustNoError(t, err, "unsetenv")
	}

	got := getEnvDir("APPDIR_TEST_DIR")
	core.AssertEqual(t, tc.want, got, "dir")
}

func newGetEnvDirTestCase(name, envValue,
	want string) getEnvDirTestCase {
	return getEnvDirTestCase{
		envValue: envValue,
		name:     name,
		want:     want,
	}
}

// newGetEnvDirTestCaseUnset declares a row where the variable is
// genuinely absent from the environment, not merely empty.
func newGetEnvDirTestCaseUnset(name string) getEnvDirTestCase {
	return getEnvDirTestCase{
		name:  name,
		want:  "",
		unset: true,
	}
}

// getEnvHomeDirTestCase tests getEnvHomeDir honouring absolute
// values of the environment variable and falling back to a
// directory under the user's home otherwise. setTestHome and the
// rows are platform-specific and platform-gated.
type getEnvHomeDirTestCase struct {
	envValue string
	name     string
	fallback string
	want     string
}

func (tc getEnvHomeDirTestCase) Name() string {
	return tc.name
}

func (tc getEnvHomeDirTestCase) Test(t *testing.T) {
	t.Helper()
	t.Setenv("APPDIR_TEST_DIR", tc.envValue)
	setTestHome(t)

	got, err := getEnvHomeDir("APPDIR_TEST_DIR", tc.fallback)
	core.AssertMustNoError(t, err, "env home dir")
	core.AssertEqual(t, tc.want, got, "dir")
}

func newGetEnvHomeDirTestCase(name, envValue, fallback,
	want string) getEnvHomeDirTestCase {
	return getEnvHomeDirTestCase{
		envValue: envValue,
		name:     name,
		fallback: fallback,
		want:     want,
	}
}

// partsFromSlashTestCase tests partsFromSlash flattening a base and
// slash-delimited sub-paths into single-level parts.
type partsFromSlashTestCase struct {
	base string
	name string
	want []string
	sub  []string
}

func (tc partsFromSlashTestCase) Name() string {
	return tc.name
}

func (tc partsFromSlashTestCase) Test(t *testing.T) {
	t.Helper()

	got := partsFromSlash(tc.base, tc.sub...)
	core.AssertSliceEqual(t, tc.want, got, "parts")
}

func newPartsFromSlashTestCase(name, base string, sub,
	want []string) partsFromSlashTestCase {
	return partsFromSlashTestCase{
		base: base,
		name: name,
		want: want,
		sub:  sub,
	}
}

func TestPartsFromSlash(t *testing.T) {
	testCases := []partsFromSlashTestCase{
		newPartsFromSlashTestCase("base only", "/base", nil,
			core.S("/base")),
		newPartsFromSlashTestCase("base and sub", "/base",
			core.S("app", "x/y"), core.S("/base", "app", "x", "y")),
		newPartsFromSlashTestCase("no base", "",
			core.S("app"), core.S("app")),
		newPartsFromSlashTestCase("all empty", "",
			core.S(""), core.S[string]()),
	}

	core.RunTestCases(t, testCases)
}

// splitFromSlashTestCase tests splitFromSlash discarding empty
// slash-delimited segments.
type splitFromSlashTestCase struct {
	input string
	name  string
	want  []string
}

func (tc splitFromSlashTestCase) Name() string {
	return tc.name
}

func (tc splitFromSlashTestCase) Test(t *testing.T) {
	t.Helper()

	got := splitFromSlash(tc.input)
	core.AssertSliceEqual(t, tc.want, got, "parts")
}

func newSplitFromSlashTestCase(name, input string,
	want []string) splitFromSlashTestCase {
	return splitFromSlashTestCase{
		input: input,
		name:  name,
		want:  want,
	}
}

func TestSplitFromSlash(t *testing.T) {
	testCases := []splitFromSlashTestCase{
		newSplitFromSlashTestCase("single", "a", core.S("a")),
		newSplitFromSlashTestCase("nested", "a/b", core.S("a", "b")),
		newSplitFromSlashTestCase("surrounding slashes", "/a/",
			core.S("a")),
		newSplitFromSlashTestCase("double slash", "a//b",
			core.S("a", "b")),
		newSplitFromSlashTestCase("empty", "", core.S[string]()),
	}

	core.RunTestCases(t, testCases)
}
