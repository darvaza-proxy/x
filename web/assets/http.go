package assets

import (
	"net/http"

	"darvaza.org/x/web"
)

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
