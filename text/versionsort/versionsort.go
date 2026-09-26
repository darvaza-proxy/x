package versionsort

import (
	"cmp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"darvaza.org/core"
)

// Compare orders two strings by version, reporting -1 when a comes
// first, +1 when b does, and 0 only for the same string. The empty
// string comes ahead of any other, a leading tilde included. Strings
// that tie on every rule, such as "ttyS1" and "ttyS01", or "tty" and
// "tty0", fall back to their bytes.
func Compare[S core.String](a, b S) int {
	sa, sb := string(a), string(b)
	switch {
	case sa == sb:
		return 0
	case sa == "" || sb == "":
		return cmp.Compare(len(sa), len(sb))
	}
	if d := compareVersion(sa, sb); d != 0 {
		return d
	}
	return strings.Compare(sa, sb)
}

// Sort sorts a slice of strings by version, in place.
func Sort[S core.String](a []S) {
	slices.SortFunc(a, Compare[S])
}

// SortBy sorts a slice by the version string fn gives each element, in
// place. The sort is stable, so elements whose strings are equal keep
// their order. fn is called once per element. A nil fn leaves the slice
// as it is.
func SortBy[S core.String, T any](a []T, fn func(T) S) {
	if fn == nil {
		return
	}

	entries := make([]keyed[S, T], len(a))
	for i, v := range a {
		entries[i] = keyed[S, T]{key: fn(v), value: v}
	}

	slices.SortStableFunc(entries, func(x, y keyed[S, T]) int {
		return Compare(x.key, y.key)
	})

	for i, e := range entries {
		a[i] = e.value
	}
}

// keyed pairs an element with the version string SortBy sorts it by.
type keyed[S core.String, T any] struct {
	key   S
	value T
}

// compareVersion compares two strings a run at a time, the text before
// each number and then the number, and reports 0 when every run
// matches.
func compareVersion(a, b string) int {
	for a != "" || b != "" {
		var d int
		if a, b, d = compareText(a, b); d != 0 {
			return d
		}
		if a, b, d = compareNumber(a, b); d != 0 {
			return d
		}
	}
	return 0
}

// compareText compares the text up to the next number of each string a
// character at a time, and returns what follows it.
func compareText(a, b string) (restA, restB string, d int) {
	for startsText(a) || startsText(b) {
		if n := sharedText(a, b); n > 0 {
			a, b = a[n:], b[n:]
			continue
		}

		oa, na := textOrder(a)
		ob, nb := textOrder(b)
		if d = cmp.Compare(oa, ob); d != 0 {
			return a, b, d
		}
		// equal orders are the same character, or the same invalid
		// byte, so neither is empty
		a, b = a[na:], b[nb:]
	}
	return a, b, 0
}

// sharedText reports how many leading bytes a and b have in common that
// are ASCII and not digits. Each is a whole character, ranked the same
// on both sides, so compareText can step over them without ranking them.
func sharedText(a, b string) int {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] &&
		a[i] < utf8.RuneSelf && !isDigit(a[i]) {
		i++
	}
	return i
}

// compareNumber compares the leading run of digits of each string by
// the number it writes, however long either is, and returns what
// follows it. A string with no run there counts as zero.
func compareNumber(a, b string) (restA, restB string, d int) {
	na, restA := cutDigits(a)
	nb, restB := cutDigits(b)
	na = strings.TrimLeft(na, "0")
	nb = strings.TrimLeft(nb, "0")
	if d = cmp.Compare(len(na), len(nb)); d == 0 {
		d = strings.Compare(na, nb)
	}
	return restA, restB, d
}

// cutDigits splits a string into its leading run of decimal digits,
// which may be empty, and the rest.
func cutDigits(s string) (digits, rest string) {
	i := 0
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	return s[:i], s[i:]
}

// startsText reports whether s opens with anything but a digit.
func startsText(s string) bool {
	return s != "" && !isDigit(s[0])
}

// otherRank and invalidRank offset the classes ranked after the letters,
// so every other character follows every letter, and every byte that is
// not valid UTF-8 follows every character.
const (
	otherRank   = unicode.MaxRune + 1
	invalidRank = 2 * otherRank
)

// textOrder ranks the first character of s for compareText, and
// reports its size in bytes. A tilde ranks first, ahead of the end of
// the string, then a digit, then the letters and then every other
// character, each by code point; compareText never weighs the end
// against a digit. A byte that is not valid UTF-8 ranks last, by value.
func textOrder(s string) (order, size int) {
	if s == "" {
		return -1, 0
	}

	r, size := utf8.DecodeRuneInString(s)
	switch {
	case r == '~':
		return -2, size
	case isDigit(s[0]):
		return 0, size
	case r == utf8.RuneError && size == 1:
		return invalidRank + int(s[0]), size
	case unicode.IsLetter(r):
		return int(r), size
	default:
		return otherRank + int(r), size
	}
}

// isDigit reports whether c is an ASCII decimal digit.
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
