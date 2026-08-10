# AGENTS.md

## Project Overview

Simple service that identifies clusters with very old data in the Insights Results Aggregator database (e.g. Insights Operator disabled / cluster gone) and can prune that data. By default it only **displays** such clusters; deletion needs explicit `-cleanup` / `-cleanup-all` flags.

**Tech Stack**: Go 1.25 (`go.mod`), PostgreSQL (`lib/pq`; sqlite3 also supported), Viper + TOML, zerolog, insights-operator-utils, ClowdApp cron job (`deploy/clowdapp.yaml`).

## Team context

This repository is owned by **ObsInt Processing**.

Before working here, **load and follow the team-info skill** (Shared Standards, PR rules, Go/Python conventions, testing norms, related services):

- Skill source: https://github.com/RedHatInsights/processing-tools/blob/master/skills/team-info/SKILL.md
- Install (example): `npx skills add RedHatInsights/processing-tools --skill team-info -g -a cursor -y`

Do **not** duplicate team-wide rules in this file. Keep this AGENTS.md limited to **this repository**.

**Related repos**:

- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator) — database this tool cleans
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec) — BDD (`cleaner_tests.sh`)

## Repository Structure

```text
.
├── cleaner.go          # main, CLI flags, operation dispatch
├── storage.go          # SQL list/delete/fill/vacuum
├── config.go           # Viper load + env overrides
├── types.go            # ClusterName, CliFlags, Summary, TableAndKey
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

- Go toolchain from `go.mod` (`go 1.25.0`)
- Optional local DB: `podman-compose up -d`, then `./insights-results-aggregator-cleaner -fill-in-db` (see README; `docker-compose.yaml` in tree)

### Running Tests

- `make test` → `./unit-tests.sh` (`go test` with 2m timeout and coverage profile)
- BDD: clone insights-behavioral-spec and run `cleaner_tests.sh`

### Code Quality

- `make lint` — golangci-lint via pre-commit
- `make style` — shellcheck + abcgo + lint
- `make before_commit` is wired to `style test integration_tests openapi-check license`, but `integration_tests` and `openapi-check` are **not** defined as Makefile targets here — prefer `make style` and `make test` (plus `./check_coverage.sh` if needed)

### Building and Running

- `make build` → `./insights-results-aggregator-cleaner`
- `make run` — build and execute (default: list old records)
- Flags include `-show-configuration`, `-cleanup`, `-cleanup-all`, `-dry-run` (default **true**; applies to cleanup-all only), `-max-age`, `-clusters`, `-fill-in-db`, `-vacuum`, `-summary`

## Key Architectural Patterns

### Data flow

1. Load config from `config.toml` or `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`, overridable by `INSIGHTS_RESULTS_CLEANER__*` env vars; connect to aggregator DB (`schema` = `ocp_recommendations` or `dvo_recommendations`).
2. Default path: query old reports / related rows by `max_age` and log (or write) cluster IDs — **no deletes**.
3. Explicit cleanup:
   - `-cleanup`: deletes rows for clusters in `cluster_list.txt` and/or `-clusters` using schema-scoped `tablesAndKeysInOCPDatabase` / `tablesAndKeysInDVODatabase`. **`-dry-run` does not apply.**
   - `-cleanup-all`: age-based work over `allTablesToDelete` (`tablesToDeleteOCP` + `tablesToDeleteDVO`) — not limited by `storage.schema`. With `-dry-run=true`, SQL `DELETE` is rewritten to `SELECT`.
4. ClowdApp CronJob `insights-aggregator-cleaner` / job `cleaner` runs `-dry-run=${DRY_RUN} -cleanup-all=${CLEANUP_ALL}`; template defaults `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days`, schedule `0 12 * * *`; DB via `sharedDbAppName: ccx-insights-results`. Env sets `SCHEMA=ocp_recommendations`, while cleanup-all SQL still includes DVO delete statements in code.

### Components

- `cleaner.go`: CLI + `doSelectedOperation`
- `storage.go`: connection, list/delete SQL, fill-in test data, vacuum
- `config.go`: `LoadConfiguration`
- `deploy/clowdapp.yaml`: schedule and job parameters

### Configuration

- `config.toml` — `[storage]`, `[logging]`, `[cleaner]` (`max_age`, `cluster_list_file`), `[sentry]`
- Config path env: `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`
- Overrides: `INSIGHTS_RESULTS_CLEANER__STORAGE__*`, `__LOGGING__*`, `__CLEANER__MAX_AGE`, …
- Details: [README](./README.md), [docs/index.md](./docs/index.md), [GitHub Pages](https://redhatinsights.github.io/insights-results-aggregator-cleaner/)

## Working with this Repository

**As an agent, you should create a TODO list** when working on tasks to track progress and ensure all steps are completed systematically.

## Code Conventions

- Follow team-info Go standards; CONTRIBUTING asks for Effective Go commentary on end-user methods.
- Flat `package main` — logic in `cleaner.go` / `storage.go` / `config.go`.
- When extending per-cluster cleanup, keep `report` last in `tablesAndKeysInOCPDatabase` (comment in `storage.go`: constraints).

## Important Notes

### Dependencies

Direct modules: see `require (` in `go.mod` (toml, go-sqlmock, insights-operator-utils, uuid, pq, sqlite3, tablewriter, app-common-go, zerolog, viper, testify, go-capture).

### Testing

- Unit tests use `go-sqlmock` (e.g. `storage_test.go`)
- BDD lives in insights-behavioral-spec
- Prefer cleanup table lists in `storage.go` over the README “Database tables affected” section when they diverge (e.g. `report_info` is in code)

### Monitoring

- Optional Sentry via `[sentry]` in config and ClowdApp secret `insights-results-aggregator-cleaner-dsn`

## Common Tasks

### Add or change which tables are cleaned

1. Per-cluster `-cleanup`: update `tablesAndKeysInOCPDatabase` or `tablesAndKeysInDVODatabase` in `storage.go`.
2. Age-based `-cleanup-all`: add SQL constants and entries in `tablesToDeleteOCP` / `tablesToDeleteDVO`.
3. Mirror in `storage_test.go` (existing cleanup tests).
4. Update README “Database tables affected” when the set changes.

### Run a dry-run cleanup-all locally

1. Point `config.toml` (or env) at a non-prod DB.
2. `make build`
3. `./insights-results-aggregator-cleaner -cleanup-all -dry-run=true -max-age="90 days"`
4. Only set `-dry-run=false` when age-based deletion is intentional.
5. Do not use `-cleanup` expecting a dry-run — that path deletes immediately for listed clusters.

## Pull Request Guidelines

### Before Creating a PR

- Run `make style` and `make test` (prefer these over `make before_commit` until Makefile deps exist)
- Also follow team-info Pull Request Requirements and Testing

### Repo-specific checklist

- Document user-visible CLI or cleanup-table changes (CONTRIBUTING)

## Deployment Information

- Configs: `deploy/clowdapp.yaml`
- ClowdApp name: `insights-aggregator-cleaner`; image under `obsint-processing-tenant/.../insights-results-aggregator-cleaner`; shared DB: `ccx-insights-results`
- Template defaults: `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days`, `JOB_SCHEDULE=0 12 * * *`
- Environments and promotion: see team-info Deployment Flow
- Local: `docker-compose.yaml` / README fill-in-db flow

## Security Considerations

- Follow team-info Security
- `-fill-in-db` is not for production (README)
- `-cleanup-all` defaults to dry-run; `-cleanup` has no dry-run and deletes immediately

## Debugging Tips

- `-show-configuration` to verify resolved storage/schema/max_age
- `-output <file>` to capture old-cluster listing
- Exit codes (`cleaner.go`): `0` OK, `1` storage, `2` fill-in, `3` cleanup, `4` vacuum (README documents `0`–`3`)
- Package docs: [cleaner](https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner.html), [storage](https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/storage.html)

## External References

- [README](./README.md)
- [GitHub Pages](https://redhatinsights.github.io/insights-results-aggregator-cleaner/)
- [docs/index.md](./docs/index.md)
- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator)
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec)
