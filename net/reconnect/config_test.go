package reconnect

import (
	"testing"

	"darvaza.org/core"
)

// Test case for validateRemote function
type validateRemoteTestCase struct {
	name    string
	input   string
	wantErr bool
}

func newValidateRemoteTestCase(name, input string, wantErr bool) validateRemoteTestCase {
	return validateRemoteTestCase{
		name:    name,
		input:   input,
		wantErr: wantErr,
	}
}

func (tc validateRemoteTestCase) Name() string {
	return tc.name
}

func (tc validateRemoteTestCase) Test(t *testing.T) {
	t.Helper()

	err := validateRemote(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "validateRemote")
	} else {
		core.AssertNoError(t, err, "validateRemote")
	}
}

var _ core.TestCase = validateRemoteTestCase{}

func TestValidateRemote(t *testing.T) {
	testCases := []validateRemoteTestCase{
		// Valid TCP addresses
		newValidateRemoteTestCase("valid tcp host:port", "example.com:8080", false),
		newValidateRemoteTestCase("valid tcp IP:port", "192.168.1.1:443", false),
		newValidateRemoteTestCase("valid tcp localhost:port", "localhost:3000", false),
		newValidateRemoteTestCase("valid tcp IPv6", "[::1]:8080", false),

		// Valid Unix socket addresses
		newValidateRemoteTestCase("valid unix with prefix", "unix:/var/run/app.sock", false),
		newValidateRemoteTestCase("valid unix absolute path", "/var/run/app.sock", false),
		newValidateRemoteTestCase("valid unix with .sock", "app.sock", false),
		newValidateRemoteTestCase("valid unix relative", "./run/app.sock", false),

		// Invalid addresses
		newValidateRemoteTestCase("empty string", "", true),
		newValidateRemoteTestCase("tcp missing port", "example.com", true),
		newValidateRemoteTestCase("tcp colon no port", "example.com:", true),
		newValidateRemoteTestCase("unix empty after prefix", "unix:", true),
	}

	core.RunTestCases(t, testCases)
}
