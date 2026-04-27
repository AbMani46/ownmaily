# OwnMaily — Session State

## Session 1A — Verify Setup (complete)

### What was verified and fixed

| Check                                        | Result                                                                           |
| -------------------------------------------- | -------------------------------------------------------------------------------- |
| `sqlc.yml` configured correctly              | Pass                                                                             |
| `Taskfile.yml` has all required tasks        | Fixed — was missing dev, build, migrate-up, migrate-down, docker-up, docker-down |
| `golang-migrate` finds migrations dir        | Pass — `db/migrations` wired correctly                                           |
| Docker Compose starts and Postgres reachable | Fixed — docker-compose.yml was empty; Postgres 18 volume mount path fixed        |
| Test migration runs up and down cleanly      | Pass                                                                             |

### Issues found and fixed

1. **docker-compose.yml was empty** — created with Postgres 18 Alpine, port 5433, healthcheck
2. **Postgres 18 volume path** — Postgres 18+ requires mount at `/var/lib/postgresql` (not `/data`); fixed in docker-compose.yml
3. **Taskfile was incomplete** — added `dev`, `build`, `migrate-up`, `migrate-down`, `docker-up`, `docker-down`; renamed `sqlc` → `sqlc-gen`; dropped non-project default task
4. **`.env` had placeholder DB_URL** — replaced with real connection string `postgres://ownmaily:ownmaily@localhost:5433/ownmaily?sslmode=disable`
5. **No entry point** — created `cmd/server/main.go` (minimal placeholder) so `dev` and `build` tasks compile

### Known limitation — sqlc-gen

`task sqlc-gen` fails until at least one `.sql` file exists in `db/queries`. This is expected sqlc behaviour. Will be populated in Session 2.

---

## Session 1 — Project Scaffold (complete)

### What was built

| Item                                                         | Status |
| ------------------------------------------------------------ | ------ |
| Go dependencies (chi, pgx/v5, godotenv, golang-migrate)      | Done   |
| `internal/config/config.go` — Config struct + Load()         | Done   |
| `internal/db/db.go` — Connect() with pgxpool, max 10 conns   | Done   |
| `internal/db/migrate.go` — RunMigrations() via iofs driver   | Done   |
| `db/embed.go` — embeds db/migrations FS for migration runner | Done   |
| Migration 000001_init (replaced 000001_setup)                | Done   |
| `cmd/server/main.go` — full server wiring                    | Done   |
| `frontend/index.html` — placeholder                          | Done   |
| All internal/ subdirs with .gitkeep                          | Done   |

### Verification checklist

| Check                          | Result                           |
| ------------------------------ | -------------------------------- |
| `task docker-up`               | Pass — Postgres starts clean     |
| `task migrate-up`              | Pass — applies migration         |
| `task migrate-up` (second run) | Pass — logs "already up to date" |
| `task build`                   | Pass — compiles to bin/ownmaily  |
| `curl localhost:4400/health`   | Pass — returns `{"status":"ok"}` |
| Server logs                    | "OwnMaily started on :4400"      |

### Deviations from plan

- **Port changed to 4400** — user preference (was 3000 in original plan)
- **Embed lives in `db/embed.go`** — `//go:embed` can't traverse `../` so the embed must live in the same directory tree as the migrations. Created `db/embed.go` (package `db`) which exports `Migrations embed.FS`. `main.go` imports it as `dbembed`. `internal/db/migrate.go` accepts `fs.FS` so it stays testable.

### Architecture note — embed strategy

```
db/
  embed.go        <- package db; //go:embed migrations; var Migrations embed.FS
  migrations/
    000001_init.up.sql
    000001_init.down.sql
internal/db/
  db.go           <- Connect(dbUrl) — pgxpool
  migrate.go      <- RunMigrations(dbUrl, fs.FS) — iofs source driver
cmd/server/
  main.go         <- wires everything; imports dbembed + internal/db
```

---

## Current Project Structure

```
ownmaily/
├── cmd/
│   └── server/
│       └── main.go
├── db/
│   ├── embed.go             # embeds migrations FS
│   ├── migrations/
│   │   ├── 000001_init.up.sql
│   │   └── 000001_init.down.sql
│   └── queries/             # empty — populated Session 2
├── frontend/
│   └── index.html           # placeholder
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── db.go
│   │   └── migrate.go
│   ├── handler/             # .gitkeep
│   ├── mailer/              # .gitkeep
│   ├── middleware/          # .gitkeep
│   ├── sqlc/                # empty — generated Session 2
│   └── worker/              # .gitkeep
├── bin/                     # git-ignored build output
├── docker-compose.yml
├── Taskfile.yml
├── sqlc.yml
├── go.mod
├── go.sum
├── .env                     # git-ignored
├── OWNMAILY_CONTEXT.MD
└── OWNMAILY_FEATURES.MD
```

## Taskfile Tasks

| Task                           | Command                                     | Notes                            |
| ------------------------------ | ------------------------------------------- | -------------------------------- |
| `task dev`                     | `go run ./cmd/server/...`                   | hot dev loop                     |
| `task build`                   | builds to `bin/ownmaily`                    |                                  |
| `task sqlc-gen`                | sqlc via Docker                             | requires at least one query file |
| `task migrate-up`              | runs all pending migrations                 |                                  |
| `task migrate-down`            | rolls back 1 migration (STEP=N to override) |                                  |
| `task migrate-create NAME=foo` | creates 000xxx_foo.up/down.sql              |                                  |
| `task docker-up`               | `docker compose up -d`                      |                                  |
| `task docker-down`             | `docker compose down`                       |                                  |
| `task psql`                    | opens psql shell                            |                                  |

## Environment

```
DB_URL=postgres://ownmaily:ownmaily@localhost:5433/ownmaily?sslmode=disable
APP_SECRET=change-me-before-production
INSTALLATION_URL=http://localhost:4400
PORT=4400
```

## Tool Versions

| Tool           | Version                      |
| -------------- | ---------------------------- |
| Go             | 1.26.2                       |
| golang-migrate | v4.19.1                      |
| sqlc           | via Docker (sqlc/sqlc image) |
| task           | 3.46.3                       |
| Docker Compose | v2.38.1                      |
| PostgreSQL     | 18-alpine                    |

---

## Session 2 — Next (Schema + SQLC + Auth)

Ready to begin:

- Full DB schema (subscribers, lists, campaigns, jobs, settings, api_keys tables)
- SQLC queries and generated Go code
- JWT auth middleware + login endpoint
- API key middleware
