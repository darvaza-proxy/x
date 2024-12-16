package buffer

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// CertKeyPairs groups a key with matching certificates.
type CertKeyPairs struct {
	Key   x509utils.PrivateKey
	Certs []*x509.Certificate
}

// Pairs returns [CertKeyPairs] for all keys in the [Buffer].
func (buf *Buffer) Pairs() ([]CertKeyPairs, error) {
	if buf == nil {
		return nil, core.ErrNilReceiver
	}

	buf.mu.Lock()
	buf.unsafeInit()
	keys := buf.keySet.Values()
	buf.mu.Unlock()

	out := make([]CertKeyPairs, 0, len(keys))
	for _, key := range keys {
		certs := buf.certSet.GetByPrivateKey(key)
		out = append(out, CertKeyPairs{
			Key:   key,
			Certs: certs,
		})
	}

	return out, nil
}
