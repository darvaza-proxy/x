package x509utils

import (
	"encoding/pem"
	"io/fs"
	"os"
	"path"

	"darvaza.org/core"
)

// DecodePEMBlockFunc is called for each PEM block coded. it returns false
// to terminate the loop
type DecodePEMBlockFunc func(fSys fs.FS, filename string, block *pem.Block) bool

// ReadPEM invoques a callback for each PEM block found
// it can receive raw PEM data
func ReadPEM(b []byte, cb DecodePEMBlockFunc) error {
	switch {
	case len(b) == 0:
		return ErrEmpty
	case cb == nil:
		// nothing do
		return nil
	}

	if block, rest := pem.Decode(b); block != nil {
		// PEM chain
		_ = readBlock(nil, "", block, rest, cb)
		return nil
	}

	// Not PEM
	return core.ErrInvalid
}

func readBlock(fSys fs.FS, filename string, block *pem.Block, rest []byte, cb DecodePEMBlockFunc) bool {
	for block != nil {
		if !cb(fSys, filename, block) {
			// cascade termination request
			return false
		} else if len(rest) == 0 {
			// EOF
			break
		}

		// next
		block, rest = pem.Decode(rest)
	}

	return true
}

// ReadFilePEM reads a PEM file calling cb for each block
func ReadFilePEM(fSys fs.FS, filename string, cb DecodePEMBlockFunc) error {
	b, err := fs.ReadFile(fSys, filename)
	if err != nil {
		return err
	}

	err = ReadPEM(b, cb)
	if err != nil {
		return &fs.PathError{
			Op:   "pem.Decode",
			Path: filename,
			Err:  err,
		}
	}
	return nil
}

// ReadDirPEM reads a directory recursively looking for PEM files
func ReadDirPEM(fSys fs.FS, dir string, cb DecodePEMBlockFunc) error {
	files, dirs, err := splitReadDir(fSys, dir)
	switch {
	case err != nil:
		// invalid directory
		return err
	case cb == nil:
		// nothing to run
		return nil
	default:
		var errs core.CompoundError

		// files first
		if err = doReadDirPEM(fSys, dir, cb, ReadFilePEM, files); err != nil {
			errs.AppendError(err)
		}

		// then sub-directories
		if err = doReadDirPEM(fSys, dir, cb, ReadDirPEM, dirs); err != nil {
			errs.AppendError(err)
		}

		return errs.AsError()
	}
}

func doReadDirPEM(fSys fs.FS, dir string, cb DecodePEMBlockFunc,
	fn func(fs.FS, string, DecodePEMBlockFunc) error, entries []fs.DirEntry) error {
	//
	var errs core.CompoundError

	for _, fi := range entries {
		fullName := path.Join(dir, fi.Name())
		if err := fn(fSys, fullName, cb); err != nil {
			errs.AppendError(err)
		}
	}

	return errs.AsError()
}

func splitReadDir(fSys fs.FS, dir string) ([]fs.DirEntry, []fs.DirEntry, error) {
	dd, err := fs.ReadDir(fSys, dir)
	if err != nil {
		return nil, nil, err
	}

	files := make([]fs.DirEntry, 0, len(dd))
	dirs := make([]fs.DirEntry, 0, len(dd))
	for _, de := range dd {
		if de.IsDir() {
			dirs = append(dirs, de)
		} else {
			files = append(files, de)
		}
	}

	return files, dirs, nil
}

// ReadStringPEM works over raw PEM data, a filename or directory reading
// PEM blocks and invoking a callback for each.
func ReadStringPEM(s string, cb DecodePEMBlockFunc) error {
	if ReadPEM([]byte(s), cb) == nil {
		// raw. done.
		return nil
	}

	st, err := os.Stat(s)
	if err == nil {
		// string is a file path
		return readOSPathPEM(s, st, cb)
	}

	if pe, ok := err.(*os.PathError); ok {
		if pe.Err == os.ErrInvalid {
			// not a path
			err = fs.ErrInvalid
		}
	}

	return err
}

func readOSPathPEM(s string, st fs.FileInfo, cb DecodePEMBlockFunc) error {
	if st.IsDir() {
		return ReadDirPEM(os.DirFS(s), ".", cb)
	}

	// file
	b, err := os.ReadFile(s)
	if err != nil {
		return err
	}

	return ReadPEM(b, cb)
}
