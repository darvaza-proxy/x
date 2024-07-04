package assets

import (
	"io/fs"
)

// ContentTypeSetter is the interface that allows a [fs.File] or its
// [fs.FileInfo] to assign a MIME Content-Type to it.
// If empty it will be ignored.
// On success the new value is returned.
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

// ETagsSetter is the interface that allows a [fs.File] or its
// [fs.FileInfo] to assign a ETags to it.
// Any previous value will be replaced, unless none is provided.
// The effective ETags set is returned.
type ETagsSetter interface {
	SetETags(...string) []string
}

// ETaged allows files to declare their ETag to enable
// proper caching.
type ETaged interface {
	ETags() []string
}

// ETags checks the [fs.File] and [fs.FileInfo] to see
// if they implement the [ETaged] interface, and return
// their output.
func ETags(file fs.File) []string {
	if s := tryETags(file); len(s) > 0 {
		return s
	}

	if fi, _ := file.Stat(); fi != nil {
		return tryETags(fi)
	}

	return nil
}

func tryETags(candidates ...any) []string {
	for _, x := range candidates {
		if v, ok := x.(ETaged); ok {
			if s := v.ETags(); len(s) > 0 {
				return s
			}
		}
	}
	return nil
}
