<template>
  <div class="login-page">
    <div class="login-card">
      <div class="card-header">
        <div class="logo-icon">
          <Mail :size="22" :stroke-width="2" />
        </div>
        <h1 class="app-name">OwnMaily</h1>
        <p class="app-tagline">Sign in to your account</p>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div v-if="error" class="error-banner">
          {{ error }}
        </div>

        <div class="field">
          <label for="email">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            autocomplete="email"
            required
          />
        </div>

        <div class="field">
          <label for="password">Password</label>
          <div class="password-wrap">
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="••••••••"
              autocomplete="current-password"
              required
            />
            <button type="button" class="toggle-pw" @click="showPassword = !showPassword" tabindex="-1">
              <Eye v-if="!showPassword" :size="16" :stroke-width="2" />
              <EyeOff v-else :size="16" :stroke-width="2" />
            </button>
          </div>
        </div>

        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>

      <div v-if="showSetupLink" class="card-footer">
        <RouterLink to="/setup" class="setup-link">Run setup wizard →</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Mail, Eye, EyeOff } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')
// Default to showing the link; hide it once we confirm setup is complete
const showSetupLink = ref(true)

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push('/dashboard')
  } catch {
    error.value = 'Invalid email or password.'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const res = await api.get('/api/setup/status')
    showSetupLink.value = !res.data.complete
  } catch {
    // If the endpoint doesn't exist or errors, keep the link visible
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-card {
  width: 380px;
  background: var(--bg-card);
  border-radius: 12px;
  padding: 28px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px;
  text-align: center;
}

.logo-icon {
  width: 44px;
  height: 44px;
  background: var(--accent);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  margin-bottom: 12px;
}

.app-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 4px;
}

.app-tagline {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.error-banner {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  border-radius: var(--radius-input);
  padding: 10px 12px;
  font-size: 13px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.field label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.field input {
  height: 36px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--bg-card);
  outline: none;
  transition: border-color 0.12s, box-shadow 0.12s;
  width: 100%;
}

.field input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.password-wrap {
  position: relative;
}

.password-wrap input {
  padding-right: 36px;
}

.toggle-pw {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  padding: 0;
  line-height: 1;
}

.toggle-pw:hover {
  color: var(--text-secondary);
}

.submit-btn {
  height: 38px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius-btn);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.12s;
  margin-top: 2px;
}

.submit-btn:hover:not(:disabled) {
  background: var(--accent-hover);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.submit-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
  border: 1px solid var(--accent);
}

.card-footer {
  margin-top: 20px;
  text-align: center;
}

.setup-link {
  font-size: 12px;
  color: var(--accent);
  text-decoration: none;
}

.setup-link:hover {
  color: var(--accent-hover);
  text-decoration: underline;
}
</style>
