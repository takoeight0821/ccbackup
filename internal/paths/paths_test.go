package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tilde with subpath", "~/.claude", filepath.Join(home, ".claude")},
		{"tilde with spaces in path", "~/OneDrive - Cybozu", filepath.Join(home, "OneDrive - Cybozu")},
		{"absolute path", "/absolute/path", "/absolute/path"},
		{"relative path", "relative/path", "relative/path"},
		{"tilde only", "~", home},
		{"tilde with slash only", "~/", home},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHome(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	err := EnsureDir(dir)
	require.NoError(t, err)
	assert.DirExists(t, dir)

	// Call again on existing dir should not error
	err = EnsureDir(dir)
	require.NoError(t, err)
}
