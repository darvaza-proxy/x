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

// returnAdd2 induces an [x509utils.ErrEmpty] error when nothing was found.
func returnAdd2(count int, err error) (int, error) {
	if count == 0 && err == nil {
		err = x509utils.ErrEmpty
	}
	return count, err
}
