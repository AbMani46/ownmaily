<template>
  <div class="email-editor" @keydown="handleKeydown">
    <div class="toolbar">
      <button
        class="tb-btn"
        :class="{ active: isBold }"
        title="Bold"
        @mousedown.prevent
        @click="editor?.chain().focus().toggleBold().run()"
      ><strong>B</strong></button>
      <button
        class="tb-btn tb-italic"
        :class="{ active: isItalic }"
        title="Italic"
        @mousedown.prevent
        @click="editor?.chain().focus().toggleItalic().run()"
      ><em>I</em></button>
      <button
        class="tb-btn tb-underline"
        :class="{ active: isUnderline }"
        title="Underline"
        @mousedown.prevent
        @click="editor?.chain().focus().toggleUnderline().run()"
      ><u>U</u></button>

      <div class="toolbar-sep" />

      <button class="tb-pill" :class="{ active: isH1 }" @mousedown.prevent @click="editor?.chain().focus().toggleHeading({ level: 1 }).run()">H1</button>
      <button class="tb-pill" :class="{ active: isH2 }" @mousedown.prevent @click="editor?.chain().focus().toggleHeading({ level: 2 }).run()">H2</button>
      <button class="tb-pill" :class="{ active: isH3 }" @mousedown.prevent @click="editor?.chain().focus().toggleHeading({ level: 3 }).run()">H3</button>

      <div class="toolbar-sep" />

      <button
        class="tb-btn"
        :class="{ active: isLink }"
        title="Link (Cmd+K)"
        @mousedown.prevent
        @click="openLinkModal()"
      >
        <Link2 :size="13" />
      </button>
      <button
        v-if="isLink"
        class="tb-btn"
        title="Remove link"
        @mousedown.prevent
        @click="editor?.chain().focus().unsetLink().run()"
      >
        <Link2Off :size="13" />
      </button>
      <button
        class="tb-btn"
        :class="{ active: isBulletList }"
        title="Bullet list"
        @mousedown.prevent
        @click="editor?.chain().focus().toggleBulletList().run()"
      >
        <List :size="13" />
      </button>
      <button
        class="tb-btn"
        :class="{ active: isOrderedList }"
        title="Ordered list"
        @mousedown.prevent
        @click="editor?.chain().focus().toggleOrderedList().run()"
      >
        <ListOrdered :size="13" />
      </button>

      <div class="toolbar-sep" />

      <div class="vars-wrap" ref="varsWrapRef">
        <button
          class="tb-pill"
          title="Insert variable"
          @mousedown.prevent
          @click="showVarsDropdown = !showVarsDropdown"
        >{ } <ChevronDown :size="10" :stroke-width="2.5" class="vars-caret" /></button>
        <div v-if="showVarsDropdown" class="vars-dropdown">
          <button
            v-for="v in VARIABLES"
            :key="v.id"
            class="vars-item"
            @mousedown.prevent
            @click="insertVariable(v.id)"
          >
            <span class="vars-item-label">{{ v.label }}</span>
            <span class="vars-item-hint">{{ v.hint }}</span>
          </button>
        </div>
      </div>

      <button
        class="tb-pill"
        title="Insert CTA button"
        @mousedown.prevent
        @click="insertCTA()"
      >CTA</button>

      <div class="toolbar-sep" />

      <button
        class="tb-btn"
        :class="{ uploading: isUploading }"
        title="Insert image"
        :disabled="isUploading"
        @mousedown.prevent
        @click="imageInputRef?.click()"
      >
        <Loader v-if="isUploading" :size="13" class="spin" />
        <ImageIcon v-else :size="13" />
      </button>
      <input
        ref="imageInputRef"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        style="display:none"
        @change="onImageFileSelected"
      />

      <div class="toolbar-spacer" />

      <button class="preview-toggle" @click="emit('toggle-preview')">
        <Eye :size="12" />
        {{ previewMode ? 'Edit' : 'Preview' }}
      </button>
    </div>

    <div class="editor-body">
      <EditorContent :editor="editor" />
    </div>

    <BaseModal :show="showLinkModal" title="Insert Link" @close="closeLinkModal">
      <form @submit.prevent="submitLink" class="link-form">
        <div class="form-field">
          <label class="field-label">URL</label>
          <input
            ref="linkInputRef"
            v-model="linkUrl"
            placeholder="https://example.com"
            type="url"
            class="link-input"
          />
        </div>
        <div class="form-actions">
          <BaseButton variant="ghost" type="button" @click="closeLinkModal">Cancel</BaseButton>
          <BaseButton variant="primary" type="submit">Insert Link</BaseButton>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent, VueNodeViewRenderer } from '@tiptap/vue-3'
import { Node, mergeAttributes } from '@tiptap/core'
import { onClickOutside } from '@vueuse/core'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import TiptapImage from '@tiptap/extension-image'
import { Link2, Link2Off, List, ListOrdered, Eye, ChevronDown, Image as ImageIcon, Loader } from 'lucide-vue-next'
import BaseModal from '@/components/BaseModal.vue'
import BaseButton from '@/components/BaseButton.vue'
import VariableChipNodeView from '@/components/VariableChipNode.vue'
import CTAButtonNodeView from '@/components/CTAButtonNode.vue'
import api from '@/lib/api'

const VARIABLES = [
  { id: 'first_name',      label: 'First name',      hint: '{{first_name}}' },
  { id: 'unsubscribe_url', label: 'Unsubscribe URL',  hint: '{{unsubscribe_url}}' },
]

const VariableChip = Node.create({
  name: 'variableChip',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return { id: { default: null } }
  },

  parseHTML() {
    return [
      {
        tag: 'span[data-variable]',
        getAttrs: (dom) => ({ id: (dom as HTMLElement).getAttribute('data-variable') }),
      },
    ]
  },

  renderHTML({ node }) {
    return ['span', mergeAttributes({ 'data-variable': node.attrs.id, class: 'variable-chip' }), `{{${node.attrs.id}}}`]
  },

  addNodeView() {
    return VueNodeViewRenderer(VariableChipNodeView)
  },
})

const CTAButton = Node.create({
  name: 'ctaButton',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return {
      href:  { default: '#' },
      label: { default: 'Learn More →' },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'a[data-cta]',
        getAttrs: (dom) => {
          const el = dom as HTMLElement
          return {
            href:  el.getAttribute('href') ?? '#',
            label: el.textContent?.trim() ?? 'Learn More →',
          }
        },
      },
    ]
  },

  renderHTML({ node }) {
    return [
      'a',
      mergeAttributes({
        'data-cta': '',
        href:        node.attrs.href,
        style:       'background:#10b981;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-weight:600;display:inline-block;',
      }),
      node.attrs.label,
    ]
  },

  addNodeView() {
    return VueNodeViewRenderer(CTAButtonNodeView)
  },
})

const props = defineProps<{
  modelValue: string
  previewMode?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'toggle-preview'): void
}>()

const transactionCount = ref(0)

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit,
    Underline,
    Link.configure({
      openOnClick: false,
      HTMLAttributes: { style: 'color:#10b981;text-decoration:underline;' },
    }),
    Placeholder.configure({ placeholder: 'Start writing your email…' }),
    TiptapImage.configure({
      inline: false,
      allowBase64: false,
      HTMLAttributes: { style: 'max-width:100%;height:auto;display:block;' },
    }),
    VariableChip,
    CTAButton,
  ],
  onUpdate({ editor }) {
    emit('update:modelValue', editor.getHTML())
  },
  onTransaction() {
    transactionCount.value++
  },
})

watch(() => props.modelValue, (val) => {
  if (!editor.value) return
  if (editor.value.getHTML() === val) return
  editor.value.commands.setContent(val, false)
})

onBeforeUnmount(() => editor.value?.destroy())

// Variables dropdown
const showVarsDropdown = ref(false)
const varsWrapRef = ref<HTMLElement | null>(null)

onClickOutside(varsWrapRef, () => { showVarsDropdown.value = false })

function insertVariable(id: string) {
  showVarsDropdown.value = false
  editor.value?.chain().focus().insertContent({ type: 'variableChip', attrs: { id } }).run()
}

function insertCTA() {
  editor.value?.chain().focus().insertContent({
    type: 'ctaButton',
    attrs: { href: '#', label: 'Learn More →' },
  }).run()
}

// Image upload
const imageInputRef = ref<HTMLInputElement | null>(null)
const isUploading = ref(false)

async function onImageFileSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  input.value = '' // reset so the same file can be re-selected
  isUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const res = await api.post('/api/uploads/images', formData)
    editor.value?.chain().focus().setImage({ src: res.data.url }).run()
  } catch (err) {
    console.error('Image upload failed:', err)
  } finally {
    isUploading.value = false
  }
}

// Active states — transactionCount creates reactive dependency on editor state changes
const isBold       = computed(() => { transactionCount.value; return editor.value?.isActive('bold') ?? false })
const isItalic     = computed(() => { transactionCount.value; return editor.value?.isActive('italic') ?? false })
const isUnderline  = computed(() => { transactionCount.value; return editor.value?.isActive('underline') ?? false })
const isH1         = computed(() => { transactionCount.value; return editor.value?.isActive('heading', { level: 1 }) ?? false })
const isH2         = computed(() => { transactionCount.value; return editor.value?.isActive('heading', { level: 2 }) ?? false })
const isH3         = computed(() => { transactionCount.value; return editor.value?.isActive('heading', { level: 3 }) ?? false })
const isBulletList   = computed(() => { transactionCount.value; return editor.value?.isActive('bulletList') ?? false })
const isOrderedList  = computed(() => { transactionCount.value; return editor.value?.isActive('orderedList') ?? false })
const isLink         = computed(() => { transactionCount.value; return editor.value?.isActive('link') ?? false })

// Link modal
const showLinkModal = ref(false)
const linkUrl = ref('')
const linkInputRef = ref<HTMLInputElement | null>(null)

function openLinkModal() {
  linkUrl.value = editor.value?.getAttributes('link').href ?? ''
  showLinkModal.value = true
  nextTick(() => linkInputRef.value?.focus())
}

function closeLinkModal() {
  showLinkModal.value = false
  linkUrl.value = ''
  editor.value?.chain().focus().run()
}

function submitLink() {
  const url = linkUrl.value.trim()
  if (!url) {
    editor.value?.chain().focus().unsetLink().run()
  } else {
    editor.value?.chain().focus().setLink({ href: url }).run()
  }
  closeLinkModal()
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    openLinkModal()
  }
}
</script>

<style scoped>
.email-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 8px 12px;
  border-bottom: 1px solid #f0ede6;
  background: #faf9f7;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.tb-btn {
  width: 28px;
  height: 28px;
  border: 1px solid #e8e5de;
  border-radius: 5px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  color: #555;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 120ms, border-color 120ms, color 120ms;
}

.tb-btn:hover {
  background: #f0ede6;
}

.tb-btn.active {
  background: var(--accent-light);
  border-color: var(--accent);
  color: var(--accent);
}

.tb-italic { font-style: italic; }
.tb-underline { text-decoration: underline; }

.tb-pill {
  padding: 4px 8px;
  border: 1px solid #e8e5de;
  border-radius: 5px;
  background: #fff;
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  color: #555;
  font-family: inherit;
  transition: background 120ms, border-color 120ms, color 120ms;
}

.tb-pill:hover {
  background: #f0ede6;
}

.tb-pill.active {
  background: var(--accent-light);
  border-color: var(--accent);
  color: var(--accent);
}

.toolbar-sep {
  width: 1px;
  height: 20px;
  background: #e8e5de;
  margin: 0 4px;
}

.toolbar-spacer {
  flex: 1;
}

.preview-toggle {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #e8e5de;
  background: #fff;
  color: #555;
  font-size: 11px;
  cursor: pointer;
  font-family: inherit;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: background 120ms;
}

.preview-toggle:hover {
  background: #f0ede6;
}

.editor-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

:deep(.ProseMirror) {
  min-height: 400px;
  outline: none;
  padding: 24px;
  font-size: 14px;
  line-height: 1.7;
  color: #222;
  font-family: 'DM Sans', sans-serif;
}

:deep(.ProseMirror:focus) {
  outline: none;
}

:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: #bbb;
  pointer-events: none;
  float: left;
  height: 0;
}

:deep(.ProseMirror h1) { font-size: 1.8em; font-weight: 700; margin: 0.67em 0; }
:deep(.ProseMirror h2) { font-size: 1.4em; font-weight: 700; margin: 0.75em 0; }
:deep(.ProseMirror h3) { font-size: 1.15em; font-weight: 600; margin: 0.83em 0; }
:deep(.ProseMirror ul) { padding-left: 1.5em; list-style: disc; }
:deep(.ProseMirror ol) { padding-left: 1.5em; list-style: decimal; }
:deep(.ProseMirror li) { margin: 0.25em 0; }
:deep(.ProseMirror a) { color: #10b981; text-decoration: underline; cursor: pointer; }
:deep(.ProseMirror p) { margin: 0.4em 0; }
:deep(.ProseMirror hr) { border: none; border-top: 1px solid #e5e7eb; margin: 1.5em 0; }
:deep(.ProseMirror img) { max-width: 100%; height: auto; display: block; border-radius: 4px; margin: 8px 0; }
:deep(.ProseMirror img.ProseMirror-selectednode) { outline: 2px solid var(--accent); border-radius: 4px; }

.tb-btn.uploading {
  opacity: 0.6;
  cursor: not-allowed;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}

.link-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: #444;
}

.link-input {
  width: 100%;
  padding: 8px 11px;
  border: 1px solid #ddd;
  border-radius: var(--radius-input);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: #fff;
  outline: none;
  box-sizing: border-box;
  transition: border-color 120ms, box-shadow 120ms;
}

.link-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 4px;
}

/* Variables dropdown */
.vars-wrap {
  position: relative;
}

.vars-caret {
  opacity: 0.6;
  margin-left: 2px;
  vertical-align: -2px;
}

.vars-dropdown {
  position: absolute;
  top: calc(100% + 5px);
  left: 0;
  background: #fff;
  border: 1px solid #e8e5de;
  border-radius: 7px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.10);
  z-index: 200;
  min-width: 200px;
  overflow: hidden;
}

.vars-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  border: none;
  background: none;
  font-family: inherit;
  cursor: pointer;
  transition: background 80ms;
}

.vars-item:hover {
  background: #f4f3ef;
}

.vars-item-label {
  font-size: 12px;
  font-weight: 500;
  color: #333;
}

.vars-item-hint {
  font-size: 10px;
  color: #aaa;
  font-family: 'DM Mono', 'Courier New', monospace;
  white-space: nowrap;
}

.vars-item:hover .vars-item-hint {
  color: var(--accent);
}
</style>
