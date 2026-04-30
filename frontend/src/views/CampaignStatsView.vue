<template>
  <div class="campaign-stats">
    <!-- Back -->
    <div>
      <BaseButton variant="ghost" @click="$router.push('/campaigns')">
        <ChevronLeft :size="13" :stroke-width="2" />
        Back to Campaigns
      </BaseButton>
    </div>

    <div v-if="loading" class="loading-state">Loading…</div>

    <template v-if="campaign && stats">
      <!-- Header card -->
      <BaseCard>
        <div class="header-row">
          <div class="header-info">
            <h2 class="camp-name">{{ campaign.name }}</h2>
            <div class="camp-meta">
              <BaseBadge :status="campaign.status" />
              <span v-if="campaign.sent_at">Sent {{ formatDate(campaign.sent_at) }}</span>
              <span v-if="campaign.sent_at">&middot;</span>
              <span>{{ stats.sent?.toLocaleString() ?? 0 }} recipients</span>
            </div>
          </div>
          <div class="header-actions">
            <BaseButton variant="secondary" @click="handleDuplicate" :loading="dupeLoading">
              <Copy :size="12" :stroke-width="2" />
              Duplicate
            </BaseButton>
          </div>
        </div>
      </BaseCard>

      <!-- Stat grid -->
      <div class="stat-grid" v-if="stats">
        <div class="stat-tile">
          <div class="stat-tile-label">Sent</div>
          <div class="stat-tile-value">{{ (stats.sent ?? 0).toLocaleString() }}</div>
        </div>
        <div class="stat-tile">
          <div class="stat-tile-label">Opens</div>
          <div class="stat-tile-value">{{ (stats.opens ?? 0).toLocaleString() }}</div>
          <div class="stat-tile-sub">{{ formatRate(stats.open_rate) }} open rate</div>
        </div>
        <div class="stat-tile">
          <div class="stat-tile-label">Clicks</div>
          <div class="stat-tile-value">{{ (stats.clicks ?? 0).toLocaleString() }}</div>
          <div class="stat-tile-sub">{{ formatRate(stats.click_rate) }} click rate</div>
        </div>
        <div class="stat-tile">
          <div class="stat-tile-label">Failed</div>
          <div class="stat-tile-value">{{ (stats.failed ?? 0).toLocaleString() }}</div>
        </div>
        <div class="stat-tile">
          <div class="stat-tile-label">Bounces</div>
          <div class="stat-tile-value">{{ (stats.bounced ?? 0).toLocaleString() }}</div>
        </div>
        <div class="stat-tile">
          <div class="stat-tile-label">Unsubscribes</div>
          <div class="stat-tile-value">{{ (stats.unsubscribed ?? 0).toLocaleString() }}</div>
        </div>
      </div>

      <!-- Link breakdown -->
      <BaseCard>
        <div class="section-title">Link Click Breakdown</div>

        <div v-if="!sortedLinks.length" class="empty-links">
          No data yet — campaign may still be sending.
        </div>

        <div v-else class="link-list">
          <div
            v-for="(link, i) in sortedLinks"
            :key="link.link_url"
            class="link-row"
          >
            <div class="link-top">
              <a :href="link.link_url" target="_blank" rel="noopener" class="link-url">
                {{ truncateUrl(link.link_url) }}
              </a>
              <div class="link-stats">
                <span class="link-count">{{ link.click_count.toLocaleString() }} clicks</span>
                <span class="link-pct">{{ linkPct(link.click_count) }}</span>
              </div>
            </div>
            <div class="link-bar-bg">
              <div
                class="link-bar-fill"
                :style="{
                  width: linkPct(link.click_count),
                  background: i === 0 ? '#10b981' : i === 1 ? '#6366f1' : i === 2 ? '#059669' : '#d1d5db',
                }"
              />
            </div>
          </div>
        </div>
      </BaseCard>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronLeft, Copy } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseCard from '@/components/BaseCard.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'

const route = useRoute()
const router = useRouter()

const campaign = ref(null)
const stats = ref(null)
const loading = ref(false)
const dupeLoading = ref(false)

const sortedLinks = computed(() => {
  if (!stats.value?.links) return []
  return [...stats.value.links].sort((a, b) => b.click_count - a.click_count)
})

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d)) return '—'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatRate(rate) {
  if (rate == null) return '0.0%'
  return (rate * 100).toFixed(1) + '%'
}

function truncateUrl(url) {
  if (!url) return '—'
  try {
    const u = new URL(url)
    const full = u.hostname + u.pathname + u.search
    return full.length > 60 ? full.slice(0, 60) + '…' : full
  } catch {
    return url.length > 60 ? url.slice(0, 60) + '…' : url
  }
}

function linkPct(count) {
  const sent = stats.value?.sent ?? 0
  if (!sent) return '0%'
  return ((count / sent) * 100).toFixed(1) + '%'
}

async function fetchData() {
  loading.value = true
  try {
    const [campRes, statsRes] = await Promise.all([
      api.get(`/api/campaigns/${route.params.id}`),
      api.get(`/api/campaigns/${route.params.id}/stats`),
    ])
    campaign.value = campRes.data
    stats.value = statsRes.data
  } catch (e) {
    if (e.response?.status === 404) router.push('/campaigns')
  } finally {
    loading.value = false
  }
}

async function handleDuplicate() {
  dupeLoading.value = true
  try {
    const res = await api.post(`/api/campaigns/${route.params.id}/duplicate`)
    const newId = res.data.id?.String || res.data.id
    router.push(`/campaigns/${newId}/edit`)
  } catch (e) {
    console.error('duplicate error', e)
  } finally {
    dupeLoading.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.campaign-stats {
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

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.header-info {
  flex: 1;
}

.camp-name {
  font-size: 18px;
  font-weight: 700;
  color: #111;
  margin: 0 0 8px;
  letter-spacing: -0.02em;
}

.camp-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: #888;
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 12px;
}

.stat-tile {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 16px 18px;
}

.stat-tile-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
}

.stat-tile-value {
  font-size: 26px;
  font-weight: 700;
  color: #111;
  letter-spacing: -0.03em;
  line-height: 1;
}

.stat-tile-sub {
  font-size: 11px;
  color: #888;
  margin-top: 4px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 16px;
}

.empty-links {
  font-size: 13px;
  color: var(--text-muted);
  padding: 20px 0;
  text-align: center;
}

.link-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.link-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.link-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.link-url {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: #6366f1;
  text-decoration: none;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-url:hover {
  text-decoration: underline;
}

.link-stats {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.link-count {
  font-weight: 600;
  font-size: 13px;
  color: #555;
}

.link-pct {
  font-size: 12px;
  color: #aaa;
  min-width: 44px;
  text-align: right;
}

.link-bar-bg {
  height: 5px;
  background: #f0ede6;
  border-radius: 3px;
  overflow: hidden;
}

.link-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 600ms ease;
}
</style>
