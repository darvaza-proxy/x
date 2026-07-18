//go:build !windows

package appdir

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"darvaza.org/core"
)

// getUserRuntimeTempDir composes the last-tier user runtime
// directory, runtime-<user>, under the volatile temporary
// directory returned by [getTempDir].
func getUserRuntimeTempDir() (string, error) {
	name := "runtime-"
	u, _ := user.Current()
	if u != nil && u.Username != "" {
		name += u.Username
	} else {
		name += strconv.Itoa(os.Getuid())
	}

	return joinFn(getTempDir, name)
}

// getTempDir returns the volatile temporary directory, honouring
// a $TMPDIR redirection and defaulting to /tmp. A relative $TMPDIR
// is resolved against the working directory to keep the result
// absolute, which fails only when the working directory is
// unavailable.
func getTempDir() (string, error) {
	return filepath.Abs(core.Coalesce(os.Getenv("TMPDIR"), "/tmp"))
}
