package lru

import (
	"darvaza.org/core"
)

var (
	// ErrNoUpstream is returned when the LRU cache does not have an upstream store.
	ErrNoUpstream = core.Wrap(core.ErrInvalid, "no upstream store")
	// ErrNotFound is returned when a certificate is not found in the LRU cache.
	ErrNotFound = core.QuietWrap(core.ErrNotExists, "certificate not found")
	// ErrExpired is returned when a certificate is expired.
	ErrExpired = core.Wrap(core.ErrInvalid, "expired")
	// ErrInvalidName is returned when a certificate name is invalid.
	ErrInvalidName = core.Wrap(core.ErrInvalid, "bad name")
	// ErrBadCert is returned when a certificate is invalid.
	ErrBadCert = core.Wrap(core.ErrInvalid, "invalid certificate")
)
