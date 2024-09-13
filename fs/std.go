package fs

import "io/fs"

type (
	// FileInfo is an alias of the standard [fs.FileInfo] type.
	FileInfo = fs.FileInfo
	// FileMode is an alias of the standard [fs.FileMode] type.
	FileMode = fs.FileMode
	// DirEntry is an alias of the standard [fs.DirEntry] type.
	DirEntry = fs.DirEntry
	// PathError is an alias of the standard [fs.PathError] type.
	PathError = fs.PathError
)

var (
	// ErrInvalid is an alias of the standard [fs.ErrInvalid] constant.
	ErrInvalid = fs.ErrInvalid
	// ErrPermission is an alias of the standard [fs.ErrPermission] constant.
	ErrPermission = fs.ErrPermission
	// ErrExist is an alias of the standard [fs.ErrExist] constant.
	ErrExist = fs.ErrExist
	// ErrNotExist is an alias of the standard [fs.ErrNotExist] constant.
	ErrNotExist = fs.ErrNotExist
	// ErrClosed is an alias of the standard [fs.ErrClosed] constant.
	ErrClosed = fs.ErrClosed
)

// ValidPath is a proxy to the standard [fs.ValidPath]
// which reports whether the given path name valid and clean
// for use in a call to Open().
func ValidPath(name string) bool {
	return fs.ValidPath(name)
}

// Stat is a proxy to the standard [fs.Stat] function
// which attempts to get [fs.FileInfo] about the given
// name on the given file system.
func Stat(fSys fs.FS, name string) (fs.FileInfo, error) {
	return fs.Stat(fSys, name)
}
