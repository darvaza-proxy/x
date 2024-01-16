package embed

import (
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"time"
)

// interface assertions
var (
	_ fs.FileInfo       = (*FileInfo)(nil)
	_ fs.DirEntry       = (*FileInfo)(nil)
	_ fs.File           = (*FileDescriptor)(nil)
	_ io.ReadSeekCloser = (*FileDescriptor)(nil)
	_ Embedded          = (*File)(nil)
)

// FileInfo describes a File.
type FileInfo struct {
	sys *File
}

// IsDir indicates the [File] is not a directory, shortcut
// for Mode().IsDir().
func (FileInfo) IsDir() bool { return false }

// Type indicates the [File] is a regular file.
func (FileInfo) Type() fs.FileMode { return 0 }

// Mode returns file mode bits of a read-only file.
func (FileInfo) Mode() fs.FileMode { return fs.FileMode(0444) }

// Name returns the base name of the file
func (fi FileInfo) Name() string { return path.Base(fi.sys.Name) }

// ModTime returns when was the file was last modified.
func (fi FileInfo) ModTime() time.Time { return fi.sys.ModTime }

// Size indicates the size of the file unencoded.
func (fi FileInfo) Size() int64 { return fi.sys.Size }

// Sys returns the underlying [File].
func (fi FileInfo) Sys() any { return fi.sys }

// Info returns the FileInfo for the [File], required by
// [fs.DirEntry].
func (fi FileInfo) Info() (fs.FileInfo, error) {
	return fi, nil
}

// ContentType indicates the MIME type of the [File].
func (fi FileInfo) ContentType() string { return fileContentType(fi.sys) }

// ContentEncoding indicates how the [File] is encoded, like "gzip".
func (fi FileInfo) ContentEncoding() string { return fileContentEncoding(fi.sys) }

// ContentDigests provides a list of checksums for the unencoded contents.
func (fi FileInfo) ContentDigests() []string { return fileContentDigests(fi.sys) }

// File contains an embedded file's content and details.
type File struct {
	Name    string
	ModTime time.Time
	Size    int64

	Content         []byte
	ContentType     string
	ContentEncoding string
	ContentDigests  []string
}

// Info returns the FileInfo for the [File].
func (file *File) Info() (fs.FileInfo, error) {
	return FileInfo{file}, nil
}

// Open returns a reader for the contents of file, encoded
// if ContentEncoding() isn't "identity".
func (file *File) Open() (fs.File, error) {
	fd := &FileDescriptor{
		ReadSeeker: bytes.NewReader(file.Content),
		f:          file,
	}
	return fd, nil
}

func fileContentType(file *File) string {
	if file != nil {
		// predefined
		ct := file.ContentType
		if ct != "" {
			return ct
		}

		// by extension
		ct = mime.TypeByExtension(path.Ext(file.Name))
		if ct != "" {
			// remember and return
			file.ContentType = ct
			return ct
		}

		// by content
		ct = http.DetectContentType(file.Content)

		// remember and return
		file.ContentType = ct
		return ct
	}

	return ""
}

func fileContentEncoding(file *File) string {
	if file != nil {
		// predefined
		ce := file.ContentEncoding
		if ce == "" {
			ce = "identity"
			file.ContentEncoding = ce
		}
		return ce
	}

	return ""
}

func fileContentDigests(file *File) []string {
	if file != nil {
		cd := file.ContentDigests
		if len(cd) > 0 {
			// predefined
			return cd
		}

		// calculate sha512
		sum, ok := fileContentSum512(file, fileContentEncoding(file))
		if ok {
			hash := base64.RawStdEncoding.EncodeToString(sum)
			cd = []string{"sha512:" + hash}

			// remember and return
			file.ContentDigests = cd
			return cd
		}
	}

	return []string{}
}

func fileContentSum512(file *File, encoding string) ([]byte, bool) {
	var r io.Reader

	switch encoding {
	case "identity", "":
		// raw
		r = bytes.NewReader(file.Content)
	case "gzip":
		// decompress gzip
		f, err := gzip.NewReader(bytes.NewReader(file.Content))
		if err == nil {
			r = f
		}
	}

	// TODO: other formats?

	if r != nil {
		s := sha512.New()
		if _, err := io.Copy(s, r); err == nil {
			return s.Sum(nil), true
		}
	}

	return nil, false
}

// FileDescriptor provides a ReadSeekCloser for the [File].
type FileDescriptor struct {
	io.ReadSeeker
	f *File
}

// Close closes the Reader.
func (fd *FileDescriptor) Close() error {
	fd.ReadSeeker = nil
	fd.f = nil
	return nil
}

// Stat returns the FileInfo for the [File].
func (fd *FileDescriptor) Stat() (fs.FileInfo, error) {
	return fd.f.Info()
}

// NewFromFS creates a [File] by reading one from a given filesystem
func NewFromFS(fsys fs.FS, name string) (*File, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}

	out, err := NewFromFile(f, name)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// NewFromFile creates a [File] from a given [fs.File].
func NewFromFile(f fs.File, name string) (*File, error) {
	var buf bytes.Buffer

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size, err := buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = fi.Name()
	}

	out := &File{
		Name:    name,
		ModTime: fi.ModTime(),
		Size:    size,

		Content: buf.Bytes(),
	}

	if p, ok := f.(interface {
		ContentType() string
	}); ok {
		out.ContentType = p.ContentType()
	}

	if p, ok := f.(interface {
		ContentEncoding() string
	}); ok {
		out.ContentEncoding = p.ContentEncoding()
	}

	if p, ok := f.(interface {
		ContentDigests() []string
	}); ok {
		out.ContentDigests = p.ContentDigests()
	}

	// compute hashes
	_ = fileContentEncoding(out)
	_ = fileContentDigests(out)
	_ = fileContentType(out)

	// return
	return out, nil
}
