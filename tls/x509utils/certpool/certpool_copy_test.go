package certpool_test

import (
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

// TestPoolCopyNilReceiver confirms Copy on a nil pool creates a destination
// when none is given and returns the supplied one otherwise.
func TestPoolCopyNilReceiver(t *testing.T) {
	fresh := (*certpool.CertPool)(nil).Copy(nil, nil)
	core.AssertMustNotNil(t, fresh, "created destination")
	core.AssertEqual(t, 0, fresh.Count(), "empty")

	dst := certpool.New()
	got := (*certpool.CertPool)(nil).Copy(dst, nil)
	core.AssertSame(t, dst, got, "returns supplied destination")
}

// TestPoolCopyToSelf confirms copying a pool onto itself is a no-op returning
// the same pool.
func TestPoolCopyToSelf(t *testing.T) {
	pool, _ := mkPopulatedPool(t)
	core.AssertSame(t, pool, pool.Copy(pool, nil), "self copy")
}

// TestPoolCopyClonesWhenNoDestination confirms Copy with no destination clones
// every certificate into a fresh, independent pool.
func TestPoolCopyClonesWhenNoDestination(t *testing.T) {
	pool, certs := mkPopulatedPool(t)

	clone := pool.Copy(nil, nil)
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertEqual(t, len(certs), clone.Count(), "all copied")

	extra := mkTestLeaf(t, "extra.example.com", nil).cert
	core.AssertMustTrue(t, clone.AddCert(extra), "add to clone")
	core.AssertEqual(t, len(certs), pool.Count(), "source unchanged")
}

// TestPoolCopyIntoDestinationSkipsKnown confirms Copy into a destination adds
// only the certificates it lacks: a second copy adds nothing.
func TestPoolCopyIntoDestinationSkipsKnown(t *testing.T) {
	pool, certs := mkPopulatedPool(t)
	dst := certpool.New()

	pool.Copy(dst, nil)
	core.AssertMustEqual(t, len(certs), dst.Count(), "all copied")

	pool.Copy(dst, nil)
	core.AssertEqual(t, len(certs), dst.Count(), "no duplicates")
}

// TestPoolCopyCondition confirms the condition filters which certificates are
// copied: only CAs land in the destination.
func TestPoolCopyCondition(t *testing.T) {
	root := mkTestCA(t, "Copy Root", nil)
	leaf := mkTestLeaf(t, "copy.example.com", &root)
	pool := certpool.New()
	core.AssertMustTrue(t, pool.AddCert(root.cert), "add root")
	core.AssertMustTrue(t, pool.AddCert(leaf.cert), "add leaf")

	caOnly := pool.Copy(nil, func(cert *x509.Certificate) bool {
		return cert.IsCA
	})
	core.AssertEqual(t, 1, caOnly.Count(), "only the CA copied")
	core.AssertTrue(t, caOnly.IsCA(), "destination holds only CAs")
}

// TestPoolClone confirms Clone duplicates a pool independently and returns nil
// for a nil receiver.
func TestPoolClone(t *testing.T) {
	core.AssertNil(t, (*certpool.CertPool)(nil).Clone(), "nil receiver")

	pool, certs := mkPopulatedPool(t)
	clone := core.AssertMustTypeIs[*certpool.CertPool](t, pool.Clone(), "clone")
	core.AssertEqual(t, len(certs), clone.Count(), "all cloned")

	extra := mkTestLeaf(t, "clone-extra.example.com", nil).cert
	core.AssertMustTrue(t, clone.AddCert(extra), "add to clone")
	core.AssertEqual(t, len(certs), pool.Count(), "source unchanged")
}
