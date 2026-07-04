//go:build !windows

package appdir

import (
	"os"
	"os/user"
	"strconv"

	"darvaza.org/core"
)

func getUserDataDir() (string, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir != "" {
		return dir, nil
	}

	dir, err := os.UserHomeDir()
	switch {
	case err != nil:
		return "", err
	default:
		return dir + "/.local/share", nil
	}
}

func getUserRuntimeDir() (string, error) {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir != "" {
		return dir, nil
	}

	// systemd special
	uid := strconv.Itoa(os.Getuid())
	dir = "/run/user/" + uid
	st, _ := os.Stat(dir)
	if st != nil && st.IsDir() {
		return dir, nil
	}

	name := "runtime-"
	u, _ := user.Current()
	if u != nil && u.Username != "" {
		name += u.Username
	} else {
		name += uid
	}

	return joinFn(getTempDir, name)
}

// getTempDir returns the volatile temporary directory, honouring
// a $TMPDIR redirection and defaulting to /tmp.
func getTempDir() (string, error) {
	return core.Coalesce(os.Getenv("TMPDIR"), "/tmp"), nil
}
