package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

// ScanCommits runs git log for repoPath on targetDate and returns commits authored
// by the local git user.email. Returns an empty slice if git is unavailable or the
// repo has no matching commits.
func ScanCommits(repoPath, repoName string, targetDate time.Time) ([]store.Commit, error) {
	// Get author email from git config
	emailOut, err := exec.Command("git", "-C", repoPath, "config", "user.email").Output()
	if err != nil {
		return nil, fmt.Errorf("getting git user.email for %s: %w", repoPath, err)
	}
	email := strings.TrimSpace(string(emailOut))

	since := targetDate.Format("2006-01-02") + " 00:00:00"
	until := targetDate.Format("2006-01-02") + " 23:59:59"

	out, err := exec.Command(
		"git", "-C", repoPath,
		"log",
		"--since="+since,
		"--until="+until,
		"--author="+email,
		"--pretty=format:%h %s",
	).Output()
	if err != nil {
		// git log exits 128 on a bad repo; treat as no commits
		return nil, fmt.Errorf("running git log in %s: %w", repoPath, err)
	}

	slug, _ := GetOriginSlug(repoPath)
	var commits []store.Commit
	for _, line := range strings.Split(string(bytes.TrimSpace(out)), "\n") {
		if line == "" {
			continue
		}
		sha, msg, ok := parseGitLogLine(line)
		if !ok {
			continue
		}
		commits = append(commits, store.Commit{
			SHA: sha, Message: msg, Repo: repoName, RepoSlug: slug,
		})
	}
	return commits, nil
}

// GetOriginSlug runs `git -C repoPath remote get-url origin` and parses the URL.
// Returns ("", false) if origin is missing or not a GitHub remote.
func GetOriginSlug(repoPath string) (string, bool) {
	out, err := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", false
	}
	return ParseOriginURL(strings.TrimSpace(string(out)))
}

// ParseOriginURL extracts the owner/repo slug from a GitHub remote URL.
// Supports both SSH (git@github.com:owner/repo.git) and HTTPS (https://github.com/owner/repo.git).
// Returns ("", false) for non-GitHub URLs.
func ParseOriginURL(rawURL string) (string, bool) {
	if rawURL == "" {
		return "", false
	}
	var path string
	switch {
	case strings.HasPrefix(rawURL, "git@github.com:"):
		path = strings.TrimPrefix(rawURL, "git@github.com:")
	case strings.HasPrefix(rawURL, "https://github.com/"):
		path = strings.TrimPrefix(rawURL, "https://github.com/")
	default:
		return "", false
	}
	path = strings.TrimSuffix(path, ".git")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0] + "/" + parts[1], true
}

// parseGitLogLine splits a `%h %s` git log line into SHA and subject.
func parseGitLogLine(line string) (sha, subject string, ok bool) {
	idx := strings.IndexByte(line, ' ')
	if idx < 0 {
		return "", "", false
	}
	sha = line[:idx]
	subject = line[idx+1:]
	if sha == "" || subject == "" {
		return "", "", false
	}
	return sha, subject, true
}
