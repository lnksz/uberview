package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
