<template>
  <NodeViewWrapper as="span" class="cta-outer">
    <span
      class="cta-preview"
      :class="{ 'cta-preview--sel': selected }"
      @click="openModal"
    >
      <span class="cta-preview-label">{{ node.attrs.label }}</span>
      <Pencil :size="10" class="cta-preview-edit" />
    </span>
    <Teleport to="body">
      <div v-if="showModal" class="cta-backdrop" @mousedown.self="closeModal">
        <div class="cta-dialog" role="dialog" aria-modal="true">
          <div class="cta-dialog-header">
            <span class="cta-dialog-title">Edit Button</span>
            <button class="cta-dialog-x" type="button" @click="closeModal" aria-label="Close">×</button>
          </div>
          <form @submit.prevent="submitModal" class="cta-dialog-form">
            <div class="cta-dialog-field">
              <label class="cta-dialog-label">Button text</label>
              <input
                ref="labelInputRef"
                v-model="form.label"
                type="text"
                class="cta-dialog-input"
                placeholder="Learn More →"
              />
            </div>
            <div class="cta-dialog-field">
              <label class="cta-dialog-label">Link URL</label>
              <input
                v-model="form.href"
                type="url"
                class="cta-dialog-input"
                placeholder="https://example.com"
              />
            </div>
            <div class="cta-dialog-actions">
              <button type="button" class="cta-dialog-cancel" @click="closeModal">Cancel</button>
              <button type="submit" class="cta-dialog-submit">Save</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </NodeViewWrapper>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick } from 'vue'
import { NodeViewWrapper } from '@tiptap/vue-3'
import { Pencil } from 'lucide-vue-next'

const props = defineProps<{
  node: { attrs: { href: string; label: string } }
  selected: boolean
  updateAttributes: (attrs: Record<string, unknown>) => void
}>()

const showModal = ref(false)
const labelInputRef = ref<HTMLInputElement | null>(null)
const form = reactive({ href: '', label: '' })

function openModal() {
  form.href  = props.node.attrs.href
  form.label = props.node.attrs.label
  showModal.value = true
  nextTick(() => labelInputRef.value?.focus())
}

function closeModal() {
  showModal.value = false
}

function submitModal() {
  props.updateAttributes({
    href:  form.href.trim()  || '#',
    label: form.label.trim() || 'Learn More →',
  })
  closeModal()
}
</script>

<style scoped>
.cta-outer {
  display: inline;
}

.cta-preview {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #10b981;
  color: #fff;
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  user-select: none;
  vertical-align: middle;
  transition: filter 120ms;
}

.cta-preview:hover {
  filter: brightness(0.91);
}

.cta-preview--sel {
  outline: 2px solid #059669;
  outline-offset: 2px;
}

.cta-preview-edit {
  opacity: 0;
  flex-shrink: 0;
  transition: opacity 120ms;
}

.cta-preview:hover .cta-preview-edit {
  opacity: 0.75;
}

/* Teleport modal — scoped attrs still apply to teleported nodes in Vue 3 */
.cta-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.42);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.cta-dialog {
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.18);
  width: 340px;
  padding: 20px;
}

.cta-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.cta-dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: #222;
}

.cta-dialog-x {
  background: none;
  border: none;
  font-size: 20px;
  color: #bbb;
  cursor: pointer;
  line-height: 1;
  padding: 0 2px;
  transition: color 100ms;
}

.cta-dialog-x:hover {
  color: #555;
}

.cta-dialog-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.cta-dialog-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.cta-dialog-label {
  font-size: 12px;
  font-weight: 600;
  color: #444;
}

.cta-dialog-input {
  width: 100%;
  padding: 8px 11px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 13px;
  font-family: inherit;
  color: #222;
  background: #fff;
  outline: none;
  box-sizing: border-box;
  transition: border-color 120ms, box-shadow 120ms;
}

.cta-dialog-input:focus {
  border-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.cta-dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 4px;
}

.cta-dialog-cancel {
  padding: 7px 14px;
  border: 1px solid #e0ddd6;
  border-radius: 6px;
  background: #fff;
  font-size: 13px;
  font-family: inherit;
  color: #555;
  cursor: pointer;
  transition: background 120ms;
}

.cta-dialog-cancel:hover {
  background: #f5f4f0;
}

.cta-dialog-submit {
  padding: 7px 16px;
  border: none;
  border-radius: 6px;
  background: #10b981;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: filter 120ms;
}

.cta-dialog-submit:hover {
  filter: brightness(0.91);
}
</style>
