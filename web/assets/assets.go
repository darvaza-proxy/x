// Package assets assists serving embedded assets via HTTP
package assets

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
	"darvaza.org/x/web/qlist"
)

var _ http.Handler = (*AssetHandler)(nil)

// Asset represents an in-memory object that doesn't change
type Asset interface {
	Content() io.ReadSeeker
}

// AssetHandler serves an Asset.
// Only Content() is required, but the object is also tested
// for:
//
// * ContentType() string
// * ETags() []string
// * Stat() (fs.FileInfo, error)
// * Info() (fs.FileInfo, error)
// * ModTime() time.Time
// * Header() http.Header
//
// AssetHandler supports:
// * 405
// * 406
// * ETag
// * Last-Modified
// * Ranges
type AssetHandler struct {
	Asset Asset

	mu       sync.Mutex
	ct       string
	parsedCT qlist.QualityList
}

func (h *AssetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch strings.ToUpper(req.Method) {
	case consts.GET, consts.HEAD:
		// GET
		h.handleGet(rw, req)
	case consts.OPTIONS:
		// OPTIONS
		h.setAllowed(rw.Header())
		rw.WriteHeader(http.StatusNoContent)
	default:
		// 405
		h.setAllowed(rw.Header())
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (*AssetHandler) setAllowed(hdr http.Header) {
	hdr[consts.Allow] = []string{"OPTIONS, GET, HEAD"}
}

func (h *AssetHandler) handleGet(rw http.ResponseWriter, req *http.Request) {
	var code int

	accepted, err := qlist.ParseMediaRangeHeader(req.Header)
	if err != nil {
		// 400
		code = http.StatusBadRequest
	} else if ok, err := h.isAcceptable(accepted); err != nil {
		// 500, bad content type
		code = http.StatusInternalServerError
	} else if !ok {
		// 406
		code = http.StatusNotAcceptable
	} else {
		// 200
		h.serveContent(rw, req)
		return
	}
	rw.WriteHeader(code)
}

func (h *AssetHandler) isAcceptable(accepted qlist.QualityList) (bool, error) {
	ql, err := h.getContentTypeParsed()
	switch {
	case err != nil:
		// bad
		return false, err
	case len(ql) == 0:
		// unknown
		return true, nil
	default:
		_, _, ok := qlist.BestQualityParsed(ql, accepted)
		return ok, nil
	}
}

func (h *AssetHandler) serveContent(rw http.ResponseWriter, req *http.Request) {
	hdr := rw.Header()
	h.copyHeaders(hdr)
	h.setContentType(hdr)
	h.setETags(hdr)

	modTime, _ := getModTime(h.Asset)
	content := h.Asset.Content()

	// at the end, Close if we can
	if f, ok := content.(io.Closer); ok {
		defer unsafeClose(f)
	}

	http.ServeContent(rw, req, "", modTime, content)
}

func (h *AssetHandler) copyHeaders(dest http.Header) {
	// copy headers
	if x, ok := h.Asset.(interface {
		Header() http.Header
	}); ok {
		for k, vv := range x.Header() {
			dest[k] = core.SliceCopy(vv)
		}
	}
}

func (h *AssetHandler) setETags(dest http.Header) {
	if !headerExist(dest, consts.ETag) {
		tags, _ := getETags(h.Asset)
		if len(tags) > 0 {
			dest[consts.ETag] = tags
		}
	}
}

func (h *AssetHandler) setContentType(dest http.Header) {
	if !headerExist(dest, consts.ContentType) {
		if s, ok := h.getContentType(); ok {
			dest[consts.ContentType] = []string{s}
		}
	}
}

func (h *AssetHandler) getContentTypeParsed() (qlist.QualityList, error) {
	if err := h.init(); err != nil {
		return nil, err
	}

	return h.parsedCT, nil
}

func (h *AssetHandler) getContentType() (string, bool) {
	if err := h.init(); err != nil {
		return "", false
	}
	return h.ct, h.ct != ""
}

func (h *AssetHandler) init() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.parsedCT == nil {
		ct, err := getContentType(h.Asset, "")
		switch {
		case err != nil:
			return err
		case ct == "":
			h.parsedCT = []qlist.QualityValue{}
		default:
			qv, err := qlist.ParseQualityValue(ct)
			if err != nil {
				return err
			}

			h.parsedCT = []qlist.QualityValue{qv}
		}
	}

	return nil
}
