package config

import "os"

// NewPathError is a shortcut to generate [os.PathError].
func NewPathError(filename, op string, err error) *os.PathError {
	return &os.PathError{
		Path: filename,
		Op:   op,
		Err:  err,
	}
}
