package fs_test

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/gobwas/glob"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var _ core.TestCase = globCompileCase{}

// globCompileCase asserts whether a single compiled pattern matches
// (or rejects) a single input path.
type globCompileCase struct {
	pattern string
	input   string
	name    string
	want    bool
}

func (tc globCompileCase) Name() string { return tc.name }

func (tc globCompileCase) Test(t *testing.T) {
	t.Helper()
	ms, err := fs.GlobCompile(tc.pattern)
	if !core.AssertNoError(t, err, "GlobCompile") {
		return
	}
	core.AssertEqual(t, tc.want, ms[0].Match(tc.input), "Match")
}

//revive:disable-next-line:flag-argument
func newGlobCompileCase(name, pattern, input string, want bool) globCompileCase {
	return globCompileCase{name: name, pattern: pattern, input: input, want: want}
}

// TestGlobCompile_AnyDepthMatchesRoot locks `**` semantics: a
// leading `**/` segment matches zero or more directories at the
// root, and embedded `/**/` matches zero or more directories
// between fixed segments. The depth-strict `*/foo` form still
// requires at least one segment ahead and is exercised so the
// any-depth wrapper doesn't quietly relax it.
func TestGlobCompile_AnyDepthMatchesRoot(t *testing.T) {
	core.RunTestCases(t, globCompileCases())
}

func globCompileCases() []globCompileCase {
	return []globCompileCase{
		// **/foo at root and at depth.
		newGlobCompileCase("**/foo at root", "**/foo", "foo", true),
		newGlobCompileCase("**/foo depth 1", "**/foo", "a/foo", true),
		newGlobCompileCase("**/foo depth 2", "**/foo", "a/b/foo", true),
		newGlobCompileCase("**/foo no match bar", "**/foo", "bar", false),
		newGlobCompileCase("**/foo no match foobar", "**/foo", "foobar", false),

		// **/foo/bar — multi-segment any-depth.
		newGlobCompileCase("**/foo/bar at root", "**/foo/bar", "foo/bar", true),
		newGlobCompileCase("**/foo/bar depth 1", "**/foo/bar", "a/foo/bar", true),
		newGlobCompileCase("**/foo/bar depth 2", "**/foo/bar", "a/b/foo/bar", true),
		newGlobCompileCase("**/foo/bar no match foo", "**/foo/bar", "foo", false),

		// **/*.log — glob basename any-depth.
		newGlobCompileCase("**/*.log at root", "**/*.log", "x.log", true),
		newGlobCompileCase("**/*.log depth 1", "**/*.log", "a/x.log", true),
		newGlobCompileCase("**/*.log depth 2", "**/*.log", "a/b/x.log", true),
		newGlobCompileCase("**/*.log no match txt", "**/*.log", "x.txt", false),

		// **/x.log — literal basename any-depth (for comparison with glob basename above).
		newGlobCompileCase("**/x.log depth 2", "**/x.log", "a/b/x.log", true),

		// Anchored variants — embedded `**` followed by glob basename.
		newGlobCompileCase("x/**/*.log depth 2", "x/**/*.log", "x/a/b/y.log", true),
		newGlobCompileCase("x/**/*.log depth 1", "x/**/*.log", "x/a/y.log", true),

		// Embedded /**/ matches zero or more directories.
		newGlobCompileCase("a/**/b zero", "a/**/b", "a/b", true),
		newGlobCompileCase("a/**/b one", "a/**/b", "a/x/b", true),
		newGlobCompileCase("a/**/b two", "a/**/b", "a/x/y/b", true),
		newGlobCompileCase("a/**/b no match a", "a/**/b", "a", false),
		newGlobCompileCase("a/**/b no match x/b", "a/**/b", "x/b", false),

		// Depth-strict */foo: at least one segment required.
		newGlobCompileCase("*/foo no match root", "*/foo", "foo", false),
		newGlobCompileCase("*/foo depth 1", "*/foo", "a/foo", true),

		// Anchored patterns are unaffected by the rewrite.
		newGlobCompileCase("anchored basename match", "foo", "foo", true),
		newGlobCompileCase("anchored basename no depth", "foo", "a/foo", false),
		newGlobCompileCase("anchored path match", "foo/bar", "foo/bar", true),
		newGlobCompileCase("anchored path no depth", "foo/bar", "a/foo/bar", false),
	}
}

// TestGobwasRaw_AnyDepth pins the raw gobwas behaviour our
// any-depth wrapper depends on: `**/X` (without brace alternation)
// must match X at any depth, including zero or more leading
// directory segments. Compose-via-anyMatcher is the workaround for
// gobwas's brace-alternation quirk that drops `**` semantics inside
// `{...}` — see compileAnyDepth in glob.go.
func TestGobwasRaw_AnyDepth(t *testing.T) {
	g, err := glob.Compile("**/*.log", '/')
	core.AssertMustNoError(t, err, "glob.Compile")
	core.AssertEqual(t, true, g.Match("a/x.log"), "raw **/*.log depth 1")
	core.AssertEqual(t, true, g.Match("a/b/x.log"), "raw **/*.log depth 2")
}

// bareFS implements only fs.FS — neither ReadDirFS nor GlobFS — used
// to confirm MatchFunc returns the capability-gap error when neither
// branch of the type switch applies.
type bareFS struct{}

func (bareFS) Open(string) (fs.File, error) { return nil, fs.ErrNotExist }

// errGlobFS implements GlobFS (but not ReadDirFS) and returns err from
// Glob, so we can verify globMatchFunc propagates the underlying error
// verbatim instead of masking it.
type errGlobFS struct {
	bareFS
	err error
}

func (e errGlobFS) Glob(string) ([]string, error) { return nil, e.err }

// listGlobFS implements GlobFS (but not ReadDirFS) returning fixed
// names from Glob, so we can exercise the GlobFS branch happy path
// independently of fstest.MapFS (which routes through ReadDirFS).
type listGlobFS struct {
	bareFS
	names []string
}

func (l listGlobFS) Glob(string) ([]string, error) { return l.names, nil }

// TestMatchFunc_UnsupportedFS pins the capability-gap contract: when
// the underlying fs.FS implements neither ReadDirFS nor GlobFS,
// MatchFunc returns *fs.PathError{Op:"match", Path:<root>, Err:
// ErrUnsupported}, detectable via errors.Is.
func TestMatchFunc_UnsupportedFS(t *testing.T) {
	_, err := fs.MatchFunc(bareFS{}, ".", nil)
	core.AssertErrorIs(t, err, fs.ErrUnsupported, "errors.Is ErrUnsupported")
	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "*fs.PathError")
	core.AssertEqual(t, "match", pe.Op, "Op")
	core.AssertEqual(t, ".", pe.Path, "Path")
}

// TestMatchFunc_GlobErrorPropagates pins the propagation contract:
// when GlobFS.Glob returns an error, MatchFunc surfaces it verbatim
// rather than masking it with ErrUnsupported or any other sentinel.
func TestMatchFunc_GlobErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	_, err := fs.MatchFunc(errGlobFS{err: sentinel}, ".", nil)
	core.AssertSame(t, sentinel, err, "propagated error")
}

// TestMatch_ReadDirFS exercises the ReadDirFS branch end-to-end:
// compiled patterns filter MapFS entries and the root-matching
// rewrite applies through Match (not just GlobCompile).
func TestMatch_ReadDirFS(t *testing.T) {
	fsys := fstest.MapFS{
		"foo.go":     {},
		"a/foo.go":   {},
		"a/b/foo.go": {},
		"a/bar.txt":  {},
	}
	ms, err := fs.GlobCompile("**/*.go")
	core.AssertMustNoError(t, err, "GlobCompile")
	out, err := fs.Match(fsys, ".", ms...)
	core.AssertMustNoError(t, err, "Match")
	core.AssertSliceEqual(t,
		core.S("a/b/foo.go", "a/foo.go", "foo.go"),
		out, "matches")
}

// TestMatch_GlobFS exercises the GlobFS branch happy path: a
// GlobFS-only filesystem feeds names into globMatchFunc, which
// filters via the compiled patterns and returns sorted results.
func TestMatch_GlobFS(t *testing.T) {
	fsys := listGlobFS{names: []string{"foo.go", "a/foo.go", "a/bar.txt"}}
	ms, err := fs.GlobCompile("**/*.go")
	core.AssertMustNoError(t, err, "GlobCompile")
	out, err := fs.Match(fsys, ".", ms...)
	core.AssertMustNoError(t, err, "Match")
	core.AssertSliceEqual(t,
		core.S("a/foo.go", "foo.go"),
		out, "matches")
}

// TestGlob exercises the Glob convenience entry point that compiles
// patterns and walks from "." in one call, covering the whole pipeline
// from rewrite through MatchFunc against a ReadDirFS.
func TestGlob(t *testing.T) {
	fsys := fstest.MapFS{
		"foo.go":     {},
		"a/foo.go":   {},
		"a/b/foo.go": {},
		"a/bar.txt":  {},
	}
	out, err := fs.Glob(fsys, "**/*.go")
	core.AssertMustNoError(t, err, "Glob")
	core.AssertSliceEqual(t,
		core.S("a/b/foo.go", "a/foo.go", "foo.go"),
		out, "matches")
}

// TestMatchFunc_InvalidRoot pins the malformed-root contract: when
// Clean rejects the root, MatchFunc returns *fs.PathError{Op:
// "readdir", Path:<root>, Err:ErrInvalid} — distinct from the
// capability-gap ErrUnsupported case.
func TestMatchFunc_InvalidRoot(t *testing.T) {
	_, err := fs.MatchFunc(fstest.MapFS{}, "/../bad", nil)
	core.AssertErrorIs(t, err, fs.ErrInvalid, "errors.Is ErrInvalid")
	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "*fs.PathError")
	core.AssertEqual(t, "readdir", pe.Op, "Op")
	core.AssertEqual(t, "/../bad", pe.Path, "Path")
}
