package fs

import "strings"

// joinRunes appends a clean path component to another in a []rune buffer,
// automatically inserting a path separator ('/') when both components are
// non-empty. If either component is empty, it returns the other component
// without modification.
//
// This function is used internally for efficient path construction during
// path cleaning operations.
func joinRunes(before, after []rune) []rune {
	switch {
	case len(after) == 0:
		return before
	case len(before) == 0:
		return append(before, after...)
	default:
		before = append(before, '/')
		return append(before, after...)
	}
}

// unsafeCutRoot attempts to remove a root path prefix from a given path name.
// It returns the remaining path and a boolean indicating success.
//
// When the root is successfully removed, the function returns the remaining
// path portion. If no path remains after removal, it returns "." representing
// the current directory.
//
// Special cases:
//   - When name equals root exactly: returns "."
//   - When root is ".": returns the original name (unless name is also ".")
//   - When name has root as a directory prefix: returns the remaining path
//   - When removal would leave an empty path: returns "."
//
// The function treats root as a directory prefix, so trailing slashes in the
// input affect matching behaviour. Paths are expected to be normalised by
// the caller.
//
// The function is marked "unsafe" because it assumes well-formed path inputs
// and does not perform validation or normalisation. Callers must ensure that
// both name and root are clean, absolute or relative paths as appropriate.
//
// Used internally for glob pattern matching and path manipulation operations.
func unsafeCutRoot(name, root string) (string, bool) {
	switch root {
	case name:
		return ".", true
	case ".":
		return name, name != "."
	default:
		s, ok := strings.CutPrefix(name, root+"/")
		switch {
		case !ok:
			return "", false
		case s == "":
			return ".", true
		default:
			return s, true
		}
	}
}
