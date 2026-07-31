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

/** Renders a semantic section heading in the shared dividing, horizontal, or plain style. */
import HelpTooltip from './HelpTooltip.vue'

// ── Props ─────────────────────────────────────────────────────
withDefaults(
  defineProps<{
    /** Semantic heading level and matching type-scale tier. */
    level?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5'
    /** Divider style. */
    variant?: 'dividing' | 'horizontal' | 'plain'
    /** Title text rendered inside the heading. */
    title?: string
    /** Optional PrimeIcon class (e.g. "pi pi-cog") rendered before the title. */
    icon?: string
    /** Optional help text — renders <HelpTooltip> after the title when present. */
    helpText?: string
    /** Stable test hook forwarded to the rendered heading. */
    testId?: string
    /** Optional stable test hook forwarded to the help icon. */
    helpTestId?: string
  }>(),
  { level: 'h2', variant: 'dividing', title: undefined, icon: undefined },
)
</script>

<template>
  <component
    :is="level"
    class="section-header"
    :class="[`section-header--${variant}`, `section-header--${level}`]"
    :data-testid="testId"
  >
    <span class="section-header__content">
      <i v-if="icon" :class="icon" class="section-header__icon" aria-hidden="true" />
      {{ title }}
      <slot />
      <HelpTooltip v-if="helpText" :text="helpText" :test-id="helpTestId" />
    </span>
    <span v-if="$slots.suffix" class="section-header__suffix">
      <slot name="suffix" />
    </span>
  </component>
</template>

<style scoped>
/* Clear any floated siblings so the header always starts on its own row */
.section-header {
  clear: both;
}

.section-header__content {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.section-header__icon {
  font-size: 0.9em;
}

/* ── Dividing (default) ─────────────────────────────────────── */
.section-header--dividing {
  display: block;
  margin: 0 0 0.75em;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--p-content-border-color);
}

/* Keep the optional suffix on the title baseline and at the opposite edge. */
.section-header--dividing:has(.section-header__suffix) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

/* ── Plain (type scale only, no rule) ───────────────────────── */
.section-header--plain {
  display: block;
  margin: 0;
}

/* ── Horizontal (line each side) ────────────────────────────── */
.section-header--horizontal {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

/* Render the horizontal divider on both sides of the title. */
.section-header--horizontal::before,
.section-header--horizontal::after {
  flex: 1;
  height: 1px;
  background: var(--p-content-border-color);
  content: '';
}

.section-header--horizontal::before {
  margin-right: 0.75rem;
}

.section-header--horizontal::after {
  margin-left: 0.75rem;
}
</style>
