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
		fmt.Fprintln(w, "(no entry)") //nolint:errcheck
		return
	}
	if len(entry.Sections["notes"]) > 0 {
		fmt.Fprintln(w, "Notes:") //nolint:errcheck
		for _, n := range entry.Sections["notes"] {
			fmt.Fprintf(w, "  • %s\n", n) //nolint:errcheck
		}
	}
	if len(entry.Commits) > 0 {
		fmt.Fprintln(w, "Commits:") //nolint:errcheck
		for _, c := range entry.Commits {
			fmt.Fprintf(w, "  • %s  %s  (%s)\n", c.SHA, c.Message, c.Repo) //nolint:errcheck
		}
	}
	if len(entry.PRs) > 0 {
		fmt.Fprintln(w, "PRs:") //nolint:errcheck
		for _, pr := range entry.PRs {
			fmt.Fprintf(w, "  • PR #%d %s [%s] (%s)\n", pr.Number, pr.Title, pr.State, pr.Repo) //nolint:errcheck
		}
	}
	if len(entry.Sections["blockers"]) > 0 {
		fmt.Fprintln(w, "Blockers:") //nolint:errcheck
		for _, b := range entry.Sections["blockers"] {
			fmt.Fprintf(w, "  • %s\n", b) //nolint:errcheck
		}
	}
}

// StandupTerminal writes a human-readable standup view to w.
// doneEntries: entries in the "since" period (done items).
// todayEntry: today's entry for Notes+Blockers (may be nil).
func StandupTerminal(since, until time.Time, doneEntries []*store.DayEntry, todayEntry *store.DayEntry, w io.Writer) {
	// Header
	fmt.Fprintf(w, "--- Standup: %s ---\n\n", until.Format("Monday, January 2")) //nolint:errcheck

	// Done section
	fmt.Fprintf(w, "Done (since %s):\n", since.Format("2006-01-02")) //nolint:errcheck
	hasDone := false
	for _, e := range doneEntries {
		for _, c := range e.Commits {
			fmt.Fprintf(w, "  • %s  %s  (%s · %s)\n", c.SHA, c.Message, c.Repo, e.Date.Format("2006-01-02")) //nolint:errcheck
			hasDone = true
		}
		for _, pr := range e.PRs {
			fmt.Fprintf(w, "  • PR #%d %s [%s] (%s)\n", pr.Number, pr.Title, pr.State, pr.Repo) //nolint:errcheck
			hasDone = true
		}
	}
	if !hasDone {
		fmt.Fprintln(w, "  (none)") //nolint:errcheck
	}

	// Blockers
	fmt.Fprintln(w, "\nBlockers:") //nolint:errcheck
	hasBlockers := false
	if todayEntry != nil {
		for _, b := range todayEntry.Sections["blockers"] {
			fmt.Fprintf(w, "  • %s\n", b) //nolint:errcheck
			hasBlockers = true
		}
	}
	if !hasBlockers {
		fmt.Fprintln(w, "  (none)") //nolint:errcheck
	}

	// Notes today
	fmt.Fprintln(w, "\nNotes (today so far):") //nolint:errcheck
	hasNotes := false
	if todayEntry != nil {
		for _, n := range todayEntry.Sections["notes"] {
			fmt.Fprintf(w, "  • %s\n", n) //nolint:errcheck
			hasNotes = true
		}
	}
	if !hasNotes {
		fmt.Fprintln(w, "  (none)") //nolint:errcheck
	}
}
