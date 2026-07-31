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
 * <SortableTh> — a clickable table header cell with a tri-state sort glyph that
 * matches PrimeVue's DataTable icons (neutral `pi-sort-alt`, directional
 * `pi-sort-amount-up-alt` / `pi-sort-amount-down` when active). Pairs with
 * {@link useGroupSort}: pass the column's `field`, the sort's `active` field and
 * `direction`, and handle `@toggle` (typically `sort.toggle`). The label is the slot.
 */
import { computed } from 'vue'

import type { SortDirection } from '@/composables/useGroupSort'

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{
  /** This column's sort key. */
  field: string
  /** The currently-active sort field (from the GroupSort). */
  active: string
  /** The active sort direction. */
  direction: SortDirection
}>()

const emit = defineEmits<{ toggle: [field: string] }>()

// ── Computed ──────────────────────────────────────────────────
const isActive = computed(() => props.active === props.field)

// Expose the current sort direction to assistive technology
const ariaSort = computed(() => {
  if (!isActive.value) return 'none'
  return props.direction === 'asc' ? 'ascending' : 'descending'
})

// Tri-state glyph matching PrimeVue's DataTable sort icons: SortAlt (neutral),
// SortAmountUpAlt (asc), SortAmountDown (desc) — as their PrimeIcons equivalents.
const sortIconClass = computed(() => {
  if (!isActive.value) return 'pi pi-sort-alt'
  return props.direction === 'asc' ? 'pi pi-sort-amount-up-alt' : 'pi pi-sort-amount-down'
})

function onToggle(): void {
  emit('toggle', props.field)
}
</script>

<template>
  <th
    class="sortable-th"
    :class="{ 'sortable-th--active': isActive }"
    scope="col"
    :aria-sort="ariaSort"
  >
    <button type="button" class="sortable-th__button" @click="onToggle">
      <span class="sortable-th__label"><slot /></span>
      <i :class="sortIconClass" class="sortable-th__icon" aria-hidden="true" />
    </button>
  </th>
</template>

<style scoped>
/* Self-contained styling for the sortable header cell. Scoped so the component
   works in any table, and so these rules override the shared data-grid base
   header rule (`.data-grid__table thead th`) when used inside a data grid:
   the scoped attribute lifts specificity above that base selector. */
.sortable-th {
  user-select: none;
}

.sortable-th:has(.sortable-th__button:hover) {
  background: var(--p-surface-50);
}

/* Active (sorted) header — teal highlight matching PrimeVue's DataTable sorted
   column (its headerCell selectedBackground / selectedColor resolve to these
   highlight tokens). */
.sortable-th--active {
  background: var(--p-highlight-background);
  color: var(--p-highlight-color);
}

.sortable-th__button {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  width: 100%;
  padding: 0;
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  text-align: inherit;
  cursor: pointer;
}

.sortable-th__button:focus-visible {
  outline: 2px solid var(--p-focus-ring-color);
  outline-offset: 2px;
}

.sortable-th__icon {
  opacity: 0.5;
}

.sortable-th__button:hover .sortable-th__icon {
  opacity: 0.8;
}

.sortable-th--active .sortable-th__icon {
  opacity: 1;
  color: inherit;
}
</style>
