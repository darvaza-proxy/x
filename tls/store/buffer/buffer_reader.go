package buffer

import (
	"crypto/x509"
	"encoding/pem"
	"io/fs"

	"darvaza.org/x/tls/x509utils"
)

// NewAddCallback returns a callback that adds all certificates and private keys
// to the [Buffer].
func (buf *Buffer) NewAddCallback() x509utils.DecodePEMBlockFunc {
	if buf == nil {
		return nil
	}
	return buf.onAdd
}

func (buf *Buffer) onAdd(fSys fs.FS, fileName string, block *pem.Block) bool {
	// cert?
	cert, err := x509utils.BlockToCertificate(block)
	switch {
	case cert != nil:
		// cert
		buf.pushCert(fSys, fileName, cert)
	case err == x509utils.ErrIgnored:
		// key?
		var key x509utils.PrivateKey

		key, err = x509utils.BlockToPrivateKey(block)
		if key != nil {
			// key
			buf.pushKey(fSys, fileName, key)
		}
	}

	if err != nil {
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() == nil
}

// NewAddCertsCallback returns a callback that adds all certificates to the [Buffer].
func (buf *Buffer) NewAddCertsCallback() x509utils.DecodePEMBlockFunc {
	if buf == nil {
		return nil
	}
	return buf.onAddCert
}

func (buf *Buffer) onAddCert(fSys fs.FS, fileName string, block *pem.Block) bool {
	cert, err := x509utils.BlockToCertificate(block)
	switch {
	case cert != nil:
		buf.pushCert(fSys, fileName, cert)
	case err != nil:
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() == nil
}

// NewAddPrivateKeysCallback returns a callback that adds private keys to the [Buffer].
func (buf *Buffer) NewAddPrivateKeysCallback() x509utils.DecodePEMBlockFunc {
	if buf == nil {
		return nil
	}
	return buf.onAddPrivateKeys
}

func (buf *Buffer) onAddPrivateKeys(fSys fs.FS, fileName string, block *pem.Block) bool {
	key, err := x509utils.BlockToPrivateKey(block)
	switch {
	case key != nil:
		buf.pushKey(fSys, fileName, key)
	case err != nil:
		buf.pushErr(fSys, fileName, err)
	}

	return buf.ctx.Err() == nil
}

func (buf *Buffer) pushErr(fSys fs.FS, fileName string, err error) {
	if err == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.unsafeInit()
	e := buf.unsafeGetSource2(fSys, fileName)
	e.Errs = append(e.Errs, err)
}

func (buf *Buffer) pushCert(fSys fs.FS, fileName string, cert *x509.Certificate) {
	if cert == nil {
		return
	}

	buf.mu.Lock()
	buf.unsafeInit()
	err := buf.unsafePushCert(fSys, fileName, cert)
	buf.mu.Unlock()

	if err != nil {
		buf.pushErr(fSys, fileName, err)
	}
}

func (buf *Buffer) unsafePushCert(fSys fs.FS, fileName string, cert *x509.Certificate) error {
	// unique cert reference
	cert, err := buf.certSet.Push(cert)
	if err != nil && !IsExists(err) {
		return err
	}

	// append cert
	src := buf.unsafeGetSource2(fSys, fileName)
	src.Certs = append(src.Certs, cert)
	return nil
}

func (buf *Buffer) pushKey(fSys fs.FS, fileName string, key x509utils.PrivateKey) {
	if key == nil {
		return
	}

	buf.mu.Lock()
	buf.unsafeInit()
	err := buf.unsafePushKey(fSys, fileName, key)
	buf.mu.Unlock()

	if err != nil {
		buf.pushErr(fSys, fileName, err)
	}
}

func (buf *Buffer) unsafePushKey(fSys fs.FS, fileName string, key x509utils.PrivateKey) error {
	// unique key reference
	key, err := buf.keySet.Push(key)
	if err != nil && !IsExists(err) {
		return err
	}

	// append key
	src := buf.unsafeGetSource2(fSys, fileName)
	src.Keys = append(src.Keys, key)
	return nil
}

func (buf *Buffer) unsafeGetSource(name SourceName) *Source {
	if buf.sources == nil {
		buf.sources = make(map[SourceName]*Source)
	}

	e, ok := buf.sources[name]
	if !ok {
		e = &Source{
			SourceName: name,
		}
		buf.sources[name] = e
	}
	return e
}

func (buf *Buffer) unsafeGetSource2(fSys fs.FS, fileName string) *Source {
	name := NewSourceName(fSys, fileName)
	return buf.unsafeGetSource(name)
}
