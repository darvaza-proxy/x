package fs

import (
	"io"
	"io/fs"
)

// File is an alias of the standard [fs.File] interface.
type File = fs.File

// ReadDirFile is an alias of the standard [fs.ReadDirFile] interface.
type ReadDirFile = fs.ReadDirFile

var (
	_ fs.File       = (*ClosedFile)(nil)
	_ io.ReadSeeker = (*ClosedFile)(nil)
	_ io.Writer     = (*ClosedFile)(nil)
)

// ClosedFile always returns [fs.ErrClosed]
type ClosedFile struct {
	FileInfo fs.FileInfo
}

// Read implements the [io.ReadSeeker] interface always returning [fs.ErrClosed]
func (*ClosedFile) Read([]byte) (int, error) { return 0, fs.ErrClosed }

// Seek implements the [io.ReadSeeker] interface always returning [fs.ErrClosed]
func (*ClosedFile) Seek(int64, int) (int64, error) { return 0, fs.ErrClosed }

// Write implements the [io.Writer] interface always returning [fs.ErrClosed]
func (*ClosedFile) Write([]byte) (int, error) { return 0, fs.ErrClosed }

// Close implements the [fs.File] interface, always succeeding as [ClosedFile]
// is always closed.
func (*ClosedFile) Close() error { return nil }

// Stat implements the [fs.File] interface, allowing the return of the original
// [fs.FileInfo], or failing with [fs.ErrInvalid] if not known.
func (f *ClosedFile) Stat() (fs.FileInfo, error) {
	if f != nil && f.FileInfo != nil {
		return f.FileInfo, nil
	}

	return nil, fs.ErrInvalid
}
