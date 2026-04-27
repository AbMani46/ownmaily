# OWNMAILY — Project Context

## What is OwnMaily

OwnMaily is a self-hosted email marketing tool. Buy it once for $99, own it forever, no subscriber limits, no monthly fees. You bring your own SMTP provider, connect your domain, and send to your list. Your data lives on your server.

The target customer is the indie founder, bootstrapper, or small newsletter operator who thinks $50/month for email marketing is insane and wants full control over their data.

---

## Business Model

- **Price:** $99 one-time purchase
- **Updates:** Lifetime — buy once, get all future updates
- **Hosting:** Customer's own server — GetMaily provides the software only
- **SMTP:** Customer brings their own (Resend, Mailgun, or SES)
- **No subscriber limits, no sending limits, no monthly fees**

---

## Tech Stack

Consistent with existing projects (Middl, SolidUptime):

- **Backend:** Go, chi router, sqlc, pgx, PostgreSQL
- **Frontend:** Vue 3 Composition API, script setup, Pinia, Vue Router 4
- **Database:** PostgreSQL in Docker
- **Deployment:** Docker Compose (primary), Railway/Render/Coolify one-click
- **File structure:** Monorepo — `/backend` and `/frontend`

---

## SMTP Provider Support

Launch with Resend first, add Mailgun and SES before v1 ships.

### Provider Interface (Go)

All providers implement a common interface:

```go
type Mailer interface {
    Send(to, from, subject, html, text string) error
    Name() string
}
```

### Bounce Normalization

Each provider has its own webhook parser. All parsers normalize to a single internal struct:

```go
type BounceEvent struct {
    Email     string
    Type      string    // "hard" or "soft"
    Reason    string
    Timestamp time.Time
}
```

Everything downstream (suppression list, bounce rate, unsubscribe handler) is provider-agnostic.

### Provider Status

| Provider | Sending | Bounce Parser | Priority   |
| -------- | ------- | ------------- | ---------- |
| Resend   | v1      | v1            | Launch     |
| Mailgun  | v1      | v1            | Pre-launch |
| SES      | v1      | v1            | Pre-launch |

---

## Core Engine

### 1. Sending Queue

Campaigns are sent via a background worker, not synchronously. Architecture:

- `jobs` table in Postgres tracks pending/in-progress/completed sends
- Background goroutine polls for pending jobs
- Sends in batches (max 100 per request per Resend recommendation)
- Respects provider rate limits
- Supports pause/resume on failure
- No Redis required — Postgres queue only

### 2. Bounce Handling

- Hard bounce → immediately add to suppression list, never retry
- Soft bounce → retry up to 3 times, then suppress
- Suppression list is checked before every send
- Bounce rate tracked per campaign and globally per sender domain
- Warn user if bounce rate exceeds 5%

### 3. Open Tracking

- 1x1 invisible pixel embedded in every campaign email
- Unique URL per recipient per campaign: `/track/open/<token>`
- Token is HMAC-signed, encodes subscriber ID + campaign ID
- First hit recorded as open, subsequent hits ignored
- Gmail image proxy acknowledged — not perfectly accurate, good enough

### 4. Click Tracking

- All links in campaign HTML rewritten to tracking URLs on save
- Format: `/track/click/<token>?url=<encoded_destination>`
- Token encodes subscriber ID + campaign ID + link index
- Redirect to destination after recording
- Click data stored per recipient per link per campaign

### 5. Unsubscribe

- Every email includes unsubscribe link and `List-Unsubscribe` header (required by Gmail/Yahoo for bulk senders)
- Token format: HMAC-signed JWT encoding subscriber ID + campaign ID
- Stateless — no DB lookup needed to validate token
- One-click, no confirmation screen, no login required (CAN-SPAM compliant)
- Processed instantly on click
- Endpoint: `/unsubscribe?token=<token>`

### 6. Public Signup Endpoint

- `POST /api/public/subscribe` — no auth required
- Accepts email + list ID
- Rate limited (prevent abuse on self-hosted installs)
- Returns success/already subscribed/invalid
- Optional double opt-in confirmation email (configurable per list)
- JavaScript embed snippet generated per list in UI

---

## REST API

The Vue frontend is built entirely on top of the REST API. The same API is available to customers for integrations and AI agent use.

### Auth

- Single user (self-hosted, one owner)
- JWT-based auth
- API key support for programmatic access

### Core Endpoints (planned)

```
POST   /api/auth/login
POST   /api/auth/logout

GET    /api/subscribers
POST   /api/subscribers
DELETE /api/subscribers/:id
POST   /api/subscribers/import        -- CSV upload
GET    /api/subscribers/export        -- CSV download

GET    /api/lists
POST   /api/lists
PUT    /api/lists/:id
DELETE /api/lists/:id

GET    /api/campaigns
POST   /api/campaigns
PUT    /api/campaigns/:id
POST   /api/campaigns/:id/send
POST   /api/campaigns/:id/schedule
GET    /api/campaigns/:id/stats

GET    /api/analytics/overview
GET    /api/analytics/campaigns/:id

GET    /api/settings
PUT    /api/settings/smtp
PUT    /api/settings/general

POST   /api/public/subscribe          -- no auth, rate limited
GET    /track/open/:token             -- no auth
GET    /track/click/:token            -- no auth
GET    /unsubscribe                   -- no auth
```

---

## Features — V1 Scope

### Subscribers

- Add manually or import via CSV
- Export to CSV
- Tags for segmentation
- Multiple lists
- Suppression list (bounced + unsubscribed)
- View per-subscriber send/open/click history

### Campaigns

- Rich text email editor
- Plain text fallback
- 2-3 prebuilt clean templates
- Send immediately or schedule
- Send to a list or a tag
- Duplicate existing campaign
- Preview before send

### Analytics

- Per-campaign: open rate, click rate, unsubscribe rate, bounce rate
- Overview dashboard: total subscribers, total campaigns sent, overall engagement
- Per-link click counts within a campaign

### Settings

- SMTP provider selection + credentials
- Sending domain + from name + reply-to
- Domain verification guidance (SPF/DKIM instructions)
- API key management
- General settings (site name, timezone)

### Signup Forms

- Embeddable JS snippet per list
- Configurable fields (email required, name optional)
- Double opt-in toggle per list
- Simple success/error states

### Setup Wizard (first launch)

- Step 1: Set sender name and from email
- Step 2: Connect SMTP (select provider, paste API key)
- Step 3: Send a test email to yourself
- Step 4: Import your first list or create a signup form

---

## Distribution

| Method          | Notes                                               |
| --------------- | --------------------------------------------------- | ----------------------------------------------- |
| Docker Compose  | Primary install method, single `docker-compose.yml` |
| Install script  | `curl url                                           | bash` — pulls Docker, sets up compose, launches |
| Railway         | One-click deploy button in README                   |
| Render          | One-click deploy button in README                   |
| Coolify/Dokploy | App listing                                         |
| Pikapods        | Marketplace listing — free distribution channel     |
| Elestio         | Marketplace listing                                 |

Goal: customer goes from purchase to first campaign in under 10 minutes.

---

## What We Are NOT Building in V1

- Automation sequences / drip campaigns (v2)
- Drag and drop visual email builder (v2)
- Multi-user / team accounts (v2)
- Landing pages (never, out of scope)
- SMS (never, out of scope)
- Ecommerce flows / abandoned cart (never, out of scope)
- Built-in SMTP (never — always bring your own)

---

## Compliance Notes

- CAN-SPAM: one-click unsubscribe, physical address in footer (user configures), honor requests instantly
- GDPR: data lives on customer's server — their responsibility, not ours. Document this clearly
- Gmail/Yahoo bulk sender requirements: List-Unsubscribe header, DKIM, low bounce rate — all enforced by GetMaily

---

## Naming

- **Product:** OwnMaily
- **Domain:** ownmaily.com
- **Tagline (draft):** Own your email list. No monthly fees. No limits.

---

## Session Notes

- Project conceived April 2026
- Stack chosen to match existing Jordan projects (Go + Vue)
- Resend chosen as launch SMTP due to clean API and bootstrapper audience fit
- Postgres jobs table chosen over Redis for simplicity
- $99 one-time price validated against market (MailWizz $299, Sendy $69, Swift Send $49)
- Lifetime updates chosen over annual renewal for v1 to reduce friction
