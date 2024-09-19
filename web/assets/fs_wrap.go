package assets

import (
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/fs"
	"darvaza.org/x/web"
)

var (
	_ FS            = (*WrapFS)(nil)
	_ fs.FS         = (*WrapFS)(nil)
	_ fs.ReadFileFS = (*WrapFS)(nil)
	_ fs.StatFS     = (*WrapFS)(nil)
	_ fs.SubFS      = (*WrapFS)(nil)
	_ httpView      = (*WrapFS)(nil)
)

// WrapFS is an assets [FS] that tests multiple layers
// and name variants.
type WrapFS struct {
	layers []fs.FS
	globs  [][]fs.Matcher
	names  []func(string) []string
	root   string

	resolver func(*http.Request) (string, error)
}

func (o *WrapFS) forEachLayer(h func(fs.FS) bool) {
	switch {
	case o == nil, h == nil, len(o.layers) == 0:
		return
	default:
		for _, sl := range o.layers {
			if !h(sl) {
				return
			}
		}
	}
}

func (o *WrapFS) forEachName(name string, h func(string) bool) {
	switch {
	case o == nil, name == "", h == nil:
		return
	case len(o.names) == 0:
		o.forName(name, h)
	default:
		o.doForEachName(name, h)
	}
}

func (o *WrapFS) doForEachName(name string, h func(string) bool) {
	m := make(map[string]struct{})
	for _, fn := range o.names {
		for _, s := range fn(name) {
			if !o.doForNameOnce(s, h, m) {
				return
			}
		}
	}
}

func (o *WrapFS) doForNameOnce(name string, h func(string) bool, m map[string]struct{}) bool {
	if _, dupe := m[name]; dupe {
		// don't interrupt the loop when we find a duplicate name
		return true
	}

	m[name] = struct{}{}
	return o.forName(name, h)
}

func (o *WrapFS) forEachNameLayer(name string, h func(fs.FS, string) bool) {
	ok := true
	o.forEachName(name, func(s string) bool {
		o.forEachLayer(func(p fs.FS) bool {
			if !h(p, s) {
				ok = false
			}
			return ok
		})
		return ok
	})
}

func (o *WrapFS) forName(name string, h func(string) bool) bool {
	// must match all sets, if any.
	for _, gg := range o.globs {
		if !Match(name, gg) {
			// skip, but don't break.
			return true
		}
	}

	// acceptable
	return h(unsafeJoin(o.root, name))
}

// Open implements [fs.FS], returning the first successful match.
func (o *WrapFS) Open(name string) (fs.File, error) {
	var out fs.File
	var err error

	o.forEachNameLayer(name, func(fSys fs.FS, fName string) bool {
		out, err = fSys.Open(fName)
		return out == nil
	})

	switch {
	case out != nil:
		return out, nil
	case err == nil:
		err = fs.ErrNotExist
	}

	return nil, newPathError("open", name, err)
}

// Stat implements [fs.StatFS], returning the first successful match.
func (o *WrapFS) Stat(name string) (fs.FileInfo, error) {
	var out fs.FileInfo
	var err error

	o.forEachNameLayer(name, func(fSys fs.FS, fName string) bool {
		out, err = fs.Stat(fSys, fName)
		return out == nil
	})

	switch {
	case out != nil:
		return out, nil
	case err == nil:
		err = fs.ErrNotExist
	}

	return nil, newPathError("stat", name, err)
}

// ReadFile implements [fs.ReadFileFS], returning the first successful match.
func (o *WrapFS) ReadFile(name string) ([]byte, error) {
	var out []byte
	var found bool
	var err error

	o.forEachNameLayer(name, func(fSys fs.FS, fName string) bool {
		out, err = fs.ReadFile(fSys, fName)
		if err == nil {
			found = true
		}
		return !found
	})

	switch {
	case found:
		return out, nil
	case err == nil:
		err = fs.ErrNotExist
	}

	return nil, newPathError("readfile", name, err)
}

// Sub implements [fs.Sub] returning a partial view of the [FS]. When using "."
// a new view is created without name options.
func (o *WrapFS) Sub(dir string) (fs.FS, error) {
	root, ok := fs.Clean(dir)
	if !ok {
		return nil, newErrInvalid("readdir", dir)
	}

	return o.unsafeSub(root), nil
}

func (o *WrapFS) unsafeSub(dir string) *WrapFS {
	root := o.root
	if dir != "." {
		root = unsafeJoin(root, dir)
	}

	return &WrapFS{
		layers:   core.SliceCopy(o.layers),
		globs:    core.SliceCopy(o.globs),
		root:     root,
		resolver: o.resolver,
	}
}

// ServeHTTP directly implements the [http.Handler] interface
func (o *WrapFS) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	serveHTTP(o, rw, req, nil)
}

// Middleware returns a middleware handler that will attempt to serve files from
// the wrapped layers, or proceed if not found.
func (o *WrapFS) Middleware() func(http.Handler) http.Handler {
	return web.NewMiddleware(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		serveHTTP(o, rw, req, next)
	})
}

// SetResolver specifies a custom name resolver to use on HTTP requests to this FS
func (o *WrapFS) SetResolver(r func(*http.Request) (string, error)) {
	if r == nil {
		r = DefaultRequestResolver
	}

	o.resolver = r
}

func (o *WrapFS) getResolver() func(*http.Request) (string, error) {
	return o.resolver
}

// getFileHandler implements the [httpView] interface returns the first named file if
// valid and existing.
func (o *WrapFS) getFileHandler(name string) http.Handler {
	var out http.Handler

	o.forEachNameLayer(name, func(fSys fs.FS, fName string) bool {
		if f, ok := o.newFileHandler(fSys, fName); ok {
			out = f
			return false
		}
		return true
	})

	return out
}

// WithLayer adds fs.FS layers to be tried before those already specified.
// layers are listed bottom up. Meaning the last to be checked is the first listed.
func (o *WrapFS) WithLayer(layers ...fs.FS) *WrapFS {
	layers = core.SliceReversedFn(layers, func(_ []fs.FS, sl fs.FS) (fs.FS, bool) {
		return sl, sl != nil
	})

	if len(layers) > 0 {
		layers = append(layers, o.layers...)
		o.layers = layers
	}

	return o
}

// WithIndex adds index options to serve when the requested name isn't available.
func (o *WrapFS) WithIndex(ss ...string) *WrapFS {
	if len(o.names) == 0 {
		// first try the requested name
		o.names = append(o.names, func(fn string) []string {
			return []string{fn}
		})
	}

	ss = core.SliceCopyFn(ss, func(_ []string, s string) (string, bool) {
		return fs.Clean(s)
	})

	if len(ss) > 0 {
		o.names = append(o.names, newIndexFunc(ss))
	}

	return o
}

// WithTry adds functions that provide alternative names to search for.
func (o *WrapFS) WithTry(funcs ...func(string) []string) *WrapFS {
	for _, fn := range funcs {
		if fn != nil {
			o.names = append(o.names, fn)
		}
	}

	return o
}

func newIndexFunc(ss []string) func(string) []string {
	return func(fn string) []string {
		if fn == "." {
			return core.SliceCopy(ss)
		}

		return core.SliceCopyFn(ss, func(_ []string, s string) (string, bool) {
			return unsafeJoin(fn, s), true
		})
	}
}

// NewWrapFS creates a [WrapFS] trying on all the given filesystem options.
// layers are listed bottom up. Meaning the last to be checked is the first listed.
func NewWrapFS(layers ...fs.FS) (*WrapFS, error) {
	layers = core.SliceReversedFn(layers, func(_ []fs.FS, sl fs.FS) (fs.FS, bool) {
		return sl, sl != nil
	})

	out := &WrapFS{
		root:     ".",
		layers:   layers,
		resolver: DefaultRequestResolver,
	}

	if len(out.layers) == 1 {
		if v, ok := out.layers[0].(*WrapFS); ok {
			// pass through
			return v, nil
		}
	}

	return out, nil
}

// unsafeNewWrapSubFS extends [NewFS] for arbitrary file systems supporting the [fs.SubFS]
// interface.
func unsafeNewWrapSubFS(v fs.SubFS, root string, gg []fs.Matcher) (*WrapFS, error) {
	var fSys fs.FS

	if root == "." {
		fSys = v
	} else if sub, err := v.Sub(root); err == nil {
		fSys = sub
	} else {
		return nil, err
	}

	return unsafeNewWrapFS(fSys, ".", gg)
}

func unsafeNewWrapFS(fSys fs.FS, root string, gg []fs.Matcher) (*WrapFS, error) {
	var globs [][]fs.Matcher
	if len(gg) > 0 {
		globs = append(globs, gg)
	}

	out := &WrapFS{
		layers: []fs.FS{fSys},
		globs:  globs,
		root:   root,
	}

	return out, nil
}

// MustWrapFS is equivalent to [NewWrapFS] but panics on error
func MustWrapFS(layers ...fs.FS) *WrapFS {
	out, err := NewWrapFS(layers...)
	if err != nil {
		panic(err)
	}
	return out
}
