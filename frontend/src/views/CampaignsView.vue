<template>
  <div class="campaigns-view">
    <div class="action-bar">
      <span class="page-title">Campaigns</span>
      <BaseButton variant="primary" @click="$router.push('/campaigns/new')">
        <Plus :size="13" :stroke-width="2" />
        Create Campaign
      </BaseButton>
    </div>

    <BaseCard :noPadding="true">
      <!-- Tabs -->
      <div class="tabs-wrap">
        <button
          v-for="t in tabs"
          :key="t.id"
          class="tab"
          :class="{ active: activeTab === t.id }"
          @click="setTab(t.id)"
        >
          {{ t.label }}
          <span class="tab-count" :class="{ 'active': activeTab === t.id }">{{ t.count }}</span>
        </button>
      </div>

      <!-- Table -->
      <BaseTable
        :columns="cols"
        :rows="campaigns"
        @row-click="onRowClick"
      >
        <template #cell-name="{ value, row }">
          <div class="camp-name">{{ value }}</div>
          <div class="camp-target">{{ row.send_to_type === 'tag' ? '#' : '' }}{{ targetLabel(row) }}</div>
        </template>
        <template #cell-status="{ value }">
          <BaseBadge :status="value" />
        </template>
        <template #cell-_date="{ row }">
          <span class="muted">{{ campaignDate(row) }}</span>
        </template>
        <template #cell-_open_rate="{ row }">
          <span v-if="row.status === 'sent'" class="rate-green">
            {{ row.open_rate != null ? formatRate(row.open_rate) : '—' }}
          </span>
          <span v-else class="muted-dash">—</span>
        </template>
        <template #cell-_click_rate="{ row }">
          <span v-if="row.status === 'sent'" class="rate-indigo">
            {{ row.click_rate != null ? formatRate(row.click_rate) : '—' }}
          </span>
          <span v-else class="muted-dash">—</span>
        </template>
        <template #cell-_actions="{ row }">
          <div class="actions-cell" @click.stop>
            <template v-if="row.status === 'sent'">
              <BaseButton variant="ghost" @click="$router.push('/campaigns/' + row.id + '/stats')">
                Stats
              </BaseButton>
            </template>
            <template v-else>
              <BaseButton variant="ghost" @click="$router.push('/campaigns/' + row.id + '/edit')">
                Edit
              </BaseButton>
              <BaseButton variant="ghost" @click="handleDelete(row)" class="delete-btn">
                <Trash2 :size="12" color="#dc2626" />
              </BaseButton>
            </template>
          </div>
        </template>
        <template #empty>
          <div class="empty-campaigns">
            <span>No campaigns yet.</span>
            <BaseButton variant="primary" @click="$router.push('/campaigns/new')">
              <Plus :size="13" :stroke-width="2" />
              Create your first campaign
            </BaseButton>
          </div>
        </template>
      </BaseTable>

      <div v-if="total > PER_PAGE" class="pagination-wrap">
        <BasePagination :page="page" :perPage="PER_PAGE" :total="total" @update:page="setPage" />
      </div>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Plus, Trash2 } from 'lucide-vue-next'
import api from '@/lib/api'
import { useConfirm } from '@/composables/useConfirm'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'
import BasePagination from '@/components/BasePagination.vue'

const router = useRouter()
const { confirm } = useConfirm()

const PER_PAGE = 50

const campaigns = ref([])
const total = ref(0)
const page = ref(1)
const activeTab = ref('all')
const counts = ref({ all: 0, sent: 0, scheduled: 0, draft: 0 })
const listsMap = ref({})
const tagsMap = ref({})

const tabs = computed(() => [
  { id: 'all', label: 'All', count: counts.value.all },
  { id: 'sent', label: 'Sent', count: counts.value.sent },
  { id: 'scheduled', label: 'Scheduled', count: counts.value.scheduled },
  { id: 'draft', label: 'Drafts', count: counts.value.draft },
])

const dateColumnLabel = computed(() => {
  if (activeTab.value === 'sent') return 'Sent'
  if (activeTab.value === 'scheduled') return 'Scheduled For'
  return 'Created'
})

const cols = computed(() => [
  { key: 'name', label: 'Campaign' },
  { key: 'status', label: 'Status' },
  { key: '_date', label: dateColumnLabel.value },
  { key: '_open_rate', label: 'Open Rate' },
  { key: '_click_rate', label: 'Click Rate' },
  { key: '_actions', label: '' },
])

function targetLabel(row) {
  if (!row.send_to_id) return '—'
  if (row.send_to_type === 'tag') {
    return tagsMap.value[row.send_to_id] || 'Tag'
  }
  return listsMap.value[row.send_to_id] || 'List'
}

function campaignDate(row) {
  const date = row.status === 'sent' ? row.sent_at
    : row.status === 'scheduled' ? row.scheduled_at
    : row.created_at
  if (!date) return '—'
  const d = new Date(date)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatRate(rate) {
  return (rate * 100).toFixed(1) + '%'
}

function onRowClick(row) {
  if (row.status === 'sent') {
    router.push('/campaigns/' + row.id + '/stats')
  } else {
    router.push('/campaigns/' + row.id + '/edit')
  }
}

async function fetchCampaigns() {
  try {
    const params = { page: page.value, per_page: PER_PAGE }
    if (activeTab.value !== 'all') params.status = activeTab.value
    const res = await api.get('/api/campaigns', { params })
    campaigns.value = res.data.campaigns ?? []
    total.value = res.data.total ?? 0
    if (activeTab.value !== 'all') {
      counts.value[activeTab.value] = res.data.total ?? 0
    }
  } catch (e) {
    console.error('fetch campaigns error', e)
  }
}

async function fetchAllCounts() {
  try {
    const [all, sent, scheduled, draft] = await Promise.all([
      api.get('/api/campaigns', { params: { per_page: 1 } }),
      api.get('/api/campaigns', { params: { status: 'sent', per_page: 1 } }),
      api.get('/api/campaigns', { params: { status: 'scheduled', per_page: 1 } }),
      api.get('/api/campaigns', { params: { status: 'draft', per_page: 1 } }),
    ])
    counts.value = {
      all: all.data.total ?? 0,
      sent: sent.data.total ?? 0,
      scheduled: scheduled.data.total ?? 0,
      draft: draft.data.total ?? 0,
    }
  } catch (e) {
    // non-critical
  }
}

function setTab(id) {
  activeTab.value = id
  page.value = 1
  fetchCampaigns()
}

function setPage(p) {
  page.value = p
  fetchCampaigns()
}

async function fetchNameMaps() {
  try {
    const [lr, tr] = await Promise.all([
      api.get('/api/lists'),
      api.get('/api/tags'),
    ])
    const lm = {}
    for (const l of lr.data.lists ?? []) lm[l.id] = l.name
    listsMap.value = lm
    const tm = {}
    for (const t of tr.data.tags ?? []) tm[t.id] = t.name
    tagsMap.value = tm
  } catch (e) {
    // non-critical
  }
}

async function handleDelete(row) {
  const scheduledNote = row.status === 'scheduled'
    ? ' The scheduled send will be cancelled.'
    : ''
  const yes = confirm(`Delete campaign "${row.name}"?${scheduledNote} This cannot be undone.`)
  if (!yes) return
  try {
    await api.delete(`/api/campaigns/${row.id}`)
    fetchCampaigns()
    fetchAllCounts()
  } catch (e) {
    const code = e.response?.data?.error
    if (code === 'campaign_locked') {
      alert('Cannot delete a campaign that is currently sending.')
    }
  }
}

onMounted(() => {
  fetchCampaigns()
  fetchAllCounts()
  fetchNameMaps()
})
</script>

<style scoped>
.campaigns-view {
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

.camp-name {
  font-weight: 500;
  color: #111;
  margin-bottom: 1px;
}

.camp-target {
  font-size: 11px;
  color: #aaa;
}

.muted {
  color: var(--text-muted);
  font-size: 13px;
}

.muted-dash {
  color: #ccc;
}

.rate-green {
  font-weight: 600;
  color: #059669;
}

.rate-indigo {
  font-weight: 600;
  color: #6366f1;
}

.actions-cell {
  display: flex;
  gap: 4px;
  align-items: center;
}

.delete-btn {
  color: #dc2626;
}

.pagination-wrap {
  padding: 0 14px 14px;
}

.empty-campaigns {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--text-muted);
  font-size: 13px;
}
</style>
