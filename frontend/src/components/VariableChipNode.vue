<template>
  <NodeViewWrapper as="span" class="var-chip-outer">
    <span class="var-chip" :class="{ 'var-chip--sel': selected }">{{ display }}</span>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { NodeViewWrapper } from '@tiptap/vue-3'
import { computed } from 'vue'

const props = defineProps<{
  node: { attrs: { id: string } }
  selected: boolean
}>()

const display = computed(() => `{{ ${props.node.attrs.id} }}`)
</script>

<style scoped>
.var-chip-outer {
  display: inline;
}

.var-chip {
  display: inline-flex;
  align-items: center;
  padding: 1px 8px;
  border-radius: 20px;
  background: var(--accent-light);
  border: 1px solid var(--accent);
  color: var(--accent);
  font-size: 11px;
  font-weight: 500;
  font-family: 'DM Mono', 'Fira Code', 'Courier New', monospace;
  line-height: 1.8;
  cursor: default;
  white-space: nowrap;
  user-select: none;
  vertical-align: middle;
  transition: background 100ms, color 100ms;
}

.var-chip--sel {
  background: var(--accent);
  color: #fff;
}
</style>
