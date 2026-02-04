package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takoeight0821/ccbackup/internal/git"
	"github.com/takoeight0821/ccbackup/internal/paths"
	"github.com/takoeight0821/ccbackup/internal/sync"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Claude Code history",
	Long:  `Backup ~/.claude/ to the backup directory.`,
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	exec := viper.GetBool("exec")
	verbose := viper.GetBool("verbose")
	out := cmd.OutOrStdout()

	sourceDir, err := paths.ExpandHome(viper.GetString("source_dir"))
	if err != nil {
		return fmt.Errorf("expand source_dir: %w", err)
	}

	backupDir, err := paths.ExpandHome(viper.GetString("backup_dir"))
	if err != nil {
		return fmt.Errorf("expand backup_dir: %w", err)
	}

	// Validate source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Validate backup directory is initialized (only when executing)
	if exec {
		gitDir := filepath.Join(backupDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("backup directory not initialized, run 'ccbackup init --exec' first")
		}
	}

	includePatterns := viper.GetStringSlice("include")

	syncer := sync.NewSyncer(sourceDir, backupDir, includePatterns)
	syncer.DryRun = !exec
	syncer.Verbose = verbose

	ctx := context.Background()

	if !exec {
		// Dry-run: show what would be copied
		items, err := syncer.Plan(ctx)
		if err != nil {
			return fmt.Errorf("plan: %w", err)
		}

		if len(items) == 0 {
			fmt.Fprintln(out, "No changes to backup.")
			return nil
		}

		for _, item := range items {
			fmt.Fprintf(out, "Would copy: %s (%s)\n", item.RelPath, formatSize(item.Size))
		}
		fmt.Fprintln(out, "\nRun with --exec to apply changes.")
		return nil
	}

	// Execute backup
	result, err := syncer.Execute(ctx)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	if result.CopiedCount == 0 {
		fmt.Fprintln(out, "No changes to backup.")
		return nil
	}

	if verbose {
		for _, item := range result.Items {
			fmt.Fprintf(out, "Copied: %s\n", item.RelPath)
		}
	}
	fmt.Fprintf(out, "Copied %d files (%s)\n", result.CopiedCount, formatSize(result.TotalBytes))

	// Git commit
	g := git.NewGit(backupDir)
	if err := g.AddAll(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	hasChanges, err := g.HasChanges()
	if err != nil {
		return fmt.Errorf("check changes: %w", err)
	}

	if hasChanges {
		commitMsg := fmt.Sprintf("Backup %s", time.Now().Format("2006-01-02 15:04"))
		if err := g.Commit(commitMsg); err != nil {
			return fmt.Errorf("git commit: %w", err)
		}
		fmt.Fprintf(out, "Committed: \"%s\"\n", commitMsg)
	}

	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
