<template>
  <div class="pagination" v-if="total > perPage">
    <span class="info">Showing {{ from }}–{{ to }} of {{ total }}</span>
    <button class="pg-btn" :disabled="page <= 1" @click="$emit('update:page', page - 1)">←</button>
    <button class="pg-btn" :disabled="page >= pages" @click="$emit('update:page', page + 1)">→</button>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  page: { type: Number, required: true },
  perPage: { type: Number, required: true },
  total: { type: Number, required: true },
})
defineEmits(['update:page'])

const pages = computed(() => Math.ceil(props.total / props.perPage))
const from = computed(() => (props.page - 1) * props.perPage + 1)
const to = computed(() => Math.min(props.page * props.perPage, props.total))
</script>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 14px;
}

.info {
  font-size: 12px;
  color: var(--text-muted);
}

.pg-btn {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background: #fff;
  cursor: pointer;
  font-size: 12px;
  font-family: inherit;
  transition: opacity 120ms;
}

.pg-btn:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.pg-btn:not(:disabled):hover {
  background: #f4f3ef;
}
</style>
