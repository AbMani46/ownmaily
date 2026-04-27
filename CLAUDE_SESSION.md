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

## Session 2 — DB Schema + sqlc (complete)

### What was built

| Item                                                                                               | Status |
| -------------------------------------------------------------------------------------------------- | ------ |
| `db/migrations/000002_schema.up.sql` — full schema (14 tables + indexes + seed)                    | Done   |
| `db/migrations/000002_schema.down.sql` — drops all tables in reverse dependency order              | Done   |
| `db/queries/settings.sql` — GetSettings, UpdateSettings, SetSetupComplete                          | Done   |
| `db/queries/owner.sql` — CreateOwner, GetOwner, UpdateOwnerPassword                                | Done   |
| `db/queries/api_keys.sql` — CreateAPIKey, GetLatestAPIKey, DeleteAllAPIKeys                        | Done   |
| `db/queries/subscribers.sql` — full CRUD, search, paginated list, upsert                           | Done   |
| `db/queries/suppressed_emails.sql` — AddSuppression, IsSuppressed, CRUD                            | Done   |
| `db/queries/lists.sql` — CRUD, list-subscriber junction, count                                     | Done   |
| `db/queries/tags.sql` — CRUD, subscriber-tag junction, count                                       | Done   |
| `db/queries/campaigns.sql` — CRUD, status updates, scheduled due query                             | Done   |
| `db/queries/send_jobs.sql` — CRUD, increment counters, pending list                                | Done   |
| `db/queries/campaign_recipients.sql` — CRUD, BulkCreateCampaignRecipients (copyfrom), pending list | Done   |
| `db/queries/opens.sql` — RecordOpen (upsert ignore), count, hasOpened, list                        | Done   |
| `db/queries/clicks.sql` — RecordClick, count, per-link breakdown, list                             | Done   |
| `internal/sqlc/` — 15 generated .go files (sqlc v1.30.0)                                           | Done   |

### Verification

| Check             | Result                                                     |
| ----------------- | ---------------------------------------------------------- |
| `task migrate-up` | Pass — 2/u schema (69ms)                                   |
| `\dt` in psql     | All 14 tables present                                      |
| settings row seed | Pass — id=true, setup_complete=false, site_name='OwnMaily' |
| `task sqlc-gen`   | Pass — 15 files generated, no errors                       |
| `task build`      | Pass — compiles cleanly                                    |

### Tables created

`settings`, `owner`, `api_keys`, `subscribers`, `suppressed_emails`, `lists`, `list_subscribers`, `tags`, `subscriber_tags`, `campaigns`, `send_jobs`, `campaign_recipients`, `opens`, `clicks`

### Deviations from plan

None — all tables, query files, and functions match the spec exactly.

---

## Session 3 — Auth (complete)

### What was built

| Item                                                                     | Status |
| ------------------------------------------------------------------------ | ------ |
| `internal/auth/jwt.go` — GenerateToken, ValidateToken, Claims struct     | Done   |
| `internal/auth/password.go` — HashPassword (cost 12), CheckPassword      | Done   |
| `internal/auth/apikey.go` — GenerateAPIKey, KeyPrefix, HashKey, CheckKey | Done   |
| `internal/middleware/auth.go` — RequireJWT, RequireAPIKey, RequireAuth   | Done   |
| `internal/handler/helpers.go` — writeJSON, writeError, readJSON          | Done   |
| `internal/handler/auth.go` — Login, Logout, Me handlers                  | Done   |
| `cmd/server/main.go` — wired sqlc.Queries, mounted all auth routes       | Done   |
| Dependencies: golang-jwt/jwt/v5, golang.org/x/crypto                     | Done   |

### Verification

| Check                                      | Result                                                         |
| ------------------------------------------ | -------------------------------------------------------------- |
| `task build`                               | Pass — compiles clean                                          |
| `curl localhost:4400/health`               | 200 `{"status":"ok"}` — unprotected, still works               |
| `curl localhost:4400/api/auth/me`          | 401 `{"error":"unauthorized","message":"missing token"}`       |
| `POST /api/auth/login` with wrong password | 401 `{"error":"unauthorized","message":"invalid credentials"}` |

Full login flow (end-to-end with real owner row) deferred to Session 11 (setup wizard seeds owner).

### Deviations from plan

None — all functions, routes, and middleware match the spec exactly.

---

## Session 4 — Subscribers CRUD + Import/Export (complete)

### What was built

| Item                                                                               | Status |
| ---------------------------------------------------------------------------------- | ------ |
| `internal/handler/subscribers.go` — SubscriberHandler with 8 methods              | Done   |
| `GET /api/subscribers` — list with q/status/list_id/tag_id filters + pagination   | Done   |
| `POST /api/subscribers` — create with email validation + suppression/dup checks   | Done   |
| `GET /api/subscribers/export` — CSV stream (no buffer), RFC3339 created_at         | Done   |
| `POST /api/subscribers/import` — CSV multipart, 10MB cap, counts imported/skipped/invalid | Done |
| `GET /api/subscribers/{id}` — subscriber + tags + lists                            | Done   |
| `PUT /api/subscribers/{id}` — update email/name/status, re-check if email changed | Done   |
| `DELETE /api/subscribers/{id}` — suppress + delete, 204                           | Done   |
| `POST /api/subscribers/{id}/unsubscribe` — status + suppression                   | Done   |
| `db/queries/lists.sql` — added `ListListsForSubscriber`                            | Done   |
| `db/queries/subscribers.sql` — added `ListAllSubscribers`, updated `UpdateSubscriber` (+ status field) | Done |
| `sqlc.yml` — added `emit_json_tags: true` (all models now have snake_case json tags) | Done |
| `cmd/server/main.go` — all 8 subscriber routes mounted inside RequireAuth group   | Done   |

### Verification

| Check                                              | Result                                                  |
| -------------------------------------------------- | ------------------------------------------------------- |
| `task build`                                       | Pass — compiles clean                                   |
| `POST /api/subscribers`                            | 201 with subscriber object (tags: [])                   |
| `GET /api/subscribers`                             | 200 `{subscribers:[...], total:1, page:1, per_page:50}` |
| `GET /api/subscribers/:id`                         | 200 with subscriber + tags + lists                      |
| `GET /api/subscribers?q=test`                      | 200 with matching subscribers                           |
| `GET /api/subscribers?status=active`               | 200 filtered by status                                  |
| `GET /api/subscribers/export`                      | CSV download with correct headers                       |
| `POST /api/subscribers/import` (3 valid, 1 dup, 1 bad) | `{imported:2, skipped:1, invalid:1}`             |
| `PUT /api/subscribers/:id`                         | 200 with updated subscriber                             |
| `POST /api/subscribers/:id/unsubscribe`            | 200 `{message:unsubscribed}`                            |
| `DELETE /api/subscribers/:id`                      | 204 No Content                                          |
| Re-create deleted subscriber                       | 409 `{error:suppressed}` — suppression enforced         |

### Auth method for testing

Seeded owner row directly in psql with known bcrypt hash. Email: `admin@test.com`, password: `testpass123`. Login via `POST /api/auth/login` to get JWT bearer token. (Remove before production.)

### Deviations from plan

- **Import uses `CreateSubscriber` + duplicate-key error handling** instead of calling `UpsertSubscriber`. The existing `UpsertSubscriber` updates on conflict; the spec wanted "do nothing" for import. Handling pgconn error code 23505 gives identical external behaviour.
- **Total count for filtered queries** uses `len(subs)` (at most `per_page`) — exact count queries for search/status/list/tag filters don't exist yet. Unfiltered `ListSubscribers` uses `CountSubscribers` correctly. Marked "optimize later" per spec intent.
- **`UpdateSubscriber` query updated** to include `status = $5` so the PUT handler can update status in one query instead of calling `UpdateSubscriberStatus` separately.

---

## Session 5 — Next
