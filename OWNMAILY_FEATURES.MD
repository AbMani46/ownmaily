# OWNMAILY — Feature Specification

This document is the source of truth for what gets built in V1. Claude Code should use this to break work into sessions. Nothing outside this document gets built until after launch.

---

## 1. Authentication

### 1.1 Login

- Single owner account (self-hosted, one user per installation)
- Email + password login
- JWT issued on successful login, stored in httpOnly cookie
- JWT expiry: 7 days, refreshed on activity
- Failed login rate limited (5 attempts then 15 minute lockout)
- No registration screen — owner account seeded on first launch via env vars or setup wizard

### 1.2 API Key

- One API key per installation
- Generated on first launch, regeneratable from settings
- Passed as `Authorization: Bearer <key>` header
- Same permissions as logged in owner
- Used by agents, Zapier, external integrations
- Displayed once on generation, stored as bcrypt hash

### 1.3 Route Protection

- All `/api/*` routes require JWT or API key except:
  - `POST /api/public/subscribe` — public, rate limited
  - `GET /track/open/:token` — public, no auth
  - `GET /track/click/:token` — public, no auth
  - `GET /unsubscribe` — public, no auth
  - `POST /webhooks/:provider` — provider signature verified instead

---

## 2. Subscribers

### 2.1 Subscriber Model

Each subscriber has:

- `id` — UUID
- `email` — unique, lowercase, trimmed
- `first_name` — optional
- `last_name` — optional
- `status` — enum: `active`, `unsubscribed`, `bounced`
- `source` — enum: `import`, `api`, `form`
- `created_at`, `updated_at`

### 2.2 Manual Add

- Add single subscriber via UI form
- Fields: email (required), first name, last name (optional)
- Duplicate email returns friendly error, does not create duplicate

### 2.3 CSV Import

- Upload CSV file
- Required column: `email`
- Optional columns: `first_name`, `last_name`
- First row treated as header
- Validation pass before import — show count of valid/invalid rows
- Invalid rows skipped, valid rows imported
- Duplicates skipped silently
- Show import summary on completion: X imported, Y skipped, Z invalid
- Max file size: 10MB

### 2.4 CSV Export

- Export all subscribers or filtered subset to CSV
- Columns: email, first_name, last_name, status, created_at
- Downloaded as `subscribers-<date>.csv`

### 2.5 Subscriber Detail View

- View individual subscriber
- Show all lists they belong to
- Show all tags
- Show send history: campaigns they received
- Show open history: campaigns they opened
- Show click history: links they clicked
- Unsubscribe manually from detail view
- Delete subscriber (removes from all lists, adds to suppression)

### 2.6 Suppression List

- Separate table from subscribers
- Emails added automatically on hard bounce or unsubscribe
- Can be manually added via settings
- Checked before every send — suppressed emails never receive campaigns
- Can be exported as CSV
- Can be manually removed (with warning)

### 2.7 Search and Filter

- Search subscribers by email or name
- Filter by status (active, unsubscribed, bounced)
- Filter by list
- Filter by tag
- Paginated results, 50 per page

---

## 3. Lists

### 3.1 List Model

- `id` — UUID
- `name` — required, unique
- `description` — optional
- `double_opt_in` — boolean, default false
- `created_at`, `updated_at`
- Subscriber count (derived)

### 3.2 List Management

- Create, rename, delete lists
- Delete warns if list has subscribers
- View all subscribers in a list
- Add subscribers to a list manually
- Remove subscribers from a list

### 3.3 Double Opt-In

- When enabled on a list, new signups via form or public API receive a confirmation email
- Confirmation email contains a unique signed link
- Subscriber status stays `pending` until confirmed
- Pending subscribers do not receive campaigns
- Confirmation link expires after 48 hours
- Resend confirmation available from subscriber detail view

### 3.4 Signup Form Embed

- Each list has a generated embeddable form
- Settings per form: show first name field (toggle), show last name field (toggle)
- Generates a JavaScript snippet: `<script src="https://yourdomain.com/embed/:list_id.js"></script>`
- Snippet renders a simple form inline wherever dropped
- Form submits to `POST /api/public/subscribe`
- Success and error states displayed inline
- Form styling is minimal and intentionally plain so it inherits site styles

---

## 4. Tags

### 4.1 Tag Model

- `id` — UUID
- `name` — unique, lowercase, trimmed
- `created_at`
- Subscriber count (derived)

### 4.2 Tag Management

- Create tags inline when editing a subscriber
- View all tags in settings
- Delete tag (removes from all subscribers)
- Rename tag

### 4.3 Tagging Subscribers

- Add/remove tags on individual subscriber detail view
- Bulk tag: select multiple subscribers, apply tag
- Tags can be used as campaign send target (send to tag instead of list)

---

## 5. Campaigns

### 5.1 Campaign Model

- `id` — UUID
- `name` — internal name, not shown to recipients
- `subject` — email subject line
- `preview_text` — shown in inbox preview (optional)
- `from_name` — sender display name
- `from_email` — sender email address
- `reply_to` — optional
- `html_body` — full HTML email content
- `text_body` — plain text fallback
- `status` — enum: `draft`, `scheduled`, `sending`, `sent`, `failed`
- `send_to_type` — enum: `list`, `tag`
- `send_to_id` — UUID of target list or tag
- `scheduled_at` — nullable, set when scheduled
- `sent_at` — nullable, set when send completes
- `created_at`, `updated_at`

### 5.2 Campaign Creation

- Name the campaign (internal)
- Write subject line and optional preview text
- Set from name and from email (pre-filled from settings, overridable)
- Choose send target: a list or a tag
- Write email content in rich text editor
- Plain text version: auto-generated from HTML, manually editable
- Save as draft at any point

### 5.3 Rich Text Editor

- Basic formatting: bold, italic, underline, strikethrough
- Headings (H1, H2, H3)
- Bullet and numbered lists
- Links
- Images (URL-based, no file upload in V1)
- Horizontal rule
- No drag and drop builder — simple editor only

### 5.4 Email Templates

- 3 prebuilt templates selectable on campaign creation:
  - **Simple** — plain centered text, logo placeholder, footer
  - **Newsletter** — header image area, content sections, footer
  - **Announcement** — bold header, single CTA button, footer
- Templates are starting points, fully editable after selection
- User can save their current campaign as a custom template (V2)

### 5.5 Preview

- Preview email as rendered HTML in a modal
- Toggle between desktop and mobile viewport widths
- Send test email to any address before sending campaign

### 5.6 Send Options

- **Send immediately** — creates send job, starts within seconds
- **Schedule** — pick date and time, scheduler worker picks it up
- Cancel a scheduled campaign before it starts sending
- Cannot edit a campaign once sending has started

### 5.7 Duplicate Campaign

- Clone any existing campaign as a new draft
- Useful for recurring newsletters

### 5.8 Campaign List View

- Shows all campaigns with status badges
- Quick stats inline: open rate, click rate, sent count
- Filter by status

---

## 6. Sending Engine

### 6.1 Send Jobs Table

```
send_jobs
  id            UUID
  campaign_id   UUID
  status        enum: pending, processing, completed, failed
  attempts      int default 0
  next_attempt  timestamp
  created_at    timestamp
  updated_at    timestamp
```

### 6.2 Send Worker

- Background goroutine, starts with the app
- Polls `send_jobs` every 5 seconds for pending jobs
- Claims a job atomically (SELECT FOR UPDATE SKIP LOCKED)
- Loads campaign + all recipients not yet sent to
- Sends in batches of 100
- Marks each `campaign_recipient` as sent after successful batch
- On completion: marks job as completed, campaign status as sent, records sent_at
- On failure: increments attempts, sets next_attempt to now + exponential backoff
- Max 3 attempts before marking job as failed and alerting in UI

### 6.3 Campaign Recipients

```
campaign_recipients
  id              UUID
  campaign_id     UUID
  subscriber_id   UUID
  status          enum: pending, sent, failed, bounced
  sent_at         timestamp nullable
  open_token      text unique  -- used for open tracking pixel
  click_token     text unique  -- used for click tracking base
```

### 6.4 Pre-Send Checks

Before creating a send job:

- SMTP provider configured and credentials valid
- Sending domain has SPF/DKIM (warn if not, don't block)
- Recipient list is not empty
- No suppressed emails in recipient list (filter them out)
- Bounce rate warning if domain bounce rate exceeds 5%

---

## 7. SMTP Providers

### 7.1 Provider Interface

```go
type Mailer interface {
    Send(msg OutboundMessage) error
    Name() string
    ValidateCredentials() error
}

type OutboundMessage struct {
    To          string
    From        string
    FromName    string
    ReplyTo     string
    Subject     string
    HTMLBody     string
    TextBody     string
    Headers     map[string]string
}
```

### 7.2 Resend

- API key authentication
- Uses Resend REST API
- Supports batch sending
- Webhook: `POST /webhooks/resend`
- Webhook signature verified via Resend-Signature header
- Events handled: `email.bounced`, `email.complained`

### 7.3 Mailgun

- API key + domain authentication
- Uses Mailgun REST API
- Webhook: `POST /webhooks/mailgun`
- Webhook signature verified via HMAC
- Events handled: `failed` (hard/soft), `complained`

### 7.4 Amazon SES

- AWS access key + secret + region
- Uses AWS SDK for Go
- Bounce/complaint notifications via SNS → `POST /webhooks/ses`
- SNS subscription confirmed automatically on first message
- Events handled: `Bounce` (Permanent/Transient), `Complaint`

### 7.5 Bounce Normalization

All providers normalize to:

```go
type BounceEvent struct {
    Email     string
    Type      string    // "hard" or "soft"
    Reason    string
    Timestamp time.Time
}
```

### 7.6 Bounce Handler (shared)

- Hard bounce → insert email to suppressions, set subscriber status to `bounced`
- Soft bounce → increment soft bounce count on subscriber, suppress after 3
- Record in `bounces` table regardless of type
- Update campaign bounce stats

---

## 8. Tracking

### 8.1 Open Tracking

- Pixel URL format: `GET /track/open/:token`
- Token is HMAC-signed, encodes `subscriber_id` + `campaign_id`
- Pixel inserted as last element in email `<body>`
- On hit: validate token, record open in `opens` table if not already recorded
- Return 1x1 transparent GIF regardless of token validity
- Never return an error to the email client

### 8.2 Click Tracking

- All `<a href>` tags in campaign HTML rewritten on save
- Rewritten format: `/track/click/:token?url=<base64_encoded_destination>`
- Token encodes `subscriber_id` + `campaign_id` + `link_index`
- On hit: validate token, record click in `clicks` table, redirect to destination
- Multiple clicks by same subscriber on same link recorded (timestamp each)
- If token invalid: redirect to destination anyway, do not error

### 8.3 Unsubscribe

- Unsubscribe link inserted in every campaign email footer
- Format: `/unsubscribe?token=<hmac_token>`
- Token encodes `subscriber_id` + `campaign_id`, signed with app secret
- On hit: validate token, set subscriber status to `unsubscribed`, add to suppressions
- Show simple confirmation page: "You've been unsubscribed."
- No login required, no confirmation step, no "are you sure"
- `List-Unsubscribe` header added to every outbound email:
  `List-Unsubscribe: <https://yourdomain.com/unsubscribe?token=...>`
  `List-Unsubscribe-Post: List-Unsubscribe=One-Click`

---

## 9. Analytics

### 9.1 Overview Dashboard

- Total subscribers (active)
- Total campaigns sent
- Average open rate across all campaigns
- Average click rate across all campaigns
- Subscriber growth chart (last 30 days)
- Recent campaigns list with quick stats

### 9.2 Campaign Stats

Per campaign analytics page:

- Recipients: total sent to
- Opens: unique opens + open rate percentage
- Clicks: unique clicks + click rate percentage
- Unsubscribes: count + rate
- Bounces: hard + soft counts + bounce rate
- Per-link click breakdown: each link URL with click count
- Timeline chart: opens and clicks over time since send

### 9.3 Subscriber Stats

On subscriber detail view:

- Total campaigns received
- Total opens
- Total clicks
- Last active (most recent open or click)

---

## 10. Settings

### 10.1 General Settings

- Site name (shown in UI header)
- Installation URL (used for tracking pixel and unsubscribe links, must be set correctly)
- Timezone (for scheduling display)
- Physical address (inserted in email footer for CAN-SPAM compliance)

### 10.2 SMTP Settings

- Select provider: Resend, Mailgun, SES
- Enter credentials per provider
- Test connection button (sends test email to owner address)
- Default from name and from email
- Default reply-to (optional)

### 10.3 Domain Verification Guidance

- Not automated — guidance only
- After setting from email, show SPF and DKIM records to add to DNS
- Show current status: verified / unverified (checked via DNS lookup)
- Warning banner on campaign create if domain not verified

### 10.4 API Key

- Display current API key (masked)
- Regenerate API key button (invalidates previous key, requires confirmation)
- Copy to clipboard button

### 10.5 Suppression List Management

- View all suppressed emails
- Search suppressions
- Manually add email to suppression list
- Remove from suppression list (with warning)
- Export suppression list as CSV

---

## 11. Setup Wizard

Shown on first launch before any other screen. Cannot be skipped.

### Step 1 — Welcome

- Brief explanation of what OwnMaily is
- What you'll need: a domain, an SMTP provider API key

### Step 2 — Owner Account

- Set email and password for the owner account
- This is the only account that exists

### Step 3 — General Settings

- Installation URL (pre-filled if detectable)
- From name and from email
- Physical address (for CAN-SPAM footer)

### Step 4 — Connect SMTP

- Select provider (Resend, Mailgun, SES)
- Enter credentials
- Test connection — must pass before proceeding

### Step 5 — Send Test Email

- Send a test email to the owner address
- Confirm it arrived
- Setup complete — redirect to dashboard

---

## 12. Public API

Available to customers for integrations and AI agent use. Same endpoints as UI uses.

### Auth

- `Authorization: Bearer <api_key>` header on all requests except public endpoints

### Error Format

All errors return consistent JSON:

```json
{
  "error": "error_code",
  "message": "Human readable description"
}
```

### Endpoints

See OWNMAILY_CONTEXT.md for full endpoint list. All endpoints return JSON. All list endpoints are paginated with `page` and `per_page` query params. Default page size: 50.

### API Documentation

- Markdown file included in repo: `API.md`
- Lists every endpoint, method, params, request body, response shape, and example
- Kept updated as part of every session that adds or changes endpoints

---

## 13. Scheduler Worker

- Background goroutine, starts with the app
- Polls every 60 seconds
- Query: campaigns where `status = scheduled` AND `scheduled_at <= now()`
- For each: create send_job, update campaign status to `queued`
- Idempotent — uses database transaction to prevent double-scheduling

---

## 14. Docker and Deployment

### 14.1 Docker Compose

Single `docker-compose.yml` in repo root:

- `app` service: Go binary + Vue frontend served as static files
- `db` service: PostgreSQL 16
- `volumes` for Postgres data persistence
- `.env` file for all configuration
- Exposed port: 3000 (configurable)

### 14.2 Environment Variables

```
DATABASE_URL=
APP_SECRET=           # used for HMAC signing tokens
INSTALLATION_URL=     # full URL of the installation e.g. https://mail.yoursite.com
PORT=3000
```

SMTP credentials stored in database after setup wizard, not in env.

### 14.3 Install Script

- `curl https://ownmaily.com/install.sh | bash`
- Checks for Docker and Docker Compose, installs if missing
- Pulls `docker-compose.yml` from releases
- Creates `.env` with random APP_SECRET
- Runs `docker compose up -d`
- Prints installation URL and instructs user to complete setup wizard

### 14.4 One-Click Deploys

- Railway: `railway.json` in repo root + Deploy on Railway button in README
- Render: `render.yaml` in repo root + Deploy on Render button in README

### 14.5 Updates

- Lifetime updates included with $99 purchase
- Update process: `docker compose pull && docker compose up -d`
- Migration runner built into app startup (runs pending migrations automatically)
- Changelog maintained in `CHANGELOG.md`

---

## 15. What Is NOT in V1

The following are explicitly out of scope. Claude Code must not build these:

- Automation sequences or drip campaigns
- Drag and drop visual email builder
- Multi-user or team accounts
- Landing page builder
- SMS or any non-email channel
- Ecommerce flows (abandoned cart, post-purchase)
- Built-in SMTP relay (always bring your own)
- Outbound webhooks (notify your app on events) — V2
- Custom template saving — V2
- A/B testing — V2
- Advanced segmentation (behavior-based) — V2

---

## Session Breakdown Guide for Claude Code

Use this to divide work into focused sessions:

| Session | Focus                                                                                                     |
| ------- | --------------------------------------------------------------------------------------------------------- |
| 1       | Project scaffold: Go module, chi, folder structure, Docker Compose, Postgres connection, migration runner |
| 2       | DB schema: all tables, sqlc setup, generated queries                                                      |
| 3       | Auth: login, JWT, API key, route middleware                                                               |
| 4       | Subscribers: CRUD, CSV import, CSV export, search/filter                                                  |
| 5       | Lists: CRUD, list-subscriber junction, double opt-in flow                                                 |
| 6       | Tags: CRUD, subscriber tagging, bulk tagging                                                              |
| 7       | Campaigns: CRUD, draft/schedule status, duplicate                                                         |
| 8       | Sending engine: send_jobs table, send worker goroutine, campaign_recipients                               |
| 9       | Resend integration: sending + bounce webhook parser                                                       |
| 10      | Mailgun integration: sending + bounce webhook parser                                                      |
| 11      | SES integration: sending + SNS bounce webhook parser                                                      |
| 12      | Open tracking: pixel endpoint, token generation, opens table                                              |
| 13      | Click tracking: URL rewriting, redirect endpoint, clicks table                                            |
| 14      | Unsubscribe: token generation, endpoint, List-Unsubscribe header                                          |
| 15      | Scheduler worker: polls scheduled campaigns, creates send jobs                                            |
| 16      | Analytics: overview stats, per-campaign stats endpoints                                                   |
| 17      | Settings: general, SMTP config, API key, suppression management                                           |
| 18      | Vue frontend: project setup, router, Pinia, auth screens, layout                                          |
| 19      | Vue frontend: subscribers screens (list, detail, import)                                                  |
| 20      | Vue frontend: lists and tags screens                                                                      |
| 21      | Vue frontend: campaigns screens (list, create, detail, stats)                                             |
| 22      | Vue frontend: analytics dashboard                                                                         |
| 23      | Vue frontend: settings screens                                                                            |
| 24      | Setup wizard: all 5 steps, first launch detection                                                         |
| 25      | Signup form embed: JS snippet generation, public subscribe endpoint, rate limiting                        |
| 26      | Email templates: 3 prebuilt templates, template selection on campaign create                              |
| 27      | Docker packaging: install script, one-click deploy configs, README                                        |
| 28      | API documentation: API.md covering all endpoints                                                          |
| 29      | End-to-end testing: full send loop, bounce handling, tracking, unsubscribe                                |
| 30      | Pre-launch: landing page, Stripe one-time payment, license delivery                                       |
