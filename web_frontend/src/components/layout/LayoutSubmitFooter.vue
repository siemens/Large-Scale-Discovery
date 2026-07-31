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
 * <LayoutSubmitFooter> — the standard "Cancel · or · Submit" action row shared by the
 * app's form dialogs (view / group / database add+edit, feedback).
 *
 * The submit button supports both wiring styles already in use:
 *   - `submitType="submit"` (default) for dialogs whose body is a <form>, so
 *     Enter submits natively — the component emits nothing for submit then.
 *   - `submitType="button"` for dialogs that drive submit via @click; the
 *     component emits `submit` on click.
 * Cancel always emits `cancel`.
 *
 * Inside LayoutForm/DialogForm, validation, external disabled state, and the
 * pending spinner are inherited automatically from the surrounding form.
 */
import Button from 'primevue/button'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import OrConnector from '@/components/elements/OrConnector.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import { useFormSubmitState } from '@/composables/useFormValidation'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    cancel: 'Cancel',
    or: 'or',
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** Label for the secondary cancel action. */
    textCancelLabel?: string
    /** Text displayed between cancel and submit actions. */
    textOr?: string
    /** Label for the primary submit action. */
    textSubmitLabel: string
    /** Stable test hook forwarded to the submit button. */
    submitTestId?: string
    cancelTestId?: string
    /** 'submit' triggers the surrounding form; 'button' emits `submit` on click. */
    submitType?: 'submit' | 'button'
    submitDisabled?: boolean
    submitLoading?: boolean
    /** Submit-button severity, e.g. 'danger' for destructive confirms. Default primary. */
    submitSeverity?: 'danger' | 'warn' | 'primary' | 'secondary'
    /** Hide the Cancel/Or controls for legacy submit-only forms. */
    showCancel?: boolean
    /** Render a submit-only action with the legacy segmented icon rail. */
    segmentedSubmit?: boolean
    /** PrimeIcon for a submit-only action. Segmented submits default to a check mark. */
    submitIcon?: string
    /** HTML form id to associate the submit button with when it lives outside the <form>. */
    form?: string
  }>(),
  {
    submitType: 'submit',
    submitDisabled: false,
    submitLoading: false,
    submitSeverity: undefined,
    showCancel: true,
    segmentedSubmit: true,
    submitIcon: undefined,
    form: undefined,
  },
)

const emit = defineEmits<{ cancel: []; submit: [] }>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

const formSubmitState = useFormSubmitState()
const effectiveSubmitLoading = computed(
  () => props.submitLoading || formSubmitState.submitLoading.value,
)
const effectiveSubmitDisabled = computed(
  () =>
    props.submitDisabled || effectiveSubmitLoading.value || formSubmitState.submitDisabled.value,
)
const usesSegmentedSubmit = computed(() => props.segmentedSubmit && !props.showCancel)
const submitComponent = computed(() => (usesSegmentedSubmit.value ? SegmentedButton : Button))
const effectiveSubmitIcon = computed(() => {
  if (props.showCancel) return undefined
  return props.submitIcon ?? (usesSegmentedSubmit.value ? 'pi pi-check' : undefined)
})

// ── Functions ─────────────────────────────────────────────────
// A native submit button is handled by the surrounding form, so only the
// explicit button variant emits on click
function onSubmitClick(): void {
  if (props.submitType === 'button') emit('submit')
}
</script>

<template>
  <div class="layout-submit-footer">
    <template v-if="showCancel">
      <Button
        :data-testid="cancelTestId"
        type="button"
        severity="secondary"
        :label="props.textCancelLabel ?? t('cancel')"
        @click="emit('cancel')"
      />
      <OrConnector :text-or="props.textOr ?? t('or')" />
    </template>
    <component
      :is="submitComponent"
      :type="submitType"
      :form="form"
      :label="textSubmitLabel"
      :icon="effectiveSubmitIcon"
      :severity="submitSeverity"
      :disabled="effectiveSubmitDisabled"
      :aria-busy="effectiveSubmitLoading"
      :data-loading="effectiveSubmitLoading ? 'true' : undefined"
      :data-testid="submitTestId ?? (showCancel ? undefined : 'scan-submit')"
      @click="onSubmitClick"
    />
  </div>
</template>

<style scoped>
/* Right-aligned action row with zero gap so the OrConnector badge bridges the buttons */
.layout-submit-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0;
  width: 100%;
}
</style>
