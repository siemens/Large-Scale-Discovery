<script setup lang="ts">
/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2025-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

/**
 * <DialogConfirm> — CSP-clean confirmation dialog.
 *
 * Replaces the legacy `confirmOverlay(icon, title, htmlMessage, onConfirm,
 * wordToConfirm?, onCancel?)`. Two modes:
 *   - plain confirm (Users / Databases delete);
 *   - word-confirm: when `confirmWord` is set, an input must match the word
 *     before the confirm button enables (ports the Groups "delete by name"
 *     gate).
 *
 * CSP note: the legacy helper injected raw HTML (`htmlMessage`). We instead
 * render the message via the default slot (typed segments) — no `v-html`,
 * nothing that needs `unsafe-inline`/`unsafe-eval`. All chrome strings route
 * through `t()`.
 */
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    confirmInputAriaLabel: 'Type "{word}" to confirm.',
  },
}

/** MessageSegment is a safe plain-text fragment with optional emphasis. */
interface MessageSegment {
  text: string
  emphasized: boolean
}

// ── Props and emits ───────────────────────────────────────────
interface DialogConfirmProps {
  /** v-model:visible */
  visible: boolean
  title: string
  icon?: string
  /** Primary message shown in the dialog body. */
  message?: string
  /** Secondary message shown below the primary message. */
  secondaryMessage?: string
  /** Label for the secondary cancel action. Defaults to "Cancel". */
  textCancelLabel?: string
  /** Label for the primary confirmation action. */
  textConfirmLabel: string
  /** Word the user must type to unlock the confirm button. When set, a text input appears. */
  confirmWord?: string
  severity?: 'danger' | 'warn' | 'primary' | 'secondary'
}

const props = withDefaults(defineProps<DialogConfirmProps>(), {
  icon: 'pi pi-exclamation-triangle',
  message: undefined,
  secondaryMessage: undefined,
  severity: 'danger',
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: []
  cancel: []
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Reactive state ────────────────────────────────────────────
// Word-confirm gate: typed value must match confirmWord before confirm enables.
const confirmationInput = ref('')
const inputTouched = ref(false)

// ── Computed ──────────────────────────────────────────────────
const requiresWord = computed(
  () => typeof props.confirmWord === 'string' && props.confirmWord.length > 0,
)
const confirmEnabled = computed(
  () => !requiresWord.value || confirmationInput.value === props.confirmWord,
)

const inputInvalid = computed(
  () => requiresWord.value && inputTouched.value && confirmationInput.value !== props.confirmWord,
)

const messageSegments = computed(() => splitQuotedText(props.message))
const secondaryMessageSegments = computed(() => splitQuotedText(props.secondaryMessage))

// ── Functions ─────────────────────────────────────────────────
// Split quoted values into safe text nodes rather than rendering backend-authored HTML
function splitQuotedText(message: string | undefined): MessageSegment[] {
  if (!message) return []

  const segments: MessageSegment[] = []
  const quotedTextPattern = /'([^']+)'/g
  let cursor = 0
  let match: RegExpExecArray | null

  while ((match = quotedTextPattern.exec(message)) !== null) {
    const quotedValue = match[1]
    if (!quotedValue) continue

    // Keep the quote characters visible while emphasizing only their contents
    segments.push({ text: message.slice(cursor, match.index + 1), emphasized: false })
    segments.push({ text: quotedValue, emphasized: true })
    cursor = match.index + match[0].length - 1
  }

  segments.push({ text: message.slice(cursor), emphasized: false })
  return segments
}

function close(): void {
  emit('update:visible', false)
}

function onConfirm(): void {
  if (!confirmEnabled.value) return
  emit('confirm')
  close()
}

// Emit cancel then close; mirrors onConfirm's close call so callers always see both events.
function onCancel(): void {
  emit('cancel')
  close()
}

// Treat every user-driven dialog close path as cancellation
function onVisibilityChange(nextVisible: boolean): void {
  if (!nextVisible) onCancel()
}

// ── Watchers and lifecycle ────────────────────────────────────
watch(
  () => props.visible,
  (open) => {
    if (open) {
      confirmationInput.value = ''
      inputTouched.value = false
    }
  },
)
</script>

<template>
  <Dialog
    data-testid="t-dialog-confirm"
    :visible="visible"
    modal
    dismissable-mask
    closable
    :draggable="false"
    class="app-dialog dialog-confirm"
    @update:visible="onVisibilityChange"
  >
    <template #header>
      <div class="app-dialog__header">
        <i :class="icon" class="app-dialog__header-icon" aria-hidden="true" />
        <span class="app-dialog__header-title">{{ title }}</span>
      </div>
    </template>

    <div class="dialog-confirm__body">
      <slot v-if="$slots.default" />
      <template v-else>
        <p v-if="message">
          <template v-for="(segment, index) in messageSegments" :key="index">
            <strong
              v-if="segment.emphasized"
              class="dialog-confirm__emphasis"
              :class="`dialog-confirm__emphasis--${severity}`"
            >
              {{ segment.text }}
            </strong>
            <template v-else>{{ segment.text }}</template>
          </template>
        </p>
        <p v-if="secondaryMessage">
          <template v-for="(segment, index) in secondaryMessageSegments" :key="index">
            <strong
              v-if="segment.emphasized"
              class="dialog-confirm__emphasis"
              :class="`dialog-confirm__emphasis--${severity}`"
            >
              {{ segment.text }}
            </strong>
            <template v-else>{{ segment.text }}</template>
          </template>
        </p>
      </template>

      <div v-if="requiresWord" class="dialog-confirm__word">
        <InputText
          v-model="confirmationInput"
          data-testid="tcd-confirm-input"
          fluid
          autocomplete="off"
          :invalid="inputInvalid"
          :placeholder="confirmWord"
          :aria-label="t('confirmInputAriaLabel', { word: confirmWord })"
          @blur="inputTouched = true"
          @keyup.enter="onConfirm"
        />
      </div>
    </div>

    <template #footer>
      <LayoutSubmitFooter
        cancel-test-id="tcd-cancel"
        submit-test-id="tcd-confirm"
        submit-type="button"
        :segmented-submit="false"
        :text-cancel-label="props.textCancelLabel"
        :text-submit-label="textConfirmLabel"
        :submit-severity="severity"
        :submit-disabled="!confirmEnabled"
        @cancel="onCancel"
        @submit="onConfirm"
      />
    </template>
  </Dialog>
</template>

<style scoped>
/* Keep confirmation dialogs readable on large screens without overflowing narrow viewports */
.dialog-confirm {
  width: 38rem;
  max-width: calc(100vw - 2rem);
}

/* Center the confirmation copy independently from optional form controls */
.dialog-confirm__body {
  text-align: center;
}

.dialog-confirm__body p {
  margin: 0 0 0.5rem;
}

.dialog-confirm__body p:last-child {
  margin-bottom: 0;
}

/* Match emphasized quoted values to the confirm button's themed background
   token. Keeping the four severity mappings here covers every caller while
   allowing the button palette to change without duplicating color values. */
.dialog-confirm__emphasis--danger {
  color: var(--p-button-danger-background);
}

.dialog-confirm__emphasis--warn {
  color: var(--p-button-warn-background);
}

.dialog-confirm__emphasis--primary {
  color: var(--p-button-primary-background);
}

.dialog-confirm__emphasis--secondary {
  color: var(--p-button-secondary-background);
}

.dialog-confirm__word {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  text-align: left;
  width: 40%;
  margin-left: auto;
  margin-right: auto;
}
</style>
