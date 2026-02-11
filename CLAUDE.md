# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

go-migrator is an "UP only" SQL migration library for Go — no down migrations by design. It can be used as a Go package or as a standalone CLI/Docker image.

Module path: `github.com/nfwGytautas/go-migrator` (published as `github.com/apphene/go-migrator`)

## Commands

```bash
make run-tests      # Run tests with -race and -coverprofile
make run-migrator   # Run migrator via docker-compose (starts postgres + migrator)
go build ./...      # Build all packages
go test -race ./... # Run all tests
```

## Architecture

Three packages plus a CLI entry point:

- **Root package (`gomigrator`)** — Public API: `LoadMigrationsFromDir`, `LoadMigrationsFromFS`, `RunMigrations`. Defines `Migration` struct, `Logger` and `MigrationDriver` interfaces.
- **`drivers/`** — Database driver implementations. PostgreSQL (`NewPostgresDriver`) runs each migration in a transaction with optional fixture data. ClickHouse (`NewClickhouseDriver`) executes migrations sequentially without transactions (ClickHouse limitation); fixtures are not supported.
- **`executor/`** — Config loading from YAML (`gomigrator.yaml`) and orchestration. Supports environment variable substitution (`${VAR}`), retries, timeouts, and concurrent execution of multiple migration jobs.
- **`cmd/`** — Minimal CLI that loads config and calls `executor.Execute`.

### Key interfaces (migrator.go)

`MigrationDriver` — implement this to add a new database backend:
- `Connect`, `CreateMigrationsTable`, `GetCurrentVersion`, `ApplyMigration`, `Close`

`Logger` — optional, pass `nil` to `RunMigrations` to disable logging:
- `Info(string)`, `Error(error)`

### Migration file conventions

- Format: `<version>_<name>.sql` (e.g., `01_create_users.sql`)
- Fixtures: `<version>_<name>.fixture.sql` (loaded when fixtures enabled)
- Directory scanning is non-recursive
- Migrations tracked in `gomigrator_migrations` table; already-applied versions are skipped

### Config format (executor/config.go)

```yaml
max-retries: 3
retry-delay: 1s
timeout: 10s
fixtures: false
migrations:
  - name: my-migration
    source: /path/to/migrations
    postgres:
      dsn: postgresql://${DB_USER}:${DB_PASS}@localhost:5432/mydb
  - name: my-clickhouse-migration
    source: /path/to/ch-migrations
    clickhouse:
      dsn: clickhouse://${CH_HOST}:9000/mydb
```

## Dependencies

Go 1.24.9. Key deps: `jackc/pgx/v5` (postgres), `ClickHouse/clickhouse-go/v2` (clickhouse), `caarlos0/env/v9` (env parsing), `stretchr/testify` (tests).
