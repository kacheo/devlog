package store

import (
	"fmt"
	"sort"
	"strings"
)

// TagCount holds a tag name and the number of day files that contain it.
type TagCount struct {
	Tag   string
	Count int
}

// ListTags scans all day files and returns tags sorted by count descending, then alphabetically.
// Case-insensitive deduplication: the first-seen casing (oldest file) is used as the display name.
func (s *Store) ListTags() ([]TagCount, error) {
	dates, err := s.AllDates()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}    // lowercase key → count
	display := map[string]string{} // lowercase key → first-seen display casing

	for _, date := range dates {
		entry, err := s.Load(date)
		if err != nil || entry == nil {
			continue
		}
		// Deduplicate within a single day before counting.
		seen := map[string]bool{}
		for _, tag := range entry.Tags {
			key := strings.ToLower(tag)
			if seen[key] {
				continue
			}
			seen[key] = true
			counts[key]++
			if _, exists := display[key]; !exists {
				display[key] = tag
			}
		}
	}

	result := make([]TagCount, 0, len(counts))
	for key, count := range counts {
		result = append(result, TagCount{Tag: display[key], Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return strings.ToLower(result[i].Tag) < strings.ToLower(result[j].Tag)
	})
	return result, nil
}

// RenameTag renames oldTag to newTag (case-insensitive match) across all day files.
// Returns the number of files modified. Returns an error if oldTag and newTag are identical
// (case-insensitively).
func (s *Store) RenameTag(oldTag, newTag string) (int, error) {
	if strings.EqualFold(oldTag, newTag) {
		return 0, fmt.Errorf("old and new tag are the same: %q", oldTag)
	}

	dates, err := s.AllDates()
	if err != nil {
		return 0, err
	}

	modified := 0
	for _, date := range dates {
		changed := false
		err := s.Modify(date, func(entry *DayEntry) error {
			for i, tag := range entry.Tags {
				if strings.EqualFold(tag, oldTag) {
					entry.Tags[i] = newTag
					changed = true
				}
			}
			return nil
		})
		if err != nil {
			return modified, fmt.Errorf("updating %s: %w", date.Format("2006-01-02"), err)
		}
		if changed {
			modified++
		}
	}
	return modified, nil
}
