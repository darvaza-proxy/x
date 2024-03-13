package embed

import (
	"io/fs"

	"darvaza.org/core"
)

func newError(op, path string, err error) error {
	return &fs.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

func newNotExistsError(op, path string) error {
	return newError(op, path, core.ErrNotExists)
}
