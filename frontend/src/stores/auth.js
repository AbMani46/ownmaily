import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/lib/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('om_token') || null)
  const isAuthenticated = computed(() => !!token.value)

  async function login(email, password) {
    const res = await api.post('/api/auth/login', { email, password })
    token.value = res.data.token
    localStorage.setItem('om_token', res.data.token)
  }

  function logout() {
    token.value = null
    localStorage.removeItem('om_token')
    api.post('/api/auth/logout').catch(() => {})
  }

  function setToken(t) {
    token.value = t
    localStorage.setItem('om_token', t)
  }

  return { token, isAuthenticated, login, logout, setToken }
})
