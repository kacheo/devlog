package render

import (
	"encoding/json"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

// showDayJSON is the versioned JSON shape for a single day (show --json).
type showDayJSON struct {
	Version  string           `json:"version"`
	Date     string           `json:"date"`
	Tags     []string         `json:"tags"`
	Sections showSectionsJSON `json:"sections"`
}

type showSectionsJSON struct {
	Notes       []string       `json:"notes"`
	ActionItems []string       `json:"action_items"`
	Commits     []store.Commit `json:"commits"`
	PRs         []store.PR     `json:"prs"`
	Blockers    []string       `json:"blockers"`
}

// ShowJSON serializes a single DayEntry to JSON (show --json).
// Returns []byte("null") for a nil entry (missing day file).
func ShowJSON(entry *store.DayEntry) ([]byte, error) {
	if entry == nil {
		return []byte("null"), nil
	}
	v := showDayJSON{
		Version: "1",
		Date:    entry.Date.Format("2006-01-02"),
		Tags:    notNilStrings(entry.Tags),
		Sections: showSectionsJSON{
			Notes:       notNilStrings(entry.Sections["notes"]),
			ActionItems: notNilStrings(entry.Sections["action_items"]),
			Commits:     notNilCommits(entry.Commits),
			PRs:         notNilPRs(entry.PRs),
			Blockers:    notNilStrings(entry.Sections["blockers"]),
		},
	}
	return json.Marshal(v)
}

// ShowJSONWeek serializes a slice of DayEntry (one per day) as a JSON array.
func ShowJSONWeek(entries []*store.DayEntry) ([]byte, error) {
	days := make([]showDayJSON, 0, len(entries))
	for _, e := range entries {
		days = append(days, toDayJSON(e))
	}
	return json.Marshal(days)
}

func toDayJSON(e *store.DayEntry) showDayJSON {
	return showDayJSON{
		Version: "1",
		Date:    e.Date.Format("2006-01-02"),
		Tags:    notNilStrings(e.Tags),
		Sections: showSectionsJSON{
			Notes:       notNilStrings(e.Sections["notes"]),
			ActionItems: notNilStrings(e.Sections["action_items"]),
			Commits:     notNilCommits(e.Commits),
			PRs:         notNilPRs(e.PRs),
			Blockers:    notNilStrings(e.Sections["blockers"]),
		},
	}
}

// standupDoneItem represents a commit or PR in the standup JSON output.
type standupDoneItem struct {
	Type    string `json:"type"`
	SHA     string `json:"sha,omitempty"`
	Message string `json:"message,omitempty"`
	Number  int    `json:"number,omitempty"`
	Title   string `json:"title,omitempty"`
	State   string `json:"state,omitempty"`
	Repo    string `json:"repo"`
	Date    string `json:"date"`
}

type standupBlockerItem struct {
	Text string `json:"text"`
	Date string `json:"date"`
}

type standupNoteItem struct {
	Text string `json:"text"`
	Date string `json:"date"`
}

type standupPeriod struct {
	Since string `json:"since"`
	Until string `json:"until"`
}

type standupJSON struct {
	Version     string               `json:"version"`
	GeneratedAt string               `json:"generated_at"`
	Period      standupPeriod        `json:"period"`
	Done        []standupDoneItem    `json:"done"`
	Blockers    []standupBlockerItem `json:"blockers"`
	Notes       []standupNoteItem    `json:"notes"`
}

// StandupJSON builds the standup JSON output.
// doneEntries: entries from the since period (for Done items).
// todayEntry: today's entry (for Blockers and Notes); may be nil.
func StandupJSON(since, until time.Time, doneEntries []*store.DayEntry, todayEntry *store.DayEntry) ([]byte, error) {
	done := make([]standupDoneItem, 0)
	for _, e := range doneEntries {
		dateStr := e.Date.Format("2006-01-02")
		for _, c := range e.Commits {
			done = append(done, standupDoneItem{
				Type:    "commit",
				SHA:     c.SHA,
				Message: c.Message,
				Repo:    c.Repo,
				Date:    dateStr,
			})
		}
		for _, pr := range e.PRs {
			done = append(done, standupDoneItem{
				Type:   "pr",
				Number: pr.Number,
				Title:  pr.Title,
				State:  pr.State,
				Repo:   pr.Repo,
				Date:   dateStr,
			})
		}
	}

	blockers := make([]standupBlockerItem, 0)
	notes := make([]standupNoteItem, 0)
	if todayEntry != nil {
		dateStr := todayEntry.Date.Format("2006-01-02")
		for _, b := range todayEntry.Sections["blockers"] {
			blockers = append(blockers, standupBlockerItem{Text: b, Date: dateStr})
		}
		for _, n := range todayEntry.Sections["notes"] {
			notes = append(notes, standupNoteItem{Text: n, Date: dateStr})
		}
	}

	v := standupJSON{
		Version:     "1",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Period: standupPeriod{
			Since: since.Format("2006-01-02"),
			Until: until.Format("2006-01-02"),
		},
		Done:     done,
		Blockers: blockers,
		Notes:    notes,
	}
	return json.Marshal(v)
}

func notNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func notNilCommits(c []store.Commit) []store.Commit {
	if c == nil {
		return []store.Commit{}
	}
	return c
}

func notNilPRs(p []store.PR) []store.PR {
	if p == nil {
		return []store.PR{}
	}
	return p
}
