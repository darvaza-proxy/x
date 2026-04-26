package web

// cspell:words Fevil

import (
	"net"
	"net/url"
	"strings"

	"darvaza.org/core"
	"golang.org/x/net/idna"

	"darvaza.org/x/fs"
)

// Clean normalises a URL path. It is aimed primarily at paths
// destined for an outbound HTTP Location header, but is usable
// wherever a URL path needs to be reduced to a canonical form.
//
// Backslashes are replaced with forward slashes before reduction
// (matching WHATWG URL-path normalisation); the path is then
// reduced with fs.Clean, and any leading rooted /.. escape
// components are stripped so the cleaned result can't escape
// above the root. A trailing slash on the input is preserved.
//
// The input must be a path, not a full URL: Clean does not
// parse scheme or authority, so "http://evil.com/x" reduces as
// a path and becomes "http:/evil.com/x". Literal /.. is
// stripped; URL-encoded variants (for example %2e%2e) are not
// decoded and pass through verbatim.
//
// The second return is false when leading rooted /.. blocks
// had to be discarded — a signal that the input attempted to
// escape above the root. Backslash normalisation and fs.Clean's
// own reduction don't flip it.
func Clean(path string) (string, bool) {
	path = strings.ReplaceAll(path, `\`, "/")
	// Trailing-slash check runs on the post-ReplaceAll path, so
	// input ending in `\` is treated as trailing-slash input.
	trailing := len(path) > 1 && strings.HasSuffix(path, "/")

	cleaned, _ := fs.Clean(path)

	ok := true
	for strings.HasPrefix(cleaned, "/../") {
		cleaned = cleaned[3:]
		ok = false
	}
	// The loop can't remove a terminal /.. on its own: the
	// "/../" prefix match requires a following slash. Handle
	// the bare "/.." case here.
	if cleaned == "/.." {
		cleaned = "/"
		ok = false
	}

	if trailing && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned, ok
}

// CleanURL normalises a URL reference intended for an outbound
// HTTP Location header. The host is validated and lowercased
// through core.SplitHostPort, converted to ASCII punycode via
// idna.Display.ToASCII so non-ASCII labels are wire-safe, and
// its port stripped when it matches the scheme's known default
// (http/ws → 80, https/wss → 443). Scheme is preserved verbatim;
// the path is reduced through Clean.
//
// Leading rooted /.. is silently stripped from the path —
// browser URL resolution clamps /.. at root, and CleanURL
// mirrors that semantics rather than signalling escape
// attempts to the caller.
//
// CleanURL is a composer, not an origin validator. When dest
// is derived from untrusted input, the caller must validate
// origin or scheme before calling; CleanURL will happily
// reassemble "https://evil.com/foo" as-is.
//
// Returns a non-nil error when url.Parse rejects the input or
// core.SplitHostPort / idna rejects the host shape — either
// indicates a malformed composition, a server-side bug rather
// than an attack. Callers should surface this as 500.
func CleanURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if u.Host != "" {
		host, err := doCleanHost(u.Scheme, u.Host)
		if err != nil {
			return "", err
		}
		u.Host = host
	}
	if u.Path == "" {
		return u.String(), nil
	}
	// Clean's ok discarded: browser /..-clamping semantics,
	// see doc comment above.
	cleaned, _ := Clean(u.Path)
	if cleaned != u.Path {
		u.Path = cleaned
		// Force re-encoding from Path. Otherwise u.String()
		// reuses RawPath and our cleaning is invisible on
		// inputs like "/%2Fevil.com".
		u.RawPath = ""
	}
	return u.String(), nil
}

// doCleanHost normalises a URL-authority host[:port] for
// CleanURL. Splits through core.SplitHostPort for IDN/IPv6-aware
// validation and lowercasing, converts the resulting host to
// ASCII punycode via idna.Display.ToASCII so u.String() doesn't
// percent-escape non-ASCII labels on serialisation, strips the
// port when it matches the scheme's known default, and
// reassembles with IPv6 brackets as URL authorities require.
//
// Returns a non-nil error when core.SplitHostPort or idna
// rejects the host shape — surfaced to CleanURL's caller as a
// malformed composition, not swallowed.
func doCleanHost(scheme, hostPort string) (string, error) {
	host, port, err := core.SplitHostPort(hostPort)
	if err != nil {
		return "", err
	}
	host, err = doCleanHostLabel(host)
	if err != nil {
		return "", err
	}
	if port == defaultPortFor(scheme) {
		port = ""
	}
	if port == "" {
		return host, nil
	}
	return host + ":" + port, nil
}

// doCleanHostLabel takes the host returned by core.SplitHostPort
// and shapes it for a URL authority: IPv4 in dotted-decimal,
// IPv6 bracketed and canonicalised (netip.Addr.String() shortens
// `::0001` → `::1` and lowercases hex digits), names converted
// to ASCII punycode via idna.Display.ToASCII. The punycode step
// sidesteps url.URL.String()'s percent-encoding of non-ASCII
// bytes in u.Host — Unicode from core.SplitHostPort would
// otherwise land as %XX on the wire.
//
// For names, RFC 1035's 63-byte-per-label limit is enforced
// after the ASCII conversion: idna.Display doesn't verify DNS
// length, and a punycode A-label expanded past 63 bytes would
// otherwise reach clients that reject the redirect. IP literals
// always fit well under the limit, so the check lives on the
// name branch only.
func doCleanHostLabel(host string) (string, error) {
	if addr, ipErr := core.ParseAddr(host); ipErr == nil {
		s := addr.String()
		if addr.Is6() {
			s = "[" + s + "]"
		}
		return s, nil
	}

	// core.SplitHostPort's validName already gated host through
	// idna.Display.ToUnicode; the matching ToASCII call in the
	// same Display profile cannot reject what ToUnicode accepted.
	// core.Must lets a future idna or core drift surface loudly.
	s := core.Must(idna.Display.ToASCII(host))
	for _, label := range strings.Split(s, ".") {
		if len(label) > 63 {
			return "", &net.AddrError{
				Err:  "label exceeds 63-byte DNS limit",
				Addr: host,
			}
		}
	}
	return s, nil
}

func defaultPortFor(scheme string) string {
	switch scheme {
	case "http", "ws":
		return "80"
	case "https", "wss":
		return "443"
	}
	return ""
}
