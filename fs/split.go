package fs

import (
	"strings"
)

// Split splits a path into two [Clean] [fs.ValidPath] components
// that make dir + "/" + file equivalent to the given path.
func Split(path string) (dir, file string) {
	s, _ := Clean(path)
	return unsafeSplit(s)
}

func unsafeSplit(path string) (dir, file string) {
	i := strings.LastIndexByte(path, '/')
	switch {
	case i < 0:
		// "foo" -> ".", "foo"
		return ".", path
	case i == 0:
		// "/foo" -> "", "foo"
		return "", path[1:]
	default:
		// "foo/bar" -> "foo", "bar"
		return path[:i], path[i+1:]
	}
}
