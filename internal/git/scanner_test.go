package git

import (
	"testing"
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
