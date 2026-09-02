package web_test

import (
	"errors"
	"net/http"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
)

// Compile-time verification that the test case type implements TestCase.
var _ core.TestCase = asErrorHeadersTestCase{}

var errPlainNoHeaders = errors.New("plain error")

// headerProviderError carries response headers under the singular
// method name, the shape [web.HeaderProvider] selects. It deliberately
// omits ServeHTTP so AsError converts it instead of returning it
// untouched as an [http.Handler].
type headerProviderError struct {
	hdr http.Header
}

func (*headerProviderError) Error() string { return "header provider" }

func (err *headerProviderError) Header() http.Header { return err.hdr }

// headersProviderError is the same in the plural spelling, the shape
// [web.HeadersProvider] selects.
type headersProviderError struct {
	hdr http.Header
}

func (*headersProviderError) Error() string { return "headers provider" }

func (err *headersProviderError) Headers() http.Header { return err.hdr }

// newRetryAfterHeader returns a header carrying a single Retry-After
// entry, enough to tell a carried header apart from an absent one.
func newRetryAfterHeader(delay string) http.Header {
	return http.Header{consts.RetryAfter: core.S(delay)}
}

// asErrorHeadersTestCase records which headers AsError carries over
// from the error it converts. Both provider spellings contribute
// theirs; anything else contributes none, leaving the field nil.
type asErrorHeadersTestCase struct {
	err  error
	want http.Header
	name string
}

func (tc asErrorHeadersTestCase) Name() string {
	return tc.name
}

func (tc asErrorHeadersTestCase) Test(t *testing.T) {
	t.Helper()

	got := core.AssertMustTypeIs[*web.HTTPError](t, web.AsError(tc.err),
		"HTTPError")
	core.AssertDeepEqual(t, tc.want, got.Hdr, "headers")
}

func newAsErrorHeadersTestCase(name string, err error,
	want http.Header) asErrorHeadersTestCase {
	return asErrorHeadersTestCase{
		name: name,
		err:  err,
		want: want,
	}
}

func TestAsErrorHeaders(t *testing.T) {
	testCases := []asErrorHeadersTestCase{
		newAsErrorHeadersTestCase("singular Header",
			&headerProviderError{hdr: newRetryAfterHeader("30")},
			newRetryAfterHeader("30")),
		newAsErrorHeadersTestCase("plural Headers",
			&headersProviderError{hdr: newRetryAfterHeader("60")},
			newRetryAfterHeader("60")),
		newAsErrorHeadersTestCase("no headers", errPlainNoHeaders, nil),
	}

	core.RunTestCases(t, testCases)
}

// TestAsErrorClonesHeaders confirms the converted error owns its
// headers rather than aliasing the source, so a later change to the
// error it was built from does not reach the response.
func TestAsErrorClonesHeaders(t *testing.T) {
	src := newRetryAfterHeader("30")

	got := core.AssertMustTypeIs[*web.HTTPError](t,
		web.AsError(&headerProviderError{hdr: src}), "HTTPError")

	src.Set(consts.RetryAfter, "120")
	core.AssertSliceEqual(t, core.S("30"),
		got.Hdr.Values(consts.RetryAfter), "carried value")
}
