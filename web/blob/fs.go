package blob

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/web"
)

var (
	_ http.Handler = (*FS)(nil)
	_ io.Closer    = (*FS)(nil)
)

// FS is an [http.Handler] running on top of a [fs.FS]
type FS struct {
	mu        sync.Mutex
	err       error
	store     Store
	trySuffix []string

	FS       fs.FS
	BaseDIR  string
	Index    string
	Suffixes []string

	OnError  func(http.ResponseWriter, *http.Request, error)
	Resolver func(*http.Request) (string, error)
}

// ServeHTTP implements [http.Handler]
func (f *FS) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f.handleRequest(rw, req, nil)
}

// NewMiddleware creates Middleware using this FS as overlay
func (f *FS) NewMiddleware(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		f.handleRequest(rw, req, next)
	}
	return http.HandlerFunc(fn)
}

func (f *FS) handleRequest(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	h, err := f.getHandlerByRequest(req)
	switch {
	case h != nil:
		// found, serve
		h.ServeHTTP(rw, req)
	case err != nil && !os.IsNotExist(err):
		// bad request, fail
		f.handleError(rw, req, err)
	case next != nil:
		// middleware, skip
		next.ServeHTTP(rw, req)
	default:
		// not found, fail
		f.handleError(rw, req, fs.ErrNotExist)
	}
}

func (f *FS) getHandlerByRequest(req *http.Request) (http.Handler, error) {
	path, err := f.resolveRequest(req)
	switch {
	case err != nil:
		return nil, err
	default:
		return f.GetHandler(path)
	}
}

func (f *FS) resolveRequest(req *http.Request) (string, error) {
	if f.Resolver != nil {
		return f.Resolver(req)
	}

	return req.URL.Path, nil
}

// GetHandler returns the [http.Handler] for the given path.
func (f *FS) GetHandler(path string) (http.Handler, error) {
	var name string
	var trailing bool

	// preflight checks
	if s, ok := CleanPath(path); !ok {
		err := web.NewHTTPError(http.StatusBadRequest, nil, "invalid path")
		return nil, err
	} else if s == "/" {
		name, trailing = "", true
	} else {
		// regular
		name, trailing = strings.CutSuffix(s[1:], "/")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// auto-init
	if err := f.init(); err != nil {
		return nil, err
	}

	return f.tryGetHandler(name, trailing)
}

func (f *FS) tryGetHandler(fileName string, trailing bool) (http.Handler, error) {
	for _, suffix := range f.trySuffix {
		name := joinNameSuffix(fileName, suffix)

		h, err := f.store.GetHandler(name)
		switch {
		case h != nil:
			// found
			h = f.prepareHandler(h, fileName, suffix, trailing)
			if h != nil {
				// ready
				return h, nil
			}
		case err != nil && !os.IsNotExist(err):
			// bad error
			return nil, err
		}
	}

	// not found
	return nil, fs.ErrNotExist
}

// revive:disable:flag-parameter
func (*FS) prepareHandler(h http.Handler, fileName, suffix string, trailing bool) http.Handler {
	// revive:enable:flag-parameter

	if trailing {
		if suffix == "" || suffix[0] != '/' {
			// remove trailing slash
			return newRedirect(fileName)
		}
	} else if suffix != "" && suffix[0] == '/' {
		// add trailing slash
		return newRedirect(fileName + "/")
	}

	return h
}
func (f *FS) init() error {
	switch {
	case f.store != nil:
		// already initialized
		return nil
	case f.err != nil:
		// previous failure
		return f.err
	default:
		// assemble store
		s, err := f.newStore()
		if err != nil {
			return f.afterClose(err)
		}

		// suffixes
		suffixes := []string{""}
		if f.Index != "" {
			suffixes = append(suffixes, "/"+f.Index)
		}
		if len(f.Suffixes) > 0 {
			suffixes = append(suffixes, f.Suffixes...)
		}

		f.store = s
		f.trySuffix = suffixes
		return nil
	}
}

// Close attempts to close the store and frees the FS.
// If the store fails to Close, the FS becomes
// permanently unusable.
func (f *FS) Close() error {
	var err error
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.store != nil {
		if e := f.store.Close(); e != nil {
			err = core.Wrap(e, "blob.FS.Close")
		}
	}

	return f.afterClose(err)
}

func (f *FS) afterClose(err error) error {
	f.store = nil
	f.trySuffix = nil
	f.err = err
	return err
}
func (f *FS) handleError(rw http.ResponseWriter, req *http.Request, err error) {
	if f.OnError != nil {
		f.OnError(rw, req, err)
	} else if h, ok := web.ErrorHandler(req.Context()); ok {
		h(rw, req, err)
	} else {
		f.defaultErrorHandler(rw, req, err)
	}
}

func (*FS) defaultErrorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	h, ok := err.(http.Handler)
	if !ok {
		var code int

		if wee, ok := err.(web.Error); ok {
			code = wee.Status()
		} else if os.IsNotExist(err) {
			code = http.StatusNotFound
		} else {
			code = http.StatusInternalServerError
		}

		h = web.NewHTTPError(code, err, "")
	}

	h.ServeHTTP(rw, req)
}
