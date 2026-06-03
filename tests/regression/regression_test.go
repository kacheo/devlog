//go:build regression

// Package regression_test pins the externally-visible JSON output format of the
// devlog CLI. Tests seed a workspace with fixed-date entries, run commands, and
// compare output against committed golden files in ./golden/.
//
// To regenerate golden files after an intentional format change:
//
//	make update-golden
//
// Run: make test-regression
package regression_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kacheo/devlog/tests/testharness"
)

// goldenPath returns the absolute path to a golden file by name.
func goldenPath(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "golden", name)
}

// checkOrUpdate either updates the golden file (DEVLOG_GOLDEN_UPDATE=1) or
// diffs got against the golden file and reports any mismatch.
func checkOrUpdate(t *testing.T, name, got string) {
	t.Helper()
	path := goldenPath(name)
	if os.Getenv("DEVLOG_GOLDEN_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir golden: %v", err)
		}
		if err := os.WriteFile(path, []byte(got+"\n"), 0644); err != nil {
			t.Fatalf("writing golden %s: %v", name, err)
		}
		t.Logf("updated golden: %s", name)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v\nRun 'make update-golden' to generate it.", name, err)
	}
	if strings.TrimSpace(string(want)) != strings.TrimSpace(got) {
		t.Errorf("output does not match golden %s\n--- want ---\n%s\n--- got ---\n%s",
			name, want, got)
	}
}

// prettyJSON re-marshals raw JSON with indentation for stable, readable diffs.
func prettyJSON(t *testing.T, raw string) string {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("prettyJSON: invalid JSON: %v\nraw: %s", err, raw)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("prettyJSON: marshal: %v", err)
	}
	return string(out)
}

// seedWorkspace populates a workspace with fixed-date entries for all regression scenarios.
func seedWorkspace(t *testing.T, ws *testharness.Workspace) {
	t.Helper()
	// Day 1: 2026-01-15 — two notes, one blocker, tags
	ws.MustRun("add", "--date", "2026-01-15", "--tag", "auth", "--tag", "backend", "fixed the OAuth token refresh loop")
	ws.MustRun("add", "--date", "2026-01-15", "reviewed PR for rate limiter")
	ws.MustRun("add", "--date", "2026-01-15", "--section", "blockers", "waiting on DevOps to provision staging DB")
	// Day 2: 2026-01-16 — one note, one action item, different tag
	ws.MustRun("add", "--date", "2026-01-16", "--tag", "infra", "deployed new cache layer")
	ws.MustRun("add", "--date", "2026-01-16", "--section", "action_items", "write runbook for cache failover")
}

// TestRegression_ShowSingleDay pins show --json for a single day.
func TestRegression_ShowSingleDay(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	seedWorkspace(t, ws)
	stdout, _, err := ws.Run("--json", "show", "2026-01-15")
	if err != nil {
		t.Fatalf("show --json 2026-01-15: %v", err)
	}
	checkOrUpdate(t, "show_single_day.json", prettyJSON(t, strings.TrimSpace(stdout)))
}

// TestRegression_ShowRange pins show --json --from/--until for a two-day range.
// Both dates are fixed, so the output is deterministic.
func TestRegression_ShowRange(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	seedWorkspace(t, ws)
	stdout, _, err := ws.Run("--json", "show", "--from", "2026-01-15", "--until", "2026-01-16")
	if err != nil {
		t.Fatalf("show --json --from/--until: %v", err)
	}
	checkOrUpdate(t, "show_range.json", prettyJSON(t, strings.TrimSpace(stdout)))
}

// TestRegression_SearchJSON pins search --json output for a known query.
func TestRegression_SearchJSON(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	seedWorkspace(t, ws)
	stdout, _, err := ws.Run("--json", "search", "OAuth")
	if err != nil {
		t.Fatalf("search --json OAuth: %v", err)
	}
	checkOrUpdate(t, "search_oauth.json", prettyJSON(t, strings.TrimSpace(stdout)))
}

// TestRegression_RootHelpText pins devlog --help to catch command/flag renames.
func TestRegression_RootHelpText(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	stdout, _, err := ws.Run("--help")
	if err != nil {
		t.Fatalf("devlog --help: %v", err)
	}
	checkOrUpdate(t, "help_root.txt", strings.TrimSpace(stdout))
}

// TestRegression_AddHelpText pins devlog add --help.
func TestRegression_AddHelpText(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	stdout, _, err := ws.Run("add", "--help")
	if err != nil {
		t.Fatalf("devlog add --help: %v", err)
	}
	checkOrUpdate(t, "help_add.txt", strings.TrimSpace(stdout))
}

// TestRegression_ShowHelpText pins devlog show --help (catches new --from/--until flags, etc.).
func TestRegression_ShowHelpText(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	stdout, _, err := ws.Run("show", "--help")
	if err != nil {
		t.Fatalf("devlog show --help: %v", err)
	}
	checkOrUpdate(t, "help_show.txt", strings.TrimSpace(stdout))
}
