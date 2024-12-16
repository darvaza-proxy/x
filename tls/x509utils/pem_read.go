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

// ReadPEM invokes a callback for each PEM block found in the input data.
// It returns ErrEmpty if the input is empty, core.ErrInvalid if it
// fails to decode.
func ReadPEM(b []byte, cb DecodePEMBlockFunc) error {
	var block *pem.Block

	if len(b) == 0 {
		return ErrEmpty
	}

	for {
		block, b = pem.Decode(b)
		switch {
		case block == nil:
			// failed to decode
			return core.ErrInvalid
		case cb != nil && !cb(nil, "", block):
			// aborted
			return nil
		case len(b) == 0:
			// EOF
			return nil
		}
	}
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
func ReadStringPEM(s string, cb DecodePEMBlockFunc, options ...ReadOption) error {
	r := &readOptions{
		cb:   cb,
		dirs: true,
	}

	for _, fn := range options {
		if err := fn(r); err != nil {
			return err
		}
	}

	return r.run(s)
}

type readOptions struct {
	cb DecodePEMBlockFunc
	fs fs.FS

	dirs bool
}

func (r *readOptions) run(s string) error {
	if ReadPEM([]byte(s), r.cb) == nil {
		// raw. done.
		return nil
	}

	st, err := r.stat(s)
	if err == nil {
		// string is a file path
		return r.readPathPEM(s, st)
	}

	if pe, ok := err.(*os.PathError); ok {
		if pe.Err == os.ErrInvalid {
			// not a path
			err = fs.ErrInvalid
		}
	}

	return err
}

func (r *readOptions) stat(s string) (fs.FileInfo, error) {
	if r.fs == nil {
		return os.Stat(s)
	}

	return fs.Stat(r.fs, s)
}

func (r *readOptions) readDirPEM(s string) error {
	switch {
	case !r.dirs:
		return &fs.PathError{
			Op:   "Read",
			Path: s,
			Err:  fs.ErrNotExist,
		}
	case r.fs == nil:
		return ReadDirPEM(os.DirFS(s), ".", r.cb)
	default:
		return ReadDirPEM(r.fs, s, r.cb)
	}
}

func (r *readOptions) readFile(s string) ([]byte, error) {
	if r.fs == nil {
		return os.ReadFile(s)
	}

	return fs.ReadFile(r.fs, s)
}

func (r *readOptions) readPathPEM(s string, st fs.FileInfo) error {
	if st.IsDir() {
		return r.readDirPEM(s)
	}

	// file
	b, err := r.readFile(s)
	if err != nil {
		return err
	}

	return ReadPEM(b, r.cb)
}

// ReadOption tunes how [ReadStringPEM] operates.
type ReadOption func(*readOptions) error

// ReadWithFS specifies a [fs.FS] to use when resolving paths.
func ReadWithFS(fSys fs.FS) ReadOption {
	return func(r *readOptions) error {
		switch {
		case r == nil:
			return core.ErrNilReceiver
		case fSys == nil:
			return core.Wrap(core.ErrInvalid, "fs not specified")
		default:
			r.fs = fSys
			return nil
		}
	}
}

// ReadWithoutDirs prevents [ReadStringPEM] from scanning directories.
func ReadWithoutDirs() ReadOption {
	return func(r *readOptions) error {
		if r == nil {
			return core.ErrNilReceiver
		}

		r.dirs = false
		return nil
	}
}

// ReadWithDirs allows [ReadStringPEM] to scan directories.
// This is the default.
func ReadWithDirs() ReadOption {
	return func(r *readOptions) error {
		if r == nil {
			return core.ErrNilReceiver
		}

		r.dirs = true
		return nil
	}
}
