//go:build windows

package certpool

import (
	"crypto/x509"
	"errors"
	"unsafe"

	"darvaza.org/core"
	"golang.org/x/sys/windows"
)

// errCryptNotFound is CRYPT_E_NOT_FOUND, the sentinel
// [windows.CertEnumCertificatesInStore] reports once a store holds no
// further certificate contexts. It marks a clean end of enumeration,
// not a failure.
const errCryptNotFound = windows.Errno(0x80092004)

// systemStoreNames are the Windows certificate stores consulted by
// [NewSystemCertPool]. Only ROOT is read: its members are the anchors
// Windows itself trusts. The CA store is deliberately excluded — it holds
// chain-building intermediates that confer no trust on Windows, and the
// current-user CA store is writeable without elevation or prompt, so
// anchoring its contents would trust more than Windows does.
var systemStoreNames = []string{"ROOT"}

// NewSystemCertPool returns a [CertPool] populated from the Windows system
// certificate stores, together with an aggregation of any errors met while
// reading them.
func NewSystemCertPool() (*CertPool, error) {
	var errs core.CompoundError
	pool := New()

	for _, name := range systemStoreNames {
		if err := loadSystemStore(pool, &errs, name); err != nil {
			_ = errs.AppendError(err)
		}
	}

	err := errs.AsError()
	switch {
	case pool.Count() > 0:
		// success, possibly alongside per-store errors
		return pool, err
	case err == nil:
		// no certs and no errors... don't bother again.
		return nil, ErrNoCertificatesFound
	default:
		// no cert, but we got errors to report.
		return nil, err
	}
}

// loadSystemStore opens a single named Windows system store and adds every
// certificate it can parse to the pool, collecting parse errors into errs.
func loadSystemStore(pool *CertPool, errs *core.CompoundError, name string) error {
	storeName, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return core.Wrapf(err, "store %q", name)
	}

	store, err := windows.CertOpenSystemStore(0, storeName)
	if err != nil {
		return core.Wrapf(err, "CertOpenSystemStore(%q)", name)
	}
	defer func() { _ = windows.CertCloseStore(store, 0) }()

	return enumStoreCerts(pool, errs, store)
}

// enumStoreCerts walks the certificate contexts of an open store, copying
// each DER blob out of store-owned memory and adding the parsed result to
// the pool. [windows.CertEnumCertificatesInStore] frees the context passed
// as prev on every call — even on error — and the terminating call frees the
// last one, so the enumeration cleans up after itself and no context is
// freed by hand.
func enumStoreCerts(pool *CertPool, errs *core.CompoundError, store windows.Handle) error {
	var prev *windows.CertContext
	for {
		ctx, err := windows.CertEnumCertificatesInStore(store, prev)
		switch {
		case ctx != nil:
			addCertContext(pool, errs, ctx)
			prev = ctx
		case errors.Is(err, errCryptNotFound), err == nil:
			// CRYPT_E_NOT_FOUND marks the clean end of the store. The
			// nil-error arm is defensive: the x/sys wrapper only reports
			// an error once the context is nil, so err is never nil here
			// today, but a nil context is the end of the store either way.
			return nil
		default:
			return core.Wrap(err, "CertEnumCertificatesInStore")
		}
	}
}

// addCertContext parses the DER carried by a certificate context and adds
// it to the pool. A context without a DER blob to reference is skipped
// silently; a blob Go's parser rejects is recorded in errs against its hash,
// so one bad entry is reported by name without sinking the rest of the load.
func addCertContext(pool *CertPool, errs *core.CompoundError, ctx *windows.CertContext) {
	if ctx.EncodedCert == nil || ctx.Length == 0 {
		return
	}

	// Copy before the next enumeration call frees this context.
	der := make([]byte, ctx.Length)
	copy(der, unsafe.Slice(ctx.EncodedCert, ctx.Length))

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		hash := Sum(der)
		_ = errs.AppendError(core.Wrapf(err, "certificate %x", hash[:]))
		return
	}

	pool.AddCert(cert)
}
