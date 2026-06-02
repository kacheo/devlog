package cmd

import (
	"github.com/spf13/cobra"
)

var (
	globalJSON bool
	globalDate string
)

var rootCmd = &cobra.Command{
	Use:   "devlog",
	Short: "Local-first engineer journal",
	Long: `devlog keeps a structured daily work journal as plain markdown files.
It auto-imports git commits and GitHub PRs, and outputs structured JSON for AI agents.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalJSON, "json", false, "emit structured JSON output")
	rootCmd.PersistentFlags().StringVar(&globalDate, "date", "", "target date (YYYY-MM-DD)")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
