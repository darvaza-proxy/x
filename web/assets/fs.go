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
// On success the new value is returned.
type ContentTypeSetterFS interface {
	fs.FS

	SetContentType(path, contentType string) (string, error)
}

// ETagedFS is the interface implemented by a file system
type ETagedFS interface {
	fs.FS

	ETags(path string) ([]string, error)
}

// ETagsSetterFS is the interface implemented by a file system
// to set a file's ETags.
// On success the new tags is returned.
type ETagsSetterFS interface {
	fs.FS

	SetETags(string, ...string) ([]string, error)
}
