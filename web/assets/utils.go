package assets

import (
	"strings"

	"darvaza.org/core"
)

// CleanETags removes empty and duplicate tags.
// Whitespace before and after the content of each tag is also removed.
func CleanETags(tags ...string) []string {
	return core.SliceReplaceFn(tags, doCleanETags)
}

func doCleanETags(s []string, tag string) (string, bool) {
	tag = strings.TrimSpace(tag)
	switch {
	case tag == "":
		// empty
		return "", false
	case core.SliceContains(s, tag):
		// duplicate
		return "", false
	default:
		// new
		return tag, true
	}
}
