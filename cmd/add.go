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
	addDeps    []string
)

var addCmd = &cobra.Command{
	Use:   "add \"text\"",
	Short: "Append a bullet to today's journal, or add a global blocker/action item",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addSection, "section", "", "section to append to (default: notes)")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "add a tag to today's frontmatter (repeatable)")
	addCmd.Flags().StringArrayVar(&addDeps, "dep", nil, "dependency UUID (8-char prefix or full; repeatable)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	text := args[0]

	// Determine target section
	targetSection := "notes"
	if addSection != "" {
		canonical, ok := store.NormalizeSection(addSection)
		if !ok {
			allSections := store.AllSections()
			return fmt.Errorf("unknown section %q; valid sections: %s",
				addSection, strings.Join(allSections, ", "))
		}
		targetSection = canonical
	}

	// Load config
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Open store
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	// Route global sections (blockers, action_items) to the items store
	if store.IsGlobalSection(targetSection) {
		itemType, _ := store.NormalizeItemType(targetSection)
		item, err := st.AddItem(itemType, text, addDeps)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "added %s: %s\n", strings.ReplaceAll(item.Type, "_", " "), store.ShortID(item.ID))
		return nil
	}

	// Deps flag only makes sense for global items
	if len(addDeps) > 0 {
		return fmt.Errorf("--dep is only valid for blockers and action_items sections")
	}

	// Determine target date for day-file sections
	date, err := resolveDate(globalDate)
	if err != nil {
		return err
	}

	return st.Modify(date, func(entry *store.DayEntry) error {
		entry.Sections[targetSection] = append(entry.Sections[targetSection], text)
		for _, tag := range addTags {
			if !containsStr(entry.Tags, tag) {
				entry.Tags = append(entry.Tags, tag)
			}
		}
		return nil
	})
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
