package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takoeight0821/ccbackup/internal/git"
	"github.com/takoeight0821/ccbackup/internal/paths"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize backup repository",
	Long:  `Initialize the backup directory with Git and Git LFS.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("backup-dir", "", "backup directory path")
}

func runInit(cmd *cobra.Command, args []string) error {
	exec := viper.GetBool("exec")
	verbose := viper.GetBool("verbose")
	out := cmd.OutOrStdout()

	// Get backup directory
	backupDir := viper.GetString("backup_dir")
	if flagDir, _ := cmd.Flags().GetString("backup-dir"); flagDir != "" {
		backupDir = flagDir
	}

	backupDir, err := paths.ExpandHome(backupDir)
	if err != nil {
		return fmt.Errorf("expand backup_dir: %w", err)
	}

	cfgPath := configFilePath()

	if !exec {
		// Dry-run output
		fmt.Fprintf(out, "Would create config: %s\n", cfgPath)
		fmt.Fprintf(out, "Would create directory: %s\n", backupDir)
		fmt.Fprintln(out, "Would run: git init")
		fmt.Fprintln(out, "Would run: git lfs install")
		fmt.Fprintln(out, "Would create: .gitattributes (LFS patterns)")
		fmt.Fprintln(out, "Would create: .gitignore")
		fmt.Fprintln(out, "Would run: git add -A && git commit")
		fmt.Fprintln(out, "\nRun with --exec to apply changes.")
		return nil
	}

	// Create config directory and file
	cfgDir := filepath.Dir(cfgPath)
	if err := paths.EnsureDir(cfgDir); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Write config file if it doesn't exist
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := writeDefaultConfig(cfgPath, backupDir); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		if verbose {
			fmt.Fprintf(out, "Created config: %s\n", cfgPath)
		}
	} else if verbose {
		fmt.Fprintf(out, "Config already exists: %s\n", cfgPath)
	}

	// Create backup directory
	if err := paths.EnsureDir(backupDir); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}
	if verbose {
		fmt.Fprintf(out, "Created backup directory: %s\n", backupDir)
	}

	// Initialize Git
	g := git.NewGit(backupDir)
	if err := g.Init(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	if verbose {
		fmt.Fprintln(out, "Initialized git repository")
	}

	// Create .gitignore
	gitignorePath := filepath.Join(backupDir, ".gitignore")
	gitignoreContent := ".DS_Store\n*.swp\n*~\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("create .gitignore: %w", err)
	}

	// Initialize Git LFS (skip if not installed)
	lfs := git.NewLFS(backupDir)
	if err := lfs.Install(); err != nil {
		if verbose {
			fmt.Fprintf(out, "Warning: git-lfs not available, skipping LFS setup\n")
		}
	} else {
		if verbose {
			fmt.Fprintln(out, "Initialized Git LFS")
		}

		// Track LFS patterns
		lfsPatterns := viper.GetStringSlice("lfs_patterns")
		for _, pattern := range lfsPatterns {
			if err := lfs.Track(pattern); err != nil {
				return fmt.Errorf("git lfs track %s: %w", pattern, err)
			}
			if verbose {
				fmt.Fprintf(out, "Configured LFS for: %s\n", pattern)
			}
		}
	}

	// Initial commit
	if err := g.AddAll(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	hasChanges, err := g.HasChanges()
	if err != nil {
		return fmt.Errorf("check changes: %w", err)
	}

	if hasChanges {
		if err := g.Commit("Initialize backup repository"); err != nil {
			return fmt.Errorf("git commit: %w", err)
		}
		if verbose {
			fmt.Fprintln(out, "Created initial commit")
		}
	}

	fmt.Fprintln(out, "Ready! Run 'ccbackup backup --exec' to start backing up.")
	return nil
}

func writeDefaultConfig(path, backupDir string) error {
	home, _ := os.UserHomeDir()
	sourceDir := filepath.Join(home, ".claude")

	content := fmt.Sprintf(`backup_dir: "%s"
source_dir: "%s"
include:
  - projects
  - history.jsonl
  - plans
  - todos
`, backupDir, sourceDir)

	return os.WriteFile(path, []byte(content), 0644)
}
