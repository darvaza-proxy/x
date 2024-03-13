package embed

import (
	"io/fs"
	"path/filepath"
	"sync"

	"darvaza.org/core"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

var (
	_ (fs.FS) = (*Wrapped)(nil)
)

// Direct implements a watched [fs.FS] over the local file-system
type Direct struct {
	mu sync.Mutex
	f  billy.Filesystem
	w  *fsnotify.Watcher
	r  string
}

// Open attempts to open a file from the cached file-system
func (p *Direct) Open(name string) (fs.File, error) {
	if p.f == nil {
		return nil, newNotExistsError("open", name)
	}

	if f, ok := p.getFile(name); ok {
		return f, nil
	}

	bf, err := p.f.Open(name)
	if err != nil {
		return nil, err
	}

	ff, err := p.addFile(name, bf)
	if err != nil {
		defer bf.Close()

		return nil, err
	}

	return ff, nil
}

//nolint:unparam
func (p *Direct) getFile(string) (fs.File, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return nil, false
}

//nolint:unparam
func (p *Direct) addFile(name string, _ billy.File) (fs.File, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return nil, newError("open", name, core.ErrTODO)
}

// NewDirect creates a direct [FS] using fsnotify to control
// caching of metadata.
func NewDirect(dir string) (*Direct, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	p := &Direct{
		w: w,
		r: root,
		f: osfs.New(root,
			osfs.WithBoundOS(),
			osfs.WithDeduplicatePath(true)),
	}

	return p, nil
}
