package tls_test

// cspell:words stdtls

import (
	stdtls "crypto/tls"
	"net"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls"
)

var _ core.TestCase = splitCHITestCase{}

// stubAddr is a net.Addr that is neither TCP nor UDP and exposes no typed
// address, so core.AddrPort cannot resolve it.
type stubAddr struct{}

func (stubAddr) Network() string { return "stub" }
func (stubAddr) String() string  { return "stub-addr" }

// stubConn is a net.Conn whose LocalAddr is fixed; every other method is
// inherited from the nil embedded Conn and must not be called.
type stubConn struct {
	net.Conn
	local net.Addr
}

func (c stubConn) LocalAddr() net.Addr { return c.local }

// splitCHITestCase exercises SplitClientHelloInfo's serverName derivation: the
// explicit SNI passes through, and an empty SNI falls back to the connection's
// local address in the bracketed form the store keys under, with any 4-in-6
// mapping and scope zone removed.
type splitCHITestCase struct {
	conn       net.Conn
	name       string
	serverName string
	wantName   string
}

func (tc splitCHITestCase) Name() string { return tc.name }

func (tc splitCHITestCase) Test(t *testing.T) {
	t.Helper()

	chi := &stdtls.ClientHelloInfo{ServerName: tc.serverName, Conn: tc.conn}
	_, name, err := tls.SplitClientHelloInfo(chi)
	core.AssertMustNoError(t, err, "err")
	core.AssertEqual(t, tc.wantName, name, "serverName")
}

func newSplitCHITestCase(name, serverName string, conn net.Conn,
	wantName string) splitCHITestCase {
	return splitCHITestCase{
		conn:       conn,
		name:       name,
		serverName: serverName,
		wantName:   wantName,
	}
}

func tcpAddr(ip string, zone string) *net.TCPAddr {
	return &net.TCPAddr{IP: net.ParseIP(ip), Zone: zone, Port: 443}
}

// TestSplitClientHelloInfo covers the F60 fallback: the typed local address is
// taken through core.AddrPort, which selects the IP, removes any 4-in-6
// mapping and drops the scope zone, landing on the bracketed key form. A
// connection with no resolvable address yields an empty serverName rather than
// a bracketed string.
func TestSplitClientHelloInfo(t *testing.T) {
	cases := []splitCHITestCase{
		newSplitCHITestCase("explicit SNI", "example.com", nil,
			"example.com"),
		newSplitCHITestCase("IPv4 local address", "",
			stubConn{local: tcpAddr("192.0.2.1", "")}, "[192.0.2.1]"),
		newSplitCHITestCase("4-in-6 unmapped", "",
			stubConn{local: &net.TCPAddr{IP: net.IPv4(192, 0, 2, 2),
				Port: 443}}, "[192.0.2.2]"),
		newSplitCHITestCase("IPv6 zone dropped", "",
			stubConn{local: tcpAddr("fe80::1", "eth0")}, "[fe80::1]"),
		newSplitCHITestCase("nil Conn", "", nil, ""),
		newSplitCHITestCase("unresolvable local address", "",
			stubConn{local: stubAddr{}}, ""),
	}

	core.RunTestCases(t, cases)
}

// TestSplitClientHelloInfoNil confirms a nil ClientHelloInfo is rejected.
func TestSplitClientHelloInfoNil(t *testing.T) {
	_, name, err := tls.SplitClientHelloInfo(nil)
	core.AssertError(t, err, "nil chi")
	core.AssertEqual(t, "", name, "serverName")
}
