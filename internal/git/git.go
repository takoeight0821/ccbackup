package git

import (
	"os/exec"
	"strings"
)

// Git wraps git commands.
type Git struct {
	Dir    string
	DryRun bool
}

// NewGit creates a new Git wrapper.
func NewGit(dir string) *Git {
	return &Git{Dir: dir}
}

// Init initializes a git repository.
func (g *Git) Init() error {
	if g.DryRun {
		return nil
	}
	return g.run("init")
}

// Add adds a file to the staging area.
func (g *Git) Add(path string) error {
	if g.DryRun {
		return nil
	}
	return g.run("add", path)
}

// AddAll adds all files to the staging area.
func (g *Git) AddAll() error {
	if g.DryRun {
		return nil
	}
	return g.run("add", "-A")
}

// Commit creates a commit with the given message.
func (g *Git) Commit(message string) error {
	if g.DryRun {
		return nil
	}
	return g.run("commit", "-m", message)
}

// HasChanges returns true if there are uncommitted changes.
func (g *Git) HasChanges() (bool, error) {
	if g.DryRun {
		return false, nil
	}
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.Dir
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// run executes a git command.
func (g *Git) run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	return cmd.Run()
}
