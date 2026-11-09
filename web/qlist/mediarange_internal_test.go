package qlist

import (
	"testing"
)

//revive:disable-next-line:argument-limit
func parsedEqual(t *testing.T, mime string,
	mtype, stype string, q float32, attrs map[string]string) {
	t.Helper()
	r, err := ParseMediaRange(mime)
	if err != nil {
		t.Fatalf("Failed to parse %q: %s", mime, err)
	}
	if got := mediaRangeType(r); mtype != got {
		t.Errorf("Failed to parse major type %q from %q, got %q",
			mtype, mime, got)
	}
	if got := mediaRangeSubType(r); stype != got {
		t.Errorf("Failed to parse minor type %q from %q, got %q",
			stype, mime, got)
	}
	if q != r.Quality() {
		t.Errorf("Failed to parse quality %v from %q, got %v",
			q, mime, r.Quality())
	}
	if !equalAttrs(attrs, r.attrs) {
		t.Errorf("Failed to parse attributes %v from %q, got %v",
			attrs, mime, r.attrs)
	}
}

func mediaRangeType(r QualityValue) string {
	var mtype string
	if len(r.value) > 0 {
		mtype = r.value[0]
	}

	if mtype == "" {
		mtype = "*"
	}

	return mtype
}

func mediaRangeSubType(r QualityValue) string {
	var stype string
	if len(r.value) > 1 {
		stype = r.value[1]
	}

	if stype == "" {
		stype = "*"
	}

	return stype
}

func equalAttrs(a, b map[string]string) bool {
	if len(a) == len(b) {
		for k, va := range a {
			vb, ok := b[k]
			if !ok || va != vb {
				return false
			}
		}
		return true
	}
	return false
}

func TestParseMimeType(t *testing.T) {
	parsedEqual(t, "Application/xhtml;q=0.5;vEr=1.2", "application", "xhtml",
		0.5, map[string]string{"ver": "1.2"})
}

func TestParseMediaRange(t *testing.T) {
	parsedEqual(t, "application/xml;q=1", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml;q=", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml;q", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml ; q=", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml ; q=1;b=other", "application", "xml",
		1, map[string]string{"b": "other"})
	parsedEqual(t, "application/xml ; q=2;b=other", "application", "xml",
		1, map[string]string{"b": "other"})
	// Java URLConnection class sends an Accept header that includes a single *
	parsedEqual(t, " *;q=.2", "*", "*", .2, nil)
}
