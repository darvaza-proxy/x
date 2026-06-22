package buffer

import (
	"context"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

// IsCancelled indicates the error represents a context cancellation or expiration.
func IsCancelled(err error) bool {
	return core.IsError(err, context.Canceled, context.DeadlineExceeded)
}

// IsExists indicates the error means something already exists.
func IsExists(err error) bool {
	return core.IsError(err, core.ErrExists, fs.ErrExist)
}

// returnAdd2 induces an [x509utils.ErrEmpty] error only when there was
// genuinely nothing to operate on (inputs == 0). It must NOT fire when inputs
// were present but all already existed: an all-duplicate apply added nothing
// (count == 0) yet is a successful idempotent no-op, not a failure. count is
// the number of items actually added.
func returnAdd2(inputs, count int, err error) (int, error) {
	if inputs == 0 && err == nil {
		err = x509utils.ErrEmpty
	}
	return count, err
}
