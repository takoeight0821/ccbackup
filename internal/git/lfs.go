package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// LFS wraps git-lfs commands.
type LFS struct {
	Dir    string
	DryRun bool
}

// NewLFS creates a new LFS wrapper.
func NewLFS(dir string) *LFS {
	return &LFS{Dir: dir}
}

// Install runs git lfs install in the repository.
func (l *LFS) Install() error {
	if l.DryRun {
		return nil
	}
	return l.run("install")
}

// Track adds a pattern to be tracked by LFS.
func (l *LFS) Track(pattern string) error {
	if l.DryRun {
		return nil
	}
	return l.run("track", pattern)
}

// run executes a git-lfs command.
func (l *LFS) run(args ...string) error {
	lfsArgs := append([]string{"lfs"}, args...)
	cmd := exec.Command("git", lfsArgs...)
	cmd.Dir = l.Dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
