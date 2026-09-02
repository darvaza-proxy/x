package assets_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/fs"
	"darvaza.org/x/web/assets"
	"darvaza.org/x/web/consts"
)

var (
	_ core.TestCase = assetHeaderTestCase{}
	_ core.TestCase = assetLastModifiedTestCase{}
)

// assetBody is the payload every fixture serves. Its content is
// immaterial; only the headers the handler derives are under test.
const assetBody = "asset body"

// assetModTime is the modification time the dated fixtures declare,
// fixed so the rendered Last-Modified value is predictable.
var assetModTime = time.Date(2026, time.March, 14, 15, 9, 26, 0, time.UTC)

// plainAsset is the least an [assets.AssetHandler] accepts: content
// and nothing else. It describes neither its headers nor its age.
type plainAsset struct{}

func (plainAsset) Content() io.ReadSeeker {
	return strings.NewReader(assetBody)
}

// headerAsset contributes headers of its own to the response.
type headerAsset struct {
	plainAsset
	hdr http.Header
}

func (a headerAsset) Header() http.Header { return a.hdr }

// modTimeAsset declares its age directly, the shape ModTimer selects.
type modTimeAsset struct {
	plainAsset
	modTime time.Time
}

func (a modTimeAsset) ModTime() time.Time { return a.modTime }

// statAsset describes itself the way an open file does, the shape
// Stater selects.
type statAsset struct {
	plainAsset
	info fs.FileInfo
	err  error
}

func (a statAsset) Stat() (fs.FileInfo, error) { return a.info, a.err }

// infoAsset describes itself the way a directory entry does, the
// shape Infoer selects.
type infoAsset struct {
	plainAsset
	info fs.FileInfo
	err  error
}

func (a infoAsset) Info() (fs.FileInfo, error) { return a.info, a.err }

// newFileInfo returns a real [fs.FileInfo] carrying the given
// modification time.
func newFileInfo(modTime time.Time) fs.FileInfo {
	fSys := fstest.MapFS{
		"asset.txt": &fstest.MapFile{
			Data:    []byte(assetBody),
			ModTime: modTime,
		},
	}

	return core.Must(fSys.Stat("asset.txt"))
}

// serveAsset runs one GET through an [assets.AssetHandler] and returns
// the recorded response. The request accepts anything, so the handler
// reaches the content rather than answering 406.
func serveAsset(t *testing.T, asset assets.Asset) *httptest.ResponseRecorder {
	t.Helper()

	h := &assets.AssetHandler{Asset: asset}

	req := httptest.NewRequest(consts.GET, "/", nil)
	req.Header.Set(consts.Accept, "*/*")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	core.AssertMustEqual(t, http.StatusOK, rw.Code, "status")

	return rw
}

// assetHeaderTestCase records whether an asset's own headers reach the
// response. An asset that declares none contributes none, leaving the
// field absent rather than blank.
type assetHeaderTestCase struct {
	asset assets.Asset
	name  string
	want  []string
}

func (tc assetHeaderTestCase) Name() string {
	return tc.name
}

func (tc assetHeaderTestCase) Test(t *testing.T) {
	t.Helper()

	rw := serveAsset(t, tc.asset)
	core.AssertSliceEqual(t, tc.want,
		rw.Header().Values(consts.RetryAfter), "Retry-After")
}

func newAssetHeaderTestCase(name string, asset assets.Asset,
	want []string) assetHeaderTestCase {
	return assetHeaderTestCase{
		name:  name,
		asset: asset,
		want:  want,
	}
}

func TestAssetHeaders(t *testing.T) {
	withHeaders := headerAsset{
		hdr: http.Header{consts.RetryAfter: core.S("30")},
	}

	testCases := []assetHeaderTestCase{
		newAssetHeaderTestCase("asset with headers", withHeaders, core.S("30")),
		newAssetHeaderTestCase("asset without headers", plainAsset{}, nil),
	}

	core.RunTestCases(t, testCases)
}

// assetLastModifiedTestCase records the Last-Modified header an asset
// earns. The handler consults ModTime first and falls back to Stat and
// then Info, so an asset describing itself through any of them is
// dated, and one that cannot answer, or answers with a zero time or an
// error, is served undated.
type assetLastModifiedTestCase struct {
	asset assets.Asset
	name  string
	want  []string
}

func (tc assetLastModifiedTestCase) Name() string {
	return tc.name
}

func (tc assetLastModifiedTestCase) Test(t *testing.T) {
	t.Helper()

	rw := serveAsset(t, tc.asset)
	core.AssertSliceEqual(t, tc.want,
		rw.Header().Values(consts.LastModified), "Last-Modified")
}

func newAssetLastModifiedTestCase(name string, asset assets.Asset,
	want []string) assetLastModifiedTestCase {
	return assetLastModifiedTestCase{
		name:  name,
		asset: asset,
		want:  want,
	}
}

// newAssetLastModifiedTestCaseDated declares a row expecting the
// header rendered for [assetModTime].
func newAssetLastModifiedTestCaseDated(name string,
	asset assets.Asset) assetLastModifiedTestCase {
	return newAssetLastModifiedTestCase(name, asset,
		core.S(assetModTime.Format(http.TimeFormat)))
}

// newAssetLastModifiedTestCaseUndated declares a row expecting no
// header at all.
func newAssetLastModifiedTestCaseUndated(name string,
	asset assets.Asset) assetLastModifiedTestCase {
	return newAssetLastModifiedTestCase(name, asset, nil)
}

func assetLastModifiedTestCases() []assetLastModifiedTestCase {
	info := newFileInfo(assetModTime)

	return []assetLastModifiedTestCase{
		newAssetLastModifiedTestCaseDated("declared ModTime",
			modTimeAsset{modTime: assetModTime}),
		newAssetLastModifiedTestCaseDated("Stat reports the time",
			statAsset{info: info}),
		newAssetLastModifiedTestCaseDated("Info reports the time",
			infoAsset{info: info}),

		newAssetLastModifiedTestCaseUndated("nothing to ask",
			plainAsset{}),
		newAssetLastModifiedTestCaseUndated("zero ModTime",
			modTimeAsset{}),
		newAssetLastModifiedTestCaseUndated("Stat fails",
			statAsset{err: fs.ErrNotExist}),
		newAssetLastModifiedTestCaseUndated("Stat answers with nothing",
			statAsset{}),
		newAssetLastModifiedTestCaseUndated("Info fails",
			infoAsset{err: fs.ErrNotExist}),
	}
}

func TestAssetLastModified(t *testing.T) {
	core.RunTestCases(t, assetLastModifiedTestCases())
}
