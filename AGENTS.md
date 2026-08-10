# AGENTS.md

## Project Overview

CLI/cron job that finds and deletes obsolete Insights Results Aggregator data — typically clusters with reports older than a configured max age (Insights Operator gone or disabled). Default invocation **lists** old clusters only; deletion requires explicit `-cleanup` / `-cleanup-all` flags. Shares the aggregator Postgres DB (`ocp_recommendations` or `dvo_recommendations` schema).

**Tech Stack**: Go 1.25 (`go.mod`), PostgreSQL (`lib/pq`; sqlite3 supported for local), Viper + TOML config, zerolog, insights-operator-utils, Clowder CronJob (`deploy/clowdapp.yaml`).

## Team context

This repository is owned by **ObsInt Processing**.

Before working here, **load and follow the team-info skill** (Shared Standards, PR rules, Go/Python conventions, testing norms, related services):

- Skill source: https://github.com/RedHatInsights/processing-tools/blob/master/skills/team-info/SKILL.md
- Install (example): `npx skills add RedHatInsights/processing-tools --skill team-info -g -a cursor -y`

Do **not** duplicate team-wide rules in this file. Keep this AGENTS.md limited to **this repository**.

**Related repos** (from team-info):

- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator) — owns the DB this tool cleans
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec) — BDD (`cleaner_tests.sh`)

## Repository Structure

```text
.
├── cleaner.go          # main, CLI flags, operation dispatch
├── storage.go          # SQL list/delete/fill/vacuum
├── config.go           # Viper load + env overrides
├── types.go            # ClusterName, CliFlags, Summary, TableAndKey types
├── config.toml         # local default config
├── cluster_list.txt    # cluster UUIDs for -cleanup
├── deploy/clowdapp.yaml
├── docs/               # GitHub Pages sources
├── tests/              # sample configs / cluster lists
├── docker-compose.yaml # local Postgres
├── unit-tests.sh
├── Makefile
└── *_test.go
```

## Development Workflow

### Setup

- Go toolchain from `go.mod` (currently `go 1.25.0`)
- Optional local DB: `podman-compose up -d` (see README), then `./insights-results-aggregator-cleaner -fill-in-db`

### Running Tests

- `make test` → `./unit-tests.sh` (`go test` with 2m timeout, coverage profile)
- BDD: clone [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec) and run `cleaner_tests.sh`

### Code Quality

- `make lint` — golangci-lint via pre-commit
- `make style` — shellcheck + abcgo + lint
- `make before_commit` — style + test + … (see note below)

Note: `before_commit` lists `integration_tests` and `openapi-check`, which are **not** defined as Makefile targets in this tree — expect that target to fail until those exist or are removed. Prefer `make style` + `make test` (+ `./check_coverage.sh` if needed).

### Building and Running

- `make build` → `./insights-results-aggregator-cleaner`
- `make run` — build and execute (default: list old records)
- Useful flags: `-show-configuration`, `-cleanup`, `-cleanup-all`, `-dry-run` (default **true**, **only for `-cleanup-all`**), `-max-age`, `-clusters`, `-fill-in-db`, `-vacuum`, `-summary`

## Key Architectural Patterns

### Data flow

1. Load config from `config.toml` or `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`, overridable by `INSIGHTS_RESULTS_CLEANER__*` env vars; connect to aggregator DB (`schema` = `ocp_recommendations` or `dvo_recommendations`).
2. Default path: query old reports / related rows by `max_age` and log (or write) cluster IDs — **no deletes**.
3. Explicit cleanup:
   - `-cleanup`: **always deletes** rows for clusters in `cluster_list.txt` and/or `-clusters` (schema-scoped table list). **`-dry-run` does not apply.**
   - `-cleanup-all`: age-based work via `tablesToDeleteOCP` **and** `tablesToDeleteDVO` (both schemas — not limited by `storage.schema`). With `-dry-run=true` (CLI default), SQL `DELETE` is rewritten to `SELECT` so nothing is removed.
4. Deployed as ClowdApp CronJob `insights-aggregator-cleaner` / job `cleaner`: command is `-dry-run=${DRY_RUN} -cleanup-all=${CLEANUP_ALL}`. Template parameter defaults are `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days` — so **template default is dry-run**. Stage/prod may override `DRY_RUN` via app-interface; confirm there before assuming deletes run. DB via shared Clowder app `ccx-insights-results`. ClowdApp sets `SCHEMA=ocp_recommendations` for env, but `-cleanup-all` still touches DVO delete statements in code.

### Components

- `cleaner.go`: CLI + `doSelectedOperation` dispatch (`cleanup`, `cleanupAll`, `displayOldRecords`, …)
- `storage.go`: connection, list/delete SQL, fill-in test data, vacuum
- `config.go`: `LoadConfiguration`
- `deploy/clowdapp.yaml`: schedule (default `0 12 * * *`), `MAX_AGE`, `DRY_RUN`, `CLEANUP_ALL`

### Configuration

- File: `config.toml` — `[storage]`, `[logging]`, `[cleaner]` (`max_age`, `cluster_list_file`), `[sentry]`
- Config path env: `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`
- Overrides: `INSIGHTS_RESULTS_CLEANER__STORAGE__*`, `__LOGGING__*`, `__CLEANER__MAX_AGE`, …
- Details: [GitHub Pages](https://redhatinsights.github.io/insights-results-aggregator-cleaner/) / [docs/index.md](./docs/index.md)

## Working with this Repository

**As an agent, you should create a TODO list** when working on tasks to track progress and ensure all steps are completed systematically.

## Code Conventions

- Follow team-info Go standards; CONTRIBUTING asks for Effective Go commentary on exported APIs and `make before_commit` when it works.
- Flat `package main` — logic lives in `cleaner.go` / `storage.go` / `config.go`, not subpackages.
- Schema-specific table lists: extend `tablesAndKeysIn*` (per-cluster cleanup) and/or `tablesToDelete*` (age-based cleanup-all) in `storage.go`.

## Important Notes

### Dependencies

- Direct: toml, viper, pq, sqlite3, zerolog, insights-operator-utils, app-common-go, uuid, tablewriter, testify, go-sqlmock

### Testing

- Unit tests use `go-sqlmock` extensively (`storage_test.go`, `cleaner_test.go`)
- BDD scenarios live in insights-behavioral-spec (not this repo)
- README “Database tables affected” may lag code — trust `tablesAndKeysIn*` / `tablesToDelete*` in `storage.go` (e.g. code includes `report_info`, README may not)

### Monitoring

- Optional Sentry via config / ClowdApp secret `insights-results-aggregator-cleaner-dsn`
- No HTTP metrics server in this binary (batch/CLI job)

## Common Tasks

### Add or change which tables are cleaned

1. For **per-cluster** `-cleanup`: update `tablesAndKeysInOCPDatabase` or `tablesAndKeysInDVODatabase` in `storage.go` (table + key column; keep `report` last for FK order).
2. For **age-based** `-cleanup-all`: add SQL constants and entries in `tablesToDeleteOCP` / `tablesToDeleteDVO`.
3. Mirror behavior in `storage_test.go` (existing cleanup tests are the pattern).
4. Update README “Database tables affected” if the list of tables changes.

### Run a safe dry-run cleanup-all locally

1. Point `config.toml` (or env) at a non-prod DB.
2. `make build`
3. `./insights-results-aggregator-cleaner -cleanup-all -dry-run=true -max-age="90 days"` (`-dry-run` defaults true; confirmed in binary `-h`)
4. Only set `-dry-run=false` when age-based deletion is intentional.
5. Do **not** use `-cleanup` expecting a dry-run — that path deletes immediately for listed clusters.

## Pull Request Guidelines

### Before Creating a PR

- Run `make style` and `make test` (prefer these over broken `make before_commit` until Makefile deps are fixed)
- Also follow team-info Pull Request Requirements and Testing

### Repo-specific checklist

- Document user-visible CLI or cleanup-table changes in README / Pages sources under `docs/`

## Deployment Information

- Configs: `deploy/clowdapp.yaml` (CronJob; image under `obsint-processing-tenant/.../insights-results-aggregator-cleaner`)
- ClowdApp name: `insights-aggregator-cleaner`; shared DB: `ccx-insights-results`
- Defaults in template: `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days`, schedule `0 12 * * *` (noon)
- Environments and promotion: see team-info Deployment Flow; check app-interface SaaS params if real deletes are expected in an env
- Local compose: `docker-compose.yaml` / README fill-in-db flow

## Security Considerations

- Follow team-info Security
- Credentials via config/env (never commit real secrets); production uses Clowder DB binding + optional Sentry DSN secret
- `-cleanup-all` is safe by default (`-dry-run=true`); `-cleanup` has no dry-run and deletes immediately
- `-fill-in-db` is for non-prod only

## Debugging Tips

- `-show-configuration` to verify resolved storage/schema/max_age
- `-output <file>` to capture old-cluster listing
- Exit codes: `0` ok, `1` storage error, `2` fill-in failed, `3` cleanup failed
- Package docs on Pages: [cleaner](https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner.html), [storage](https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/storage.html)

## External References

- [README](./README.md) — CLI usage, DB schema notes, Makefile overview
- [GitHub Pages](https://redhatinsights.github.io/insights-results-aggregator-cleaner/) — configuration and usage
- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator) — DB owner
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec) — BDD for this service
