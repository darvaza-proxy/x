package assets

import (
	"io"
	"io/fs"
	"net/http"

	"darvaza.org/x/web"
)

// File represents seek-able [fs.File]
type File interface {
	fs.File
	io.ReadSeekCloser
}

// ServeFile serves the contents of the given [File].
func ServeFile(rw http.ResponseWriter, req *http.Request, file File) {
	fi, err := file.Stat()
	if err != nil {
		serve500(rw, req, err)
		return
	}

	name := fi.Name()

	// Content-Type
	if err := setContentType(rw.Header(), file, name); err != nil {
		serve500(rw, req, err)
		return
	}

	// ETag
	setETag(rw.Header(), file)

	// TODO: Content-Encoding

	http.ServeContent(rw, req, name, fi.ModTime(), file)
}

func setContentType(hdr http.Header, file File, name string) error {
	_, haveType := hdr["Content-Type"]
	if !haveType {
		cType, err := getContentType(file, name)
		switch {
		case err != nil:
			return err
		case cType != "":
			hdr["Content-Type"] = []string{cType}
		}
	}
	return nil
}

func getContentType(file File, name string) (string, error) {
	if ct := ContentType(file); ct != "" {
		// self-describing
		return ct, nil
	}

	if ct := TypeByFilename(name); ct != "" {
		// inferred from file extension
		return ct, nil
	}

	// detect from content
	return TypeBySniffing(file)
}

func setETag(hdr http.Header, file File) {
	_, haveETag := hdr["Etag"]
	if !haveETag {
		if tags := ETags(file); len(tags) > 0 {
			hdr["Etag"] = tags
		}
	}
}

func serve500(rw http.ResponseWriter, req *http.Request, err error) {
	h := &web.HTTPError{
		Code: http.StatusInternalServerError,
		Err:  err,
	}

	h.ServeHTTP(rw, req)
}
