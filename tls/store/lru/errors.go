package lru

import (
	"io/fs"

	"darvaza.org/core"
)

var (
	ErrNoUpstream  = core.Wrap(core.ErrInvalid, "no upstream store")
	ErrNotFound    = core.QuietWrap(fs.ErrNotExist, "%s", "certificate not found")
	ErrExpired     = core.Wrap(core.ErrInvalid, "expired")
	ErrInvalidName = core.Wrap(core.ErrInvalid, "bad name")
	ErrBadCert     = core.Wrap(core.ErrInvalid, "invalid certificate")
)
