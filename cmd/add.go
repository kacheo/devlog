package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var (
	addSection string
	addTags    []string
)

var addCmd = &cobra.Command{
	Use:   "add \"text\"",
	Short: "Append a bullet to today's journal",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addSection, "section", "", "section to append to (default: notes)")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "add a tag to today's frontmatter (repeatable)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	text := args[0]

	// Determine target section
	targetSection := "notes"
	if addSection != "" {
		canonical, ok := store.NormalizeSection(addSection)
		if !ok {
			return fmt.Errorf("unknown section %q; valid sections: %s",
				addSection, strings.Join(store.KnownSections, ", "))
		}
		targetSection = canonical
	}

	// Load config
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine target date
	date, err := resolveDate(globalDate)
	if err != nil {
		return err
	}

	// Open store
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	// Load or create entry
	entry, err := st.LoadOrCreate(date)
	if err != nil {
		return fmt.Errorf("loading day file: %w", err)
	}

	// Append bullet
	entry.Sections[targetSection] = append(entry.Sections[targetSection], text)

	// Merge tags (deduplicate)
	for _, tag := range addTags {
		if !containsStr(entry.Tags, tag) {
			entry.Tags = append(entry.Tags, tag)
		}
	}

	// Save
	if err := st.Save(entry); err != nil {
		return fmt.Errorf("saving day file: %w", err)
	}
	return nil
}

// resolveDate returns the target date from a string ("today", "yesterday", "YYYY-MM-DD", or "").
// Empty string resolves to today.
func resolveDate(s string) (time.Time, error) {
	if s == "" {
		s = "today"
	}
	return store.ParseDate(s)
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
