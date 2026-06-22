package buffer_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls"
	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils"
)

// feed pushes the given PEM blocks into buf as a single named source via the
// general add callback (certs and keys both routed by block type).
func feed(t *testing.T, buf *buffer.Buffer, file string, blocks ...*pem.Block) {
	t.Helper()
	cb := buf.NewAddCallback()
	core.AssertMustNotNil(t, cb, "callback")
	for _, b := range blocks {
		cb(nil, file, b)
	}
}

// TestNew confirms a fresh buffer is usable and empty.
func TestNew(t *testing.T) {
	buf := buffer.New(context.Background(), nil)
	core.AssertMustNotNil(t, buf, "buffer")

	core.AssertMustNotNil(t, buf.Certs(), "certs")
	core.AssertMustNotNil(t, buf.Keys(), "keys")

	pairs, err := buf.Pairs()
	core.AssertMustNoError(t, err, "pairs")
	core.AssertEqual(t, 0, len(pairs), "pair count")
}

// TestNilReceiver confirms every accessor is nil-receiver safe.
func TestNilReceiver(t *testing.T) {
	var buf *buffer.Buffer

	core.AssertNil(t, buf.Certs(), "certs")
	core.AssertNil(t, buf.Keys(), "keys")
	core.AssertNil(t, buf.Clone(), "clone")
	core.AssertNil(t, buf.NewAddCallback(), "callback")

	_, err := buf.Pairs()
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "pairs err")

	// ForEach on a nil receiver is a silent no-op.
	buf.ForEach(func(fs.FS, string, []x509utils.PrivateKey,
		[]*x509.Certificate, []error) bool {
		t.Fatal("nil receiver must not iterate")
		return false
	})
}

// bufferGuardOp names a Buffer flush method for the shared argument-guard
// checks in TestAddGuards.
type bufferGuardOp struct {
	fn   func(*buffer.Buffer, context.Context, tls.StoreX509Writer) (int, error)
	name string
}

// addOps lists the four flush methods as method expressions so the guards are
// exercised uniformly across all of them.
var addOps = []bufferGuardOp{
	{name: "AddCACerts", fn: (*buffer.Buffer).AddCACerts},
	{name: "AddPrivateKey", fn: (*buffer.Buffer).AddPrivateKey},
	{name: "AddCert", fn: (*buffer.Buffer).AddCert},
	{name: "AddCertPair", fn: (*buffer.Buffer).AddCertPair},
}

// runAddGuardCancelled confirms a flush method aborts on a cancelled context.
func runAddGuardCancelled(t *testing.T, op bufferGuardOp) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	buf := buffer.New(context.Background(), nil)
	_, err := op.fn(buf, ctx, newFakeStore())
	core.AssertErrorIs(t, err, context.Canceled, "err")
}

// TestAddGuards confirms the argument guards on every flush method.
func TestAddGuards(t *testing.T) {
	for _, op := range addOps {
		t.Run(op.name+"/nil receiver", func(t *testing.T) {
			_, err := op.fn(nil, context.Background(), newFakeStore())
			core.AssertErrorIs(t, err, core.ErrNilReceiver, "err")
		})
		t.Run(op.name+"/nil store", func(t *testing.T) {
			buf := buffer.New(context.Background(), nil)
			_, err := op.fn(buf, context.Background(), nil)
			core.AssertErrorIs(t, err, tls.ErrNoStore, "err")
		})
		t.Run(op.name+"/cancelled", func(t *testing.T) {
			runAddGuardCancelled(t, op)
		})
	}
}

// TestPopulateAndRead feeds a leaf and its key as one source and confirms the
// buffer exposes both and pairs them.
func TestPopulateAndRead(t *testing.T) {
	ca := mkCA(t, "Read CA")
	leaf := mkLeaf(t, ca, "read.example.com", "read.example.com")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "leaf.pem", certBlock(leaf), keyBlock(t, leaf))

	core.AssertEqual(t, 1, buf.Certs().Len(), "cert count")
	core.AssertEqual(t, 1, buf.Keys().Len(), "key count")

	pairs, err := buf.Pairs()
	core.AssertMustNoError(t, err, "pairs")
	core.AssertMustEqual(t, 1, len(pairs), "pair count")
	core.AssertEqual(t, 1, len(pairs[0].Certs), "paired certs")
	core.AssertTrue(t, x509utils.ValidCertKeyPair(leaf.Leaf, pairs[0].Key),
		"pair matches")

	var sources int
	buf.ForEach(func(_ fs.FS, name string, keys []x509utils.PrivateKey,
		certs []*x509.Certificate, _ []error) bool {
		sources++
		core.AssertEqual(t, "leaf.pem", name, "source name")
		core.AssertEqual(t, 1, len(keys), "source keys")
		core.AssertEqual(t, 1, len(certs), "source certs")
		return true
	})
	core.AssertEqual(t, 1, sources, "source count")
}

// TestAddCACerts confirms trusted roots flush to the store.
func TestAddCACerts(t *testing.T) {
	ca := mkCA(t, "Trust CA")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "ca.pem", certBlock(ca))

	store := newFakeStore()
	n, err := buf.AddCACerts(context.Background(), store)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertEqual(t, 1, len(store.caCerts), "stored roots")
}

// TestAddCert confirms plain certificates flush to the store.
func TestAddCert(t *testing.T) {
	ca := mkCA(t, "Cert CA")
	leaf := mkLeaf(t, ca, "cert.example.com", "cert.example.com")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "leaf.pem", certBlock(leaf))

	store := newFakeStore()
	n, err := buf.AddCert(context.Background(), store)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertEqual(t, 1, len(store.certs), "stored certs")
}

// TestAddPrivateKeyKeyOnly is the regression guard for F51: a key-only source
// (no certificate) must still flush its key. The old cond filtered on
// len(Certs)>0, so AddPrivateKey silently added nothing.
func TestAddPrivateKeyKeyOnly(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "key.example.org")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "key.pem", keyBlock(t, leaf))

	store := newFakeStore()
	n, err := buf.AddPrivateKey(context.Background(), store)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertMustEqual(t, 1, len(store.keys), "stored keys")
	core.AssertTrue(t, x509utils.ValidCertKeyPair(leaf.Leaf, store.keys[0]),
		"stored key matches")
}

// feedFn stages a leaf and its key into a buffer in a particular order.
type feedFn func(t *testing.T, buf *buffer.Buffer, leaf *tls.Certificate)

// addCertPairCase confirms AddCertPair pairs a leaf with its key regardless of
// the injection order staged by feed; the assertion path is shared.
type addCertPairCase struct {
	feed feedFn
	name string
}

var _ core.TestCase = addCertPairCase{}

func newAddCertPairCase(name string, feed feedFn) addCertPairCase {
	return addCertPairCase{name: name, feed: feed}
}

// Name returns the test case name.
func (tc addCertPairCase) Name() string { return tc.name }

// Test stages the buffer via feed and confirms the leaf is paired with its key.
func (tc addCertPairCase) Test(t *testing.T) {
	t.Helper()

	ca := mkCA(t, "Pair CA")
	leaf := mkLeaf(t, ca, "pair.example.com", "pair.example.com")
	buf := buffer.New(context.Background(), nil)
	tc.feed(t, buf, leaf)

	store := newFakeStore()
	n, err := buf.AddCertPair(context.Background(), store)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertMustEqual(t, 1, len(store.pairs), "stored pairs")
	core.AssertMustNotNil(t, store.pairs[0].key, "paired key")
	core.AssertTrue(t,
		x509utils.ValidCertKeyPair(leaf.Leaf, store.pairs[0].key),
		"pair key matches leaf")
}

// TestAddCertPairOrderIndependent is the core design guarantee: a leaf is
// paired with its key regardless of injection order or source — key before
// cert, cert before key, or each in its own source.
func TestAddCertPairOrderIndependent(t *testing.T) {
	core.RunTestCases(t, []addCertPairCase{
		newAddCertPairCase("key then cert, one source",
			func(t *testing.T, buf *buffer.Buffer, leaf *tls.Certificate) {
				feed(t, buf, "pair.pem", keyBlock(t, leaf), certBlock(leaf))
			}),
		newAddCertPairCase("cert then key, one source",
			func(t *testing.T, buf *buffer.Buffer, leaf *tls.Certificate) {
				feed(t, buf, "pair.pem", certBlock(leaf), keyBlock(t, leaf))
			}),
		newAddCertPairCase("separate sources",
			func(t *testing.T, buf *buffer.Buffer, leaf *tls.Certificate) {
				feed(t, buf, "cert.pem", certBlock(leaf))
				feed(t, buf, "key.pem", keyBlock(t, leaf))
			}),
	})
}

// TestAddCertPairIntermediates confirms the leaf is paired and the rest of the
// source travels as intermediates.
func TestAddCertPairIntermediates(t *testing.T) {
	root := mkCA(t, "Root CA")
	sub := mkSubCA(t, root, "Sub CA")
	leaf := mkLeaf(t, sub, "chain.example.com", "chain.example.com")

	buf := buffer.New(context.Background(), nil)
	// leaf first (it is the pair leaf), then the intermediate, plus the key.
	feed(t, buf, "bundle.pem", certBlock(leaf), certBlock(sub), keyBlock(t, leaf))

	store := newFakeStore()
	n, err := buf.AddCertPair(context.Background(), store)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, 1, n, "count")
	core.AssertMustEqual(t, 1, len(store.pairs), "stored pairs")
	core.AssertEqual(t, 1, len(store.pairs[0].inter), "intermediates")
}

// TestAddPrivateKeyIdempotent is the regression guard for F49: re-applying keys
// already present is a successful no-op (count 0, nil), not ErrEmpty.
func TestAddPrivateKeyIdempotent(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "key.example.org")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "key.pem", keyBlock(t, leaf))

	store := newFakeStore()
	n, err := buf.AddPrivateKey(context.Background(), store)
	core.AssertMustNoError(t, err, "first add")
	core.AssertEqual(t, 1, n, "first count")

	// second flush against the same store: all already present.
	n, err = buf.AddPrivateKey(context.Background(), store)
	core.AssertNoError(t, err, "re-apply err")
	core.AssertEqual(t, 0, n, "re-apply count")
}

// TestAddEmptyBuffer confirms a genuinely empty buffer still reports ErrEmpty —
// the distinction F49 preserves.
func TestAddEmptyBuffer(t *testing.T) {
	buf := buffer.New(context.Background(), nil)
	store := newFakeStore()

	_, err := buf.AddPrivateKey(context.Background(), store)
	core.AssertErrorIs(t, err, x509utils.ErrEmpty, "empty buffer")
}

// runFlushErrAddCACerts confirms a store error from AddCACerts surfaces.
func runFlushErrAddCACerts(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	buf := buffer.New(ctx, nil)
	feed(t, buf, "ca.pem", certBlock(mkCA(t, "Err CA")))
	store := newFakeStore()
	store.errAddCACerts = core.ErrInvalid
	n, err := buf.AddCACerts(ctx, store)
	core.AssertErrorIs(t, err, core.ErrInvalid, "err")
	core.AssertEqual(t, 0, n, "count")
}

// runFlushErrAddCert confirms a store error from AddCert surfaces.
func runFlushErrAddCert(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	leaf := mkLeaf(t, mkCA(t, "Err CA"), "err.example.com", "err.example.com")
	buf := buffer.New(ctx, nil)
	feed(t, buf, "leaf.pem", certBlock(leaf))
	store := newFakeStore()
	store.errAddCert = core.ErrInvalid
	n, err := buf.AddCert(ctx, store)
	core.AssertErrorIs(t, err, core.ErrInvalid, "err")
	core.AssertEqual(t, 0, n, "count")
}

// runFlushErrAddPrivateKey confirms a store error from AddPrivateKey surfaces.
func runFlushErrAddPrivateKey(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	leaf := mkSelfSignedLeaf(t, "err.example.org")
	buf := buffer.New(ctx, nil)
	feed(t, buf, "key.pem", keyBlock(t, leaf))
	store := newFakeStore()
	store.errAddPrivateKey = core.ErrInvalid
	n, err := buf.AddPrivateKey(ctx, store)
	core.AssertErrorIs(t, err, core.ErrInvalid, "err")
	core.AssertEqual(t, 0, n, "count")
}

// runFlushErrAddCertPair confirms a store error from AddCertPair surfaces.
func runFlushErrAddCertPair(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	leaf := mkLeaf(t, mkCA(t, "Err CA"), "err.example.com", "err.example.com")
	buf := buffer.New(ctx, nil)
	feed(t, buf, "pair.pem", certBlock(leaf), keyBlock(t, leaf))
	store := newFakeStore()
	store.errAddCertPair = core.ErrInvalid
	n, err := buf.AddCertPair(ctx, store)
	core.AssertErrorIs(t, err, core.ErrInvalid, "err")
	core.AssertEqual(t, 0, n, "count")
}

// TestFlushPropagatesError confirms a real add error (not a duplicate) from
// the store surfaces through the compound error rather than being swallowed,
// across every flush method.
func TestFlushPropagatesError(t *testing.T) {
	t.Run("AddCACerts", runFlushErrAddCACerts)
	t.Run("AddCert", runFlushErrAddCert)
	t.Run("AddPrivateKey", runFlushErrAddPrivateKey)
	t.Run("AddCertPair", runFlushErrAddCertPair)
}

// TestClone confirms a clone carries the content and is independent of the
// source.
func TestClone(t *testing.T) {
	ca := mkCA(t, "Clone CA")
	leaf := mkLeaf(t, ca, "clone.example.com", "clone.example.com")

	buf := buffer.New(context.Background(), nil)
	feed(t, buf, "leaf.pem", certBlock(leaf), keyBlock(t, leaf))

	clone := buf.Clone()
	core.AssertMustNotNil(t, clone, "clone")
	core.AssertEqual(t, 1, clone.Certs().Len(), "clone certs")
	core.AssertEqual(t, 1, clone.Keys().Len(), "clone keys")

	// adding to the clone must not affect the original.
	extra := mkLeaf(t, ca, "extra.example.com", "extra.example.com")
	feed(t, clone, "extra.pem", certBlock(extra))
	core.AssertEqual(t, 2, clone.Certs().Len(), "clone after add")
	core.AssertEqual(t, 1, buf.Certs().Len(), "source unchanged")
}
