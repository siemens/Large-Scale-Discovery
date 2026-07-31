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
 * <FieldSlider> — numeric slider field: a label above a drag slider, with the
 * shared help/error message below.
 *
 * When `showValue` is set it renders a paired numeric readout input kept in sync
 * with the slider. Add `unit` to append a gray addon chip (e.g. "Days") to the
 * right of that input. Without `showValue` it renders a plain full-width slider.
 */
import Slider from 'primevue/slider'
import { computed, useId } from 'vue'

import FieldHint from '@/components/elements/FieldHint.vue'
import FieldNumber from '@/components/elements/FieldNumber.vue'
import HelpTooltip from '@/components/elements/HelpTooltip.vue'

// ── Props and model ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    label: string
    required?: boolean
    help?: string | undefined
    /**
     * Present the help hint as a `?` HelpTooltip beside the label instead of an
     * inline bubble below the field. Per field. Errors always render inline.
     */
    tooltip?: boolean
    error?: string | undefined
    min?: number
    max?: number
    step?: number
    disabled?: boolean
    /** Validate after focus leaves the complete slider/value control. */
    validate?: (() => void) | undefined
    /** Show a paired numeric input beside the slider, kept in sync with it. */
    showValue?: boolean
    /** Place the label to the left of the slider row instead of above it (legacy
     * agent-process layout). Only meaningful together with `showValue`. */
    inlineLabel?: boolean
    /** Optional unit label rendered as a gray addon chip to the right of the numeric
     *  input (e.g. "Days", "Seconds"). Only meaningful together with `showValue`. */
    unit?: string | undefined
  }>(),
  {
    help: undefined,
    required: false,
    tooltip: false,
    error: undefined,
    min: 0,
    max: 100,
    step: 1,
    disabled: false,
    validate: undefined,
    showValue: false,
    inlineLabel: false,
    unit: undefined,
  },
)

// v-model number shared by the slider and the optional numeric readout.
const model = defineModel<number>({ required: true })

// ── Composables ───────────────────────────────────────────────
// Associate the label with the numeric input when present, else the slider
const controlId = useId()

// ── Computed ──────────────────────────────────────────────────
// In tooltip mode the hint moves beside the label as a `?` icon; errors still
// render inline below the field.
const inlineHelp = computed(() => (props.tooltip ? undefined : props.help))
const tooltipHelp = computed(() => (props.tooltip ? props.help : undefined))

// ── Functions ─────────────────────────────────────────────────
// Slider emits number | number[]; this field is single-value, so normalize.
function updateSliderValue(value: number | number[]): void {
  model.value = Array.isArray(value) ? (value[0] ?? props.min) : value
}

function updateNumberValue(value: number | null | undefined): void {
  model.value = value ?? props.min
}

function validateOnFocusout(event: FocusEvent): void {
  const field = event.currentTarget
  const nextTarget = event.relatedTarget
  if (field instanceof HTMLElement && nextTarget instanceof Node && field.contains(nextTarget))
    return
  props.validate?.()
}
</script>

<template>
  <div
    class="field-slider"
    :class="{
      'field-slider--disabled': disabled,
      'field-slider--invalid': !!error,
      'field-slider--inline-label': inlineLabel && showValue,
    }"
    @focusout="validateOnFocusout"
  >
    <label class="field-slider__label" :for="controlId"
      >{{ label }}<span v-if="required" class="field-slider__required" aria-hidden="true">*</span
      ><HelpTooltip v-if="tooltipHelp" :text="tooltipHelp" />
    </label>

    <div v-if="showValue" class="field-slider__row">
      <Slider
        :model-value="model"
        :min="min"
        :max="max"
        :step="step"
        :disabled="disabled"
        :invalid="!!error"
        :aria-required="required || undefined"
        :aria-invalid="error ? 'true' : undefined"
        class="field-slider__slider"
        :aria-label="label"
        @update:model-value="updateSliderValue"
      />
      <FieldNumber
        :id="controlId"
        :model-value="model"
        :min="min"
        :max="max"
        :disabled="disabled"
        :unit="unit"
        class="field-slider__value"
        @update:model-value="updateNumberValue"
      />
    </div>

    <Slider
      v-else
      :id="controlId"
      :model-value="model"
      :min="min"
      :max="max"
      :step="step"
      :disabled="disabled"
      :aria-required="required || undefined"
      :aria-invalid="error ? 'true' : undefined"
      class="field-slider__slider"
      :aria-label="label"
      @update:model-value="updateSliderValue"
    />

    <FieldHint :error="error" :text="inlineHelp" />
  </div>
</template>

<style scoped>
/* Label above, slider (or slider+value row) below, message last. */
.field-slider {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
  min-width: 0;
}

/* inline-flex + center so an inline HelpTooltip (`?`) sits on the label line. */
.field-slider__label {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  font-weight: 600;
}

.field-slider__required {
  margin-left: 0.25rem;
  color: var(--p-red-500);
}

/* Slider + numeric readout on one line (showValue mode). Extra gap so the value
   box is clearly separated from the slider track (legacy parity). */
.field-slider__row {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

/* Inline-label layout (legacy agent-process rows): label sits to the LEFT of the
   slider row on a single line, rather than stacked above it. The label is given a
   fixed share so the sliders across the two grid columns line up vertically. */
.field-slider--inline-label {
  flex-direction: row;
  flex-wrap: wrap;
  align-items: center;
  gap: 1rem;
}

/* Center the label (and its inline `?` tooltip) against the slider row. The base
   label rule pins align-self to flex-start for the stacked layout; here the label
   shares a row with the slider, so it must center vertically instead. Lighter
   weight matches the compact row style (vs. 600 for above-field labels). */
.field-slider--inline-label .field-slider__label {
  flex: 0 0 6rem;
  align-self: center;
  line-height: 1;
  font-weight: 500;
}

.field-slider--inline-label .field-slider__row {
  flex: 1 1 auto;
  min-width: 0;
}

.field-slider__slider {
  width: 100%;
  min-width: 4rem;
}

/* In stacked layout the thumb at position 0 would sit left of the label text.
   A small left padding brings it flush under the label's first character. Not
   applied in inline-label mode where the slider sits beside the label, not below it. */
.field-slider:not(.field-slider--inline-label) .field-slider__slider {
  margin-left: 0.4rem;
}

/* Slider thumb border. Aura exposes no handle border token, so the thin all-around
   border that separates the surface-50 thumb from the track is set here (the thumb
   fill/size/shadow come from the LsdPreset in main.ts). */
.field-slider__slider :deep(.p-slider-handle) {
  border: 1px solid var(--p-surface-300);
}

.field-slider--invalid .field-slider__slider :deep(.p-slider-range),
.field-slider--invalid .field-slider__slider :deep(.p-slider-handle) {
  border-color: var(--p-form-field-invalid-border-color, var(--p-red-500));
}

/* Suppress the persistent focus ring on the handle after dragging. */
.field-slider__slider :deep(.p-slider-handle:focus-visible) {
  outline: none;
  box-shadow: none;
}

/* Prevent the FieldNumber from stretching to fill the slider row.
   The inner .p-inputnumber is capped at 4.5rem; the InputGroup wrapper follows suit. */
.field-slider__value {
  flex: 0 0 auto;
  width: auto;
}

.field-slider__value :deep(.p-inputgroup) {
  width: auto;
}

.field-slider__value :deep(.p-inputnumber) {
  flex: 0 0 6rem;
  width: 6rem;
}

.field-slider__value :deep(.p-inputnumber-input) {
  /* !important: overrides runtime-injected @primeuix/styles which loads after scoped CSS */
  padding-inline: 0.25rem !important;
}

.field-slider__value :deep(.p-inputnumber-button) {
  /* Buttons use a fixed width token, not padding — shrink width instead */
  width: 1.75rem !important;
}

.field-slider--disabled {
  opacity: 0.55;
}
</style>
