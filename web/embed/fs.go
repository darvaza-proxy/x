// Package embed implements a filesystem of embedded files.
package embed

import "io/fs"

// interface assertions
var (
	_ fs.FS     = (*FS)(nil)
	_ fs.StatFS = (*FS)(nil)
)

// Embedded ...
type Embedded interface {
	Info() (fs.FileInfo, error)
	Open() (fs.File, error)
}

// FS implements a [fs.FS] of embedded []
type FS struct {
	files map[string]Embedded
}

// Add adds a [Embedded] file to the [fs.FS].
func (fsys *FS) Add(name string, file Embedded) error {
	if file == nil || !fs.ValidPath(name) {
		return fs.ErrInvalid
	}

	if fsys.files == nil {
		fsys.files = make(map[string]Embedded)
	}

	fsys.files[name] = file
	return nil
}

// Open gets a descriptor to read an [Embedded] [fs.File].
func (fsys *FS) Open(name string) (fs.File, error) {
	f, err := fsys.getFile(name, "open")
	if err != nil {
		return nil, err
	}
	return f.Open()
}

// Stat gets information aboutn an an [Embedded] [fs.File].
func (fsys *FS) Stat(name string) (fs.FileInfo, error) {
	f, err := fsys.getFile(name, "stat")
	if err != nil {
		return nil, err
	}
	return f.Info()
}

func (fsys *FS) getFile(name, op string) (Embedded, *fs.PathError) {
	if f, ok := fsys.files[name]; ok {
		return f, nil
	}

	err := &fs.PathError{
		Op:   op,
		Path: name,
		Err:  fs.ErrNotExist,
	}
	return nil, err
}
