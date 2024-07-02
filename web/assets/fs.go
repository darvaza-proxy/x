package assets

import (
	"io/fs"
)

// ContentTypedFS is the interface implemented by a file system
// to get a file's MIME Content-Type
type ContentTypedFS interface {
	fs.FS

	ContentType(path string) (string, error)
}

// ContentTypeSetterFS is the interface implemented by a file system
// to set a file's MIME Content-Type.
type ContentTypeSetterFS interface {
	fs.FS

	SetContentType(path, contentType string) (string, error)
}
