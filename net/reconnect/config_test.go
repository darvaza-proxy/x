package reconnect

import (
	"context"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = validateRemoteTestCase{}

// Test case for ValidateRemote function
type validateRemoteTestCase struct {
	input   string
	name    string
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

	err := ValidateRemote(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "ValidateRemote")
	} else {
		core.AssertNoError(t, err, "ValidateRemote")
	}
}

func makeValidateRemoteTestCases() []core.TestCase {
	return []core.TestCase{
		// Valid TCP addresses
		newValidateRemoteTestCase("valid tcp host:port", "example.com:8080", false),
		newValidateRemoteTestCase("valid tcp IP:port", "192.168.1.1:443", false),
		newValidateRemoteTestCase("valid tcp localhost:port", "localhost:3000", false),
		newValidateRemoteTestCase("valid tcp IPv6", "[::1]:8080", false),

		// Valid Unix socket addresses
		newValidateRemoteTestCase("valid unix with prefix", "unix:/var/run/app.sock", false),
		newValidateRemoteTestCase("valid unix with triple slash", "unix:///var/run/app.sock", false),
		newValidateRemoteTestCase("valid unix absolute path", "/var/run/app.sock", false),
		newValidateRemoteTestCase("valid unix with .sock", "app.sock", false),
		newValidateRemoteTestCase("valid unix relative", "./run/app.sock", false),

		// Valid TCP addresses (edge cases)
		newValidateRemoteTestCase("valid tcp hostname ending with .sock", "foo.sock:443", false),

		// Invalid addresses
		newValidateRemoteTestCase("empty string", "", true),
		newValidateRemoteTestCase("tcp missing port", "example.com", true),
		newValidateRemoteTestCase("tcp colon no port", "example.com:", true),
		newValidateRemoteTestCase("unix empty after prefix", "unix:", true),
	}
}

func TestValidateRemote(t *testing.T) {
	core.RunTestCases(t, makeValidateRemoteTestCases())
}

func runTestConfigValidValidConfig(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "example.com:8080",
	}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "set defaults")

	err = cfg.Valid()
	core.AssertNoError(t, err, "valid config")
}

func runTestConfigValidInvalidRemote(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "invalid-remote",
	}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "set defaults")

	err = cfg.Valid()
	core.AssertError(t, err, "invalid remote")
}

func runTestConfigValidMissingContext(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Remote: "example.com:8080",
	}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "set defaults")

	// Manually clear context to test validation
	cfg.Context = nil
	err = cfg.Valid()
	core.AssertError(t, err, "missing context")
}

func runTestConfigValidMissingLogger(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "example.com:8080",
	}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "set defaults")

	// Manually clear logger to test validation
	cfg.Logger = nil
	err = cfg.Valid()
	core.AssertError(t, err, "missing logger")
}

func runTestConfigValidMissingWaitReconnect(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "example.com:8080",
	}
	err := cfg.SetDefaults()
	core.AssertNoError(t, err, "set defaults")

	// Manually clear wait reconnect to test validation
	cfg.WaitReconnect = nil
	err = cfg.Valid()
	core.AssertError(t, err, "missing wait reconnect")
}

func TestConfigValid(t *testing.T) {
	t.Run("valid config", runTestConfigValidValidConfig)
	t.Run("invalid remote", runTestConfigValidInvalidRemote)
	t.Run("missing context", runTestConfigValidMissingContext)
	t.Run("missing logger", runTestConfigValidMissingLogger)
	t.Run("missing wait reconnect", runTestConfigValidMissingWaitReconnect)
}

func runTestNewValidConfig(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "example.com:8080",
	}

	client, err := New(cfg)
	core.AssertNoError(t, err, "new client")
	core.AssertNotNil(t, client, "client")
}

func runTestNewUnixSocketConfig(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "/var/run/app.sock",
	}

	client, err := New(cfg)
	core.AssertNoError(t, err, "new unix client")
	core.AssertNotNil(t, client, "client")
}

func runTestNewInvalidConfig(t *testing.T) {
	t.Helper()
	cfg := &Config{
		Context: context.Background(),
		Remote:  "invalid-remote",
	}

	client, err := New(cfg)
	core.AssertError(t, err, "new client with invalid config")
	core.AssertNil(t, client, "client should be nil")
}

func TestNew(t *testing.T) {
	t.Run("valid config", runTestNewValidConfig)
	t.Run("unix socket config", runTestNewUnixSocketConfig)
	t.Run("invalid config", runTestNewInvalidConfig)
}
