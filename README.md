# Uberview

A lightweight dashboard that aggregates open issues assigned to you from multiple task providers 
(GitLab, Jira Cloud, Jira Server) into a single view.

This is a read-only dashboard, not trying to replace the native interfaces of the providers,
but help keep me sane with keeping the overview (GER = "Überblick") of my tasks.

## Features

- Single binary, no dependencies
- Supports GitLab, Jira Cloud, and Jira Server
- Auto light/dark theme based on system preference
- Per-provider refresh buttons
- Real-time provider status indicators

## Usage

```bash
go build -o uberview .
./uberview config.yaml
```

Access the dashboard at `http://localhost:8080`

## Configuration

See `config.example.yaml` for an example configuration file (maintained).
Rename it to `config.yaml` and adjust the settings as needed.

```yaml
server:
  port: 8080

task_providers:
  - type: "gitlab"
    name: "GitLab"
    url: "https://10.0.0.1"
    token: "glpat-xxxx"
    user: "username"

  - type: "jira_cloud"
    name: "Jira Cloud"
    url: "https://company.atlassian.net"
    token: "jira-api-token"
    email: "user@company.com"
    user: "Jon Doe"

  - type: "jira_server"
    name: "Jira Server"
    url: "https://10.0.0.2"
    user: "username"
    token: "personal-access-token"
```

### Provider Authentication

| Provider | Authentication |
|----------|----------------|
| GitLab | Personal Access Token (`token`) |
| Jira Cloud | Email + API Token (Basic Auth) |
| Jira Server | Personal Access Token (Bearer) or username:password (Basic Auth) |

## Development

```bash
go mod tidy
go build -o uberview .
./uberview config.yaml
```

or quick run:

```bash
go run . config.yaml
```
