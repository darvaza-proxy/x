package assets

import (
	"embed"
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

// FS represents an HTTP capable file system.
type FS interface {
	// FS is a standard Go file system
	fs.FS

	// ServeHTTP handles requests as normal and fails with a standard 404
	http.Handler
	// Middleware handles requests for known files and pass over the rest
	Middleware() func(http.Handler) http.Handler

	// SetResolver sets a helper that will extract the request path from
	// the [http.Request]. by default req.URL.Path will be used.
	SetResolver(func(*http.Request) (string, error))
}

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

// NewFS creates a new assets [FS] from the given [fs.FS], optional root, and an
// list of patterns. [github.com/gobwas/glob] patterns supported.
func NewFS(base fs.FS, root string, patterns ...string) (FS, error) {
	dir, globs, err := sanitizeNewFS(root, patterns)
	if err != nil {
		return nil, err
	}

	switch v := base.(type) {
	case embed.FS:
		return unsafeNewEmbedFS(&v, dir, globs)
	case *embed.FS:
		return unsafeNewEmbedFS(v, dir, globs)
	case nil:
		err = core.QuietWrap(core.ErrInvalid, "base fs.FS not provided")
	default:
		err = core.Wrap(core.ErrTODO, "%T not yet supported", base)
	}

	return nil, err
}

func sanitizeNewFS(root string, patterns []string) (string, []fs.Matcher, error) {
	dir, ok := fs.Clean(root)
	if !ok {
		return "", nil, newErrInvalid("readdir", root)
	}

	globs, err := fs.GlobCompile(patterns...)
	if err != nil {
		return "", nil, err
	}

	return dir, globs, err
}

// MustFS is equivalent to [NewFS] but panics on error.
func MustFS(base fs.FS, root string, patterns ...string) FS {
	o, err := NewFS(base, root, patterns...)
	if err != nil {
		panic(err)
	}
	return o
}
