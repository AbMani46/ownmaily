<template>
  <div class="settings-section">
    <div class="form-section">
      <div class="section-head">
        <h3 class="section-title">Email Provider</h3>
        <p class="section-desc">Configure how OwnMaily sends emails on your behalf.</p>
      </div>
      <div class="card form-card">
        <div class="field">
          <label>Provider</label>
          <select v-model="provider" @change="testResult = null">
            <option value="resend">Resend</option>
            <option value="mailgun">Mailgun</option>
            <option value="ses">Amazon SES</option>
          </select>
        </div>

        <!-- Resend -->
        <template v-if="provider === 'resend'">
          <div class="field">
            <label>API Key</label>
            <input v-model="creds.api_key" type="password" placeholder="re_••••••••••••••••" class="mono-input" />
          </div>
        </template>

        <!-- Mailgun -->
        <template v-else-if="provider === 'mailgun'">
          <div class="field">
            <label>API Key</label>
            <input v-model="creds.api_key" type="password" placeholder="key-••••••••••••••••" class="mono-input" />
          </div>
          <div class="field">
            <label>Domain</label>
            <input v-model="creds.domain" type="text" placeholder="mg.example.com" />
          </div>
          <div class="field">
            <label>Region</label>
            <select v-model="creds.region">
              <option value="US">US</option>
              <option value="EU">EU</option>
            </select>
          </div>
        </template>

        <!-- SES -->
        <template v-else-if="provider === 'ses'">
          <div class="field">
            <label>Access Key ID</label>
            <input v-model="creds.access_key_id" type="text" placeholder="AKIA••••••••••••••••" class="mono-input" />
          </div>
          <div class="field">
            <label>Secret Access Key</label>
            <input v-model="creds.secret_access_key" type="password" placeholder="••••••••••••••••••••••••••••••••••••••••" class="mono-input" />
          </div>
          <div class="field">
            <label>Region</label>
            <select v-model="creds.region">
              <option value="us-east-1">us-east-1</option>
              <option value="us-west-2">us-west-2</option>
              <option value="eu-west-1">eu-west-1</option>
              <option value="ap-southeast-1">ap-southeast-1</option>
            </select>
          </div>
        </template>
      </div>
    </div>

    <div class="action-row">
      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? 'Saving…' : 'Save configuration' }}
      </button>
      <button class="btn-secondary" :disabled="testing" @click="testConnection">
        {{ testing ? 'Testing…' : 'Test connection' }}
      </button>
      <transition name="fade">
        <span v-if="saved" class="saved-msg">Saved ✓</span>
      </transition>
    </div>

    <div v-if="testResult" class="test-result" :class="testResult.ok ? 'ok' : 'fail'">
      {{ testResult.message }}
    </div>
    <div v-if="error" class="error-msg">{{ error }}</div>
  </div>
</template>

<script>
import { ref, watch, onMounted } from 'vue'
import api from '@/lib/api'

export default {
  name: 'SettingsSMTPView',
  setup() {
    const provider = ref('resend')
    const creds = ref({ api_key: '', domain: '', region: 'US', access_key_id: '', secret_access_key: '' })
    const saving = ref(false)
    const testing = ref(false)
    const saved = ref(false)
    const testResult = ref(null)
    const error = ref('')

    watch(provider, () => {
      creds.value = { api_key: '', domain: '', region: provider.value === 'ses' ? 'us-east-1' : 'US', access_key_id: '', secret_access_key: '' }
    })

    onMounted(async () => {
      try {
        const { data } = await api.get('/api/settings')
        if (data.smtp_provider) provider.value = data.smtp_provider
      } catch {}
    })

    function buildCredentials() {
      if (provider.value === 'resend') return { api_key: creds.value.api_key }
      if (provider.value === 'mailgun') return { api_key: creds.value.api_key, domain: creds.value.domain, region: creds.value.region }
      return { access_key_id: creds.value.access_key_id, secret_access_key: creds.value.secret_access_key, region: creds.value.region }
    }

    async function save() {
      saving.value = true
      error.value = ''
      try {
        await api.put('/api/settings/smtp', { provider: provider.value, credentials: buildCredentials() })
        saved.value = true
        setTimeout(() => { saved.value = false }, 2000)
      } catch (e) {
        error.value = e.response?.data?.message || 'Save failed'
      } finally {
        saving.value = false
      }
    }

    async function testConnection() {
      testing.value = true
      testResult.value = null
      error.value = ''
      try {
        await api.post('/api/settings/smtp/test')
        testResult.value = { ok: true, message: 'Connection successful — test email sent.' }
      } catch (e) {
        testResult.value = { ok: false, message: e.response?.data?.message || 'Connection failed.' }
      } finally {
        testing.value = false
      }
    }

    return { provider, creds, saving, testing, saved, testResult, error, save, testConnection }
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
.form-card { display: flex; flex-direction: column; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field label { font-size: 12px; font-weight: 600; color: #444; letter-spacing: 0.02em; }
.field input, .field select {
  padding: 8px 11px;
  border: 1px solid var(--border);
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  outline: none;
  background: #fff;
  transition: border-color 120ms;
}
.field input:focus, .field select:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16,185,129,0.12);
}
.mono-input { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
.action-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
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
.btn-secondary {
  padding: 8px 16px;
  background: #fff;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
}
.btn-secondary:hover:not(:disabled) { border-color: #aaa; color: var(--text-primary); }
.btn-secondary:disabled { opacity: 0.6; cursor: not-allowed; }
.saved-msg { font-size: 13px; font-weight: 500; color: var(--accent); }
.error-msg { font-size: 13px; color: #dc2626; }
.test-result {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}
.test-result.ok { background: #f0fdf4; border: 1px solid #bbf7d0; color: #166534; }
.test-result.fail { background: #fef2f2; border: 1px solid #fecaca; color: #991b1b; }
.fade-enter-active, .fade-leave-active { transition: opacity 200ms; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
