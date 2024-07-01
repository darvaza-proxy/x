package assets

import (
	"embed"
	"io/fs"
	"net/http"

	"darvaza.org/core"
)

// FS represents an HTTP capable file system.
type FS interface {
	fs.FS

	http.Handler
	Middleware() func(http.Handler) http.Handler

	SetResolver(func(*http.Request) (string, error))
}

// NewFS creates a new assets [FS] from the given [fs.FS], optional root, and an
// list of patterns. [github.com/gobwas/glob] patterns supported.
func NewFS(base fs.FS, root string, patterns ...string) (FS, error) {
	switch v := base.(type) {
	case embed.FS:
		return NewEmbedFS(&v, root, patterns...)
	case *embed.FS:
		return NewEmbedFS(v, root, patterns...)
	case nil:
		err := core.QuietWrap(core.ErrInvalid, "base fs.FS not provided")
		return nil, err
	default:
		err := core.Wrap(core.ErrTODO, "%T not yet supported", base)
		return nil, err
	}
}

// MustFS is equivalent to [NewFS] but panics on error.
func MustFS(base fs.FS, root string, patterns ...string) FS {
	o, err := NewFS(base, root, patterns...)
	if err != nil {
		panic(err)
	}
	return o
}
