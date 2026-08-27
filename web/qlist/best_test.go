package qlist_test

import (
	"net/http"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/web/qlist"
)

var (
	_ core.TestCase = bestEncodingTestCase{}
	_ core.TestCase = bestQualityTestCase{}
	_ core.TestCase = bestQualityWithIdentityTestCase{}
)

// parseAccept turns a raw header value into the [qlist.QualityList] the Best
// functions consume. An empty value stands for an absent header and yields an
// empty list.
func parseAccept(t *testing.T, accept string) qlist.QualityList {
	t.Helper()

	if accept == "" {
		return nil
	}

	ql, err := qlist.ParseQualityString(accept)
	core.AssertMustNoError(t, err, "parse")
	return ql
}

// bestQualityTestCase exercises [qlist.BestQuality] against a raw header value.
type bestQualityTestCase struct {
	accept      string
	name        string
	want        string
	supported   []string
	wantQuality float32
}

func (tc bestQualityTestCase) Name() string {
	return tc.name
}

func (tc bestQualityTestCase) Test(t *testing.T) {
	t.Helper()

	got, quality, ok := qlist.BestQuality(tc.supported, parseAccept(t, tc.accept))

	core.AssertEqual(t, tc.want, got, "option")
	core.AssertEqual(t, tc.wantQuality, quality, "quality")
	core.AssertEqual(t, got != "", ok, "ok")
}

func newBestQualityTestCase(name string, supported []string, accept, want string,
	wantQuality float32) bestQualityTestCase {
	return bestQualityTestCase{
		accept:      accept,
		name:        name,
		want:        want,
		supported:   supported,
		wantQuality: wantQuality,
	}
}

func bestQualityTestCases() []bestQualityTestCase {
	return []bestQualityTestCase{
		newBestQualityTestCase("supported value accepted",
			core.S("gzip", "deflate"), "gzip", "gzip", 1),
		newBestQualityTestCase("highest quality wins",
			core.S("gzip", "br"), "gzip;q=0.5, br;q=0.8", "br", 0.8),
		newBestQualityTestCase("wildcard accepts a supported value",
			core.S("br"), "*", "br", 1),
		newBestQualityTestCase("unsupported value",
			core.S("br"), "gzip", "", 0),
		newBestQualityTestCase("zero quality is not a match",
			core.S("gzip"), "gzip;q=0", "", 0),
		newBestQualityTestCase("nothing supported",
			core.S[string](), "gzip", "", 0),
		newBestQualityTestCase("nothing accepted",
			core.S("gzip"), "", "", 0),
		newBestQualityTestCase("invalid supported value is dropped",
			core.S("gzip;q=nonsense", "br"), "br", "br", 1),
	}
}

func TestBestQuality(t *testing.T) {
	core.RunTestCases(t, bestQualityTestCases())
}

// bestQualityWithIdentityTestCase exercises [qlist.BestQualityWithIdentity].
// The identity is set by the factory rather than by the row, since it selects
// which of the two behaviours the row covers.
type bestQualityWithIdentityTestCase struct {
	accept      string
	identity    string
	name        string
	want        string
	supported   []string
	wantQuality float32
}

func (tc bestQualityWithIdentityTestCase) Name() string {
	return tc.name
}

func (tc bestQualityWithIdentityTestCase) Test(t *testing.T) {
	t.Helper()

	got, quality, ok := qlist.BestQualityWithIdentity(tc.supported,
		parseAccept(t, tc.accept), tc.identity)

	core.AssertEqual(t, tc.want, got, "option")
	core.AssertEqual(t, tc.wantQuality, quality, "quality")
	core.AssertEqual(t, got != "", ok, "ok")
}

// newIdentityTestCase covers a call offering the "identity" encoding as the
// fallback, the value [qlist.BestEncoding] passes.
func newIdentityTestCase(name string, supported []string, accept, want string,
	wantQuality float32) bestQualityWithIdentityTestCase {
	return bestQualityWithIdentityTestCase{
		accept:      accept,
		identity:    "identity",
		name:        name,
		want:        want,
		supported:   supported,
		wantQuality: wantQuality,
	}
}

// newNoIdentityTestCase covers a call offering no fallback, where
// [qlist.BestQualityWithIdentity] reduces to [qlist.BestQuality].
func newNoIdentityTestCase(name string, supported []string, accept, want string,
	wantQuality float32) bestQualityWithIdentityTestCase {
	return bestQualityWithIdentityTestCase{
		accept:      accept,
		identity:    "",
		name:        name,
		want:        want,
		supported:   supported,
		wantQuality: wantQuality,
	}
}

func bestQualityWithIdentityTestCases() []bestQualityWithIdentityTestCase {
	return []bestQualityWithIdentityTestCase{
		newIdentityTestCase("identity taken when nothing else matches",
			core.S("br"), "identity", "identity", 1),
		newIdentityTestCase("supported value beats the identity",
			core.S("gzip"), "gzip;q=0.8, identity;q=0.3", "gzip", 0.8),
		newIdentityTestCase("unmentioned identity is acceptable",
			core.S("br"), "gzip", "identity", 1),
		newIdentityTestCase("identity refused by name",
			core.S("br"), "gzip, identity;q=0", "", 0),
		newIdentityTestCase("identity refused by wildcard",
			core.S("br"), "*;q=0", "", 0),
		newIdentityTestCase("named identity outranks a refusing wildcard",
			core.S("br"), "*;q=0, identity;q=1", "identity", 1),
		newNoIdentityTestCase("no fallback offered",
			core.S("br"), "identity", "", 0),
		newNoIdentityTestCase("no fallback needed",
			core.S("gzip"), "gzip", "gzip", 1),
	}
}

func TestBestQualityWithIdentity(t *testing.T) {
	core.RunTestCases(t, bestQualityWithIdentityTestCases())
}

// bestEncodingTestCase exercises [qlist.BestEncoding] over an Accept-Encoding
// header. An empty accept stands for the header being absent.
type bestEncodingTestCase struct {
	accept    string
	name      string
	want      string
	supported []string
}

func (tc bestEncodingTestCase) Name() string {
	return tc.name
}

func (tc bestEncodingTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := qlist.BestEncoding(tc.supported, tc.header())

	core.AssertEqual(t, tc.want, got, "encoding")
	core.AssertEqual(t, got != "", ok, "ok")
}

func (tc bestEncodingTestCase) header() http.Header {
	hdr := make(http.Header)
	if tc.accept != "" {
		hdr.Set(qlist.AcceptEncoding, tc.accept)
	}
	return hdr
}

func newBestEncodingTestCase(name string, supported []string,
	accept, want string) bestEncodingTestCase {
	return bestEncodingTestCase{
		accept:    accept,
		name:      name,
		want:      want,
		supported: supported,
	}
}

func bestEncodingTestCases() []bestEncodingTestCase {
	return []bestEncodingTestCase{
		newBestEncodingTestCase("negotiates the best supported encoding",
			core.S("gzip", "br"), "br;q=0.9, gzip;q=0.5", "br"),
		newBestEncodingTestCase("wildcard takes a supported encoding",
			core.S("gzip"), "*", "gzip"),
		newBestEncodingTestCase("identity when the client asks for it",
			core.S("gzip"), "identity", "identity"),
		newBestEncodingTestCase("absent header falls back to identity",
			core.S("gzip"), "", "identity"),
		newBestEncodingTestCase("invalid header falls back to identity",
			core.S("gzip"), "gzip;q=nonsense", "identity"),
		newBestEncodingTestCase("refused identity selects nothing",
			core.S("br"), "gzip, identity;q=0", ""),
	}
}

func TestBestEncoding(t *testing.T) {
	core.RunTestCases(t, bestEncodingTestCases())
}
