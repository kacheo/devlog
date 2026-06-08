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
	ActionItems []itemJSON     `json:"action_items"`
	Commits     []store.Commit `json:"commits"`
	PRs         []store.PR     `json:"prs"`
	Blockers    []itemJSON     `json:"blockers"`
}

type itemJSON struct {
	ID           string   `json:"id"`
	Text         string   `json:"text"`
	Resolved     bool     `json:"resolved"`
	Dependencies []string `json:"dependencies"`
}

// ShowJSON serializes a single DayEntry plus unresolved global items to JSON (show --json).
// Returns []byte("null") for a nil entry (missing day file).
func ShowJSON(entry *store.DayEntry, blockers []store.Item, actionItems []store.Item) ([]byte, error) {
	if entry == nil {
		return []byte("null"), nil
	}
	v := showDayJSON{
		Version: "1",
		Date:    entry.Date.Format("2006-01-02"),
		Tags:    notNilStrings(entry.Tags),
		Sections: showSectionsJSON{
			Notes:       notNilStrings(entry.Sections["notes"]),
			ActionItems: toItemJSONSlice(actionItems),
			Commits:     notNilCommits(entry.Commits),
			PRs:         notNilPRs(entry.PRs),
			Blockers:    toItemJSONSlice(blockers),
		},
	}
	return json.Marshal(v)
}

// ShowJSONWeek serializes a slice of DayEntry (one per day) as a JSON array.
// Global items are attached to every day entry in the range.
func ShowJSONWeek(entries []*store.DayEntry, blockers []store.Item, actionItems []store.Item) ([]byte, error) {
	days := make([]showDayJSON, 0, len(entries))
	for _, e := range entries {
		days = append(days, toDayJSON(e, blockers, actionItems))
	}
	return json.Marshal(days)
}

func toDayJSON(e *store.DayEntry, blockers []store.Item, actionItems []store.Item) showDayJSON {
	return showDayJSON{
		Version: "1",
		Date:    e.Date.Format("2006-01-02"),
		Tags:    notNilStrings(e.Tags),
		Sections: showSectionsJSON{
			Notes:       notNilStrings(e.Sections["notes"]),
			ActionItems: toItemJSONSlice(actionItems),
			Commits:     notNilCommits(e.Commits),
			PRs:         notNilPRs(e.PRs),
			Blockers:    toItemJSONSlice(blockers),
		},
	}
}

// itemsEntryJSON is the JSON shape for a single item in the items command output.
type itemsEntryJSON struct {
	ID           string   `json:"id"`
	ShortID      string   `json:"short_id"`
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	Resolved     bool     `json:"resolved"`
	ResolvedAt   *string  `json:"resolved_at,omitempty"` // RFC3339, omitted when nil
	Dependencies []string `json:"dependencies"`
	Due          *string  `json:"due,omitempty"`     // YYYY-MM-DD, action_item only
	ETA          *string  `json:"eta,omitempty"`     // YYYY-MM-DD, blocker only
	Overdue      bool     `json:"overdue"`           // true when past due and unresolved
}

// ItemsJSON serializes a slice of items for the items command (--json output).
func ItemsJSON(items []store.Item) ([]byte, error) {
	out := make([]itemsEntryJSON, len(items))
	for i, it := range items {
		deps := it.Dependencies
		if deps == nil {
			deps = []string{}
		}
		var resolvedAt *string
		if it.ResolvedAt != nil {
			s := it.ResolvedAt.UTC().Format("2006-01-02T15:04:05Z")
			resolvedAt = &s
		}
		var due, eta *string
		if it.Due != nil {
			s := it.Due.String()
			due = &s
		}
		if it.ETA != nil {
			s := it.ETA.String()
			eta = &s
		}
		out[i] = itemsEntryJSON{
			ID:           it.ID,
			ShortID:      store.ShortID(it.ID),
			Type:         it.Type,
			Text:         it.Text,
			Resolved:     it.Resolved,
			ResolvedAt:   resolvedAt,
			Dependencies: deps,
			Due:          due,
			ETA:          eta,
			Overdue:      it.IsOverdue(),
		}
	}
	return json.Marshal(out)
}

func toItemJSONSlice(items []store.Item) []itemJSON {
	if len(items) == 0 {
		return []itemJSON{}
	}
	out := make([]itemJSON, len(items))
	for i, it := range items {
		deps := it.Dependencies
		if deps == nil {
			deps = []string{}
		}
		out[i] = itemJSON{
			ID:           it.ID,
			Text:         it.Text,
			Resolved:     it.Resolved,
			Dependencies: deps,
		}
	}
	return out
}

type tagsJSON struct {
	Version string         `json:"version"`
	Tags    []tagCountJSON `json:"tags"`
}

type tagCountJSON struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// TagsJSON serializes tag counts for the tags list command (--json output).
func TagsJSON(tags []store.TagCount) ([]byte, error) {
	out := make([]tagCountJSON, len(tags))
	for i, tc := range tags {
		out[i] = tagCountJSON{Tag: tc.Tag, Count: tc.Count}
	}
	return json.Marshal(tagsJSON{Version: "1", Tags: out})
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
