# OwnMaily Session Log

---

## Session: UI Polish — Pre-Launch Audit Fixes
**Date:** 2026-04-30
**Status:** Complete

### What was done

Implemented all "Fix Before Launch" and "Fix Before Launch if Quick" items from UI_AUDIT.MD and UI_AUDIT_SPLIT.MD. Changes listed by component:

---

#### BaseModal (`components/BaseModal.vue`)
- Added `keydown` Escape listener (`onMounted`/`onUnmounted`) that emits `close` when modal is open

---

#### Campaign Editor (`CampaignEditView.vue`)
- Added `@mousedown.prevent` to every toolbar button (Bold, Italic, Underline, H1/H2/H3, Link, List) — preserves text selection when clicking formatting actions
- Replaced `window.prompt` for insert link with an inline `BaseModal` that saves/restores the selection via `window.getSelection().getRangeAt(0).cloneRange()`
- Added "✓ Saved" state on the Save Draft button that appears for 2.5 seconds after a successful save (`saveSuccess` ref)
- Added template pill confirmation modal before overwriting non-empty editor content (`confirmApplyTemplate` → `showTemplateConfirm` → `applyPendingTemplate`)
- Added CSS `:empty:before` placeholder text on the editor area via `data-placeholder` attribute
- Removed Unicode box character icons from template pills (they rendered poorly)

---

#### Dashboard (`DashboardView.vue`)
- Fixed hardcoded `—` for open rate and click rate in the Recent Campaigns table — now renders actual values when present (with green `.rate-value` style), falls back to `—` only when `null`

---

#### Analytics (`AnalyticsView.vue`)
- Removed the period selector pills entirely (they were decorative and misleading)
- Fixed hardcoded `—` for open rate and click rate in the Campaign Performance table — wired to `c.open_rate` and `c.click_rate`
- Converted from Options API (`export default {}`) to `<script setup>` for consistency with other views

---

#### Settings — API Key (`settings/SettingsAPIKeyView.vue`)
- Copy button is now hidden until a key has been regenerated (`v-if="newKey"`)
- Both warning boxes (`.new-key-box` "Save this" and `.warn-box` full-access warning) now only render after a key is generated (`v-if="newKey"`)

---

#### Setup Wizard (`SetupWizardView.vue`)
- Added "Test the connection above to continue" hint text next to the disabled Continue button on the SMTP step (step 3, index 3), visible when `!smtpTested`
- Fixed test email fallback: `testSmtp()` now errors clearly if `owner.value.email` is empty (instead of silently falling back to `installation_url`). The test is only sent to the owner's email address.

---

#### Login (`LoginView.vue`)
- Added `onMounted` call to `GET /api/setup/status` — hides the "Run setup wizard" link when `res.data.complete === true`
- **Deviation:** The spec says to check `GET /api/settings`, but that endpoint requires auth and returns 401 on the login page regardless of setup state. Used `GET /api/setup/status` instead (assumed to exist as a public endpoint). Default state is `showSetupLink = true` (visible), so the link still shows on fresh installs or if the endpoint doesn't exist.

---

#### Campaigns List (`CampaignsView.vue`)
- "Date" column label is now dynamically computed: "Sent" for the Sent tab, "Scheduled For" for the Scheduled tab, "Created" for All/Drafts
- Empty state now includes a "Create your first campaign" primary button that navigates to `/campaigns/new`
- Delete confirmation for scheduled campaigns now reads: "The scheduled send will be cancelled. This cannot be undone."

---

#### Subscribers (`SubscribersView.vue`)
- Added skeleton rows (8 rows) while the subscriber table is loading
- Fixed search behavior: search now passes `status` filter alongside the `q` parameter, so searching while on the "Bounced" tab correctly searches only bounced subscribers. Added an inline comment documenting the fix.

---

#### Lists (`ListsView.vue`)
- Added an Edit button to each list row (via `_actions` column) that opens an Edit List modal (name + description fields)
- Replaced native `<input type="checkbox">` for double opt-in with a styled CSS toggle switch (`toggle-switch` / `toggle-thumb` CSS classes)

---

#### Subscriber Detail (`SubscriberDetailView.vue`)
- Added `onDocumentClick` click-outside handler that closes the tag dropdown when clicking anywhere outside the tag card (`ref="tagCardRef"`)
- "Last Active" tile now renders `subscriber.last_opened_at` via `formatDate()` instead of hardcoded `—`
  - **Deviation:** Assumed the backend field is `last_opened_at`. If the actual field name differs (e.g., `last_active`, `updated_at`), update the template accordingly.

---

#### Settings — Suppression List (`settings/SettingsSuppressionView.vue`)
- Refactored custom `<teleport>` modal to use `BaseModal` component — now consistent with the rest of the app (same backdrop, animation, close button, and Escape key handling)

---

#### Sidebar (`components/AppSidebar.vue`)
- Shows the logged-in user's email in the sidebar footer above the Log out button
- Email is derived from `auth.userEmail` (new computed in `stores/auth.js`)
- `userEmail` decodes the JWT payload from `localStorage` (`atob(token.split('.')[1])`) and reads the `email` field
  - **Deviation:** Assumes the JWT payload contains an `email` field. If the backend puts it under a different key (e.g., `sub`), update the auth store computed.

---

### Deviations from spec
1. **Login setup link check:** used `GET /api/setup/status` instead of `GET /api/settings` (which requires auth)
2. **Last Active field:** used `last_opened_at` — verify against actual API response field name
3. **Sidebar user email:** decoded from JWT `email` claim — verify the backend encodes email in the JWT payload

---

## Session: Polish Fixes (post-audit)
**Date:** 2026-04-30
**Status:** Complete

### What was done

Three targeted fixes resolving the deviations noted above:

1. **`GET /api/setup/status` endpoint** (`internal/handler/setup.go`, `cmd/server/main.go`)
   - Added `SetupHandler.Status` — calls `GetSettings`, returns `{"complete": true/false}`, falls back to `false` on DB error.
   - Registered as `r.Get("/api/setup/status", setupHandler.Status)` outside the auth group. The existing `setupGuard` already skips all `/api/` prefixed paths so no additional skip list entry needed.

2. **`last_active` field fix** (`frontend/src/views/SubscriberDetailView.vue:30`)
   - Changed `subscriber.last_opened_at` → `subscriber.last_active` to match actual API response field.

3. **JWT `sub` claim fix** (`frontend/src/stores/auth.js:14`)
   - Changed `payload.email` → `payload.sub` — the backend encodes owner email as `RegisteredClaims.Subject`, not a custom `email` field. Resolves blank email in sidebar footer.

---

## Session: Interactive Installer (`install.sh`)
**Date:** 2026-04-30
**Status:** Complete

### What was done

Rewrote `install.sh` as a full production-quality interactive installer using whiptail.

**Flow:**
1. Windows detection — exits immediately with WSL2 / one-click deploy instructions
2. whiptail check/install — tries `apt-get`, `yum`, `brew` in order; exits with clear manual instructions if all fail
3. Docker check/install — uses `curl -fsSL https://get.docker.com | sh`, adds user to docker group
4. Docker Compose v2 check — exits with distro-specific install instructions if missing
5. whiptail welcome screen
6. Installation URL — inputbox, validates non-empty, strips trailing slash
7. Admin email — inputbox, validates contains `@`, loops until valid
8. Admin password — passwordbox, minimum 8 chars, confirm passwordbox, loops until match
9. `APP_SECRET` — generated via `openssl rand -hex 32`
10. Creates `$HOME/ownmaily`, downloads `docker-compose.yml` from GitHub
11. Writes `.env` (`APP_SECRET`, `INSTALLATION_URL`, `PORT=4400`, `POSTGRES_PASSWORD=ownmaily`), `chmod 600`
12. Writes `setup_credentials.txt` with email + password + instructions, `chmod 600`
13. `docker compose pull` + `docker compose up -d`
14. Health check — polls `http://localhost:4400/health` every 3s, 60s timeout, dot progress
15. whiptail success/partial-success screen
16. Stdout summary with URL, email, credentials path, and useful docker commands

**Design decisions:**
- Uses `#!/usr/bin/env bash` + `set -euo pipefail`
- `_sudo()` helper — no-op if root, uses `sudo` otherwise
- whiptail I/O redirection: `3>&1 1>&2 2>&3` captures inputbox value to variable; each call guarded with `|| die "Installation cancelled."` so ESC/Cancel exits cleanly under `set -e`
- Health check distinguishes "ready" vs "started but timed out" — shows appropriate whiptail screen and stdout message in both cases
- `setup_credentials.txt` is `chmod 600` with a prominent reminder to delete after wizard completion

---

## Session: README Quick Start Update
**Date:** 2026-05-01
**Status:** Complete

### What was done

Updated the Quick Start section of `README.md` to match the current install script flow exactly.

**Changes:**
- Added paragraph under the one-line install command explaining the three questions and automatic behaviour (Docker install, secret generation, port 4400)
- Added "Open the URL in your browser to complete setup via the setup wizard." line
- Renamed "Docker Compose (manual)" heading to "Manual install (Docker Compose)"
- Added Windows section directing users to WSL2 or Railway/Render one-click buttons
- Confirmed no em dashes anywhere in the file

---

## Session: Installer Rewrite (whiptail -> plain read)
**Date:** 2026-05-01
**Status:** Complete

### What was done

Rewrote `install.sh` from scratch to remove the whiptail dependency entirely. The
new script uses plain `read` prompts via `/dev/tty` so it works correctly when piped
from curl.

**Changes from the previous installer:**
- Removed all whiptail calls (welcome screen, inputboxes, passwordboxes, success msgbox)
- All three prompts use `printf ... >/dev/tty` + `read -r ... </dev/tty` (or `read -rs` for passwords) so stdin from the curl pipe does not interfere
- Error symbol changed from `[✗]` (unicode) to `[x]` (plain ASCII) per spec
- No em dashes anywhere in output strings
- Removed whiptail install logic entirely (no apt-get/yum/brew for whiptail)
- Installation URL now defaults to `http://localhost:4400` (shown in bracket prompt) instead of pre-filled `https://`
- Final summary is plain stdout only -- no secondary TUI screen

**Flow unchanged:**
- Windows detection, Docker auto-install via get.docker.com, Docker Compose v2 check
- 3 questions: installation URL, admin email, admin password (min 8 chars, confirm loop)
- Generates APP_SECRET and POSTGRES_PASSWORD via openssl rand
- Creates ~/ownmaily, downloads docker-compose.yml, writes .env (chmod 600) and setup_credentials.txt (chmod 600)
- docker compose pull + up -d
- Polls http://localhost:4400/health every 3s, 60s timeout
- Prints URL, admin email, credentials path, and delete reminder

---

## Session: Docker Hub Release Tasks
**Date:** 2026-05-01
**Status:** Complete

### What was done

Added two new tasks to `Taskfile.yml` for publishing to Docker Hub under `knobbly007/ownmaily`.

- **`docker-push`** — builds the image tagged `knobbly007/ownmaily:latest` and pushes it. Standard release flow: `task docker-push`.
- **`docker-release`** — builds and pushes both `latest` and a short-SHA tag (e.g. `knobbly007/ownmaily:a1b2c3d`) in one command. SHA is resolved at run time via `git rev-parse --short HEAD`.

No other files changed.

---

## Session: Secret Audit + .gitignore Fix
**Date:** 2026-04-30
**Status:** Complete

### What was done

Full codebase audit for hardcoded secrets. Three files scanned: all `.go`, `.vue`, `.js`, `.sh`, `.yml`, `.sql`, config files.

---

#### Findings

| File | Finding | Verdict |
|------|---------|---------|
| `install.sh:189` | `POSTGRES_PASSWORD=ownmaily` hardcoded into production `.env` | **Fixed** |
| `docker-compose.yml:10,23` | `${POSTGRES_PASSWORD:-ownmaily}` fallback exposes known weak default | **Fixed** |
| `internal/testutil/fixtures.go:22` | `testpass123` hardcoded as test owner password | Acceptable — test-only package, never shipped |
| `internal/testutil/server.go:18` | `TestAppSecret = "test-secret-do-not-use"` | Acceptable — test-only package |
| `internal/testutil/db.go:17` | `postgres://ownmaily:ownmaily@localhost:5433/...` hardcoded fallback | Acceptable — test code, overridden by `TEST_DB_URL` env var |
| `internal/config/config.go` | No hardcoded secrets — reads `APP_SECRET`, `DB_URL` from env, errors if missing | Clean |
| `.env` | `APP_SECRET=change-me-before-production` | Git-ignored — not committed |
| `.env.example` | Placeholder values only | Clean |

No hardcoded API keys, tokens, real email addresses, or JWT secrets found outside `.env` (which is git-ignored).

---

#### Fixes Applied

**`install.sh`** — generated `POSTGRES_PASSWORD` alongside `APP_SECRET`:
- Added `POSTGRES_PASSWORD=$(openssl rand -hex 16)` in the "Generate secrets" block
- Changed `.env` write from `POSTGRES_PASSWORD=ownmaily` → `POSTGRES_PASSWORD=${POSTGRES_PASSWORD}`

**`docker-compose.yml`** — removed `:-ownmaily` weak fallback defaults:
- `DB_URL` line: `${POSTGRES_PASSWORD:-ownmaily}` → `${POSTGRES_PASSWORD}`
- `POSTGRES_PASSWORD` env: `${POSTGRES_PASSWORD:-ownmaily}` → `${POSTGRES_PASSWORD}`
- Result: compose now fails clearly if no `.env` is present rather than silently using a known credential

**`.gitignore`** — added exception to track `.env.example`:
- Changed `**/.env*` → two lines: `**/.env*` + `!.env.example`
- Verified: `.env` remains ignored; `.env.example` is now untracked (visible to git)
