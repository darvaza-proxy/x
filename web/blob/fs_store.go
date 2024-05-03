package blob

import (
	"net/http"

	"darvaza.org/core"
)

// Store wraps a [fs.FS]
type Store interface {
	GetHandler(path string) (http.Handler, error)

	Close() error
}

func (*FS) newStore() (Store, error) {
	return nil, core.ErrTODO
}
