package memfs

import (
	"strings"

	"darvaza.org/x/fs"
)

var (
	_ fs.MkdirFS    = (*FS)(nil)
	_ fs.MkdirAllFS = (*FS)(nil)
)

// Mkdir creates a directory in the tree
func (m *FS) Mkdir(path string, _ fs.FileMode) error {
	if !fs.ValidPath(path) {
		return errInvalid("mkdir", path)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.unsafeMkdir(path, false)
	return err
}

func (m *FS) unsafeMkdir(dir string, allowExists bool) (*dirNode, error)

// MkdirAll creates all intermediate directories if needed
// and doesn't complain if any already exists.
func (m *FS) MkdirAll(path string, _ fs.FileMode) error {
	if !fs.ValidPath(path) {
		return errInvalid("mkdir", path)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.unsafeMkdirAll(path)
	return err
}

func (m *FS) unsafeMkdirAll(path string) (*dirNode, error) {
	dir, rest := "", path
	for {
		i := strings.IndexRune(rest, '/')
		if i < 0 {
			// last
			return m.unsafeMkdir(path, true)
		}

		dir, rest = path[:i-1], rest[i+1:]
		if _, err := m.unsafeMkdir(dir, true); err != nil {
			return nil, err
		}
	}
}
