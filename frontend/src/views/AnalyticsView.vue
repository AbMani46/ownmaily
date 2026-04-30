<template>
  <div class="analytics-page">
    <!-- Period selector -->
    <div class="period-row">
      <button
        v-for="p in periods"
        :key="p.id"
        class="period-pill"
        :class="{ active: activePeriod === p.id }"
        @click="activePeriod = p.id"
      >{{ p.label }}</button>
      <span class="period-note">Period filter is display-only — data shows all time</span>
    </div>

    <!-- Row 1 stat cards -->
    <div v-if="loading" class="skeleton-grid">
      <div v-for="i in 8" :key="i" class="skeleton-card"></div>
    </div>

    <template v-else>
      <div class="stat-grid">
        <StatCard
          label="Total Subscribers"
          :value="fmt(overview.total_subscribers)"
          sub="all time"
        />
        <StatCard
          label="Emails Sent"
          :value="fmt(overview.total_emails_sent)"
          sub="all time"
        />
        <StatCard
          label="Avg Open Rate"
          :value="pct(overview.overall_open_rate)"
          sub="across all campaigns"
        />
        <StatCard
          label="Avg Click Rate"
          :value="pct(overview.overall_click_rate)"
          sub="across all campaigns"
        />
      </div>

      <!-- Row 2 stat cards -->
      <div class="stat-grid">
        <StatCard
          label="Unsubscribed"
          :value="fmt(overview.unsubscribed)"
          sub="all time"
        />
        <StatCard
          label="Bounced"
          :value="fmt(overview.bounced)"
          sub="all time"
        />
        <StatCard
          label="Active Subscribers"
          :value="fmt(overview.active_subscribers)"
          sub="currently active"
        />
        <StatCard
          label="Deliverability"
          :value="deliverability"
          sub="delivery rate"
        />
      </div>

      <!-- Campaign performance table -->
      <div class="card table-card">
        <div class="table-header">
          <h3 class="section-title">Campaign Performance</h3>
        </div>
        <div v-if="campaignsLoading" class="table-loading">Loading campaigns…</div>
        <template v-else>
          <div v-if="campaigns.length === 0" class="empty-state">
            No campaigns sent yet.
          </div>
          <table v-else class="perf-table">
            <thead>
              <tr>
                <th>Campaign</th>
                <th>Sent date</th>
                <th>Sent count</th>
                <th>Open rate</th>
                <th>Click rate</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="c in campaigns"
                :key="c.id"
                class="table-row"
                @click="$router.push('/campaigns/' + c.id + '/stats')"
              >
                <td class="campaign-name">{{ c.name }}</td>
                <td class="muted">{{ formatDate(c.sent_at) }}</td>
                <td class="num">{{ c.sent_count ? fmt(c.sent_count) : '—' }}</td>
                <td class="rate-dash">—</td>
                <td class="rate-dash">—</td>
              </tr>
            </tbody>
          </table>
        </template>
      </div>
    </template>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import api from '@/lib/api'
import StatCard from '@/components/StatCard.vue'

export default {
  name: 'AnalyticsView',
  components: { StatCard },
  setup() {
    const overview = ref({
      total_subscribers: 0,
      total_emails_sent: 0,
      overall_open_rate: 0,
      overall_click_rate: 0,
      unsubscribed: 0,
      bounced: 0,
      active_subscribers: 0,
    })
    const campaigns = ref([])
    const loading = ref(true)
    const campaignsLoading = ref(true)
    const activePeriod = ref('all')

    const periods = [
      { id: '7d', label: '7 days' },
      { id: '30d', label: '30 days' },
      { id: '90d', label: '90 days' },
      { id: 'all', label: 'All time' },
    ]

    const deliverability = computed(() => {
      const sent = overview.value.total_emails_sent
      const bounced = overview.value.bounced
      if (!sent) return '—'
      return (((sent - bounced) / sent) * 100).toFixed(1) + '%'
    })

    function fmt(n) {
      if (n == null) return '0'
      return Number(n).toLocaleString()
    }

    function pct(r) {
      if (r == null) return '—'
      return (Number(r) * 100).toFixed(1) + '%'
    }

    function formatDate(s) {
      if (!s) return '—'
      return new Date(s).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    }

    onMounted(async () => {
      try {
        const { data } = await api.get('/api/analytics/overview')
        overview.value = data
      } finally {
        loading.value = false
      }

      try {
        const { data } = await api.get('/api/campaigns', { params: { status: 'sent', per_page: 20 } })
        campaigns.value = data.campaigns || []
      } finally {
        campaignsLoading.value = false
      }
    })

    return { overview, campaigns, loading, campaignsLoading, activePeriod, periods, deliverability, fmt, pct, formatDate }
  },
}
</script>

<style scoped>
.analytics-page {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.period-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.period-pill {
  padding: 5px 14px;
  border-radius: 20px;
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
}

.period-pill.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

.period-note {
  font-size: 11px;
  color: var(--text-muted);
  margin-left: 6px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.skeleton-card {
  height: 90px;
  border-radius: var(--radius-card);
  background: linear-gradient(90deg, #f0ede6 25%, #faf8f4 50%, #f0ede6 75%);
  background-size: 200% 100%;
  animation: shimmer 1.2s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
}

.table-card {
  overflow: hidden;
}

.table-header {
  padding: 16px 20px 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 14px;
  letter-spacing: -0.01em;
}

.table-loading {
  padding: 32px 20px;
  color: var(--text-muted);
  font-size: 13px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.perf-table {
  width: 100%;
  border-collapse: collapse;
}

.perf-table thead tr {
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
}

.perf-table th {
  padding: 8px 20px;
  text-align: left;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
}

.table-row {
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background 80ms;
}

.table-row:hover {
  background: #faf8f4;
}

.table-row:last-child {
  border-bottom: none;
}

.perf-table td {
  padding: 12px 20px;
  font-size: 13px;
  color: var(--text-primary);
}

.campaign-name {
  font-weight: 500;
}

.muted {
  color: var(--text-muted);
}

.num {
  font-variant-numeric: tabular-nums;
}

.rate-dash {
  color: #ccc;
  font-weight: 400;
}
</style>
