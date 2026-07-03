//go:build !windows

package appdir

import (
	"os"
	"os/user"
	"strconv"
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

	dir = "/tmp/runtime-"
	u, _ := user.Current()
	if u != nil && u.Username != "" {
		dir += u.Username
	} else {
		dir += uid
	}

	return dir, nil
}
