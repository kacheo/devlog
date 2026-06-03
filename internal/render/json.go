package render

import (
	"encoding/json"

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
