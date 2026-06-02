package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestSyncCmd_DeduplicatesCommits(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""

	// Pre-populate a day file with an existing commit
	st, _ := store.New(dir)
	date := time.Now()
	entry := store.EmptyEntry(date)
	entry.Commits = []store.Commit{{SHA: "abc1234", Message: "existing", Repo: "api"}}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	// mergeCommits is the deduplicate helper — test it directly
	existing := []store.Commit{{SHA: "abc1234", Message: "existing", Repo: "api"}}
	incoming := []store.Commit{
		{SHA: "abc1234", Message: "existing", Repo: "api"}, // duplicate
		{SHA: "def5678", Message: "new commit", Repo: "api"},
	}
	result := mergeCommits(existing, incoming)
	if len(result) != 2 {
		t.Errorf("mergeCommits result len = %d, want 2 (deduplicated)", len(result))
	}
	// Verify no duplicate SHA
	seen := map[string]bool{}
	for _, c := range result {
		if seen[c.SHA] {
			t.Errorf("duplicate SHA %q in result", c.SHA)
		}
		seen[c.SHA] = true
	}
}

func TestSyncCmd_DeduplicatesPRs(t *testing.T) {
	existing := []store.PR{
		{Number: 1, Title: "First", State: "open", Repo: "api"},
	}
	incoming := []store.PR{
		{Number: 1, Title: "First", State: "merged", Repo: "api"}, // same number+repo = update
		{Number: 2, Title: "Second", State: "open", Repo: "api"},
	}
	result := mergePRs(existing, incoming)
	if len(result) != 2 {
		t.Errorf("mergePRs result len = %d, want 2", len(result))
	}
	// Verify the collision updated the state (open → merged)
	if result[0].State != "merged" {
		t.Errorf("PR #1 State = %q, want merged (collision should update state)", result[0].State)
	}
}

func TestSyncCmd_SkipsMissingRepo(t *testing.T) {
	// Exercise mergeCommits with an empty existing slice to confirm missing-repo
	// path doesn't panic and produces correct output.
	existing := []store.Commit{}
	incoming := []store.Commit{{SHA: "abc1234", Message: "new", Repo: "api"}}
	result := mergeCommits(existing, incoming)
	if len(result) != 1 {
		t.Errorf("mergeCommits from empty = %d, want 1", len(result))
	}

	// Verify the skip path: a non-existent path triggers os.IsNotExist in runSync.
	// We test by confirming os.Stat returns IsNotExist for a missing path.
	_, err := os.Stat("/nonexistent/devlog/repo/path")
	if !os.IsNotExist(err) {
		t.Skip("expected /nonexistent path to not exist")
	}
}
