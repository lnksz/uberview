# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Uberview is for individual technical users who work across multiple GitLab and Jira instances. They use it during day-to-day task triage and planning to see their own assigned work without checking each provider separately.

## Product Purpose

Uberview gives users one read-only overview of their open assigned issues across supported task providers. It exists to reduce the mental overhead of keeping several issue trackers in view while leaving detailed work and issue changes in each provider's native interface. Success means a user can quickly understand what is assigned to them, where it lives, and which dates or priorities need attention.

## Positioning

Uberview is a provider-neutral personal overview rather than another system of record. It normalizes assigned issues from GitLab, Jira Cloud, and Jira Server into one table or timeline while preserving direct links back to each source.

## Operating Context

- Users self-host the dashboard and configure providers through a local YAML file.
- The dashboard checks provider reachability and loads issues per provider on demand.
- Users can combine, sort, and temporarily hide provider results in a table, or inspect dated work in a Gantt view.
- View, zoom, hidden-provider, and Gantt column-width preferences remain in the browser's local storage.
- Issue details and edits continue in GitLab or Jira through outbound links.

## Capabilities and Constraints

- The product is read-only by design.
- It supports GitLab, Jira Cloud, and Jira Server without privileging one provider's workflow.
- It aggregates open issues assigned to the configured user and retains partial results when some providers fail.
- It exposes provider status, per-provider refresh, issue counts, filtering by provider visibility, sortable issue metadata, and Gantt date ranges.
- It must remain self-hosted and private, with no hosted service or external task-data processing.
- It must remain deployable as a single Go binary with the frontend embedded and no runtime application dependencies.
- Provider credentials remain in the local YAML configuration; their storage and host security are the operator's responsibility.
- The interface follows the browser's light or dark system preference.

## Brand Commitments

- The product name is Uberview, referencing the German word "Überblick" for overview.
- Product language should be direct, practical, and modest. It should not imply that Uberview replaces GitLab or Jira.

## Evidence on Hand

- `README.md` documents the product purpose, supported providers, local deployment, and configuration.
- `config.example.yaml` documents the supported provider and authentication configurations.
- `main.go` contains the provider integrations, normalized issue model, concurrent aggregation, partial-failure behavior, and embedded web server.
- `index.html` contains the working dashboard, including provider controls, table and Gantt views, local preferences, loading, empty, and provider-error states.
- `favicon.ico` is the existing product icon.
- No testimonials, customer claims, usage benchmarks, pricing, or compliance claims are present and future work must not fabricate them.

## Product Principles

1. Make the whole workload legible before asking users to act.
2. Stay provider-neutral while preserving clear source context.
3. Keep actions in the systems of record and make returning to them effortless.
4. Prefer private, low-maintenance operation over platform complexity.
5. Degrade gracefully so one unavailable provider does not erase the rest of the overview.
