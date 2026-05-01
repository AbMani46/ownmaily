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
