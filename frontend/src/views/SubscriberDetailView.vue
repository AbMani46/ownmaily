<template>
  <div class="detail" v-if="!loading">
    <BaseButton variant="ghost" @click="$router.push('/subscribers')">
      <ChevronLeft :size="12" :stroke-width="2.5" />
      Back to Subscribers
    </BaseButton>

    <div class="two-col">
      <!-- Left column -->
      <div class="left-col">
        <!-- Subscriber card -->
        <BaseCard>
          <div class="sub-header">
            <div class="avatar" :style="{ background: avatarBg }">
              {{ avatarInitial }}
            </div>
            <div class="sub-info">
              <div class="sub-name">{{ fullName }}</div>
              <div class="sub-email">{{ subscriber.email }}</div>
            </div>
            <BaseBadge :status="subscriber.status" />
          </div>
          <div class="info-grid">
            <div class="info-tile">
              <div class="tile-key">Added</div>
              <div class="tile-val">{{ formatDate(subscriber.created_at) }}</div>
            </div>
            <div class="info-tile">
              <div class="tile-key">Last Active</div>
              <div class="tile-val">{{ formatDate(subscriber.last_active) }}</div>
            </div>
            <div class="info-tile">
              <div class="tile-key">Source</div>
              <div class="tile-val">{{ subscriber.source || '—' }}</div>
            </div>
            <div class="info-tile">
              <div class="tile-key">Status</div>
              <div class="tile-val">{{ subscriber.status }}</div>
            </div>
          </div>
        </BaseCard>

        <!-- Tags card -->
        <BaseCard>
          <div class="section-header" ref="tagCardRef">
            <span class="section-title">Tags</span>
            <BaseButton variant="ghost" @click="showTagDropdown = !showTagDropdown">
              <Plus :size="11" :stroke-width="2.5" />
              Add tag
            </BaseButton>
          </div>

          <div v-if="showTagDropdown" class="tag-dropdown">
            <input
              v-model="tagSearch"
              class="tag-search"
              placeholder="Search tags…"
              @focus="fetchAvailableTags"
            />
            <div class="tag-list">
              <div
                v-for="tag in filteredAvailableTags"
                :key="tag.id"
                class="tag-option"
                @click="addTag(tag)"
              >
                <span class="tag-dot" :style="{ background: tagColor(tag.name) }" />
                {{ tag.name }}
              </div>
              <div v-if="filteredAvailableTags.length === 0" class="tag-empty">No tags found</div>
            </div>
          </div>

          <div class="tags-pills">
            <span
              v-for="tag in subscriberTags"
              :key="tag.id"
              class="tag-pill"
              :style="{ background: tagColor(tag.name) + '18', color: tagColor(tag.name), borderColor: tagColor(tag.name) + '30' }"
            >
              {{ tag.name }}
              <button class="tag-remove" @click="removeTag(tag)" :style="{ color: tagColor(tag.name) }">×</button>
            </span>
            <span v-if="subscriberTags.length === 0" class="no-tags">No tags assigned</span>
          </div>
        </BaseCard>

        <!-- List memberships -->
        <BaseCard>
          <div class="section-header">
            <span class="section-title">List Memberships</span>
          </div>
          <div class="list-memberships">
            <div
              v-for="list in subscriberLists"
              :key="list.id"
              class="list-row"
            >
              <div>
                <div class="list-name">{{ list.name }}</div>
                <div class="list-date">Subscribed {{ formatDate(subscriber.created_at) }}</div>
              </div>
              <BaseBadge :status="list.double_opt_in ? 'double_optin' : 'single_optin'" />
            </div>
            <div v-if="subscriberLists.length === 0" class="no-lists">Not on any lists.</div>
          </div>
        </BaseCard>
      </div>

      <!-- Right column -->
      <div class="right-col">
        <!-- Stat cards -->
        <div class="stats-row">
          <StatCard label="Campaigns" :value="stats?.campaigns_received ?? '—'" sub="received" />
          <StatCard label="Opens" :value="stats?.total_opens ?? '—'" sub="total" />
          <StatCard label="Clicks" :value="stats?.total_clicks ?? '—'" sub="total" />
        </div>

        <!-- Actions card -->
        <BaseCard>
          <div class="section-header">
            <span class="section-title">Actions</span>
          </div>
          <div class="action-buttons">
            <BaseButton
              variant="secondary"
              :loading="unsubLoading"
              :disabled="subscriber.status === 'unsubscribed'"
              @click="handleUnsubscribe"
            >
              <UserMinus :size="13" :stroke-width="2" />
              {{ subscriber.status === 'unsubscribed' ? 'Unsubscribed' : 'Unsubscribe' }}
            </BaseButton>
            <BaseButton
              variant="danger"
              :loading="deleteLoading"
              @click="handleDelete"
            >
              <Trash2 :size="13" :stroke-width="2" />
              Delete Subscriber
            </BaseButton>
          </div>
        </BaseCard>
      </div>
    </div>
  </div>

  <!-- Loading state -->
  <div v-else class="loading-state">
    <div class="sk sk-block" />
    <div class="two-col" style="margin-top:20px">
      <div class="left-col">
        <div class="sk sk-card" />
        <div class="sk sk-card" />
      </div>
      <div class="right-col">
        <div class="sk sk-card" style="height:80px" />
        <div class="sk sk-card" style="height:120px" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronLeft, Plus, UserMinus, Trash2 } from 'lucide-vue-next'
import api from '@/lib/api'
import BaseCard from '@/components/BaseCard.vue'
import BaseBadge from '@/components/BaseBadge.vue'
import BaseButton from '@/components/BaseButton.vue'
import StatCard from '@/components/StatCard.vue'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id)

const loading = ref(true)
const subscriber = ref({})
const subscriberTags = ref([])
const subscriberLists = ref([])
const stats = ref(null)

const showTagDropdown = ref(false)
const tagSearch = ref('')
const availableTags = ref([])
const tagCardRef = ref(null)

function onDocumentClick(e) {
  if (showTagDropdown.value && tagCardRef.value && !tagCardRef.value.contains(e.target)) {
    showTagDropdown.value = false
  }
}

const unsubLoading = ref(false)
const deleteLoading = ref(false)

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

const fullName = computed(() => {
  const parts = [subscriber.value.first_name, subscriber.value.last_name].filter(Boolean)
  return parts.length ? parts.join(' ') : subscriber.value.email || ''
})

const avatarInitial = computed(() => {
  const name = fullName.value
  return name ? name.charAt(0).toUpperCase() : '?'
})

const avatarBg = computed(() => {
  const initial = avatarInitial.value
  const colors = ['#d1fae5', '#dbeafe', '#ede9fe', '#fef3c7', '#fee2e2']
  const idx = (initial.charCodeAt(0) || 0) % colors.length
  return colors[idx]
})

const filteredAvailableTags = computed(() => {
  const q = tagSearch.value.toLowerCase()
  const existingIds = new Set(subscriberTags.value.map(t => t.id))
  return availableTags.value
    .filter(t => !existingIds.has(t.id))
    .filter(t => !q || t.name.toLowerCase().includes(q))
})

async function fetchData() {
  loading.value = true
  try {
    const [subRes, statsRes] = await Promise.all([
      api.get(`/api/subscribers/${id.value}`),
      api.get(`/api/subscribers/${id.value}/stats`),
    ])
    const data = subRes.data
    subscriber.value = data
    subscriberTags.value = data.tags ?? []
    subscriberLists.value = data.lists ?? []
    stats.value = statsRes.data
  } catch (e) {
    console.error('fetch subscriber error', e)
  } finally {
    loading.value = false
  }
}

async function fetchAvailableTags() {
  if (availableTags.value.length > 0) return
  try {
    const res = await api.get('/api/tags')
    availableTags.value = res.data.tags ?? []
  } catch (e) {
    console.error('fetch tags error', e)
  }
}

async function addTag(tag) {
  try {
    await api.post(`/api/subscribers/${id.value}/tags`, { tag_id: tag.id })
    subscriberTags.value.push(tag)
    showTagDropdown.value = false
    tagSearch.value = ''
  } catch (e) {
    console.error('add tag error', e)
  }
}

async function removeTag(tag) {
  try {
    await api.delete(`/api/subscribers/${id.value}/tags/${tag.id}`)
    subscriberTags.value = subscriberTags.value.filter(t => t.id !== tag.id)
  } catch (e) {
    console.error('remove tag error', e)
  }
}

async function handleUnsubscribe() {
  if (!confirm('Unsubscribe this contact? They will no longer receive emails.')) return
  unsubLoading.value = true
  try {
    await api.post(`/api/subscribers/${id.value}/unsubscribe`)
    subscriber.value = { ...subscriber.value, status: 'unsubscribed' }
  } catch (e) {
    console.error('unsubscribe error', e)
  } finally {
    unsubLoading.value = false
  }
}

async function handleDelete() {
  if (!confirm('Permanently delete this subscriber? This cannot be undone.')) return
  deleteLoading.value = true
  try {
    await api.delete(`/api/subscribers/${id.value}`)
    router.push('/subscribers')
  } catch (e) {
    console.error('delete error', e)
    deleteLoading.value = false
  }
}

onMounted(() => {
  fetchData()
  document.addEventListener('click', onDocumentClick)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
})
</script>

<style scoped>
.detail {
  padding: 24px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.two-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.left-col, .right-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Subscriber card header */
.sub-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 18px;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  color: var(--accent);
  flex-shrink: 0;
}

.sub-info {
  flex: 1;
  min-width: 0;
}

.sub-name {
  font-size: 16px;
  font-weight: 600;
  color: #111;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sub-email {
  font-size: 13px;
  color: var(--text-muted);
}

/* Info grid */
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.info-tile {
  background: #faf8f4;
  border-radius: 7px;
  padding: 9px 12px;
}

.tile-key {
  font-size: 10px;
  color: #aaa;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 2px;
}

.tile-val {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}

/* Section header */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

/* Tag dropdown */
.tag-dropdown {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
  padding: 8px;
  margin-bottom: 10px;
  position: relative;
  z-index: 10;
}

.tag-search {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
  margin-bottom: 6px;
}

.tag-search:focus {
  border-color: var(--accent);
}

.tag-list {
  max-height: 160px;
  overflow-y: auto;
}

.tag-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 13px;
  color: #333;
  transition: background 80ms;
}

.tag-option:hover {
  background: #faf8f4;
}

.tag-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tag-empty {
  font-size: 12px;
  color: #bbb;
  padding: 6px 8px;
}

/* Tags pills */
.tags-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  align-items: center;
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px 2px 10px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid;
}

.tag-remove {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  font-size: 14px;
  line-height: 1;
  opacity: 0.7;
  transition: opacity 120ms;
}

.tag-remove:hover {
  opacity: 1;
}

.no-tags {
  font-size: 13px;
  color: #bbb;
}

/* List memberships */
.list-memberships {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.list-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: #faf8f4;
  border-radius: 7px;
}

.list-name {
  font-size: 13px;
  font-weight: 500;
  color: #333;
}

.list-date {
  font-size: 11px;
  color: #aaa;
}

.no-lists {
  font-size: 13px;
  color: #bbb;
}

/* Stats row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

/* Action buttons */
.action-buttons {
  display: flex;
  gap: 8px;
}

/* Loading skeleton */
.loading-state {
  padding: 24px 28px;
}

.sk {
  background: #e8e5de;
  border-radius: 4px;
  animation: pulse 1.5s ease-in-out infinite;
}

.sk-block {
  height: 32px;
  width: 140px;
  border-radius: 7px;
}

.sk-card {
  height: 200px;
  border-radius: var(--radius-card);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}
</style>
