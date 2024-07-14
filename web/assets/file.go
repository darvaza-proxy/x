package assets

import (
	"io"
	"io/fs"
	"net/http"

	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
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
	if err := setETag(rw.Header(), file); err != nil {
		serve500(rw, req, err)
		return
	}

	// TODO: Content-Encoding

	http.ServeContent(rw, req, name, fi.ModTime(), file)
}

func setContentType(hdr http.Header, file File, name string) error {
	_, haveType := hdr[consts.ContentType]
	if !haveType {
		cType, err := getContentType(file, name)
		switch {
		case err != nil:
			return err
		case cType != "":
			hdr[consts.ContentType] = []string{cType}
		}
	}
	return nil
}

func getContentType(file File, name string) (string, error) {
	if ct := ContentType(file); ct != "" {
		// self-describing
		return ct, nil
	}

	// infer from file extension
	ct := TypeByFilename(name)
	if ct == "" {
		var err error

		// detect from content
		ct, err = TypeBySniffing(file)
		if err != nil {
			return "", err
		}
	}

	if ct != "" {
		// try to remember
		if f := getContentTypeSetter(file); f != nil {
			f.SetContentType(ct)
		}
	}

	return ct, nil
}

func getContentTypeSetter(file fs.File) ContentTypeSetter {
	if f, ok := file.(ContentTypeSetter); ok {
		return f
	}

	if fi, _ := file.Stat(); fi != nil {
		if f, ok := fi.(ContentTypeSetter); ok {
			return f
		}
	}

	return nil
}

func setETag(hdr http.Header, file File) error {
	_, haveETag := hdr[consts.ETag]
	if !haveETag {
		tags, err := getETags(file)
		switch {
		case err != nil:
			return err
		case len(tags) > 0:
			hdr[consts.ETag] = tags
		}
	}
	return nil
}

func getETags(file File) ([]string, error) {
	if tags := ETags(file); len(tags) > 0 {
		return tags, nil
	}

	hash, err := BLAKE3SumFile(file)
	if err != nil {
		return nil, err
	}

	tags := []string{hash}

	// try to remember
	if f := getETagsSetter(file); f != nil {
		f.SetETags(tags...)
	}

	return tags, nil
}

func getETagsSetter(file fs.File) ETagsSetter {
	if f, ok := file.(ETagsSetter); ok {
		return f
	}

	if fi, _ := file.Stat(); fi != nil {
		if f, ok := fi.(ETagsSetter); ok {
			return f
		}
	}

	return nil
}

func serve500(rw http.ResponseWriter, req *http.Request, err error) {
	h := &web.HTTPError{
		Code: http.StatusInternalServerError,
		Err:  err,
	}

	h.ServeHTTP(rw, req)
}
