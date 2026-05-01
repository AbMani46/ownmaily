<template>
  <span class="badge" :style="{ background: style.bg, color: style.color }">
    <span class="dot" :style="{ background: style.color }" />
    {{ displayLabel }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const BADGE_STYLES = {
  active:       { bg: '#d1fae5', color: '#065f46', label: 'Active' },
  sent:         { bg: '#d1fae5', color: '#065f46', label: 'Sent' },
  draft:        { bg: '#fef3c7', color: '#92400e', label: 'Draft' },
  scheduled:    { bg: '#dbeafe', color: '#1e40af', label: 'Scheduled' },
  pending:      { bg: '#fef3c7', color: '#92400e', label: 'Pending' },
  bounced:      { bg: '#fee2e2', color: '#991b1b', label: 'Bounced' },
  failed:       { bg: '#fee2e2', color: '#991b1b', label: 'Failed' },
  unsubscribed: { bg: '#f3f4f6', color: '#4b5563', label: 'Unsubscribed' },
  double_optin: { bg: '#ede9fe', color: '#5b21b6', label: 'Double Opt-in' },
  single_optin: { bg: '#e0f2fe', color: '#075985', label: 'Single Opt-in' },
  suppressed:   { bg: '#fee2e2', color: '#991b1b', label: 'Suppressed' },
}

const props = defineProps({
  status: { type: String, required: true },
  label: { type: String, default: '' },
})

const style = computed(() => BADGE_STYLES[props.status] || { bg: '#f3f4f6', color: '#4b5563', label: props.status })
const displayLabel = computed(() => props.label || style.value.label)
</script>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.03em;
  padding: 2px 8px;
  border-radius: var(--radius-badge);
  white-space: nowrap;
  text-transform: uppercase;
}

.dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
}
</style>
