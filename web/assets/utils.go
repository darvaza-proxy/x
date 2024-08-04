package assets

import (
	"io"
	"net/http"
	"strings"
	"time"

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

// headerExist checks if a canonical header has been specified
func headerExist(hdr http.Header, name string) bool {
	if v, ok := hdr[name]; ok {
		if len(v) > 0 && v[0] != "" {
			return true
		}
	}
	return false
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
	switch f := v.(type) {
	case io.ReadSeeker:
		return f, true
	case Asset:
		return f.Content(), true
	default:
		return nil, false
	}
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

func tryModTime(v any) (time.Time, bool) {
	var modTime time.Time
	if f, ok := v.(interface {
		ModTime() time.Time
	}); ok {
		modTime = f.ModTime()
	}

	return modTime, !modTime.IsZero()
}

func getModTime(v any) (time.Time, bool) {
	if t, ok := tryModTime(v); ok {
		return t, true
	}

	if fi, ok := tryStat(v); ok {
		return tryModTime(fi)
	}

	return time.Time{}, false
}
