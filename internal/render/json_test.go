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
	return e
}

func makeBlocker(id, text string) store.Item {
	return store.Item{ID: id, Type: "blocker", Text: text, Resolved: false, Dependencies: []string{}}
}


func TestShowJSON_SingleDay(t *testing.T) {
	entry := makeEntry("2026-06-01")
	blockers := []store.Item{makeBlocker("deadbeef-0000-4000-8000-aabbccddeeff", "Blocked on review")}
	out, err := ShowJSON(entry, blockers, nil)
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
	// Blockers are now objects with id/text/resolved/dependencies
	blockersJSON := sections["blockers"].([]interface{})
	if len(blockersJSON) != 1 {
		t.Fatalf("blockers len = %d, want 1", len(blockersJSON))
	}
	b := blockersJSON[0].(map[string]interface{})
	if b["text"] != "Blocked on review" {
		t.Errorf("blocker text = %v", b["text"])
	}
	if b["resolved"] != false {
		t.Errorf("blocker resolved = %v, want false", b["resolved"])
	}
}

func TestShowJSON_NilEntry_IsNull(t *testing.T) {
	out, err := ShowJSON(nil, nil, nil)
	if err != nil {
		t.Fatalf("ShowJSON(nil) error = %v", err)
	}
	if string(out) != "null" {
		t.Errorf("ShowJSON(nil) = %q, want null", out)
	}
}

func TestShowJSONWeek_Array(t *testing.T) {
	entries := []*store.DayEntry{makeEntry("2026-06-01"), makeEntry("2026-06-02")}
	out, err := ShowJSONWeek(entries, nil, nil)
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

func TestShowJSON_EmptyItemsSlices(t *testing.T) {
	entry := makeEntry("2026-06-01")
	out, err := ShowJSON(entry, []store.Item{}, []store.Item{})
	if err != nil {
		t.Fatalf("ShowJSON() error = %v", err)
	}
	var v map[string]interface{}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	sections := v["sections"].(map[string]interface{})
	blockers := sections["blockers"].([]interface{})
	if len(blockers) != 0 {
		t.Errorf("blockers = %v, want []", blockers)
	}
	ai := sections["action_items"].([]interface{})
	if len(ai) != 0 {
		t.Errorf("action_items = %v, want []", ai)
	}
}

func TestItemsJSON_Unresolved(t *testing.T) {
	items := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "blocker", Text: "waiting on infra", Resolved: false, Dependencies: []string{}},
	}
	b, err := ItemsJSON(items)
	if err != nil {
		t.Fatalf("ItemsJSON: %v", err)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	if out[0]["short_id"] != "deadbeef" {
		t.Errorf("short_id = %v, want deadbeef", out[0]["short_id"])
	}
	if out[0]["type"] != "blocker" {
		t.Errorf("type = %v, want blocker", out[0]["type"])
	}
	if _, ok := out[0]["resolved_at"]; ok {
		t.Error("resolved_at should be omitted for unresolved items")
	}
}

func TestItemsJSON_Resolved(t *testing.T) {
	ts := time.Date(2026, 6, 3, 14, 22, 0, 0, time.UTC)
	items := []store.Item{
		{ID: "aabbccdd-0000-4000-8000-112233445566", Type: "action_item", Text: "write docs", Resolved: true, ResolvedAt: &ts, Dependencies: []string{}},
	}
	b, err := ItemsJSON(items)
	if err != nil {
		t.Fatalf("ItemsJSON: %v", err)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out[0]["resolved"] != true {
		t.Errorf("resolved = %v, want true", out[0]["resolved"])
	}
	if out[0]["resolved_at"] != "2026-06-03T14:22:00Z" {
		t.Errorf("resolved_at = %v, want 2026-06-03T14:22:00Z", out[0]["resolved_at"])
	}
}

func TestItemsJSON_Empty(t *testing.T) {
	b, err := ItemsJSON([]store.Item{})
	if err != nil {
		t.Fatalf("ItemsJSON: %v", err)
	}
	if string(b) != "[]" {
		t.Errorf("expected [], got %s", b)
	}
}
