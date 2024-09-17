package assets

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"sync"

	"darvaza.org/core"
)

var (
	_ Asset         = (*wrapFSFile)(nil)
	_ fs.File       = (*wrapFSFile)(nil)
	_ io.ReadSeeker = (*wrapFSFile)(nil)
	_ io.WriterTo   = (*wrapFSFile)(nil)
	_ http.Handler  = (*wrapFSFile)(nil)
)

type wrapFSFile struct {
	mu sync.Mutex
	r  *bytes.Reader

	h  AssetHandler
	fs fs.FS
	fi fs.FileInfo
	fn string
}

func (*WrapFS) newFileHandler(fSys fs.FS, fn string) (*wrapFSFile, bool) {
	fi, err := fs.Stat(fSys, fn)
	switch {
	case err != nil, fi == nil:
		return nil, false
	case fi.Mode().IsRegular():
		out := &wrapFSFile{
			fs: fSys,
			fi: fi,
			fn: fn,
		}
		out.h.Asset = out
		return out, true
	default:
		return nil, false
	}
}

// WriteTo writes the contents of the file to a given
// [io.Writer] implementing the [io.WriterTo].
func (f *wrapFSFile) WriteTo(w io.Writer) (int64, error) {
	r, err := f.reader()
	if err != nil {
		return 0, err
	}

	return r.WriteTo(w)
}

// Content implements the [Asset] interface returning the seekable
// buffer positioned at the start of the file.
func (f *wrapFSFile) Content() io.ReadSeeker {
	r, err := f.reader()
	if err != nil {
		return nil
	}

	// rewind if needed
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		unsafeClose(f)

		core.PanicWrap(err, "unreachable")
		return nil
	}
	return f
}

// Stat implements the [fs.File] interface returning itself
func (f *wrapFSFile) Stat() (fs.FileInfo, error) {
	return f.fi, nil
}

// Read implements the [fs.File] interface
func (f *wrapFSFile) Read(b []byte) (int, error) {
	r, err := f.reader()
	if err != nil {
		return 0, err
	}
	return r.Read(b)
}

// Seek implements the [io.ReadSeeker] interface
func (f *wrapFSFile) Seek(offset int64, whence int) (int64, error) {
	r, err := f.reader()
	if err != nil {
		return 0, err
	}
	return r.Seek(offset, whence)
}

// Close implements the [fs.File] interface
func (f *wrapFSFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.r = nil
	return nil
}

func (f *wrapFSFile) reader() (*bytes.Reader, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.r == nil {
		b, err := fs.ReadFile(f.fs, f.fn)
		if err != nil {
			return nil, err
		}
		f.r = bytes.NewReader(b)
	}
	return f.r, nil
}

func (f *wrapFSFile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f.h.ServeHTTP(rw, req)
}
