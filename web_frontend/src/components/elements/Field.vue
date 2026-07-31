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
 * <Field> — labeled field wrapper rendering an optional leading icon,
 * a pointing help label, and an error message.
 *
 * Owns none of the control — the consumer slots it — so the same wrapper fits
 * inputs, selects, autocompletes, and any other control type. The generated
 * `fieldId` is exposed as a slot prop so the consumer can bind it to the control
 * for proper `<label for>` association.
 */
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import { computed, useId } from 'vue'

import FieldHint from '@/components/elements/FieldHint.vue'
import HelpTooltip from '@/components/elements/HelpTooltip.vue'
import { registerFormValidator } from '@/composables/useFormValidation'

import type { FormValidator } from '@/composables/useFormValidation'

// Keep Vue's runtime component name unambiguous while exposing the requested Field.vue API
defineOptions({ name: 'BaseField' })

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    label?: string | undefined
    required?: boolean
    /** Optional help text — shown as an always-visible pointing FieldHint below the field. */
    help?: string | undefined
    /**
     * Present the help hint as a `?` HelpTooltip beside the label instead of an
     * inline bubble below the field (keeps dense forms compact). Per field, so a
     * form can mix tooltip and inline hints. Errors always render inline regardless.
     */
    tooltip?: boolean
    /** Optional leading PrimeIcon class (e.g. `pi pi-user`); wraps the slot in an IconField. */
    icon?: string | undefined
    error?: string | undefined
    /**
     * Optional on-blur validator (typically a useZodForm `validateField` bound to
     * this field). Fired when focus leaves the complete field, not while focus
     * moves between controls inside a composite field. Its return value is ignored.
     */
    validate?: (() => boolean | void) | undefined
    /** Run the field validator when focus leaves the complete field. */
    validateOnBlur?: boolean
    /** Validator registered with the surrounding form independently of blur handling. */
    formValidator?: FormValidator
  }>(),
  { formValidator: undefined, tooltip: false, validateOnBlur: true },
)

// ── Composables ───────────────────────────────────────────────
// Stable id linking the label to the slotted control. Exposed as a slot prop so
// the consumer can bind it (`<Field v-slot="{ id }"><InputText :id="id" />`),
// giving proper label/control association without each caller inventing an id.
const fieldId = useId()
if (props.formValidator) registerFormValidator(props.formValidator)

// ── Computed ──────────────────────────────────────────────────
// In tooltip mode the neutral help hint moves onto the label as a `?` icon, so
// it must not also render inline below the field. Errors are unaffected — they
// always pass through to the inline FieldHint.
const inlineHelp = computed(() => (props.tooltip ? undefined : props.help))
const tooltipHelp = computed(() => (props.tooltip ? props.help : undefined))

// ── Functions ─────────────────────────────────────────────────
// Validate only after focus leaves every control belonging to this field
function validateOnFocusout(event: FocusEvent): void {
  if (!props.validateOnBlur) return

  const field = event.currentTarget
  const nextTarget = event.relatedTarget
  if (!(field instanceof HTMLElement)) return
  if (nextTarget instanceof Node && field.contains(nextTarget)) return

  props.validate?.()
}
</script>

<template>
  <div class="field" @focusout="validateOnFocusout">
    <label v-if="label" class="field__label" :for="fieldId">
      {{ label }}
      <span v-if="required" class="field__required" aria-hidden="true">*</span>
      <HelpTooltip v-if="tooltipHelp" :text="tooltipHelp" />
    </label>
    <!-- Leading-icon variant wraps the slotted control in an IconField. The id
         is exposed so the consumer can bind it to the control for label/for. -->
    <IconField v-if="icon">
      <InputIcon :class="icon" />
      <slot :id="fieldId" />
    </IconField>
    <slot v-else :id="fieldId" />
    <FieldHint :error="error" :text="inlineHelp" />
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 1rem;
}

/* Field label emphasis. inline-flex + center so an inline HelpTooltip (`?`) sits
   optically centered on the label line rather than dropping below its baseline
   (mirrors SectionHeader's .section-header__content). */
.field__label {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  font-weight: 600;
}

.field__required {
  color: var(--p-red-500);
  margin-left: 0.25rem;
}

/* Keep invalid controls red while PrimeVue applies its higher-specificity hover and focus states */
.field :deep(.p-invalid),
.field :deep(.p-invalid:enabled:hover),
.field :deep(.p-invalid:enabled:focus),
.field :deep(.p-invalid.p-focus),
.field :deep(.p-invalid:focus-within),
.field :deep(.p-autocomplete.p-invalid .p-autocomplete-input-multiple),
.field :deep(.p-autocomplete.p-invalid:not(.p-disabled):hover .p-autocomplete-input-multiple),
.field :deep(.p-autocomplete.p-invalid.p-focus .p-autocomplete-input-multiple) {
  border-color: var(--p-form-field-invalid-border-color, var(--p-red-500));
}

/* Darken the leading input icon while the field is focused. The base color is the
   themed iconfield token (main.ts, ~0.87 body text); focus deepens it to full text color. */
.field :deep(.p-iconfield:focus-within .p-inputicon) {
  color: var(--p-text-color);
}

/* When a FloatLabel is slotted inside the IconField wrapper its InputText is
   the first child of FloatLabel (not of IconField), so PrimeVue's native
   `:not(:first-child)` padding rule and `has(.p-inputicon:first-child)` label
   rule both miss. Restore both manually using PrimeVue's own CSS variables so
   the values stay in sync with any theme changes. */
.field :deep(.p-iconfield .p-floatlabel .p-inputtext) {
  padding-inline-start: calc((var(--p-inputtext-padding-x) * 2) + var(--p-icon-size));
}

.field :deep(.p-iconfield .p-floatlabel label) {
  inset-inline-start: calc((var(--p-inputtext-padding-x) * 2) + var(--p-icon-size));
}
</style>
