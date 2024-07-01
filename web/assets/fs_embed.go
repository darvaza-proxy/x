package assets

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"strings"
	"sync"

	"darvaza.org/x/fs"
)

var (
	_ FS            = (*EmbedFS)(nil)
	_ fs.GlobFS     = (*EmbedFS)(nil)
	_ fs.ReadFileFS = (*EmbedFS)(nil)
	_ fs.SubFS      = (*EmbedFS)(nil)
	_ httpView      = (*EmbedFS)(nil)

	_ fs.FileInfo = (*EmbedMeta)(nil)
	_ fs.DirEntry = (*EmbedMeta)(nil)

	_ File         = (*EmbedFile)(nil)
	_ http.Handler = (*EmbedFile)(nil)
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

// getFileHandler implements the [httpView] interface returning the named file
// if valid and existing.
func (o *EmbedFS) getFileHandler(path string) http.Handler {
	if fm, ok := o.files[path]; ok {
		return fm
	}
	return nil
}

// NewEmbedFS creates an assets FS from the given [embed.FS], optional root, and an optional
// list of patterns. [github.com/gobwas/glob] patterns supported.
func NewEmbedFS(base *embed.FS, root string, patterns ...string) (*EmbedFS, error) {
	dir, globs, err := sanitizeNewFS(root, patterns)
	if err != nil {
		return nil, err
	}

	return unsafeNewEmbedFS(base, dir, globs)
}

func unsafeNewEmbedFS(base *embed.FS, root string, globs []fs.Matcher) (*EmbedFS, error) {
	if base == nil {
		// no base file system, no root either.
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: root,
			Err:  fs.ErrNotExist,
		}
	}

	o := &EmbedFS{
		base: &embedFS{
			FS: base,
		},
		root:  root,
		files: make(map[string]*EmbedMeta),
	}

	if err := o.init(globs); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *EmbedFS) init(globs []fs.Matcher) error {
	var err error

	check := func(path string, di fs.DirEntry) bool {
		switch {
		case err != nil:
			// abort
			return false
		case !di.Type().IsRegular():
			// skip non-file entries
			return false
		default:
			fi, e := di.Info()
			if e != nil {
				err = &fs.PathError{
					Op:   "stat",
					Path: path,
					Err:  e,
				}
				// abort
				return false
			}

			return o.initAddFile(path, fi, globs)
		}
	}

	_, e := fs.MatchFunc(o.base, o.root, check)
	switch {
	case err != nil:
		// aborted
		return err
	case e != nil:
		// failed to walk the given root
		return e
	default:
		// success
		return nil
	}
}

// initAddFile is called for every regular file in the [embed.FS] during init restricted
// by the provided root to effectively include them in our new [EmbedFS] instance.
func (o *EmbedFS) initAddFile(realPath string, fi fs.FileInfo, globs []fs.Matcher) bool {
	path := o.subPath(realPath)

	if Match(path, globs) {
		_, name := fs.Split(path)
		sm := o.base.unsafeAddFile(fi, name, realPath)
		sm.unsafeAddToFS(o, path)
		return true
	}

	return false
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
// sub-directory. It will fail with [fs.ErrNotExist]
// if there are no files in it.
func (o *EmbedFS) Sub(dir string) (fs.FS, error) {
	root, ok := fs.Clean(dir)
	switch {
	case !ok:
		// bad directory
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: dir,
			Err:  fs.ErrInvalid,
		}
	case root == ".":
		// NO-OP
		return o, nil
	default:
		// create new view
		o.lock()
		defer o.unlock()

		return o.unsafeSub(root)
	}
}

func (o *EmbedFS) unsafeSub(root string) (*EmbedFS, error) {
	out := &EmbedFS{
		base:     o.base,
		root:     unsafeJoin(o.root, root),
		resolver: o.resolver,
		files:    make(map[string]*EmbedMeta),
	}

	prefix := out.root + "/"
	for _, f := range o.files {
		path, ok := strings.CutPrefix(f.Path(), prefix)
		if ok {
			f.unsafeAddToFS(out, path)
		}
	}

	if len(out.files) > 0 {
		return out, nil
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
		if Match(path, globs) {
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

	if r == nil {
		r = DefaultRequestResolver
	}

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
	serveHTTP(o, rw, req, nil)
}

// Middleware returns a middleware handler to be used with this [EmbedFS],
// allowing us to proceed when the requested file wasn't found.
func (o *EmbedFS) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, req *http.Request) {
			serveHTTP(o, rw, req, next)
		}

		return http.HandlerFunc(fn)
	}
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

	fd.reader = &fs.ClosedFile{
		FileInfo: fd.meta,
	}
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
	defer unsafeClose(fd)

	ServeFile(rw, req, fd)
}

// embedFS is the File data shared by all views of the given [embed.FS]
type embedMeta struct {
	fs.FileInfo

	name string
	path string
}

func (fm *embedMeta) Name() string { return fm.name }
func (fm *embedMeta) Path() string { return fm.path }

func (fm *embedMeta) unsafeAddToFS(o *EmbedFS, path string) {
	o.files[path] = &EmbedMeta{
		embedMeta: fm,
		fs:        o,
	}
}

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
