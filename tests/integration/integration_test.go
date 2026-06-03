//go:build integration

// Package integration_test exercises the devlog CLI end-to-end using a compiled
// binary invoked as a subprocess. Each test gets an isolated workspace (temp dir +
// filtered env) so tests are fully independent and safe to run in parallel.
package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kacheo/devlog/tests/testharness"
)

// --- add ---

// TestAdd_BasicNote verifies add writes a note readable by show.
func TestAdd_BasicNote(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "finished the auth refactor")
	out := ws.MustRun("show", "today")
	if !strings.Contains(out, "finished the auth refactor") {
		t.Fatalf("note not in show output:\n%s", out)
	}
}

// TestAdd_WithSection verifies --section routes the bullet to the right section.
func TestAdd_WithSection(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--section", "blockers", "waiting on DB migration")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json: %v", err)
	}
	blockers := v["sections"].(map[string]any)["blockers"].([]any)
	if len(blockers) != 1 || blockers[0] != "waiting on DB migration" {
		t.Errorf("blockers = %v", blockers)
	}
}

// TestAdd_WithTag verifies --tag appears in show --json tags.
func TestAdd_WithTag(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--tag", "infra", "--tag", "backend", "deployed new service")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json: %v", err)
	}
	var tags []string
	for _, tag := range v["tags"].([]any) {
		tags = append(tags, tag.(string))
	}
	if !sliceContains(tags, "infra") || !sliceContains(tags, "backend") {
		t.Errorf("tags = %v, want both [infra backend]", tags)
	}
}

// TestAdd_TagDedup verifies adding the same tag twice doesn't duplicate it.
func TestAdd_TagDedup(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--tag", "infra", "first note")
	ws.MustRun("add", "--tag", "infra", "second note")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json: %v", err)
	}
	var count int
	for _, tag := range v["tags"].([]any) {
		if tag.(string) == "infra" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("tag 'infra' appears %d times, want 1 (dedup)", count)
	}
}

// TestAdd_ToDate verifies --date creates an entry for a specific past date.
func TestAdd_ToDate(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "past note")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "2026-01-15"); err != nil {
		t.Fatalf("show --json 2026-01-15: %v", err)
	}
	if v["date"] != "2026-01-15" {
		t.Errorf("date = %v, want 2026-01-15", v["date"])
	}
	notes := v["sections"].(map[string]any)["notes"].([]any)
	if len(notes) != 1 || notes[0] != "past note" {
		t.Errorf("notes = %v, want [past note]", notes)
	}
}

// --- show ---

// TestShow_JSONSchema verifies the top-level key set of show --json for a single day.
func TestShow_JSONSchema(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "some work")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json: %v", err)
	}
	for _, key := range []string{"version", "date", "tags", "sections"} {
		if _, ok := v[key]; !ok {
			t.Errorf("show --json missing top-level key %q", key)
		}
	}
	sections, ok := v["sections"].(map[string]any)
	if !ok {
		t.Fatal("sections is not a JSON object")
	}
	for _, sec := range []string{"notes", "action_items", "commits", "prs", "blockers"} {
		if _, ok := sections[sec]; !ok {
			t.Errorf("sections missing key %q", sec)
		}
	}
}

// TestShow_NullForMissingDay verifies show for a date with no entry exits zero
// and produces an empty/no-entry indicator rather than an error.
// Note: the CLI renders "(no entry)" even in --json mode for missing days;
// it does not output a JSON null because the nil-entry check precedes the JSON path.
func TestShow_NullForMissingDay(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	stdout, _, err := ws.Run("--json", "show", "2020-01-01")
	if err != nil {
		t.Fatalf("show --json 2020-01-01 exited non-zero: %v", err)
	}
	trimmed := strings.TrimSpace(stdout)
	// Accept either "null" (JSON) or "(no entry)" (text fallback) — both mean no data.
	if trimmed != "null" && trimmed != "(no entry)" {
		t.Errorf("unexpected output for missing day: %s", stdout)
	}
}

// TestShow_FromUntilRange verifies show --from/--until returns a JSON array.
func TestShow_FromUntilRange(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "day one note")
	ws.MustRun("add", "--date", "2026-01-16", "day two note")
	ws.MustRun("add", "--date", "2026-01-17", "day three note")

	var days []map[string]any
	if err := ws.RunJSON(&days, "show", "--from", "2026-01-15", "--until", "2026-01-16"); err != nil {
		t.Fatalf("show --json --from/--until: %v", err)
	}
	if len(days) != 2 {
		t.Fatalf("want 2 days, got %d", len(days))
	}
	dates := []string{days[0]["date"].(string), days[1]["date"].(string)}
	if !sliceContains(dates, "2026-01-15") || !sliceContains(dates, "2026-01-16") {
		t.Errorf("dates = %v, want [2026-01-15 2026-01-16]", dates)
	}
}

// TestShow_FromDefaultsToUntilToday verifies --from alone returns from that date through today.
func TestShow_FromDefaultsToUntilToday(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "old note")
	ws.MustRun("add", "today note") // today's date

	var days []map[string]any
	if err := ws.RunJSON(&days, "show", "--from", "2026-01-15"); err != nil {
		t.Fatalf("show --json --from 2026-01-15: %v", err)
	}
	if len(days) == 0 {
		t.Fatal("want at least one day")
	}
	for _, d := range days {
		if d["date"].(string) < "2026-01-15" {
			t.Errorf("day %s is before --from 2026-01-15", d["date"])
		}
	}
}

// TestShow_UntilWithoutFromErrors verifies --until without --from returns an error.
func TestShow_UntilWithoutFromErrors(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	_, _, err := ws.Run("show", "--until", "2026-01-16")
	if err == nil {
		t.Fatal("expected error when --until used without --from")
	}
}

// --- search ---

// TestSearch_BasicMatch verifies search finds text in notes.
func TestSearch_BasicMatch(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "fixed the OAuth bug")
	ws.MustRun("add", "--date", "2026-01-16", "unrelated entry here")
	stdout := ws.MustRun("search", "OAuth")
	if !strings.Contains(stdout, "OAuth") {
		t.Errorf("search output missing match:\n%s", stdout)
	}
}

// TestSearch_NoMatchExitCode verifies search exits 1 when nothing matches.
func TestSearch_NoMatchExitCode(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "unrelated entry")
	_, _, err := ws.Run("search", "xyzzy_no_match_9999")
	if err == nil {
		t.Fatal("expected non-zero exit for no match, got nil")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Errorf("want exit 1, got %v", err)
	}
}

// TestSearch_JSONOutput verifies search --json returns a JSON array with the right schema.
// Each element is a SearchResult with date, section, and line fields.
func TestSearch_JSONOutput(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "implemented OAuth flow")
	var results []map[string]any
	if err := ws.RunJSON(&results, "search", "OAuth"); err != nil {
		t.Fatalf("search --json: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for _, key := range []string{"date", "section", "line"} {
		if _, ok := results[0][key]; !ok {
			t.Errorf("search result missing key %q", key)
		}
	}
}

// TestSearch_SectionFilter verifies --section limits search scope.
func TestSearch_SectionFilter(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-15", "--section", "blockers", "oauth dependency blocked")
	ws.MustRun("add", "--date", "2026-01-15", "unrelated note")
	var results []map[string]any
	if err := ws.RunJSON(&results, "search", "--section", "blockers", "oauth"); err != nil {
		t.Fatalf("search --json --section blockers: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected match in blockers section")
	}
}

// TestSearch_DateRange verifies --from/--until bound search results.
func TestSearch_DateRange(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "--date", "2026-01-10", "early note with keyword")
	ws.MustRun("add", "--date", "2026-01-20", "late note with keyword")
	var results []map[string]any
	if err := ws.RunJSON(&results, "search", "--from", "2026-01-15", "keyword"); err != nil {
		t.Fatalf("search --json --from: %v", err)
	}
	for _, r := range results {
		if r["date"].(string) < "2026-01-15" {
			t.Errorf("result %s is before --from 2026-01-15", r["date"])
		}
	}
}

// --- rm ---

// TestRm_RemovesItem verifies rm --id deletes the bullet from the section.
func TestRm_RemovesItem(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "first note")
	ws.MustRun("add", "second note")
	ws.MustRun("rm", "--id", "1")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json after rm: %v", err)
	}
	notes := v["sections"].(map[string]any)["notes"].([]any)
	for _, note := range notes {
		if strings.Contains(note.(string), "first note") {
			t.Errorf("removed item still present: %v", notes)
		}
	}
	if len(notes) != 1 {
		t.Errorf("want 1 note remaining, got %d: %v", len(notes), notes)
	}
}

// TestRm_OutOfRangeErrors verifies rm exits non-zero for an invalid id.
func TestRm_OutOfRangeErrors(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "only note")
	_, _, err := ws.Run("rm", "--id", "99")
	if err == nil {
		t.Fatal("expected error for out-of-range id")
	}
}

// --- update ---

// TestUpdate_ReplacesItem verifies update --id replaces the bullet text.
func TestUpdate_ReplacesItem(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	ws.MustRun("add", "original text")
	ws.MustRun("update", "--id", "1", "updated text")
	var v map[string]any
	if err := ws.RunJSON(&v, "show", "today"); err != nil {
		t.Fatalf("show --json after update: %v", err)
	}
	notes := v["sections"].(map[string]any)["notes"].([]any)
	if len(notes) != 1 || notes[0] != "updated text" {
		t.Errorf("notes after update = %v, want [updated text]", notes)
	}
}

// --- help / init ---

// TestHelp_ExitZero verifies `devlog --help` exits 0.
func TestHelp_ExitZero(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)
	if _, _, err := ws.Run("--help"); err != nil {
		t.Fatalf("devlog --help exited non-zero: %v", err)
	}
}

// TestInit_InstallsHook verifies `devlog init --non-interactive --add-repo <path>`
// creates an executable post-commit hook containing "devlog sync".
func TestInit_InstallsHook(t *testing.T) {
	t.Parallel()
	ws := testharness.New(t)

	// Create a git repo inside the workspace dir so devlog can discover it.
	repoDir := filepath.Join(ws.Dir, "myrepo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("mkdir repoDir: %v", err)
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", args[0], err, out)
		}
	}

	hookPath := filepath.Join(repoDir, ".git", "hooks", "post-commit")

	// installHook is the internal mechanism; call it directly via the exported
	// helper path by writing the hook ourselves to match what devlog init does.
	if err := installHookViaDevlog(ws, repoDir, hookPath); err != nil {
		t.Fatalf("hook installation: %v", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("reading post-commit hook: %v", err)
	}
	if !strings.Contains(string(data), "devlog sync") {
		t.Errorf("hook missing 'devlog sync':\n%s", data)
	}
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("post-commit hook is not executable")
	}
}

// installHookViaDevlog drives devlog init interactively to install the hook,
// using an isolated XDG_CONFIG_HOME so it does not touch the real user config.
func installHookViaDevlog(ws *testharness.Workspace, repoDir, hookPath string) error {
	xdgDir, err := os.MkdirTemp("", "devlog-xdg-*")
	if err != nil {
		return err
	}

	// Write a minimal config that points journal.dir at ws.Dir and registers repoDir.
	cfgDir := filepath.Join(xdgDir, "devlog")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return err
	}
	cfgContent := "[journal]\ndir = \"" + ws.Dir + "\"\n\n[[repos]]\npath = \"" + repoDir + "\"\nname = \"myrepo\"\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(cfgContent), 0644); err != nil {
		return err
	}

	// Build env: workspace env + isolated XDG_CONFIG_HOME.
	env := ws.Env()
	env = append(env, "XDG_CONFIG_HOME="+xdgDir)

	// Drive interactive init: answer journal dir prompt, skip GitHub token,
	// then answer "y" to install hook for myrepo.
	input := ws.Dir + "\n" + "\n" + "y\n"

	cmd := exec.Command(ws.Binary, "init")
	cmd.Dir = ws.Dir
	cmd.Env = env
	cmd.Stdin = strings.NewReader(input)
	if out, cmdErr := cmd.CombinedOutput(); cmdErr != nil {
		return fmt.Errorf("devlog init: %v\n%s", cmdErr, out)
	}
	return nil
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
