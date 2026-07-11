//go:build !windows

package appdir

import (
	"os"
	"os/user"
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
// a $TMPDIR redirection and defaulting to /tmp.
func getTempDir() (string, error) {
	return core.Coalesce(os.Getenv("TMPDIR"), "/tmp"), nil
}
