package fs

import "strings"

// joinRunes appends a clean path to another in a []rune buffer.
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

func unsafeCutRoot(name, root string) (string, bool) {
	switch {
	case root == name:
		return ".", true
	case root == ".":
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
