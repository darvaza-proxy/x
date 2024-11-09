package autocert

import (
	"golang.org/x/crypto/acme"
)

const (
	// LetsEncryptURL is the Directory endpoint of Let's Encrypt CA.
	LetsEncryptURL = acme.LetsEncryptURL
	// LetsEncryptStagingURL is the Directory endpoint of Let's Encrypt CA Staging Environment.
	LetsEncryptStagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

	// ALPNProto is the ALPN protocol name used by a CA server when validating
	// tls-alpn-01 challenges.
	ALPNProto = acme.ALPNProto
)
