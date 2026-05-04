package fs

import (
	"io/fs"
	"strings"

	"github.com/gobwas/glob"

	"darvaza.org/core"
)

// Matcher is a compiled globbing pattern from https://github.com/gobwas/glob
type Matcher = glob.Glob

// GlobCompile compiles a list of file globbing patterns using
// https://github.com/gobwas/glob with one adjustment: a leading
// `**/` matches at any depth including the root, so `**/foo`
// matches `foo`, `a/foo` and `a/b/foo`. Embedded `/**/`
// (e.g. `a/**/b`) matches zero or more directory segments.
// The depth-strict `*/foo` form is unaffected — it still
// requires at least one segment ahead. The bare `**/` form
// (no tail) is left to gobwas.
func GlobCompile(patterns ...string) ([]Matcher, error) {
	out := make([]Matcher, 0, len(patterns))

	for _, pat := range patterns {
		g, err := compileAnyDepth(pat)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}

	return out, nil
}

// compileAnyDepth compiles a pattern, with leading `**/X` semantics
// implemented as the OR of the root form `X` and the original
// `**/X`. Composing two compiled matchers sidesteps a gobwas brace-
// alternation quirk that drops `**` semantics under `{...}`.
// Patterns that don't start with `**/`, or whose tail is empty,
// compile straight through.
func compileAnyDepth(pat string) (Matcher, error) {
	rest, ok := strings.CutPrefix(pat, "**/")
	if !ok || rest == "" {
		return glob.Compile(pat, '/')
	}
	root, err := glob.Compile(rest, '/')
	if err != nil {
		return nil, err
	}
	deep, err := glob.Compile(pat, '/')
	if err != nil {
		return nil, err
	}
	return anyMatcher{root, deep}, nil
}

// anyMatcher composes two or more compiled patterns into a single
// Matcher that returns true if any constituent matches.
type anyMatcher []glob.Glob

// Match returns true if any of the wrapped matchers accept s.
func (am anyMatcher) Match(s string) bool {
	for _, g := range am {
		if g.Match(s) {
			return true
		}
	}
	return false
}

// Glob returns all entries matching any of the given patterns.
func Glob(fSys fs.FS, patterns ...string) ([]string, error) {
	g, err := GlobCompile(patterns...)
	if err != nil {
		return nil, err
	}

	return Match(fSys, ".", g...)
}

// Match returns all entries matching any of the given compiled glob patterns.
func Match(fSys fs.FS, root string, globs ...Matcher) ([]string, error) {
	return MatchFunc(fSys, root, newCheckerMatchAny(globs))
}

func newCheckerMatchAny(globs []Matcher) func(string, fs.DirEntry) bool {
	if len(globs) == 0 {
		return nil
	}

	return func(path string, _ fs.DirEntry) bool {
		for _, g := range globs {
			if g.Match(path) {
				return true
			}
		}
		return false
	}
}

// MatchFunc returns all entries satisfying the given checker function.
// If no function is provided, all entries will be listed.
// Entries giving Stat error will be ignored.
func MatchFunc(fSys fs.FS, root string, check func(string, fs.DirEntry) bool) ([]string, error) {
	dir, ok := Clean(root)
	if !ok {
		err := &fs.PathError{
			Op:   "readdir",
			Path: root,
			Err:  fs.ErrInvalid,
		}
		return nil, err
	}

	if check == nil {
		check = func(string, fs.DirEntry) bool { return true }
	}

	switch x := fSys.(type) {
	case ReadDirFS:
		return walkMatchFunc(x, dir, check)
	case GlobFS:
		return globMatchFunc(x, dir, check)
	default:
		return nil, &fs.PathError{Op: "match", Path: dir, Err: ErrUnsupported}
	}
}

func walkMatchFunc(fSys ReadDirFS, dir string, check func(string, fs.DirEntry) bool) ([]string, error) {
	var out []string
	err := fs.WalkDir(fSys, dir, func(path string, di fs.DirEntry, err error) error {
		switch {
		case err != nil:
			// only pass root errors
			return core.IIf(dir == path, err, nil)
		case check(path, di):
			// match
			out = append(out, path)
		default:
		}
		return nil
	})

	return out, err
}

func globMatchFunc(fSys fs.GlobFS, root string, check func(string, fs.DirEntry) bool) ([]string, error) {
	ss, err := fSys.Glob("**")
	switch {
	case err != nil:
		return nil, err
	case len(ss) == 0:
		return ss, nil
	default:
		m := make(map[string]struct{})
		for _, s := range ss {
			if s, ok := globMatchFuncOne(fSys, root, s, check); ok {
				m[s] = struct{}{}
			}
		}
		return core.SortedKeys(m), nil
	}
}

func globMatchFuncOne(fSys fs.FS, root, fullName string, check func(string, fs.DirEntry) bool) (string, bool) {
	name, ok := unsafeCutRoot(fullName, root)
	if ok {
		fi, _ := fs.Stat(fSys, fullName)
		if check(name, fs.FileInfoToDirEntry(fi)) {
			return name, true
		}
	}

	return "", false
}
