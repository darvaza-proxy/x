package assets

import (
	"io"
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

// unsafeClose does a Close() discarding errors
func unsafeClose(f io.Closer) {
	_ = f.Close()
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

func tryContentType(v any) (string, bool) {
	var s string
	if f, ok := v.(ContentTyped); ok {
		s = f.ContentType()
	}

	return s, s != ""
}

func tryContentTypeSetter(v any) (ContentTypeSetter, bool) {
	if f, ok := v.(ContentTypeSetter); ok {
		return f, f != nil
	}
	return nil, false
}

func tryETags(v any) ([]string, bool) {
	var tags []string

	if f, ok := v.(ETaged); ok {
		tags = f.ETags()
	}

	return tags, len(tags) > 0
}

func tryETagsSetter(v any) (ETagsSetter, bool) {
	if f, ok := v.(ETagsSetter); ok {
		return f, f != nil
	}
	return nil, false
}

func tryReadSeeker(v any) (fs.ReadSeeker, bool) {
	f, ok := v.(io.ReadSeeker)
	return f, ok
}

func tryStat(v any) (fs.FileInfo, bool) {
	if f, ok := v.(interface {
		Stat() (fs.FileInfo, error)
	}); ok {
		fi, err := f.Stat()
		return fi, fi != nil && err == nil
	}

	if f, ok := v.(interface {
		Info() (fs.FileInfo, error)
	}); ok {
		fi, err := f.Info()
		return fi, fi != nil && err == nil
	}

	return nil, false
}
