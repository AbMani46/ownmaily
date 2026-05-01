<template>
  <div class="settings-section">
    <div class="form-section">
      <div class="section-head">
        <h3 class="section-title">API Key</h3>
        <p class="section-desc">Use this key to interact with the OwnMaily REST API.</p>
      </div>
      <div class="card form-card">
        <div class="field">
          <label>Secret key</label>
          <div class="key-row">
            <div class="key-display">
              <span v-if="newKey">{{ newKey }}</span>
              <span v-else>{{ keyPrefix }}{{ '•'.repeat(40) }}</span>
            </div>
            <button class="icon-btn" @click="copyKey" :title="copied ? 'Copied!' : 'Copy'">
              <span v-if="copied" class="copy-check">✓</span>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </button>
          </div>
        </div>

        <!-- One-time reveal after regenerate -->
        <transition name="slide">
          <div v-if="newKey" class="new-key-box">
            <div class="new-key-label">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#d97706" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              Save this — it won't be shown again.
            </div>
          </div>
        </transition>

        <div class="warn-box">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#d97706" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          <p>This key grants full access to your OwnMaily installation. Keep it secret. It is shown in full only once after generation.</p>
        </div>
      </div>
    </div>

    <!-- Regenerate confirm -->
    <div v-if="showConfirm" class="confirm-box">
      <h4>Regenerate API key?</h4>
      <p>Your current key will be immediately invalidated. Any integrations using it will stop working until updated.</p>
      <div class="confirm-actions">
        <button class="btn-danger" :disabled="regenerating" @click="doRegenerate">
          {{ regenerating ? 'Regenerating…' : 'Yes, regenerate' }}
        </button>
        <button class="btn-ghost" @click="showConfirm = false">Cancel</button>
      </div>
    </div>
    <button v-else class="btn-outline" @click="showConfirm = true">
      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
      Regenerate key
    </button>

    <div v-if="error" class="error-msg">{{ error }}</div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import api from '@/lib/api'

export default {
  name: 'SettingsAPIKeyView',
  setup() {
    const keyPrefix = ref('om_')
    const newKey = ref('')
    const copied = ref(false)
    const showConfirm = ref(false)
    const regenerating = ref(false)
    const error = ref('')

    onMounted(async () => {
      try {
        const { data } = await api.get('/api/settings/api-key')
        keyPrefix.value = data.key_prefix || 'om_'
      } catch {}
    })

    function copyKey() {
      const key = newKey.value || keyPrefix.value + '•'.repeat(40)
      if (newKey.value) {
        navigator.clipboard?.writeText(newKey.value)
      }
      copied.value = true
      setTimeout(() => { copied.value = false }, 1000)
    }

    async function doRegenerate() {
      regenerating.value = true
      error.value = ''
      try {
        const { data } = await api.post('/api/settings/api-key/regenerate')
        newKey.value = data.key || ''
        keyPrefix.value = (data.key || '').split('_').slice(0, 2).join('_') + '_'
        showConfirm.value = false
      } catch (e) {
        error.value = e.response?.data?.message || 'Regeneration failed'
      } finally {
        regenerating.value = false
      }
    }

    return { keyPrefix, newKey, copied, showConfirm, regenerating, error, copyKey, doRegenerate }
  },
}
</script>

<style scoped>
.settings-section { max-width: 560px; display: flex; flex-direction: column; gap: 24px; }
.form-section { display: flex; flex-direction: column; gap: 12px; }
.section-head { }
.section-title { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0 0 4px; }
.section-desc { font-size: 13px; color: var(--text-muted); margin: 0; line-height: 1.5; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-card); padding: 18px; }
.form-card { display: flex; flex-direction: column; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 12px; font-weight: 600; color: #444; letter-spacing: 0.02em; }
.key-row { display: flex; gap: 8px; align-items: center; }
.key-display {
  flex: 1;
  padding: 8px 12px;
  background: #0c0c0f;
  border: 1px solid #222;
  border-radius: 7px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: #9aefbc;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.icon-btn {
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: #fff;
  cursor: pointer;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  transition: all 120ms;
}
.icon-btn:hover { border-color: #aaa; }
.copy-check { font-size: 14px; color: var(--accent); font-weight: 700; }
.new-key-box {
  padding: 10px 14px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
}
.new-key-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #92400e;
  font-weight: 500;
}
.warn-box {
  display: flex;
  gap: 10px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  padding: 12px 14px;
}
.warn-box p { margin: 0; font-size: 12px; color: #92400e; line-height: 1.5; }
.confirm-box {
  background: #fff;
  border: 1px solid #fca5a5;
  border-radius: 10px;
  padding: 20px 22px;
}
.confirm-box h4 { margin: 0 0 8px; color: #991b1b; font-size: 14px; font-weight: 600; }
.confirm-box p { margin: 0 0 14px; font-size: 13px; color: #666; line-height: 1.5; }
.confirm-actions { display: flex; gap: 8px; }
.btn-danger {
  padding: 7px 14px;
  background: #dc2626;
  color: #fff;
  border: none;
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
}
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-ghost {
  padding: 7px 14px;
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
}
.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fff;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
}
.btn-outline:hover { border-color: #aaa; color: var(--text-primary); }
.error-msg { font-size: 13px; color: #dc2626; }
.slide-enter-active, .slide-leave-active { transition: all 200ms; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
