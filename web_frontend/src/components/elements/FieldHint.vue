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

/** Renders a field hint or validation error and registers visible errors with its form. */
import { computed } from 'vue'

import { registerFormValidationError } from '@/composables/useFormValidation'

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** Neutral hint text (alternative to the default slot). */
    text?: string
    /** Validation error — takes precedence over the hint when set (error scheme). */
    error?: string
    /** Color scheme for the hint (ignored when `error` is set). */
    severity?: 'help' | 'info' | 'warn' | 'error'
    /** Which side of the annotated element the bubble sits on (arrow points at it). */
    placement?: 'below' | 'above'
    /** Whether the hint draws an arrow toward an individual annotated field. */
    showArrow?: boolean
  }>(),
  {
    text: '',
    error: undefined,
    severity: 'help',
    placement: 'below',
    showArrow: true,
  },
)

// ── Computed ──────────────────────────────────────────────────
// A validation error replaces the resting hint and changes its accessibility role.
const showError = computed(() => !!props.error)
registerFormValidationError(showError)
</script>

<template>
  <div
    v-if="error || text || $slots.default"
    class="field-hint"
    :class="[
      `field-hint--${showError ? 'error' : severity}`,
      `field-hint--${placement}`,
      { 'field-hint--without-arrow': !showArrow },
    ]"
    :role="showError ? 'alert' : undefined"
    :data-testid="showError ? 'field-error' : 'field-hint'"
  >
    <template v-if="showError">{{ error }}</template>
    <slot v-else>{{ text }}</slot>
  </div>
</template>

<style scoped>
/* Pointing hint bubble anchored to the field it annotates. Colors default to
   the neutral surface scheme and are overridden per severity below; the vertical
   offset and arrow direction are set per placement (below/above). */
.field-hint {
  position: relative;
  display: block;
  padding: 0.3rem 0.55rem;
  background: var(--field-hint-bg);
  border: 1px solid var(--field-hint-border);
  border-radius: var(--p-content-border-radius);
  color: var(--field-hint-text);
  font-size: var(--app-font-size-small);
  line-height: 1.3;
}

/* Border-triangle: a rotated square with two adjacent borders, positioned on the
   arrow side. Both placements share the corner selection; only its vertical
   anchor differs, set per placement below. */
.field-hint::before {
  content: '';
  position: absolute;
  left: 0.9rem;
  width: 0.6rem;
  height: 0.6rem;
  background: var(--field-hint-bg);
  border-left: 1px solid var(--field-hint-border);
  border-top: 1px solid var(--field-hint-border);
  transform: rotate(45deg);
}

/* Group-level hints omit the pointer because they annotate multiple sibling fields */
.field-hint--without-arrow::before {
  display: none;
}

/* Below the field (default): bubble sits under it, arrow points up. */
.field-hint--below {
  margin-top: 0.6rem;
}

.field-hint--below::before {
  top: -0.34rem;
}

/* Above the field: bubble sits over it, arrow points down. The lit borders swap
   to the bottom/right pair so the triangle's solid edge faces the field below. */
.field-hint--above {
  margin-bottom: 0.6rem;
}

.field-hint--above::before {
  bottom: -0.34rem;
  border-left: none;
  border-top: none;
  border-right: 1px solid var(--field-hint-border);
  border-bottom: 1px solid var(--field-hint-border);
}

/* Neutral hint (default) reuses the popover/tooltip surface look (main.ts popover
   tokens): white background, subtle surface-300 border, muted text — so a field
   hint reads as the same neutral "informational popup" language. */
.field-hint--help {
  --field-hint-bg: var(--p-surface-50);
  --field-hint-border: var(--p-surface-300);
  --field-hint-text: var(--p-text-muted-color);
}

/* Info/warn/error schemes mirror the <Message> palette (main.ts message.colorScheme). */
.field-hint--info {
  --field-hint-bg: var(--app-info-bg);
  --field-hint-border: var(--app-info-border);
  --field-hint-text: var(--app-info-text);
}

.field-hint--warn {
  --field-hint-bg: var(--app-warn-bg);
  --field-hint-border: var(--app-warn-border);
  --field-hint-text: var(--app-warn-text);
}

.field-hint--error {
  --field-hint-bg: var(--app-error-bg);
  --field-hint-border: var(--app-error-border);
  --field-hint-text: var(--app-error-text);
}
</style>
