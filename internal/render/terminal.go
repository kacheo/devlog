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
			if a.Due != nil {
				if a.IsOverdue() {
					line += "  ⚠ OVERDUE (due " + a.Due.String() + ")"
				} else {
					line += "  (due " + a.Due.String() + ")"
				}
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
			if b.ETA != nil {
				line += "  (eta " + b.ETA.String() + ")"
			}
			fmt.Fprintln(w, line)
		}
	}
}

// ItemsTerminal writes a human-readable list of items to w.
// Each line shows the short ID, text, and (for resolved items) the resolution date.
func ItemsTerminal(items []store.Item, w io.Writer) {
	if len(items) == 0 {
		fmt.Fprintln(w, "(none)")
		return
	}
	for _, it := range items {
		line := fmt.Sprintf("  [%s] %s", store.ShortID(it.ID), it.Text)
		if it.Due != nil {
			if it.IsOverdue() {
				line += "  ⚠ OVERDUE (due " + it.Due.String() + ")"
			} else {
				line += "  (due " + it.Due.String() + ")"
			}
		}
		if it.ETA != nil {
			line += "  (eta " + it.ETA.String() + ")"
		}
		if it.Resolved && it.ResolvedAt != nil {
			line += "  (resolved " + it.ResolvedAt.Format("2006-01-02") + ")"
		}
		fmt.Fprintln(w, line)
	}
}

// TagsTerminal writes a two-column tag/count list to w.
func TagsTerminal(tags []store.TagCount, w io.Writer) {
	if len(tags) == 0 {
		fmt.Fprintln(w, "(no tags)")
		return
	}
	for _, tc := range tags {
		fmt.Fprintf(w, "  %-30s %d\n", tc.Tag, tc.Count)
	}
}

func formatDeps(deps []string) string {
	shorts := make([]string, len(deps))
	for i, d := range deps {
		shorts[i] = store.ShortID(d)
	}
	return strings.Join(shorts, ", ")
}
