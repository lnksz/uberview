package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed index.html
var staticFiles embed.FS

// TaskProviderType represents the type of task provider
type TaskProviderType string

const (
	TaskProviderGitLab TaskProviderType = "gitlab"
)

// TaskProvider represents a task provider configuration
type TaskProvider struct {
	Type  TaskProviderType `yaml:"type"`
	Name  string           `yaml:"name"`
	URL   string           `yaml:"url"`
	Token string           `yaml:"token"`
	User  string           `yaml:"user"`
}

// Config represents the YAML configuration structure
type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	TaskProvider TaskProvider `yaml:"task_provider"`
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
	Labels      []string  `json:"labels"`
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

// Issue represents a generic issue from any task provider
type Issue struct {
	Source    string    `json:"source"`
	Title     string    `json:"title"`
	WebURL    string    `json:"web_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderStatus represents the task provider server status
type ProviderStatus struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Type   string `json:"type"`
	Online bool   `json:"online"`
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

// checkProviderStatus checks if the task provider server is reachable
func (a *App) checkProviderStatus() ProviderStatus {
	status := ProviderStatus{
		Name:   a.config.TaskProvider.Name,
		URL:    a.config.TaskProvider.URL,
		Type:   string(a.config.TaskProvider.Type),
		Online: false,
	}

	switch a.config.TaskProvider.Type {
	case TaskProviderGitLab:
		apiURL := fmt.Sprintf("%s/api/v4/version", a.config.TaskProvider.URL)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return status
		}

		req.Header.Set("PRIVATE-TOKEN", a.config.TaskProvider.Token)

		resp, err := a.client.Do(req)
		if err != nil {
			return status
		}
		defer resp.Body.Close()

		status.Online = resp.StatusCode == http.StatusOK
	}

	return status
}

// fetchIssues fetches all open issues assigned to the configured user
func (a *App) fetchIssues() ([]Issue, error) {
	switch a.config.TaskProvider.Type {
	case TaskProviderGitLab:
		return a.fetchGitLabIssues()
	default:
		return nil, fmt.Errorf("unsupported task provider type: %s", a.config.TaskProvider.Type)
	}
}

// fetchGitLabIssues fetches all open issues from GitLab assigned to the configured user
func (a *App) fetchGitLabIssues() ([]Issue, error) {
	var allIssues []Issue
	page := 1
	perPage := 100

	for {
		params := url.Values{}
		params.Set("assignee_username", a.config.TaskProvider.User)
		params.Set("state", "opened")
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))

		apiURL := fmt.Sprintf("%s/api/v4/issues?%s", a.config.TaskProvider.URL, params.Encode())
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", a.config.TaskProvider.Token)

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

		// Convert GitLab issues to generic Issue format
		for _, gi := range gitlabIssues {
			issue := Issue{
				Source:    a.config.TaskProvider.Name,
				Title:     gi.Title,
				WebURL:    gi.WebURL,
				CreatedAt: gi.CreatedAt,
				UpdatedAt: gi.UpdatedAt,
			}
			allIssues = append(allIssues, issue)
		}

		// Check if there are more pages
		if len(gitlabIssues) < perPage {
			break
		}
		page++
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

// handleStatus returns the task provider server status as JSON
func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := a.checkProviderStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleIssues returns the issues as JSON
func (a *App) handleIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := a.fetchIssues()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
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

	// Setup routes
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/api/status", app.handleStatus)
	http.HandleFunc("/api/issues", app.handleIssues)

	// Start server
	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Printf("Task Provider: %s (%s) - Type: %s", config.TaskProvider.Name, config.TaskProvider.URL, config.TaskProvider.Type)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
