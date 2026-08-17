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

// Aliases of the standard [fs.FileMode] mode bits, and their
// combined type and permission masks.
const (
	ModeDir        = fs.ModeDir
	ModeAppend     = fs.ModeAppend
	ModeExclusive  = fs.ModeExclusive
	ModeTemporary  = fs.ModeTemporary
	ModeSymlink    = fs.ModeSymlink
	ModeDevice     = fs.ModeDevice
	ModeNamedPipe  = fs.ModeNamedPipe
	ModeSocket     = fs.ModeSocket
	ModeSetuid     = fs.ModeSetuid
	ModeSetgid     = fs.ModeSetgid
	ModeCharDevice = fs.ModeCharDevice
	ModeSticky     = fs.ModeSticky
	ModeIrregular  = fs.ModeIrregular

	ModeType = fs.ModeType
	ModePerm = fs.ModePerm
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

// ReadDir is a proxy to the standard [fs.ReadDir] function
// which attempts to read the named directory on the given file
// system, returning its entries sorted by filename.
func ReadDir(fSys fs.FS, name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(fSys, name)
}

// ReadFile is a proxy to the standard [fs.ReadFile] function
// which attempts to read the content of a file with a given name
// on the given file system.
func ReadFile(fSys fs.FS, name string) ([]byte, error) {
	return fs.ReadFile(fSys, name)
}

// Stat is a proxy to the standard [fs.Stat] function
// which attempts to get [fs.FileInfo] about the given
// name on the given file system.
func Stat(fSys fs.FS, name string) (fs.FileInfo, error) {
	return fs.Stat(fSys, name)
}
