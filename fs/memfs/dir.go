package memfs

import (
	"io/fs"
	"time"

	"github.com/DmitriyVTitov/size"
)

var (
	_ Node           = (*dirNode)(nil)
	_ fs.FileInfo    = (*dirNode)(nil)
	_ fs.File        = (*dirNode)(nil)
	_ fs.ReadDirFile = (*dirNode)(nil)

	_ fs.DirEntry = (*dirEntry)(nil)
)

type dirNode struct {
	parent  *dirNode
	path    string
	name    string
	entries []Node
	modTime time.Time
}

func (d *dirNode) Open() (fs.File, error)     { return d, nil }
func (d *dirNode) Stat() (fs.FileInfo, error) { return d, nil }
func (d *dirNode) Sys() any                   { return d }
func (d *dirNode) Name() string               { return d.name }
func (d *dirNode) Path() string               { return d.path }
func (d *dirNode) ModTime() time.Time         { return d.modTime }

func (*dirNode) Close() error      { return nil }
func (*dirNode) IsDir() bool       { return true }
func (*dirNode) Mode() fs.FileMode { return 0755 }

func (d *dirNode) Read([]byte) (int, error) {
	return 0, errPermission("read", d.path)
}

func (d *dirNode) ReadDir(n int) ([]fs.DirEntry, error)

func (d *dirNode) Size() int64 {
	return int64(size.Of(d))
}

type dirEntry struct {
	n    Node
	name string
}

func (e *dirEntry) mustInfo() fs.FileInfo {
	fi, err := e.n.Stat()
	if err != nil {
		panic(err)
	}
	return fi
}

func (e *dirEntry) Name() string               { return e.name }
func (e *dirEntry) Info() (fs.FileInfo, error) { return e.n.Stat() }
func (e *dirEntry) IsDir() bool                { return e.mustInfo().IsDir() }
func (e *dirEntry) Type() fs.FileMode          { return e.mustInfo().Mode() }
