package store

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Commit struct {
	SHA      string `yaml:"sha"      json:"sha"`
	Message  string `yaml:"message"  json:"message"`
	Repo     string `yaml:"repo"     json:"repo"`
	RepoSlug string `yaml:"repo_slug" json:"repoSlug"`
}

type PR struct {
	Number int    `yaml:"number" json:"number"`
	Title  string `yaml:"title"  json:"title"`
	State  string `yaml:"state"  json:"state"`
	Repo   string `yaml:"repo"   json:"repo"`
	URL    string `yaml:"url,omitempty" json:"url"`
}

type frontmatter struct {
	Date    string   `yaml:"date"`
	Tags    []string `yaml:"tags"`
	Commits []Commit `yaml:"commits"`
	PRs     []PR     `yaml:"prs"`
}

type DayEntry struct {
	Date    time.Time
	Tags    []string
	Commits []Commit
	PRs     []PR
	Sections map[string][]string
}

var KnownSections = []string{"notes", "commits", "prs", "blockers"}

func NormalizeSection(s string) (string, bool) {
	switch strings.ToLower(s) {
	case "note", "notes":
		return "notes", true
	case "blocker", "blockers":
		return "blockers", true
	case "commit", "commits":
		return "commits", true
	case "pr", "prs":
		return "prs", true
	}
	return "", false
}

func EmptyEntry(date time.Time) *DayEntry {
	sections := make(map[string][]string, len(KnownSections))
	for _, s := range KnownSections {
		sections[s] = []string{}
	}
	return &DayEntry{
		Date:     date,
		Tags:     []string{},
		Sections: sections,
	}
}

func Parse(content []byte) (*DayEntry, error) {
	entry := &DayEntry{
		Sections: make(map[string][]string, len(KnownSections)),
	}
	for _, s := range KnownSections {
		entry.Sections[s] = []string{}
	}

	body := content

	if bytes.HasPrefix(content, []byte("---\n")) {
		rest := content[4:]
		bodyOffset := -1
		idx := bytes.Index(rest, []byte("\n---\n"))
		if idx >= 0 {
			bodyOffset = idx + 5
		} else if bytes.HasSuffix(rest, []byte("\n---")) {
			idx = len(rest) - 4
			bodyOffset = len(rest)
		}
		if idx >= 0 {
			fmBytes := rest[:idx]
			body = rest[bodyOffset:]

			var fm frontmatter
			if err := yaml.Unmarshal(fmBytes, &fm); err != nil {
				return nil, fmt.Errorf("parsing frontmatter: %w", err)
			}
			if fm.Date != "" {
				t, err := time.ParseInLocation("2006-01-02", fm.Date, time.Local)
				if err != nil {
					return nil, fmt.Errorf("parsing date %q: %w", fm.Date, err)
				}
				entry.Date = t
			}
			entry.Tags = fm.Tags
			if entry.Tags == nil {
				entry.Tags = []string{}
			}
			entry.Commits = fm.Commits
			if entry.Commits == nil {
				entry.Commits = []Commit{}
			}
			entry.PRs = fm.PRs
			if entry.PRs == nil {
				entry.PRs = []PR{}
			}
		}
	}

	var currentSection string
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")
			canonical, _ := NormalizeSection(heading)
			currentSection = canonical
			continue
		}
		if strings.HasPrefix(line, "- ") && currentSection != "" {
			if _, ok := entry.Sections[currentSection]; ok {
				if currentSection == "notes" || currentSection == "blockers" {
					entry.Sections[currentSection] = append(
						entry.Sections[currentSection],
						strings.TrimPrefix(line, "- "),
					)
				}
			}
		}
	}

	return entry, nil
}

func Serialize(e *DayEntry) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString("---\n")
	fmt.Fprintf(&b, "date: %s\n", e.Date.Format("2006-01-02"))

	tags := e.Tags
	if len(tags) == 0 {
		b.WriteString("tags: []\n")
	} else {
		b.WriteString("tags: [")
		for i, t := range tags {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t)
		}
		b.WriteString("]\n")
	}

	commits := e.Commits
	if len(commits) == 0 {
		b.WriteString("commits: []\n")
	} else {
		b.WriteString("commits:\n")
		for _, c := range commits {
			fmt.Fprintf(&b, "  - {sha: %s, message: %q, repo: %s, repo_slug: %s}\n", c.SHA, c.Message, c.Repo, c.RepoSlug)
		}
	}

	prs := e.PRs
	if len(prs) == 0 {
		b.WriteString("prs: []\n")
	} else {
		b.WriteString("prs:\n")
		for _, pr := range prs {
			if pr.URL != "" {
				fmt.Fprintf(&b, "  - {number: %d, title: %q, state: %s, repo: %s, url: %s}\n",
					pr.Number, pr.Title, pr.State, pr.Repo, pr.URL)
			} else {
				fmt.Fprintf(&b, "  - {number: %d, title: %q, state: %s, repo: %s}\n",
					pr.Number, pr.Title, pr.State, pr.Repo)
			}
		}
	}

	b.WriteString("---\n")

	sectionTitles := map[string]string{
		"notes":    "Notes",
		"commits":  "Commits",
		"prs":      "PRs",
		"blockers": "Blockers",
	}

	for _, sec := range KnownSections {
		b.WriteString("\n## ")
		b.WriteString(sectionTitles[sec])
		b.WriteString("\n")
		switch sec {
		case "commits":
			for _, c := range e.Commits {
				fmt.Fprintf(&b, "- %s %s (%s)\n", c.SHA, c.Message, c.Repo)
			}
		case "prs":
			for _, pr := range e.PRs {
				fmt.Fprintf(&b, "- #%d %s [%s] (%s)\n", pr.Number, pr.Title, pr.State, pr.Repo)
			}
		default:
			for _, line := range e.Sections[sec] {
				fmt.Fprintf(&b, "- %s\n", line)
			}
		}
	}

	return b.Bytes(), nil
}