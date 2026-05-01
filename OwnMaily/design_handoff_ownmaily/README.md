# Handoff: OwnMaily — Self-hosted Email Marketing Dashboard

## Overview

OwnMaily is a self-hosted email marketing platform. This handoff covers the full admin UI: authentication, subscriber management, list/tag management, campaign authoring and analytics, and settings (SMTP, API key, suppression list).

## About the Design Files

The files in this bundle are **high-fidelity HTML prototypes** — interactive design references showing intended look, layout, and behavior. They are **not** production code to ship directly.

Your task is to **recreate these designs in your target codebase** (React, Next.js, Vue, etc.) using its established patterns, routing library, component library, and data layer. The HTML files are the source of truth for visual design and interaction; treat them the way you'd treat a Figma export.

Open `OwnMaily.html` in a browser to explore all screens interactively. Use the **Tweaks panel** (top-right toolbar icon) to jump between screens.

## Fidelity

**High-fidelity.** Colors, typography, spacing, border radii, shadow values, component states, and copy are all final or near-final. Implement pixel-accurately using your codebase's design system where it overlaps; deviate only where your existing components enforce a different pattern.

---

## Design Tokens

### Colors

```
--accent:          #e07d3c   (primary CTA, active nav, focus rings)
--accent-hover:    #c96a2e
--accent-light:    #fff8f4   (active nav bg, tinted backgrounds)

Background (page):  #f4f3ef
Background (card):  #ffffff
Background (sidebar): #0a0a0c

Border default:    #e8e5de
Border subtle:     #f0ede6
Border dark:       #1a1a22

Text primary:      #111110
Text secondary:    #555
Text muted:        #888 / #aaa
Text inverse:      #fff / #ccc

-- Status colors --
Active/Sent:       bg #d1fae5  text #065f46
Draft/Pending:     bg #fef3c7  text #92400e
Scheduled:         bg #dbeafe  text #1e40af
Bounced/Failed:    bg #fee2e2  text #991b1b
Unsubscribed:      bg #f3f4f6  text #4b5563
Double opt-in:     bg #ede9fe  text #5b21b6
Single opt-in:     bg #e0f2fe  text #075985
Suppressed:        bg #fee2e2  text #991b1b

-- Chart colors --
Opens line:        #e07d3c
Clicks line:       #6366f1
Growth/success:    #059669
```

### Typography

```
Font family (UI):       'DM Sans', sans-serif
Font family (mono):     'JetBrains Mono', monospace

Page title / H1:        14px, weight 600, color #111, letter-spacing -0.02em
Section heading:        13px, weight 600, color #555, uppercase, letter-spacing 0.04em
Body:                   13px, weight 400, color #222
Label (form):           12px, weight 600, color #444, letter-spacing 0.02em
Caption / muted:        12px, weight 400, color #888
Table header:           11px, weight 600, color #888, uppercase, letter-spacing 0.05em
Badge:                  11px, weight 600, uppercase, letter-spacing 0.03em
Stat value (large):     28–30px, weight 700, color #111, letter-spacing -0.04em
Code:                   12px, JetBrains Mono, color #9aefbc on #0c0c0f bg
```

### Spacing

```
Page padding:           24px top/bottom, 28px left/right
Card padding:           20px 22px
Section gap:            18–22px
Form field gap:         14px
Table cell padding:     11px 14px
Table header padding:   9px 14px
Nav item padding:       9px 12px
Sidebar width:          216px (expanded), 56px (collapsed)
Top bar height:         52px
```

### Shape & Elevation

```
Border radius (card):   10px
Border radius (button): 7px
Border radius (input):  7px
Border radius (badge):  20px (pill)
Border radius (small):  5–6px

Card shadow:            none (border only: 1px solid #e8e5de)
Login card shadow:      0 4px 24px rgba(0,0,0,0.06)
Focus ring:             0 0 0 3px rgba(224,125,60,0.12) with border #e07d3c
```

---

## Screens / Views

### 1. Login (`/login`)
**File:** `ownmaily-screens-auth.jsx` → `LoginScreen`

Centered card on a `#f4f3ef` page background.

- **Logo block** — 44×44px rounded square (`border-radius: 12px`) in accent color with white inbox icon, "OwnMaily" title (22px/700) below, subtitle (13px/#888)
- **Card** — 380px wide, `border-radius: 12px`, white bg, 28px padding, shadow `0 4px 24px rgba(0,0,0,0.06)`
  - Email field (full width)
  - Password field with show/hide toggle button (eye icon) absolutely positioned right side
  - Primary button full width, "Sign in", triggers loading state for 800ms then navigates to dashboard
- **Footer link** — "Run setup wizard" in accent color navigates to setup wizard

**States:** Loading ("Signing in…"), default

---

### 2. Setup Wizard (`/setup-wizard`)
**File:** `ownmaily-screens-auth.jsx` → `SetupWizard`

Centered, 520px wide. 5-step flow.

**Step progress bar** — horizontal row of numbered circles connected by lines. Completed steps show a check icon in accent bg. Current step shows accent border, accent text. Future steps are grey. Labels below each node (10px).

**Steps:**
1. **Welcome** — centered icon + headline + description paragraph
2. **Owner Account** — name, email, password, confirm password fields
3. **General Settings** — site name, installation URL (with `https://` prefix), timezone select, physical address textarea
4. **Connect SMTP** — provider select (Resend/Mailgun/Amazon SES/Custom SMTP), API key, from email, from name
5. **Send Test Email** — centered layout, email input, success state with green checkmark

**Navigation:** "Previous" / "Continue →" buttons. Final step says "Finish Setup" and navigates to dashboard after 1200ms.

**Progress bar** — thin line below step title, width animates from 20% → 100% across steps, accent color fill.

---

### 3. Dashboard (`/dashboard`)
**File:** `ownmaily-screens-main.jsx` → `DashboardScreen`

4-column stat card row → 2-column chart row → full-width recent campaigns table.

**Stat cards (row 1):**
- Total Subscribers: **4,218** · +88 this month
- Campaigns Sent: **24** · all time
- Avg. Open Rate: **47.7%** · +2.3% vs last month
- Avg. Click Rate: **8.9%** · +0.4% vs last month

**Charts (row 2):**
- Left card: "Subscriber Growth" — large number + delta text + sparkline SVG (polyline + polygon fill, accent color)
- Right card: "Opens vs Clicks" — two inline metrics + sparkline

**Recent Campaigns table** — last 4 campaigns, columns: Campaign (name + list subtitle), Status badge, Sent date, Open rate (bold green when real), Click rate (bold indigo when real). Clicking a row navigates to campaign stats.

---

### 4. Subscribers (`/subscribers`)
**File:** `ownmaily-screens-subscribers.jsx` → `SubscribersScreen`

**Action bar** — search input (full-width, left-padded search icon), Import CSV button, Export button, Add Subscriber primary button.

**Tabs** — All / Active / Unsubscribed / Bounced, each with count pill.

**Table columns:** Email+Name (stacked, name muted below), Status badge, Tags (pill row), Added date.

Clicking a row navigates to subscriber detail. Pagination: 5 per page.

---

### 5. Subscriber Detail (`/subscribers/detail/:id`)
**File:** `ownmaily-screens-subscribers.jsx` → `SubscriberDetailScreen`

Two-column layout (left: info + tags + lists / right: stats + timeline).

**Left column:**
- **Subscriber card** — avatar circle with initial (48px, accent-tinted bg), name (16px/600), email (13px/muted), status badge. Below: 2×2 grid of info tiles (Added, Last Active, Source, IP Country) on `#faf8f4` bg, 9px padding, 7px radius.
- **Tags card** — pill row with "+ Add tag" ghost button in header
- **List memberships card** — each list shown as row with name, subscribed date, opt-in badge

**Right column:**
- **3 stat cards** (Campaigns received, Opens, Clicks) in a grid
- **Activity timeline** — vertical connector line, each event has a colored circle icon, event text + date. Event types: open (accent), click (indigo), join (green), confirm (blue)

---

### 6. Lists (`/lists`)
**File:** `ownmaily-screens-subscribers.jsx` → `ListsScreen`

Top 3 lists shown as **card grid** (3 columns). Each card: icon square (36px, accent-tinted), list name (15px/600), subscriber count (22px/700 in accent color), created date. Hover: accent border + subtle box shadow.

Below: full table of all lists. Columns: List Name, Subscribers (bold), Opt-in Type badge, Created.

---

### 7. List Detail (`/lists/detail/:id`)
**File:** `ownmaily-screens-subscribers.jsx` → `ListDetailScreen`

3-column stat row (Total Subscribers, Opt-in Type, Created date) → subscriber table → embed code card.

**Embed code card** — monospace code block on `#0c0c0f` bg, green text (`#9aefbc`), with Copy button in header.

---

### 8. Tags (`/tags`)
**File:** `ownmaily-screens-subscribers.jsx` → `TagsScreen`

Full-width table. Columns: Tag Name (with 8px color dot), Subscribers, Actions (Edit + Delete buttons). Delete is in red ghost style.

---

### 9. Campaigns (`/campaigns`)
**File:** `ownmaily-screens-campaigns.jsx` → `CampaignsScreen`

Tabs: All / Sent / Scheduled / Drafts with counts.

Table columns: Campaign (name + list), Status badge, Date, Sent count, Open Rate (green bold), Click Rate (indigo bold), Actions (Edit for drafts/scheduled, Stats for sent). Row click also navigates.

---

### 10. Campaign Create/Edit (`/campaigns/create/:id`)
**File:** `ownmaily-screens-campaigns.jsx` → `CampaignCreateScreen`

**Template selector bar** — row of template pills (Blank, Newsletter, Announcement, Weekly Digest). Selected pill has accent border + tinted bg.

**Two-column layout** (360px left / flex-1 right):

**Left — Settings panel:**
- Campaign name
- Subject line
- Preview text (with hint copy)
- Divider
- From name, From email, Reply-to
- Divider
- "Send to" list select
- Info box showing recipient count (`#faf8f4` bg, 10px radius)

**Right — Editor:**
- Toolbar: Bold/Italic/Underline buttons (28px square), H1/H2/H3 pills, Link icon, Image button, List icon, right-aligned Preview toggle
- Content area: `contentEditable` div with HTML email content, 24px padding
- Preview mode: centered 600px email preview on `#f4f3ef` bg, white card with 32px padding, serif font rendering
- Bottom bar: Save Draft + Schedule (with calendar icon) + Send Now (with send icon) buttons

---

### 11. Campaign Stats (`/campaigns/stats/:id`)
**File:** `ownmaily-screens-campaigns.jsx` → `CampaignStatsScreen`

**Header card** — campaign name (18px/700), status badge, sent date, list name, recipient count. Duplicate + View Email buttons.

**6-column stat grid** — Sent, Delivered (with %), Opens (with rate), Unique Opens, Clicks (with rate), Bounces.

**Timeline chart (SVG)** — opens (accent) and clicks (indigo) plotted as lines with filled polygon below. X axis labeled in hours (0h → 48h).

**Link breakdown** — each link URL in monospace, click count + percentage right-aligned, progress bar below (accent/indigo/green/grey by rank).

---

### 12. Analytics (`/analytics`)
**File:** `ownmaily-screens-main.jsx` → `AnalyticsScreen`

**Period selector** — pill buttons: 7 days / 30 days / 90 days / All time. Selected = accent bg.

**Row 1 stats (4 cols):** Total Subscribers, Emails Sent, Avg Open Rate, Avg Click Rate

**Row 2 stats (4 cols):** Unsubscribes, Bounces, New Subscribers, Deliverability

**Charts row (2:1 grid):**
- Left: "Opens by Day of Week" — vertical bar chart SVG, 7 bars (Mon–Sun), each labeled below
- Right: "Subscriber Breakdown" — three labeled progress bars (Active/Unsubscribed/Bounced) with counts

**Campaign performance table** — same columns as campaigns list, filtered to sent campaigns only.

---

### 13–16. Settings (`/settings/general`, `/settings/smtp`, `/settings/apikey`, `/settings/suppression`)
**File:** `ownmaily-screens-settings.jsx`

**Layout:** Left settings sidebar (200px) + right content area. Sidebar items have accent left-border when active.

**Settings sidebar items:** General, SMTP, API Key, Suppression List — each with icon.

#### 13. General Settings
Sections: Site Information (name, URL, timezone), Sender Defaults (from name/email/reply-to), Compliance (physical address textarea). Save button with animated "Saved ✓" confirmation.

#### 14. SMTP Settings
Provider select changes the fields shown below it:
- **Resend:** API key (masked), sending domain
- **Mailgun:** API key, domain, region
- **Amazon SES:** Access key ID, secret key, region
- **Custom SMTP:** host, port, username, password, encryption

"Test connection" button cycles through loading → success states.

#### 15. API Key
- Masked key display in dark monospace block (`#0c0c0f` bg)
- Show/hide toggle, copy button (shows check on copy)
- Warning callout (amber bg, amber border)
- "Regenerate" → confirmation card with destructive warning, confirm/cancel buttons
- API usage code snippet

#### 16. Suppression List
- Search bar, Export + Add Email buttons
- Orange warning callout explaining suppression behavior
- Table: email (monospace), reason badge, date added, remove button (red)
- Rows can be removed from local state (remove button)

---

## Interactions & Behavior

### Navigation
All navigation is client-side state (`screen` string, e.g. `"subscribers/detail/1"`). Implement with your router (React Router, Next.js App Router, etc.).

### Status Badges
All badges use a 5px dot + uppercase label pattern. Map status strings → color pairs (see Design Tokens above).

### Tables
- Row hover: `background: #faf8f4`
- Clickable rows: `cursor: pointer`, navigate on click
- Column headers: uppercase, 11px, muted

### Buttons
Four variants:
- **primary** — accent bg, white text, hover darkens (`#c96a2e`)
- **secondary** — `#eae7e0` bg, dark text, grey border
- **ghost** — transparent, hover `#f0ede6` bg
- **danger** — `#dc2626` bg, white text

All buttons: 7px radius, 120ms transition, 13px font, `500` weight.

### Inputs & Selects
- Default border: `#ddd`
- Focus border: `#e07d3c` + `0 0 0 3px rgba(224,125,60,0.12)` box shadow
- 7px radius, 13px font, 8px 11px padding

### Sidebar
- Expanded: 216px, shows icon + label
- Collapsed: 56px, shows icon only, label as tooltip (`title` attr)
- Active item: `#1e1e26` bg, accent left border (2px), accent icon color, white text
- Dark background (`#0a0a0c`), dark border (`#1a1a22`)

### Top Bar (App Shell)
- 52px height, white bg, `#ece9e1` bottom border
- Breadcrumb built from screen path (e.g. `subscribers/detail/1` → "Subscribers › Detail")
- Right side: contextual action buttons

---

## App Shell Structure

```
<div style="display:flex; height:100vh">
  <Sidebar />                          // 216px fixed, sticky
  <div style="flex:1; display:flex; flex-direction:column">
    <TopBar />                         // 52px sticky
    <main style="flex:1; overflow-y:auto">
      {/* screen content */}
    </main>
  </div>
</div>
```

Page content is always `padding: 24px 28px` inside the scrollable main area.

---

## Component File Map

| File | Contents |
|------|----------|
| `ownmaily-icons.jsx` | SVG icon components (Lucide-style). All icons used in the UI. |
| `ownmaily-components.jsx` | `Badge`, `Button`, `StatCard`, `Input`, `Select`, `Table`, `Card`, `SectionTitle`, `Tabs`, `Pill`, `CodeBlock`, `Pagination` |
| `ownmaily-layout.jsx` | `Sidebar`, `TopBar`, `AppShell` |
| `ownmaily-screens-auth.jsx` | `LoginScreen`, `SetupWizard` |
| `ownmaily-screens-main.jsx` | `DashboardScreen`, `AnalyticsScreen`, mock data |
| `ownmaily-screens-subscribers.jsx` | `SubscribersScreen`, `SubscriberDetailScreen`, `ListsScreen`, `ListDetailScreen`, `TagsScreen` |
| `ownmaily-screens-campaigns.jsx` | `CampaignsScreen`, `CampaignCreateScreen`, `CampaignStatsScreen` |
| `ownmaily-screens-settings.jsx` | `SettingsGeneralScreen`, `SettingsSMTPScreen`, `SettingsAPIKeyScreen`, `SettingsSuppressionScreen` |
| `ownmaily-app.jsx` | Root app, routing, tweaks wiring |
| `OwnMaily.html` | Entry point — open this in a browser to explore all screens |

---

## Mock Data

All data is local to `ownmaily-screens-main.jsx`:
- `MOCK_CAMPAIGNS` — 5 campaigns (sent/scheduled/draft)
- `MOCK_SUBSCRIBERS` — 7 subscribers with varied statuses
- `MOCK_LISTS` — 5 lists
- `MOCK_TAGS` — 6 tags with colors

Replace with real API calls in your implementation.

---

## Assets

No external image assets. All icons are inline SVG. Fonts loaded from Google Fonts:
- `DM Sans` (UI font, weights 300–700)
- `JetBrains Mono` (code/monospace blocks, weights 400/500)

In production, self-host these fonts or substitute your design system's equivalents.
