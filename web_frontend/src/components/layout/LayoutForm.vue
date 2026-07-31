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

/** Provides an inline form with validation, duplicate-submit protection, and optional actions. */
import SectionHeader from '@/components/elements/SectionHeader.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormSubmitLock } from '@/composables/useFormSubmitLock'
import { provideFormValidation } from '@/composables/useFormValidation'

// ── Props and emits ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** Section header text (callers pass their static or create/edit title). */
    title?: string
    /** Whether to render the standard Cancel/Save footer. Off = form owns it. */
    footer?: boolean
    /** Label for the secondary cancel action. */
    textCancelLabel?: string
    /** Text displayed between cancel and submit actions. */
    textOr?: string
    /** Label for the primary submit action. */
    textSubmitLabel: string
    /** Stable test hook forwarded to the standard submit button. */
    submitTestId?: string
    /** Optional non-validation reason to disable submit; visible errors are automatic. */
    submitDisabled?: boolean
    /** Submit action awaited so pending state and duplicate attempts stay centralized. */
    submitHandler: () => Promise<void> | void
    submitSeverity?: 'danger' | 'warn' | 'primary' | 'secondary'
    /** Hide Cancel/Or controls for submit-only forms. */
    showCancel?: boolean
    /** Optional PrimeIcon displayed before the submit label. */
    submitIcon?: string
  }>(),
  {
    footer: true,
    submitDisabled: false,
    submitSeverity: undefined,
    showCancel: true,
    submitIcon: undefined,
  },
)

const emit = defineEmits<{ cancel: [] }>()

// ── Composables ───────────────────────────────────────────────
const { isSubmitting, run: runSubmit } = useFormSubmitLock()
const { submitDisabled: effectiveSubmitDisabled, validate } = provideFormValidation({
  submitDisabled: () => props.submitDisabled,
  submitLoading: isSubmitting,
})

// ── Functions ─────────────────────────────────────────────────
// Ignore native submits while disabled; otherwise hold the lock until the
// caller's complete async submit flow has finished.
async function submitForm(): Promise<void> {
  if (effectiveSubmitDisabled.value) return
  if (!(await validate())) return
  await runSubmit(props.submitHandler)
}
</script>

<template>
  <section class="layout-form">
    <SectionHeader v-if="title">{{ title }}</SectionHeader>

    <!-- Fields (and, when footer=false, the form's own integrated action row) -->
    <form novalidate class="layout-form__form" @submit.prevent="submitForm">
      <slot />

      <LayoutSubmitFooter
        v-if="footer"
        :text-cancel-label="textCancelLabel"
        :text-or="textOr"
        :text-submit-label="textSubmitLabel"
        :submit-test-id="submitTestId"
        :submit-severity="submitSeverity"
        :show-cancel="showCancel"
        :submit-icon="submitIcon"
        @cancel="emit('cancel')"
      />
    </form>
  </section>
</template>

<style scoped>
/* Vertical field stack; individual panels set their own inner grids/rows. */
.layout-form__form {
  display: flex;
  flex-direction: column;
  gap: 0;
}
</style>
