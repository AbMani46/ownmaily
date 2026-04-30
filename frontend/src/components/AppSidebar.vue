<template>
  <aside class="sidebar">
    <div class="logo-block">
      <div class="logo-icon">
        <Mail :size="18" :stroke-width="2" />
      </div>
      <span class="logo-text">OwnMaily</span>
    </div>

    <nav class="nav">
      <RouterLink
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <component :is="item.icon" class="nav-icon" :size="20" :stroke-width="1.75" />
        <span class="nav-label">{{ item.label }}</span>
      </RouterLink>
    </nav>

    <div class="sidebar-footer">
      <button class="logout-btn" @click="handleLogout">
        <LogOut class="nav-icon" :size="20" :stroke-width="1.75" />
        <span>Log out</span>
      </button>
    </div>
  </aside>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Mail, LayoutGrid, Users, List, Tag, Activity, BarChart2, Settings, LogOut } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

function isActive(path) {
  return route.path.startsWith(path)
}

async function handleLogout() {
  auth.logout()
  router.push('/login')
}

const navItems = [
  { path: '/dashboard',   label: 'Dashboard',   icon: LayoutGrid },
  { path: '/subscribers', label: 'Subscribers', icon: Users },
  { path: '/lists',       label: 'Lists',        icon: List },
  { path: '/tags',        label: 'Tags',         icon: Tag },
  { path: '/campaigns',   label: 'Campaigns',    icon: Activity },
  { path: '/analytics',   label: 'Analytics',    icon: BarChart2 },
  { path: '/settings',    label: 'Settings',     icon: Settings },
]
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  min-height: 100vh;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-dark);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.logo-block {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 16px 16px;
  border-bottom: 1px solid var(--border-dark);
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: var(--accent);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.logo-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-inverse);
  letter-spacing: -0.01em;
}

.nav {
  flex: 1;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: var(--radius-btn);
  text-decoration: none;
  transition: background 0.12s;
  border-left: 2px solid transparent;
  margin-left: -2px;
}

.nav-item .nav-icon {
  color: #aaaaaa;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.nav-item .nav-label {
  font-size: 13px;
  color: #888888;
  font-weight: 400;
}

.nav-item:hover {
  background: #161618;
}

.nav-item.active {
  background: #1e1e26;
  border-left-color: var(--accent);
}

.nav-item.active .nav-icon {
  color: var(--accent);
}

.nav-item.active .nav-label {
  color: var(--text-inverse);
  font-weight: 500;
}

.sidebar-footer {
  padding: 12px 8px;
  border-top: 1px solid var(--border-dark);
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: var(--radius-btn);
  background: none;
  border: none;
  cursor: pointer;
  width: 100%;
  font-size: 13px;
  color: #888888;
  font-family: inherit;
  transition: background 0.12s;
}

.logout-btn .nav-icon {
  color: #aaaaaa;
  display: flex;
  align-items: center;
}

.logout-btn:hover {
  background: #161618;
  color: #cccccc;
}
</style>
