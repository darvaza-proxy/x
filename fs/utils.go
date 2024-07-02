package fs

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
