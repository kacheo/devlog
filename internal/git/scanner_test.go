package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestParseOriginURL(t *testing.T) {
	tests := []struct {
		input     string
		wantSlug  string
		wantFound bool
	}{
		{"git@github.com:owner/repo.git", "owner/repo", true},
		{"git@github.com:owner/repo", "owner/repo", true},
		{"https://github.com/owner/repo.git", "owner/repo", true},
		{"https://github.com/owner/repo", "owner/repo", true},
		{"https://gitlab.com/owner/repo.git", "", false},
		{"git@gitlab.com:owner/repo.git", "", false},
		{"", "", false},
		{"not-a-url", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			slug, found := ParseOriginURL(tt.input)
			if found != tt.wantFound {
				t.Errorf("ParseOriginURL(%q) found = %v, want %v", tt.input, found, tt.wantFound)
			}
			if slug != tt.wantSlug {
				t.Errorf("ParseOriginURL(%q) slug = %q, want %q", tt.input, slug, tt.wantSlug)
			}
		})
	}
}

func TestParseGitLogLine(t *testing.T) {
	tests := []struct {
		line    string
		wantSHA string
		wantMsg string
		wantOK  bool
	}{
		{"abc1234 fix: oauth token refresh loop", "abc1234", "fix: oauth token refresh loop", true},
		{"def5678 feat: add rate limiter to /v2/users", "def5678", "feat: add rate limiter to /v2/users", true},
		{"abc1234 single-word", "abc1234", "single-word", true},
		{"", "", "", false},
		{"nospace", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			sha, msg, ok := parseGitLogLine(tt.line)
			if ok != tt.wantOK {
				t.Errorf("parseGitLogLine(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
			}
			if sha != tt.wantSHA {
				t.Errorf("parseGitLogLine(%q) sha = %q, want %q", tt.line, sha, tt.wantSHA)
			}
			if msg != tt.wantMsg {
				t.Errorf("parseGitLogLine(%q) msg = %q, want %q", tt.line, msg, tt.wantMsg)
			}
		})
	}
}

// initTempRepo creates a minimal git repository in a temp dir with one commit
// authored by the given email, and returns the directory path.
func initTempRepo(t *testing.T, email string) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}
	run("git", "init")
	run("git", "config", "user.email", email)
	run("git", "config", "user.name", "Test User")
	// Disable commit signing for this test repo so environments with mandatory
	// GPG/SSH signing configuration don't break the commit.
	run("git", "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "init: first commit")
	return dir
}

func TestGetOriginSlug_NoOrigin(t *testing.T) {
	dir := initTempRepo(t, "test@example.com")
	slug, ok := GetOriginSlug(dir)
	if ok {
		t.Errorf("GetOriginSlug(no origin) found = true, want false")
	}
	if slug != "" {
		t.Errorf("GetOriginSlug(no origin) slug = %q, want empty", slug)
	}
}

func TestGetOriginSlug_BadPath(t *testing.T) {
	slug, ok := GetOriginSlug("/nonexistent-path-devlog-test")
	if ok {
		t.Errorf("GetOriginSlug(bad path) found = true, want false")
	}
	if slug != "" {
		t.Errorf("GetOriginSlug(bad path) slug = %q, want empty", slug)
	}
}

func TestGetOriginSlug_WithGitHubRemote(t *testing.T) {
	dir := initTempRepo(t, "test@example.com")
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/owner/testrepo.git")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %v\n%s", err, out)
	}
	slug, ok := GetOriginSlug(dir)
	if !ok {
		t.Fatal("GetOriginSlug(with GitHub remote) found = false, want true")
	}
	if slug != "owner/testrepo" {
		t.Errorf("GetOriginSlug slug = %q, want \"owner/testrepo\"", slug)
	}
}

func TestScanCommits_BadRepo(t *testing.T) {
	_, err := ScanCommits("/nonexistent-path-devlog-test", "repo", time.Now())
	if err == nil {
		t.Error("ScanCommits(bad repo) expected error, got nil")
	}
}

func TestScanCommits_NoCommitsOnDate(t *testing.T) {
	dir := initTempRepo(t, "test@example.com")
	ancient := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	commits, err := ScanCommits(dir, "test-repo", ancient)
	if err != nil {
		t.Fatalf("ScanCommits(no commits on date) error = %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("ScanCommits(no commits on date) = %d commits, want 0", len(commits))
	}
}

func TestScanCommits_WithCommits(t *testing.T) {
	const email = "scanner-test@example.com"
	dir := initTempRepo(t, email)

	commits, err := ScanCommits(dir, "test-repo", time.Now())
	if err != nil {
		t.Fatalf("ScanCommits(with commits) error = %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("ScanCommits(with commits) = %d commits, want 1", len(commits))
	}
	if commits[0].Repo != "test-repo" {
		t.Errorf("Commit.Repo = %q, want \"test-repo\"", commits[0].Repo)
	}
	if commits[0].SHA == "" {
		t.Error("Commit.SHA is empty")
	}
	if commits[0].Message == "" {
		t.Error("Commit.Message is empty")
	}
}

func TestScanCommits_WithGitHubRemote(t *testing.T) {
	const email = "scanner-gh-test@example.com"
	dir := initTempRepo(t, email)

	cmd := exec.Command("git", "remote", "add", "origin", "git@github.com:owner/myrepo.git")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %v\n%s", err, out)
	}

	commits, err := ScanCommits(dir, "myrepo", time.Now())
	if err != nil {
		t.Fatalf("ScanCommits error = %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("ScanCommits = %d commits, want 1", len(commits))
	}
	if commits[0].RepoSlug != "owner/myrepo" {
		t.Errorf("Commit.RepoSlug = %q, want \"owner/myrepo\"", commits[0].RepoSlug)
	}
}
