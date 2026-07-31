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
 * <FieldStepper> — the labelled −/value/+ control used for cycle retention
 * and network segment size in the scope Add forms. Like <FieldMultiSelect>,
 * it bundles <Field> (label / help / id) so call sites stay a single tag
 * rather than wrapping the control themselves.
 *
 * Retention is either a positive integer or `'All'`. Network sizes step only
 * through powers of two up to the Nmap limit. Two-way bound via `v-model`.
 */
import Button from 'primevue/button'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import Field from '@/components/elements/Field.vue'

// Network sizes must map to IPv4 CIDR blocks and stay within Nmap's hard limit
const NETWORK_SIZE_STEPS = [1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1_024, 2_048]

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    all: 'All Cycles',
    cycles: '{count} Cycle | {count} Cycles',
    splitSize: '{count} IP | {count} IPs',
    decreaseAriaLabel: 'Decrease cycle retention',
    increaseAriaLabel: 'Increase cycle retention',
    decreaseNetworkSizeAriaLabel: 'Decrease network size',
    increaseNetworkSizeAriaLabel: 'Increase network size',
  },
}

// ── Props and model ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    label: string
    kind?: 'retention' | 'splitSize'
    disabled?: boolean
    error?: string
    /** Optional help text shown under the field. */
    help?: string
    /** Present the help hint as a `?` tooltip beside the label instead of inline. */
    tooltip?: boolean
    /** Optional prefix for stable test hooks on the control and its three parts. */
    testId?: string
  }>(),
  {
    kind: 'retention',
    disabled: false,
    error: undefined,
    help: undefined,
    tooltip: false,
  },
)

// Two-way value: a positive integer or the retention sentinel 'All'.
const model = defineModel<number | 'All'>({ required: true })

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Computed ────────────────────────────────────────────────────────────────
// Network-size controls enforce the CIDR range at the interaction boundary
const isNetworkSize = computed(() => props.kind === 'splitSize')
const canDecrease = computed(
  () => !isNetworkSize.value || (typeof model.value === 'number' && model.value > 1),
)
const canIncrease = computed(
  () => !isNetworkSize.value || (typeof model.value === 'number' && model.value < 2_048),
)

// Accessible and visible labels reflect the domain represented by the shared control
const decreaseAriaLabel = computed(() =>
  t(isNetworkSize.value ? 'decreaseNetworkSizeAriaLabel' : 'decreaseAriaLabel'),
)
const increaseAriaLabel = computed(() =>
  t(isNetworkSize.value ? 'increaseNetworkSizeAriaLabel' : 'increaseAriaLabel'),
)
const displayValue = computed(() => {
  if (isNetworkSize.value && typeof model.value === 'number') {
    return t('splitSize', { count: model.value }, model.value)
  }
  return model.value === 'All' ? t('all') : t('cycles', { count: model.value }, model.value)
})

// ── Functions ─────────────────────────────────────────────────
// Increment retention by one or select the next CIDR-compatible network size
function increaseValue(): void {
  if (props.disabled || !canIncrease.value) return

  // Network sizes double through the fixed CIDR-compatible list
  if (isNetworkSize.value) {
    const current = typeof model.value === 'number' ? model.value : 1
    const next = NETWORK_SIZE_STEPS.find((splitSize) => splitSize > current)
    if (next !== undefined) model.value = next
    return
  }

  // Retention restarts at one after the all-cycles sentinel
  model.value = model.value === 'All' ? 1 : model.value + 1
}

// Decrement retention by one or select the previous CIDR-compatible network size
function decreaseValue(): void {
  if (props.disabled || !canDecrease.value) return

  // Network sizes halve through the fixed CIDR-compatible list
  if (isNetworkSize.value) {
    const current = typeof model.value === 'number' ? model.value : 1
    const currentIndex = NETWORK_SIZE_STEPS.findIndex((splitSize) => splitSize >= current)
    const previous = currentIndex > 0 ? NETWORK_SIZE_STEPS[currentIndex - 1] : undefined
    if (previous !== undefined) model.value = previous
    return
  }

  // Retention collapses to the sentinel after the final stored cycle
  const current = model.value
  model.value = typeof current === 'number' && current > 1 ? current - 1 : 'All'
}
</script>

<template>
  <Field
    class="field-stepper__field"
    :class="{ 'field-stepper__field--disabled': disabled }"
    :label="label"
    :help="help"
    :tooltip="tooltip"
    :error="error"
    :aria-disabled="disabled ? 'true' : undefined"
  >
    <div
      class="field-stepper"
      :data-testid="testId"
      :class="{ 'field-stepper--invalid': !!error }"
      role="group"
      :aria-label="label"
      :aria-invalid="error ? 'true' : undefined"
    >
      <Button
        type="button"
        icon="pi pi-minus"
        text
        size="small"
        :disabled="disabled || !canDecrease"
        :aria-label="decreaseAriaLabel"
        :data-testid="testId ? `${testId}-sub` : undefined"
        @click="decreaseValue"
      />
      <span
        class="field-stepper__value"
        :data-testid="testId ? `${testId}-value` : undefined"
        aria-live="polite"
      >
        {{ displayValue }}
      </span>
      <Button
        type="button"
        icon="pi pi-plus"
        text
        size="small"
        :disabled="disabled || !canIncrease"
        :aria-label="increaseAriaLabel"
        :data-testid="testId ? `${testId}-add` : undefined"
        @click="increaseValue"
      />
    </div>
  </Field>
</template>

<style scoped>
/* Full-width control shaped like an input box: −/value/+ across the field, with
   the value filling the space between the two edge buttons. */
.field-stepper {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  width: 100%;
  padding: 0.125rem 0.25rem;
  border: 1px solid var(--p-inputtext-border-color, var(--p-content-border-color));
  border-radius: var(--p-content-border-radius);
  background: var(--p-inputtext-background, var(--p-surface-0));
}

.field-stepper__value {
  flex: 1;
  text-align: center;
  font-weight: 600;
}

/* Match scan-settings fields by dimming the label, help, and control together */
.field-stepper__field--disabled {
  opacity: 0.55;
}

.field-stepper--invalid {
  border-color: var(--p-form-field-invalid-border-color, var(--p-red-500));
}
</style>
