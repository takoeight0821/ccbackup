package sync

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileInfo holds file metadata for comparison.
type FileInfo struct {
	Size    int64
	ModTime time.Time
}

// SyncItem represents a file to be synced.
type SyncItem struct {
	RelPath string
	SrcPath string
	DstPath string
	Size    int64
}

// SyncResult holds the result of a sync operation.
type SyncResult struct {
	CopiedCount int
	TotalBytes  int64
	Items       []SyncItem
}

// Syncer handles file synchronization between source and destination.
type Syncer struct {
	SrcDir  string
	DstDir  string
	Filter  *Filter
	DryRun  bool
	Verbose bool
}

// NewSyncer creates a new Syncer.
func NewSyncer(srcDir, dstDir string, includePatterns []string) *Syncer {
	return &Syncer{
		SrcDir: srcDir,
		DstDir: dstDir,
		Filter: NewFilter(includePatterns),
	}
}

// NeedsSync returns true if src needs to be synced to dst.
func NeedsSync(src, dst *FileInfo) bool {
	if dst == nil {
		return true
	}
	if src.Size != dst.Size {
		return true
	}
	if src.ModTime.After(dst.ModTime) {
		return true
	}
	return false
}

// Plan scans the source directory and returns items that need syncing.
func (s *Syncer) Plan(ctx context.Context) ([]SyncItem, error) {
	var items []SyncItem

	err := filepath.Walk(s.SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.SrcDir, path)
		if err != nil {
			return err
		}

		// Check filter
		if !s.Filter.ShouldInclude(relPath) {
			return nil
		}

		// Check if destination file exists and needs sync
		dstPath := filepath.Join(s.DstDir, relPath)
		srcInfo := &FileInfo{Size: info.Size(), ModTime: info.ModTime()}

		var dstInfo *FileInfo
		if dstStat, err := os.Stat(dstPath); err == nil {
			dstInfo = &FileInfo{Size: dstStat.Size(), ModTime: dstStat.ModTime()}
		}

		if NeedsSync(srcInfo, dstInfo) {
			items = append(items, SyncItem{
				RelPath: relPath,
				SrcPath: path,
				DstPath: dstPath,
				Size:    info.Size(),
			})
		}

		return nil
	})

	return items, err
}

// Execute performs the sync operation.
func (s *Syncer) Execute(ctx context.Context) (*SyncResult, error) {
	items, err := s.Plan(ctx)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{Items: items}

	for _, item := range items {
		result.CopiedCount++
		result.TotalBytes += item.Size

		if s.DryRun {
			continue
		}

		if err := s.copyFile(item.SrcPath, item.DstPath); err != nil {
			return result, err
		}
	}

	return result, nil
}

// copyFile copies a file from src to dst, preserving modtime.
func (s *Syncer) copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Preserve modification time
	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}
