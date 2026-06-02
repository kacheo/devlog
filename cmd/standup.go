package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var standupSince string

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Print a standup summary",
	Args:  cobra.NoArgs,
	RunE:  runStandup,
}

func init() {
	standupCmd.Flags().StringVar(&standupSince, "since", "", "date from which to collect done items (default: yesterday)")
	rootCmd.AddCommand(standupCmd)
}

func runStandup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	// Resolve since date (default: yesterday)
	sinceStr := standupSince
	if sinceStr == "" {
		sinceStr = "yesterday"
	}
	since, err := store.ParseDate(sinceStr)
	if err != nil {
		return fmt.Errorf("invalid --since date: %w", err)
	}

	// Until = today (midnight)
	now := time.Now()
	y, m, d := now.Date()
	until := time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	// Collect done entries: all days from since up to (but not including) until
	var doneEntries []*store.DayEntry
	for cursor := since; cursor.Before(until); cursor = cursor.AddDate(0, 0, 1) {
		entry, err := st.Load(cursor)
		if err != nil {
			return fmt.Errorf("loading entry for %s: %w", cursor.Format("2006-01-02"), err)
		}
		if entry != nil {
			doneEntries = append(doneEntries, entry)
		}
	}

	// Today's entry for blockers/notes (missing today is fine)
	todayEntry, err := st.Load(until)
	if err != nil {
		return fmt.Errorf("loading today's entry: %w", err)
	}

	w := cmd.OutOrStdout()

	if globalJSON {
		out, err := render.StandupJSON(since, until, doneEntries, todayEntry)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	render.StandupTerminal(since, until, doneEntries, todayEntry, w)
	return nil
}
