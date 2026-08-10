# AGENTS.md

## Project Overview

Simple service that identifies clusters with very old data in the Insights Results Aggregator database (e.g. Insights Operator disabled / cluster gone) and can prune that data. By default it only **displays** such clusters; deletion needs explicit flags. (Source: [README.md](./README.md) Description + Default operation / Data cleanup.)

**Tech Stack** (sources: `go.mod` `require`, [deploy/clowdapp.yaml](./deploy/clowdapp.yaml)):

- Go `1.25.0`
- Postgres via `github.com/lib/pq`; also `github.com/mattn/go-sqlite3`
- Config: `github.com/spf13/viper`, `github.com/BurntSushi/toml`
- Logging: `github.com/rs/zerolog`; utils: `github.com/RedHatInsights/insights-operator-utils`
- Deployed as a ClowdApp **job** (cron schedule) in `deploy/clowdapp.yaml`

## Team context

This repository is owned by **ObsInt Processing**.

Before working here, **load and follow the team-info skill** (Shared Standards, PR rules, Go/Python conventions, testing norms, related services):

- Skill source: https://github.com/RedHatInsights/processing-tools/blob/master/skills/team-info/SKILL.md
- Install (example): `npx skills add RedHatInsights/processing-tools --skill team-info -g -a cursor -y`

Do **not** duplicate team-wide rules in this file. Keep this AGENTS.md limited to **this repository**.

**Related repos**:

- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator) — DB this tool cleans (README Description; team-info EDP table)
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec) — BDD via `cleaner_tests.sh` ([README.md](./README.md) BDD tests)

## Repository Structure

```text
.
├── cleaner.go          # main, flags, doSelectedOperation (cleaner.go)
├── storage.go          # SQL list/delete/fill/vacuum (storage.go)
├── config.go           # LoadConfiguration / Viper (config.go)
├── types.go            # ClusterName, CliFlags, Summary, TableAndKey (types.go)
├── config.toml         # default local config (config.toml)
├── cluster_list.txt    # default cluster list file name (config.toml cluster_list_file)
├── deploy/clowdapp.yaml
├── docs/               # Pages sources (README Documentation)
├── tests/              # sample configs / cluster lists
├── docker-compose.yaml # postgres:13.9 + related services
├── unit-tests.sh
├── Makefile            # BINARY:=insights-results-aggregator-cleaner
└── *_test.go
```

## Development Workflow

### Setup

- Toolchain: `go 1.25.0` in `go.mod`
- Local DB: README says `podman-compose up -d`, then `./insights-results-aggregator-cleaner -fill-in-db` ([README.md](./README.md) Test data generation). Compose file present: [docker-compose.yaml](./docker-compose.yaml).

### Running Tests

- `make test` runs `./unit-tests.sh` ([Makefile](./Makefile)); script runs `go test -v -timeout 2m -coverprofile coverage.out` ([unit-tests.sh](./unit-tests.sh))
- BDD: clone insights-behavioral-spec and run `cleaner_tests.sh` ([README.md](./README.md) BDD tests)

### Code Quality

- `make lint` → `pre-commit run --all-files golangci-lint-full` ([Makefile](./Makefile))
- `make style` → `shellcheck` + `abcgo` + `lint` ([Makefile](./Makefile))
- `make before_commit` is defined as `style test integration_tests openapi-check license` ([Makefile](./Makefile)), but **`integration_tests` and `openapi-check` are not defined** as targets in this Makefile — the target as written will fail. Prefer `make style` and `make test`. Coverage helper: [check_coverage.sh](./check_coverage.sh) (referenced by `before_commit`).

### Building and Running

- `make build` / `make run` ([Makefile](./Makefile)); binary name `insights-results-aggregator-cleaner`
- Flags defined in `cleaner.go` `main` / printed by `-h`, including: `-cleanup`, `-cleanup-all`, `-dry-run` (default **true**; help text: applies to cleanup-all), `-max-age`, `-clusters`, `-fill-in-db`, `-vacuum`, `-summary`, `-show-configuration`, `-output`, …

## Key Architectural Patterns

### Data flow

1. Config from `config.toml` (or file named by `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`), with `INSIGHTS_RESULTS_CLEANER_` env prefix / `__` nested overrides ([config.go](./config.go), [docs/index.md](./docs/index.md), [README.md](./README.md) Configuration). Storage `schema` is `ocp_recommendations` or `dvo_recommendations` ([docs/index.md](./docs/index.md); constants in [storage.go](./storage.go)).
2. Default operation: list old records only — no delete ([README.md](./README.md) Default operation; `doSelectedOperation` default → `displayOldRecords` in [cleaner.go](./cleaner.go)).
3. Cleanup modes ([cleaner.go](./cleaner.go), [storage.go](./storage.go)):
   - `-cleanup`: `performCleanupInDB` deletes by cluster list from file and/or `-clusters`; schema selects `tablesAndKeysInOCPDatabase` or `tablesAndKeysInDVODatabase`. **No `DryRun` argument on this path.**
   - `-cleanup-all`: `performCleanupAllInDB` iterates `allTablesToDelete` (`tablesToDeleteOCP` **appended with** `tablesToDeleteDVO`) — not filtered by `storage.schema`. If `dryRun`, `deleteOldRecordsFromTable` replaces `DELETE` with `SELECT`.
4. Deploy job ([deploy/clowdapp.yaml](./deploy/clowdapp.yaml)): ClowdApp `insights-aggregator-cleaner`, job `cleaner`, command `./insights-results-aggregator-cleaner -dry-run=${DRY_RUN} -cleanup-all=${CLEANUP_ALL}`, `sharedDbAppName: ccx-insights-results`, env `INSIGHTS_RESULTS_CLEANER__STORAGE__SCHEMA=ocp_recommendations`. Template parameter **defaults**: `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days`, `JOB_SCHEDULE=0 12 * * *`.

### Components

- [cleaner.go](./cleaner.go) — CLI + `doSelectedOperation`
- [storage.go](./storage.go) — DB connection and SQL operations
- [config.go](./config.go) — `LoadConfiguration`
- [deploy/clowdapp.yaml](./deploy/clowdapp.yaml) — CronJob parameters above

### Configuration

- [config.toml](./config.toml): `[storage]`, `[logging]`, `[cleaner]` (`max_age`, `cluster_list_file`), `[sentry]`
- Env list and schema/driver notes: [README.md](./README.md) / [docs/index.md](./docs/index.md)
- Pages: https://redhatinsights.github.io/insights-results-aggregator-cleaner/

## Working with this Repository

**As an agent, you should create a TODO list** when working on tasks to track progress and ensure all steps are completed systematically.

## Code Conventions

- CONTRIBUTING: Effective Go commentary for end-user methods; run checks before commit ([CONTRIBUTING.md](./CONTRIBUTING.md)). Language-wide rules: team-info.
- All application `.go` files use `package main` (no internal packages).
- Per-cluster table order: `report` is last in `tablesAndKeysInOCPDatabase` with comment `must be at the end due to constraints` ([storage.go](./storage.go)).

## Important Notes

### Dependencies

See direct `require (` block in [go.mod](./go.mod) (toml, go-sqlmock, insights-operator-utils, uuid, pq, sqlite3, tablewriter, app-common-go, zerolog, viper, testify, go-capture).

### Testing

- Unit tests import `github.com/DATA-DOG/go-sqlmock` (e.g. [storage_test.go](./storage_test.go))
- BDD: [README.md](./README.md) BDD tests → insights-behavioral-spec
- Table lists for cleanup: code in [storage.go](./storage.go) (`tablesAndKeysIn*`, `tablesToDelete*`). README section “Database tables affected by this service” does not list `report_info`, which **is** in `tablesAndKeysInOCPDatabase` — prefer the Go lists when extending cleanup.

### Monitoring

- Sentry: `[sentry]` in [config.toml](./config.toml); ClowdApp wires secret `insights-results-aggregator-cleaner-dsn` ([deploy/clowdapp.yaml](./deploy/clowdapp.yaml))

## Common Tasks

### Add or change which tables are cleaned

1. Per-cluster `-cleanup`: edit `tablesAndKeysInOCPDatabase` / `tablesAndKeysInDVODatabase` in [storage.go](./storage.go) (keep `report` last as noted there).
2. Age-based `-cleanup-all`: add SQL + entries in `tablesToDeleteOCP` / `tablesToDeleteDVO` in [storage.go](./storage.go).
3. Extend tests in [storage_test.go](./storage_test.go) (existing `TestPerformCleanup*` patterns).
4. Update README “Database tables affected by this service” if you change the set ([README.md](./README.md)); CONTRIBUTING asks for docs on user-visible behavior changes.

### Run a dry-run cleanup-all locally

1. Point config/env at a non-prod DB.
2. `make build`
3. `./insights-results-aggregator-cleaner -cleanup-all -dry-run=true -max-age="90 days"` — `-dry-run` defaults true (`cleaner.go` / `-h`); dry-run path in `deleteOldRecordsFromTable` ([storage.go](./storage.go)).
4. `-cleanup` does not take dry-run into account (`cleanup` → `performCleanupInDB` only).

## Pull Request Guidelines

### Before Creating a PR

- Prefer `make style` and `make test` (see Makefile note above). CONTRIBUTING still mentions `make before_commit`.
- Also follow team-info Pull Request Requirements and Testing

### Repo-specific checklist

- CONTRIBUTING: include docs for behavior / end-user capability changes ([CONTRIBUTING.md](./CONTRIBUTING.md))

## Deployment Information

- [deploy/clowdapp.yaml](./deploy/clowdapp.yaml): image default `quay.io/redhat-services-prod/obsint-processing-tenant/insights-results-aggregator-cleaner/insights-results-aggregator-cleaner`; ClowdApp name `insights-aggregator-cleaner`; `sharedDbAppName: ccx-insights-results`
- Template defaults: `DRY_RUN=true`, `CLEANUP_ALL=true`, `MAX_AGE=90 days`, `JOB_SCHEDULE=0 12 * * *`
- Stage/prod promotion process: team-info Deployment Flow (not defined in this repo)
- Local: [docker-compose.yaml](./docker-compose.yaml) + README fill-in-db

## Security Considerations

- Follow team-info Security
- `-fill-in-db`: README — “Don't use it on production, of course.”
- `-dry-run` default true for cleanup-all (`cleaner.go`); `-cleanup` always deletes listed clusters (no dry-run in that function path)
- Sentry DSN via optional K8s secret in ClowdApp (see Monitoring)

## Debugging Tips

- `-show-configuration`, `-output` ([cleaner.go](./cleaner.go) flags)
- Exit codes in [cleaner.go](./cleaner.go) (`ExitStatus*` iota): `0` OK, `1` storage, `2` fill-in, `3` cleanup, `4` vacuum. README Exit status documents `0`–`3` only.
- Generated package docs: https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner.html and `…/storage.html` ([README.md](./README.md) Documentation for source files)

## External References

- [README.md](./README.md)
- https://redhatinsights.github.io/insights-results-aggregator-cleaner/
- [docs/index.md](./docs/index.md)
- [insights-results-aggregator](https://github.com/RedHatInsights/insights-results-aggregator)
- [insights-behavioral-spec](https://github.com/RedHatInsights/insights-behavioral-spec)
