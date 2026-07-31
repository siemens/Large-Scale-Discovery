<script setup lang="ts" generic="T">
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
 * <DataGrid> — the shared, grouped, bordered table used by the Scopes and Views
 * (views + access-tokens) config pages.
 *
 * It owns the repeated chrome those pages used to hand-roll: the group `v-for`,
 * the `.data-grid__group` wrapper + group header, the `<table>`, the sort-header
 * row, and the row loop (with its `:key` and optional row class). Per-group sort
 * state is resolved from the passed {@link KeyedGroupSort} and handed to the
 * header slot, so callers stop repeating `sort.forGroup(groupName)…` on every
 * SortableTh.
 *
 * Cells stay caller-authored (the columns are highly bespoke): the `#header` slot
 * renders the `<th>`/`<SortableTh>` set and the `#row` slot renders the `<td>`s.
 * The three-state loading/empty wrapper and section-level test IDs stay in the
 * page (outside this component).
 *
 * The grouped-table styling lives in this component's `<style scoped>`: the frame
 * elements are DataGrid's own DOM, and the slotted `<th>`/`<td>` structure is
 * reached via `:deep()` anchored on the DataGrid-owned `.data-grid__table`.
 *
 * One style concern stays OUT of here on purpose:
 *  - The PrimeVue <DataTable> skin lives in DataTable.vue's scoped styles.
 *
 * Everything else in the data-grid__* namespace — cell modifiers (ellipsis,
 * nowrap) and row-action chrome (btn, actions, cell--actions) — lives here in
 * scoped :deep() rules anchored on .data-grid__table.
 */
import type { GroupSort, KeyedGroupSort } from '@/composables/useGroupSort'

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** Grouped rows as produced by groupByKey (`@/lib/collections`): [name, rows][]. */
    groups: ReadonlyArray<readonly [string, readonly T[]]>
    /** Per-group sort family; DataGrid resolves forGroup(name) for each group. */
    sort: KeyedGroupSort
    /** Stable per-row key (e.g. `s => s.id`, or a composite string). */
    rowKey: (row: T) => string | number
    /** CSS class(es) added to the <table>. Omit when no extra class is needed. */
    tableClass?: string
    /** Static class, or a per-row function, added to each body <tr>. */
    rowClass?: string | ((row: T) => string | undefined)
    /** Vertical alignment of body cells. Default 'middle'; use 'top' when cells can
     *  hold multi-line content (e.g. the Views grant lists). */
    cellAlign?: 'middle' | 'top'
    testId?: string
  }>(),
  {
    cellAlign: 'middle',
  },
)

// ── Functions ─────────────────────────────────────────────────
function resolveRowClass(row: T): string | undefined {
  return typeof props.rowClass === 'function' ? props.rowClass(row) : props.rowClass
}

function resolveGroupSort(groupName: string): GroupSort {
  return props.sort.forGroup(groupName)
}
</script>

<template>
  <div v-for="[groupName, rows] in groups" :key="groupName" class="data-grid__group">
    <div class="data-grid__group-header">
      <slot name="group-header" :group-name="groupName" :rows="rows">
        <h4>
          <span class="sensitive">{{ groupName }}</span>
        </h4>
      </slot>
    </div>
    <table
      class="data-grid__table"
      :data-testid="testId"
      :class="[{ 'data-grid__table--align-top': cellAlign === 'top' }, tableClass]"
    >
      <caption class="data-grid__caption">
        <slot name="caption" :group-name="groupName" :rows="rows">{{ groupName }}</slot>
      </caption>
      <thead>
        <tr>
          <!-- Header cells (SortableTh / plain th). groupSort is the resolved
               GroupSort for this group so SortableTh binds directly to it. -->
          <slot name="header" :group-name="groupName" :group-sort="resolveGroupSort(groupName)" />
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="rowKey(row)" :class="resolveRowClass(row)">
          <slot name="row" :row="row" :group-name="groupName" />
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
/*
 * Scoped to DataGrid. The frame elements (group / header / table) are DataGrid's
 * own DOM, so they scope directly. The cell rules target the `<th>`/`<td>` the
 * host page renders via slots — reached with `:deep()` anchored on the DataGrid-
 * owned `.data-grid__table`, so they still can't leak outside a DataGrid.
 *
 * Naming follows the project BEM convention: `data-grid__<element>` for parts and
 * `data-grid__<element>--<modifier>` for variants.
 */

.data-grid__group {
  padding-bottom: 1.5rem;
}

.data-grid__group-header {
  color: var(--p-text-muted-color);
  padding: 0.75rem;
  background: var(--p-surface-50);
  border: 1px solid var(--p-content-border-color);
  border-bottom: none;
  /* Rounded top only. The table below supplies the rounded bottom so the
       header + table read as one attached card. */
  border-radius: var(--p-border-radius-sm) var(--p-border-radius-sm) 0 0;
}

.data-grid__table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid var(--p-content-border-color);
  border-top: none;
  /* Rounded bottom only. Pairs with the group header's rounded top. */
  border-radius: 0 0 var(--p-border-radius-sm) var(--p-border-radius-sm);
}

/* Name each table for assistive technology without duplicating the visible
   group heading. Consumers may replace the caption through its slot. */
.data-grid__caption {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* Header cells (slotted) — grey header shared with the PrimeVue DataTables;
   per-cell grid borders are specific to these tables. */
.data-grid__table :deep(thead th) {
  padding: 0.5rem 0.75rem;
  background: var(--p-surface-50);
  border-bottom: 1px solid var(--p-content-border-color);
  border-right: 1px solid var(--p-content-border-color);
  font-weight: 700;
  text-align: left;
  color: var(--p-text-color);
  white-space: nowrap;
}

.data-grid__table :deep(thead th.sortable-th:hover) {
  background: var(--p-surface-100);
}

.data-grid__table :deep(thead th.sortable-th--active) {
  background: var(--p-highlight-background);
  color: var(--p-highlight-color);
}

.data-grid__table :deep(thead th.sortable-th--active:hover) {
  background: var(--p-highlight-background);
  color: var(--p-highlight-color);
}

/* Last header cell: no right border (table border closes it) */
.data-grid__table :deep(thead th:last-child) {
  border-right: none;
}

/* Body cells (slotted) — token colors match thead for consistency. */
.data-grid__table :deep(tbody td) {
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--p-content-border-color);
  border-right: 1px solid var(--p-content-border-color);
  color: var(--p-text-color);
  vertical-align: middle;
}

.data-grid__table :deep(tbody td:last-child) {
  border-right: none;
}

/* Last body row: no bottom border (table border closes it) */
.data-grid__table :deep(tbody tr:last-child td) {
  border-bottom: none;
}

/* Opt-in top alignment for tables whose cells hold multi-line content. */
.data-grid__table--align-top :deep(tbody td) {
  vertical-align: top;
}

/* Row-action chrome. Specificity (0,2,0) from scoped :deep() beats the PrimeVue
   preset's (0,1,0), so !important is not needed. */
.data-grid__table :deep(td.data-grid__cell--actions) {
  background-color: var(--p-surface-50);
  text-align: right;
}

.data-grid__table :deep(tr:hover td.data-grid__cell--actions) {
  background-color: var(--p-surface-100);
}

.data-grid__table :deep(.data-grid__actions) {
  display: flex;
  flex-wrap: wrap;
  gap: 0.142857rem;
  justify-content: flex-end;
}

.data-grid__table :deep(.data-grid__btn) {
  width: 1.8rem;
  height: 1.8rem;
  padding: 0;
  margin: 0;
  font-size: 0.75rem;
}

/* Actions-column header: right-align to match the right-aligned cell content below. */
.data-grid__table :deep(thead th.data-grid__th-actions) {
  text-align: right;
}

/* Cell content modifiers — applied by callers to their slotted <td>s. */
.data-grid__table :deep(.data-grid__cell--ellipsis) {
  max-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.data-grid__table :deep(.data-grid__cell--nowrap) {
  white-space: nowrap;
}
</style>
