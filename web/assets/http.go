package assets

import (
	"io/fs"
)

// ContentTypeSetter is the interface that allows a [fs.File] or its
// [fs.FileInfo] to assign a MIME Content-Type to it.
// if empty it will be ignored
type ContentTypeSetter interface {
	SetContentType(string) string
}

// ContentTyped allows files to declare their Content-Type.
type ContentTyped interface {
	ContentType() string
}

// ContentType checks if the [fs.File] or its [fs.FileInfo] effectively
// provides the Content-Type, and returns it if so.
func ContentType(file fs.File) string {
	if ct := tryContentType(file); ct != "" {
		return ct
	}

	fi, _ := file.Stat()
	if fi != nil {
		return tryContentType(fi)
	}

	return ""
}

func tryContentType(candidates ...any) string {
	for _, x := range candidates {
		if v, ok := x.(ContentTyped); ok {
			if ct := v.ContentType(); ct != "" {
				return ct
			}
		}
	}

	return ""
}
