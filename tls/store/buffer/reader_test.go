package buffer_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/store/buffer"
	"darvaza.org/x/tls/x509utils"
)

// feedCerts pushes blocks through the certificates-only callback.
func feedCerts(t *testing.T, buf *buffer.Buffer, file string,
	blocks ...*pem.Block) {
	t.Helper()
	cb := buf.NewAddCertsCallback()
	core.AssertMustNotNil(t, cb, "callback")
	for _, b := range blocks {
		cb(nil, file, b)
	}
}

// feedKeys pushes blocks through the private-keys-only callback.
func feedKeys(t *testing.T, buf *buffer.Buffer, file string,
	blocks ...*pem.Block) {
	t.Helper()
	cb := buf.NewAddPrivateKeysCallback()
	core.AssertMustNotNil(t, cb, "callback")
	for _, b := range blocks {
		cb(nil, file, b)
	}
}

// TestNewAddCertsCallback confirms the certificates-only callback records the
// certificate and nothing else.
func TestNewAddCertsCallback(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "certs.example.org")

	buf := buffer.New(context.Background(), nil)
	feedCerts(t, buf, "cert.pem", certBlock(leaf))

	core.AssertEqual(t, 1, buf.Certs().Len(), "certs")
	core.AssertEqual(t, 0, buf.Keys().Len(), "keys")
}

// TestNewAddPrivateKeysCallback confirms the keys-only callback records the key
// and no certificate.
func TestNewAddPrivateKeysCallback(t *testing.T) {
	leaf := mkSelfSignedLeaf(t, "keys.example.org")

	buf := buffer.New(context.Background(), nil)
	feedKeys(t, buf, "key.pem", keyBlock(t, leaf))

	core.AssertEqual(t, 1, buf.Keys().Len(), "keys")
	core.AssertEqual(t, 0, buf.Certs().Len(), "certs")
}

// TestCallbacksNilReceiver confirms the typed callbacks are nil-receiver safe.
func TestCallbacksNilReceiver(t *testing.T) {
	var buf *buffer.Buffer
	core.AssertNil(t, buf.NewAddCertsCallback(), "certs cb")
	core.AssertNil(t, buf.NewAddPrivateKeysCallback(), "keys cb")
}

// TestPushErr confirms a malformed block records an error against its source,
// surfaced through ForEach.
func TestPushErr(t *testing.T) {
	buf := buffer.New(context.Background(), nil)
	bad := &pem.Block{Type: "CERTIFICATE", Bytes: []byte("not a certificate")}
	feedCerts(t, buf, "bad.pem", bad)

	core.AssertEqual(t, 0, buf.Certs().Len(), "no certs")

	var errs int
	buf.ForEach(func(_ fs.FS, name string, _ []x509utils.PrivateKey,
		_ []*x509.Certificate, e []error) bool {
		core.AssertEqual(t, "bad.pem", name, "source")
		errs += len(e)
		return true
	})
	core.AssertEqual(t, 1, errs, "recorded error")
}
