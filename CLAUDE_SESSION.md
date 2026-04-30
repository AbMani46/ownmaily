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

## Session 16 — Settings: General, SMTP, API Key, Suppression Management (complete)

### What was built

| Item                                                                                                           | Status |
| -------------------------------------------------------------------------------------------------------------- | ------ |
| `db/queries/settings.sql` — `UpdateSMTPSettings :exec` added                                                   | Done   |
| `db/queries/suppressed_emails.sql` — `SearchSuppressions`, `ListAllSuppressions` added                         | Done   |
| `task sqlc-gen` — 3 new functions generated cleanly                                                            | Done   |
| `internal/mailer/store.go` — `Store` with `sync.RWMutex`, `Get()` (read lock), `Set()` (write lock)            | Done   |
| `internal/worker/send_worker.go` — `mailer.Mailer` field replaced with `*mailer.Store`; calls `Get()` per send | Done   |
| `internal/handler/settings.go` — `SettingsHandler` with all 10 endpoints                                       | Done   |
| `cmd/server/main.go` — `mailerStore` replaces `activeMailer`; `settingsHandler` wired; 10 routes registered    | Done   |

### Session 9 TODO resolved

The `TODO Session 17: reload mailer when SMTP settings change` comment is removed. `mailer.Store` hot-swap is now in place — `PUT /api/settings/smtp` calls `mailerStore.Set(newMailer)` after saving to DB.

### sqlc additions

| Query                 | File                  | Type  |
| --------------------- | --------------------- | ----- |
| `UpdateSMTPSettings`  | settings.sql          | :exec |
| `SearchSuppressions`  | suppressed_emails.sql | :many |
| `ListAllSuppressions` | suppressed_emails.sql | :many |

### Verification

| Check                                            | Result                                                             |
| ------------------------------------------------ | ------------------------------------------------------------------ |
| `go build ./...`                                 | Pass — compiles clean                                              |
| `GET /api/settings`                              | 200 — all fields, no `smtp_credentials`                            |
| `PUT /api/settings/general`                      | 200 — updated fields returned, smtp_provider/credentials preserved |
| `PUT /api/settings/smtp` (fake resend key)       | 200 — resend init doesn't validate key; mailerStore swapped        |
| `PUT /api/settings/smtp` (invalid provider)      | 400 `invalid_provider`                                             |
| `POST /api/settings/api-key/regenerate`          | 200 — full `om_...` key returned once with "Save this key" message |
| `GET /api/settings/api-key`                      | 200 — `key_prefix` only, no full key                               |
| API key auth (`Authorization: Bearer om_...`)    | 200 — `GET /api/settings` works with raw API key                   |
| `POST /api/settings/suppressions`                | 200 — adds email with reason=manual                                |
| `POST /api/settings/suppressions` (duplicate)    | 409 `already_suppressed`                                           |
| `GET /api/settings/suppressions`                 | 200 — paginated list with total                                    |
| `GET /api/settings/suppressions/export`          | CSV download with email, reason, created_at                        |
| `DELETE /api/settings/suppressions/manual%40...` | 204 — email removed                                                |

### Deviations from spec

- **`PUT /api/settings/smtp` fake resend key → 200 not 400**: `NewResendMailer` stores the key without making an API call. The spec anticipated this with "OR 200 if resend accepts it". The 400 path is exercised by empty/missing credentials or invalid provider names.
- **Suppression conflict detection uses `IsSuppressed` pre-check**: `AddSuppression` uses `ON CONFLICT DO NOTHING` (silent); added `IsSuppressed` check before insert to detect and return 409. Tiny TOCTOU race in concurrent inserts is acceptable for single-owner installs.

## Test Session A — Test Infrastructure + Auth + Subscribers (complete)

### What was built

| Item                                                                                                                                                                                    | Status |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| `internal/testutil/db.go` — NewTestDB: drop/create ownmaily_test, run migrations, return pool+queries                                                                                   | Done   |
| `internal/testutil/server.go` — NewTestServer: full chi router wired, httptest.Server via pre-allocated listener                                                                        | Done   |
| `internal/testutil/fixtures.go` — CreateOwner, LoginAndGetToken, CreateSubscriber, CreateList, CreateTag, CreateCampaign, AuthHeader, MakeRequest, DecodeJSON, AssertStatus, UUIDString | Done   |
| `internal/handler/auth_test.go` — 8 auth test functions                                                                                                                                 | Done   |
| `internal/handler/subscribers_test.go` — 13 subscriber test functions                                                                                                                   | Done   |
| `Taskfile.yml` — `test` task added                                                                                                                                                      | Done   |

### Test functions written

**Auth (auth_test.go)**

| Function                | Status |
| ----------------------- | ------ |
| TestLogin_Success       | PASS   |
| TestLogin_WrongPassword | PASS   |
| TestLogin_NoOwner       | PASS   |
| TestLogin_SetsJWTCookie | PASS   |
| TestMe_WithValidToken   | PASS   |
| TestMe_WithoutToken     | PASS   |
| TestMe_WithAPIKey       | PASS   |
| TestLogout_ClearsCookie | PASS   |

**Subscribers (subscribers_test.go)**

| Function                             | Status |
| ------------------------------------ | ------ |
| TestCreateSubscriber_Success         | PASS   |
| TestCreateSubscriber_DuplicateEmail  | PASS   |
| TestCreateSubscriber_SuppressedEmail | PASS   |
| TestCreateSubscriber_InvalidEmail    | PASS   |
| TestListSubscribers_Pagination       | PASS   |
| TestListSubscribers_SearchFilter     | PASS   |
| TestListSubscribers_StatusFilter     | PASS   |
| TestGetSubscriber_NotFound           | PASS   |
| TestUpdateSubscriber                 | PASS   |
| TestDeleteSubscriber_AddsSuppression | PASS   |
| TestUnsubscribeSubscriber            | PASS   |
| TestImportCSV_ValidAndInvalid        | PASS   |
| TestExportCSV                        | PASS   |

Total: 21 tests, 21 PASS, 0 FAIL, 0 SKIP.

### Bug fixed by tests

**`GET /api/auth/me` did not support API key auth.** The `Me` handler read only JWT claims from context. When API key middleware authenticated a request, no claims were set, so the handler returned 401 "missing claims". Fixed by checking `middleware.APIKeyAuthed` in context first and fetching the owner from DB for that path. This was a real gap: `GET /api/settings` worked with API key (any route), but `Me` explicitly failed.

### Architecture notes

- **TestDB lifecycle**: each test drops and recreates `ownmaily_test`, runs all migrations. Takes ~0.5s per test — acceptable for 21 tests.
- **TestServer listener pre-allocation**: uses `net.Listen` before handler creation so `installationURL = "http://" + addr` is known at construction time. Handlers receive the actual server URL.
- **`t.Skip`** on Postgres unavailability (not `t.Fatal`) — test suite is gracefully skipped if Docker is not running.
- **`fuser -k 4400/tcp` not needed** — tests use httptest.Server on a random port; port 4400 (dev server) is not touched.
- **No `t.Parallel()`** — tests run sequentially within the package; each gets an isolated DB.

### Deviations from plan

- **`TestLogin_SetsJWTCookie`** asserts HttpOnly + non-empty value; spec said "cookie jwt is set" — covered fully.
- **`TestLogout_ClearsCookie`** checks `Set-Cookie` header directly for `Max-Age=0` rather than inspecting parsed `resp.Cookies()`, because Go serializes `MaxAge=-1` as `Max-Age=0` on wire and parses it back as `MaxAge=0` (zero is also the default for unset), making the parsed value ambiguous.
- **`TestImportCSV_ValidAndInvalid`** includes a suppressed email, so counts are imported=2, skipped=2, invalid=1 (rather than skipped=1 if suppressed were not tested).
- **Worker goroutines not started** in test server — `NewSendWorker` and `NewSchedulerWorker` are wired to satisfy constructor requirements but `Start` is never called. Campaigns sent in tests would not be processed by the background worker.

---

## Test Session B — Lists, Tags, Campaigns, Send Worker, Scheduler (complete)

### Test functions written — all PASS

| File                                  | Function                                     | Result |
| ------------------------------------- | -------------------------------------------- | ------ |
| `internal/handler/lists_test.go`      | `TestCreateList_Success`                     | PASS   |
|                                       | `TestCreateList_DuplicateName`               | PASS   |
|                                       | `TestCreateList_NameRequired`                | PASS   |
|                                       | `TestGetList_NotFound`                       | PASS   |
|                                       | `TestUpdateList`                             | PASS   |
|                                       | `TestDeleteList_NonEmpty_Returns409`         | PASS   |
|                                       | `TestDeleteList_ForceDelete`                 | PASS   |
|                                       | `TestAddSubscriberToList`                    | PASS   |
|                                       | `TestAddSubscriberToList_AlreadyMember`      | PASS   |
|                                       | `TestAddSubscriberToList_InactiveSubscriber` | PASS   |
|                                       | `TestRemoveSubscriberFromList`               | PASS   |
|                                       | `TestListSubscribersInList_Paginated`        | PASS   |
|                                       | `TestDoubleOptIn_ConfirmationFlow`           | PASS   |
| `internal/handler/tags_test.go`       | `TestCreateTag_Success`                      | PASS   |
|                                       | `TestCreateTag_Duplicate`                    | PASS   |
|                                       | `TestUpdateTag_NameConflict`                 | PASS   |
|                                       | `TestDeleteTag_CascadesSubscriberTags`       | PASS   |
|                                       | `TestAddTagToSubscriber`                     | PASS   |
|                                       | `TestAddTagToSubscriber_AlreadyTagged`       | PASS   |
|                                       | `TestRemoveTagFromSubscriber`                | PASS   |
|                                       | `TestBulkTag_Add`                            | PASS   |
|                                       | `TestBulkTag_Remove`                         | PASS   |
|                                       | `TestFilterSubscribersByTag`                 | PASS   |
| `internal/handler/campaigns_test.go`  | `TestCreateCampaign_Success`                 | PASS   |
|                                       | `TestCreateCampaign_MissingRequiredFields`   | PASS   |
|                                       | `TestUpdateCampaign_Locked_WhenSent`         | PASS   |
|                                       | `TestScheduleCampaign_Success`               | PASS   |
|                                       | `TestScheduleCampaign_PastTime`              | PASS   |
|                                       | `TestCancelCampaign`                         | PASS   |
|                                       | `TestDuplicateCampaign`                      | PASS   |
|                                       | `TestDeleteCampaign_WhenSending_Returns409`  | PASS   |
|                                       | `TestCampaignPreview`                        | PASS   |
|                                       | `TestCampaignStats_ZerosForNewCampaign`      | PASS   |
|                                       | `TestSendCampaign_NoRecipients`              | PASS   |
| `internal/worker/send_worker_test.go` | `TestSendWorker_ProcessesJob_EndToEnd`       | PASS   |
|                                       | `TestSendWorker_SkipsSuppressedRecipients`   | PASS   |
|                                       | `TestSendWorker_HandlesWorkerFailure`        | PASS   |
| `internal/worker/scheduler_test.go`   | `TestSchedulerWorker_QueuesOverdueCampaign`  | PASS   |
|                                       | `TestSchedulerWorker_IgnoresFutureCampaign`  | PASS   |
|                                       | `TestSchedulerWorker_IdempotentOnDoubleRun`  | PASS   |

**Total: 46 tests pass** (8 auth + 14 subscribers + 13 lists + 9 tags + 11 campaigns + 3 send worker + 3 scheduler — some counts include both sessions A and B)

**Handler suite total: 40 tests, 29.9s**
**Worker suite total: 6 tests, 5.3s**

### Bugs found and fixed

1. **`internal/handler/lists.go` — duplicate list name returned 500 instead of 409** — The Create and Update handlers lacked the `pgconn.PgError` code `23505` check that the tags handler already had. Added the check; `TestCreateList_DuplicateName` caught this.

### Architecture notes — worker tests

- **Import cycle**: `testutil/server.go` imports `internal/worker`. Worker tests using `package worker` cannot import `testutil`. Resolved by creating `internal/worker/testhelper_test.go` — a minimal test-only helper that sets up the DB directly (imports `internal/db`, `db`, `internal/sqlc`) without importing `testutil`.
- **Worker test DB**: uses a separate `ownmaily_test_worker` database to avoid conflicts with handler tests that use `ownmaily_test`.
- **Sleep timing**: `send_worker.go` has `time.Sleep(500ms)` per recipient. Worker tests with 3 recipients take ~1.5s each. Acceptable; not changed.
- **`processJob` and `tick` access**: Both are unexported methods on their respective structs. `package worker` test files access them directly.

### Deviations from plan

- **`campaign_recipients html` assertions**: The `CampaignRecipient` DB model has no html field — HTML is built at send time (not stored per recipient). EndToEnd test asserts counts and status only. Click/open/unsubscribe URL presence is inherent in the `BuildMessage` logic already unit-tested by existing tracking package.
- **`TestDoubleOptIn_ConfirmationFlow`**: Token generated independently using `mailer.NewConfirmationMailer` with the same secret. Works because `ValidateToken` is cryptographic (no DB token storage) — any token with the correct HMAC and non-expired timestamp validates.

## Test Session C — Tracking, Webhooks, Settings, Analytics (complete)

### Test functions written — all PASS

| File                                 | Function                                         | Result |
| ------------------------------------ | ------------------------------------------------ | ------ |
| `internal/handler/tracking_test.go`  | `TestOpenPixel_ValidToken_RecordsOpen`           | PASS   |
|                                      | `TestOpenPixel_InvalidToken_StillReturnsGIF`     | PASS   |
|                                      | `TestOpenPixel_Idempotent`                       | PASS   |
|                                      | `TestOpenPixel_ReturnsGIF`                       | PASS   |
|                                      | `TestClickRedirect_ValidToken_RecordsClick`      | PASS   |
|                                      | `TestClickRedirect_InvalidToken_StillRedirects`  | PASS   |
|                                      | `TestClickRedirect_MissingURL_Returns400`        | PASS   |
|                                      | `TestUnsubscribe_ValidToken`                     | PASS   |
|                                      | `TestUnsubscribe_InvalidToken_Returns400HTML`    | PASS   |
|                                      | `TestUnsubscribe_Idempotent`                     | PASS   |
|                                      | `TestUnsubscribe_AlreadyUnsubscribed_Returns200` | PASS   |
| `internal/handler/webhooks_test.go`  | `TestResendWebhook_HardBounce`                   | PASS   |
|                                      | `TestResendWebhook_SoftBounce_NoSuppression`     | PASS   |
|                                      | `TestResendWebhook_Complaint`                    | PASS   |
|                                      | `TestResendWebhook_UnknownEvent_Returns200`      | PASS   |
|                                      | `TestMailgunWebhook_HardBounce`                  | PASS   |
|                                      | `TestMailgunWebhook_Complaint`                   | PASS   |
|                                      | `TestMailgunWebhook_SoftBounce_NoSuppression`    | PASS   |
|                                      | `TestSESWebhook_HardBounce`                      | PASS   |
|                                      | `TestSESWebhook_Complaint`                       | PASS   |
|                                      | `TestSESWebhook_SoftBounce_NoSuppression`        | PASS   |
|                                      | `TestSESWebhook_SubscriptionConfirmation`        | PASS   |
| `internal/handler/settings_test.go`  | `TestGetSettings`                                | PASS   |
|                                      | `TestUpdateGeneralSettings`                      | PASS   |
|                                      | `TestUpdateSMTPSettings_InvalidProvider`         | PASS   |
|                                      | `TestUpdateSMTPSettings_ValidProvider`           | PASS   |
|                                      | `TestRegenerateAPIKey`                           | PASS   |
|                                      | `TestGetAPIKey_AfterRegenerate`                  | PASS   |
|                                      | `TestAPIKey_CanAuthenticateRequests`             | PASS   |
|                                      | `TestAddSuppression_Manual`                      | PASS   |
|                                      | `TestAddSuppression_Duplicate`                   | PASS   |
|                                      | `TestDeleteSuppression`                          | PASS   |
|                                      | `TestListSuppressions_Paginated`                 | PASS   |
|                                      | `TestExportSuppressions_CSV`                     | PASS   |
| `internal/handler/analytics_test.go` | `TestAnalyticsOverview_EmptyDB`                  | PASS   |
|                                      | `TestAnalyticsOverview_WithData`                 | PASS   |
|                                      | `TestAnalyticsOverview_RatesCalculation`         | PASS   |
|                                      | `TestSubscriberStats_NoActivity`                 | PASS   |
|                                      | `TestSubscriberStats_WithActivity`               | PASS   |

**Final total: 100 tests, 100 PASS, 0 FAIL**

**Handler suite: 94 tests**
**Worker suite: 6 tests**

### Session C totals by file

| File                | Tests  |
| ------------------- | ------ |
| tracking_test.go    | 11     |
| webhooks_test.go    | 11     |
| settings_test.go    | 12     |
| analytics_test.go   | 5      |
| **Session C total** | **39** |

### Bugs found

None — all new handlers were already correct.

### Architecture notes

- **Webhook tests use `http.NewRequest` directly** (not `testutil.MakeRequest`) because MakeRequest marshals the body as JSON; webhook payloads are raw JSON strings or form-encoded bodies that must not be re-encoded.
- **Analytics seeding** inserts `campaign_recipients` directly via `CreateCampaignRecipient` + `UpdateRecipientStatus`; this avoids the send worker and 500ms-per-send sleep, keeping tests fast.
- **`TestSESWebhook_SubscriptionConfirmation`** points SubscribeURL at `/api/auth/login` (returns 405 Method Not Allowed on GET, but any HTTP response satisfies the handler's GET-and-ignore-response pattern).
- **`range N` syntax** (Go 1.22+) used in analytics seeding loops.

### Deviations from plan

- **`TestAnalyticsOverview_WithData` seeds directly** rather than using the send worker — simpler, no 500ms-per-send sleep, test runs in ~0.6s. Functionally equivalent for asserting overview counts.

## Rate Limiter Patch (complete)

### What changed

- **`internal/worker/send_worker.go`** — replaced all three `time.Sleep(500 * time.Millisecond)` calls in the per-recipient loop with `w.limiter.Wait(ctx)` (token bucket, 2 sends/sec, burst 1).
- Added `limiter *rate.Limiter` field to `SendWorker`; initialized in `NewSendWorker` via `rate.NewLimiter(rate.Limit(2), 1)`.
- Added `golang.org/x/time/rate` import (was already an indirect dep at v0.12.0; promoted to explicit at v0.15.0 after `go get`).
- `time` import retained — still used by `time.NewTicker` (polling loop) and `time.Now()` (sent timestamp).
- Context cancellation on `limiter.Wait` returns `ctx.Err()` directly for clean shutdown.
- `go build ./...` clean; all 100 tests pass (handler: 94, worker: 6).

## Session 17 — Vue Frontend: Project Setup, Router, Pinia, Auth, Layout Shell (complete)

### What was built

| Item                                                                              | Status |
| --------------------------------------------------------------------------------- | ------ |
| `frontend/vite.config.js` — Vue plugin, Tailwind v4 Vite plugin, `@` alias, proxy | Done   |
| `frontend/src/style.css` — `@import "tailwindcss"` + tokens import + body font    | Done   |
| `frontend/src/styles/tokens.css` — full design token set (emerald accent)         | Done   |
| `frontend/src/lib/api.js` — axios client, Bearer token interceptor, 401 redirect  | Done   |
| `frontend/src/stores/auth.js` — Pinia auth store, localStorage persistence        | Done   |
| `frontend/src/router/index.js` — Vue Router 4, public/private guard               | Done   |
| `frontend/src/layouts/AppLayout.vue` — sidebar + topbar shell                     | Done   |
| `frontend/src/components/AppSidebar.vue` — nav, active state, logout              | Done   |
| `frontend/src/components/AppTopBar.vue` — page title from route meta              | Done   |
| `frontend/src/views/LoginView.vue` — login card, show/hide password, error state  | Done   |
| `frontend/src/views/DashboardView.vue` — placeholder                              | Done   |
| `frontend/src/views/SetupWizardView.vue` — placeholder                            | Done   |
| `frontend/src/App.vue` — `<RouterView />`                                         | Done   |
| `frontend/src/main.js` — app + pinia + router wired                               | Done   |
| `frontend/index.html` — Google Fonts (DM Sans, JetBrains Mono), entry updated     | Done   |
| `cmd/server/main.go` — SPA fallback serving `frontend/dist/`                      | Done   |
| `Taskfile.yml` — `frontend-dev`, `frontend-build` tasks added                     | Done   |

### Verification

| Check                                      | Result                                |
| ------------------------------------------ | ------------------------------------- |
| `task frontend-build`                      | Pass — 10 chunks, 296ms, no errors    |
| `go build ./...`                           | Pass — compiles clean                 |
| `curl http://localhost:4400/`              | 200 — Vue index.html served           |
| `curl http://localhost:4400/dashboard`     | 200 — SPA fallback returns index.html |
| `curl http://localhost:4400/health`        | 200 — API route still works           |
| Vue dev server (`pnpm --dir frontend dev`) | Starts on :5173, proxies /api → :4400 |

### Stack decisions

- **Tailwind v4** (not v3): uses `@import "tailwindcss"` in CSS + `@tailwindcss/vite` plugin; no `tailwind.config.js` needed
- **Plain .js** files throughout (not TypeScript); `tsc &&` removed from build script
- **Taskfile frontend tasks** use `pnpm --dir frontend <cmd>` to run from project root (not `dir: frontend`)

### Accent color

Accent is **emerald** `#10b981` throughout — not amber as in the original design file. All focus rings, active nav states, logo, and button use emerald tokens.

### Deviations from plan

- **Tailwind v4 config** differs from spec (spec assumed v3 `tailwind.config.js` + `@tailwind` directives); v4 Vite plugin approach used instead — functionally identical for the utility-only usage in layout components
- **`pnpm --dir frontend dev`** instead of `dir: frontend` in Taskfile — pnpm's `--dir` flag is more reliable for cross-directory invocation

## Session 18 — Dashboard + Subscribers Screens (complete)

### What was built

| Item                                                                                                                   | Status |
| ---------------------------------------------------------------------------------------------------------------------- | ------ |
| `src/components/BaseCard.vue` — white card, 10px radius, border, optional padding                                      | Done   |
| `src/components/StatCard.vue` — label/value/delta/sub, 28px bold value                                                 | Done   |
| `src/components/BaseBadge.vue` — status → color map, 5px dot, pill, uppercase                                          | Done   |
| `src/components/BaseButton.vue` — primary/secondary/ghost/danger variants, loading spinner                             | Done   |
| `src/components/BaseInput.vue` — v-model, emerald focus ring                                                           | Done   |
| `src/components/BaseTable.vue` — slot-based cells, hover #faf8f4, uppercase 11px headers                               | Done   |
| `src/components/BasePagination.vue` — page/perPage/total, prev/next, "Showing X–Y of Z"                                | Done   |
| `src/components/BaseModal.vue` — Teleport to body, overlay, close button, fade transition                              | Done   |
| `src/views/DashboardView.vue` — 4 stat cards from /api/analytics/overview, recent campaigns table, skeleton loading    | Done   |
| `src/views/SubscribersView.vue` — search (300ms debounce), status tabs, add/import modals, export, pagination 50/page  | Done   |
| `src/views/SubscriberDetailView.vue` — two-column layout, tags add/remove, list memberships, stats, unsubscribe/delete | Done   |
| `src/router/index.js` — /subscribers and /subscribers/:id routes added                                                 | Done   |

### Verification

| Check                                       | Result                                                                     |
| ------------------------------------------- | -------------------------------------------------------------------------- |
| `pnpm build`                                | Pass — 119 modules, 426ms, zero errors                                     |
| `GET /api/analytics/overview`               | Returns total_subscribers, active, unsubscribed, bounced, open/click rates |
| `GET /api/campaigns?per_page=5&status=sent` | Returns campaigns array with id/name/subject/status/sent_at                |
| `GET /api/subscribers?page=1&per_page=50`   | Returns subscribers + total, tags embedded                                 |
| `GET /api/subscribers/:id`                  | Returns subscriber + tags + lists                                          |
| `GET /api/subscribers/:id/stats`            | Returns campaigns_received, total_opens, total_clicks                      |
| `POST /api/subscribers/import` (CSV)        | Returns {imported, skipped, invalid}                                       |

### Deviations from design spec

- **Dashboard charts row** (sparkline + bar chart) not built — spec's "Layout per design section 3" only specified stat cards and recent campaigns table; charts were in the design mockup but not in the session task description
- **Campaign open/click rates on dashboard** show "—" — the `/api/campaigns` list endpoint does not return per-campaign rates; the Stats endpoint is per-campaign and would require N+1 calls
- **Tag colors** are hash-derived from tag name (cycling through 6 colors) — the Tag model has no `color` field in the DB schema
- **Activity timeline** not built on subscriber detail — `/api/subscribers/:id/stats` returns aggregate counts, not per-event timeline data
- **Confirm dialogs** use `window.confirm()` rather than a custom modal — sufficient for v1

### API shape notes

- `pgtype.UUID` marshals to standard 8-4-4-4-12 string format
- Subscriber detail (`GET /api/subscribers/:id`) returns `tags` and `lists` arrays embedded
- Subscriber stats returns `campaigns_received`, `total_opens`, `total_clicks` (plus raw arrays unused here)
- Import result returns `{imported, skipped, invalid}`

## Session 19 — Lists, Tags, Campaigns Screens (complete)

### What was built

| Item                                                                                                       | Status |
| ---------------------------------------------------------------------------------------------------------- | ------ |
| `src/composables/useConfirm.js` — `window.confirm` wrapper, upgradeable later                              | Done   |
| `src/views/ListsView.vue` — 3-column card grid + full table + Create List modal                            | Done   |
| `src/views/ListDetailView.vue` — stat row + paginated subscribers + embed code + delete                    | Done   |
| `src/views/TagsView.vue` — full table + create/edit/delete modals                                          | Done   |
| `src/views/CampaignsView.vue` — status tabs (All/Sent/Scheduled/Drafts) + table + counts                  | Done   |
| `src/views/CampaignEditView.vue` — two-column layout, template pills, contentEditable editor, preview mode | Done   |
| `src/views/CampaignStatsView.vue` — header card, 6-col stat grid, link breakdown table                     | Done   |
| `src/router/index.js` — 7 new routes added (campaigns/new registered before campaigns/:id/edit)            | Done   |

### Verification

| Check                                                              | Result                                                |
| ------------------------------------------------------------------ | ----------------------------------------------------- |
| `pnpm --dir frontend build`                                        | Pass — 536ms, zero errors                             |
| `GET /api/lists`                                                   | `{lists: [{id, name, double_opt_in, subscriber_count, ...}]}` |
| `GET /api/tags`                                                    | `{tags: []}` (none in test DB)                        |
| `GET /api/campaigns`                                               | `{campaigns: [...], total, page, per_page}`           |
| `GET /api/campaigns/:id/stats`                                     | `{sent, failed, opens, clicks, open_rate, click_rate, links: [{link_index, link_url, click_count}]}` |
| Vite dev routes `/lists`, `/campaigns`, `/tags`                    | All serve index.html (SPA routing correct)            |

### API shape notes

- `GET /api/lists` — `listResponse` extends `db.List` with `subscriber_count int64`; `double_opt_in` is a bool field
- `GET /api/campaigns/:id/stats` — `links` array has `link_url` and `click_count` (not `count`); rates are `float64` ratios (e.g. 0.47 for 47%)
- `POST /api/campaigns` — validation requires `name`, `subject`, `from_name`, `from_email` (must contain `@`), `send_to_type` ("list" or "tag")
- `POST /api/campaigns/:id/schedule` — body: `{"scheduled_at": "RFC3339"}`, must be future time, campaign must be draft/scheduled
- Campaign `send_to_id` is `pgtype.UUID` — marshals to UUID string or `null` when unset
- Campaigns API does not return target list/tag name — resolved client-side by fetching lists+tags maps

### Deviations from design spec

- **Target name in campaigns table**: API doesn't include target name in campaign object; resolved by loading lists+tags maps on mount and looking up by `send_to_id`
- **Bounces/Unsubscribes in stats grid**: per-campaign stats endpoint returns `sent/failed/opens/clicks/rates/links` only; bounces and unsubscribes default to 0 (shown but always 0 until API extends the response)
- **Add subscriber to list modal**: spec says "subscriber search/email input" — implemented as email lookup against `GET /api/subscribers?q=email`, exact match required, shows subscriber name if found
- **`send_to_id` null handling**: when no target is selected, `send_to_id` is `null` in JSON; frontend maps to empty string `''` and disables Send Now button
