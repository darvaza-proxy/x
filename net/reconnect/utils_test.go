package reconnect

import (
	"testing"

	"darvaza.org/core"
)

// Test case for parseRemote function
type parseRemoteTestCase struct {
	name            string
	input           string
	expectedNetwork string
	expectedAddress string
}

func newParseRemoteTestCase(name, input, expectedNetwork, expectedAddress string) parseRemoteTestCase {
	return parseRemoteTestCase{
		name:            name,
		input:           input,
		expectedNetwork: expectedNetwork,
		expectedAddress: expectedAddress,
	}
}

func (tc parseRemoteTestCase) Name() string {
	return tc.name
}

func (tc parseRemoteTestCase) Test(t *testing.T) {
	t.Helper()

	network, address := parseRemote(tc.input)
	core.AssertEqual(t, tc.expectedNetwork, network, "network")
	core.AssertEqual(t, tc.expectedAddress, address, "address")
}

var _ core.TestCase = parseRemoteTestCase{}

func TestParseRemote(t *testing.T) {
	testCases := []parseRemoteTestCase{
		// TCP cases
		newParseRemoteTestCase("tcp with host and port", "example.com:8080", "tcp", "example.com:8080"),
		newParseRemoteTestCase("tcp with IP and port", "192.168.1.1:443", "tcp", "192.168.1.1:443"),
		newParseRemoteTestCase("tcp with localhost", "localhost:3000", "tcp", "localhost:3000"),

		// Unix socket cases with unix: prefix
		newParseRemoteTestCase("unix with prefix", "unix:/var/run/app.sock", "unix", "/var/run/app.sock"),
		newParseRemoteTestCase("unix with prefix no extension", "unix:/tmp/socket", "unix", "/tmp/socket"),
		newParseRemoteTestCase("unix with prefix relative", "unix:./app.sock", "unix", "./app.sock"),

		// Unix socket cases with absolute path
		newParseRemoteTestCase("unix absolute path", "/var/run/app.sock", "unix", "/var/run/app.sock"),
		newParseRemoteTestCase("unix absolute no extension", "/tmp/socket", "unix", "/tmp/socket"),
		newParseRemoteTestCase("unix absolute deep", "/usr/local/var/run/app.sock", "unix",
			"/usr/local/var/run/app.sock"),

		// Unix socket cases with .sock extension
		newParseRemoteTestCase("unix relative with .sock", "app.sock", "unix", "app.sock"),
		newParseRemoteTestCase("unix relative path with .sock", "./run/app.sock", "unix", "./run/app.sock"),
		newParseRemoteTestCase("unix relative nested with .sock", "var/run/app.sock", "unix", "var/run/app.sock"),
	}

	core.RunTestCases(t, testCases)
}
