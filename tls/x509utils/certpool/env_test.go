package certpool_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

var _ core.TestCase = envFnTestCase{}

// envFnTestCase exercises the argument and lookup guards of NewFromEnvFn: a
// missing accessor or name is invalid; an undefined or empty variable does
// not exist.
type envFnTestCase struct {
	getEnv  func(string) (string, bool)
	wantErr error
	name    string
	varName string
}

func (tc envFnTestCase) Name() string { return tc.name }

func (tc envFnTestCase) Test(t *testing.T) {
	t.Helper()
	_, err := certpool.NewFromEnvFn(tc.getEnv, tc.varName)
	core.AssertErrorIs(t, err, tc.wantErr, "error")
}

func newEnvFnTestCase(name string, getEnv func(string) (string, bool),
	varName string, wantErr error) envFnTestCase {
	return envFnTestCase{
		getEnv:  getEnv,
		name:    name,
		varName: varName,
		wantErr: wantErr,
	}
}

func envFnTestCases() []envFnTestCase {
	absent := func(string) (string, bool) { return "", false }
	empty := func(string) (string, bool) { return "", true }

	return core.S(
		newEnvFnTestCase("nil accessor", nil, "VAR", core.ErrInvalid),
		newEnvFnTestCase("empty name", absent, "", core.ErrInvalid),
		newEnvFnTestCase("undefined variable", absent, "VAR", core.ErrNotExists),
		newEnvFnTestCase("empty value", empty, "VAR", core.ErrNotExists),
	)
}

func TestNewFromEnvFnErrors(t *testing.T) {
	core.RunTestCases(t, envFnTestCases())
}

// TestNewFromEnvFnSuccess confirms a PEM value resolved through the accessor
// yields a populated pool.
func TestNewFromEnvFnSuccess(t *testing.T) {
	leaf := mkTestLeaf(t, "env.example.com", nil)
	getEnv := func(string) (string, bool) {
		return string(certPEM(t, leaf.cert)), true
	}

	pool, err := certpool.NewFromEnvFn(getEnv, "VAR")
	core.AssertMustNoError(t, err, "from env fn")
	core.AssertMustNotNil(t, pool, "pool")
	core.AssertEqual(t, 1, pool.Count(), "one cert")
}

// TestNewFromEnv confirms NewFromEnv reads the process environment: a set
// variable yields the pool, an empty one is treated as absent.
func TestNewFromEnv(t *testing.T) {
	leaf := mkTestLeaf(t, "process-env.example.com", nil)
	t.Setenv("CERTPOOL_TEST_PEM", string(certPEM(t, leaf.cert)))

	pool, err := certpool.NewFromEnv("CERTPOOL_TEST_PEM")
	core.AssertMustNoError(t, err, "from env")
	core.AssertEqual(t, 1, pool.Count(), "one cert")

	// an empty value counts as absent; set it explicitly so the assertion
	// does not depend on the ambient environment.
	t.Setenv("CERTPOOL_TEST_EMPTY", "")
	_, err = certpool.NewFromEnv("CERTPOOL_TEST_EMPTY")
	core.AssertErrorIs(t, err, core.ErrNotExists, "empty variable")
}

// TestNewFromStrings confirms PEM content is parsed into a pool and that empty
// input reports no certificates found.
func TestNewFromStrings(t *testing.T) {
	one := mkTestLeaf(t, "s-one.example.com", nil)
	two := mkTestLeaf(t, "s-two.example.com", nil)

	pool, err := certpool.NewFromStrings(string(certPEM(t, one.cert, two.cert)))
	core.AssertMustNoError(t, err, "from strings")
	core.AssertEqual(t, 2, pool.Count(), "two certs")

	// PEM that parses cleanly but holds no certificate (only a key) leaves the
	// pool empty, which reports ErrNoCertificatesFound.
	_, err = certpool.NewFromStrings(string(keyPEM(t, one)))
	core.AssertErrorIs(t, err, certpool.ErrNoCertificatesFound, "no certs")
}
