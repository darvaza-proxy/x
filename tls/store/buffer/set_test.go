package buffer_test

// cspell:words csclone cscopy csinit

import (
	"crypto/x509"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils"
)

// TestNewKeySet confirms construction and deduplication by public key.
func TestNewKeySet(t *testing.T) {
	a := keyOf(t, mkSelfSignedLeaf(t, "a.example.org"))
	b := keyOf(t, mkSelfSignedLeaf(t, "b.example.org"))

	ks, err := buffer.NewKeySet(a, b)
	core.AssertMustNoError(t, err, "new")
	core.AssertEqual(t, 2, ks.Len(), "len")

	dup, err := buffer.NewKeySet(a, a)
	core.AssertMustNoError(t, err, "new dup")
	core.AssertEqual(t, 1, dup.Len(), "unique len")
}

// TestNewKeySetInvalid confirms a nil key is rejected and MustKeySet panics.
func TestNewKeySetInvalid(t *testing.T) {
	_, err := buffer.NewKeySet(nil)
	core.AssertErrorIs(t, err, core.ErrInvalid, "nil key")

	core.AssertPanic(t, func() { buffer.MustKeySet(nil) }, nil, "must panics")
}

// TestKeySetPublic confirms the public-key type-cast helper.
func TestKeySetPublic(t *testing.T) {
	ks := buffer.MustKeySet()
	key := keyOf(t, mkSelfSignedLeaf(t, "pub.example.org"))

	core.AssertNotNil(t, ks.Public(key), "public")
	core.AssertNil(t, ks.Public(nil), "nil key")
}

// TestKeySetGetFromCertificate confirms lookup by a certificate's public key,
// covering the found, not-present and invalid-certificate paths plus the nil
// receiver.
func TestKeySetGetFromCertificate(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "get.example.org")
	other := mkSelfSignedLeaf(t, "other.example.org")
	ks := buffer.MustKeySet(keyOf(t, leaf))

	got, err := ks.GetFromCertificate(leaf.Leaf)
	core.AssertMustNoError(t, err, "found")
	core.AssertTrue(t, x509utils.ValidCertKeyPair(leaf.Leaf, got), "match")

	_, err = ks.GetFromCertificate(other.Leaf)
	core.AssertErrorIs(t, err, core.ErrNotExists, "absent")

	_, err = ks.GetFromCertificate(new(x509.Certificate))
	core.AssertErrorIs(t, err, core.ErrInvalid, "no public key")

	var nilKS *buffer.KeySet
	_, err = nilKS.GetFromCertificate(leaf.Leaf)
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "nil receiver")
}

// TestKeySetClone confirms Clone is independent and nil-receiver safe.
func TestKeySetClone(t *testing.T) {
	a := keyOf(t, mkSelfSignedLeaf(t, "clone-a.example.org"))
	b := keyOf(t, mkSelfSignedLeaf(t, "clone-b.example.org"))
	ks := buffer.MustKeySet(a)

	clone := ks.Clone()
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertEqual(t, 1, clone.Len(), "clone len")

	_, err := clone.Push(b)
	core.AssertMustNoError(t, err, "push")
	core.AssertEqual(t, 2, clone.Len(), "clone grew")
	core.AssertEqual(t, 1, ks.Len(), "source unchanged")

	var nilKS *buffer.KeySet
	core.AssertNil(t, nilKS.Clone(), "nil clone")
}

// TestKeySetCopyCond confirms Copy honours the filter condition.
func TestKeySetCopyCond(t *testing.T) {
	a := keyOf(t, mkSelfSignedLeaf(t, "copy-a.example.org"))
	b := keyOf(t, mkSelfSignedLeaf(t, "copy-b.example.org"))
	ks := buffer.MustKeySet(a, b)

	dst := ks.Copy(nil, func(k x509utils.PrivateKey) bool {
		return x509utils.PublicKeyEqual(k.Public(), a.Public())
	})
	core.AssertMustNotNil(t, dst, "dst")
	core.AssertEqual(t, 1, dst.Len(), "filtered copy")
}

// TestInitKeySet confirms in-place initialisation and its guards.
func TestInitKeySet(t *testing.T) {
	a := keyOf(t, mkSelfSignedLeaf(t, "init.example.org"))

	var ks buffer.KeySet
	core.AssertMustNoError(t, buffer.InitKeySet(&ks, a), "init")
	core.AssertEqual(t, 1, ks.Len(), "init len")

	core.AssertErrorIs(t, buffer.InitKeySet(nil, a), core.ErrInvalid, "nil out")

	var must buffer.KeySet
	buffer.MustInitKeySet(&must, a)
	core.AssertEqual(t, 1, must.Len(), "must init len")

	core.AssertPanic(t, func() { buffer.MustInitKeySet(nil, a) }, nil,
		"must init panics")
}

// TestNewCertSet confirms construction and deduplication by leaf.
func TestNewCertSet(t *testing.T) {
	a := mkSelfSignedLeaf(t, "cs-a.example.org")
	b := mkSelfSignedLeaf(t, "cs-b.example.org")

	cs, err := buffer.NewCertSet(a, b)
	core.AssertMustNoError(t, err, "new")
	core.AssertEqual(t, 2, cs.Len(), "len")

	dup, err := buffer.NewCertSet(a, a)
	core.AssertMustNoError(t, err, "new dup")
	core.AssertEqual(t, 1, dup.Len(), "unique len")
}

// TestNewCertSetInvalid confirms a nil certificate is rejected and MustCertSet
// panics.
func TestNewCertSetInvalid(t *testing.T) {
	_, err := buffer.NewCertSet(nil)
	core.AssertErrorIs(t, err, core.ErrInvalid, "nil cert")

	core.AssertPanic(t, func() { buffer.MustCertSet(nil) }, nil, "must panics")
}

// TestCertSetClone confirms Clone is independent and nil-receiver safe.
func TestCertSetClone(t *testing.T) {
	a := mkSelfSignedLeaf(t, "csclone-a.example.org")
	b := mkSelfSignedLeaf(t, "csclone-b.example.org")
	cs := buffer.MustCertSet(a)

	clone := cs.Clone()
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertEqual(t, 1, clone.Len(), "clone len")

	_, err := clone.Push(b)
	core.AssertMustNoError(t, err, "push")
	core.AssertEqual(t, 2, clone.Len(), "clone grew")
	core.AssertEqual(t, 1, cs.Len(), "source unchanged")

	var nilCS *buffer.CertSet
	core.AssertNil(t, nilCS.Clone(), "nil clone")
}

// TestCertSetCopyCond confirms Copy honours the filter condition.
func TestCertSetCopyCond(t *testing.T) {
	a := mkSelfSignedLeaf(t, "cscopy-a.example.org")
	b := mkSelfSignedLeaf(t, "cscopy-b.example.org")
	cs := buffer.MustCertSet(a, b)

	dst := cs.Copy(nil, func(c *tls.Certificate) bool {
		return c.Leaf.Equal(a.Leaf)
	})
	core.AssertMustNotNil(t, dst, "dst")
	core.AssertEqual(t, 1, dst.Len(), "filtered copy")
}

// TestInitCertSet confirms in-place initialisation and its guards.
func TestInitCertSet(t *testing.T) {
	a := mkSelfSignedLeaf(t, "csinit.example.org")

	var cs buffer.CertSet
	core.AssertMustNoError(t, buffer.InitCertSet(&cs, a), "init")
	core.AssertEqual(t, 1, cs.Len(), "init len")

	core.AssertErrorIs(t, buffer.InitCertSet(nil, a), core.ErrInvalid, "nil out")

	var must buffer.CertSet
	buffer.MustInitCertSet(&must, a)
	core.AssertEqual(t, 1, must.Len(), "must init len")

	core.AssertPanic(t, func() { buffer.MustInitCertSet(nil, a) }, nil,
		"must init panics")
}
