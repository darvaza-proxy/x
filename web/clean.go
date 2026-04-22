package web

import (
	"strings"

	"darvaza.org/x/fs"
)

// Clean normalises a URL path. It is aimed primarily at paths
// destined for an outbound HTTP Location header, but is usable
// wherever a URL path needs to be reduced to a canonical form.
//
// Backslashes are replaced with forward slashes before reduction
// (matching WHATWG URL-path normalisation); the path is then
// reduced with fs.Clean, and any leading rooted /.. escape
// components are stripped so the cleaned result can't escape
// above the root. A trailing slash on the input is preserved.
//
// The input must be a path, not a full URL: Clean does not
// parse scheme or authority, so "http://evil.com/x" reduces as
// a path and becomes "http:/evil.com/x". Literal /.. is
// stripped; URL-encoded variants (for example %2e%2e) are not
// decoded and pass through verbatim.
//
// The second return is false when leading rooted /.. blocks
// had to be discarded — a signal that the input attempted to
// escape above the root. Backslash normalisation and fs.Clean's
// own reduction don't flip it.
func Clean(path string) (string, bool) {
	path = strings.ReplaceAll(path, `\`, "/")
	// Trailing-slash check runs on the post-ReplaceAll path, so
	// input ending in `\` is treated as trailing-slash input.
	trailing := len(path) > 1 && strings.HasSuffix(path, "/")

	cleaned, _ := fs.Clean(path)

	ok := true
	for strings.HasPrefix(cleaned, "/../") {
		cleaned = cleaned[3:]
		ok = false
	}
	// The loop can't remove a terminal /.. on its own: the
	// "/../" prefix match requires a following slash. Handle
	// the bare "/.." case here.
	if cleaned == "/.." {
		cleaned = "/"
		ok = false
	}

	if trailing && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned, ok
}
