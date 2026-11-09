package web

// cspell:words Fevil

import (
	"net"
	"net/url"
	"strings"

	"darvaza.org/core"

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
// HTTP Location header. The host is validated and put in its
// canonical ASCII form through core.SplitHostPortASCII, and its
// port stripped when it matches the scheme's known default
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
// Returns a non-nil error when url.Parse rejects the input,
// core.SplitHostPortASCII rejects the host or port, or the
// authority has no host or has port 0 — each indicates a malformed
// composition, a server-side bug rather than an attack. Callers
// should surface this as 500.
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
// CleanURL. It validates the host and port through
// checkCleanHost, drops the port when it matches the scheme's
// known default, and reassembles with IPv6 brackets as URL
// authorities require.
//
// Returns a non-nil error when checkCleanHost rejects the
// authority — surfaced to CleanURL's caller as a malformed
// composition, not swallowed.
func doCleanHost(scheme, hostPort string) (string, error) {
	host, port, err := checkCleanHost(hostPort)
	switch {
	case err != nil:
		// bad host or port
		return "", err
	case port == "", port == defaultPortFor(scheme):
		// no port, or the scheme's default
		if strings.Contains(host, ":") {
			// IPv6
			return "[" + host + "]", nil
		}
		return host, nil
	default:
		// host and port
		return net.JoinHostPort(host, port), nil
	}
}

// checkCleanHost splits hostPort through core.SplitHostPortASCII,
// which validates the host and port and returns the host in its
// canonical ASCII form, an IPv6 zone kept as given. It refuses an
// authority with no host or with port 0.
func checkCleanHost(hostPort string) (host, port string, err error) {
	host, port, err = core.SplitHostPortASCII(hostPort)
	switch {
	case err != nil:
		// bad host or port
		return "", "", err
	case host == "":
		// missing host
		return "", "", addrErr(hostPort, "missing host")
	case port == "0":
		// port 0, which the split returns without its padding
		return "", "", addrErr(hostPort, "invalid port")
	default:
		// host, with or without a port
		return host, port, nil
	}
}

// addrErr returns a *net.AddrError naming the address refused and why.
func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
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
