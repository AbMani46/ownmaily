<template>
  <div class="base-card" :class="{ 'no-padding': noPadding }">
    <slot />
  </div>
</template>

<script setup>
defineProps({
  noPadding: { type: Boolean, default: false },
})
</script>

<style scoped>
.base-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  overflow: hidden;
}

.base-card:not(.no-padding) {
  padding: 20px 22px;
}
</style>
