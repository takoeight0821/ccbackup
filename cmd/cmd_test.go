package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestViper(t *testing.T, sourceDir, backupDir string) func() {
	t.Helper()
	viper.Reset()
	viper.Set("source_dir", sourceDir)
	viper.Set("backup_dir", backupDir)
	viper.Set("exclude", []string{"debug", "*.json", "!history.jsonl"})
	viper.Set("lfs_patterns", []string{"file-history/**/*"})
	viper.Set("exec", false)
	viper.Set("verbose", false)

	return func() {
		viper.Reset()
	}
}

func TestInitCommand_DryRun(t *testing.T) {
	backupDir := t.TempDir()
	sourceDir := t.TempDir()
	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Would create config:")
	assert.Contains(t, output, "Would run: git init")
	assert.Contains(t, output, "Run with --exec to apply changes.")

	// Verify nothing was created
	_, err = os.Stat(filepath.Join(backupDir, ".git"))
	assert.True(t, os.IsNotExist(err))
}

func TestInitCommand_Exec(t *testing.T) {
	backupDir := t.TempDir()
	sourceDir := t.TempDir()
	configDir := t.TempDir()

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	// Override config file path for test
	cfgFile = filepath.Join(configDir, "config.yaml")

	viper.Set("exec", true)
	viper.Set("verbose", true)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"init", "--exec", "-v"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify git was initialized
	assert.DirExists(t, filepath.Join(backupDir, ".git"))

	// Verify .gitignore was created
	assert.FileExists(t, filepath.Join(backupDir, ".gitignore"))

	output := stdout.String()
	assert.Contains(t, output, "Ready!")
}

func TestBackupCommand_DryRun(t *testing.T) {
	sourceDir := t.TempDir()
	backupDir := t.TempDir()

	// Create test files in source
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "history.jsonl"), []byte("data"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "projects"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "projects", "session.jsonl"), []byte("session"), 0644))

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"backup"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Would copy:")
	assert.Contains(t, output, "history.jsonl")
	assert.Contains(t, output, "Run with --exec to apply changes.")
}

func TestBackupCommand_Exec(t *testing.T) {
	sourceDir := t.TempDir()
	backupDir := t.TempDir()

	// Create test files in source
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "history.jsonl"), []byte("data"), 0644))

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	// Initialize git in backup dir
	initGitRepo(t, backupDir)

	viper.Set("exec", true)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"backup", "--exec"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify file was copied
	assert.FileExists(t, filepath.Join(backupDir, "history.jsonl"))
}

func TestRestoreCommand_DryRun(t *testing.T) {
	sourceDir := t.TempDir()
	backupDir := t.TempDir()

	// Create test files in backup
	require.NoError(t, os.WriteFile(filepath.Join(backupDir, "history.jsonl"), []byte("data"), 0644))

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"restore"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Would restore:")
	assert.Contains(t, output, "history.jsonl")
}

func TestRestoreCommand_Exec(t *testing.T) {
	sourceDir := t.TempDir()
	backupDir := t.TempDir()

	// Create test files in backup
	require.NoError(t, os.WriteFile(filepath.Join(backupDir, "history.jsonl"), []byte("restored"), 0644))

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	viper.Set("exec", true)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"restore", "--exec"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify file was restored
	content, err := os.ReadFile(filepath.Join(sourceDir, "history.jsonl"))
	require.NoError(t, err)
	assert.Equal(t, "restored", string(content))
}

func TestConfigShow(t *testing.T) {
	sourceDir := t.TempDir()
	backupDir := t.TempDir()

	cleanup := setupTestViper(t, sourceDir, backupDir)
	defer cleanup()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "show"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "source_dir:")
	assert.Contains(t, output, "backup_dir:")
	assert.Contains(t, output, "exclude:")
}

func TestConfigPath(t *testing.T) {
	cleanup := setupTestViper(t, "", "")
	defer cleanup()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "path"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "config.yaml")
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := filepath.Join(dir, ".git")
	if _, err := os.Stat(cmd); os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))
		// Minimal git init
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git", "refs", "heads"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]\n\trepositoryformatversion = 0\n\tfilemode = true\n\tbare = false\n"), 0644))
	}
}
