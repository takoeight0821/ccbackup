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

// SyncError records a per-file error that did not abort the sync.
type SyncError struct {
	RelPath string
	Err     error
}

// SyncResult holds the result of a sync operation.
type SyncResult struct {
	CopiedCount int
	TotalBytes  int64
	Items       []SyncItem
	Errors      []SyncError
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

// PlanResult holds items to sync and any warnings encountered during planning.
type PlanResult struct {
	Items    []SyncItem
	Warnings []SyncError
}

// Plan scans the source directory and returns items that need syncing.
func (s *Syncer) Plan(ctx context.Context) (*PlanResult, error) {
	result := &PlanResult{}

	err := filepath.Walk(s.SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// For walk errors on individual files, record and continue.
			// If info is nil, we can try to get the relPath from the path.
			relPath, relErr := filepath.Rel(s.SrcDir, path)
			if relErr != nil {
				relPath = path
			}
			result.Warnings = append(result.Warnings, SyncError{RelPath: relPath, Err: err})
			return nil
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
			result.Items = append(result.Items, SyncItem{
				RelPath: relPath,
				SrcPath: path,
				DstPath: dstPath,
				Size:    info.Size(),
			})
		}

		return nil
	})

	return result, err
}

// Execute performs the sync operation.
func (s *Syncer) Execute(ctx context.Context) (*SyncResult, error) {
	plan, err := s.Plan(ctx)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Items:  plan.Items,
		Errors: plan.Warnings,
	}

	for _, item := range plan.Items {
		if s.DryRun {
			result.CopiedCount++
			result.TotalBytes += item.Size
			continue
		}

		if err := s.copyFile(item.SrcPath, item.DstPath); err != nil {
			result.Errors = append(result.Errors, SyncError{RelPath: item.RelPath, Err: err})
			continue
		}

		result.CopiedCount++
		result.TotalBytes += item.Size
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
