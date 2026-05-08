<template>
  <div class="list-detail">
    <!-- Back -->
    <div>
      <BaseButton variant="ghost" @click="$router.push('/lists')">
        <ChevronLeft :size="13" :stroke-width="2" />
        Back to Lists
      </BaseButton>
    </div>

    <div v-if="loading" class="loading-state">Loading…</div>

    <template v-if="list">
      <!-- Header with delete -->
      <div class="detail-header">
        <div>
          <h1 class="detail-title">{{ list.name }}</h1>
          <p v-if="list.description" class="detail-desc">{{ list.description }}</p>
        </div>
        <BaseButton variant="danger" @click="handleDelete" :loading="deleteLoading">
          <Trash2 :size="13" :stroke-width="2" />
          Delete List
        </BaseButton>
      </div>

      <!-- Stat row -->
      <div class="stat-row">
        <div class="stat-card">
          <div class="stat-label">Total Subscribers</div>
          <div class="stat-value">{{ (list.subscriber_count || 0).toLocaleString() }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Opt-in Type</div>
          <div class="stat-value-badge">
            <BaseBadge :status="list.double_opt_in ? 'double_optin' : 'single_optin'" />
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Created</div>
          <div class="stat-value stat-value--sm">{{ formatDate(list.created_at) }}</div>
        </div>
      </div>

      <!-- Subscribers table -->
      <BaseCard :noPadding="true">
        <div class="table-header-row">
          <span class="section-title">{{ list.name }} — Subscribers</span>
          <div class="table-header-actions">
            <BaseButton variant="secondary" @click="showImport = true">
              <Upload :size="12" :stroke-width="2" />
              Import CSV
            </BaseButton>
            <BaseButton variant="primary" @click="showAddSub = true">
              <Plus :size="12" :stroke-width="2" />
              Add Subscriber
            </BaseButton>
          </div>
        </div>

        <BaseTable
          :columns="subCols"
          :rows="subscribers"
          @row-click="row => $router.push('/subscribers/' + row.id)"
        >
          <template #cell-email="{ value, row }">
            <div class="email-cell">{{ value }}</div>
            <div class="name-cell">{{ row.first_name }} {{ row.last_name }}</div>
          </template>
          <template #cell-status="{ value }">
            <BaseBadge :status="value" />
          </template>
          <template #cell-created_at="{ value }">
            <span class="muted">{{ formatDate(value) }}</span>
          </template>
          <template #cell-_actions="{ row }">
            <BaseButton
              variant="ghost"
              @click.stop="handleRemoveSub(row)"
              class="remove-btn"
            >
              <X :size="12" color="#dc2626" />
              Remove
            </BaseButton>
          </template>
          <template #empty>
            <div>No subscribers in this list yet.</div>
          </template>
        </BaseTable>

        <div v-if="subTotal > PER_PAGE" class="pagination-wrap">
          <BasePagination :page="subPage" :perPage="PER_PAGE" :total="subTotal" @update:page="setSubPage" />
        </div>
      </BaseCard>

      <!-- Embed code -->
      <BaseCard>
        <div class="embed-header">
          <span class="section-title">Embed Subscribe Form</span>
          <BaseButton variant="ghost" @click="copyEmbed">
            <template v-if="copied">
              <CheckCheck :size="12" color="#10b981" />
              Copied!
            </template>
            <template v-else>
              <Copy :size="12" />
              Copy
            </template>
          </BaseButton>
        </div>
        <div class="code-block">
          <code>{{ embedCode }}</code>
        </div>
        <p class="embed-hint">Paste this snippet into any HTML page to embed a subscribe form.</p>
      </BaseCard>
    </template>

    <!-- Import CSV Modal -->
    <BaseModal :show="showImport" title="Import Subscribers" @close="closeImport">
      <div class="form">
        <div class="form-field">
          <div
            class="dropzone"
            :class="{ 'dropzone--active': isDragging, 'dropzone--selected': importFile }"
            @click="importFileInput.click()"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop.prevent="onDrop"
          >
            <input
              ref="importFileInput"
              type="file"
              accept=".csv"
              class="file-input-hidden"
              @change="onFileSelect"
            />
            <Upload :size="22" :stroke-width="1.5" class="dropzone-icon" />
            <span v-if="importFile" class="dropzone-filename">{{ importFile.name }}</span>
            <span v-else class="dropzone-label">Click to browse or drag a CSV here</span>
            <span class="dropzone-hint">Required: <code>email</code> &nbsp;·&nbsp; Optional: <code>first_name</code>, <code>last_name</code></span>
          </div>
          <p class="import-ai-hint">Already have subscribers? Export your list, filter with AI, and import back.</p>
        </div>

        <div v-if="importPreview" class="import-preview">
          <span v-if="importPreview.hasEmail">
            <strong>{{ importPreview.valid }}</strong> valid rows detected
          </span>
          <span v-else class="form-error">No <code>email</code> column found in CSV</span>
        </div>

        <div v-if="importResult" class="import-result">
          Imported <strong>{{ importResult.imported }}</strong>,
          Skipped <strong>{{ importResult.skipped }}</strong>,
          Invalid <strong>{{ importResult.invalid }}</strong>
        </div>

        <p v-if="importError" class="form-error">{{ importError }}</p>

        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeImport">Close</BaseButton>
          <BaseButton
            variant="primary"
            :loading="importLoading"
            :disabled="!importFile || (importPreview && !importPreview.hasEmail)"
            @click="submitImport"
          >
            Import
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- Add Subscriber Modal -->
    <BaseModal :show="showAddSub" title="Add Subscriber to List" @close="closeAddSub">
      <form @submit.prevent="submitAddSub" class="form">
        <div class="form-field">
          <label class="field-label">Email address <span class="required">*</span></label>
          <BaseInput
            v-model="addSubEmail"
            placeholder="subscriber@example.com"
            type="email"
            @input="onAddSubInput"
          />
          <div v-if="addSubSearchResult" class="search-result">
            <span v-if="addSubSearchResult.found" class="search-found">
              Found: <strong>{{ addSubSearchResult.name }}</strong>
            </span>
            <span v-else class="search-not-found">
              No subscriber found with that email.
              <a href="/subscribers" class="create-link">Add them first →</a>
            </span>
          </div>
        </div>
        <p v-if="addSubError" class="form-error">{{ addSubError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeAddSub">Cancel</BaseButton>
          <BaseButton
            variant="primary"
            type="submit"
            :loading="addSubLoading"
            :disabled="!addSubSearchResult?.found"
          >
            Add to List
          </BaseButton>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronLeft, Plus, Trash2, X, Copy, CheckCheck, Upload } from 'lucide-vue-next'
import api from '@/lib/api'
import { useConfirm } from '@/composables/useConfirm'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseInput from '@/components/BaseInput.vue'
import BaseModal from '@/components/BaseModal.vue'
import BasePagination from '@/components/BasePagination.vue'

const route = useRoute()
const router = useRouter()
const { confirm } = useConfirm()

const PER_PAGE = 50

const list = ref(null)
const loading = ref(false)
const deleteLoading = ref(false)

const subscribers = ref([])
const subTotal = ref(0)
const subPage = ref(1)

const showAddSub = ref(false)
const addSubLoading = ref(false)
const addSubEmail = ref('')
const addSubError = ref('')
const addSubSearchResult = ref(null)
const addSubFoundId = ref(null)
let addSubTimer = null

const copied = ref(false)

const showImport = ref(false)
const importLoading = ref(false)
const importFile = ref(null)
const importPreview = ref(null)
const importResult = ref(null)
const importError = ref('')
const importFileInput = ref(null)
const isDragging = ref(false)

const embedCode = computed(() => {
  if (!list.value) return ''
  return `<script src="${window.location.origin}/embed/${list.value.id}.js"><\/script>`
})

const subCols = [
  { key: 'email', label: 'Email / Name' },
  { key: 'status', label: 'Status' },
  { key: 'created_at', label: 'Added' },
  { key: '_actions', label: '' },
]

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

async function fetchList() {
  loading.value = true
  try {
    const res = await api.get(`/api/lists/${route.params.id}`)
    list.value = res.data
  } catch (e) {
    if (e.response?.status === 404) router.push('/lists')
  } finally {
    loading.value = false
  }
}

async function fetchSubscribers() {
  try {
    const res = await api.get(`/api/lists/${route.params.id}/subscribers`, {
      params: { page: subPage.value, per_page: PER_PAGE },
    })
    subscribers.value = res.data.subscribers ?? []
    subTotal.value = res.data.total ?? 0
  } catch (e) {
    console.error('fetch list subscribers error', e)
  }
}

function setSubPage(p) {
  subPage.value = p
  fetchSubscribers()
}

async function handleDelete() {
  if (!list.value) return
  const count = list.value.subscriber_count || 0
  let confirmed
  if (count > 0) {
    confirmed = confirm(`This list has ${count} subscriber${count !== 1 ? 's' : ''}. Force delete anyway?`)
  } else {
    confirmed = confirm(`Delete list "${list.value.name}"? This cannot be undone.`)
  }
  if (!confirmed) return

  deleteLoading.value = true
  try {
    await api.delete(`/api/lists/${route.params.id}${count > 0 ? '?force=true' : ''}`)
    router.push('/lists')
  } catch (e) {
    console.error('delete list error', e)
  } finally {
    deleteLoading.value = false
  }
}

async function handleRemoveSub(row) {
  const yes = confirm(`Remove ${row.email} from this list?`)
  if (!yes) return
  try {
    await api.delete(`/api/lists/${route.params.id}/subscribers/${row.id}`)
    fetchList()
    fetchSubscribers()
  } catch (e) {
    console.error('remove subscriber error', e)
  }
}

function onAddSubInput() {
  addSubSearchResult.value = null
  addSubFoundId.value = null
  clearTimeout(addSubTimer)
  if (!addSubEmail.value.trim()) return
  addSubTimer = setTimeout(searchSubscriber, 400)
}

async function searchSubscriber() {
  try {
    const res = await api.get('/api/subscribers', {
      params: { q: addSubEmail.value.trim(), per_page: 1 },
    })
    const subs = res.data.subscribers ?? []
    const found = subs.find(s => s.email.toLowerCase() === addSubEmail.value.trim().toLowerCase())
    if (found) {
      addSubSearchResult.value = { found: true, name: `${found.first_name} ${found.last_name}`.trim() || found.email }
      addSubFoundId.value = found.id
    } else {
      addSubSearchResult.value = { found: false }
      addSubFoundId.value = null
    }
  } catch (e) {
    addSubSearchResult.value = null
  }
}

function closeAddSub() {
  showAddSub.value = false
  addSubEmail.value = ''
  addSubError.value = ''
  addSubSearchResult.value = null
  addSubFoundId.value = null
}

async function submitAddSub() {
  if (!addSubFoundId.value) return
  addSubError.value = ''
  addSubLoading.value = true
  try {
    await api.post(`/api/lists/${route.params.id}/subscribers`, {
      subscriber_id: addSubFoundId.value,
    })
    closeAddSub()
    fetchList()
    fetchSubscribers()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'already_member') {
      addSubError.value = 'This subscriber is already in the list.'
    } else if (code === 'inactive_subscriber') {
      addSubError.value = 'Cannot add an inactive subscriber.'
    } else {
      addSubError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    addSubLoading.value = false
  }
}

function closeImport() {
  showImport.value = false
  importFile.value = null
  importPreview.value = null
  importResult.value = null
  importError.value = ''
  if (importFileInput.value) importFileInput.value.value = ''
}

async function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files?.[0]
  if (!file) return
  await processImportFile(file)
}

async function onFileSelect(e) {
  const file = e.target.files?.[0]
  if (!file) return
  await processImportFile(file)
}

async function processImportFile(file) {
  importFile.value = file
  importPreview.value = null
  importResult.value = null
  importError.value = ''
  importPreview.value = await parseCSVPreview(file)
}

function parseCSVPreview(file) {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target.result
      const lines = text.trim().split(/\r?\n/)
      if (lines.length < 2) { resolve({ valid: 0, hasEmail: false }); return }
      const headers = lines[0].split(',').map(h => h.trim().replace(/^["']|["']$/g, '').toLowerCase())
      const emailCol = headers.findIndex(h => h === 'email')
      if (emailCol === -1) { resolve({ valid: 0, hasEmail: false }); return }
      let count = 0
      for (let i = 1; i < lines.length; i++) {
        const cols = lines[i].split(',')
        const email = (cols[emailCol] || '').trim().replace(/^["']|["']$/g, '')
        if (email.includes('@')) count++
      }
      resolve({ valid: count, hasEmail: true })
    }
    reader.readAsText(file)
  })
}

async function submitImport() {
  if (!importFile.value) return
  importError.value = ''
  importResult.value = null
  importLoading.value = true
  try {
    const form = new FormData()
    form.append('file', importFile.value)
    form.append('list_id', route.params.id)
    const res = await api.post('/api/subscribers/import', form)
    importResult.value = res.data
    fetchList()
    fetchSubscribers()
  } catch (e) {
    importError.value = e.response?.data?.message || 'Import failed. Please try again.'
  } finally {
    importLoading.value = false
  }
}

async function copyEmbed() {
  try {
    await navigator.clipboard.writeText(embedCode.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (e) {
    console.error('copy failed', e)
  }
}

onMounted(() => {
  fetchList()
  fetchSubscribers()
})
</script>

<style scoped>
.list-detail {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.loading-state {
  color: var(--text-muted);
  font-size: 13px;
  padding: 20px 0;
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.detail-title {
  font-size: 18px;
  font-weight: 700;
  color: #111;
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}

.detail-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0;
}

.stat-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.stat-card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 16px 20px;
}

.stat-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #111;
  letter-spacing: -0.04em;
}

.stat-value--sm {
  font-size: 16px;
  letter-spacing: -0.02em;
}

.stat-value-badge {
  margin-top: 2px;
}

.table-header-row {
  padding: 16px 22px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.pagination-wrap {
  padding: 0 14px 14px;
}

.email-cell {
  font-weight: 500;
  color: #111;
  margin-bottom: 1px;
}

.name-cell {
  font-size: 11px;
  color: #aaa;
}

.muted {
  color: var(--text-muted);
}

.remove-btn {
  color: #dc2626;
  font-size: 12px;
}

.embed-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.code-block {
  background: #0c0c0f;
  border-radius: 7px;
  padding: 16px 18px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: #9aefbc;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
}

.embed-hint {
  font-size: 12px;
  color: #aaa;
  margin: 10px 0 0;
}

.form {
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

.search-result {
  font-size: 12px;
  margin-top: 4px;
}

.search-found {
  color: #059669;
}

.search-not-found {
  color: #dc2626;
}

.create-link {
  color: var(--accent);
  text-decoration: none;
  margin-left: 4px;
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

.file-input-hidden {
  display: none;
}

.dropzone {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 24px 20px;
  border: 1.5px dashed #d0ccc4;
  border-radius: var(--radius-input);
  background: #faf9f7;
  cursor: pointer;
  transition: border-color 120ms, background 120ms;
  text-align: center;
  user-select: none;
}

.dropzone:hover,
.dropzone--active {
  border-color: var(--accent);
  background: var(--accent-light);
}

.dropzone--selected {
  border-color: #a8d5c3;
  background: var(--accent-light);
}

.dropzone-icon {
  color: #aaa;
  margin-bottom: 2px;
}

.dropzone--active .dropzone-icon,
.dropzone--selected .dropzone-icon {
  color: var(--accent);
}

.dropzone-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.dropzone-filename {
  font-size: 13px;
  font-weight: 500;
  color: var(--accent);
}

.dropzone-hint {
  font-size: 11.5px;
  color: var(--text-muted);
  margin-top: 2px;
}

.dropzone-hint code {
  font-family: monospace;
  background: #ede9e1;
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 11px;
}

.import-ai-hint {
  font-size: 11.5px;
  color: var(--text-muted);
  margin: 4px 0 0;
  font-style: italic;
}

.import-preview {
  font-size: 13px;
  color: #059669;
  background: #f0fdf9;
  border: 1px solid #d1fae5;
  border-radius: 7px;
  padding: 8px 12px;
}

.import-result {
  font-size: 13px;
  color: #065f46;
  background: #f0fdf9;
  border: 1px solid #d1fae5;
  border-radius: 7px;
  padding: 8px 12px;
}
</style>
