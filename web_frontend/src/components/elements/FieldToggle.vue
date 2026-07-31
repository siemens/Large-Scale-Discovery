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
 * <FieldToggle> — inline boolean field: a ToggleSwitch with its label beside it
 * and the shared help/error message below. The label/control association is wired via
 * a generated id so clicking the label toggles the switch.
 */
import ToggleSwitch from 'primevue/toggleswitch'
import { computed, useId } from 'vue'

import FieldHint from '@/components/elements/FieldHint.vue'
import HelpTooltip from '@/components/elements/HelpTooltip.vue'

// ── Props and model ───────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    label: string
    required?: boolean
    /** Neutral help hint shown below the row. */
    help?: string | undefined
    /**
     * Present the help hint as a `?` HelpTooltip beside the label instead of an
     * inline bubble below the row. Per field. Errors always render inline.
     */
    tooltip?: boolean
    /** Validation error — takes precedence over `help`. */
    error?: string | undefined
    disabled?: boolean
    /** Validate after focus leaves the switch and its label/help row. */
    validate?: (() => void) | undefined
    /** Optional stable test hook applied to the complete toggle field. */
    testId?: string
  }>(),
  {
    required: false,
    help: undefined,
    tooltip: false,
    error: undefined,
    disabled: false,
    validate: undefined,
  },
)

// Two-way boolean value; clicking the label or the switch itself toggles it.
const model = defineModel<boolean>({ required: true })

// ── Composables ───────────────────────────────────────────────
// Associate the label with the switch via a generated id so clicking it toggles the switch
const controlId = useId()

// ── Computed ──────────────────────────────────────────────────
// In tooltip mode the hint moves beside the label as a `?` icon; errors still
// render inline below the row.
const inlineHelp = computed(() => (props.tooltip ? undefined : props.help))
const tooltipHelp = computed(() => (props.tooltip ? props.help : undefined))

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
    class="field-toggle"
    :data-testid="testId"
    :class="{ 'field-toggle--disabled': disabled }"
    @focusout="validateOnFocusout"
  >
    <div class="field-toggle__row">
      <ToggleSwitch v-model="model" :input-id="controlId" :disabled="disabled" :invalid="!!error" />
      <label class="field-toggle__label" :for="controlId">
        {{ label }}
        <span v-if="required" class="field-toggle__required" aria-hidden="true">*</span>
      </label>
      <HelpTooltip v-if="tooltipHelp" :text="tooltipHelp" />
    </div>
    <FieldHint :error="error" :text="inlineHelp" />
  </div>
</template>

<style scoped>
/* Row: switch then label; message stacks below on its own line. */
.field-toggle {
  margin-bottom: 0.5rem;
}

.field-toggle__row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Inline labels read as normal weight (they sit beside the control, not above). */
.field-toggle__label {
  font-weight: normal;
}

.field-toggle__required {
  margin-left: 0.25rem;
  color: var(--p-red-500);
}

.field-toggle--disabled {
  opacity: 0.55;
}
</style>
