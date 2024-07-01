package assets

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/fs"
	"darvaza.org/x/web"
)

var (
	_ FS            = (*EmbedFS)(nil)
	_ fs.GlobFS     = (*EmbedFS)(nil)
	_ fs.ReadFileFS = (*EmbedFS)(nil)
	_ fs.SubFS      = (*EmbedFS)(nil)

	_ fs.FileInfo = (*EmbedMeta)(nil)
	_ fs.DirEntry = (*EmbedMeta)(nil)

	_ fs.File           = (*EmbedFile)(nil)
	_ io.ReadSeekCloser = (*EmbedFile)(nil)
	_ http.Handler      = (*EmbedFile)(nil)
)

// embedFS is the data shared by all views of the given [embed.FS]
type embedFS struct {
	*embed.FS
	sync.Mutex

	files []*embedMeta
}

// unsafeAddFile registers a static file in the file system.
func (o *embedFS) unsafeAddFile(fi fs.FileInfo, name, path string) *embedMeta {
	fm := &embedMeta{
		FileInfo: fi,
		name:     name,
		path:     path,
	}

	o.files = append(o.files, fm)
	return fm
}

// EmbedFS extends [embed.FS] for serving static assets.
type EmbedFS struct {
	base *embedFS
	root string

	files    map[string]*EmbedMeta
	resolver func(*http.Request) (string, error)
}

func (o *EmbedFS) lock()   { o.base.Lock() }
func (o *EmbedFS) unlock() { o.base.Unlock() }

// subPath returns the root-relative path of a file
func (o *EmbedFS) subPath(realPath string) string {
	if o.root == "." {
		return realPath
	}

	return realPath[len(o.root)+1:]
}

// getPath returns the root-relative path of a file
func (o *EmbedFS) getPath(fm *EmbedMeta) string {
	return o.subPath(fm.embedMeta.Path())
}

// getFile attempts to find a file in the file-system, and
// return [fs.PathError] in case of errors.
func (o *EmbedFS) getFile(op, path string) (*EmbedMeta, error) {
	if !fs.ValidPath(path) {
		return nil, &fs.PathError{
			Op:   op,
			Path: path,
			Err:  fs.ErrInvalid,
		}
	}

	fm, ok := o.files[path]
	if !ok {
		return nil, &fs.PathError{
			Op:   op,
			Path: path,
			Err:  fs.ErrNotExist,
		}
	}

	return fm, nil
}

func (o *EmbedFS) unsafeAddFile(sm *embedMeta, path string) {
	fm := &EmbedMeta{
		embedMeta: sm,
		fs:        o,
	}

	o.files[path] = fm
}

// NewEmbedFS creates an assets FS from the given [embed.FS], optional root, and an optional
// list of patterns. [github.com/gobwas/glob] patterns supported.
func NewEmbedFS(base *embed.FS, root string, patterns ...string) (*EmbedFS, error) {
	if base == nil {
		base = &embed.FS{} // empty
	}

	dir, ok := fs.Clean(root)
	if !ok {
		err := &fs.PathError{
			Op:   "readdir",
			Path: root,
			Err:  fs.ErrInvalid,
		}
		return nil, err
	}

	o := &EmbedFS{
		base: &embedFS{
			FS: base,
		},
		root:  dir,
		files: make(map[string]*EmbedMeta),
	}

	if err := o.init(patterns...); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *EmbedFS) init(patterns ...string) error {
	g, err := fs.GlobCompile(patterns...)
	if err != nil {
		return err
	}

	check := func(path string, di fs.DirEntry) bool {
		if di.Type().IsRegular() {
			return o.initAddFile(path, di, g)
		}

		return false
	}

	_, err = fs.MatchFunc(o.base, o.root, check)
	return err
}

func (o *EmbedFS) initAddFile(realPath string, di fs.DirEntry, globs []fs.Matcher) bool {
	var match bool

	path := o.subPath(realPath)
	for _, g := range globs {
		if g.Match(path) {
			match = true
			break
		}
	}

	if len(globs) > 0 && !match {
		// skip
		return false
	}

	fi, err := di.Info()
	if err != nil {
		core.PanicWrap(err, "Info:%q", path)
	}

	_, name := fs.Split(path)

	sm := o.base.unsafeAddFile(fi, name, realPath)
	o.unsafeAddFile(sm, path)
	return true
}

// Open opens the named file from the file system.
func (o *EmbedFS) Open(name string) (fs.File, error) {
	fm, err := o.getFile("stat", name)
	if err != nil {
		return nil, err
	}

	return fm.newFile(), nil
}

// Stat returns a [fs.FileInfo] describing the named file from the file system.
func (o *EmbedFS) Stat(name string) (fs.FileInfo, error) {
	fm, err := o.getFile("stat", name)
	if err != nil {
		return nil, err
	}

	return fm, nil
}

// Sub creates a view of the file system restricted to a particular
// sub-directory.
func (o *EmbedFS) Sub(root string) (fs.FS, error) {
	s, ok := fs.Clean(root)
	if !ok {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: root,
			Err:  fs.ErrInvalid,
		}
	}

	if s == "." {
		// NO-OP
		return o, nil
	}

	o.lock()
	defer o.unlock()

	return o.unsafeSub(s)
}

func (o *EmbedFS) unsafeSub(root string) (*EmbedFS, error) {
	o2 := &EmbedFS{
		base:     o.base,
		root:     unsafeJoin(o.root, root),
		resolver: o.resolver,
		files:    make(map[string]*EmbedMeta),
	}

	prefix := o2.root + "/"
	for _, m1 := range o.files {
		path, ok := strings.CutPrefix(m1.path, prefix)
		if ok {
			o2.unsafeAddFile(m1.embedMeta, path)
		}
	}

	if len(o2.files) > 0 {
		return o2, nil
	}

	return nil, &fs.PathError{
		Op:   "readdir",
		Path: root,
		Err:  fs.ErrNotExist,
	}
}

// Glob returns a list of files matching the given pattern.
// [github.com/gobwas/glob] patterns supported.
func (o *EmbedFS) Glob(pattern string) ([]string, error) {
	g, err := fs.GlobCompile(pattern)
	if err != nil {
		return nil, err
	}

	return o.match(g), nil
}

func (o *EmbedFS) match(globs []fs.Matcher) []string {
	out := make([]string, 0, len(o.files))

	for path := range o.files {
		if isGlobMatch(path, globs) {
			out = append(out, path)
		}
	}

	return out
}

// ReadFile reads the named file and returns its contents.
func (o *EmbedFS) ReadFile(name string) ([]byte, error) {
	fm, err := o.getFile("read", name)
	if err != nil {
		return nil, err
	}

	return o.doReadFile(fm.path)
}

func (o *EmbedFS) doReadFile(realPath string) ([]byte, error) {
	return o.base.ReadFile(realPath)
}

// SetResolver provides an optional function to identify the requested resource name.
func (o *EmbedFS) SetResolver(r func(*http.Request) (string, error)) {
	o.lock()
	defer o.unlock()

	o.resolver = r
}

func (o *EmbedFS) getResolver() func(*http.Request) (string, error) {
	o.lock()
	defer o.unlock()

	if r := o.resolver; r != nil {
		return r
	}

	return DefaultRequestResolver
}

// ServeHTTP directly implements the [http.Handler] interface
func (o *EmbedFS) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	o.doServeHTTP(rw, req, nil)
}

// Middleware returns a middleware handler to be used with this [EmbedFS],
// allowing us to proceed when the requested file wasn't found.
func (o *EmbedFS) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, req *http.Request) {
			o.doServeHTTP(rw, req, next)
		}

		return http.HandlerFunc(fn)
	}
}

func (o *EmbedFS) doServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	var h http.Handler

	fm, err := o.getFileFromRequest(req)
	switch {
	case fm != nil:
		h = fm
	case err.Status() != http.StatusNotFound:
		h = err
	case next != nil:
		h = next
	default:
		h = errNotFound
	}

	h.ServeHTTP(rw, req)
}

func (o *EmbedFS) getFileFromRequest(req *http.Request) (*EmbedMeta, *web.HTTPError) {
	r := o.getResolver()
	path, err := r(req)

	if err == nil && path != "" && path[0] == '/' {
		path, ok := fs.Clean(path[1:])
		if ok {
			fm, _ := o.getFile("stat", path)
			if fm != nil {
				return fm, nil
			}

			return nil, errNotFound
		}
	}

	return nil, errBadRequest
}

// EmbedFile is a readable instance of an embedded static file.
type EmbedFile struct {
	mu     sync.Mutex
	meta   *EmbedMeta
	reader io.ReadSeeker
}

func (fd *EmbedFile) getReader() (io.ReadSeeker, error) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if fd.reader == nil {
		r, err := fd.meta.newReader()
		if err != nil {
			return nil, err
		}

		fd.reader = r
	}

	return fd.reader, nil
}

// Stat implements the [fs.File] interface returning the [EmbedMeta] associated
// with this descriptor.
func (fd *EmbedFile) Stat() (fs.FileInfo, error) { return fd.meta, nil }

// Close implements the [fs.File] interface, but doesn't really do anything.
func (fd *EmbedFile) Close() error {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	fd.reader = newClosedReader()
	return nil
}

// Read reads up to len(b) bytes from the File and stores them in b. It returns
// the number of bytes read and any error encountered. At end of file,
// Read returns 0, io.EOF.
func (fd *EmbedFile) Read(b []byte) (int, error) {
	r, err := fd.getReader()
	if err != nil {
		return 0, err
	}

	return r.Read(b)
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means relative
// to the current offset, and 2 means relative to the end. It returns the new offset
// and an error, if any. The behavior of Seek on a file opened with O_APPEND is not
// specified.
func (fd *EmbedFile) Seek(offset int64, whence int) (int64, error) {
	r, err := fd.getReader()
	if err != nil {
		return 0, err
	}
	return r.Seek(offset, whence)
}

func (fd *EmbedFile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer fd.Close()

	name := fd.meta.Name()
	modTime := fd.meta.ModTime()

	http.ServeContent(rw, req, name, modTime, fd)
}

// embedFS is the File data shared by all views of the given [embed.FS]
type embedMeta struct {
	fs.FileInfo

	name string
	path string
}

func (fm *embedMeta) Name() string { return fm.name }
func (fm *embedMeta) Path() string { return fm.path }

// EmbedMeta contains all information we know about the embedded assets
type EmbedMeta struct {
	*embedMeta

	fs *EmbedFS
}

// Name returns the name of the [File] in the directory.
func (fm *EmbedMeta) Name() string { return fm.name }

// Path returns the full path to the [File] in the base [fs.FS]
func (fm *EmbedMeta) Path() string { return fm.fs.getPath(fm) }

// Info implements the [fs.DirEntry] interface returning itself as [fs.FileInfo]
func (fm *EmbedMeta) Info() (fs.FileInfo, error) { return fm, nil }

// Type implements the [fs.DirEntry] interface as an alias of [EmbedMeta.Mode].
func (fm *EmbedMeta) Type() fs.FileMode { return fm.Mode() }

func (fm *EmbedMeta) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fd := fm.newFile()
	fd.ServeHTTP(rw, req)
}

func (fm *EmbedMeta) newFile() *EmbedFile {
	return &EmbedFile{
		meta: fm,
	}
}

func (fm *EmbedMeta) newReader() (*bytes.Reader, error) {
	b, err := fm.fs.doReadFile(fm.path)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)
	return r, nil
}
