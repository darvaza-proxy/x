package assets

import (
	"io"

	"darvaza.org/x/fs"
)

var (
	_ io.ReadSeeker = (*closedReader)(nil)
)

// closedReader always returns [fs.ErrClosed]
type closedReader struct{}

func (closedReader) Read([]byte) (int, error)       { return 0, fs.ErrClosed }
func (closedReader) Seek(int64, int) (int64, error) { return 0, fs.ErrClosed }

func newClosedReader() io.ReadSeeker {
	return &closedReader{}
}

// isGlobMatch tests if the given path matches any of the given globs.
// if no matchers are provided, it will be understood as unconditional.
func isGlobMatch(path string, globs []fs.Matcher) bool {
	var match bool

	for _, g := range globs {
		if g.Match(path) {
			match = true
			break
		}
	}

	return match || len(globs) == 0
}

// unsafeJoin joins two clean paths.
func unsafeJoin(base, dir string) string {
	switch {
	case dir == ".":
		return base
	case base == ".":
		return dir
	default:
		return base + "/" + dir
	}
}
