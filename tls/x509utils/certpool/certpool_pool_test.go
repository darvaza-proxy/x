package certpool_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// TestIsZero confirms IsZero: a nil pool is never zero, a fresh pool is, and a
// populated one is not.
func TestIsZero(t *testing.T) {
	core.AssertFalse(t, (*certpool.CertPool)(nil).IsZero(), "nil pool")
	core.AssertTrue(t, certpool.New().IsZero(), "empty pool")

	pool, _ := mkPopulatedPool(t)
	core.AssertFalse(t, pool.IsZero(), "populated pool")
}

// TestIsCA confirms IsCA: neither a nil pool nor an empty one is a CA pool, a
// pool of only CAs is, and a pool holding any leaf is not.
func TestIsCA(t *testing.T) {
	core.AssertFalse(t, (*certpool.CertPool)(nil).IsCA(), "nil pool")
	core.AssertFalse(t, certpool.New().IsCA(), "empty pool")

	root := mkTestCA(t, "CA Root", nil)
	sub := mkTestCA(t, "CA Sub", &root)
	caPool := certpool.New()
	core.AssertMustTrue(t, caPool.AddCert(root.cert), "add root")
	core.AssertMustTrue(t, caPool.AddCert(sub.cert), "add sub")
	core.AssertTrue(t, caPool.IsCA(), "all CAs")

	leaf := mkTestLeaf(t, "leaf.example.com", &root)
	core.AssertMustTrue(t, caPool.AddCert(leaf.cert), "add leaf")
	core.AssertFalse(t, caPool.IsCA(), "one leaf breaks it")
}

// TestReset confirms Reset empties a populated pool and reports ErrNilReceiver
// for a nil one.
func TestReset(t *testing.T) {
	core.AssertErrorIs(t, (*certpool.CertPool)(nil).Reset(),
		core.ErrNilReceiver, "nil receiver")

	pool, _ := mkPopulatedPool(t)
	core.AssertMustNoError(t, pool.Reset(), "reset")
	core.AssertEqual(t, 0, pool.Count(), "empty after reset")
	core.AssertTrue(t, pool.IsZero(), "zero after reset")
}
