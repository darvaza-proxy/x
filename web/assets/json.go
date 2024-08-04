package assets

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
	"darvaza.org/x/web/forms"
)

var (
	_ Asset        = (*jsonAsset)(nil)
	_ ContentTyped = (*jsonAsset)(nil)
	_ ETaged       = (*jsonAsset)(nil)
)

// DefaultJSONIndent is the indentation used by NewJSONAssetHandler
// when serializing the data.
const DefaultJSONIndent = "\t"

type jsonAsset struct {
	hdr  http.Header
	sum  string
	body []byte
}

func (jsonAsset) ContentType() string      { return consts.JSON }
func (o jsonAsset) Content() io.ReadSeeker { return bytes.NewReader(o.body) }
func (o jsonAsset) ETags() []string        { return []string{o.sum} }
func (o jsonAsset) Header() http.Header    { return o.hdr }

// NewJSONAssetHandler creates an [AssetHandler] serving a constant JSON resource.
func NewJSONAssetHandler[T any](d time.Duration, v T) (http.Handler, error) {
	body, err := json.MarshalIndent(v, "", DefaultJSONIndent)
	if err != nil {
		return nil, err
	}

	hash := BLAKE3Sum(body)

	hdr := make(http.Header)
	hdr[consts.ContentType] = []string{consts.JSON}
	hdr[consts.ContentLength] = []string{forms.FormatSigned(len(body), 10)}
	hdr[consts.ETag] = []string{hash}
	web.SetCache(hdr, d)

	h := &AssetHandler{
		Asset: jsonAsset{
			hdr:  hdr,
			sum:  hash,
			body: body,
		},
	}
	return h, nil
}

// MustJSONAssetHandler creates an [AssetHandler] serving a constant JSON resource,
// and panics if marshalling fails.
func MustJSONAssetHandler[T any](d time.Duration, v T) http.Handler {
	h, err := NewJSONAssetHandler(d, v)
	if err != nil {
		core.Panic(err)
	}
	return h
}
