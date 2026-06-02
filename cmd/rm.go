package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var (
	rmSection string
	rmID      int
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a bullet from a section by its 1-based index",
	Args:  cobra.NoArgs,
	RunE:  runRm,
}

func init() {
	rmCmd.Flags().StringVar(&rmSection, "section", "notes", "section to remove from")
	rmCmd.Flags().IntVar(&rmID, "id", 0, "1-based index of the bullet to remove")
	_ = rmCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	canonical, ok := store.NormalizeSection(rmSection)
	if !ok {
		return fmt.Errorf("unknown section %q; valid sections: %s",
			rmSection, strings.Join(store.KnownSections, ", "))
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
		if rmID < 1 || rmID > len(bullets) {
			return fmt.Errorf("id %d out of range: section %q has %d item(s)", rmID, canonical, len(bullets))
		}
		entry.Sections[canonical] = append(bullets[:rmID-1], bullets[rmID:]...)
		return nil
	})
}
