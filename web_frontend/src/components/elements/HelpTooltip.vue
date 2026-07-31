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

/** Renders keyboard-accessible contextual help with a right-positioned tooltip. */

// ── Props ─────────────────────────────────────────────────────
defineProps<{
  /** The help text shown in the tooltip. */
  text: string
  /** Optional stable test hook; defaults to the generic help-tooltip hook. */
  testId?: string
}>()
</script>

<template>
  <i
    v-tooltip.right="text"
    class="pi pi-question-circle help-tooltip"
    :data-testid="testId ?? 'help-tooltip'"
    role="img"
    :aria-label="text"
    tabindex="0"
  />
</template>

<style scoped>
/* Scale and align the icon with either heading or field-label text. */
.help-tooltip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1em;
  height: 1em;
  margin-left: 0.25rem;
  border-radius: 50%;
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
  line-height: 1;
  vertical-align: middle;
  cursor: help;
}
</style>
