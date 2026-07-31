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
 * <FieldNumber> — a numeric stepper with an optional unit chip on the right.
 *
 * Wraps PrimeVue InputGroup + InputNumber (show-buttons, button-layout="horizontal")
 * + optional InputGroupAddon. The +/- buttons are styled like <FieldStepper>:
 * teal icon color, no internal border, rounded-rectangle teal hover.
 */
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import InputNumber from 'primevue/inputnumber'

const model = defineModel<number | null>()

const props = withDefaults(
  defineProps<{
    id?: string
    min?: number
    max?: number
    step?: number
    placeholder?: string
    disabled?: boolean
    invalid?: boolean
    /** Label shown in the right-hand chip. Omit to render without an addon. */
    unit?: string
  }>(),
  {
    id: undefined,
    min: undefined,
    max: undefined,
    step: 1,
    placeholder: undefined,
    disabled: false,
    invalid: false,
    unit: undefined,
  },
)
</script>

<template>
  <InputGroup :class="['field-number', { 'field-number--with-unit': props.unit }]">
    <InputNumber
      v-model="model"
      :input-id="props.id"
      :min="props.min"
      :max="props.max"
      :step="props.step"
      :placeholder="props.placeholder"
      :disabled="props.disabled"
      :invalid="props.invalid"
      show-buttons
      button-layout="horizontal"
    />
    <InputGroupAddon v-if="props.unit">{{ props.unit }}</InputGroupAddon>
  </InputGroup>
</template>

<style scoped>
.field-number {
  width: 100%;
}

/* Outer border wraps all three parts (decrement | input | increment) as one unit.
   overflow:hidden clips button hover backgrounds to the border-radius. */
.field-number :deep(.p-inputnumber) {
  width: auto;
  flex: 1 1 auto;
  border: 1px solid var(--p-content-border-color);
  border-radius: var(--p-content-border-radius);
  overflow: hidden;
}

/* When a unit addon follows, square off the right side of the number input. */
.field-number--with-unit :deep(.p-inputnumber) {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

/* Unit chip: muted gray pill aligned with the legacy "right labeled input" look. */
.field-number :deep(.p-inputgroupaddon) {
  width: 5.5rem;
  flex-shrink: 0;
  justify-content: center;
  background: var(--p-surface-100);
  border-color: var(--p-content-border-color);
  color: var(--p-text-muted-color);
}

/* Teal icon color: InputNumber icons are bare SVGs with fill="currentColor";
   color must be set on the button itself, not on a child element. */
.field-number :deep(.p-inputnumber-button) {
  background: transparent;
  border: none;
  box-shadow: none;
  color: var(--p-primary-500);
}

/* Strip the text input's own border — the .p-inputnumber wrapper provides it. */
.field-number :deep(.p-inputnumber-input) {
  border: none;
  box-shadow: none;
  text-align: center;
}

/* Rounded-rectangle teal hover, matching the FieldStepper hover pattern.
   border-radius not needed here — overflow:hidden on the parent clips it. */
.field-number :deep(.p-inputnumber-button:hover) {
  background: color-mix(in srgb, var(--p-primary-500) 10%, transparent);
}
</style>
