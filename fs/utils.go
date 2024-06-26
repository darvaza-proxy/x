package fs

func joinRunes(before, after []rune) []rune {
	switch {
	case len(after) == 0:
		return before
	case len(before) == 0:
		return append(before, after...)
	default:
		s := append(before, '/')
		return append(s, after...)
	}
}
