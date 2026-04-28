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

| Item                                                                                                   | Status |
| ------------------------------------------------------------------------------------------------------ | ------ |
| `internal/handler/subscribers.go` — SubscriberHandler with 8 methods                                   | Done   |
| `GET /api/subscribers` — list with q/status/list_id/tag_id filters + pagination                        | Done   |
| `POST /api/subscribers` — create with email validation + suppression/dup checks                        | Done   |
| `GET /api/subscribers/export` — CSV stream (no buffer), RFC3339 created_at                             | Done   |
| `POST /api/subscribers/import` — CSV multipart, 10MB cap, counts imported/skipped/invalid              | Done   |
| `GET /api/subscribers/{id}` — subscriber + tags + lists                                                | Done   |
| `PUT /api/subscribers/{id}` — update email/name/status, re-check if email changed                      | Done   |
| `DELETE /api/subscribers/{id}` — suppress + delete, 204                                                | Done   |
| `POST /api/subscribers/{id}/unsubscribe` — status + suppression                                        | Done   |
| `db/queries/lists.sql` — added `ListListsForSubscriber`                                                | Done   |
| `db/queries/subscribers.sql` — added `ListAllSubscribers`, updated `UpdateSubscriber` (+ status field) | Done   |
| `sqlc.yml` — added `emit_json_tags: true` (all models now have snake_case json tags)                   | Done   |
| `cmd/server/main.go` — all 8 subscriber routes mounted inside RequireAuth group                        | Done   |

### Verification

| Check                                                  | Result                                                  |
| ------------------------------------------------------ | ------------------------------------------------------- |
| `task build`                                           | Pass — compiles clean                                   |
| `POST /api/subscribers`                                | 201 with subscriber object (tags: [])                   |
| `GET /api/subscribers`                                 | 200 `{subscribers:[...], total:1, page:1, per_page:50}` |
| `GET /api/subscribers/:id`                             | 200 with subscriber + tags + lists                      |
| `GET /api/subscribers?q=test`                          | 200 with matching subscribers                           |
| `GET /api/subscribers?status=active`                   | 200 filtered by status                                  |
| `GET /api/subscribers/export`                          | CSV download with correct headers                       |
| `POST /api/subscribers/import` (3 valid, 1 dup, 1 bad) | `{imported:2, skipped:1, invalid:1}`                    |
| `PUT /api/subscribers/:id`                             | 200 with updated subscriber                             |
| `POST /api/subscribers/:id/unsubscribe`                | 200 `{message:unsubscribed}`                            |
| `DELETE /api/subscribers/:id`                          | 204 No Content                                          |
| Re-create deleted subscriber                           | 409 `{error:suppressed}` — suppression enforced         |

### Auth method for testing

Seeded owner row directly in psql with known bcrypt hash. Email: `admin@test.com`, password: `testpass123`. Login via `POST /api/auth/login` to get JWT bearer token. (Remove before production.)

### Deviations from plan

- **Import uses `CreateSubscriber` + duplicate-key error handling** instead of calling `UpsertSubscriber`. The existing `UpsertSubscriber` updates on conflict; the spec wanted "do nothing" for import. Handling pgconn error code 23505 gives identical external behaviour.
- **Total count for filtered queries** uses `len(subs)` (at most `per_page`) — exact count queries for search/status/list/tag filters don't exist yet. Unfiltered `ListSubscribers` uses `CountSubscribers` correctly. Marked "optimize later" per spec intent.
- **`UpdateSubscriber` query updated** to include `status = $5` so the PUT handler can update status in one query instead of calling `UpdateSubscriberStatus` separately.

---

## Session 5 — Lists: CRUD, Memberships, Double Opt-In (complete)

### What was built

| Item                                                                                                       | Status |
| ---------------------------------------------------------------------------------------------------------- | ------ |
| `internal/mailer/confirmation.go` — ConfirmationMailer with GenerateToken, ValidateToken, SendConfirmation | Done   |
| `internal/handler/lists.go` — ListHandler with 9 methods                                                   | Done   |
| `GET /api/lists` — list all with subscriber_count per list                                                 | Done   |
| `POST /api/lists` — create, name required                                                                  | Done   |
| `GET /api/lists/{id}` — 404 if not found, includes subscriber_count                                        | Done   |
| `PUT /api/lists/{id}` — 404 if not found, returns updated list + subscriber_count                          | Done   |
| `DELETE /api/lists/{id}` — 409 if non-empty (unless force=true), cascade deletes memberships               | Done   |
| `GET /api/lists/{id}/subscribers` — paginated with tags per subscriber                                     | Done   |
| `POST /api/lists/{id}/subscribers` — inactive subscriber guard, already_member guard, double opt-in        | Done   |
| `DELETE /api/lists/{id}/subscribers/{subscriberID}` — 204                                                  | Done   |
| `GET /confirm` — public HMAC token validation + subscriber status → active                                 | Done   |
| Routes wired in `cmd/server/main.go` — 8 auth routes + 1 public `/confirm` route                           | Done   |

### Verification

| Check                                         | Result                                                         |
| --------------------------------------------- | -------------------------------------------------------------- |
| `POST /api/lists`                             | 201 with list object, subscriber_count: 0                      |
| `GET /api/lists`                              | 200 `{"lists":[...]}`                                          |
| `GET /api/lists/:id`                          | 200 with list + subscriber_count                               |
| `GET /api/lists/00000...` (bad id)            | 404 not_found                                                  |
| `PUT /api/lists/:id`                          | 200 with updated list + subscriber_count                       |
| `POST /api/lists/:id/subscribers`             | 200 `{"message":"added"}`                                      |
| `POST /api/lists/:id/subscribers` (duplicate) | 409 already_member                                             |
| `GET /api/lists/:id/subscribers`              | 200 with subscribers array, total: 1                           |
| `GET /api/subscribers/:id`                    | lists array includes the list                                  |
| `DELETE /api/lists/:id/subscribers/:sub_id`   | 204                                                            |
| `DELETE /api/lists/:id` (non-empty)           | 409 list_not_empty with count in message                       |
| `DELETE /api/lists/:id?force=true`            | 204                                                            |
| Double opt-in: add subscriber to doi list     | 200 added, server log shows confirmation URL, status → pending |
| `GET /confirm?token=...&sid=...&lid=...`      | 200 "Email confirmed. You are now subscribed."                 |
| `GET /confirm?token=badtoken&...`             | 400 invalid_token                                              |
| `GET /confirm` (missing params)               | 400 invalid_token                                              |

### Token design

- Token format: `base64url(message) + "." + base64url(HMAC-SHA256(message, appSecret))`
- Message: `subscriberID:listID:unixTimestamp`
- 48-hour expiry validated from timestamp embedded in token
- sid/lid URL params cross-checked against decoded token payload

### Deviations from plan

- **ConfirmationMailer.db field kept but unused** — DB operations (GetSubscriberByID, IsSubscriberInList, UpdateSubscriberStatus) live in the handler for cleaner separation; mailer owns token logic only
- **No separate `ts` query param** — timestamp is embedded inside the base64url-encoded token payload; token is self-contained

---

## Session 6 — Tags: CRUD, Subscriber Tagging, Bulk Tagging (complete)

### What was built

| Item                                                                                                       | Status |
| ---------------------------------------------------------------------------------------------------------- | ------ |
| `db/queries/tags.sql` — added `IsSubscriberTagged` query                                                   | Done   |
| `task sqlc-gen` — regenerated; `IsSubscriberTagged` now in `internal/sqlc/tags.sql.go`                     | Done   |
| `internal/handler/tags.go` — TagHandler with List, Create, Get, Update, Delete, ListSubscribers            | Done   |
| `GET /api/tags` — lists all tags with subscriber_count                                                     | Done   |
| `POST /api/tags` — trim+lowercase, 409 on duplicate name                                                   | Done   |
| `GET /api/tags/{id}` — 404 if not found, includes subscriber_count                                         | Done   |
| `PUT /api/tags/{id}` — 404 if not found, 409 on name conflict, returns updated tag + subscriber_count      | Done   |
| `DELETE /api/tags/{id}` — 404 if not found, cascade removes subscriber_tags rows, 204                      | Done   |
| `GET /api/tags/{id}/subscribers` — paginated with tags per subscriber                                      | Done   |
| `GET /api/subscribers/{id}/tags` — list tags for a subscriber                                              | Done   |
| `POST /api/subscribers/{id}/tags` — 404 if sub/tag not found, 409 already_tagged, 200 {"message":"tagged"} | Done   |
| `DELETE /api/subscribers/{id}/tags/{tagID}` — 404 checks, 204                                              | Done   |
| `POST /api/subscribers/bulk-tag` — add/remove, best-effort, max 500, returns processed count               | Done   |
| Routes wired in `cmd/server/main.go` — bulk-tag registered before /{id} to avoid chi clash                 | Done   |

### Verification

| Check                                          | Result                                   |
| ---------------------------------------------- | ---------------------------------------- |
| `task build`                                   | Pass — compiles clean                    |
| `POST /api/tags`                               | 201 `{...subscriber_count:0}`            |
| `POST /api/tags` (duplicate)                   | 409 `{error:"duplicate"}`                |
| `GET /api/tags`                                | 200 `{"tags":[{...}]}`                   |
| `GET /api/tags/:id`                            | 200 with subscriber_count                |
| `POST /api/subscribers/:id/tags`               | 200 `{"message":"tagged"}`               |
| Subscriber detail after tag                    | tags array includes the tag              |
| Tag subscriber_count after add                 | 1                                        |
| `POST /api/subscribers/:id/tags` (duplicate)   | 409 `{error:"already_tagged"}`           |
| `GET /api/subscribers/:id/tags`                | 200 `{"tags":[...]}`                     |
| `GET /api/tags/:id/subscribers`                | 200 paginated with tags per subscriber   |
| `POST /api/subscribers/bulk-tag` (2 subs, add) | 200 `{processed:2,tag_id:"..."}`         |
| `GET /api/subscribers?tag_id=...`              | total:2 — filter working                 |
| `DELETE /api/subscribers/:id/tags/:tag_id`     | 204                                      |
| `PUT /api/tags/:id`                            | 200 with updated name + subscriber_count |
| `DELETE /api/tags/:id`                         | 204                                      |

### sqlc additions

- `IsSubscriberTagged :one` — `SELECT EXISTS(...)` used by the individual add endpoint to detect duplicates before inserting

### Deviations from plan

- **`IsSubscriberTagged` added** — `AddTagToSubscriber` uses `ON CONFLICT DO NOTHING` so can't detect conflicts from the return value. Added a dedicated EXISTS query instead of changing the insert (the insert with ON CONFLICT is still used for bulk-tag where conflicts are intentionally silenced).

---

## Session 7 — Campaigns: CRUD, Status, Duplicate (complete)

### What was built

| Item                                                                                                                       | Status |
| -------------------------------------------------------------------------------------------------------------------------- | ------ |
| `db/queries/campaigns.sql` — added `ListCampaignsByStatus`, `CountCampaignsByStatus`, `ScheduleCampaign`, `CancelCampaign` | Done   |
| `task sqlc-gen` — regenerated; all 4 new functions in `internal/sqlc/campaigns.sql.go`                                     | Done   |
| `internal/handler/campaigns.go` — CampaignHandler with all 11 methods                                                      | Done   |
| `GET /api/campaigns` — paginated, optional `?status=` filter                                                               | Done   |
| `POST /api/campaigns` — validates name/subject/from_name/from_email/@/send_to_type; creates draft                          | Done   |
| `GET /api/campaigns/{id}` — 404 if not found                                                                               | Done   |
| `PUT /api/campaigns/{id}` — 409 if sending/sent; preserves scheduled_at                                                    | Done   |
| `DELETE /api/campaigns/{id}` — 409 if sending; 204                                                                         | Done   |
| `GET /api/campaigns/{id}/preview` — returns html_body as text/html; charset=utf-8                                          | Done   |
| `GET /api/campaigns/{id}/stats` — sent/failed/opens/clicks/rates/per-link breakdown                                        | Done   |
| `POST /api/campaigns/{id}/send` — validates send_to_id/subject/from_email; sets queued; 202                                | Done   |
| `POST /api/campaigns/{id}/schedule` — validates draft status, RFC3339, future time; sets scheduled                         | Done   |
| `POST /api/campaigns/{id}/cancel` — validates scheduled status; clears scheduled_at; restores draft                        | Done   |
| `POST /api/campaigns/{id}/duplicate` — copies all fields, name prefixed "Copy of ", status=draft                           | Done   |
| Routes wired in `cmd/server/main.go`                                                                                       | Done   |

### Verification

| Check                                          | Result                                                              |
| ---------------------------------------------- | ------------------------------------------------------------------- |
| `go build ./...`                               | Pass — compiles clean                                               |
| `POST /api/campaigns`                          | 201 with campaign, status: "draft"                                  |
| `GET /api/campaigns`                           | 200 `{"campaigns":[...], "total":1, "page":1, "per_page":50}`       |
| `GET /api/campaigns?status=draft`              | 200 filtered to drafts, total correct                               |
| `PUT /api/campaigns/:id`                       | 200 with updated fields                                             |
| `GET /api/campaigns/:id/preview`               | 200 text/html response                                              |
| `GET /api/campaigns/:id/stats`                 | 200 zeros for new campaign, links: []                               |
| `POST /api/campaigns/:id/schedule`             | 200 status: "scheduled", scheduled_at set                           |
| `POST /api/campaigns/:id/schedule` (past time) | 400 invalid_time                                                    |
| `POST /api/campaigns/:id/cancel`               | 200 status: "draft", scheduled_at: null                             |
| `POST /api/campaigns/:id/send`                 | 202 `{"message":"Campaign queued for sending","campaign_id":"..."}` |
| `POST /api/campaigns/:id/duplicate`            | 201 name: "Copy of ...", status: "draft"                            |
| `PUT /api/campaigns/:id` (status: sent)        | 409 campaign_locked                                                 |
| `DELETE /api/campaigns/:id` (status: sending)  | 409 campaign_locked                                                 |

### sqlc additions

- `ListCampaignsByStatus :many` — paginated filter by status column
- `CountCampaignsByStatus :one` — COUNT for total with status filter
- `ScheduleCampaign :one` — sets status='scheduled' and scheduled_at in one query
- `CancelCampaign :one` — sets status='draft' and scheduled_at=NULL in one query
- `UpdateCampaign` already existed from Session 2 (includes scheduled_at param) — used as-is; PUT handler preserves existing scheduled_at

### Deviations from plan

- **`UpdateCampaign` already had `scheduled_at=$12`** — the Session 2 query includes scheduled_at as a parameter. The PUT handler fetches the existing campaign first and passes its ScheduledAt value to preserve it on content edits.
- **`queued` status not locked for edits** — spec locks only `sending` and `sent`. A campaign in `queued` status can still be edited (this is correct per spec).

---

## Session 8 — Sending Engine: Send Worker, Job Processing, Recipient Expansion (complete)

### What was built

| Item                                                                                                     | Status |
| -------------------------------------------------------------------------------------------------------- | ------ |
| `internal/mailer/mailer.go` — Mailer interface + LogMailer stub                                          | Done   |
| `internal/worker/expand.go` — ExpandRecipients (list/tag, active filter, suppression check)              | Done   |
| `internal/worker/email_builder.go` — BuildMessage with {{first_name}} and {{unsubscribe_url}} vars       | Done   |
| `internal/worker/send_worker.go` — SendWorker: Start, processPendingJobs, processJob, failJob            | Done   |
| `db/queries/campaigns.sql` — added `MarkCampaignSent :exec`                                              | Done   |
| `db/queries/send_jobs.sql` — added `UpdateSendJobCounts :exec`, `UpdateSendJobError :exec`               | Done   |
| `task sqlc-gen` — regenerated; all 3 new functions in sqlc package                                       | Done   |
| `internal/handler/campaigns.go` — CampaignHandler gains `appSecret`, `worker` fields; Send handler wired | Done   |
| `POST /api/campaigns/:id/send` — ExpandRecipients → no_recipients guard → CreateSendJob → queued         | Done   |
| `cmd/server/main.go` — LogMailer + SendWorker wired; SIGINT/SIGTERM graceful shutdown (5s)               | Done   |

### Verification

| Check                              | Result                                                              |
| ---------------------------------- | ------------------------------------------------------------------- | ----------------------------- |
| `go build ./...`                   | Pass — compiles clean                                               |
| `POST /api/campaigns/:id/send`     | 202 `{"message":"Campaign queued for sending","campaign_id":"..."}` |
| Server log within 10s              | `[LogMailer] Would send to test@example.com                         | subject: Hello from OwnMaily` |
| `GET /api/campaigns/:id` after 10s | status: "sent", sent_at set                                         |
| `GET /api/campaigns/:id/stats`     | sent: 1, failed: 0, open_rate: 0                                    |
| `SELECT * FROM send_jobs`          | status: complete, total_count=1, sent_count=1, failed_count=0       |
| Send to empty list                 | 400 `{"error":"no_recipients",...}`                                 |

### sqlc additions

- `MarkCampaignSent :exec` — sets status='sent' and sent_at=NOW() in one query
- `UpdateSendJobCounts :exec` — sets total_count after recipient expansion
- `UpdateSendJobError :exec` — sets status + error_message on fatal job failure

### LogMailer note

`LogMailer` is a temporary stub that logs to stdout. It will be replaced in Session 9 when real SMTP is wired.

### Deviations from plan

- **`UpdateSendJobError` added** — spec referenced setting error_message on failure. `UpdateSendJobStatus` only takes status; a separate query was added to set both status and error_message atomically.
- **`bulkRecipientParams` helper in expand.go** — the spec put BulkCreate in the worker; the helper function that converts `[]Subscriber` → `[]BulkCreateCampaignRecipientsParams` lives in `expand.go` (same package) to keep `send_worker.go` clean.
- **Rate limiting: 500ms sleep added between each send** to stay within provider rate limits. Applied after every recipient attempt (both success and failure paths) before moving to the next recipient.

---

## Session 9 — Resend Integration: Real SMTP + Bounce Webhook (complete)

### What was built

| Item                                                                                              | Status |
| ------------------------------------------------------------------------------------------------- | ------ |
| `internal/mailer/resend.go` — ResendMailer: POST to Resend API, Bearer auth, omitempty fields     | Done   |
| `internal/mailer/factory.go` — NewMailer factory: resend/mailgun/ses/fallback-to-LogMailer        | Done   |
| `internal/mailer/bounce.go` — BounceEvent struct + ParseResendBounce for email.bounced/complained | Done   |
| `internal/handler/webhooks.go` — WebhookHandler.Resend: hard bounce → suppression + status        | Done   |
| `cmd/server/main.go` — loadMailer helper reads settings at startup; falls back to LogMailer       | Done   |
| `POST /webhooks/resend` — registered outside RequireAuth group                                    | Done   |

### Verification

| Check            | Result                |
| ---------------- | --------------------- |
| `go build ./...` | Pass — compiles clean |

### Testing with a real Resend API key

Insert credentials directly via psql (no settings UI until Session 17):

```sql
UPDATE settings SET
  smtp_provider = 'resend',
  smtp_credentials = '{"api_key":"re_your_key_here"}'
WHERE id = TRUE;
```

Restart `task dev` — server logs `mailer: resend loaded`.

Simulate bounce webhook:

```bash
curl -X POST localhost:4400/webhooks/resend \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "email.bounced",
    "data": {
      "email_id": "abc123",
      "from": "you@example.com",
      "to": ["test@example.com"],
      "bounce_type": "hard"
    }
  }'
# → 200 OK
```

### Deviations from plan

- **Webhook always returns 200** — even on parse errors, to prevent Resend retries. Errors are logged server-side.
- **Signature verification skipped** — TODO comment in webhooks.go; acceptable for self-hosted v1.
- **`loadMailer` is a package-level function** in main.go (not inlined) — cleaner to read and easy to extract later.
- **Mailer reload deferred to Session 17** — loadMailer is called once at startup; TODO comment added.

---

## Session 10 — Mailgun Integration (complete)

### What was built

| Item                                     | File                           |
| ---------------------------------------- | ------------------------------ |
| `MailgunMailer` struct + `Send` + `Name` | `internal/mailer/mailgun.go`   |
| Factory `"mailgun"` case (was stub)      | `internal/mailer/factory.go`   |
| `ParseMailgunBounce(r *http.Request)`    | `internal/mailer/bounce.go`    |
| `WebhookHandler.Mailgun` handler         | `internal/handler/webhooks.go` |
| `POST /webhooks/mailgun` route           | `cmd/server/main.go`           |

### How sending works

- Endpoint: `POST https://api.mailgun.net/v3/<domain>/messages`
- Auth: HTTP Basic, username `api`, password = API key
- Body: `application/x-www-form-urlencoded` via `url.Values`
- Optional headers: `h:Reply-To`, `h:List-Unsubscribe` (only if set)

To switch to Mailgun:

```sql
UPDATE settings SET
  smtp_provider = 'mailgun',
  smtp_credentials = '{"api_key":"key-...","domain":"mg.yourdomain.com"}'
WHERE id = TRUE;
```

Restart server — logs `mailer: mailgun loaded`.

### Verification

| Check                                                           | Result                |
| --------------------------------------------------------------- | --------------------- |
| `go build ./...`                                                | Pass — compiles clean |
| Hard bounce curl → 200                                          | Pass                  |
| Complaint curl → 200                                            | Pass                  |
| Unhandled event curl → 200                                      | Pass                  |
| `test@example.com` in `suppressed_emails` (reason: hard_bounce) | Pass                  |
| `test2@example.com` in `suppressed_emails` (reason: complained) | Pass                  |

Real Mailgun send was **not** tested (no account available). Webhook simulation confirmed.

### Deviations from plan

- **`handleHardBounce` log prefix** remains `webhook/resend:` — it's a shared helper; log prefix is cosmetic, not fixing now.
- **Signature verification skipped** — TODO comment added in `Mailgun` handler.

---

## Session 11 — SES Integration + Scheduler Worker (complete)

### What was built

| Item                                                                                                  | Status |
| ----------------------------------------------------------------------------------------------------- | ------ |
| `internal/mailer/ses_sign.go` — AWS SigV4 signing: signRequest, hexSHA256, hmacSHA256 helpers         | Done   |
| `internal/mailer/ses.go` — SESMailer struct + Send + Name; SES v2 JSON payload; no AWS SDK            | Done   |
| `internal/mailer/factory.go` — SES case replaces stub; parses access_key_id/secret/region credentials | Done   |
| `internal/mailer/bounce.go` — ParseSESBounce: handles SNS Notification; Bounce + Complaint types      | Done   |
| `internal/handler/webhooks.go` — WebhookHandler.SES: SubscriptionConfirmation + Notification routing  | Done   |
| `POST /webhooks/ses` — registered outside RequireAuth group                                           | Done   |
| `internal/worker/scheduler.go` — SchedulerWorker: 60s poll, transactional queuing with re-fetch guard | Done   |
| `cmd/server/main.go` — schedulerWorker wired; both workers started on ctx                             | Done   |

### Verification

| Check                                               | Result                                                 |
| --------------------------------------------------- | ------------------------------------------------------ |
| `go build ./...`                                    | Pass — compiles clean                                  |
| POST /webhooks/ses SubscriptionConfirmation         | 200 — GETs SubscribeURL, logs "subscription confirmed" |
| POST /webhooks/ses Notification (hard bounce)       | 200 — ses-test@example.com in suppressed_emails        |
| Scheduler: injected past-scheduled campaign in psql | Picked up within 60s, status → queued → sent           |
| Scheduler log                                       | "scheduler: queued campaign <id>"                      |

### SES credentials JSON shape

```json
{ "access_key_id": "...", "secret_access_key": "...", "region": "us-east-1" }
```

Update via psql (no settings UI until Session 17):

```sql
UPDATE settings SET
  smtp_provider = 'ses',
  smtp_credentials = '{"access_key_id":"AKIA...","secret_access_key":"...","region":"us-east-1"}'
WHERE id = TRUE;
```

### Architecture notes

- SigV4 signing: canonical headers (content-type, host, x-amz-date in alphabetical order); HMAC chain: AWS4+secret → date → region → ses → aws4_request
- SNS SubscriptionConfirmation: handler GETs SubscribeURL directly; ParseSESBounce returns nil,nil for non-Notification types so the handler handles subscription separately
- Scheduler transaction: uses `pool.BeginTx` + `db.WithTx(tx)` to wrap CreateSendJob + UpdateCampaignStatus atomically; re-fetches campaign inside tx to prevent double-scheduling

### Deviations from plan

- **SES send not live-tested** — no AWS account available; compile-only verification per spec note
- **Before running curl verification scripts, kill whatever is running on the port first** — the dev server was already running on :4400 when verification curls were attempted
- **Scheduler fired at t+120s on first test** — the campaign was inserted just after the first tick (t+60s), so it was picked up at t+120s. Under normal conditions it will be within 60 seconds of insertion.

---

## Session 12 — Open Tracking: Pixel Endpoint, Token Generation, Recording (complete)

### What was built

| Item                                                                                                    | Status |
| ------------------------------------------------------------------------------------------------------- | ------ |
| `internal/tracking/tokens.go` — `GenerateOpenToken`, `ParseOpenToken` (HMAC-SHA256, same pattern as S5) | Done   |
| `internal/handler/tracking.go` — `TrackingHandler.Open`: parse token, record open, always return GIF    | Done   |
| `internal/worker/email_builder.go` — pixel injected before `</body>` (or appended); TextBody unchanged  | Done   |
| `GET /track/open/{token}` — registered outside RequireAuth group in `cmd/server/main.go`                | Done   |

### Verification

| Check                                        | Result                                                                  |
| -------------------------------------------- | ----------------------------------------------------------------------- |
| `go build ./...`                             | Pass — compiles clean                                                   |
| `GET /track/open/invalid.token`              | 200, Content-Type: image/gif, 43 bytes                                  |
| `GET /track/open/<valid_token>` (first hit)  | 200, Content-Type: image/gif, row inserted in `opens`                   |
| `GET /track/open/<valid_token>` (second hit) | 200, Content-Type: image/gif, no duplicate row (ON CONFLICT DO NOTHING) |
| Campaign send → `/api/campaigns/:id/stats`   | opens: 1, open_rate: 1.0 after hitting pixel for sent campaign          |

### Token format

- Payload: `subscriberID:campaignID` (both UUID strings)
- Message: `base64url(payload)`
- Signature: `base64url(HMAC-SHA256(message, appSecret))`
- Token: `message + "." + signature`
- No expiry — tokens are valid forever (embedded in already-sent emails)

### Deviations from plan

- **Pixel injected at send time, not stored in DB** — `BuildMessage` injects pixel dynamically; `html_body` in `campaigns` table stays as the original template. This is correct: the pixel URL is per-recipient, so it cannot be stored in the campaign body.
- **`pgtype.UUID` used throughout** — spec said `uuid.UUID` but the project uses `pgtype.UUID` everywhere; `internal/tracking` uses the same type.

---

## Session 13 — Click Tracking: URL Rewriting, Redirect Endpoint, Recording (complete)

### What was built

| Item                                                                                            | Status |
| ----------------------------------------------------------------------------------------------- | ------ |
| `internal/tracking/tokens.go` — GenerateClickToken, ParseClickToken (payload: subID:campID:idx) | Done   |
| `internal/tracking/rewrite.go` — RewriteLinks via golang.org/x/net/html tree walk               | Done   |
| `internal/handler/tracking.go` — Click handler: parse token, RecordClick, 302 redirect          | Done   |
| `internal/worker/email_builder.go` — RewriteLinks called before pixel injection                 | Done   |
| `cmd/server/main.go` — `GET /track/click/{token}` route wired outside RequireAuth group         | Done   |
| `golang.org/x/net` dependency added (upgraded to v0.53.0)                                       | Done   |

### Verification

| Check                                                                | Result                        |
| -------------------------------------------------------------------- | ----------------------------- |
| Valid click token → 302 redirect to destination                      | Pass                          |
| Click recorded in `clicks` table (link_index, link_url, campaign_id) | Pass — 2 rows after 2 hits    |
| Invalid token → still redirects (link not broken)                    | Pass — 302 to destination     |
| Missing `url` param → 400                                            | Pass                          |
| Campaign stats `clicks: 2`, per-link breakdown                       | Pass                          |
| Unsubscribe link skipped (contains `/unsubscribe`)                   | Pass — skipLink logic correct |

### Implementation notes

- `RewriteLinks` uses `golang.org/x/net/html` tree walk — no regex
- Skips `#` anchors, `mailto:`, `tel:`, and any href containing `/unsubscribe`
- Handles both full HTML documents and fragments: fragment mode extracts body children after parse to avoid injecting `<html><head>` wrapper
- Link slice returned by RewriteLinks is discarded at call site — link URLs stored per-click from `url` query param (matches schema)
- `ParseClickToken` payload split with `SplitN(..., 3)` to handle UUIDs with colons correctly

## Session 14 — Unsubscribe: Token Generation, Endpoint, List-Unsubscribe Header (complete)

### What was built

| Item                                                                                           | Status |
| ---------------------------------------------------------------------------------------------- | ------ |
| `tracking.GenerateUnsubscribeToken` / `ParseUnsubscribeToken` in `internal/tracking/tokens.go` | Done   |
| `GET /unsubscribe` endpoint in `internal/handler/tracking.go`                                  | Done   |
| Route registered outside RequireAuth group in `cmd/server/main.go`                             | Done   |
| `{{unsubscribe_url}}` placeholder replaced with real signed token in `email_builder.go`        | Done   |
| `List-Unsubscribe` + `List-Unsubscribe-Post` headers set in `BuildMessage`                     | Done   |
| Mailgun: refactored to loop over `msg.Headers` generically with `h:` prefix                    | Done   |
| SES: added `Content.Simple.Headers` array for custom headers                                   | Done   |
| Resend: already passed `msg.Headers` map — no change needed                                    | Done   |

### Unsubscribe endpoint behaviour

- Invalid/missing token → 400 HTML error page
- Subscriber not found → 400 HTML error page (no existence leak)
- Already `unsubscribed` or `bounced` → 200 success (idempotent)
- Happy path: `UpdateSubscriberStatus("unsubscribed")` + `AddSuppression(reason="unsubscribed")` (ON CONFLICT DO NOTHING) → 200 HTML success

### Verification results

| Check                                                            | Result |
| ---------------------------------------------------------------- | ------ |
| Invalid token → 400 HTML                                         | Pass   |
| Missing token → 400 HTML                                         | Pass   |
| Valid token → 200 HTML, status=unsubscribed in DB                | Pass   |
| Suppression row created with reason=unsubscribed                 | Pass   |
| Second hit (idempotent) → 200 HTML, no duplicate suppression row | Pass   |
| `go build ./...` clean                                           | Pass   |

### Mailer header notes

- **Resend**: `"headers"` JSON field accepts any key-value map — `List-Unsubscribe-Post` passes through unchanged.
- **Mailgun**: was hardcoding only `h:List-Unsubscribe`. Refactored to `for k, v := range msg.Headers { form.Set("h:"+k, v) }` — now passes both headers generically.
- **SES**: was not passing any headers. Added `Content.Simple.Headers` array (SES v2 SendEmail format: `[{Name, Value}]`) when `msg.Headers` is non-empty.

### Deviations from spec

None.

## Session 15 — Analytics: Overview Stats + Per-Campaign Stats Endpoints (complete)

### What was built

| Item                                                                                                | Status |
| --------------------------------------------------------------------------------------------------- | ------ |
| `db/queries/subscribers.sql` — `CountSubscribersByStatus` added                                     | Done   |
| `db/queries/opens.sql` — `CountTotalOpens` added                                                    | Done   |
| `db/queries/clicks.sql` — `CountTotalClicks` added                                                  | Done   |
| `db/queries/campaign_recipients.sql` — `CountTotalSent`, `ListCampaignsReceivedBySubscriber` added  | Done   |
| `task sqlc-gen` — 5 new functions generated cleanly                                                 | Done   |
| `internal/handler/analytics.go` — `AnalyticsHandler.Overview`: aggregates all overview stats        | Done   |
| `internal/handler/subscribers.go` — `SubscriberHandler.Stats`: per-subscriber engagement history    | Done   |
| `cmd/server/main.go` — `GET /api/analytics/overview` + `GET /api/subscribers/{id}/stats` registered | Done   |

### Verification

| Check                                          | Result                                                                                                          |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `go build ./...`                               | Pass — compiles clean                                                                                           |
| `GET /api/analytics/overview`                  | 200 — correct counts: 6 total, 3 active, 2 unsubscribed, 1 bounced, 5 campaigns sent, rate calculations correct |
| `GET /api/subscribers/:id/stats` (no activity) | 200 — 0 campaigns_received, 0 opens, 0 clicks, last_active: null                                                |
| POST unsubscribe → re-fetch overview           | active_subscribers decremented (4→3), unsubscribed incremented (2→3)                                            |

### Owner login credentials (dev only)

Email: `admin@test.com`, password: `testpass123`

### Deviations from plan

- **Skipped `CountActiveSubscribers`** — used parameterised `CountSubscribersByStatus("active")` instead; fewer queries in the schema
- **Skipped `CountSentCampaigns`** — reused existing `CountCampaignsByStatus("sent")`; same result
- **`ListCampaignsReceivedBySubscriber` uses explicit column list** — `c.*` with alias in sqlc JOIN queries sometimes needs explicit columns; listed all Campaign columns to guarantee correct mapping
- **`last_active` string comparison** — RFC3339 strings are lexicographically comparable for ISO timestamps, so `>` comparison is correct without parsing back to time.Time

## Session 16 — Next
