package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ccbackup",
	Short: "Claude Code history backup tool",
	Long:  `Backup ~/.claude/ history with Git version control.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default $HOME/.config/ccbackup/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("exec", false, "actually execute (default: dry-run)")
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("exec", rootCmd.PersistentFlags().Lookup("exec"))
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(filepath.Join(home, ".config", "ccbackup"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvPrefix("CCBACKUP")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	// Set defaults
	setDefaults(home)
}

func setDefaults(home string) {
	viper.SetDefault("source_dir", filepath.Join(home, ".claude"))
	viper.SetDefault("backup_dir", filepath.Join(home, "claude-backup"))
	viper.SetDefault("include", []string{
		"projects",
		"history.jsonl",
		"plans",
		"todos",
		"usage-data",
		"stats-cache.json",
	})
	viper.SetDefault("lfs_patterns", []string{})
}

// configFilePath returns the path to the config file.
func configFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ccbackup", "config.yaml")
}
