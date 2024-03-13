package embed

import (
	"io/fs"

	"darvaza.org/core"
)

var (
	_ (fs.FS) = (*Wrapped)(nil)
)

// Wrapped implements a fs.FS over the standard embed.FS
type Wrapped struct {
	FS fs.FS
}

// Open attempts to open a file from the wrapped file-system
func (p *Wrapped) Open(name string) (fs.File, error) {
	if p.FS != nil {
		return p.FS.Open(name)
	}

	return nil, core.ErrInvalid
}
