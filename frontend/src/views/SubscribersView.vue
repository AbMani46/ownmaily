<template>
  <div class="subscribers">
    <!-- Action bar -->
    <div class="action-bar">
      <div class="search-wrap">
        <Search class="search-icon" :size="14" :stroke-width="2" />
        <input
          v-model="search"
          class="search-input"
          placeholder="Search by email or name…"
          @input="onSearchInput"
        />
      </div>
      <BaseButton variant="secondary" @click="showImport = true">
        <Upload :size="13" :stroke-width="2" />
        Import CSV
      </BaseButton>
      <BaseButton variant="secondary" @click="handleExport">
        <Download :size="13" :stroke-width="2" />
        Export
      </BaseButton>
      <BaseButton variant="primary" @click="showAdd = true">
        <Plus :size="13" :stroke-width="2" />
        Add Subscriber
      </BaseButton>
    </div>

    <!-- Selection bar -->
    <div v-if="selectedIds.size > 0" class="selection-bar">
      <span class="sel-count">{{ selectedIds.size }} selected</span>
      <div class="sel-actions">
        <div class="sel-list-wrap">
          <select v-model="bulkListId" class="sel-list-select">
            <option value="">Add to list…</option>
            <option v-for="l in lists" :key="l.id" :value="l.id">{{ l.name }}</option>
          </select>
          <BaseButton
            variant="primary"
            size="sm"
            :disabled="!bulkListId"
            :loading="bulkLoading"
            @click="submitBulkAdd"
          >
            Add to List
          </BaseButton>
        </div>
        <span class="sel-divider">·</span>
        <button class="sel-link" @click="selectAll">
          Select all ({{ subscribers.length }})
        </button>
        <span class="sel-divider">·</span>
        <button class="sel-link" @click="clearSelection">Deselect all</button>
      </div>
      <span v-if="bulkSuccess" class="sel-success">{{ bulkSuccess }}</span>
    </div>

    <!-- Tabs + Table -->
    <BaseCard :noPadding="true">
      <div class="tabs-wrap">
        <button
          v-for="t in tabs"
          :key="t.id"
          class="tab"
          :class="{ active: activeTab === t.id }"
          @click="setTab(t.id)"
        >
          {{ t.label }}
          <span v-if="t.count != null" class="tab-count">{{ t.count }}</span>
        </button>
      </div>

      <!-- Skeleton rows while loading -->
      <div v-if="loading" class="sk-table">
        <div v-for="n in 8" :key="n" class="sk-row">
          <div class="sk sk-check" />
          <div class="sk sk-text-wide" />
          <div class="sk sk-badge" />
          <div class="sk sk-tags" />
          <div class="sk sk-text" />
        </div>
      </div>

      <BaseTable
        v-else
        :columns="cols"
        :rows="subscribers"
        @row-click="row => $router.push('/subscribers/' + row.id)"
      >
        <template #cell-_select="{ row }">
          <input
            type="checkbox"
            class="row-checkbox"
            :checked="selectedIds.has(row.id)"
            @click.stop
            @change="toggleSelect(row.id)"
          />
        </template>
        <template #cell-email="{ value, row }">
          <div class="email-cell">{{ value }}</div>
          <div class="name-cell">{{ row.first_name }} {{ row.last_name }}</div>
        </template>
        <template #cell-status="{ value }">
          <BaseBadge :status="value" />
        </template>
        <template #cell-tags="{ value }">
          <div class="tags-cell">
            <span
              v-for="(tag, i) in (value || []).slice(0, 3)"
              :key="tag.id || i"
              class="tag-pill"
              :style="{ background: tagColor(tag.name) + '18', color: tagColor(tag.name), borderColor: tagColor(tag.name) + '30' }"
            >{{ tag.name }}</span>
            <span v-if="(value || []).length > 3" class="tag-more">+{{ value.length - 3 }} more</span>
          </div>
        </template>
        <template #cell-created_at="{ value }">
          <span class="muted">{{ formatDate(value) }}</span>
        </template>
        <template #empty>
          <div>No subscribers found.</div>
        </template>
      </BaseTable>

      <div v-if="total > 0" class="pagination-wrap">
        <BasePagination :page="page" :perPage="PER_PAGE" :total="total" @update:page="setPage" />
      </div>
    </BaseCard>

    <!-- Add Subscriber Modal -->
    <BaseModal :show="showAdd" title="Add Subscriber" @close="closeAdd">
      <form @submit.prevent="submitAdd" class="form">
        <div class="form-field">
          <label class="field-label">Email <span class="required">*</span></label>
          <BaseInput v-model="addForm.email" placeholder="user@example.com" type="email" />
        </div>
        <div class="form-row">
          <div class="form-field">
            <label class="field-label">First name</label>
            <BaseInput v-model="addForm.firstName" placeholder="Alice" />
          </div>
          <div class="form-field">
            <label class="field-label">Last name</label>
            <BaseInput v-model="addForm.lastName" placeholder="Smith" />
          </div>
        </div>
        <p v-if="addError" class="form-error">{{ addError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeAdd">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="addLoading">Add Subscriber</BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- Import CSV Modal -->
    <BaseModal :show="showImport" title="Import Subscribers" @close="closeImport">
      <div class="form">
        <div class="form-field">
          <label class="field-label">Add to list <span class="field-optional">(optional)</span></label>
          <select v-model="importListId" class="field-select">
            <option value="">No specific list</option>
            <option v-for="l in lists" :key="l.id" :value="l.id">{{ l.name }}</option>
          </select>
        </div>

        <div class="form-field">
          <div
            class="dropzone"
            :class="{ 'dropzone--active': isDragging, 'dropzone--selected': importFile }"
            @click="fileInput.click()"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @drop.prevent="onDrop"
          >
            <input
              ref="fileInput"
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Upload, Download, Plus } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseInput from '@/components/BaseInput.vue'
import BasePagination from '@/components/BasePagination.vue'
import BaseModal from '@/components/BaseModal.vue'

const PER_PAGE = 50

const subscribers = ref([])
const total = ref(0)
const page = ref(1)
const activeTab = ref('all')
const search = ref('')
const loading = ref(false)
const overview = ref(null)
const lists = ref([])

// Selection state
const selectedIds = ref(new Set())
const bulkListId = ref('')
const bulkLoading = ref(false)
const bulkSuccess = ref('')

const showAdd = ref(false)
const addLoading = ref(false)
const addError = ref('')
const addForm = ref({ email: '', firstName: '', lastName: '' })

const showImport = ref(false)
const importLoading = ref(false)
const importFile = ref(null)
const importPreview = ref(null)
const importResult = ref(null)
const importError = ref('')
const importListId = ref('')
const fileInput = ref(null)
const isDragging = ref(false)

let searchTimer = null

const tabs = computed(() => [
  { id: 'all', label: 'All', count: overview.value?.total_subscribers ?? null },
  { id: 'active', label: 'Active', count: overview.value?.active_subscribers ?? null },
  { id: 'unsubscribed', label: 'Unsubscribed', count: overview.value?.unsubscribed ?? null },
  { id: 'bounced', label: 'Bounced', count: overview.value?.bounced ?? null },
])

const cols = [
  { key: '_select', label: '' },
  { key: 'email', label: 'Email / Name' },
  { key: 'status', label: 'Status' },
  { key: 'tags', label: 'Tags' },
  { key: 'created_at', label: 'Added', muted: true },
]

const TAG_COLORS = ['#10b981', '#8b5cf6', '#0ea5e9', '#f59e0b', '#6b7280', '#ec4899']

function tagColor(name) {
  let h = 0
  for (const c of name) h = (h * 31 + c.charCodeAt(0)) % TAG_COLORS.length
  return TAG_COLORS[Math.abs(h)]
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function toggleSelect(id) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  selectedIds.value = s
}

function selectAll() {
  selectedIds.value = new Set(subscribers.value.map(s => s.id))
}

function clearSelection() {
  selectedIds.value = new Set()
  bulkListId.value = ''
  bulkSuccess.value = ''
}

async function submitBulkAdd() {
  if (!bulkListId.value || selectedIds.value.size === 0) return
  bulkLoading.value = true
  bulkSuccess.value = ''
  try {
    const res = await api.post(`/api/lists/${bulkListId.value}/subscribers/bulk`, {
      subscriber_ids: [...selectedIds.value],
    })
    const added = res.data.added ?? selectedIds.value.size
    bulkSuccess.value = `Added ${added} to list.`
    clearSelection()
  } catch (e) {
    console.error('bulk add error', e)
  } finally {
    bulkLoading.value = false
  }
}

async function fetchSubscribers() {
  loading.value = true
  try {
    const params = { page: page.value, per_page: PER_PAGE }
    // Search and status filter are applied together — search respects the active tab
    if (search.value.trim()) params.q = search.value.trim()
    if (activeTab.value !== 'all') params.status = activeTab.value
    const res = await api.get('/api/subscribers', { params })
    subscribers.value = res.data.subscribers ?? []
    total.value = res.data.total ?? 0
    // Clear selection when the page changes so stale IDs don't persist
    clearSelection()
  } catch (e) {
    console.error('fetch subscribers error', e)
  } finally {
    loading.value = false
  }
}

async function fetchOverview() {
  try {
    const res = await api.get('/api/analytics/overview')
    overview.value = res.data
  } catch (e) {
    // non-critical
  }
}

async function fetchLists() {
  try {
    const res = await api.get('/api/lists')
    lists.value = res.data.lists ?? []
  } catch (e) {
    // non-critical
  }
}

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchSubscribers()
  }, 300)
}

function setTab(id) {
  activeTab.value = id
  search.value = ''
  page.value = 1
  fetchSubscribers()
}

function setPage(p) {
  page.value = p
  fetchSubscribers()
}

async function handleExport() {
  try {
    const res = await api.get('/api/subscribers/export', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = 'subscribers.csv'
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    console.error('export error', e)
  }
}

function closeAdd() {
  showAdd.value = false
  addForm.value = { email: '', firstName: '', lastName: '' }
  addError.value = ''
}

async function submitAdd() {
  addError.value = ''
  if (!addForm.value.email) return
  addLoading.value = true
  try {
    await api.post('/api/subscribers', {
      email: addForm.value.email,
      first_name: addForm.value.firstName,
      last_name: addForm.value.lastName,
    })
    closeAdd()
    page.value = 1
    fetchSubscribers()
    fetchOverview()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'duplicate') {
      addError.value = 'This email is already subscribed.'
    } else if (code === 'suppressed') {
      addError.value = 'This email is suppressed and cannot be added.'
    } else {
      addError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    addLoading.value = false
  }
}

function closeImport() {
  showImport.value = false
  importFile.value = null
  importPreview.value = null
  importResult.value = null
  importError.value = ''
  importListId.value = ''
  if (fileInput.value) fileInput.value.value = ''
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

  const preview = await parseCSVPreview(file)
  importPreview.value = preview
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
    if (importListId.value) form.append('list_id', importListId.value)
    const res = await api.post('/api/subscribers/import', form)
    importResult.value = res.data
    fetchSubscribers()
    fetchOverview()
  } catch (e) {
    importError.value = e.response?.data?.message || 'Import failed. Please try again.'
  } finally {
    importLoading.value = false
  }
}

onMounted(() => {
  fetchSubscribers()
  fetchOverview()
  fetchLists()
})
</script>

<style scoped>
.subscribers {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.action-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-wrap {
  flex: 1;
  position: relative;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: #bbb;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 32px;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  background: #fff;
  color: var(--text-primary);
  outline: none;
  transition: border-color 120ms, box-shadow 120ms;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

/* Selection bar */
.selection-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  background: #f0fdf9;
  border: 1px solid #d1fae5;
  border-radius: var(--radius-card);
  font-size: 13px;
}

.sel-count {
  font-weight: 600;
  color: #065f46;
  white-space: nowrap;
}

.sel-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.sel-list-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.sel-list-select {
  padding: 5px 10px;
  border: 1px solid #a7f3d0;
  border-radius: var(--radius-input);
  font-size: 12px;
  font-family: inherit;
  background: #fff;
  color: var(--text-primary);
  outline: none;
  cursor: pointer;
}

.sel-list-select:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.sel-divider {
  color: #a7f3d0;
  font-size: 12px;
}

.sel-link {
  background: none;
  border: none;
  padding: 0;
  font-size: 12px;
  font-family: inherit;
  color: #059669;
  cursor: pointer;
  white-space: nowrap;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.sel-link:hover {
  color: #047857;
}

.sel-success {
  font-size: 12px;
  color: #059669;
  font-weight: 500;
  margin-left: auto;
}

/* Tabs */
.tabs-wrap {
  display: flex;
  gap: 2px;
  border-bottom: 1px solid var(--border);
  padding: 0 14px;
}

.tab {
  padding: 10px 14px 9px;
  font-size: 13px;
  font-weight: 400;
  color: #666;
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: all 120ms;
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab.active {
  font-weight: 600;
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.tab-count {
  font-size: 11px;
  background: #f0ede6;
  padding: 1px 6px;
  border-radius: 10px;
  color: #888;
}

.tab.active .tab-count {
  background: #d1fae5;
  color: #065f46;
}

.pagination-wrap {
  padding: 0 14px 14px;
}

/* Checkbox column */
.row-checkbox {
  width: 15px;
  height: 15px;
  accent-color: var(--accent);
  cursor: pointer;
}

/* Table cell content */
.email-cell {
  font-weight: 500;
  color: #111;
  margin-bottom: 1px;
}

.name-cell {
  font-size: 11px;
  color: #aaa;
}

.tags-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  align-items: center;
}

.tag-pill {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid;
}

.tag-more {
  font-size: 11px;
  color: var(--text-muted);
}

.muted {
  color: var(--text-muted);
}

/* Loading skeleton */
.sk-table {
  padding: 4px 14px 14px;
  display: flex;
  flex-direction: column;
}

.sk-row {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 14px 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.sk {
  background: #e8e5de;
  border-radius: 4px;
  animation: pulse 1.5s ease-in-out infinite;
}

.sk-check { height: 15px; width: 15px; border-radius: 3px; flex-shrink: 0; }
.sk-text { height: 12px; width: 70px; }
.sk-text-wide { height: 12px; width: 180px; }
.sk-badge { height: 20px; width: 60px; border-radius: 20px; }
.sk-tags { height: 20px; width: 120px; border-radius: 20px; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

/* Form */
.form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
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

.field-optional {
  font-weight: 400;
  color: var(--text-muted);
}

.field-select {
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  background: #fff;
  color: var(--text-primary);
  outline: none;
  cursor: pointer;
  transition: border-color 120ms, box-shadow 120ms;
}

.field-select:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.required {
  color: #dc2626;
}

.import-ai-hint {
  font-size: 11.5px;
  color: var(--text-muted);
  margin: 4px 0 0;
  font-style: italic;
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
