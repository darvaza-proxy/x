package web_test

// cspell:words Fevil

import (
	"net/http"
	"strings"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web"
	"darvaza.org/x/web/consts"
)

var (
	_ core.TestCase = redirectTestCase{}
	_ core.TestCase = redirectStatusTestCase{}
)

// redirectTestCase exercises redirect composition through
// web.NewStatusFound: args formatting, CleanURL normalisation
// and Location header wiring. wantErr selects between the
// success path (302, Location set) and the error path where
// CleanURL rejects via url.Parse, core.SplitHostPort or the
// DNS label-length check — caller asked for 302, factory
// wraps the failure as a 500.
type redirectTestCase struct {
	name    string
	dest    string
	want    string
	args    []any
	wantErr bool
}

func (tc redirectTestCase) Name() string {
	return tc.name
}

func (tc redirectTestCase) Test(t *testing.T) {
	t.Helper()

	err := web.NewStatusFound(tc.dest, tc.args...)
	core.AssertMustNotNil(t, err, "HTTPError")

	wantCode := core.IIf(tc.wantErr, http.StatusInternalServerError, http.StatusFound)
	core.AssertEqual(t, wantCode, err.HTTPStatus(), "status code")

	hdr := err.Header()
	if tc.wantErr {
		core.AssertFalse(t, web.HasHeader(hdr, consts.Location), "no Location")
		core.AssertNotNil(t, err.Err, "wrapped error")
		return
	}
	core.AssertSliceEqual(t, []string{tc.want},
		hdr.Values(consts.Location), "Location")
}

func newRedirectTestCase(name, dest, want string,
	args ...any) redirectTestCase {
	return redirectTestCase{
		name: name,
		dest: dest,
		args: args,
		want: want,
	}
}

func newRedirectErrorTestCase(name, dest string) redirectTestCase {
	return redirectTestCase{
		name:    name,
		dest:    dest,
		wantErr: true,
	}
}

func redirectTestCases() []redirectTestCase {
	return []redirectTestCase{
		newRedirectTestCase("plain rooted", "/foo", "/foo"),
		newRedirectTestCase("rooted trailing slash", "/foo/", "/foo/"),
		newRedirectTestCase("rooted dotdot reduces", "/foo/../bar", "/bar"),
		newRedirectTestCase("relative", "bar", "bar"),
		newRedirectTestCase("root", "/", "/"),
		newRedirectTestCase("empty", "", ""),
		newRedirectTestCase("sprintf args", "/foo/%d", "/foo/42", 42),
		// Format verb without args skips Sprintf and passes through
		// verbatim — pins the len(args) > 0 guard.
		newRedirectTestCase("format verb without args",
			"/search?q=%20", "/search?q=%20"),
		// Interpretation vectors get cleaned on the path.
		newRedirectTestCase("backslash-slash path",
			`/\evil.com`, "/evil.com"),
		newRedirectTestCase("backslash root path", `/\`, "/"),
		newRedirectTestCase("dot-backslash bypass", `/./\evil.com`, "/evil.com"),
		newRedirectTestCase("dotdot-backslash bypass",
			`/a/../\evil.com`, "/evil.com"),
		newRedirectTestCase("percent-encoded slash vector",
			"/%2Fevil.com", "/evil.com"),
		// Leading rooted /.. is stripped.
		newRedirectTestCase("escape root", "/..", "/"),
		newRedirectTestCase("escape then path", "/../evil.com", "/evil.com"),
		newRedirectTestCase("escape twice", "/../../evil.com", "/evil.com"),
		// Full URL and protocol-relative authority are application
		// intent; preserved verbatim, path cleaned.
		newRedirectTestCase("absolute URL", "https://sso.example.com/login",
			"https://sso.example.com/login"),
		newRedirectTestCase("absolute URL escape cleaned",
			"https://x.com/../y", "https://x.com/y"),
		newRedirectTestCase("protocol-relative authority",
			"//evil.com", "//evil.com"),
		// url.Parse rejects the input — newRedirect wraps
		// the failure as a 500 rather than emit a broken
		// Location.
		newRedirectErrorTestCase("invalid percent escape", "%ZZ"),
		// SplitHostPort rejects the host shape — same 500 path
		// as the url.Parse branch, different trigger.
		newRedirectErrorTestCase("host starts with dot", "//.com/x"),
		newRedirectErrorTestCase("port out of range",
			"https://example.com:99999/x"),
		// DNS label-length check rejects labels over 63 bytes
		// — the third CleanURL error source.
		newRedirectErrorTestCase("dns label too long",
			"https://"+strings.Repeat("a", 64)+".example.com/x"),
	}
}

func TestNewRedirect(t *testing.T) {
	core.RunTestCases(t, redirectTestCases())
}

// redirectStatusTestCase verifies each generated dispatcher hands
// off to newRedirect with the right status code, and — via the
// dirty-input row — that it actually routes through Clean.
type redirectStatusTestCase struct {
	factory func(string, ...any) *web.HTTPError
	name    string
	dest    string
	want    string
	code    int
	wantErr bool
}

func (tc redirectStatusTestCase) Name() string {
	return tc.name
}

func (tc redirectStatusTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.factory(tc.dest)
	core.AssertMustNotNil(t, err, "HTTPError")
	core.AssertEqual(t, tc.code, err.HTTPStatus(), "status code")

	hdr := err.Header()
	if tc.wantErr {
		core.AssertFalse(t, web.HasHeader(hdr, consts.Location), "no Location")
		core.AssertNotNil(t, err.Err, "wrapped error")
		return
	}
	core.AssertSliceEqual(t, []string{tc.want},
		hdr.Values(consts.Location), "Location")
}

func newRedirectStatusTestCase(name, dest, want string,
	factory func(string, ...any) *web.HTTPError,
	code int) redirectStatusTestCase {
	return redirectStatusTestCase{
		name:    name,
		dest:    dest,
		want:    want,
		factory: factory,
		code:    code,
	}
}

func newRedirectStatusErrorTestCase(name string,
	factory func(string, ...any) *web.HTTPError) redirectStatusTestCase {
	return redirectStatusTestCase{
		name:    name,
		dest:    "%ZZ",
		factory: factory,
		code:    http.StatusInternalServerError,
		wantErr: true,
	}
}

func redirectStatusTestCases() []redirectStatusTestCase {
	return []redirectStatusTestCase{
		newRedirectStatusTestCase("MovedPermanently", "/foo", "/foo",
			web.NewStatusMovedPermanently, http.StatusMovedPermanently),
		newRedirectStatusTestCase("Found", "/foo", "/foo",
			web.NewStatusFound, http.StatusFound),
		// Dirty input confirms the dispatcher runs through Clean,
		// not just that the status code is wired up.
		newRedirectStatusTestCase("Found cleans escape", "/../x", "/x",
			web.NewStatusFound, http.StatusFound),
		newRedirectStatusTestCase("SeeOther", "/foo", "/foo",
			web.NewStatusSeeOther, http.StatusSeeOther),
		newRedirectStatusTestCase("TemporaryRedirect", "/foo", "/foo",
			web.NewStatusTemporaryRedirect, http.StatusTemporaryRedirect),
		newRedirectStatusTestCase("PermanentRedirect", "/foo", "/foo",
			web.NewStatusPermanentRedirect, http.StatusPermanentRedirect),
		// Error-path rows confirm each dispatcher delegates the
		// CleanURL failure to newRedirect's 500 wrap; Found's
		// error path is exhaustively covered in redirectTestCase.
		newRedirectStatusErrorTestCase("MovedPermanently error",
			web.NewStatusMovedPermanently),
		newRedirectStatusErrorTestCase("SeeOther error",
			web.NewStatusSeeOther),
		newRedirectStatusErrorTestCase("TemporaryRedirect error",
			web.NewStatusTemporaryRedirect),
		newRedirectStatusErrorTestCase("PermanentRedirect error",
			web.NewStatusPermanentRedirect),
	}
}

func TestRedirectStatusHelpers(t *testing.T) {
	core.RunTestCases(t, redirectStatusTestCases())
}
