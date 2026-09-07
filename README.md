# Uberview

A lightweight dashboard that aggregates open issues assigned to you from multiple task providers 
(GitLab, Jira Cloud, Jira Server) into a single view.

This is a read-only dashboard, not trying to replace the native interfaces of the providers,
but help keep me sane with keeping the overview (GER = "Überblick") of my tasks.

> Disclaimer: Next to scratching a real itch of mine, this is also a pet project to play
> around with agent tools, models etc.

## Features

- Single binary, no dependencies
- Supports GitLab, Jira Cloud, and Jira Server
- Shows GitLab configurable work item statuses when available
- Auto light/dark theme based on system preference
- Per-provider refresh buttons
- Real-time provider status indicators
- 15-second SQLite issue cache in `sqlite.db`
- Inline, file-backed, or command-backed provider tokens

## Usage

```bash
go build -o uberview .
./uberview config.yaml
```

Access the dashboard at `http://localhost:8080`

## Running with systemd

This repo ships a sample unit file at `systemd/uberview.service`.

This is intended to run as a user service (no root required).

1) Build and install the binary:

```bash
go build -o uberview .
install -d -m 0755 ~/.local/bin
install -m 0755 uberview ~/.local/bin/uberview
```

2) Install config:

```bash
mkdir -p ~/.config/uberview
install -m 0640 config.yaml ~/.config/uberview/config.yaml
```

3) Install and start the service:

```bash
mkdir -p ~/.config/systemd/user
install -m 0644 systemd/uberview.service ~/.config/systemd/user/uberview.service
systemctl --user daemon-reload
systemctl --user enable --now uberview
```

If your systemd/user setup does not allow capability dropping or unprivileged namespacing, you may need to use the provided unit as-is (it avoids those features). If you still hit errors like `status=218/CAPABILITIES`, remove hardening lines such as `SystemCallFilter=`, `SystemCallArchitectures=`, `RestrictSUIDSGID=`, `LockPersonality=`, and `MemoryDenyWriteExecute=` from your user unit.

Logs:

```bash
journalctl --user -u uberview -f
```

To keep it running after logout:

```bash
sudo loginctl enable-linger "$USER"
```

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
    token-file: "secrets/gitlab.token"
    user: "username"

  - type: "jira_cloud"
    name: "Jira Cloud"
    url: "https://company.atlassian.net"
    token-prog: "pass show jira-api-token"
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
| GitLab | Personal Access Token from `token`, `token-file`, or `token-prog` |
| Jira Cloud | Email + API token from `token`, `token-file`, or `token-prog` |
| Jira Server | PAT from `token`, `token-file`, or `token-prog`; or username:password |

Configure no more than one token source for each provider. Relative `token-file` paths are resolved next to the configuration file. `token-prog` is split into an executable and whitespace-separated arguments, runs once at startup without a shell, and must print exactly one non-empty line. Trailing CR/LF characters are trimmed, so normal command output ending in a newline works directly. Its output is kept in memory and is never written to `sqlite.db`.

Successful provider issue responses are cached for 15 seconds in `sqlite.db`, created with mode `0600` in the process working directory. The cache stores normalized issue data only; it does not store provider credentials. Run Uberview from a private directory because ticket titles and metadata may themselves be sensitive.

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
