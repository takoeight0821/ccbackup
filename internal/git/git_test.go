package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_InitAndCommit(t *testing.T) {
	dir := t.TempDir()

	g := NewGit(dir)

	// Init
	err := g.Init()
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(dir, ".git"))

	// Create a file
	testFile := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	// Add
	err = g.Add("test.txt")
	require.NoError(t, err)

	// Commit
	err = g.Commit("Initial commit")
	require.NoError(t, err)

	// Verify commit exists
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Initial commit")
}

func TestGit_AddAll(t *testing.T) {
	dir := t.TempDir()

	g := NewGit(dir)
	require.NoError(t, g.Init())

	// Create multiple files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("2"), 0644))

	// Add all
	err := g.AddAll()
	require.NoError(t, err)

	// Commit and verify
	err = g.Commit("Add all files")
	require.NoError(t, err)

	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Add all files")
}

func TestGit_HasChanges(t *testing.T) {
	dir := t.TempDir()

	g := NewGit(dir)
	require.NoError(t, g.Init())

	// Initially no changes
	hasChanges, err := g.HasChanges()
	require.NoError(t, err)
	assert.False(t, hasChanges)

	// Create a file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644))

	// Now has changes
	hasChanges, err = g.HasChanges()
	require.NoError(t, err)
	assert.True(t, hasChanges)

	// Add and commit
	require.NoError(t, g.AddAll())
	require.NoError(t, g.Commit("commit"))

	// No changes after commit
	hasChanges, err = g.HasChanges()
	require.NoError(t, err)
	assert.False(t, hasChanges)
}

func TestGit_DryRun(t *testing.T) {
	dir := t.TempDir()

	g := NewGit(dir)
	g.DryRun = true

	// All operations should succeed but not actually run
	err := g.Init()
	require.NoError(t, err)

	// .git should NOT exist
	_, err = os.Stat(filepath.Join(dir, ".git"))
	assert.True(t, os.IsNotExist(err))
}

func TestLFS_Track(t *testing.T) {
	// Skip if git-lfs is not installed
	if _, err := exec.LookPath("git-lfs"); err != nil {
		t.Skip("git-lfs not installed")
	}

	dir := t.TempDir()

	g := NewGit(dir)
	require.NoError(t, g.Init())

	lfs := NewLFS(dir)
	require.NoError(t, lfs.Install())

	// Track pattern
	err := lfs.Track("file-history/**/*")
	require.NoError(t, err)

	// Check .gitattributes
	content, err := os.ReadFile(filepath.Join(dir, ".gitattributes"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "file-history/**/*")
	assert.Contains(t, string(content), "filter=lfs")
}

func TestLFS_DryRun(t *testing.T) {
	// Skip if git-lfs is not installed
	if _, err := exec.LookPath("git-lfs"); err != nil {
		t.Skip("git-lfs not installed")
	}

	dir := t.TempDir()

	g := NewGit(dir)
	require.NoError(t, g.Init())

	lfs := NewLFS(dir)
	lfs.DryRun = true

	err := lfs.Install()
	require.NoError(t, err)

	err = lfs.Track("*.bin")
	require.NoError(t, err)

	// .gitattributes should NOT exist (dry run)
	_, err = os.Stat(filepath.Join(dir, ".gitattributes"))
	assert.True(t, os.IsNotExist(err))
}
