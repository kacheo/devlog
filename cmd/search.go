package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/store"
)

var (
	searchSections []string
	searchFrom     string
	searchUntil    string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search across all journal entries",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSliceVar(&searchSections, "section", nil,
		"sections to search: notes, blockers, action_items, commits, prs (repeatable or comma-separated)")
	searchCmd.Flags().StringVar(&searchFrom, "from", "", "earliest date to search (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchUntil, "until", "", "latest date to search (YYYY-MM-DD)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	opts := store.SearchOpts{
		Query:    args[0],
		Sections: searchSections,
	}

	if searchFrom != "" {
		t, err := time.ParseInLocation("2006-01-02", searchFrom, time.Local)
		if err != nil {
			return fmt.Errorf("invalid --from %q: expected YYYY-MM-DD", searchFrom)
		}
		opts.From = t
	}
	if searchUntil != "" {
		t, err := time.ParseInLocation("2006-01-02", searchUntil, time.Local)
		if err != nil {
			return fmt.Errorf("invalid --until %q: expected YYYY-MM-DD", searchUntil)
		}
		if !opts.From.IsZero() && t.Before(opts.From) {
			return fmt.Errorf("--until %s is before --from %s", searchUntil, searchFrom)
		}
		opts.Until = t
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	results, err := st.Search(opts)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return fmt.Errorf("no matches for %q", opts.Query)
	}

	w := cmd.OutOrStdout()

	if globalJSON {
		out, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
		return nil
	}

	for _, r := range results {
		fmt.Fprintf(w, "%s  |  %-12s  |  %s\n", r.Date, r.Section, r.Line)
	}
	return nil
}
