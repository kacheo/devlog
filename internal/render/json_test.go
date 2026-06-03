package render

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)


func makeEntry(dateStr string) *store.DayEntry {
	date, _ := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	e := store.EmptyEntry(date)
	e.Tags = []string{"auth"}
	e.Commits = []store.Commit{{SHA: "abc1234", Message: "fix: auth bug", Repo: "api"}}
	e.PRs = []store.PR{{Number: 10, Title: "Fix auth", State: "merged", Repo: "api"}}
	e.Sections["notes"] = []string{"Did some work"}
	e.Sections["blockers"] = []string{"Blocked on review"}
	return e
}

func TestShowJSON_SingleDay(t *testing.T) {
	entry := makeEntry("2026-06-01")
	out, err := ShowJSON(entry)
	if err != nil {
		t.Fatalf("ShowJSON() error = %v", err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if v["version"] != "1" {
		t.Errorf("version = %v, want '1'", v["version"])
	}
	if v["date"] != "2026-06-01" {
		t.Errorf("date = %v, want '2026-06-01'", v["date"])
	}
	sections := v["sections"].(map[string]interface{})
	notes := sections["notes"].([]interface{})
	if len(notes) != 1 || notes[0] != "Did some work" {
		t.Errorf("notes = %v", notes)
	}
	commits := sections["commits"].([]interface{})
	if len(commits) != 1 {
		t.Fatalf("commits len = %d, want 1", len(commits))
	}
	c := commits[0].(map[string]interface{})
	if c["sha"] != "abc1234" {
		t.Errorf("commit sha = %v", c["sha"])
	}
	prs := sections["prs"].([]interface{})
	if len(prs) != 1 {
		t.Fatalf("prs len = %d, want 1", len(prs))
	}
	pr := prs[0].(map[string]interface{})
	if pr["number"] != float64(10) {
		t.Errorf("pr number = %v", pr["number"])
	}
}

func TestShowJSON_NilEntry_IsNull(t *testing.T) {
	out, err := ShowJSON(nil)
	if err != nil {
		t.Fatalf("ShowJSON(nil) error = %v", err)
	}
	if string(out) != "null" {
		t.Errorf("ShowJSON(nil) = %q, want null", out)
	}
}

func TestShowJSONWeek_Array(t *testing.T) {
	entries := []*store.DayEntry{makeEntry("2026-06-01"), makeEntry("2026-06-02")}
	out, err := ShowJSONWeek(entries)
	if err != nil {
		t.Fatalf("ShowJSONWeek() error = %v", err)
	}
	var v []interface{}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(v) != 2 {
		t.Errorf("len = %d, want 2", len(v))
	}
	day := v[0].(map[string]interface{})
	if day["version"] != "1" {
		t.Errorf("version = %v", day["version"])
	}
}

