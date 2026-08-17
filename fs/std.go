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
	// WalkDirFunc is an alias of the standard [fs.WalkDirFunc] type.
	WalkDirFunc = fs.WalkDirFunc
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

	// SkipDir is an alias of the standard [fs.SkipDir] sentinel.
	SkipDir = fs.SkipDir
	// SkipAll is an alias of the standard [fs.SkipAll] sentinel.
	SkipAll = fs.SkipAll
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

// Sub is a proxy to the standard [fs.Sub] function
// which returns a file system corresponding to the subtree
// rooted at the given directory.
func Sub(fSys fs.FS, dir string) (fs.FS, error) {
	return fs.Sub(fSys, dir)
}

// WalkDir is a proxy to the standard [fs.WalkDir] function
// which walks the file tree rooted at the given directory,
// calling fn for each file or directory in the tree.
func WalkDir(fSys fs.FS, root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(fSys, root, fn)
}

// FileInfoToDirEntry is a proxy to the standard
// [fs.FileInfoToDirEntry] function which returns a [fs.DirEntry]
// reporting the given [fs.FileInfo].
func FileInfoToDirEntry(info fs.FileInfo) fs.DirEntry {
	return fs.FileInfoToDirEntry(info)
}

// FormatDirEntry is a proxy to the standard [fs.FormatDirEntry]
// function which returns a formatted version of the given
// [fs.DirEntry] for human readability.
func FormatDirEntry(dir fs.DirEntry) string {
	return fs.FormatDirEntry(dir)
}

// FormatFileInfo is a proxy to the standard [fs.FormatFileInfo]
// function which returns a formatted version of the given
// [fs.FileInfo] for human readability.
func FormatFileInfo(info fs.FileInfo) string {
	return fs.FormatFileInfo(info)
}
