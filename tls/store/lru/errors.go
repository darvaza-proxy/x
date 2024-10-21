package lru

import (
	"darvaza.org/core"
)

var (
	// ErrNotInitialized is returned when the LRU cache has not been initialized.
	// [LRU] must be created via [Config.New] or [LRU.Init].
	ErrNotInitialized = core.Wrap(core.ErrInvalid, "store not initialized")
	// ErrAlreadyInitialized is returned by [LRU.Init] when the LRU cache has already been initialized.
	ErrAlreadyInitialized = core.Wrap(core.ErrExists, "store already initialized")
	// ErrNoClientHelloInfo is returned when the Get() interface is used but the upstream
	// only supports GetCertificate()
	ErrNoClientHelloInfo = core.QuietWrap(core.ErrInvalid, "upstream requires ClientHelloInfo")

	// ErrNotFound is returned when a certificate is not found in the LRU cache.
	ErrNotFound = core.QuietWrap(core.ErrNotExists, "certificate not found")
	// ErrExpired is returned when a certificate is expired.
	ErrExpired = core.Wrap(core.ErrInvalid, "expired")
	// ErrBadCert is returned when a certificate is invalid.
	ErrBadCert = core.Wrap(core.ErrInvalid, "invalid certificate")
)
