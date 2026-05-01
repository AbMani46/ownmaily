<template>
  <div class="settings-page">
    <!-- Sidebar nav -->
    <nav class="settings-sidebar">
      <router-link
        v-for="item in navItems"
        :key="item.section"
        :to="'/settings/' + item.section"
        class="nav-item"
        :class="{ active: currentSection === item.section }"
      >
        <component :is="item.icon" class="nav-icon" />
        {{ item.label }}
      </router-link>
    </nav>

    <!-- Content area -->
    <div class="settings-content">
      <component :is="currentComponent" v-if="currentComponent" />
      <div v-else class="not-found">Section not found.</div>
    </div>
  </div>
</template>

<script>
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'

const GlobeIcon = { template: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>' }
const ServerIcon = { template: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>' }
const KeyIcon = { template: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>' }
const ShieldIcon = { template: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>' }

export default {
  name: 'SettingsView',
  setup() {
    const route = useRoute()

    const currentSection = computed(() => route.params.section || 'general')

    const navItems = [
      { section: 'general', label: 'General', icon: GlobeIcon },
      { section: 'smtp', label: 'SMTP', icon: ServerIcon },
      { section: 'api-key', label: 'API Key', icon: KeyIcon },
      { section: 'suppression', label: 'Suppression List', icon: ShieldIcon },
    ]

    const sectionMap = {
      general: defineAsyncComponent(() => import('./settings/SettingsGeneralView.vue')),
      smtp: defineAsyncComponent(() => import('./settings/SettingsSMTPView.vue')),
      'api-key': defineAsyncComponent(() => import('./settings/SettingsAPIKeyView.vue')),
      suppression: defineAsyncComponent(() => import('./settings/SettingsSuppressionView.vue')),
    }

    const currentComponent = computed(() => sectionMap[currentSection.value] || null)

    return { navItems, currentSection, currentComponent }
  },
}
</script>

<style scoped>
.settings-page {
  display: flex;
  padding: 24px 28px;
  gap: 32px;
  min-height: 100%;
}

.settings-sidebar {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 200px;
  flex-shrink: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 9px 12px;
  border-radius: 7px;
  border-left: 2px solid transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 400;
  text-decoration: none;
  transition: all 100ms;
  cursor: pointer;
}

.nav-item:hover {
  background: var(--bg-page);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-light);
  border-left-color: var(--accent);
  color: var(--accent);
  font-weight: 600;
  padding-left: 10px;
}

.nav-item.active .nav-icon {
  color: var(--accent);
}

.nav-icon {
  flex-shrink: 0;
  color: #aaa;
}

.nav-item.active .nav-icon {
  color: var(--accent);
}

.settings-content {
  flex: 1;
  min-width: 0;
}

.not-found {
  color: var(--text-muted);
  font-size: 13px;
}
</style>
