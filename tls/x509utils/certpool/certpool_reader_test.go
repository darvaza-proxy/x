package certpool_test

import (
	"context"
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

var _ core.TestCase = getTestCase{}

// mkWildcardLeaf builds a self-signed leaf whose only SAN is a wildcard, so it
// is reachable through the pattern index rather than an exact name.
func mkWildcardLeaf(t *testing.T, wildcard string) testCert {
	t.Helper()
	tmpl := baseTemplate(t, wildcard, []string{wildcard})
	return signAndParse(t, tmpl, nil)
}

// getTestCase exercises Get's name resolution against a shared pool: exact
// names, wildcard suffixes and IP SANs resolve to their certificate; a bad
// name and an absent name surface their errors.
type getTestCase struct {
	pool    *certpool.CertPool
	want    *x509.Certificate
	wantErr error
	name    string
	query   string
}

func (tc getTestCase) Name() string { return tc.name }

func (tc getTestCase) Test(t *testing.T) {
	t.Helper()
	got, err := tc.pool.Get(context.Background(), tc.query)
	if tc.wantErr != nil {
		core.AssertErrorIs(t, err, tc.wantErr, "error")
		core.AssertNil(t, got, "cert")
		return
	}
	core.AssertMustNoError(t, err, "get")
	core.AssertMustNotNil(t, got, "cert")
	core.AssertTrue(t, got.Equal(tc.want), "cert")
}

func newGetTestCase(name string, pool *certpool.CertPool, query string,
	want *x509.Certificate, wantErr error) getTestCase {
	return getTestCase{
		pool:    pool,
		want:    want,
		name:    name,
		query:   query,
		wantErr: wantErr,
	}
}

func getTestCases(t *testing.T) []getTestCase {
	t.Helper()

	exact := mkTestLeaf(t, "exact.example.com", nil)
	wild := mkWildcardLeaf(t, "*.wild.example.com")
	ipLeaf := mkIPLeaf(t, "203.0.113.7")

	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(exact.cert), "add exact")
	core.AssertMustTrue(t, pool.AddCert(wild.cert), "add wildcard")
	core.AssertMustTrue(t, pool.AddCert(ipLeaf.cert), "add ip")

	return core.S(
		newGetTestCase("exact name", pool, "exact.example.com", exact.cert, nil),
		newGetTestCase("wildcard suffix", pool, "host.wild.example.com",
			wild.cert, nil),
		newGetTestCase("ip address", pool, "203.0.113.7", ipLeaf.cert, nil),
		newGetTestCase("absent name", pool, "absent.example.com", nil,
			core.ErrNotExists),
		newGetTestCase("bad name", pool, "", nil, core.ErrInvalid),
	)
}

func TestGet(t *testing.T) {
	core.RunTestCases(t, getTestCases(t))
}

// TestGetNilReceiver confirms a nil pool reports ErrNilReceiver rather than
// panicking.
func TestGetNilReceiver(t *testing.T) {
	got, err := (*certpool.CertPool)(nil).Get(context.Background(), "x")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error")
	core.AssertNil(t, got, "cert")
}

// TestGetCancelledContext confirms a cancelled context short-circuits the
// lookup after the name check.
func TestGetCancelledContext(t *testing.T) {
	leaf := mkTestLeaf(t, "cancel.example.com", nil)
	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(leaf.cert), "add")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := pool.Get(ctx, "cancel.example.com")
	core.AssertErrorIs(t, err, context.Canceled, "error")
	core.AssertNil(t, got, "cert")
}

// TestExportCachesAndInvalidates confirms Export assembles an x509.CertPool,
// caches it across calls, and rebuilds it after the pool changes.
func TestExportCachesAndInvalidates(t *testing.T) {
	leaf := mkTestLeaf(t, "export.example.com", nil)
	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(leaf.cert), "add")

	first := pool.Export()
	core.AssertMustNotNil(t, first, "export")
	core.AssertSame(t, first, pool.Export(), "cached")

	other := mkTestLeaf(t, "export2.example.com", nil)
	core.AssertMustTrue(t, pool.AddCert(other.cert), "add second")
	core.AssertNotSame(t, first, pool.Export(), "rebuilt after change")
}

// TestExportNilReceiver confirms a nil pool still yields a usable empty pool.
func TestExportNilReceiver(t *testing.T) {
	got := (*certpool.CertPool)(nil).Export()
	core.AssertNotNil(t, got, "export")
}

// TestForEachVisitsAll confirms ForEach visits every certificate exactly once.
func TestForEachVisitsAll(t *testing.T) {
	pool, certs := mkPopulatedPool(t)

	var visited int
	pool.ForEach(context.Background(),
		func(_ context.Context, _ *x509.Certificate) bool {
			visited++
			return true
		})
	core.AssertEqual(t, len(certs), visited, "visited all")
}

// TestForEachAborts confirms returning false stops the walk early.
func TestForEachAborts(t *testing.T) {
	pool, _ := mkPopulatedPool(t)

	var visited int
	pool.ForEach(context.Background(),
		func(_ context.Context, _ *x509.Certificate) bool {
			visited++
			return false
		})
	core.AssertEqual(t, 1, visited, "stopped after first")
}

// TestForEachCancelledMidWalk confirms cancelling the context between
// iterations stops the walk: the callback cancels on its first call and the
// second iteration never runs.
func TestForEachCancelledMidWalk(t *testing.T) {
	pool, certs := mkPopulatedPool(t)
	core.AssertMustTrue(t, len(certs) > 1, "need multiple certs")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var visited int
	pool.ForEach(ctx, func(_ context.Context, _ *x509.Certificate) bool {
		visited++
		cancel()
		return true
	})
	core.AssertEqual(t, 1, visited, "stopped after cancel")
}

// TestForEachNoOp confirms the guard arms: a nil receiver, a nil callback and a
// pre-cancelled context each visit nothing.
func TestForEachNoOp(t *testing.T) {
	pool, _ := mkPopulatedPool(t)
	ctx := context.Background()

	(*certpool.CertPool)(nil).ForEach(ctx,
		func(context.Context, *x509.Certificate) bool { return true })
	pool.ForEach(ctx, nil)

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	var visited int
	pool.ForEach(cancelled, func(_ context.Context, _ *x509.Certificate) bool {
		visited++
		return true
	})
	core.AssertEqual(t, 0, visited, "cancelled visits nothing")
}
