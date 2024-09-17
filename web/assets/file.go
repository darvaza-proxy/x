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

func setContentType(hdr http.Header, file any, name string) error {
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

func getContentType(file any, name string) (string, error) {
	if ct := ContentType(file); ct != "" {
		// self-describing
		return ct, nil
	}

	// infer from file extension
	ct := TypeByFilename(name)
	if ct == "" {
		var err error

		// detect from content
		ct, err = sniffContentType(file)
		if err != nil {
			return "", err
		}
	}

	if ct != "" {
		// try to remember
		if f, ok := getContentTypeSetter(file); ok {
			f.SetContentType(ct)
		}
	}

	return ct, nil
}

func sniffContentType(v any) (string, error) {
	if file, ok := tryReadSeeker(v); ok {
		return TypeBySniffing(file)
	}

	return "", nil
}

func getContentTypeSetter(v any) (ContentTypeSetter, bool) {
	if f, ok := tryContentTypeSetter(v); ok {
		return f, true
	}

	if fi, ok := tryStat(v); ok {
		if f, ok := tryContentTypeSetter(fi); ok {
			return f, true
		}
	}

	return nil, false
}

func setETag(hdr http.Header, file any) error {
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

func getETags(file any) ([]string, error) {
	if tags := ETags(file); len(tags) > 0 {
		return tags, nil
	}

	hash, err := sniffBLAKE3Sum(file)
	switch {
	case err != nil:
		return nil, err
	case hash == "":
		return nil, nil
	}

	tags := []string{hash}

	// try to remember
	if f, _ := getETagsSetter(file); f != nil {
		f.SetETags(tags...)
	}

	return tags, nil
}

func sniffBLAKE3Sum(v any) (string, error) {
	if f, ok := tryReadSeeker(v); ok {
		return BLAKE3SumFile(f)
	}

	return "", nil
}

func getETagsSetter(v any) (ETagsSetter, bool) {
	if f, ok := tryETagsSetter(v); ok {
		return f, true
	}

	if fi, ok := tryStat(v); ok {
		if f, ok := tryETagsSetter(fi); ok {
			return f, true
		}
	}

	return nil, false
}

func getFileName(v any) (string, bool) {
	if fi, ok := tryStat(v); ok {
		return fi.Name(), true
	}
	return "", false
}

func serve500(rw http.ResponseWriter, req *http.Request, err error) {
	h := &web.HTTPError{
		Code: http.StatusInternalServerError,
		Err:  err,
	}

	h.ServeHTTP(rw, req)
}
