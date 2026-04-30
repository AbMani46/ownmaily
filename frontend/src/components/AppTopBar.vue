<template>
  <header class="topbar">
    <div class="topbar-left">
      <h1 class="page-title">{{ title }}</h1>
    </div>
    <div class="topbar-right">
      <slot />
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const title = computed(() => {
  return route.meta?.title || formatRouteName(route.name)
})

function formatRouteName(name) {
  if (!name) return 'OwnMaily'
  return String(name).replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}
</script>

<style scoped>
.topbar {
  height: var(--topbar-height);
  background: var(--bg-card);
  border-bottom: 1px solid #ece9e1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
}

.page-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
