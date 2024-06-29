// Package fs provides tools to work with [fs.FS]
package fs

import (
	"io/fs"
	"time"
)

// FS is an alias of the standard [fs.FS] interface.
type FS = fs.FS

// ChmodFS is the interface implemented by a file system that
// provides the functionality of [os.Chmod].
type ChmodFS interface {
	FS
	Chmod(path string, mode FileMode) error
}

// ChtimesFS is the interface implemented by a file system that
// provides the functionality of [os.Chtimes].
type ChtimesFS interface {
	FS
	Chtimes(path string, atime, mtime time.Time) error
}

// GlobFS is an alias of the standard [fs.GlobFS] interface.
type GlobFS = fs.GlobFS

// MkdirFS is the interface implemented by a file system that
// provides the functionality of [os.Mkdir].
type MkdirFS interface {
	FS
	Mkdir(path string, mode FileMode) error
}

// MkdirAllFS is the interface implemented by a file system that
// provides the functionality of [os.MkdirAll].
type MkdirAllFS interface {
	FS
	MkdirAll(path string, mode FileMode) error
}

// MkdirTempFS is the interface implemented by a file system that
// provides the functionality of [os.MkdirTemp].
type MkdirTempFS interface {
	FS
	MkdirTemp(dir string, pattern string) error
}

// ReadDirFS is an alias of the standard [fs.ReadDirFS] type.
type ReadDirFS = fs.ReadDirFS

// ReadFileFS is an alias of the standard [fs.ReadFileFS] type.
type ReadFileFS = fs.ReadFileFS

// ReadlinkFS is the interface implemented by a file system that
// provides the functionality of [os.Readlink].
type ReadlinkFS interface {
	FS
	Readlink(path string) (string, error)
}

// RemoveFS is the interface implemented by a file system that
// provides the functionality of [os.Remove].
type RemoveFS interface {
	FS
	Remove(path string) error
}

// RemoveAllFS is the interface implemented by a file system that
// provides the functionality of [os.RemoveAll].
type RemoveAllFS interface {
	FS
	RemoveAll(path string) error
}

// RenameFS is the interface implemented by a file system that
// provides the functionality of [os.Rename].
type RenameFS interface {
	FS
	Rename(oldPath, newPath string) error
}

// StatFS is an alias of the standard [fs.StatFS] type.
type StatFS = fs.StatFS

// SubFS is an alias of the standard [fs.ReadFileFS] type.
type SubFS = fs.SubFS

// SymlinkFS is the interface implemented by a file system that
// provides the functionality of [os.Symlink].
type SymlinkFS interface {
	FS
	Symlink(target, path string) error
}

// WriteFileFS is the interface implemented by a file system that
// provides the functionality of [os.WriteFile].
type WriteFileFS interface {
	FS
	WriteFile(path string, data []byte, perm FileMode) error
}
