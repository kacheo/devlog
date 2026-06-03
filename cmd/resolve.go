package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <id>",
	Short: "Mark a blocker or action item as resolved",
	Long: `Mark a blocker or action item as resolved.

<id> may be a full UUID or an 8-character prefix.

Examples:
  devlog resolve a1b2c3d4
  devlog resolve a1b2c3d4-e5f6-4789-b012-c3d4e5f67890`,
	Args: cobra.ExactArgs(1),
	RunE: runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	id := args[0]

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	item, err := st.ResolveItem(id)
	if err != nil {
		return err
	}

	typeName := strings.ReplaceAll(item.Type, "_", " ")
	fmt.Fprintf(cmd.OutOrStdout(), "resolved: %s %s — %q\n", typeName, store.ShortID(item.ID), item.Text)
	return nil
}
