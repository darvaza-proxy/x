// Package tls aids working with TLS Certificates
package tls

import "crypto/tls"

type (
	// Certificate is an alias of the standard [tls.Certificate]
	Certificate = tls.Certificate

	// ClientHelloInfo is an alias of the standard [tls.ClientHelloInfo].
	ClientHelloInfo = tls.ClientHelloInfo

	// Config is an alias of the standard [tls.Config]
	Config = tls.Config
)
