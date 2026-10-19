package fs

import (
	"slices"
	"strings"
)

// Clean reduces a path and tells if it's a valid root
// for fs.FS operations.
func Clean(path string) (string, bool) {
	switch path {
	case "", ".":
		return ".", true
	case "/":
		return "/", false
	}
	if path[0] == '/' {
		return cleanRooted(path[1:])
	}
	return cleanRelative(path)
}

// cleanRooted classifies a rooted path (one that originally began
// with '/'), with the leading slash stripped by the caller.
func cleanRooted(rest string) (string, bool) {
	switch s := doClean(rest); s {
	case "", ".":
		return "/", false
	default:
		return "/" + s, false
	}
}

// cleanRelative classifies a relative path (one that did not begin
// with '/').
func cleanRelative(path string) (string, bool) {
	switch s := doClean(path); s {
	case "", ".":
		return ".", true
	case "..":
		return s, false
	default:
		return s, !strings.HasPrefix(s, "../")
	}
}

func doClean(path string) string {
	s := doCleanRunes([]rune(path))

	switch {
	case len(s) == 0:
		// root
		return "."
	case len(s) == len(path):
		// same
		return path
	default:
		// cleaned
		return string(s)
	}
}

func doCleanRunes(path []rune) []rune {
	buf, rest := make([]rune, 0, len(path)), path

	for len(rest) > 0 {
		var part []rune

		i := slices.Index(rest, '/')
		switch {
		case i < 0:
			// last
			part, rest = rest, nil
			buf = cleanRunesApply(buf, part)
		default:
			// next
			part, rest = rest[:i], rest[i+1:]
			buf = cleanRunesApply(buf, part)
		}
	}

	return buf
}

func cleanRunesApply(buf, next []rune) []rune {
	var dotdot, dot = []rune(".."), []rune{'.'}

	switch {
	case len(next) == 0, slices.Equal(next, dot):
		// ignore "" and "."
		return buf
	case slices.Equal(next, dotdot):
		// dot dot
		return cleanRunesDotDot(buf)
	default:
		// anything else, append
		return joinRunes(buf, next)
	}
}

func cleanRunesDotDot(buf []rune) []rune {
	var dotdot = []rune("..")

	if len(buf) == 0 || slices.Equal(buf, dotdot) {
		// ".." or "../.."
		return joinRunes(buf, dotdot)
	}

	i := lastIndexRune(buf, '/')
	switch {
	case i < 0:
		// "foo/.." → ""
		return buf[:0]
	case slices.Equal(buf[i+1:], dotdot):
		// "../../.."
		return joinRunes(buf, dotdot)
	default:
		// "a/b/.."
		return buf[:i]
	}
}

// lastIndexRune returns the index of the last r in s, or -1 when s
// holds none.
func lastIndexRune(s []rune, r rune) int {
	for i, c := range slices.Backward(s) {
		if c == r {
			return i
		}
	}
	return -1
}
