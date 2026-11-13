package net_test

// cspell:words rhqv

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/net"
)

var (
	_ core.TestCase = splitHostPortTestCase{}
	_ core.TestCase = cleanDialAddressTestCase{}
	_ core.TestCase = joinHostPortTestCase{}
	_ core.TestCase = joinAllHostPortsTestCase{}
)

type splitHostPortTestCase struct {
	name     string
	hostPort string
	host     string
	port     uint16
	wantErr  bool
}

func (tc splitHostPortTestCase) Name() string {
	return tc.name
}

func (tc splitHostPortTestCase) Test(t *testing.T) {
	host, port, err := net.SplitHostPort(tc.hostPort)
	if tc.wantErr {
		assertAddrError(t, err, tc.hostPort)
	} else {
		core.AssertNoError(t, err, "error")
	}
	core.AssertEqual(t, tc.host, host, "host")
	core.AssertEqual(t, tc.port, port, "port")
}

// assertAddrError asserts that err is a *net.AddrError naming addr, and
// stops the test when it is not one.
func assertAddrError(t *testing.T, err error, addr string) {
	addrErr := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
	core.AssertEqual(t, addr, addrErr.Addr, "error address")
}

func newSplitHostPortTestCase(name, hostPort, host string, port uint16) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostPort: hostPort,
		host:     host,
		port:     port,
		wantErr:  false,
	}
}

func newSplitHostPortTestCasePortOnly(name, hostPort string, port uint16) splitHostPortTestCase {
	return newSplitHostPortTestCase(name, hostPort, "", port)
}

func newSplitHostPortTestCaseRefused(name, hostPort string) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostPort: hostPort,
		host:     "",
		port:     0,
		wantErr:  true,
	}
}

func splitHostPortTestCases() []core.TestCase {
	return []core.TestCase{
		// The port, absent or given.
		newSplitHostPortTestCase("name", "name", "name", 0),
		newSplitHostPortTestCase("name and port", "name:1234", "name", 1234),
		newSplitHostPortTestCase("name and port 65535", "name:65535", "name", 65535),
		newSplitHostPortTestCase("name and padded port", "name:0080", "name", 80),

		// A name comes back mapped through UTS 46, in its ASCII form.
		newSplitHostPortTestCase("uppercase name", "Example.COM:80", "example.com", 80),
		newSplitHostPortTestCase("international name", "Hello.世界:80", "hello.xn--rhqv96g", 80),
		newSplitHostPortTestCase("punycode name", "hello.xn--rhqv96g:80", "hello.xn--rhqv96g", 80),
		newSplitHostPortTestCase("underscore label", "a_b.name:80", "a_b.name", 80),

		// Addresses.
		newSplitHostPortTestCase("IPv4 and port", "192.0.2.1:80", "192.0.2.1", 80),
		newSplitHostPortTestCase("ideographic full stops IPv4", "１２７。０。０。１:80", "127.0.0.1", 80),
		newSplitHostPortTestCase("bracketed IPv6 and port", "[::1]:1234", "::1", 1234),
		newSplitHostPortTestCase("unspecified IPv6", "[::]:6060", "::", 6060),

		// A port alone splits with an empty host.
		newSplitHostPortTestCasePortOnly("port only", ":6060", 6060),
		newSplitHostPortTestCasePortOnly("port 0 only", ":0", 0),

		// Refused.
		newSplitHostPortTestCaseRefused("space", "bad name:80"),
		newSplitHostPortTestCaseRefused("numeric last label", "name.1:80"),
		newSplitHostPortTestCaseRefused("invalid punycode", "xn--a.name:80"),
		newSplitHostPortTestCaseRefused("port out of range", "name:65536"),
		newSplitHostPortTestCaseRefused("empty", ""),
	}
}

func TestSplitHostPort(t *testing.T) {
	core.RunTestCases(t, splitHostPortTestCases())
}

type cleanDialAddressTestCase struct {
	name     string
	hostPort string
	out      string
}

func (tc cleanDialAddressTestCase) Name() string {
	return tc.name
}

// wantErr tells whether the row expects an error. A cleaned address has
// at least ":port", so a row with an empty out expects one.
func (tc cleanDialAddressTestCase) wantErr() bool {
	return tc.out == ""
}

func (tc cleanDialAddressTestCase) Test(t *testing.T) {
	out, err := net.CleanDialAddress(tc.hostPort)
	if tc.wantErr() {
		assertAddrError(t, err, tc.hostPort)
	} else {
		core.AssertNoError(t, err, "error")
	}
	core.AssertEqual(t, tc.out, out, "cleaned")
}

func newCleanDialAddressTestCase(name, hostPort, out string) cleanDialAddressTestCase {
	return cleanDialAddressTestCase{
		name:     name,
		hostPort: hostPort,
		out:      mustNotZero(out, "cleanDialAddressTestCase: every accepted row must state what it cleans to"),
	}
}

func newCleanDialAddressTestCaseRefused(name, hostPort string) cleanDialAddressTestCase {
	return cleanDialAddressTestCase{
		name:     name,
		hostPort: hostPort,
		out:      "",
	}
}

func cleanDialAddressTestCases() []core.TestCase {
	return []core.TestCase{
		// Normalised as SplitHostPortASCII gives the host and port.
		newCleanDialAddressTestCase("name and port", "name:80", "name:80"),
		newCleanDialAddressTestCase("uppercase name", "Example.COM:80", "example.com:80"),
		newCleanDialAddressTestCase("international name", "Hello.世界:80", "hello.xn--rhqv96g:80"),
		newCleanDialAddressTestCase("padded port", "name:0080", "name:80"),
		newCleanDialAddressTestCase("ideographic full stops IPv4", "１２７。０。０。１:80", "127.0.0.1:80"),
		newCleanDialAddressTestCase("IPv4", "192.0.2.1:80", "192.0.2.1:80"),
		newCleanDialAddressTestCase("IPv6", "[::1]:80", "[::1]:80"),
		newCleanDialAddressTestCase("non-canonical IPv6", "[2001:DB8:0::1]:80", "[2001:db8::1]:80"),
		newCleanDialAddressTestCase("IPv6 with zone", "[fe80::1%eth0]:80", "[fe80::1%eth0]:80"),

		// An IPv4-mapped address without a zone comes back as the IPv4
		// address it wraps.
		newCleanDialAddressTestCase("IPv4-mapped", "[::ffff:192.0.2.1]:80", "192.0.2.1:80"),
		newCleanDialAddressTestCase("IPv4-mapped with zone", "[::ffff:192.0.2.1%eth0]:80",
			"[::ffff:192.0.2.1%eth0]:80"),

		// The split takes brackets around any host.
		newCleanDialAddressTestCase("bracketed IPv4", "[192.0.2.1]:80", "192.0.2.1:80"),

		// The split refuses an IPv4 address with a zone, and a name with
		// a trailing dot.
		newCleanDialAddressTestCaseRefused("IPv4 with zone", "192.0.2.1%eth0:80"),
		newCleanDialAddressTestCaseRefused("trailing dot", "example.com.:80"),

		// An empty or unspecified host, which dials the local system.
		newCleanDialAddressTestCase("port only", ":80", ":80"),
		newCleanDialAddressTestCase("unspecified IPv4", "0.0.0.0:80", "0.0.0.0:80"),
		newCleanDialAddressTestCase("unspecified IPv4 short", "0:80", "0.0.0.0:80"),
		newCleanDialAddressTestCase("unspecified IPv6", "[::]:80", "[::]:80"),
		newCleanDialAddressTestCase("unspecified IPv6 with zone", "[::%eth0]:80", "[::%eth0]:80"),
		newCleanDialAddressTestCase("unspecified IPv4-mapped", "[::ffff:0.0.0.0]:80", "0.0.0.0:80"),

		// Refused.
		newCleanDialAddressTestCaseRefused("name without port", "name"),
		newCleanDialAddressTestCaseRefused("IPv6 without port", "[::1]"),
		newCleanDialAddressTestCaseRefused("IPv6 without brackets", "::1"),
		newCleanDialAddressTestCaseRefused("port 0", "name:0"),
		newCleanDialAddressTestCaseRefused("padded port 0", "name:00"),
		newCleanDialAddressTestCaseRefused("port 0 only", ":0"),
		newCleanDialAddressTestCaseRefused("empty port", "name:"),
		newCleanDialAddressTestCaseRefused("port out of range", "name:65536"),
		newCleanDialAddressTestCaseRefused("named port", "name:http"),
		newCleanDialAddressTestCaseRefused("space", "bad name:80"),
		newCleanDialAddressTestCaseRefused("empty", ""),
	}
}

func TestCleanDialAddress(t *testing.T) {
	core.RunTestCases(t, cleanDialAddressTestCases())
}

type joinHostPortTestCase struct {
	name string
	host string
	out  string
	port uint16
}

func (tc joinHostPortTestCase) Name() string {
	return tc.name
}

// wantErr tells whether the row expects an error. A join gives at least
// ":port", so a row with an empty out expects one.
func (tc joinHostPortTestCase) wantErr() bool {
	return tc.out == ""
}

func (tc joinHostPortTestCase) Test(t *testing.T) {
	out, err := net.JoinHostPort(tc.host, tc.port)
	if tc.wantErr() {
		assertAddrError(t, err, tc.host)
	} else {
		core.AssertNoError(t, err, "error")
	}
	core.AssertEqual(t, tc.out, out, "joined")
}

// mustNotZero returns value, and panics with message when it is the zero
// value of its type.
func mustNotZero[T any](value T, message string) T {
	if core.IsZero(value) {
		panic(message)
	}
	return value
}

func newJoinHostPortTestCase(name, host string, port uint16, out string) joinHostPortTestCase {
	return joinHostPortTestCase{
		name: name,
		host: host,
		port: port,
		out:  mustNotZero(out, "joinHostPortTestCase: every joining row must state what it joins as"),
	}
}

func newJoinHostPortTestCaseRefused(name, host string, port uint16) joinHostPortTestCase {
	return joinHostPortTestCase{
		name: name,
		host: host,
		port: port,
		out:  "",
	}
}

func joinHostPortTestCases() []core.TestCase {
	return []core.TestCase{
		// The port.
		newJoinHostPortTestCase("name and port", "name", 1234, "name:1234"),
		newJoinHostPortTestCase("name and port 0", "name", 0, "name:0"),

		// A name joins in its canonical ASCII form.
		newJoinHostPortTestCase("uppercase name", "Example.COM", 80, "example.com:80"),
		newJoinHostPortTestCase("international name", "hello.世界", 80, "hello.xn--rhqv96g:80"),

		// Addresses.
		newJoinHostPortTestCase("IPv4", "192.0.2.1", 80, "192.0.2.1:80"),
		newJoinHostPortTestCase("IPv6", "::1", 1234, "[::1]:1234"),
		newJoinHostPortTestCase("non-canonical IPv6", "2001:DB8:0::1", 80, "[2001:db8::1]:80"),
		newJoinHostPortTestCase("IPv6 with zone", "fe80::1%eth0", 80, "[fe80::1%eth0]:80"),

		// An IPv4-mapped address without a zone joins as the IPv4
		// address it wraps.
		newJoinHostPortTestCase("IPv4-mapped", "::ffff:192.0.2.1", 80, "192.0.2.1:80"),
		newJoinHostPortTestCase("IPv4-mapped with zone", "::ffff:192.0.2.1%eth0", 80,
			"[::ffff:192.0.2.1%eth0]:80"),

		// The split takes brackets around any host.
		newJoinHostPortTestCase("bracketed IPv6", "[::1]", 80, "[::1]:80"),
		newJoinHostPortTestCase("bracketed IPv4", "[192.0.2.1]", 80, "192.0.2.1:80"),

		// An unspecified address keeps its family.
		newJoinHostPortTestCase("unspecified IPv4", "0.0.0.0", 6060, "0.0.0.0:6060"),
		newJoinHostPortTestCase("unspecified IPv4 short", "0", 6060, "0.0.0.0:6060"),
		newJoinHostPortTestCase("unspecified IPv6", "::", 6060, "[::]:6060"),
		newJoinHostPortTestCase("unspecified IPv4-mapped", "::ffff:0.0.0.0", 6060, "0.0.0.0:6060"),
		newJoinHostPortTestCase("unspecified IPv6 with zone", "::%eth0", 6060, "[::%eth0]:6060"),
		newJoinHostPortTestCase("unspecified IPv4 and port 0", "0", 0, "0.0.0.0:0"),

		// No host at all joins as the port alone.
		newJoinHostPortTestCase("empty host", "", 6060, ":6060"),

		// Refused.
		newJoinHostPortTestCaseRefused("unclosed bracket", "[::1", 80),
		newJoinHostPortTestCaseRefused("space", "bad name", 80),
		newJoinHostPortTestCaseRefused("numeric last label", "name.1", 80),
		newJoinHostPortTestCaseRefused("invalid punycode", "xn--a.name", 80),
		newJoinHostPortTestCaseRefused("punycode of an ASCII label", "xn--ab-.name", 80),
		newJoinHostPortTestCaseRefused("IPv4 with zone", "192.0.2.1%eth0", 80),
		newJoinHostPortTestCaseRefused("trailing dot", "example.com.", 80),
		newJoinHostPortTestCaseRefused("host with another port", "name:8080", 80),
		newJoinHostPortTestCaseRefused("host with the same port", "name:8080", 8080),
	}
}

func TestJoinHostPort(t *testing.T) {
	core.RunTestCases(t, joinHostPortTestCases())
}

type joinAllHostPortsTestCase struct {
	name      string
	refused   string
	addresses []string
	ports     []uint16
	out       []string
}

func (tc joinAllHostPortsTestCase) Name() string {
	return tc.name
}

func (tc joinAllHostPortsTestCase) Test(t *testing.T) {
	out, err := net.JoinAllHostPorts(tc.addresses, tc.ports)
	if tc.refused != "" {
		// a row naming the address refused expects an error
		assertAddrError(t, err, tc.refused)
	} else {
		core.AssertNoError(t, err, "error")
	}
	core.AssertSliceEqual(t, tc.out, out, "joined")
}

func newJoinAllHostPortsTestCase(name string, addresses []string, ports []uint16,
	out []string) joinAllHostPortsTestCase {
	return joinAllHostPortsTestCase{
		name:      name,
		refused:   "",
		addresses: addresses,
		ports:     ports,
		out:       out,
	}
}

// newJoinAllHostPortsTestCaseRefused takes the address refused, and the
// host:port pairs joined before it, which JoinAllHostPorts returns with
// the error.
func newJoinAllHostPortsTestCaseRefused(name string, addresses []string, ports []uint16,
	refused string, out []string) joinAllHostPortsTestCase {
	return joinAllHostPortsTestCase{
		name:      name,
		refused:   mustNotZero(refused, "joinAllHostPortsTestCase: every refused row must name the address refused"),
		addresses: addresses,
		ports:     ports,
		out:       out,
	}
}

func joinAllHostPortsTestCases() []core.TestCase {
	return []core.TestCase{
		newJoinAllHostPortsTestCase("every address with every port",
			core.S("name", "::1"), core.S[uint16](80, 443),
			core.S("name:80", "name:443", "[::1]:80", "[::1]:443")),
		newJoinAllHostPortsTestCaseRefused("refused after a joined address",
			core.S("name", "xn--ab-.name", "::1"), core.S[uint16](80),
			"xn--ab-.name", core.S("name:80")),
		newJoinAllHostPortsTestCaseRefused("first address refused",
			core.S("xn--ab-.name", "name"), core.S[uint16](80),
			"xn--ab-.name", nil),
		// Every address is checked, with no ports to join it with.
		newJoinAllHostPortsTestCaseRefused("refused with no ports",
			core.S("name", "xn--ab-.name"), core.S[uint16](),
			"xn--ab-.name", nil),
	}
}

func TestJoinAllHostPorts(t *testing.T) {
	core.RunTestCases(t, joinAllHostPortsTestCases())
}
