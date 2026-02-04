package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNeedsSync(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	tests := []struct {
		name string
		src  *FileInfo
		dst  *FileInfo
		want bool
	}{
		{
			name: "dst is nil (new file)",
			src:  &FileInfo{Size: 100, ModTime: now},
			dst:  nil,
			want: true,
		},
		{
			name: "size differs",
			src:  &FileInfo{Size: 200, ModTime: now},
			dst:  &FileInfo{Size: 100, ModTime: now},
			want: true,
		},
		{
			name: "src is newer",
			src:  &FileInfo{Size: 100, ModTime: now},
			dst:  &FileInfo{Size: 100, ModTime: earlier},
			want: true,
		},
		{
			name: "same size and time",
			src:  &FileInfo{Size: 100, ModTime: now},
			dst:  &FileInfo{Size: 100, ModTime: now},
			want: false,
		},
		{
			name: "dst is newer (no sync needed)",
			src:  &FileInfo{Size: 100, ModTime: earlier},
			dst:  &FileInfo{Size: 100, ModTime: now},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsSync(tt.src, tt.dst)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSyncer_Plan(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create test files in src
	require.NoError(t, os.MkdirAll(filepath.Join(src, "projects"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "projects", "session.jsonl"), []byte("hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "history.jsonl"), []byte("world"), 0644))
	// Create files that should NOT be included
	require.NoError(t, os.MkdirAll(filepath.Join(src, "debug"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "debug", "log.txt"), []byte("debug"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "settings.json"), []byte("settings"), 0644))

	syncer := NewSyncer(src, dst, []string{"projects", "history.jsonl"})
	items, err := syncer.Plan(context.Background())
	require.NoError(t, err)

	// Should find 2 files (projects/session.jsonl and history.jsonl), not debug/log.txt or settings.json
	assert.Len(t, items, 2)

	paths := make([]string, len(items))
	for i, item := range items {
		paths[i] = item.RelPath
	}
	assert.Contains(t, paths, "projects/session.jsonl")
	assert.Contains(t, paths, "history.jsonl")
}

func TestSyncer_Plan_ExistingDstFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create test file in src
	srcFile := filepath.Join(src, "history.jsonl")
	require.NoError(t, os.WriteFile(srcFile, []byte("hello"), 0644))

	// Create same file in dst (same size, same time)
	dstFile := filepath.Join(dst, "history.jsonl")
	require.NoError(t, os.WriteFile(dstFile, []byte("hello"), 0644))

	// Set same modtime
	srcInfo, _ := os.Stat(srcFile)
	require.NoError(t, os.Chtimes(dstFile, srcInfo.ModTime(), srcInfo.ModTime()))

	syncer := NewSyncer(src, dst, []string{"history.jsonl"})
	items, err := syncer.Plan(context.Background())
	require.NoError(t, err)

	// Should find 0 files (already synced)
	assert.Len(t, items, 0)
}

func TestSyncer_Execute(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create test files in src
	require.NoError(t, os.MkdirAll(filepath.Join(src, "projects"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "projects", "session.jsonl"), []byte("hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "history.jsonl"), []byte("world"), 0644))

	syncer := NewSyncer(src, dst, []string{"projects", "history.jsonl"})
	result, err := syncer.Execute(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, result.CopiedCount)
	assert.Equal(t, int64(10), result.TotalBytes) // "hello" + "world" = 10 bytes

	// Verify files exist in dst
	assert.FileExists(t, filepath.Join(dst, "projects", "session.jsonl"))
	assert.FileExists(t, filepath.Join(dst, "history.jsonl"))

	// Verify content
	content1, _ := os.ReadFile(filepath.Join(dst, "projects", "session.jsonl"))
	assert.Equal(t, "hello", string(content1))
	content2, _ := os.ReadFile(filepath.Join(dst, "history.jsonl"))
	assert.Equal(t, "world", string(content2))
}

func TestSyncer_RoundTrip(t *testing.T) {
	// Simulate backup -> restore -> compare
	original := t.TempDir()
	backup := t.TempDir()
	restored := t.TempDir()

	// Create original files
	require.NoError(t, os.WriteFile(filepath.Join(original, "history.jsonl"), []byte("data1"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(original, "projects"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(original, "projects", "session.jsonl"), []byte("data2"), 0644))

	includePatterns := []string{"projects", "history.jsonl"}

	// Backup: original -> backup
	backupSyncer := NewSyncer(original, backup, includePatterns)
	_, err := backupSyncer.Execute(context.Background())
	require.NoError(t, err)

	// Restore: backup -> restored
	restoreSyncer := NewSyncer(backup, restored, includePatterns)
	_, err = restoreSyncer.Execute(context.Background())
	require.NoError(t, err)

	// Compare original and restored
	original1, _ := os.ReadFile(filepath.Join(original, "history.jsonl"))
	restored1, _ := os.ReadFile(filepath.Join(restored, "history.jsonl"))
	assert.Equal(t, original1, restored1)

	original2, _ := os.ReadFile(filepath.Join(original, "projects", "session.jsonl"))
	restored2, _ := os.ReadFile(filepath.Join(restored, "projects", "session.jsonl"))
	assert.Equal(t, original2, restored2)
}

func TestSyncer_DryRun(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create test file in src
	require.NoError(t, os.WriteFile(filepath.Join(src, "history.jsonl"), []byte("hello"), 0644))

	syncer := NewSyncer(src, dst, []string{"history.jsonl"})
	syncer.DryRun = true

	result, err := syncer.Execute(context.Background())
	require.NoError(t, err)

	// DryRun should report what would happen
	assert.Equal(t, 1, result.CopiedCount)

	// But file should NOT exist in dst
	_, err = os.Stat(filepath.Join(dst, "history.jsonl"))
	assert.True(t, os.IsNotExist(err))
}
