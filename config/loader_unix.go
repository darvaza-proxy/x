//go:build !windows

package config

import (
	"io/fs"
	"os"
	"path/filepath"
)

// NewFromFileOS returns the first successfully decoded option.
func (l *Loader[T]) NewFromFileOS(names ...string) (*T, error) {
	for i, s := range names {
		// TODO: handle `~`
		name, err := filepath.Abs(s)
		switch {
		case err != nil:
			return nil, err
		case name == "/":
			return nil, NewPathError(s, "load", fs.ErrInvalid)
		default:
			names[i] = name[1:]
		}
	}

	fSys := os.DirFS(`/`)
	return l.NewFromFile(fSys, names...)
}
