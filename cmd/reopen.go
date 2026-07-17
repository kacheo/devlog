package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a resolved blocker or action item",
	Long: `Reopen a resolved blocker or action item.

<id> may be a full UUID or an 8-character prefix.

Examples:
  devlog reopen a1b2c3d4
  devlog reopen a1b2c3d4-e5f6-4789-b012-c3d4e5f67890`,
	Args: cobra.ExactArgs(1),
	RunE: runReopen,
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}

func runReopen(cmd *cobra.Command, args []string) error {
	id := args[0]

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	item, err := st.ReopenItem(id)
	if err != nil {
		return err
	}

	typeName := strings.ReplaceAll(item.Type, "_", " ")
	fmt.Fprintf(cmd.OutOrStdout(), "reopened: %s %s — %q\n", typeName, store.ShortID(item.ID), item.Text)
	return nil
}
