package reconnect_test

// cspell:words rhqv

import (
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/net/reconnect"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = parseRemoteTestCase{}
var _ core.TestCase = timeoutToAbsoluteTimeTestCase{}

// Test case for ParseRemote function
type parseRemoteTestCase struct {
	input           string
	expectedNetwork string
	expectedAddress string
	name            string
	wantErr         bool
}

func newParseRemoteTestCase(name, input, expectedNetwork, expectedAddress string, wantErr bool) parseRemoteTestCase {
	return parseRemoteTestCase{
		name:            name,
		input:           input,
		expectedNetwork: expectedNetwork,
		expectedAddress: expectedAddress,
		wantErr:         wantErr,
	}
}

func (tc parseRemoteTestCase) Name() string {
	return tc.name
}

func (tc parseRemoteTestCase) Test(t *testing.T) {
	network, address, err := reconnect.ParseRemote(tc.input)
	validateErr := reconnect.ValidateRemote(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "ParseRemote")
		core.AssertError(t, validateErr, "ValidateRemote")
		return
	}

	core.AssertNoError(t, err, "ParseRemote")
	core.AssertNoError(t, validateErr, "ValidateRemote")
	core.AssertEqual(t, tc.expectedNetwork, network, "network")
	core.AssertEqual(t, tc.expectedAddress, address, "address")
}

func makeParseRemoteTestCases() []core.TestCase {
	return []core.TestCase{
		// TCP cases
		newParseRemoteTestCase("tcp with host and port", "example.com:8080", "tcp", "example.com:8080", false),
		newParseRemoteTestCase("tcp with IP and port", "192.168.1.1:443", "tcp", "192.168.1.1:443", false),
		newParseRemoteTestCase("tcp with localhost", "localhost:3000", "tcp", "localhost:3000", false),
		newParseRemoteTestCase("tcp with .sock in hostname", "service.sock.example.com:8080", "tcp",
			"service.sock.example.com:8080", false),
		newParseRemoteTestCase("tcp with IPv6 and port", "[::1]:8080", "tcp", "[::1]:8080", false),
		newParseRemoteTestCase("tcp with IPv6 zone", "[fe80::1%eth0]:80", "tcp", "[fe80::1%eth0]:80", false),

		// TCP cases returning the address checked, in its canonical ASCII form
		newParseRemoteTestCase("tcp uppercase name", "Example.COM:8080", "tcp", "example.com:8080", false),
		newParseRemoteTestCase("tcp international name", "Hello.世界:443", "tcp", "hello.xn--rhqv96g:443", false),
		newParseRemoteTestCase("tcp name mapped into IPv4", "１２７。０。０。１:80", "tcp", "127.0.0.1:80", false),
		newParseRemoteTestCase("tcp padded port", "example.com:0080", "tcp", "example.com:80", false),

		// TCP cases with an empty or unspecified host, which dial the local system
		newParseRemoteTestCase("tcp empty host", ":8080", "tcp", ":8080", false),
		newParseRemoteTestCase("tcp unspecified IPv4", "0.0.0.0:80", "tcp", "0.0.0.0:80", false),
		newParseRemoteTestCase("tcp unspecified IPv4 short", "0:80", "tcp", "0.0.0.0:80", false),
		newParseRemoteTestCase("tcp unspecified IPv6", "[::]:80", "tcp", "[::]:80", false),
		newParseRemoteTestCase("tcp unspecified IPv6 with zone", "[::%eth0]:80", "tcp", "[::%eth0]:80", false),
		newParseRemoteTestCase("tcp unspecified IPv4-mapped", "[::ffff:0.0.0.0]:80", "tcp", "0.0.0.0:80", false),

		// Unix socket cases with unix: prefix
		newParseRemoteTestCase("unix with prefix", "unix:/var/run/app.sock", "unix", "/var/run/app.sock", false),
		newParseRemoteTestCase("unix with prefix no extension", "unix:/tmp/socket", "unix", "/tmp/socket", false),
		newParseRemoteTestCase("unix with prefix relative", "unix:./app.sock", "unix", "./app.sock", false),

		// Unix socket cases with absolute path
		newParseRemoteTestCase("unix absolute path", "/var/run/app.sock", "unix", "/var/run/app.sock", false),
		newParseRemoteTestCase("unix absolute no extension", "/tmp/socket", "unix", "/tmp/socket", false),
		newParseRemoteTestCase("unix absolute deep", "/usr/local/var/run/app.sock", "unix",
			"/usr/local/var/run/app.sock", false),

		// Unix socket cases with .sock extension
		newParseRemoteTestCase("unix relative with .sock", "app.sock", "unix", "app.sock", false),
		newParseRemoteTestCase("unix relative path with .sock", "./run/app.sock", "unix", "./run/app.sock", false),
		newParseRemoteTestCase("unix relative nested with .sock", "var/run/app.sock", "unix",
			"var/run/app.sock", false),

		// Unix socket cases with abstract sockets
		newParseRemoteTestCase("abstract socket", "@abstract-name", "unix", "@abstract-name", false),
		newParseRemoteTestCase("abstract socket with dash", "@my-service", "unix", "@my-service", false),

		// Error cases
		newParseRemoteTestCase("empty string", "", "", "", true),
		newParseRemoteTestCase("tcp missing port", "example.com", "", "", true),
		newParseRemoteTestCase("tcp colon no port", "example.com:", "", "", true),
		newParseRemoteTestCase("tcp port zero", "localhost:0", "", "", true),
		newParseRemoteTestCase("tcp padded port zero", "localhost:00", "", "", true),
		newParseRemoteTestCase("unix empty after prefix", "unix:", "", "", true),
		newParseRemoteTestCase("abstract socket empty name", "@", "", "", true),
		newParseRemoteTestCase("unix path with null byte", "/tmp/sock\x00et", "", "", true),
	}
}

func TestParseRemote(t *testing.T) {
	core.RunTestCases(t, makeParseRemoteTestCases())
}

// Test case for TimeoutToAbsoluteTime function
type timeoutToAbsoluteTimeTestCase struct {
	base     time.Time
	expected time.Time
	name     string
	duration time.Duration
}

func newTimeoutToAbsoluteTimeTestCase(name string, base time.Time, duration time.Duration,
	expected time.Time) timeoutToAbsoluteTimeTestCase {
	return timeoutToAbsoluteTimeTestCase{
		name:     name,
		base:     base,
		duration: duration,
		expected: expected,
	}
}

func (tc timeoutToAbsoluteTimeTestCase) Name() string {
	return tc.name
}

func (tc timeoutToAbsoluteTimeTestCase) Test(t *testing.T) {
	result := reconnect.TimeoutToAbsoluteTime(tc.base, tc.duration)
	core.AssertEqual(t, tc.expected, result, "timeout")
}

func makeTimeoutToAbsoluteTimeTestCases() []core.TestCase {
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	return []core.TestCase{
		// Positive duration with non-zero base
		newTimeoutToAbsoluteTimeTestCase("positive duration with base",
			baseTime, 5*time.Second, baseTime.Add(5*time.Second)),

		// Zero duration
		newTimeoutToAbsoluteTimeTestCase("zero duration",
			baseTime, 0, time.Time{}),

		// Negative duration
		newTimeoutToAbsoluteTimeTestCase("negative duration",
			baseTime, -5*time.Second, time.Time{}),
	}
}

func runTestTimeoutToAbsoluteTimePositiveDurationWithZeroBase(t *testing.T) {
	before := time.Now()
	result := reconnect.TimeoutToAbsoluteTime(time.Time{}, 5*time.Second)
	after := time.Now()

	// Result should be between before+5s and after+5s
	expectedMin := before.Add(5 * time.Second)
	expectedMax := after.Add(5 * time.Second)

	core.AssertTrue(t, result.After(expectedMin) || result.Equal(expectedMin), "result after min")
	core.AssertTrue(t, result.Before(expectedMax) || result.Equal(expectedMax), "result before max")
}

func TestTimeoutToAbsoluteTime(t *testing.T) {
	// Handle the special case where zero base time should use current time
	t.Run("positive duration with zero base", runTestTimeoutToAbsoluteTimePositiveDurationWithZeroBase)

	core.RunTestCases(t, makeTimeoutToAbsoluteTimeTestCases())
}
