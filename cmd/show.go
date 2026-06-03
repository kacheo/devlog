package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var (
	showTag      string
	showRepo     string
	showSections []string
	showFrom     string
	showUntil    string
)

var showCmd = &cobra.Command{
	Use:   "show [today|yesterday|YYYY-MM-DD|week]",
	Short: "Print journal entries",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringVar(&showTag, "tag", "", "include only entries with this tag")
	showCmd.Flags().StringVar(&showRepo, "repo", "", "include only entries with commits/PRs from this repo")
	showCmd.Flags().StringSliceVar(&showSections, "section", nil,
		"render only these sections: notes, blockers, action_items (repeatable or comma-separated)")
	showCmd.Flags().StringVar(&showFrom, "from", "", "start of date range (inclusive)")
	showCmd.Flags().StringVar(&showUntil, "until", "", "end of date range (inclusive), defaults to today")
}

func runShow(cmd *cobra.Command, args []string) error {
	// Mutual exclusions
	if len(args) > 0 && globalDate != "" {
		return fmt.Errorf("--date and positional date argument are mutually exclusive")
	}
	if (showFrom != "" || showUntil != "") && globalDate != "" {
		return fmt.Errorf("--date and --from/--until are mutually exclusive")
	}
	if (showFrom != "" || showUntil != "") && len(args) > 0 {
		return fmt.Errorf("--from/--until and positional date argument are mutually exclusive")
	}
	if showUntil != "" && showFrom == "" {
		return fmt.Errorf("--until requires --from")
	}

	if err := validateShowSections(showSections); err != nil {
		return err
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	w := cmd.OutOrStdout()

	// Date range mode
	if showFrom != "" {
		from, err := store.ParseDate(showFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		var until time.Time
		if showUntil != "" {
			until, err = store.ParseDate(showUntil)
			if err != nil {
				return fmt.Errorf("invalid --until date: %w", err)
			}
		} else {
			now := time.Now()
			y, m, d := now.Date()
			until = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		}
		return showRange(st, from, until, w)
	}

	// Determine what to show (single day or week shorthand)
	target := "today"
	if globalDate != "" {
		target = globalDate
	} else if len(args) > 0 {
		target = args[0]
	}

	if target == "week" {
		now := time.Now()
		y, m, d := now.Date()
		until := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		from := until.AddDate(0, 0, -6)
		return showRange(st, from, until, w)
	}

	// Single day
	date, err := store.ParseDate(target)
	if err != nil {
		return err
	}
	entry, err := st.Load(date)
	if err != nil {
		return fmt.Errorf("loading day file: %w", err)
	}

	if !entryMatchesFilters(entry) {
		fmt.Fprintln(w, "(no entry)")
		return nil
	}

	if globalJSON {
		out, err := render.ShowJSON(entry)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	if entry == nil {
		fmt.Fprintln(w, "(no entry)")
		return nil
	}
	render.ShowTerminal(filteredEntry(entry), w)
	return nil
}

func showRange(st *store.Store, from, until time.Time, w io.Writer) error {
	var entries []*store.DayEntry
	for cursor := from; !cursor.After(until); cursor = cursor.AddDate(0, 0, 1) {
		entry, err := st.Load(cursor)
		if err != nil {
			continue // skip days with errors
		}
		if entry != nil && entryMatchesFilters(entry) {
			entries = append(entries, entry)
		}
	}

	if globalJSON {
		out, err := render.ShowJSONWeek(entries)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	if len(entries) == 0 {
		fmt.Fprintln(w, "(no entries)")
		return nil
	}
	for _, entry := range entries {
		fmt.Fprintf(w, "=== %s ===\n", entry.Date.Format("2006-01-02"))
		render.ShowTerminal(filteredEntry(entry), w)
		fmt.Fprintln(w)
	}
	return nil
}

// entryMatchesFilters returns true if entry satisfies the active showTag and showRepo filters.
func entryMatchesFilters(entry *store.DayEntry) bool {
	if entry == nil {
		return false
	}
	if showTag != "" {
		found := false
		for _, t := range entry.Tags {
			if strings.EqualFold(t, showTag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if showRepo != "" {
		found := false
		for _, c := range entry.Commits {
			if strings.EqualFold(c.Repo, showRepo) {
				found = true
				break
			}
		}
		if !found {
			for _, pr := range entry.PRs {
				if strings.EqualFold(pr.Repo, showRepo) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// filteredEntry returns a shallow copy of entry with only the sections requested by showSections.
// If showSections is empty, the original entry is returned unchanged.
func filteredEntry(entry *store.DayEntry) *store.DayEntry {
	if len(showSections) == 0 {
		return entry
	}
	copy := *entry
	copy.Sections = make(map[string][]string, len(showSections))
	for _, sec := range showSections {
		copy.Sections[sec] = entry.Sections[sec]
	}
	return &copy
}

func validateShowSections(sections []string) error {
	for _, s := range sections {
		if _, ok := store.NormalizeSection(s); !ok {
			return fmt.Errorf("unknown section %q: valid sections are %s",
				s, strings.Join(store.KnownSections, ", "))
		}
	}
	return nil
}
