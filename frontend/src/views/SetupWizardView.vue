<template>
  <div class="wizard-shell">
    <div class="wizard-wrap">
      <!-- Logo -->
      <div class="wizard-logo">
        <div class="logo-mark">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/><path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>
        </div>
        <span class="logo-text">OwnMaily Setup</span>
      </div>

      <!-- Step progress bar -->
      <div class="step-track">
        <template v-for="(s, i) in steps" :key="s.id">
          <div class="step-node" :class="{ complete: i < currentStep, current: i === currentStep }">
            <div class="step-circle">
              <svg v-if="i < currentStep" xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
              <span v-else>{{ i + 1 }}</span>
            </div>
            <span class="step-label">{{ s.title }}</span>
          </div>
          <div v-if="i < steps.length - 1" class="step-line" :class="{ done: i < currentStep }"></div>
        </template>
      </div>

      <!-- Card -->
      <div class="wizard-card">
        <div class="card-head">
          <h2>Step {{ currentStep + 1 }}: {{ steps[currentStep].title }}</h2>
          <!-- Progress bar -->
          <div class="progress-track">
            <div class="progress-fill" :style="{ width: ((currentStep + 1) / steps.length * 100) + '%' }"></div>
          </div>
        </div>

        <!-- Step content -->
        <div class="step-body">
          <!-- Step 1: Welcome -->
          <template v-if="currentStep === 0">
            <div class="welcome-body">
              <div class="welcome-icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
              </div>
              <h3>Welcome to OwnMaily</h3>
              <p>You're minutes away from owning your email list. No SaaS, no limits, no monthly fees.</p>
              <div class="checklist">
                <div class="check-item"><span class="check-dot"></span>A domain you control</div>
                <div class="check-item"><span class="check-dot"></span>An SMTP provider API key (Resend, Mailgun, or SES)</div>
                <div class="check-item"><span class="check-dot"></span>About 5 minutes</div>
              </div>
            </div>
          </template>

          <!-- Step 2: Owner Account -->
          <template v-else-if="currentStep === 1">
            <div class="form-fields">
              <div class="field">
                <label>Email address</label>
                <input v-model="owner.email" type="email" placeholder="you@example.com" />
              </div>
              <div class="field">
                <label>Password</label>
                <input v-model="owner.password" type="password" placeholder="At least 8 characters" />
              </div>
              <div class="field">
                <label>Confirm password</label>
                <input v-model="owner.confirm" type="password" placeholder="Repeat password" />
                <span v-if="passwordMismatch" class="field-error">Passwords do not match</span>
              </div>
            </div>
          </template>

          <!-- Step 3: General Settings -->
          <template v-else-if="currentStep === 2">
            <div class="form-fields">
              <div class="field">
                <label>Site name</label>
                <input v-model="general.site_name" type="text" placeholder="My Newsletter" />
              </div>
              <div class="field">
                <label>Installation URL</label>
                <input v-model="general.installation_url" type="url" placeholder="https://mail.example.com" />
              </div>
              <div class="field">
                <label>Timezone</label>
                <select v-model="general.timezone">
                  <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
                </select>
              </div>
              <div class="field">
                <label>Physical address <span class="field-hint-inline">Required for CAN-SPAM</span></label>
                <textarea v-model="general.physical_address" rows="2" placeholder="123 Main St, City, Country"></textarea>
              </div>
            </div>
          </template>

          <!-- Step 4: Connect SMTP -->
          <template v-else-if="currentStep === 3">
            <div class="form-fields">
              <div class="field">
                <label>SMTP provider</label>
                <select v-model="smtp.provider" @change="smtpTested = false">
                  <option value="resend">Resend</option>
                  <option value="mailgun">Mailgun</option>
                  <option value="ses">Amazon SES</option>
                </select>
              </div>

              <template v-if="smtp.provider === 'resend'">
                <div class="field">
                  <label>API Key</label>
                  <input v-model="smtp.api_key" type="password" placeholder="re_••••••••••••••••" class="mono-input" />
                </div>
              </template>
              <template v-else-if="smtp.provider === 'mailgun'">
                <div class="field"><label>API Key</label><input v-model="smtp.api_key" type="password" placeholder="key-••••••••" class="mono-input" /></div>
                <div class="field"><label>Domain</label><input v-model="smtp.domain" type="text" placeholder="mg.example.com" /></div>
                <div class="field"><label>Region</label><select v-model="smtp.region"><option value="US">US</option><option value="EU">EU</option></select></div>
              </template>
              <template v-else-if="smtp.provider === 'ses'">
                <div class="field"><label>Access Key ID</label><input v-model="smtp.access_key_id" type="text" placeholder="AKIA••••••••••••••••" class="mono-input" /></div>
                <div class="field"><label>Secret Access Key</label><input v-model="smtp.secret_access_key" type="password" placeholder="••••••••••••••••••••••••••••••••••••••••" class="mono-input" /></div>
                <div class="field"><label>Region</label><select v-model="smtp.region"><option value="us-east-1">us-east-1</option><option value="us-west-2">us-west-2</option><option value="eu-west-1">eu-west-1</option><option value="ap-southeast-1">ap-southeast-1</option></select></div>
              </template>

              <button class="btn-test" :disabled="testingSmtp" @click="testSmtp">
                {{ testingSmtp ? 'Testing…' : 'Test connection' }}
              </button>

              <div v-if="smtpTestResult" class="test-result" :class="smtpTested ? 'ok' : 'fail'">
                {{ smtpTestResult }}
              </div>
            </div>
          </template>

          <!-- Step 5: Send Test Email -->
          <template v-else-if="currentStep === 4">
            <div class="test-email-body">
              <div class="send-icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
              </div>
              <p class="send-desc">We'll send a test email to confirm everything is working.</p>
              <div class="field" style="width:100%;text-align:left">
                <label>Send test to</label>
                <input v-model="testEmail" type="email" placeholder="you@example.com" />
              </div>
              <button class="btn-send" :disabled="sendingTest" @click="sendTestEmail">
                {{ sendingTest ? 'Sending…' : 'Send test email' }}
              </button>
              <transition name="fade">
                <div v-if="testEmailSent" class="test-success">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                  Check your inbox — test email sent!
                </div>
              </transition>
            </div>
          </template>
        </div>

        <!-- Error -->
        <div v-if="stepError" class="step-error">{{ stepError }}</div>

        <!-- Navigation -->
        <div class="card-footer">
          <button class="btn-ghost" @click="goBack">
            {{ currentStep === 0 ? 'Back to login' : '← Previous' }}
          </button>
          <div class="footer-right">
            <span v-if="currentStep === 3 && !smtpTested" class="continue-hint">
              Test the connection above to continue
            </span>
            <button
              class="btn-primary"
              :disabled="nextDisabled || stepLoading"
              @click="goNext"
            >
              <span v-if="stepLoading" class="loading-dots">…</span>
              <span v-else-if="currentStep === steps.length - 1">Finish Setup</span>
              <span v-else>Continue →</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

export default {
  name: 'SetupWizardView',
  setup() {
    const router = useRouter()
    const auth = useAuthStore()

    const currentStep = ref(0)
    const stepLoading = ref(false)
    const stepError = ref('')

    const steps = [
      { id: 'welcome', title: 'Welcome' },
      { id: 'account', title: 'Owner Account' },
      { id: 'general', title: 'General Settings' },
      { id: 'smtp', title: 'Connect SMTP' },
      { id: 'test', title: 'Send Test Email' },
    ]

    const owner = ref({ email: '', password: '', confirm: '' })
    const general = ref({
      site_name: 'OwnMaily',
      installation_url: '',
      timezone: 'UTC',
      physical_address: '',
    })
    const smtp = ref({ provider: 'resend', api_key: '', domain: '', region: 'US', access_key_id: '', secret_access_key: '' })
    const testEmail = ref('')
    const testEmailSent = ref(false)
    const sendingTest = ref(false)
    const smtpTested = ref(false)
    const testingSmtp = ref(false)
    const smtpTestResult = ref('')

    const timezones = [
      'UTC', 'America/New_York', 'America/Chicago', 'America/Denver',
      'America/Los_Angeles', 'Europe/London', 'Europe/Berlin', 'Europe/Paris',
      'Asia/Tokyo', 'Asia/Shanghai', 'Asia/Singapore', 'Australia/Sydney',
    ]

    onMounted(() => {
      general.value.installation_url = window.location.origin
    })

    const passwordMismatch = computed(() =>
      owner.value.confirm.length > 0 && owner.value.password !== owner.value.confirm
    )

    const nextDisabled = computed(() => {
      if (currentStep.value === 3) return !smtpTested.value
      return false
    })

    function buildSmtpCredentials() {
      const p = smtp.value.provider
      if (p === 'resend') return { api_key: smtp.value.api_key }
      if (p === 'mailgun') return { api_key: smtp.value.api_key, domain: smtp.value.domain, region: smtp.value.region }
      return { access_key_id: smtp.value.access_key_id, secret_access_key: smtp.value.secret_access_key, region: smtp.value.region }
    }

    async function testSmtp() {
      smtpTestResult.value = ''
      smtpTested.value = false
      stepError.value = ''

      if (!owner.value.email) {
        smtpTestResult.value = 'Owner email not set — go back to step 2 and enter your email address before testing.'
        return
      }

      testingSmtp.value = true
      try {
        // Save SMTP first, then send a test to the owner's email
        await api.put('/api/setup/smtp', { provider: smtp.value.provider, credentials: buildSmtpCredentials() })
        await api.post('/api/setup/test-smtp', { to: owner.value.email })
        smtpTested.value = true
        smtpTestResult.value = `Connection successful — test email sent to ${owner.value.email}!`
      } catch (e) {
        smtpTestResult.value = e.response?.data?.message || 'Connection failed. Check your credentials.'
      } finally {
        testingSmtp.value = false
      }
    }

    async function sendTestEmail() {
      sendingTest.value = true
      try {
        await api.post('/api/setup/test-smtp', { to: testEmail.value })
        testEmailSent.value = true
      } catch (e) {
        stepError.value = e.response?.data?.message || 'Failed to send test email'
      } finally {
        sendingTest.value = false
      }
    }

    async function goNext() {
      stepError.value = ''
      stepLoading.value = true
      try {
        if (currentStep.value === 1) {
          if (owner.value.password !== owner.value.confirm) {
            stepError.value = 'Passwords do not match'
            return
          }
          await api.post('/api/setup/owner', { email: owner.value.email, password: owner.value.password })
          // Pre-fill test email
          testEmail.value = owner.value.email
        } else if (currentStep.value === 2) {
          await api.put('/api/setup/settings', {
            site_name: general.value.site_name,
            installation_url: general.value.installation_url,
            timezone: general.value.timezone,
            physical_address: general.value.physical_address,
          })
        } else if (currentStep.value === 4) {
          // Finish setup
          const { data } = await api.post('/api/setup/complete')
          auth.setToken(data.token)
          router.push('/dashboard')
          return
        }
        currentStep.value++
      } catch (e) {
        stepError.value = e.response?.data?.message || 'An error occurred. Please try again.'
      } finally {
        stepLoading.value = false
      }
    }

    function goBack() {
      stepError.value = ''
      if (currentStep.value === 0) {
        router.push('/login')
      } else {
        currentStep.value--
      }
    }

    return {
      currentStep, stepLoading, stepError, steps,
      owner, general, smtp, testEmail, testEmailSent, sendingTest,
      smtpTested, testingSmtp, smtpTestResult,
      timezones, passwordMismatch, nextDisabled,
      goNext, goBack, testSmtp, sendTestEmail,
    }
  },
}
</script>

<style scoped>
.wizard-shell {
  min-height: 100vh;
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  font-family: 'DM Sans', sans-serif;
}

.wizard-wrap {
  width: 100%;
  max-width: 520px;
}

.wizard-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 28px;
}

.logo-mark {
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

/* Step track */
.step-track {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  margin-bottom: 28px;
}

.step-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.step-circle {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e8e5de;
  border: 2px solid #e8e5de;
  color: #aaa;
  font-size: 12px;
  font-weight: 700;
  transition: all 200ms;
}

.step-node.complete .step-circle {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

.step-node.current .step-circle {
  background: #fff;
  border-color: var(--accent);
  color: var(--accent);
}

.step-label {
  font-size: 10px;
  color: #aaa;
  font-weight: 400;
  white-space: nowrap;
}

.step-node.current .step-label { color: var(--accent); font-weight: 600; }
.step-node.complete .step-label { color: var(--accent); }

.step-line {
  flex: 1;
  height: 2px;
  background: #e8e5de;
  margin: 15px 4px 18px;
  transition: background 300ms;
}

.step-line.done { background: var(--accent); }

/* Card */
.wizard-card {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 28px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
}

.card-head {
  margin-bottom: 22px;
}

.card-head h2 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 10px;
  letter-spacing: -0.02em;
}

.progress-track {
  height: 2px;
  background: #f0ede6;
  border-radius: 2px;
}

.progress-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 2px;
  transition: width 300ms ease;
}

/* Welcome step */
.welcome-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 10px 0;
  gap: 12px;
}

.welcome-icon { margin-bottom: 4px; }

.welcome-body h3 {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.02em;
}

.welcome-body p {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
  max-width: 340px;
  margin: 0;
}

.checklist {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: left;
  margin-top: 4px;
}

.check-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
}

.check-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  flex-shrink: 0;
}

/* Form */
.form-fields { display: flex; flex-direction: column; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field label {
  font-size: 12px;
  font-weight: 600;
  color: #444;
  letter-spacing: 0.02em;
  display: flex;
  align-items: center;
  gap: 6px;
}
.field-hint-inline { font-size: 11px; color: var(--text-muted); font-weight: 400; }
.field input, .field select, .field textarea {
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
.field input:focus, .field select:focus, .field textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16,185,129,0.12);
}
.field textarea { resize: vertical; line-height: 1.5; }
.field-error { font-size: 12px; color: #dc2626; }
.mono-input { font-family: 'JetBrains Mono', monospace; font-size: 12px; }

/* SMTP test */
.btn-test {
  padding: 8px 16px;
  background: #fff;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  align-self: flex-start;
  transition: all 120ms;
}
.btn-test:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
.btn-test:disabled { opacity: 0.6; cursor: not-allowed; }
.test-result {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}
.test-result.ok { background: #f0fdf4; border: 1px solid #bbf7d0; color: #166534; }
.test-result.fail { background: #fef2f2; border: 1px solid #fecaca; color: #991b1b; }

/* Test email step */
.test-email-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 14px;
}

.send-icon { }

.send-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
  max-width: 300px;
}

.btn-send {
  padding: 9px 22px;
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
.btn-send:hover:not(:disabled) { background: var(--accent-hover); }
.btn-send:disabled { opacity: 0.6; cursor: not-allowed; }

.test-success {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--accent);
}

/* Step error */
.step-error {
  margin-top: 12px;
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  font-size: 13px;
  color: #991b1b;
}

/* Footer */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
}

.btn-primary {
  padding: 9px 20px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 120ms;
  min-width: 110px;
  text-align: center;
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-ghost {
  padding: 9px 14px;
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid transparent;
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  transition: all 120ms;
}
.btn-ghost:hover { border-color: var(--border); color: var(--text-primary); }

.loading-dots { letter-spacing: 2px; }

.footer-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.continue-hint {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

.fade-enter-active, .fade-leave-active { transition: opacity 200ms; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.step-body { min-height: 180px; }
</style>
