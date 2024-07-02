package assets

import (
	"io/fs"
	"net/http"

	"darvaza.org/x/web"
)

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

// ETaged allows files to declare their ETag to enable
// proper caching.
type ETaged interface {
	ETag() []string
}

// ETag checks the [fs.File] and [fs.FileInfo] to see
// if they implement the [ETaged] interface, and return
// their output.
func ETag(file fs.File) []string {
	if s := tryETag(file); len(s) > 0 {
		return s
	}

	if fi, _ := file.Stat(); fi != nil {
		return tryETag(fi)
	}

	return nil
}

func tryETag(candidates ...any) []string {
	for _, x := range candidates {
		if v, ok := x.(ETaged); ok {
			if s := v.ETag(); len(s) > 0 {
				return s
			}
		}
	}
	return nil
}

// DefaultRequestResolver returns the requested path
func DefaultRequestResolver(req *http.Request) (string, error) {
	var path string

	if req != nil {
		path = req.URL.Path
	}

	return path, nil
}

var (
	errBadRequest = &web.HTTPError{Code: http.StatusBadRequest}
	errNotFound   = &web.HTTPError{Code: http.StatusNotFound}
)
