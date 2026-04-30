<template>
  <div class="dashboard">
    <!-- Stat cards -->
    <div class="stats-grid">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="skeleton-card">
          <div class="sk sk-label" />
          <div class="sk sk-value" />
          <div class="sk sk-sub" />
        </div>
      </template>
      <template v-else>
        <StatCard
          label="Total Subscribers"
          :value="fmt(overview?.total_subscribers)"
        />
        <StatCard
          label="Campaigns Sent"
          :value="fmt(overview?.total_campaigns_sent)"
          sub="all time"
        />
        <StatCard
          label="Avg Open Rate"
          :value="fmtRate(overview?.overall_open_rate)"
        />
        <StatCard
          label="Avg Click Rate"
          :value="fmtRate(overview?.overall_click_rate)"
        />
      </template>
    </div>

    <!-- Recent campaigns -->
    <BaseCard :noPadding="true">
      <div class="card-header">
        <span class="section-title">Recent Campaigns</span>
        <RouterLink to="/campaigns" class="view-all">View all →</RouterLink>
      </div>

      <template v-if="loading">
        <div class="sk-table">
          <div v-for="n in 4" :key="n" class="sk-row">
            <div class="sk sk-text-wide" />
            <div class="sk sk-badge" />
            <div class="sk sk-text" />
            <div class="sk sk-text" />
            <div class="sk sk-text" />
          </div>
        </div>
      </template>

      <template v-else>
        <BaseTable
          :columns="campaignCols"
          :rows="campaigns"
          @row-click="row => $router.push('/campaigns/' + row.id + '/stats')"
        >
          <template #cell-name="{ value, row }">
            <div class="campaign-name">{{ value }}</div>
            <div class="campaign-sub">{{ row.subject }}</div>
          </template>
          <template #cell-status="{ value }">
            <BaseBadge :status="value" />
          </template>
          <template #cell-sent_at="{ value }">
            <span class="muted">{{ formatDate(value) }}</span>
          </template>
          <template #cell-open_rate="{ value }">
            <span class="rate-dash">—</span>
          </template>
          <template #cell-click_rate="{ value }">
            <span class="rate-dash">—</span>
          </template>
          <template #empty>
            <div class="empty-state">
              No campaigns sent yet. Create your first campaign.
            </div>
          </template>
        </BaseTable>
      </template>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '@/lib/api'
import StatCard from '@/components/StatCard.vue'
import BaseCard from '@/components/BaseCard.vue'
import BaseTable from '@/components/BaseTable.vue'
import BaseBadge from '@/components/BaseBadge.vue'

const loading = ref(true)
const overview = ref(null)
const campaigns = ref([])

const campaignCols = [
  { key: 'name', label: 'Campaign' },
  { key: 'status', label: 'Status' },
  { key: 'sent_at', label: 'Sent' },
  { key: 'open_rate', label: 'Opens' },
  { key: 'click_rate', label: 'Clicks' },
]

function fmt(n) {
  if (n == null) return '—'
  return Number(n).toLocaleString()
}

function fmtRate(r) {
  if (r == null) return '—'
  return (r * 100).toFixed(1) + '%'
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

onMounted(async () => {
  try {
    const [ovRes, campRes] = await Promise.all([
      api.get('/api/analytics/overview'),
      api.get('/api/campaigns', { params: { per_page: 5, status: 'sent' } }),
    ])
    overview.value = ovRes.data
    campaigns.value = campRes.data.campaigns ?? []
  } catch (e) {
    console.error('dashboard fetch error', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dashboard {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.card-header {
  padding: 16px 22px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.view-all {
  font-size: 12px;
  color: var(--text-muted);
  text-decoration: none;
  transition: color 120ms;
}
.view-all:hover {
  color: var(--accent);
}

.campaign-name {
  font-weight: 500;
  color: #111;
  margin-bottom: 1px;
}

.campaign-sub {
  font-size: 11px;
  color: #aaa;
}

.muted {
  color: var(--text-muted);
}

.rate-dash {
  color: #ccc;
}

.empty-state {
  font-size: 13px;
  color: var(--text-muted);
}

/* Skeleton */
.skeleton-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sk {
  background: #e8e5de;
  border-radius: 4px;
  animation: pulse 1.5s ease-in-out infinite;
}

.sk-label { height: 10px; width: 60%; }
.sk-value { height: 28px; width: 50%; margin-top: 4px; }
.sk-sub { height: 10px; width: 40%; }
.sk-text { height: 12px; width: 60px; }
.sk-text-wide { height: 12px; width: 160px; }
.sk-badge { height: 20px; width: 56px; border-radius: 20px; }

.sk-table {
  padding: 8px 14px 14px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.sk-row {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 13px 0;
  border-bottom: 1px solid var(--border-subtle);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}
</style>
