<template>
  <div class="stat-card">
    <div class="stat-header">
      <span class="stat-label">{{ label }}</span>
    </div>
    <div class="stat-value">{{ value }}</div>
    <div v-if="delta || sub" class="stat-footer">
      <span v-if="delta" class="stat-delta" :class="delta.startsWith('+') ? 'pos' : 'neg'">{{ delta }}</span>
      <span v-if="sub" class="stat-sub">{{ sub }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], required: true },
  delta: { type: String, default: '' },
  sub: { type: String, default: '' },
})
</script>

<style scoped>
.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-card);
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.04em;
  line-height: 1.1;
  margin-top: 4px;
}

.stat-footer {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-delta {
  font-weight: 600;
}

.stat-delta.pos { color: #059669; }
.stat-delta.neg { color: #dc2626; }
</style>
