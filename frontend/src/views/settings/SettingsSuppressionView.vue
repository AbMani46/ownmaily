<template>
  <div class="suppression-section">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="search-wrap">
        <svg class="search-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input v-model="searchInput" class="search-input" placeholder="Search suppressed emails…" />
      </div>
      <button class="btn-secondary" @click="exportList">
        <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
        Export
      </button>
      <button class="btn-primary" @click="showAddModal = true">
        <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        Add Email
      </button>
    </div>

    <!-- Warning callout -->
    <div class="warn-callout">
      <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#c2410c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      <p>Suppressed emails will never receive campaigns.</p>
    </div>

    <!-- Table -->
    <div class="card table-card">
      <div v-if="loading" class="table-placeholder">Loading…</div>
      <template v-else>
        <div v-if="suppressions.length === 0 && !searchInput" class="table-placeholder">No suppressed emails yet.</div>
        <div v-else-if="suppressions.length === 0" class="table-placeholder">No results for "{{ searchInput }}".</div>
        <table v-else class="sup-table">
          <thead>
            <tr>
              <th>Email</th>
              <th>Reason</th>
              <th>Added</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in suppressions" :key="row.email" class="table-row">
              <td class="email-cell">{{ row.email }}</td>
              <td><span class="reason-badge" :class="reasonClass(row.reason)">{{ formatReason(row.reason) }}</span></td>
              <td class="muted">{{ formatDate(row.created_at) }}</td>
              <td>
                <button class="remove-btn" @click="removeEmail(row.email)">Remove</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="total > perPage" class="pagination-row">
          <button :disabled="page <= 1" class="page-btn" @click="page--">← Prev</button>
          <span class="page-info">{{ page }} / {{ Math.ceil(total / perPage) }}</span>
          <button :disabled="page >= Math.ceil(total / perPage)" class="page-btn" @click="page++">Next →</button>
        </div>
      </template>
    </div>

    <!-- Add modal -->
    <teleport to="body">
      <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
        <div class="modal">
          <h3>Add to suppression list</h3>
          <div class="field">
            <label>Email address</label>
            <input v-model="addEmail" type="email" placeholder="bad@example.com" autofocus @keyup.enter="submitAdd" />
          </div>
          <div v-if="addError" class="error-msg">{{ addError }}</div>
          <div class="modal-actions">
            <button class="btn-primary" :disabled="addLoading" @click="submitAdd">
              {{ addLoading ? 'Adding…' : 'Add email' }}
            </button>
            <button class="btn-ghost" @click="showAddModal = false">Cancel</button>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script>
import { ref, watch, onMounted } from 'vue'
import api from '@/lib/api'

export default {
  name: 'SettingsSuppressionView',
  setup() {
    const suppressions = ref([])
    const loading = ref(true)
    const searchInput = ref('')
    const page = ref(1)
    const perPage = 25
    const total = ref(0)
    const showAddModal = ref(false)
    const addEmail = ref('')
    const addLoading = ref(false)
    const addError = ref('')

    let searchTimer = null

    async function load() {
      loading.value = true
      try {
        const params = { page: page.value, per_page: perPage }
        if (searchInput.value) params.q = searchInput.value
        const { data } = await api.get('/api/settings/suppressions', { params })
        suppressions.value = data.suppressions || []
        total.value = data.total || 0
      } finally {
        loading.value = false
      }
    }

    watch(searchInput, () => {
      clearTimeout(searchTimer)
      page.value = 1
      searchTimer = setTimeout(load, 300)
    })

    watch(page, load)
    onMounted(load)

    function formatDate(s) {
      if (!s) return '—'
      return new Date(s).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    }

    function formatReason(r) {
      const map = { hard_bounce: 'Hard bounce', complained: 'Spam complaint', unsubscribed: 'Unsubscribed', manual: 'Manually added' }
      return map[r] || r
    }

    function reasonClass(r) {
      if (r === 'hard_bounce') return 'reason-bounce'
      if (r === 'complained') return 'reason-spam'
      if (r === 'unsubscribed') return 'reason-unsub'
      return 'reason-manual'
    }

    async function removeEmail(email) {
      if (!confirm(`Remove ${email} from suppression list?`)) return
      try {
        await api.delete('/api/settings/suppressions/' + encodeURIComponent(email))
        await load()
      } catch (e) {
        alert(e.response?.data?.message || 'Remove failed')
      }
    }

    function exportList() {
      window.location.href = '/api/settings/suppressions/export'
    }

    async function submitAdd() {
      if (!addEmail.value || !addEmail.value.includes('@')) {
        addError.value = 'Enter a valid email address'
        return
      }
      addLoading.value = true
      addError.value = ''
      try {
        await api.post('/api/settings/suppressions', { email: addEmail.value })
        addEmail.value = ''
        showAddModal.value = false
        await load()
      } catch (e) {
        addError.value = e.response?.data?.message || 'Add failed'
      } finally {
        addLoading.value = false
      }
    }

    return {
      suppressions, loading, searchInput, page, perPage, total,
      showAddModal, addEmail, addLoading, addError,
      formatDate, formatReason, reasonClass, removeEmail, exportList, submitAdd,
    }
  },
}
</script>

<style scoped>
.suppression-section { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; gap: 8px; align-items: center; }
.search-wrap { flex: 1; position: relative; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); color: #bbb; }
.search-input {
  width: 100%;
  padding: 8px 12px 8px 32px;
  border: 1px solid var(--border);
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
  transition: border-color 120ms;
}
.search-input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(16,185,129,0.12); }
.btn-primary {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 8px 14px; background: var(--accent); color: #fff;
  border: none; border-radius: var(--radius-btn); font-size: 13px; font-weight: 600;
  font-family: inherit; cursor: pointer; white-space: nowrap; transition: background 120ms;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-secondary {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 8px 14px; background: #fff; color: var(--text-secondary);
  border: 1px solid var(--border); border-radius: var(--radius-btn); font-size: 13px;
  font-family: inherit; cursor: pointer; white-space: nowrap; transition: all 120ms;
}
.btn-secondary:hover { border-color: #aaa; color: var(--text-primary); }
.warn-callout {
  display: flex; gap: 10px; align-items: flex-start;
  background: #fff7ed; border: 1px solid #fed7aa; border-radius: 8px; padding: 12px 14px;
}
.warn-callout p { margin: 0; font-size: 12px; color: #9a3412; line-height: 1.5; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-card); overflow: hidden; }
.table-card { }
.table-placeholder { padding: 40px 20px; text-align: center; color: var(--text-muted); font-size: 13px; }
.sup-table { width: 100%; border-collapse: collapse; }
.sup-table thead tr { border-bottom: 1px solid var(--border-subtle); }
.sup-table th {
  padding: 8px 20px; text-align: left;
  font-size: 10px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted);
}
.table-row { border-bottom: 1px solid var(--border-subtle); }
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: #faf8f4; }
.sup-table td { padding: 11px 20px; font-size: 13px; color: var(--text-primary); }
.email-cell { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
.reason-badge {
  display: inline-block; padding: 2px 8px; border-radius: 20px;
  font-size: 11px; font-weight: 500;
}
.reason-bounce { background: #fef2f2; color: #991b1b; }
.reason-spam { background: #fff7ed; color: #9a3412; }
.reason-unsub { background: #f0fdf4; color: #166534; }
.reason-manual { background: #f5f5f5; color: #555; }
.muted { color: var(--text-muted); }
.remove-btn {
  padding: 4px 10px;
  background: transparent; color: #dc2626;
  border: 1px solid #fca5a5; border-radius: 5px;
  font-size: 12px; font-family: inherit; cursor: pointer; transition: all 120ms;
}
.remove-btn:hover { background: #fef2f2; }
.pagination-row {
  display: flex; align-items: center; gap: 12px; padding: 12px 20px;
  border-top: 1px solid var(--border-subtle); font-size: 13px;
}
.page-btn {
  padding: 5px 12px; border: 1px solid var(--border); border-radius: 6px;
  background: #fff; font-size: 12px; font-family: inherit; cursor: pointer;
}
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { color: var(--text-muted); font-size: 12px; }
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.35);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: #fff; border-radius: 12px; padding: 28px;
  width: 400px; box-shadow: 0 8px 40px rgba(0,0,0,0.18);
  display: flex; flex-direction: column; gap: 16px;
}
.modal h3 { font-size: 15px; font-weight: 600; color: var(--text-primary); margin: 0; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 12px; font-weight: 600; color: #444; letter-spacing: 0.02em; }
.field input {
  padding: 8px 11px; border: 1px solid var(--border); border-radius: var(--radius-input);
  font-size: 13px; font-family: inherit; outline: none; transition: border-color 120ms;
}
.field input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(16,185,129,0.12); }
.error-msg { font-size: 12px; color: #dc2626; }
.modal-actions { display: flex; gap: 8px; }
.btn-ghost {
  padding: 8px 14px; background: transparent; color: var(--text-secondary);
  border: 1px solid var(--border); border-radius: var(--radius-btn);
  font-size: 13px; font-family: inherit; cursor: pointer;
}
</style>
