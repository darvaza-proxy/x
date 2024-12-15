// Package signer provides utilities for signing X.509 certificates
package signer

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// Signer signs certificates
type Signer struct {
	key  x509utils.PrivateKey
	cert []*x509.Certificate
}

func (*Signer) init() error { return nil }

// IsRSA tells if the [Signer] uses an RSA key.
func (s *Signer) IsRSA() bool {
	if s == nil || s.key == nil {
		return false
	}

	_, ok := s.key.(*rsa.PrivateKey)
	return ok
}

// Public returns the [x509utils.PublicKey] of the [Signer]'s key.
func (s *Signer) Public() x509utils.PublicKey {
	if s == nil || s.key == nil {
		return nil
	}

	pub, ok := s.key.Public().(x509utils.PublicKey)
	if ok {
		return pub
	}

	return nil
}

// Match tells if the public key of the [Signer] is matches the argument.
func (s *Signer) Match(pub crypto.PublicKey) bool {
	if s == nil || s.key == nil || pub == nil {
		return false
	}

	return x509utils.PublicKeyEqual(s.key.Public(), pub)
}

// Sign signs the digest using the [Signer]'s key.
func (s *Signer) Sign(digest []byte) (signature []byte, err error) {
	switch {
	case s == nil:
		return nil, core.ErrNilReceiver
	case s.key == nil:
		return nil, core.Wrap(core.ErrInvalid, "signer not initialized")
	default:
		return s.key.Sign(rand.Reader, digest, nil)
	}
}

// NewFromKeyPair creates a new [Signer] using the provided [x509utils.PrivateKey].
func NewFromKeyPair(key x509utils.PrivateKey, cert []*x509.Certificate) (*Signer, error) {
	if key == nil || key.Public() == nil {
		return nil, core.ErrInvalid
	}

	switch k := key.(type) {
	case *rsa.PrivateKey:
		if err := k.Validate(); err != nil {
			return nil, err
		}
		return makeNew(k, cert)
	case *ecdsa.PrivateKey, ed25519.PrivateKey:
		return makeNew(k, cert)
	case *ed25519.PrivateKey:
		if k != nil {
			return makeNew(*k, cert)
		}
	}

	return nil, core.ErrInvalid
}

func makeNew(key x509utils.PrivateKey, cert []*x509.Certificate) (*Signer, error) {
	s := &Signer{
		key:  key,
		cert: cert,
	}

	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}
