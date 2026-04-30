<template>
  <div class="input-wrap" :class="{ focused }">
    <input
      class="input"
      :type="type"
      :placeholder="placeholder"
      :value="modelValue"
      :disabled="disabled"
      @input="$emit('update:modelValue', $event.target.value)"
      @focus="focused = true"
      @blur="focused = false"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  type: { type: String, default: 'text' },
  disabled: { type: Boolean, default: false },
})
defineEmits(['update:modelValue'])

const focused = ref(false)
</script>

<style scoped>
.input-wrap {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  background: #fff;
  transition: border-color 120ms, box-shadow 120ms;
  overflow: hidden;
}

.input-wrap.focused {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.input {
  flex: 1;
  padding: 8px 11px;
  border: none;
  outline: none;
  font-size: 13px;
  font-family: inherit;
  background: transparent;
  color: var(--text-primary);
  min-width: 0;
  width: 100%;
}

.input:disabled {
  background: #f9f9f7;
}
</style>
