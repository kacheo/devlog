package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var showCmd = &cobra.Command{
	Use:   "show [today|yesterday|YYYY-MM-DD|week]",
	Short: "Print journal entries",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	// Mutual exclusion
	if len(args) > 0 && globalDate != "" {
		return fmt.Errorf("--date and positional date argument are mutually exclusive")
	}

	// Determine what to show
	target := "today"
	if globalDate != "" {
		target = globalDate
	} else if len(args) > 0 {
		target = args[0]
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

	if target == "week" {
		return showWeek(st, w)
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
	render.ShowTerminal(entry, w)
	return nil
}

func showWeek(st *store.Store, w io.Writer) error {
	now := time.Now()
	var entries []*store.DayEntry
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		y, m, d := date.Date()
		midnight := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		entry, err := st.Load(midnight)
		if err != nil {
			continue // skip days with errors
		}
		if entry != nil {
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
		fmt.Fprintln(w, "(no entries this week)")
		return nil
	}
	for _, entry := range entries {
		fmt.Fprintf(w, "=== %s ===\n", entry.Date.Format("2006-01-02"))
		render.ShowTerminal(entry, w)
		fmt.Fprintln(w)
	}
	return nil
}
