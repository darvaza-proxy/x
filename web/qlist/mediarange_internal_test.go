package qlist

import (
	"runtime"
	"testing"
)

//revive:disable-next-line:argument-limit
func parsedEqual(test *testing.T, mime string,
	t string, st string, q float32, attrs map[string]string) {
	r, err := ParseMediaRange(mime)
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		test.Errorf("%s:%d Failed to parse: %s", file, line, err)
		test.FailNow()
	}
	if got := mediaRangeType(r); t != got {
		test.Errorf("%s:%d Failed to parse major type %q from %q, got %q",
			file, line, t, mime, got)
	}
	if got := mediaRangeSubType(r); st != got {
		test.Errorf("%s:%d Failed to parse minor type %q from %q, got %q",
			file, line, st, mime, got)
	}
	if q != r.Quality() {
		test.Errorf("%s:%d Failed to parse quality %v from %q, got %v",
			file, line, st, mime, r.Quality())
	}
	if !equalAttrs(attrs, r.attrs) {
		test.Errorf("%s:%d Failed to parse attributes, expected %v, got %v",
			file, line, attrs, r.attrs)
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
