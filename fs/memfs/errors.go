package memfs

import (
	"io/fs"
	"syscall"
)

func newPathError(op, name string, err error) error {
	return &fs.PathError{
		Op:   op,
		Err:  err,
		Path: name,
	}
}

func errExist(op, name string) error {
	return newPathError(op, name, fs.ErrExist)
}

func errNotExist(op, name string) error {
	return newPathError(op, name, fs.ErrNotExist)
}

func errInvalid(op, name string) error {
	return newPathError(op, name, fs.ErrInvalid)
}

func errNotDir(op, name string) error {
	return newPathError(op, name, syscall.ENOTDIR)
}

func errPermission(op, name string) error {
	return newPathError(op, name, fs.ErrPermission)
}
