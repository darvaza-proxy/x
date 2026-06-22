package buffer_test

import (
	"context"
	"crypto/x509"
	"io/fs"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils"
)

// mkSource builds a Source from a leaf and its key in one place.
func mkSource(name string, certs []*x509.Certificate,
	keys []x509utils.PrivateKey) *buffer.Source {
	return &buffer.Source{
		SourceName: buffer.NewSourceName(nil, name),
		Certs:      certs,
		Keys:       keys,
	}
}

// TestSourceAddCertPairSameSourceKey is the F34 regression: with a nil KeySet
// the same-source loop in findKeyForCert is the ONLY way to pair the leaf with
// its key. The bug (key.Public method value vs key.Public()) made the loop
// never match, so the pair went out with a nil key.
func TestSourceAddCertPairSameSourceKey(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "same-source.example.org")
	src := mkSource("pair.pem",
		core.S(leaf.Leaf),
		core.S(keyOf(t, leaf)))

	store := newFakeStore()
	// nil KeySet on purpose: no buffer-wide fallback is available.
	n, err := src.AddCertPair(context.Background(), store, nil)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertMustEqual(t, 1, len(store.pairs), "pairs")
	core.AssertMustNotNil(t, store.pairs[0].key, "paired key (F34)")
	core.AssertTrue(t,
		x509utils.ValidCertKeyPair(leaf.Leaf, store.pairs[0].key),
		"paired key matches leaf")
}

// TestSourceAddCertPairBufferWideKey confirms the fallback: an empty source
// KeySet but the key present in the buffer-wide KeySet still pairs.
func TestSourceAddCertPairBufferWideKey(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "wide.example.org")
	keys := buffer.MustKeySet(keyOf(t, leaf))

	// the source carries the leaf but NOT its key.
	src := mkSource("cert.pem", core.S(leaf.Leaf), nil)

	store := newFakeStore()
	n, err := src.AddCertPair(context.Background(), store, keys)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertMustEqual(t, 1, len(store.pairs), "pairs")
	core.AssertTrue(t,
		x509utils.ValidCertKeyPair(leaf.Leaf, store.pairs[0].key),
		"paired key matches leaf")
}

// sourceGuardOp names a Source flush method for the shared argument-guard
// checks in TestSourceGuards.
type sourceGuardOp struct {
	fn   func(*buffer.Source, context.Context, tls.StoreX509Writer) (int, error)
	name string
}

// sourceOps lists the Source flush methods that share the argument guards.
// AddCertPair takes an extra KeySet, so it is wrapped to the common signature
// with a nil set.
var sourceOps = []sourceGuardOp{
	{name: "AddCACerts", fn: (*buffer.Source).AddCACerts},
	{name: "AddCert", fn: (*buffer.Source).AddCert},
	{name: "AddPrivateKeys", fn: (*buffer.Source).AddPrivateKeys},
	{name: "AddCertPair", fn: func(src *buffer.Source, ctx context.Context,
		out tls.StoreX509Writer) (int, error) {
		return src.AddCertPair(ctx, out, nil)
	}},
}

// runSourceGuardCancelled confirms a Source flush method aborts on a cancelled
// context.
func runSourceGuardCancelled(t *testing.T, op sourceGuardOp) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := mkSource("x.pem", nil, nil)
	_, err := op.fn(src, ctx, newFakeStore())
	core.AssertErrorIs(t, err, context.Canceled, "err")
}

// TestSourceGuards confirms the argument guards on every Source flush method.
func TestSourceGuards(t *testing.T) {
	for _, op := range sourceOps {
		t.Run(op.name+"/nil receiver", func(t *testing.T) {
			_, err := op.fn(nil, context.Background(), newFakeStore())
			core.AssertErrorIs(t, err, core.ErrNilReceiver, "err")
		})
		t.Run(op.name+"/nil store", func(t *testing.T) {
			src := mkSource("x.pem", nil, nil)
			_, err := op.fn(src, context.Background(), nil)
			core.AssertErrorIs(t, err, tls.ErrNoStore, "err")
		})
		t.Run(op.name+"/cancelled", func(t *testing.T) {
			runSourceGuardCancelled(t, op)
		})
	}
}

// TestSourceClone confirms Clone produces an independent copy.
func TestSourceClone(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "clone.example.org")
	src := mkSource("s.pem",
		core.S(leaf.Leaf),
		core.S(keyOf(t, leaf)))

	clone := src.Clone()
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertEqual(t, 1, len(clone.Certs), "clone certs")
	core.AssertEqual(t, 1, len(clone.Keys), "clone keys")

	// mutating the clone's slices must not touch the source.
	clone.Certs = append(clone.Certs, leaf.Leaf)
	core.AssertEqual(t, 1, len(src.Certs), "source unchanged")
}

// TestSourceCloneNil confirms Clone is nil-receiver safe.
func TestSourceCloneNil(t *testing.T) {
	var src *buffer.Source
	core.AssertNil(t, src.Clone(), "clone")
}

func TestSourceNameIsFile(t *testing.T) {
	core.AssertTrue(t, buffer.NewSourceName(nil, "f.pem").IsFile(), "named")
	core.AssertFalse(t, buffer.NewSourceName(nil, "").IsFile(), "anonymous")
}

// TestSourceNameNewErrorPathError confirms a named source yields an
// [fs.PathError] carrying the file name and op.
func TestSourceNameNewErrorPathError(t *testing.T) {
	sn := buffer.NewSourceName(nil, "bad.pem")
	err := sn.NewError(core.ErrInvalid, "AddCert", "boom")

	pe := core.AssertMustTypeIs[*fs.PathError](t, err, "PathError")
	core.AssertEqual(t, "bad.pem", pe.Path, "path")
	core.AssertEqual(t, "AddCert", pe.Op, "op")
	core.AssertErrorIs(t, pe, core.ErrInvalid, "wrapped")
}

// TestSourceNameAppendErrorFlattens confirms a compound error is appended
// element by element rather than nested.
func TestSourceNameAppendErrorFlattens(t *testing.T) {
	inner := new(core.CompoundError)
	_ = inner.AppendError(core.ErrInvalid)
	_ = inner.AppendError(core.ErrExists)

	sn := buffer.NewSourceName(nil, "")
	out := new(core.CompoundError)
	sn.AppendError(out, inner, "op", "")

	core.AssertEqual(t, 2, len(out.Errs), "flattened count")
}
