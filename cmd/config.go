package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takoeight0821/ccbackup/internal/paths"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Show or manage ccbackup configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Run:   runConfigPath,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	sourceDir, err := paths.ExpandHome(viper.GetString("source_dir"))
	if err != nil {
		sourceDir = viper.GetString("source_dir") + " (expansion failed)"
	}
	backupDir, err := paths.ExpandHome(viper.GetString("backup_dir"))
	if err != nil {
		backupDir = viper.GetString("backup_dir") + " (expansion failed)"
	}

	fmt.Fprintf(out, "source_dir: %s\n", sourceDir)
	fmt.Fprintf(out, "backup_dir: %s\n", backupDir)

	fmt.Fprintln(out, "include:")
	for _, pattern := range viper.GetStringSlice("include") {
		fmt.Fprintf(out, "  - %s\n", pattern)
	}

	fmt.Fprintln(out, "lfs_patterns:")
	for _, pattern := range viper.GetStringSlice("lfs_patterns") {
		fmt.Fprintf(out, "  - %s\n", pattern)
	}

	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, configFilePath())
}
