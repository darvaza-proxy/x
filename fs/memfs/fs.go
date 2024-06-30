// Package memfs implements an in-memory file system abstraction
package memfs

import (
	"sync"

	"github.com/derekparker/trie/v2"

	"darvaza.org/core"
	"darvaza.org/x/fs"
)

var (
	_ fs.FS        = (*FS)(nil)
	_ fs.ReadDirFS = (*FS)(nil)
	_ fs.StatFS    = (*FS)(nil)
)

// Node represents a file or directory in the [FS] tree.
type Node interface {
	Open() (fs.File, error)
	Stat() (fs.FileInfo, error)
}

// MemFS is a filesystem tree implemented in memory.
type FS struct {
	mu sync.RWMutex
	t  *trie.Trie[Node]
}

// New creates a new [MemFS]
func New() *FS {
	m := &FS{
		t: trie.New[Node](),
	}
	return m
}

// Open gets a node from the tree
func (m *FS) Open(path string) (fs.File, error) {
	node, err := m.getPath("open", path)
	if err != nil {
		return nil, err
	}

	return node.Open()
}

// ReadDir gets the list of subnodes in a directory.
func (m *FS) ReadDir(path string) ([]fs.DirEntry, error) {
	node, err := m.getPath("readdir", path)
	if err != nil {
		return nil, err
	}

	fi, err := node.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		f, err := node.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()

		d, ok := f.(fs.ReadDirFile)
		if ok {
			return d.ReadDir(0)
		}
	}

	return nil, errNotDir("readdir", path)
}

// Stat returns information about a node in the tree
func (m *FS) Stat(path string) (fs.FileInfo, error) {
	node, err := m.getPath("stat", path)
	if err != nil {
		return nil, err
	}

	return node.Stat()
}

// Add adds a file to the tree.
func (m *FS) Add(path string, data Node) error {
	dir, file, err := m.validateAdd(path, data)
	if err != nil {
		return err
	}

	path = unsafeJoin(dir, file)

	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.unsafeGet(path)
	if ok {
		return errExist("add", path)
	}

	dirNode, err := m.unsafeMkdirAll(dir)
	if err != nil {
		return err
	}

	return m.unsafeAddFile(dirNode, file, data)
}

func (*FS) validateAdd(path string, data Node) (dir, file string, err error) {
	if data == nil {
		return "", "", errInvalid("add", path)
	}

	dir, file = fs.Split(path)
	if fs.ValidPath(dir) {
		return "", "", errInvalid("add", path)
	}

	fi, err := data.Stat()
	switch {
	case err != nil:
		return "", "", err
	case !fi.Mode().IsRegular():
		// files only
		err = errInvalid("add", path)
		return "", "", core.Wrap(err, "only regular files allowed")
	default:
		name := fi.Name()
		if name != "" && name != file {
			err = errInvalid("add", path)
			return "", "", core.Wrapf(err, "name:%q ≠ %q", name, file)
		}
	}

	return dir, file, nil
}

func (m *FS) unsafeAddFile(dir *dirNode, fileName string, data Node) error

func (m *FS) getPath(op string, origPath string) (Node, error) {
	path, ok := fs.Clean(origPath)
	if !ok {
		return nil, errInvalid(op, path)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	fi, ok := m.unsafeGet(path)
	if !ok {
		return nil, errNotExist(op, path)
	}

	return fi, nil
}

func (m *FS) unsafeGet(path string) (Node, bool) {
	tn, ok := m.t.Find(path)
	if !ok {
		return nil, false
	}

	node := tn.Meta()
	return node, node != nil
}

func (m *FS) unsafeAdd(path string, node Node) {
	m.t.Add(path, node)
}

func unsafeJoin(dir, file string) string {
	switch {
	case dir == ".":
		return file
	case file == ".":
		return dir
	default:
		return dir + "/" + file
	}
}
