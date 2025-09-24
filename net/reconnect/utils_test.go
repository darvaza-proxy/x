package reconnect

import (
	"testing"
	"time"

	"darvaza.org/core"
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
	t.Helper()

	network, address, err := ParseRemote(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "ParseRemote")
		return
	}

	core.AssertNoError(t, err, "ParseRemote")
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
		newParseRemoteTestCase("tcp empty host", ":8080", "", "", true),
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
	duration time.Duration
	expected time.Time
	name     string
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
	t.Helper()

	result := TimeoutToAbsoluteTime(tc.base, tc.duration)
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
	t.Helper()
	before := time.Now()
	result := TimeoutToAbsoluteTime(time.Time{}, 5*time.Second)
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
