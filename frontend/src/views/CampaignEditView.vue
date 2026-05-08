<template>
  <div class="campaign-edit">
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
        <!-- Content area -->
        <div class="editor-content" :class="{ 'preview-bg': previewMode }">
          <!-- Edit mode -->
          <EmailEditor v-show="!previewMode" v-model="editorHTML" :preview-mode="previewMode" @toggle-preview="previewMode = !previewMode" />
          <!-- Preview mode -->
          <div v-show="previewMode" class="preview-wrap">
            <div class="preview-controls">
              <div class="device-toggle">
                <button class="device-btn" :class="{ active: previewDevice === 'desktop' }" @click="previewDevice = 'desktop'">
                  <Monitor :size="13" />
                  Desktop
                </button>
                <button class="device-btn" :class="{ active: previewDevice === 'mobile' }" @click="previewDevice = 'mobile'">
                  <Smartphone :size="13" />
                  Mobile
                </button>
              </div>
              <button class="preview-edit-btn" @click="previewMode = false">
                <Pencil :size="12" />
                Edit
              </button>
            </div>

            <div class="email-viewport" :class="`device-${previewDevice}`">
              <div v-if="previewDevice === 'mobile'" class="phone-chrome">
                <span class="phone-time">9:41</span>
                <span class="phone-icons">··· ▲ ▮</span>
              </div>
              <div class="email-shell">
                <div class="shell-meta">
                  <div class="shell-from">From: {{ form.from_name }} &lt;{{ form.from_email }}&gt;</div>
                  <div class="shell-subject">{{ form.subject || 'No subject yet' }}</div>
                  <div class="shell-preheader">{{ form.preview_text || 'No preview text' }}</div>
                </div>
                <div class="shell-divider" />
                <div class="email-body" v-html="previewEmailHTML" />
              </div>
            </div>
          </div>
        </div>

        <!-- Bottom action bar -->
        <div class="action-bar">
          <div class="action-left">
            <BaseButton variant="secondary" @click="saveDraft" :loading="saveLoading">
              <template v-if="saveSuccess">
                <span class="saved-indicator">✓ Saved</span>
              </template>
              <template v-else>
                Save Draft
              </template>
            </BaseButton>
            <BaseButton variant="secondary" @click="openSchedule" :disabled="!form.send_to_id">
              <CalendarDays :size="13" :stroke-width="2" />
              Schedule
            </BaseButton>
            <p v-if="saveError" class="save-error">{{ saveError }}</p>
          </div>
          <div class="action-right">
            <span v-if="sendPreflight" class="send-hint send-error-hint">{{ sendPreflight }}</span>
            <span v-else-if="!form.send_to_id" class="send-hint">Select a list first</span>
            <BaseButton variant="primary" @click="openSendConfirm" :disabled="!form.send_to_id">
              <SendHorizonal :size="13" :stroke-width="2" />
              Send Now
            </BaseButton>
          </div>
        </div>
      </div>
    </div>

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
        <div class="schedule-summary">
          <p class="schedule-name">{{ form.name || 'Untitled campaign' }}</p>
          <p class="schedule-meta">{{ form.subject || '(no subject)' }} &middot; {{ selectedTargetCount?.toLocaleString() ?? '?' }} subscriber{{ selectedTargetCount !== 1 ? 's' : '' }}</p>
        </div>
        <div class="form-field">
          <label class="field-label">Send at <span class="required">*</span></label>
          <input
            v-model="scheduleAt"
            type="datetime-local"
            class="datetime-input"
            :min="minDatetime"
          />
          <span class="field-hint">Your local time ({{ localTimezoneLabel }})</span>
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
        <div class="confirm-summary">
          <div class="summary-row">
            <span class="summary-label">Subject</span>
            <span class="summary-val">{{ form.subject || '(no subject)' }}</span>
          </div>
          <div class="summary-row">
            <span class="summary-label">From</span>
            <span class="summary-val">{{ form.from_name }} &lt;{{ form.from_email }}&gt;</span>
          </div>
        </div>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CalendarDays, SendHorizonal, Pencil, Monitor, Smartphone } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseInput from '@/components/BaseInput.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseModal from '@/components/BaseModal.vue'
import EmailEditor from '@/components/EmailEditor.vue'

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
const installationURL = ref('')

const editorHTML = ref('')
const previewMode = ref(false)
const previewDevice = ref('desktop')
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
const sendPreflight = ref('')

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
  newsletter: `<h2>Hello {{first_name}},</h2>
<p>Here's what's been happening this month.</p>
<h3>What's new</h3>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
<hr />
<p style="color:#888;font-size:13px">You're receiving this because you subscribed. <a href="{{unsubscribe_url}}">Unsubscribe</a></p>`,
  announcement: `<h2>Exciting news!</h2>
<p>We're thrilled to share something special with you.</p>
<p><a data-cta href="#" style="background:#10b981;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;display:inline-block">Learn More →</a></p>
<hr />
<p style="color:#888;font-size:13px">You're receiving this because you subscribed. <a href="{{unsubscribe_url}}">Unsubscribe</a></p>`,
}

const minDatetime = computed(() => {
  const now = new Date(Date.now() + 5 * 60000)
  const pad = n => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`
})

const localTimezoneLabel = computed(() => {
  const offset = -new Date().getTimezoneOffset()
  const sign = offset >= 0 ? '+' : '-'
  const h = String(Math.floor(Math.abs(offset) / 60)).padStart(2, '0')
  const m = String(Math.abs(offset) % 60).padStart(2, '0')
  return `UTC${sign}${h}:${m}`
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

const previewEmailHTML = computed(() => toEmailHTML(editorHTML.value || '', installationURL.value))

function confirmApplyTemplate(id) {
  const currentContent = editorHTML.value?.trim()
  const isEmpty = !currentContent || currentContent === '<p></p>'
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

function applyTemplate(id) {
  selectedTemplate.value = id
  editorHTML.value = fromEmailHTML(TEMPLATE_HTML[id] || '')
}

async function loadSettings() {
  try {
    const res = await api.get('/api/settings')
    const s = res.data
    if (!form.value.from_name) form.value.from_name = s.from_name || ''
    if (!form.value.from_email) form.value.from_email = s.from_email || ''
    if (!form.value.reply_to) form.value.reply_to = s.reply_to || ''
    if (s.installation_url) installationURL.value = s.installation_url.replace(/\/$/, '')
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
    editorHTML.value = fromEmailHTML(c.html_body || '')
  } catch (e) {
    console.error('load campaign error', e)
  }
}

function toEmailHTML(html, baseURL = '') {
  if (!html) return ''

  const doc = new DOMParser().parseFromString(html, 'text/html')
  const body = doc.body

  // Variable chips → {{var}}
  body.querySelectorAll('span[data-variable]').forEach(el => {
    el.replaceWith(`{{${el.getAttribute('data-variable')}}}`)
  })

  // Headings: inline font/spacing (email clients strip <style> blocks)
  const HEADING_STYLES = {
    H1: 'font-size:28px;font-weight:700;margin:0 0 16px 0;line-height:1.2;',
    H2: 'font-size:22px;font-weight:700;margin:0 0 14px 0;line-height:1.3;',
    H3: 'font-size:18px;font-weight:600;margin:0 0 12px 0;line-height:1.4;',
  }
  body.querySelectorAll('h1, h2, h3').forEach(el => {
    el.setAttribute('style', HEADING_STYLES[el.tagName])
  })

  // Paragraphs: margin only (preserve existing color/size on template footer paragraphs)
  body.querySelectorAll('p').forEach(el => {
    if (!el.closest('li') && !el.style.margin) el.style.margin = '0 0 12px 0'
  })

  // Bullet lists: padding-left for Outlook (ignores CSS-only indent)
  body.querySelectorAll('ul').forEach(el => {
    el.setAttribute('style', 'padding:0 0 0 20px;margin:0 0 14px 0;')
  })

  // Ordered lists: same treatment as bullet lists
  body.querySelectorAll('ol').forEach(el => {
    el.setAttribute('style', 'padding:0 0 0 20px;margin:0 0 14px 0;list-style-type:decimal;')
  })

  // List items: unwrap the <p> Tiptap nests inside each <li> (causes double-spacing in Gmail/Yahoo)
  body.querySelectorAll('li > p').forEach(p => {
    const parent = p.parentNode
    while (p.firstChild) parent.insertBefore(p.firstChild, p)
    parent.removeChild(p)
  })

  // Links: inline color so Yahoo/Android Mail don't override with their defaults
  body.querySelectorAll('a[href]').forEach(el => {
    if (!el.style.color) el.style.color = '#10b981'
    if (!el.style.textDecoration) el.style.textDecoration = 'underline'
  })

  // HR: flat border (Outlook renders plain <hr> as a 3D embossed rule)
  body.querySelectorAll('hr').forEach(el => {
    el.setAttribute('style', 'border:none;border-top:1px solid #e5e7eb;margin:24px 0;')
  })

  // Images: email-safe inline styles. When baseURL is provided (preview pane), convert
  // root-relative src paths to absolute so the browser can fetch them.
  body.querySelectorAll('img').forEach(el => {
    if (baseURL && el.getAttribute('src')?.startsWith('/')) {
      el.setAttribute('src', baseURL + el.getAttribute('src'))
    }
    el.style.maxWidth = '100%'
    el.style.height = 'auto'
    el.style.display = 'block'
    if (!el.style.margin) el.style.margin = '0 auto 12px'
  })

  return body.innerHTML
}

function fromEmailHTML(html) {
  if (!html) return html
  const doc = new DOMParser().parseFromString(html, 'text/html')

  // Mark old-style inline-block links as CTA buttons so Tiptap's parseHTML rule can rehydrate them
  doc.body.querySelectorAll('a[href]').forEach(el => {
    const bg = el.style.background || el.style.backgroundColor
    if (el.style.display === 'inline-block' && bg) el.setAttribute('data-cta', '')
  })

  const pattern = /\{\{([a-z_]+)\}\}/g

  function walk(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      pattern.lastIndex = 0
      if (!pattern.test(node.textContent)) return
      pattern.lastIndex = 0
      const frag = document.createDocumentFragment()
      let last = 0, match
      while ((match = pattern.exec(node.textContent)) !== null) {
        if (match.index > last) frag.appendChild(document.createTextNode(node.textContent.slice(last, match.index)))
        const span = document.createElement('span')
        span.setAttribute('data-variable', match[1])
        span.className = 'variable-chip'
        span.textContent = `{{${match[1]}}}`
        frag.appendChild(span)
        last = match.index + match[0].length
      }
      if (last < node.textContent.length) frag.appendChild(document.createTextNode(node.textContent.slice(last)))
      node.replaceWith(frag)
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      Array.from(node.childNodes).forEach(walk)
    }
  }

  walk(doc.body)
  return doc.body.innerHTML
}

async function doSave() {
  const payload = {
    name: form.value.name.trim(),
    subject: form.value.subject.trim(),
    preview_text: form.value.preview_text.trim(),
    from_name: form.value.from_name.trim(),
    from_email: form.value.from_email.trim(),
    reply_to: form.value.reply_to.trim(),
    html_body: toEmailHTML(editorHTML.value || ''),
    send_to_type: form.value.send_to_type,
    send_to_id: form.value.send_to_id || '',
  }
  if (campaignId.value) {
    await api.put(`/api/campaigns/${campaignId.value}`, payload)
  } else {
    const res = await api.post('/api/campaigns', payload)
    campaignId.value = res.data.id?.String || res.data.id
    router.replace(`/campaigns/${campaignId.value}/edit`)
  }
}

async function saveDraft() {
  saveError.value = ''
  saveSuccess.value = false
  saveLoading.value = true
  try {
    await doSave()
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
  if (!form.value.send_to_id) {
    scheduleError.value = 'Select a recipient list or tag before scheduling.'
    return
  }
  if (!scheduleAt.value) {
    scheduleError.value = 'Please select a date and time.'
    return
  }

  scheduleLoading.value = true
  try {
    await doSave()
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
  sendPreflight.value = ''
  sendError.value = ''
  if (!form.value.name.trim()) {
    sendPreflight.value = 'Campaign name is required.'
    return
  }
  if (!form.value.subject.trim()) {
    sendPreflight.value = 'Subject line is required.'
    return
  }
  showSendConfirm.value = true
}

async function submitSend() {
  sendError.value = ''
  if (!form.value.name.trim()) {
    sendError.value = 'Campaign name is required before sending.'
    return
  }
  if (!form.value.subject.trim()) {
    sendError.value = 'Subject line is required before sending.'
    return
  }
  sendLoading.value = true
  try {
    await doSave()
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
    editorHTML.value = TEMPLATE_HTML.blank
  }
})
</script>

<style scoped>
.campaign-edit {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--topbar-height));
  box-sizing: border-box;
}

.template-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
  background: #faf9f7;
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

.editor-content {
  flex: 1;
  min-height: 0;
}

.editor-content.preview-bg {
  background: #f4f3ef;
  padding: 24px;
}

.preview-wrap {
  padding: 20px 24px 24px;
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.preview-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.device-toggle {
  display: flex;
  gap: 3px;
  background: #e8e5de;
  border-radius: 8px;
  padding: 3px;
}

.device-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 11px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #666;
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
}

.device-btn.active {
  background: #fff;
  color: #333;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.preview-edit-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border: 1px solid #e8e5de;
  border-radius: 6px;
  background: #fff;
  color: #555;
  font-size: 11px;
  font-family: inherit;
  cursor: pointer;
  transition: background 120ms;
}

.preview-edit-btn:hover {
  background: #f0ede6;
}

.email-viewport {
  margin: 0 auto;
  overflow: hidden;
}

.device-desktop {
  max-width: 600px;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}

.device-mobile {
  max-width: 375px;
  border: 8px solid #1e1e1e;
  border-radius: 40px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.18);
}

.phone-chrome {
  background: #1e1e1e;
  color: #fff;
  font-size: 11px;
  padding: 8px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-family: -apple-system, BlinkMacSystemFont, sans-serif;
}

.phone-time {
  font-weight: 600;
  letter-spacing: 0.01em;
}

.phone-icons {
  opacity: 0.85;
  letter-spacing: 2px;
}

.email-shell {
  background: #fff;
}

.shell-meta {
  padding: 20px 24px 16px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.shell-from {
  font-size: 12px;
  color: #aaa;
}

.shell-subject {
  font-size: 18px;
  font-weight: 700;
  font-family: Georgia, serif;
  color: #111;
}

.shell-preheader {
  font-size: 12px;
  color: #aaa;
}

.shell-divider {
  height: 1px;
  background: #eee;
  margin: 0 24px;
}

.email-body {
  padding: 24px;
  font-family: Georgia, serif;
  font-size: 14px;
  line-height: 1.7;
  color: #333;
}

:deep(.email-body .variable-chip) {
  display: inline;
  padding: 0 5px;
  border-radius: 10px;
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #059669;
  font-size: 11px;
  font-family: 'DM Mono', 'Courier New', monospace;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #f0ede6;
  background: #faf9f7;
  flex-shrink: 0;
}

.action-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.action-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.save-error {
  font-size: 12px;
  color: #dc2626;
  margin: 0;
}

.send-hint {
  font-size: 11px;
  color: #aaa;
}

.send-error-hint {
  color: #dc2626;
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

.confirm-summary {
  background: #f9f8f5;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.summary-row {
  display: flex;
  gap: 10px;
  font-size: 12px;
  line-height: 1.4;
}

.summary-label {
  color: #999;
  min-width: 46px;
  flex-shrink: 0;
}

.summary-val {
  color: #333;
  font-weight: 500;
  word-break: break-word;
}

.schedule-summary {
  background: #f9f8f5;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 10px 14px;
}

.schedule-name {
  font-size: 13px;
  font-weight: 600;
  color: #333;
  margin: 0 0 3px 0;
}

.schedule-meta {
  font-size: 12px;
  color: #888;
  margin: 0;
}
</style>
