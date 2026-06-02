package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var (
	updateSection string
	updateID      int
)

var updateCmd = &cobra.Command{
	Use:   "update \"new text\"",
	Short: "Replace a bullet in a section by its 1-based index",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateSection, "section", "notes", "section to update")
	updateCmd.Flags().IntVar(&updateID, "id", 0, "1-based index of the bullet to replace")
	_ = updateCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	text := args[0]

	canonical, ok := store.NormalizeSection(updateSection)
	if !ok {
		return fmt.Errorf("unknown section %q; valid sections: %s",
			updateSection, strings.Join(store.KnownSections, ", "))
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	date, err := resolveDate(globalDate)
	if err != nil {
		return err
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	return st.Modify(date, func(entry *store.DayEntry) error {
		bullets := entry.Sections[canonical]
		if updateID < 1 || updateID > len(bullets) {
			return fmt.Errorf("id %d out of range: section %q has %d item(s)", updateID, canonical, len(bullets))
		}
		entry.Sections[canonical][updateID-1] = text
		return nil
	})
}
