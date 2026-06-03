package render

import (
	"fmt"
	"io"

	"github.com/kacheo/devlog/internal/store"
)

// ShowTerminal writes a human-readable day entry to w.
func ShowTerminal(entry *store.DayEntry, w io.Writer) {
	if entry == nil {
		fmt.Fprintln(w, "(no entry)")
		return
	}
	if len(entry.Sections["notes"]) > 0 {
		fmt.Fprintln(w, "Notes:")
		for i, n := range entry.Sections["notes"] {
			fmt.Fprintf(w, "  [%d] %s\n", i+1, n)
		}
	}
	if len(entry.Sections["action_items"]) > 0 {
		fmt.Fprintln(w, "Action Items:")
		for i, a := range entry.Sections["action_items"] {
			fmt.Fprintf(w, "  [%d] %s\n", i+1, a)
		}
	}
	if len(entry.Commits) > 0 {
		fmt.Fprintln(w, "Commits:")
		for _, c := range entry.Commits {
			fmt.Fprintf(w, "  • %s  %s  (%s)\n", c.SHA, c.Message, c.Repo)
		}
	}
	if len(entry.PRs) > 0 {
		fmt.Fprintln(w, "PRs:")
		for _, pr := range entry.PRs {
			fmt.Fprintf(w, "  • PR #%d %s [%s] (%s)\n", pr.Number, pr.Title, pr.State, pr.Repo)
		}
	}
	if len(entry.Sections["blockers"]) > 0 {
		fmt.Fprintln(w, "Blockers:")
		for i, b := range entry.Sections["blockers"] {
			fmt.Fprintf(w, "  [%d] %s\n", i+1, b)
		}
	}
}
