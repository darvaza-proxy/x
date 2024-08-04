package assets

import (
	"net/http"

	"darvaza.org/x/fs"
	"darvaza.org/x/web"
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
func ContentType(v any) string {
	// direct test
	ct, ok := tryContentType(v)
	if ok {
		return ct
	}

	// via fs.FileInfo
	if fi, ok := tryStat(v); ok {
		ct, _ = tryContentType(fi)
	}

	return ct
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
func ETags(v any) []string {
	if s, ok := tryETags(v); ok {
		return s
	}

	if fi, ok := tryStat(v); ok {
		if s, ok := tryETags(fi); ok {
			return s
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

// httpView is a private interface used to provide the basis for
// [http.Handler] and Middleware() to a [FS].
type httpView interface {
	getResolver() func(*http.Request) (string, error)
	getFileHandler(string) http.Handler
}

// httpError is an error that can render itself.
type httpError interface {
	web.Error
	http.Handler
}

func serveHTTP(v httpView, rw http.ResponseWriter, req *http.Request, next http.Handler) {
	var h http.Handler

	f, err := getFileFromRequest(v, req)
	switch {
	case f != nil:
		h = f
	case err != errNotFound:
		h = err
	case next != nil:
		h = next
	default:
		h = errNotFound
	}

	h.ServeHTTP(rw, req)
}

func getFileFromRequest(v httpView, req *http.Request) (http.Handler, httpError) {
	r := v.getResolver()
	path, err := r(req)

	if err == nil && path != "" && path[0] == '/' {
		path, ok := fs.Clean(path[1:])
		if ok {
			if h := v.getFileHandler(path); h != nil {
				return h, nil
			}

			return nil, errNotFound
		}
	}

	return nil, newBadRequest(err)
}

func newBadRequest(err error) httpError {
	var code int

	switch e := err.(type) {
	case httpError:
		return e
	case web.Error:
		code = e.HTTPStatus()
	}

	if code < 400 {
		code = http.StatusBadRequest
	}

	return &web.HTTPError{
		Code: code,
		Err:  err,
	}
}

var errNotFound = &web.HTTPError{Code: http.StatusNotFound}
