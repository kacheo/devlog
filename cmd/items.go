package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
	"github.com/kacheo/devlog/internal/render"
	"github.com/kacheo/devlog/internal/store"
)

var (
	itemsResolved bool
	itemsAll      bool
	itemsOverdue  bool
	itemsType     string
	itemsFrom     string
	itemsUntil    string
)

var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "List blockers and action items",
	Long: `List blockers and action items.

By default shows only unresolved items. Use --resolved to see what has been
resolved, or --all to see everything.

--from and --until filter on the resolution date when used with --resolved or --all.

Examples:
  devlog items                              # open items
  devlog items --resolved                   # resolved items
  devlog items --resolved --from yesterday  # resolved since yesterday
  devlog items --all --type blockers        # every blocker ever
  devlog items --overdue                    # action items past their due date
  devlog items --json                       # machine-readable`,
	Args: cobra.NoArgs,
	RunE: runItems,
}

func init() {
	itemsCmd.Flags().BoolVar(&itemsResolved, "resolved", false, "show only resolved items")
	itemsCmd.Flags().BoolVar(&itemsAll, "all", false, "show both resolved and unresolved items")
	itemsCmd.Flags().BoolVar(&itemsOverdue, "overdue", false, "show only overdue action items (past due date)")
	itemsCmd.Flags().StringVar(&itemsType, "type", "", "filter by type: blockers or action_items")
	itemsCmd.Flags().StringVar(&itemsFrom, "from", "", "show resolved items resolved on or after this date (YYYY-MM-DD)")
	itemsCmd.Flags().StringVar(&itemsUntil, "until", "", "show resolved items resolved on or before this date (YYYY-MM-DD)")
	rootCmd.AddCommand(itemsCmd)
}

func runItems(cmd *cobra.Command, _ []string) error {
	if itemsResolved && itemsAll {
		return fmt.Errorf("--resolved and --all are mutually exclusive")
	}
	if itemsOverdue && (itemsResolved || itemsAll) {
		return fmt.Errorf("--overdue is mutually exclusive with --resolved and --all")
	}

	// Validate --type
	var wantType string
	if itemsType != "" {
		canonical, ok := store.NormalizeSection(itemsType)
		if !ok || !store.IsGlobalSection(canonical) {
			return fmt.Errorf("unknown --type %q: valid values are blockers, action_items", itemsType)
		}
		wantType = canonical
	}

	// Parse date range for resolved filter
	var from, until time.Time
	if itemsFrom != "" {
		t, err := store.ParseDate(itemsFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		from = t
	}
	if itemsUntil != "" {
		t, err := store.ParseDate(itemsUntil)
		if err != nil {
			return fmt.Errorf("invalid --until date: %w", err)
		}
		// Include the full until day
		until = t.Add(24*time.Hour - time.Second)
	}
	if !until.IsZero() && !from.IsZero() && until.Before(from) {
		return fmt.Errorf("--until must not be before --from")
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	st, err := store.New(cfg.Journal.Dir)
	if err != nil {
		return fmt.Errorf("journal not configured: %w\nRun 'devlog init' to set up", err)
	}

	all, err := st.LoadAllItems()
	if err != nil {
		return fmt.Errorf("loading items: %w", err)
	}

	var items []store.Item
	switch {
	case itemsAll:
		resolved := store.FilterResolved(all, from, until)
		unresolved := store.FilterUnresolved(all)
		items = append(unresolved, resolved...)
	case itemsResolved:
		items = store.FilterResolved(all, from, until)
	case itemsOverdue:
		items = store.FilterOverdue(all)
	default:
		items = store.FilterUnresolved(all)
	}

	// Filter by type
	if wantType != "" {
		filtered := make([]store.Item, 0, len(items))
		for _, it := range items {
			itSection, _ := store.NormalizeItemType(it.Type)
			// NormalizeItemType returns "blocker"/"action_item"; wantType is the section form
			// Map section → item type for comparison
			wantItemType, _ := store.NormalizeItemType(wantType)
			if itSection == wantItemType {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}

	w := cmd.OutOrStdout()

	if globalJSON {
		b, err := render.ItemsJSON(items)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(b))
		return nil
	}

	label := itemsLabel()
	fmt.Fprintf(w, "%s:\n", label)
	render.ItemsTerminal(items, w)
	return nil
}

func itemsLabel() string {
	parts := []string{}
	switch {
	case itemsAll:
		parts = append(parts, "All items")
	case itemsResolved:
		parts = append(parts, "Resolved items")
	case itemsOverdue:
		parts = append(parts, "Overdue items")
	default:
		parts = append(parts, "Open items")
	}
	if itemsType != "" {
		parts = append(parts, "("+strings.ReplaceAll(itemsType, "_", " ")+")")
	}
	return strings.Join(parts, " ")
}
