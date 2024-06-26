package fs

import (
	"github.com/gobwas/glob"
)

// Glob is a compiled globbing pattern from https://github.com/gobwas/glob
type Glob = glob.Glob

// GlobCompile compiles a list of file globbing patterns using
// https://github.com/gobwas/glob
func GlobCompile(patterns ...string) ([]Glob, error) {
	out := make([]Glob, 0, len(patterns))

	for _, pat := range patterns {
		g, err := glob.Compile(pat, '/')
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}

	return out, nil
}
