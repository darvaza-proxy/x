package certpool_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

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
