# OwnMaily UI Audit

Audited: 2026-04-30
Scope: all files in `frontend/src/views/` and `frontend/src/components/`

---

## Campaign Editor (`CampaignEditView.vue`)

This is the highest-risk view. The toolbar has a fundamental interaction bug plus several secondary UX issues.

### Rich Text Toolbar — Root Bug

**Highlighted text is not affected by formatting buttons.**

The cause: toolbar buttons have no `@mousedown.prevent`. When a user selects text in the `contenteditable` editor and then clicks a toolbar button, the sequence is:
1. `mousedown` fires on the button → browser blurs the editor → text selection is destroyed
2. `click` fires → `execCmd` runs on an empty selection → nothing happens

The `editorRef.value?.focus()` call inside `execCmd` restores focus after the fact, but the selection is already gone. This affects every toolbar action: Bold, Italic, Underline, H1/H2/H3, unordered list. All of them silently fail on selected text.

Note: H1/H2/H3 use `formatBlock` which operates at block level (wraps the whole paragraph), so the selection loss doesn't matter for those — they apply to the current paragraph regardless. But the user doesn't know this and will perceive them as broken too, since selecting a word and clicking H2 formats the whole paragraph, not just the word.

**Specific toolbar issues:**

1. Bold/Italic/Underline have no active/toggle state. When the cursor is inside bold text, the B button shows no visual indication of the active state. Users can't tell what formatting is currently applied.

2. H1/H2/H3 buttons have no active state either. No visual feedback when the cursor is in a heading.

3. Template selector uses raw Unicode box characters (`□`, `▦`, `◉`) as icons — these render poorly and look like placeholder symbols, not real UI.

4. Clicking a template pill immediately overwrites the editor content with no confirmation dialog. A user who accidentally clicks "Newsletter" while mid-draft loses all their work with no undo.

5. `insertLink` uses `window.prompt('Enter URL:')` — a native browser alert. This is visually inconsistent with the rest of the app and loses any highlighted text in some browsers.

6. No placeholder text in the empty editor area. A blank white box gives no affordance to new users.

### Settings Panel Issues

7. Required field validation is deferred to server: Name, Subject, From Name, From Email have `*` labels but no client-side validation. The user can click "Send Now" with all fields empty and only discovers the error after the API rejects it.

8. "Send Now" button is only disabled when `send_to_id` is empty (`!form.send_to_id`). It does not validate name, subject, or from_email. Users can attempt to send campaigns with no subject line.

9. Schedule button has a confusing disabled condition: `:disabled="!campaignId && saveLoading"`. It is enabled even when a new campaign has never been saved — clicking it opens the schedule modal, which then auto-saves, which is fine in practice but the intent isn't obvious and the button state is misleading.

10. Schedule modal doesn't validate that `send_to_id` is selected. A user can schedule a campaign with no recipient list/tag set.

### Action Bar

11. "Save Draft" success has no visual confirmation — the loading spinner disappears and nothing else changes. No toast, no "Saved" state change. Users are left wondering if the save worked.

12. No autosave. Any navigation away from the editor (clicking sidebar, browser back) silently discards unsaved content.

### Preview Mode

13. The Edit/Preview toggle is small and positioned in the top-right corner of the toolbar. When in preview mode there is no persistent banner or mode indicator. It's easy to be confused about which mode you're in.

14. Preview mode does not accurately simulate an email client — fonts, line spacing, and layout differ from what Outlook/Gmail actually renders. The disclaimer should note this.

---

## Campaigns List (`CampaignsView.vue`)

1. The "Date" column label is static but shows different things per status: `created_at` for drafts, `scheduled_at` for scheduled, `sent_at` for sent. A user scanning the column doesn't know what date they're looking at.

2. Row clicks navigate to edit for drafts and to stats for sent campaigns. There is no visual affordance (hover state is subtle background change, no arrow indicator) to tell users rows are clickable or that clicking a sent campaign goes somewhere different from clicking a draft.

3. Delete action uses `window.confirm()` — a native browser dialog that looks out of place and isn't styled.

4. Deleting a "scheduled" campaign shows the same generic confirm as deleting a draft, with no mention that the scheduled send will not go out.

5. Tab counts require 4 separate API calls on mount (`fetchAllCounts`). There's no loading state on the count badges — they show the number or the previous stale count while fetching.

6. Empty state is a plain text string with no CTA ("No campaigns yet. Create your first campaign.") — no button to actually create one.

---

## Dashboard (`DashboardView.vue`)

1. **Bug:** The "Recent Campaigns" table has hardcoded `—` for Open Rate and Click Rate regardless of actual data:
   ```
   <template #cell-open_rate="{ value }"><span class="rate-dash">—</span></template>
   ```
   The template slot ignores the `value` entirely. Even if a campaign has tracking data, the dashboard always shows dashes.

2. The `delta` prop on `StatCard` exists and supports positive/negative trend formatting, but it is never used anywhere in the dashboard. All stat cards show raw numbers with no growth/change context.

3. No onboarding or "get started" state. A new user with zero campaigns, zero subscribers, and a freshly configured SMTP sees four "—" stat cards and an empty table. No guidance on what to do next.

4. The 4-column stat grid is fixed at `repeat(4, 1fr)` with no responsive breakpoints.

---

## Analytics View (`AnalyticsView.vue`)

1. **Critical UX issue:** The period selector (7 days / 30 days / 90 days / All time) is purely decorative. A note next to it says "Period filter is display-only — data shows all time". Users will click these pills expecting filtered data and see no change. Having a row of styled, interactive-looking buttons that do nothing is a significant UX problem.

2. **Bug (same as dashboard):** The campaign performance table hardcodes `—` for open rate and click rate on every row. The template slots `#cell-_open_rate` and `#cell-_click_rate` don't exist in `AnalyticsView` — it uses a plain `<table>` with hardcoded `<td class="rate-dash">—</td>`.

3. The Analytics view uses the Options API (`export default {}`), while all other views use `<script setup>`. This is a style inconsistency and means the view doesn't benefit from automatic `<script setup>` optimizations.

4. The skeleton loading only covers the 8 stat cards. The campaign performance table below has its own separate `campaignsLoading` state but just shows text "Loading campaigns…" instead of a skeleton.

5. Two rows of 4 stat cards (8 total) means a lot of scrolling before reaching the campaign table. "Deliverability" and "Active Subscribers" feel secondary compared to sent/opens/clicks.

---

## Lists View (`ListsView.vue`)

1. **Redundant layout:** The top 3 lists are shown as clickable cards, then the same lists appear again in the table below. This duplicates information without adding clarity. If a user has only 1–2 lists, the cards section feels half-empty.

2. No way to edit a list's name or description. The only actions available after creation are: view subscribers, add a subscriber, and delete.

3. The Create List modal uses a raw `<input type="checkbox">` for the double opt-in toggle. All other forms in the app use styled custom inputs. This native checkbox is visually inconsistent.

4. No search or filter on the lists table.

5. `double_optin` badge (`double_optin` / `single_optin`) text is all-caps ("DOUBLE OPT-IN", "SINGLE OPT-IN") due to `text-transform: uppercase` on `.badge`. The hyphen in "OPT-IN" looks odd in this style. Most first-time users won't know what this means without tooltip/explanation.

---

## List Detail View (`ListDetailView.vue`)

1. **Delete List** is a prominent danger button sitting directly in the page header alongside the list name. It takes one click + one browser confirm to permanently delete a list with all its subscriber associations. The danger button color red is sufficient warning but the placement is too accessible for a destructive action.

2. The "Add Subscriber to List" modal searches by email via debounce, then shows "No subscriber found with that email. Add them first →". This link is a plain `<a href="/subscribers">` anchor — it navigates away from the list detail page without returning the user to the list afterward.

3. The subscriber membership date in List Detail shows `formatDate(subscriber.created_at)` — this is the subscriber's global signup date, not when they were added to this specific list. Users see incorrect dates if a subscriber was added to the list later.

4. The embed code section shows an embed script tag (`/embed/:id.js`) but there is no corresponding frontend or backend implementation of this embed endpoint visible in the codebase. The button and code block exist for unimplemented functionality.

5. No success confirmation when removing a subscriber from a list — the row disappears (after a refetch) but no feedback acknowledges the action.

---

## Subscriber Detail View (`SubscriberDetailView.vue`)

1. **"Last Active"** field always shows `—`. This field is rendered in the info grid but the data is never populated from the API response.

2. **"Source"** field shows `—` for most subscribers. No explanation of what "source" means.

3. The tag dropdown stays open until you click "Add tag" again (the toggle button). Clicking anywhere else on the page does not close it. There's no click-outside handler.

4. The Actions card contains "Unsubscribe" (secondary button style) and "Delete Subscriber" (danger/red button style) side by side. The visual distinction is appropriate but the lack of any text explanation or spacing between them makes the section feel rushed.

5. Stats row shows campaign count, total opens, total clicks. These are raw numbers with no rates or context — "3 opens" out of how many campaigns? No open rate is calculated for the subscriber.

6. The info grid shows "Status" as a text value (e.g. "active") in a tile AND a `BaseBadge` with the same status appears in the subscriber card header. The status is shown twice in different formats on the same view.

---

## Subscriber List (`SubscribersView.vue`)

1. Search and tab filters are mutually exclusive in behavior (search ignores the active status tab) but the UI gives no indication of this. A user filtering "Bounced" subscribers, then typing in the search box, will silently start searching all statuses.

2. Export triggers a silent download. There's no loading state on the Export button and no success message. If the export fails, the error is logged to the console only.

3. Import modal: after a successful import, the result summary ("Imported X, Skipped Y, Invalid Z") appears in the modal and stays. The only way to dismiss it is the "Close" button. There's no "Import another file" action — the user has to close and reopen.

4. The subscribers table has a `loading` ref but no loading indicator is shown in the table when fetching (no skeleton rows, no spinner). The table just appears empty while loading.

5. Tags column shows max 3 tags with "+N more" text. There's no tooltip or popover to show the remaining tags without navigating to the subscriber detail.

6. No bulk actions — no way to select multiple subscribers and tag/untag/delete in one operation.

---

## Tags View (`TagsView.vue`)

1. No click-through from the Tags table to a filtered subscriber list. You can see "124 subscribers" but can't navigate to see who they are.

2. Tag colors are auto-generated from a hash of the name (deterministic but not user-controllable). No way to assign a preferred color.

3. The delete confirmation message says "This will remove it from all subscribers" but doesn't show the subscriber count in the confirm dialog. A user deleting a high-value tag has no warning of impact.

4. Actions column has a visible "Actions" header label — most other tables in the app use an empty label for the actions column. Minor inconsistency.

---

## Settings — General (`SettingsGeneralView.vue`)

1. No inline validation before saving. Empty "From email" or malformed "Installation URL" is only caught server-side. The user submits and waits for the API error.

2. The "Saved ✓" confirmation disappears after 2 seconds. If the user is looking at a different part of the form when save completes, they see nothing.

3. The timezone selector shows 12 hardcoded options. Users in timezones not listed (e.g., India/Mumbai, America/Sao_Paulo) have no option and default to UTC silently.

---

## Settings — SMTP (`SettingsSMTPView.vue`)

1. After saving SMTP credentials and returning to this page, the API key fields are blank (expected — keys aren't returned for security). But there's no "configured" indicator. The page looks identical whether SMTP is set up or not, leaving users unsure if their configuration is active.

2. "Test connection" tests the currently saved SMTP (calls `POST /api/settings/smtp/test`), not the current unsaved form values. A user who changes the provider to Mailgun and clicks "Test connection" without saving first tests the old Resend config.

3. No links to API key dashboards for each provider. First-time users setting up Resend need to navigate away and find their key independently.

---

## Settings — API Key (`SettingsAPIKeyView.vue`)

1. The copy button works only after regenerating a key (when `newKey` is set). Before that, clicking it copies bullet characters (the masked display) and shows a `✓` checkmark as if it succeeded — this is misleading. The function explicitly only calls `navigator.clipboard.writeText` when `newKey.value` is truthy.

2. The "Save this — it won't be shown again" warning and the main warning box are both present before any key is generated, which is confusing. These should only appear after regenerating.

3. The `keyPrefix` initializes to `om_` and renders as `om_` + 40 bullets, implying a key already exists even on a fresh install with no key generated yet.

---

## Settings — Suppression List (`SettingsSuppressionView.vue`)

1. Uses its own custom inline modal (`<teleport to="body">` + custom CSS) instead of the `BaseModal` component used everywhere else. Visual inconsistency — different backdrop opacity, different animation, different styling.

2. Remove action uses `window.alert()` for errors and `window.confirm()` for confirmation. Both are native browser dialogs.

3. Export uses `window.location.href = '/api/settings/suppressions/export'`. This causes a brief page redirect before the browser triggers the download, which is visually jarring.

---

## Setup Wizard (`SetupWizardView.vue`)

1. Step 4 (Connect SMTP): the "Continue" button is disabled until SMTP is tested (`nextDisabled = !smtpTested`). But there's no visible explanation that testing is required — users may try to click Continue, see it disabled, and not understand why.

2. Step 2 (Account): no client-side validation prevents advancing with an empty email or short password. The error only appears after the API rejects the payload. The step feels broken — you click Continue and it fails.

3. The "Test connection" in step 4 sends a test email to `owner.value.email || general.value.installation_url`. If the owner email field wasn't filled in step 2, it falls back to the installation URL string — sending a test email to a URL is obviously wrong.

4. The Welcome step checklist uses filled green dots (`check-dot` class) instead of checkmarks, which looks like a "you must have these" requirements list. No item is checked off as the user proceeds through setup.

5. No skip option for the test email step (step 5). If SMTP works but the user doesn't want to send a test right now, they're blocked — they must enter an email and send.

6. After completing setup and being redirected to `/dashboard`, if the user navigates back to `/setup`, the wizard re-runs. On step 2 it will fail ("owner already exists") but shows a generic error with no guidance.

---

## Login View (`LoginView.vue`)

1. "Run setup wizard →" is always visible on the login page. On a fully configured production system, this link is confusing (what does "setup" mean if I'm already using this?). It should only appear if setup is incomplete.

2. No auto-focus on the email input on page load. Users must manually click into the field.

---

## AppSidebar (`AppSidebar.vue`)

1. No user account information shown anywhere in the sidebar or topbar — no email address, no name, no indication of which installation you're connected to.

2. No logout confirmation. Clicking Log Out immediately calls `auth.logout()` with no warning.

3. `isActive(path)` uses `startsWith`, which is correct for nested routes. However, `/dashboard` also starts with `/dashboard` for all sub-routes, so this works fine. But `/settings` will be active for all `/settings/*` routes — that's intentional but worth noting that clicking "Settings" in the nav doesn't navigate to a specific section.

---

## AppTopBar (`AppTopBar.vue`)

1. The `<slot />` in the topbar right side is defined but not used by any view. All views add their own action bars within the page body. The topbar is just a title bar.

2. The title derives from `route.meta?.title || formatRouteName(route.name)`. Route names like `campaign-edit` become "Campaign Edit" and `campaign-stats` becomes "Campaign Stats" — grammatically odd. Some pages (e.g. `CampaignEditView`) override with their own `page-title` element inside the view, creating two visual "titles" (topbar title + page content title) — although in CampaignEditView, there's no in-view page title since the template bar takes that slot.

---

## BaseTable (`BaseTable.vue`)

1. All rows always have `cursor: pointer` and fire `@click`. In views where no `@row-click` handler is wired (or where the row is read-only), clicking still shows a pointer cursor and emits the event. No `clickable` prop to disable this.

2. No integrated loading state. Views must implement their own skeletons outside the table — this creates inconsistency (dashboard has skeletons, subscribers table has nothing).

3. No `<caption>` or `aria-label` on the `<table>` element. Accessibility issue.

---

## BaseModal (`BaseModal.vue`)

1. Pressing Escape does not close the modal. There's no `keydown` event listener.

2. Modal animation is opacity-only (`opacity` fade). No scale or slide transform. The appearance feels abrupt.

3. Max-width is hardcoded at 480px and cannot be overridden by callers.

---

## BasePagination (`BasePagination.vue`)

1. No page number display — only prev/next arrows. Users with large datasets (e.g. 500 subscribers, 50 per page) have no visibility into how many pages exist or which page they're on beyond "Showing 51–100 of 500".

2. The pagination component conditionally renders itself (`v-if="total > perPage"`). In `SubscribersView`, the wrapper `<div v-if="total > 0">` wraps the pagination. If total = exactly `perPage`, the component renders but its internal `v-if` hides it, while the wrapper div shows. Minor but creates an invisible DOM element.

---

## useConfirm (`composables/useConfirm.js`)

`useConfirm` wraps `window.confirm()` directly. Every deletion and destructive action in the app uses this. Native browser confirms are unstyled, block the thread, and behave inconsistently across browsers. This affects:
- Campaign delete
- Subscriber unsubscribe/delete  
- List delete
- Removing subscriber from list
- Suppression list remove

---

## Cross-Cutting Issues

1. **No toast/notification system.** Success feedback is either silent (export download), transient inline text (General Settings "Saved ✓"), or navigational (send campaign → redirect). Errors appear as inline red text. There's no unified notification layer.

2. **No responsive design.** Fixed grids (`repeat(4, 1fr)`, `repeat(6, 1fr)`, `grid-template-columns: 360px 1fr`) will break on any screen narrower than ~1100px. The campaign editor 6-stat grid at `repeat(6, 1fr)` will overflow at normal laptop widths.

3. **Inconsistent confirm dialogs.** Some destructive actions use `window.confirm()` (delete campaign, remove from suppression list), others use the styled `BaseModal` (none currently — all destructive confirms use the native dialog). The pattern should be unified.

4. **No empty state illustrations.** Every empty state is plain text. First-time users see blank tables with a single muted sentence. No visual guidance, no CTAs, no onboarding nudges.

5. **`BaseBadge` color collision.** `active` and `sent` statuses share the same green color scheme. In the campaigns table where a sent campaign's badge is green, and an active subscriber's badge is also green, the semantic difference is lost.

6. **`document.execCommand` is deprecated.** The entire rich text toolbar is built on `execCommand`, which browsers have deprecated and may remove. This is a technical debt item but has UX impact (inconsistent behavior across browsers already).

7. **No keyboard navigation in the editor.** No keyboard shortcuts (Ctrl+B for bold, Ctrl+I for italic) since the toolbar handles formatting and there are no key listeners on the editor area.
