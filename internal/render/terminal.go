package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/kacheo/devlog/internal/store"
)

// ShowTerminal writes a human-readable day entry plus unresolved global items to w.
func ShowTerminal(entry *store.DayEntry, blockers []store.Item, actionItems []store.Item, w io.Writer) {
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
	if len(actionItems) > 0 {
		fmt.Fprintln(w, "Action Items:")
		for _, a := range actionItems {
			line := fmt.Sprintf("  [%s] %s", store.ShortID(a.ID), a.Text)
			if len(a.Dependencies) > 0 {
				line += " (needs: " + formatDeps(a.Dependencies) + ")"
			}
			fmt.Fprintln(w, line)
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
	if len(blockers) > 0 {
		fmt.Fprintln(w, "Blockers:")
		for _, b := range blockers {
			line := fmt.Sprintf("  [%s] %s", store.ShortID(b.ID), b.Text)
			if len(b.Dependencies) > 0 {
				line += " (needs: " + formatDeps(b.Dependencies) + ")"
			}
			fmt.Fprintln(w, line)
		}
	}
}

func formatDeps(deps []string) string {
	shorts := make([]string, len(deps))
	for i, d := range deps {
		shorts[i] = store.ShortID(d)
	}
	return strings.Join(shorts, ", ")
}
