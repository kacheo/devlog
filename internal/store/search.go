package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SearchableSections lists all valid section names accepted by SearchOpts.Sections.
var SearchableSections = append(append([]string{}, KnownSections...), "commits", "prs")

// SearchOpts controls which day files and sections are scanned.
type SearchOpts struct {
	Query    string
	Sections []string  // nil/empty = all sections
	From     time.Time // zero = no lower bound
	Until    time.Time // zero = no upper bound
}

// SearchResult is a single matched line from a day file.
type SearchResult struct {
	Date    string `json:"date"`
	Section string `json:"section"`
	Line    string `json:"line"`
}

// Search scans all YYYY-MM-DD.md files in the store directory and returns lines
// containing opts.Query (case-insensitive). opts.Sections restricts which sections
// are searched; an unknown section name returns an error.
func (s *Store) Search(opts SearchOpts) ([]SearchResult, error) {
	if err := validateSearchSections(opts.Sections); err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	query := strings.ToLower(opts.Query)
	var results []SearchResult

	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		name := de.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		dateStr := strings.TrimSuffix(name, ".md")
		t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			continue
		}
		if !opts.From.IsZero() && t.Before(opts.From) {
			continue
		}
		if !opts.Until.IsZero() && t.After(opts.Until) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dir, name))
		if err != nil {
			continue
		}
		entry, err := Parse(data)
		if err != nil {
			continue
		}

		dateLabel := t.Format("2006-01-02")
		results = appendSectionMatches(results, entry, query, opts.Sections, dateLabel)
	}

	return results, nil
}

func appendSectionMatches(results []SearchResult, entry *DayEntry, query string, sections []string, dateLabel string) []SearchResult {
	searchAll := len(sections) == 0

	for _, sec := range KnownSections {
		if !searchAll && !containsSection(sections, sec) {
			continue
		}
		for _, line := range entry.Sections[sec] {
			if strings.Contains(strings.ToLower(line), query) {
				results = append(results, SearchResult{Date: dateLabel, Section: sec, Line: line})
			}
		}
	}

	if searchAll || containsSection(sections, "commits") {
		for _, c := range entry.Commits {
			if strings.Contains(strings.ToLower(c.Message), query) {
				results = append(results, SearchResult{Date: dateLabel, Section: "commits", Line: c.Message})
			}
		}
	}

	if searchAll || containsSection(sections, "prs") {
		for _, pr := range entry.PRs {
			if strings.Contains(strings.ToLower(pr.Title), query) {
				results = append(results, SearchResult{Date: dateLabel, Section: "prs", Line: pr.Title})
			}
		}
	}

	return results
}

func validateSearchSections(sections []string) error {
	for _, s := range sections {
		if !containsSection(SearchableSections, s) {
			return fmt.Errorf("unknown section %q: valid sections are %s",
				s, strings.Join(SearchableSections, ", "))
		}
	}
	return nil
}

func containsSection(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
