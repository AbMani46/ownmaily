<template>
  <div class="settings-section">
    <div class="form-section">
      <div class="section-head">
        <h3 class="section-title">Site Information</h3>
        <p class="section-desc">Basic details about your installation.</p>
      </div>
      <div class="card form-card">
        <div class="field">
          <label>Site name</label>
          <input v-model="form.site_name" type="text" placeholder="My Newsletter" />
        </div>
        <div class="field">
          <label>Installation URL</label>
          <input v-model="form.installation_url" type="url" placeholder="https://mail.example.com" />
        </div>
        <div class="field">
          <label>Timezone</label>
          <select v-model="form.timezone">
            <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
          </select>
        </div>
      </div>
    </div>

    <div class="form-section">
      <div class="section-head">
        <h3 class="section-title">Sender Defaults</h3>
        <p class="section-desc">Used as defaults for new campaigns.</p>
      </div>
      <div class="card form-card">
        <div class="field">
          <label>From name</label>
          <input v-model="form.from_name" type="text" placeholder="OwnMaily Team" />
        </div>
        <div class="field">
          <label>From email</label>
          <input v-model="form.from_email" type="email" placeholder="hello@example.com" />
        </div>
        <div class="field">
          <label>Reply-to</label>
          <input v-model="form.reply_to" type="email" placeholder="hello@example.com" />
        </div>
      </div>
    </div>

    <div class="form-section">
      <div class="section-head">
        <h3 class="section-title">Compliance</h3>
        <p class="section-desc">Required for CAN-SPAM and GDPR compliance.</p>
      </div>
      <div class="card form-card">
        <div class="field">
          <label>Physical address</label>
          <textarea v-model="form.physical_address" rows="3" placeholder="123 Main St&#10;City, State ZIP&#10;Country"></textarea>
          <span class="field-hint">Shown in the footer of every email you send.</span>
        </div>
      </div>
    </div>

    <div class="save-row">
      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? 'Saving…' : 'Save changes' }}
      </button>
      <transition name="fade">
        <span v-if="saved" class="saved-msg">Saved ✓</span>
      </transition>
      <span v-if="error" class="error-msg">{{ error }}</span>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import api from '@/lib/api'

export default {
  name: 'SettingsGeneralView',
  setup() {
    const form = ref({
      site_name: '',
      installation_url: '',
      timezone: 'UTC',
      physical_address: '',
      from_name: '',
      from_email: '',
      reply_to: '',
    })
    const saving = ref(false)
    const saved = ref(false)
    const error = ref('')

    const timezones = [
      'UTC', 'America/New_York', 'America/Chicago', 'America/Denver',
      'America/Los_Angeles', 'Europe/London', 'Europe/Berlin', 'Europe/Paris',
      'Asia/Tokyo', 'Asia/Shanghai', 'Asia/Singapore', 'Australia/Sydney',
    ]

    onMounted(async () => {
      try {
        const { data } = await api.get('/api/settings')
        form.value = {
          site_name: data.site_name || '',
          installation_url: data.installation_url || '',
          timezone: data.timezone || 'UTC',
          physical_address: data.physical_address || '',
          from_name: data.from_name || '',
          from_email: data.from_email || '',
          reply_to: data.reply_to || '',
        }
      } catch {}
    })

    async function save() {
      saving.value = true
      error.value = ''
      try {
        await api.put('/api/settings/general', form.value)
        saved.value = true
        setTimeout(() => { saved.value = false }, 2000)
      } catch (e) {
        error.value = e.response?.data?.message || 'Save failed'
      } finally {
        saving.value = false
      }
    }

    return { form, saving, saved, error, timezones, save }
  },
}
</script>

<style scoped>
.settings-section { max-width: 560px; display: flex; flex-direction: column; gap: 28px; }
.form-section { display: flex; flex-direction: column; gap: 12px; }
.section-head { }
.section-title { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0 0 4px; }
.section-desc { font-size: 13px; color: var(--text-muted); margin: 0; line-height: 1.5; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-card); padding: 18px; }
.form-card { display: flex; flex-direction: column; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 12px; font-weight: 600; color: #444; letter-spacing: 0.02em; }
.field input, .field select, .field textarea {
  padding: 8px 11px;
  border: 1px solid var(--border);
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  outline: none;
  transition: border-color 120ms;
  background: #fff;
}
.field input:focus, .field select:focus, .field textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16,185,129,0.12);
}
.field textarea { resize: vertical; line-height: 1.5; }
.field-hint { font-size: 11px; color: var(--text-muted); }
.save-row { display: flex; align-items: center; gap: 10px; }
.btn-primary {
  padding: 8px 18px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 120ms;
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.saved-msg { font-size: 13px; font-weight: 500; color: var(--accent); }
.error-msg { font-size: 13px; color: #dc2626; }
.fade-enter-active, .fade-leave-active { transition: opacity 200ms; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
