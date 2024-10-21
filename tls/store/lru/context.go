package lru

import (
	"context"
	"crypto/tls"

	"darvaza.org/core"
)

// WithClientHelloInfo attaches a [tls.ClientHelloInfo] to a [context.Context].
func WithClientHelloInfo(ctx context.Context, chi *tls.ClientHelloInfo) context.Context {
	return chiKey.WithValue(ctx, chi)
}

// ClientHelloInfoFormContext extracts a [tls.ClientHelloInfo] from a [context.Context].
func ClientHelloInfo(ctx context.Context) (*tls.ClientHelloInfo, bool) {
	return chiKey.Get(ctx)
}

var chiKey = core.NewContextKey[*tls.ClientHelloInfo]("tls-chi")
