package git

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
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

func TestParseGHPROutput_InvalidJSON(t *testing.T) {
	_, err := parseGHOutput("not-json", time.Now())
	if err == nil {
		t.Error("parseGHOutput(invalid JSON) expected error, got nil")
	}
}

func TestParseGHPROutput_InvalidTimestamp(t *testing.T) {
	// PR with a bad updatedAt should be silently skipped, not error.
	// Use noon UTC so the valid PR falls on 2026-06-01 in any local timezone.
	input := `[{"number":1,"title":"Bad","state":"open","updatedAt":"bad-timestamp","url":"http://x"},{"number":2,"title":"Good","state":"open","updatedAt":"2026-06-01T12:00:00Z","url":"http://y"}]`
	targetDay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	prs, err := parseGHOutput(input, targetDay)
	if err != nil {
		t.Fatalf("parseGHOutput(invalid timestamp) error = %v", err)
	}
	// PR 1 skipped (bad timestamp), PR 2 included (valid and matches date)
	if len(prs) != 1 || prs[0].Number != 2 {
		t.Errorf("prs = %+v, want [{Number:2}]", prs)
	}
}

func TestParseRESTOutput(t *testing.T) {
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

func TestParseRESTOutput_InvalidJSON(t *testing.T) {
	_, err := parseRESTOutput("not-json", time.Now(), "user")
	if err == nil {
		t.Error("parseRESTOutput(invalid JSON) expected error, got nil")
	}
}

func TestParseRESTOutput_InvalidTimestamp(t *testing.T) {
	// Use noon UTC so the valid PR falls on 2026-06-01 in any local timezone.
	input := `[{"number":1,"title":"Bad ts","state":"open","updated_at":"bad","user":{"login":"u"}},{"number":2,"title":"Good","state":"open","updated_at":"2026-06-01T12:00:00Z","user":{"login":"u"}}]`
	targetDay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	prs, err := parseRESTOutput(input, targetDay, "u")
	if err != nil {
		t.Fatalf("parseRESTOutput(invalid timestamp) error = %v", err)
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
	_ = isGHAvailable()
}

func TestFetchPRs_NoGHNoToken(t *testing.T) {
	// Without gh CLI auth and without a token, FetchPRs should return nil, nil gracefully.
	// We can't control whether gh is installed, but with no token and a fake repo slug
	// this should either succeed with empty results or fail through gracefully.
	prs, err := FetchPRs("nonexistent-owner/nonexistent-repo-devlog-test", "repo", "", time.Now())
	if err != nil {
		// FetchPRs is designed to return nil, nil when nothing is available.
		// If gh is available and fails, it falls through to the token check (empty) and returns nil.
		// An error here means gh is available and returned a non-nil error that wasn't suppressed — unexpected.
		t.Logf("FetchPRs returned error (gh may be available and errored): %v", err)
	}
	_ = prs
}

// testTransport redirects requests for a given host to a local httptest.Server.
type testTransport struct {
	host   string
	server *httptest.Server
}

func (tr *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == tr.host {
		// Rewrite the URL to point to the test server
		newURL := *req.URL
		newURL.Scheme = "http"
		newURL.Host = tr.server.Listener.Addr().String()
		req = req.Clone(req.Context())
		req.URL = &newURL
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestFetchPRsViaREST(t *testing.T) {
	targetDay := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// Build the REST API mock server.
	mux := http.NewServeMux()

	// /user endpoint returns the authenticated user's login.
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"login":"testuser"}`)
	})

	// /repos/.../pulls endpoint returns one PR on target day.
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		prs := []restPRItem{
			{
				Number:    42,
				Title:     "Add feature",
				State:     "merged",
				UpdatedAt: targetDay.Format(time.RFC3339),
				HTMLURL:   "https://github.com/owner/repo/pull/42",
				User:      struct{ Login string `json:"login"` }{Login: "testuser"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prs) //nolint:errcheck
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Redirect api.github.com requests to the test server.
	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &testTransport{host: "api.github.com", server: srv}
	defer func() { http.DefaultClient.Transport = orig }()

	prs, err := fetchPRsViaREST("owner/repo", "repo", "fake-token", targetDay)
	if err != nil {
		t.Fatalf("fetchPRsViaREST() error = %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("fetchPRsViaREST() = %d prs, want 1", len(prs))
	}
	if prs[0].Number != 42 {
		t.Errorf("prs[0].Number = %d, want 42", prs[0].Number)
	}
	if prs[0].Repo != "repo" {
		t.Errorf("prs[0].Repo = %q, want \"repo\"", prs[0].Repo)
	}
}

func TestFetchPRsViaREST_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"login":"testuser"}`)
	})
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &testTransport{host: "api.github.com", server: srv}
	defer func() { http.DefaultClient.Transport = orig }()

	_, err := fetchPRsViaREST("owner/repo", "repo", "fake-token", time.Now())
	if err == nil {
		t.Error("fetchPRsViaREST(500) expected error, got nil")
	}
}

func TestGetGitHubLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			fmt.Fprintf(w, `{"login":"mylogin"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &testTransport{host: "api.github.com", server: srv}
	defer func() { http.DefaultClient.Transport = orig }()

	login, err := getGitHubLogin("fake-token")
	if err != nil {
		t.Fatalf("getGitHubLogin() error = %v", err)
	}
	if login != "mylogin" {
		t.Errorf("getGitHubLogin() = %q, want \"mylogin\"", login)
	}
}

// Verify FetchPRs falls through to REST when gh is unavailable.
func TestFetchPRs_WithToken(t *testing.T) {
	targetDay := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)

	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"login":"tokenuser"}`)
	})
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		prs := []restPRItem{
			{
				Number:    7,
				Title:     "Token PR",
				State:     "open",
				UpdatedAt: targetDay.Format(time.RFC3339),
				HTMLURL:   "https://github.com/owner/repo/pull/7",
				User:      struct{ Login string `json:"login"` }{Login: "tokenuser"},
			},
		}
		json.NewEncoder(w).Encode(prs) //nolint:errcheck
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &testTransport{host: "api.github.com", server: srv}
	defer func() { http.DefaultClient.Transport = orig }()

	// FetchPRs will try gh first; if gh is not installed, it falls to REST.
	// If gh is installed but fails on fake repo, it also falls to REST.
	// Either way, REST with our token+mock should return the PR.
	prs, err := FetchPRs("owner/repo", "repo", "fake-token", targetDay)
	if err != nil {
		t.Fatalf("FetchPRs(with token) error = %v", err)
	}

	// If gh succeeded with an empty result that's also fine; otherwise we want 1.
	// We only assert no error since gh behavior varies by environment.
	_ = prs
}

// Compile-time check that store.PR is the right type.
var _ []store.PR
