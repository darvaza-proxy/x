package qlist_test

import (
	"testing"

	"darvaza.org/x/web/qlist"
)

// Media types the BestMatch tables offer as supported and expect back. The
// Accept headers driving those tables stay verbatim, as each is the input
// under test rather than a fixture.
const (
	mimeJSON     = "application/json"
	mimeXBEL     = "application/xbel+xml"
	mimeXML      = "application/xml"
	mimeAnyImage = "image/*"
	mimeHTML     = "text/html"
	mimeTextXML  = "text/xml"
)

func TestRFC2616Example(t *testing.T) {
	//revive:disable-next-line:line-length-limit
	accept := "text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, * /*;q=0.5"
	cond := map[string]float32{
		"text/html;level=1": 1.0,
		"text/html":         0.7,
		"text/plain":        0.3,
		"image/jpeg":        0.5,
		"text/html;level=2": 0.4,
		"text/html;level=3": 0.7,
	}
	for mime, q := range cond {
		if q != qlist.MediaRangeQuality(mime, accept) {
			t.Errorf("Failed to match %v at %f, got %f instead",
				mime, q, qlist.MediaRangeQuality(mime, accept))
		}
	}
}

func doTestBestMatch(t *testing.T, supported []string, headers map[string]string) {
	for header, result := range headers {
		match := qlist.MediaRangeBestQuality(supported, header)
		if match != result {
			t.Errorf("BestMatch(%v, %v) == %s, not %s\n", supported, header, match, result)
		}
	}
}

func TestBestMatch(t *testing.T) {
	supported := []string{mimeXML, mimeXBEL}
	headers := map[string]string{
		"application/xbel+xml":      mimeXBEL,
		"application/xbel+xml; q=1": mimeXBEL,
		"application/xml; q=1":      mimeXML,
		"application/*; q=1":        mimeXML,
		"*/*":                       mimeXML,
	}
	doTestBestMatch(t, supported, headers)
}

func TestBestMatchDirect(t *testing.T) {
	supported := []string{mimeXBEL, mimeTextXML}
	headers := map[string]string{
		"text/*;q=0.5,*/*; q=0.1":               mimeTextXML,
		"text/html,application/atom+xml; q=0.9": "",
	}
	doTestBestMatch(t, supported, headers)
}

func TestBestMatchAjax(t *testing.T) {
	// Common AJAX scenario
	supported := []string{mimeJSON, mimeHTML}
	headers := map[string]string{
		"application/json, text/javascript, */*": mimeJSON,
		"application/json, text/html;q=0.9":      mimeJSON,
	}
	doTestBestMatch(t, supported, headers)
}

func TestSupportWildcards(t *testing.T) {
	supported := []string{mimeAnyImage, mimeXML}
	headers := map[string]string{
		"image/png": mimeAnyImage,
		"image/*":   mimeAnyImage,
	}
	doTestBestMatch(t, supported, headers)
}
