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
 * `DialogForm` — a reusable modal wrapper for a validated form, the form-shaped
 * sibling of `DialogConfirm`. It owns the PrimeVue `<Dialog>` chrome, the
 * `<form>` element, and the Cancel/Submit footer, and exposes the form body as
 * the default slot. Validation and the submit action stay with the consumer;
 * rendered FieldHint errors automatically disable the footer submit button.
 * The supplied async submit handler is awaited so duplicate submissions are
 * blocked; the consumer still listens for `cancel`.
 *
 * Close affordances differ deliberately: the Cancel button resets and closes
 * immediately, while a click outside on the mask first shows the shared
 * "discard unsaved changes?" confirm (the same one used by `UnchangedGuard`)
 * when the caller reports the form is `dirty`. When clean, every path closes
 * silently. The ✕ button, Cancel, and an outside click all route through the
 * same close/discard logic.
 */
import Dialog from 'primevue/dialog'
import { computed, ref, useId } from 'vue'
import { useI18n } from 'vue-i18n'

import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormSubmitLock } from '@/composables/useFormSubmitLock'
import { provideFormValidation } from '@/composables/useFormValidation'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    discard: 'Discard',
    keepEditing: 'Keep editing',
    unsavedTitle: 'Discard unsaved changes?',
    unsavedMessage: 'You have unsaved changes. If you leave this page, they will be discarded.',
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** v-model:visible */
    visible: boolean
    title: string
    /** Optional PrimeIcon class (e.g. `pi pi-comment`); renders a centered icon+title header matching DialogConfirm. */
    icon?: string
    /** Label for the primary submit action. */
    textSubmitLabel: string
    /** Stable test hook forwarded to the submit button. */
    submitTestId?: string
    /** Optional non-validation reason to disable submit; visible errors are automatic. */
    submitDisabled?: boolean
    /** Submit action awaited so pending state and duplicate attempts stay centralized. */
    submitHandler: () => Promise<void> | void
    submitSeverity?: 'danger' | 'warn' | 'primary' | 'secondary'
    /**
     * Whether the form has unsaved input. When true, closing via an outside
     * click asks for confirmation first; the Cancel button always closes directly.
     */
    dirty?: boolean
    /** Dialog width, e.g. '30rem'. */
    width?: string
    /** Extra class applied to the teleported dialog root. */
    dialogClass?: string
  }>(),
  {
    icon: undefined,
    submitDisabled: false,
    submitSeverity: undefined,
    dirty: false,
    width: '30rem',
    dialogClass: undefined,
  },
)

const emit = defineEmits<{
  'update:visible': [value: boolean]
  cancel: []
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

const { isSubmitting, run: runSubmit } = useFormSubmitLock()
const { submitDisabled: effectiveSubmitDisabled, validate } = provideFormValidation({
  submitDisabled: () => props.submitDisabled,
  submitLoading: isSubmitting,
})
const formId = useId()

// ── Reactive state ────────────────────────────────────────────
// Whether the "discard unsaved changes?" confirm is currently open
const discardPromptVisible = ref(false)

// ── Computed ──────────────────────────────────────────────────
// Dialog width as an inline style object
const dialogStyle = computed(() => ({ width: props.width }))
const dialogHeader = computed(() => (props.icon ? undefined : props.title))

// ── Functions ─────────────────────────────────────────────────
// Close the dialog by notifying the parent via the v-model event
function closeDialog(): void {
  emit('update:visible', false)
}

// Ignore native submits while disabled; otherwise hold the lock until the
// caller's complete async submit flow has finished.
async function submitForm(): Promise<void> {
  if (effectiveSubmitDisabled.value) return
  if (!(await validate())) return
  await runSubmit(props.submitHandler)
}

// Cancel button: always resets and closes directly, no discard confirm
function cancelDialog(): void {
  emit('cancel')
  closeDialog()
}

// Outside click on the mask: a dirty form must confirm the discard first, a
// clean one closes straight away like Cancel
function requestClose(): void {
  if (props.dirty) {
    discardPromptVisible.value = true
    return
  }
  cancelDialog()
}

// Treat every user-driven PrimeVue close path as a request to leave the form
function onVisibilityChange(nextVisible: boolean): void {
  if (!nextVisible) requestClose()
}

// Discard confirmed: treat as a cancel (the caller's cancel handler resets the form)
function discardChangesAndClose(): void {
  cancelDialog()
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    dismissable-mask
    :header="dialogHeader"
    closable
    :draggable="false"
    :style="dialogStyle"
    :class="['app-dialog', dialogClass]"
    @update:visible="onVisibilityChange"
  >
    <template v-if="icon" #header>
      <div class="app-dialog__header">
        <i :class="icon" class="app-dialog__header-icon" aria-hidden="true" />
        <span class="app-dialog__header-title">{{ title }}</span>
      </div>
    </template>
    <form :id="formId" novalidate class="dialog-form__form" @submit.prevent="submitForm">
      <slot />
    </form>

    <template #footer>
      <LayoutSubmitFooter
        submit-type="submit"
        :form="formId"
        :text-submit-label="textSubmitLabel"
        :submit-test-id="submitTestId"
        :submit-severity="submitSeverity"
        @cancel="cancelDialog"
      />
    </template>
  </Dialog>

  <!-- Discard-changes confirm, shown only when closing a dirty form via an
       outside click (the same warning UnchangedGuard uses on navigation) -->
  <DialogConfirm
    v-model:visible="discardPromptVisible"
    :title="t('unsavedTitle')"
    :text-cancel-label="t('keepEditing')"
    :text-confirm-label="t('discard')"
    @confirm="discardChangesAndClose"
  >
    {{ t('unsavedMessage') }}
  </DialogConfirm>
</template>
