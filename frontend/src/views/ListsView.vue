<template>
  <div class="lists-view">
    <!-- Header -->
    <div class="action-bar">
      <span class="page-title">Lists</span>
      <BaseButton variant="primary" @click="showCreate = true">
        <Plus :size="13" :stroke-width="2" />
        Create List
      </BaseButton>
    </div>

    <!-- Top 3 cards -->
    <div v-if="lists.length > 0" class="list-cards">
      <div
        v-for="list in lists.slice(0, 3)"
        :key="list.id"
        class="list-card"
        @click="$router.push('/lists/' + list.id)"
      >
        <div class="card-header-row">
          <div class="list-icon">
            <LayoutList :size="16" :stroke-width="2" color="#10b981" />
          </div>
          <BaseBadge :status="list.double_opt_in ? 'double_optin' : 'single_optin'" />
        </div>
        <div class="card-name">{{ list.name }}</div>
        <div class="card-count">{{ (list.subscriber_count || 0).toLocaleString() }}</div>
        <div class="card-meta">subscribers &middot; created {{ formatDate(list.created_at) }}</div>
      </div>
    </div>

    <!-- Full table -->
    <BaseCard :noPadding="true">
      <div class="table-header">
        <span class="section-title">All Lists</span>
      </div>
      <BaseTable
        :columns="cols"
        :rows="lists"
        @row-click="row => $router.push('/lists/' + row.id)"
      >
        <template #cell-name="{ value }">
          <span class="list-name">{{ value }}</span>
        </template>
        <template #cell-subscriber_count="{ value }">
          <span class="count-cell">{{ (value || 0).toLocaleString() }}</span>
        </template>
        <template #cell-double_opt_in="{ value }">
          <BaseBadge :status="value ? 'double_optin' : 'single_optin'" />
        </template>
        <template #cell-created_at="{ value }">
          <span class="muted">{{ formatDate(value) }}</span>
        </template>
        <template #cell-_actions="{ row }">
          <div class="actions-cell" @click.stop>
            <BaseButton variant="ghost" @click="openEdit(row)">Edit</BaseButton>
          </div>
        </template>
        <template #empty>
          <div>No lists yet. Create your first list to get started.</div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Create List Modal -->
    <BaseModal :show="showCreate" title="Create List" @close="closeCreate">
      <form @submit.prevent="submitCreate" class="form">
        <div class="form-field">
          <label class="field-label">Name <span class="required">*</span></label>
          <BaseInput v-model="createForm.name" placeholder="e.g. Weekly Newsletter" />
        </div>
        <div class="form-field">
          <label class="field-label">Description</label>
          <BaseInput v-model="createForm.description" placeholder="Optional description" />
        </div>
        <div class="form-field">
          <div class="toggle-row">
            <button
              type="button"
              class="toggle-switch"
              :class="{ on: createForm.double_opt_in }"
              @click="createForm.double_opt_in = !createForm.double_opt_in"
              role="switch"
              :aria-checked="createForm.double_opt_in"
            >
              <span class="toggle-thumb" />
            </button>
            <span class="toggle-text">
              <strong>Double opt-in</strong> — subscribers confirm via email before being added
            </span>
          </div>
        </div>
        <p v-if="createError" class="form-error">{{ createError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeCreate">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="createLoading">Create List</BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- Edit List Modal -->
    <BaseModal :show="showEdit" title="Edit List" @close="closeEdit">
      <form @submit.prevent="submitEdit" class="form">
        <div class="form-field">
          <label class="field-label">Name <span class="required">*</span></label>
          <BaseInput v-model="editForm.name" placeholder="e.g. Weekly Newsletter" />
        </div>
        <div class="form-field">
          <label class="field-label">Description</label>
          <BaseInput v-model="editForm.description" placeholder="Optional description" />
        </div>
        <p v-if="editError" class="form-error">{{ editError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeEdit">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="editLoading">Save Changes</BaseButton>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, LayoutList } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseInput from '@/components/BaseInput.vue'
import BaseModal from '@/components/BaseModal.vue'

const lists = ref([])
const loading = ref(false)

const showCreate = ref(false)
const createLoading = ref(false)
const createError = ref('')
const createForm = ref({ name: '', description: '', double_opt_in: false })

const showEdit = ref(false)
const editLoading = ref(false)
const editError = ref('')
const editForm = ref({ id: '', name: '', description: '' })

const cols = [
  { key: 'name', label: 'List Name' },
  { key: 'subscriber_count', label: 'Subscribers' },
  { key: 'double_opt_in', label: 'Opt-in Type' },
  { key: 'created_at', label: 'Created' },
  { key: '_actions', label: '' },
]

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

async function fetchLists() {
  loading.value = true
  try {
    const res = await api.get('/api/lists')
    lists.value = res.data.lists ?? []
  } catch (e) {
    console.error('fetch lists error', e)
  } finally {
    loading.value = false
  }
}

function closeCreate() {
  showCreate.value = false
  createForm.value = { name: '', description: '', double_opt_in: false }
  createError.value = ''
}

async function submitCreate() {
  createError.value = ''
  if (!createForm.value.name.trim()) {
    createError.value = 'Name is required.'
    return
  }
  createLoading.value = true
  try {
    await api.post('/api/lists', {
      name: createForm.value.name.trim(),
      description: createForm.value.description.trim(),
      double_opt_in: createForm.value.double_opt_in,
    })
    closeCreate()
    fetchLists()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'duplicate') {
      createError.value = 'A list with that name already exists.'
    } else {
      createError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    createLoading.value = false
  }
}

function openEdit(list) {
  editForm.value = { id: list.id, name: list.name, description: list.description || '' }
  editError.value = ''
  showEdit.value = true
}

function closeEdit() {
  showEdit.value = false
  editForm.value = { id: '', name: '', description: '' }
  editError.value = ''
}

async function submitEdit() {
  editError.value = ''
  if (!editForm.value.name.trim()) {
    editError.value = 'Name is required.'
    return
  }
  editLoading.value = true
  try {
    await api.put(`/api/lists/${editForm.value.id}`, {
      name: editForm.value.name.trim(),
      description: editForm.value.description.trim(),
    })
    closeEdit()
    fetchLists()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'duplicate') {
      editError.value = 'A list with that name already exists.'
    } else {
      editError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    editLoading.value = false
  }
}

onMounted(fetchLists)
</script>

<style scoped>
.lists-view {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.list-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.list-card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 18px 20px;
  cursor: pointer;
  transition: border-color 120ms, box-shadow 120ms;
}

.list-card:hover {
  border-color: var(--accent);
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.08);
}

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.list-icon {
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: var(--accent-light);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: #111;
  margin-bottom: 4px;
}

.card-count {
  font-size: 22px;
  font-weight: 700;
  color: var(--accent);
  line-height: 1;
}

.card-meta {
  font-size: 11px;
  color: #aaa;
  margin-top: 3px;
}

.table-header {
  padding: 16px 22px 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.list-name {
  font-weight: 500;
  color: #111;
}

.count-cell {
  font-weight: 600;
}

.muted {
  color: var(--text-muted);
}

.actions-cell {
  display: flex;
  gap: 4px;
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

/* Styled toggle switch */
.toggle-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.toggle-switch {
  position: relative;
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background: #d1d5db;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
  margin-top: 1px;
  transition: background 150ms;
  padding: 0;
}

.toggle-switch.on {
  background: var(--accent);
}

.toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  transition: transform 150ms;
  display: block;
}

.toggle-switch.on .toggle-thumb {
  transform: translateX(16px);
}

.toggle-text {
  font-size: 13px;
  color: #444;
  line-height: 1.5;
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
