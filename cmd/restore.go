package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takoeight0821/ccbackup/internal/paths"
	"github.com/takoeight0821/ccbackup/internal/sync"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore Claude Code history",
	Long:  `Restore backup to ~/.claude/.`,
	RunE:  runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
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

	// Restore is the reverse of backup: from backupDir to sourceDir
	// Use include patterns to filter what gets restored
	includePatterns := viper.GetStringSlice("include")

	syncer := sync.NewSyncer(backupDir, sourceDir, includePatterns)
	syncer.DryRun = !exec
	syncer.Verbose = verbose

	ctx := context.Background()

	if !exec {
		// Dry-run: show what would be restored
		plan, err := syncer.Plan(ctx)
		if err != nil {
			return fmt.Errorf("plan: %w", err)
		}

		for _, w := range plan.Warnings {
			fmt.Fprintf(out, "Warning: %s: %v\n", w.RelPath, w.Err)
		}

		if len(plan.Items) == 0 {
			fmt.Fprintln(out, "No changes to restore.")
			return nil
		}

		for _, item := range plan.Items {
			fmt.Fprintf(out, "Would restore: %s (%s)\n", item.RelPath, formatSize(item.Size))
		}
		fmt.Fprintln(out, "\nRun with --exec to apply changes.")
		return nil
	}

	// Execute restore
	result, err := syncer.Execute(ctx)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	if result.CopiedCount == 0 && len(result.Errors) == 0 {
		fmt.Fprintln(out, "No changes to restore.")
		return nil
	}

	if verbose {
		for _, item := range result.Items {
			fmt.Fprintf(out, "Restored: %s\n", item.RelPath)
		}
	}
	if result.CopiedCount > 0 {
		fmt.Fprintf(out, "Restored %d files (%s)\n", result.CopiedCount, formatSize(result.TotalBytes))
	}

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(out, "Failed: %s: %v\n", e.RelPath, e.Err)
		}
		return fmt.Errorf("%d file(s) failed to sync", len(result.Errors))
	}

	return nil
}
