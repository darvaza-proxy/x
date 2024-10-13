package certbotdir

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"os"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

func (s *Store) getByName(domain string) (*tls.Certificate, error) {
	certFile, keyFile := s.cfg.FilePair(domain)
	return s.getByFilePair(certFile, keyFile)
}

func (*Store) getByFilePair(certFile, keyFile string) (*tls.Certificate, error) {
	key, err := readKey(keyFile)
	if err != nil {
		return nil, err
	}

	certs, err := readCerts(certFile)
	if err != nil {
		return nil, err
	}

	rawCerts, err := buildRawChain(certs)
	if err != nil {
		return nil, err
	}

	// TODO: validate

	return &tls.Certificate{
		Certificate: rawCerts,
		Leaf:        certs[0],
		PrivateKey:  key,
	}, nil
}

func readCerts(certFile string) ([]*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	certs, err := readPEMCerts(certPEM)
	if e, ok := err.(*core.CompoundError); ok {
		err = asPathErrors(e, certFile, "pem.Decode")
	}
	return certs, err
}

func readKey(keyFile string) (x509utils.PrivateKey, error) {
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	key, err := readPEMKeys(keyPEM)
	if e, ok := err.(*core.CompoundError); ok {
		err = asPathErrors(e, keyFile, "pem.Decode")
	}
	return key, err
}

func readPEMCerts(certPEM []byte) ([]*x509.Certificate, error) {
	var errs core.CompoundError
	var certs []*x509.Certificate

	err := x509utils.ReadPEM(certPEM, func(_ fs.FS, _ string, block *pem.Block) bool {
		cert, err := x509utils.BlockToCertificate(block)
		switch {
		case err == x509utils.ErrIgnored:
			errs.Append(core.ErrInvalid, "%s", "not certificate block")
		case err != nil:
			errs.AppendError(err)
		case cert != nil:
			certs = append(certs, cert)
		}
		return true
	})

	if err != nil {
		errs.AppendError(err)
	}

	return certs, errs.AsError()
}

func readPEMKeys(keyPEM []byte) (x509utils.PrivateKey, error) {
	var errs core.CompoundError
	var keys []x509utils.PrivateKey
	var key x509utils.PrivateKey

	err := x509utils.ReadPEM(keyPEM, func(_ fs.FS, _ string, block *pem.Block) bool {
		key, err := x509utils.BlockToPrivateKey(block)
		switch {
		case err == x509utils.ErrIgnored:
			errs.Append(core.ErrInvalid, "%s", "not private key block")
		case err != nil:
			errs.AppendError(err)
		case key != nil:
			keys = append(keys, key)
		}
		return true
	})

	if err != nil {
		errs.AppendError(err)
	}

	if l := len(keys); l == 0 {
		errs.Append(core.ErrInvalid, "file contains no key")
	} else if l > 1 {
		errs.Append(core.ErrInvalid, "file contains %v keys", l)
	} else {
		key = keys[0]
	}

	return key, errs.AsError()
}

func buildRawChain(certs []*x509.Certificate) ([][]byte, error) {
	raw := make([][]byte, len(certs))
	for i, c := range certs {
		if len(c.Raw) == 0 {
			err := &x509utils.ErrInvalidCert{
				Cert:   c,
				Reason: "missing DER raw data",
			}
			return nil, err
		}

		raw[i] = c.Raw
	}
	return raw, nil
}

func asPathErrors(errs *core.CompoundError, fileName, op string) error {
	if len(errs.Errs) == 0 {
		return nil
	}

	for i, err := range errs.Errs {
		errs.Errs[i] = asPathError(err, fileName, op)
	}

	return errs
}

func asPathError(err error, fileName, op string) *fs.PathError {
	e, ok := err.(*fs.PathError)
	if ok {
		// Set Path and possibly Op
		return &fs.PathError{
			Path: fileName,
			Op:   core.Coalesce(e.Op, op),
			Err:  e.Err,
		}
	}

	// wrap
	return &fs.PathError{
		Path: fileName,
		Op:   op,
		Err:  err,
	}
}
