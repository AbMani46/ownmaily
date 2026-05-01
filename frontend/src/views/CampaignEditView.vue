<template>
  <div class="campaign-edit">
    <!-- Template selector -->
    <div class="template-bar">
      <span class="template-label">Template:</span>
      <button
        v-for="t in templates"
        :key="t.id"
        class="template-pill"
        :class="{ active: selectedTemplate === t.id }"
        @click="confirmApplyTemplate(t.id)"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- Two-column layout -->
    <div class="editor-layout">
      <!-- Left: Settings panel -->
      <div class="settings-panel">
        <div class="settings-inner">
          <div class="form-field">
            <label class="field-label">Campaign name <span class="required">*</span></label>
            <BaseInput v-model="form.name" placeholder="e.g. May Newsletter" />
          </div>

          <div class="divider" />

          <div class="form-field">
            <label class="field-label">Subject line <span class="required">*</span></label>
            <BaseInput v-model="form.subject" placeholder="What's in it for them?" />
          </div>
          <div class="form-field">
            <label class="field-label">Preview text</label>
            <BaseInput v-model="form.preview_text" placeholder="Short teaser shown in inbox" />
            <span class="field-hint">Appears after the subject in most email clients</span>
          </div>

          <div class="divider" />

          <div class="form-field">
            <label class="field-label">From name <span class="required">*</span></label>
            <BaseInput v-model="form.from_name" placeholder="Your Name" />
          </div>
          <div class="form-field">
            <label class="field-label">From email <span class="required">*</span></label>
            <BaseInput v-model="form.from_email" placeholder="you@example.com" type="email" />
          </div>
          <div class="form-field">
            <label class="field-label">Reply-to</label>
            <BaseInput v-model="form.reply_to" placeholder="replies@example.com" type="email" />
          </div>

          <div class="divider" />

          <!-- Send to -->
          <div class="form-field">
            <label class="field-label">Send to</label>
            <div class="radio-row">
              <label class="radio-label">
                <input type="radio" value="list" v-model="form.send_to_type" @change="form.send_to_id = ''" />
                List
              </label>
              <label class="radio-label">
                <input type="radio" value="tag" v-model="form.send_to_type" @change="form.send_to_id = ''" />
                Tag
              </label>
            </div>

            <select
              v-model="form.send_to_id"
              class="select-input"
            >
              <option value="">
                {{ form.send_to_type === 'tag' ? 'Select a tag…' : 'Select a list…' }}
              </option>
              <template v-if="form.send_to_type === 'list'">
                <option v-for="l in lists" :key="l.id" :value="l.id">
                  {{ l.name }} ({{ (l.subscriber_count || 0).toLocaleString() }})
                </option>
              </template>
              <template v-else>
                <option v-for="t in tagsList" :key="t.id" :value="t.id">
                  {{ t.name }} ({{ (t.subscriber_count || 0).toLocaleString() }})
                </option>
              </template>
            </select>

            <div v-if="selectedTargetCount != null" class="recipient-info">
              <strong class="recipient-count">{{ selectedTargetCount.toLocaleString() }}</strong>
              subscribers will receive this campaign
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Editor -->
      <div class="editor-panel">
        <!-- Toolbar -->
        <div class="toolbar">
          <button class="tb-btn" title="Bold" @mousedown.prevent @click="execCmd('bold')"><strong>B</strong></button>
          <button class="tb-btn tb-italic" title="Italic" @mousedown.prevent @click="execCmd('italic')"><em>I</em></button>
          <button class="tb-btn tb-underline" title="Underline" @mousedown.prevent @click="execCmd('underline')"><u>U</u></button>

          <div class="toolbar-sep" />

          <button class="tb-pill" @mousedown.prevent @click="execCmd('formatBlock', 'h1')">H1</button>
          <button class="tb-pill" @mousedown.prevent @click="execCmd('formatBlock', 'h2')">H2</button>
          <button class="tb-pill" @mousedown.prevent @click="execCmd('formatBlock', 'h3')">H3</button>

          <div class="toolbar-sep" />

          <button class="tb-btn" title="Link" @mousedown.prevent @click="openLinkModal">
            <Link2 :size="13" color="#555" />
          </button>
          <button class="tb-btn" title="Unordered list" @mousedown.prevent @click="execCmd('insertUnorderedList')">
            <List :size="13" color="#555" />
          </button>

          <div class="toolbar-spacer" />

          <button class="preview-toggle" @click="previewMode = !previewMode">
            <Eye :size="12" />
            {{ previewMode ? 'Edit' : 'Preview' }}
          </button>
        </div>

        <!-- Content area -->
        <div class="editor-content" :class="{ 'preview-bg': previewMode }">
          <!-- Edit mode -->
          <div
            v-show="!previewMode"
            ref="editorRef"
            class="editor-area"
            contenteditable="true"
            data-placeholder="Start writing your email…"
          />
          <!-- Preview mode -->
          <div v-show="previewMode" class="preview-wrap">
            <div class="preview-card">
              <div class="preview-from">From: {{ form.from_name }} &lt;{{ form.from_email }}&gt;</div>
              <div class="preview-subject">{{ form.subject || 'No subject yet' }}</div>
              <div class="preview-preheader">{{ form.preview_text || 'No preview text' }}</div>
              <div class="preview-body" v-html="editorHTML" />
            </div>
          </div>
        </div>

        <!-- Bottom action bar -->
        <div class="action-bar">
          <p v-if="saveError" class="save-error">{{ saveError }}</p>
          <div class="action-buttons">
            <BaseButton variant="secondary" @click="saveDraft" :loading="saveLoading">
              <template v-if="saveSuccess">
                <span class="saved-indicator">✓ Saved</span>
              </template>
              <template v-else>
                Save Draft
              </template>
            </BaseButton>
            <BaseButton variant="secondary" @click="openSchedule" :disabled="!campaignId && saveLoading">
              <CalendarDays :size="13" :stroke-width="2" />
              Schedule
            </BaseButton>
            <BaseButton variant="primary" @click="openSendConfirm" :disabled="!form.send_to_id">
              <SendHorizonal :size="13" :stroke-width="2" />
              Send Now
            </BaseButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Insert Link Modal -->
    <BaseModal :show="showLinkModal" title="Insert Link" @close="closeLinkModal">
      <form @submit.prevent="submitLink" class="form">
        <div class="form-field">
          <label class="field-label">URL</label>
          <BaseInput v-model="linkUrl" ref="linkInputRef" placeholder="https://example.com" type="url" />
        </div>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeLinkModal">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit">Insert Link</BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- Template Overwrite Confirmation Modal -->
    <BaseModal :show="showTemplateConfirm" title="Replace editor content?" @close="showTemplateConfirm = false">
      <div class="send-confirm">
        <p class="confirm-text">
          Applying the <strong>{{ pendingTemplateLabel }}</strong> template will replace your current content. This cannot be undone.
        </p>
        <div class="form-actions">
          <BaseButton variant="ghost" @click="showTemplateConfirm = false">Cancel</BaseButton>
          <BaseButton variant="primary" @click="applyPendingTemplate">Apply Template</BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- Schedule Modal -->
    <BaseModal :show="showSchedule" title="Schedule Campaign" @close="showSchedule = false">
      <form @submit.prevent="submitSchedule" class="form">
        <div class="form-field">
          <label class="field-label">Send at <span class="required">*</span></label>
          <input
            v-model="scheduleAt"
            type="datetime-local"
            class="datetime-input"
            :min="minDatetime"
          />
        </div>
        <p v-if="scheduleError" class="form-error">{{ scheduleError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="showSchedule = false">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="scheduleLoading">Schedule</BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- Send Confirmation Modal -->
    <BaseModal :show="showSendConfirm" title="Send Campaign" @close="showSendConfirm = false">
      <div class="send-confirm">
        <p class="confirm-text">
          Send <strong>{{ form.name || 'this campaign' }}</strong> to
          <strong>{{ selectedTargetCount?.toLocaleString() ?? '?' }}</strong>
          subscriber{{ selectedTargetCount !== 1 ? 's' : '' }}?
        </p>
        <p class="confirm-sub">This action cannot be undone.</p>
        <p v-if="sendError" class="form-error">{{ sendError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" @click="showSendConfirm = false">Cancel</BaseButton>
          <BaseButton variant="primary" @click="submitSend" :loading="sendLoading">
            <SendHorizonal :size="13" :stroke-width="2" />
            Send Now
          </BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Link2, List, Eye, CalendarDays, SendHorizonal } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseInput from '@/components/BaseInput.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseModal from '@/components/BaseModal.vue'

const route = useRoute()
const router = useRouter()

const isNew = computed(() => !route.params.id)
const campaignId = ref(route.params.id || null)

const form = ref({
  name: '',
  subject: '',
  preview_text: '',
  from_name: '',
  from_email: '',
  reply_to: '',
  send_to_type: 'list',
  send_to_id: '',
})

const lists = ref([])
const tagsList = ref([])

const editorRef = ref(null)
const editorHTML = ref('')
const previewMode = ref(false)
const selectedTemplate = ref('blank')

const saveLoading = ref(false)
const saveError = ref('')
const saveSuccess = ref(false)

const showSchedule = ref(false)
const scheduleLoading = ref(false)
const scheduleAt = ref('')
const scheduleError = ref('')

const showSendConfirm = ref(false)
const sendLoading = ref(false)
const sendError = ref('')

const showLinkModal = ref(false)
const linkUrl = ref('')
let savedSelection = null

const showTemplateConfirm = ref(false)
const pendingTemplate = ref(null)

const templates = [
  { id: 'blank', label: 'Blank' },
  { id: 'newsletter', label: 'Newsletter' },
  { id: 'announcement', label: 'Announcement' },
]

const pendingTemplateLabel = computed(() => {
  const t = templates.find(t => t.id === pendingTemplate.value)
  return t?.label ?? ''
})

const TEMPLATE_HTML = {
  blank: '<p></p>',
  newsletter: `<h2>Hello {{first_name | "there"}},</h2>
<p>Here's what's been happening this month.</p>
<h3>What's new</h3>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
<hr />
<p style="color:#888;font-size:13px">You're receiving this because you subscribed. <a href="{{unsubscribe_url}}">Unsubscribe</a></p>`,
  announcement: `<h2>Exciting news!</h2>
<p>We're thrilled to share something special with you.</p>
<p><a href="#" style="background:#10b981;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;display:inline-block">Learn More →</a></p>
<hr />
<p style="color:#888;font-size:13px">You're receiving this because you subscribed. <a href="{{unsubscribe_url}}">Unsubscribe</a></p>`,
}

const minDatetime = computed(() => {
  const now = new Date()
  now.setMinutes(now.getMinutes() + 5)
  return now.toISOString().slice(0, 16)
})

const selectedTargetCount = computed(() => {
  if (!form.value.send_to_id) return null
  if (form.value.send_to_type === 'list') {
    const l = lists.value.find(x => x.id === form.value.send_to_id)
    return l?.subscriber_count ?? null
  } else {
    const t = tagsList.value.find(x => x.id === form.value.send_to_id)
    return t?.subscriber_count ?? null
  }
})

function execCmd(cmd, value = null) {
  document.execCommand(cmd, false, value)
  editorRef.value?.focus()
  syncEditorHTML()
}

function openLinkModal() {
  // Save current selection so it survives the modal opening
  const sel = window.getSelection()
  savedSelection = sel && sel.rangeCount > 0 ? sel.getRangeAt(0).cloneRange() : null
  linkUrl.value = ''
  showLinkModal.value = true
}

function closeLinkModal() {
  showLinkModal.value = false
  linkUrl.value = ''
  savedSelection = null
}

function submitLink() {
  const url = linkUrl.value.trim()
  if (!url) { closeLinkModal(); return }
  showLinkModal.value = false
  // Restore selection before executing command
  if (savedSelection) {
    const sel = window.getSelection()
    sel.removeAllRanges()
    sel.addRange(savedSelection)
  }
  editorRef.value?.focus()
  document.execCommand('createLink', false, url)
  syncEditorHTML()
  savedSelection = null
  linkUrl.value = ''
}

function confirmApplyTemplate(id) {
  const currentContent = editorRef.value?.innerHTML?.trim()
  const isEmpty = !currentContent || currentContent === '<p></p>' || currentContent === ''
  if (!isEmpty && id !== selectedTemplate.value) {
    pendingTemplate.value = id
    showTemplateConfirm.value = true
  } else {
    applyTemplate(id)
  }
}

function applyPendingTemplate() {
  showTemplateConfirm.value = false
  if (pendingTemplate.value) {
    applyTemplate(pendingTemplate.value)
    pendingTemplate.value = null
  }
}

function syncEditorHTML() {
  if (editorRef.value) {
    editorHTML.value = editorRef.value.innerHTML
  }
}

function applyTemplate(id) {
  selectedTemplate.value = id
  if (editorRef.value) {
    editorRef.value.innerHTML = TEMPLATE_HTML[id] || ''
    editorHTML.value = TEMPLATE_HTML[id] || ''
  }
}

async function loadSettings() {
  try {
    const res = await api.get('/api/settings')
    const s = res.data
    if (!form.value.from_name) form.value.from_name = s.from_name || ''
    if (!form.value.from_email) form.value.from_email = s.from_email || ''
    if (!form.value.reply_to) form.value.reply_to = s.reply_to || ''
  } catch (e) {
    // non-critical
  }
}

async function loadLists() {
  try {
    const res = await api.get('/api/lists')
    lists.value = res.data.lists ?? []
  } catch (e) {}
}

async function loadTags() {
  try {
    const res = await api.get('/api/tags')
    tagsList.value = res.data.tags ?? []
  } catch (e) {}
}

async function loadCampaign() {
  if (!campaignId.value) return
  try {
    const res = await api.get(`/api/campaigns/${campaignId.value}`)
    const c = res.data
    form.value = {
      name: c.name || '',
      subject: c.subject || '',
      preview_text: c.preview_text || '',
      from_name: c.from_name || '',
      from_email: c.from_email || '',
      reply_to: c.reply_to || '',
      send_to_type: c.send_to_type || 'list',
      send_to_id: c.send_to_id?.String || c.send_to_id || '',
    }
    await nextTick()
    if (editorRef.value) {
      editorRef.value.innerHTML = c.html_body || ''
      editorHTML.value = c.html_body || ''
    }
  } catch (e) {
    console.error('load campaign error', e)
  }
}

async function saveDraft() {
  saveError.value = ''
  saveSuccess.value = false
  if (editorRef.value) syncEditorHTML()

  const payload = {
    name: form.value.name.trim(),
    subject: form.value.subject.trim(),
    preview_text: form.value.preview_text.trim(),
    from_name: form.value.from_name.trim(),
    from_email: form.value.from_email.trim(),
    reply_to: form.value.reply_to.trim(),
    html_body: editorRef.value?.innerHTML || '',
    send_to_type: form.value.send_to_type,
    send_to_id: form.value.send_to_id || '',
  }

  saveLoading.value = true
  try {
    if (campaignId.value) {
      await api.put(`/api/campaigns/${campaignId.value}`, payload)
    } else {
      const res = await api.post('/api/campaigns', payload)
      campaignId.value = res.data.id?.String || res.data.id
      router.replace(`/campaigns/${campaignId.value}/edit`)
    }
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 2500)
  } catch (e) {
    const msg = e.response?.data?.message || 'Failed to save draft.'
    saveError.value = msg
  } finally {
    saveLoading.value = false
  }
}

function openSchedule() {
  scheduleError.value = ''
  scheduleAt.value = minDatetime.value
  showSchedule.value = true
}

async function submitSchedule() {
  scheduleError.value = ''
  if (!scheduleAt.value) {
    scheduleError.value = 'Please select a date and time.'
    return
  }

  scheduleLoading.value = true
  try {
    if (!campaignId.value) {
      await saveDraft()
      if (!campaignId.value) {
        scheduleError.value = 'Please save the campaign first.'
        return
      }
    }
    const iso = new Date(scheduleAt.value).toISOString()
    await api.post(`/api/campaigns/${campaignId.value}/schedule`, { scheduled_at: iso })
    showSchedule.value = false
    router.push('/campaigns')
  } catch (e) {
    const msg = e.response?.data?.message || 'Failed to schedule campaign.'
    scheduleError.value = msg
  } finally {
    scheduleLoading.value = false
  }
}

function openSendConfirm() {
  sendError.value = ''
  showSendConfirm.value = true
}

async function submitSend() {
  sendError.value = ''
  sendLoading.value = true
  try {
    if (!campaignId.value) {
      await saveDraft()
      if (!campaignId.value) {
        sendError.value = 'Please save the campaign first.'
        return
      }
    } else {
      await saveDraft()
    }
    await api.post(`/api/campaigns/${campaignId.value}/send`)
    showSendConfirm.value = false
    router.push('/campaigns')
  } catch (e) {
    const code = e.response?.data?.error
    const msg = e.response?.data?.message
    if (code === 'no_recipients') {
      sendError.value = 'No active subscribers found for this target.'
    } else {
      sendError.value = msg || 'Failed to send campaign.'
    }
  } finally {
    sendLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadLists(), loadTags()])

  if (campaignId.value) {
    await loadCampaign()
  } else {
    await loadSettings()
    await nextTick()
    if (editorRef.value) {
      editorRef.value.innerHTML = TEMPLATE_HTML.blank
      editorHTML.value = TEMPLATE_HTML.blank
    }
  }

  if (editorRef.value) {
    editorRef.value.addEventListener('input', syncEditorHTML)
  }
})
</script>

<style scoped>
.campaign-edit {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100vh - var(--topbar-height));
  box-sizing: border-box;
}

.template-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.template-label {
  font-size: 12px;
  color: #888;
  font-weight: 500;
}

.template-pill {
  padding: 5px 12px;
  border-radius: 7px;
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
  border: 1px solid #ddd;
  background: #fff;
  color: #666;
  font-weight: 400;
}

.template-pill.active {
  border-color: var(--accent);
  background: var(--accent-light);
  color: var(--accent);
  font-weight: 600;
}

.editor-layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 20px;
  flex: 1;
  min-height: 0;
}

/* Settings panel */
.settings-panel {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  overflow-y: auto;
}

.settings-inner {
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: #444;
  letter-spacing: 0.02em;
}

.required {
  color: #dc2626;
}

.field-hint {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.divider {
  height: 1px;
  background: #f0ede6;
}

.radio-row {
  display: flex;
  gap: 16px;
  margin-bottom: 6px;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #444;
  cursor: pointer;
}

.radio-label input[type="radio"] {
  accent-color: var(--accent);
}

.select-input {
  width: 100%;
  padding: 8px 11px;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: #fff;
  outline: none;
  cursor: pointer;
  transition: border-color 120ms, box-shadow 120ms;
}

.select-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.recipient-info {
  background: #faf8f4;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 12px;
  color: #666;
  line-height: 1.5;
}

.recipient-count {
  color: #333;
  font-size: 13px;
}

/* Editor panel */
.editor-panel {
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  overflow: hidden;
  min-height: 0;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 8px 12px;
  border-bottom: 1px solid #f0ede6;
  background: #faf9f7;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.tb-btn {
  width: 28px;
  height: 28px;
  border: 1px solid #e8e5de;
  border-radius: 5px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  color: #555;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 120ms;
}

.tb-btn:hover {
  background: #f0ede6;
}

.tb-pill {
  padding: 4px 8px;
  border: 1px solid #e8e5de;
  border-radius: 5px;
  background: #fff;
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  color: #555;
  font-family: inherit;
  transition: background 120ms;
}

.tb-pill:hover {
  background: #f0ede6;
}

.toolbar-sep {
  width: 1px;
  height: 20px;
  background: #e8e5de;
  margin: 0 4px;
}

.toolbar-spacer {
  flex: 1;
}

.preview-toggle {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #e8e5de;
  background: #fff;
  color: #555;
  font-size: 11px;
  cursor: pointer;
  font-family: inherit;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: background 120ms;
}

.preview-toggle:hover {
  background: #f0ede6;
}

.editor-content {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.editor-content.preview-bg {
  background: #f4f3ef;
  padding: 24px;
}

.editor-area {
  min-height: 400px;
  outline: none;
  padding: 24px;
  font-size: 14px;
  line-height: 1.7;
  color: #222;
  font-family: 'DM Sans', sans-serif;
}

.editor-area:focus {
  outline: none;
}

/* Placeholder shown when editor is empty */
.editor-area:empty:before {
  content: attr(data-placeholder);
  color: #bbb;
  pointer-events: none;
  display: block;
}

.preview-wrap {
  padding: 24px;
}

.preview-card {
  max-width: 600px;
  margin: 0 auto;
  background: #fff;
  border-radius: 8px;
  padding: 32px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}

.preview-from {
  font-size: 12px;
  color: #aaa;
  margin-bottom: 16px;
}

.preview-subject {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 6px;
  font-family: Georgia, serif;
  color: #111;
}

.preview-preheader {
  color: #aaa;
  font-size: 12px;
  margin-bottom: 24px;
}

.preview-body {
  font-family: Georgia, serif;
  line-height: 1.7;
  color: #333;
  font-size: 14px;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #f0ede6;
  background: #faf9f7;
  flex-shrink: 0;
}

.save-error {
  font-size: 12px;
  color: #dc2626;
  margin: 0;
  flex: 1;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.saved-indicator {
  color: #059669;
  font-weight: 600;
  font-size: 12px;
}

/* Modals */
.form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.datetime-input {
  width: 100%;
  padding: 8px 11px;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: #fff;
  outline: none;
  box-sizing: border-box;
  transition: border-color 120ms, box-shadow 120ms;
}

.datetime-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.send-confirm {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.confirm-text {
  font-size: 14px;
  color: #333;
  margin: 0;
  line-height: 1.5;
}

.confirm-sub {
  font-size: 12px;
  color: #888;
  margin: 0;
}

.form-error {
  font-size: 12px;
  color: #dc2626;
  margin: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 4px;
}
</style>
