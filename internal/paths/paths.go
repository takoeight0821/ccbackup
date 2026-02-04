package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHome expands a path that starts with ~ to the user's home directory.
func ExpandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" || path == "~/" {
		return home, nil
	}

	return filepath.Join(home, path[2:]), nil
}

// EnsureDir creates a directory and all parent directories if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
