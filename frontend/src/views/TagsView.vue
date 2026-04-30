<template>
  <div class="tags-view">
    <div class="action-bar">
      <span class="page-title">Tags</span>
      <BaseButton variant="primary" @click="showCreate = true">
        <Plus :size="13" :stroke-width="2" />
        Create Tag
      </BaseButton>
    </div>

    <BaseCard :noPadding="true">
      <div class="table-header">
        <span class="section-title">All Tags</span>
      </div>
      <BaseTable :columns="cols" :rows="tags">
        <template #cell-name="{ value, row }">
          <div class="tag-name-cell">
            <span class="tag-dot" :style="{ background: tagColor(value) }" />
            <span class="tag-name">{{ value }}</span>
          </div>
        </template>
        <template #cell-subscriber_count="{ value }">
          <span class="count-cell">{{ (value || 0).toLocaleString() }}</span>
        </template>
        <template #cell-_actions="{ row }">
          <div class="actions-cell" @click.stop>
            <BaseButton variant="ghost" @click="openEdit(row)">Edit</BaseButton>
            <BaseButton variant="ghost" @click="handleDelete(row)" class="delete-btn">
              <Trash2 :size="12" color="#dc2626" />
              Delete
            </BaseButton>
          </div>
        </template>
        <template #empty>
          <div>No tags yet. Create a tag to start organizing subscribers.</div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Create Tag Modal -->
    <BaseModal :show="showCreate" title="Create Tag" @close="closeCreate">
      <form @submit.prevent="submitCreate" class="form">
        <div class="form-field">
          <label class="field-label">Tag name <span class="required">*</span></label>
          <BaseInput v-model="createName" placeholder="e.g. vip-customers" />
        </div>
        <p v-if="createError" class="form-error">{{ createError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeCreate">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="createLoading">Create Tag</BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- Edit Tag Modal -->
    <BaseModal :show="showEdit" title="Rename Tag" @close="closeEdit">
      <form @submit.prevent="submitEdit" class="form">
        <div class="form-field">
          <label class="field-label">Tag name <span class="required">*</span></label>
          <BaseInput v-model="editName" placeholder="Tag name" />
        </div>
        <p v-if="editError" class="form-error">{{ editError }}</p>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeEdit">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit" :loading="editLoading">Save</BaseButton>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, Trash2 } from 'lucide-vue-next'
import api from '@/lib/api'
import { useConfirm } from '@/composables/useConfirm'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseInput from '@/components/BaseInput.vue'
import BaseModal from '@/components/BaseModal.vue'

const { confirm } = useConfirm()

const tags = ref([])

const showCreate = ref(false)
const createLoading = ref(false)
const createName = ref('')
const createError = ref('')

const showEdit = ref(false)
const editLoading = ref(false)
const editName = ref('')
const editError = ref('')
const editingTag = ref(null)

const TAG_COLORS = ['#10b981', '#8b5cf6', '#0ea5e9', '#f59e0b', '#6b7280', '#ec4899']

function tagColor(name) {
  let h = 0
  for (const c of name) h = (h * 31 + c.charCodeAt(0)) % TAG_COLORS.length
  return TAG_COLORS[Math.abs(h)]
}

const cols = [
  { key: 'name', label: 'Tag Name' },
  { key: 'subscriber_count', label: 'Subscribers' },
  { key: '_actions', label: 'Actions' },
]

async function fetchTags() {
  try {
    const res = await api.get('/api/tags')
    tags.value = res.data.tags ?? []
  } catch (e) {
    console.error('fetch tags error', e)
  }
}

function closeCreate() {
  showCreate.value = false
  createName.value = ''
  createError.value = ''
}

async function submitCreate() {
  createError.value = ''
  if (!createName.value.trim()) {
    createError.value = 'Name is required.'
    return
  }
  createLoading.value = true
  try {
    await api.post('/api/tags', { name: createName.value.trim() })
    closeCreate()
    fetchTags()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'duplicate') {
      createError.value = 'A tag with that name already exists.'
    } else {
      createError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    createLoading.value = false
  }
}

function openEdit(tag) {
  editingTag.value = tag
  editName.value = tag.name
  editError.value = ''
  showEdit.value = true
}

function closeEdit() {
  showEdit.value = false
  editingTag.value = null
  editName.value = ''
  editError.value = ''
}

async function submitEdit() {
  editError.value = ''
  if (!editName.value.trim()) {
    editError.value = 'Name is required.'
    return
  }
  editLoading.value = true
  try {
    await api.put(`/api/tags/${editingTag.value.id}`, { name: editName.value.trim() })
    closeEdit()
    fetchTags()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'duplicate') {
      editError.value = 'A tag with that name already exists.'
    } else {
      editError.value = 'Something went wrong. Please try again.'
    }
  } finally {
    editLoading.value = false
  }
}

async function handleDelete(tag) {
  const yes = confirm(`Delete tag "${tag.name}"? This will remove it from all subscribers.`)
  if (!yes) return
  try {
    await api.delete(`/api/tags/${tag.id}`)
    fetchTags()
  } catch (e) {
    console.error('delete tag error', e)
  }
}

onMounted(fetchTags)
</script>

<style scoped>
.tags-view {
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

.tag-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tag-name {
  font-weight: 500;
  color: #111;
}

.count-cell {
  font-weight: 600;
}

.actions-cell {
  display: flex;
  gap: 6px;
}

.delete-btn {
  color: #dc2626;
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
