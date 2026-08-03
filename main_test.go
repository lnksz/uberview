package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchGitLabIssuesNormalizesStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"title":"Scheduled work","web_url":"https://gitlab.example/issues/1","created_at":"2026-01-02T12:00:00Z","updated_at":"2026-01-02T12:00:00Z","start_date":"2026-03-10","due_date":"2026-03-20"},
			{"title":"Unscheduled work","web_url":"https://gitlab.example/issues/2","created_at":"2026-01-03T12:00:00Z","updated_at":"2026-01-03T12:00:00Z"}
		]`))
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
