# AGENTS.md

This repository is a small Go web app that serves a single-page dashboard (embedded `index.html`) and exposes a JSON API that aggregates “open issues assigned to me” from multiple task providers.

## Project type

- Language: Go (`go.mod`)
- Entry point: `main.go`
- UI assets are embedded at build time via `//go:embed` (`index.html`, `favicon.ico`).

## Repo layout

- `main.go` — all application code (HTTP server, config, providers, API handlers)
- `index.html` — embedded frontend UI
- `favicon.ico` — embedded icon
- `config.example.yaml` — maintained example config; copy to `config.yaml`
- `systemd/uberview.service` — sample unit file for running the built binary
- `README.md` — usage notes

There are no additional packages/directories for Go code yet; everything is currently in one file.

## Essential commands

### Build

```bash
go build -o uberview .
```

### Run

- With a config file (default is `config.yaml` if no arg is given):

```bash
./uberview config.yaml
```

- Quick run without building:

```bash
go run . config.yaml
```

### Tests

No Go tests are present in this repository at time of writing (no `*_test.go` files observed). The closest “smoke test” is building and running the server.

## Configuration

- Config file is YAML. Example: `config.example.yaml`.
- Top-level keys:
  - `server.port` (defaults to `8080` if omitted; see `loadConfig` in `main.go:146-165`)
  - `task_providers`: list of providers

Provider types (see `TaskProviderType` in `main.go:25-32`):

- `gitlab`
- `jira_cloud`
- `jira_server`

Provider fields (see `TaskProvider` in `main.go:34-44`):

- Common: `type`, `name`, `url`, `token`, `user`
- Jira Cloud: `email` is required and Basic Auth is `email:token`
- Jira Server: either `token` (Bearer) or `user` + `password` (Basic)

## Runtime behavior / endpoints

Server:

- Binds on `:<port>` and logs the URL as `http://localhost:<port>` (`main.go:714-720`).

Routes (registered in `main.go:707-712`):

- `GET /` — serves embedded `index.html`
- `GET /favicon.ico` — serves embedded favicon
- `GET /api/status` — provider reachability check only (no issue fetch)
- `GET /api/issues` — fetch issues from all providers concurrently
- `GET /api/provider/{name}/issues` — fetch issues from a single provider by *provider Name*

The single-provider endpoint extracts `{name}` by string slicing (`main.go:601-632`); the `name` must match `task_providers[].name` exactly.

## Provider implementations (what to look for)

All provider code is currently in `main.go`:

- GitLab
  - Status check: `GET /api/v4/version` with `PRIVATE-TOKEN` header (`main.go:177-200`)
  - Issues: `GET /api/v4/issues` with query params:
    - `assignee_username`, `state=opened`, `scope=all`, pagination via `per_page` and `page` (`main.go:346-402`)

- Jira Cloud
  - Status check: `GET /rest/api/3/myself` with Basic auth (`main.go:201-226`)
  - Issues: `GET /rest/api/3/search/jql` (note comment: old `/search` is deprecated and returns 410) (`main.go:404-477`)
  - Pagination uses `nextPageToken` / `isLast` (`main.go:410-474`)

- Jira Server
  - Status check: `GET /rest/api/2/myself` with Bearer PAT or Basic user/pass (`main.go:227-256`)
  - Issues: `GET /rest/api/2/search` with `startAt`/`maxResults` pagination (`main.go:479-554`)

Issue aggregation:

- `fetchAllIssues` runs providers concurrently, returns partial results if at least one provider succeeds; it only returns an error when *all* providers fail (`main.go:281-344`).

## Conventions and style observed

- Single `package main` app; types and functions are in one file.
- Standard library HTTP server (`net/http`) with `http.HandleFunc`.
- Concurrency via `sync.WaitGroup` + shared slices guarded by a mutex where needed.
- Logging via `log.Printf` / `log.Fatalf`. Errors from providers are logged; response JSON includes a top-level `error` string for `/api/issues` and `/api/provider/...`.

## Gotchas / non-obvious details

- Assets are embedded: if you add new static files, you must update `//go:embed` declarations accordingly (currently only `index.html` and `favicon.ico`).
- Provider name matching for `/api/provider/{name}/issues` is exact and path-based. Names containing `/` will break routing because the handler is registered at `/api/provider/` and then slices the raw path.
- Jira date parsing ignores parse errors (`time.Parse(...); createdAt, _ := ...`). If you change fields or formats, ensure you handle failures intentionally.

## Deployment notes (observed)

- A hardened systemd unit template is included at `systemd/uberview.service`.
  - Expects the binary at `/usr/local/bin/uberview` and config at `/etc/uberview/config.yaml`.

## No additional agent rule files

No `.cursor/rules`, `.cursorrules`, `.github/copilot-instructions.md`, or `claude.md` files were found in this repo at time of writing.