package assets

import (
	"strings"

	"darvaza.org/core"
	"darvaza.org/x/fs"
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

// Match tests if the given path matches any of the given globs.
// if no matchers are provided, it will be understood as unconditional.
func Match(path string, globs []fs.Matcher) bool {
	var match bool

	for _, g := range globs {
		if g.Match(path) {
			match = true
			break
		}
	}

	return match || len(globs) == 0
}
