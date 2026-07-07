package x509utils_test

import (
	"crypto/x509"
	"net"
	"net/url"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils"
)

var (
	_ core.TestCase = nameAsIPTestCase{}
	_ core.TestCase = nameAsSuffixTestCase{}
	_ core.TestCase = namesIPTestCase{}
	_ core.TestCase = namesTestCase{}
	_ core.TestCase = sanitizeNameTestCase{}
)

// nameAsIPTestCase exercises NameAsIP: a parseable address is bracketed and
// accepted, anything else is rejected. Acceptance and a non-empty result are
// inseparable, so ok is asserted as the invariant expected != "" rather than
// declared as a redundant column.
type nameAsIPTestCase struct {
	input    string
	expected string
	name     string
}

func (tc nameAsIPTestCase) Name() string { return tc.name }

func (tc nameAsIPTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := x509utils.NameAsIP(tc.input)
	core.AssertEqual(t, tc.expected, got, "addr")
	core.AssertEqual(t, tc.expected != "", ok, "ok")
}

func newNameAsIPTestCase(name, input, expected string) nameAsIPTestCase {
	return nameAsIPTestCase{
		input:    input,
		expected: expected,
		name:     name,
	}
}

func nameAsIPTestCases() []nameAsIPTestCase {
	return core.S(
		newNameAsIPTestCase("hostname", "a.b.c", ""),
		newNameAsIPTestCase("bare zero", "0", "[0.0.0.0]"),
		newNameAsIPTestCase("ipv4", "1.2.3.4", "[1.2.3.4]"),
		newNameAsIPTestCase("overflow octet", "1.2.3.400", ""),
		newNameAsIPTestCase("ipv6 unspecified", "::", "[::]"),
		newNameAsIPTestCase("ipv4-mapped ipv6", "::ffff:1.2.3.4", "[1.2.3.4]"),
		newNameAsIPTestCase("zoned link-local", "fe80::1%eth0", "[fe80::1]"),
		newNameAsIPTestCase("fqdn", "foo.example.org", ""),
	)
}

func TestNameAsIP(t *testing.T) {
	core.RunTestCases(t, nameAsIPTestCases())
}

// nameAsSuffixTestCase exercises NameAsSuffix: the name is reduced to the
// dotted suffix after its first label, accepted only when that suffix is
// longer than a lone dot. ok is therefore the invariant len(expected) > 1, not
// expected != "": "a." yields "." yet is rejected.
type nameAsSuffixTestCase struct {
	input    string
	expected string
	name     string
}

func (tc nameAsSuffixTestCase) Name() string { return tc.name }

func (tc nameAsSuffixTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := x509utils.NameAsSuffix(tc.input)
	core.AssertEqual(t, tc.expected, got, "suffix")
	core.AssertEqual(t, len(tc.expected) > 1, ok, "ok")
}

func newNameAsSuffixTestCase(name, input,
	expected string) nameAsSuffixTestCase {
	return nameAsSuffixTestCase{
		input:    input,
		expected: expected,
		name:     name,
	}
}

func nameAsSuffixTestCases() []nameAsSuffixTestCase {
	return core.S(
		newNameAsSuffixTestCase("dotted host", "foo.example.com",
			".example.com"),
		newNameAsSuffixTestCase("leading dot", ".example.com", ""),
		newNameAsSuffixTestCase("two labels", "a.b.c", ".b.c"),
		newNameAsSuffixTestCase("leading dot two labels", ".b.c", ""),
		newNameAsSuffixTestCase("single inner label", "b.c", ".c"),
		newNameAsSuffixTestCase("leading dot one label", ".c", ""),
		newNameAsSuffixTestCase("bare label", "c", ""),
		// idx>0 yields a lone dot, which fails the len>1 gate: a non-empty
		// result with ok=false, the row that makes the invariant load-bearing.
		newNameAsSuffixTestCase("trailing dot", "a.", "."),
		newNameAsSuffixTestCase("empty", "", ""),
	)
}

func TestNameAsSuffix(t *testing.T) {
	core.RunTestCases(t, nameAsSuffixTestCases())
}

// sanitizeNameTestCase exercises SanitizeName: the name is split from its
// optional port and case-folded to the form Names stores under (RFC 6125),
// with IP addresses normalised. Acceptance and a non-empty result are
// inseparable, so ok is asserted as the invariant expected != "".
type sanitizeNameTestCase struct {
	input    string
	expected string
	name     string
}

func (tc sanitizeNameTestCase) Name() string { return tc.name }

func (tc sanitizeNameTestCase) Test(t *testing.T) {
	t.Helper()
	got, ok := x509utils.SanitizeName(tc.input)
	core.AssertEqual(t, tc.expected, got, "name")
	core.AssertEqual(t, tc.expected != "", ok, "ok")
}

func newSanitizeNameTestCase(name, input,
	expected string) sanitizeNameTestCase {
	return sanitizeNameTestCase{
		input:    input,
		expected: expected,
		name:     name,
	}
}

func sanitizeNameTestCases() []sanitizeNameTestCase {
	return core.S(
		// DNS names are case-insensitive (RFC 6125): an uppercase query
		// folds to the lower-cased form Names stores under.
		newSanitizeNameTestCase("uppercase host", "WWW.EXAMPLE.COM",
			"www.example.com"),
		newSanitizeNameTestCase("lowercase host", "www.example.com",
			"www.example.com"),
		newSanitizeNameTestCase("mixed case with port", "Mixed.Case.ORG:443",
			"mixed.case.org"),
		newSanitizeNameTestCase("ipv4", "1.2.3.4", "1.2.3.4"),
		newSanitizeNameTestCase("ipv6 with port", "[2001:DB8::1]:443",
			"2001:db8::1"),
		// An IP scope zone is stripped on the IP branch by WithZone(""),
		// never by removeZone.
		newSanitizeNameTestCase("ipv6 zone", "[fe80::1%eth0]:443",
			"fe80::1"),
		// A '%' can only survive nameRE inside a non-final label, and idna's
		// STD3 rules then reject it, so SplitHostPort fails before
		// doSanitizeName: the removeZone name-branch is never reached.
		newSanitizeNameTestCase("percent in host", "foo%bar.example", ""),
		newSanitizeNameTestCase("empty", "", ""),
	)
}

func TestSanitizeName(t *testing.T) {
	core.RunTestCases(t, sanitizeNameTestCases())
}

// namesTestCase exercises Names: DNS literals fold to lower case, wildcards
// become dotted patterns, and IP SANs are bracketed. The certificates are
// minted and re-parsed so their SANs are in the DER-canonical form the basic
// store feeds Names via cert.Leaf.
type namesTestCase struct {
	cert     *x509.Certificate
	name     string
	names    []string
	patterns []string
}

func (tc namesTestCase) Name() string { return tc.name }

func (tc namesTestCase) Test(t *testing.T) {
	t.Helper()
	names, patterns := x509utils.Names(tc.cert)
	core.AssertSliceEqual(t, tc.names, names, "names")
	core.AssertSliceEqual(t, tc.patterns, patterns, "patterns")
}

func newNamesTestCase(name string, cert *x509.Certificate,
	names, patterns []string) namesTestCase {
	return namesTestCase{
		cert:     cert,
		name:     name,
		names:    names,
		patterns: patterns,
	}
}

func namesTestCases(t *testing.T) []namesTestCase {
	t.Helper()

	dns := func(ds ...string) *x509.Certificate {
		return certSpec{cn: ds[0], dns: ds}.build(t).cert
	}
	ip := func(cn string, ips ...net.IP) *x509.Certificate {
		return certSpec{cn: cn, ips: ips}.build(t).cert
	}
	none := core.S[string]()

	return core.S(
		newNamesTestCase("dns literal", dns("a.example.com"),
			core.S("a.example.com"), none),
		// RFC 6125: DNS names are case-insensitive; Names lower-cases them.
		newNamesTestCase("uppercase folds", dns("Foo.Example.COM"),
			core.S("foo.example.com"), none),
		// a wildcard becomes a dotted suffix pattern, not a literal name.
		newNamesTestCase("wildcard pattern", dns("*.example.com"),
			none, core.S(".example.com")),
		newNamesTestCase("literal and wildcard",
			dns("a.example.com", "*.example.com"),
			core.S("a.example.com"), core.S(".example.com")),
		newNamesTestCase("ipv4 bracketed",
			ip("host", net.ParseIP("1.2.3.4")),
			core.S("[1.2.3.4]"), none),
		newNamesTestCase("ipv6 bracketed",
			ip("host", net.ParseIP("2001:db8::1")),
			core.S("[2001:db8::1]"), none),
	)
}

func TestNames(t *testing.T) {
	core.RunTestCases(t, namesTestCases(t))
}

// TestNamesNil confirms a nil certificate yields no names rather than
// dereferencing into a panic.
func TestNamesNil(t *testing.T) {
	names, patterns := x509utils.Names(nil)
	core.AssertNil(t, names, "names")
	core.AssertNil(t, patterns, "patterns")
}

// TestNamesDropsAndDeduplicates covers splitDNSNames' empty-name drop, the
// SliceUnique deduplication, and appendIPAddresses' malformed-IP drop, using
// certificate literals so Names reads the SANs verbatim (a minted cert could
// not carry an empty DNSName or a 3-byte IP address).
func TestNamesDropsAndDeduplicates(t *testing.T) {
	goodIP, err := core.ParseNetIP("1.2.3.4")
	core.AssertMustNoError(t, err, "parse ip")

	// a 3-byte address fails netip.AddrFromSlice and is dropped; the
	// parsed one is canonical (unmapped) and is bracketed.
	badIP := net.IP{1, 2, 3}

	cert := &x509.Certificate{
		DNSNames: []string{"a.example.com", "", "a.example.com", "*.example.com",
			"*.example.com"},
		IPAddresses: []net.IP{badIP, goodIP},
	}

	names, patterns := x509utils.Names(cert)
	core.AssertSliceEqual(t, core.S("a.example.com", "[1.2.3.4]"), names,
		"names")
	core.AssertSliceEqual(t, core.S(".example.com"), patterns, "patterns")
}

// TestNamesMappedIP confirms a 16-byte IPv4-mapped-IPv6 iPAddress SAN is
// stored under the canonical unmapped key ([1.2.3.4]), matching what the
// lookup paths produce for the same address, rather than [::ffff:1.2.3.4]
// (F61). A certificate literal carries the mapped SAN verbatim; a minted cert
// would normalise it to a 4-byte SAN.
func TestNamesMappedIP(t *testing.T) {
	mapped := net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 1, 2, 3, 4}
	cert := &x509.Certificate{IPAddresses: []net.IP{mapped}}

	names, _ := x509utils.Names(cert)
	core.AssertSliceEqual(t, core.S("[1.2.3.4]"), names, "names")
}

// namesIPTestCase feeds appendIPAddresses (through the exported Names) IP SANs
// of every boundary length. netip.AddrFromSlice accepts only 4- and 16-byte
// slices, both of which yield a valid address, so nothing shorter or longer
// survives and no ok-but-invalid address exists. That is why appendIPAddresses
// needs no IsValid() guard: these rows produce the same output with or without
// it, and would catch a bad address leaking through if one ever could (F72).
type namesIPTestCase struct {
	name  string
	ips   []net.IP
	names []string
}

func (tc namesIPTestCase) Name() string { return tc.name }

func (tc namesIPTestCase) Test(t *testing.T) {
	t.Helper()
	cert := &x509.Certificate{IPAddresses: tc.ips}
	names, _ := x509utils.Names(cert)
	core.AssertSliceEqual(t, tc.names, names, "names")
}

func newNamesIPTestCase(name string, ips []net.IP,
	names []string) namesIPTestCase {
	return namesIPTestCase{
		ips:   ips,
		names: names,
		name:  name,
	}
}

func namesIPTestCases() []namesIPTestCase {
	none := core.S[string]()
	return core.S(
		newNamesIPTestCase("nil element", core.S[net.IP](nil), none),
		newNamesIPTestCase("empty", []net.IP{{}}, none),
		newNamesIPTestCase("three bytes", []net.IP{{1, 2, 3}}, none),
		newNamesIPTestCase("ipv4", []net.IP{{1, 2, 3, 4}},
			core.S("[1.2.3.4]")),
		newNamesIPTestCase("five bytes", []net.IP{{1, 2, 3, 4, 5}}, none),
		newNamesIPTestCase("fifteen bytes", []net.IP{make([]byte, 15)}, none),
		newNamesIPTestCase("ipv4-mapped ipv6",
			[]net.IP{{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 1, 2, 3, 4}},
			core.S("[1.2.3.4]")),
		newNamesIPTestCase("ipv6", []net.IP{net.ParseIP("2001:db8::1")},
			core.S("[2001:db8::1]")),
		newNamesIPTestCase("seventeen bytes", []net.IP{make([]byte, 17)}, none),
	)
}

func TestNamesIPLengths(t *testing.T) {
	core.RunTestCases(t, namesIPTestCases())
}

// TestHostname confirms Hostname sanitises a URL host the same way SanitizeName
// does: it strips the port and folds case.
func TestHostname(t *testing.T) {
	u := &url.URL{Host: "WWW.Example.COM:443"}
	got, ok := x509utils.Hostname(u)
	core.AssertMustTrue(t, ok, "ok")
	core.AssertEqual(t, "www.example.com", got, "host")
}
