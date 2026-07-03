package appdir

import (
	"path/filepath"
	"strings"
)

// Join combines file path parts in an OS-specific way. Parts
// are allowed to be multipart themselves, using `/` as
// delimiter.
func Join(base string, sub ...string) string {
	a := partsFromSlash(base, sub...)
	return filepath.Join(a...)
}

// joinFn joins sub-paths onto a base directory obtained
// from fn, propagating fn's error.
func joinFn(fn func() (string, error),
	sub ...string) (string, error) {
	dir, err := fn()
	if err != nil {
		return "", err
	}
	return Join(dir, sub...), nil
}

func partsFromSlash(base string, sub ...string) []string {
	var n int

	pp := make([][]string, len(sub))
	for i, s := range sub {
		pp[i] = splitFromSlash(s)
		n += len(pp[i])
	}

	if base != "" {
		n++
	}

	a := make([]string, 0, n)
	if base != "" {
		a = append(a, base)
	}
	for _, p := range pp {
		a = append(a, p...)
	}

	return a
}

func splitFromSlash(s string) []string {
	parts := strings.Split(s, "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
