package git

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

// FetchPRs returns PRs updated on targetDate for the given owner/repo.
// It tries gh CLI first, then REST API with token, then returns nil (no error).
// repoName is the human-readable label used in PR.Repo.
func FetchPRs(ownerRepo, repoName, githubToken string, targetDate time.Time) ([]store.PR, error) {
	if isGHAvailable() {
		prs, err := fetchPRsViaGH(ownerRepo, repoName, targetDate)
		if err == nil {
			return prs, nil
		}
		// gh failed — fall through to REST or skip
	}
	if githubToken != "" {
		prs, err := fetchPRsViaREST(ownerRepo, repoName, githubToken, targetDate)
		if err == nil {
			return prs, nil
		}
	}
	return nil, nil
}

// isGHAvailable returns true if `gh` is in PATH.
func isGHAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// fetchPRsViaGH uses the `gh` CLI to fetch PRs.
func fetchPRsViaGH(ownerRepo, repoName string, targetDate time.Time) ([]store.PR, error) {
	out, err := exec.Command(
		"gh", "pr", "list",
		"--repo", ownerRepo,
		"--author=@me",
		"--state=all",
		"--json", "number,title,state,updatedAt",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	prs, err := parseGHOutput(string(out), targetDate)
	if err != nil {
		return nil, err
	}
	// Stamp the repo name
	for i := range prs {
		prs[i].Repo = repoName
	}
	return prs, nil
}

// ghPRItem matches the gh CLI JSON output for a PR.
type ghPRItem struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	UpdatedAt string `json:"updatedAt"`
}

// parseGHOutput parses `gh pr list --json` output and filters to targetDate.
func parseGHOutput(raw string, targetDate time.Time) ([]store.PR, error) {
	var items []ghPRItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("parsing gh output: %w", err)
	}
	var prs []store.PR
	for _, item := range items {
		t, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			continue
		}
		if sameDay(t.In(time.Local), targetDate) {
			prs = append(prs, store.PR{
				Number: item.Number,
				Title:  item.Title,
				State:  item.State,
			})
		}
	}
	return prs, nil
}

// fetchPRsViaREST uses the GitHub REST API to fetch PRs.
func fetchPRsViaREST(ownerRepo, repoName, token string, targetDate time.Time) ([]store.PR, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=all&per_page=100", ownerRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub REST API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub REST API status %d", resp.StatusCode)
	}

	// Get the authenticated user's login (to filter by author)
	login, err := getGitHubLogin(token)
	if err != nil {
		return nil, err
	}

	prs, err := parseRESTOutput(string(body), targetDate, login)
	if err != nil {
		return nil, err
	}
	for i := range prs {
		prs[i].Repo = repoName
	}
	return prs, nil
}

// getGitHubLogin returns the authenticated GitHub user's login via REST API.
func getGitHubLogin(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var u struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return "", err
	}
	return u.Login, nil
}

// restPRItem matches the GitHub REST API pull request shape.
type restPRItem struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// parseRESTOutput parses GitHub REST API PR list JSON and filters by author and date.
func parseRESTOutput(raw string, targetDate time.Time, authorLogin string) ([]store.PR, error) {
	var items []restPRItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("parsing REST output: %w", err)
	}
	var prs []store.PR
	for _, item := range items {
		if item.User.Login != authorLogin {
			continue
		}
		t, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			continue
		}
		if sameDay(t.In(time.Local), targetDate) {
			prs = append(prs, store.PR{
				Number: item.Number,
				Title:  item.Title,
				State:  item.State,
			})
		}
	}
	return prs, nil
}

// sameDay returns true if a and b fall on the same calendar day in local time.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.In(time.Local).Date()
	by, bm, bd := b.In(time.Local).Date()
	return ay == by && am == bm && ad == bd
}

// getUserEmailFromGit is a helper used by callers to get the local git user email.
func getUserEmailFromGit(repoPath string) (string, error) {
	out, err := exec.Command("git", "-C", repoPath, "config", "user.email").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
