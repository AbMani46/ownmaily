<template>
  <button
    class="btn"
    :class="`btn-${variant}`"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <svg v-if="loading" class="spinner" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
      <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
    </svg>
    <slot v-if="!loading" />
    <span v-if="loading">Loading…</span>
  </button>
</template>

<script setup>
defineProps({
  variant: { type: String, default: 'primary' },
  loading: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
})
defineEmits(['click'])
</script>

<style scoped>
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 7px 14px;
  border-radius: var(--radius-btn);
  border: none;
  cursor: pointer;
  transition: all 120ms ease;
  white-space: nowrap;
  outline: none;
  line-height: 1;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
}
.btn-primary:hover:not(:disabled) {
  background: var(--accent-hover);
}

.btn-secondary {
  background: #eae7e0;
  color: #1a1a1a;
  border: 1px solid #d6d2c8;
}
.btn-secondary:hover:not(:disabled) {
  background: #f0ede6;
}

.btn-ghost {
  background: transparent;
  color: #4a4a4a;
  border: 1px solid transparent;
}
.btn-ghost:hover:not(:disabled) {
  background: #f0ede6;
}

.btn-danger {
  background: #dc2626;
  color: #fff;
}
.btn-danger:hover:not(:disabled) {
  background: #b91c1c;
}

.spinner {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
