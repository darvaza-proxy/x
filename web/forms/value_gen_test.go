package forms_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web/consts"
	"darvaza.org/x/web/forms"
)

var (
	_ core.TestCase = valueNaNTestCase{}
	_ core.TestCase = valueTestCase{}
)

// postBody returns a factory of POST requests carrying form as their
// URL-encoded body.
func postBody(form string) func() *http.Request {
	return func() *http.Request {
		return newPostRequest("", form)
	}
}

// postQuery returns a factory of POST requests carrying form as their
// query string, with an empty URL-encoded body.
func postQuery(form string) func() *http.Request {
	return func() *http.Request {
		return newPostRequest(form, "")
	}
}

func newPostRequest(query, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/?"+query,
		strings.NewReader(body))
	req.Header.Set(consts.ContentType, consts.URLEncodedForm)
	return req
}

type valueTestCase struct {
	expected any
	target   error
	get      func() (any, bool, error)
	name     string
	found    bool
}

func (tc valueTestCase) Name() string {
	return tc.name
}

func (tc valueTestCase) Test(t *testing.T) {
	t.Helper()

	value, found, err := tc.get()
	if tc.target == nil {
		core.AssertNoError(t, err, "error")
	} else {
		core.AssertErrorIs(t, err, tc.target, "error")
	}
	core.AssertEqual(t, tc.found, found, "found")
	core.AssertEqual(t, tc.expected, value, "value")
}

// newValueRangeTestCase expects fn(req, field, minV, maxV) to find the
// field and return expected with no error.
//
//revive:disable-next-line:argument-limit
func newValueRangeTestCase[T any](name string,
	fn func(*http.Request, string, T, T) (T, bool, error),
	req func() *http.Request, field string,
	minV, maxV, expected T) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueRange(fn, req, field, minV, maxV),
		expected: expected,
		found:    true,
	}
}

// newValueRangeErrorTestCase expects fn(req, field, minV, maxV) to find
// the field and return expected with an error matching target.
//
//revive:disable-next-line:argument-limit
func newValueRangeErrorTestCase[T any](name string,
	fn func(*http.Request, string, T, T) (T, bool, error),
	req func() *http.Request, field string,
	minV, maxV, expected T, target error) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueRange(fn, req, field, minV, maxV),
		expected: expected,
		target:   target,
		found:    true,
	}
}

// newValueRangeMissingTestCase expects fn(req, field, minV, maxV) not to
// find the field, and to return expected with no error.
//
//revive:disable-next-line:argument-limit
func newValueRangeMissingTestCase[T any](name string,
	fn func(*http.Request, string, T, T) (T, bool, error),
	req func() *http.Request, field string,
	minV, maxV, expected T) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueRange(fn, req, field, minV, maxV),
		expected: expected,
		found:    false,
	}
}

func bindValueRange[T any](
	fn func(*http.Request, string, T, T) (T, bool, error),
	req func() *http.Request, field string,
	minV, maxV T) func() (any, bool, error) {
	return func() (any, bool, error) {
		v, found, err := fn(req(), field, minV, maxV)
		return v, found, err
	}
}

// newValueBaseRangeTestCase expects fn(req, field, base, minV, maxV) to
// find the field and return expected with no error.
//
//revive:disable-next-line:argument-limit
func newValueBaseRangeTestCase[T any](name string,
	fn func(*http.Request, string, int, T, T) (T, bool, error),
	req func() *http.Request, field string, base int,
	minV, maxV, expected T) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueBaseRange(fn, req, field, base, minV, maxV),
		expected: expected,
		found:    true,
	}
}

// newValueBaseRangeErrorTestCase expects fn(req, field, base, minV, maxV)
// to find the field and return expected with an error matching target.
//
//revive:disable-next-line:argument-limit
func newValueBaseRangeErrorTestCase[T any](name string,
	fn func(*http.Request, string, int, T, T) (T, bool, error),
	req func() *http.Request, field string, base int,
	minV, maxV, expected T, target error) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueBaseRange(fn, req, field, base, minV, maxV),
		expected: expected,
		target:   target,
		found:    true,
	}
}

// newValueBaseRangeMissingTestCase expects fn(req, field, base, minV, maxV)
// not to find the field, and to return expected with no error.
//
//revive:disable-next-line:argument-limit
func newValueBaseRangeMissingTestCase[T any](name string,
	fn func(*http.Request, string, int, T, T) (T, bool, error),
	req func() *http.Request, field string, base int,
	minV, maxV, expected T) valueTestCase {
	return valueTestCase{
		name:     name,
		get:      bindValueBaseRange(fn, req, field, base, minV, maxV),
		expected: expected,
		found:    false,
	}
}

//revive:disable-next-line:argument-limit
func bindValueBaseRange[T any](
	fn func(*http.Request, string, int, T, T) (T, bool, error),
	req func() *http.Request, field string, base int,
	minV, maxV T) func() (any, bool, error) {
	return func() (any, bool, error) {
		v, found, err := fn(req(), field, base, minV, maxV)
		return v, found, err
	}
}

// valueSignedInRangeTestCases holds the rows shared by the Form and
// PostForm variants of fn.
func valueSignedInRangeTestCases(
	fn func(*http.Request, string, int, int, int) (int, bool, error),
) []valueTestCase {
	return core.S(
		newValueBaseRangeTestCase("within",
			fn, postBody("n=5"), "n", 10, -50, 50,
			5),
		newValueBaseRangeTestCase("at min",
			fn, postBody("n=-50"), "n", 10, -50, 50,
			-50),
		newValueBaseRangeTestCase("at max",
			fn, postBody("n=50"), "n", 10, -50, 50,
			50),
		newValueBaseRangeErrorTestCase("below min",
			fn, postBody("n=-51"), "n", 10, -50, 50,
			-51, forms.ErrRange),
		newValueBaseRangeErrorTestCase("above max",
			fn, postBody("n=51"), "n", 10, -50, 50,
			51, forms.ErrRange),
		newValueBaseRangeErrorTestCase("false range",
			fn, postBody("n=5"), "n", 10, 50, -50,
			5, forms.ErrRange),
		newValueBaseRangeTestCase("base 16",
			fn, postBody("n=a"), "n", 16, -50, 50,
			10),

		newValueBaseRangeMissingTestCase("missing, zero out of bounds",
			fn, postBody("m=5"), "n", 10, 1, 50,
			0),
		newValueBaseRangeErrorTestCase("letters",
			fn, postBody("n=x"), "n", 10, -50, 50,
			0, forms.ErrSyntax),
		newValueBaseRangeErrorTestCase("empty",
			fn, postBody("n="), "n", 10, -50, 50,
			0, forms.ErrSyntax),
	)
}

func TestFormValueSignedInRange(t *testing.T) {
	fn := forms.FormValueSignedInRange[int]
	core.RunTestCases(t, append(valueSignedInRangeTestCases(fn),
		newValueBaseRangeTestCase("in the query",
			fn, postQuery("n=5"), "n", 10, -50, 50,
			5),
	))
}

func TestPostFormValueSignedInRange(t *testing.T) {
	fn := forms.PostFormValueSignedInRange[int]
	core.RunTestCases(t, append(valueSignedInRangeTestCases(fn),
		newValueBaseRangeMissingTestCase("in the query",
			fn, postQuery("n=5"), "n", 10, -50, 50,
			0),
	))
}

// valueUnsignedInRangeTestCases holds the rows shared by the Form and
// PostForm variants of fn.
func valueUnsignedInRangeTestCases(
	fn func(*http.Request, string, int, uint, uint) (uint, bool, error),
) []valueTestCase {
	return core.S(
		newValueBaseRangeTestCase("within",
			fn, postBody("n=5"), "n", 10, 2, 50,
			5),
		newValueBaseRangeTestCase("at min",
			fn, postBody("n=2"), "n", 10, 2, 50,
			2),
		newValueBaseRangeTestCase("at max",
			fn, postBody("n=50"), "n", 10, 2, 50,
			50),
		newValueBaseRangeErrorTestCase("below min",
			fn, postBody("n=1"), "n", 10, 2, 50,
			1, forms.ErrRange),
		newValueBaseRangeErrorTestCase("above max",
			fn, postBody("n=51"), "n", 10, 2, 50,
			51, forms.ErrRange),
		newValueBaseRangeErrorTestCase("false range",
			fn, postBody("n=5"), "n", 10, 50, 2,
			5, forms.ErrRange),
		newValueBaseRangeTestCase("base 16",
			fn, postBody("n=a"), "n", 16, 2, 50,
			10),

		newValueBaseRangeMissingTestCase("missing, zero out of bounds",
			fn, postBody("m=5"), "n", 10, 2, 50,
			0),
		newValueBaseRangeErrorTestCase("negative",
			fn, postBody("n=-1"), "n", 10, 2, 50,
			0, forms.ErrSyntax),
		newValueBaseRangeErrorTestCase("letters",
			fn, postBody("n=x"), "n", 10, 2, 50,
			0, forms.ErrSyntax),
		newValueBaseRangeErrorTestCase("empty",
			fn, postBody("n="), "n", 10, 2, 50,
			0, forms.ErrSyntax),
	)
}

func TestFormValueUnsignedInRange(t *testing.T) {
	fn := forms.FormValueUnsignedInRange[uint]
	core.RunTestCases(t, append(valueUnsignedInRangeTestCases(fn),
		newValueBaseRangeTestCase("in the query",
			fn, postQuery("n=5"), "n", 10, 2, 50,
			5),
	))
}

func TestPostFormValueUnsignedInRange(t *testing.T) {
	fn := forms.PostFormValueUnsignedInRange[uint]
	core.RunTestCases(t, append(valueUnsignedInRangeTestCases(fn),
		newValueBaseRangeMissingTestCase("in the query",
			fn, postQuery("n=5"), "n", 10, 2, 50,
			0),
	))
}

// valueFloatInRangeTestCases holds the rows shared by the Form and
// PostForm variants of fn.
func valueFloatInRangeTestCases(
	fn func(*http.Request, string, float64, float64) (float64, bool, error),
) []valueTestCase {
	return core.S(
		newValueRangeTestCase("within",
			fn, postBody("n=0.5"), "n", 0, 1,
			0.5),
		newValueRangeTestCase("at min",
			fn, postBody("n=0"), "n", 0, 1,
			0),
		newValueRangeTestCase("at max",
			fn, postBody("n=1"), "n", 0, 1,
			1),
		newValueRangeErrorTestCase("below min",
			fn, postBody("n=-0.5"), "n", 0, 1,
			-0.5, forms.ErrRange),
		newValueRangeErrorTestCase("above max",
			fn, postBody("n=1.5"), "n", 0, 1,
			1.5, forms.ErrRange),
		newValueRangeErrorTestCase("false range",
			fn, postBody("n=0.5"), "n", 1, 0,
			0.5, forms.ErrRange),
		newValueRangeErrorTestCase("infinity",
			fn, postBody("n=Inf"), "n", 0, 1,
			math.Inf(1), forms.ErrRange),

		newValueRangeMissingTestCase("missing, zero out of bounds",
			fn, postBody("m=0.5"), "n", 0.5, 1,
			0),
		newValueRangeErrorTestCase("letters",
			fn, postBody("n=x"), "n", 0, 1,
			0, forms.ErrSyntax),
		newValueRangeErrorTestCase("empty",
			fn, postBody("n="), "n", 0, 1,
			0, forms.ErrSyntax),
	)
}

func TestFormValueFloatInRange(t *testing.T) {
	fn := forms.FormValueFloatInRange[float64]
	core.RunTestCases(t, append(valueFloatInRangeTestCases(fn),
		newValueRangeTestCase("in the query",
			fn, postQuery("n=0.5"), "n", 0, 1,
			0.5),
	))
}

func TestPostFormValueFloatInRange(t *testing.T) {
	fn := forms.PostFormValueFloatInRange[float64]
	core.RunTestCases(t, append(valueFloatInRangeTestCases(fn),
		newValueRangeMissingTestCase("in the query",
			fn, postQuery("n=0.5"), "n", 0, 1,
			0),
	))
}

type valueNaNTestCase struct {
	target error
	get    func() (float64, bool, error)
	name   string
}

func (tc valueNaNTestCase) Name() string {
	return tc.name
}

func (tc valueNaNTestCase) Test(t *testing.T) {
	t.Helper()

	value, found, err := tc.get()
	core.AssertErrorIs(t, err, tc.target, "error")
	core.AssertTrue(t, found, "found")
	core.AssertTrue(t, math.IsNaN(value), "NaN")
}

// newValueRangeNaNErrorTestCase expects fn(req, field, minV, maxV) to find
// the field and return NaN with an error matching target.
//
//revive:disable-next-line:argument-limit
func newValueRangeNaNErrorTestCase[T core.Float](name string,
	fn func(*http.Request, string, T, T) (T, bool, error),
	req func() *http.Request, field string,
	minV, maxV T, target error) valueNaNTestCase {
	return valueNaNTestCase{
		name: name,
		get: func() (float64, bool, error) {
			v, found, err := fn(req(), field, minV, maxV)
			return float64(v), found, err
		},
		target: target,
	}
}

// valueFloatInRangeNaNTestCases holds the rows shared by the Form and
// PostForm variants of fn.
func valueFloatInRangeNaNTestCases(
	fn func(*http.Request, string, float64, float64) (float64, bool, error),
) []valueNaNTestCase {
	return core.S(
		newValueRangeNaNErrorTestCase("finite bounds",
			fn, postBody("n=NaN"), "n", 0, 1,
			forms.ErrRange),
		newValueRangeNaNErrorTestCase("infinite bounds",
			fn, postBody("n=NaN"), "n", math.Inf(-1), math.Inf(1),
			forms.ErrRange),
	)
}

func TestFormValueFloatInRangeNaN(t *testing.T) {
	core.RunTestCases(t,
		valueFloatInRangeNaNTestCases(forms.FormValueFloatInRange[float64]))
}

func TestPostFormValueFloatInRangeNaN(t *testing.T) {
	core.RunTestCases(t,
		valueFloatInRangeNaNTestCases(forms.PostFormValueFloatInRange[float64]))
}
