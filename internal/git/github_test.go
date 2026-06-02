package git

import (
	"testing"
	"time"
)

func TestParseGHPROutput(t *testing.T) {
	input := `[{"number":142,"title":"Add rate limiter","state":"merged","updatedAt":"2026-06-01T14:00:00Z","url":"https://github.com/kacheo/api-server/pull/142"},{"number":143,"title":"Fix auth","state":"open","updatedAt":"2026-06-02T09:00:00Z","url":"https://github.com/kacheo/api-server/pull/143"}]`
	targetDay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)

	prs, err := parseGHOutput(input, targetDay)
	if err != nil {
		t.Fatalf("parseGHOutput() error = %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("len(prs) = %d, want 1 (only 2026-06-01)", len(prs))
	}
	if prs[0].Number != 142 {
		t.Errorf("prs[0].Number = %d, want 142", prs[0].Number)
	}
	if prs[0].Title != "Add rate limiter" {
		t.Errorf("prs[0].Title = %q", prs[0].Title)
	}
	if prs[0].State != "merged" {
		t.Errorf("prs[0].State = %q", prs[0].State)
	}
	if prs[0].URL != "https://github.com/kacheo/api-server/pull/142" {
		t.Errorf("prs[0].URL = %q, want https://github.com/kacheo/api-server/pull/142", prs[0].URL)
	}
}

func TestParseGHPROutput_Empty(t *testing.T) {
	prs, err := parseGHOutput("[]", time.Now())
	if err != nil {
		t.Fatalf("parseGHOutput() error = %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 prs, got %d", len(prs))
	}
}

func TestParseRESTOutput(t *testing.T) {
	// REST API format uses updated_at, user.login, and html_url
	input := `[{"number":142,"title":"Add rate limiter","state":"open","updated_at":"2026-06-01T14:00:00Z","html_url":"https://github.com/kacheo/api-server/pull/142","user":{"login":"testuser"}},{"number":143,"title":"Fix auth","state":"closed","updated_at":"2026-06-02T09:00:00Z","html_url":"https://github.com/kacheo/api-server/pull/143","user":{"login":"testuser"}}]`
	targetDay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	authorLogin := "testuser"

	prs, err := parseRESTOutput(input, targetDay, authorLogin)
	if err != nil {
		t.Fatalf("parseRESTOutput() error = %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("len(prs) = %d, want 1", len(prs))
	}
	if prs[0].Number != 142 {
		t.Errorf("prs[0].Number = %d, want 142", prs[0].Number)
	}
	if prs[0].URL != "https://github.com/kacheo/api-server/pull/142" {
		t.Errorf("prs[0].URL = %q, want https://github.com/kacheo/api-server/pull/142", prs[0].URL)
	}
}

func TestParseRESTOutput_FiltersByAuthor(t *testing.T) {
	input := `[{"number":1,"title":"PR by other","state":"open","updated_at":"2026-06-01T14:00:00Z","user":{"login":"otheruser"}},{"number":2,"title":"My PR","state":"open","updated_at":"2026-06-01T15:00:00Z","user":{"login":"testuser"}}]`
	targetDay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)

	prs, err := parseRESTOutput(input, targetDay, "testuser")
	if err != nil {
		t.Fatalf("parseRESTOutput() error = %v", err)
	}
	if len(prs) != 1 || prs[0].Number != 2 {
		t.Errorf("prs = %+v, want [{Number:2}]", prs)
	}
}

func TestSameDay(t *testing.T) {
	local := time.Local
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, local)
	tests := []struct {
		a, b time.Time
		want bool
	}{
		{time.Date(2026, 6, 1, 0, 0, 0, 0, local), base, true},
		{time.Date(2026, 6, 1, 23, 59, 59, 0, local), base, true},
		{time.Date(2026, 6, 2, 0, 0, 0, 0, local), base, false},
		{time.Date(2026, 5, 31, 23, 59, 59, 0, local), base, false},
	}
	for _, tt := range tests {
		got := sameDay(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("sameDay(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsGHAvailable(t *testing.T) {
	// Just verify the function doesn't panic — result depends on host environment
	_ = isGHAvailable()
}
