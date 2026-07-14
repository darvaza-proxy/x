package certpool

import (
	"context"
	"crypto/x509"
	"testing"

	"darvaza.org/core"
)

// TestCertPoolEntryEqual confirms certPoolEntry.Equal: identity and equal DER
// match, a nil operand or receiver does not, and differing DER does not.
func TestCertPoolEntryEqual(t *testing.T) {
	a := &certPoolEntry{cert: &x509.Certificate{Raw: []byte{0x01}}}
	sameDER := &certPoolEntry{cert: &x509.Certificate{Raw: []byte{0x01}}}
	otherDER := &certPoolEntry{cert: &x509.Certificate{Raw: []byte{0x02}}}

	core.AssertTrue(t, a.Equal(a), "identity")
	core.AssertTrue(t, a.Equal(sameDER), "equal DER")
	core.AssertFalse(t, a.Equal(otherDER), "different DER")
	core.AssertFalse(t, a.Equal(nil), "nil operand")
	core.AssertFalse(t, (*certPoolEntry)(nil).Equal(a), "nil receiver")
}

// TestCertPoolEntryIsCA confirms certPoolEntry.IsCA: only a valid entry over a
// CA certificate is a CA; a leaf and an invalid entry are not.
func TestCertPoolEntryIsCA(t *testing.T) {
	ca := &certPoolEntry{cert: &x509.Certificate{Raw: []byte{0x01}, IsCA: true}}
	leaf := &certPoolEntry{cert: &x509.Certificate{Raw: []byte{0x01}}}

	core.AssertTrue(t, ca.IsCA(), "CA entry")
	core.AssertFalse(t, leaf.IsCA(), "leaf entry")
	core.AssertFalse(t, (&certPoolEntry{}).IsCA(), "invalid entry")
}

// TestBySubjectEmptyBucketRetained pins both arms of GetBySubjectHash's
// `!ok || l.Len() == 0` guard. Deleting the sole holder of a subject empties
// the bucket's list but keeps its map key, so the l.Len()==0 arm is reachable
// and distinct from the !ok arm a never-indexed hash produces.
func TestBySubjectEmptyBucketRetained(t *testing.T) {
	cert := fullCert()
	pool := New()
	core.AssertMustTrue(t, pool.AddCert(cert), "add")

	h, ok := HashSubject(cert)
	core.AssertMustTrue(t, ok, "hash subject")

	l, ok := pool.bySubject[h]
	core.AssertMustTrue(t, ok, "bucket present")
	core.AssertMustEqual(t, 1, l.Len(), "one entry")

	core.AssertMustNoError(t, pool.DeleteCert(context.Background(), cert),
		"delete")

	l, ok = pool.bySubject[h]
	core.AssertTrue(t, ok, "bucket key retained")
	core.AssertEqual(t, 0, l.Len(), "bucket emptied")
	core.AssertNil(t, pool.GetBySubjectHash(h), "empty bucket misses")

	// a hash never indexed exercises the !ok arm.
	core.AssertNil(t, pool.GetBySubjectHash(Hash{}), "absent key misses")
}
