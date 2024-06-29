package fs

import (
	"io/fs"

	"github.com/gobwas/glob"

	"darvaza.org/core"
)

// Glob is a compiled globbing pattern from https://github.com/gobwas/glob
type Glob = glob.Glob

// GlobCompile compiles a list of file globbing patterns using
// https://github.com/gobwas/glob
func GlobCompile(patterns ...string) ([]Glob, error) {
	out := make([]Glob, 0, len(patterns))

	for _, pat := range patterns {
		g, err := glob.Compile(pat, '/')
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}

	return out, nil
}

// GlobFS returns all entries matching any of the given patterns.
func GlobFS(fSys fs.FS, patterns ...string) ([]string, error) {
	g, err := GlobCompile(patterns...)
	if err != nil {
		return nil, err
	}

	return MatchFS(fSys, ".", g...)
}

// MatchFS returns all entries matching any of the given compiled glob patterns.
func MatchFS(fSys fs.FS, root string, globs ...Glob) ([]string, error) {
	return MatchFuncFS(fSys, root, newCheckerMatchAnyFS(globs))
}

func newCheckerMatchAnyFS(globs []Glob) func(string, fs.DirEntry) bool {
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

// MatchFuncFS returns all entries satisfying the given checker function.
// If no function is provided, all entries will be listed.
// Entries giving Stat error will be ignored.
func MatchFuncFS(fSys fs.FS, root string, check func(string, fs.DirEntry) bool) ([]string, error) {
	var out []string

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

	err := fs.WalkDir(fSys, dir, func(path string, di fs.DirEntry, err error) error {
		switch {
		case err != nil:
			// only pass root errors
			return core.IIf(dir == path, err, nil)
		case check(path, di):
			// match
			out = append(out, path)
		}
		return nil
	})

	return out, err
}
