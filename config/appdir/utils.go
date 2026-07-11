package appdir

import (
	"os"
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

// getEnvDir returns the absolute directory named by the given
// environment variable, or "" when it holds no absolute path,
// whether unset, empty, or relative. Ignoring a relative value
// matches the XDG Base Directory Specification, which treats a
// relative path in its variables as invalid.
func getEnvDir(key string) string {
	if dir := os.Getenv(key); filepath.IsAbs(dir) {
		return dir
	}

	return ""
}

// getEnvHomeDir returns the directory designated by the given
// environment variable via [getEnvDir], or the given fallback
// joined to the user's home directory when the variable is
// unset.
func getEnvHomeDir(key, fallback string) (string, error) {
	if dir := getEnvDir(key); dir != "" {
		return dir, nil
	}

	return joinFn(os.UserHomeDir, fallback)
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
