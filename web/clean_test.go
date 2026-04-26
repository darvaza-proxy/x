package web

// cspell:words Fevil

import (
	"strings"
	"testing"

	"darvaza.org/core"
)

var _ core.TestCase = cleanTestCase{}

type cleanTestCase struct {
	name string
	path string
	out  string
	ok   bool
}

func (tc cleanTestCase) Name() string {
	return tc.name
}

func (tc cleanTestCase) Test(t *testing.T) {
	t.Helper()

	s, ok := Clean(tc.path)
	core.AssertEqual(t, tc.out, s, "cleaned")
	core.AssertEqual(t, tc.ok, ok, "ok")
}

func newCleanTestCase(name, path, out string, ok bool) cleanTestCase {
	return cleanTestCase{
		name: name,
		path: path,
		out:  out,
		ok:   ok,
	}
}

func cleanTestCases() []cleanTestCase {
	return []cleanTestCase{
		// fs.Clean parity for ordinary paths.
		newCleanTestCase("empty", "", ".", true),
		newCleanTestCase("dot", ".", ".", true),
		newCleanTestCase("single component", "a", "a", true),
		newCleanTestCase("two components", "a/b", "a/b", true),
		newCleanTestCase("double slash", "a//b", "a/b", true),
		newCleanTestCase("self-cancel", "a/..", ".", true),
		newCleanTestCase("cancel and continue", "a/../b", "b", true),
		newCleanTestCase("root", "/", "/", true),
		newCleanTestCase("rooted single", "/a", "/a", true),
		newCleanTestCase("rooted trailing slash", "/a/", "/a/", true),
		newCleanTestCase("rooted cancel to root", "/a/..", "/", true),
		newCleanTestCase("rooted two", "/a/b", "/a/b", true),
		newCleanTestCase("rooted cancel", "/a/b/..", "/a", true),
		newCleanTestCase("rooted cancel and continue", "/a/b/../foo", "/a/foo", true),
		// WHATWG backslash normalisation: \ becomes / before fs.Clean.
		newCleanTestCase("backslash root", `/\`, "/", true),
		newCleanTestCase("protocol-relative backslash", `/\evil.com`, "/evil.com", true),
		newCleanTestCase("protocol-relative backslash with path",
			`/\evil.com/path`, "/evil.com/path", true),
		newCleanTestCase("dot-backslash bypass", `/./\evil.com`, "/evil.com", true),
		newCleanTestCase("dotdot-backslash bypass",
			`/a/../\evil.com`, "/evil.com", true),
		// Protocol-relative shapes flatten via fs.Clean's empty-leading drop.
		newCleanTestCase("double slash root", "//", "/", true),
		newCleanTestCase("protocol-relative slash", "//evil.com", "/evil.com", true),
		newCleanTestCase("protocol-relative slash with path",
			"//evil.com/path", "/evil.com/path", true),
		// /.. escape stripping — ok=false signals discarded blocks.
		newCleanTestCase("escape root", "/..", "/", false),
		newCleanTestCase("escape root trailing", "/../", "/", false),
		newCleanTestCase("escape root twice", "/../..", "/", false),
		newCleanTestCase("escape then path", "/../evil.com", "/evil.com", false),
		newCleanTestCase("escape then path with trailing",
			"/../foo/", "/foo/", false),
		newCleanTestCase("escape twice then path",
			"/../../evil.com", "/evil.com", false),
		newCleanTestCase("escape three deep",
			"/../../../x", "/x", false),
		newCleanTestCase("escape with trailing double slash",
			"/../..//", "/", false),
		newCleanTestCase("dotdot prefix of filename",
			"/..foo", "/..foo", true),
		// Relative escape: fs.Clean leaves it; we don't strip (not /..).
		newCleanTestCase("relative dotdot", "..", "..", true),
		newCleanTestCase("relative dotdot trailing", "../", "../", true),
		newCleanTestCase("relative dotdot with path", "../a", "../a", true),
		// ok invariant: non-/.. reductions keep ok=true even
		// when the output differs from the input.
		newCleanTestCase("ok-invariant backslash normalisation",
			`/a\b`, "/a/b", true),
		newCleanTestCase("ok-invariant fs.Clean reduction",
			"/a/b/../c", "/a/c", true),
		// Pathological inputs: documented, not guarded against.
		newCleanTestCase("carriage return passes through",
			"/a\rb", "/a\rb", true),
		newCleanTestCase("newline passes through",
			"/a\nb", "/a\nb", true),
		newCleanTestCase("null byte passes through",
			"/a\x00b", "/a\x00b", true),
		newCleanTestCase("url-encoded dotdot not decoded",
			"/%2e%2e/evil.com", "/%2e%2e/evil.com", true),
		newCleanTestCase("full URL input mangled",
			"http://evil.com/foo", "http:/evil.com/foo", true),
	}
}

func TestClean(t *testing.T) {
	core.RunTestCases(t, cleanTestCases())
}

var _ core.TestCase = cleanURLTestCase{}

type cleanURLTestCase struct {
	name    string
	in      string
	out     string
	wantErr bool
}

func (tc cleanURLTestCase) Name() string {
	return tc.name
}

func (tc cleanURLTestCase) Test(t *testing.T) {
	t.Helper()

	s, err := CleanURL(tc.in)
	core.AssertEqual(t, tc.out, s, "cleaned")
	if tc.wantErr {
		core.AssertError(t, err, "err expected")
	} else {
		core.AssertNoError(t, err, "err unexpected")
	}
}

func newCleanURLTestCase(name, in, out string) cleanURLTestCase {
	return cleanURLTestCase{
		name: name,
		in:   in,
		out:  out,
	}
}

func newCleanURLErrorTestCase(name, in string) cleanURLTestCase {
	return cleanURLTestCase{
		name:    name,
		in:      in,
		out:     "",
		wantErr: true,
	}
}

func cleanURLTestCases() []cleanURLTestCase {
	return []cleanURLTestCase{
		// Path-only inputs: identical to Clean.
		newCleanURLTestCase("plain rooted", "/foo", "/foo"),
		newCleanURLTestCase("rooted trailing slash", "/foo/", "/foo/"),
		newCleanURLTestCase("rooted reduce", "/a/../b", "/b"),
		// Backslash normalisation on the path component.
		newCleanURLTestCase("protocol-relative backslash",
			`/\evil.com`, "/evil.com"),
		newCleanURLTestCase("dotdot-backslash bypass",
			`/a/../\evil.com`, "/evil.com"),
		// Browser /..-clamp: leading /.. silently stripped.
		newCleanURLTestCase("escape root", "/..", "/"),
		newCleanURLTestCase("escape then path", "/../evil.com", "/evil.com"),
		newCleanURLTestCase("escape twice", "/../../evil.com", "/evil.com"),
		// Full URL: scheme + authority preserved, path cleaned.
		newCleanURLTestCase("absolute URL clean path",
			"https://sso.example.com/login", "https://sso.example.com/login"),
		newCleanURLTestCase("absolute URL path escape stripped",
			"https://x.com/../y", "https://x.com/y"),
		newCleanURLTestCase("absolute URL backslash in path",
			`https://x.com/a\b`, "https://x.com/a/b"),
		// Protocol-relative with authority: caller intent, pass through.
		newCleanURLTestCase("protocol-relative authority",
			"//evil.com", "//evil.com"),
		newCleanURLTestCase("protocol-relative authority with path",
			"//evil.com/a/../b", "//evil.com/b"),
		newCleanURLTestCase("protocol-relative authority with leading escape",
			"//evil.com/../x", "//evil.com/x"),
		// Percent-encoded interpretation vector: %2F decodes to /,
		// path becomes "//evil.com", fs.Clean collapses empty-leading.
		newCleanURLTestCase("percent-encoded slash vector",
			"/%2Fevil.com", "/evil.com"),
		// Percent-encoded benign: Path unchanged, RawPath preserved.
		newCleanURLTestCase("percent-encoded benign preserved",
			"/a%2Fb", "/a%2Fb"),
		// Query and fragment survive around path cleaning.
		newCleanURLTestCase("query preserved",
			"/foo?q=a/../b", "/foo?q=a/../b"),
		newCleanURLTestCase("fragment preserved",
			"/foo#bar", "/foo#bar"),
		newCleanURLTestCase("query and fragment around clean",
			"/a/../b?x=1#y", "/b?x=1#y"),
		// Host normalisation: lowercase, default-port strip.
		newCleanURLTestCase("uppercase host lowercased",
			"https://Example.COM/path", "https://example.com/path"),
		newCleanURLTestCase("https default port stripped",
			"https://example.com:443/x", "https://example.com/x"),
		newCleanURLTestCase("http default port stripped",
			"http://example.com:80/", "http://example.com/"),
		newCleanURLTestCase("non-default port preserved",
			"https://example.com:8443/x", "https://example.com:8443/x"),
		newCleanURLTestCase("ws default port stripped",
			"ws://example.com:80/x", "ws://example.com/x"),
		newCleanURLTestCase("wss default port stripped",
			"wss://example.com:443/x", "wss://example.com/x"),
		newCleanURLTestCase("IPv4 default port stripped",
			"http://192.168.1.1:80/x", "http://192.168.1.1/x"),
		newCleanURLTestCase("IPv4 non-default port preserved",
			"https://192.168.1.1:8080/x", "https://192.168.1.1:8080/x"),
		newCleanURLTestCase("IPv6 default port stripped",
			"https://[::1]:443/x", "https://[::1]/x"),
		newCleanURLTestCase("IPv6 non-default port preserved",
			"https://[::1]:8080/x", "https://[::1]:8080/x"),
		newCleanURLTestCase("IPv6 canonicalised",
			"https://[2001:0DB8::0001]/x", "https://[2001:db8::1]/x"),
		newCleanURLTestCase("protocol-relative host lowercased",
			"//Example.COM/x", "//example.com/x"),
		// IDN normalisation: punycode / raw Unicode / mixed-case
		// punycode all converge on lowercase punycode on output
		// — the canonical wire form for IDN in URL hosts.
		newCleanURLTestCase("IDN punycode lowercased",
			"https://xn--fsqu00a.example/x",
			"https://xn--fsqu00a.example/x"),
		newCleanURLTestCase("IDN punycode uppercase folded",
			"https://XN--FSQU00A.example/x",
			"https://xn--fsqu00a.example/x"),
		newCleanURLTestCase("IDN raw Unicode converted to punycode",
			"https://例子.example/x",
			"https://xn--fsqu00a.example/x"),
		// Userinfo round-trip: username:password should pass
		// through untouched, host component still lowercased.
		newCleanURLTestCase("userinfo preserved host lowercased",
			"https://user:pass@Example.COM/x",
			"https://user:pass@example.com/x"),
		// Empty input: url.Parse accepts, no host/path work to do.
		newCleanURLTestCase("empty input", "", ""),
		// Unknown scheme: defaultPortFor returns "" so the port
		// is preserved verbatim.
		newCleanURLTestCase("unknown scheme port preserved",
			"ftp://example.com:21/x", "ftp://example.com:21/x"),
		// Error path: url.Parse rejects the input — signals a
		// server-side bug, caller routes to 500.
		newCleanURLErrorTestCase("invalid percent escape", "%ZZ"),
		// Error path: trailing-dot FQDN — DNS-valid but
		// core.SplitHostPort rejects, surfaced as a malformed
		// composition (500).
		newCleanURLErrorTestCase("trailing dot host rejected",
			"https://example.com./x"),
		// Error path: url.Parse accepts but SplitHostPort rejects
		// the host shape (leading dot). Same 500 signal.
		newCleanURLErrorTestCase("host starts with dot", "//.com/x"),
		// Error path: url.Parse accepts the port but it's out of
		// the 1-65535 range; SplitHostPort rejects.
		newCleanURLErrorTestCase("port out of range",
			"https://example.com:99999/x"),
		// Error path: a Unicode label long enough that its
		// punycode A-label exceeds the 63-byte DNS limit
		// idna.Display.ToASCII enforces but
		// idna.Display.ToUnicode (used by core.SplitHostPort)
		// tolerates.
		newCleanURLErrorTestCase("label too long after punycode",
			"https://"+strings.Repeat("日", 100)+".example/x"),
	}
}

func TestCleanURL(t *testing.T) {
	core.RunTestCases(t, cleanURLTestCases())
}
