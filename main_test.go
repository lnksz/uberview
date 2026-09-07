package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchGitLabIssuesNormalizesStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/issues":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"id":101,"title":"Scheduled work","state":"opened","web_url":"https://gitlab.example/issues/1","created_at":"2026-01-02T12:00:00Z","updated_at":"2026-01-02T12:00:00Z","start_date":"2026-03-10","due_date":"2026-03-20"},
				{"id":102,"title":"Unscheduled work","state":"opened","web_url":"https://gitlab.example/issues/2","created_at":"2026-01-03T12:00:00Z","updated_at":"2026-01-03T12:00:00Z"}
			]`))
		case "/api/graphql":
			if r.Method != http.MethodPost {
				t.Errorf("GraphQL method = %s, want POST", r.Method)
			}
			var request struct {
				Query string `json:"query"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decoding GraphQL request: %v", err)
			}
			if !strings.Contains(request.Query, "gid://gitlab/WorkItem/101") {
				t.Errorf("GraphQL query does not request the GitLab work item: %s", request.Query)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"i0":{"widgets":[{"__typename":"WorkItemWidgetStatus","status":{"name":"In progress"}}]},"i1":{"widgets":[]}}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	app := &App{client: server.Client()}
	issues, err := app.fetchGitLabIssues(TaskProvider{
		Type: TaskProviderGitLab,
		URL:  server.URL,
		User: "me",
	})
	if err != nil {
		t.Fatalf("fetchGitLabIssues() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("fetchGitLabIssues() returned %d issues, want 2", len(issues))
	}

	wantStart, err := time.Parse("2006-01-02", "2026-03-10")
	if err != nil {
		t.Fatalf("parsing expected start date: %v", err)
	}
	if issues[0].StartAt == nil || !issues[0].StartAt.Equal(wantStart) {
		t.Fatalf("first issue StartAt = %v, want %v", issues[0].StartAt, wantStart)
	}
	if issues[1].StartAt != nil {
		t.Fatalf("second issue StartAt = %v, want nil", issues[1].StartAt)
	}
	if issues[0].Status != "In progress" {
		t.Errorf("first issue Status = %q, want work item status", issues[0].Status)
	}
	if issues[1].Status != "opened" {
		t.Errorf("second issue Status = %q, want state fallback", issues[1].Status)
	}
}

func TestFetchGitLabWorkItemStatusesKeepsCompletedBatches(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			_, _ = w.Write([]byte(`{"data":{"i0":{"widgets":[{"__typename":"WorkItemWidgetStatus","status":{"name":"In progress"}}]}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{},"errors":[{"message":"status unavailable"}]}`))
	}))
	defer server.Close()

	issues := make([]GitLabIssue, 26)
	for i := range issues {
		issues[i].ID = i + 1
	}
	app := &App{client: server.Client()}
	statuses, err := app.fetchGitLabWorkItemStatuses(TaskProvider{URL: server.URL}, issues)
	if err == nil {
		t.Fatal("fetchGitLabWorkItemStatuses() error = nil, want second-batch error")
	}
	if requestCount != 2 {
		t.Fatalf("GraphQL requests = %d, want 2", requestCount)
	}
	if statuses[1] != "In progress" {
		t.Fatalf("first-batch status = %q, want In progress", statuses[1])
	}
}

func TestFetchJiraCloudIssuesNormalizesStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.Query().Get("fields"), "customfield_10702") {
			t.Error("Jira Cloud request does not include customfield_10702")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"isLast": true,
			"issues": [
				{"key":"PLAN-1","fields":{"summary":"Scheduled work","created":"2026-01-02T12:00:00.000+0000","updated":"2026-01-02T12:00:00.000+0000","customfield_10702":"2026-03-10","duedate":"2026-03-20"}},
				{"key":"PLAN-2","fields":{"summary":"Unscheduled work","created":"2026-01-03T12:00:00.000+0000","updated":"2026-01-03T12:00:00.000+0000"}}
			]
		}`))
	}))
	defer server.Close()

	app := &App{client: server.Client()}
	issues, err := app.fetchJiraCloudIssues(TaskProvider{
		Type:  TaskProviderJiraCloud,
		URL:   server.URL,
		User:  "me",
		Email: "me@example.com",
		Token: "token",
	})
	if err != nil {
		t.Fatalf("fetchJiraCloudIssues() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("fetchJiraCloudIssues() returned %d issues, want 2", len(issues))
	}

	wantStart, err := time.Parse("2006-01-02", "2026-03-10")
	if err != nil {
		t.Fatalf("parsing expected start date: %v", err)
	}
	if issues[0].StartAt == nil || !issues[0].StartAt.Equal(wantStart) {
		t.Fatalf("first issue StartAt = %v, want %v", issues[0].StartAt, wantStart)
	}
	if issues[1].StartAt != nil {
		t.Fatalf("second issue StartAt = %v, want nil", issues[1].StartAt)
	}
}

func TestResolveProviderTokensFromFileAndProgram(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gitlab.token"), []byte("file-secret\n"), 0o600); err != nil {
		t.Fatalf("writing token file: %v", err)
	}
	config := Config{TaskProviders: []TaskProvider{
		{Name: "GitLab", TokenFile: "gitlab.token"},
		{Name: "Jira", TokenProg: "printf prog-secret\\r\\n"},
	}}

	if err := resolveProviderTokens(&config, dir); err != nil {
		t.Fatalf("resolveProviderTokens() error = %v", err)
	}
	if got := config.TaskProviders[0].Token; got != "file-secret" {
		t.Fatalf("file token = %q, want %q", got, "file-secret")
	}
	if got := config.TaskProviders[1].Token; got != "prog-secret" {
		t.Fatalf("program token = %q, want %q", got, "prog-secret")
	}
}

func TestResolveProviderTokensRejectsMultipleSources(t *testing.T) {
	config := Config{TaskProviders: []TaskProvider{{
		Name:      "GitLab",
		Token:     "inline",
		TokenFile: "secret.txt",
	}}}

	err := resolveProviderTokens(&config, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "only one") {
		t.Fatalf("resolveProviderTokens() error = %v, want source exclusivity error", err)
	}
}

func TestResolveProviderTokensRejectsMultilineSecret(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "token"), []byte("\nsecret\n"), 0o600); err != nil {
		t.Fatalf("writing token file: %v", err)
	}
	config := Config{TaskProviders: []TaskProvider{{Name: "GitLab", TokenFile: "token"}}}
	if err := resolveProviderTokens(&config, dir); err == nil || !strings.Contains(err.Error(), "multiple lines") {
		t.Fatalf("multiline token error = %v, want multiple lines error", err)
	}
}

func TestRunTokenProgramBoundsOutputAndRuntime(t *testing.T) {
	if _, err := runTokenProgram("head -c 9000 /dev/zero", t.TempDir(), time.Second); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized token-prog error = %v, want output limit error", err)
	}
	if _, err := runTokenProgram("sleep 1", t.TempDir(), 20*time.Millisecond); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("slow token-prog error = %v, want timeout error", err)
	}
}

func TestIssueCacheExpiresEntries(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "sqlite.db")
	cache, err := openIssueCache(cachePath, 15*time.Second)
	if err != nil {
		t.Fatalf("openIssueCache() error = %v", err)
	}
	defer cache.db.Close()
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("stat cache database: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("cache permissions = %o, want 600", got)
	}

	provider := TaskProvider{Type: TaskProviderGitLab, Name: "GitLab", URL: "https://gitlab.example", User: "me"}
	want := []Issue{{Source: provider.Name, Title: "Cached ticket", WebURL: "https://gitlab.example/issues/1"}}
	if err := cache.Put(provider, want); err != nil {
		t.Fatalf("IssueCache.Put() error = %v", err)
	}
	got, cached, err := cache.Get(provider)
	if err != nil {
		t.Fatalf("IssueCache.Get() error = %v", err)
	}
	if !cached || len(got) != 1 || got[0].Title != want[0].Title {
		t.Fatalf("IssueCache.Get() = (%v, %v), want cached ticket", got, cached)
	}

	_, err = cache.db.Exec(
		"UPDATE issue_cache SET fetched_at = ? WHERE provider_key = ?",
		time.Now().Add(-16*time.Second).UnixNano(), providerCacheKey(provider),
	)
	if err != nil {
		t.Fatalf("expiring cache row: %v", err)
	}
	if _, cached, err := cache.Get(provider); err != nil || cached {
		t.Fatalf("IssueCache.Get() after expiry = cached %v, error %v", cached, err)
	}
	_, err = cache.db.Exec(
		"UPDATE issue_cache SET fetched_at = ? WHERE provider_key = ?",
		time.Now().Add(time.Minute).UnixNano(), providerCacheKey(provider),
	)
	if err != nil {
		t.Fatalf("setting future cache timestamp: %v", err)
	}
	if _, cached, err := cache.Get(provider); err != nil || cached {
		t.Fatalf("IssueCache.Get() with future timestamp = cached %v, error %v", cached, err)
	}
}

func TestOpenIssueCacheRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatalf("writing symlink target: %v", err)
	}
	path := filepath.Join(dir, "sqlite.db")
	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("creating symlink: %v", err)
	}
	if _, err := openIssueCache(path, 15*time.Second); err == nil || !strings.Contains(err.Error(), "refusing symlink") {
		t.Fatalf("openIssueCache() error = %v, want symlink rejection", err)
	}
}

func TestFetchProviderIssuesUsesCache(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"title":"Cached ticket",
			"web_url":"https://gitlab.example/issues/1",
			"created_at":"2026-01-02T12:00:00Z",
			"updated_at":"2026-01-02T12:00:00Z"
		}]`))
	}))
	defer server.Close()

	cache, err := openIssueCache(filepath.Join(t.TempDir(), "sqlite.db"), 15*time.Second)
	if err != nil {
		t.Fatalf("openIssueCache() error = %v", err)
	}
	defer cache.db.Close()
	app := &App{client: server.Client(), cache: cache}
	provider := TaskProvider{Type: TaskProviderGitLab, Name: "GitLab", URL: server.URL, User: "me"}

	if _, cached, err := app.fetchProviderIssues(provider); err != nil || cached {
		t.Fatalf("first fetch = cached %v, error %v", cached, err)
	}
	if _, cached, err := app.fetchProviderIssues(provider); err != nil || !cached {
		t.Fatalf("second fetch = cached %v, error %v", cached, err)
	}
	if requests != 1 {
		t.Fatalf("provider requests = %d, want 1", requests)
	}
}

func TestFetchProviderIssuesCoalescesConcurrentMisses(t *testing.T) {
	var requests atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	cache, err := openIssueCache(filepath.Join(t.TempDir(), "sqlite.db"), 15*time.Second)
	if err != nil {
		t.Fatalf("openIssueCache() error = %v", err)
	}
	defer cache.db.Close()
	app := &App{client: server.Client(), cache: cache}
	provider := TaskProvider{Type: TaskProviderGitLab, Name: "GitLab", URL: server.URL, User: "me"}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, _, err := app.fetchProviderIssues(provider); err != nil {
				t.Errorf("fetchProviderIssues() error = %v", err)
			}
		}()
	}
	close(start)
	for requests.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	if got := requests.Load(); got != 1 {
		t.Fatalf("provider requests = %d, want 1", got)
	}
}

func TestHandleProviderIssuesMarksCachedReachabilityUnknown(t *testing.T) {
	cache, err := openIssueCache(filepath.Join(t.TempDir(), "sqlite.db"), 15*time.Second)
	if err != nil {
		t.Fatalf("openIssueCache() error = %v", err)
	}
	defer cache.db.Close()
	provider := TaskProvider{Type: TaskProviderGitLab, Name: "GitLab", URL: "https://gitlab.example", User: "me"}
	if err := cache.Put(provider, []Issue{{Source: provider.Name, Title: "Cached ticket"}}); err != nil {
		t.Fatalf("IssueCache.Put() error = %v", err)
	}
	app := &App{config: Config{TaskProviders: []TaskProvider{provider}}, cache: cache}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/provider/GitLab/issues", nil)

	app.handleProviderIssues(recorder, request)
	var response struct {
		Status ProviderStatus `json:"status"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if !response.Status.Cached || response.Status.Online {
		t.Fatalf("cached status = %+v, want cached with unknown reachability", response.Status)
	}
}
