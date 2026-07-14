package certpool_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

var (
	_ core.TestCase = deleteCertErrorTestCase{}
	_ core.TestCase = deleteErrorTestCase{}
	_ core.TestCase = putErrorTestCase{}
)

// keyPEM marshals a leaf's private key as a PKCS8 "PRIVATE KEY" block, so
// import paths can be handed a non-certificate block to ignore.
func keyPEM(t *testing.T, tc testCert) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(tc.key)
	core.AssertMustNoError(t, err, "marshal key")
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

// TestPutStoresAndAliases confirms Put stores a certificate under a sanitised
// name and that a second Put adds a further alias reachable by Get.
func TestPutStoresAndAliases(t *testing.T) {
	leaf := mkTestLeaf(t, "put.example.com", nil)
	pool := certpool.New()
	ctx := context.Background()

	core.AssertMustNoError(t, pool.Put(ctx, "put.example.com", leaf.cert),
		"put")
	core.AssertMustNoError(t, pool.Put(ctx, "alias.example.com", leaf.cert),
		"alias")

	got, err := pool.Get(ctx, "alias.example.com")
	core.AssertMustNoError(t, err, "get alias")
	core.AssertMustNotNil(t, got, "cert")
	core.AssertTrue(t, got.Equal(leaf.cert), "same cert by alias")
}

// TestPutDuplicateReportsExists confirms re-putting the same name for the same
// certificate is a no-op reported as ErrExists.
func TestPutDuplicateReportsExists(t *testing.T) {
	leaf := mkTestLeaf(t, "dup.example.com", nil)
	pool := certpool.New()
	ctx := context.Background()

	core.AssertMustNoError(t, pool.Put(ctx, "dup.example.com", leaf.cert),
		"put")
	err := pool.Put(ctx, "dup.example.com", leaf.cert)
	core.AssertErrorIs(t, err, core.ErrExists, "duplicate")
}

// putErrorTestCase exercises Put's rejection arms: a bad name, an invalid
// certificate, a nil receiver and a cancelled context each surface their error
// without storing anything.
type putErrorTestCase struct {
	pool     *certpool.CertPool
	cert     *x509.Certificate
	wantErr  error
	name     string
	caseName string
	cancel   bool
}

func (tc putErrorTestCase) Name() string { return tc.caseName }

func (tc putErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if tc.cancel {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
	}
	core.AssertErrorIs(t, tc.pool.Put(ctx, tc.name, tc.cert), tc.wantErr,
		"error")
}

// revive:disable:argument-limit
func newPutErrorTestCase(caseName string, pool *certpool.CertPool,
	name string, cert *x509.Certificate, cancel bool,
	wantErr error) putErrorTestCase {
	// revive:enable:argument-limit
	return putErrorTestCase{
		pool:     pool,
		cert:     cert,
		name:     name,
		caseName: caseName,
		wantErr:  wantErr,
		cancel:   cancel,
	}
}

func putErrorTestCases(t *testing.T) []putErrorTestCase {
	t.Helper()
	cert := mkTestLeaf(t, "err.example.com", nil).cert
	pool := certpool.New()

	return core.S(
		newPutErrorTestCase("bad name", pool, "", cert, false, core.ErrInvalid),
		newPutErrorTestCase("invalid cert", pool, "ok.example.com", nil, false,
			core.ErrInvalid),
		newPutErrorTestCase("nil receiver", nil, "ok.example.com", cert, false,
			core.ErrNilReceiver),
		newPutErrorTestCase("cancelled", pool, "ok.example.com", cert, true,
			context.Canceled),
	)
}

func TestPutErrors(t *testing.T) {
	core.RunTestCases(t, putErrorTestCases(t))
}

// TestDeleteRemovesByName confirms Delete removes every certificate held under
// a name, after which Get misses.
func TestDeleteRemovesByName(t *testing.T) {
	leaf := mkTestLeaf(t, "del.example.com", nil)
	pool := certpool.New()
	ctx := context.Background()
	core.AssertMustNoError(t, pool.Put(ctx, "del.example.com", leaf.cert),
		"put")

	core.AssertMustNoError(t, pool.Delete(ctx, "del.example.com"), "delete")

	_, err := pool.Get(ctx, "del.example.com")
	core.AssertErrorIs(t, err, core.ErrNotExists, "gone")
}

// deleteErrorTestCase exercises Delete's rejection arms: a bad name, a nil
// receiver, a cancelled context and an absent (but valid) name.
type deleteErrorTestCase struct {
	pool     *certpool.CertPool
	wantErr  error
	name     string
	caseName string
	cancel   bool
}

func (tc deleteErrorTestCase) Name() string { return tc.caseName }

func (tc deleteErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if tc.cancel {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
	}
	core.AssertErrorIs(t, tc.pool.Delete(ctx, tc.name), tc.wantErr, "error")
}

func newDeleteErrorTestCase(caseName string, pool *certpool.CertPool,
	name string, cancel bool, wantErr error) deleteErrorTestCase {
	return deleteErrorTestCase{
		pool:     pool,
		name:     name,
		caseName: caseName,
		wantErr:  wantErr,
		cancel:   cancel,
	}
}

func deleteErrorTestCases(t *testing.T) []deleteErrorTestCase {
	t.Helper()
	pool := certpool.New()

	return core.S(
		newDeleteErrorTestCase("bad name", pool, "", false, core.ErrInvalid),
		newDeleteErrorTestCase("nil receiver", nil, "ok.example.com", false,
			core.ErrNilReceiver),
		newDeleteErrorTestCase("cancelled", pool, "ok.example.com", true,
			context.Canceled),
		newDeleteErrorTestCase("absent name", pool, "absent.example.com", false,
			core.ErrNotExists),
	)
}

func TestDeleteErrors(t *testing.T) {
	core.RunTestCases(t, deleteErrorTestCases(t))
}

// deleteCertErrorTestCase exercises DeleteCert's rejection arms: an invalid
// certificate, a nil receiver, a cancelled context and an absent certificate.
type deleteCertErrorTestCase struct {
	pool     *certpool.CertPool
	cert     *x509.Certificate
	wantErr  error
	caseName string
	cancel   bool
}

func (tc deleteCertErrorTestCase) Name() string { return tc.caseName }

func (tc deleteCertErrorTestCase) Test(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if tc.cancel {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
	}
	core.AssertErrorIs(t, tc.pool.DeleteCert(ctx, tc.cert), tc.wantErr, "error")
}

func newDeleteCertErrorTestCase(caseName string, pool *certpool.CertPool,
	cert *x509.Certificate, cancel bool, wantErr error) deleteCertErrorTestCase {
	return deleteCertErrorTestCase{
		pool:     pool,
		cert:     cert,
		caseName: caseName,
		wantErr:  wantErr,
		cancel:   cancel,
	}
}

func deleteCertErrorTestCases(t *testing.T) []deleteCertErrorTestCase {
	t.Helper()
	pool, _ := mkPopulatedPool(t)
	valid := mkTestLeaf(t, "valid.example.com", nil).cert
	absent := mkTestLeaf(t, "absent.example.com", nil).cert

	return core.S(
		newDeleteCertErrorTestCase("invalid cert", pool, nil, false,
			core.ErrInvalid),
		newDeleteCertErrorTestCase("nil receiver", nil, valid, false,
			core.ErrNilReceiver),
		newDeleteCertErrorTestCase("cancelled", pool, valid, true,
			context.Canceled),
		newDeleteCertErrorTestCase("absent", pool, absent, false,
			core.ErrNotExists),
	)
}

func TestDeleteCertErrors(t *testing.T) {
	core.RunTestCases(t, deleteCertErrorTestCases(t))
}

// TestAddCertGuards confirms AddCert rejects a nil receiver, an invalid
// certificate and a duplicate, while accepting a fresh one.
func TestAddCertGuards(t *testing.T) {
	pool := certpool.New()
	leaf := mkTestLeaf(t, "add.example.com", nil).cert

	core.AssertFalse(t, (*certpool.CertPool)(nil).AddCert(leaf), "nil receiver")
	core.AssertFalse(t, pool.AddCert(nil), "invalid cert")
	core.AssertTrue(t, pool.AddCert(leaf), "fresh")
	core.AssertFalse(t, pool.AddCert(leaf), "duplicate")
}

// TestImportGuards confirms Import's short-circuit arms: nil receiver,
// cancelled context, nil source and self-import each add nothing.
func TestImportGuards(t *testing.T) {
	pool, _ := mkPopulatedPool(t)
	ctx := context.Background()

	n, err := (*certpool.CertPool)(nil).Import(ctx, pool)
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "nil receiver")
	core.AssertEqual(t, 0, n, "count")

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	n, err = pool.Import(cancelled, pool)
	core.AssertErrorIs(t, err, context.Canceled, "cancelled")
	core.AssertEqual(t, 0, n, "count")

	n, err = pool.Import(ctx, nil)
	core.AssertMustNoError(t, err, "nil source")
	core.AssertEqual(t, 0, n, "count")

	n, err = pool.Import(ctx, pool)
	core.AssertMustNoError(t, err, "self import")
	core.AssertEqual(t, 0, n, "count")
}

// TestImportCopiesFreshCerts confirms Import adds the source's certificates the
// destination lacks and skips the ones it already holds.
func TestImportCopiesFreshCerts(t *testing.T) {
	src, certs := mkPopulatedPool(t)
	dst := certpool.New()
	ctx := context.Background()

	n, err := dst.Import(ctx, src)
	core.AssertMustNoError(t, err, "import")
	core.AssertEqual(t, len(certs), n, "all fresh")

	n, err = dst.Import(ctx, src)
	core.AssertMustNoError(t, err, "re-import")
	core.AssertEqual(t, 0, n, "none fresh")
}

// TestImportPEMGuards confirms ImportPEM's short-circuit arms: nil receiver,
// cancelled context and empty input each add nothing.
func TestImportPEMGuards(t *testing.T) {
	pool := certpool.New()
	ctx := context.Background()
	leaf := mkTestLeaf(t, "pem.example.com", nil)

	n, err := (*certpool.CertPool)(nil).ImportPEM(ctx, certPEM(t, leaf.cert))
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "nil receiver")
	core.AssertEqual(t, 0, n, "count")

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	n, err = pool.ImportPEM(cancelled, certPEM(t, leaf.cert))
	core.AssertErrorIs(t, err, context.Canceled, "cancelled")
	core.AssertEqual(t, 0, n, "count")

	n, err = pool.ImportPEM(ctx, nil)
	core.AssertMustNoError(t, err, "empty")
	core.AssertEqual(t, 0, n, "count")
}

// TestImportPEMParsesAndIgnores confirms ImportPEM adds certificate blocks and
// silently skips a non-certificate block (a private key).
func TestImportPEMParsesAndIgnores(t *testing.T) {
	one := mkTestLeaf(t, "one.pem.example.com", nil)
	two := mkTestLeaf(t, "two.pem.example.com", nil)
	blob := append(certPEM(t, one.cert, two.cert), keyPEM(t, one)...)

	pool := certpool.New()
	n, err := pool.ImportPEM(context.Background(), blob)
	core.AssertMustNoError(t, err, "import pem")
	core.AssertEqual(t, 2, n, "two certs, key ignored")
}

// TestImportPEMSurfacesParseError confirms a malformed certificate block
// aborts the import with the decoder's error.
func TestImportPEMSurfacesParseError(t *testing.T) {
	blob := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("not a certificate"),
	})

	pool := certpool.New()
	n, err := pool.ImportPEM(context.Background(), blob)
	core.AssertError(t, err, "parse error")
	core.AssertEqual(t, 0, n, "count")
}
