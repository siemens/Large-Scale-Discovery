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
 * <TriStateFilter> — polyfill for the PrimeVue TriStateCheckbox
 * that was deprecated and removed in v4 (see memory lsd-web-primevue-v4-gotchas).
 *
 * Cycles through three filter states on click:
 *   null  → true  → false  → null (then repeats)
 *
 * Visually renders as a checkbox-like square rather than a bare icon button,
 * matching the legacy Tabulator `tickCross` header filter appearance.
 *
 * Used by <DataTable> for columns with filterType='tristate'.
 */

// ── State ─────────────────────────────────────────────────────
const model = defineModel<boolean | null>({ default: null })
defineProps<{
  label: string
}>()

// ── Functions ─────────────────────────────────────────────────
function cycle(): void {
  if (model.value === null) {
    model.value = true
  } else if (model.value) {
    model.value = false
  } else {
    model.value = null
  }
}
</script>

<template>
  <button
    type="button"
    class="data-table-tri-state-filter"
    :class="{
      'data-table-tri-state-filter--checked': model === true,
      'data-table-tri-state-filter--crossed': model === false,
    }"
    :aria-checked="model === null ? 'mixed' : model"
    :aria-label="label"
    role="checkbox"
    @click="cycle"
  >
    <i
      v-if="model === true"
      class="pi pi-check data-table-tri-state-filter__icon data-table-tri-state-filter__icon--check"
      aria-hidden="true"
    />
    <i
      v-else-if="model === false"
      class="pi pi-times data-table-tri-state-filter__icon data-table-tri-state-filter__icon--cross"
      aria-hidden="true"
    />
  </button>
</template>

<style scoped>
/* Checkbox-like square (matches the legacy Tabulator tickCross filter look). */
.data-table-tri-state-filter {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  border: 0.142857rem solid var(--p-checkbox-border-color, var(--p-content-border-color));
  border-radius: var(--p-border-radius-sm);
  background: var(--p-checkbox-background, var(--p-content-background));
  transition:
    border-color 0.15s,
    background 0.15s;
  cursor: pointer;
  outline-offset: 2px;
}

.data-table-tri-state-filter--checked {
  border-color: var(--p-primary-500);
  background: var(--p-primary-500);
}

.data-table-tri-state-filter--crossed {
  border-color: var(--p-red-500);
  background: var(--p-red-500);
}

.data-table-tri-state-filter__icon {
  font-size: var(--app-font-size-small);
  line-height: 1;
}

.data-table-tri-state-filter__icon--check {
  color: var(--p-primary-contrast-color);
}

.data-table-tri-state-filter__icon--cross {
  color: var(--app-color-offwhite);
}
</style>
