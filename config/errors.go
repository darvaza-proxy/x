package config

import (
	"os"

	"darvaza.org/core"
)

// NewPathError is a shortcut to generate [os.PathError].
func NewPathError(filename, op string, err error) *os.PathError {
	if e, ok := err.(*os.PathError); ok {
		// unwrap and reuse information if possible.
		return &os.PathError{
			Path: core.Coalesce(e.Path, filename),
			Op:   core.Coalesce(e.Op, op),
			Err:  e.Err,
		}
	}

	// wrap
	return &os.PathError{
		Path: filename,
		Op:   op,
		Err:  err,
	}
}
