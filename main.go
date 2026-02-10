package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed index.html
var staticFiles embed.FS

//go:embed favicon.ico
var faviconData []byte

// TaskProviderType represents the type of task provider
type TaskProviderType string

const (
	TaskProviderGitLab     TaskProviderType = "gitlab"
	TaskProviderJiraCloud  TaskProviderType = "jira_cloud"
	TaskProviderJiraServer TaskProviderType = "jira_server"
)

// TaskProvider represents a task provider configuration
type TaskProvider struct {
	Type  TaskProviderType `yaml:"type"`
	Name  string           `yaml:"name"`
	URL   string           `yaml:"url"`
	Token string           `yaml:"token"`
	User  string           `yaml:"user"`
	// Jira specific fields
	Email    string `yaml:"email"`    // Required for Jira Cloud API authentication
	Password string `yaml:"password"` // Required for Jira Server API authentication (or use Token for PAT)
}

// Config represents the YAML configuration structure
type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	TaskProviders []TaskProvider `yaml:"task_providers"`
}

// GitLabIssue represents a GitLab issue from the API
type GitLabIssue struct {
	ID          int       `json:"id"`
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	WebURL      string    `json:"web_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DueDate     *string   `json:"due_date"`
	Labels      []string  `json:"labels"`
	Type        string    `json:"type"`
	IssueType   string    `json:"issue_type"`
	Project     struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		PathWithNS string `json:"path_with_namespace"`
	} `json:"project"`
	Author struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"author"`
}

// JiraSearchResponse represents the Jira Cloud search API response
type JiraSearchResponse struct {
	Total         int         `json:"total"`
	Issues        []JiraIssue `json:"issues"`
	IsLast        bool        `json:"isLast"`
	NextPageToken string      `json:"nextPageToken"`
}

// JiraServerSearchResponse represents the Jira Server search API response (v2)
type JiraServerSearchResponse struct {
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []JiraIssue `json:"issues"`
}

// JiraIssue represents a Jira issue from the API
type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary   string `json:"summary"`
		Created   string `json:"created"`
		Updated   string `json:"updated"`
		DueDate   string `json:"duedate"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Project struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"project"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
	} `json:"fields"`
}

// Issue represents a generic issue from any task provider
type Issue struct {
	Source    string     `json:"source"`
	Title     string     `json:"title"`
	WebURL    string     `json:"web_url"`
	Status    string     `json:"status,omitempty"`
	Priority  string     `json:"priority,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DueAt     *time.Time `json:"due_at,omitempty"`
}

func sortIssuesDefault(issues []Issue) {
	// Default order: due date ascending; issues without a due date last.
	sort.SliceStable(issues, func(i, j int) bool {
		di, dj := issues[i].DueAt, issues[j].DueAt
		if di == nil && dj == nil {
			return issues[i].CreatedAt.Before(issues[j].CreatedAt)
		}
		if di == nil {
			return false
		}
		if dj == nil {
			return true
		}
		if di.Equal(*dj) {
			return issues[i].CreatedAt.Before(issues[j].CreatedAt)
		}
		return di.Before(*dj)
	})
}

// ProviderStatus represents the task provider server status
type ProviderStatus struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	Online     bool   `json:"online"`
	IssueCount int    `json:"issue_count"`
	Error      string `json:"error,omitempty"`
}

// App holds the application state
type App struct {
	config Config
	client *http.Client
}

func loadConfig(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("reading config file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("parsing config file: %w", err)
	}

	// Set default port if not specified
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	return config, nil
}

// checkProviderStatus checks if a single task provider server is reachable
func (a *App) checkProviderStatus(provider TaskProvider) ProviderStatus {
	status := ProviderStatus{
		Name:       provider.Name,
		URL:        provider.URL,
		Type:       string(provider.Type),
		Online:     false,
		IssueCount: 0,
	}

	switch provider.Type {
	case TaskProviderGitLab:
		apiURL := fmt.Sprintf("%s/api/v4/version", provider.URL)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			status.Error = fmt.Sprintf("creating request: %v", err)
			return status
		}

		req.Header.Set("PRIVATE-TOKEN", provider.Token)

		resp, err := a.client.Do(req)
		if err != nil {
			status.Error = fmt.Sprintf("connection failed: %v", err)
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			status.Online = true
		} else {
			status.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}

	case TaskProviderJiraCloud:
		apiURL := fmt.Sprintf("%s/rest/api/3/myself", provider.URL)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			status.Error = fmt.Sprintf("creating request: %v", err)
			return status
		}

		// Jira Cloud uses Basic Auth with email:api_token
		auth := base64.StdEncoding.EncodeToString([]byte(provider.Email + ":" + provider.Token))
		req.Header.Set("Authorization", "Basic "+auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			status.Error = fmt.Sprintf("connection failed: %v", err)
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			status.Online = true
		} else {
			status.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}

	case TaskProviderJiraServer:
		apiURL := fmt.Sprintf("%s/rest/api/2/myself", provider.URL)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			status.Error = fmt.Sprintf("creating request: %v", err)
			return status
		}

		// Jira Server: use Bearer token for PAT, or Basic Auth for username:password
		if provider.Token != "" {
			req.Header.Set("Authorization", "Bearer "+provider.Token)
		} else {
			auth := base64.StdEncoding.EncodeToString([]byte(provider.User + ":" + provider.Password))
			req.Header.Set("Authorization", "Basic "+auth)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			status.Error = fmt.Sprintf("connection failed: %v", err)
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			status.Online = true
		} else {
			status.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}

	default:
		status.Error = fmt.Sprintf("unsupported provider type: %s", provider.Type)
	}

	return status
}

// checkAllProvidersStatus checks status of all providers
func (a *App) checkAllProvidersStatus() []ProviderStatus {
	statuses := make([]ProviderStatus, len(a.config.TaskProviders))
	var wg sync.WaitGroup

	for i, provider := range a.config.TaskProviders {
		wg.Add(1)
		go func(idx int, p TaskProvider) {
			defer wg.Done()
			statuses[idx] = a.checkProviderStatus(p)
		}(i, provider)
	}

	wg.Wait()
	return statuses
}

// fetchAllIssues fetches issues from all configured providers
func (a *App) fetchAllIssues() ([]Issue, []ProviderStatus, error) {
	var allIssues []Issue
	var mu sync.Mutex
	var wg sync.WaitGroup
	statuses := make([]ProviderStatus, len(a.config.TaskProviders))
	errors := make([]error, len(a.config.TaskProviders))

	for i, provider := range a.config.TaskProviders {
		wg.Add(1)
		go func(idx int, p TaskProvider) {
			defer wg.Done()

			var issues []Issue
			var err error

			switch p.Type {
			case TaskProviderGitLab:
				issues, err = a.fetchGitLabIssues(p)
			case TaskProviderJiraCloud:
				issues, err = a.fetchJiraCloudIssues(p)
			case TaskProviderJiraServer:
				issues, err = a.fetchJiraServerIssues(p)
			default:
				err = fmt.Errorf("unsupported task provider type: %s", p.Type)
			}

			status := ProviderStatus{
				Name:       p.Name,
				URL:        p.URL,
				Type:       string(p.Type),
				Online:     err == nil,
				IssueCount: len(issues),
			}

			mu.Lock()
			statuses[idx] = status
			errors[idx] = err
			if err == nil {
				allIssues = append(allIssues, issues...)
			}
			mu.Unlock()
		}(i, provider)
	}

	wg.Wait()

	// Collect any errors
	var errMessages []string
	for i, err := range errors {
		if err != nil {
			log.Printf("ERROR: Failed to fetch issues from provider %q (%s): %v", a.config.TaskProviders[i].Name, a.config.TaskProviders[i].URL, err)
			errMessages = append(errMessages, fmt.Sprintf("%s: %v", a.config.TaskProviders[i].Name, err))
		} else {
			log.Printf("INFO: Fetched %d issues from provider %q", statuses[i].IssueCount, a.config.TaskProviders[i].Name)
		}
	}

	if len(errMessages) > 0 && len(allIssues) == 0 {
		return nil, statuses, fmt.Errorf("all providers failed: %v", errMessages)
	}

	return allIssues, statuses, nil
}

// fetchGitLabIssues fetches all open issues from GitLab assigned to the configured user
func (a *App) fetchGitLabIssues(provider TaskProvider) ([]Issue, error) {
	var allIssues []Issue
	page := 1
	perPage := 100

	for {
		params := url.Values{}
		params.Set("assignee_username", provider.User)
		params.Set("state", "opened")
		params.Set("scope", "all")
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))

		apiURL := fmt.Sprintf("%s/api/v4/issues?%s", provider.URL, params.Encode())
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", provider.Token)

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching issues: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
		}

		var gitlabIssues []GitLabIssue
		if err := json.NewDecoder(resp.Body).Decode(&gitlabIssues); err != nil {
			return nil, fmt.Errorf("decoding issues: %w", err)
		}

		for _, gi := range gitlabIssues {
			var dueAt *time.Time
			if gi.DueDate != nil && *gi.DueDate != "" {
				if t, err := time.Parse("2006-01-02", *gi.DueDate); err == nil {
					dueAt = &t
				}
			}

			issue := Issue{
				Source:    provider.Name,
				Title:     gi.Title,
				WebURL:    gi.WebURL,
				Status:    gi.State,
				CreatedAt: gi.CreatedAt,
				UpdatedAt: gi.UpdatedAt,
				DueAt:     dueAt,
			}
			allIssues = append(allIssues, issue)
		}

		if len(gitlabIssues) < perPage {
			break
		}
		page++
	}

	return allIssues, nil
}

// fetchJiraCloudIssues fetches all open issues from Jira Cloud assigned to the configured user
func (a *App) fetchJiraCloudIssues(provider TaskProvider) ([]Issue, error) {
	var allIssues []Issue
	nextPageToken := ""
	maxResults := 100

	for {
		// JQL to find open issues assigned to the user (excluding Done status category)
		jql := fmt.Sprintf("assignee = \"%s\" AND statusCategory != Done ORDER BY created ASC", provider.User)

		params := url.Values{}
		params.Set("jql", jql)
		params.Set("maxResults", fmt.Sprintf("%d", maxResults))
		params.Set("fields", "summary,created,updated,duedate,issuetype,status,project,priority")
		if nextPageToken != "" {
			params.Set("nextPageToken", nextPageToken)
		}

		// Use the new /search/jql endpoint (the old /search endpoint is deprecated and returns 410)
		apiURL := fmt.Sprintf("%s/rest/api/3/search/jql?%s", provider.URL, params.Encode())
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Jira Cloud uses Basic Auth with email:api_token
		auth := base64.StdEncoding.EncodeToString([]byte(provider.Email + ":" + provider.Token))
		req.Header.Set("Authorization", "Basic "+auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching issues: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Jira API returned status %d", resp.StatusCode)
		}

		var searchResp JiraSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return nil, fmt.Errorf("decoding issues: %w", err)
		}

		for _, ji := range searchResp.Issues {
			// Parse Jira date format (2024-01-15T10:30:00.000+0000)
			createdAt, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Created)
			updatedAt, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Updated)

			var dueAt *time.Time
			if ji.Fields.DueDate != "" {
				if t, err := time.Parse("2006-01-02", ji.Fields.DueDate); err == nil {
					dueAt = &t
				}
			}

			// Construct the web URL for the issue
			webURL := fmt.Sprintf("%s/browse/%s", provider.URL, ji.Key)

			issue := Issue{
				Source:    provider.Name,
				Title:     ji.Fields.Summary,
				WebURL:    webURL,
				Status:    ji.Fields.Status.Name,
				Priority:  ji.Fields.Priority.Name,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				DueAt:     dueAt,
			}
			allIssues = append(allIssues, issue)
		}

		// Check if there are more pages using the new pagination style
		if searchResp.IsLast || searchResp.NextPageToken == "" {
			break
		}
		nextPageToken = searchResp.NextPageToken
	}

	return allIssues, nil
}

// fetchJiraServerIssues fetches all open issues from Jira Server assigned to the configured user
func (a *App) fetchJiraServerIssues(provider TaskProvider) ([]Issue, error) {
	var allIssues []Issue
	startAt := 0
	maxResults := 100

	for {
		// JQL to find open issues assigned to the user (excluding Done status category)
		jql := fmt.Sprintf("assignee = \"%s\" AND statusCategory != Done ORDER BY created ASC", provider.User)

		params := url.Values{}
		params.Set("jql", jql)
		params.Set("startAt", fmt.Sprintf("%d", startAt))
		params.Set("maxResults", fmt.Sprintf("%d", maxResults))
		params.Set("fields", "summary,created,updated,duedate,issuetype,status,project,priority")

		// Jira Server 8.20 uses API v2
		apiURL := fmt.Sprintf("%s/rest/api/2/search?%s", provider.URL, params.Encode())
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Jira Server: use Bearer token for PAT, or Basic Auth for username:password
		if provider.Token != "" {
			req.Header.Set("Authorization", "Bearer "+provider.Token)
		} else {
			auth := base64.StdEncoding.EncodeToString([]byte(provider.User + ":" + provider.Password))
			req.Header.Set("Authorization", "Basic "+auth)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching issues: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Jira Server API returned status %d", resp.StatusCode)
		}

		var searchResp JiraServerSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return nil, fmt.Errorf("decoding issues: %w", err)
		}

		for _, ji := range searchResp.Issues {
			// Parse Jira date format (2024-01-15T10:30:00.000+0000)
			createdAt, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Created)
			updatedAt, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Updated)

			var dueAt *time.Time
			if ji.Fields.DueDate != "" {
				if t, err := time.Parse("2006-01-02", ji.Fields.DueDate); err == nil {
					dueAt = &t
				}
			}

			// Construct the web URL for the issue
			webURL := fmt.Sprintf("%s/browse/%s", provider.URL, ji.Key)

			issue := Issue{
				Source:    provider.Name,
				Title:     ji.Fields.Summary,
				WebURL:    webURL,
				Status:    ji.Fields.Status.Name,
				Priority:  ji.Fields.Priority.Name,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				DueAt:     dueAt,
			}
			allIssues = append(allIssues, issue)
		}

		// Check if there are more pages using offset-based pagination
		if startAt+len(searchResp.Issues) >= searchResp.Total {
			break
		}
		startAt += maxResults
	}

	return allIssues, nil
}

// handleIndex serves the main HTML page
func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// handleFavicon serves the favicon
func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(faviconData)
}

// handleStatus returns the task providers server status as JSON
func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	statuses := a.checkAllProvidersStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

// handleIssues returns the issues as JSON
func (a *App) handleIssues(w http.ResponseWriter, r *http.Request) {
	issues, statuses, err := a.fetchAllIssues()
	sortIssuesDefault(issues)

	response := struct {
		Issues   []Issue          `json:"issues"`
		Statuses []ProviderStatus `json:"statuses"`
		Error    string           `json:"error,omitempty"`
	}{
		Issues:   issues,
		Statuses: statuses,
	}

	if err != nil {
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleProviderIssues returns issues from a single provider as JSON
func (a *App) handleProviderIssues(w http.ResponseWriter, r *http.Request) {
	// Extract provider name from URL path: /api/provider/{name}/issues
	path := r.URL.Path
	const prefix = "/api/provider/"
	const suffix = "/issues"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	providerName := path[len(prefix) : len(path)-len(suffix)]
	if providerName == "" {
		http.Error(w, "Provider name required", http.StatusBadRequest)
		return
	}

	// Find the provider by name
	var provider *TaskProvider
	for i := range a.config.TaskProviders {
		if a.config.TaskProviders[i].Name == providerName {
			provider = &a.config.TaskProviders[i]
			break
		}
	}

	if provider == nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Fetch issues from this provider
	var issues []Issue
	var fetchErr error

	switch provider.Type {
	case TaskProviderGitLab:
		issues, fetchErr = a.fetchGitLabIssues(*provider)
	case TaskProviderJiraCloud:
		issues, fetchErr = a.fetchJiraCloudIssues(*provider)
	case TaskProviderJiraServer:
		issues, fetchErr = a.fetchJiraServerIssues(*provider)
	default:
		fetchErr = fmt.Errorf("unsupported task provider type: %s", provider.Type)
	}

	sortIssuesDefault(issues)

	status := ProviderStatus{
		Name:       provider.Name,
		URL:        provider.URL,
		Type:       string(provider.Type),
		Online:     fetchErr == nil,
		IssueCount: len(issues),
	}

	response := struct {
		Issues []Issue        `json:"issues"`
		Status ProviderStatus `json:"status"`
		Error  string         `json:"error,omitempty"`
	}{
		Issues: issues,
		Status: status,
	}

	if fetchErr != nil {
		log.Printf("ERROR: Failed to fetch issues from provider %q (%s): %v", provider.Name, provider.URL, fetchErr)
		response.Error = fetchErr.Error()
	} else {
		log.Printf("INFO: Fetched %d issues from provider %q", len(issues), provider.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load configuration
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create app instance
	app := &App{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Check provider status at startup
	log.Printf("Configured %d task provider(s), checking connectivity...", len(config.TaskProviders))
	statuses := app.checkAllProvidersStatus()
	for _, status := range statuses {
		if status.Online {
			log.Printf("  - %s (%s) [%s]: OK", status.Name, status.URL, status.Type)
		} else {
			log.Printf("  - %s (%s) [%s]: OFFLINE - %s", status.Name, status.URL, status.Type, status.Error)
		}
	}

	// Setup routes
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/favicon.ico", app.handleFavicon)
	http.HandleFunc("/api/status", app.handleStatus)
	http.HandleFunc("/api/issues", app.handleIssues)
	http.HandleFunc("/api/provider/", app.handleProviderIssues)

	// Start server
	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Starting server on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
