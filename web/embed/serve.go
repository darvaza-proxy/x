package embed

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"time"
)

// ServeFile ...
func ServeFile(rw http.ResponseWriter, req *http.Request, f fs.File) error {
	var modTime time.Time

	hdr := rw.Header()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	// Last-Modified
	if t := fi.ModTime(); !t.IsZero() {
		modTime = t.UTC()
	}

	// Content-Type
	if p, ok := fi.(interface {
		ContentType() string
	}); ok {
		hdr.Set("Content-Type", p.ContentType())
	}

	// TODO: Content-Encoding

	r, ok := f.(io.ReadSeeker)
	if !ok {
		var buf bytes.Buffer

		_, err = io.Copy(&buf, f)
		if err != nil {
			return err
		}

		r = bytes.NewReader(buf.Bytes())
	}

	http.ServeContent(rw, req, fi.Name(), modTime, r)
	return nil
}
