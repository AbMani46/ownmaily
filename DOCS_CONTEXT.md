# OwnMaily — Documentation Context

Technical reference for docs writers. Covers every self-hosting concern in detail. Not user-facing; be thorough.

---

## Table of Contents

1. [Environment Variables](#1-environment-variables)
2. [Installation URL](#2-installation-url)
3. [SMTP Setup](#3-smtp-setup)
4. [Docker Compose Setup](#4-docker-compose-setup)
5. [Image Uploads](#5-image-uploads)
6. [Railway and Render Deployment](#6-railway-and-render-deployment)
7. [Setup Wizard Flow](#7-setup-wizard-flow)
8. [Known Gotchas](#8-known-gotchas)

---

## 1. Environment Variables

Loaded in `internal/config/config.go` via `os.Getenv`. The app calls `godotenv.Load()` at startup so a `.env` file in the working directory is sourced automatically — but plain env vars (e.g., set via Docker Compose `environment:` or a platform dashboard) work the same way and take precedence.

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_SECRET` | **Required** | — | Secret used to sign JWTs and HMAC tracking tokens. Generate with `openssl rand -hex 32`. Minimum 32 characters. If missing, the server refuses to start. |
| `DB_URL` | **Required** | — | Full Postgres connection string. Format: `postgres://user:password@host:5432/dbname?sslmode=disable`. If missing, the server refuses to start. |
| `INSTALLATION_URL` | Optional | `http://localhost:4400` | Public base URL of your OwnMaily instance. Used for tracking pixels, click URLs, unsubscribe links, double opt-in confirmation links, and image absolutization. **This must be set to a real public URL for production.** See Section 2. |
| `PORT` | Optional | `4400` | TCP port the Go HTTP server listens on inside the container. The Docker Compose external port mapping is `${PORT:-4400}:4400`, so changing `PORT` in `.env` only changes the external port, not the internal one (which is always 4400). |
| `UPLOAD_DIR` | Optional | `./uploads` | Filesystem directory where uploaded images are stored. Inside the Docker container this resolves to `/app/uploads`. The images subdirectory is `/app/uploads/images/`. |
| `POSTGRES_PASSWORD` | Optional* | — | Password for the Postgres `ownmaily` user. Used by `docker-compose.yml` to set `POSTGRES_PASSWORD` on the `db` service and construct `DB_URL`. Not read by the Go application directly — it uses `DB_URL`. The install script sets this to the fixed string `ownmaily`. |
| `DOCKER_IMAGE` | Optional | `knobbly007/ownmaily:latest` | Docker image to pull for the `app` service. Override when self-building: `DOCKER_IMAGE=myregistry/myimage:tag`. |

### Notes

- **`APP_SECRET` rotation**: Rotating `APP_SECRET` invalidates all active sessions and all outstanding tracking tokens (open pixels, click links, unsubscribe links) for already-sent campaigns. Those links will return 403/invalid for recipients who click after the rotation.
- **`DB_URL` with SSL**: The install script uses `?sslmode=disable` because the database is internal to the Docker network and unreachable externally. For external managed Postgres (Railway, Render, Neon, etc.), set `sslmode=require` or `sslmode=verify-full`.
- **`.env` file permissions**: The install script writes `.env` with `chmod 600`. On manual installs, enforce the same restriction — the file contains the app secret and database password.

---

## 2. Installation URL

### What it is

`INSTALLATION_URL` is the public base URL of the OwnMaily server — the root your subscribers will reach in their browser. It has no trailing slash.

Example values:
- `http://localhost:4400` (local testing only)
- `http://203.0.113.42:4400` (bare IP, testing on a VPS)
- `https://mail.yourdomain.com` (production)

### Why it matters

Every link OwnMaily inserts into outgoing emails is built from this URL. These include:

| Link type | Format | Impact if wrong |
|---|---|---|
| Open tracking pixel | `<installationURL>/track/open/<token>` | Opens not recorded; open rate shows 0% |
| Click tracking | `<installationURL>/track/click/<token>` | Clicks not recorded; link redirects break |
| Unsubscribe URL | `<installationURL>/unsubscribe?token=<token>` | Unsubscribe links break (CAN-SPAM violation) |
| Double opt-in confirm | `<installationURL>/confirm?token=<token>` | New subscribers can't confirm; pending forever |
| Uploaded images | `<installationURL>/uploads/images/<file>` | Images don't load in email clients |

### Where it's read

There are two sources in priority order:

1. **Settings database** (`settings.installation_url`) — set during the setup wizard (Step 3: General Settings) and editable at any time via Settings > General. This is what gets used when sending.
2. **`INSTALLATION_URL` env var** — the startup fallback. The send worker falls back to this only if the database read fails or the stored value is empty.

The send worker re-reads the value from the database at the start of each campaign send job (one DB read per campaign, not per recipient). This means changing the URL in Settings takes effect on the next campaign send without a container restart.

### How to set it correctly

**Step 1** — Set it in the `.env` file before first launch:
```
INSTALLATION_URL=https://mail.yourdomain.com
```

**Step 2** — The setup wizard (Step 3: General Settings) pre-fills the Installation URL field from `window.location.origin`. If you open the wizard at `https://mail.yourdomain.com/setup`, it will pre-fill the correct value. If you access it via a local port forward or SSH tunnel, the pre-fill will be wrong — correct it manually.

**Step 3** — After wizard completion, the value is stored in the database. You can update it later in Settings > General.

**HTTPS requirement**: Use `https://` only if SSL is actually configured in front of the container (reverse proxy, Cloudflare, platform CDN, etc.). The Go app itself speaks plain HTTP — TLS termination must happen upstream. Using `https://` when traffic is plain HTTP will cause mixed-content errors in browsers.

### What happens if it's wrong after sending

- Past campaigns: existing tracking tokens are already built with the old URL and won't decode correctly with a new URL (they're HMAC-signed). Links in already-sent emails break.
- New campaigns: once you update the URL in Settings, future campaigns use the new URL.

---

## 3. SMTP Setup

OwnMaily does not ship a built-in mail server. You must connect one of three supported providers. The provider is configured in the setup wizard (Step 4) or in Settings > SMTP.

### How credentials are stored

Credentials are stored as JSON in the `settings.smtp_credentials` (JSONB) column. The provider name is stored in `settings.smtp_provider`. From name and from email are stored in campaign rows (copied at draft-save time) and also in `settings.from_name` / `settings.from_email`.

At startup, the app reads credentials from the database and constructs a mailer. If SMTP is not configured or credentials are invalid, it falls back to `LogMailer`, which logs email sends to stdout but doesn't actually deliver them.

### Provider: Resend

**Credentials needed:**
- `api_key`: your Resend API key (starts with `re_`)

**Domain requirement:** The from email's domain must be verified in your Resend dashboard under Domains. Example: if you want to send from `newsletter@yourdomain.com`, add and verify `yourdomain.com` in Resend. Using an unverified domain returns a `403` error.

**Regions:** Resend has one global API endpoint (`https://api.resend.com/emails`). No region selection needed.

**Webhook (bounce handling):** Register `POST /webhooks/resend` in your Resend dashboard under Webhooks. Events: `email.bounced`, `email.complained`.

### Provider: Mailgun

**Credentials needed:**
- `api_key`: Mailgun private API key
- `domain`: your Mailgun sending domain (e.g., `mg.yourdomain.com`)
- `region`: `US` or `EU` — controls which API endpoint is used (`api.mailgun.net` vs `api.eu.mailgun.net`)

**Domain requirement:** Add and verify the sending domain in Mailgun under Sending > Domains. The from email must use this domain.

**Webhook (bounce handling):** Register `POST /webhooks/mailgun` in Mailgun under Webhooks. All events are parsed; bounces and complaints update the suppression list.

### Provider: Amazon SES

**Credentials needed:**
- `access_key_id`: AWS IAM access key ID
- `secret_access_key`: AWS IAM secret access key
- `region`: AWS region (choices in the wizard: `us-east-1`, `us-west-2`, `eu-west-1`, `ap-southeast-1`)

**Sandbox mode:** New SES accounts start in sandbox mode. In sandbox, you can only send from and to verified email addresses. You must request production access from AWS to send to arbitrary recipients.

**IAM permissions required:** `ses:SendEmail` (and `ses:SendRawEmail` if applicable) on the sending identity resource.

**Domain requirement:** Verify your sending domain (or individual from-address) in the SES console under Verified identities.

**Webhook (bounce handling):** Register `POST /webhooks/ses` via SNS. SES sends bounce/complaint notifications via SNS → HTTP endpoint.

**Request signing:** OwnMaily signs SES API requests with AWS Signature Version 4 in `internal/mailer/ses_sign.go`. No SDK dependency.

### From email domain — most common setup mistake

The from email's domain must be verified with your SMTP provider before any email can be sent. This applies to all three providers. Unverified domains cause sends to fail at the provider level — either with an explicit error (Resend returns 403, Mailgun returns 401/domain error) or silently depending on the provider's behavior. The OwnMaily UI will show the send as failed or the campaign may appear stuck in "sending" with delivery errors in the container logs.

**The domain in the from email and the domain verified with your provider must match exactly.** If you verified `yourdomain.com` in Resend but set the from email to `newsletter@mail.yourdomain.com`, the send will fail — `mail.yourdomain.com` is a different domain that also needs verification.

### Where "From" is configured

- **From name** and **From email** are set in the SMTP wizard step (Step 4) and saved with `PUT /api/setup/smtp` or `PUT /api/settings/smtp`.
- These values are also saved **per campaign** at draft-save time. If you change them in Settings, existing saved drafts retain the old values. New campaigns pick up the new Settings values.
- From name and from email are shown in the Send confirmation modal so the user can verify before sending.

### What the wizard does vs. what Settings does

| Action | Wizard (Step 4) | Settings > SMTP |
|---|---|---|
| Set provider | Yes | Yes |
| Set API credentials | Yes | Yes |
| Set from name / from email | Yes | Yes |
| Test connection | Yes (`POST /api/setup/test-smtp`) | Yes (`POST /api/settings/smtp/test`) |
| Save without testing | Yes (click Continue) | Yes (click Save) |
| Skip entirely | Yes (Skip for now) | N/A |

The test sends a real email to the owner's email address (wizard) or whatever address is entered (Settings test). The test call first saves credentials via `PUT`, then fires a test email via `POST /test-smtp`.

---

## 4. Docker Compose Setup

### Services

**`app` service:**
- Image: `${DOCKER_IMAGE:-knobbly007/ownmaily:latest}`
- Internal port: 4400 (always)
- External port: `${PORT:-4400}` (configurable via `.env`)
- Restart policy: `unless-stopped`
- Waits for `db` to pass health check before starting
- Env vars passed in: `DB_URL`, `APP_SECRET`, `INSTALLATION_URL`, `PORT=4400`
- `UPLOAD_DIR` is not passed explicitly — defaults to `./uploads` inside the container, which resolves to `/app/uploads`

**`db` service:**
- Image: `postgres:18-alpine`
- No external port (unreachable from outside the Docker network)
- Restart policy: `unless-stopped`
- Health check: `pg_isready -U ownmaily` every 5s, 5 retries
- Postgres credentials: user `ownmaily`, password `${POSTGRES_PASSWORD}`, database `ownmaily`

### Volumes

| Volume name | Mounted at | Contents | Persistent? |
|---|---|---|---|
| `postgres_data` | `/var/lib/postgresql` (inside `db`) | All database files | Yes — named volume |
| `uploads_data` | `/app/uploads` (inside `app`) | Uploaded images | Yes — named volume |

**Critical note on Postgres mount path:** The volume is mounted at `/var/lib/postgresql`, not `/var/lib/postgresql/data`. Postgres 18 alpine creates its data directory as `/var/lib/postgresql/data` — a subdirectory of the mount point. This is intentional and correct. Do not change this path.

### What persists across restarts

- All database data (subscribers, campaigns, settings, etc.) — persisted in `postgres_data`
- Uploaded images — persisted in `uploads_data`

### What does NOT persist

- Application logs (stdout only)
- Nothing else — OwnMaily is fully stateless at the application layer

### Updating

```bash
cd ~/ownmaily
docker compose pull
docker compose up -d
```

Migrations run automatically at startup via `db.RunMigrations`. Downtime is limited to the container restart window (typically a few seconds).

### Uninstalling (destructive)

```bash
cd ~/ownmaily
docker compose down -v   # -v destroys named volumes — all data is gone
```

Without `-v`, `docker compose down` stops and removes containers but leaves volumes intact (data preserved).

### Ports

The default external port is 4400. To change it, set `PORT=8080` in `.env`. Only the external port changes — the internal container port stays 4400 because the app's `environment:` block hardcodes `PORT=4400`.

To put OwnMaily behind a reverse proxy (nginx, Caddy, Traefik) and serve on port 80/443:
1. Remove or remap the port mapping in `docker-compose.yml`
2. Configure the reverse proxy to forward to `localhost:4400`
3. Set `INSTALLATION_URL` to the public HTTPS URL

### Self-building

Set `DOCKER_IMAGE=myregistry/myimage:tag` in `.env` and run `task docker-push` (or `docker build` + `docker push` manually). The Dockerfile is a two-stage build: Go + Node (pnpm) builder stage, then a minimal Alpine runtime image.

---

## 5. Image Uploads

### Upload endpoint

`POST /api/uploads/images` — requires authentication (JWT or API key).

Form field: `file` (multipart).

### Constraints

- **Max size:** 5 MB hard limit enforced at the HTTP layer (`r.ParseMultipartForm(5 << 20)`). Requests larger than 5 MB return `413 Request Entity Too Large`.
- **Accepted formats:** JPEG, PNG, GIF, WebP.
- **Format detection:** The server sniffs the first 512 bytes of the uploaded file using Go's `http.DetectContentType` and ignores the client-supplied `Content-Type` header. A file renamed to `.png` that is actually a GIF will be detected as GIF and saved with `.gif` extension.
- **No in-browser editing:** OwnMaily does not offer image cropping, resizing, or any editing. Prepare images to the correct dimensions and file size in an external tool before uploading.

### Where images are stored

Inside the container: `$UPLOAD_DIR/images/<hex>.<ext>` — defaults to `/app/uploads/images/<32-char-hex>.<ext>`.

The filename is a randomly generated 32-character hex string. No original filename is preserved.

### How images are served

`GET /uploads/*` is served publicly (no auth required) via `http.FileServer`. This path is intentionally unauthenticated so email clients (Gmail, Apple Mail, etc.) can fetch images when rendering emails.

The `/uploads/` path prefix is in the `setupGuard` skip list, so it works even before setup is complete.

### How image URLs work

The image upload endpoint returns a **root-relative URL**: `/uploads/images/<filename>`.

This URL is what gets stored in the campaign's `html_body` in the database. Absolute URLs are NOT stored.

At send time, `BuildMessage` in `internal/worker/email_builder.go` absolutizes all `/uploads/...` paths using a regex before sending:
```go
var relativeImgSrc = regexp.MustCompile(`(<img\s[^>]*src=")(/uploads/[^"]+)(")`)
```

It replaces `/uploads/...` with `<installationURL>/uploads/...` using the current `installation_url` from the settings database.

This design means:
- **Changing `INSTALLATION_URL` does not break stored campaigns.** The stored path is always relative; the absolute URL is built at send time from the current settings value.
- **Images in the preview pane** in the campaign editor are absolutized client-side using `installationURL` from the settings API response, so they render correctly even when the Vite dev server runs on a different port than the Go backend.

### What happens if `INSTALLATION_URL` changes

- **Old sent campaigns:** links in already-delivered emails still point to the old URL. Images break if the old domain/IP is no longer reachable.
- **New campaigns:** next send uses the new URL; images load correctly.
- **Stored HTML:** unchanged (root-relative paths). Only the send-time behavior changes.

### Cloud platforms with ephemeral filesystems

On platforms like Railway (free tier) and Render (free web services), the local filesystem is wiped on every deploy. Uploaded images stored under `UPLOAD_DIR` will be permanently lost after each deploy. This is a known limitation for v1 — the `Storage` interface in `internal/handler/uploads.go` is a seam for a future R2/S3 backend, but currently only `LocalStorage` is implemented. For anyone using image uploads in campaigns, self-hosting on a VPS with Docker Compose is the recommended approach.

---

## 6. Railway and Render Deployment

Railway and Render one-click deploy are deferred to post-launch. For v1, Docker Compose is the only supported deployment method. If there is user demand after launch, Railway/Render config files (`railway.toml`, `render.yaml`) will be added at that point.

---

## 7. Setup Wizard Flow

The wizard runs automatically on first visit if `settings.setup_complete = false`. The `setupGuard` middleware redirects all non-API, non-tracking, non-asset paths to `/setup` until setup is marked complete.

### Steps

**Step 1 — Welcome**
- Display only. Explains prerequisites: a domain, an SMTP API key, ~5 minutes.
- No API calls. Always can proceed.

**Step 2 — Owner Account**
- Fields: email address, password, confirm password.
- **Cannot be skipped.** This is the only admin account — there's no way to add more users later.
- Validation: passwords must match (checked client-side and server-side).
- API: `POST /api/setup/owner` — creates the owner row and sets the password hash.
- After success: the owner's email is pre-filled into the test email field on Step 5.

**Step 3 — General Settings**
- Fields: site name, installation URL, timezone, physical address.
- **Cannot be skipped.** (Continue is always enabled and always calls the API.)
- Installation URL is pre-filled from `window.location.origin` on mount. If you opened the wizard via a local tunnel or proxy, correct this manually.
- Physical address is required for CAN-SPAM compliance. It appears in campaign footer templates.
- Timezone affects scheduled campaign display.
- API: `PUT /api/setup/settings`

**Step 4 — Connect SMTP**
- Fields: from name, from email, provider selector, provider-specific credentials.
- **Skippable** via "Skip for now" button. SMTP must be configured before sending any campaign.
- **Clicking "Continue" saves credentials** (calls `PUT /api/setup/smtp` regardless of whether you tested first). This is intentional — you can configure now and test later.
- **Clicking "Skip for now" does NOT save credentials.** Nothing is written to the database.
- "Test connection" button: saves credentials then fires a test email to the owner's email.
- Domain hint: as soon as you type a from email address containing `@`, a gray info box shows the domain and reminds you to verify it with your provider.
- Provider hint: a blue info box explains the verification requirement for each provider (Resend 403, Mailgun domains, SES sandbox).

**Step 5 — Send Test Email**
- **Skippable** via "Skip for now" button (which completes setup, same as "Finish Setup").
- Sends a test email to whatever address is entered (pre-filled from the owner email).
- "Finish Setup" / "Skip for now": calls `POST /api/setup/complete`, receives a JWT, stores it, redirects to `/dashboard`.
- After this step, `settings.setup_complete = true` — the `setupGuard` no longer redirects.

### What can be skipped

| Step | Skippable | Consequence of skipping |
|---|---|---|
| Welcome | Yes (Continue freely) | None |
| Owner Account | No | App unusable — no login possible |
| General Settings | No | `INSTALLATION_URL` won't be in DB; tracking URLs use env var fallback |
| Connect SMTP | Yes | Campaigns go to LogMailer (console log only); no actual emails sent |
| Send Test Email | Yes | Setup completes without verifying delivery |

### Must be configured before first send

Before a campaign can actually be sent, the following must all be true:
1. SMTP provider configured with valid credentials (otherwise LogMailer silently swallows sends)
2. From email uses a domain verified with that provider
3. `INSTALLATION_URL` set to a publicly reachable URL (otherwise tracking and unsubscribe links are broken)
4. Campaign has a recipient list or tag selected

---

## 8. Known Gotchas

### 1. LogMailer silent fallback

If SMTP is not configured or credentials fail to load at startup, the app uses `LogMailer`. This logs email send attempts to stdout but never actually delivers anything. There is no visible error in the UI. Campaigns show as "sent" with 0 opens.

**Detection:** Check container logs for `[LogMailer] Would send to ...`. If you see this, SMTP is not configured.

**Fix:** Configure SMTP in Settings > SMTP or re-run the wizard.

### 2. INSTALLATION_URL must be reachable by email clients

The installation URL is embedded in tracking pixels and image `src` attributes in outgoing emails. Email clients (Gmail, Outlook, Apple Mail) fetch these URLs when rendering. If your server is behind a firewall, private IP, or localhost, tracking and images will silently fail. The server must be reachable over the public internet on port 80/443 (or whatever port is in the URL).

### 3. Changing INSTALLATION_URL breaks old tracking links

HMAC tokens in tracking URLs are signed with `APP_SECRET` and embed the installation URL domain only indirectly (via the token payload). However, once an email is delivered with tracking pixels pointing to `old-url.com`, those pixels always point to `old-url.com`. Changing the URL only affects future sends.

### 4. POSTGRES_PASSWORD is a fixed string

The install script sets `POSTGRES_PASSWORD=ownmaily` (a fixed value, not random). This is intentional — the Postgres service is not exposed on any external port so the password is not a meaningful attack surface. The database is only reachable within the Docker internal network under the service name `db`.

### 5. Single-user only

There is one owner account. The `owner` table enforces a single row via `PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE)`. There is no "add user" flow.

### 6. Send rate is 2 emails/second

The send worker has a hard-coded rate limiter of 2 emails per second (burst 1). Large lists take significant time to send. A list of 10,000 subscribers takes approximately 83 minutes. This is configurable only by changing the code (`rate.NewLimiter(rate.Limit(2), 1)` in `internal/worker/send_worker.go`).

### 7. Bulk subscriber select is page-scoped

The bulk select UI on the Subscribers page selects subscribers on the current page only (50 per page). "Select all N" selects all on the current page, not all subscribers matching the filter.

### 8. Wizard Installation URL pre-fill from browser origin

The wizard pre-fills Installation URL from `window.location.origin`. If you open the setup wizard through an SSH tunnel (`localhost:4400`) or via a domain that differs from the public URL (e.g., an IP while the domain isn't yet pointed), the pre-fill is wrong. **Always verify this field before clicking Continue on Step 3.**

### 9. SMTP Continue without testing

Prior to a bug fix during development, clicking "Continue" on the SMTP step did not save credentials — only the "Test connection" button did. Users who filled in their credentials and clicked Continue lost their SMTP config silently. This is fixed: "Continue" now calls `PUT /api/setup/smtp` before advancing. "Skip for now" intentionally bypasses the save.

### 10. Schedule datetime uses browser local time

The schedule date/time picker uses `datetime-local`, which operates in browser local time. The timezone label shown below the picker (e.g., "Your local time (UTC+05:00)") confirms what timezone is being used. The API receives the time converted to UTC via `new Date(value).toISOString()`. This is correct but can surprise users who expect the app to schedule in the instance timezone from Settings.

### 11. No image resize in editor

The Tiptap image extension inserts images at `max-width: 100%`. There is no drag-to-resize handle. Prepare images to the correct dimensions before uploading.

### 12. CTA button `data-cta` attribute

The announcement email template includes a styled CTA button implemented as a Tiptap atom node. The node is identified by a `data-cta` attribute on the `<a>` tag. This attribute is harmless in email clients (unknown attributes are silently ignored) but is load-bearing for the editor — without it, the button loses all styling when a saved campaign is loaded back into the editor. The `fromEmailHTML` function in the frontend detects old-style buttons by their `display:inline-block` + background style and adds `data-cta` automatically, so campaigns saved before this feature was added still rehydrate correctly.

### 13. Windows not supported by installer

The `install.sh` script detects Windows (`$OSTYPE` = `msys`/`cygwin` or `$OS` = `Windows_NT`) and exits immediately. Windows users must use WSL2 or a cloud deploy (Railway, Render).

### 14. Reinstalling wipes the database

If you run `install.sh` when an existing install is detected and confirm the overwrite, the script runs `docker compose down -v`. This destroys both `postgres_data` and `uploads_data` named volumes — all subscriber data, campaign history, and uploaded images are permanently deleted.

### 15. Uploaded images are publicly accessible

The `/uploads/*` route has no authentication. Any URL of the form `https://yourinstance.com/uploads/images/<filename>` is publicly accessible. This is required for email clients to render images. Filenames are 32-character random hex strings so enumeration is not practical, but the content is not secret.

### 16. API key is single-active

The `api_keys` table has no `UNIQUE` constraint on the hash, but regenerating the API key (`POST /api/settings/api-key/regenerate`) replaces the previous key. Only one API key can be active at a time. Old keys are invalidated on regeneration.
