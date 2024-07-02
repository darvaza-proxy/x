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
