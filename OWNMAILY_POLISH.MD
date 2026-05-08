# OwnMaily Polish Log

Tracks incremental UI/UX improvement sessions. Each session documents decisions, what changed, and what's next.

---

## Session 1 — Tiptap Editor (foundations)

**Goal:** Replace the contenteditable-based email editor with Tiptap. Foundations only — no toolbar redesign, no active state indicators, no advanced features.

### What was installed

```
@tiptap/vue-3              3.22.5   — Vue 3 component + useEditor composable
@tiptap/pm                 3.22.5   — ProseMirror peer dep (required explicitly)
@tiptap/starter-kit        3.22.5   — Bold, Italic, Heading, BulletList, HorizontalRule, History, etc.
@tiptap/extension-underline 3.22.5  — Ctrl+U; not in StarterKit
@tiptap/extension-link      3.22.5  — Insert link; openOnClick: false so links don't navigate in editor
@tiptap/extension-placeholder 3.22.5 — "Start writing your email…" placeholder via Tiptap's own mechanism
```

### What changed

**New file:** `frontend/src/components/EmailEditor.vue`

- `<script setup lang="ts">`
- Props: `modelValue: string` / emits `update:modelValue` — fully v-model compatible
- Extensions: StarterKit + Underline + Link + Placeholder
- Emits new HTML on every `onUpdate` via `editor.getHTML()`
- Watches `modelValue` prop; calls `editor.commands.setContent()` when parent changes it (handles template apply + campaign load)
- Link modal lives **inside** this component. Parent toolbar calls `emailEditorRef?.openLinkModal()` — the modal, state (linkUrl), and `chain().focus().setLink()` are all internal.
- `defineExpose` surface (called by parent toolbar): `execBold`, `execItalic`, `execUnderline`, `execHeading(1|2|3)`, `execBulletList`, `openLinkModal`
- Scoped styles on `:deep(.ProseMirror)` replicate the original editor-area look

**Modified:** `frontend/src/views/CampaignEditView.vue`

- `<div contenteditable>` → `<EmailEditor ref="emailEditorRef" v-model="editorHTML" />`
- All `document.execCommand()` calls gone; toolbar buttons delegate to `emailEditorRef?.exec*()`
- Link modal, `showLinkModal`, `linkUrl`, `savedSelection`, `linkInputRef` removed from parent
- `syncEditorHTML()`, `execCmd()`, DOM `input` listener, all `editorRef.value.innerHTML` reads/writes removed
- `saveDraft()` reads `editorHTML.value` directly (always in sync via v-model)
- `nextTick` import removed — no DOM manipulation left

### Keyboard shortcuts (work out of the box)

| Action    | Shortcut                   |
| --------- | -------------------------- |
| Bold      | Ctrl+B / Cmd+B             |
| Italic    | Ctrl+I / Cmd+I             |
| Underline | Ctrl+U / Cmd+U             |
| Undo      | Ctrl+Z / Cmd+Z             |
| Redo      | Ctrl+Shift+Z / Cmd+Shift+Z |

### Architecture decisions

- **Link modal stays in EmailEditor, not parent.** Parent toolbar calls `openLinkModal()` via `defineExpose`. Editor instance never leaks to parent.
- **No toolbar redesign this session.** Existing toolbar markup kept; only the click handlers were rewired.
- **No active state on toolbar buttons.** Toolbar doesn't visually reflect bold-on / heading-active etc. Deferred to a later session.

### Known gaps / next sessions

- Toolbar buttons have no active state (e.g., Bold button doesn't highlight when cursor is inside bold text). This requires passing `editor.isActive('bold')` etc. to the parent — suggest moving toolbar inside `EmailEditor` when tackling this. ✅ Done in Session 2.
- No image support yet.
- The preview pane uses `v-html="editorHTML"` — same as before, no change to preview rendering.
- Template HTML strings (`TEMPLATE_HTML` in `CampaignEditView`) still use raw HTML; Tiptap renders them fine but the `blank` template (`<p></p>`) triggers the watch no-op correctly (editor already starts empty).

---

## Session 2 — Toolbar active states + Cmd+K link

**Goal:** Move toolbar inside `EmailEditor.vue` so it has direct access to the editor instance. Active state on all formatting buttons. Clean Cmd+K link insertion.

### What changed

**Modified:** `frontend/src/components/EmailEditor.vue`

- Added `previewMode?: boolean` prop and `toggle-preview` emit — toolbar now owns the preview toggle button
- Toolbar markup (Bold, Italic, Underline, H1/H2/H3, Link, Bullet, Preview toggle) moved fully inside the component
- Imported `Link2`, `List`, `Eye` from lucide-vue-next
- **Active state reactivity:** added `transactionCount` ref incremented in `onTransaction`. Each `isActive` computed (isBold, isItalic, isUnderline, isH1, isH2, isH3, isBulletList, isLink) reads `transactionCount.value` to create a reactive dependency, then calls `editor.value?.isActive(...)`. Buttons bind `:class="{ active: isBold }"` etc.
- **Active style:** `.tb-btn.active` / `.tb-pill.active` → `background: var(--accent-light); border-color: var(--accent); color: var(--accent)`. Icon color follows CSS `currentColor` (no hard-coded `color="#555"` on lucide icons).
- **Cmd+K / Ctrl+K:** `@keydown="handleKeydown"` on `.email-editor` wrapper; handler checks `e.metaKey || e.ctrlKey` + `e.key === 'k'`, calls `openLinkModal()`.
- **Link input auto-focus:** modal now uses a plain `<input ref="linkInputRef">` instead of `BaseInput`; `openLinkModal()` calls `nextTick(() => linkInputRef.value?.focus())`.
- **Layout:** `.email-editor` is now `display: flex; flex-direction: column; height: 100%`. Added `.editor-body { flex: 1; min-height: 0; overflow-y: auto }` wrapper around `EditorContent` so toolbar stays fixed while text scrolls.
- Removed `defineExpose` entirely — parent no longer calls any methods on the component.

**Modified:** `frontend/src/views/CampaignEditView.vue`

- Removed entire `.toolbar` div (28 lines of markup)
- `<EmailEditor>` tag: removed `ref="emailEditorRef"`, added `:preview-mode="previewMode"` and `@toggle-preview="previewMode = !previewMode"`
- Removed `Link2, List, Eye` from lucide imports
- Removed `const emailEditorRef = ref(null)`
- Removed all toolbar scoped CSS (`.toolbar`, `.tb-btn`, `.tb-pill`, `.toolbar-sep`, `.toolbar-spacer`, `.preview-toggle` + hover variants — ~70 lines)
- `.editor-content`: removed `overflow-y: auto` (EmailEditor handles its own scrolling now)
- `.preview-wrap`: added `height: 100%; overflow-y: auto; box-sizing: border-box` so preview pane scrolls independently

### Keyboard shortcuts (full list)

| Action    | Shortcut                   |
| --------- | -------------------------- |
| Bold      | Ctrl+B / Cmd+B             |
| Italic    | Ctrl+I / Cmd+I             |
| Underline | Ctrl+U / Cmd+U             |
| Link      | Ctrl+K / Cmd+K             |
| Undo      | Ctrl+Z / Cmd+Z             |
| Redo      | Ctrl+Shift+Z / Cmd+Shift+Z |

### Architecture decisions

- **`transactionCount` pattern for active states.** Tiptap's `Editor` instance doesn't change identity on each keystroke — `editor.value` stays the same ref. So computed properties that call `isActive()` don't re-run unless forced. Incrementing a `transactionCount` on every `onTransaction` gives Vue's reactivity system a dependency to track without any extra watcher overhead.
- **Toolbar owns `previewMode` display, parent owns state.** Preview content lives in the parent; EmailEditor reflects the current mode via a prop and fires an event when toggled. No two-way ref bleed.
- **Plain `<input>` in link modal.** Swapped from `BaseInput` wrapper to a native `<input ref="linkInputRef">` so `focus()` can be called directly after `nextTick`. Styled to match `BaseInput` visually.

### Known gaps / next sessions

- No image support yet.
- Toolbar doesn't have a "remove link" affordance beyond clearing the URL field and submitting.
- The preview pane uses `v-html="editorHTML"` — no change to preview rendering.

---

## Session 3 — Variable chips (inline atom nodes)

**Goal:** Allow inserting merge-tag variables (`{{first_name}}`, `{{unsubscribe_url}}`) as styled inline chips that are atomic (deleted whole, not character-by-character). Chips serialize correctly to the backend's expected `{{double_braces}}` syntax.

### What was installed

```
@tiptap/core  3.22.5  — explicit dep; was a transitive dep, not directly importable under pnpm
```

### What changed

**New file:** `frontend/src/components/VariableChipNode.vue`

- Vue NodeView component for the chip; used by `VueNodeViewRenderer`
- `<NodeViewWrapper as="span">` — renders inline, not as a block div
- `selected` prop drives `.var-chip--sel` class (accent background when node-selected)
- Displays `{{ first_name }}` / `{{ unsubscribe_url }}` (with spaces) in monospace font
- Scoped styles: accent-colored pill, `user-select: none`, `vertical-align: middle`

**Modified:** `frontend/src/components/EmailEditor.vue`

- `Node, mergeAttributes` imported from `@tiptap/core`; `VueNodeViewRenderer` from `@tiptap/vue-3`; `onClickOutside` from `@vueuse/core`
- `VariableChip` extension defined inline (before `useEditor`):
  - `inline: true, group: 'inline', atom: true` — `atom` is the key flag; ProseMirror treats the node as indivisible
  - `addAttributes()` → `id` attribute (variable name)
  - `parseHTML()` → matches `span[data-variable]`; restores chips when loading saved HTML
  - `renderHTML()` → `<span data-variable="first_name" class="variable-chip">{{first_name}}</span>`
  - `addNodeView()` → `VueNodeViewRenderer(VariableChipNodeView)`
- `VariableChip` added to `extensions` array
- `VARIABLES` const: `[{ id, label, hint }]` — hint is pre-built `{{id}}` string (avoids Vue template parser choking on nested `{{`)
- `showVarsDropdown` ref + `varsWrapRef` ref; `onClickOutside(varsWrapRef, ...)` closes dropdown on outside click
- `insertVariable(id)` — closes dropdown, then `chain().focus().insertContent({ type: 'variableChip', attrs: { id } }).run()`
- Toolbar: new separator + `.vars-wrap` div after the List button; button labeled `{ } ▾`; absolute-positioned `.vars-dropdown` lists both variables with name + hint columns
- `.vars-item:hover .vars-item-hint` → hint turns accent color on hover

**Modified:** `frontend/src/views/CampaignEditView.vue`

- `toEmailHTML(html)` — regex strips chip spans back to `{{id}}` before sending to backend: `html.replace(/<span\s[^>]*data-variable="([^"]+)"[^>]*>.*?<\/span>/g, '{{$1}}')`
- `saveDraft()` now sends `html_body: toEmailHTML(editorHTML.value || '')` — chips serialize as `{{first_name}}` to match `email_builder.go`'s `strings.ReplaceAll` replacements
- `:deep(.preview-body .variable-chip)` — chips in the preview pane are styled (muted accent pill) so they're visually clear even in `v-html` context

### Architecture decisions

- **`atom: true` for atomicity.** ProseMirror treats atom nodes as single characters — arrow keys jump over them whole, Backspace/Delete removes the entire node. No custom keymap needed.
- **NodeView for editor display, `renderHTML` for serialization.** These are separate paths. The NodeView is what the user sees while typing; `renderHTML` is what `getHTML()` outputs to `editorHTML`. Both paths are needed.
- **Chip span preserved in `editorHTML`, stripped only at save time.** `editorHTML` retains `<span data-variable="...">` so the editor can round-trip (load saved campaign → chips reappear). `toEmailHTML` converts only when writing to the backend. Preview pane uses `editorHTML` directly and sees the spans (styled by `:deep`).
- **`{{...}}` hint strings in JS, not in template.** Vue's template compiler tokenizes `{{` as interpolation syntax — putting `'{{' + v.id + '}}'` inside a template expression crashes the parser. Pre-computing `hint: '{{first_name}}'` in the JS constant sidesteps this.
- **`onClickOutside` from `@vueuse/core`.** Already a project dep. Cleaner than manual document event listeners + watch cleanup.

### Variables supported

| Label           | Chip display            | Serialized            | Backend replacement    |
| --------------- | ----------------------- | --------------------- | ---------------------- |
| First name      | `{{ first_name }}`      | `{{first_name}}`      | `subscriber.FirstName` |
| Unsubscribe URL | `{{ unsubscribe_url }}` | `{{unsubscribe_url}}` | signed token URL       |

### Known gaps / next sessions

- No image support yet.
- Template HTML strings in `TEMPLATE_HTML` (newsletter, announcement) still contain raw `{{first_name}}` text — they load into the editor as plain text, not chips. This is acceptable for v1; a migration to pre-convert template HTML to chip format could come later.
- Toolbar has no "remove link" affordance (pre-existing gap from Session 2).

---

## Session 4 — HTML email output + email client compatibility

**Goal:** Audit the HTML Tiptap outputs via `editor.getHTML()`, identify what breaks in Gmail/Outlook/Apple Mail, inline styles where needed, and verify variable chip serialization.

### What was audited

Tiptap's raw `getHTML()` output for a typical email (heading, bold paragraph, bullet list, link, chip, HR):

```html
<h2>Hello there,</h2>
<p>Some <strong>bold</strong> text.</p>
<ul>
  <li><p>First item</p></li>
  <li><p>Second item</p></li>
</ul>
<p>Check out our <a href="https://example.com">website</a>.</p>
<p>
  <span data-variable="first_name" class="variable-chip">{{first_name}}</span>
</p>
<hr />
```

**Issues found:**

| Issue                                                        | Client                   | Severity                                         |
| ------------------------------------------------------------ | ------------------------ | ------------------------------------------------ |
| `<li><p>` nesting — StarterKit wraps list item text in `<p>` | Gmail, Yahoo             | Structural — causes double-height bullet spacing |
| No `padding-left` on `<ul>`                                  | Outlook 2007–2019        | Structural — bullets render flush-left           |
| No inline styles on headings                                 | Outlook, cross-client    | Spacing inconsistency                            |
| No `margin` on paragraphs                                    | Outlook                  | Spacing inconsistency                            |
| Links have no inline `color`                                 | Yahoo Mail, Android Mail | Accent color overridden by client defaults       |
| Plain `<hr>`                                                 | Outlook                  | Renders as 3D embossed border                    |
| Variable chip `<span>`                                       | All                      | ✓ Already serialized correctly via `toEmailHTML` |

### What changed

**Modified:** `frontend/src/components/EmailEditor.vue`

- `Link.configure()` now includes `HTMLAttributes: { style: 'color:#10b981;text-decoration:underline;' }` — links in `getHTML()` output carry inline color, so the preview pane reflects the correct email link color and `editorHTML` is already email-safe for links

**Modified:** `frontend/src/views/CampaignEditView.vue`

- `toEmailHTML(html)` replaced: was a single regex; now uses `DOMParser` to manipulate the DOM before serializing. Full pipeline:
  1. **Variable chips** — `span[data-variable]` → `{{id}}` text node (same behavior as before)
  2. **Headings** — `setAttribute('style', ...)` with `font-size`, `font-weight`, `margin`, `line-height` per level
  3. **Paragraphs** — adds `margin:0 0 12px 0` only if no existing `margin` (preserves template footer `<p style="color:#888;font-size:13px">`)
  4. **Bullet lists** — `padding:0 0 0 20px;margin:0 0 14px 0;` on `<ul>` for Outlook
  5. **List items** — unwraps `<li><p>` → `<li>` by hoisting `<p>`'s children in-place before removing the `<p>` node
  6. **Links** — adds `color:#10b981` and `text-decoration:underline` only if `el.style.color` is not already set (preserves CTA button links with `color:#fff`)
  7. **HR** — `border:none;border-top:1px solid #e5e7eb;margin:24px 0;`

### Architecture decisions

- **DOMParser in `toEmailHTML`, not `renderHTML` overrides.** Custom Tiptap node extensions that override `renderHTML` would require disabling StarterKit's built-in extensions via `StarterKit.configure({ heading: false, ... })` — an API with some uncertainty in Tiptap 3.22.5. The DOMParser approach achieves identical output without touching the editor's schema or document model. The editor experience is unchanged; only the payload sent to the backend is email-safe.
- **`toEmailHTML` is the single transform boundary.** All email-specific mutations happen exactly once, at save/send time. `editorHTML` stays as Tiptap's raw output; `html_body` in the API payload is the processed version.
- **Selective style injection for paragraphs.** `margin` is only added if the element has no existing `margin` inline style. This preserves the footer paragraph in the newsletter template (`style="color:#888;font-size:13px"`) without overriding its author intent.
- **Link color check before injection.** `!el.style.color` means any link that already has an explicit inline `color` (e.g., a white-on-green CTA button) is left untouched. Without this guard, template buttons with `color:#fff` would have their color silently overridden to green.

### Variable chip serialization — confirmed correct

The full round-trip is:

1. User inserts chip via toolbar → `insertContent({ type: 'variableChip', attrs: { id: 'first_name' } })`
2. `getHTML()` → `<span data-variable="first_name" class="variable-chip">{{first_name}}</span>` (via `renderHTML` in the VariableChip extension)
3. `toEmailHTML` → `{{first_name}}` (text node replacement via DOMParser)
4. Backend stores `{{first_name}}`; `email_builder.go`'s `strings.ReplaceAll` substitutes the subscriber's first name
5. Load saved campaign → `editorHTML.value = c.html_body` → Tiptap `parseHTML` matches `span[data-variable]` → chip rehydrates ✓

### Sample output from `toEmailHTML` for the newsletter template

```html
<h2 style="font-size:22px;font-weight:700;margin:0 0 14px 0;line-height:1.3;">
  Hello {{first_name | "there"}},
</h2>
<p style="margin:0 0 12px 0;">Here's what's been happening this month.</p>
<h3 style="font-size:18px;font-weight:600;margin:0 0 12px 0;line-height:1.4;">
  What's new
</h3>
<p style="margin:0 0 12px 0;">Lorem ipsum dolor sit amet...</p>
<hr style="border:none;border-top:1px solid #e5e7eb;margin:24px 0;" />
<p style="color:#888;font-size:13px;margin:0 0 12px 0;">
  You're receiving this because you subscribed.
  <a href="{{unsubscribe_url}}" style="color:#10b981;text-decoration:underline;"
    >Unsubscribe</a
  >
</p>
```

### Known gaps / next sessions

- No image support yet.
- Announcement template's CTA button (`<a style="background:#10b981;color:#fff;padding:...">`) loses its button styling when loaded into Tiptap — the Link extension only preserves `href`/`target`, not arbitrary style attributes. Pre-existing gap; not introduced by this session. A dedicated "button" node extension would fix this.
- Template HTML strings still load `{{first_name}}` as plain text (not chips). Pre-existing gap from Session 3.
- Toolbar has no "remove link" affordance. Pre-existing gap from Session 2.
- No ordered list (`<ol>`) support — StarterKit includes it but no toolbar button exists.

---

## Session 5 — Send Now fix + visual polish (Week 1 wrap-up)

**Goal:** Fix the broken Send Now flow, resolve the preview one-way-door UX bug, and clean up the visual awkwardness in the campaign editor so the whole thing feels shippable.

### Root causes identified

**Send Now bug — `saveDraft()` error swallowing:**
`saveDraft()` caught its own errors internally and returned normally regardless of outcome. When `submitSend()` called it, a failed save was invisible: for new campaigns `campaignId` stayed null and the modal showed a misleading "Please save the campaign first" message; for existing campaigns `submitSend` proceeded to POST `/send` even when the draft update had failed.

**Preview mode one-way door:**
`EmailEditor` was hidden with `v-show="!previewMode"` from the parent. The only "Edit" toggle button lived inside the EmailEditor toolbar — hidden along with it. Once the user clicked "Preview", there was no visible path back to the editor.

### What changed

**Modified:** `frontend/src/views/CampaignEditView.vue`

**Fix 1 — Send Now error propagation:**

- Extracted `doSave()` from `saveDraft()` — the raw async save logic with no catch block; throws on any API failure
- `saveDraft()` now calls `doSave()` and wraps it with catch (UI behavior unchanged for the Save Draft button)
- `submitSend()` calls `doSave()` directly — errors bubble into `submitSend()`'s own catch block, which already surfaces them as `sendError` in the confirmation modal. The old "Please save the campaign first" fallback is gone because `doSave()` throws instead of silently returning.

**Fix 2 — Preview stuck:**

- Added `.preview-header` wrapper inside `.preview-card` containing `.preview-meta` (from/subject/preheader) and a new `.preview-edit-btn` in the top-right corner
- Clicking "Edit" sets `previewMode = false` directly — no event bus, no prop, just a click handler on the parent's ref
- `.preview-meta` uses `gap: 5px` column layout; `.preview-header` gets `margin-bottom: 24px` replacing the former `margin-bottom` on `.preview-preheader`
- Added `Pencil` import from `lucide-vue-next`

**Fix 3 — Action bar restructure:**

- Replaced flat single-row layout with two groups: `.action-left` (Save Draft + Schedule + optional save error) and `.action-right` (optional "Select a list first" hint + Send Now)
- `justify-content: space-between` separates the draft operations from the send action visually
- `save-error` now renders inline in `.action-left` rather than with `flex: 1` pushing everything right — cleaner when no error is present
- `send-hint` (`"Select a list first"`) appears right of the Send Now button's left side only when `send_to_id` is empty

**Fix 4 — Template bar integration:**

- Moved `.template-bar` from top-level `.campaign-edit` to first child of `.editor-panel`
- Added `padding: 10px 14px`, `border-bottom: 1px solid var(--border)`, `background: #faf9f7` — matches the toolbar styling in EmailEditor; the panel now reads top-to-bottom as: template selector → editor/preview → action bar
- Removed `gap: 16px` from `.campaign-edit` (only `.editor-layout` remains as a direct child)

**Fix 5 — Variable chip button consistency:**

- Replaced `▾` unicode caret with `<ChevronDown :size="10" :stroke-width="2.5">` from lucide-vue-next in `EmailEditor.vue`
- Updated `.vars-caret` CSS: removed `font-size: 9px`, added `vertical-align: -2px` to align the SVG with surrounding text

### Architecture decisions

- **`doSave()` throws, `saveDraft()` catches, `submitSend()` propagates.** The simplest fix that preserves existing Save Draft behavior while giving `submitSend()` accurate error information. No return-value dance, no second error state.
- **Preview "Edit" button in the card, not the action bar.** The preview card is the artifact the user is looking at; the exit from preview should be co-located with it. The action bar already has enough going on.
- **Template bar inside `editor-panel`, not floating above it.** The template bar controls editor content, not page-level layout. Placing it inside the panel creates a clear visual hierarchy: everything in the right column belongs to editing. The settings panel on the left is unaffected.

### Known gaps carrying into Week 2

- No image support.
- Announcement template CTA button (`<a style="background:#10b981;color:#fff;...">`) loses button styling in Tiptap (Link extension drops arbitrary style attributes). Needs a dedicated button node extension.
- Template HTML strings load `{{first_name}}` as plain text (not chips). Pre-existing gap from Session 3.
- Toolbar has no "remove link" affordance. Pre-existing gap from Session 2.
- No ordered list (`<ol>`) toolbar button. Pre-existing gap.
- `submitSchedule()` still uses the old pattern of calling `saveDraft()` and then checking `campaignId` — not fixed this session since schedule is lower priority than send. Should be migrated to `doSave()` in Week 2.

---

## Week 1 Summary

**Completed across Sessions 1–5:**

| Session | What shipped                                                                                                                                              |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1       | Replaced `contenteditable` + `execCommand` with Tiptap; fully v-model compatible `EmailEditor` component                                                  |
| 2       | Toolbar moved inside `EmailEditor`; active states on all formatting buttons via `transactionCount` pattern; Cmd+K link insertion                          |
| 3       | Variable chip atoms (`{{first_name}}`, `{{unsubscribe_url}}`); atomic delete, round-trip serialization, chip rehydration on load                          |
| 4       | Email-safe HTML output via `DOMParser` pipeline in `toEmailHTML`; inline styles for Gmail/Outlook/Yahoo compatibility                                     |
| 5       | Send Now fixed (error propagation via `doSave()`); preview one-way-door fixed; action bar restructured; template bar integrated; toolbar icon consistency |

**Carries into Week 2:**

- Image support (node extension + upload endpoint)
- Ordered list toolbar button
- "Remove link" affordance in toolbar
- CTA button node extension (preserves full styling through Tiptap round-trip)
- Template strings pre-converted to chip format on load
- `submitSchedule()` migration to `doSave()`

---

## Hotfix — Tracking URL

**Date:** 2026-05-07

`SendWorker` was baking `INSTALLATION_URL` from the env var at startup and never re-reading it. Tracking pixels and click URLs used `localhost` in dev, and Settings UI changes to the installation URL had no effect until a server restart.

**Fix (`internal/worker/send_worker.go`):**

Added two lines in `processJob` before the batch loop — calls `w.db.GetSettings(ctx)` once per job and uses `s.InstallationUrl` as the URL passed to `BuildMessage`. Falls back to `w.installationURL` (the env-var value) if the query fails or the field is empty. One DB round-trip per campaign send; no per-recipient overhead. Worker lifecycle, SMTP, and `NewSendWorker` signature all untouched.

---

## Week 2 Session 1 — Email preview simulation

**Goal:** Replace raw `v-html="editorHTML"` preview with a proper email simulation — desktop and mobile views, realistic email shell, `toEmailHTML()` output so preview matches what gets sent.

### What changed

**Modified:** `frontend/src/views/CampaignEditView.vue`

- Added `previewDevice` ref (`'desktop' | 'mobile'`, default `'desktop'`)
- Added `previewEmailHTML` computed: `computed(() => toEmailHTML(editorHTML.value || ''))` — preview now renders exactly what the backend receives, inline styles and all
- Imported `Monitor` and `Smartphone` from lucide-vue-next
- Replaced old `.preview-card` / `.preview-header` / `.preview-meta` / `.preview-body` markup with:
  - `.preview-controls` bar (outside the email card): device toggle left, Edit button right
  - `.email-viewport` (handles width + phone frame) with `:class="\`device-${previewDevice}\`"`
    - `.device-desktop`: `max-width: 600px; border-radius: 8px; box-shadow: 0 4px 24px rgba(0,0,0,0.08)`
    - `.device-mobile`: `max-width: 375px; border: 8px solid #1e1e1e; border-radius: 40px`
  - `.phone-chrome` (mobile only): dark bar with "9:41" time and glyph icons (rendered via `v-if="previewDevice === 'mobile'"`)
  - `.email-shell`: white card containing `.shell-meta` / `.shell-divider` / `.email-body`
- `.device-toggle`: segmented pill control (sand background, active button lifts to white with shadow) — same visual language as template bar pills
- Variable chip `:deep` selector updated from `.preview-body .variable-chip` to `.email-body .variable-chip`

### Architecture decisions

- **`previewEmailHTML` is a computed, not a method call in the template.** Vue caches it; re-renders only when `editorHTML` changes. No redundant `toEmailHTML` calls while typing.
- **Controls bar outside the email card.** The card is the artifact; putting UI controls (toggle, edit button) inside it breaks the visual illusion. The bar floats above, controls beneath the gray background.
- **Phone chrome via `v-if`, not CSS.** The status bar row only exists in the DOM in mobile mode. No hidden-but-present markup; no CSS `display` toggling a decorative element.
- **`overflow: hidden` on `.email-viewport`.** Clips the dark border-radius corners of `.device-mobile` cleanly without needing border-radius on `.email-shell` itself.

### Known gaps carrying forward

- No image support.
- Announcement template CTA button loses styling through Tiptap round-trip (pre-existing, Session 4).
- Template strings load `{{first_name}}` as plain text, not chips (pre-existing, Session 3). ✅ Fixed in Week 2 Session 2.
- Toolbar has no "remove link" affordance (pre-existing, Session 2). ✅ Fixed in Week 2 Session 2.
- No ordered list toolbar button (pre-existing). ✅ Fixed in Week 2 Session 2.
- `submitSchedule()` still uses old `saveDraft()` pattern instead of `doSave()` (pre-existing, Session 5). ✅ Fixed in Week 2 Session 2.

---

## Week 2 Session 2 — Carry-forward editor gaps

**Goal:** Close four pre-existing gaps from Week 1: ordered list button, remove-link affordance, template/campaign variable chip rehydration, and `submitSchedule()` error propagation.

### What changed

**Modified:** `frontend/src/components/EmailEditor.vue`

- Added `Link2Off`, `ListOrdered` to lucide-vue-next import
- Added `isOrderedList` computed (same `transactionCount` pattern as `isBulletList`) — calls `editor.value?.isActive('orderedList')`
- **Remove link button:** added a `v-if="isLink"` button directly after the link button. Uses `Link2Off` icon. Calls `chain().focus().unsetLink().run()` inline — no modal, one click. Appears only when the cursor is inside a link (same condition as `isLink` active state).
- **Ordered list button:** added after the bullet list button, same `.tb-btn` class. Uses `ListOrdered` icon. Calls `toggleOrderedList()`.

**Modified:** `frontend/src/views/CampaignEditView.vue`

- **`TEMPLATE_HTML.newsletter`:** changed `{{first_name | "there"}}` to `{{first_name}}`. The Liquid-style fallback was aspirational — `email_builder.go` uses `strings.ReplaceAll` and has no fallback support. `{{first_name}}` is now a clean chip target.
- **`toEmailHTML`:** added `<ol>` transform after the `<ul>` block — `padding:0 0 0 20px;margin:0 0 14px 0;list-style-type:decimal;`. The existing `li > p` unwrap already covers ordered list items generically.
- **`fromEmailHTML(html)`:** new function — the inverse of `toEmailHTML`'s chip-strip step. Uses `DOMParser` + a recursive text-node walker that converts `{{variable_id}}` patterns (matching `[a-z_]+`) to `<span data-variable="..." class="variable-chip">` elements. Text-node-only walk means `href="{{unsubscribe_url}}"` attribute values are never touched — only visible text content is converted. Tiptap's existing `parseHTML` rule for `VariableChip` then rehydrates the spans as atom chip nodes.
- **`applyTemplate(id)`:** now calls `fromEmailHTML(TEMPLATE_HTML[id] || '')` — templates load with chips pre-inserted.
- **`loadCampaign()`:** `editorHTML.value` now set to `fromEmailHTML(c.html_body || '')` — campaigns loaded from the backend (which stores `{{first_name}}` plain text after `toEmailHTML` strip) now rehydrate their chips on load.
- **`submitSchedule()`:** replaced `if (!campaignId.value) { await saveDraft(); if (!campaignId.value) { error } }` with `await doSave()`. Errors from save now bubble into `submitSchedule()`'s catch block and surface as `scheduleError`. Behavior matches `submitSend()`.

### Architecture decisions

- **Remove-link as toolbar affordance, not modal action.** The link button going active signals "you're in a link." Adding a contextual `Link2Off` button right next to it is the fastest path: one click, no modal, no URL field to clear. The modal path (edit URL → clear → submit) still works for users who prefer it.
- **`fromEmailHTML` walks text nodes only.** A raw regex on the full HTML string would also match `href="{{unsubscribe_url}}"` and produce a malformed attribute. The DOMParser walk targets only `TEXT_NODE` entries — attribute values are never in scope. This is the same strategy `toEmailHTML` uses (DOMParser → DOM mutation → `innerHTML`).
- **`submitSchedule` always saves.** Unlike the old code which only called `saveDraft()` if `campaignId` was null, `doSave()` runs unconditionally. This ensures the latest form state is committed before a schedule is set — consistent with `submitSend()`.

### Keyboard shortcuts (unchanged)

| Action    | Shortcut                   |
| --------- | -------------------------- |
| Bold      | Ctrl+B / Cmd+B             |
| Italic    | Ctrl+I / Cmd+I             |
| Underline | Ctrl+U / Cmd+U             |
| Link      | Ctrl+K / Cmd+K             |
| Undo      | Ctrl+Z / Cmd+Z             |
| Redo      | Ctrl+Shift+Z / Cmd+Shift+Z |

### Known gaps carrying forward

- No image support.
- Announcement template CTA button loses styling through Tiptap round-trip (pre-existing, Session 4).

---

## Week 2 Session 3 — Bulk subscriber management

**Goal:** Checkbox selection on the subscribers table with bulk "add to list", CSV import into a specific list, and an AI-workflow hint near the import button.

### What changed

**New backend endpoint: `POST /api/lists/{id}/subscribers/bulk`** (`internal/handler/lists.go`)

- Body: `{ "subscriber_ids": ["uuid1", ...] }`, no cap (self-hosted tool, user affects only their own instance)
- Verifies the list exists, then loops calling `AddSubscriberToList` (SQL has `ON CONFLICT DO NOTHING`, so idempotent)
- Returns `{ "added": N }`
- Does not trigger double opt-in email — this is an admin bulk action, not a user-initiated subscription

**Modified: `POST /api/subscribers/import`** (`internal/handler/subscribers.go`)

- Accepts optional `list_id` multipart form field
- If provided and valid UUID, newly created subscribers are added to the list via `AddSubscriberToList`
- For duplicate-email rows (previously just `skipped++`): looks up the existing subscriber by email and adds them to the list silently — they still count as skipped in the response, but join the list
- Suppressed addresses: skipped for both import and list membership as before
- Response shape unchanged: `{ imported, skipped, invalid }`

**New route registered** (`cmd/server/main.go`):

- `r.Post("/api/lists/{id}/subscribers/bulk", listHandler.BulkAddSubscribers)`

**Modified: `frontend/src/views/SubscribersView.vue`**

- **Checkbox column:** prepended `{ key: '_select', label: '' }` to `cols`; rendered via `#cell-_select` slot as `<input type="checkbox" @click.stop @change="toggleSelect(row.id)">`. The `@click.stop` prevents the row-click navigation from firing. The `<th>` label is empty — no BaseTable changes needed.
- **Selection state:** `selectedIds = ref(new Set())`. Mutations always replace with a new Set so Vue's reactivity system detects changes correctly. Selection is cleared on page/tab change.
- **Selection bar:** appears between the action bar and `BaseCard` when `selectedIds.size > 0`. Shows: selected count · list dropdown + "Add to List" button · "Select all (N)" link · "Deselect all" link · success message. Green-tinted (`#f0fdf9` background, `#d1fae5` border) to be clearly readable without being alarming.
- **Bulk add:** calls `POST /api/lists/{id}/subscribers/bulk` with all selected IDs; on success shows inline `"Added N to list."` and clears selection.
- **Import modal — list picker:** added "Add to list (optional)" `<select>` above the file dropzone, populated from `GET /api/lists` (fetched on mount). If a list is selected, `list_id` is appended to FormData before POST.
- **Import modal — hint:** _"Already have subscribers? Export your list, filter with AI, and import back."_ appears below the dropzone in italic muted text.
- **Lists fetched on mount:** `fetchLists()` called alongside `fetchSubscribers()` and `fetchOverview()`. Lists are reused for both the selection bar dropdown and the import modal picker.
- **Skeleton row update:** added a `sk-check` skeleton block as the first element in skeleton rows to match the new checkbox column.

**Modified: `frontend/src/views/ListDetailView.vue`**

- **Import CSV button:** added to `.table-header-row` alongside the existing "Add Subscriber" button, inside a new `.table-header-actions` flex wrapper.
- **Import modal:** inline modal with the same dropzone/preview/result/error UI as `SubscribersView`. List ID is pre-set to `route.params.id` — no list picker shown since the context is already a specific list.
- **Hint text:** same _"Already have subscribers? Export your list, filter with AI, and import back."_ line below the dropzone.
- After import: calls both `fetchList()` (updates the subscriber count stat card) and `fetchSubscribers()` (refreshes the table).
- `Upload` icon added to lucide import.

### Architecture decisions

- **`new Set(...)` replacement pattern for reactivity.** `Set` mutations (`.add()`, `.delete()`) don't trigger Vue's reactivity system because the ref identity doesn't change. Replacing with `selectedIds.value = new Set(selectedIds.value)` forces a reactive update. Alternative (a plain array with `.includes()`) would work but is O(n) for the `:checked` binding on every row render.
- **"Select all" in the selection bar, not a header checkbox.** `BaseTable` renders `col.label` as plain text in `<th>` — there is no header slot. Adding one would require modifying the shared component for a single use case. The selection bar "Select all (N)" link covers current-page selection cleanly without touching `BaseTable`.
- **Bulk add skips double opt-in.** The single-subscriber `AddSubscriber` endpoint sends a confirmation email for double opt-in lists. Bulk add bypasses this — it's an admin operation. Consistent with how `BulkTag` works.
- **Import + existing subscribers join the list.** When importing a CSV into a list, subscribers whose emails already exist (which would otherwise just be `skipped`) are still added to the list. This supports the primary use case: import a filtered export back into a new list segment.
- **`list_id` is optional on the import endpoint.** The existing global import behavior (no list assignment) is preserved when no `list_id` is provided.
- **CSV parsing logic is duplicated in `ListDetailView`, not extracted.** The `parseCSVPreview` function is ~20 lines and appears in both views. A shared utility would be premature for two usages; if a third view needs it, extract then.

### Known gaps carrying forward

- Announcement template CTA button loses styling through Tiptap round-trip (pre-existing, Session 4).
- Bulk select is page-scoped (max 50 per page). Cross-page selection would require tracking all matching subscriber IDs server-side, which is out of scope.

---

## Week 2 Session 4 — Image support

**Goal:** Image upload endpoint on the Go backend, Tiptap Image extension in the editor, toolbar button, preview rendering, and email-safe serialization.

### What changed

**New file: `internal/handler/uploads.go`**

- `Storage` interface — `Save(filename string, r io.Reader) (url string, err error)`. Clean seam for a future R2/S3 backend without changing the handler.
- `LocalStorage` — writes to a configurable directory (`UPLOAD_DIR` env var, default `./uploads`); returns a **root-relative URL** (`/uploads/images/<hex>.<ext>`). Absolute URL construction is deferred to the caller (preview pane) and to send time (BuildMessage) so changing `INSTALLATION_URL` doesn't break stored paths.
- `UploadHandler.UploadImage`:
  1. Enforces 5 MB multipart limit
  2. Sniffs actual MIME type via `http.DetectContentType` (first 512 bytes) — does not trust the client-supplied Content-Type
  3. Accepts `image/jpeg`, `image/png`, `image/gif`, `image/webp`; rejects everything else with 415
  4. Seeks back to byte 0 after sniffing so the full file is written
  5. Generates a 32-char random hex filename (no uuid dependency)
  6. Returns `{ "url": "/uploads/images/<hex>.<ext>" }`
- Uses existing `writeJSON` / `writeError` helpers from `helpers.go`

**Modified: `internal/config/config.go`**

- Added `UploadDir string` field
- Reads `UPLOAD_DIR` env var; defaults to `./uploads`

**Modified: `cmd/server/main.go`**

- `uploadStorage := handler.NewLocalStorage(cfg.UploadDir + "/images")` — path is `./uploads/images/`
- `uploadHandler := handler.NewUploadHandler(uploadStorage)`
- `r.Post("/api/uploads/images", uploadHandler.UploadImage)` inside the `RequireAuth` group
- `r.Get("/uploads/*", ...)` outside the auth group — `http.StripPrefix` + `http.FileServer(http.Dir(uploadDir))` serves uploaded files publicly
- Added `/uploads/` to `setupGuard` skip list so image requests don't redirect to `/setup`

**Modified: `internal/worker/email_builder.go`**

- `relativeImgSrc` regex: `(<img\s[^>]*src=")(/uploads/[^"]+)(")`
- In `BuildMessage`, before `RewriteLinks`: replaces all matching `src="/uploads/..."` with `src="<installationURL>/uploads/..."` using the current `installationURL` (read from settings DB at send time — same pattern as the tracking URL hotfix). Stored HTML stays portable; absolute URL is only materialized at the moment an email is actually sent.

**Modified: `docker-compose.yml`**

- Added `uploads_data:/app/uploads` volume to the `app` service and `uploads_data:` to the top-level `volumes` block — uploaded images persist across container restarts.

**Modified: `frontend/src/components/EmailEditor.vue`**

- Installed `@tiptap/extension-image 3.22.5`
- `TiptapImage.configure({ inline: false, allowBase64: false, HTMLAttributes: { style: 'max-width:100%;height:auto;display:block;' } })` added to extensions
- `ImageIcon` (aliased from lucide `Image`) and `Loader` added to lucide import
- `api` imported from `@/lib/api` for authenticated upload
- `imageInputRef` ref + hidden `<input type="file" accept="image/jpeg,image/png,image/gif,image/webp">` in template
- `isUploading` ref drives disabled state and spinner on the toolbar button
- `onImageFileSelected(e)` — POSTs FormData to `/api/uploads/images` via `api.post`, inserts `{ type: 'image', attrs: { src: res.data.url } }` at cursor via `editor.commands.setImage`. Resets input value so the same file can be re-selected.
- Toolbar: separator + image button (camera icon) with `Loader` spinner while uploading; placed after the ordered list button, before the toolbar spacer
- CSS: `:deep(.ProseMirror img)` — `max-width:100%;height:auto;display:block;border-radius:4px;margin:8px 0;`. Selection ring via `:deep(.ProseMirror img.ProseMirror-selectednode)` — 2px accent outline. Spinner keyframe animation on `.spin`.

**Modified: `frontend/src/views/CampaignEditView.vue`**

- `installationURL` ref — populated by `loadSettings()` from `s.installation_url` (already in settings API response)
- `toEmailHTML(html, baseURL = '')` — added optional second parameter. When `baseURL` is provided (preview pane), image `src` attributes starting with `/` are prepended with `baseURL` before the inline style transform. The DB-save path calls `toEmailHTML(editorHTML.value)` with no `baseURL` — stored `html_body` retains relative paths.
- Image transform step (after HR): sets `max-width:100%`, `height:auto`, `display:block`, `margin:0 auto 12px` (preserving any existing `margin`) on every `<img>` element
- `previewEmailHTML` computed now passes `installationURL.value` as `baseURL` so images render in the preview pane even when running the Vite dev server on a different port than the Go backend

### Architecture decisions

- **`Storage` interface at the handler layer, not config.** The interface lives in `handler/uploads.go` alongside its only current implementation. A future `R2Storage` would be instantiated in `main.go` based on an env var (`UPLOAD_BACKEND=r2`) and passed to `NewUploadHandler` — no change to the handler itself.
- **Root-relative URLs stored in DB, absolute only at use time.** `installationURL` is mutable (user changes it in Settings). Storing absolute URLs would silently break all historical images after a domain change. Root-relative paths (`/uploads/images/...`) remain valid regardless of domain.
- **`http.DetectContentType` sniff, not MIME parameter trust.** Clients can send any `Content-Type` header. Reading the first 512 bytes and using Go's built-in sniff logic is the authoritative check.
- **`toEmailHTML` baseURL only for preview, not for DB save.** `doSave()` calls `toEmailHTML(editorHTML.value)` without `baseURL`; `previewEmailHTML` calls it with `installationURL.value`. This keeps stored HTML portable while making the preview pane work correctly in dev (Vite on 5173, Go on 4400).
- **Regex absolutization in `BuildMessage`, not a DOM parser.** `email_builder.go` is pure Go with no HTML-parse dependency. The pattern `/uploads/[^"]+` is narrow enough to be safe — it only matches paths inside the uploads directory. The pixel URL (`/track/open/...`) would never match because it uses `src="<absolute URL>"` and is added after the regex runs.

### Known gaps carrying forward

- No image resize in the editor — Tiptap's Image extension renders at `max-width:100%` but drag-to-resize needs `@tiptap/extension-image-resize`. Deferred.
- No image library/gallery — each upload is a one-shot insert with no way to browse or re-use uploaded images.
- Announcement template CTA button loses styling through Tiptap round-trip (pre-existing, Session 4). ✅ Fixed in Week 2 Session 5.
- Bulk select is page-scoped (pre-existing, Session 3).

---

## Week 2 Session 5 — CTA button node extension

**Goal:** Fix the announcement template's styled CTA button (`<a style="background:#10b981;...;display:inline-block">`) losing all its styling when loaded into Tiptap. The Link extension (a mark) only preserves `href`/`target` — a dedicated inline atom node is required to round-trip the full button appearance.

### Root cause

The `Link` extension is a **mark** — it annotates text ranges and only captures `href` + `target`. When the announcement template's `<a style="background:...;padding:...;border-radius:...">` was parsed by Tiptap, all style attributes were stripped and the button became a plain link. A mark cannot carry arbitrary presentation state; a **node** can.

### What changed

**New file: `frontend/src/components/CTAButtonNode.vue`**

- VueNodeView component for the CTA button atom
- `<NodeViewWrapper as="span">` — inline wrapper, preserves flow
- Renders as a styled green pill button that visually matches the email output (same background, border-radius, font-weight)
- Hover reveals a pencil icon indicating editability
- Click opens an edit modal via `<Teleport to="body">` — teleport avoids z-index and clip issues inside ProseMirror's DOM tree; Vue 3 scoped styles still apply to teleported nodes via the `data-v-*` attribute mechanism
- Edit modal: two fields — Button text (label) and Link URL (href)
- Submit calls `updateAttributes({ href, label })` from VueNodeViewRenderer props — no external state or callbacks needed
- `selected` prop drives `.cta-preview--sel` accent outline (same pattern as VariableChip)

**Modified: `frontend/src/components/EmailEditor.vue`**

- Imported `CTAButtonNodeView`
- Defined `CTAButton` inline node extension (same file as `VariableChip`, after it):
  - `group: 'inline'`, `inline: true`, `atom: true`
  - `addAttributes()` → `href` (default `'#'`), `label` (default `'Learn More →'`)
  - `parseHTML()` → matches `a[data-cta]`; `getAttrs` reads `href` attribute and `textContent` as label
  - `renderHTML()` → `['a', { 'data-cta': '', href, style: 'background:#10b981;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;display:inline-block;' }, label]` — full button HTML baked into the node's serialization
  - `addNodeView()` → `VueNodeViewRenderer(CTAButtonNodeView)`
- `CTAButton` added to `extensions` array after `VariableChip`
- `insertCTA()` function — `chain().focus().insertContent({ type: 'ctaButton', attrs: { href: '#', label: 'Learn More →' } })`
- Toolbar: `CTA` pill button (`.tb-pill`) added after the `{ }` vars dropdown and before the image separator

**Modified: `frontend/src/views/CampaignEditView.vue`**

- `TEMPLATE_HTML.announcement` — added `data-cta` attribute to the existing `<a>`: `<a data-cta href="#" style="...">`. The `data-cta` marker is what the `CTAButton` extension's `parseHTML` rule matches; it's harmless in email clients (unknown attributes are silently ignored).
- `fromEmailHTML()` — added migration step before the text-node walk: detects `<a href>` elements with `display:inline-block` + a `background` style and adds `data-cta` to them. Handles any campaigns saved with the pre-Session 5 announcement template without `data-cta`.

### Architecture decisions

- **Node, not mark.** A mark applies to a text range and only captures attributes whitelisted by the extension. A node owns its entire representation — `renderHTML` outputs whatever HTML is needed, and `parseHTML` reconstructs the node's attributes on load. Full round-trip fidelity without touching the Link extension at all.
- **`data-cta` persists in `html_body`.** Unlike variable chips (which are stripped to `{{var}}` in `toEmailHTML`), the `data-cta` attribute is kept in the stored email HTML. This is intentional: it's the rehydration marker. Email clients ignore unknown attributes. The alternative (stripping it and using `fromEmailHTML` heuristics exclusively) would be fragile — `fromEmailHTML`'s `display:inline-block` detection covers old saves, but new campaigns rely on `data-cta` for clean rehydration.
- **Edit modal in the NodeView, not in `EmailEditor`.** The NodeView has direct access to `updateAttributes` via VueNodeViewRenderer props. Co-locating the modal in the NodeView means `EmailEditor.vue` needs no new state, no new refs, and no event plumbing. Consistent with how the link modal lives inside `EmailEditor` rather than in the parent campaign view — state owned closest to where it's used.
- **`Teleport to="body"` for the modal.** The NodeView renders inside ProseMirror's content-editable DOM. A fixed-position modal inside it would be clipped by overflow rules or have stacking context issues. Teleporting to body guarantees it renders at the top layer regardless of the ProseMirror DOM hierarchy.
- **Button color locked to `#10b981`.** No color picker. The brand accent is baked into `renderHTML`'s style string and the NodeView's preview styles. One less field in the edit modal; consistent with the app's existing green-on-white aesthetic.
- **`toEmailHTML` unchanged.** The existing link color guard (`if (!el.style.color) el.style.color = '#10b981'`) already skips `a[data-cta]` since `color:#fff` is explicitly set in the button's inline style. No other `toEmailHTML` step touches the button. The paragraph it lives in gets `margin:0 0 12px 0` as normal.

### Round-trip verification

| Step                       | Value                                                                                                                                       |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Template string            | `<a data-cta href="#" style="background:#10b981;color:#fff;...">Learn More →</a>`                                                           |
| `fromEmailHTML`            | adds `data-cta` if missing (migration); passes through unchanged if already present                                                         |
| Tiptap `parseHTML`         | matches `a[data-cta]` → `ctaButton` node with `href="#"`, `label="Learn More →"`                                                            |
| Editor display             | styled green button (NodeView), pencil icon on hover                                                                                        |
| User edits label/href      | `updateAttributes` → node attrs updated                                                                                                     |
| `getHTML()` → `editorHTML` | `<p><a data-cta href="https://..." style="...">Shop Now</a></p>`                                                                            |
| `toEmailHTML`              | paragraph gets `margin:0 0 12px 0`; link color guard skips (color already set)                                                              |
| `html_body` in DB          | `<p style="margin:0 0 12px 0;"><a data-cta href="https://..." style="background:#10b981;color:#fff;padding:12px 24px;...">Shop Now</a></p>` |
| Reload campaign            | `fromEmailHTML` → `data-cta` present → Tiptap `parseHTML` → `ctaButton` rehydrated ✓                                                        |

### Known gaps carrying forward

- No image resize / gallery (pre-existing, Session 4).
- Bulk select is page-scoped (pre-existing, Session 3).

---

## Week 2 Session 6 — Send + Schedule flow polish + SMTP wizard fixes

**Goal:** Make the send flow feel confident end-to-end; make scheduling feel intentional; fix two silent data-loss bugs in the SMTP wizard.

### Root causes identified

**SMTP wizard: "Continue →" on step 3 didn't save credentials**
`goNext()` had branches for steps 1, 2, and 4 — nothing for step 3. Clicking Continue on the SMTP step just incremented `currentStep` without calling the API. SMTP credentials were only persisted if the user clicked "Test connection" first. Users who filled in their provider/API key and clicked Continue (without testing) lost all their SMTP config silently and had to reconfigure it in Settings after setup.

**`minDatetime` used UTC, not local time**
`minDatetime` returned `now.toISOString().slice(0, 16)` — UTC format. The `datetime-local` input's `min` attribute and its own value are both in local browser time. For users in UTC+ zones, the min constraint was in the past (allowing past scheduling) and the pre-filled value showed the wrong time. The `new Date(scheduleAt.value).toISOString()` conversion back to UTC was correct; only the min and initial value were wrong.

**Schedule modal had no recipient list guard**
The Schedule button had `:disabled="!campaignId && saveLoading"` but no check for `send_to_id`. A user could open and submit the schedule modal with no list or tag selected. `doSave()` would succeed (drafts don't require a list), the schedule API would create a job, and the send worker would expand zero recipients — campaign scheduled, nobody gets it, no error shown.

**No pre-flight validation before Send Now**
`openSendConfirm()` opened the confirmation modal unconditionally. An empty campaign name or empty subject would pass through, hit `doSave()`, and if the backend accepted blanks, go out as an untitled/no-subject send.

### What changed

**Modified: `frontend/src/views/SetupWizardView.vue`**

- **Fix 1 — SMTP step save in `goNext()`:** Added `else if (currentStep.value === 3)` branch that calls `await api.put('/api/setup/smtp', { provider, from_email, from_name, credentials: buildSmtpCredentials() })` before incrementing. Now clicking "Continue →" on the SMTP step persists credentials exactly as "Test connection" did. Errors (bad API key, network failure) surface via `stepError` — the Continue button already has loading state via `stepLoading`. "Skip for now" still bypasses this call intentionally — skipping means skip.
- **Fix 9 — Footer layout:** Extracted `[Skip for now]` and `[Continue →]` into a `<div class="footer-btns">` row. Moved the continue-hint `<span>` below that row. CSS: `.footer-right` is now `flex-direction: column; align-items: flex-end` with a new `.footer-btns` flex row. The hint no longer reads as content between two buttons.

**Modified: `frontend/src/views/CampaignEditView.vue`**

- **Fix 2 — `minDatetime` local time:** Replaced `now.toISOString().slice(0, 16)` with a manual local format: `${year}-${month}-${day}T${hours}:${mins}` using `getFullYear/getMonth/getDate/getHours/getMinutes`. Also switched to `new Date(Date.now() + 5 * 60000)` to compute 5-minutes-from-now atomically. The downstream `new Date(scheduleAt.value).toISOString()` conversion to UTC for the API was already correct and is unchanged.
- **Fix 3 + 7 — Schedule list guard + button disable:** `submitSchedule()` now checks `!form.send_to_id` first and returns with `scheduleError = 'Select a recipient list or tag before scheduling.'` The Schedule button in the action bar now uses `:disabled="!form.send_to_id"` (matching the Send Now button), replacing the old `!campaignId && saveLoading` condition which was both semantically wrong and insufficient.
- **Fix 4 — Send confirm modal summary:** Added a `.confirm-summary` box between the main confirm-text and the "cannot be undone" line. Shows two rows: `Subject` (value or "(no subject)") and `From` (from_name + from_email). Lets the user verify what they're actually sending before the irreversible click.
- **Fix 5 — Schedule modal summary:** Added a `.schedule-summary` card at the top of the schedule form. Shows `form.name` (bold) and a subtitle combining `form.subject` · subscriber count. Users see what they're scheduling before picking the time.
- **Fix 6 — Timezone label:** Added `localTimezoneLabel` computed — reads `new Date().getTimezoneOffset()`, formats as `UTC+HH:MM` or `UTC-HH:MM`. Displayed as a `field-hint` below the `datetime-local` input: "Your local time (UTC+05:00)". No ambiguity about what timezone the picker uses.
- **Fix 7:** See Fix 3 above (same edit).
- **Fix 8 — Pre-flight validation before send:** `openSendConfirm()` now checks `form.name.trim()` and `form.subject.trim()` before opening the modal. On failure: sets `sendPreflight` ref, returns early (modal stays closed). `sendPreflight` displays as `.send-error-hint` (red) in `.action-right`, using `v-if="sendPreflight"` before the existing "Select a list first" hint. `submitSend()` also validates at the start (belt-and-suspenders) so even if the modal is reached with bad state, the submit button is blocked with a clear error in-modal.

### Architecture decisions

- **`goNext()` SMTP save uses same payload as `testSmtp()`.** Both paths call `buildSmtpCredentials()` — the helper is already in scope and handles all three providers. No duplication; both paths now produce the same DB state.
- **`minDatetime` as local format.** `datetime-local` is a local-time input throughout: its displayed value, its min constraint, and its parsed output (parsed as local time by `new Date(string)` when there's no timezone indicator). Making minDatetime local keeps all three aligned. The UTC conversion only happens at the boundary where data leaves the frontend (`new Date(scheduleAt.value).toISOString()`).
- **`sendPreflight` is separate from `sendError`.** `sendError` lives inside the modal (set during/after modal open). `sendPreflight` lives in the action bar and prevents modal open. Keeping them separate means the modal error area is never polluted with pre-modal validation noise, and the action bar hint is not confused with save errors.
- **Schedule disabled same condition as Send Now.** Both now use `!form.send_to_id`. This enforces the invariant that both delivery actions require a target, consistently. The "Select a list first" hint in `.action-right` provides the explanation for why both are inactive.
- **SMTP Continue saves even with empty fields.** If a user clicks Continue on the SMTP step without filling anything in, `PUT /api/setup/smtp` is called with empty/default values. The backend will store the default (Resend, no API key). This is no worse than what happened before (nothing stored). Users who skip SMTP entirely use "Skip for now" — that path is unchanged.

### Known gaps carrying forward

- No image resize / gallery (pre-existing, Session 4).
- Bulk select is page-scoped (pre-existing, Session 3).

---

## Week 2 Summary

**Sessions 1–6 completed across Week 2:**

| Session | What shipped                                                                                                                                                                                                                           |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| W2 S1   | Email preview simulation — desktop/mobile device frames, `toEmailHTML()` output in preview, device toggle                                                                                                                              |
| W2 S2   | Carry-forward editor gaps — ordered list button, remove-link button, variable chip rehydration on template/campaign load, `submitSchedule()` migrated to `doSave()`                                                                    |
| W2 S3   | Bulk subscriber management — checkbox selection, bulk add-to-list, CSV import with list assignment, import hint                                                                                                                        |
| W2 S4   | Image upload — Go `Storage` interface + `LocalStorage`, `POST /api/uploads/images`, Tiptap Image extension, email-safe absolutization at send time                                                                                     |
| W2 S5   | CTA button node extension — atom node with NodeView, full round-trip preservation of styled buttons, `fromEmailHTML` migration for old saves                                                                                           |
| W2 S6   | Send + schedule flow polish — SMTP wizard Continue saves credentials, schedule datetime fixed for non-UTC zones, schedule/send list guard, send confirm + schedule modals show campaign summary, timezone label, pre-flight validation |

**Remaining known gaps entering Week 3:**

- No image resize or gallery (draggable resize needs `@tiptap/extension-image-resize`)
- Bulk select is page-scoped (cross-page selection requires server-side ID tracking)
- Announcement template CTA loses styling through Tiptap round-trip if `data-cta` is absent (pre-existing; `fromEmailHTML` migration handles old saves)
