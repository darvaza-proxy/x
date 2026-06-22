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

// NewListener is an alias of the standard [tls.NewListener]. It wraps a
// net.Listener so accepted connections negotiate TLS using the given [Config] —
// pair it with an sni.Dispatcher to serve unclaimed connections.
var NewListener = tls.NewListener
