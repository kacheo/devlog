package render

import (
	"fmt"
	"io"
	"time"

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
	if len(entry.Sections["action_items"]) > 0 {
		fmt.Fprintln(w, "Action Items:")
		for i, a := range entry.Sections["action_items"] {
			fmt.Fprintf(w, "  [%d] %s\n", i+1, a)
		}
	}
	if len(entry.Sections["blockers"]) > 0 {
		fmt.Fprintln(w, "Blockers:")
		for i, b := range entry.Sections["blockers"] {
			fmt.Fprintf(w, "  [%d] %s\n", i+1, b)
		}
	}
}

// StandupTerminal writes a human-readable standup view to w.
// doneEntries: entries in the "since" period (done items).
// todayEntry: today's entry for Notes+Blockers (may be nil).
func StandupTerminal(since, until time.Time, doneEntries []*store.DayEntry, todayEntry *store.DayEntry, w io.Writer) {
	// Header
	fmt.Fprintf(w, "--- Standup: %s ---\n\n", until.Format("Monday, January 2"))

	// Done section
	fmt.Fprintf(w, "Done (since %s):\n", since.Format("2006-01-02"))
	hasDone := false
	for _, e := range doneEntries {
		for _, c := range e.Commits {
			fmt.Fprintf(w, "  • %s  %s  (%s · %s)\n", c.SHA, c.Message, c.Repo, e.Date.Format("2006-01-02"))
			hasDone = true
		}
		for _, pr := range e.PRs {
			fmt.Fprintf(w, "  • PR #%d %s [%s] (%s)\n", pr.Number, pr.Title, pr.State, pr.Repo)
			hasDone = true
		}
	}
	if !hasDone {
		fmt.Fprintln(w, "  (none)")
	}

	// Blockers
	fmt.Fprintln(w, "\nBlockers:")
	hasBlockers := false
	if todayEntry != nil {
		for _, b := range todayEntry.Sections["blockers"] {
			fmt.Fprintf(w, "  • %s\n", b)
			hasBlockers = true
		}
	}
	if !hasBlockers {
		fmt.Fprintln(w, "  (none)")
	}

	// Notes today
	fmt.Fprintln(w, "\nNotes (today so far):")
	hasNotes := false
	if todayEntry != nil {
		for _, n := range todayEntry.Sections["notes"] {
			fmt.Fprintf(w, "  • %s\n", n)
			hasNotes = true
		}
	}
	if !hasNotes {
		fmt.Fprintln(w, "  (none)")
	}
}
