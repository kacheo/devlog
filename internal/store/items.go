package store

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Item is a global blocker or action item that persists until resolved.
type Item struct {
	ID           string     `yaml:"id"`
	Type         string     `yaml:"type"`                   // "blocker" or "action_item"
	Text         string     `yaml:"text"`
	Resolved     bool       `yaml:"resolved"`
	ResolvedAt   *time.Time `yaml:"resolved_at,omitempty"`  // set when Resolved becomes true
	Dependencies []string   `yaml:"dependencies"`           // UUIDs of other items
}

type itemFile struct {
	Items []Item `yaml:"items"`
}

// ItemsFilePath returns the path to the global items YAML file.
func (s *Store) ItemsFilePath() string {
	return s.dir + "/items.yaml"
}

// LoadAllItems reads items.yaml and returns all items. Returns an empty slice if the file is missing.
func (s *Store) LoadAllItems() ([]Item, error) {
	data, err := os.ReadFile(s.ItemsFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("reading items file: %w", err)
	}
	var items []Item
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parsing items file: %w", err)
	}
	if items == nil {
		items = []Item{}
	}
	return items, nil
}

// AddItem validates and atomically appends a new item to items.yaml.
// itemType must be "blocker" or "action_item". deps are 8-char prefixes or full UUIDs.
func (s *Store) AddItem(itemType, text string, deps []string) (*Item, error) {
	if itemType != "blocker" && itemType != "action_item" {
		return nil, fmt.Errorf("unknown item type %q: must be blocker or action_item", itemType)
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("item text cannot be empty")
	}
	var added *Item
	err := s.modifyItems(func(f *itemFile) error {
		// Resolve dep prefixes to full UUIDs.
		fullDeps, err := resolveDepIDs(f.Items, deps)
		if err != nil {
			return err
		}
		// Type constraint: blockers may only depend on other blockers.
		if itemType == "blocker" {
			for _, depID := range fullDeps {
				dep := findItemByID(f.Items, depID)
				if dep != nil && dep.Type != "blocker" {
					return fmt.Errorf("blocker cannot depend on %s %s (blockers are external constraints)", dep.Type, depID[:8])
				}
			}
		}
		// DAG check: build adjacency map and verify no cycle would be introduced.
		newID := newUUID()
		if err := checkNoCycle(f.Items, newID, fullDeps); err != nil {
			return err
		}
		item := Item{
			ID:           newID,
			Type:         itemType,
			Text:         text,
			Resolved:     false,
			Dependencies: fullDeps,
		}
		if item.Dependencies == nil {
			item.Dependencies = []string{}
		}
		f.Items = append(f.Items, item)
		added = &f.Items[len(f.Items)-1]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return added, nil
}

// ResolveItem marks the item matching id (8-char prefix or full UUID) as resolved,
// recording the resolution timestamp.
func (s *Store) ResolveItem(id string) (*Item, error) {
	var resolved *Item
	err := s.modifyItems(func(f *itemFile) error {
		var matches []*Item
		for i := range f.Items {
			if matchesID(f.Items[i].ID, id) {
				matches = append(matches, &f.Items[i])
			}
		}
		switch len(matches) {
		case 0:
			return fmt.Errorf("no item found matching id %q", id)
		case 1:
			now := time.Now()
			matches[0].Resolved = true
			matches[0].ResolvedAt = &now
			resolved = matches[0]
			return nil
		default:
			return fmt.Errorf("ambiguous id %q: matches %d items", id, len(matches))
		}
	})
	if err != nil {
		return nil, err
	}
	return resolved, nil
}

// modifyItems atomically loads, mutates, and saves items.yaml under an exclusive lock.
// The file is stored as a top-level YAML sequence of Item objects.
func (s *Store) modifyItems(fn func(*itemFile) error) error {
	path := s.ItemsFilePath()
	return s.withLock(path, func() error {
		var f itemFile
		data, readErr := os.ReadFile(path)
		if readErr != nil && !os.IsNotExist(readErr) {
			return readErr
		}
		if readErr == nil {
			if err := yaml.Unmarshal(data, &f.Items); err != nil {
				return fmt.Errorf("parsing items file: %w", err)
			}
		}
		if f.Items == nil {
			f.Items = []Item{}
		}
		if err := fn(&f); err != nil {
			return err
		}
		out, err := yaml.Marshal(f.Items)
		if err != nil {
			return fmt.Errorf("serializing items: %w", err)
		}
		return s.saveRaw(path, out)
	})
}

// newUUID generates a random UUID v4.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// matchesID returns true if candidate equals id exactly or starts with it (8-char prefix).
func matchesID(candidate, id string) bool {
	if id == "" {
		return false
	}
	return candidate == id || (len(id) <= len(candidate) && candidate[:len(id)] == id)
}

// resolveDepIDs resolves a list of dep prefixes/full UUIDs to canonical full UUIDs.
func resolveDepIDs(items []Item, prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		return []string{}, nil
	}
	out := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		var matches []Item
		for _, item := range items {
			if matchesID(item.ID, p) {
				matches = append(matches, item)
			}
		}
		switch len(matches) {
		case 0:
			return nil, fmt.Errorf("no item found matching id %q", p)
		case 1:
			out = append(out, matches[0].ID)
		default:
			return nil, fmt.Errorf("ambiguous id %q: matches %d items", p, len(matches))
		}
	}
	return out, nil
}

func findItemByID(items []Item, id string) *Item {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}

// checkNoCycle verifies that adding newID → deps does not introduce a cycle.
// Builds a graph of existing items plus the hypothetical new node, then runs DFS from newID.
// Since newID is freshly generated no existing item points to it, making this a no-op for
// AddItem. Kept as a correct guard if deps ever become mutable.
func checkNoCycle(items []Item, newID string, deps []string) error {
	adj := make(map[string][]string, len(items)+1)
	for _, item := range items {
		adj[item.ID] = item.Dependencies
	}
	adj[newID] = deps

	visited := make(map[string]bool)
	var dfs func(cur string) bool
	dfs = func(cur string) bool {
		if visited[cur] {
			return false
		}
		visited[cur] = true
		for _, next := range adj[cur] {
			if next == newID || dfs(next) {
				return true
			}
		}
		return false
	}
	if dfs(newID) {
		return fmt.Errorf("dependencies would create a cycle")
	}
	return nil
}

// ShortID returns the first 8 characters of a UUID.
func ShortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

// FilterUnresolved returns only items where Resolved is false.
func FilterUnresolved(items []Item) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if !it.Resolved {
			out = append(out, it)
		}
	}
	return out
}

// FilterResolved returns only items where Resolved is true, optionally bounded by
// from/until applied to ResolvedAt. Zero-value from/until means unbounded.
func FilterResolved(items []Item, from, until time.Time) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if !it.Resolved {
			continue
		}
		if !from.IsZero() && it.ResolvedAt != nil && it.ResolvedAt.Before(from) {
			continue
		}
		if !until.IsZero() && it.ResolvedAt != nil && it.ResolvedAt.After(until) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// SplitByType separates items into blockers and action items.
func SplitByType(items []Item) (blockers []Item, actionItems []Item) {
	for _, it := range items {
		switch it.Type {
		case "blocker":
			blockers = append(blockers, it)
		case "action_item":
			actionItems = append(actionItems, it)
		}
	}
	if blockers == nil {
		blockers = []Item{}
	}
	if actionItems == nil {
		actionItems = []Item{}
	}
	return
}

// NormalizeItemType maps section names to canonical item types.
func NormalizeItemType(section string) (string, bool) {
	switch strings.ToLower(section) {
	case "blocker", "blockers":
		return "blocker", true
	case "action_item", "action_items":
		return "action_item", true
	}
	return "", false
}
